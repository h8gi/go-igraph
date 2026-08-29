package main

import (
	"fmt"
	"log"

	igraph "github.com/h8gi/go-igraph"
)

func main() {
	path, err := igraph.NewGraphFromEdges(3, []igraph.Edge{{From: 0, To: 1}, {From: 1, To: 2}}, false)
	if err != nil {
		log.Fatal(err)
	}
	defer path.Close()
	chord, err := igraph.NewGraphFromEdges(3, []igraph.Edge{{From: 0, To: 2}}, false)
	if err != nil {
		log.Fatal(err)
	}
	defer chord.Close()

	union, err := igraph.UnionMany([]*igraph.Graph{path, chord, path}, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer union.Graph.Close()
	power, err := union.Graph.GraphPower(2, false)
	if err != nil {
		log.Fatal(err)
	}
	defer power.Graph.Close()
	selected, err := igraph.VertexIDs(0, 2)
	if err != nil {
		log.Fatal(err)
	}
	subgraph, err := power.Graph.InducedSubgraph(selected)
	if err != nil {
		log.Fatal(err)
	}
	defer subgraph.Graph.Close()
	mycielski, err := chord.Mycielskian(1)
	if err != nil {
		log.Fatal(err)
	}
	defer mycielski.Graph.Close()

	printShape("union", union.Graph)
	printShape("power", power.Graph)
	printShape("induced subgraph", subgraph.Graph)
	printShape("Mycielski", mycielski.Graph)
	fmt.Printf("union input-0 vertex mapping: %v\n", union.Inputs[0].Vertices.OldToNew)
	fmt.Printf("subgraph vertex mapping: %v\n", subgraph.Vertices.OldToNew)
	fmt.Printf("Mycielski source descendants: %v\n", mycielski.SourceToResult)
}

func printShape(name string, graph *igraph.Graph) {
	vertices, err := graph.VertexCount()
	if err != nil {
		log.Fatal(err)
	}
	edges, err := graph.EdgeCount()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s: %d vertices, %d edges\n", name, vertices, edges)
}
