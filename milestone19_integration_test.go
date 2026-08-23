package igraph_test

import (
	"reflect"
	"sync"
	"testing"

	igraph "github.com/h8gi/go-igraph"
)

func TestMilestone19ConstructionWorkflows(t *testing.T) {
	t.Run("weighted matrix round trip and ownership", func(t *testing.T) {
		want := [][]float64{{0, 2, 0}, {0, 0, 3}, {4, 0, 0}}
		matrix, err := igraph.NewMatrixFromRows(want)
		if err != nil {
			t.Fatal(err)
		}
		weighted, err := igraph.NewWeightedAdjacency(matrix, igraph.AdjacencyOptions{})
		if err != nil {
			t.Fatal(err)
		}
		adjacency, err := weighted.Graph.AdjacencyMatrix(weighted.Weights, igraph.AdjacencyMatrixOptions{})
		if err != nil {
			weighted.Graph.Close()
			t.Fatal(err)
		}
		weighted.Graph.Close()
		if !reflect.DeepEqual(adjacency.Rows(), want) {
			t.Errorf("round trip = %v, want %v", adjacency.Rows(), want)
		}
		if len(weighted.Weights) != 3 {
			t.Errorf("weights after Close = %v", weighted.Weights)
		}
	})

	t.Run("degree realization analysis", func(t *testing.T) {
		want := []int{3, 2, 2, 1}
		graph, err := igraph.RealizeDegreeSequence(want, nil, igraph.DegreeSequenceRealizationOptions{})
		if err != nil {
			t.Fatal(err)
		}
		defer graph.Close()
		got, err := graph.Degree(igraph.AllVertices(), igraph.DegreeOptions{Direction: igraph.DirectionAll, CountLoops: true})
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("degrees = %v, want %v", got, want)
		}
	})

	t.Run("tree encoding round trips", func(t *testing.T) {
		wantPrufer := []int{3, 3, 4}
		graph, err := igraph.NewTreeFromPrufer(wantPrufer)
		if err != nil {
			t.Fatal(err)
		}
		gotPrufer, err := graph.PruferSequence()
		graph.Close()
		if err != nil || !reflect.DeepEqual(gotPrufer, wantPrufer) {
			t.Errorf("Prüfer = %v, %v", gotPrufer, err)
		}

		parents := []int{igraph.NoParent, 0, 0, 1, 1}
		tree, err := igraph.NewTreeFromParents(parents, igraph.TreeOut)
		if err != nil {
			t.Fatal(err)
		}
		defer tree.Close()
		bfs, err := tree.BreadthFirstSearch(igraph.BFSOptions{Roots: []int{0}, Direction: igraph.DirectionOut})
		if err != nil || !reflect.DeepEqual(bfs.Parents, parents) {
			t.Errorf("parents = %v, %v", bfs.Parents, err)
		}
	})

	t.Run("family composed with analysis", func(t *testing.T) {
		graph, err := igraph.NewGeneralizedPetersen(5, 2)
		if err != nil {
			t.Fatal(err)
		}
		defer graph.Close()
		degrees, err := graph.Degree(igraph.AllVertices(), igraph.DegreeOptions{Direction: igraph.DirectionAll})
		if err != nil {
			t.Fatal(err)
		}
		for vertex, degree := range degrees {
			if degree != 3 {
				t.Errorf("degree[%d] = %d", vertex, degree)
			}
		}
	})
}

func TestMilestone19ConcurrentConstructionAndConversion(t *testing.T) {
	matrix, err := igraph.NewMatrixFromRows([][]float64{{0, 1}, {1, 0}})
	if err != nil {
		t.Fatal(err)
	}
	shared, err := igraph.NewAdjacency(matrix, igraph.AdjacencyOptions{Mode: igraph.AdjacencyUndirected})
	if err != nil {
		t.Fatal(err)
	}
	defer shared.Close()

	var wait sync.WaitGroup
	errors := make(chan error, 24)
	for worker := 0; worker < 8; worker++ {
		wait.Add(1)
		go func() {
			defer wait.Done()
			for iteration := 0; iteration < 3; iteration++ {
				converted, err := shared.AdjacencyMatrix(nil, igraph.AdjacencyMatrixOptions{})
				if err != nil {
					errors <- err
					continue
				}
				if rows, columns := converted.Dims(); rows != 2 || columns != 2 {
					t.Errorf("dims = %dx%d", rows, columns)
				}
				independent, err := igraph.NewCirculant(7, []int{1, 2}, false)
				if err != nil {
					errors <- err
					continue
				}
				independent.Close()
			}
		}()
	}
	wait.Wait()
	close(errors)
	for err := range errors {
		t.Error(err)
	}
}
