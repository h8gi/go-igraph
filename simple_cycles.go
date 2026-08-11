package igraph

// #cgo pkg-config: igraph
// #include <igraph.h>
// #include "cycle_cgo.h"
import "C"

import (
	"errors"
	"fmt"
)

// SimpleCycleOptions controls bounded simple-cycle enumeration. Direction uses
// the existing outgoing, incoming, or all-edge traversal interpretation and is
// ignored for undirected graphs. Non-nil MinLength and MaxLength values are
// positive inclusive bounds. MaxResults must be positive; the zero value is
// intentionally invalid so every potentially exponential call is explicit.
type SimpleCycleOptions struct {
	Direction  DirectionMode
	MinLength  *int
	MaxLength  *int
	MaxResults int
}

// SimpleCyclesResult is a Go-owned bounded collection. Cycles is always
// non-nil. Truncated is true only when at least one additional cycle matching
// the requested direction and length bounds existed beyond MaxResults.
type SimpleCyclesResult struct {
	Cycles    []Cycle
	Truncated bool
}

// SimpleCycles enumerates at most MaxResults simple cycles. This binds the
// experimental igraph_simple_cycles API in pinned igraph 1.0.1; its public
// shape may therefore be revisited if upstream changes incompatibly.
//
// Cycle start, orientation, and outer result order are not compatibility
// promises. Graph and option inputs are borrowed only for the synchronous
// call. Every returned slice is non-nil, Go-owned, and remains valid after the
// graph is closed.
//
//igraph:bind igraph_simple_cycles
func (g *Graph) SimpleCycles(options SimpleCycleOptions) (SimpleCyclesResult, error) {
	return g.simpleCycles(options, nil)
}

type simpleCycleParameters struct {
	direction       DirectionMode
	minLength       int
	maxLength       int
	requestedResult int
}

func prepareSimpleCycleParameters(options SimpleCycleOptions) (simpleCycleParameters, error) {
	if _, err := options.Direction.cValue(); err != nil {
		return simpleCycleParameters{}, err
	}
	if options.MaxResults <= 0 {
		return simpleCycleParameters{}, fmt.Errorf("igraph: simple-cycle maximum results must be positive: %d", options.MaxResults)
	}
	if _, err := intToIgraphInt(options.MaxResults, "simple-cycle maximum results"); err != nil {
		return simpleCycleParameters{}, err
	}
	if options.MaxResults == int(^uint(0)>>1) {
		return simpleCycleParameters{}, errors.New("igraph: simple-cycle maximum results plus one is out of range")
	}
	requested := options.MaxResults + 1
	if _, err := intToIgraphInt(requested, "simple-cycle requested results"); err != nil {
		return simpleCycleParameters{}, err
	}
	minLength, err := optionalSimpleCycleLength(options.MinLength, "minimum")
	if err != nil {
		return simpleCycleParameters{}, err
	}
	maxLength, err := optionalSimpleCycleLength(options.MaxLength, "maximum")
	if err != nil {
		return simpleCycleParameters{}, err
	}
	if minLength >= 0 && maxLength >= 0 && minLength > maxLength {
		return simpleCycleParameters{}, fmt.Errorf("igraph: simple-cycle minimum length %d exceeds maximum length %d", minLength, maxLength)
	}
	return simpleCycleParameters{
		direction:       options.Direction,
		minLength:       minLength,
		maxLength:       maxLength,
		requestedResult: requested,
	}, nil
}

func optionalSimpleCycleLength(value *int, name string) (int, error) {
	if value == nil {
		return -1, nil
	}
	if *value <= 0 {
		return 0, fmt.Errorf("igraph: simple-cycle %s length must be positive: %d", name, *value)
	}
	if _, err := intToIgraphInt(*value, "simple-cycle "+name+" length"); err != nil {
		return 0, err
	}
	return *value, nil
}

type cycleListInitializer func() (*intVectorList, error)
type cycleListCloser func(*intVectorList)
type cycleListConverter func(*intVectorList) ([][]int, error)

type simpleCycleAdapters struct {
	initialize cycleListInitializer
	close      cycleListCloser
	call       func(*Graph, *intVectorList, *intVectorList, simpleCycleParameters) int
	convert    cycleListConverter
}

func defaultSimpleCycleAdapters() simpleCycleAdapters {
	return simpleCycleAdapters{
		initialize: newIntVectorList,
		close:      func(list *intVectorList) { list.close() },
		call: func(g *Graph, vertices, edges *intVectorList, parameters simpleCycleParameters) int {
			cMode, _ := parameters.direction.cValue()
			return int(C.go_igraph_simple_cycles(
				&g.graph, &vertices.value, &edges.value, cMode,
				C.igraph_int_t(parameters.minLength),
				C.igraph_int_t(parameters.maxLength),
				C.igraph_int_t(parameters.requestedResult),
			))
		},
		convert: func(list *intVectorList) ([][]int, error) { return list.slices() },
	}
}

func newCycleListPair(
	initialize cycleListInitializer,
	closeList cycleListCloser,
) (*intVectorList, *intVectorList, error) {
	vertices, err := initialize()
	if err != nil {
		return nil, nil, err
	}
	edges, err := initialize()
	if err != nil {
		closeList(vertices)
		return nil, nil, err
	}
	return vertices, edges, nil
}

func (g *Graph) simpleCycles(options SimpleCycleOptions, adapters *simpleCycleAdapters) (SimpleCyclesResult, error) {
	if g == nil {
		return SimpleCyclesResult{}, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return SimpleCyclesResult{}, ErrClosed
	}
	parameters, err := prepareSimpleCycleParameters(options)
	if err != nil {
		return SimpleCyclesResult{}, err
	}
	resolved := defaultSimpleCycleAdapters()
	if adapters != nil {
		resolved = *adapters
	}
	vertexLists, edgeLists, err := newCycleListPair(resolved.initialize, resolved.close)
	if err != nil {
		return SimpleCyclesResult{}, err
	}
	defer resolved.close(vertexLists)
	defer resolved.close(edgeLists)
	if code := resolved.call(g, vertexLists, edgeLists, parameters); code != int(C.IGRAPH_SUCCESS) {
		return SimpleCyclesResult{}, igraphError("enumerate simple cycles", code)
	}
	vertices, err := resolved.convert(vertexLists)
	if err != nil {
		return SimpleCyclesResult{}, err
	}
	edges, err := resolved.convert(edgeLists)
	if err != nil {
		return SimpleCyclesResult{}, err
	}
	if len(vertices) != len(edges) {
		return SimpleCyclesResult{}, fmt.Errorf("igraph: simple-cycle vertex list count %d does not match edge list count %d", len(vertices), len(edges))
	}
	truncated := len(vertices) > options.MaxResults
	count := len(vertices)
	if truncated {
		count = options.MaxResults
	}
	cycles := make([]Cycle, count)
	for i := range cycles {
		cycles[i], err = newCycle(vertices[i], edges[i])
		if err != nil {
			return SimpleCyclesResult{}, fmt.Errorf("igraph: convert simple cycle %d: %w", i, err)
		}
	}
	return SimpleCyclesResult{Cycles: cycles, Truncated: truncated}, nil
}
