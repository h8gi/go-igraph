package igraph

import (
	"errors"
	"math"
	"reflect"
	"sync"
	"testing"
)

func testHRGModel(t *testing.T) HRGModel {
	t.Helper()
	model, err := NewHRGModel([]int{-2, 0}, []int{2, 1}, []float64{0.25, 0.75}, []int{2, 1})
	if err != nil {
		t.Fatal(err)
	}
	return model
}

func TestHRGModelCopiesAndValidates(t *testing.T) {
	left := []int{-2, 0}
	model, err := NewHRGModel(left, []int{2, 1}, []float64{0.25, 0.75}, []int{2, 1})
	if err != nil {
		t.Fatal(err)
	}
	left[0] = 0
	got := model.LeftChildren()
	got[0] = 0
	if !reflect.DeepEqual(model.LeftChildren(), []int{-2, 0}) {
		t.Fatal("model storage was mutated")
	}
	if !reflect.DeepEqual(model.RightChildren(), []int{2, 1}) || !reflect.DeepEqual(model.EdgeCounts(), []int{2, 1}) {
		t.Fatal("unexpected model accessors")
	}
	zero, err := NewHRGModel(nil, nil, nil, nil)
	if err != nil || zero.LeafCount() != 1 {
		t.Fatalf("zero model = %#v, %v", zero, err)
	}
	invalid := []struct {
		name        string
		left, right []int
		prob        []float64
		edges       []int
	}{
		{"lengths", []int{0}, nil, []float64{.5}, []int{0}},
		{"duplicate leaf", []int{0}, []int{0}, []float64{.5}, []int{0}},
		{"leaf range", []int{0}, []int{2}, []float64{.5}, []int{0}},
		{"root child", []int{-1}, []int{0}, []float64{.5}, []int{0}},
		{"repeated internal", []int{-2, 0}, []int{-2, 1}, []float64{.5, .5}, []int{0, 0}},
		{"unreachable", []int{0, 1}, []int{1, 2}, []float64{.5, .5}, []int{0, 0}},
		{"probability", []int{0}, []int{1}, []float64{math.NaN()}, []int{0}},
		{"infinite probability", []int{0}, []int{1}, []float64{math.Inf(1)}, []int{0}},
		{"edges", []int{0}, []int{1}, []float64{.5}, []int{-1}},
	}
	for _, tc := range invalid {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := NewHRGModel(tc.left, tc.right, tc.prob, tc.edges); err == nil {
				t.Fatal("expected error")
			}
		})
	}
}

func TestHRGSingletonDendrogram(t *testing.T) {
	model, err := NewHRGModel(nil, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	graph, probabilities, err := model.Dendrogram()
	if err != nil {
		t.Fatal(err)
	}
	defer graph.Close()
	vertices, err := graph.VertexCount()
	if err != nil || vertices != 1 {
		t.Fatalf("vertices = %d, %v", vertices, err)
	}
	if len(probabilities) != 1 || !math.IsNaN(probabilities[0]) {
		t.Fatalf("probabilities = %#v", probabilities)
	}
	if _, err := NewHRGModelFromDendrogram(graph, probabilities); err == nil {
		t.Fatal("pinned igraph should reject singleton dendrogram creation")
	}
}

func TestCHRGResize(t *testing.T) {
	h, err := newCHRG(testHRGModel(t))
	if err != nil {
		t.Fatal(err)
	}
	defer h.close()
	if err := h.resize(3); err != nil {
		t.Fatal(err)
	}
	if model, err := h.model(); err != nil || model.LeafCount() != 3 {
		t.Fatalf("resized model = %#v, %v", model, err)
	}
	if err := h.resize(-1); err == nil {
		t.Fatal("expected upstream resize error")
	}
}

func TestCHRGFailureCleanupAndExtraction(t *testing.T) {
	forced := errors.New("forced failure")
	if h, err := newCHRGWithInitializer(testHRGModel(t), func(*cHRG, int) int { return 4 }); err == nil || h != nil {
		t.Fatal("expected initialization failure")
	}
	h, err := newCHRG(testHRGModel(t))
	if err != nil {
		t.Fatal(err)
	}
	defer h.close()
	valid := hrgModelReaders{
		size:          func(*cHRG) (int, error) { return 3, nil },
		left:          func(*cHRG) ([]int, error) { return []int{-2, 0}, nil },
		right:         func(*cHRG) ([]int, error) { return []int{2, 1}, nil },
		edges:         func(*cHRG) ([]int, error) { return []int{2, 1}, nil },
		probabilities: func(*cHRG, int) ([]float64, error) { return []float64{.25, .75}, nil },
	}
	badSize := valid
	badSize.size = func(*cHRG) (int, error) { return 0, nil }
	if _, err := h.modelWithReaders(badSize); err == nil {
		t.Fatal("expected invalid size")
	}
	badSize.size = func(*cHRG) (int, error) { return 0, forced }
	if _, err := h.modelWithReaders(badSize); !errors.Is(err, forced) {
		t.Fatalf("size error = %v", err)
	}
	for name, mutate := range map[string]func(*hrgModelReaders){
		"left":  func(r *hrgModelReaders) { r.left = func(*cHRG) ([]int, error) { return nil, forced } },
		"right": func(r *hrgModelReaders) { r.right = func(*cHRG) ([]int, error) { return nil, forced } },
		"edges": func(r *hrgModelReaders) { r.edges = func(*cHRG) ([]int, error) { return nil, forced } },
	} {
		readers := valid
		mutate(&readers)
		if _, err := h.modelWithReaders(readers); !errors.Is(err, forced) {
			t.Fatalf("%s conversion error = %v", name, err)
		}
	}
	badReal := valid
	badReal.probabilities = func(*cHRG, int) ([]float64, error) { return nil, forced }
	if _, err := h.modelWithReaders(badReal); !errors.Is(err, forced) {
		t.Fatalf("real conversion error = %v", err)
	}
}

func TestHRGDendrogramRoundTrip(t *testing.T) {
	model := testHRGModel(t)
	graph, probabilities, err := model.Dendrogram()
	if err != nil {
		t.Fatal(err)
	}
	defer graph.Close()
	vertices, err := graph.VertexCount()
	if err != nil {
		t.Fatal(err)
	}
	if vertices != 5 || len(probabilities) != vertices {
		t.Fatalf("vertices/probabilities = %d/%d", vertices, len(probabilities))
	}
	directed, err := graph.IsDirected()
	if err != nil || !directed {
		t.Fatalf("directed = %t, %v", directed, err)
	}
	edgeCount, err := graph.EdgeCount()
	if err != nil || edgeCount != 4 {
		t.Fatalf("edges = %d, %v", edgeCount, err)
	}
	converted, err := NewHRGModelFromDendrogram(graph, probabilities)
	if err != nil {
		t.Fatal(err)
	}
	if converted.LeafCount() != model.LeafCount() || !reflect.DeepEqual(converted.Probabilities(), model.Probabilities()) {
		t.Fatalf("round trip = %#v, want leaf count and probabilities from %#v", converted, model)
	}
	if err := graph.Close(); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(converted.Probabilities(), model.Probabilities()) {
		t.Fatal("model changed after graph close")
	}
	if _, err := NewHRGModelFromDendrogram(graph, probabilities); err != ErrClosed {
		t.Fatalf("closed graph error = %v", err)
	}
}

func TestHRGDendrogramConcurrentReadsAndCloseRace(t *testing.T) {
	model := testHRGModel(t)
	graph, probabilities, err := model.Dendrogram()
	if err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			converted, err := NewHRGModelFromDendrogram(graph, probabilities)
			if err != nil && !errors.Is(err, ErrClosed) {
				t.Errorf("conversion error = %v", err)
			}
			if err == nil && converted.LeafCount() != 3 {
				t.Errorf("leaf count = %d", converted.LeafCount())
			}
		}()
	}
	wg.Add(1)
	go func() { defer wg.Done(); _ = graph.Close() }()
	wg.Wait()
}

func TestHRGDendrogramRejectsInvalidInputs(t *testing.T) {
	if _, err := NewHRGModelFromDendrogram(nil, nil); err != ErrClosed {
		t.Fatalf("nil graph error = %v", err)
	}
	graph, err := NewGraphFromEdges(2, []Edge{{0, 1}}, false)
	if err != nil {
		t.Fatal(err)
	}
	defer graph.Close()
	if _, err := NewHRGModelFromDendrogram(graph, []float64{.5}); err == nil {
		t.Fatal("expected alignment error")
	}
	if _, err := NewHRGModelFromDendrogram(graph, []float64{.5, .5}); err == nil {
		t.Fatal("expected internal probability error")
	}
	directed, err := NewGraphFromEdges(3, []Edge{{0, 1}, {1, 2}}, false)
	if err != nil {
		t.Fatal(err)
	}
	defer directed.Close()
	if _, err := NewHRGModelFromDendrogram(directed, []float64{math.NaN(), math.NaN(), .5}); err == nil {
		t.Fatal("expected upstream tree error")
	}
	bad := HRGModel{left: []int{0}}
	if graph, _, err := bad.Dendrogram(); err == nil || graph != nil {
		t.Fatal("expected invalid model error")
	}
}
