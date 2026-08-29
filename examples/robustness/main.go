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
	}, false)
	if err != nil {
		log.Fatal(err)
	}
	defer graph.Close()

	articulation, err := graph.ArticulationPoints()
	if err != nil {
		log.Fatal(err)
	}
	biconnected, err := graph.IsBiconnected()
	if err != nil {
		log.Fatal(err)
	}
	blocks, err := graph.CohesiveBlocks()
	if err != nil {
		log.Fatal(err)
	}
	defer blocks.BlockTree.Close()
	bond, err := graph.BondPercolation([]int{0, 1, 2, 3, 4, 5})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("biconnected: %v\n", biconnected)
	fmt.Printf("articulation points: %v\n", articulation)
	fmt.Printf("cohesive blocks: %v\n", blocks.Blocks)
	fmt.Printf("cohesion: %v, parents: %v\n", blocks.Cohesion, blocks.Parents)
	fmt.Printf("bond giant-component curve: %v\n", bond.GiantComponentSizes)
}
