package igraph

// #cgo pkg-config: igraph
// #include <igraph.h>
// #include "algorithm_cgo.h"
import "C"

import "fmt"

// ConnectednessMode controls how edge directions are interpreted by
// connected-component queries. Weak connectedness ignores edge directions;
// strong connectedness requires directed reachability in both directions.
// Its zero value is ConnectednessWeak. The distinction is ignored for
// undirected graphs.
type ConnectednessMode uint8

const (
	ConnectednessWeak ConnectednessMode = iota
	ConnectednessStrong
)

func (mode ConnectednessMode) cValue() (C.igraph_connectedness_t, error) {
	switch mode {
	case ConnectednessWeak:
		return C.IGRAPH_WEAK, nil
	case ConnectednessStrong:
		return C.IGRAPH_STRONG, nil
	default:
		return 0, fmt.Errorf("igraph: invalid connectedness mode: %d", mode)
	}
}

// ConnectedComponents is a Go-owned connected-component result. Membership is
// indexed by vertex ID and contains the corresponding component ID. Sizes is
// indexed by component ID, and Count is always equal to len(Sizes).
//
// Component IDs and their ordering are defined by upstream igraph. In strong
// mode, igraph currently assigns them in topological order. The slices remain
// valid and mutable after the source graph is closed.
type ConnectedComponents struct {
	Membership []int
	Sizes      []int
	Count      int
}

// ConnectedComponents returns all weakly or strongly connected components of
// the graph. Weak mode ignores edge directions; strong mode requires mutual
// directed reachability. The mode distinction is ignored for undirected
// graphs, although mode must still be a valid ConnectednessMode.
//
// The returned value owns all of its storage and remains valid after the graph
// is closed.
//
//igraph:bind igraph_connected_components
func (g *Graph) ConnectedComponents(mode ConnectednessMode) (ConnectedComponents, error) {
	if g == nil {
		return ConnectedComponents{}, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return ConnectedComponents{}, ErrClosed
	}
	cMode, err := mode.cValue()
	if err != nil {
		return ConnectedComponents{}, err
	}

	membership, err := newIntVector(nil)
	if err != nil {
		return ConnectedComponents{}, err
	}
	defer membership.close()
	sizes, err := newIntVector(nil)
	if err != nil {
		return ConnectedComponents{}, err
	}
	defer sizes.close()

	var cCount C.igraph_int_t
	if code := C.go_igraph_connected_components(&g.graph, &membership.value, &sizes.value, &cCount, cMode); code != C.IGRAPH_SUCCESS {
		return ConnectedComponents{}, igraphError("get connected components", int(code))
	}

	result := ConnectedComponents{}
	result.Membership, err = membership.slice()
	if err == nil {
		result.Sizes, err = sizes.slice()
	}
	if err == nil {
		result.Count, err = igraphIntToInt(cCount, "connected component count")
	}
	if err != nil {
		return ConnectedComponents{}, err
	}
	if err := validateConnectedComponents(result, int(C.igraph_vcount(&g.graph))); err != nil {
		return ConnectedComponents{}, err
	}
	return result, nil
}

func validateConnectedComponents(result ConnectedComponents, vertexCount int) error {
	if len(result.Membership) != vertexCount {
		return fmt.Errorf("igraph: connected component membership length %d does not match vertex count %d", len(result.Membership), vertexCount)
	}
	if len(result.Sizes) != result.Count {
		return fmt.Errorf("igraph: connected component size length %d does not match count %d", len(result.Sizes), result.Count)
	}
	observedSizes := make([]int, result.Count)
	for vertexID, componentID := range result.Membership {
		if componentID < 0 || componentID >= result.Count {
			return fmt.Errorf("igraph: connected component ID %d for vertex %d out of range [0, %d)", componentID, vertexID, result.Count)
		}
		observedSizes[componentID]++
	}
	for componentID, size := range result.Sizes {
		if size != observedSizes[componentID] {
			return fmt.Errorf("igraph: connected component %d size %d does not match membership size %d", componentID, size, observedSizes[componentID])
		}
	}
	return nil
}

// IsConnected reports whether the graph is weakly or strongly connected.
// Weak mode ignores edge directions; strong mode requires every vertex to be
// reachable from every other vertex. The mode distinction is ignored for
// undirected graphs, although mode must still be valid. A graph with no
// vertices is not connected.
//
//igraph:bind igraph_is_connected
func (g *Graph) IsConnected(mode ConnectednessMode) (bool, error) {
	if g == nil {
		return false, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return false, ErrClosed
	}
	cMode, err := mode.cValue()
	if err != nil {
		return false, err
	}

	var result C.igraph_bool_t
	if code := C.go_igraph_is_connected(&g.graph, &result, cMode); code != C.IGRAPH_SUCCESS {
		return false, igraphError("check connectedness", int(code))
	}
	return result != booltoint(false), nil
}
