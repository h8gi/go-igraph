package main

import (
	"fmt"
	"log"

	igraph "github.com/h8gi/go-igraph"
)

func main() {
	seed := uint64(2026)

	fmt.Println("=== Reproducible and Advanced Random Graph Models ===")

	// 1. Generate a Barabási-Albert scale-free network reproducibly.
	baGraph, err := igraph.BarabasiGame(20, 2, 1.0, 1.0, false, igraph.BarabasiOptions{Seed: &seed})
	if err != nil {
		log.Fatalf("BarabasiGame failed: %v", err)
	}
	defer baGraph.Close()

	vCount, eCount := mustCounts(baGraph)
	fmt.Printf("1. Generated Barabási-Albert Graph: %d vertices, %d edges\n", vCount, eCount)

	// 2. Perform in-place rewiring while preserving degree sequence.
	if err := baGraph.Rewire(50, igraph.RewireSimple, igraph.RewireOptions{Seed: &seed}); err != nil {
		log.Fatalf("Rewire failed: %v", err)
	}
	fmt.Println("2. Successfully rewired graph in-place while preserving degree sequence.")

	// 3. Sample a random walk on the rewired network.
	vPath, ePath, err := baGraph.RandomWalk(0, 10, igraph.DirectionOut, nil, igraph.RandomWalkOptions{Seed: &seed})
	if err != nil {
		log.Fatalf("RandomWalk failed: %v", err)
	}
	fmt.Printf("3. Random Walk (start=0, steps=10):\n   Visited Vertices: %v\n   Traversed Edges: %v\n", vPath, ePath)

	// 4. Generate a Stochastic Block Model (SBM) graph with 2 communities.
	prefMat, err := igraph.NewMatrixFromRows([][]float64{
		{0.6, 0.05},
		{0.05, 0.6},
	})
	if err != nil {
		log.Fatalf("NewMatrixFromRows failed: %v", err)
	}

	sbmGraph, err := igraph.SBMGame(10, prefMat, []int{5, 5}, false, false, igraph.SBMOptions{Seed: &seed})
	if err != nil {
		log.Fatalf("SBMGame failed: %v", err)
	}
	defer sbmGraph.Close()

	sbmVCount, sbmECount := mustCounts(sbmGraph)
	fmt.Printf("4. Generated Stochastic Block Model Graph: %d vertices, %d edges\n", sbmVCount, sbmECount)

	// 5. Sample a uniform random spanning tree.
	treeEdges, err := sbmGraph.RandomSpanningTree(igraph.SpanningTreeOptions{Seed: &seed})
	if err != nil {
		log.Fatalf("RandomSpanningTree failed: %v", err)
	}
	fmt.Printf("5. Sampled Random Spanning Tree Edge IDs: %v\n", treeEdges)

	// 6. Sample row-oriented latent positions and generate a random-dot-product graph.
	positions, err := igraph.SampleDirichlet(8, []float64{1, 1, 1}, igraph.LatentSampleOptions{Seed: &seed})
	if err != nil {
		log.Fatalf("SampleDirichlet failed: %v", err)
	}
	latentGraph, err := igraph.DotProductGame(positions, igraph.LatentGraphOptions{Seed: &seed})
	if err != nil {
		log.Fatalf("DotProductGame failed: %v", err)
	}
	defer latentGraph.Close()
	rows, dimensions := positions.Dims()
	fmt.Printf("6. Generated latent graph from a %d-by-%d row-per-vertex matrix.\n", rows, dimensions)

	// 7. Reuse Go-owned geometric coordinates in spatial analysis.
	geometric, err := igraph.GeometricRandomGame(12, 0.3, igraph.GeometricGraphOptions{Seed: &seed})
	if err != nil {
		log.Fatalf("GeometricRandomGame failed: %v", err)
	}
	defer geometric.Graph.Close()
	lengths, err := geometric.Graph.SpatialEdgeLengths(geometric.Coordinates, igraph.SpatialEuclidean)
	if err != nil {
		log.Fatalf("SpatialEdgeLengths failed: %v", err)
	}
	fmt.Printf("7. Reused geometric coordinates to measure %d spatial edges.\n", len(lengths))
}

func mustCounts(g *igraph.Graph) (int, int) {
	v, err := g.VertexCount()
	if err != nil {
		log.Fatalf("VertexCount failed: %v", err)
	}
	e, err := g.EdgeCount()
	if err != nil {
		log.Fatalf("EdgeCount failed: %v", err)
	}
	return v, e
}
