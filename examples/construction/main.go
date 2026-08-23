package main

import (
	"fmt"
	"log"

	igraph "github.com/h8gi/go-igraph"
)

func main() {
	matrix, err := igraph.NewMatrixFromRows([][]float64{
		{0, 2, 0},
		{0, 0, 3},
		{4, 0, 0},
	})
	if err != nil {
		log.Fatal(err)
	}

	weighted, err := igraph.NewWeightedAdjacency(matrix, igraph.AdjacencyOptions{})
	if err != nil {
		log.Fatal(err)
	}
	defer weighted.Graph.Close()

	roundTrip, err := weighted.Graph.AdjacencyMatrix(weighted.Weights, igraph.AdjacencyMatrixOptions{})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("weighted adjacency round trip: %v\n", roundTrip.Rows())

	tree, err := igraph.NewTreeFromPrufer([]int{3, 3, 4})
	if err != nil {
		log.Fatal(err)
	}
	defer tree.Close()
	encoding, err := tree.PruferSequence()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Prüfer round trip: %v\n", encoding)

	petersen, err := igraph.NewGeneralizedPetersen(5, 2)
	if err != nil {
		log.Fatal(err)
	}
	defer petersen.Close()
	degrees, err := petersen.Degree(igraph.AllVertices(), igraph.DegreeOptions{Direction: igraph.DirectionAll})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Petersen degrees: %v\n", degrees)
}
