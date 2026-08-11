// Command cliques demonstrates bounded clique-family analysis.
package main

import (
	"fmt"
	"log"

	igraph "github.com/h8gi/go-igraph"
)

func main() {
	graph, err := igraph.NewGraphFromEdges(5, []igraph.Edge{
		{From: 0, To: 1}, {From: 0, To: 2}, {From: 1, To: 2}, {From: 3, To: 4},
	}, false)
	if err != nil {
		log.Fatal(err)
	}
	defer graph.Close()

	largest, err := graph.LargestCliques(2)
	if err != nil {
		log.Fatal(err)
	}
	maximal, err := graph.MaximalCliques(igraph.VertexSetEnumerationOptions{MaxResults: 2})
	if err != nil {
		log.Fatal(err)
	}
	weighted, err := graph.MaximumWeightCliques([]int{1, 1, 1, 10, 10}, 1)
	if err != nil {
		log.Fatal(err)
	}
	independent, err := graph.LargestIndependentVertexSets(6)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("largest cliques: %v (truncated=%t)\n", largest.Sets, largest.Truncated)
	fmt.Printf("maximal cliques: %v (truncated=%t)\n", maximal.Sets, maximal.Truncated)
	fmt.Printf("maximum-weight cliques: %v\n", weighted.Sets)
	fmt.Printf("largest independent sets: %d\n", len(independent.Sets))
}
