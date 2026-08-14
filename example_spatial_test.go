package igraph_test

import (
	"fmt"
	"log"

	igraph "github.com/h8gi/go-igraph"
)

func ExampleGraph_SpatialEdgeLengths() {
	points, _ := igraph.NewMatrixFromRows([][]float64{{0, 0}, {1, 0}, {3, 0}, {6, 0}})
	maximum := 1
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
	fmt.Println(path.Vertices)
	fmt.Println(distance)
	// Output:
	// [0 1 2 3]
	// 6
}

func ExampleConvexHull2D() {
	points, _ := igraph.NewMatrixFromRows([][]float64{{0, 0}, {2, 0}, {0, 2}})
	hull, err := igraph.ConvexHull2D(points)
	if err != nil {
		log.Fatal(err)
	}
	delaunay, _ := igraph.NewDelaunayGraph(points)
	defer delaunay.Close()
	gabriel, _ := igraph.NewGabrielGraph(points)
	defer gabriel.Close()
	relative, _ := igraph.NewRelativeNeighborhoodGraph(points)
	defer relative.Close()
	delaunayEdges, _ := delaunay.EdgeCount()
	gabrielEdges, _ := gabriel.EdgeCount()
	relativeEdges, _ := relative.EdgeCount()
	fmt.Println(len(hull.PointIndices), delaunayEdges, gabrielEdges, relativeEdges)
	// Output:
	// 3 3 2 2
}
