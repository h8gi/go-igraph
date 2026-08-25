package igraph

// #cgo pkg-config: igraph
// #include <igraph.h>
// #include "percolation_cgo.h"
import "C"

import (
	"errors"
	"fmt"
)

// BondPercolationResult is a Go-owned bond-percolation curve. Entry i records
// the state immediately after adding edge order[i]. GiantComponentSizes is the
// largest component size and ActiveVertexCounts is the number of vertices
// incident to at least one active edge. Both slices are non-nil and aligned.
type BondPercolationResult struct {
	GiantComponentSizes []int
	ActiveVertexCounts  []int
}

// SitePercolationResult is a Go-owned site-percolation curve. Entry i records
// the state immediately after adding vertex order[i]. GiantComponentSizes is
// the largest component size and ActiveEdgeCounts is the number of active edge
// endpoints. Non-loop edges contribute one and loops contribute two, following
// igraph's site-percolation convention. Both slices are non-nil and aligned.
type SitePercolationResult struct {
	GiantComponentSizes []int
	ActiveEdgeCounts    []int
}

// BondPercolation returns the component-growth curve produced by adding every
// source edge in edgeOrder. edgeOrder is borrowed only for the synchronous
// call and must be a complete, duplicate-free permutation of edge IDs. Edge
// directions are ignored; loops and parallel edges retain their source edge
// IDs. Empty graphs return non-nil empty slices.
//
// This API wraps an experimental igraph function whose upstream signature may
// change in a minor release.
//
//igraph:bind igraph_bond_percolation
func (g *Graph) BondPercolation(edgeOrder []int) (BondPercolationResult, error) {
	if g == nil {
		return BondPercolationResult{}, ErrClosed
	}
	g.mu.RLock()
	defer g.mu.RUnlock()
	if g.closed {
		return BondPercolationResult{}, ErrClosed
	}
	edgeCount := int(C.igraph_ecount(&g.graph))
	if err := validatePercolationOrder(edgeOrder, edgeCount, "edge"); err != nil {
		return BondPercolationResult{}, err
	}
	series, err := collectPercolation(edgeOrder, edgeCount, "bond percolation", percolationOperations{
		newVector: newIntVector,
		close:     (*intVector).close,
		query: func(order, giant, active *intVector) error {
			if code := C.go_igraph_bond_percolation(&g.graph, &giant.value, &active.value, &order.value); code != C.IGRAPH_SUCCESS {
				return igraphError("calculate bond percolation", int(code))
			}
			return nil
		},
		slice: func(vector *intVector) ([]int, error) { return vector.slice() },
	})
	if err != nil {
		return BondPercolationResult{}, err
	}
	return BondPercolationResult{GiantComponentSizes: series.giant, ActiveVertexCounts: series.active}, nil
}

// SitePercolation returns the component-growth curve produced by adding every
// source vertex in vertexOrder. vertexOrder is borrowed only for the
// synchronous call and must be a complete, duplicate-free permutation of
// vertex IDs. Edge directions are ignored. A loop becomes active with its
// vertex and contributes two to ActiveEdgeCounts; parallel edges are counted
// separately. Empty graphs return non-nil empty slices.
//
// This API wraps an experimental igraph function whose upstream signature may
// change in a minor release.
//
//igraph:bind igraph_site_percolation
func (g *Graph) SitePercolation(vertexOrder []int) (SitePercolationResult, error) {
	if g == nil {
		return SitePercolationResult{}, ErrClosed
	}
	g.mu.RLock()
	defer g.mu.RUnlock()
	if g.closed {
		return SitePercolationResult{}, ErrClosed
	}
	vertexCount := int(C.igraph_vcount(&g.graph))
	if err := validatePercolationOrder(vertexOrder, vertexCount, "vertex"); err != nil {
		return SitePercolationResult{}, err
	}
	series, err := collectPercolation(vertexOrder, vertexCount, "site percolation", percolationOperations{
		newVector: newIntVector,
		close:     (*intVector).close,
		query: func(order, giant, active *intVector) error {
			if code := C.go_igraph_site_percolation(&g.graph, &giant.value, &active.value, &order.value); code != C.IGRAPH_SUCCESS {
				return igraphError("calculate site percolation", int(code))
			}
			return nil
		},
		slice: func(vector *intVector) ([]int, error) { return vector.slice() },
	})
	if err != nil {
		return SitePercolationResult{}, err
	}
	return SitePercolationResult{GiantComponentSizes: series.giant, ActiveEdgeCounts: series.active}, nil
}

// EdgeListPercolation returns the bond-percolation curve produced by adding
// the supplied endpoint pairs in slice order. The input is borrowed only for
// the synchronous call and copied into C-owned storage. Vertex count is
// inferred as one plus the largest endpoint, so isolated vertices cannot be
// represented. Endpoints must be non-negative. Loops and parallel pairs are
// allowed, and endpoint direction has no meaning. Empty input returns non-nil
// empty slices.
//
// This graph-independent API wraps an experimental igraph function whose
// upstream signature may change in a minor release.
//
//igraph:bind igraph_edgelist_percolation
func EdgeListPercolation(edges []Edge) (BondPercolationResult, error) {
	endpoints, err := percolationEndpoints(edges)
	if err != nil {
		return BondPercolationResult{}, err
	}
	series, err := collectPercolation(endpoints, len(edges), "edge-list percolation", percolationOperations{
		newVector: newIntVector,
		close:     (*intVector).close,
		query: func(edgeVector, giant, active *intVector) error {
			if code := C.go_igraph_edgelist_percolation(&edgeVector.value, &giant.value, &active.value); code != C.IGRAPH_SUCCESS {
				return igraphError("calculate edge-list percolation", int(code))
			}
			return nil
		},
		slice: func(vector *intVector) ([]int, error) { return vector.slice() },
	})
	if err != nil {
		return BondPercolationResult{}, err
	}
	return BondPercolationResult{GiantComponentSizes: series.giant, ActiveVertexCounts: series.active}, nil
}

func validatePercolationOrder(order []int, size int, kind string) error {
	if len(order) != size {
		return fmt.Errorf("igraph: %s order length %d does not match %s count %d", kind, len(order), kind, size)
	}
	seen := make([]bool, size)
	for index, id := range order {
		if id < 0 || id >= size || seen[id] {
			return fmt.Errorf("igraph: %s order must be a permutation; invalid %s ID %d at position %d", kind, kind, id, index)
		}
		seen[id] = true
	}
	return nil
}

func percolationEndpoints(edges []Edge) ([]int, error) {
	if len(edges) > int(^uint(0)>>1)/2 {
		return nil, errors.New("igraph: percolation edge list is too large")
	}
	endpoints := make([]int, 0, 2*len(edges))
	for index, edge := range edges {
		for _, endpoint := range []struct {
			name  string
			value int
		}{{"from", edge.From}, {"to", edge.To}} {
			if endpoint.value < 0 {
				return nil, fmt.Errorf("igraph: percolation edge %d %s endpoint must be non-negative: %d", index, endpoint.name, endpoint.value)
			}
			if _, err := intToIgraphInt(endpoint.value, fmt.Sprintf("percolation edge %d %s endpoint", index, endpoint.name)); err != nil {
				return nil, err
			}
			endpoints = append(endpoints, endpoint.value)
		}
	}
	return endpoints, nil
}

type percolationSeries struct {
	giant  []int
	active []int
}

type percolationOperations struct {
	newVector func([]int) (*intVector, error)
	close     func(*intVector)
	query     func(input, giant, active *intVector) error
	slice     func(*intVector) ([]int, error)
}

func collectPercolation(input []int, expectedLength int, label string, operations percolationOperations) (percolationSeries, error) {
	inputVector, err := operations.newVector(input)
	if err != nil {
		return percolationSeries{}, err
	}
	defer operations.close(inputVector)
	giant, err := operations.newVector(nil)
	if err != nil {
		return percolationSeries{}, err
	}
	defer operations.close(giant)
	active, err := operations.newVector(nil)
	if err != nil {
		return percolationSeries{}, err
	}
	defer operations.close(active)
	if err := operations.query(inputVector, giant, active); err != nil {
		return percolationSeries{}, err
	}
	result := percolationSeries{}
	result.giant, err = operations.slice(giant)
	if err == nil {
		result.active, err = operations.slice(active)
	}
	if err != nil {
		return percolationSeries{}, err
	}
	if err := validatePercolationSeries(result, expectedLength, label); err != nil {
		return percolationSeries{}, err
	}
	return result, nil
}

func validatePercolationSeries(series percolationSeries, expectedLength int, label string) error {
	if series.giant == nil || series.active == nil {
		return fmt.Errorf("igraph: %s returned nil output storage", label)
	}
	if len(series.giant) != expectedLength || len(series.active) != expectedLength {
		return fmt.Errorf("igraph: %s returned lengths %d and %d, want %d", label, len(series.giant), len(series.active), expectedLength)
	}
	for index := range expectedLength {
		if series.giant[index] < 0 || series.active[index] < 0 {
			return fmt.Errorf("igraph: %s returned a negative value at step %d", label, index)
		}
		if index > 0 && (series.giant[index] < series.giant[index-1] || series.active[index] < series.active[index-1]) {
			return fmt.Errorf("igraph: %s returned a decreasing value at step %d", label, index)
		}
	}
	return nil
}
