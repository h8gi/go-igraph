// Command cycles demonstrates a cycle-analysis pipeline.
package main

import (
	"fmt"
	"log"

	igraph "github.com/h8gi/go-igraph"
)

func main() {
	graph, err := igraph.NewGraphFromEdges(5, []igraph.Edge{
		{From: 0, To: 1}, {From: 1, To: 2}, {From: 2, To: 0},
		{From: 2, To: 3}, {From: 3, To: 4}, {From: 4, To: 2},
	}, true)
	if err != nil {
		log.Fatal(err)
	}
	defer graph.Close()

	witness, err := graph.FindCycle(igraph.DirectionOut)
	if err != nil {
		log.Fatal(err)
	}
	cycles, err := graph.SimpleCycles(igraph.SimpleCycleOptions{
		Direction:  igraph.DirectionOut,
		MaxResults: 1,
	})
	if err != nil {
		log.Fatal(err)
	}
	basis, err := graph.MinimumCycleBasis(igraph.MinimumCycleBasisOptions{})
	if err != nil {
		log.Fatal(err)
	}
	feedback, err := graph.FeedbackEdgeSet(igraph.FeedbackEdgeOptions{})
	if err != nil {
		log.Fatal(err)
	}

	reduced, err := graph.Clone()
	if err != nil {
		log.Fatal(err)
	}
	defer reduced.Close()
	selector, err := igraph.EdgeIDs(feedback...)
	if err != nil {
		log.Fatal(err)
	}
	if _, err := reduced.DeleteEdges(selector); err != nil {
		log.Fatal(err)
	}
	order, err := reduced.TopologicalSort(igraph.DirectionOut)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("cycle witness length: %d\n", len(witness.Vertices))
	fmt.Printf("bounded cycles: %d (truncated=%t)\n", len(cycles.Cycles), cycles.Truncated)
	fmt.Printf("minimum cycle-basis rank: %d\n", len(basis))
	fmt.Printf("feedback edges: %d\n", len(feedback))
	fmt.Printf("topological vertices after removal: %d\n", len(order))
}
