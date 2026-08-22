package igraph

import (
	"errors"
	"testing"
)

func TestAdjacencyConstructionFailureAdapters(t *testing.T) {
	matrix, _ := NewMatrixFromRows([][]float64{{0, 1}, {0, 0}})
	failure := errors.New("injected failure")

	initialization := defaultAdjacencyAdapters()
	initialization.newMatrix = func(Matrix) (*cMatrix, error) { return nil, failure }
	if _, err := newAdjacency(matrix, AdjacencyOptions{}, &initialization); !errors.Is(err, failure) {
		t.Fatalf("matrix init = %v", err)
	}

	weightInitialization := defaultAdjacencyAdapters()
	weightInitialization.newReal = func([]float64) (*realVector, error) { return nil, failure }
	if _, err := newWeightedAdjacency(matrix, AdjacencyOptions{}, &weightInitialization); !errors.Is(err, failure) {
		t.Fatalf("weight init = %v", err)
	}

	upstream := defaultAdjacencyAdapters()
	upstream.create = func(*cMatrix, AdjacencyOptions) adjacencyCallResult {
		return adjacencyCallResult{code: 1}
	}
	if _, err := newAdjacency(matrix, AdjacencyOptions{}, &upstream); err == nil {
		t.Fatal("upstream error nil")
	}

	weightedUpstream := defaultAdjacencyAdapters()
	weightedUpstream.weighted = func(*cMatrix, *realVector, AdjacencyOptions) adjacencyCallResult {
		return adjacencyCallResult{code: 1}
	}
	if _, err := newWeightedAdjacency(matrix, AdjacencyOptions{}, &weightedUpstream); err == nil {
		t.Fatal("weighted upstream error nil")
	}

	conversion := defaultAdjacencyAdapters()
	conversion.convertReal = func(*realVector) ([]float64, error) { return nil, failure }
	if _, err := newWeightedAdjacency(matrix, AdjacencyOptions{}, &conversion); !errors.Is(err, failure) {
		t.Fatalf("weight conversion = %v", err)
	}
}
