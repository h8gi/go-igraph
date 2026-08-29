package igraph_test

import (
	"fmt"

	igraph "github.com/h8gi/go-igraph"
)

func Example_advancedRandomModels() {
	seed := uint64(2026)
	positions, err := igraph.SampleDirichlet(5, []float64{1, 1, 1}, igraph.LatentSampleOptions{Seed: &seed})
	if err != nil {
		panic(err)
	}
	graph, err := igraph.DotProductGame(positions, igraph.LatentGraphOptions{Seed: &seed})
	if err != nil {
		panic(err)
	}
	defer graph.Close()
	rows, dimensions := positions.Dims()
	vertices, _ := graph.VertexCount()
	fmt.Printf("latent samples: %d x %d; graph vertices: %d\n", rows, dimensions, vertices)

	geometric, err := igraph.GeometricRandomGame(4, 0.5, igraph.GeometricGraphOptions{Seed: &seed})
	if err != nil {
		panic(err)
	}
	defer geometric.Graph.Close()
	rows, dimensions = geometric.Coordinates.Dims()
	fmt.Printf("geometric coordinates: %d x %d\n", rows, dimensions)
	// Output:
	// latent samples: 5 x 3; graph vertices: 5
	// geometric coordinates: 4 x 2
}
