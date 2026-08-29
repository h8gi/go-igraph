package igraph_test

import (
	"fmt"

	igraph "github.com/h8gi/go-igraph"
)

func ExampleUnionMany_graphAlgebra() {
	path, _ := igraph.NewGraphFromEdges(3, []igraph.Edge{{From: 0, To: 1}, {From: 1, To: 2}}, false)
	chord, _ := igraph.NewGraphFromEdges(3, []igraph.Edge{{From: 0, To: 2}}, false)
	defer path.Close()
	defer chord.Close()

	union, err := igraph.UnionMany([]*igraph.Graph{path, chord, path}, nil)
	if err != nil {
		panic(err)
	}
	defer union.Graph.Close()
	power, err := union.Graph.GraphPower(2, false)
	if err != nil {
		panic(err)
	}
	defer power.Graph.Close()

	vertices, _ := power.Graph.VertexCount()
	edges, _ := power.Graph.EdgeCount()
	fmt.Printf("power: %d vertices, %d edges\n", vertices, edges)
	fmt.Printf("first input vertices: %v\n", union.Inputs[0].Vertices.OldToNew)
	// Output:
	// power: 3 vertices, 3 edges
	// first input vertices: [0 1 2]
}

func ExampleGraph_Mycielskian_workflow() {
	edge, _ := igraph.NewGraphFromEdges(2, []igraph.Edge{{From: 0, To: 1}}, false)
	defer edge.Close()
	result, err := edge.Mycielskian(1)
	if err != nil {
		panic(err)
	}
	defer result.Graph.Close()

	vertices, _ := result.Graph.VertexCount()
	edges, _ := result.Graph.EdgeCount()
	fmt.Printf("Mycielski graph: %d vertices, %d edges\n", vertices, edges)
	fmt.Printf("source descendants: %v\n", result.SourceToResult)
	// Output:
	// Mycielski graph: 5 vertices, 5 edges
	// source descendants: [[0 2] [1 3]]
}
