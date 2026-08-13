package main

import (
	"fmt"
	"log"

	igraph "github.com/h8gi/go-igraph"
)

func main() {
	matrix, err := igraph.NewMatrixFromRows([][]float64{
		{5, 1},
		{4, 6},
		{0, 3},
	})
	if err != nil {
		log.Fatal(err)
	}
	network, err := igraph.NewWeightedBiadjacency(matrix, igraph.WeightedBiadjacencyOptions{})
	if err != nil {
		log.Fatal(err)
	}
	defer network.Graph.Close()

	projections, err := network.Graph.BipartiteProjections(network.Partition)
	if err != nil {
		log.Fatal(err)
	}
	defer projections.False.Graph.Close()
	defer projections.True.Graph.Close()

	fmt.Println("Affiliation projection")
	fmt.Printf("  row source IDs: %v\n", projections.False.SourceVertexIDs)
	fmt.Printf("  shared affiliations: %v\n", projections.False.Multiplicities)

	matching, err := network.Graph.MaximumBipartiteMatching(
		network.Partition,
		igraph.BipartiteMatchingOptions{Weights: network.Weights},
	)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Assignment matching")
	fmt.Printf("  pairs: %v\n", matching.Pairs)
	fmt.Printf("  total weight: %.0f\n", matching.Weight)
}
