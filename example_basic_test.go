package igraph_test

import (
	"fmt"
	"log"

	igraph "github.com/h8gi/go-igraph"
)

func ExampleNewGraphFromEdges() {
	// Create a simple directed graph with 3 vertices and 2 edges
	graph, err := igraph.NewGraphFromEdges(3, []igraph.Edge{
		{From: 0, To: 1},
		{From: 1, To: 2},
	}, true)
	if err != nil {
		log.Fatal(err)
	}
	defer graph.Close()

	vc, err := graph.VertexCount()
	if err != nil {
		log.Fatal(err)
	}
	ec, err := graph.EdgeCount()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Vertices: %d, Edges: %d\n", vc, ec)
	// Output:
	// Vertices: 3, Edges: 2
}

func ExampleGraph_DeleteVertices() {
	graph, err := igraph.NewGraphFromEdges(4, []igraph.Edge{
		{From: 0, To: 1}, {From: 1, To: 2},
		{From: 2, To: 3}, {From: 1, To: 1},
	}, false)
	if err != nil {
		log.Fatal(err)
	}
	defer graph.Close()

	vertices, _ := igraph.VertexIDs(1)
	mapping, err := graph.DeleteVertices(vertices)
	if err != nil {
		log.Fatal(err)
	}

	edges, err := graph.Edges()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(mapping.Vertices.OldToNew, mapping.Vertices.NewToOld)
	fmt.Println(mapping.Edges.OldToNew, mapping.Edges.NewToOld)
	fmt.Println(edges)
	// Output:
	// [0 -1 1 2] [0 2 3]
	// [-1 -1 0 -1] [2]
	// [{1 2}]
}

func ExampleGraph_Decompose() {
	graph, err := igraph.NewGraphFromEdges(6, []igraph.Edge{
		{From: 0, To: 1}, {From: 1, To: 2}, {From: 2, To: 0},
		{From: 3, To: 4},
	}, false)
	if err != nil {
		log.Fatal(err)
	}
	defer graph.Close()

	components, err := graph.Decompose(igraph.DecomposeOptions{})
	if err != nil {
		log.Fatal(err)
	}

	for _, component := range components {
		vertices, _ := component.VertexCount()
		edges, _ := component.EdgeCount()
		fmt.Println(vertices, edges)
		_ = component.Close()
	}
	// Output:
	// 3 3
	// 2 1
	// 1 0
}
