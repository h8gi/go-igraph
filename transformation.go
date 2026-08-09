package igraph

// #cgo pkg-config: igraph
// #include <igraph.h>
// #include "algorithm_cgo.h"
import "C"

import (
	"fmt"
)

// SimplifyOptions selects which non-simple edge structures to remove.
// RemoveParallel merges each group of parallel edges into one edge, while
// RemoveLoops removes every self-loop. The zero value requests no changes.
type SimplifyOptions struct {
	RemoveParallel bool
	RemoveLoops    bool
}

// DirectedConversionMode controls how an undirected edge becomes directed.
type DirectedConversionMode uint8

const (
	// DirectedConversionArbitrary gives every edge one deterministic upstream-
	// defined orientation and preserves the edge count.
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
// The operation intentionally returns no edge ID mapping. igraph 1.0.1 does
// not report simplification provenance, and merging parallel edges has no
// unambiguous one-to-one inverse. The binding passes no attribute-combination
// policy; edge attributes would be discarded, including on unaffected edges.
// Graph attributes are not currently exposed by this package.
//
//igraph:bind igraph_simplify
func (g *Graph) SimplifyInPlace(options SimplifyOptions) error {
	if g == nil {
		return ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return ErrClosed
	}
	if !options.RemoveParallel && !options.RemoveLoops {
		return nil
	}
	return g.applyAtomicGraphTransformation("simplify graph", func(replacement *C.igraph_t) C.igraph_error_t {
		return C.go_igraph_simplify(
			replacement, booltoint(options.RemoveParallel), booltoint(options.RemoveLoops),
		)
	})
}

// ConvertToDirectedInPlace atomically converts an undirected graph according
// to mode. A graph that is already directed is unchanged after mode validation.
//
// No edge ID mapping is returned: mutual conversion duplicates edges and
// igraph 1.0.1 exposes no provenance for the generated edge IDs. Vertex IDs
// remain unchanged. Existing graph, vertex, and edge attributes follow
// upstream conversion behavior; this package does not yet expose attributes.
//
//igraph:bind igraph_to_directed
func (g *Graph) ConvertToDirectedInPlace(mode DirectedConversionMode) error {
	if g == nil {
		return ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return ErrClosed
	}
	cMode, err := mode.cValue()
	if err != nil {
		return err
	}
	if C.igraph_is_directed(&g.graph) != booltoint(false) {
		return nil
	}
	return g.applyAtomicGraphTransformation("convert graph to directed", func(replacement *C.igraph_t) C.igraph_error_t {
		return C.go_igraph_to_directed(replacement, cMode)
	})
}

// ConvertToUndirectedInPlace atomically converts a directed graph according to
// mode. A graph that is already undirected is unchanged after mode validation.
//
// No edge ID mapping is returned: collapse and mutual modes merge, discard, or
// pair edges, and igraph 1.0.1 exposes no provenance for those changes. Vertex
// IDs remain unchanged. The binding passes no attribute-combination policy;
// collapse and mutual conversion therefore discard edge attributes, while each
// conversion preserves them. Graph attributes are not currently exposed.
//
//igraph:bind igraph_to_undirected
func (g *Graph) ConvertToUndirectedInPlace(mode UndirectedConversionMode) error {
	if g == nil {
		return ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return ErrClosed
	}
	cMode, err := mode.cValue()
	if err != nil {
		return err
	}
	if C.igraph_is_directed(&g.graph) == booltoint(false) {
		return nil
	}
	return g.applyAtomicGraphTransformation("convert graph to undirected", func(replacement *C.igraph_t) C.igraph_error_t {
		return C.go_igraph_to_undirected(replacement, cMode)
	})
}

func (g *Graph) applyAtomicGraphTransformation(
	action string,
	transform func(*C.igraph_t) C.igraph_error_t,
) error {
	return executeGraphTransformation(&g.graph, graphTransformationOperations{
		clone: func(replacement, source *C.igraph_t) error {
			if code := C.go_igraph_copy(replacement, source); code != C.IGRAPH_SUCCESS {
				return igraphError(action+" copy", int(code))
			}
			return nil
		},
		transform: func(replacement *C.igraph_t) error {
			if code := transform(replacement); code != C.IGRAPH_SUCCESS {
				return igraphError(action, int(code))
			}
			return nil
		},
		destroy: func(replacement *C.igraph_t) { C.igraph_destroy(replacement) },
		commit:  replaceInitializedGraph,
	})
}

type graphTransformationOperations struct {
	clone     func(*C.igraph_t, *C.igraph_t) error
	transform func(*C.igraph_t) error
	destroy   func(*C.igraph_t)
	commit    func(*C.igraph_t, *C.igraph_t)
}

func executeGraphTransformation(
	destination *C.igraph_t,
	operations graphTransformationOperations,
) error {
	var replacement C.igraph_t
	return executeAtomicTransformation(
		func() error { return operations.clone(&replacement, destination) },
		func() error { return operations.transform(&replacement) },
		func() { operations.destroy(&replacement) },
		func() { operations.commit(destination, &replacement) },
	)
}

func executeAtomicTransformation(
	clone func() error,
	transform func() error,
	destroy func(),
	commit func(),
) error {
	if err := clone(); err != nil {
		return err
	}
	committed := false
	defer func() {
		if !committed {
			destroy()
		}
	}()
	if err := transform(); err != nil {
		return err
	}
	commit()
	committed = true
	return nil
}
