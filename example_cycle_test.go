package igraph_test

import (
	"fmt"
	"log"

	igraph "github.com/h8gi/go-igraph"
)

func ExampleGraph_SimpleCycles() {
	graph, err := igraph.NewGraphFromEdges(5, []igraph.Edge{
		{From: 0, To: 1}, {From: 1, To: 2}, {From: 2, To: 0},
		{From: 2, To: 3}, {From: 3, To: 4}, {From: 4, To: 2},
	}, true)
	if err != nil {
		log.Fatal(err)
	}
	defer graph.Close()

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

	fmt.Println("returned cycles:", len(cycles.Cycles))
	fmt.Println("more cycles exist:", cycles.Truncated)
	fmt.Println("cycle-basis rank:", len(basis))
	// Output:
	// returned cycles: 1
	// more cycles exist: true
	// cycle-basis rank: 2
}

func ExampleGraph_FeedbackEdgeSet() {
	graph, err := igraph.NewGraphFromEdges(4, []igraph.Edge{
		{From: 0, To: 1}, {From: 1, To: 2}, {From: 2, To: 0}, {From: 2, To: 3},
	}, true)
	if err != nil {
		log.Fatal(err)
	}
	defer graph.Close()

	feedback, err := graph.FeedbackEdgeSet(igraph.FeedbackEdgeOptions{})
	if err != nil {
		log.Fatal(err)
	}
	selector, err := igraph.EdgeIDs(feedback...)
	if err != nil {
		log.Fatal(err)
	}
	if _, err := graph.DeleteEdges(selector); err != nil {
		log.Fatal(err)
	}
	acyclic, err := graph.IsAcyclic()
	if err != nil {
		log.Fatal(err)
	}
	order, err := graph.TopologicalSort(igraph.DirectionOut)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("feedback edges:", len(feedback))
	fmt.Println("acyclic after removal:", acyclic)
	fmt.Println("topological vertices:", len(order))
	// Output:
	// feedback edges: 1
	// acyclic after removal: true
	// topological vertices: 4
}
