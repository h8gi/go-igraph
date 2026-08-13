package igraph

// #cgo pkg-config: igraph
// #include <igraph.h>
// #include "motifs_cgo.h"
import "C"

import (
	"fmt"
	"math"
)

const maximumExactMotifCount = float64(1 << 53)

// DyadCensusResult contains the three Holland-Leinhardt dyad classes. Counts
// are Go-owned non-negative exact integers and remain valid after graph closure.
type DyadCensusResult struct {
	Mutual     int64
	Asymmetric int64
	Null       int64
}

// MotifsRandesuOptions configures RANDESU motif enumeration. Size must be 3
// or 4. CutProb is borrowed for the synchronous call; nil disables cutting,
// otherwise it must contain Size finite probabilities in [0, 1]. A non-nil
// Seed makes stochastic cutting reproducible under the package RNG lock.
type MotifsRandesuOptions struct {
	Size    int
	CutProb []float64
	Seed    *uint64
}

// MotifsRandesuEstimateOptions configures stochastic total-motif estimation.
// Exactly one sampling mode is used: AllVertices requires a positive
// SampleSize for random sampling, while any other non-empty SampleVertices
// selector supplies the sample explicitly and requires SampleSize to be zero.
// Options and selector storage are borrowed only for the synchronous call.
type MotifsRandesuEstimateOptions struct {
	Size           int
	CutProb        []float64
	SampleSize     int
	SampleVertices VertexSelector
	Seed           *uint64
}

// DyadCensus classifies every unordered vertex pair as mutual, asymmetric, or
// null. Loops and parallel-edge multiplicity are ignored. In an undirected
// graph, connected pairs are mutual and Asymmetric is always zero.
//
//igraph:bind igraph_dyad_census
func (g *Graph) DyadCensus() (DyadCensusResult, error) {
	return g.dyadCensus(nil)
}

// TriadCensus returns the Davis-Leinhardt census in its standard 16-class
// order. The non-nil slice and its exact integer counts are Go-owned and remain
// valid after graph closure. Directed graphs use directed classes; undirected
// graphs occupy the corresponding subset of classes.
//
//igraph:bind igraph_triad_census
func (g *Graph) TriadCensus() ([]int64, error) {
	return g.triadCensus(nil)
}

// AdjacentTrianglesCount returns one triangle count per selected vertex in
// materialized selector order, including duplicates. The selector is borrowed
// only for the synchronous call. Edge directions, multiplicities, and loops
// are ignored; the returned non-nil slice is Go-owned.
//
//igraph:bind igraph_count_adjacent_triangles
func (g *Graph) AdjacentTrianglesCount(vertices VertexSelector) ([]int64, error) {
	return g.adjacentTrianglesCount(vertices, nil)
}

// TrianglesCount returns the number of fully connected vertex triples. Edge
// directions, multiplicities, and loops are ignored.
//
//igraph:bind igraph_count_triangles
func (g *Graph) TrianglesCount() (int64, error) {
	return g.trianglesCount(nil)
}

// TrianglesList returns every fully connected vertex triple exactly once.
// Edge directions, multiplicities, and loops are ignored. Vertex order within
// a triple and outer result order are not compatibility promises. The non-nil
// result is Go-owned and remains valid after graph closure.
//
//igraph:bind igraph_list_triangles
func (g *Graph) TrianglesList() ([][3]int, error) {
	return g.trianglesList(nil)
}

// MotifsRandesu returns the motif-class histogram for induced subgraphs of
// size 3 or 4. Counts are finite non-negative exact integers represented as
// float64; NaN marks isomorphism classes that cannot occur for the graph's
// directedness. The non-nil result is Go-owned and survives graph closure.
//
//igraph:bind igraph_motifs_randesu
func (g *Graph) MotifsRandesu(options MotifsRandesuOptions) ([]float64, error) {
	return g.motifsRandesu(options, nil)
}

// MotifsRandesuEstimate estimates the total number of induced motifs of size
// 3 or 4 from either a random vertex sample or an explicit vertex selector.
// The estimate is a finite non-negative float64 and may be fractional.
//
//igraph:bind igraph_motifs_randesu_estimate
func (g *Graph) MotifsRandesuEstimate(options MotifsRandesuEstimateOptions) (float64, error) {
	return g.motifsRandesuEstimate(options, nil)
}

// MotifsRandesuNo returns the total number of induced motifs of size 3 or 4.
// The exact integer count is Go-owned. Options are borrowed synchronously.
//
//igraph:bind igraph_motifs_randesu_no
func (g *Graph) MotifsRandesuNo(options MotifsRandesuOptions) (int64, error) {
	return g.motifsRandesuNo(options, nil)
}

// The callback variant is deliberately not exposed across the cgo boundary;
// MotifsRandesu provides a Go-owned result instead.
//
//igraph:unsupported igraph_motifs_randesu_callback

type motifAdapters struct {
	initializeReal func() (*realVector, error)
	createReal     func([]float64) (*realVector, error)
	closeReal      func(*realVector)
	convertReal    func(*realVector) ([]float64, error)
	initializeInt  func() (*intVector, error)
	createInt      func([]int) (*intVector, error)
	closeInt       func(*intVector)
	convertInt     func(*intVector) ([]int, error)
	dyadCall       func(*Graph) ([3]float64, int)
	triadCall      func(*Graph, *realVector) int
	adjacentCall   func(*Graph, *realVector, *cVertexSelector) int
	countCall      func(*Graph) (float64, int)
	listCall       func(*Graph, *intVector) int
	randesuCall    func(*Graph, *realVector, int, *realVector) int
	estimateCall   func(*Graph, int, *realVector, int, *intVector) (float64, int)
	randesuNoCall  func(*Graph, int, *realVector) (float64, int)
}

func defaultMotifAdapters() motifAdapters {
	return motifAdapters{
		initializeReal: func() (*realVector, error) { return newRealVectorSize(0) },
		createReal:     newRealVector,
		closeReal:      func(vector *realVector) { vector.close() },
		convertReal:    func(vector *realVector) ([]float64, error) { return vector.slice() },
		initializeInt:  func() (*intVector, error) { return newIntVector(nil) },
		createInt:      newIntVector,
		closeInt:       func(vector *intVector) { vector.close() },
		convertInt:     func(vector *intVector) ([]int, error) { return vector.slice() },
		dyadCall: func(g *Graph) ([3]float64, int) {
			var mutual, asymmetric, nullDyads C.igraph_real_t
			code := C.go_igraph_dyad_census(&g.graph, &mutual, &asymmetric, &nullDyads)
			return [3]float64{float64(mutual), float64(asymmetric), float64(nullDyads)}, int(code)
		},
		triadCall: func(g *Graph, result *realVector) int {
			return int(C.go_igraph_triad_census(&g.graph, &result.value))
		},
		adjacentCall: func(g *Graph, result *realVector, vertices *cVertexSelector) int {
			return int(C.go_igraph_count_adjacent_triangles(&g.graph, &result.value, vertices.value))
		},
		countCall: func(g *Graph) (float64, int) {
			var result C.igraph_real_t
			code := C.go_igraph_count_triangles(&g.graph, &result)
			return float64(result), int(code)
		},
		listCall: func(g *Graph, result *intVector) int {
			return int(C.go_igraph_list_triangles(&g.graph, &result.value))
		},
		randesuCall: func(g *Graph, result *realVector, size int, cutProb *realVector) int {
			return int(C.go_igraph_motifs_randesu(
				&g.graph, &result.value, C.igraph_int_t(size), realVectorPointer(cutProb),
			))
		},
		estimateCall: func(g *Graph, size int, cutProb *realVector, sampleSize int, sample *intVector) (float64, int) {
			var estimate C.igraph_real_t
			code := C.go_igraph_motifs_randesu_estimate(
				&g.graph, &estimate, C.igraph_int_t(size), realVectorPointer(cutProb),
				C.igraph_int_t(sampleSize), intVectorPointer(sample),
			)
			return float64(estimate), int(code)
		},
		randesuNoCall: func(g *Graph, size int, cutProb *realVector) (float64, int) {
			var count C.igraph_real_t
			code := C.go_igraph_motifs_randesu_no(
				&g.graph, &count, C.igraph_int_t(size), realVectorPointer(cutProb),
			)
			return float64(count), int(code)
		},
	}
}

func realVectorPointer(vector *realVector) *C.igraph_vector_t {
	if vector == nil {
		return nil
	}
	return &vector.value
}

func intVectorPointer(vector *intVector) *C.igraph_vector_int_t {
	if vector == nil {
		return nil
	}
	return &vector.value
}

func resolvedMotifAdapters(adapters *motifAdapters) motifAdapters {
	if adapters == nil {
		return defaultMotifAdapters()
	}
	return *adapters
}

func (g *Graph) dyadCensus(adapters *motifAdapters) (DyadCensusResult, error) {
	if g == nil {
		return DyadCensusResult{}, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return DyadCensusResult{}, ErrClosed
	}
	resolved := resolvedMotifAdapters(adapters)
	values, code := resolved.dyadCall(g)
	if code != int(C.IGRAPH_SUCCESS) {
		return DyadCensusResult{}, igraphError("calculate dyad census", code)
	}
	converted, err := checkedMotifCounts(values[:], 3, "dyad census")
	if err != nil {
		return DyadCensusResult{}, err
	}
	return DyadCensusResult{Mutual: converted[0], Asymmetric: converted[1], Null: converted[2]}, nil
}

func (g *Graph) triadCensus(adapters *motifAdapters) ([]int64, error) {
	if g == nil {
		return nil, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return nil, ErrClosed
	}
	resolved := resolvedMotifAdapters(adapters)
	result, err := resolved.initializeReal()
	if err != nil {
		return nil, err
	}
	defer resolved.closeReal(result)
	if code := resolved.triadCall(g, result); code != int(C.IGRAPH_SUCCESS) {
		return nil, igraphError("calculate triad census", code)
	}
	values, err := resolved.convertReal(result)
	if err != nil {
		return nil, err
	}
	return checkedMotifCounts(values, 16, "triad census")
}

func (g *Graph) adjacentTrianglesCount(vertices VertexSelector, adapters *motifAdapters) ([]int64, error) {
	if g == nil {
		return nil, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return nil, ErrClosed
	}
	vertexCount := int(C.igraph_vcount(&g.graph))
	if err := validateVertexSelector(vertices, vertexCount); err != nil {
		return nil, err
	}
	vertexIDs, err := materializeVertexIDs(&g.graph, vertices)
	if err != nil {
		return nil, fmt.Errorf("igraph: materialize adjacent-triangle selector: %w", err)
	}
	selected, err := VertexIDs(vertexIDs...)
	if err != nil {
		return nil, err
	}
	cVertices, err := newCVertexSelector(selected)
	if err != nil {
		return nil, err
	}
	defer cVertices.close()
	resolved := resolvedMotifAdapters(adapters)
	result, err := resolved.initializeReal()
	if err != nil {
		return nil, err
	}
	defer resolved.closeReal(result)
	if code := resolved.adjacentCall(g, result, cVertices); code != int(C.IGRAPH_SUCCESS) {
		return nil, igraphError("count adjacent triangles", code)
	}
	values, err := resolved.convertReal(result)
	if err != nil {
		return nil, err
	}
	return checkedMotifCounts(values, len(vertexIDs), "adjacent triangle")
}

func (g *Graph) trianglesCount(adapters *motifAdapters) (int64, error) {
	if g == nil {
		return 0, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return 0, ErrClosed
	}
	resolved := resolvedMotifAdapters(adapters)
	value, code := resolved.countCall(g)
	if code != int(C.IGRAPH_SUCCESS) {
		return 0, igraphError("count triangles", code)
	}
	return checkedMotifCount(value, "triangle count")
}

func (g *Graph) trianglesList(adapters *motifAdapters) ([][3]int, error) {
	if g == nil {
		return nil, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return nil, ErrClosed
	}
	resolved := resolvedMotifAdapters(adapters)
	result, err := resolved.initializeInt()
	if err != nil {
		return nil, err
	}
	defer resolved.closeInt(result)
	if code := resolved.listCall(g, result); code != int(C.IGRAPH_SUCCESS) {
		return nil, igraphError("list triangles", code)
	}
	values, err := resolved.convertInt(result)
	if err != nil {
		return nil, err
	}
	if len(values)%3 != 0 {
		return nil, fmt.Errorf("igraph: triangle list length %d is not divisible by three", len(values))
	}
	triangles := make([][3]int, len(values)/3)
	for index := range triangles {
		copy(triangles[index][:], values[index*3:index*3+3])
	}
	return triangles, nil
}

func (g *Graph) motifsRandesu(options MotifsRandesuOptions, adapters *motifAdapters) ([]float64, error) {
	if g == nil {
		return nil, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return nil, ErrClosed
	}
	if err := validateMotifsRandesuOptions(options.Size, options.CutProb); err != nil {
		return nil, err
	}
	resolved := resolvedMotifAdapters(adapters)
	cutProb, err := newOptionalCutProbability(options.CutProb, resolved.createReal)
	if err != nil {
		return nil, err
	}
	if cutProb != nil {
		defer resolved.closeReal(cutProb)
	}
	result, err := resolved.initializeReal()
	if err != nil {
		return nil, err
	}
	defer resolved.closeReal(result)
	err = withRNG(options.Seed, func() error {
		if code := resolved.randesuCall(g, result, options.Size, cutProb); code != int(C.IGRAPH_SUCCESS) {
			return igraphError("calculate RANDESU motif histogram", code)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	values, err := resolved.convertReal(result)
	if err != nil {
		return nil, err
	}
	expectedLength := motifHistogramLength(
		C.igraph_is_directed(&g.graph) != booltoint(false), options.Size,
	)
	if len(values) != expectedLength {
		return nil, fmt.Errorf(
			"igraph: RANDESU histogram length %d does not match expected length %d",
			len(values), expectedLength,
		)
	}
	for index, value := range values {
		if math.IsNaN(value) {
			continue
		}
		if _, err := checkedMotifCount(value, fmt.Sprintf("RANDESU histogram count at index %d", index)); err != nil {
			return nil, err
		}
	}
	return values, nil
}

func motifHistogramLength(directed bool, size int) int {
	if directed {
		if size == 3 {
			return 16
		}
		return 218
	}
	if size == 3 {
		return 4
	}
	return 11
}

func (g *Graph) motifsRandesuEstimate(options MotifsRandesuEstimateOptions, adapters *motifAdapters) (float64, error) {
	if g == nil {
		return 0, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return 0, ErrClosed
	}
	if err := validateMotifsRandesuOptions(options.Size, options.CutProb); err != nil {
		return 0, err
	}
	vertexCount := int(C.igraph_vcount(&g.graph))
	sampleIDs, sampleSize, err := validateMotifSample(&g.graph, vertexCount, options)
	if err != nil {
		return 0, err
	}
	resolved := resolvedMotifAdapters(adapters)
	cutProb, err := newOptionalCutProbability(options.CutProb, resolved.createReal)
	if err != nil {
		return 0, err
	}
	if cutProb != nil {
		defer resolved.closeReal(cutProb)
	}
	var sample *intVector
	if sampleIDs != nil {
		sample, err = resolved.createInt(sampleIDs)
		if err != nil {
			return 0, err
		}
		defer resolved.closeInt(sample)
	}
	var estimate float64
	err = withRNG(options.Seed, func() error {
		var code int
		estimate, code = resolved.estimateCall(g, options.Size, cutProb, sampleSize, sample)
		if code != int(C.IGRAPH_SUCCESS) {
			return igraphError("estimate RANDESU motif count", code)
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	if math.IsNaN(estimate) || math.IsInf(estimate, 0) || estimate < 0 {
		return 0, fmt.Errorf("igraph: RANDESU motif estimate is not finite and non-negative: %g", estimate)
	}
	return estimate, nil
}

func (g *Graph) motifsRandesuNo(options MotifsRandesuOptions, adapters *motifAdapters) (int64, error) {
	if g == nil {
		return 0, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return 0, ErrClosed
	}
	if err := validateMotifsRandesuOptions(options.Size, options.CutProb); err != nil {
		return 0, err
	}
	resolved := resolvedMotifAdapters(adapters)
	cutProb, err := newOptionalCutProbability(options.CutProb, resolved.createReal)
	if err != nil {
		return 0, err
	}
	if cutProb != nil {
		defer resolved.closeReal(cutProb)
	}
	var count float64
	err = withRNG(options.Seed, func() error {
		var code int
		count, code = resolved.randesuNoCall(g, options.Size, cutProb)
		if code != int(C.IGRAPH_SUCCESS) {
			return igraphError("count RANDESU motifs", code)
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	return checkedMotifCount(count, "RANDESU total motif count")
}

func validateMotifsRandesuOptions(size int, cutProb []float64) error {
	if size != 3 && size != 4 {
		return fmt.Errorf("igraph: motif size must be 3 or 4: %d", size)
	}
	if _, err := intToIgraphInt(size, "motif size"); err != nil {
		return err
	}
	if cutProb != nil && len(cutProb) != size {
		return fmt.Errorf("igraph: cut probability length %d must match motif size %d", len(cutProb), size)
	}
	for index, probability := range cutProb {
		if math.IsNaN(probability) || math.IsInf(probability, 0) || probability < 0 || probability > 1 {
			return fmt.Errorf("igraph: cut probability at index %d must be finite and in [0, 1]: %v", index, probability)
		}
	}
	return nil
}

func newOptionalCutProbability(
	values []float64,
	create func([]float64) (*realVector, error),
) (*realVector, error) {
	if values == nil {
		return nil, nil
	}
	return create(values)
}

func validateMotifSample(
	graph *C.igraph_t,
	vertexCount int,
	options MotifsRandesuEstimateOptions,
) ([]int, int, error) {
	if options.SampleVertices.kind == vertexSelectorAll {
		if options.SampleSize <= 0 {
			return nil, 0, fmt.Errorf("igraph: RANDESU sample size must be positive: %d", options.SampleSize)
		}
		if options.SampleSize > vertexCount {
			return nil, 0, fmt.Errorf("igraph: RANDESU sample size %d exceeds vertex count %d", options.SampleSize, vertexCount)
		}
		if _, err := intToIgraphInt(options.SampleSize, "RANDESU sample size"); err != nil {
			return nil, 0, err
		}
		return nil, options.SampleSize, nil
	}
	if options.SampleSize != 0 {
		return nil, 0, fmt.Errorf("igraph: RANDESU sample size must be zero when explicit sample vertices are provided: %d", options.SampleSize)
	}
	if err := validateVertexSelector(options.SampleVertices, vertexCount); err != nil {
		return nil, 0, err
	}
	ids, err := materializeVertexIDs(graph, options.SampleVertices)
	if err != nil {
		return nil, 0, fmt.Errorf("igraph: materialize RANDESU sample selector: %w", err)
	}
	if len(ids) == 0 {
		return nil, 0, fmt.Errorf("igraph: RANDESU explicit sample must not be empty")
	}
	seen := make(map[int]struct{}, len(ids))
	for index, id := range ids {
		if _, duplicate := seen[id]; duplicate {
			return nil, 0, fmt.Errorf("igraph: RANDESU sample vertex at index %d is duplicated: %d", index, id)
		}
		seen[id] = struct{}{}
	}
	return ids, 0, nil
}

func checkedMotifCounts(values []float64, expected int, description string) ([]int64, error) {
	if len(values) != expected {
		return nil, fmt.Errorf("igraph: %s result length %d does not match expected length %d", description, len(values), expected)
	}
	result := make([]int64, len(values))
	for index, value := range values {
		converted, err := checkedMotifCount(value, fmt.Sprintf("%s count at index %d", description, index))
		if err != nil {
			return nil, err
		}
		result[index] = converted
	}
	return result, nil
}

func checkedMotifCount(value float64, description string) (int64, error) {
	if math.IsNaN(value) || math.IsInf(value, 0) || value < 0 || math.Trunc(value) != value {
		return 0, fmt.Errorf("igraph: %s is not a finite non-negative integer: %g", description, value)
	}
	if value > maximumExactMotifCount {
		return 0, fmt.Errorf("igraph: %s exceeds the exact integer range of C double: %g", description, value)
	}
	return int64(value), nil
}
