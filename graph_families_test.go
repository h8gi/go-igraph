package igraph

import (
	"errors"
	"reflect"
	"testing"
)

func TestCommonDeterministicGraphFamilies(t *testing.T) {
	circulant, err := NewCirculant(5, []int{1, -1, 0, 6}, false)
	circulant = cleanupConstructedGraph(t, circulant, err)
	assertGraphShape(t, circulant, 5, 5, false)

	for _, test := range []struct {
		mode     WheelMode
		directed bool
		edges    int
		from, to int
	}{
		{WheelUndirected, false, 8, 0, 1},
		{WheelOut, true, 8, 0, 1},
		{WheelIn, true, 8, 1, 0},
		{WheelMutual, true, 16, 0, 1},
	} {
		wheel, err := NewWheel(5, 0, test.mode)
		wheel = cleanupConstructedGraph(t, wheel, err)
		assertGraphShape(t, wheel, 5, test.edges, test.directed)
		adjacent, err := wheel.AreAdjacent(test.from, test.to)
		if err != nil || !adjacent {
			t.Errorf("wheel adjacency %d->%d = %t, %v", test.from, test.to, adjacent, err)
		}
	}

	petersen, err := NewGeneralizedPetersen(5, 2)
	petersen = cleanupConstructedGraph(t, petersen, err)
	assertGraphShape(t, petersen, 10, 15, false)
	degrees, err := petersen.Degree(AllVertices(), DegreeOptions{Direction: DirectionAll})
	if err != nil {
		t.Fatal(err)
	}
	for vertex, degree := range degrees {
		if degree != 3 {
			t.Errorf("Petersen degree[%d] = %d, want 3", vertex, degree)
		}
	}

	citation, err := NewFullCitation(4, true)
	citation = cleanupConstructedGraph(t, citation, err)
	assertGraphShape(t, citation, 4, 6, true)
	if adjacent, err := citation.AreAdjacent(3, 0); err != nil || !adjacent {
		t.Errorf("citation 3->0 = %t, %v", adjacent, err)
	}
}

func TestMultipartiteAndTuranFamilies(t *testing.T) {
	result, err := NewFullMultipartite([]int{2, 1, 2}, false, DirectionAll)
	if err != nil {
		t.Fatalf("NewFullMultipartite failed: %v", err)
	}
	t.Cleanup(func() { result.Graph.Close() })
	assertGraphShape(t, result.Graph, 5, 8, false)
	if want := []int{0, 0, 1, 2, 2}; !reflect.DeepEqual(result.Parts, want) {
		t.Errorf("parts = %v, want %v", result.Parts, want)
	}
	result.Graph.Close()
	if want := []int{0, 0, 1, 2, 2}; !reflect.DeepEqual(result.Parts, want) {
		t.Errorf("parts changed after Close: %v", result.Parts)
	}

	directed, err := NewFullMultipartite([]int{1, 1}, true, DirectionIn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { directed.Graph.Close() })
	if adjacent, err := directed.Graph.AreAdjacent(1, 0); err != nil || !adjacent {
		t.Errorf("directed multipartite 1->0 = %t, %v", adjacent, err)
	}

	turan, err := NewTuran(5, 2)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { turan.Graph.Close() })
	assertGraphShape(t, turan.Graph, 5, 6, false)
	if want := []int{0, 0, 0, 1, 1}; !reflect.DeepEqual(turan.Parts, want) {
		t.Errorf("Turán parts = %v, want %v", turan.Parts, want)
	}
}

func TestLineGraphProvenanceAndClose(t *testing.T) {
	source, err := NewPath(4, false, false)
	source = cleanupConstructedGraph(t, source, err)
	line, err := source.LineGraph()
	line = cleanupConstructedGraph(t, line, err)
	assertGraphShape(t, line, 3, 2, false)
	for _, endpoints := range [][2]int{{0, 1}, {1, 2}} {
		adjacent, err := line.AreAdjacent(endpoints[0], endpoints[1])
		if err != nil || !adjacent {
			t.Errorf("line graph adjacency %v = %t, %v", endpoints, adjacent, err)
		}
	}
	source.Close()
	assertGraphShape(t, line, 3, 2, false)
	if _, err := source.LineGraph(); !errors.Is(err, ErrClosed) {
		t.Errorf("LineGraph after Close error = %v, want ErrClosed", err)
	}
}

func TestGraphFamiliesEmptyAndInvalid(t *testing.T) {
	emptyGraphs := []func() (*Graph, error){
		func() (*Graph, error) { return NewCirculant(0, []int{1}, true) },
		func() (*Graph, error) { return NewFullCitation(0, false) },
	}
	for index, construct := range emptyGraphs {
		graph, err := construct()
		graph = cleanupConstructedGraph(t, graph, err)
		assertGraphShape(t, graph, 0, 0, index == 0)
	}
	emptyMultipartite, err := NewFullMultipartite(nil, true, DirectionOut)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { emptyMultipartite.Graph.Close() })
	assertGraphShape(t, emptyMultipartite.Graph, 0, 0, true)
	if emptyMultipartite.Parts == nil {
		t.Error("empty multipartite parts are nil")
	}
	emptyTuran, err := NewTuran(0, 2)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { emptyTuran.Graph.Close() })
	if emptyTuran.Parts == nil {
		t.Error("empty Turán parts are nil")
	}
	singleWheel, err := NewWheel(1, 0, WheelUndirected)
	singleWheel = cleanupConstructedGraph(t, singleWheel, err)
	assertGraphShape(t, singleWheel, 1, 0, false)

	invalid := []func() (*Graph, error){
		func() (*Graph, error) { return NewCirculant(-1, nil, false) },
		func() (*Graph, error) { return NewWheel(0, 0, WheelOut) },
		func() (*Graph, error) { return NewWheel(2, 2, WheelOut) },
		func() (*Graph, error) { return NewWheel(2, 0, WheelMode(99)) },
		func() (*Graph, error) { return NewGeneralizedPetersen(2, 1) },
		func() (*Graph, error) { return NewGeneralizedPetersen(5, 0) },
		func() (*Graph, error) { return NewGeneralizedPetersen(5, 3) },
		func() (*Graph, error) { return NewFullCitation(-1, false) },
	}
	for index, construct := range invalid {
		if graph, err := construct(); err == nil {
			graph.Close()
			t.Errorf("invalid constructor %d succeeded", index)
		}
	}
	if result, err := NewFullMultipartite([]int{1, -1}, false, DirectionAll); err == nil {
		result.Graph.Close()
		t.Error("negative multipartite size succeeded")
	}
	if result, err := NewFullMultipartite(nil, false, DirectionMode(99)); err == nil {
		result.Graph.Close()
		t.Error("invalid multipartite direction succeeded")
	}
	if result, err := NewTuran(1, 0); err == nil {
		result.Graph.Close()
		t.Error("zero-part Turán graph succeeded")
	}
}
