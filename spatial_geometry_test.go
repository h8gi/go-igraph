package igraph_test

import (
	"errors"
	"math"
	"testing"

	igraph "github.com/h8gi/go-igraph"
)

func TestConvexHull2DAlignmentAndOwnership(t *testing.T) {
	points, err := igraph.NewMatrixFromRows([][]float64{
		{0, 0}, {2, 0}, {2, 2}, {0, 2}, {1, 1},
	})
	if err != nil {
		t.Fatal(err)
	}
	result, err := igraph.ConvexHull2D(points)
	if err != nil {
		t.Fatal(err)
	}
	rows, columns := result.Coordinates.Dims()
	if result.PointIndices == nil || rows != len(result.PointIndices) || columns != 2 {
		t.Fatalf("hull shape = indices %v, matrix %dx%d", result.PointIndices, rows, columns)
	}
	if len(result.PointIndices) != 4 {
		t.Fatalf("hull indices = %v, want four corners", result.PointIndices)
	}
	wantCorners := map[int]bool{0: true, 1: true, 2: true, 3: true}
	coordinates := result.Coordinates.Rows()
	for index, pointID := range result.PointIndices {
		if !wantCorners[pointID] {
			t.Fatalf("hull contains non-corner point %d", pointID)
		}
		wantX, _ := points.At(pointID, 0)
		wantY, _ := points.At(pointID, 1)
		if coordinates[index][0] != wantX || coordinates[index][1] != wantY {
			t.Fatalf("hull row %d = %v, want point row %d", index, coordinates[index], pointID)
		}
	}
	result.PointIndices[0] = 99
	coordinates[0][0] = 99
	if value, _ := points.At(0, 0); value != 0 {
		t.Fatalf("mutating hull result changed input point: %v", value)
	}
}

func TestConvexHull2DEmptyAndDegenerate(t *testing.T) {
	empty, err := igraph.ConvexHull2D(igraph.Matrix{})
	if err != nil {
		t.Fatal(err)
	}
	rows, columns := empty.Coordinates.Dims()
	if empty.PointIndices == nil || len(empty.PointIndices) != 0 || rows != 0 || columns != 2 {
		t.Fatalf("empty hull = %#v, dimensions %dx%d", empty, rows, columns)
	}

	line, _ := igraph.NewMatrixFromRows([][]float64{{0, 0}, {1, 0}, {2, 0}, {1, 0}})
	result, err := igraph.ConvexHull2D(line)
	if err != nil {
		t.Fatal(err)
	}
	rows, columns = result.Coordinates.Dims()
	if result.PointIndices == nil || rows != len(result.PointIndices) || columns != 2 {
		t.Fatalf("degenerate hull = %#v, dimensions %dx%d", result, rows, columns)
	}
}

func TestSpatialEdgeLengthsMetricsAndEdgeOrder(t *testing.T) {
	graph, err := igraph.NewGraphFromEdges(3, []igraph.Edge{
		{From: 0, To: 1},
		{From: 0, To: 1},
		{From: 1, To: 1},
		{From: 1, To: 2},
	}, true)
	if err != nil {
		t.Fatal(err)
	}
	points, _ := igraph.NewMatrixFromRows([][]float64{{0, 0}, {3, 4}, {6, 0}})

	euclidean, err := graph.SpatialEdgeLengths(points, igraph.SpatialEuclidean)
	if err != nil {
		t.Fatal(err)
	}
	manhattan, err := graph.SpatialEdgeLengths(points, igraph.SpatialManhattan)
	if err != nil {
		t.Fatal(err)
	}
	for index, want := range []float64{5, 5, 0, 5} {
		if euclidean[index] != want {
			t.Errorf("Euclidean length %d = %v, want %v", index, euclidean[index], want)
		}
	}
	for index, want := range []float64{7, 7, 0, 7} {
		if manhattan[index] != want {
			t.Errorf("Manhattan length %d = %v, want %v", index, manhattan[index], want)
		}
	}
	if err := graph.Close(); err != nil {
		t.Fatal(err)
	}
	if euclidean[0] != 5 || manhattan[0] != 7 {
		t.Fatal("Go-owned lengths changed after graph closure")
	}
	if _, err := graph.SpatialEdgeLengths(points, igraph.SpatialEuclidean); !errors.Is(err, igraph.ErrClosed) {
		t.Fatalf("closed graph error = %v, want %v", err, igraph.ErrClosed)
	}
}

func TestSpatialEdgeLengthsEmptyGraph(t *testing.T) {
	graph, err := igraph.NewGraph()
	if err != nil {
		t.Fatal(err)
	}
	defer graph.Close()
	lengths, err := graph.SpatialEdgeLengths(igraph.Matrix{}, igraph.SpatialEuclidean)
	if err != nil {
		t.Fatal(err)
	}
	if lengths == nil || len(lengths) != 0 {
		t.Fatalf("empty graph lengths = %#v, want non-nil empty slice", lengths)
	}
}

func TestSpatialGeometryRejectsInvalidInputs(t *testing.T) {
	oneDimensional, _ := igraph.NewMatrixFromRows([][]float64{{0}, {1}})
	if _, err := igraph.ConvexHull2D(oneDimensional); err == nil {
		t.Fatal("ConvexHull2D accepted one-dimensional points")
	}
	nonFinite, _ := igraph.NewMatrixFromRows([][]float64{{0, 0}, {math.NaN(), 1}})
	if _, err := igraph.ConvexHull2D(nonFinite); err == nil {
		t.Fatal("ConvexHull2D accepted non-finite points")
	}

	graph, _ := igraph.NewGraphFromEdges(2, []igraph.Edge{{From: 0, To: 1}}, false)
	t.Cleanup(func() { _ = graph.Close() })
	wrongCount, _ := igraph.NewMatrixFromRows([][]float64{{0, 0}})
	zeroDimensions, _ := igraph.NewMatrix(2, 0)
	valid, _ := igraph.NewMatrixFromRows([][]float64{{0, 0}, {1, 1}})
	for name, call := range map[string]func() error{
		"point count": func() error {
			_, err := graph.SpatialEdgeLengths(wrongCount, igraph.SpatialEuclidean)
			return err
		},
		"zero dimensions": func() error {
			_, err := graph.SpatialEdgeLengths(zeroDimensions, igraph.SpatialEuclidean)
			return err
		},
		"metric": func() error {
			_, err := graph.SpatialEdgeLengths(valid, igraph.SpatialMetric(99))
			return err
		},
	} {
		t.Run(name, func(t *testing.T) {
			if err := call(); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
	if _, err := (*igraph.Graph)(nil).SpatialEdgeLengths(valid, igraph.SpatialEuclidean); !errors.Is(err, igraph.ErrClosed) {
		t.Fatalf("nil graph error = %v, want %v", err, igraph.ErrClosed)
	}
}
