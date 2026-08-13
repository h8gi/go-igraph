package igraph_test

import (
	"fmt"
	"log"

	igraph "github.com/h8gi/go-igraph"
)

func ExampleNewWeightedBiadjacency() {
	matrix, err := igraph.NewMatrixFromRows([][]float64{
		{1.5, 0, 2},
		{0, 3, 1},
	})
	if err != nil {
		log.Fatal(err)
	}
	result, err := igraph.NewWeightedBiadjacency(matrix, igraph.WeightedBiadjacencyOptions{})
	if err != nil {
		log.Fatal(err)
	}
	defer result.Graph.Close()

	roundTrip, err := result.Graph.Biadjacency(result.Partition, result.Weights)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(roundTrip.RowVertexIDs)
	fmt.Println(roundTrip.ColumnVertexIDs)
	fmt.Println(roundTrip.Matrix.Rows())
	// Output:
	// [0 1]
	// [2 3 4]
	// [[1.5 0 2] [0 3 1]]
}

func ExampleGraph_BipartiteProjections() {
	graph, err := igraph.NewBipartite(
		igraph.BipartitePartition{false, false, false, true, true},
		[]igraph.Edge{{From: 0, To: 3}, {From: 0, To: 4}, {From: 1, To: 3}, {From: 1, To: 4}, {From: 2, To: 4}},
		false,
	)
	if err != nil {
		log.Fatal(err)
	}
	defer graph.Graph.Close()

	projections, err := graph.Graph.BipartiteProjections(graph.Partition)
	if err != nil {
		log.Fatal(err)
	}
	defer projections.False.Graph.Close()
	defer projections.True.Graph.Close()

	fmt.Println(projections.False.SourceVertexIDs)
	fmt.Println(projections.False.Multiplicities)
	fmt.Println(projections.True.SourceVertexIDs)
	fmt.Println(projections.True.Multiplicities)
	// Output:
	// [0 1 2]
	// [2 1 1]
	// [3 4]
	// [2]
}
