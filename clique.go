package igraph

// #cgo pkg-config: igraph
// #include <igraph.h>
// #include "algorithm_cgo.h"
// #include "clique_cgo.h"
import "C"

import "fmt"

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
	// Pinned igraph's independent-set implementation assumes a simple graph.
	// Normalize a temporary copy so loops and parallel edges consistently have
	// adjacency-only semantics across the clique family.
	var simple C.igraph_t
	if code := C.go_igraph_copy(&simple, &g.graph); code != C.IGRAPH_SUCCESS {
		return 0, igraphError("copy graph for "+operation, int(code))
	}
	defer C.igraph_destroy(&simple)
	if code := C.go_igraph_simplify(&simple, booltoint(true), booltoint(true)); code != C.IGRAPH_SUCCESS {
		return 0, igraphError("simplify graph for "+operation, int(code))
	}
	var result C.igraph_int_t
	var code C.igraph_error_t
	if independent {
		code = C.go_igraph_independence_number(&simple, &result)
	} else {
		code = C.go_igraph_clique_number(&simple, &result)
	}
	if code != C.IGRAPH_SUCCESS {
		return 0, igraphError(operation, int(code))
	}
	return igraphIntToInt(result, operation)
}
