package igraph_test

import (
	"fmt"
	"log"

	igraph "github.com/h8gi/go-igraph"
)

func ExampleStaticFitnessGame() {
	seed := uint64(42)
	graph, err := igraph.StaticFitnessGame(
		4,
		[]float64{1, 2, 3, 4},
		nil,
		igraph.StaticFitnessOptions{Seed: &seed},
	)
	if err != nil {
		log.Fatal(err)
	}
	defer graph.Close()

	vertices, err := graph.VertexCount()
	if err != nil {
		log.Fatal(err)
	}
	edges, err := graph.EdgeCount()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("vertices: %d, edges: %d\n", vertices, edges)

	// Output:
	// vertices: 4, edges: 4
}
