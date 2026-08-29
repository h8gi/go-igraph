package igraph

// #cgo pkg-config: igraph
// #include <igraph.h>
// #include "algorithm_cgo.h"
// #include "operators_cgo.h"
import "C"

import (
	"fmt"
	"sort"
)

// GraphTransformationResult describes ID provenance from a successful in-place
// graph transformation. Vertex mappings are always exact; operation-specific
// documentation states whether they are identity or many-to-one. When
// EdgeMappingAvailable is true, Mapping.Edges maps source edge IDs to result
// edge IDs and may be many-to-one. When it is false, Mapping.Edges contains
// non-nil empty slices and must not be interpreted as an empty source graph.
// Endpoint-equivalent parallel edges use ascending source and result edge ID
// order as a deterministic structural convention for one-to-one and reciprocal
// pairing; this does not claim attribute provenance. All slices are Go-owned
// and remain valid after the graph is closed.
type GraphTransformationResult struct {
	Mapping              GraphIDMapping
	EdgeMappingAvailable bool
}

// SimplifyOptions selects which non-simple edge structures to remove.
// RemoveParallel merges each group of parallel edges into one edge, while
// RemoveLoops removes every self-loop. The zero value requests no changes.
type SimplifyOptions struct {
	RemoveParallel bool
	RemoveLoops    bool
	// EdgeAttributes is required when RemoveParallel may merge attributed
	// edges. It is borrowed only for the call. Loop-only simplification
	// preserves surviving edge attributes without requiring a policy.
	EdgeAttributes *AttributeCombinationPolicy
}

// ConnectNeighborhoodInPlace atomically adds an edge from each vertex to every
// vertex reachable within at most order steps. order must be non-negative;
// zero is an explicit no-op. On directed graphs, DirectionOut adds source-to-
// reachable edges and DirectionIn adds predecessor-to-source edges.
// DirectionAll ignores direction for reachability and adds one directed edge
// for each missing unordered connection in the pinned upstream ordering; it
// does not add a reciprocal pair. Undirected graphs always use symmetric
// reachability and add at most one undirected edge per vertex pair.
// Existing connections are not duplicated; existing loops and parallel edges
// are retained.
//
// Vertex IDs are unchanged. Mapping.Edges maps every pre-existing edge to the
// result; newly added edges have RemovedID in NewToOld. Graph, vertex, and
// existing edge attributes are preserved. New edges receive igraph's typed
// missing defaults: NaN for numeric, the empty string for string, and false for
// Boolean attributes. The operation uses clone-and-swap, so validation,
// initialization, upstream, or mapping failure leaves g unchanged.
//
//igraph:bind igraph_connect_neighborhood
func (g *Graph) ConnectNeighborhoodInPlace(order int, direction DirectionMode) (GraphTransformationResult, error) {
	return g.connectNeighborhoodInPlace(order, direction, nil)
}

func (g *Graph) connectNeighborhoodInPlace(order int, direction DirectionMode, hook graphTransformationFailureHook) (GraphTransformationResult, error) {
	if g == nil {
		return GraphTransformationResult{}, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return GraphTransformationResult{}, ErrClosed
	}
	if order < 0 {
		return GraphTransformationResult{}, fmt.Errorf("igraph: neighborhood order must be non-negative: %d", order)
	}
	cDirection, err := direction.cValue()
	if err != nil {
		return GraphTransformationResult{}, err
	}
	if order == 0 {
		return identityGraphTransformationResult(int(C.igraph_vcount(&g.graph)), int(C.igraph_ecount(&g.graph)))
	}
	return g.applyAtomicGraphTransformation("connect graph neighborhood", hook, func(replacement *C.igraph_t) C.igraph_error_t {
		return C.go_igraph_connect_neighborhood(replacement, C.igraph_int_t(order), cDirection)
	}, func(before, after []Edge, directed bool) (IDMapping, bool, error) {
		mapping, err := prefixEdgeMapping(before, after, directed)
		return mapping, true, err
	})
}

// ContractVerticesInPlace atomically merges vertices according to mapping.
// mapping is indexed by the current vertex ID and must have exactly one
// non-negative target label per vertex. Labels may contain gaps: distinct
// labels are ranked in ascending order to produce consecutive result IDs.
// Mapping.Vertices is the exact old-to-normalized-result mapping and chooses
// the lowest source ID as each NewToOld representative. All edges retain their
// IDs and attributes; contraction may turn them into loops or parallel edges.
// Directedness, graph attributes, and edge attributes are preserved.
//
// When contraction merges vertices that have attributes, vertexAttributes is
// required and controls each many-to-one value. The mapping and policy are
// borrowed only for the synchronous call. An identity mapping, including one
// produced by normalizing gapped labels, is a validated no-op. The operation
// uses clone-and-swap, so validation, allocation, upstream, attribute
// combination, or provenance failure leaves g unchanged.
//
//igraph:bind igraph_contract_vertices
func (g *Graph) ContractVerticesInPlace(mapping []int, vertexAttributes *AttributeCombinationPolicy) (GraphTransformationResult, error) {
	return g.contractVerticesInPlace(mapping, vertexAttributes, nil)
}

func (g *Graph) contractVerticesInPlace(mapping []int, vertexAttributes *AttributeCombinationPolicy, hook graphTransformationFailureHook) (GraphTransformationResult, error) {
	if g == nil {
		return GraphTransformationResult{}, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return GraphTransformationResult{}, ErrClosed
	}
	vertexCount := int(C.igraph_vcount(&g.graph))
	normalized, resultVertexCount, identity, err := normalizeContractionMapping(mapping, vertexCount)
	if err != nil {
		return GraphTransformationResult{}, err
	}
	metadata, err := attributeMetadataLocked(&g.graph, AttributeVertex)
	if err != nil {
		return GraphTransformationResult{}, err
	}
	if identity {
		if vertexAttributes != nil {
			if err := validateCombinationPolicy(vertexAttributes, metadata); err != nil {
				return GraphTransformationResult{}, err
			}
		}
		return contractionResult(normalized, resultVertexCount, int(C.igraph_ecount(&g.graph)))
	}
	combination, err := newAttributeCombination(vertexAttributes, metadata)
	if err != nil {
		return GraphTransformationResult{}, err
	}
	defer combination.close()
	cMapping, err := newIntVector(normalized)
	if err != nil {
		return GraphTransformationResult{}, err
	}
	defer cMapping.close()
	result, err := contractionResult(normalized, resultVertexCount, int(C.igraph_ecount(&g.graph)))
	if err != nil {
		return GraphTransformationResult{}, err
	}
	if err := runGraphTransformationHook(hook, graphTransformationAtClone); err != nil {
		return GraphTransformationResult{}, err
	}
	var replacement C.igraph_t
	if code := C.go_igraph_copy(&replacement, &g.graph); code != C.IGRAPH_SUCCESS {
		return GraphTransformationResult{}, igraphError("contract vertices copy", int(code))
	}
	committed := false
	defer func() {
		if !committed {
			C.igraph_destroy(&replacement)
		}
	}()
	if err := runGraphTransformationHook(hook, graphTransformationAtTransform); err != nil {
		return GraphTransformationResult{}, err
	}
	var cCombination *C.igraph_attribute_combination_t
	if combination != nil {
		cCombination = &combination.value
	}
	if code := C.go_igraph_contract_vertices(&replacement, &cMapping.value, cCombination); code != C.IGRAPH_SUCCESS {
		return GraphTransformationResult{}, igraphError("contract vertices", int(code))
	}
	if err := runGraphTransformationHook(hook, graphTransformationAfterTransform); err != nil {
		return GraphTransformationResult{}, err
	}
	if got := int(C.igraph_vcount(&replacement)); got != resultVertexCount {
		return GraphTransformationResult{}, fmt.Errorf("igraph: contracted vertex count is %d, want %d", got, resultVertexCount)
	}
	if got, want := int(C.igraph_ecount(&replacement)), len(result.Mapping.Edges.OldToNew); got != want {
		return GraphTransformationResult{}, fmt.Errorf("igraph: contracted edge count is %d, want %d", got, want)
	}
	replaceInitializedGraph(&g.graph, &replacement)
	committed = true
	return result, nil
}

func normalizeContractionMapping(mapping []int, vertexCount int) ([]int, int, bool, error) {
	if len(mapping) != vertexCount {
		return nil, 0, false, fmt.Errorf("igraph: contraction mapping length is %d, want %d", len(mapping), vertexCount)
	}
	labels := make([]int, 0, vertexCount)
	seen := make(map[int]struct{}, vertexCount)
	for vertex, label := range mapping {
		if label < 0 {
			return nil, 0, false, fmt.Errorf("igraph: contraction target for vertex %d must be non-negative: %d", vertex, label)
		}
		if _, ok := seen[label]; !ok {
			seen[label] = struct{}{}
			labels = append(labels, label)
		}
	}
	sort.Ints(labels)
	ranks := make(map[int]int, len(labels))
	for rank, label := range labels {
		ranks[label] = rank
	}
	normalized := make([]int, vertexCount)
	identity := len(labels) == vertexCount
	for vertex, label := range mapping {
		normalized[vertex] = ranks[label]
		identity = identity && normalized[vertex] == vertex
	}
	return normalized, len(labels), identity, nil
}

func contractionResult(vertices []int, resultVertexCount, edgeCount int) (GraphTransformationResult, error) {
	vertexMapping, err := newIDMapping(vertices, resultVertexCount)
	if err != nil {
		return GraphTransformationResult{}, err
	}
	edgeMapping, err := identityIDMapping(edgeCount)
	if err != nil {
		return GraphTransformationResult{}, err
	}
	return GraphTransformationResult{
		Mapping:              GraphIDMapping{Vertices: vertexMapping, Edges: edgeMapping},
		EdgeMappingAvailable: true,
	}, nil
}

// ReverseEdgesInPlace atomically reverses selected edges of a directed graph.
// The selector is validated and materialized before cloning. Duplicate IDs or
// endpoint pairs select an edge once, so they never cancel by double reversal.
// An empty selector is a validated no-op. Vertex and edge IDs, ordering, loops,
// parallel multiplicity, and all attributes are preserved; both returned
// mappings are exact non-nil identities. Undirected graphs are rejected.
//
//igraph:bind igraph_reverse_edges
func (g *Graph) ReverseEdgesInPlace(selector EdgeSelector) (GraphTransformationResult, error) {
	return g.reverseEdgesInPlace(selector, nil)
}

func (g *Graph) reverseEdgesInPlace(selector EdgeSelector, hook graphTransformationFailureHook) (GraphTransformationResult, error) {
	if g == nil {
		return GraphTransformationResult{}, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return GraphTransformationResult{}, ErrClosed
	}
	if C.igraph_is_directed(&g.graph) == booltoint(false) {
		return GraphTransformationResult{}, fmt.Errorf("igraph: edge reversal requires a directed graph")
	}
	selected, err := materializeSelectedEdgeIDs(&g.graph, selector)
	if err != nil {
		return GraphTransformationResult{}, err
	}
	selected = uniqueSelectedEdgeIDs(selected)
	if len(selected) == 0 {
		return identityGraphTransformationResult(int(C.igraph_vcount(&g.graph)), int(C.igraph_ecount(&g.graph)))
	}
	selectedSelector, err := EdgeIDs(selected...)
	if err != nil {
		return GraphTransformationResult{}, err
	}
	cSelector, err := newCEdgeSelector(selectedSelector)
	if err != nil {
		return GraphTransformationResult{}, err
	}
	defer cSelector.close()
	selectedSet := make(map[int]struct{}, len(selected))
	for _, edgeID := range selected {
		selectedSet[edgeID] = struct{}{}
	}
	return g.applyAtomicGraphTransformation("reverse graph edges", hook, func(replacement *C.igraph_t) C.igraph_error_t {
		return C.go_igraph_reverse_edges(replacement, cSelector.value)
	}, func(before, after []Edge, _ bool) (IDMapping, bool, error) {
		if len(before) != len(after) {
			return IDMapping{}, false, fmt.Errorf("edge count changed from %d to %d", len(before), len(after))
		}
		for edgeID, edge := range before {
			if _, reverse := selectedSet[edgeID]; reverse {
				edge.From, edge.To = edge.To, edge.From
			}
			if edge != after[edgeID] {
				return IDMapping{}, false, fmt.Errorf("edge %d is %v after reversal, want %v", edgeID, after[edgeID], edge)
			}
		}
		mapping, err := identityIDMapping(len(before))
		return mapping, true, err
	})
}

func uniqueSelectedEdgeIDs(ids []int) []int {
	result := make([]int, 0, len(ids))
	seen := make(map[int]struct{}, len(ids))
	for _, id := range ids {
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}
	return result
}

// DirectedConversionMode controls how an undirected edge becomes directed.
type DirectedConversionMode uint8

const (
	// DirectedConversionArbitrary gives every edge one upstream-defined
	// orientation and preserves the edge count.
	DirectedConversionArbitrary DirectedConversionMode = iota
	// DirectedConversionMutual creates two oppositely directed edges from each
	// undirected edge. A loop therefore produces two directed loops.
	DirectedConversionMutual
	// DirectedConversionAcyclic directs each non-loop edge from the lower vertex
	// ID to the higher vertex ID. Existing loops remain loops.
	DirectedConversionAcyclic
	// DirectedConversionRandom gives every edge one random orientation and
	// preserves the edge count.
	DirectedConversionRandom
)

func (mode DirectedConversionMode) cValue() (C.igraph_to_directed_t, error) {
	switch mode {
	case DirectedConversionArbitrary:
		return C.IGRAPH_TO_DIRECTED_ARBITRARY, nil
	case DirectedConversionMutual:
		return C.IGRAPH_TO_DIRECTED_MUTUAL, nil
	case DirectedConversionAcyclic:
		return C.IGRAPH_TO_DIRECTED_ACYCLIC, nil
	case DirectedConversionRandom:
		return C.IGRAPH_TO_DIRECTED_RANDOM, nil
	default:
		return 0, fmt.Errorf("igraph: invalid directed conversion mode: %d", mode)
	}
}

// UndirectedConversionMode controls how directed edges become undirected.
type UndirectedConversionMode uint8

const (
	// UndirectedConversionEach removes direction from every edge independently,
	// preserving the edge count and possibly creating parallel edges.
	UndirectedConversionEach UndirectedConversionMode = iota
	// UndirectedConversionCollapse creates one undirected edge for every
	// connected unordered vertex pair, collapsing all parallel directions.
	UndirectedConversionCollapse
	// UndirectedConversionMutual creates one undirected edge for each matched
	// pair of opposite directed edges. Unmatched non-loop edges are removed;
	// loops are retained unconditionally.
	UndirectedConversionMutual
)

func (mode UndirectedConversionMode) cValue() (C.igraph_to_undirected_t, error) {
	switch mode {
	case UndirectedConversionEach:
		return C.IGRAPH_TO_UNDIRECTED_EACH, nil
	case UndirectedConversionCollapse:
		return C.IGRAPH_TO_UNDIRECTED_COLLAPSE, nil
	case UndirectedConversionMutual:
		return C.IGRAPH_TO_UNDIRECTED_MUTUAL, nil
	default:
		return 0, fmt.Errorf("igraph: invalid undirected conversion mode: %d", mode)
	}
}

// SimplifyInPlace atomically removes the requested parallel edges and loops.
// Vertices are unchanged, but igraph may reorder edges even when the input was
// already simple. A zero options value is a no-op.
//
// The result contains an exact structural edge mapping; parallel sources may
// map to the same result edge and NewToOld selects their lowest source ID.
// Graph and vertex attributes are preserved. Surviving edge attributes are
// preserved when only loops are removed. When parallel edges may be merged,
// EdgeAttributes must explicitly select how every edge attribute is combined;
// validation and upstream failures leave the receiver unchanged.
//
//igraph:bind igraph_simplify
func (g *Graph) SimplifyInPlace(options SimplifyOptions) (GraphTransformationResult, error) {
	return g.simplifyInPlace(options, nil)
}

func (g *Graph) simplifyInPlace(options SimplifyOptions, hook graphTransformationFailureHook) (GraphTransformationResult, error) {
	if g == nil {
		return GraphTransformationResult{}, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return GraphTransformationResult{}, ErrClosed
	}
	if !options.RemoveParallel && !options.RemoveLoops {
		return identityGraphTransformationResult(int(C.igraph_vcount(&g.graph)), int(C.igraph_ecount(&g.graph)))
	}
	edgeAttributes, err := attributeMetadataLocked(&g.graph, AttributeEdge)
	if err != nil {
		return GraphTransformationResult{}, err
	}
	policy := options.EdgeAttributes
	if !options.RemoveParallel && policy == nil {
		policy = &AttributeCombinationPolicy{Default: AttributeCombineFirst}
	}
	combination, err := newAttributeCombination(policy, edgeAttributes)
	if err != nil {
		return GraphTransformationResult{}, err
	}
	defer combination.close()
	return g.applyAtomicGraphTransformation("simplify graph", hook, func(replacement *C.igraph_t) C.igraph_error_t {
		var cCombination *C.igraph_attribute_combination_t
		if combination != nil {
			cCombination = &combination.value
		}
		return C.go_igraph_simplify(
			replacement, booltoint(options.RemoveParallel), booltoint(options.RemoveLoops), cCombination,
		)
	}, func(before, after []Edge, directed bool) (IDMapping, bool, error) {
		return simplifyEdgeMapping(before, after, directed, options)
	})
}

// ConvertToDirectedInPlace atomically converts an undirected graph according
// to mode. A graph that is already directed is unchanged after mode validation.
//
// The result contains an edge mapping for arbitrary, acyclic, and random
// conversion. When mutual conversion actually duplicates edges it is
// one-to-many, which IDMapping cannot represent, so EdgeMappingAvailable is
// false and its edge slices are non-nil and empty. No-op and empty conversions
// have available identity mappings. Vertex IDs remain unchanged. Graph and
// vertex attributes are preserved. Existing edge attributes are
// copied to each derived edge and remain independently owned.
//
//igraph:bind igraph_to_directed
func (g *Graph) ConvertToDirectedInPlace(mode DirectedConversionMode) (GraphTransformationResult, error) {
	return g.convertToDirectedInPlace(mode, nil)
}

func (g *Graph) convertToDirectedInPlace(mode DirectedConversionMode, hook graphTransformationFailureHook) (GraphTransformationResult, error) {
	if g == nil {
		return GraphTransformationResult{}, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return GraphTransformationResult{}, ErrClosed
	}
	cMode, err := mode.cValue()
	if err != nil {
		return GraphTransformationResult{}, err
	}
	if C.igraph_is_directed(&g.graph) != booltoint(false) {
		return identityGraphTransformationResult(int(C.igraph_vcount(&g.graph)), int(C.igraph_ecount(&g.graph)))
	}
	return g.applyAtomicGraphTransformation("convert graph to directed", hook, func(replacement *C.igraph_t) C.igraph_error_t {
		return C.go_igraph_to_directed(replacement, cMode)
	}, func(before, after []Edge, _ bool) (IDMapping, bool, error) {
		if mode == DirectedConversionMutual {
			if len(before) == 0 {
				mapping, err := identityIDMapping(0)
				return mapping, true, err
			}
			mapping, err := newIDMapping([]int{}, 0)
			return mapping, false, err
		}
		mapping, err := identityEndpointEdgeMapping(before, after, false)
		return mapping, true, err
	})
}

// ConvertToUndirectedInPlace atomically converts a directed graph according to
// mode. A graph that is already undirected is unchanged after mode validation.
//
// Each, collapse, and mutual return structural edge provenance. For equivalent
// parallel edges, reciprocal pairing follows ascending source and result edge
// ID order; it is an ID convention, not attribute lineage. Vertex IDs remain
// unchanged. Each preserves edge attributes without a policy. Collapse and
// mutual require an explicit policy when edge attributes exist; the policy is
// borrowed only for the call. Every successful result owns its attributes.
//
//igraph:bind igraph_to_undirected
func (g *Graph) ConvertToUndirectedInPlace(mode UndirectedConversionMode, edgeAttributes *AttributeCombinationPolicy) (GraphTransformationResult, error) {
	return g.convertToUndirectedInPlace(mode, edgeAttributes, nil)
}

func (g *Graph) convertToUndirectedInPlace(mode UndirectedConversionMode, edgeAttributePolicy *AttributeCombinationPolicy, hook graphTransformationFailureHook) (GraphTransformationResult, error) {
	if g == nil {
		return GraphTransformationResult{}, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return GraphTransformationResult{}, ErrClosed
	}
	cMode, err := mode.cValue()
	if err != nil {
		return GraphTransformationResult{}, err
	}
	if C.igraph_is_directed(&g.graph) == booltoint(false) {
		return identityGraphTransformationResult(int(C.igraph_vcount(&g.graph)), int(C.igraph_ecount(&g.graph)))
	}
	edgeAttributes, err := attributeMetadataLocked(&g.graph, AttributeEdge)
	if err != nil {
		return GraphTransformationResult{}, err
	}
	if mode != UndirectedConversionEach && edgeAttributePolicy == nil && len(edgeAttributes) != 0 {
		return GraphTransformationResult{}, fmt.Errorf("igraph: an explicit edge attribute combination policy is required")
	}
	var combination *attributeCombination
	if edgeAttributePolicy != nil {
		combination, err = newAttributeCombination(edgeAttributePolicy, edgeAttributes)
		if err != nil {
			return GraphTransformationResult{}, err
		}
	}
	defer combination.close()
	return g.applyAtomicGraphTransformation("convert graph to undirected", hook, func(replacement *C.igraph_t) C.igraph_error_t {
		var cCombination *C.igraph_attribute_combination_t
		if combination != nil {
			cCombination = &combination.value
		}
		return C.go_igraph_to_undirected(replacement, cMode, cCombination)
	}, func(before, after []Edge, _ bool) (IDMapping, bool, error) {
		var mapping IDMapping
		var err error
		switch mode {
		case UndirectedConversionEach:
			mapping, err = identityEndpointEdgeMapping(before, after, false)
		case UndirectedConversionCollapse:
			mapping, err = manyToOneEdgeMapping(before, after, false, nil)
		case UndirectedConversionMutual:
			mapping, err = mutualEdgeMapping(before, after)
		}
		return mapping, true, err
	})
}

func (g *Graph) applyAtomicGraphTransformation(
	action string,
	hook graphTransformationFailureHook,
	transform func(*C.igraph_t) C.igraph_error_t,
	buildMapping func([]Edge, []Edge, bool) (IDMapping, bool, error),
) (GraphTransformationResult, error) {
	before, err := edgeSlice(&g.graph)
	if err != nil {
		return GraphTransformationResult{}, fmt.Errorf("igraph: snapshot edges before %s: %w", action, err)
	}
	vertices, err := identityIDMapping(int(C.igraph_vcount(&g.graph)))
	if err != nil {
		return GraphTransformationResult{}, err
	}
	directed := C.igraph_is_directed(&g.graph) != booltoint(false)
	if err := runGraphTransformationHook(hook, graphTransformationAtClone); err != nil {
		return GraphTransformationResult{}, err
	}
	var replacement C.igraph_t
	if code := C.go_igraph_copy(&replacement, &g.graph); code != C.IGRAPH_SUCCESS {
		return GraphTransformationResult{}, igraphError(action+" copy", int(code))
	}
	committed := false
	defer func() {
		if !committed {
			C.igraph_destroy(&replacement)
		}
	}()
	if err := runGraphTransformationHook(hook, graphTransformationAtTransform); err != nil {
		return GraphTransformationResult{}, err
	}
	if code := transform(&replacement); code != C.IGRAPH_SUCCESS {
		return GraphTransformationResult{}, igraphError(action, int(code))
	}
	if err := runGraphTransformationHook(hook, graphTransformationAfterTransform); err != nil {
		return GraphTransformationResult{}, err
	}
	after, err := edgeSlice(&replacement)
	if err != nil {
		return GraphTransformationResult{}, fmt.Errorf("igraph: snapshot edges after %s: %w", action, err)
	}
	edges, available, err := buildMapping(before, after, directed)
	if err != nil {
		return GraphTransformationResult{}, fmt.Errorf("igraph: map edges after %s: %w", action, err)
	}
	replaceInitializedGraph(&g.graph, &replacement)
	committed = true
	return GraphTransformationResult{
		Mapping:              GraphIDMapping{Vertices: vertices, Edges: edges},
		EdgeMappingAvailable: available,
	}, nil
}

type graphTransformationStage uint8

const (
	graphTransformationAtClone graphTransformationStage = iota
	graphTransformationAtTransform
	graphTransformationAfterTransform
)

type graphTransformationFailureHook func(graphTransformationStage) error

func runGraphTransformationHook(hook graphTransformationFailureHook, stage graphTransformationStage) error {
	if hook == nil {
		return nil
	}
	if err := hook(stage); err != nil {
		return fmt.Errorf("igraph: injected graph transformation failure at stage %d: %w", stage, err)
	}
	return nil
}

func identityGraphTransformationResult(vertexCount, edgeCount int) (GraphTransformationResult, error) {
	vertices, err := identityIDMapping(vertexCount)
	if err != nil {
		return GraphTransformationResult{}, err
	}
	edges, err := identityIDMapping(edgeCount)
	if err != nil {
		return GraphTransformationResult{}, err
	}
	return GraphTransformationResult{
		Mapping:              GraphIDMapping{Vertices: vertices, Edges: edges},
		EdgeMappingAvailable: true,
	}, nil
}

type edgeEndpointKey struct {
	first  int
	second int
}

func endpointKey(edge Edge, directed bool) edgeEndpointKey {
	if directed || edge.From <= edge.To {
		return edgeEndpointKey{first: edge.From, second: edge.To}
	}
	return edgeEndpointKey{first: edge.To, second: edge.From}
}

func simplifyEdgeMapping(before, after []Edge, directed bool, options SimplifyOptions) (IDMapping, bool, error) {
	eligible := func(edge Edge) bool { return !options.RemoveLoops || edge.From != edge.To }
	if options.RemoveParallel {
		mapping, err := manyToOneEdgeMapping(before, after, directed, eligible)
		return mapping, true, err
	}
	mapping, err := stableFilteredEdgeMapping(before, after, directed, eligible)
	return mapping, true, err
}

func identityEndpointEdgeMapping(before, after []Edge, directed bool) (IDMapping, error) {
	if len(before) != len(after) {
		return IDMapping{}, fmt.Errorf("edge count changed from %d to %d", len(before), len(after))
	}
	for id := range before {
		if endpointKey(before[id], directed) != endpointKey(after[id], directed) {
			return IDMapping{}, fmt.Errorf("edge %d endpoints changed from %v to %v", id, endpointKey(before[id], directed), endpointKey(after[id], directed))
		}
	}
	return identityIDMapping(len(before))
}

func prefixEdgeMapping(before, after []Edge, directed bool) (IDMapping, error) {
	if len(after) < len(before) {
		return IDMapping{}, fmt.Errorf("edge count decreased from %d to %d", len(before), len(after))
	}
	oldToNew := make([]int, len(before))
	for id := range before {
		if endpointKey(before[id], directed) != endpointKey(after[id], directed) {
			return IDMapping{}, fmt.Errorf("edge %d endpoints changed from %v to %v", id, endpointKey(before[id], directed), endpointKey(after[id], directed))
		}
		oldToNew[id] = id
	}
	return newIDMapping(oldToNew, len(after))
}

func stableFilteredEdgeMapping(
	before, after []Edge,
	directed bool,
	eligible func(Edge) bool,
) (IDMapping, error) {
	oldToNew := make([]int, len(before))
	resultID := 0
	for sourceID, edge := range before {
		oldToNew[sourceID] = RemovedID
		if eligible != nil && !eligible(edge) {
			continue
		}
		if resultID >= len(after) {
			return IDMapping{}, fmt.Errorf("edge %d has no result", sourceID)
		}
		if endpointKey(edge, directed) != endpointKey(after[resultID], directed) {
			return IDMapping{}, fmt.Errorf("edge %d endpoints %v do not match result %d endpoints %v", sourceID, endpointKey(edge, directed), resultID, endpointKey(after[resultID], directed))
		}
		oldToNew[sourceID] = resultID
		resultID++
	}
	if resultID != len(after) {
		return IDMapping{}, fmt.Errorf("mapped %d of %d result edges", resultID, len(after))
	}
	return completeEdgeMapping(oldToNew, len(after))
}

func manyToOneEdgeMapping(
	before, after []Edge,
	directed bool,
	eligible func(Edge) bool,
) (IDMapping, error) {
	resultID := make(map[edgeEndpointKey]int, len(after))
	for id, edge := range after {
		key := endpointKey(edge, directed)
		if _, exists := resultID[key]; exists {
			return IDMapping{}, fmt.Errorf("result contains duplicate endpoints %v", key)
		}
		resultID[key] = id
	}
	oldToNew := make([]int, len(before))
	used := make([]bool, len(after))
	for sourceID, edge := range before {
		oldToNew[sourceID] = RemovedID
		if eligible != nil && !eligible(edge) {
			continue
		}
		key := endpointKey(edge, directed)
		id, exists := resultID[key]
		if !exists {
			return IDMapping{}, fmt.Errorf("edge %d with endpoints %v has no result", sourceID, key)
		}
		oldToNew[sourceID] = id
		used[id] = true
	}
	for id, wasUsed := range used {
		if !wasUsed {
			return IDMapping{}, fmt.Errorf("result edge %d has no source", id)
		}
	}
	return completeEdgeMapping(oldToNew, len(after))
}

func mutualEdgeMapping(before, after []Edge) (IDMapping, error) {
	type sourceGroup struct {
		forward []int
		reverse []int
	}
	groups := make(map[edgeEndpointKey]*sourceGroup)
	for sourceID, edge := range before {
		key := endpointKey(edge, false)
		group := groups[key]
		if group == nil {
			group = &sourceGroup{}
			groups[key] = group
		}
		if edge.From == key.first && edge.To == key.second {
			group.forward = append(group.forward, sourceID)
		} else {
			group.reverse = append(group.reverse, sourceID)
		}
	}
	resultIDs := make(map[edgeEndpointKey][]int)
	for resultID, edge := range after {
		key := endpointKey(edge, false)
		resultIDs[key] = append(resultIDs[key], resultID)
	}
	oldToNew := make([]int, len(before))
	for id := range oldToNew {
		oldToNew[id] = RemovedID
	}
	used := 0
	for key, group := range groups {
		ids := resultIDs[key]
		if key.first == key.second {
			if len(ids) != len(group.forward) {
				return IDMapping{}, fmt.Errorf("loop endpoints %v produced %d edges from %d", key, len(ids), len(group.forward))
			}
			for index, sourceID := range group.forward {
				oldToNew[sourceID] = ids[index]
			}
			used += len(ids)
			continue
		}
		pairs := len(group.forward)
		if len(group.reverse) < pairs {
			pairs = len(group.reverse)
		}
		if len(ids) != pairs {
			return IDMapping{}, fmt.Errorf("reciprocal endpoints %v produced %d edges from %d pairs", key, len(ids), pairs)
		}
		for index := 0; index < pairs; index++ {
			oldToNew[group.forward[index]] = ids[index]
			oldToNew[group.reverse[index]] = ids[index]
		}
		used += len(ids)
	}
	if used != len(after) {
		return IDMapping{}, fmt.Errorf("mapped %d of %d result edges", used, len(after))
	}
	return completeEdgeMapping(oldToNew, len(after))
}

func completeEdgeMapping(oldToNew []int, newCount int) (IDMapping, error) {
	mapping, err := newIDMapping(oldToNew, newCount)
	if err != nil {
		return IDMapping{}, err
	}
	for resultID, sourceID := range mapping.NewToOld {
		if sourceID == RemovedID {
			return IDMapping{}, fmt.Errorf("result edge %d has no source", resultID)
		}
	}
	return mapping, nil
}
