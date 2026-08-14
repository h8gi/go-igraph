package igraph_test

import (
	"errors"
	"sync"
	"testing"

	igraph "github.com/h8gi/go-igraph"
)

func TestMilestone16SpatialRoutingPipeline(t *testing.T) {
	points, _ := igraph.NewMatrixFromRows([][]float64{
		{0, 0}, {1, 0}, {3, 0}, {6, 0}, {3, 2},
	})
	maximum := 2
	graph, err := igraph.NewNearestNeighborGraph(points, igraph.NearestNeighborOptions{
		MaxNeighbors: &maximum,
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = graph.Close() })

	lengths, err := graph.SpatialEdgeLengths(points, igraph.SpatialEuclidean)
	if err != nil {
		t.Fatal(err)
	}
	if edges, _ := graph.EdgeCount(); len(lengths) != edges {
		t.Fatalf("edge lengths = %d, want %d", len(lengths), edges)
	}
	path, err := graph.ShortestPath(0, 3, igraph.PathOptions{
		Direction: igraph.DirectionAll,
		Weights:   lengths,
	})
	if err != nil || !path.Found || path.Vertices[0] != 0 || path.Vertices[len(path.Vertices)-1] != 3 {
		t.Fatalf("spatial weighted path = %#v, %v", path, err)
	}
	for _, edgeID := range path.Edges {
		if edgeID < 0 || edgeID >= len(lengths) {
			t.Fatalf("path edge ID %d is not aligned with %d lengths", edgeID, len(lengths))
		}
	}

	if err := graph.Close(); err != nil {
		t.Fatal(err)
	}
	if len(lengths) == 0 || len(path.Vertices) == 0 || len(path.Edges) == 0 {
		t.Fatal("Go-owned routing results were cleared by graph closure")
	}
	if _, err := graph.SpatialEdgeLengths(points, igraph.SpatialEuclidean); !errors.Is(err, igraph.ErrClosed) {
		t.Fatalf("post-close SpatialEdgeLengths error = %v, want ErrClosed", err)
	}
}

func TestMilestone16PlanarGeometryPipeline(t *testing.T) {
	points, _ := igraph.NewMatrixFromRows([][]float64{
		{0, 0}, {3, 0}, {4, 2}, {2, 4}, {0, 3}, {1.4, 1.1},
	})
	hull, err := igraph.ConvexHull2D(points)
	if err != nil || len(hull.PointIndices) != 5 || len(hull.Coordinates.Rows()) != 5 {
		t.Fatalf("convex hull = %#v, %v", hull, err)
	}

	delaunay := newMilestone16Graph(t, func() (*igraph.Graph, error) { return igraph.NewDelaunayGraph(points) })
	gabriel := newMilestone16Graph(t, func() (*igraph.Graph, error) { return igraph.NewGabrielGraph(points) })
	relative := newMilestone16Graph(t, func() (*igraph.Graph, error) { return igraph.NewRelativeNeighborhoodGraph(points) })
	lune := newMilestone16Graph(t, func() (*igraph.Graph, error) { return igraph.NewLuneBetaSkeleton(points, 1) })
	circle := newMilestone16Graph(t, func() (*igraph.Graph, error) { return igraph.NewCircleBetaSkeleton(points, 1) })
	weighted, err := igraph.NewBetaWeightedGabrielGraph(points, igraph.BetaWeightedGabrielOptions{})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = weighted.Graph.Close() })

	assertEdgeSubset(t, relative, gabriel)
	assertEdgeSubset(t, gabriel, delaunay)
	assertSameEdges(t, lune, gabriel)
	assertSameEdges(t, circle, gabriel)
	assertSameEdges(t, weighted.Graph, gabriel)
	if edges, _ := weighted.Graph.EdgeCount(); len(weighted.ThresholdBetas) != edges {
		t.Fatalf("threshold count = %d, want %d", len(weighted.ThresholdBetas), edges)
	}

	if err := weighted.Graph.Close(); err != nil {
		t.Fatal(err)
	}
	if len(weighted.ThresholdBetas) == 0 {
		t.Fatal("Go-owned thresholds were cleared by graph closure")
	}
}

func TestMilestone16ConcurrentSpatialReadsAndClose(t *testing.T) {
	points, _ := igraph.NewMatrixFromRows([][]float64{
		{0, 0}, {3, 0}, {4, 2}, {2, 4}, {0, 3}, {1.4, 1.1},
	})
	graph := newMilestone16Graph(t, func() (*igraph.Graph, error) { return igraph.NewDelaunayGraph(points) })
	start := make(chan struct{})
	errorsByCall := make(chan error, 16)
	var wait sync.WaitGroup
	for index := 0; index < 16; index++ {
		wait.Add(1)
		go func() {
			defer wait.Done()
			<-start
			_, err := graph.SpatialEdgeLengths(points, igraph.SpatialEuclidean)
			errorsByCall <- err
		}()
	}
	close(start)
	if err := graph.Close(); err != nil {
		t.Fatal(err)
	}
	wait.Wait()
	close(errorsByCall)
	for err := range errorsByCall {
		if err != nil && !errors.Is(err, igraph.ErrClosed) {
			t.Errorf("Close-race SpatialEdgeLengths error = %v", err)
		}
	}

	var constructors sync.WaitGroup
	for index := 0; index < 12; index++ {
		constructors.Add(1)
		go func() {
			defer constructors.Done()
			created, err := igraph.NewGabrielGraph(points)
			if err != nil {
				t.Errorf("concurrent spatial constructor = %v", err)
				return
			}
			_ = created.Close()
		}()
	}
	constructors.Wait()
}

func newMilestone16Graph(t *testing.T, construct func() (*igraph.Graph, error)) *igraph.Graph {
	t.Helper()
	graph, err := construct()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = graph.Close() })
	return graph
}
