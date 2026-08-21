package igraph_test

import (
	"fmt"
	"log"

	igraph "github.com/h8gi/go-igraph"
)

func ExampleGraph_VertexNumericAttributes() {
	graph, err := igraph.NewGraphFromEdges(3, []igraph.Edge{{From: 0, To: 1}, {From: 1, To: 2}}, true)
	if err != nil {
		log.Fatal(err)
	}
	defer graph.Close()

	if err := graph.SetGraphStringAttribute("name", "routes"); err != nil {
		log.Fatal(err)
	}
	if err := graph.SetVertexNumericAttributes("score", []float64{4, 7, 9}); err != nil {
		log.Fatal(err)
	}
	if err := graph.SetEdgeBooleanAttributes("open", []bool{true, false}); err != nil {
		log.Fatal(err)
	}
	if err := graph.SetVertexNumericAttribute("score", 1, 8); err != nil {
		log.Fatal(err)
	}

	name, _ := graph.GraphStringAttribute("name")
	scores, _ := graph.VertexNumericAttributes("score")
	open, _ := graph.EdgeBooleanAttributes("open")
	fmt.Println(name)
	fmt.Println(scores)
	fmt.Println(open)
	// Output:
	// routes
	// [4 8 9]
	// [true false]
}
