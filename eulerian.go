package igraph

// #cgo pkg-config: igraph
// #include <igraph.h>
// #include "eulerian_cgo.h"
import "C"

// EulerianStatus reports whether a graph admits an Eulerian open path and/or
// cycle. HasPath is also true for graphs that have an Eulerian cycle.
type EulerianStatus struct {
	HasPath  bool
	HasCycle bool
}

// EulerianStatus returns Eulerian existence information without constructing
// a traversal.
//
//igraph:bind igraph_is_eulerian
func (g *Graph) EulerianStatus() (EulerianStatus, error) {
	if g == nil {
		return EulerianStatus{}, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return EulerianStatus{}, ErrClosed
	}
	var hasPath, hasCycle C.igraph_bool_t
	if code := C.go_igraph_is_eulerian(&g.graph, &hasPath, &hasCycle); code != C.IGRAPH_SUCCESS {
		return EulerianStatus{}, igraphError("check Eulerian status", int(code))
	}
	return EulerianStatus{HasPath: hasPath != booltoint(false), HasCycle: hasCycle != booltoint(false)}, nil
}

// EulerianPath returns an aligned Go-owned edge and vertex traversal. Found is
// false with non-nil empty slices when the graph has no Eulerian path.
//
//igraph:bind igraph_eulerian_path
func (g *Graph) EulerianPath() (Path, error) {
	return g.eulerianTraversal(false)
}

// EulerianCycle returns an aligned Go-owned closed traversal. Found is false
// with non-nil empty slices when the graph has no Eulerian cycle.
//
//igraph:bind igraph_eulerian_cycle
func (g *Graph) EulerianCycle() (Path, error) {
	return g.eulerianTraversal(true)
}

func (g *Graph) eulerianTraversal(cycle bool) (Path, error) {
	if g == nil {
		return Path{}, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return Path{}, ErrClosed
	}
	var hasPath, hasCycle C.igraph_bool_t
	if code := C.go_igraph_is_eulerian(&g.graph, &hasPath, &hasCycle); code != C.IGRAPH_SUCCESS {
		return Path{}, igraphError("check Eulerian status", int(code))
	}
	found := hasPath != booltoint(false)
	if cycle {
		found = hasCycle != booltoint(false)
	}
	if !found {
		return Path{Vertices: []int{}, Edges: []int{}, Found: false}, nil
	}
	edges, err := newIntVector(nil)
	if err != nil {
		return Path{}, err
	}
	defer edges.close()
	vertices, err := newIntVector(nil)
	if err != nil {
		return Path{}, err
	}
	defer vertices.close()
	var code C.igraph_error_t
	if cycle {
		code = C.go_igraph_eulerian_cycle(&g.graph, &edges.value, &vertices.value)
	} else {
		code = C.go_igraph_eulerian_path(&g.graph, &edges.value, &vertices.value)
	}
	if code != C.IGRAPH_SUCCESS {
		return Path{}, igraphError("construct Eulerian traversal", int(code))
	}
	vertexIDs, err := vertices.slice()
	if err != nil {
		return Path{}, err
	}
	edgeIDs, err := edges.slice()
	if err != nil {
		return Path{}, err
	}
	return Path{Vertices: vertexIDs, Edges: edgeIDs, Found: true}, nil
}
