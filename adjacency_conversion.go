package igraph

// #cgo pkg-config: igraph
// #include <igraph.h>
// #include "adjacency_cgo.h"
import "C"

import (
	"fmt"
	"math"
)

// AdjacencyMatrixMode controls which triangle is populated for undirected
// graphs. It is ignored for directed graphs. The zero value populates both
// triangles and produces a symmetric matrix.
type AdjacencyMatrixMode uint8

const (
	AdjacencyMatrixBoth AdjacencyMatrixMode = iota
	AdjacencyMatrixUpper
	AdjacencyMatrixLower
)

func (mode AdjacencyMatrixMode) cValue() (C.igraph_get_adjacency_t, error) {
	switch mode {
	case AdjacencyMatrixBoth:
		return C.IGRAPH_GET_ADJACENCY_BOTH, nil
	case AdjacencyMatrixUpper:
		return C.IGRAPH_GET_ADJACENCY_UPPER, nil
	case AdjacencyMatrixLower:
		return C.IGRAPH_GET_ADJACENCY_LOWER, nil
	default:
		return 0, fmt.Errorf("igraph: invalid adjacency matrix mode: %d", mode)
	}
}

// AdjacencyMatrixOptions controls dense adjacency extraction. Its zero value
// returns both triangles for an undirected graph and ignores loops.
type AdjacencyMatrixOptions struct {
	Mode  AdjacencyMatrixMode
	Loops AdjacencyLoops
}

// StochasticMatrixOptions controls dense stochastic-matrix extraction. Its
// zero value normalizes rows. ColumnWise instead normalizes columns.
type StochasticMatrixOptions struct {
	ColumnWise bool
}

// AdjacencyMatrix returns a dense vertex-ID-aligned adjacency matrix. Nil
// weights count every edge as one. Non-nil weights must contain one finite
// value per edge ID; parallel-edge weights are summed. For undirected graphs,
// Mode selects which triangle is populated. Loops controls the diagonal.
// Weights is borrowed only for the call. The returned Matrix is Go-owned and
// remains valid after g is closed.
//
//igraph:bind igraph_get_adjacency
func (g *Graph) AdjacencyMatrix(weights []float64, options AdjacencyMatrixOptions) (Matrix, error) {
	return g.adjacencyMatrix(weights, options, nil)
}

// StochasticMatrix returns the adjacency matrix normalized so every nonzero
// row, or column when ColumnWise is true, sums to one. Rows or columns for
// vertices with zero total weight remain zero. Nil weights count every edge as
// one. Non-nil weights must contain one non-negative finite value per edge ID.
// Weights is borrowed only for the call. The returned Matrix is Go-owned and
// remains valid after g is closed.
//
//igraph:bind igraph_get_stochastic
func (g *Graph) StochasticMatrix(weights []float64, options StochasticMatrixOptions) (Matrix, error) {
	return g.stochasticMatrix(weights, options, nil)
}

type adjacencyConversionAdapters struct {
	newMatrix     func(Matrix) (*cMatrix, error)
	newReal       func([]float64) (*realVector, error)
	convertMatrix func(*cMatrix) (Matrix, error)
	adjacency     func(*Graph, *cMatrix, *realVector, AdjacencyMatrixOptions) int
	stochastic    func(*Graph, *cMatrix, *realVector, StochasticMatrixOptions) int
}

func defaultAdjacencyConversionAdapters() adjacencyConversionAdapters {
	return adjacencyConversionAdapters{
		newMatrix: newCMatrix,
		newReal:   newRealVector,
		convertMatrix: func(matrix *cMatrix) (Matrix, error) {
			return matrix.matrix()
		},
		adjacency: func(graph *Graph, matrix *cMatrix, weights *realVector, options AdjacencyMatrixOptions) int {
			mode, _ := options.Mode.cValue()
			loops, _ := options.Loops.cValue()
			return int(C.go_igraph_get_adjacency(
				&graph.graph, &matrix.value, mode, edgeWeightPointer(weights), loops,
			))
		},
		stochastic: func(graph *Graph, matrix *cMatrix, weights *realVector, options StochasticMatrixOptions) int {
			return int(C.go_igraph_get_stochastic(
				&graph.graph, &matrix.value, booltoint(options.ColumnWise), edgeWeightPointer(weights),
			))
		},
	}
}

func resolvedAdjacencyConversionAdapters(adapters *adjacencyConversionAdapters) adjacencyConversionAdapters {
	if adapters == nil {
		return defaultAdjacencyConversionAdapters()
	}
	return *adapters
}

func newConversionWeightsLocked(g *Graph, values []float64, nonNegative bool, create func([]float64) (*realVector, error)) (*realVector, error) {
	if values == nil {
		return nil, nil
	}
	edgeCount := int(C.igraph_ecount(&g.graph))
	if len(values) != edgeCount {
		return nil, fmt.Errorf("igraph: weights slice length (%d) must match number of edges (%d)", len(values), edgeCount)
	}
	for index, value := range values {
		if math.IsNaN(value) || math.IsInf(value, 0) || (nonNegative && value < 0) {
			if nonNegative {
				return nil, fmt.Errorf("igraph: weight value at index %d must be non-negative finite: %v", index, value)
			}
			return nil, fmt.Errorf("igraph: weight value at index %d must be finite: %v", index, value)
		}
	}
	return create(values)
}

func (g *Graph) adjacencyMatrix(weightValues []float64, options AdjacencyMatrixOptions, adapters *adjacencyConversionAdapters) (Matrix, error) {
	if g == nil {
		return Matrix{}, ErrClosed
	}
	g.mu.RLock()
	defer g.mu.RUnlock()
	if g.closed {
		return Matrix{}, ErrClosed
	}
	if _, err := options.Mode.cValue(); err != nil {
		return Matrix{}, err
	}
	if _, err := options.Loops.cValue(); err != nil {
		return Matrix{}, err
	}
	resolved := resolvedAdjacencyConversionAdapters(adapters)
	weights, err := newConversionWeightsLocked(g, weightValues, false, resolved.newReal)
	if err != nil {
		return Matrix{}, err
	}
	if weights != nil {
		defer weights.close()
	}
	result, err := resolved.newMatrix(Matrix{})
	if err != nil {
		return Matrix{}, err
	}
	defer result.close()
	if code := resolved.adjacency(g, result, weights, options); code != int(C.IGRAPH_SUCCESS) {
		return Matrix{}, igraphError("convert graph to adjacency matrix", code)
	}
	return resolved.convertMatrix(result)
}

func (g *Graph) stochasticMatrix(weightValues []float64, options StochasticMatrixOptions, adapters *adjacencyConversionAdapters) (Matrix, error) {
	if g == nil {
		return Matrix{}, ErrClosed
	}
	g.mu.RLock()
	defer g.mu.RUnlock()
	if g.closed {
		return Matrix{}, ErrClosed
	}
	resolved := resolvedAdjacencyConversionAdapters(adapters)
	weights, err := newConversionWeightsLocked(g, weightValues, true, resolved.newReal)
	if err != nil {
		return Matrix{}, err
	}
	if weights != nil {
		defer weights.close()
	}
	result, err := resolved.newMatrix(Matrix{})
	if err != nil {
		return Matrix{}, err
	}
	defer result.close()
	if code := resolved.stochastic(g, result, weights, options); code != int(C.IGRAPH_SUCCESS) {
		return Matrix{}, igraphError("convert graph to stochastic matrix", code)
	}
	return resolved.convertMatrix(result)
}
