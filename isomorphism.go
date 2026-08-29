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

// VF2IsomorphismEnumerationOptions configures bounded equal-size mapping
// enumeration. MaxMappings must be positive.
type VF2IsomorphismEnumerationOptions struct {
	Colors      VF2IsomorphismOptions
	MaxMappings int
}

// VF2SubgraphEnumerationOptions configures bounded pattern mapping
// enumeration. MaxMappings must be positive.
type VF2SubgraphEnumerationOptions struct {
	Colors      VF2SubgraphOptions
	MaxMappings int
}

// MappingEnumerationResult contains a bounded set of mappings. Mappings is
// always non-nil. Truncated reports that at least one additional mapping was
// found after the limit was reached.
type MappingEnumerationResult struct {
	Mappings  [][]int
	Truncated bool
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

// CountIsomorphismsVF2 counts color-compatible mappings from g (source) to
// target. Color inputs are borrowed and copied for the synchronous call.
//
//igraph:bind igraph_count_isomorphisms_vf2
func (g *Graph) CountIsomorphismsVF2(target *Graph, options VF2IsomorphismOptions) (int, error) {
	return countVF2(g, target, options.SourceVertexColors, options.TargetVertexColors, options.SourceEdgeColors, options.TargetEdgeColors, false)
}

// CountSubgraphIsomorphismsVF2 counts color-compatible mappings of pattern
// into g (the target). Color inputs are borrowed and copied for the
// synchronous call.
//
//igraph:bind igraph_count_subisomorphisms_vf2
func (g *Graph) CountSubgraphIsomorphismsVF2(pattern *Graph, options VF2SubgraphOptions) (int, error) {
	return countVF2(g, pattern, options.TargetVertexColors, options.PatternVertexColors, options.TargetEdgeColors, options.PatternEdgeColors, true)
}

// EnumerateIsomorphismsVF2 returns at most MaxMappings source-to-target
// mappings. Enumeration stops internally after discovering one additional
// mapping, which sets Truncated. Returned slices are independently Go-owned.
func (g *Graph) EnumerateIsomorphismsVF2(target *Graph, options VF2IsomorphismEnumerationOptions) (MappingEnumerationResult, error) {
	return enumerateVF2(g, target, options.Colors.SourceVertexColors, options.Colors.TargetVertexColors, options.Colors.SourceEdgeColors, options.Colors.TargetEdgeColors, options.MaxMappings, false)
}

// EnumerateSubgraphIsomorphismsVF2 returns at most MaxMappings
// pattern-to-target mappings. Enumeration stops internally after discovering
// one additional mapping, which sets Truncated. Returned slices are
// independently Go-owned.
func (g *Graph) EnumerateSubgraphIsomorphismsVF2(pattern *Graph, options VF2SubgraphEnumerationOptions) (MappingEnumerationResult, error) {
	return enumerateVF2(g, pattern, options.Colors.TargetVertexColors, options.Colors.PatternVertexColors, options.Colors.TargetEdgeColors, options.Colors.PatternEdgeColors, options.MaxMappings, true)
}

func countVF2(first, second *Graph, firstVertices, secondVertices, firstEdges, secondEdges []int, subgraph bool) (int, error) {
	var count C.igraph_int_t
	err := withLockedGraphs([]*Graph{first, second}, func(graphs []*C.igraph_t) error {
		colors, err := prepareVF2Colors(graphs[0], graphs[1], firstVertices, secondVertices, firstEdges, secondEdges)
		if err != nil {
			return err
		}
		defer colors.close()
		firstVertex, secondVertex, firstEdge, secondEdge := colors.pointers()
		var code C.igraph_error_t
		if subgraph {
			code = C.go_igraph_count_subisomorphisms_vf2(graphs[0], graphs[1], firstVertex, secondVertex, firstEdge, secondEdge, &count)
		} else {
			code = C.go_igraph_count_isomorphisms_vf2(graphs[0], graphs[1], firstVertex, secondVertex, firstEdge, secondEdge, &count)
		}
		if code != C.IGRAPH_SUCCESS {
			return igraphError("count VF2 isomorphisms", int(code))
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	return igraphIntToInt(count, "VF2 isomorphism count")
}

//igraph:internal igraph_get_isomorphisms_vf2_callback
//igraph:internal igraph_get_subisomorphisms_vf2_callback
func enumerateVF2(first, second *Graph, firstVertices, secondVertices, firstEdges, secondEdges []int, maxMappings int, subgraph bool) (MappingEnumerationResult, error) {
	result := MappingEnumerationResult{Mappings: make([][]int, 0)}
	if maxMappings <= 0 {
		return result, fmt.Errorf("igraph: MaxMappings must be positive: %d", maxMappings)
	}
	cMax, err := intToIgraphInt(maxMappings, "MaxMappings")
	if err != nil {
		return result, err
	}
	err = withLockedGraphs([]*Graph{first, second}, func(graphs []*C.igraph_t) error {
		colors, err := prepareVF2Colors(graphs[0], graphs[1], firstVertices, secondVertices, firstEdges, secondEdges)
		if err != nil {
			return err
		}
		defer colors.close()
		mappings, err := newIntVectorList()
		if err != nil {
			return err
		}
		defer mappings.close()
		firstVertex, secondVertex, firstEdge, secondEdge := colors.pointers()
		var truncated C.igraph_bool_t
		code := C.go_igraph_enumerate_isomorphisms_vf2(
			graphs[0], graphs[1], firstVertex, secondVertex, firstEdge, secondEdge,
			cMax, booltoint(subgraph), &mappings.value, &truncated,
		)
		if code != C.IGRAPH_SUCCESS {
			return igraphError("enumerate VF2 isomorphisms", int(code))
		}
		result.Mappings, err = mappings.slices()
		if err != nil {
			return err
		}
		result.Truncated = truncated != booltoint(false)
		return nil
	})
	return result, err
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
	firstSimple, err := graphIsSimple(first)
	if err != nil {
		return nil, err
	}
	secondSimple, err := graphIsSimple(second)
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
func graphIsSimple(graph *C.igraph_t) (bool, error) {
	var result C.igraph_bool_t
	if code := C.go_igraph_is_simple(graph, &result); code != C.IGRAPH_SUCCESS {
		return false, igraphError("check graph simplicity", int(code))
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
