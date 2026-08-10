package igraph_test

import (
	"fmt"
	"log"

	igraph "github.com/h8gi/go-igraph"
)

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

func ExampleGraph_VertexBetweenness() {
	graph, err := igraph.NewPath(3, false, false)
	if err != nil {
		log.Fatal(err)
	}
	defer graph.Close()

	scores, err := graph.VertexBetweenness(igraph.AllVertices(), igraph.BetweennessOptions{
		Normalized: false,
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(scores)
	// Output: [0 1 0]
}
