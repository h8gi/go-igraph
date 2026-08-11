package igraph

/*
#cgo pkg-config: igraph
#include <igraph.h>
#include "isomorphism_cgo.h"
*/
import "C"

import "fmt"

// VF2IsomorphismOptions configures color-aware equal-size matching. Nil color
// slices disable that color dimension. A color dimension must be provided for
// both operands or neither operand.
type VF2IsomorphismOptions struct {
	SourceVertexColors []int
	TargetVertexColors []int
	SourceEdgeColors   []int
	TargetEdgeColors   []int
}

// VF2SubgraphOptions configures color-aware pattern matching. Nil color slices
// disable that color dimension. A color dimension must be provided for both
// operands or neither operand.
type VF2SubgraphOptions struct {
	TargetVertexColors  []int
	PatternVertexColors []int
	TargetEdgeColors    []int
	PatternEdgeColors   []int
}

// IsomorphismResult contains the first equal-size VF2 match. SourceToTarget is
// indexed by source vertex and TargetToSource is indexed by target vertex.
// Both mappings are non-nil and empty when Found is false.
type IsomorphismResult struct {
	Found          bool
	SourceToTarget []int
	TargetToSource []int
}

// SubgraphIsomorphismResult contains the first VF2 pattern match.
// PatternToTarget is indexed by pattern vertex. TargetToPattern is indexed by
// target vertex and contains RemovedID for target vertices outside the match.
// Both mappings are non-nil and empty when Found is false.
type SubgraphIsomorphismResult struct {
	Found           bool
	PatternToTarget []int
	TargetToPattern []int
}

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

// IsomorphicVF2 finds the first color-compatible isomorphism from g (source)
// to target. VF2 accepts directed or undirected simple graphs; both operands
// must have the same directedness. Color slices are borrowed and copied for
// the synchronous call. Returned mappings are Go-owned.
//
//igraph:bind igraph_isomorphic_vf2
func (g *Graph) IsomorphicVF2(target *Graph, options VF2IsomorphismOptions) (IsomorphismResult, error) {
	var result IsomorphismResult
	err := withLockedGraphs([]*Graph{g, target}, func(graphs []*C.igraph_t) error {
		colors, err := prepareVF2Colors(
			graphs[0], graphs[1], options.SourceVertexColors,
			options.TargetVertexColors, options.SourceEdgeColors,
			options.TargetEdgeColors,
		)
		if err != nil {
			return err
		}
		defer colors.close()
		return runVF2(graphs[0], graphs[1], colors, false, &result)
	})
	return result, err
}

// ContainsSubgraphIsomorphicToVF2 finds the first color-compatible mapping of
// pattern into g (the target). VF2 accepts directed or undirected simple
// graphs; both operands must have the same directedness. Color slices are
// borrowed and copied for the synchronous call. Returned mappings are
// Go-owned.
//
//igraph:bind igraph_subisomorphic_vf2
func (g *Graph) ContainsSubgraphIsomorphicToVF2(pattern *Graph, options VF2SubgraphOptions) (SubgraphIsomorphismResult, error) {
	var public SubgraphIsomorphismResult
	err := withLockedGraphs([]*Graph{g, pattern}, func(graphs []*C.igraph_t) error {
		colors, err := prepareVF2Colors(
			graphs[0], graphs[1], options.TargetVertexColors,
			options.PatternVertexColors, options.TargetEdgeColors,
			options.PatternEdgeColors,
		)
		if err != nil {
			return err
		}
		defer colors.close()
		var internal IsomorphismResult
		if err := runVF2(graphs[0], graphs[1], colors, true, &internal); err != nil {
			return err
		}
		public.Found = internal.Found
		public.PatternToTarget = internal.TargetToSource
		public.TargetToPattern = internal.SourceToTarget
		return nil
	})
	return public, err
}

type vf2Colors struct {
	firstVertex  *intVector
	secondVertex *intVector
	firstEdge    *intVector
	secondEdge   *intVector
}

func prepareVF2Colors(first, second *C.igraph_t, firstVertices, secondVertices, firstEdges, secondEdges []int) (*vf2Colors, error) {
	if (firstVertices == nil) != (secondVertices == nil) {
		return nil, fmt.Errorf("igraph: VF2 vertex colors must be provided for both operands or neither")
	}
	if (firstEdges == nil) != (secondEdges == nil) {
		return nil, fmt.Errorf("igraph: VF2 edge colors must be provided for both operands or neither")
	}
	firstSimple, err := vf2GraphIsSimple(first)
	if err != nil {
		return nil, err
	}
	secondSimple, err := vf2GraphIsSimple(second)
	if err != nil {
		return nil, err
	}
	if !firstSimple || !secondSimple {
		return nil, fmt.Errorf("igraph: VF2 requires simple graphs")
	}
	if firstVertices != nil && (len(firstVertices) != int(C.igraph_vcount(first)) || len(secondVertices) != int(C.igraph_vcount(second))) {
		return nil, fmt.Errorf("igraph: VF2 vertex color lengths must match their graph vertex counts")
	}
	if firstEdges != nil && (len(firstEdges) != int(C.igraph_ecount(first)) || len(secondEdges) != int(C.igraph_ecount(second))) {
		return nil, fmt.Errorf("igraph: VF2 edge color lengths must match their graph edge counts")
	}
	colors := &vf2Colors{}
	values := [4][]int{firstVertices, secondVertices, firstEdges, secondEdges}
	destinations := [4]**intVector{
		&colors.firstVertex, &colors.secondVertex, &colors.firstEdge, &colors.secondEdge,
	}
	for index, value := range values {
		if value == nil {
			continue
		}
		if *destinations[index], err = newIntVector(value); err != nil {
			colors.close()
			return nil, err
		}
	}
	return colors, nil
}

//igraph:internal igraph_is_simple
func vf2GraphIsSimple(graph *C.igraph_t) (bool, error) {
	var result C.igraph_bool_t
	if code := C.go_igraph_is_simple(graph, &result); code != C.IGRAPH_SUCCESS {
		return false, igraphError("check VF2 graph shape", int(code))
	}
	return result != booltoint(false), nil
}

func (colors *vf2Colors) close() {
	if colors == nil {
		return
	}
	for _, vector := range []*intVector{
		colors.firstVertex, colors.secondVertex, colors.firstEdge, colors.secondEdge,
	} {
		if vector != nil {
			vector.close()
		}
	}
}

func (colors *vf2Colors) pointers() (*C.igraph_vector_int_t, *C.igraph_vector_int_t, *C.igraph_vector_int_t, *C.igraph_vector_int_t) {
	var firstVertex, secondVertex, firstEdge, secondEdge *C.igraph_vector_int_t
	if colors.firstVertex != nil {
		firstVertex = &colors.firstVertex.value
	}
	if colors.secondVertex != nil {
		secondVertex = &colors.secondVertex.value
	}
	if colors.firstEdge != nil {
		firstEdge = &colors.firstEdge.value
	}
	if colors.secondEdge != nil {
		secondEdge = &colors.secondEdge.value
	}
	return firstVertex, secondVertex, firstEdge, secondEdge
}

func runVF2(first, second *C.igraph_t, colors *vf2Colors, subgraph bool, result *IsomorphismResult) error {
	firstMap, err := newIntVector(nil)
	if err != nil {
		return err
	}
	defer firstMap.close()
	secondMap, err := newIntVector(nil)
	if err != nil {
		return err
	}
	defer secondMap.close()
	firstVertex, secondVertex, firstEdge, secondEdge := colors.pointers()
	var found C.igraph_bool_t
	var code C.igraph_error_t
	if subgraph {
		code = C.go_igraph_subisomorphic_vf2(first, second, firstVertex, secondVertex, firstEdge, secondEdge, &found, &firstMap.value, &secondMap.value)
	} else {
		code = C.go_igraph_isomorphic_vf2(first, second, firstVertex, secondVertex, firstEdge, secondEdge, &found, &firstMap.value, &secondMap.value)
	}
	if code != C.IGRAPH_SUCCESS {
		return igraphError("run VF2 isomorphism", int(code))
	}
	result.Found = found != booltoint(false)
	result.SourceToTarget, err = firstMap.slice()
	if err != nil {
		return err
	}
	result.TargetToSource, err = secondMap.slice()
	return err
}
