package igraph

// #cgo pkg-config: igraph
// #include <igraph.h>
// #include "clique_cgo.h"
// #include "graphlets_cgo.h"
// #include "isomorphism_cgo.h"
import "C"

import (
	"fmt"
	"math"
	"sort"
)

// GraphletsResult contains an aligned graphlet basis and its projected
// coefficients. Cliques[i] has coefficient Mu[i]. All slices and nested
// slices are Go-owned and remain valid after the source graph is closed.
type GraphletsResult struct {
	Cliques [][]int
	Mu      []float64
}

// GraphletsCandidateBasisResult contains an aligned candidate basis and the
// highest weight threshold at which each clique appears. Cliques[i] has
// threshold Thresholds[i]. All returned storage is Go-owned.
type GraphletsCandidateBasisResult struct {
	Cliques    [][]int
	Thresholds []float64
}

// Graphlets calculates a candidate graphlet basis, projects the graph onto
// it for niter iterations, and orders graphlets by decreasing coefficient.
// The graph must be simple when directions are ignored. Nil weights mean one
// per edge; non-nil weights are borrowed synchronously and must contain one
// finite non-negative value per edge. niter must be non-negative.
//
//igraph:bind igraph_graphlets
func (g *Graph) Graphlets(weights []float64, niter int) (GraphletsResult, error) {
	return g.graphlets(weights, niter, nil)
}

// GraphletsCandidateBasis calculates the cliques considered by graphlet
// projection and their aligned thresholds. Graph shape and weight semantics
// match Graphlets.
//
//igraph:bind igraph_graphlets_candidate_basis
func (g *Graph) GraphletsCandidateBasis(weights []float64) (GraphletsCandidateBasisResult, error) {
	return g.graphletsCandidateBasis(weights, nil)
}

// GraphletsProject projects edge weights onto a caller-supplied graphlet
// basis. Every inner clique must contain at least two unique in-range vertices
// that form a clique when directions are ignored; duplicate basis cliques are
// rejected. Non-empty initialMu supplies one finite non-negative starting
// coefficient per clique; nil or empty initialMu starts every coefficient at
// one. Inputs are borrowed synchronously and the returned slice is Go-owned.
//
//igraph:bind igraph_graphlets_project
func (g *Graph) GraphletsProject(
	cliques [][]int,
	initialMu []float64,
	weights []float64,
	niter int,
) ([]float64, error) {
	return g.graphletsProject(cliques, initialMu, weights, niter, nil)
}

type graphletAdapters struct {
	initializeList func() (*intVectorList, error)
	createList     func([][]int) (*intVectorList, error)
	closeList      func(*intVectorList)
	convertList    func(*intVectorList) ([][]int, error)
	initializeReal func() (*realVector, error)
	createReal     func([]float64) (*realVector, error)
	closeReal      func(*realVector)
	convertReal    func(*realVector) ([]float64, error)
	graphletsCall  func(*Graph, *realVector, *intVectorList, *realVector, int) int
	candidateCall  func(*Graph, *realVector, *intVectorList, *realVector) int
	projectCall    func(*Graph, *realVector, *intVectorList, *realVector, bool, int) int
	simpleCall     func(*Graph) (bool, int)
}

func defaultGraphletAdapters() graphletAdapters {
	return graphletAdapters{
		initializeList: newIntVectorList,
		createList:     newGraphletCliqueList,
		closeList:      func(list *intVectorList) { list.close() },
		convertList:    func(list *intVectorList) ([][]int, error) { return list.slices() },
		initializeReal: func() (*realVector, error) { return newRealVectorSize(0) },
		createReal:     newRealVector,
		closeReal:      func(vector *realVector) { vector.close() },
		convertReal:    func(vector *realVector) ([]float64, error) { return vector.slice() },
		graphletsCall: func(g *Graph, weights *realVector, cliques *intVectorList, mu *realVector, niter int) int {
			return int(C.go_igraph_graphlets(
				&g.graph, &weights.value, &cliques.value, &mu.value, C.igraph_int_t(niter),
			))
		},
		candidateCall: func(g *Graph, weights *realVector, cliques *intVectorList, thresholds *realVector) int {
			return int(C.go_igraph_graphlets_candidate_basis(
				&g.graph, &weights.value, &cliques.value, &thresholds.value,
			))
		},
		projectCall: func(g *Graph, weights *realVector, cliques *intVectorList, mu *realVector, startMu bool, niter int) int {
			return int(C.go_igraph_graphlets_project(
				&g.graph, &weights.value, &cliques.value, &mu.value,
				booltoint(startMu), C.igraph_int_t(niter),
			))
		},
		simpleCall: func(g *Graph) (bool, int) {
			var simple C.igraph_bool_t
			code := C.go_igraph_graphlets_is_simple(&g.graph, &simple)
			return simple != booltoint(false), int(code)
		},
	}
}

func resolvedGraphletAdapters(adapters *graphletAdapters) graphletAdapters {
	if adapters == nil {
		return defaultGraphletAdapters()
	}
	return *adapters
}

func (g *Graph) graphlets(weights []float64, niter int, adapters *graphletAdapters) (GraphletsResult, error) {
	if g == nil {
		return GraphletsResult{}, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return GraphletsResult{}, ErrClosed
	}
	if err := validateGraphletIterations(niter); err != nil {
		return GraphletsResult{}, err
	}
	resolved := resolvedGraphletAdapters(adapters)
	cWeights, err := newGraphletWeights(&g.graph, weights, resolved.createReal)
	if err != nil {
		return GraphletsResult{}, err
	}
	defer resolved.closeReal(cWeights)
	if err := validateGraphletGraph(g, resolved); err != nil {
		return GraphletsResult{}, err
	}
	if C.igraph_ecount(&g.graph) == 0 {
		return GraphletsResult{Cliques: make([][]int, 0), Mu: make([]float64, 0)}, nil
	}
	cliques, mu, err := initializeGraphletOutputs(resolved)
	if err != nil {
		return GraphletsResult{}, err
	}
	defer resolved.closeList(cliques)
	defer resolved.closeReal(mu)
	if code := resolved.graphletsCall(g, cWeights, cliques, mu, niter); code != int(C.IGRAPH_SUCCESS) {
		return GraphletsResult{}, igraphError("calculate graphlets", code)
	}
	return convertGraphletResult(cliques, mu, int(C.igraph_vcount(&g.graph)), resolved)
}

func (g *Graph) graphletsCandidateBasis(weights []float64, adapters *graphletAdapters) (GraphletsCandidateBasisResult, error) {
	if g == nil {
		return GraphletsCandidateBasisResult{}, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return GraphletsCandidateBasisResult{}, ErrClosed
	}
	resolved := resolvedGraphletAdapters(adapters)
	cWeights, err := newGraphletWeights(&g.graph, weights, resolved.createReal)
	if err != nil {
		return GraphletsCandidateBasisResult{}, err
	}
	defer resolved.closeReal(cWeights)
	if err := validateGraphletGraph(g, resolved); err != nil {
		return GraphletsCandidateBasisResult{}, err
	}
	if C.igraph_ecount(&g.graph) == 0 {
		return GraphletsCandidateBasisResult{
			Cliques: make([][]int, 0), Thresholds: make([]float64, 0),
		}, nil
	}
	cliques, thresholds, err := initializeGraphletOutputs(resolved)
	if err != nil {
		return GraphletsCandidateBasisResult{}, err
	}
	defer resolved.closeList(cliques)
	defer resolved.closeReal(thresholds)
	if code := resolved.candidateCall(g, cWeights, cliques, thresholds); code != int(C.IGRAPH_SUCCESS) {
		return GraphletsCandidateBasisResult{}, igraphError("calculate graphlet candidate basis", code)
	}
	convertedCliques, convertedThresholds, err := convertAlignedGraphletValues(
		cliques, thresholds, int(C.igraph_vcount(&g.graph)), "threshold", resolved,
	)
	if err != nil {
		return GraphletsCandidateBasisResult{}, err
	}
	return GraphletsCandidateBasisResult{Cliques: convertedCliques, Thresholds: convertedThresholds}, nil
}

func (g *Graph) graphletsProject(
	cliques [][]int,
	initialMu []float64,
	weights []float64,
	niter int,
	adapters *graphletAdapters,
) ([]float64, error) {
	if g == nil {
		return nil, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return nil, ErrClosed
	}
	if err := validateGraphletIterations(niter); err != nil {
		return nil, err
	}
	canonicalCliques, err := validateGraphletCliques(&g.graph, cliques)
	if err != nil {
		return nil, err
	}
	if err := validateInitialMu(initialMu, len(canonicalCliques)); err != nil {
		return nil, err
	}
	resolved := resolvedGraphletAdapters(adapters)
	cWeights, err := newGraphletWeights(&g.graph, weights, resolved.createReal)
	if err != nil {
		return nil, err
	}
	defer resolved.closeReal(cWeights)
	if err := validateGraphletGraph(g, resolved); err != nil {
		return nil, err
	}
	if len(canonicalCliques) == 0 {
		return make([]float64, 0), nil
	}
	cCliques, err := resolved.createList(canonicalCliques)
	if err != nil {
		return nil, err
	}
	defer resolved.closeList(cCliques)
	startMu := len(initialMu) > 0
	muValues := initialMu
	if !startMu {
		muValues = nil
	}
	mu, err := resolved.createReal(muValues)
	if err != nil {
		return nil, err
	}
	defer resolved.closeReal(mu)
	if code := resolved.projectCall(g, cWeights, cCliques, mu, startMu, niter); code != int(C.IGRAPH_SUCCESS) {
		return nil, igraphError("project graphlet basis", code)
	}
	values, err := resolved.convertReal(mu)
	if err != nil {
		return nil, err
	}
	return validateGraphletValues(values, len(canonicalCliques), "coefficient")
}

func initializeGraphletOutputs(resolved graphletAdapters) (*intVectorList, *realVector, error) {
	cliques, err := resolved.initializeList()
	if err != nil {
		return nil, nil, err
	}
	values, err := resolved.initializeReal()
	if err != nil {
		resolved.closeList(cliques)
		return nil, nil, err
	}
	return cliques, values, nil
}

func convertGraphletResult(
	cliques *intVectorList,
	mu *realVector,
	vertexCount int,
	resolved graphletAdapters,
) (GraphletsResult, error) {
	convertedCliques, convertedMu, err := convertAlignedGraphletValues(
		cliques, mu, vertexCount, "coefficient", resolved,
	)
	if err != nil {
		return GraphletsResult{}, err
	}
	return GraphletsResult{Cliques: convertedCliques, Mu: convertedMu}, nil
}

func convertAlignedGraphletValues(
	cliques *intVectorList,
	values *realVector,
	vertexCount int,
	description string,
	resolved graphletAdapters,
) ([][]int, []float64, error) {
	convertedCliques, err := resolved.convertList(cliques)
	if err != nil {
		return nil, nil, err
	}
	convertedValues, err := resolved.convertReal(values)
	if err != nil {
		return nil, nil, err
	}
	canonicalCliques, err := validateConvertedGraphletCliques(convertedCliques, vertexCount)
	if err != nil {
		return nil, nil, err
	}
	convertedValues, err = validateGraphletValues(convertedValues, len(convertedCliques), description)
	if err != nil {
		return nil, nil, err
	}
	return canonicalCliques, convertedValues, nil
}

func validateGraphletIterations(niter int) error {
	if niter < 0 {
		return fmt.Errorf("igraph: graphlet iteration count must be non-negative: %d", niter)
	}
	_, err := intToIgraphInt(niter, "graphlet iteration count")
	return err
}

func validateGraphletGraph(g *Graph, resolved graphletAdapters) error {
	simple, code := resolved.simpleCall(g)
	if code != int(C.IGRAPH_SUCCESS) {
		return igraphError("validate graphlet graph shape", code)
	}
	if !simple {
		return fmt.Errorf("igraph: graphlet methods require a graph that is simple when directions are ignored")
	}
	return nil
}

func newGraphletWeights(
	graph *C.igraph_t,
	weights []float64,
	create func([]float64) (*realVector, error),
) (*realVector, error) {
	edgeCount := int(C.igraph_ecount(graph))
	values := weights
	if weights == nil {
		values = make([]float64, edgeCount)
		for index := range values {
			values[index] = 1
		}
	} else if len(weights) != edgeCount {
		return nil, fmt.Errorf(
			"igraph: graphlet weight length %d must match edge count %d",
			len(weights), edgeCount,
		)
	}
	for index, weight := range values {
		if math.IsNaN(weight) || math.IsInf(weight, 0) || weight < 0 {
			return nil, fmt.Errorf(
				"igraph: graphlet weight at index %d must be finite and non-negative: %v",
				index, weight,
			)
		}
	}
	return create(values)
}

func validateInitialMu(values []float64, cliqueCount int) error {
	if len(values) == 0 {
		return nil
	}
	if len(values) != cliqueCount {
		return fmt.Errorf(
			"igraph: initial graphlet coefficient length %d must match clique count %d",
			len(values), cliqueCount,
		)
	}
	for index, value := range values {
		if math.IsNaN(value) || math.IsInf(value, 0) || value < 0 {
			return fmt.Errorf(
				"igraph: initial graphlet coefficient at index %d must be finite and non-negative: %v",
				index, value,
			)
		}
	}
	return nil
}

func validateGraphletCliques(graph *C.igraph_t, cliques [][]int) ([][]int, error) {
	vertexCount := int(C.igraph_vcount(graph))
	canonical, err := validateConvertedGraphletCliques(cliques, vertexCount)
	if err != nil {
		return nil, err
	}
	for index, clique := range canonical {
		selector, err := VertexIDs(clique...)
		if err != nil {
			return nil, err
		}
		cSelector, err := newCVertexSelector(selector)
		if err != nil {
			return nil, err
		}
		var complete C.igraph_bool_t
		code := C.go_igraph_is_clique(graph, cSelector.value, booltoint(false), &complete)
		cSelector.close()
		if code != C.IGRAPH_SUCCESS {
			return nil, igraphError("validate graphlet clique", int(code))
		}
		if complete == booltoint(false) {
			return nil, fmt.Errorf("igraph: graphlet clique at index %d is not complete", index)
		}
	}
	return canonical, nil
}

func validateConvertedGraphletCliques(cliques [][]int, vertexCount int) ([][]int, error) {
	canonical := make([][]int, len(cliques))
	seen := make(map[string]struct{}, len(cliques))
	for index, clique := range cliques {
		if len(clique) < 2 {
			return nil, fmt.Errorf("igraph: graphlet clique at index %d must contain at least two vertices", index)
		}
		canonical[index] = append([]int(nil), clique...)
		sort.Ints(canonical[index])
		for position, vertex := range canonical[index] {
			if vertex < 0 || vertex >= vertexCount {
				return nil, fmt.Errorf(
					"igraph: graphlet clique %d vertex at position %d is %d, outside [0, %d)",
					index, position, vertex, vertexCount,
				)
			}
			if position > 0 && vertex == canonical[index][position-1] {
				return nil, fmt.Errorf("igraph: graphlet clique %d contains duplicate vertex %d", index, vertex)
			}
		}
		key := fmt.Sprint(canonical[index])
		if _, duplicate := seen[key]; duplicate {
			return nil, fmt.Errorf("igraph: graphlet clique at index %d duplicates an earlier clique", index)
		}
		seen[key] = struct{}{}
	}
	return canonical, nil
}

func validateGraphletValues(values []float64, expected int, description string) ([]float64, error) {
	if len(values) != expected {
		return nil, fmt.Errorf(
			"igraph: graphlet %s length %d must match clique count %d",
			description, len(values), expected,
		)
	}
	for index, value := range values {
		if math.IsNaN(value) || math.IsInf(value, 0) || value < 0 {
			return nil, fmt.Errorf(
				"igraph: graphlet %s at index %d must be finite and non-negative: %v",
				description, index, value,
			)
		}
	}
	return values, nil
}

func newGraphletCliqueList(cliques [][]int) (*intVectorList, error) {
	list, err := newIntVectorList()
	if err != nil {
		return nil, err
	}
	for _, clique := range cliques {
		vector, err := newIntVector(clique)
		if err != nil {
			list.close()
			return nil, err
		}
		code := C.go_igraph_vector_int_list_push_back_copy(&list.value, &vector.value)
		vector.close()
		if code != C.IGRAPH_SUCCESS {
			list.close()
			return nil, igraphError("copy graphlet clique", int(code))
		}
	}
	return list, nil
}
