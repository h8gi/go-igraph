package igraph_test

import (
	"math"
	"sync"
	"testing"

	igraph "github.com/h8gi/go-igraph"
)

func TestMilestone20StructuralInspectionPipeline(t *testing.T) {
	graph, err := igraph.NewGraphFromEdges(3, []igraph.Edge{
		{From: 0, To: 0}, {From: 0, To: 1}, {From: 0, To: 1},
		{From: 1, To: 0}, {From: 1, To: 2}, {From: 2, To: 1},
	}, true)
	if err != nil {
		t.Fatal(err)
	}
	defer graph.Close()
	loops, err := graph.HasLoopEdges()
	if err != nil || !loops {
		t.Fatalf("loops = %v, %v", loops, err)
	}
	multiple, err := graph.HasMultipleEdges()
	if err != nil || !multiple {
		t.Fatalf("multiple = %v, %v", multiple, err)
	}
	mutual, err := graph.HasMutualEdges(false)
	if err != nil || !mutual {
		t.Fatalf("mutual = %v, %v", mutual, err)
	}
	mean, err := graph.MeanDegree(true)
	if err != nil || mean <= 0 {
		t.Fatalf("mean degree = %v, %v", mean, err)
	}
	reciprocity, err := graph.Reciprocity(igraph.ReciprocityRatio, true)
	if err != nil || math.IsNaN(reciprocity) || reciprocity <= 0 {
		t.Fatalf("reciprocity = %v, %v", reciprocity, err)
	}

	if _, err := graph.SimplifyInPlace(igraph.SimplifyOptions{RemoveLoops: true, RemoveParallel: true}); err != nil {
		t.Fatal(err)
	}
	loops, _ = graph.HasLoopEdges()
	multiple, _ = graph.HasMultipleEdges()
	if loops || multiple {
		t.Fatalf("simplified graph still non-simple: loops=%v multiple=%v", loops, multiple)
	}
	if _, err := graph.ConvertToUndirectedInPlace(igraph.UndirectedConversionCollapse, nil); err != nil {
		t.Fatal(err)
	}
	weights := []float64{2, 3}
	diversity, err := graph.Diversity(igraph.AllVertices(), weights)
	if err != nil || len(diversity) != 3 {
		t.Fatalf("diversity = %v, %v", diversity, err)
	}
}

func TestMilestone20StructureAndMatrixPipeline(t *testing.T) {
	graph, err := igraph.NewGraphFromEdges(5, []igraph.Edge{
		{From: 0, To: 1}, {From: 1, To: 2}, {From: 2, To: 3},
		{From: 3, To: 0}, {From: 0, To: 2}, {From: 3, To: 4},
	}, false)
	if err != nil {
		t.Fatal(err)
	}
	defer graph.Close()

	forestEdges, err := graph.MinimumSpanningForest([]float64{4, 3, 2, 8, 1, 5})
	if err != nil {
		t.Fatal(err)
	}
	selector, err := igraph.EdgeIDs(forestEdges...)
	if err != nil {
		t.Fatal(err)
	}
	forestResult, err := graph.EdgeSubgraph(selector, false)
	if err != nil {
		t.Fatal(err)
	}
	defer forestResult.Graph.Close()
	forest, err := forestResult.Graph.IsForest(igraph.DirectionAll)
	if err != nil || !forest.IsForest {
		t.Fatalf("minimum spanning forest = %#v, %v", forest, err)
	}

	unfolded, err := graph.UnfoldTree([]int{0}, igraph.DirectionAll)
	if err != nil {
		t.Fatal(err)
	}
	defer unfolded.Graph.Close()
	unfoldedTree, err := unfolded.Graph.IsTree(igraph.DirectionAll)
	if err != nil || !unfoldedTree.IsTree {
		t.Fatalf("unfolded tree = %#v, %v", unfoldedTree, err)
	}

	order, err := graph.MaximumCardinalityOrder()
	if err != nil {
		t.Fatal(err)
	}
	chordal, err := graph.Chordality(igraph.ChordalityOptions{Ordering: order.Vertices, Complete: true})
	if err != nil {
		t.Fatal(err)
	}
	defer chordal.Completion.Close()
	completed, err := chordal.Completion.Chordality(igraph.ChordalityOptions{})
	if err != nil || !completed.Chordal {
		t.Fatalf("completion = %#v, %v", completed, err)
	}

	laplacian, err := graph.Laplacian(igraph.LaplacianOptions{})
	if err != nil {
		t.Fatal(err)
	}
	adjacency, err := graph.AdjacencyMatrix(nil, igraph.AdjacencyMatrixOptions{})
	if err != nil {
		t.Fatal(err)
	}
	rows, columns := laplacian.Dims()
	ar, ac := adjacency.Dims()
	if rows != 5 || columns != 5 || ar != rows || ac != columns {
		t.Fatalf("matrix dimensions Laplacian=%dx%d adjacency=%dx%d", rows, columns, ar, ac)
	}
	graph.Close()
	if got := laplacian.Rows(); len(got) != 5 {
		t.Fatalf("Laplacian after source close = %v", got)
	}
	if count, err := chordal.Completion.VertexCount(); err != nil || count != 5 {
		t.Fatalf("completion after source close = %d, %v", count, err)
	}
}

func TestMilestone20ConcurrentReadOnlyAnalysis(t *testing.T) {
	graph, err := igraph.NewGraphFromEdges(5, []igraph.Edge{{0, 1}, {1, 2}, {2, 3}, {3, 0}, {3, 4}}, false)
	if err != nil {
		t.Fatal(err)
	}
	defer graph.Close()

	var wait sync.WaitGroup
	errors := make(chan error, 8*3*5)
	for worker := 0; worker < 8; worker++ {
		wait.Add(1)
		go func() {
			defer wait.Done()
			for iteration := 0; iteration < 3; iteration++ {
				if _, err := graph.EdgeMultiplicities(igraph.AllEdges()); err != nil {
					errors <- err
				}
				if _, err := graph.AverageNearestNeighborDegree(igraph.AllVertices(), igraph.NearestNeighborDegreeOptions{Direction: igraph.DirectionAll, NeighborDegreeDirection: igraph.DirectionAll}); err != nil {
					errors <- err
				}
				if _, err := graph.IsForest(igraph.DirectionAll); err != nil {
					errors <- err
				}
				if _, err := graph.Chordality(igraph.ChordalityOptions{}); err != nil {
					errors <- err
				}
				if _, err := graph.Laplacian(igraph.LaplacianOptions{}); err != nil {
					errors <- err
				}
			}
		}()
	}
	wait.Wait()
	close(errors)
	for err := range errors {
		t.Error(err)
	}
}
