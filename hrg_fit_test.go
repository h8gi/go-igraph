package igraph

import (
	"errors"
	"reflect"
	"sync"
	"testing"
)

func newHRGFitGraph(t *testing.T) *Graph {
	t.Helper()
	g, err := NewGraphFromEdges(6, []Edge{{0, 1}, {1, 2}, {2, 3}, {3, 4}, {4, 5}, {5, 0}, {0, 3}, {1, 4}}, false)
	if err != nil {
		t.Fatal(err)
	}
	return g
}

func TestFitHRGSeedReplayAndWarmStart(t *testing.T) {
	g := newHRGFitGraph(t)
	defer g.Close()
	seed := uint64(42)
	first, err := g.FitHRG(HRGFitOptions{Steps: 100, Seed: &seed})
	if err != nil {
		t.Fatal(err)
	}
	second, err := g.FitHRG(HRGFitOptions{Steps: 100, Seed: &seed})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("seed replay differs: %#v != %#v", first, second)
	}
	before := first.Probabilities()
	warm, err := g.FitHRG(HRGFitOptions{Steps: 25, Seed: &seed, StartingModel: &first})
	if err != nil {
		t.Fatal(err)
	}
	if warm.LeafCount() != 6 || !reflect.DeepEqual(first.Probabilities(), before) {
		t.Fatal("warm fit changed starting model or alignment")
	}
	if err := g.Close(); err != nil {
		t.Fatal(err)
	}
	if first.LeafCount() != 6 || len(first.Probabilities()) != 5 {
		t.Fatal("Go-owned fit result changed after graph close")
	}
	if _, err := g.FitHRG(HRGFitOptions{}); !errors.Is(err, ErrClosed) {
		t.Fatalf("closed error = %v", err)
	}
}

func TestFitHRGValidation(t *testing.T) {
	if _, err := (*Graph)(nil).FitHRG(HRGFitOptions{}); !errors.Is(err, ErrClosed) {
		t.Fatalf("nil error = %v", err)
	}
	g := newHRGFitGraph(t)
	defer g.Close()
	seed := uint64(42)
	if model, err := g.FitHRG(HRGFitOptions{Steps: 0, Seed: &seed}); err != nil || model.LeafCount() != 6 {
		t.Fatalf("zero-step fit = %#v, %v", model, err)
	}
	if _, err := g.FitHRG(HRGFitOptions{Steps: -1}); err == nil {
		t.Fatal("expected negative step error")
	}
	wrong := testHRGModel(t)
	if _, err := g.FitHRG(HRGFitOptions{StartingModel: &wrong}); err == nil {
		t.Fatal("expected starting-model alignment error")
	}
	for name, edges := range map[string][]Edge{"loop": {{0, 0}}, "parallel": {{0, 1}, {0, 1}}} {
		t.Run(name, func(t *testing.T) {
			graph, err := NewGraphFromEdges(2, edges, false)
			if err != nil {
				t.Fatal(err)
			}
			defer graph.Close()
			if _, err := graph.FitHRG(HRGFitOptions{}); err == nil {
				t.Fatal("expected simplicity error")
			}
		})
	}
	directed, err := NewGraphFromEdges(2, []Edge{{0, 1}}, true)
	if err != nil {
		t.Fatal(err)
	}
	defer directed.Close()
	if _, err := directed.FitHRG(HRGFitOptions{}); err == nil {
		t.Fatal("expected directed error")
	}
	empty, err := NewGraph()
	if err != nil {
		t.Fatal(err)
	}
	defer empty.Close()
	if _, err := empty.FitHRG(HRGFitOptions{}); err == nil {
		t.Fatal("pinned igraph should reject empty fit")
	}
}

func TestFitHRGFailuresCloseTemporaryModel(t *testing.T) {
	g := newHRGFitGraph(t)
	defer g.Close()
	forced := errors.New("forced failure")
	base := defaultHRGFitAdapters()
	fresh := base
	fresh.fresh = func(int) (*cHRG, error) { return nil, forced }
	if _, err := g.fitHRG(HRGFitOptions{}, &fresh); !errors.Is(err, forced) {
		t.Fatalf("fresh error = %v", err)
	}
	for name, mutate := range map[string]func(*hrgFitAdapters){
		"run":     func(a *hrgFitAdapters) { a.run = func(*Graph, *cHRG, bool, int) error { return forced } },
		"extract": func(a *hrgFitAdapters) { a.extract = func(*cHRG) (HRGModel, error) { return HRGModel{}, forced } },
	} {
		t.Run(name, func(t *testing.T) {
			closed := 0
			a := base
			mutate(&a)
			a.close = func(h *cHRG) { closed++; h.close() }
			if _, err := g.fitHRG(HRGFitOptions{}, &a); !errors.Is(err, forced) {
				t.Fatalf("error = %v", err)
			}
			if closed != 1 {
				t.Fatalf("closes = %d", closed)
			}
		})
	}
}

func TestFitHRGConcurrentCalls(t *testing.T) {
	g := newHRGFitGraph(t)
	defer g.Close()
	var wg sync.WaitGroup
	for range 4 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			seed := uint64(42)
			model, err := g.FitHRG(HRGFitOptions{Steps: 100, Seed: &seed})
			if err != nil {
				t.Errorf("fit error = %v", err)
				return
			}
			if model.LeafCount() != 6 {
				t.Errorf("leaves = %d", model.LeafCount())
			}
		}()
	}
	wg.Wait()
}
