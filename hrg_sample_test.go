package igraph

import (
	"errors"
	"reflect"
	"sync"
	"testing"
)

func closeSampledGraphs(t *testing.T, graphs []*Graph) {
	t.Helper()
	for _, graph := range graphs {
		if err := graph.Close(); err != nil {
			t.Fatal(err)
		}
	}
}

func sampledEdgeLists(t *testing.T, graphs []*Graph) [][]Edge {
	t.Helper()
	result := make([][]Edge, len(graphs))
	for i, graph := range graphs {
		edges, err := graph.Edges()
		if err != nil {
			t.Fatal(err)
		}
		result[i] = edges
	}
	return result
}

func TestHRGSampleSeedReplayAndIndependentOwnership(t *testing.T) {
	model := testHRGModel(t)
	seed := uint64(42)
	first, err := model.Sample(4, &seed)
	if err != nil {
		t.Fatal(err)
	}
	defer closeSampledGraphs(t, first)
	second, err := model.Sample(4, &seed)
	if err != nil {
		t.Fatal(err)
	}
	defer closeSampledGraphs(t, second)
	if first == nil || len(first) != 4 || !reflect.DeepEqual(sampledEdgeLists(t, first), sampledEdgeLists(t, second)) {
		t.Fatalf("seeded samples differ")
	}
	for _, graph := range first {
		directed, err := graph.IsDirected()
		if err != nil {
			t.Fatal(err)
		}
		loops, err := graph.HasLoopEdges()
		if err != nil {
			t.Fatal(err)
		}
		multiple, err := graph.HasMultipleEdges()
		if err != nil {
			t.Fatal(err)
		}
		vertices, err := graph.VertexCount()
		if err != nil {
			t.Fatal(err)
		}
		if directed || loops || multiple || vertices != model.LeafCount() {
			t.Fatalf("sample topology = directed %v, loops %v, multiple %v, vertices %d", directed, loops, multiple, vertices)
		}
	}
	if err := first[0].Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := first[1].VertexCount(); err != nil {
		t.Fatalf("closing sibling affected graph: %v", err)
	}
	if _, err := second[0].VertexCount(); err != nil {
		t.Fatalf("closing another call affected graph: %v", err)
	}
}

func TestHRGSampleProbabilityBoundariesAndMinimumModels(t *testing.T) {
	seed := uint64(7)
	for name, probability := range map[string]float64{"zero": 0, "one": 1} {
		t.Run(name, func(t *testing.T) {
			model, err := NewHRGModel([]int{0}, []int{1}, []float64{probability}, []int{0})
			if err != nil {
				t.Fatal(err)
			}
			graphs, err := model.Sample(1, &seed)
			if err != nil {
				t.Fatal(err)
			}
			defer closeSampledGraphs(t, graphs)
			vertices, err := graphs[0].VertexCount()
			if err != nil {
				t.Fatal(err)
			}
			edges, err := graphs[0].EdgeCount()
			if err != nil {
				t.Fatal(err)
			}
			if vertices != 2 || edges != int(probability) {
				t.Fatalf("vertices, edges = %d, %d", vertices, edges)
			}
		})
	}
	if graphs, err := (HRGModel{}).Sample(2, &seed); graphs != nil || err == nil {
		t.Fatalf("one-leaf sample = %v, %v", graphs, err)
	}
}

func TestHRGSampleValidationAndInvalidModel(t *testing.T) {
	model := testHRGModel(t)
	if graphs, err := model.Sample(0, nil); graphs != nil || err == nil {
		t.Fatalf("zero count = %v, %v", graphs, err)
	}
	invalid := HRGModel{left: []int{0}}
	if graphs, err := invalid.Sample(1, nil); graphs != nil || err == nil {
		t.Fatalf("invalid model = %v, %v", graphs, err)
	}
}

func TestHRGSampleFailuresCleanUp(t *testing.T) {
	model := testHRGModel(t)
	forced := errors.New("forced failure")
	base := defaultHRGSampleAdapters()
	for name, mutate := range map[string]func(*hrgSampleAdapters){
		"model-init": func(a *hrgSampleAdapters) { a.model = func(HRGModel) (*cHRG, error) { return nil, forced } },
		"list-init":  func(a *hrgSampleAdapters) { a.newList = func() (*graphList, error) { return nil, forced } },
		"upstream":   func(a *hrgSampleAdapters) { a.run = func(*cHRG, *graphList, int) error { return forced } },
		"extraction": func(a *hrgSampleAdapters) { a.take = func(*graphList) ([]*Graph, error) { return nil, forced } },
	} {
		t.Run(name, func(t *testing.T) {
			a := base
			modelCloses, listCloses := 0, 0
			a.closeModel = func(h *cHRG) { modelCloses++; h.close() }
			a.closeList = func(list *graphList) { listCloses++; list.close() }
			mutate(&a)
			if _, err := model.sample(2, nil, &a); !errors.Is(err, forced) {
				t.Fatalf("error = %v", err)
			}
			wantModel, wantList := 1, 1
			if name == "model-init" {
				wantModel, wantList = 0, 0
			}
			if name == "list-init" {
				wantList = 0
			}
			if modelCloses != wantModel || listCloses != wantList {
				t.Fatalf("closes = %d, %d; want %d, %d", modelCloses, listCloses, wantModel, wantList)
			}
		})
	}
}

func TestHRGSamplePartialExtractionClosesAdoptedPrefix(t *testing.T) {
	model := testHRGModel(t)
	forced := errors.New("forced partial extraction")
	a := defaultHRGSampleAdapters()
	var adopted *Graph
	a.take = func(list *graphList) ([]*Graph, error) {
		return list.takeGraphsWithHooks(graphListExtractionHooks{afterAdopt: func(index int, graph *Graph) error {
			if index == 0 {
				adopted = graph
				return forced
			}
			return nil
		}})
	}
	if graphs, err := model.sample(3, nil, &a); graphs != nil || !errors.Is(err, forced) {
		t.Fatalf("partial extraction = %v, %v", graphs, err)
	}
	if adopted == nil {
		t.Fatal("no adopted prefix")
	}
	if _, err := adopted.VertexCount(); !errors.Is(err, ErrClosed) {
		t.Fatalf("adopted prefix error = %v", err)
	}
}

func TestHRGSampleCountMismatchClosesReturnedGraphs(t *testing.T) {
	model := testHRGModel(t)
	graph, err := NewGraphFromEdges(1, nil, false)
	if err != nil {
		t.Fatal(err)
	}
	a := defaultHRGSampleAdapters()
	a.take = func(*graphList) ([]*Graph, error) { return []*Graph{graph}, nil }
	if graphs, err := model.sample(2, nil, &a); graphs != nil || err == nil {
		t.Fatalf("count mismatch = %v, %v", graphs, err)
	}
	if _, err := graph.VertexCount(); !errors.Is(err, ErrClosed) {
		t.Fatalf("mismatched result error = %v", err)
	}
}

func TestHRGSampleConcurrentCalls(t *testing.T) {
	model := testHRGModel(t)
	seed := uint64(42)
	var wg sync.WaitGroup
	for range 6 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			graphs, err := model.Sample(3, &seed)
			if err != nil {
				t.Errorf("sample: %v", err)
				return
			}
			for _, graph := range graphs {
				_ = graph.Close()
			}
		}()
	}
	wg.Wait()
}
