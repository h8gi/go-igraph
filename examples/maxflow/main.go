package main

import (
	"fmt"
	"log"

	igraph "github.com/h8gi/go-igraph"
)

func main() {
	// Create a network flow graph with 6 vertices (Source: 0, Sink: 5)
	// Edges represent network links with specified capacities.
	g, err := igraph.NewGraphFromEdges(6, []igraph.Edge{
		{From: 0, To: 1}, // Capacity 16
		{From: 0, To: 2}, // Capacity 13
		{From: 1, To: 2}, // Capacity 10
		{From: 1, To: 3}, // Capacity 12
		{From: 2, To: 1}, // Capacity 4
		{From: 2, To: 4}, // Capacity 14
		{From: 3, To: 2}, // Capacity 9
		{From: 3, To: 5}, // Capacity 20
		{From: 4, To: 3}, // Capacity 7
		{From: 4, To: 5}, // Capacity 4
	}, true)
	if err != nil {
		log.Fatalf("failed to create graph: %v", err)
	}
	defer g.Close()

	capacities := []float64{16, 13, 10, 12, 4, 14, 9, 20, 7, 4}
	res, err := g.MaxFlow(0, 5, capacities)
	if err != nil {
		log.Fatalf("failed to compute max flow: %v", err)
	}

	fmt.Printf("=== Network Maximum Flow Demo ===\n")
	fmt.Printf("Source vertex: 0, Sink vertex: 5\n")
	fmt.Printf("Maximum Flow Value: %.1f\n", res.Value)
	fmt.Printf("Flow along edges: %v\n", res.Flow)
	fmt.Printf("Partition - Source side vertices: %v\n", res.Partition)
	fmt.Printf("Partition - Target side vertices: %v\n", res.Partition2)
}
