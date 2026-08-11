package igraph

// #cgo pkg-config: igraph
// #include <igraph.h>
// #include "cycle_cgo.h"
import "C"

import (
	"errors"
	"fmt"
	"math"
)

// FundamentalCycleBasisOptions controls the breadth-first tree used to build
// a fundamental cycle basis. A nil Root covers all weak components; a non-nil
// Root limits the result to that vertex's weak component. A nil BFSCutoff
// computes a complete basis. A non-nil finite non-negative cutoff includes
// only cycles of length at most 2*cutoff+1 and therefore requires
// AllowIncomplete.
type FundamentalCycleBasisOptions struct {
	Root            *int
	BFSCutoff       *float64
	AllowIncomplete bool
}

// MinimumCycleBasisOptions controls minimum cycle-basis computation. A nil
// BFSCutoff requests an exact complete minimum basis. A non-nil finite
// non-negative cutoff limits candidate breadth-first searches; when
// AllowIncomplete is false upstream completes the basis, while true permits a
// non-spanning result containing only cycles of length at most 2*cutoff+1.
// NaturalOrder requests edge IDs in traversal order around every cycle; when
// false, each element is an unordered edge set.
type MinimumCycleBasisOptions struct {
	BFSCutoff       *float64
	AllowIncomplete bool
	NaturalOrder    bool
}

// FundamentalCycleBasis returns a fundamental cycle basis associated with an
// upstream breadth-first spanning forest. Edge directions are ignored and
// self-loops and parallel edges are supported.
//
// This binds an experimental API in pinned igraph 1.0.1. The upstream weights
// parameter is deliberately omitted because that release documents and
// implements it as unused. The graph and options are borrowed synchronously;
// the non-nil outer and inner result slices are Go-owned and survive Close.
//
//igraph:bind igraph_fundamental_cycles
func (g *Graph) FundamentalCycleBasis(options FundamentalCycleBasisOptions) ([][]int, error) {
	return g.fundamentalCycleBasis(options, nil)
}

// MinimumCycleBasis returns an unweighted minimum cycle basis. Edge directions
// are ignored and self-loops and parallel edges are supported. With a cutoff
// and AllowIncomplete, the result may not span the cycle space. With a cutoff
// and no AllowIncomplete, it spans the cycle space, but cycles longer than
// 2*cutoff+1 are not guaranteed to be the smallest possible choices.
//
// This binds an experimental API in pinned igraph 1.0.1. Its unused weights
// parameter is deliberately omitted. Results follow the same Go ownership
// contract as FundamentalCycleBasis.
//
//igraph:bind igraph_minimum_cycle_basis
func (g *Graph) MinimumCycleBasis(options MinimumCycleBasisOptions) ([][]int, error) {
	return g.minimumCycleBasis(options, nil)
}

type cycleBasisAdapters struct {
	initialize      func() (*intVectorList, error)
	close           func(*intVectorList)
	convert         func(*intVectorList) ([][]int, error)
	fundamentalCall func(
		*Graph, *intVectorList, int, float64,
	) int
	minimumCall func(
		*Graph, *intVectorList, float64, bool, bool,
	) int
}

func defaultCycleBasisAdapters() cycleBasisAdapters {
	return cycleBasisAdapters{
		initialize: newIntVectorList,
		close:      func(list *intVectorList) { list.close() },
		convert:    func(list *intVectorList) ([][]int, error) { return list.slices() },
		fundamentalCall: func(g *Graph, result *intVectorList, root int, cutoff float64) int {
			return int(C.go_igraph_fundamental_cycles(
				&g.graph, &result.value, C.igraph_int_t(root), C.igraph_real_t(cutoff),
			))
		},
		minimumCall: func(g *Graph, result *intVectorList, cutoff float64, complete, naturalOrder bool) int {
			return int(C.go_igraph_minimum_cycle_basis(
				&g.graph, &result.value, C.igraph_real_t(cutoff),
				booltoint(complete), booltoint(naturalOrder),
			))
		},
	}
}

func (g *Graph) fundamentalCycleBasis(options FundamentalCycleBasisOptions, adapters *cycleBasisAdapters) ([][]int, error) {
	if g == nil {
		return nil, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return nil, ErrClosed
	}
	root := -1
	if options.Root != nil {
		converted, err := intToIgraphInt(*options.Root, "fundamental cycle-basis root")
		if err != nil {
			return nil, err
		}
		root = int(converted)
		vertexCount := int(C.igraph_vcount(&g.graph))
		if root < 0 || root >= vertexCount {
			return nil, fmt.Errorf("igraph: fundamental cycle-basis root %d out of range [0, %d)", root, vertexCount)
		}
	}
	cutoff, err := cycleBasisCutoff(options.BFSCutoff)
	if err != nil {
		return nil, err
	}
	if options.BFSCutoff != nil && !options.AllowIncomplete {
		return nil, errors.New("igraph: fundamental cycle-basis cutoff requires AllowIncomplete")
	}
	resolved := defaultCycleBasisAdapters()
	if adapters != nil {
		resolved = *adapters
	}
	result, err := resolved.initialize()
	if err != nil {
		return nil, err
	}
	defer resolved.close(result)
	if code := resolved.fundamentalCall(g, result, root, cutoff); code != int(C.IGRAPH_SUCCESS) {
		return nil, igraphError("calculate fundamental cycle basis", code)
	}
	return resolved.convert(result)
}

func (g *Graph) minimumCycleBasis(options MinimumCycleBasisOptions, adapters *cycleBasisAdapters) ([][]int, error) {
	if g == nil {
		return nil, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return nil, ErrClosed
	}
	cutoff, err := cycleBasisCutoff(options.BFSCutoff)
	if err != nil {
		return nil, err
	}
	if options.BFSCutoff == nil && options.AllowIncomplete {
		return nil, errors.New("igraph: minimum cycle basis cannot allow incomplete output without a cutoff")
	}
	resolved := defaultCycleBasisAdapters()
	if adapters != nil {
		resolved = *adapters
	}
	result, err := resolved.initialize()
	if err != nil {
		return nil, err
	}
	defer resolved.close(result)
	complete := options.BFSCutoff == nil || !options.AllowIncomplete
	if code := resolved.minimumCall(g, result, cutoff, complete, options.NaturalOrder); code != int(C.IGRAPH_SUCCESS) {
		return nil, igraphError("calculate minimum cycle basis", code)
	}
	return resolved.convert(result)
}

func cycleBasisCutoff(value *float64) (float64, error) {
	if value == nil {
		return -1, nil
	}
	if math.IsNaN(*value) || math.IsInf(*value, 0) || *value < 0 {
		return 0, fmt.Errorf("igraph: cycle-basis BFS cutoff must be finite and non-negative: %v", *value)
	}
	return *value, nil
}
