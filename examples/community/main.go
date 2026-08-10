package main

import (
	"fmt"
	"log"

	igraph "github.com/h8gi/go-igraph"
)

func main() {
	// Create a graph with two densely connected clusters joined by a single edge
	g, err := igraph.NewGraphFromEdges(8, []igraph.Edge{
		// Cluster 1 (Vertices 0, 1, 2, 3)
		{From: 0, To: 1}, {From: 0, To: 2}, {From: 0, To: 3},
		{From: 1, To: 2}, {From: 2, To: 3},

		// Inter-cluster bridge
		{From: 3, To: 4},

		// Cluster 2 (Vertices 4, 5, 6, 7)
		{From: 4, To: 5}, {From: 4, To: 6}, {From: 4, To: 7},
		{From: 5, To: 6}, {From: 6, To: 7},
	}, false)
	if err != nil {
		log.Fatalf("failed to create graph: %v", err)
	}
	defer g.Close()

	seed := uint64(42)
	partition, err := g.CommunityMultilevel(igraph.MultilevelOptions{
		Seed: &seed,
	})
	if err != nil {
		log.Fatalf("failed to compute multilevel communities: %v", err)
	}

	fmt.Printf("=== Community Detection Demo ===\n")
	fmt.Printf("Detected communities count: %d\n", partition.CommunityCount)
	fmt.Printf("Modularity score: %.4f\n", partition.Modularity)
	fmt.Printf("Vertex membership assignments: %v\n", partition.Membership)
}
