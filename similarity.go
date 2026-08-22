package igraph

/*
#cgo pkg-config: igraph
#include <igraph.h>
#include "similarity_cgo.h"
*/
import "C"

import "fmt"

// NeighborhoodSimilarityMetric identifies a neighborhood-set similarity.
type NeighborhoodSimilarityMetric uint8

const (
	// SimilarityJaccard divides the common-neighbor count by the union size.
	SimilarityJaccard NeighborhoodSimilarityMetric = iota
	// SimilarityDice divides twice the common-neighbor count by the sum of sizes.
	SimilarityDice
)

// NeighborhoodSimilarityOptions controls a selector-aligned similarity matrix.
// Its zero value computes outgoing-neighbor Jaccard similarity without adding
// each compared vertex to its own neighborhood.
type NeighborhoodSimilarityOptions struct {
	Metric       NeighborhoodSimilarityMetric
	Direction    DirectionMode
	IncludeLoops bool
}

type similarityHooks struct {
	newResult func() (*cMatrix, error)
	run       func() error
}

type similarityVectorHooks struct {
	newResult func() (*realVector, error)
	run       func() error
}

// CitationCouplingKind identifies which directed citation relationship is
// counted. The two kinds are identical on undirected graphs.
type CitationCouplingKind uint8

const (
	// CouplingCocitation counts vertices that cite both compared vertices.
	CouplingCocitation CitationCouplingKind = iota
	// CouplingBibliographic counts vertices cited by both compared vertices.
	CouplingBibliographic
)

// NeighborhoodSimilarity returns a Go-owned matrix whose rows and columns
// follow the exact materialized orders of rows and columns, including
// duplicates. Selectors and options are borrowed only for this synchronous
// call. IncludeLoops adds each compared vertex to its own neighborhood.
//
//igraph:bind igraph_similarity_jaccard
//igraph:bind igraph_similarity_dice
func (g *Graph) NeighborhoodSimilarity(rows, columns VertexSelector, options NeighborhoodSimilarityOptions) (Matrix, error) {
	return g.neighborhoodSimilarity(rows, columns, options, similarityHooks{})
}

func (g *Graph) neighborhoodSimilarity(
	rows, columns VertexSelector,
	options NeighborhoodSimilarityOptions,
	hooks similarityHooks,
) (Matrix, error) {
	if g == nil {
		return Matrix{}, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return Matrix{}, ErrClosed
	}

	vertexCount := int(C.igraph_vcount(&g.graph))
	if err := validateVertexSelector(rows, vertexCount); err != nil {
		return Matrix{}, err
	}
	if err := validateVertexSelector(columns, vertexCount); err != nil {
		return Matrix{}, err
	}
	rowIDs, err := materializeVertexIDs(&g.graph, rows)
	if err != nil {
		return Matrix{}, err
	}
	columnIDs, err := materializeVertexIDs(&g.graph, columns)
	if err != nil {
		return Matrix{}, err
	}
	if _, err := matrixSize(len(rowIDs), len(columnIDs)); err != nil {
		return Matrix{}, err
	}
	combinedIDs := make([]int, 0, len(rowIDs)+len(columnIDs))
	combinedIDs = append(combinedIDs, rowIDs...)
	combinedIDs = append(combinedIDs, columnIDs...)
	uniqueIDs, positions := deduplicateIDs(combinedIDs)
	if _, err := matrixSize(len(uniqueIDs), len(uniqueIDs)); err != nil {
		return Matrix{}, err
	}
	mode, err := options.Direction.cValue()
	if err != nil {
		return Matrix{}, err
	}
	if options.Metric != SimilarityJaccard && options.Metric != SimilarityDice {
		return Matrix{}, fmt.Errorf("igraph: invalid neighborhood similarity metric: %d", options.Metric)
	}

	uniqueSelector, err := VertexIDs(uniqueIDs...)
	if err != nil {
		return Matrix{}, err
	}
	cVertices, err := newCVertexSelector(uniqueSelector)
	if err != nil {
		return Matrix{}, err
	}
	defer cVertices.close()
	result, err := newSimilarityResult(hooks)
	if err != nil {
		return Matrix{}, err
	}
	defer result.close()

	operation := func() error {
		var code C.igraph_error_t
		switch options.Metric {
		case SimilarityJaccard:
			code = C.go_igraph_similarity_jaccard(
				&g.graph, &result.value, cVertices.value, cVertices.value,
				mode, booltoint(options.IncludeLoops),
			)
		case SimilarityDice:
			code = C.go_igraph_similarity_dice(
				&g.graph, &result.value, cVertices.value, cVertices.value,
				mode, booltoint(options.IncludeLoops),
			)
		}
		if code != C.IGRAPH_SUCCESS {
			return igraphError("calculate neighborhood similarity", int(code))
		}
		return nil
	}
	if hooks.run != nil {
		operation = hooks.run
	}
	if err := operation(); err != nil {
		return Matrix{}, err
	}
	matrix, err := checkedSimilarityMatrix(result, len(uniqueIDs), len(uniqueIDs))
	if err != nil {
		return Matrix{}, err
	}
	return projectSimilarityMatrix(
		matrix,
		positions[:len(rowIDs)],
		positions[len(rowIDs):],
	)
}

// InverseLogWeightedSimilarity returns inverse-log-weighted common-neighbor
// scores. Rows follow the exact materialized selector order, including
// duplicates; columns are all vertices in vertex-ID order. The selector and
// direction are borrowed only for this synchronous call, and the result is
// Go-owned. Upstream always accounts for graph loops in this measure.
//
//igraph:bind igraph_similarity_inverse_log_weighted
func (g *Graph) InverseLogWeightedSimilarity(vertices VertexSelector, direction DirectionMode) (Matrix, error) {
	return g.inverseLogWeightedSimilarity(vertices, direction, similarityHooks{})
}

func (g *Graph) inverseLogWeightedSimilarity(
	vertices VertexSelector,
	direction DirectionMode,
	hooks similarityHooks,
) (Matrix, error) {
	var mode C.igraph_neimode_t
	return g.selectedToAllSimilarity(
		vertices,
		"calculate inverse-log-weighted similarity",
		func() error {
			var err error
			mode, err = direction.cValue()
			return err
		},
		func(graph *C.igraph_t, result *C.igraph_matrix_t, selector C.igraph_vs_t) C.igraph_error_t {
			return C.go_igraph_similarity_inverse_log_weighted(graph, result, selector, mode)
		},
		hooks,
	)
}

// CitationCoupling returns selected-to-all coupling scores. Rows follow the
// exact materialized selector order, including duplicates; columns are all
// vertices in vertex-ID order. Cocitation compares common citing vertices and
// bibliographic coupling compares common cited vertices. The selector is
// borrowed for the call and the returned matrix is Go-owned.
//
//igraph:bind igraph_cocitation
//igraph:bind igraph_bibcoupling
func (g *Graph) CitationCoupling(vertices VertexSelector, kind CitationCouplingKind) (Matrix, error) {
	return g.selectedToAllSimilarity(
		vertices,
		"calculate citation coupling",
		func() error {
			if kind != CouplingCocitation && kind != CouplingBibliographic {
				return fmt.Errorf("igraph: invalid citation coupling kind: %d", kind)
			}
			return nil
		},
		func(graph *C.igraph_t, result *C.igraph_matrix_t, selector C.igraph_vs_t) C.igraph_error_t {
			if kind == CouplingCocitation {
				return C.go_igraph_cocitation(graph, result, selector)
			}
			return C.go_igraph_bibcoupling(graph, result, selector)
		},
		similarityHooks{},
	)
}

// NeighborhoodSimilarityPairs returns one score per vertex pair in input
// order, including repeated pairs. Pair orientation does not change these
// symmetric scores, while Direction controls which directed neighborhoods are
// compared. Pairs and options are borrowed only for this synchronous call; the
// returned non-nil slice is Go-owned.
//
//igraph:bind igraph_similarity_jaccard_pairs
//igraph:bind igraph_similarity_dice_pairs
func (g *Graph) NeighborhoodSimilarityPairs(pairs []Edge, options NeighborhoodSimilarityOptions) ([]float64, error) {
	return g.neighborhoodSimilarityPairs(pairs, options, similarityVectorHooks{})
}

func (g *Graph) neighborhoodSimilarityPairs(
	pairs []Edge,
	options NeighborhoodSimilarityOptions,
	hooks similarityVectorHooks,
) ([]float64, error) {
	if g == nil {
		return nil, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return nil, ErrClosed
	}
	mode, err := similarityOptions(options)
	if err != nil {
		return nil, err
	}
	vertexCount := int(C.igraph_vcount(&g.graph))
	if len(pairs) > int(^uint(0)>>1)/2 {
		return nil, fmt.Errorf("igraph: vertex pair collection is too large")
	}
	endpoints := make([]int, 0, 2*len(pairs))
	for index, pair := range pairs {
		if err := validateEdge(pair, vertexCount, index); err != nil {
			return nil, err
		}
		endpoints = append(endpoints, pair.From, pair.To)
	}
	cPairs, err := newIntVector(endpoints)
	if err != nil {
		return nil, err
	}
	defer cPairs.close()
	result, err := newSimilarityVectorResult(hooks)
	if err != nil {
		return nil, err
	}
	defer result.close()
	run := func() error {
		var code C.igraph_error_t
		if options.Metric == SimilarityJaccard {
			code = C.go_igraph_similarity_jaccard_pairs(
				&g.graph, &result.value, &cPairs.value,
				mode, booltoint(options.IncludeLoops),
			)
		} else {
			code = C.go_igraph_similarity_dice_pairs(
				&g.graph, &result.value, &cPairs.value,
				mode, booltoint(options.IncludeLoops),
			)
		}
		if code != C.IGRAPH_SUCCESS {
			return igraphError("calculate pair neighborhood similarity", int(code))
		}
		return nil
	}
	if hooks.run != nil {
		run = hooks.run
	}
	if err := run(); err != nil {
		return nil, err
	}
	return checkedSimilarityValues(result, len(pairs))
}

// NeighborhoodSimilarityEdges returns one score for the endpoints of each
// selected edge. Results follow the selector's materialized edge-ID order,
// including duplicates. The selector and options are borrowed only for this
// synchronous call; the returned non-nil slice is Go-owned.
//
//igraph:bind igraph_similarity_jaccard_es
//igraph:bind igraph_similarity_dice_es
func (g *Graph) NeighborhoodSimilarityEdges(edges EdgeSelector, options NeighborhoodSimilarityOptions) ([]float64, error) {
	return g.neighborhoodSimilarityEdges(edges, options, similarityVectorHooks{})
}

func (g *Graph) neighborhoodSimilarityEdges(
	edges EdgeSelector,
	options NeighborhoodSimilarityOptions,
	hooks similarityVectorHooks,
) ([]float64, error) {
	if g == nil {
		return nil, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return nil, ErrClosed
	}
	mode, err := similarityOptions(options)
	if err != nil {
		return nil, err
	}
	edgeIDs, err := materializeSelectedEdgeIDs(&g.graph, edges)
	if err != nil {
		return nil, err
	}
	materialized, err := EdgeIDs(edgeIDs...)
	if err != nil {
		return nil, err
	}
	cEdges, err := newCEdgeSelector(materialized)
	if err != nil {
		return nil, err
	}
	defer cEdges.close()
	result, err := newSimilarityVectorResult(hooks)
	if err != nil {
		return nil, err
	}
	defer result.close()
	run := func() error {
		var code C.igraph_error_t
		if options.Metric == SimilarityJaccard {
			code = C.go_igraph_similarity_jaccard_es(
				&g.graph, &result.value, cEdges.value,
				mode, booltoint(options.IncludeLoops),
			)
		} else {
			code = C.go_igraph_similarity_dice_es(
				&g.graph, &result.value, cEdges.value,
				mode, booltoint(options.IncludeLoops),
			)
		}
		if code != C.IGRAPH_SUCCESS {
			return igraphError("calculate edge neighborhood similarity", int(code))
		}
		return nil
	}
	if hooks.run != nil {
		run = hooks.run
	}
	if err := run(); err != nil {
		return nil, err
	}
	return checkedSimilarityValues(result, len(edgeIDs))
}

func similarityOptions(options NeighborhoodSimilarityOptions) (C.igraph_neimode_t, error) {
	mode, err := options.Direction.cValue()
	if err != nil {
		return 0, err
	}
	if options.Metric != SimilarityJaccard && options.Metric != SimilarityDice {
		return 0, fmt.Errorf("igraph: invalid neighborhood similarity metric: %d", options.Metric)
	}
	return mode, nil
}

func newSimilarityVectorResult(hooks similarityVectorHooks) (*realVector, error) {
	if hooks.newResult != nil {
		return hooks.newResult()
	}
	return newRealVectorSize(0)
}

func checkedSimilarityValues(value *realVector, size int) ([]float64, error) {
	result, err := value.slice()
	if err != nil {
		return nil, err
	}
	if len(result) != size {
		return nil, fmt.Errorf(
			"igraph: similarity result length is %d, want %d",
			len(result), size,
		)
	}
	return result, nil
}

type selectedToAllOperation func(
	*C.igraph_t, *C.igraph_matrix_t, C.igraph_vs_t,
) C.igraph_error_t

func (g *Graph) selectedToAllSimilarity(
	vertices VertexSelector,
	description string,
	validate func() error,
	operation selectedToAllOperation,
	hooks similarityHooks,
) (Matrix, error) {
	if g == nil {
		return Matrix{}, ErrClosed
	}
	if operation == nil {
		return Matrix{}, fmt.Errorf("igraph: similarity operation is nil")
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return Matrix{}, ErrClosed
	}
	if validate != nil {
		if err := validate(); err != nil {
			return Matrix{}, err
		}
	}

	vertexCount := int(C.igraph_vcount(&g.graph))
	if err := validateVertexSelector(vertices, vertexCount); err != nil {
		return Matrix{}, err
	}
	selectedIDs, err := materializeVertexIDs(&g.graph, vertices)
	if err != nil {
		return Matrix{}, err
	}
	if _, err := matrixSize(len(selectedIDs), vertexCount); err != nil {
		return Matrix{}, err
	}
	uniqueIDs, rowPositions := deduplicateIDs(selectedIDs)
	uniqueSelector, err := VertexIDs(uniqueIDs...)
	if err != nil {
		return Matrix{}, err
	}
	cVertices, err := newCVertexSelector(uniqueSelector)
	if err != nil {
		return Matrix{}, err
	}
	defer cVertices.close()
	result, err := newSimilarityResult(hooks)
	if err != nil {
		return Matrix{}, err
	}
	defer result.close()
	run := func() error {
		if code := operation(&g.graph, &result.value, cVertices.value); code != C.IGRAPH_SUCCESS {
			return igraphError(description, int(code))
		}
		return nil
	}
	if hooks.run != nil {
		run = hooks.run
	}
	if err := run(); err != nil {
		return Matrix{}, err
	}
	matrix, err := checkedSimilarityMatrix(result, len(uniqueIDs), vertexCount)
	if err != nil {
		return Matrix{}, err
	}
	if len(uniqueIDs) == len(selectedIDs) {
		return matrix, nil
	}
	return expandMatrixRows(matrix, rowPositions)
}

func newSimilarityResult(hooks similarityHooks) (*cMatrix, error) {
	if hooks.newResult != nil {
		return hooks.newResult()
	}
	return newCMatrix(Matrix{})
}

func checkedSimilarityMatrix(value *cMatrix, rows, columns int) (Matrix, error) {
	result, err := value.matrix()
	if err != nil {
		return Matrix{}, err
	}
	actualRows, actualColumns := result.Dims()
	if actualRows != rows || actualColumns != columns {
		return Matrix{}, fmt.Errorf(
			"igraph: similarity result dimensions are %d by %d, want %d by %d",
			actualRows, actualColumns, rows, columns,
		)
	}
	return result, nil
}

func expandMatrixRows(matrix Matrix, positions []int) (Matrix, error) {
	_, columns := matrix.Dims()
	result, err := NewMatrix(len(positions), columns)
	if err != nil {
		return Matrix{}, err
	}
	for resultRow, sourceRow := range positions {
		copy(
			result.values[resultRow*columns:(resultRow+1)*columns],
			matrix.values[sourceRow*columns:(sourceRow+1)*columns],
		)
	}
	return result, nil
}

func projectSimilarityMatrix(matrix Matrix, rows, columns []int) (Matrix, error) {
	result, err := NewMatrix(len(rows), len(columns))
	if err != nil {
		return Matrix{}, err
	}
	for resultRow, sourceRow := range rows {
		for resultColumn, sourceColumn := range columns {
			result.values[resultRow*result.columns+resultColumn] =
				matrix.values[sourceRow*matrix.columns+sourceColumn]
		}
	}
	return result, nil
}
