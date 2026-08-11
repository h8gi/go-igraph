package igraph

/*
#cgo pkg-config: igraph
#include <igraph.h>
#include "isomorphism_cgo.h"
*/
import "C"

// Isomorphic reports whether g and other are isomorphic. The two operands must
// have the same directedness. General algorithm selection is intentionally
// internal; loops and parallel edges are supported according to igraph's
// general isomorphism dispatcher.
//
// Both graphs are borrowed for this synchronous call. The returned Boolean is
// Go-owned.
//
//igraph:bind igraph_isomorphic
func (g *Graph) Isomorphic(other *Graph) (bool, error) {
	var result C.igraph_bool_t
	err := withLockedGraphs([]*Graph{g, other}, func(graphs []*C.igraph_t) error {
		code := C.go_igraph_isomorphic(graphs[0], graphs[1], &result)
		if code != C.IGRAPH_SUCCESS {
			return igraphError("check graph isomorphism", int(code))
		}
		return nil
	})
	return result != booltoint(false), err
}

// ContainsSubgraphIsomorphicTo reports whether g, the target graph, contains a
// subgraph isomorphic to pattern. The operands must have the same directedness.
// The general dispatcher is intended for simple graphs; use the specialized
// matching APIs for explicit multigraph behavior.
//
// Both graphs are borrowed for this synchronous call. The returned Boolean is
// Go-owned.
//
//igraph:bind igraph_subisomorphic
func (g *Graph) ContainsSubgraphIsomorphicTo(pattern *Graph) (bool, error) {
	var result C.igraph_bool_t
	err := withLockedGraphs([]*Graph{g, pattern}, func(graphs []*C.igraph_t) error {
		code := C.go_igraph_subisomorphic(graphs[0], graphs[1], &result)
		if code != C.IGRAPH_SUCCESS {
			return igraphError("check subgraph isomorphism", int(code))
		}
		return nil
	})
	return result != booltoint(false), err
}
