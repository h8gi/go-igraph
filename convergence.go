package igraph

// #cgo pkg-config: igraph
// #include <igraph.h>
// #include "algorithm_cgo.h"
import "C"

import "fmt"

// ConvergenceDegreeResult contains edge-ID-aligned convergence analysis.
// Every field is a non-nil Go-owned slice with one value per edge.
type ConvergenceDegreeResult struct {
	Convergence    []float64
	InputSetSizes  []float64
	OutputSetSizes []float64
}

// EdgeConvergenceDegree computes the convergence degree and supporting input
// and output set sizes for every edge. Directed convergence is
// (input-output)/(input+output): positive values are convergent and negative
// values divergent. For an undirected graph, igraph arbitrarily orients each
// edge, reports the associated set sizes, and returns the absolute convergence.
// Loops or otherwise unused edges can have zero input and output sizes and NaN
// convergence. Parallel edges remain independently indexed by edge ID.
//
// The graph is borrowed for the synchronous call. All returned slices are
// Go-owned and remain valid after graph closure.
//
//igraph:bind igraph_convergence_degree
func (g *Graph) EdgeConvergenceDegree() (ConvergenceDegreeResult, error) {
	if g == nil {
		return ConvergenceDegreeResult{}, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return ConvergenceDegreeResult{}, ErrClosed
	}

	return collectConvergenceDegree(int(C.igraph_ecount(&g.graph)), func(result, ins, outs *realVector) int {
		return int(C.go_igraph_convergence_degree(
			&g.graph, &result.value, &ins.value, &outs.value,
		))
	})
}

func collectConvergenceDegree(
	edgeCount int,
	calculate func(result, ins, outs *realVector) int,
) (ConvergenceDegreeResult, error) {
	result, err := newRealVectorSize(0)
	if err != nil {
		return ConvergenceDegreeResult{}, err
	}
	defer result.close()
	ins, err := newRealVectorSize(0)
	if err != nil {
		return ConvergenceDegreeResult{}, err
	}
	defer ins.close()
	outs, err := newRealVectorSize(0)
	if err != nil {
		return ConvergenceDegreeResult{}, err
	}
	defer outs.close()
	if code := calculate(result, ins, outs); code != int(C.IGRAPH_SUCCESS) {
		return ConvergenceDegreeResult{}, igraphError("calculate edge convergence degree", code)
	}
	convergence, err := result.slice()
	if err != nil {
		return ConvergenceDegreeResult{}, err
	}
	inputSetSizes, err := ins.slice()
	if err != nil {
		return ConvergenceDegreeResult{}, err
	}
	outputSetSizes, err := outs.slice()
	if err != nil {
		return ConvergenceDegreeResult{}, err
	}
	if len(convergence) != edgeCount || len(inputSetSizes) != edgeCount || len(outputSetSizes) != edgeCount {
		return ConvergenceDegreeResult{}, fmt.Errorf(
			"igraph: edge convergence result lengths (%d, %d, %d) do not match edge count %d",
			len(convergence), len(inputSetSizes), len(outputSetSizes), edgeCount,
		)
	}
	return ConvergenceDegreeResult{
		Convergence: convergence, InputSetSizes: inputSetSizes, OutputSetSizes: outputSetSizes,
	}, nil
}
