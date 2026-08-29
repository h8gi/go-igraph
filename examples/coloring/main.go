package main

import (
	"fmt"
	"log"

	igraph "github.com/h8gi/go-igraph"
)

func main() {
	graph, err := igraph.NewGraphFromEdges(5, []igraph.Edge{{From: 0, To: 1}, {From: 1, To: 2}, {From: 2, To: 3}, {From: 3, To: 4}, {From: 4, To: 0}}, false)
	if err != nil {
		log.Fatal(err)
	}
	defer graph.Close()
	colors, err := graph.GreedyVertexColoring(igraph.ColoringDSatur)
	if err != nil {
		log.Fatal(err)
	}
	valid, err := graph.IsVertexColoring(colors)
	if err != nil {
		log.Fatal(err)
	}
	lowerBound, err := graph.CliqueNumber()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("colors=%v valid=%v clique-lower-bound=%d\n", colors, valid, lowerBound)
}
