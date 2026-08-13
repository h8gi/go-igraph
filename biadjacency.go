package igraph

// #cgo pkg-config: igraph
// #include <igraph.h>
import "C"

import (
	"fmt"
	"math"
)

// BiadjacencyOptions controls graph construction from a biadjacency matrix.
// Rows become false-mode vertices and columns become true-mode vertices.
type BiadjacencyOptions struct {
	Directed  bool
	Direction DirectionMode
	// Multiple interprets each matrix value as an edge multiplicity. When false,
	// every nonzero matrix entry creates exactly one edge.
	Multiple bool
}

// WeightedBiadjacencyOptions controls weighted graph construction from a
// biadjacency matrix. Weighted construction creates one edge per nonzero entry,
// so it intentionally has no multiplicity option.
type WeightedBiadjacencyOptions struct {
	Directed  bool
	Direction DirectionMode
}

// WeightedBipartiteGraphResult contains an independently owned bipartite graph,
// its explicit partition, and one weight per edge ID. Graph must be closed by
// the caller. Partition and Weights are non-nil Go-owned values that survive
// graph closure.
type WeightedBipartiteGraphResult struct {
	Graph     *Graph
	Partition BipartitePartition
	Weights   []float64
}

// BiadjacencyResult contains a Go-owned biadjacency matrix and the source
// vertex IDs corresponding to its rows and columns. Rows are false-mode
// vertices and columns are true-mode vertices. All returned values remain
// valid after the source graph is closed.
type BiadjacencyResult struct {
	Matrix          Matrix
	RowVertexIDs    []int
	ColumnVertexIDs []int
}

// NewBiadjacency constructs a bipartite graph from matrix. Matrix entries must
// be finite and non-negative. With Multiple false, each nonzero entry creates
// one edge. With Multiple true, entries must be integers and create that many
// parallel edges. DirectionOut points from rows to columns, DirectionIn from
// columns to rows, and DirectionAll creates mutual edges on a directed graph;
// direction is ignored for an undirected graph. Matrix is borrowed only for the
// call. The returned graph and partition are independently Go-owned.
//
//igraph:bind igraph_biadjacency
func NewBiadjacency(matrix Matrix, options BiadjacencyOptions) (BipartiteGraphResult, error) {
	return newBiadjacency(matrix, options, nil)
}

// NewWeightedBiadjacency constructs a bipartite graph with one edge for every
// nonzero matrix entry and uses that entry as the edge weight. Entries must be
// finite; negative weights are preserved and zero means no edge. Direction
// follows NewBiadjacency. Matrix is borrowed only for the call. The returned
// graph, partition, and edge-ID-aligned weights are independently Go-owned.
//
//igraph:bind igraph_weighted_biadjacency
func NewWeightedBiadjacency(matrix Matrix, options WeightedBiadjacencyOptions) (WeightedBipartiteGraphResult, error) {
	return newWeightedBiadjacency(matrix, options, nil)
}

// Biadjacency converts g to a biadjacency matrix using partition. Edge
// directions are ignored. Nil weights count every edge as one; non-nil weights
// must contain one finite value per edge, and parallel-edge weights are summed.
// Partition and weights are borrowed only for the call. Returned matrix and ID
// slices are Go-owned and remain valid after g is closed.
//
//igraph:bind igraph_get_biadjacency
func (g *Graph) Biadjacency(partition BipartitePartition, weights []float64) (BiadjacencyResult, error) {
	return g.biadjacency(partition, weights, nil)
}

type biadjacencyGraphCallResult struct {
	graph C.igraph_t
	code  int
}

type weightedBiadjacencyCallResult struct {
	graph C.igraph_t
	code  int
}

type biadjacencyAdapters struct {
	newMatrix      func(Matrix) (*cMatrix, error)
	newBool        func([]bool) (*boolVector, error)
	newReal        func([]float64) (*realVector, error)
	newInt         func([]int) (*intVector, error)
	convertMatrix  func(*cMatrix) (Matrix, error)
	convertBool    func(*boolVector) ([]bool, error)
	convertReal    func(*realVector) ([]float64, error)
	convertInt     func(*intVector) ([]int, error)
	create         func(*cMatrix, *boolVector, BiadjacencyOptions) biadjacencyGraphCallResult
	createWeighted func(*cMatrix, *boolVector, *realVector, WeightedBiadjacencyOptions) weightedBiadjacencyCallResult
	get            func(*Graph, *boolVector, *realVector, *cMatrix, *intVector, *intVector) int
}

func defaultBiadjacencyAdapters() biadjacencyAdapters {
	return biadjacencyAdapters{
		newMatrix: newCMatrix, newBool: newBoolVector, newReal: newRealVector,
		newInt: newIntVector, convertMatrix: (*cMatrix).matrix,
		convertBool: (*boolVector).slice, convertReal: (*realVector).slice,
		convertInt: (*intVector).slice,
		create: func(matrix *cMatrix, partition *boolVector, options BiadjacencyOptions) biadjacencyGraphCallResult {
			var graph C.igraph_t
			mode, _ := options.Direction.cValue()
			code := C.igraph_biadjacency(&graph, &partition.value, &matrix.value, booltoint(options.Directed), mode, booltoint(options.Multiple))
			return biadjacencyGraphCallResult{graph: graph, code: int(code)}
		},
		createWeighted: func(matrix *cMatrix, partition *boolVector, weights *realVector, options WeightedBiadjacencyOptions) weightedBiadjacencyCallResult {
			var graph C.igraph_t
			mode, _ := options.Direction.cValue()
			code := C.igraph_weighted_biadjacency(&graph, &partition.value, &weights.value, &matrix.value, booltoint(options.Directed), mode)
			return weightedBiadjacencyCallResult{graph: graph, code: int(code)}
		},
		get: func(graph *Graph, partition *boolVector, weights *realVector, matrix *cMatrix, rows, columns *intVector) int {
			return int(C.igraph_get_biadjacency(&graph.graph, &partition.value, realVectorPointer(weights), &matrix.value, &rows.value, &columns.value))
		},
	}
}

func resolvedBiadjacencyAdapters(adapters *biadjacencyAdapters) biadjacencyAdapters {
	if adapters == nil {
		return defaultBiadjacencyAdapters()
	}
	return *adapters
}

func validateBiadjacencyMatrix(matrix Matrix, multiple, weighted bool) error {
	rows, columns := matrix.Dims()
	if _, err := matrixSize(rows, columns); err != nil {
		return err
	}
	if len(matrix.values) != rows*columns {
		return fmt.Errorf("igraph: biadjacency matrix has %d values, want %d", len(matrix.values), rows*columns)
	}
	totalMultiplicity := 0
	for index, value := range matrix.values {
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return fmt.Errorf("igraph: biadjacency matrix value at index %d must be finite: %v", index, value)
		}
		if !weighted && value < 0 {
			return fmt.Errorf("igraph: unweighted biadjacency matrix value at index %d must be non-negative: %v", index, value)
		}
		if multiple && value != math.Trunc(value) {
			return fmt.Errorf("igraph: biadjacency multiplicity at index %d must be an integer: %v", index, value)
		}
		if multiple && value > float64(math.MaxInt) {
			return fmt.Errorf("igraph: biadjacency multiplicity at index %d is too large: %v", index, value)
		}
		if multiple {
			multiplicity := int(value)
			if totalMultiplicity > math.MaxInt-multiplicity {
				return fmt.Errorf("igraph: total biadjacency multiplicity is too large")
			}
			totalMultiplicity += multiplicity
		}
	}
	return nil
}

func validateBiadjacencyOptions(options BiadjacencyOptions) error {
	if _, err := options.Direction.cValue(); err != nil {
		return err
	}
	return nil
}

func validateWeightedBiadjacencyOptions(options WeightedBiadjacencyOptions) error {
	if _, err := options.Direction.cValue(); err != nil {
		return err
	}
	return nil
}

func newBiadjacency(matrix Matrix, options BiadjacencyOptions, adapters *biadjacencyAdapters) (BipartiteGraphResult, error) {
	if err := validateBiadjacencyOptions(options); err != nil {
		return BipartiteGraphResult{}, err
	}
	if err := validateBiadjacencyMatrix(matrix, options.Multiple, false); err != nil {
		return BipartiteGraphResult{}, err
	}
	if options.Directed && options.Direction == DirectionAll && options.Multiple {
		total := 0
		for _, value := range matrix.values {
			multiplicity := int(value)
			if total > math.MaxInt-multiplicity {
				return BipartiteGraphResult{}, fmt.Errorf("igraph: total directed biadjacency multiplicity is too large")
			}
			total += multiplicity
		}
		if total > math.MaxInt/2 {
			return BipartiteGraphResult{}, fmt.Errorf("igraph: mutual biadjacency edge count is too large")
		}
	}
	resolved := resolvedBiadjacencyAdapters(adapters)
	cMatrix, err := resolved.newMatrix(matrix)
	if err != nil {
		return BipartiteGraphResult{}, err
	}
	defer cMatrix.close()
	partition, err := resolved.newBool(nil)
	if err != nil {
		return BipartiteGraphResult{}, err
	}
	defer partition.close()
	call := resolved.create(cMatrix, partition, options)
	if call.code != int(C.IGRAPH_SUCCESS) {
		return BipartiteGraphResult{}, igraphError("construct graph from biadjacency matrix", call.code)
	}
	values, err := resolved.convertBool(partition)
	if err != nil {
		C.igraph_destroy(&call.graph)
		return BipartiteGraphResult{}, err
	}
	return BipartiteGraphResult{Graph: adoptInitializedGraph(&call.graph), Partition: BipartitePartition(values)}, nil
}

func newWeightedBiadjacency(matrix Matrix, options WeightedBiadjacencyOptions, adapters *biadjacencyAdapters) (WeightedBipartiteGraphResult, error) {
	if err := validateWeightedBiadjacencyOptions(options); err != nil {
		return WeightedBipartiteGraphResult{}, err
	}
	if err := validateBiadjacencyMatrix(matrix, false, true); err != nil {
		return WeightedBipartiteGraphResult{}, err
	}
	resolved := resolvedBiadjacencyAdapters(adapters)
	cMatrix, err := resolved.newMatrix(matrix)
	if err != nil {
		return WeightedBipartiteGraphResult{}, err
	}
	defer cMatrix.close()
	partition, err := resolved.newBool(nil)
	if err != nil {
		return WeightedBipartiteGraphResult{}, err
	}
	defer partition.close()
	weights, err := resolved.newReal(nil)
	if err != nil {
		return WeightedBipartiteGraphResult{}, err
	}
	defer weights.close()
	call := resolved.createWeighted(cMatrix, partition, weights, options)
	if call.code != int(C.IGRAPH_SUCCESS) {
		return WeightedBipartiteGraphResult{}, igraphError("construct weighted graph from biadjacency matrix", call.code)
	}
	values, err := resolved.convertBool(partition)
	if err != nil {
		C.igraph_destroy(&call.graph)
		return WeightedBipartiteGraphResult{}, err
	}
	weightValues, err := resolved.convertReal(weights)
	if err != nil {
		C.igraph_destroy(&call.graph)
		return WeightedBipartiteGraphResult{}, err
	}
	return WeightedBipartiteGraphResult{Graph: adoptInitializedGraph(&call.graph), Partition: BipartitePartition(values), Weights: weightValues}, nil
}

func (g *Graph) biadjacency(partition BipartitePartition, weightValues []float64, adapters *biadjacencyAdapters) (BiadjacencyResult, error) {
	if g == nil {
		return BiadjacencyResult{}, ErrClosed
	}
	g.mu.RLock()
	defer g.mu.RUnlock()
	if g.closed {
		return BiadjacencyResult{}, ErrClosed
	}
	if err := validateBipartitePartitionLength(partition, int(C.igraph_vcount(&g.graph))); err != nil {
		return BiadjacencyResult{}, err
	}
	if err := validatePartitionEdgesLocked(g, partition); err != nil {
		return BiadjacencyResult{}, err
	}
	resolved := resolvedBiadjacencyAdapters(adapters)
	types, err := resolved.newBool([]bool(partition))
	if err != nil {
		return BiadjacencyResult{}, err
	}
	defer types.close()
	var weights *realVector
	if weightValues != nil {
		if len(weightValues) != int(C.igraph_ecount(&g.graph)) {
			return BiadjacencyResult{}, fmt.Errorf("igraph: weight count %d does not match edge count %d", len(weightValues), int(C.igraph_ecount(&g.graph)))
		}
		for i, value := range weightValues {
			if math.IsNaN(value) || math.IsInf(value, 0) {
				return BiadjacencyResult{}, fmt.Errorf("igraph: weight at index %d must be finite: %v", i, value)
			}
		}
		weights, err = resolved.newReal(weightValues)
		if err != nil {
			return BiadjacencyResult{}, err
		}
		defer weights.close()
	}
	matrix, err := resolved.newMatrix(Matrix{})
	if err != nil {
		return BiadjacencyResult{}, err
	}
	defer matrix.close()
	rows, err := resolved.newInt(nil)
	if err != nil {
		return BiadjacencyResult{}, err
	}
	defer rows.close()
	columns, err := resolved.newInt(nil)
	if err != nil {
		return BiadjacencyResult{}, err
	}
	defer columns.close()
	if code := resolved.get(g, types, weights, matrix, rows, columns); code != int(C.IGRAPH_SUCCESS) {
		return BiadjacencyResult{}, igraphError("convert graph to biadjacency matrix", code)
	}
	resultMatrix, err := resolved.convertMatrix(matrix)
	if err != nil {
		return BiadjacencyResult{}, err
	}
	rowIDs, err := resolved.convertInt(rows)
	if err != nil {
		return BiadjacencyResult{}, err
	}
	columnIDs, err := resolved.convertInt(columns)
	if err != nil {
		return BiadjacencyResult{}, err
	}
	return BiadjacencyResult{Matrix: resultMatrix, RowVertexIDs: rowIDs, ColumnVertexIDs: columnIDs}, nil
}

func validatePartitionEdgesLocked(g *Graph, partition BipartitePartition) error {
	for edgeID := 0; edgeID < int(C.igraph_ecount(&g.graph)); edgeID++ {
		var from, to C.igraph_int_t
		if code := C.igraph_edge(&g.graph, C.igraph_int_t(edgeID), &from, &to); code != C.IGRAPH_SUCCESS {
			return igraphError("inspect bipartite edge", int(code))
		}
		if partition[int(from)] == partition[int(to)] {
			return fmt.Errorf("igraph: edge %d connects vertices in the same bipartite mode", edgeID)
		}
	}
	return nil
}
