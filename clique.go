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
