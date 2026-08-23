package igraph_test

import (
	"fmt"

	igraph "github.com/h8gi/go-igraph"
)

func ExampleGraph_Chordality() {
	graph, _ := igraph.NewGraphFromEdges(4, []igraph.Edge{{0, 1}, {1, 2}, {2, 3}, {3, 0}}, false)
	defer graph.Close()

	result, _ := graph.Chordality(igraph.ChordalityOptions{Complete: true})
	defer result.Completion.Close()
	completed, _ := result.Completion.Chordality(igraph.ChordalityOptions{})
	fmt.Println(result.Chordal, len(result.FillEdges), completed.Chordal)
	// Output: false 1 true
}

func ExampleGraph_Laplacian() {
	graph, _ := igraph.NewGraphFromEdges(3, []igraph.Edge{{0, 1}, {1, 2}}, false)
	defer graph.Close()

	matrix, _ := graph.Laplacian(igraph.LaplacianOptions{})
	fmt.Println(matrix.Rows())
	// Output: [[1 -1 0] [-1 2 -1] [0 -1 1]]
}
