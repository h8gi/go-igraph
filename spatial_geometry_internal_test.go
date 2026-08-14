package igraph

import (
	"errors"
	"testing"
)

func TestConvexHullFailureAdapters(t *testing.T) {
	points, _ := NewMatrixFromRows([][]float64{{0, 0}, {1, 0}, {0, 1}})
	failure := errors.New("injected failure")

	for name, modify := range map[string]func(*convexHullAdapters){
		"point matrix initialization": func(adapters *convexHullAdapters) {
			adapters.newMatrix = func(Matrix) (*cMatrix, error) { return nil, failure }
		},
		"index initialization": func(adapters *convexHullAdapters) {
			adapters.newInt = func([]int) (*intVector, error) { return nil, failure }
		},
		"coordinate initialization": func(adapters *convexHullAdapters) {
			calls := 0
			adapters.newMatrix = func(matrix Matrix) (*cMatrix, error) {
				calls++
				if calls == 2 {
					return nil, failure
				}
				return newCMatrix(matrix)
			}
		},
		"upstream": func(adapters *convexHullAdapters) {
			adapters.call = func(*cMatrix, *intVector, *cMatrix) int { return 4 }
		},
		"index conversion": func(adapters *convexHullAdapters) {
			adapters.convertInt = func(*intVector) ([]int, error) { return nil, failure }
		},
		"coordinate conversion": func(adapters *convexHullAdapters) {
			adapters.convertMatrix = func(*cMatrix) (Matrix, error) { return Matrix{}, failure }
		},
		"alignment": func(adapters *convexHullAdapters) {
			adapters.convertInt = func(*intVector) ([]int, error) { return []int{0}, nil }
			adapters.convertMatrix = func(*cMatrix) (Matrix, error) { return Matrix{}, nil }
		},
	} {
		t.Run(name, func(t *testing.T) {
			adapters := defaultConvexHullAdapters()
			modify(&adapters)
			if _, err := convexHull2D(points, &adapters); err == nil {
				t.Fatal("expected injected failure")
			}
		})
	}
}

func TestSpatialEdgeLengthFailureAdapters(t *testing.T) {
	graph, _ := NewGraphFromEdges(2, []Edge{{From: 0, To: 1}}, false)
	t.Cleanup(func() { _ = graph.Close() })
	points, _ := NewMatrixFromRows([][]float64{{0, 0}, {1, 0}})
	failure := errors.New("injected failure")

	for name, modify := range map[string]func(*spatialEdgeLengthAdapters){
		"matrix initialization": func(adapters *spatialEdgeLengthAdapters) {
			adapters.newMatrix = func(Matrix) (*cMatrix, error) { return nil, failure }
		},
		"vector initialization": func(adapters *spatialEdgeLengthAdapters) {
			adapters.newReal = func([]float64) (*realVector, error) { return nil, failure }
		},
		"upstream": func(adapters *spatialEdgeLengthAdapters) {
			adapters.call = func(*Graph, *realVector, *cMatrix, SpatialMetric) int { return 4 }
		},
		"conversion": func(adapters *spatialEdgeLengthAdapters) {
			adapters.convert = func(*realVector) ([]float64, error) { return nil, failure }
		},
		"alignment": func(adapters *spatialEdgeLengthAdapters) {
			adapters.convert = func(*realVector) ([]float64, error) { return []float64{}, nil }
		},
	} {
		t.Run(name, func(t *testing.T) {
			adapters := defaultSpatialEdgeLengthAdapters()
			modify(&adapters)
			if _, err := graph.spatialEdgeLengths(points, SpatialEuclidean, &adapters); err == nil {
				t.Fatal("expected injected failure")
			}
		})
	}
}
