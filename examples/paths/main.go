package main

import (
	"fmt"
	"log"

	igraph "github.com/h8gi/go-igraph"
)

func main() {
	graph, err := igraph.NewGraphFromEdges(5, []igraph.Edge{
		{From: 0, To: 1}, {From: 0, To: 2}, {From: 1, To: 3},
		{From: 2, To: 3}, {From: 3, To: 4},
	}, true)
	if err != nil {
		log.Fatal(err)
	}
	defer graph.Close()

	alternatives, err := graph.KShortestPaths(0, 4, 2, igraph.PathOptions{Direction: igraph.DirectionOut})
	if err != nil {
		log.Fatal(err)
	}
	for i, path := range alternatives {
		fmt.Printf("route %d: %v\n", i+1, path.Vertices)
	}

	reachable, err := graph.Reachability(igraph.DirectionOut)
	if err != nil {
		log.Fatal(err)
	}
	for source, vertices := range reachable.Reachable {
		fmt.Printf("reachable from %d: %v\n", source, vertices)
	}
}
