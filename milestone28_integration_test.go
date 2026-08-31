package igraph_test

import (
	"sync"
	"testing"

	"github.com/h8gi/go-igraph"
)

func TestMilestone28AdvancedLayoutWorkflow(t *testing.T) {
	graph, err := igraph.NewGraphFromEdges(5, []igraph.Edge{
		{From: 0, To: 1}, {From: 0, To: 1}, {From: 1, To: 2},
		{From: 2, To: 3}, {From: 3, To: 4}, {From: 4, To: 0}, {From: 2, To: 2},
	}, true)
	if err != nil {
		t.Fatal(err)
	}
	defer graph.Close()
	seed := uint64(28)
	circle, err := graph.LayoutCircle(nil)
	if err != nil {
		t.Fatal(err)
	}
	classic := []struct {
		name string
		call func() (igraph.Matrix, error)
	}{
		{"Davidson-Harel", func() (igraph.Matrix, error) {
			return graph.LayoutDavidsonHarel(igraph.DavidsonHarelOptions{Seed: &seed, MaxIter: 1, FineIter: 1, InitialCoordinates: &circle})
		}},
		{"GEM", func() (igraph.Matrix, error) {
			return graph.LayoutGEM(igraph.GEMOptions{Seed: &seed, MaxIter: 20, InitialCoordinates: &circle})
		}},
		{"Graphopt", func() (igraph.Matrix, error) {
			return graph.LayoutGraphopt(igraph.GraphoptOptions{Seed: &seed, NIter: 2, InitialCoordinates: &circle})
		}},
	}
	for _, layout := range classic {
		coordinates, err := layout.call()
		if err != nil {
			t.Fatalf("%s failed: %v", layout.name, err)
		}
		assertAdvancedLayout(t, layout.name, coordinates, 5)
	}

	drl, err := graph.LayoutDrL(igraph.DrLOptions{Seed: &seed, Weights: []float64{1, 1, 1, 1, 1, 1, 1}})
	if err != nil {
		t.Fatalf("DrL failed: %v", err)
	}
	aligned, err := graph.AlignLayout(drl)
	if err != nil {
		t.Fatalf("alignment failed: %v", err)
	}
	assertAdvancedLayout(t, "aligned DrL", aligned, 5)

	tree, err := igraph.NewGraphFromEdges(4, []igraph.Edge{{From: 0, To: 1}, {From: 1, To: 2}, {From: 1, To: 3}}, true)
	if err != nil {
		t.Fatal(err)
	}
	defer tree.Close()
	roots, err := tree.TreeLayoutRoots(igraph.DegOut, igraph.TreeRootByDegree)
	if err != nil {
		t.Fatal(err)
	}
	treeCoordinates, err := tree.LayoutReingoldTilford(igraph.DegOut, roots)
	if err != nil {
		t.Fatal(err)
	}
	merged, err := igraph.MergeLayoutsDLA(
		[]*igraph.Graph{graph, tree},
		[]igraph.Matrix{aligned, treeCoordinates},
		igraph.DLAMergeOptions{Seed: &seed},
	)
	if err != nil {
		t.Fatalf("DLA merge failed: %v", err)
	}
	assertAdvancedLayout(t, "integrated DLA", merged, 9)
	graph.Close()
	tree.Close()
	assertAdvancedLayout(t, "owned integrated result", merged, 9)
}

func TestMilestone28ConcurrentCloseRaces(t *testing.T) {
	graph, err := igraph.NewRing(12, false, false)
	if err != nil {
		t.Fatal(err)
	}
	seed := uint64(29)
	circle, err := graph.LayoutCircle(nil)
	if err != nil {
		t.Fatal(err)
	}
	calls := []func() error{
		func() error { _, err := graph.LayoutDrL(igraph.DrLOptions{Seed: &seed}); return err },
		func() error { _, err := graph.AlignLayout(circle); return err },
		func() error { _, err := graph.TreeLayoutRoots(igraph.DegAll, igraph.TreeRootByDegree); return err },
	}
	var wait sync.WaitGroup
	for _, call := range calls {
		call := call
		wait.Add(1)
		go func() {
			defer wait.Done()
			if err := call(); err != nil && err != igraph.ErrClosed {
				t.Errorf("close-race layout failed: %v", err)
			}
		}()
	}
	graph.Close()
	wait.Wait()
}
