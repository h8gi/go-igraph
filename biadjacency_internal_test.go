package igraph

import (
	"errors"
	"testing"
)

func TestBiadjacencyConstructionFailureAdapters(t *testing.T) {
	matrix, _ := NewMatrixFromRows([][]float64{{1}})
	failure := errors.New("injected failure")

	initialization := defaultBiadjacencyAdapters()
	initialization.newMatrix = func(Matrix) (*cMatrix, error) { return nil, failure }
	if _, err := newBiadjacency(matrix, BiadjacencyOptions{}, &initialization); !errors.Is(err, failure) {
		t.Fatalf("matrix init = %v", err)
	}

	partitionInit := defaultBiadjacencyAdapters()
	partitionInit.newBool = func([]bool) (*boolVector, error) { return nil, failure }
	if _, err := newBiadjacency(matrix, BiadjacencyOptions{}, &partitionInit); !errors.Is(err, failure) {
		t.Fatalf("partition init = %v", err)
	}

	weightInit := defaultBiadjacencyAdapters()
	weightInit.newReal = func([]float64) (*realVector, error) { return nil, failure }
	if _, err := newWeightedBiadjacency(matrix, WeightedBiadjacencyOptions{}, &weightInit); !errors.Is(err, failure) {
		t.Fatalf("weight init = %v", err)
	}

	upstream := defaultBiadjacencyAdapters()
	upstream.create = func(*cMatrix, *boolVector, BiadjacencyOptions) biadjacencyGraphCallResult {
		return biadjacencyGraphCallResult{code: 1}
	}
	if _, err := newBiadjacency(matrix, BiadjacencyOptions{}, &upstream); err == nil {
		t.Fatal("upstream error nil")
	}

	weightedUpstream := defaultBiadjacencyAdapters()
	weightedUpstream.createWeighted = func(*cMatrix, *boolVector, *realVector, WeightedBiadjacencyOptions) weightedBiadjacencyCallResult {
		return weightedBiadjacencyCallResult{code: 1}
	}
	if _, err := newWeightedBiadjacency(matrix, WeightedBiadjacencyOptions{}, &weightedUpstream); err == nil {
		t.Fatal("weighted upstream error nil")
	}

	partitionConversion := defaultBiadjacencyAdapters()
	partitionConversion.convertBool = func(*boolVector) ([]bool, error) { return nil, failure }
	if _, err := newBiadjacency(matrix, BiadjacencyOptions{}, &partitionConversion); !errors.Is(err, failure) {
		t.Fatalf("partition conversion = %v", err)
	}

	weightConversion := defaultBiadjacencyAdapters()
	weightConversion.convertReal = func(*realVector) ([]float64, error) { return nil, failure }
	if _, err := newWeightedBiadjacency(matrix, WeightedBiadjacencyOptions{}, &weightConversion); !errors.Is(err, failure) {
		t.Fatalf("weight conversion = %v", err)
	}
}

func TestBiadjacencyExtractionFailureAdapters(t *testing.T) {
	graph, _ := NewGraphFromEdges(2, []Edge{{0, 1}}, false)
	t.Cleanup(func() { _ = graph.Close() })
	partition := BipartitePartition{false, true}
	failure := errors.New("injected failure")

	for name, modify := range map[string]func(*biadjacencyAdapters){
		"types":  func(a *biadjacencyAdapters) { a.newBool = func([]bool) (*boolVector, error) { return nil, failure } },
		"matrix": func(a *biadjacencyAdapters) { a.newMatrix = func(Matrix) (*cMatrix, error) { return nil, failure } },
		"rows":   func(a *biadjacencyAdapters) { a.newInt = func([]int) (*intVector, error) { return nil, failure } },
	} {
		t.Run(name, func(t *testing.T) {
			adapters := defaultBiadjacencyAdapters()
			modify(&adapters)
			if _, err := graph.biadjacency(partition, nil, &adapters); !errors.Is(err, failure) {
				t.Fatalf("error = %v", err)
			}
		})
	}

	upstream := defaultBiadjacencyAdapters()
	upstream.get = func(*Graph, *boolVector, *realVector, *cMatrix, *intVector, *intVector) int { return 1 }
	if _, err := graph.biadjacency(partition, nil, &upstream); err == nil {
		t.Fatal("upstream error nil")
	}

	for name, modify := range map[string]func(*biadjacencyAdapters){
		"matrix": func(a *biadjacencyAdapters) {
			a.convertMatrix = func(*cMatrix) (Matrix, error) { return Matrix{}, failure }
		},
		"rows": func(a *biadjacencyAdapters) { a.convertInt = func(*intVector) ([]int, error) { return nil, failure } },
	} {
		t.Run("convert-"+name, func(t *testing.T) {
			adapters := defaultBiadjacencyAdapters()
			modify(&adapters)
			if _, err := graph.biadjacency(partition, nil, &adapters); !errors.Is(err, failure) {
				t.Fatalf("error = %v", err)
			}
		})
	}
}
