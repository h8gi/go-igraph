package igraph

// #cgo pkg-config: igraph
// #include <igraph.h>
// #include "adjacency_cgo.h"
import "C"

import (
	"fmt"
	"math"
)

// AdjacencyMode controls how a square matrix is interpreted. Directed uses
// every matrix entry as an edge from its row vertex to its column vertex.
// Undirected requires a symmetric matrix. Upper and Lower use only the named
// triangle. Min, Plus, and Max combine each pair of transposed entries.
type AdjacencyMode uint8

const (
	AdjacencyDirected AdjacencyMode = iota
	AdjacencyUndirected
	AdjacencyUpper
	AdjacencyLower
	AdjacencyMin
	AdjacencyPlus
	AdjacencyMax
)

func (mode AdjacencyMode) cValue() (C.igraph_adjacency_t, error) {
	switch mode {
	case AdjacencyDirected:
		return C.IGRAPH_ADJ_DIRECTED, nil
	case AdjacencyUndirected:
		return C.IGRAPH_ADJ_UNDIRECTED, nil
	case AdjacencyUpper:
		return C.IGRAPH_ADJ_UPPER, nil
	case AdjacencyLower:
		return C.IGRAPH_ADJ_LOWER, nil
	case AdjacencyMin:
		return C.IGRAPH_ADJ_MIN, nil
	case AdjacencyPlus:
		return C.IGRAPH_ADJ_PLUS, nil
	case AdjacencyMax:
		return C.IGRAPH_ADJ_MAX, nil
	default:
		return 0, fmt.Errorf("igraph: invalid adjacency mode: %d", mode)
	}
}

// AdjacencyLoops controls how diagonal entries are interpreted. NoLoops
// ignores the diagonal. LoopsOnce creates/counts loops once. LoopsTwice uses
// the standard undirected adjacency convention where each loop contributes
// twice, so diagonal values must be even for unweighted construction.
type AdjacencyLoops uint8

const (
	AdjacencyNoLoops AdjacencyLoops = iota
	AdjacencyLoopsTwice
	AdjacencyLoopsOnce
)

func (loops AdjacencyLoops) cValue() (C.igraph_loops_t, error) {
	switch loops {
	case AdjacencyNoLoops:
		return C.IGRAPH_NO_LOOPS, nil
	case AdjacencyLoopsTwice:
		return C.IGRAPH_LOOPS_TWICE, nil
	case AdjacencyLoopsOnce:
		return C.IGRAPH_LOOPS_ONCE, nil
	default:
		return 0, fmt.Errorf("igraph: invalid adjacency loop mode: %d", loops)
	}
}

// AdjacencyOptions controls dense adjacency-matrix construction. Its zero value
// constructs a directed graph and ignores diagonal entries.
type AdjacencyOptions struct {
	Mode  AdjacencyMode
	Loops AdjacencyLoops
}

// WeightedAdjacencyResult contains an independently owned graph and one weight
// per edge ID. Graph must be closed by the caller. Weights is a non-nil,
// Go-owned slice that remains valid after graph closure.
type WeightedAdjacencyResult struct {
	Graph   *Graph
	Weights []float64
}

// NewAdjacency constructs a graph from a square dense matrix. Entries must be
// finite, non-negative integers and represent edge multiplicities. Matrix is
// borrowed only for the call. The returned graph is independently owned and
// must be closed.
//
//igraph:bind igraph_adjacency
func NewAdjacency(matrix Matrix, options AdjacencyOptions) (*Graph, error) {
	return newAdjacency(matrix, options, nil)
}

// NewWeightedAdjacency constructs a graph with one edge for each nonzero
// interpreted matrix entry and uses that entry as its weight. Entries must be
// finite; negative weights are preserved. Matrix is borrowed only for the call.
// The graph and edge-ID-aligned weights are independently Go-owned.
//
//igraph:bind igraph_weighted_adjacency
func NewWeightedAdjacency(matrix Matrix, options AdjacencyOptions) (WeightedAdjacencyResult, error) {
	return newWeightedAdjacency(matrix, options, nil)
}

type adjacencyCallResult struct {
	graph C.igraph_t
	code  int
}

type adjacencyAdapters struct {
	newMatrix   func(Matrix) (*cMatrix, error)
	newReal     func([]float64) (*realVector, error)
	convertReal func(*realVector) ([]float64, error)
	create      func(*cMatrix, AdjacencyOptions) adjacencyCallResult
	weighted    func(*cMatrix, *realVector, AdjacencyOptions) adjacencyCallResult
}

func defaultAdjacencyAdapters() adjacencyAdapters {
	return adjacencyAdapters{
		newMatrix: newCMatrix,
		newReal:   newRealVector,
		convertReal: func(vector *realVector) ([]float64, error) {
			return vector.slice()
		},
		create: func(matrix *cMatrix, options AdjacencyOptions) adjacencyCallResult {
			mode, _ := options.Mode.cValue()
			loops, _ := options.Loops.cValue()
			var graph C.igraph_t
			code := C.go_igraph_adjacency(&graph, &matrix.value, mode, loops)
			return adjacencyCallResult{graph: graph, code: int(code)}
		},
		weighted: func(matrix *cMatrix, weights *realVector, options AdjacencyOptions) adjacencyCallResult {
			mode, _ := options.Mode.cValue()
			loops, _ := options.Loops.cValue()
			var graph C.igraph_t
			code := C.go_igraph_weighted_adjacency(&graph, &matrix.value, mode, &weights.value, loops)
			return adjacencyCallResult{graph: graph, code: int(code)}
		},
	}
}

func resolvedAdjacencyAdapters(adapters *adjacencyAdapters) adjacencyAdapters {
	if adapters == nil {
		return defaultAdjacencyAdapters()
	}
	return *adapters
}

func validateAdjacencyMatrix(matrix Matrix, weighted bool) error {
	rows, columns := matrix.Dims()
	if rows != columns {
		return fmt.Errorf("igraph: adjacency matrix must be square: %d by %d", rows, columns)
	}
	size, err := matrixSize(rows, columns)
	if err != nil {
		return err
	}
	if len(matrix.values) != size {
		return fmt.Errorf("igraph: adjacency matrix has invalid storage length")
	}
	for index, value := range matrix.values {
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return fmt.Errorf("igraph: adjacency matrix entry %d must be finite: %v", index, value)
		}
		if !weighted && (value < 0 || math.Trunc(value) != value) {
			return fmt.Errorf("igraph: adjacency matrix entry %d must be a non-negative integer: %v", index, value)
		}
		if !weighted && value >= float64(math.MaxInt64) {
			return fmt.Errorf("igraph: adjacency matrix entry %d is too large: %v", index, value)
		}
	}
	return nil
}

func validateAdjacencyOptions(options AdjacencyOptions) error {
	if _, err := options.Mode.cValue(); err != nil {
		return err
	}
	_, err := options.Loops.cValue()
	return err
}

func newAdjacency(matrix Matrix, options AdjacencyOptions, adapters *adjacencyAdapters) (*Graph, error) {
	if err := validateAdjacencyOptions(options); err != nil {
		return nil, err
	}
	if err := validateAdjacencyMatrix(matrix, false); err != nil {
		return nil, err
	}
	resolved := resolvedAdjacencyAdapters(adapters)
	cMatrix, err := resolved.newMatrix(matrix)
	if err != nil {
		return nil, err
	}
	defer cMatrix.close()
	call := resolved.create(cMatrix, options)
	if call.code != int(C.IGRAPH_SUCCESS) {
		return nil, igraphError("construct graph from adjacency matrix", call.code)
	}
	return adoptInitializedGraph(&call.graph), nil
}

func newWeightedAdjacency(matrix Matrix, options AdjacencyOptions, adapters *adjacencyAdapters) (WeightedAdjacencyResult, error) {
	if err := validateAdjacencyOptions(options); err != nil {
		return WeightedAdjacencyResult{}, err
	}
	if err := validateAdjacencyMatrix(matrix, true); err != nil {
		return WeightedAdjacencyResult{}, err
	}
	resolved := resolvedAdjacencyAdapters(adapters)
	cMatrix, err := resolved.newMatrix(matrix)
	if err != nil {
		return WeightedAdjacencyResult{}, err
	}
	defer cMatrix.close()
	weights, err := resolved.newReal(nil)
	if err != nil {
		return WeightedAdjacencyResult{}, err
	}
	defer weights.close()
	call := resolved.weighted(cMatrix, weights, options)
	if call.code != int(C.IGRAPH_SUCCESS) {
		return WeightedAdjacencyResult{}, igraphError("construct weighted graph from adjacency matrix", call.code)
	}
	values, err := resolved.convertReal(weights)
	if err != nil {
		C.igraph_destroy(&call.graph)
		return WeightedAdjacencyResult{}, err
	}
	return WeightedAdjacencyResult{Graph: adoptInitializedGraph(&call.graph), Weights: values}, nil
}
