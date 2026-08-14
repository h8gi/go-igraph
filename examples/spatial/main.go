package main

import (
	"fmt"
	"log"

	igraph "github.com/h8gi/go-igraph"
)

func main() {
	points, err := igraph.NewMatrixFromRows([][]float64{
		{0, 0}, {1, 0}, {3, 0}, {6, 0}, {3, 2},
	})
	if err != nil {
		log.Fatal(err)
	}
	maximum := 2
	graph, err := igraph.NewNearestNeighborGraph(points, igraph.NearestNeighborOptions{MaxNeighbors: &maximum})
	if err != nil {
		log.Fatal(err)
	}
	defer graph.Close()
	lengths, err := graph.SpatialEdgeLengths(points, igraph.SpatialEuclidean)
	if err != nil {
		log.Fatal(err)
	}
	path, err := graph.ShortestPath(0, 3, igraph.PathOptions{Direction: igraph.DirectionAll, Weights: lengths})
	if err != nil {
		log.Fatal(err)
	}
	distance := 0.0
	for _, edgeID := range path.Edges {
		distance += lengths[edgeID]
	}
	fmt.Printf("weighted spatial route: %v (length %.1f)\n", path.Vertices, distance)

	hull, err := igraph.ConvexHull2D(points)
	if err != nil {
		log.Fatal(err)
	}
	weighted, err := igraph.NewBetaWeightedGabrielGraph(points, igraph.BetaWeightedGabrielOptions{})
	if err != nil {
		log.Fatal(err)
	}
	defer weighted.Graph.Close()
	fmt.Printf("hull point rows: %v\n", hull.PointIndices)
	fmt.Printf("Gabriel thresholds by edge ID: %v\n", weighted.ThresholdBetas)
}
