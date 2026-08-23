package igraph

import (
	"errors"
	"reflect"
	"testing"
)

func TestPruferRoundTripAndOwnership(t *testing.T) {
	want := []int{3, 3, 3, 4}
	graph, err := NewTreeFromPrufer(want)
	graph = cleanupConstructedGraph(t, graph, err)
	assertGraphShape(t, graph, 6, 5, false)

	got, err := graph.PruferSequence()
	if err != nil {
		t.Fatalf("PruferSequence failed: %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("PruferSequence = %v, want %v", got, want)
	}
	graph.Close()
	if !reflect.DeepEqual(got, want) {
		t.Errorf("sequence changed after graph close: %v", got)
	}
	if _, err := graph.PruferSequence(); !errors.Is(err, ErrClosed) {
		t.Errorf("PruferSequence after Close error = %v, want ErrClosed", err)
	}

	empty, err := NewTreeFromPrufer(nil)
	empty = cleanupConstructedGraph(t, empty, err)
	assertGraphShape(t, empty, 2, 1, false)
	sequence, err := empty.PruferSequence()
	if err != nil || sequence == nil || len(sequence) != 0 {
		t.Errorf("two-vertex PruferSequence = %#v, %v, want non-nil empty, nil", sequence, err)
	}
}

func TestNewTreeFromParents(t *testing.T) {
	parents := []int{NoParent, 0, 0, NoParent, 3}
	for _, test := range []struct {
		mode     TreeMode
		directed bool
		from     int
		to       int
	}{
		{TreeOut, true, 0, 1},
		{TreeIn, true, 1, 0},
		{TreeUndirected, false, 0, 1},
	} {
		graph, err := NewTreeFromParents(parents, test.mode)
		graph = cleanupConstructedGraph(t, graph, err)
		assertGraphShape(t, graph, 5, 3, test.directed)
		adjacent, err := graph.AreAdjacent(test.from, test.to)
		if err != nil || !adjacent {
			t.Errorf("AreAdjacent(%d, %d) = %t, %v", test.from, test.to, adjacent, err)
		}
	}

	roundTripParents := []int{NoParent, 0, 0, 1, 1}
	roundTrip, err := NewTreeFromParents(roundTripParents, TreeOut)
	roundTrip = cleanupConstructedGraph(t, roundTrip, err)
	traversal, err := roundTrip.BreadthFirstSearch(BFSOptions{Roots: []int{0}, Direction: DirectionOut})
	if err != nil {
		t.Fatalf("BreadthFirstSearch failed: %v", err)
	}
	if !reflect.DeepEqual(traversal.Parents, roundTripParents) {
		t.Errorf("parent-vector round trip = %v, want %v", traversal.Parents, roundTripParents)
	}

	for name, parents := range map[string][]int{
		"empty":     {},
		"singleton": {NoParent},
	} {
		t.Run(name, func(t *testing.T) {
			graph, err := NewTreeFromParents(parents, TreeOut)
			graph = cleanupConstructedGraph(t, graph, err)
			assertGraphShape(t, graph, len(parents), 0, true)
		})
	}
}

func TestRegularAndSymmetricTrees(t *testing.T) {
	symmetric, err := NewSymmetricTree([]int{2, 3}, TreeOut)
	symmetric = cleanupConstructedGraph(t, symmetric, err)
	assertGraphShape(t, symmetric, 9, 8, true)
	degrees, err := symmetric.Degree(AllVertices(), DegreeOptions{Direction: DirectionOut})
	if err != nil || !reflect.DeepEqual(degrees, []int{2, 3, 3, 0, 0, 0, 0, 0, 0}) {
		t.Errorf("symmetric out-degrees = %v, %v", degrees, err)
	}

	singleton, err := NewSymmetricTree(nil, TreeUndirected)
	singleton = cleanupConstructedGraph(t, singleton, err)
	assertGraphShape(t, singleton, 1, 0, false)

	regular, err := NewRegularTree(2, 3, TreeUndirected)
	regular = cleanupConstructedGraph(t, regular, err)
	assertGraphShape(t, regular, 10, 9, false)
	degrees, err = regular.Degree(AllVertices(), DegreeOptions{Direction: DirectionAll})
	if err != nil {
		t.Fatalf("regular Degree failed: %v", err)
	}
	if degrees[0] != 3 || degrees[1] != 3 || degrees[9] != 1 {
		t.Errorf("regular degrees = %v", degrees)
	}
}

func TestTreeConstructionRejectsInvalidInputs(t *testing.T) {
	invalidParents := map[string][]int{
		"negative sentinel": {-2},
		"out of range":      {1},
		"self parent":       {0},
		"cycle":             {1, 0},
	}
	for name, parents := range invalidParents {
		t.Run(name, func(t *testing.T) {
			if graph, err := NewTreeFromParents(parents, TreeOut); err == nil {
				graph.Close()
				t.Error("NewTreeFromParents succeeded")
			}
		})
	}

	constructors := []func() (*Graph, error){
		func() (*Graph, error) { return NewTreeFromPrufer([]int{3}) },
		func() (*Graph, error) { return NewTreeFromParents(nil, TreeMode(99)) },
		func() (*Graph, error) { return NewSymmetricTree([]int{0}, TreeOut) },
		func() (*Graph, error) { return NewSymmetricTree([]int{int(^uint(0) >> 1), 2}, TreeOut) },
		func() (*Graph, error) { return NewSymmetricTree(nil, TreeMode(99)) },
		func() (*Graph, error) { return NewRegularTree(0, 2, TreeOut) },
		func() (*Graph, error) { return NewRegularTree(1, 1, TreeOut) },
		func() (*Graph, error) { return NewRegularTree(1, 2, TreeMode(99)) },
		func() (*Graph, error) { return NewRegularTree(int(^uint(0)>>1), 2, TreeOut) },
	}
	for index, construct := range constructors {
		if graph, err := construct(); err == nil {
			graph.Close()
			t.Errorf("invalid constructor %d succeeded", index)
		}
	}

	cycle, err := NewRing(3, false, false)
	cycle = cleanupConstructedGraph(t, cycle, err)
	if _, err := cycle.PruferSequence(); err == nil {
		t.Error("PruferSequence accepted a cycle")
	}
	singleton, err := NewGraphFromEdges(1, nil, false)
	singleton = cleanupConstructedGraph(t, singleton, err)
	if _, err := singleton.PruferSequence(); err == nil {
		t.Error("PruferSequence accepted a singleton")
	}
}
