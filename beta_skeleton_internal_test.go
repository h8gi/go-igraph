package igraph

import (
	"errors"
	"testing"
)

func TestBetaSkeletonFailureAdapters(t *testing.T) {
	points, _ := NewMatrixFromRows([][]float64{{0, 0}, {1, 0}})
	requirements := spatialPointRequirements{operation: "test skeleton", exactDimensions: 2, distinct: true}
	failure := errors.New("injected failure")

	initialization := defaultBetaSkeletonAdapters()
	initialization.newMatrix = func(Matrix) (*cMatrix, error) { return nil, failure }
	if graph, err := newBetaSkeleton(points, 1, requirements, betaSkeletonLune, &initialization); graph != nil || !errors.Is(err, failure) {
		t.Fatalf("matrix initialization = %v, %v", graph, err)
	}

	upstream := defaultBetaSkeletonAdapters()
	upstream.call = func(betaSkeletonKind, *cMatrix, float64) betaSkeletonCallResult {
		return betaSkeletonCallResult{code: 4}
	}
	if graph, err := newBetaSkeleton(points, 1, requirements, betaSkeletonLune, &upstream); graph != nil || err == nil {
		t.Fatalf("upstream failure = %v, %v", graph, err)
	}
}

func TestBetaWeightedGabrielFailureAdapters(t *testing.T) {
	points, _ := NewMatrixFromRows([][]float64{{0}, {1}})
	failure := errors.New("injected failure")

	for name, modify := range map[string]func(*betaWeightedGabrielAdapters){
		"matrix initialization": func(adapters *betaWeightedGabrielAdapters) {
			adapters.newMatrix = func(Matrix) (*cMatrix, error) { return nil, failure }
		},
		"vector initialization": func(adapters *betaWeightedGabrielAdapters) {
			adapters.newReal = func([]float64) (*realVector, error) { return nil, failure }
		},
		"upstream": func(adapters *betaWeightedGabrielAdapters) {
			adapters.call = func(*cMatrix, *realVector, float64) betaWeightedGabrielCallResult {
				return betaWeightedGabrielCallResult{code: 4}
			}
		},
		"conversion": func(adapters *betaWeightedGabrielAdapters) {
			adapters.convert = func(*realVector) ([]float64, error) { return nil, failure }
		},
		"alignment": func(adapters *betaWeightedGabrielAdapters) {
			adapters.convert = func(*realVector) ([]float64, error) { return []float64{}, nil }
		},
	} {
		t.Run(name, func(t *testing.T) {
			adapters := defaultBetaWeightedGabrielAdapters()
			modify(&adapters)
			if result, err := newBetaWeightedGabrielGraph(points, BetaWeightedGabrielOptions{}, &adapters); err == nil || result.Graph != nil {
				t.Fatalf("result = %#v, error = %v", result, err)
			}
		})
	}
}
