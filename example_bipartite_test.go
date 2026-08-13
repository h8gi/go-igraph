package igraph_test

import (
	"fmt"
	"log"

	igraph "github.com/h8gi/go-igraph"
)

func ExampleNewBipartiteGNM() {
	seed := uint64(42)
	result, err := igraph.NewBipartiteGNM(3, 2, 4, igraph.BipartiteRandomOptions{Seed: &seed})
	if err != nil {
		log.Fatal(err)
	}
	defer result.Graph.Close()

	vertices, _ := result.Graph.VertexCount()
	edges, _ := result.Graph.EdgeCount()
	fmt.Println(vertices, edges)
	fmt.Println(result.Partition)
	// Output:
	// 5 4
	// [false false false true true]
}

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

func ExampleGraph_MaximumBipartiteMatching() {
	graph, err := igraph.NewBipartite(
		igraph.BipartitePartition{false, false, true, true},
		[]igraph.Edge{{From: 0, To: 2}, {From: 0, To: 3}, {From: 1, To: 2}, {From: 1, To: 3}},
		false,
	)
	if err != nil {
		log.Fatal(err)
	}
	defer graph.Graph.Close()

	matching, err := graph.Graph.MaximumBipartiteMatching(
		graph.Partition,
		igraph.BipartiteMatchingOptions{Weights: []float64{1, 5, 4, 1}},
	)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(matching.Size, matching.Weight)
	fmt.Println(matching.Pairs)
	// Output:
	// 2 9
	// [{0 3} {1 2}]
}
