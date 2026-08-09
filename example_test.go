package igraph_test

import (
	"fmt"
	"log"

	igraph "github.com/h8gi/go-igraph"
)

func ExampleGraph_Distances() {
	graph, err := igraph.NewGraphFromEdges(4, []igraph.Edge{
		{From: 0, To: 1}, {From: 1, To: 2},
		{From: 0, To: 2}, {From: 2, To: 3},
	}, true)
	if err != nil {
		log.Fatal(err)
	}
	defer graph.Close()
	sources, _ := igraph.VertexIDs(0, 2)
	targets, _ := igraph.VertexIDs(3, 1, 3)
	distances, err := graph.Distances(sources, targets, igraph.PathOptions{
		Direction: igraph.DirectionOut,
		Weights:   []float64{2, 3, 10, 1},
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(distances.Rows())
	// Output: [[6 2 6] [1 +Inf 1]]
}

func ExampleGraph_BreadthFirstSearch() {
	graph, err := igraph.NewGraphFromEdges(5, []igraph.Edge{
		{From: 0, To: 1}, {From: 0, To: 2},
		{From: 1, To: 3}, {From: 2, To: 4},
	}, true)
	if err != nil {
		log.Fatal(err)
	}
	defer graph.Close()
	restriction, _ := igraph.VertexIDs(0, 2, 4)
	result, err := graph.BreadthFirstSearch(igraph.BFSOptions{
		Roots:       []int{0},
		Direction:   igraph.DirectionOut,
		Restriction: restriction,
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(result.Order)
	// Output: [0 2 4]
}

func ExampleGraph_Closeness() {
	graph, err := igraph.NewPath(4, false, false)
	if err != nil {
		log.Fatal(err)
	}
	defer graph.Close()
	vertices, _ := igraph.VertexIDs(2, 0, 2)
	result, err := graph.Closeness(vertices, igraph.DistanceCentralityOptions{
		Normalized: true,
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(result.Scores)
	fmt.Println(result.ReachableCounts, result.AllReachable)
	// Output:
	// [0.75 0.5 0.75]
	// [3 3 3] true
}

func ExampleGraph_PageRank_personalized() {
	graph, err := igraph.NewGraphFromEdges(3, []igraph.Edge{{From: 0, To: 1}, {From: 0, To: 2}}, true)
	if err != nil {
		log.Fatal(err)
	}
	defer graph.Close()
	reset, _ := igraph.VertexIDs(1, 2)
	zero := 0.0
	result, err := graph.PageRank(igraph.AllVertices(), igraph.PageRankOptions{
		Damping:       &zero,
		ResetVertices: &reset,
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(result.Scores)
	// Output: [0 0.5 0.5]
}
