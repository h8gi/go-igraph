package igraph

// #cgo pkg-config: igraph
// #include <igraph.h>
// #include "laplacian_cgo.h"
import "C"

import (
	"fmt"
	"math"
)

// LaplacianNormalization selects the definition of the graph Laplacian. Its
// zero value selects D-A.
type LaplacianNormalization uint8

const (
	// LaplacianUnnormalized computes D-A.
	LaplacianUnnormalized LaplacianNormalization = iota
	// LaplacianSymmetric computes I-D^(-1/2) A D^(-1/2).
	LaplacianSymmetric
	// LaplacianLeft computes I-D^(-1) A.
	LaplacianLeft
	// LaplacianRight computes I-A D^(-1).
	LaplacianRight
)

func (normalization LaplacianNormalization) cValue() (C.igraph_laplacian_normalization_t, error) {
	switch normalization {
	case LaplacianUnnormalized:
		return C.IGRAPH_LAPLACIAN_UNNORMALIZED, nil
	case LaplacianSymmetric:
		return C.IGRAPH_LAPLACIAN_SYMMETRIC, nil
	case LaplacianLeft:
		return C.IGRAPH_LAPLACIAN_LEFT, nil
	case LaplacianRight:
		return C.IGRAPH_LAPLACIAN_RIGHT, nil
	default:
		return 0, fmt.Errorf("igraph: invalid Laplacian normalization: %d", normalization)
	}
}

// LaplacianOptions controls dense Laplacian construction. Direction selects
// in-, out-, or total degree on directed graphs and is ignored for undirected
// graphs. Weights may be nil for unit weights; otherwise it must contain one
// finite, non-negative value per edge and is borrowed only for the call.
type LaplacianOptions struct {
	Direction     DirectionMode
	Normalization LaplacianNormalization
	Weights       []float64
}

// Laplacian returns a vertex-ID-indexed square dense matrix. Parallel edges
// contribute additively. Loops contribute to degree and adjacency and cancel
// from the unnormalized diagonal according to the upstream definition. The
// returned immutable Matrix is Go-owned and remains valid after g is closed.
// Zero-degree vertices have zero rows and columns in normalized results.
//
//igraph:bind igraph_get_laplacian
func (g *Graph) Laplacian(options LaplacianOptions) (Matrix, error) {
	return g.laplacian(options, nil)
}

type laplacianAdapters struct {
	newMatrix func(Matrix) (*cMatrix, error)
	newReal   func([]float64) (*realVector, error)
	call      func(*Graph, *cMatrix, laplacianDirection, cLaplacianNormalization, *realVector) int
	convert   func(*cMatrix) (Matrix, error)
}

type laplacianDirection = C.igraph_neimode_t
type cLaplacianNormalization = C.igraph_laplacian_normalization_t

func defaultLaplacianAdapters() laplacianAdapters {
	return laplacianAdapters{
		newMatrix: newCMatrix,
		newReal:   newRealVector,
		call: func(graph *Graph, result *cMatrix, mode C.igraph_neimode_t, normalization C.igraph_laplacian_normalization_t, weights *realVector) int {
			var pointer *C.igraph_vector_t
			if weights != nil {
				pointer = &weights.value
			}
			return int(C.go_igraph_get_laplacian(&graph.graph, &result.value, mode, normalization, pointer))
		},
		convert: (*cMatrix).matrix,
	}
}

func resolvedLaplacianAdapters(adapters *laplacianAdapters) laplacianAdapters {
	if adapters == nil {
		return defaultLaplacianAdapters()
	}
	return *adapters
}

func validateLaplacianWeights(weights []float64, edgeCount int) error {
	if weights == nil {
		return nil
	}
	if len(weights) != edgeCount {
		return fmt.Errorf("igraph: weight count %d does not match edge count %d", len(weights), edgeCount)
	}
	for index, weight := range weights {
		if math.IsNaN(weight) || math.IsInf(weight, 0) || weight < 0 {
			return fmt.Errorf("igraph: weight at index %d must be finite and non-negative: %v", index, weight)
		}
	}
	return nil
}

func (g *Graph) laplacian(options LaplacianOptions, adapters *laplacianAdapters) (Matrix, error) {
	if g == nil {
		return Matrix{}, ErrClosed
	}
	g.mu.RLock()
	defer g.mu.RUnlock()
	if g.closed {
		return Matrix{}, ErrClosed
	}
	mode, err := options.Direction.cValue()
	if err != nil {
		return Matrix{}, err
	}
	normalization, err := options.Normalization.cValue()
	if err != nil {
		return Matrix{}, err
	}
	if err := validateLaplacianWeights(options.Weights, int(C.igraph_ecount(&g.graph))); err != nil {
		return Matrix{}, err
	}
	resolved := resolvedLaplacianAdapters(adapters)
	result, err := resolved.newMatrix(Matrix{})
	if err != nil {
		return Matrix{}, err
	}
	defer result.close()
	var weights *realVector
	if options.Weights != nil {
		weights, err = resolved.newReal(options.Weights)
		if err != nil {
			return Matrix{}, err
		}
		defer weights.close()
	}
	if code := resolved.call(g, result, mode, normalization, weights); code != int(C.IGRAPH_SUCCESS) {
		return Matrix{}, igraphError("calculate graph Laplacian", code)
	}
	matrix, err := resolved.convert(result)
	if err != nil {
		return Matrix{}, err
	}
	rows, columns := matrix.Dims()
	vertices := int(C.igraph_vcount(&g.graph))
	if rows != vertices || columns != vertices {
		return Matrix{}, fmt.Errorf("igraph: Laplacian dimensions are %d by %d, want %d by %d", rows, columns, vertices, vertices)
	}
	return matrix, nil
}
