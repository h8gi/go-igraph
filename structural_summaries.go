package igraph

// #cgo pkg-config: igraph
// #include <igraph.h>
// #include "structural_summaries_cgo.h"
import "C"

import "fmt"

// MeanDegree returns the average in- or out-degree. In undirected graphs this
// is the ordinary mean degree. IncludeLoops controls whether self-loops
// contribute. A graph with no vertices returns NaN.
//
//igraph:bind igraph_mean_degree
func (g *Graph) MeanDegree(includeLoops bool) (float64, error) {
	if g == nil {
		return 0, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return 0, ErrClosed
	}
	var result C.igraph_real_t
	if code := C.go_igraph_mean_degree(&g.graph, &result, booltoint(includeLoops)); code != C.IGRAPH_SUCCESS {
		return 0, igraphError("calculate mean degree", int(code))
	}
	return float64(result), nil
}

// MaxDegree returns the largest degree among the selected vertices. Direction
// and loop handling match Degree. Empty selections return zero.
//
//igraph:bind igraph_maxdegree
func (g *Graph) MaxDegree(vertices VertexSelector, options DegreeOptions) (int, error) {
	if g == nil {
		return 0, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return 0, ErrClosed
	}
	mode, err := options.Direction.cValue()
	if err != nil {
		return 0, err
	}
	if err := validateVertexSelector(vertices, int(C.igraph_vcount(&g.graph))); err != nil {
		return 0, err
	}
	selector, err := newCVertexSelector(vertices)
	if err != nil {
		return 0, err
	}
	defer selector.close()
	var result C.igraph_int_t
	if code := C.go_igraph_maxdegree(
		&g.graph, &result, selector.value, mode, degreeLoops(options.CountLoops),
	); code != C.IGRAPH_SUCCESS {
		return 0, igraphError("calculate maximum degree", int(code))
	}
	return igraphIntToInt(result, "maximum degree")
}

// NearestNeighborDegreeOptions controls average neighbor-degree calculations.
// Direction selects which neighbors of each source are considered, while
// NeighborDegreeDirection selects which degree of those neighbors is averaged.
// Weights is optional, borrowed only for the call, and aligned with edge IDs.
type NearestNeighborDegreeOptions struct {
	Direction               DirectionMode
	NeighborDegreeDirection DirectionMode
	Weights                 []float64
}

// NearestNeighborDegreeResult contains selector-ordered averages and averages
// grouped by source degree. ByDegree index zero corresponds to degree one;
// absent degree classes and isolated selected vertices are NaN. Both slices are
// non-nil, Go-owned, and remain valid after graph closure.
type NearestNeighborDegreeResult struct {
	ByVertex []float64
	ByDegree []float64
}

// AverageNearestNeighborDegree computes the mean degree of each selected
// vertex's neighbors and the same quantity grouped by source degree. Explicit
// selector order and duplicates are preserved in ByVertex.
//
//igraph:bind igraph_avg_nearest_neighbor_degree
func (g *Graph) AverageNearestNeighborDegree(
	vertices VertexSelector,
	options NearestNeighborDegreeOptions,
) (NearestNeighborDegreeResult, error) {
	if g == nil {
		return NearestNeighborDegreeResult{}, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return NearestNeighborDegreeResult{}, ErrClosed
	}
	mode, err := options.Direction.cValue()
	if err != nil {
		return NearestNeighborDegreeResult{}, err
	}
	neighborMode, err := options.NeighborDegreeDirection.cValue()
	if err != nil {
		return NearestNeighborDegreeResult{}, err
	}
	if err := validateVertexSelector(vertices, int(C.igraph_vcount(&g.graph))); err != nil {
		return NearestNeighborDegreeResult{}, err
	}
	selector, err := newCVertexSelector(vertices)
	if err != nil {
		return NearestNeighborDegreeResult{}, err
	}
	defer selector.close()
	weights, err := newOptionalEdgeWeights(options.Weights, int(C.igraph_ecount(&g.graph)))
	if err != nil {
		return NearestNeighborDegreeResult{}, err
	}
	if weights != nil {
		defer weights.close()
	}
	byVertex, err := newRealVectorSize(0)
	if err != nil {
		return NearestNeighborDegreeResult{}, err
	}
	defer byVertex.close()
	byDegree, err := newRealVectorSize(0)
	if err != nil {
		return NearestNeighborDegreeResult{}, err
	}
	defer byDegree.close()
	if code := C.go_igraph_avg_nearest_neighbor_degree(
		&g.graph, selector.value, mode, neighborMode, &byVertex.value,
		&byDegree.value, edgeWeightPointer(weights),
	); code != C.IGRAPH_SUCCESS {
		return NearestNeighborDegreeResult{}, igraphError("calculate average nearest-neighbor degree", int(code))
	}
	result := NearestNeighborDegreeResult{}
	result.ByVertex, err = byVertex.slice()
	if err == nil {
		result.ByDegree, err = byDegree.slice()
	}
	if err != nil {
		return NearestNeighborDegreeResult{}, err
	}
	return result, nil
}

// DegreeCorrelationOptions controls degree correlation. FromDirection and
// ToDirection select the endpoint degrees. DirectedNeighbors controls whether
// directed edges contribute only source-to-target or in both orientations.
type DegreeCorrelationOptions struct {
	FromDirection     DirectionMode
	ToDirection       DirectionMode
	DirectedNeighbors bool
	Weights           []float64
}

// DegreeCorrelation returns k_nn(k), including degree zero at index zero.
// Missing degree classes are NaN. The returned non-nil slice is Go-owned.
//
//igraph:bind igraph_degree_correlation_vector
func (g *Graph) DegreeCorrelation(options DegreeCorrelationOptions) ([]float64, error) {
	if g == nil {
		return nil, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return nil, ErrClosed
	}
	fromMode, err := options.FromDirection.cValue()
	if err != nil {
		return nil, err
	}
	toMode, err := options.ToDirection.cValue()
	if err != nil {
		return nil, err
	}
	weights, err := newOptionalEdgeWeights(options.Weights, int(C.igraph_ecount(&g.graph)))
	if err != nil {
		return nil, err
	}
	if weights != nil {
		defer weights.close()
	}
	result, err := newRealVectorSize(0)
	if err != nil {
		return nil, err
	}
	defer result.close()
	if code := C.go_igraph_degree_correlation_vector(
		&g.graph, edgeWeightPointer(weights), &result.value, fromMode, toMode,
		booltoint(options.DirectedNeighbors),
	); code != C.IGRAPH_SUCCESS {
		return nil, igraphError("calculate degree correlation", int(code))
	}
	return result.slice()
}

// ReciprocityMode selects the definition of directed reciprocity.
type ReciprocityMode uint8

const (
	ReciprocityDefault ReciprocityMode = iota
	ReciprocityRatio
)

func (mode ReciprocityMode) cValue() (C.igraph_reciprocity_t, error) {
	switch mode {
	case ReciprocityDefault:
		return C.IGRAPH_RECIPROCITY_DEFAULT, nil
	case ReciprocityRatio:
		return C.IGRAPH_RECIPROCITY_RATIO, nil
	default:
		return 0, fmt.Errorf("igraph: invalid reciprocity mode: %d", mode)
	}
}

// Reciprocity returns the proportion of reciprocal directed connections.
// IgnoreLoops excludes self-loops. A directed graph with no contributing edges
// returns NaN; an undirected graph returns one.
//
//igraph:bind igraph_reciprocity
func (g *Graph) Reciprocity(mode ReciprocityMode, ignoreLoops bool) (float64, error) {
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
	if code := C.go_igraph_reciprocity(&g.graph, &result, booltoint(ignoreLoops), cMode); code != C.IGRAPH_SUCCESS {
		return 0, igraphError("calculate reciprocity", int(code))
	}
	return float64(result), nil
}

// Diversity returns normalized Shannon entropy of incident edge weights for
// each selected vertex. Weights is required, borrowed only for the call, must
// be finite and non-negative, and must align with edge IDs. The graph must be
// undirected and have no parallel edges. Isolates return NaN and vertices with
// one connection return zero. The non-nil result is selector-ordered and
// Go-owned.
//
//igraph:bind igraph_diversity
func (g *Graph) Diversity(vertices VertexSelector, weights []float64) ([]float64, error) {
	if g == nil {
		return nil, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return nil, ErrClosed
	}
	if weights == nil {
		return nil, fmt.Errorf("igraph: diversity weights must not be nil")
	}
	if err := validateVertexSelector(vertices, int(C.igraph_vcount(&g.graph))); err != nil {
		return nil, err
	}
	selector, err := newCVertexSelector(vertices)
	if err != nil {
		return nil, err
	}
	defer selector.close()
	cWeights, err := newOptionalNonNegativeEdgeWeights(weights, int(C.igraph_ecount(&g.graph)))
	if err != nil {
		return nil, err
	}
	defer cWeights.close()
	result, err := newRealVectorSize(0)
	if err != nil {
		return nil, err
	}
	defer result.close()
	if code := C.go_igraph_diversity(&g.graph, &cWeights.value, &result.value, selector.value); code != C.IGRAPH_SUCCESS {
		return nil, igraphError("calculate structural diversity", int(code))
	}
	return result.slice()
}

// RichClubOptions configures the experimental upstream rich-club sequence.
// VertexOrder must be a permutation of all vertex IDs and is borrowed only for
// the call. Result index i describes the graph remaining after i vertices have
// been removed. When Normalized is false, values are remaining edge counts or
// total weights and IncludeLoops is ignored.
type RichClubOptions struct {
	VertexOrder  []int
	Weights      []float64
	Normalized   bool
	IncludeLoops bool
	Directed     bool
}

// RichClubSequence returns the experimental density or edge-weight sequence
// produced by removing vertices in VertexOrder. The non-nil result is Go-owned.
//
//igraph:bind igraph_rich_club_sequence
func (g *Graph) RichClubSequence(options RichClubOptions) ([]float64, error) {
	if g == nil {
		return nil, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return nil, ErrClosed
	}
	vertexCount := int(C.igraph_vcount(&g.graph))
	if err := validateVertexPermutation(options.VertexOrder, vertexCount); err != nil {
		return nil, err
	}
	order, err := newIntVector(options.VertexOrder)
	if err != nil {
		return nil, err
	}
	defer order.close()
	weights, err := newOptionalEdgeWeights(options.Weights, int(C.igraph_ecount(&g.graph)))
	if err != nil {
		return nil, err
	}
	if weights != nil {
		defer weights.close()
	}
	result, err := newRealVectorSize(0)
	if err != nil {
		return nil, err
	}
	defer result.close()
	if code := C.go_igraph_rich_club_sequence(
		&g.graph, edgeWeightPointer(weights), &result.value, &order.value,
		booltoint(options.Normalized), booltoint(options.IncludeLoops),
		booltoint(options.Directed),
	); code != C.IGRAPH_SUCCESS {
		return nil, igraphError("calculate rich-club sequence", int(code))
	}
	return result.slice()
}

// VerticesByDegree returns selected vertex IDs ordered by degree. Descending
// is useful as the removal order for RichClubSequence. Ties follow upstream's
// deterministic ordering. The selector is borrowed and the non-nil result is
// Go-owned.
//
//igraph:bind igraph_sort_vertex_ids_by_degree
func (g *Graph) VerticesByDegree(
	vertices VertexSelector,
	options DegreeOptions,
	descending bool,
) ([]int, error) {
	if g == nil {
		return nil, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return nil, ErrClosed
	}
	mode, err := options.Direction.cValue()
	if err != nil {
		return nil, err
	}
	if err := validateVertexSelector(vertices, int(C.igraph_vcount(&g.graph))); err != nil {
		return nil, err
	}
	selector, err := newCVertexSelector(vertices)
	if err != nil {
		return nil, err
	}
	defer selector.close()
	result, err := newIntVector(nil)
	if err != nil {
		return nil, err
	}
	defer result.close()
	order := C.igraph_order_t(C.IGRAPH_ASCENDING)
	if descending {
		order = C.igraph_order_t(C.IGRAPH_DESCENDING)
	}
	if code := C.go_igraph_sort_vertex_ids_by_degree(
		&g.graph, &result.value, selector.value, mode,
		degreeLoops(options.CountLoops), order,
	); code != C.IGRAPH_SUCCESS {
		return nil, igraphError("sort vertices by degree", int(code))
	}
	return result.slice()
}

func degreeLoops(include bool) C.igraph_loops_t {
	if include {
		return C.igraph_loops_t(C.IGRAPH_LOOPS)
	}
	return C.igraph_loops_t(C.IGRAPH_NO_LOOPS)
}

func validateVertexPermutation(order []int, vertexCount int) error {
	if len(order) != vertexCount {
		return fmt.Errorf("igraph: vertex order length %d does not match vertex count %d", len(order), vertexCount)
	}
	seen := make([]bool, vertexCount)
	for index, vertex := range order {
		if vertex < 0 || vertex >= vertexCount {
			return fmt.Errorf("igraph: vertex order at index %d is %d, outside [0, %d)", index, vertex, vertexCount)
		}
		if seen[vertex] {
			return fmt.Errorf("igraph: vertex order contains duplicate vertex %d", vertex)
		}
		seen[vertex] = true
	}
	return nil
}
