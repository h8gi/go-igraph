package igraph

import (
	"errors"
	"testing"
)

func TestProximityGraphFailureAdapters(t *testing.T) {
	points, _ := NewMatrixFromRows([][]float64{{0}, {1}})
	operation := proximityGraphOperation{name: "test proximity graph"}
	failure := errors.New("injected failure")

	initialization := defaultProximityGraphAdapters()
	initialization.newMatrix = func(Matrix) (*cMatrix, error) { return nil, failure }
	if graph, err := newProximityGraph(points, operation, &initialization); graph != nil || !errors.Is(err, failure) {
		t.Fatalf("matrix initialization = %v, %v", graph, err)
	}

	upstream := defaultProximityGraphAdapters()
	upstream.call = func(proximityGraphOperation, *cMatrix) proximityGraphCallResult {
		return proximityGraphCallResult{code: 4}
	}
	if graph, err := newProximityGraph(points, operation, &upstream); graph != nil || err == nil {
		t.Fatalf("upstream failure = %v, %v", graph, err)
	}
}
