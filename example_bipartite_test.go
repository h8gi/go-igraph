package igraph_test

import (
	"fmt"
	"log"

	igraph "github.com/h8gi/go-igraph"
)

func ExampleNewWeightedBiadjacency() {
	matrix, err := igraph.NewMatrixFromRows([][]float64{
		{1.5, 0, 2},
		{0, 3, 1},
	})
	if err != nil {
		log.Fatal(err)
	}
	result, err := igraph.NewWeightedBiadjacency(matrix, igraph.WeightedBiadjacencyOptions{})
	if err != nil {
		log.Fatal(err)
	}
	defer result.Graph.Close()

	roundTrip, err := result.Graph.Biadjacency(result.Partition, result.Weights)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(roundTrip.RowVertexIDs)
	fmt.Println(roundTrip.ColumnVertexIDs)
	fmt.Println(roundTrip.Matrix.Rows())
	// Output:
	// [0 1]
	// [2 3 4]
	// [[1.5 0 2] [0 3 1]]
}
