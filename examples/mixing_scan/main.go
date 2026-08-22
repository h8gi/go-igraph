// Command mixing_scan demonstrates attributed mixing analysis and aligned
// two-snapshot local scans.
package main

import (
	"fmt"
	"log"
	"os"

	igraph "github.com/h8gi/go-igraph"
)

func main() {
	graph, err := igraph.NewGraphFromEdges(3, []igraph.Edge{{From: 0, To: 1}, {From: 0, To: 2}, {From: 2, To: 1}}, true)
	check(err)
	if err := graph.SetVertexStringAttributes("category", []string{"a", "b", "a"}); err != nil {
		log.Fatal(err)
	}
	if err := graph.SetEdgeNumericAttributes("weight", []float64{2, 3, 5}); err != nil {
		log.Fatal(err)
	}

	file, err := os.CreateTemp("", "go-igraph-mixing-*.graphml")
	check(err)
	path := file.Name()
	defer os.Remove(path)
	check(graph.WriteGraphML(file, false))
	check(graph.Close())
	_, err = file.Seek(0, 0)
	check(err)
	imported, err := igraph.ReadGraphML(file, 0)
	check(err)
	check(file.Close())
	defer imported.Close()

	categories, err := imported.VertexStringAttributes("category")
	check(err)
	weights, err := imported.EdgeNumericAttributes("weight")
	check(err)
	distribution, err := imported.CategoricalJointDistribution(igraph.StringCategories(categories), igraph.CategoryJointDistributionOptions{Weights: weights, Directed: true})
	check(err)
	scan, err := imported.LocalScan(igraph.LocalScanOptions{Direction: igraph.DirectionOut, Weights: weights})
	check(err)
	fmt.Printf("mixing: %v\nscan: %v\n", distribution.Matrix.Rows(), scan)

	previous, err := igraph.NewGraphFromEdges(3, []igraph.Edge{{From: 0, To: 1}, {From: 1, To: 2}}, false)
	check(err)
	defer previous.Close()
	current, err := igraph.NewGraphFromEdges(3, []igraph.Edge{{From: 0, To: 1}, {From: 0, To: 2}}, false)
	check(err)
	defer current.Close()
	cross, err := previous.CrossGraphLocalScan(current, igraph.LocalScanOptions{Radius: 1, Direction: igraph.DirectionAll})
	check(err)
	fmt.Printf("cross-snapshot scan: %v\n", cross)
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
