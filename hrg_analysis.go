package igraph

// #cgo pkg-config: igraph
// #include <igraph.h>
// #include "algorithm_cgo.h"
// #include "igraph_error_cgo.h"
// static igraph_error_t go_hrg_consensus(const igraph_t *g, igraph_vector_int_t *parents, igraph_vector_t *weights, igraph_hrg_t *h, igraph_bool_t start, igraph_int_t samples) { GO_IGRAPH_CALL(igraph_hrg_consensus(g, parents, weights, h, start, samples)); }
// static igraph_error_t go_hrg_predict(const igraph_t *g, igraph_vector_int_t *edges, igraph_vector_t *prob, igraph_hrg_t *h, igraph_bool_t start, igraph_int_t samples, igraph_int_t bins) { GO_IGRAPH_CALL(igraph_hrg_predict(g, edges, prob, h, start, samples, bins)); }
// static igraph_error_t go_hrg_analysis_is_simple(const igraph_t *g, igraph_bool_t *simple) { GO_IGRAPH_CALL(igraph_is_simple(g, simple, IGRAPH_UNDIRECTED)); }
import "C"

import (
	"errors"
	"fmt"
	"math"
)

// HRGAnalysisOptions controls consensus hierarchy estimation. Samples must be
// positive. Seed optionally resets the package C/igraph RNG. StartingModel is
// borrowed and copied for a warm start; nil starts from a fresh model.
type HRGAnalysisOptions struct {
	Samples       int
	Seed          *uint64
	StartingModel *HRGModel
}

// HRGPredictionOptions controls missing-edge prediction. Samples and Bins must
// both be positive. Seed and StartingModel follow HRGAnalysisOptions.
type HRGPredictionOptions struct {
	Samples       int
	Bins          int
	Seed          *uint64
	StartingModel *HRGModel
}

// HRGConsensus is a Go-owned consensus hierarchy. Original graph vertices
// occupy parent IDs 0..n-1 and larger IDs are groups. Weights contains one
// finite split occurrence count in [0, Samples] for each group, so Weights[i]
// describes consensus vertex n+i. Exactly one vertex has parent -1. Model is
// the immutable Go-owned model left by upstream analysis.
type HRGConsensus struct {
	Parents []int
	Weights []float64
	Model   HRGModel
}

// HRGPrediction contains missing edges and aligned existence probabilities in
// upstream order. All endpoint IDs refer to the input graph. Model is the
// immutable Go-owned model left by upstream analysis.
type HRGPrediction struct {
	Edges         []Edge
	Probabilities []float64
	Model         HRGModel
}

type hrgAnalysisAdapters struct {
	fresh      func(int) (*cHRG, error)
	start      func(HRGModel) (*cHRG, error)
	closeModel func(*cHRG)
	newInt     func() (*intVector, error)
	newReal    func() (*realVector, error)
	closeInt   func(*intVector)
	closeReal  func(*realVector)
	readInt    func(*intVector) ([]int, error)
	readReal   func(*realVector) ([]float64, error)
	extract    func(*cHRG) (HRGModel, error)
	consensus  func(*Graph, *intVector, *realVector, *cHRG, bool, int) error
	predict    func(*Graph, *intVector, *realVector, *cHRG, bool, int, int) error
}

func defaultHRGAnalysisAdapters() hrgAnalysisAdapters {
	return hrgAnalysisAdapters{
		fresh: newEmptyCHRG, start: newCHRG, closeModel: (*cHRG).close,
		newInt:   func() (*intVector, error) { return newIntVector(nil) },
		newReal:  func() (*realVector, error) { return newRealVector(nil) },
		closeInt: (*intVector).close, closeReal: (*realVector).close,
		readInt: (*intVector).slice, readReal: (*realVector).slice, extract: (*cHRG).model,
		consensus: func(g *Graph, parents *intVector, weights *realVector, h *cHRG, start bool, samples int) error {
			if code := C.go_hrg_consensus(&g.graph, &parents.value, &weights.value, &h.value, booltoint(start), C.igraph_int_t(samples)); code != C.IGRAPH_SUCCESS {
				return igraphError("calculate HRG consensus", int(code))
			}
			return nil
		},
		predict: func(g *Graph, edges *intVector, prob *realVector, h *cHRG, start bool, samples, bins int) error {
			if code := C.go_hrg_predict(&g.graph, &edges.value, &prob.value, &h.value, booltoint(start), C.igraph_int_t(samples), C.igraph_int_t(bins)); code != C.IGRAPH_SUCCESS {
				return igraphError("predict missing edges with HRG", int(code))
			}
			return nil
		},
	}
}

func validateHRGAnalysisGraphLocked(g *Graph, operation string) (int, error) {
	if C.igraph_is_directed(&g.graph) != booltoint(false) {
		return 0, fmt.Errorf("igraph: %s requires an undirected graph", operation)
	}
	var simple C.igraph_bool_t
	if code := C.go_hrg_analysis_is_simple(&g.graph, &simple); code != C.IGRAPH_SUCCESS {
		return 0, igraphError("check "+operation+" input simplicity", int(code))
	}
	if simple == booltoint(false) {
		return 0, fmt.Errorf("igraph: %s requires a simple graph without loops or parallel edges", operation)
	}
	return igraphIntToInt(C.igraph_vcount(&g.graph), operation+" vertex count")
}

// ConsensusHRG estimates a consensus hierarchy for an undirected simple graph.
// The graph and optional starting model are borrowed synchronously. The graph
// read lock is held before the RNG lock through extraction; returned values
// remain valid after graph closure. Equal non-nil seeds replay exactly. Empty
// and singleton graphs are rejected by pinned igraph 1.0.1. Edgeless,
// disconnected, and complete simple graphs are supported.
//
//igraph:bind igraph_hrg_consensus
func (g *Graph) ConsensusHRG(options HRGAnalysisOptions) (HRGConsensus, error) {
	return g.consensusHRG(options, nil)
}

func (g *Graph) consensusHRG(options HRGAnalysisOptions, adapters *hrgAnalysisAdapters) (HRGConsensus, error) {
	if options.Samples <= 0 {
		return HRGConsensus{}, fmt.Errorf("igraph: HRG consensus sample count must be positive: %d", options.Samples)
	}
	if _, err := intToIgraphInt(options.Samples, "HRG consensus sample count"); err != nil {
		return HRGConsensus{}, err
	}
	if g == nil {
		return HRGConsensus{}, ErrClosed
	}
	g.mu.RLock()
	defer g.mu.RUnlock()
	if g.closed {
		return HRGConsensus{}, ErrClosed
	}
	vertices, err := validateHRGAnalysisGraphLocked(g, "HRG consensus")
	if err != nil {
		return HRGConsensus{}, err
	}
	op := defaultHRGAnalysisAdapters()
	if adapters != nil {
		op = *adapters
	}
	var h *cHRG
	start := options.StartingModel != nil
	if start {
		if options.StartingModel.LeafCount() != vertices {
			return HRGConsensus{}, fmt.Errorf("igraph: HRG starting model has %d leaves, want %d", options.StartingModel.LeafCount(), vertices)
		}
		h, err = op.start(*options.StartingModel)
	} else {
		h, err = op.fresh(vertices)
	}
	if err != nil {
		return HRGConsensus{}, err
	}
	defer op.closeModel(h)
	parents, err := op.newInt()
	if err != nil {
		return HRGConsensus{}, err
	}
	defer op.closeInt(parents)
	weights, err := op.newReal()
	if err != nil {
		return HRGConsensus{}, err
	}
	defer op.closeReal(weights)
	var result HRGConsensus
	err = withRNG(options.Seed, func() error {
		if err := op.consensus(g, parents, weights, h, start, options.Samples); err != nil {
			return err
		}
		result.Parents, err = op.readInt(parents)
		if err != nil {
			return err
		}
		result.Weights, err = op.readReal(weights)
		if err != nil {
			return err
		}
		result.Model, err = op.extract(h)
		return err
	})
	if err != nil {
		return HRGConsensus{}, err
	}
	if err := validateHRGConsensus(result, vertices, options.Samples); err != nil {
		return HRGConsensus{}, err
	}
	return result, nil
}

func validateHRGConsensus(result HRGConsensus, vertexCount, samples int) error {
	if result.Parents == nil || result.Weights == nil {
		return errors.New("igraph: HRG consensus returned nil output storage")
	}
	if len(result.Parents) < vertexCount || len(result.Weights) != len(result.Parents)-vertexCount {
		return fmt.Errorf("igraph: HRG consensus returned %d parents and %d group weights for %d graph vertices", len(result.Parents), len(result.Weights), vertexCount)
	}
	roots := 0
	for id, parent := range result.Parents {
		if parent == -1 {
			roots++
		} else if parent < 0 || parent >= len(result.Parents) || parent == id {
			return fmt.Errorf("igraph: HRG consensus vertex %d has invalid parent %d", id, parent)
		}
		if id < vertexCount {
			continue
		}
		weight := result.Weights[id-vertexCount]
		if math.IsNaN(weight) || math.IsInf(weight, 0) || weight < 0 || weight > float64(samples) {
			return fmt.Errorf("igraph: HRG consensus vertex %d has invalid weight %v", id, weight)
		}
	}
	if roots != 1 {
		return fmt.Errorf("igraph: HRG consensus has %d roots, want 1", roots)
	}
	for start := range result.Parents {
		seen := make(map[int]struct{})
		for id := start; id != -1; id = result.Parents[id] {
			if _, ok := seen[id]; ok {
				return fmt.Errorf("igraph: HRG consensus contains a parent cycle at %d", id)
			}
			seen[id] = struct{}{}
		}
	}
	return nil
}

// PredictHRG estimates probabilities for missing edges of an undirected simple
// graph. Results are Go-owned and ordered exactly as pinned upstream returns
// them; endpoint pairs and probabilities are explicitly aligned. Complete
// graphs return non-nil empty edge and probability slices. Empty and singleton
// graphs are rejected by pinned igraph 1.0.1; edgeless and disconnected simple
// graphs are supported.
//
//igraph:bind igraph_hrg_predict
func (g *Graph) PredictHRG(options HRGPredictionOptions) (HRGPrediction, error) {
	return g.predictHRG(options, nil)
}

func (g *Graph) predictHRG(options HRGPredictionOptions, adapters *hrgAnalysisAdapters) (HRGPrediction, error) {
	if options.Samples <= 0 {
		return HRGPrediction{}, fmt.Errorf("igraph: HRG prediction sample count must be positive: %d", options.Samples)
	}
	if options.Bins <= 0 {
		return HRGPrediction{}, fmt.Errorf("igraph: HRG prediction bin count must be positive: %d", options.Bins)
	}
	if _, err := intToIgraphInt(options.Samples, "HRG prediction sample count"); err != nil {
		return HRGPrediction{}, err
	}
	if _, err := intToIgraphInt(options.Bins, "HRG prediction bin count"); err != nil {
		return HRGPrediction{}, err
	}
	if g == nil {
		return HRGPrediction{}, ErrClosed
	}
	g.mu.RLock()
	defer g.mu.RUnlock()
	if g.closed {
		return HRGPrediction{}, ErrClosed
	}
	vertices, err := validateHRGAnalysisGraphLocked(g, "HRG prediction")
	if err != nil {
		return HRGPrediction{}, err
	}
	op := defaultHRGAnalysisAdapters()
	if adapters != nil {
		op = *adapters
	}
	var h *cHRG
	start := options.StartingModel != nil
	if start {
		if options.StartingModel.LeafCount() != vertices {
			return HRGPrediction{}, fmt.Errorf("igraph: HRG starting model has %d leaves, want %d", options.StartingModel.LeafCount(), vertices)
		}
		h, err = op.start(*options.StartingModel)
	} else {
		h, err = op.fresh(vertices)
	}
	if err != nil {
		return HRGPrediction{}, err
	}
	defer op.closeModel(h)
	edges, err := op.newInt()
	if err != nil {
		return HRGPrediction{}, err
	}
	defer op.closeInt(edges)
	prob, err := op.newReal()
	if err != nil {
		return HRGPrediction{}, err
	}
	defer op.closeReal(prob)
	var result HRGPrediction
	err = withRNG(options.Seed, func() error {
		if err := op.predict(g, edges, prob, h, start, options.Samples, options.Bins); err != nil {
			return err
		}
		endpoints, convertErr := op.readInt(edges)
		if convertErr != nil {
			return convertErr
		}
		result.Probabilities, convertErr = op.readReal(prob)
		if convertErr != nil {
			return convertErr
		}
		if len(endpoints)%2 != 0 {
			return errors.New("igraph: HRG prediction returned an odd endpoint count")
		}
		result.Edges = make([]Edge, len(endpoints)/2)
		for i := range result.Edges {
			result.Edges[i] = Edge{From: endpoints[2*i], To: endpoints[2*i+1]}
		}
		result.Model, convertErr = op.extract(h)
		return convertErr
	})
	if err != nil {
		return HRGPrediction{}, err
	}
	if len(result.Edges) != len(result.Probabilities) {
		return HRGPrediction{}, fmt.Errorf("igraph: HRG prediction returned %d edges and %d probabilities", len(result.Edges), len(result.Probabilities))
	}
	for i, edge := range result.Edges {
		if err := validateEdge(edge, vertices, i); err != nil {
			return HRGPrediction{}, err
		}
		if edge.From == edge.To {
			return HRGPrediction{}, fmt.Errorf("igraph: HRG prediction %d is a loop", i)
		}
		p := result.Probabilities[i]
		if math.IsNaN(p) || math.IsInf(p, 0) || p < 0 || p > 1 {
			return HRGPrediction{}, fmt.Errorf("igraph: HRG prediction %d has invalid probability %v", i, p)
		}
	}
	if result.Edges == nil {
		result.Edges = []Edge{}
	}
	if result.Probabilities == nil {
		result.Probabilities = []float64{}
	}
	return result, nil
}
