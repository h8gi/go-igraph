package igraph_test

import (
	"fmt"
	"log"

	"github.com/h8gi/go-igraph"
)

func ExampleGraph_MaxFlow() {
	// Create a transport network graph with 4 vertices and 5 edges
	g, err := igraph.NewGraphFromEdges(4, []igraph.Edge{
		{From: 0, To: 1},
		{From: 0, To: 2},
		{From: 1, To: 2},
		{From: 1, To: 3},
		{From: 2, To: 3},
	}, true)
	if err != nil {
		log.Fatal(err)
	}
	defer g.Close()

	capacities := []float64{10.0, 10.0, 2.0, 4.0, 8.0}
	res, err := g.MaxFlow(0, 3, capacities)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Max flow: %.1f\n", res.Value)
	// Output:
	// Max flow: 12.0
}

func ExampleGraph_EdgeConnectivity() {
	// Create two triangles connected by a single bridge edge
	g, err := igraph.NewGraphFromEdges(6, []igraph.Edge{
		{From: 0, To: 1}, {From: 1, To: 2}, {From: 2, To: 0},
		{From: 2, To: 3}, // bridge
		{From: 3, To: 4}, {From: 4, To: 5}, {From: 5, To: 3},
	}, false)
	if err != nil {
		log.Fatal(err)
	}
	defer g.Close()

	ec, err := g.EdgeConnectivity(true)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Edge connectivity: %d\n", ec)
	// Output:
	// Edge connectivity: 1
}

func ExampleGraph_AllSTCuts() {
	g, err := igraph.NewGraphFromEdges(4, []igraph.Edge{
		{From: 0, To: 1},
		{From: 0, To: 2},
		{From: 1, To: 3},
		{From: 2, To: 3},
	}, true)
	if err != nil {
		log.Fatal(err)
	}
	defer g.Close()

	cuts, err := g.AllSTCuts(0, 3)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Number of s-t cuts: %d\n", len(cuts))
	// Output:
	// Number of s-t cuts: 4
}
