package igraph

import (
	"errors"
	"testing"
)

func TestAdjacencyConversionFailureAdapters(t *testing.T) {
	graph, _ := NewGraphFromEdges(2, []Edge{{From: 0, To: 1}}, true)
	t.Cleanup(func() { _ = graph.Close() })
	failure := errors.New("injected failure")

	for name, call := range map[string]func(*adjacencyConversionAdapters) error{
		"weights": func(adapters *adjacencyConversionAdapters) error {
			adapters.newReal = func([]float64) (*realVector, error) { return nil, failure }
			_, err := graph.adjacencyMatrix([]float64{1}, AdjacencyMatrixOptions{}, adapters)
			return err
		},
		"matrix-adjacency": func(adapters *adjacencyConversionAdapters) error {
			adapters.newMatrix = func(Matrix) (*cMatrix, error) { return nil, failure }
			_, err := graph.adjacencyMatrix(nil, AdjacencyMatrixOptions{}, adapters)
			return err
		},
		"matrix-stochastic": func(adapters *adjacencyConversionAdapters) error {
			adapters.newMatrix = func(Matrix) (*cMatrix, error) { return nil, failure }
			_, err := graph.stochasticMatrix(nil, StochasticMatrixOptions{}, adapters)
			return err
		},
		"convert-adjacency": func(adapters *adjacencyConversionAdapters) error {
			adapters.convertMatrix = func(*cMatrix) (Matrix, error) { return Matrix{}, failure }
			_, err := graph.adjacencyMatrix(nil, AdjacencyMatrixOptions{}, adapters)
			return err
		},
		"convert-stochastic": func(adapters *adjacencyConversionAdapters) error {
			adapters.convertMatrix = func(*cMatrix) (Matrix, error) { return Matrix{}, failure }
			_, err := graph.stochasticMatrix(nil, StochasticMatrixOptions{}, adapters)
			return err
		},
	} {
		t.Run(name, func(t *testing.T) {
			adapters := defaultAdjacencyConversionAdapters()
			if err := call(&adapters); !errors.Is(err, failure) {
				t.Fatalf("error = %v", err)
			}
		})
	}

	adjacencyUpstream := defaultAdjacencyConversionAdapters()
	adjacencyUpstream.adjacency = func(*Graph, *cMatrix, *realVector, AdjacencyMatrixOptions) int { return 1 }
	if _, err := graph.adjacencyMatrix(nil, AdjacencyMatrixOptions{}, &adjacencyUpstream); err == nil {
		t.Fatal("adjacency upstream error nil")
	}

	stochasticUpstream := defaultAdjacencyConversionAdapters()
	stochasticUpstream.stochastic = func(*Graph, *cMatrix, *realVector, StochasticMatrixOptions) int { return 1 }
	if _, err := graph.stochasticMatrix(nil, StochasticMatrixOptions{}, &stochasticUpstream); err == nil {
		t.Fatal("stochastic upstream error nil")
	}
}
