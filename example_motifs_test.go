package igraph_test

import (
	"fmt"
	"log"
	"math"

	igraph "github.com/h8gi/go-igraph"
)

func ExampleGraph_MotifsRandesu() {
	graph, err := igraph.NewGraphFromEdges(4, []igraph.Edge{
		{From: 0, To: 1}, {From: 0, To: 2}, {From: 0, To: 3},
		{From: 1, To: 2}, {From: 1, To: 3}, {From: 2, To: 3},
	}, false)
	if err != nil {
		log.Fatal(err)
	}
	defer graph.Close()

	dyads, err := graph.DyadCensus()
	if err != nil {
		log.Fatal(err)
	}
	triangles, err := graph.TrianglesCount()
	if err != nil {
		log.Fatal(err)
	}
	histogram, err := graph.MotifsRandesu(igraph.MotifsRandesuOptions{Size: 3})
	if err != nil {
		log.Fatal(err)
	}
	var histogramTotal float64
	for _, count := range histogram {
		if !math.IsNaN(count) {
			histogramTotal += count
		}
	}

	fmt.Println("mutual dyads:", dyads.Mutual)
	fmt.Println("triangles:", triangles)
	fmt.Println("connected 3-vertex motifs:", histogramTotal)
	// Output:
	// mutual dyads: 6
	// triangles: 4
	// connected 3-vertex motifs: 4
}

func ExampleGraph_Graphlets() {
	graph, err := igraph.NewGraphFromEdges(3, []igraph.Edge{
		{From: 0, To: 1}, {From: 0, To: 2}, {From: 1, To: 2},
	}, false)
	if err != nil {
		log.Fatal(err)
	}
	defer graph.Close()

	basis, err := graph.GraphletsCandidateBasis(nil)
	if err != nil {
		log.Fatal(err)
	}
	coefficients, err := graph.GraphletsProject(basis.Cliques, nil, nil, 0)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("basis:", basis.Cliques)
	fmt.Println("thresholds:", basis.Thresholds)
	fmt.Println("coefficients:", coefficients)
	// Output:
	// basis: [[0 1 2]]
	// thresholds: [1]
	// coefficients: [1]
}
