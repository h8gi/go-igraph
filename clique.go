package igraph

// #cgo pkg-config: igraph
// #include <igraph.h>
// #include "algorithm_cgo.h"
// #include "clique_cgo.h"
import "C"

import (
	"fmt"
	"math"
	"sort"
)

// VertexSetRange gives optional inclusive size bounds for clique-family
// enumeration. Nil bounds mean unbounded. Non-nil bounds must be positive and
// Minimum must not exceed Maximum. Bound pointers are read only during a call.
type VertexSetRange struct {
	Minimum *int
	Maximum *int
}

// VertexSetEnumerationOptions controls an exponential clique-family query.
// MaxResults must be positive. Range uses inclusive optional bounds.
type VertexSetEnumerationOptions struct {
	Range      VertexSetRange
	MaxResults int
}

// VertexSetEnumeration is a bounded, entirely Go-owned collection of vertex
// sets. Sets and every nested slice are non-nil. Truncated reports that at
// least one additional matching set existed.
type VertexSetEnumeration struct {
	Sets      [][]int
	Truncated bool
}

func (r VertexSetRange) validate() error {
	if r.Minimum != nil && *r.Minimum <= 0 {
		return fmt.Errorf("igraph: minimum vertex-set size must be positive: %d", *r.Minimum)
	}
	if r.Maximum != nil && *r.Maximum <= 0 {
		return fmt.Errorf("igraph: maximum vertex-set size must be positive: %d", *r.Maximum)
	}
	if r.Minimum != nil {
		if _, err := intToIgraphInt(*r.Minimum, "minimum vertex-set size"); err != nil {
			return err
		}
	}
	if r.Maximum != nil {
		if _, err := intToIgraphInt(*r.Maximum, "maximum vertex-set size"); err != nil {
			return err
		}
	}
	if r.Minimum != nil && r.Maximum != nil && *r.Minimum > *r.Maximum {
		return fmt.Errorf("igraph: minimum vertex-set size %d exceeds maximum %d", *r.Minimum, *r.Maximum)
	}
	return nil
}

func (o VertexSetEnumerationOptions) validate() error {
	if o.MaxResults <= 0 {
		return fmt.Errorf("igraph: maximum results must be positive: %d", o.MaxResults)
	}
	if _, err := intToIgraphInt(o.MaxResults, "maximum results"); err != nil {
		return err
	}
	if o.MaxResults == int(^uint(0)>>1) {
		return fmt.Errorf("igraph: maximum results is too large to detect truncation: %d", o.MaxResults)
	}
	if _, err := intToIgraphInt(o.MaxResults+1, "maximum results plus one"); err != nil {
		return err
	}
	return o.Range.validate()
}

func (r VertexSetRange) cBounds() (C.igraph_int_t, C.igraph_int_t) {
	var minimum, maximum C.igraph_int_t
	if r.Minimum != nil {
		minimum = C.igraph_int_t(*r.Minimum)
	}
	if r.Maximum != nil {
		maximum = C.igraph_int_t(*r.Maximum)
	}
	return minimum, maximum
}

// IsComplete reports whether every pair of distinct vertices is adjacent. The
// null and singleton graphs are complete. Loops do not affect the result;
// parallel edges do not change adjacency.
//
//igraph:bind igraph_is_complete
func (g *Graph) IsComplete() (bool, error) {
	if g == nil {
		return false, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return false, ErrClosed
	}
	var result C.igraph_bool_t
	if code := C.go_igraph_is_complete(&g.graph, &result); code != C.IGRAPH_SUCCESS {
		return false, igraphError("check whether graph is complete", int(code))
	}
	return result != booltoint(false), nil
}

// IsClique reports whether the selected vertices form a clique. Empty and
// singleton selections are cliques. When directed is false, edge directions
// are ignored; when true, adjacency is required in both directions. Loops and
// parallel edges do not affect adjacency. Duplicate selected IDs are rejected.
// The selector is borrowed only for the synchronous call and explicit IDs are
// copied into temporary C-owned storage.
//
//igraph:bind igraph_is_clique
func (g *Graph) IsClique(candidate VertexSelector, directed bool) (bool, error) {
	return g.vertexSetDecision(candidate, "clique", func(selector C.igraph_vs_t, result *C.igraph_bool_t) C.igraph_error_t {
		return C.go_igraph_is_clique(&g.graph, selector, booltoint(directed), result)
	})
}

// IsIndependentVertexSet reports whether no two selected vertices are
// adjacent. Empty and singleton selections are independent. Edge directions
// are ignored; loops and parallel edges do not affect the result. Duplicate
// selected IDs are rejected. The selector is borrowed only for the synchronous
// call and explicit IDs are copied into temporary C-owned storage.
//
//igraph:bind igraph_is_independent_vertex_set
func (g *Graph) IsIndependentVertexSet(candidate VertexSelector) (bool, error) {
	return g.vertexSetDecision(candidate, "independent vertex set", func(selector C.igraph_vs_t, result *C.igraph_bool_t) C.igraph_error_t {
		return C.go_igraph_is_independent_vertex_set(&g.graph, selector, result)
	})
}

func (g *Graph) vertexSetDecision(candidate VertexSelector, description string, query func(C.igraph_vs_t, *C.igraph_bool_t) C.igraph_error_t) (bool, error) {
	if g == nil {
		return false, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return false, ErrClosed
	}
	if err := validateVertexSelector(candidate, int(C.igraph_vcount(&g.graph))); err != nil {
		return false, err
	}
	if err := validateVertexSetSelector(candidate); err != nil {
		return false, err
	}
	selector, err := newCVertexSelector(candidate)
	if err != nil {
		return false, err
	}
	defer selector.close()
	var result C.igraph_bool_t
	if code := query(selector.value, &result); code != C.IGRAPH_SUCCESS {
		return false, igraphError("check "+description, int(code))
	}
	return result != booltoint(false), nil
}

func validateVertexSetSelector(selector VertexSelector) error {
	if selector.kind != vertexSelectorIDs {
		return nil
	}
	seen := make(map[int]struct{}, len(selector.ids))
	for index, id := range selector.ids {
		if _, exists := seen[id]; exists {
			return fmt.Errorf("igraph: duplicate vertex ID %d at selector index %d", id, index)
		}
		seen[id] = struct{}{}
	}
	return nil
}

// CliqueNumber returns the number of vertices in a largest clique. Edge
// directions, loops, and parallel edges are ignored.
//
//igraph:bind igraph_clique_number
func (g *Graph) CliqueNumber() (int, error) {
	return g.cliqueScalar("calculate clique number", false)
}

// IndependenceNumber returns the number of vertices in a largest independent
// vertex set. Edge directions, loops, and parallel edges are ignored.
//
//igraph:bind igraph_independence_number
func (g *Graph) IndependenceNumber() (int, error) {
	return g.cliqueScalar("calculate independence number", true)
}

func (g *Graph) cliqueScalar(operation string, independent bool) (int, error) {
	if g == nil {
		return 0, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return 0, ErrClosed
	}
	simple, err := newSimpleCliqueGraph(&g.graph, operation)
	if err != nil {
		return 0, err
	}
	defer C.igraph_destroy(simple)
	var result C.igraph_int_t
	var code C.igraph_error_t
	if independent {
		code = C.go_igraph_independence_number(simple, &result)
	} else {
		code = C.go_igraph_clique_number(simple, &result)
	}
	if code != C.IGRAPH_SUCCESS {
		return 0, igraphError(operation, int(code))
	}
	return igraphIntToInt(result, operation)
}

func newSimpleCliqueGraph(source *C.igraph_t, operation string) (*C.igraph_t, error) {
	// Pinned igraph's independent-set implementation assumes a simple graph.
	// Normalization also gives every clique-family API the same adjacency-only
	// semantics for loops and parallel edges.
	simple := &C.igraph_t{}
	if code := C.go_igraph_copy(simple, source); code != C.IGRAPH_SUCCESS {
		return nil, igraphError("copy graph for "+operation, int(code))
	}
	if code := C.go_igraph_simplify(simple, booltoint(true), booltoint(true)); code != C.IGRAPH_SUCCESS {
		C.igraph_destroy(simple)
		return nil, igraphError("simplify graph for "+operation, int(code))
	}
	return simple, nil
}

// Cliques returns at most MaxResults cliques whose sizes lie in the inclusive
// optional range. Edge directions, loops, and parallel edges are ignored.
// Vertex IDs within each clique are sorted; no outer result order is promised.
// The result is entirely Go-owned and remains valid after graph closure.
//
//igraph:bind igraph_cliques
func (g *Graph) Cliques(options VertexSetEnumerationOptions) (VertexSetEnumeration, error) {
	result := VertexSetEnumeration{Sets: make([][]int, 0)}
	if err := options.validate(); err != nil {
		return result, err
	}
	if g == nil {
		return result, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return result, ErrClosed
	}
	return enumerateCliquesLocked(&g.graph, options)
}

// LargestCliques returns at most maxResults cliques of maximum cardinality.
// It composes CliqueNumber with bounded clique enumeration and reports exact
// truncation. maxResults must be positive.
func (g *Graph) LargestCliques(maxResults int) (VertexSetEnumeration, error) {
	result := VertexSetEnumeration{Sets: make([][]int, 0)}
	options := VertexSetEnumerationOptions{MaxResults: maxResults}
	if err := options.validate(); err != nil {
		return result, err
	}
	if g == nil {
		return result, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return result, ErrClosed
	}
	simple, err := newSimpleCliqueGraph(&g.graph, "enumerate largest cliques")
	if err != nil {
		return result, err
	}
	defer C.igraph_destroy(simple)
	var number C.igraph_int_t
	if code := C.go_igraph_clique_number(simple, &number); code != C.IGRAPH_SUCCESS {
		return result, igraphError("calculate clique number", int(code))
	}
	size, err := igraphIntToInt(number, "clique number")
	if err != nil {
		return result, err
	}
	if size == 0 {
		return result, nil
	}
	options.Range.Minimum = &size
	options.Range.Maximum = &size
	return enumerateCliquesOnSimpleGraph(simple, options)
}

func enumerateCliquesLocked(graph *C.igraph_t, options VertexSetEnumerationOptions) (VertexSetEnumeration, error) {
	result := VertexSetEnumeration{Sets: make([][]int, 0)}
	simple, err := newSimpleCliqueGraph(graph, "enumerate cliques")
	if err != nil {
		return result, err
	}
	defer C.igraph_destroy(simple)
	return enumerateCliquesOnSimpleGraph(simple, options)
}

func enumerateCliquesOnSimpleGraph(graph *C.igraph_t, options VertexSetEnumerationOptions) (VertexSetEnumeration, error) {
	minimum, maximum := options.Range.cBounds()
	request := C.igraph_int_t(options.MaxResults + 1)
	return collectCliqueEnumeration(options, cliqueEnumerationOperations{
		newList:   newIntVectorList,
		closeList: (*intVectorList).close,
		query: func(list *intVectorList) error {
			if code := C.go_igraph_cliques(graph, &list.value, minimum, maximum, request); code != C.IGRAPH_SUCCESS {
				return igraphError("enumerate cliques", int(code))
			}
			return nil
		},
		listSlices: func(list *intVectorList) ([][]int, error) { return list.slices() },
	})
}

type cliqueEnumerationOperations struct {
	newList    func() (*intVectorList, error)
	closeList  func(*intVectorList)
	query      func(*intVectorList) error
	listSlices func(*intVectorList) ([][]int, error)
}

func collectCliqueEnumeration(options VertexSetEnumerationOptions, operations cliqueEnumerationOperations) (VertexSetEnumeration, error) {
	result := VertexSetEnumeration{Sets: make([][]int, 0)}
	list, err := operations.newList()
	if err != nil {
		return result, err
	}
	defer operations.closeList(list)
	if err := operations.query(list); err != nil {
		return result, err
	}
	sets, err := operations.listSlices(list)
	if err != nil {
		return result, err
	}
	for _, set := range sets {
		sort.Ints(set)
	}
	if len(sets) > options.MaxResults {
		sets = sets[:options.MaxResults]
		result.Truncated = true
	}
	result.Sets = sets
	return result, nil
}

// CliqueSizeHistogram returns counts indexed by clique size minus one. Counts
// outside the inclusive optional range are zero. Edge directions, loops, and
// parallel edges are ignored. The returned non-nil slice is Go-owned.
//
//igraph:bind igraph_clique_size_hist
func (g *Graph) CliqueSizeHistogram(sizeRange VertexSetRange) ([]int, error) {
	if err := sizeRange.validate(); err != nil {
		return []int{}, err
	}
	if g == nil {
		return []int{}, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return []int{}, ErrClosed
	}
	simple, err := newSimpleCliqueGraph(&g.graph, "calculate clique size histogram")
	if err != nil {
		return []int{}, err
	}
	defer C.igraph_destroy(simple)
	histogram, err := newRealVectorSize(0)
	if err != nil {
		return []int{}, err
	}
	defer histogram.close()
	minimum, maximum := sizeRange.cBounds()
	if code := C.go_igraph_clique_size_hist(simple, &histogram.value, minimum, maximum); code != C.IGRAPH_SUCCESS {
		return []int{}, igraphError("calculate clique size histogram", int(code))
	}
	values, err := histogram.slice()
	if err != nil {
		return []int{}, err
	}
	return cliqueHistogramCounts(values)
}

func cliqueHistogramCounts(values []float64) ([]int, error) {
	result := make([]int, len(values))
	for index, value := range values {
		if math.IsNaN(value) || math.IsInf(value, 0) || value < 0 || math.Trunc(value) != value {
			return nil, fmt.Errorf("igraph: clique histogram count at index %d is not a non-negative integer: %g", index, value)
		}
		converted := int(value)
		if converted < 0 || float64(converted) != value {
			return nil, fmt.Errorf("igraph: clique histogram count at index %d is out of Go int range: %g", index, value)
		}
		result[index] = converted
	}
	return result, nil
}

// MaximalCliques returns bounded cliques that cannot be extended by another
// vertex. Maximal does not mean maximum cardinality. The shared inclusive range,
// exact truncation, canonicalization, graph-shape, ordering, and ownership
// contracts are the same as for Cliques.
//
//igraph:bind igraph_maximal_cliques
func (g *Graph) MaximalCliques(options VertexSetEnumerationOptions) (VertexSetEnumeration, error) {
	return g.maximalCliques(nil, options)
}

// MaximalCliquesFromVertices enumerates maximal cliques from the specified
// initial/search vertices. initialVertices is not an induced-subgraph filter:
// a returned clique need not contain an initial vertex and may contain any
// graph vertex. The input partitions internal search roots, is borrowed and
// copied for the synchronous call. Empty input returns an empty result;
// duplicate and out-of-range IDs are rejected.
//
//igraph:bind igraph_maximal_cliques_subset
func (g *Graph) MaximalCliquesFromVertices(initialVertices []int, options VertexSetEnumerationOptions) (VertexSetEnumeration, error) {
	return g.maximalCliques(initialVertices, options)
}

func (g *Graph) maximalCliques(initialVertices []int, options VertexSetEnumerationOptions) (VertexSetEnumeration, error) {
	result := VertexSetEnumeration{Sets: make([][]int, 0)}
	if err := options.validate(); err != nil {
		return result, err
	}
	if g == nil {
		return result, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return result, ErrClosed
	}
	if initialVertices != nil {
		if err := validateInitialCliqueVertices(initialVertices, int(C.igraph_vcount(&g.graph))); err != nil {
			return result, err
		}
		if len(initialVertices) == 0 {
			return result, nil
		}
	}
	simple, err := newSimpleCliqueGraph(&g.graph, "enumerate maximal cliques")
	if err != nil {
		return result, err
	}
	defer C.igraph_destroy(simple)
	minimum, maximum := options.Range.cBounds()
	request := C.igraph_int_t(options.MaxResults + 1)
	if initialVertices == nil {
		return collectCliqueEnumeration(options, cliqueEnumerationOperations{
			newList: newIntVectorList, closeList: (*intVectorList).close,
			query: func(list *intVectorList) error {
				if code := C.go_igraph_maximal_cliques(simple, &list.value, minimum, maximum, request); code != C.IGRAPH_SUCCESS {
					return igraphError("enumerate maximal cliques", int(code))
				}
				return nil
			},
			listSlices: func(list *intVectorList) ([][]int, error) { return list.slices() },
		})
	}
	vertices, err := newIntVector(initialVertices)
	if err != nil {
		return result, err
	}
	defer vertices.close()
	return collectCliqueEnumeration(options, cliqueEnumerationOperations{
		newList: newIntVectorList, closeList: (*intVectorList).close,
		query: func(list *intVectorList) error {
			if code := C.go_igraph_maximal_cliques_subset(simple, &vertices.value, &list.value, minimum, maximum, request); code != C.IGRAPH_SUCCESS {
				return igraphError("enumerate maximal cliques from initial vertices", int(code))
			}
			return nil
		},
		listSlices: func(list *intVectorList) ([][]int, error) { return list.slices() },
	})
}

func validateInitialCliqueVertices(vertices []int, vertexCount int) error {
	seen := make(map[int]struct{}, len(vertices))
	for index, vertex := range vertices {
		if vertex < 0 || vertex >= vertexCount {
			return fmt.Errorf("igraph: initial clique vertex at index %d is %d, outside [0, %d)", index, vertex, vertexCount)
		}
		if _, exists := seen[vertex]; exists {
			return fmt.Errorf("igraph: duplicate initial clique vertex %d at index %d", vertex, index)
		}
		seen[vertex] = struct{}{}
	}
	return nil
}

// MaximalCliqueCount returns the number of maximal cliques in the inclusive
// optional size range using checked C-to-Go integer conversion.
//
//igraph:bind igraph_maximal_cliques_count
func (g *Graph) MaximalCliqueCount(sizeRange VertexSetRange) (int, error) {
	if err := sizeRange.validate(); err != nil {
		return 0, err
	}
	if g == nil {
		return 0, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return 0, ErrClosed
	}
	simple, err := newSimpleCliqueGraph(&g.graph, "count maximal cliques")
	if err != nil {
		return 0, err
	}
	defer C.igraph_destroy(simple)
	minimum, maximum := sizeRange.cBounds()
	var count C.igraph_int_t
	if code := C.go_igraph_maximal_cliques_count(simple, &count, minimum, maximum); code != C.IGRAPH_SUCCESS {
		return 0, igraphError("count maximal cliques", int(code))
	}
	return igraphIntToInt(count, "maximal clique count")
}

// MaximalCliqueSizeHistogram returns maximal-clique counts indexed by clique
// size minus one. The non-nil result is Go-owned and uses checked counts.
//
//igraph:bind igraph_maximal_cliques_hist
func (g *Graph) MaximalCliqueSizeHistogram(sizeRange VertexSetRange) ([]int, error) {
	if err := sizeRange.validate(); err != nil {
		return []int{}, err
	}
	if g == nil {
		return []int{}, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return []int{}, ErrClosed
	}
	simple, err := newSimpleCliqueGraph(&g.graph, "calculate maximal clique size histogram")
	if err != nil {
		return []int{}, err
	}
	defer C.igraph_destroy(simple)
	histogram, err := newRealVectorSize(0)
	if err != nil {
		return []int{}, err
	}
	defer histogram.close()
	minimum, maximum := sizeRange.cBounds()
	if code := C.go_igraph_maximal_cliques_hist(simple, &histogram.value, minimum, maximum); code != C.IGRAPH_SUCCESS {
		return []int{}, igraphError("calculate maximal clique size histogram", int(code))
	}
	values, err := histogram.slice()
	if err != nil {
		return []int{}, err
	}
	return cliqueHistogramCounts(values)
}

// WeightRange gives optional inclusive positive-integer bounds for weighted
// clique enumeration. Nil bounds mean unbounded. Pointers are borrowed only
// while validating and executing a synchronous call.
type WeightRange struct {
	Minimum *int
	Maximum *int
}

func (r WeightRange) validate() error {
	if r.Minimum != nil {
		if err := validateCliqueWeight(*r.Minimum, "minimum clique weight"); err != nil {
			return err
		}
	}
	if r.Maximum != nil {
		if err := validateCliqueWeight(*r.Maximum, "maximum clique weight"); err != nil {
			return err
		}
	}
	if r.Minimum != nil && r.Maximum != nil && *r.Minimum > *r.Maximum {
		return fmt.Errorf("igraph: minimum clique weight %d exceeds maximum %d", *r.Minimum, *r.Maximum)
	}
	return nil
}

func (r WeightRange) cBounds() (C.igraph_real_t, C.igraph_real_t) {
	var minimum, maximum C.igraph_real_t
	if r.Minimum != nil {
		minimum = C.igraph_real_t(*r.Minimum)
	}
	if r.Maximum != nil {
		maximum = C.igraph_real_t(*r.Maximum)
	}
	return minimum, maximum
}

// WeightedCliqueOptions controls bounded weighted-clique enumeration.
// MaxResults must be positive. MaximalOnly returns only cliques that cannot be
// extended, independently of their total weight.
type WeightedCliqueOptions struct {
	Range       WeightRange
	MaxResults  int
	MaximalOnly bool
}

func (o WeightedCliqueOptions) validate() error {
	if err := (VertexSetEnumerationOptions{MaxResults: o.MaxResults}).validate(); err != nil {
		return err
	}
	return o.Range.validate()
}

// WeightedCliques returns bounded cliques whose positive integer vertex-weight
// sums lie in the inclusive optional range. weights must contain exactly one
// value per vertex. The input is borrowed and copied into temporary C storage.
// Edge direction, loops, and parallel edges are ignored. Result ownership,
// canonicalization, ordering, and exact truncation match Cliques.
//
//igraph:bind igraph_weighted_cliques
func (g *Graph) WeightedCliques(weights []int, options WeightedCliqueOptions) (VertexSetEnumeration, error) {
	result := VertexSetEnumeration{Sets: make([][]int, 0)}
	if err := options.validate(); err != nil {
		return result, err
	}
	if g == nil {
		return result, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return result, ErrClosed
	}
	weightVector, err := newCliqueWeights(weights, int(C.igraph_vcount(&g.graph)))
	if err != nil {
		return result, err
	}
	defer weightVector.close()
	simple, err := newSimpleCliqueGraph(&g.graph, "enumerate weighted cliques")
	if err != nil {
		return result, err
	}
	defer C.igraph_destroy(simple)
	return enumerateWeightedCliquesOnSimple(simple, weightVector, options)
}

func enumerateWeightedCliquesOnSimple(simple *C.igraph_t, weightVector *realVector, options WeightedCliqueOptions) (VertexSetEnumeration, error) {
	minimum, maximum := options.Range.cBounds()
	request := C.igraph_int_t(options.MaxResults + 1)
	return collectCliqueEnumeration(
		VertexSetEnumerationOptions{MaxResults: options.MaxResults},
		cliqueEnumerationOperations{
			newList: newIntVectorList, closeList: (*intVectorList).close,
			query: func(list *intVectorList) error {
				if code := C.go_igraph_weighted_cliques(
					simple, &weightVector.value, &list.value, booltoint(options.MaximalOnly),
					minimum, maximum, request,
				); code != C.IGRAPH_SUCCESS {
					return igraphError("enumerate weighted cliques", int(code))
				}
				return nil
			},
			listSlices: func(list *intVectorList) ([][]int, error) { return list.slices() },
		},
	)
}

// WeightedCliqueNumber returns the maximum total positive integer vertex
// weight among all cliques. weights is borrowed and copied for the call.
//
//igraph:bind igraph_weighted_clique_number
func (g *Graph) WeightedCliqueNumber(weights []int) (int, error) {
	if g == nil {
		return 0, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return 0, ErrClosed
	}
	weightVector, err := newCliqueWeights(weights, int(C.igraph_vcount(&g.graph)))
	if err != nil {
		return 0, err
	}
	defer weightVector.close()
	simple, err := newSimpleCliqueGraph(&g.graph, "calculate weighted clique number")
	if err != nil {
		return 0, err
	}
	defer C.igraph_destroy(simple)
	return weightedCliqueNumberOnSimple(simple, weightVector)
}

// MaximumWeightCliques returns bounded cliques attaining WeightedCliqueNumber.
// It composes the scalar query with WeightedCliques and never exposes the
// upstream unbounded largest-weighted-clique collector.
func (g *Graph) MaximumWeightCliques(weights []int, maxResults int) (VertexSetEnumeration, error) {
	result := VertexSetEnumeration{Sets: make([][]int, 0)}
	if err := (VertexSetEnumerationOptions{MaxResults: maxResults}).validate(); err != nil {
		return result, err
	}
	if g == nil {
		return result, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return result, ErrClosed
	}
	weightVector, err := newCliqueWeights(weights, int(C.igraph_vcount(&g.graph)))
	if err != nil {
		return result, err
	}
	defer weightVector.close()
	simple, err := newSimpleCliqueGraph(&g.graph, "enumerate maximum-weight cliques")
	if err != nil {
		return result, err
	}
	defer C.igraph_destroy(simple)
	maximum, err := weightedCliqueNumberOnSimple(simple, weightVector)
	if err != nil {
		return result, err
	}
	if maximum == 0 {
		return result, nil
	}
	return enumerateWeightedCliquesOnSimple(simple, weightVector, WeightedCliqueOptions{
		Range: WeightRange{Minimum: &maximum, Maximum: &maximum}, MaxResults: maxResults,
	})
}

func weightedCliqueNumberOnSimple(simple *C.igraph_t, weights *realVector) (int, error) {
	var value C.igraph_real_t
	if code := C.go_igraph_weighted_clique_number(simple, &weights.value, &value); code != C.IGRAPH_SUCCESS {
		return 0, igraphError("calculate weighted clique number", int(code))
	}
	return checkedCliqueWeight(float64(value), "weighted clique number")
}

func newCliqueWeights(weights []int, vertexCount int) (*realVector, error) {
	if len(weights) != vertexCount {
		return nil, fmt.Errorf("igraph: vertex weights length is %d; want %d", len(weights), vertexCount)
	}
	values := make([]float64, len(weights))
	var total uint64
	for index, weight := range weights {
		if err := validateCliqueWeight(weight, fmt.Sprintf("vertex weight at index %d", index)); err != nil {
			return nil, err
		}
		if uint64(weight) > (uint64(1)<<53)-total {
			return nil, fmt.Errorf("igraph: sum of vertex weights exceeds the exact C-igraph integer range")
		}
		total += uint64(weight)
		values[index] = float64(weight)
	}
	return newRealVector(values)
}

func validateCliqueWeight(weight int, description string) error {
	if weight <= 0 {
		return fmt.Errorf("igraph: %s must be positive: %d", description, weight)
	}
	if int(float64(weight)) != weight {
		return fmt.Errorf("igraph: %s cannot be represented exactly by C-igraph: %d", description, weight)
	}
	return nil
}

func checkedCliqueWeight(value float64, description string) (int, error) {
	if math.IsNaN(value) || math.IsInf(value, 0) || value < 0 || math.Trunc(value) != value {
		return 0, fmt.Errorf("igraph: %s is not a non-negative integer: %g", description, value)
	}
	converted := int(value)
	if converted < 0 || float64(converted) != value {
		return 0, fmt.Errorf("igraph: %s is out of Go int range: %g", description, value)
	}
	return converted, nil
}
