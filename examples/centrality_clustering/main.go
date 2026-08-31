package main

import (
	"fmt"
	"log"

	igraph "github.com/h8gi/go-igraph"
)

func main() {
	graph, err := igraph.NewGraphFromEdges(4, []igraph.Edge{
		{From: 0, To: 1}, {From: 1, To: 2}, {From: 2, To: 0}, {From: 2, To: 3},
	}, false)
	if err != nil {
		log.Fatal(err)
	}
	defer graph.Close()
	weights := []float64{1, 2, 3, 4}
	sources, _ := igraph.VertexIDs(0)
	targets, _ := igraph.VertexIDs(3)

	vertexSubset, err := graph.VertexBetweennessSubset(igraph.AllVertices(), sources, targets, igraph.SubsetBetweennessOptions{})
	if err != nil {
		log.Fatal(err)
	}
	edgeSubset, err := graph.EdgeBetweennessSubset(igraph.AllEdges(), sources, targets, igraph.SubsetBetweennessOptions{})
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
	barrat, err := graph.BarratTransitivity(igraph.AllVertices(), weights, igraph.TransitivityZero)
	if err != nil {
		log.Fatal(err)
	}
	clustering, err := graph.EdgeClustering(igraph.AllEdges(), igraph.EdgeClusteringOptions{CycleSize: 3, Normalize: true})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("subset betweenness: %d vertex and %d edge scores\n", len(vertexSubset), len(edgeSubset))
	fmt.Printf("Burt constraint: %d scores\n", len(constraint))
	fmt.Printf("edge convergence aligned: %t\n", len(convergence.Convergence) == len(convergence.InputSetSizes) && len(convergence.Convergence) == len(convergence.OutputSetSizes))
	fmt.Printf("Barrat transitivity: %d scores\n", len(barrat))
	fmt.Printf("triangle edge clustering: %d scores\n", len(clustering))
}
