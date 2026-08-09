package igraph

// #cgo pkg-config: igraph
// #include <igraph.h>
//
// static igraph_error_t go_igraph_density(
//     const igraph_t *graph, const igraph_vector_t *weights,
//     igraph_real_t *result, igraph_bool_t loops) {
//   igraph_error_handler_t *old_error =
//       igraph_set_error_handler(&igraph_error_handler_ignore);
//   igraph_warning_handler_t *old_warning =
//       igraph_set_warning_handler(&igraph_warning_handler_ignore);
//   igraph_error_t code = igraph_density(graph, weights, result, loops);
//   igraph_set_warning_handler(old_warning);
//   igraph_set_error_handler(old_error);
//   return code;
// }
//
// static igraph_error_t go_igraph_diameter(
//     const igraph_t *graph, const igraph_vector_t *weights,
//     igraph_real_t *length, igraph_int_t *from, igraph_int_t *to,
//     igraph_vector_int_t *vertices, igraph_vector_int_t *edges,
//     igraph_bool_t directed, igraph_bool_t unconnected) {
//   igraph_error_handler_t *old_error =
//       igraph_set_error_handler(&igraph_error_handler_ignore);
//   igraph_warning_handler_t *old_warning =
//       igraph_set_warning_handler(&igraph_warning_handler_ignore);
//   igraph_error_t code = igraph_diameter(
//       graph, weights, length, from, to, vertices, edges,
//       directed, unconnected);
//   igraph_set_warning_handler(old_warning);
//   igraph_set_error_handler(old_error);
//   return code;
// }
//
// static igraph_error_t go_igraph_average_path_length(
//     const igraph_t *graph, const igraph_vector_t *weights,
//     igraph_real_t *result, igraph_real_t *unconnected_pairs,
//     igraph_bool_t directed, igraph_bool_t unconnected) {
//   igraph_error_handler_t *old_error =
//       igraph_set_error_handler(&igraph_error_handler_ignore);
//   igraph_warning_handler_t *old_warning =
//       igraph_set_warning_handler(&igraph_warning_handler_ignore);
//   igraph_error_t code = igraph_average_path_length(
//       graph, weights, result, unconnected_pairs, directed, unconnected);
//   igraph_set_warning_handler(old_warning);
//   igraph_set_error_handler(old_error);
//   return code;
// }
//
// static igraph_error_t go_igraph_transitivity_undirected(
//     const igraph_t *graph, igraph_real_t *result,
//     igraph_transitivity_mode_t mode) {
//   igraph_error_handler_t *old_error =
//       igraph_set_error_handler(&igraph_error_handler_ignore);
//   igraph_warning_handler_t *old_warning =
//       igraph_set_warning_handler(&igraph_warning_handler_ignore);
//   igraph_error_t code = igraph_transitivity_undirected(graph, result, mode);
//   igraph_set_warning_handler(old_warning);
//   igraph_set_error_handler(old_error);
//   return code;
// }
//
// static igraph_error_t go_igraph_transitivity_local_undirected(
//     const igraph_t *graph, igraph_vector_t *result, igraph_vs_t vertices,
//     igraph_transitivity_mode_t mode) {
//   igraph_error_handler_t *old_error =
//       igraph_set_error_handler(&igraph_error_handler_ignore);
//   igraph_warning_handler_t *old_warning =
//       igraph_set_warning_handler(&igraph_warning_handler_ignore);
//   igraph_error_t code = igraph_transitivity_local_undirected(
//       graph, result, vertices, mode);
//   igraph_set_warning_handler(old_warning);
//   igraph_set_error_handler(old_error);
//   return code;
// }
//
// static igraph_error_t go_igraph_transitivity_avglocal_undirected(
//     const igraph_t *graph, igraph_real_t *result,
//     igraph_transitivity_mode_t mode) {
//   igraph_error_handler_t *old_error =
//       igraph_set_error_handler(&igraph_error_handler_ignore);
//   igraph_warning_handler_t *old_warning =
//       igraph_set_warning_handler(&igraph_warning_handler_ignore);
//   igraph_error_t code = igraph_transitivity_avglocal_undirected(
//       graph, result, mode);
//   igraph_set_warning_handler(old_warning);
//   igraph_set_error_handler(old_error);
//   return code;
// }
import "C"

import "fmt"

// DensityOptions controls the interpretation of density. IncludeLoops says
// that self-loops are possible and includes them in the denominator. A nil
// Weights slice counts edges; a non-nil slice is borrowed only for the call,
// copied into temporary C storage, and must contain one finite value per edge.
type DensityOptions struct {
	IncludeLoops bool
	Weights      []float64
}

// Density returns the ratio of the edge count, or total edge weight, to the
// number of possible adjacent vertex pairs. The graph must not have parallel
// edges. When IncludeLoops is false, callers must ensure that it has no loops.
// Multigraph and loop preconditions follow igraph and are not checked.
//
//igraph:bind igraph_density
func (g *Graph) Density(options DensityOptions) (float64, error) {
	if g == nil {
		return 0, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return 0, ErrClosed
	}

	weights, err := newOptionalEdgeWeights(options.Weights, int(C.igraph_ecount(&g.graph)))
	if err != nil {
		return 0, err
	}
	if weights != nil {
		defer weights.close()
	}
	var result C.igraph_real_t
	if code := C.go_igraph_density(
		&g.graph, edgeWeightPointer(weights), &result, booltoint(options.IncludeLoops),
	); code != C.IGRAPH_SUCCESS {
		return 0, igraphError("calculate density", int(code))
	}
	return float64(result), nil
}

// DistanceSummaryOptions controls whole-graph path summaries. Direction uses
// the same outgoing, incoming, and all-edge interpretation as other path
// queries. IgnoreUnreachable limits a disconnected result to reachable pairs;
// otherwise Diameter and AveragePathLength report positive infinity. A nil
// Weights slice selects an unweighted calculation. A non-nil slice is borrowed
// only for the call, copied into temporary C storage, and must contain one
// finite value per edge. The upstream summary algorithms reject negative
// weights.
type DistanceSummaryOptions struct {
	Direction         DirectionMode
	IgnoreUnreachable bool
	Weights           []float64
}

// DiameterResult is a Go-owned diameter length, endpoint pair, and one path
// attaining that length. Path.Found is false, From and To are -1, and the path
// slices are non-nil empty slices when no diameter path exists, including for
// an empty graph or a disconnected graph when unreachable pairs are not
// ignored.
type DiameterResult struct {
	Length float64
	From   int
	To     int
	Path   Path
}

// Diameter returns the longest shortest-path distance and a corresponding
// path. Direction is ignored for an undirected graph. All returned data is
// Go-owned and remains valid after the graph is closed.
//
//igraph:bind igraph_diameter
func (g *Graph) Diameter(options DistanceSummaryOptions) (DiameterResult, error) {
	if g == nil {
		return DiameterResult{}, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return DiameterResult{}, ErrClosed
	}

	directed, reverse, err := summaryDirection(options.Direction, C.igraph_is_directed(&g.graph) != booltoint(false))
	if err != nil {
		return DiameterResult{}, err
	}
	weights, err := newOptionalEdgeWeights(options.Weights, int(C.igraph_ecount(&g.graph)))
	if err != nil {
		return DiameterResult{}, err
	}
	if weights != nil {
		defer weights.close()
	}
	vertices, err := newIntVector(nil)
	if err != nil {
		return DiameterResult{}, err
	}
	defer vertices.close()
	edges, err := newIntVector(nil)
	if err != nil {
		return DiameterResult{}, err
	}
	defer edges.close()

	var length C.igraph_real_t
	var from, to C.igraph_int_t
	if code := C.go_igraph_diameter(
		&g.graph, edgeWeightPointer(weights), &length, &from, &to,
		&vertices.value, &edges.value, directed,
		booltoint(options.IgnoreUnreachable),
	); code != C.IGRAPH_SUCCESS {
		return DiameterResult{}, igraphError("calculate diameter", int(code))
	}
	vertexIDs, err := vertices.slice()
	if err != nil {
		return DiameterResult{}, err
	}
	edgeIDs, err := edges.slice()
	if err != nil {
		return DiameterResult{}, err
	}
	fromID, err := igraphIntToInt(from, "diameter source")
	if err != nil {
		return DiameterResult{}, err
	}
	toID, err := igraphIntToInt(to, "diameter target")
	if err != nil {
		return DiameterResult{}, err
	}
	if reverse && len(vertexIDs) != 0 {
		reverseInts(vertexIDs)
		reverseInts(edgeIDs)
		fromID, toID = toID, fromID
	}
	return DiameterResult{
		Length: float64(length),
		From:   fromID,
		To:     toID,
		Path: Path{
			Vertices: vertexIDs,
			Edges:    edgeIDs,
			Found:    len(vertexIDs) != 0,
		},
	}, nil
}

// AveragePathLengthResult is a Go-owned average shortest-path length and the
// number of ordered pairs whose target is unreachable from their source.
type AveragePathLengthResult struct {
	Length           float64
	UnreachablePairs float64
}

// AveragePathLength averages over all distinct ordered vertex pairs. With
// IgnoreUnreachable, only reachable pairs contribute; otherwise a graph with
// unreachable pairs has positive-infinite Length. Length is NaN when no pair
// can contribute. Direction is ignored for an undirected graph.
//
//igraph:bind igraph_average_path_length
func (g *Graph) AveragePathLength(options DistanceSummaryOptions) (AveragePathLengthResult, error) {
	if g == nil {
		return AveragePathLengthResult{}, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return AveragePathLengthResult{}, ErrClosed
	}

	directed, _, err := summaryDirection(options.Direction, C.igraph_is_directed(&g.graph) != booltoint(false))
	if err != nil {
		return AveragePathLengthResult{}, err
	}
	weights, err := newOptionalEdgeWeights(options.Weights, int(C.igraph_ecount(&g.graph)))
	if err != nil {
		return AveragePathLengthResult{}, err
	}
	if weights != nil {
		defer weights.close()
	}
	var length, unreachable C.igraph_real_t
	if code := C.go_igraph_average_path_length(
		&g.graph, edgeWeightPointer(weights), &length, &unreachable,
		directed, booltoint(options.IgnoreUnreachable),
	); code != C.IGRAPH_SUCCESS {
		return AveragePathLengthResult{}, igraphError("calculate average path length", int(code))
	}
	return AveragePathLengthResult{
		Length:           float64(length),
		UnreachablePairs: float64(unreachable),
	}, nil
}

// TransitivityMode controls how undefined clustering coefficients are treated.
// For global and per-vertex results, its zero value, TransitivityNaN, preserves
// undefined values as NaN; TransitivityZero substitutes zero. For an average
// local result, TransitivityNaN excludes undefined vertices from the mean and
// TransitivityZero includes them with a zero coefficient.
type TransitivityMode uint8

const (
	TransitivityNaN TransitivityMode = iota
	TransitivityZero
)

func (mode TransitivityMode) cValue() (C.igraph_transitivity_mode_t, error) {
	switch mode {
	case TransitivityNaN:
		return C.IGRAPH_TRANSITIVITY_NAN, nil
	case TransitivityZero:
		return C.IGRAPH_TRANSITIVITY_ZERO, nil
	default:
		return 0, fmt.Errorf("igraph: invalid transitivity mode: %d", mode)
	}
}

// GlobalTransitivity returns the ratio of closed connected triples to all
// connected triples. Edge directions and multiplicities are ignored.
//
//igraph:bind igraph_transitivity_undirected
func (g *Graph) GlobalTransitivity(mode TransitivityMode) (float64, error) {
	if g == nil {
		return 0, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return 0, ErrClosed
	}
	cMode, err := mode.cValue()
	if err != nil {
		return 0, err
	}
	var result C.igraph_real_t
	if code := C.go_igraph_transitivity_undirected(&g.graph, &result, cMode); code != C.IGRAPH_SUCCESS {
		return 0, igraphError("calculate global transitivity", int(code))
	}
	return float64(result), nil
}

// LocalTransitivity returns one clustering coefficient per selected vertex in
// materialized selector order, including duplicates. The selector is borrowed
// only for the call and the returned non-nil slice is Go-owned. Edge directions
// and multiplicities are ignored.
//
//igraph:bind igraph_transitivity_local_undirected
func (g *Graph) LocalTransitivity(vertices VertexSelector, mode TransitivityMode) ([]float64, error) {
	if g == nil {
		return nil, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return nil, ErrClosed
	}
	cMode, err := mode.cValue()
	if err != nil {
		return nil, err
	}
	if err := validateVertexSelector(vertices, int(C.igraph_vcount(&g.graph))); err != nil {
		return nil, err
	}
	vertexIDs, err := materializeVertexIDs(&g.graph, vertices)
	if err != nil {
		return nil, fmt.Errorf("igraph: materialize local transitivity selector: %w", err)
	}
	uniqueIDs, resultIndexes := uniqueVertexIDs(vertexIDs)
	cVertices := vertices
	if len(uniqueIDs) != len(vertexIDs) {
		cVertices, err = VertexIDs(uniqueIDs...)
		if err != nil {
			return nil, fmt.Errorf("igraph: build unique local transitivity selector: %w", err)
		}
	}
	selector, err := newCVertexSelector(cVertices)
	if err != nil {
		return nil, err
	}
	defer selector.close()
	result, err := newRealVector(nil)
	if err != nil {
		return nil, err
	}
	defer result.close()
	if code := C.go_igraph_transitivity_local_undirected(
		&g.graph, &result.value, selector.value, cMode,
	); code != C.IGRAPH_SUCCESS {
		return nil, igraphError("calculate local transitivity", int(code))
	}
	values, err := result.slice()
	if err != nil {
		return nil, err
	}
	if len(uniqueIDs) == len(vertexIDs) {
		return values, nil
	}
	expanded := make([]float64, len(resultIndexes))
	for index, resultIndex := range resultIndexes {
		expanded[index] = values[resultIndex]
	}
	return expanded, nil
}

// AverageLocalTransitivity returns the mean of vertex-level clustering
// coefficients. With TransitivityNaN, vertices with fewer than two neighbors
// are excluded from the mean, and the result is NaN if no vertex remains. With
// TransitivityZero, those vertices are included with coefficient zero. Edge
// directions and multiplicities are ignored.
//
//igraph:bind igraph_transitivity_avglocal_undirected
func (g *Graph) AverageLocalTransitivity(mode TransitivityMode) (float64, error) {
	if g == nil {
		return 0, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return 0, ErrClosed
	}
	cMode, err := mode.cValue()
	if err != nil {
		return 0, err
	}
	var result C.igraph_real_t
	if code := C.go_igraph_transitivity_avglocal_undirected(&g.graph, &result, cMode); code != C.IGRAPH_SUCCESS {
		return 0, igraphError("calculate average local transitivity", int(code))
	}
	return float64(result), nil
}

func summaryDirection(mode DirectionMode, graphDirected bool) (directed, reverse C.igraph_bool_t, err error) {
	if _, err := mode.cValue(); err != nil {
		return booltoint(false), booltoint(false), err
	}
	if !graphDirected || mode == DirectionAll {
		return booltoint(false), booltoint(false), nil
	}
	return booltoint(true), booltoint(mode == DirectionIn), nil
}

func reverseInts(values []int) {
	for left, right := 0, len(values)-1; left < right; left, right = left+1, right-1 {
		values[left], values[right] = values[right], values[left]
	}
}
