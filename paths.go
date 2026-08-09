package igraph

// #cgo pkg-config: igraph
// #include <igraph.h>
//
// static igraph_error_t go_igraph_distances(
//     const igraph_t *graph, const igraph_vector_t *weights,
//     igraph_matrix_t *result, igraph_vs_t sources, igraph_vs_t targets,
//     igraph_neimode_t mode) {
//   igraph_error_handler_t *old_error =
//       igraph_set_error_handler(&igraph_error_handler_ignore);
//   igraph_warning_handler_t *old_warning =
//       igraph_set_warning_handler(&igraph_warning_handler_ignore);
//   igraph_error_t code =
//       igraph_distances(graph, weights, result, sources, targets, mode);
//   igraph_set_warning_handler(old_warning);
//   igraph_set_error_handler(old_error);
//   return code;
// }
//
// static igraph_error_t go_igraph_get_shortest_path(
//     const igraph_t *graph, const igraph_vector_t *weights,
//     igraph_vector_int_t *vertices, igraph_vector_int_t *edges,
//     igraph_int_t source, igraph_int_t target, igraph_neimode_t mode) {
//   igraph_error_handler_t *old_error =
//       igraph_set_error_handler(&igraph_error_handler_ignore);
//   igraph_warning_handler_t *old_warning =
//       igraph_set_warning_handler(&igraph_warning_handler_ignore);
//   igraph_error_t code = igraph_get_shortest_path(
//       graph, weights, vertices, edges, source, target, mode);
//   igraph_set_warning_handler(old_warning);
//   igraph_set_error_handler(old_error);
//   return code;
// }
import "C"

import (
	"fmt"
	"math"
)

// The pinned thread-safe igraph build stores handlers in thread-local state.
// Each C wrapper installs and restores its handlers within one cgo call, so
// handler state remains on the same OS thread without a Go-side mutex.

// PathOptions controls path and distance calculations. Direction is ignored
// for undirected graphs. A nil Weights slice requests an unweighted
// calculation. A non-nil slice is borrowed only for the call, copied into
// temporary C storage, and must contain one finite value per edge.
type PathOptions struct {
	Direction DirectionMode
	Weights   []float64
}

// Path is a Go-owned vertex and edge sequence. Found is false when the target
// is unreachable; in that case Vertices and Edges are non-nil empty slices.
// For a source equal to its target, Found is true, Vertices contains that
// vertex, and Edges is empty.
type Path struct {
	Vertices []int
	Edges    []int
	Found    bool
}

// Distances returns a dense distance matrix whose rows and columns follow the
// materialized order of sources and targets, including duplicates. Unreachable
// entries are positive infinity. Selector and weight inputs are borrowed only
// for the call; the returned Matrix is Go-owned and remains valid after the
// graph is closed.
//
//igraph:bind igraph_distances
func (g *Graph) Distances(sources, targets VertexSelector, options PathOptions) (Matrix, error) {
	if g == nil {
		return Matrix{}, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return Matrix{}, ErrClosed
	}

	cMode, err := options.Direction.cValue()
	if err != nil {
		return Matrix{}, err
	}
	vertexCount := int(C.igraph_vcount(&g.graph))
	if err := validateVertexSelector(sources, vertexCount); err != nil {
		return Matrix{}, fmt.Errorf("igraph: invalid source selector: %w", err)
	}
	if err := validateVertexSelector(targets, vertexCount); err != nil {
		return Matrix{}, fmt.Errorf("igraph: invalid target selector: %w", err)
	}
	targetIDs, err := materializeVertexIDs(&g.graph, targets)
	if err != nil {
		return Matrix{}, fmt.Errorf("igraph: materialize target selector: %w", err)
	}
	uniqueTargetIDs, targetColumns := uniqueVertexIDs(targetIDs)
	cTargetSelector := targets
	if len(uniqueTargetIDs) != len(targetIDs) {
		cTargetSelector, err = VertexIDs(uniqueTargetIDs...)
		if err != nil {
			return Matrix{}, fmt.Errorf("igraph: build unique target selector: %w", err)
		}
	}
	cSources, err := newCVertexSelector(sources)
	if err != nil {
		return Matrix{}, err
	}
	defer cSources.close()
	cTargets, err := newCVertexSelector(cTargetSelector)
	if err != nil {
		return Matrix{}, err
	}
	defer cTargets.close()

	weights, err := newOptionalEdgeWeights(options.Weights, int(C.igraph_ecount(&g.graph)))
	if err != nil {
		return Matrix{}, err
	}
	if weights != nil {
		defer weights.close()
	}
	result, err := newCMatrix(Matrix{})
	if err != nil {
		return Matrix{}, err
	}
	defer result.close()

	code := C.go_igraph_distances(
		&g.graph,
		edgeWeightPointer(weights),
		&result.value,
		cSources.value,
		cTargets.value,
		cMode,
	)
	if code != C.IGRAPH_SUCCESS {
		return Matrix{}, igraphError("calculate distances", int(code))
	}
	distances, err := result.matrix()
	if err != nil {
		return Matrix{}, err
	}
	if len(uniqueTargetIDs) == len(targetIDs) {
		return distances, nil
	}
	return expandDistanceColumns(distances, targetColumns)
}

// ShortestPath returns one shortest path between source and target. The path
// and options follow the same ownership, direction, and weight rules as
// Distances. When no path exists, Found is false rather than treating
// unreachability as an error.
//
//igraph:bind igraph_get_shortest_path
func (g *Graph) ShortestPath(source, target int, options PathOptions) (Path, error) {
	if g == nil {
		return Path{}, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return Path{}, ErrClosed
	}

	cMode, err := options.Direction.cValue()
	if err != nil {
		return Path{}, err
	}
	vertexCount := int(C.igraph_vcount(&g.graph))
	if err := validateVertexID(source, vertexCount); err != nil {
		return Path{}, fmt.Errorf("igraph: invalid path source: %w", err)
	}
	if err := validateVertexID(target, vertexCount); err != nil {
		return Path{}, fmt.Errorf("igraph: invalid path target: %w", err)
	}
	weights, err := newOptionalEdgeWeights(options.Weights, int(C.igraph_ecount(&g.graph)))
	if err != nil {
		return Path{}, err
	}
	if weights != nil {
		defer weights.close()
	}
	vertices, err := newIntVector(nil)
	if err != nil {
		return Path{}, err
	}
	defer vertices.close()
	edges, err := newIntVector(nil)
	if err != nil {
		return Path{}, err
	}
	defer edges.close()

	code := C.go_igraph_get_shortest_path(
		&g.graph,
		edgeWeightPointer(weights),
		&vertices.value,
		&edges.value,
		C.igraph_int_t(source),
		C.igraph_int_t(target),
		cMode,
	)
	if code != C.IGRAPH_SUCCESS {
		return Path{}, igraphError("calculate shortest path", int(code))
	}
	vertexIDs, err := vertices.slice()
	if err != nil {
		return Path{}, err
	}
	edgeIDs, err := edges.slice()
	if err != nil {
		return Path{}, err
	}
	return Path{Vertices: vertexIDs, Edges: edgeIDs, Found: len(vertexIDs) != 0}, nil
}

// uniqueVertexIDs preserves first-occurrence order and returns, for every
// original ID, the column containing that ID in the unique result.
func uniqueVertexIDs(ids []int) ([]int, []int) {
	unique := make([]int, 0, len(ids))
	columns := make([]int, len(ids))
	seen := make(map[int]int, len(ids))
	for index, id := range ids {
		column, ok := seen[id]
		if !ok {
			column = len(unique)
			seen[id] = column
			unique = append(unique, id)
		}
		columns[index] = column
	}
	return unique, columns
}

func expandDistanceColumns(distances Matrix, columns []int) (Matrix, error) {
	rows, _ := distances.Dims()
	result, err := NewMatrix(rows, len(columns))
	if err != nil {
		return Matrix{}, err
	}
	for row := 0; row < rows; row++ {
		for resultColumn, sourceColumn := range columns {
			result.values[row*result.columns+resultColumn] =
				distances.values[row*distances.columns+sourceColumn]
		}
	}
	return result, nil
}

func newOptionalEdgeWeights(values []float64, edgeCount int) (*realVector, error) {
	if values == nil {
		return nil, nil
	}
	if len(values) != edgeCount {
		return nil, fmt.Errorf(
			"igraph: weight count %d does not match edge count %d",
			len(values), edgeCount,
		)
	}
	for index, value := range values {
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return nil, fmt.Errorf("igraph: weight at index %d must be finite: %v", index, value)
		}
	}
	return newRealVector(values)
}

//igraph:internal igraph_set_error_handler
//igraph:internal igraph_set_warning_handler
func edgeWeightPointer(weights *realVector) *C.igraph_vector_t {
	if weights == nil {
		return nil
	}
	return &weights.value
}
