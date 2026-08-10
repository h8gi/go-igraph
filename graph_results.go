package igraph

/*
#include <igraph.h>
#include "graph_results_cgo.h"
*/
import "C"
import (
	"fmt"
	"math"
)

// ResidualGraphResult represents the residual graph and residual edge capacities.
type ResidualGraphResult struct {
	// Graph is the derived residual graph, independently owned and caller-closed.
	Graph *Graph
	// ResidualCapacities contains the residual capacity for each edge in the residual graph.
	ResidualCapacities []float64
}

// GomoryHuTreeResult represents the Gomory-Hu tree graph and edge flows/capacities.
type GomoryHuTreeResult struct {
	// Tree is the Gomory-Hu tree graph, independently owned and caller-closed.
	Tree *Graph
	// Flows contains the capacity/flow value for each edge in the Gomory-Hu tree.
	Flows []float64
}

// DominatorTreeResult represents the dominator tree graph and dominator vertex mappings.
type DominatorTreeResult struct {
	// Tree is the dominator tree graph, independently owned and caller-closed.
	Tree *Graph
	// Dominators contains the immediate dominator vertex ID for each vertex (-1 for root/unreachable).
	Dominators []int
	// LeftOut contains the vertex IDs omitted from the dominator tree.
	LeftOut []int
}

// TarjanReductionResult represents the Even-Tarjan reduction graph and its edge capacities.
type TarjanReductionResult struct {
	// Graph is the reduced graph, independently owned and caller-closed.
	Graph *Graph
	// Capacities contains the edge capacities of the reduced graph.
	Capacities []float64
}

func validateFlows(g *Graph, flows []float64) (*realVector, error) {
	if flows == nil {
		return nil, nil
	}
	numEdges := int(C.igraph_ecount(&g.graph))
	if len(flows) != numEdges {
		return nil, fmt.Errorf("igraph: flows slice length (%d) must match number of edges (%d)", len(flows), numEdges)
	}
	for i, f := range flows {
		if math.IsNaN(f) || f < 0 {
			return nil, fmt.Errorf("igraph: flow value at index %d must be non-negative: %v", i, f)
		}
	}
	return newRealVector(flows)
}

// ResidualGraph computes the residual graph and residual capacities for a network.
// If capacities is nil, unit capacities (1.0) are assumed.
// If flows is nil, zero flows (0.0) are assumed.
//
//igraph:bind igraph_residual_graph
func (g *Graph) ResidualGraph(capacities []float64, flows []float64) (*ResidualGraphResult, error) {
	if g == nil {
		return nil, ErrClosed
	}
	g.mu.RLock()
	defer g.mu.RUnlock()
	if g.closed {
		return nil, ErrClosed
	}

	capVec, err := validateCapacities(g, capacities)
	if err != nil {
		return nil, err
	}
	if capVec != nil {
		defer capVec.close()
	}

	flowVec, err := validateFlows(g, flows)
	if err != nil {
		return nil, err
	}
	if flowVec != nil {
		defer flowVec.close()
	}

	resCapVec, _ := newRealVectorSize(0)
	defer resCapVec.close()

	var capPtr, flowPtr *C.igraph_vector_t
	if capVec != nil {
		capPtr = &capVec.value
	}
	if flowVec != nil {
		flowPtr = &flowVec.value
	}

	var residual C.igraph_t
	code := C.go_igraph_residual_graph(&g.graph, capPtr, &residual, &resCapVec.value, flowPtr)
	if code != C.IGRAPH_SUCCESS {
		return nil, igraphError("igraph_residual_graph", int(code))
	}

	resCaps, _ := resCapVec.slice()
	return &ResidualGraphResult{
		Graph:              adoptInitializedGraph(&residual),
		ResidualCapacities: resCaps,
	}, nil
}

// ReverseResidualGraph computes the reverse residual graph for a network.
//
//igraph:bind igraph_reverse_residual_graph
func (g *Graph) ReverseResidualGraph(capacities []float64, flows []float64) (*Graph, error) {
	if g == nil {
		return nil, ErrClosed
	}
	g.mu.RLock()
	defer g.mu.RUnlock()
	if g.closed {
		return nil, ErrClosed
	}

	capVec, err := validateCapacities(g, capacities)
	if err != nil {
		return nil, err
	}
	if capVec != nil {
		defer capVec.close()
	}

	flowVec, err := validateFlows(g, flows)
	if err != nil {
		return nil, err
	}
	if flowVec != nil {
		defer flowVec.close()
	}

	var capPtr, flowPtr *C.igraph_vector_t
	if capVec != nil {
		capPtr = &capVec.value
	}
	if flowVec != nil {
		flowPtr = &flowVec.value
	}

	var residual C.igraph_t
	code := C.go_igraph_reverse_residual_graph(&g.graph, capPtr, &residual, flowPtr)
	if code != C.IGRAPH_SUCCESS {
		return nil, igraphError("igraph_reverse_residual_graph", int(code))
	}

	return adoptInitializedGraph(&residual), nil
}

// GomoryHuTree computes the Gomory-Hu tree graph and edge flows for an undirected graph.
//
//igraph:bind igraph_gomory_hu_tree
func (g *Graph) GomoryHuTree(capacities []float64) (*GomoryHuTreeResult, error) {
	if g == nil {
		return nil, ErrClosed
	}
	g.mu.RLock()
	defer g.mu.RUnlock()
	if g.closed {
		return nil, ErrClosed
	}

	capVec, err := validateCapacities(g, capacities)
	if err != nil {
		return nil, err
	}
	if capVec != nil {
		defer capVec.close()
	}

	flowVec, _ := newRealVectorSize(0)
	defer flowVec.close()

	var capPtr *C.igraph_vector_t
	if capVec != nil {
		capPtr = &capVec.value
	}

	var tree C.igraph_t
	code := C.go_igraph_gomory_hu_tree(&g.graph, &tree, &flowVec.value, capPtr)
	if code != C.IGRAPH_SUCCESS {
		return nil, igraphError("igraph_gomory_hu_tree", int(code))
	}

	flows, _ := flowVec.slice()
	return &GomoryHuTreeResult{
		Tree:  adoptInitializedGraph(&tree),
		Flows: flows,
	}, nil
}

// DominatorTree computes the dominator tree of a directed graph from root vertex.
//
//igraph:bind igraph_dominator_tree
func (g *Graph) DominatorTree(root int, mode DirectionMode) (*DominatorTreeResult, error) {
	if g == nil {
		return nil, ErrClosed
	}
	g.mu.RLock()
	defer g.mu.RUnlock()
	if g.closed {
		return nil, ErrClosed
	}

	cRoot, err := validateVertex(g, root, "root")
	if err != nil {
		return nil, err
	}

	cMode, err := mode.cValue()
	if err != nil {
		return nil, err
	}

	domVec, _ := newIntVector(nil)
	defer domVec.close()

	leftoutVec, _ := newIntVector(nil)
	defer leftoutVec.close()

	var domtree C.igraph_t
	code := C.go_igraph_dominator_tree(&g.graph, cRoot, &domVec.value, &domtree, &leftoutVec.value, cMode)
	if code != C.IGRAPH_SUCCESS {
		return nil, igraphError("igraph_dominator_tree", int(code))
	}

	doms, _ := domVec.slice()
	leftout, _ := leftoutVec.slice()

	return &DominatorTreeResult{
		Tree:       adoptInitializedGraph(&domtree),
		Dominators: doms,
		LeftOut:    leftout,
	}, nil
}

// EvenTarjanReduction computes the Even-Tarjan reduction of a graph.
//
//igraph:bind igraph_even_tarjan_reduction
func (g *Graph) EvenTarjanReduction() (*TarjanReductionResult, error) {
	if g == nil {
		return nil, ErrClosed
	}
	g.mu.RLock()
	defer g.mu.RUnlock()
	if g.closed {
		return nil, ErrClosed
	}

	capVec, _ := newRealVectorSize(0)
	defer capVec.close()

	var graphbar C.igraph_t
	code := C.go_igraph_even_tarjan_reduction(&g.graph, &graphbar, &capVec.value)
	if code != C.IGRAPH_SUCCESS {
		return nil, igraphError("igraph_even_tarjan_reduction", int(code))
	}

	caps, _ := capVec.slice()
	return &TarjanReductionResult{
		Graph:      adoptInitializedGraph(&graphbar),
		Capacities: caps,
	}, nil
}
