package igraph_test

import (
	"fmt"
	"log"

	igraph "github.com/h8gi/go-igraph"
)

func ExampleGraph_CategoricalJointDistribution() {
	graph, err := igraph.NewGraphFromEdges(3, []igraph.Edge{{From: 0, To: 1}, {From: 0, To: 2}, {From: 2, To: 1}}, true)
	if err != nil {
		log.Fatal(err)
	}
	defer graph.Close()
	categories := igraph.StringCategories([]string{"a", "b", "a"})
	distribution, err := graph.CategoricalJointDistribution(categories, igraph.CategoryJointDistributionOptions{Weights: []float64{2, 3, 5}, Directed: true})
	if err != nil {
		log.Fatal(err)
	}
	rows, _ := distribution.RowCategories.StringValues()
	columns, _ := distribution.ColumnCategories.StringValues()
	fmt.Println(rows, columns)
	fmt.Println(distribution.Matrix.Rows())

	pairs, err := graph.NeighborhoodSimilarityPairs([]igraph.Edge{{From: 0, To: 2}}, igraph.NeighborhoodSimilarityOptions{Direction: igraph.DirectionOut})
	if err != nil {
		log.Fatal(err)
	}
	scan, err := graph.LocalScan(igraph.LocalScanOptions{Direction: igraph.DirectionOut, Weights: []float64{2, 3, 5}})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(pairs)
	fmt.Println(scan)
	// Output:
	// [a b] [a b]
	// [[3 7] [0 0]]
	// [0.5]
	// [5 0 5]
}

func ExampleGraph_CrossGraphLocalScan() {
	neighborhoods, err := igraph.NewGraphFromEdges(3, []igraph.Edge{{From: 0, To: 1}, {From: 1, To: 2}}, false)
	if err != nil {
		log.Fatal(err)
	}
	defer neighborhoods.Close()
	comparison, err := igraph.NewGraphFromEdges(3, []igraph.Edge{{From: 0, To: 1}, {From: 0, To: 2}}, false)
	if err != nil {
		log.Fatal(err)
	}
	defer comparison.Close()
	values, err := neighborhoods.CrossGraphLocalScan(comparison, igraph.LocalScanOptions{Radius: 1, Direction: igraph.DirectionAll})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(values)
	// Output: [1 2 0]
}
