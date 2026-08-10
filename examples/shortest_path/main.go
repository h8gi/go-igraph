package main

import (
	"fmt"
	"log"

	igraph "github.com/h8gi/go-igraph"
)

func main() {
	// Create a weighted graph with 5 vertices and 6 directed edges
	g, err := igraph.NewGraphFromEdges(5, []igraph.Edge{
		{From: 0, To: 1}, // Weight: 2
		{From: 0, To: 2}, // Weight: 5
		{From: 1, To: 2}, // Weight: 1
		{From: 1, To: 3}, // Weight: 4
		{From: 2, To: 3}, // Weight: 1
		{From: 3, To: 4}, // Weight: 3
	}, true)
	if err != nil {
		log.Fatalf("failed to create graph: %v", err)
	}
	defer g.Close()

	weights := []float64{2.0, 5.0, 1.0, 4.0, 1.0, 3.0}
	pathRes, err := g.ShortestPath(0, 4, igraph.PathOptions{
		Direction: igraph.DirectionOut,
		Weights:   weights,
	})
	if err != nil {
		log.Fatalf("failed to compute shortest path: %v", err)
	}

	fmt.Printf("=== Shortest Path Demo ===\n")
	fmt.Printf("Shortest path vertices from 0 to 4: %v\n", pathRes.Vertices)
	fmt.Printf("Shortest path edge IDs: %v\n", pathRes.Edges)

	sourceIDs := []int{0, 1}
	targetIDs := []int{3, 4}
	sources, _ := igraph.VertexIDs(sourceIDs...)
	targets, _ := igraph.VertexIDs(targetIDs...)
	distMat, err := g.Distances(sources, targets, igraph.PathOptions{
		Direction: igraph.DirectionOut,
		Weights:   weights,
	})
	if err != nil {
		log.Fatalf("failed to compute distance matrix: %v", err)
	}

	fmt.Printf("Distance matrix [sources x targets]:\n")
	for i, row := range distMat.Rows() {
		fmt.Printf(" Source %d: %v\n", sourceIDs[i], row)
	}
}
