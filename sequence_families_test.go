package igraph

import "testing"

func TestPlanarLatticeFamilies(t *testing.T) {
	for _, test := range []struct {
		name      string
		construct func([]int, bool, bool) (*Graph, error)
		maxDegree int
	}{
		{"hexagonal", NewHexagonalLattice, 3},
		{"triangular", NewTriangularLattice, 6},
	} {
		t.Run(test.name, func(t *testing.T) {
			graph, err := test.construct([]int{2, 3}, false, false)
			graph = cleanupConstructedGraph(t, graph, err)
			vertices, err := graph.VertexCount()
			if err != nil || vertices == 0 {
				t.Fatalf("vertex count = %d, %v", vertices, err)
			}
			degrees, err := graph.Degree(AllVertices(), DegreeOptions{Direction: DirectionAll})
			if err != nil {
				t.Fatal(err)
			}
			for vertex, degree := range degrees {
				if degree > test.maxDegree {
					t.Errorf("degree[%d] = %d, max %d", vertex, degree, test.maxDegree)
				}
			}
			directed, err := test.construct([]int{2}, true, true)
			directed = cleanupConstructedGraph(t, directed, err)
			if isDirected, err := directed.IsDirected(); err != nil || !isDirected {
				t.Errorf("directed = %t, %v", isDirected, err)
			}
			empty, err := test.construct([]int{0}, true, false)
			empty = cleanupConstructedGraph(t, empty, err)
			assertGraphShape(t, empty, 0, 0, true)
		})
	}
}

func TestDeBruijnAndKautzFamilies(t *testing.T) {
	deBruijn, err := NewDeBruijn(2, 2)
	deBruijn = cleanupConstructedGraph(t, deBruijn, err)
	assertGraphShape(t, deBruijn, 4, 8, true)
	degrees, err := deBruijn.Degree(AllVertices(), DegreeOptions{Direction: DirectionOut, CountLoops: true})
	if err != nil {
		t.Fatal(err)
	}
	for vertex, degree := range degrees {
		if degree != 2 {
			t.Errorf("de Bruijn degree[%d] = %d", vertex, degree)
		}
	}

	kautz, err := NewKautz(2, 1)
	kautz = cleanupConstructedGraph(t, kautz, err)
	assertGraphShape(t, kautz, 6, 12, true)

	unit, err := NewDeBruijn(7, 0)
	unit = cleanupConstructedGraph(t, unit, err)
	assertGraphShape(t, unit, 1, 0, true)
	empty, err := NewKautz(0, 2)
	empty = cleanupConstructedGraph(t, empty, err)
	assertGraphShape(t, empty, 0, 0, true)
}

func TestLCFFamily(t *testing.T) {
	graph, err := NewLCF(10, []int{5, -5}, 5)
	graph = cleanupConstructedGraph(t, graph, err)
	assertGraphShape(t, graph, 10, 15, false)
	degrees, err := graph.Degree(AllVertices(), DegreeOptions{Direction: DirectionAll})
	if err != nil {
		t.Fatal(err)
	}
	for vertex, degree := range degrees {
		if degree != 3 {
			t.Errorf("LCF degree[%d] = %d", vertex, degree)
		}
	}
	empty, err := NewLCF(0, nil, 0)
	empty = cleanupConstructedGraph(t, empty, err)
	assertGraphShape(t, empty, 0, 0, false)
}

func TestSequenceFamiliesRejectInvalidInputs(t *testing.T) {
	constructors := []func() (*Graph, error){
		func() (*Graph, error) { return NewHexagonalLattice(nil, false, false) },
		func() (*Graph, error) { return NewTriangularLattice([]int{1, 1, 1, 1}, false, false) },
		func() (*Graph, error) { return NewHexagonalLattice([]int{-1}, false, false) },
		func() (*Graph, error) { return NewDeBruijn(-1, 1) },
		func() (*Graph, error) { return NewDeBruijn(2, int(^uint(0)>>1)) },
		func() (*Graph, error) { return NewKautz(-1, 1) },
		func() (*Graph, error) { return NewKautz(2, int(^uint(0)>>1)) },
		func() (*Graph, error) { return NewLCF(-1, nil, 0) },
		func() (*Graph, error) { return NewLCF(3, nil, -1) },
		func() (*Graph, error) { return NewLCF(0, []int{1}, 1) },
	}
	for index, construct := range constructors {
		if graph, err := construct(); err == nil {
			graph.Close()
			t.Errorf("invalid constructor %d succeeded", index)
		}
	}
}
