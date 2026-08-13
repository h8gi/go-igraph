// Command motifs demonstrates census, motif, and graphlet analysis.
package main

import (
	"fmt"
	"log"
	"math"

	igraph "github.com/h8gi/go-igraph"
)

func main() {
	edges := []igraph.Edge{
		{From: 0, To: 1}, {From: 0, To: 2}, {From: 1, To: 2}, {From: 1, To: 3},
		{From: 1, To: 4}, {From: 2, To: 3}, {From: 2, To: 4}, {From: 3, To: 4},
	}
	weights := []float64{2, 2, 3, 1, 1, 4, 4, 4}
	graph, err := igraph.NewGraphFromEdges(5, edges, false)
	if err != nil {
		log.Fatal(err)
	}
	defer graph.Close()

	dyads, err := graph.DyadCensus()
	if err != nil {
		log.Fatal(err)
	}
	triangles, err := graph.TrianglesList()
	if err != nil {
		log.Fatal(err)
	}
	histogram, err := graph.MotifsRandesu(igraph.MotifsRandesuOptions{Size: 3})
	if err != nil {
		log.Fatal(err)
	}
	var motifCount float64
	for _, count := range histogram {
		if !math.IsNaN(count) {
			motifCount += count
		}
	}
	graphlets, err := graph.Graphlets(weights, 1000)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("dyads: mutual=%d asymmetric=%d null=%d\n", dyads.Mutual, dyads.Asymmetric, dyads.Null)
	fmt.Printf("triangles: %d\n", len(triangles))
	fmt.Printf("connected 3-vertex motifs: %.0f\n", motifCount)
	fmt.Printf("graphlets: %d\n", len(graphlets.Cliques))
	fmt.Printf("leading graphlet: %v (mu=%.5f)\n", graphlets.Cliques[0], graphlets.Mu[0])
}
