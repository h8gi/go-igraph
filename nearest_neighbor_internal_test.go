package igraph

import (
	"errors"
	"testing"
)

func TestNearestNeighborFailureAdapters(t *testing.T) {
	points, _ := NewMatrixFromRows([][]float64{{0}, {1}})
	failure := errors.New("injected failure")

	initialization := defaultNearestNeighborGraphAdapters()
	initialization.newMatrix = func(Matrix) (*cMatrix, error) { return nil, failure }
	if graph, err := newNearestNeighborGraph(points, NearestNeighborOptions{}, &initialization); graph != nil || !errors.Is(err, failure) {
		t.Fatalf("matrix initialization = %v, %v", graph, err)
	}

	upstream := defaultNearestNeighborGraphAdapters()
	upstream.call = func(*cMatrix, validatedNearestNeighborOptions) nearestNeighborGraphCallResult {
		return nearestNeighborGraphCallResult{code: 4}
	}
	if graph, err := newNearestNeighborGraph(points, NearestNeighborOptions{}, &upstream); graph != nil || err == nil {
		t.Fatalf("upstream failure = %v, %v", graph, err)
	}
}
