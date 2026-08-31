package igraph_test

import (
	"fmt"
	"log"

	igraph "github.com/h8gi/go-igraph"
)

func ExampleGraph_centralityAndLocalClustering() {
	graph, err := igraph.NewGraphFromEdges(3, []igraph.Edge{
		{From: 0, To: 1}, {From: 1, To: 2}, {From: 2, To: 0},
	}, false)
	if err != nil {
		log.Fatal(err)
	}
	defer graph.Close()
	weights := []float64{1, 1, 1}
	sources, _ := igraph.VertexIDs(0)
	targets, _ := igraph.VertexIDs(2)

	vertices, err := graph.VertexBetweennessSubset(igraph.AllVertices(), sources, targets, igraph.SubsetBetweennessOptions{})
	if err != nil {
		log.Fatal(err)
	}
	edges, err := graph.EdgeBetweennessSubset(igraph.AllEdges(), sources, targets, igraph.SubsetBetweennessOptions{})
	if err != nil {
		log.Fatal(err)
	}
	constraint, err := graph.BurtConstraint(igraph.AllVertices(), weights)
	if err != nil {
		log.Fatal(err)
	}
	convergence, err := graph.EdgeConvergenceDegree()
	if err != nil {
		log.Fatal(err)
	}
	barrat, err := graph.BarratTransitivity(igraph.AllVertices(), weights, igraph.TransitivityNaN)
	if err != nil {
		log.Fatal(err)
	}
	clustering, err := graph.EdgeClustering(igraph.AllEdges(), igraph.EdgeClusteringOptions{CycleSize: 3, Normalize: true})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(vertices)
	fmt.Println(edges)
	fmt.Println(constraint)
	fmt.Println(len(convergence.Convergence) == len(convergence.InputSetSizes) && len(convergence.Convergence) == len(convergence.OutputSetSizes))
	fmt.Println(barrat)
	fmt.Println(clustering)
	// Output:
	// [0 0 0]
	// [0 0 0.5]
	// [1.125 1.125 1.125]
	// true
	// [1 1 1]
	// [1 1 1]
}
