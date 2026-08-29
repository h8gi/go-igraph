package igraph

/*
#include <igraph.h>
#include "random_games_cgo.h"
*/
import "C"

import (
	"fmt"
	"math"
)

// HierarchicalBlockSpec describes one top-level block. Size vertices are split
// according to Proportions and Preference gives the within-block probabilities.
type HierarchicalBlockSpec struct {
	Size        int
	Proportions []float64
	Preference  Matrix
}

type HierarchicalBlockOptions struct{ Seed *uint64 }
type PreferenceOptions struct {
	Seed            *uint64
	Directed, Loops bool
	TypeWeights     []float64
	TypeCounts      []int
}
type AsymmetricPreferenceOptions struct {
	Seed             *uint64
	Loops            bool
	TypeDistribution *Matrix
}
type IslandOptions struct{ Seed *uint64 }

// HierarchicalBlockGraphResult contains independently Go-owned graph and
// vertex-ID-aligned block and cluster assignments. Graph must be closed.
type HierarchicalBlockGraphResult struct {
	Graph            *Graph
	Blocks, Clusters []int
}
type PreferenceGraphResult struct {
	Graph *Graph
	Types []int
}
type AsymmetricPreferenceGraphResult struct {
	Graph             *Graph
	OutTypes, InTypes []int
}
type IslandGraphResult struct {
	Graph   *Graph
	Islands []int
}

func probability(value float64, name string) error {
	if math.IsNaN(value) || value < 0 || value > 1 {
		return fmt.Errorf("igraph: %s must be in [0, 1]: %g", name, value)
	}
	return nil
}

func validateProbabilityMatrix(m Matrix, rows, columns int, symmetric bool, name string) error {
	r, c := m.Dims()
	if r != rows || c != columns {
		return fmt.Errorf("igraph: %s dimensions are %dx%d, want %dx%d", name, r, c, rows, columns)
	}
	for i := 0; i < rows; i++ {
		for j := 0; j < columns; j++ {
			v, err := m.At(i, j)
			if err != nil {
				return err
			}
			if err := probability(v, fmt.Sprintf("%s[%d,%d]", name, i, j)); err != nil {
				return err
			}
			if symmetric && j < i {
				w, _ := m.At(j, i)
				if v != w {
					return fmt.Errorf("igraph: %s must be symmetric", name)
				}
			}
		}
	}
	return nil
}

func validateProportions(values []float64, size int, name string) error {
	if len(values) == 0 {
		return fmt.Errorf("igraph: %s must not be empty", name)
	}
	sum := 0.0
	for i, v := range values {
		if math.IsNaN(v) || math.IsInf(v, 0) || v < 0 {
			return fmt.Errorf("igraph: %s[%d] must be finite and non-negative", name, i)
		}
		sum += v
		x := v * float64(size)
		if math.Abs(x-math.Round(x)) > 1e-10 {
			return fmt.Errorf("igraph: %s[%d] times block size must be integral", name, i)
		}
	}
	if math.Abs(sum-1) > math.Sqrt(math.Nextafter(1, 2)-1) {
		return fmt.Errorf("igraph: %s must sum to one", name)
	}
	return nil
}

func assignments(sizes []int, proportions [][]float64) ([]int, []int) {
	blocks, clusters := []int{}, []int{}
	clusterOffset := 0
	for b, size := range sizes {
		for local, p := range proportions[b] {
			count := int(math.Round(p * float64(size)))
			for i := 0; i < count; i++ {
				blocks = append(blocks, b)
				clusters = append(clusters, clusterOffset+local)
			}
		}
		clusterOffset += len(proportions[b])
	}
	return blocks, clusters
}

// HierarchicalSBMGame samples equal-sized top-level blocks. Inputs are borrowed
// synchronously; returned assignments remain valid after Graph.Close.
//
//igraph:bind igraph_hsbm_game
func HierarchicalSBMGame(vertexCount, blockSize int, proportions []float64, preference Matrix, betweenProbability float64, options HierarchicalBlockOptions) (HierarchicalBlockGraphResult, error) {
	if vertexCount <= 0 || blockSize <= 0 || vertexCount%blockSize != 0 {
		return HierarchicalBlockGraphResult{}, fmt.Errorf("igraph: positive vertex count must be divisible by positive block size")
	}
	if err := validateConstructorSize("vertex count", vertexCount); err != nil {
		return HierarchicalBlockGraphResult{}, err
	}
	if err := probability(betweenProbability, "between-block probability"); err != nil {
		return HierarchicalBlockGraphResult{}, err
	}
	if err := validateProportions(proportions, blockSize, "cluster proportions"); err != nil {
		return HierarchicalBlockGraphResult{}, err
	}
	if err := validateProbabilityMatrix(preference, len(proportions), len(proportions), true, "preference matrix"); err != nil {
		return HierarchicalBlockGraphResult{}, err
	}
	rho, e := newRealVector(proportions)
	if e != nil {
		return HierarchicalBlockGraphResult{}, e
	}
	defer rho.close()
	cm, e := newCMatrix(preference)
	if e != nil {
		return HierarchicalBlockGraphResult{}, e
	}
	defer cm.close()
	g, e := generateGraph("igraph_hsbm_game", options.Seed, func(out *C.igraph_t) C.igraph_error_t {
		return C.go_igraph_hsbm_game(out, C.igraph_int_t(vertexCount), C.igraph_int_t(blockSize), &rho.value, &cm.value, C.igraph_real_t(betweenProbability))
	})
	if e != nil {
		return HierarchicalBlockGraphResult{}, e
	}
	sizes := make([]int, vertexCount/blockSize)
	props := make([][]float64, len(sizes))
	for i := range sizes {
		sizes[i] = blockSize
		props[i] = proportions
	}
	b, c := assignments(sizes, props)
	return HierarchicalBlockGraphResult{g, b, c}, nil
}

// HierarchicalSBMListGame samples heterogeneous top-level blocks.
//
//igraph:bind igraph_hsbm_list_game
func HierarchicalSBMListGame(specs []HierarchicalBlockSpec, betweenProbability float64, options HierarchicalBlockOptions) (HierarchicalBlockGraphResult, error) {
	if len(specs) == 0 {
		return HierarchicalBlockGraphResult{}, fmt.Errorf("igraph: block specifications must not be empty")
	}
	if err := probability(betweenProbability, "between-block probability"); err != nil {
		return HierarchicalBlockGraphResult{}, err
	}
	sizes := make([]int, len(specs))
	lengths := make([]int, len(specs))
	props := make([][]float64, len(specs))
	var rhos, mats []float64
	n := 0
	for i, s := range specs {
		if s.Size <= 0 {
			return HierarchicalBlockGraphResult{}, fmt.Errorf("igraph: block %d size must be positive", i)
		}
		if n > int(^uint(0)>>1)-s.Size {
			return HierarchicalBlockGraphResult{}, fmt.Errorf("igraph: vertex count overflows int")
		}
		n += s.Size
		sizes[i] = s.Size
		lengths[i] = len(s.Proportions)
		props[i] = s.Proportions
		if err := validateProportions(s.Proportions, s.Size, fmt.Sprintf("block %d proportions", i)); err != nil {
			return HierarchicalBlockGraphResult{}, err
		}
		if err := validateProbabilityMatrix(s.Preference, len(s.Proportions), len(s.Proportions), true, fmt.Sprintf("block %d preference matrix", i)); err != nil {
			return HierarchicalBlockGraphResult{}, err
		}
		rhos = append(rhos, s.Proportions...)
		for _, row := range s.Preference.Rows() {
			mats = append(mats, row...)
		}
	}
	if err := validateConstructorSize("vertex count", n); err != nil {
		return HierarchicalBlockGraphResult{}, err
	}
	cs, e := newIntVector(sizes)
	if e != nil {
		return HierarchicalBlockGraphResult{}, e
	}
	defer cs.close()
	cl, e := newIntVector(lengths)
	if e != nil {
		return HierarchicalBlockGraphResult{}, e
	}
	defer cl.close()
	cr, e := newRealVector(rhos)
	if e != nil {
		return HierarchicalBlockGraphResult{}, e
	}
	defer cr.close()
	cm, e := newRealVector(mats)
	if e != nil {
		return HierarchicalBlockGraphResult{}, e
	}
	defer cm.close()
	g, e := generateGraph("igraph_hsbm_list_game", options.Seed, func(out *C.igraph_t) C.igraph_error_t {
		return C.go_igraph_hsbm_list_game(out, C.igraph_int_t(n), &cs.value, &cl.value, &cr.value, &cm.value, C.igraph_real_t(betweenProbability))
	})
	if e != nil {
		return HierarchicalBlockGraphResult{}, e
	}
	b, c := assignments(sizes, props)
	return HierarchicalBlockGraphResult{g, b, c}, nil
}

// PreferenceGame samples a categorical mixing graph. TypeCounts requests fixed
// counts; TypeWeights requests sampled types; they are mutually exclusive.
//
//igraph:bind igraph_preference_game
func PreferenceGame(vertexCount int, preference Matrix, options PreferenceOptions) (PreferenceGraphResult, error) {
	if err := validateConstructorSize("vertex count", vertexCount); err != nil {
		return PreferenceGraphResult{}, err
	}
	types, cols := preference.Dims()
	if types < 1 || cols != types {
		return PreferenceGraphResult{}, fmt.Errorf("igraph: preference matrix must be non-empty and square")
	}
	if err := validateProbabilityMatrix(preference, types, types, !options.Directed, "preference matrix"); err != nil {
		return PreferenceGraphResult{}, err
	}
	if options.TypeCounts != nil && options.TypeWeights != nil {
		return PreferenceGraphResult{}, fmt.Errorf("igraph: type counts and weights are mutually exclusive")
	}
	var dist []float64
	fixed := false
	if options.TypeCounts != nil {
		if len(options.TypeCounts) != types {
			return PreferenceGraphResult{}, fmt.Errorf("igraph: type count length must match preference matrix")
		}
		sum := 0
		for i, v := range options.TypeCounts {
			if v < 0 {
				return PreferenceGraphResult{}, fmt.Errorf("igraph: type count %d is negative", i)
			}
			sum += v
			dist = append(dist, float64(v))
		}
		if sum != vertexCount {
			return PreferenceGraphResult{}, fmt.Errorf("igraph: type counts sum to %d, want %d", sum, vertexCount)
		}
		fixed = true
	} else if options.TypeWeights != nil {
		if len(options.TypeWeights) != types {
			return PreferenceGraphResult{}, fmt.Errorf("igraph: type weight length must match preference matrix")
		}
		sum := 0.0
		for i, v := range options.TypeWeights {
			if v < 0 || math.IsNaN(v) || math.IsInf(v, 0) {
				return PreferenceGraphResult{}, fmt.Errorf("igraph: type weight %d must be finite and non-negative", i)
			}
			sum += v
		}
		if !(sum > 0) {
			return PreferenceGraphResult{}, fmt.Errorf("igraph: type weights must have positive sum")
		}
		dist = options.TypeWeights
	}
	cm, e := newCMatrix(preference)
	if e != nil {
		return PreferenceGraphResult{}, e
	}
	defer cm.close()
	out, e := newIntVector(nil)
	if e != nil {
		return PreferenceGraphResult{}, e
	}
	defer out.close()
	var cd *realVector
	if dist != nil {
		cd, e = newRealVector(dist)
		if e != nil {
			return PreferenceGraphResult{}, e
		}
		defer cd.close()
	}
	var dp *C.igraph_vector_t
	if cd != nil {
		dp = &cd.value
	}
	g, e := generateGraph("igraph_preference_game", options.Seed, func(gr *C.igraph_t) C.igraph_error_t {
		return C.go_igraph_preference_game(gr, C.igraph_int_t(vertexCount), C.igraph_int_t(types), dp, booltoint(fixed), &cm.value, &out.value, booltoint(options.Directed), booltoint(options.Loops))
	})
	if e != nil {
		return PreferenceGraphResult{}, e
	}
	values, e := out.slice()
	if e != nil {
		g.Close()
		return PreferenceGraphResult{}, e
	}
	return PreferenceGraphResult{g, values}, nil
}

// AsymmetricPreferenceGame samples a directed graph with distinct outgoing and incoming types.
//
//igraph:bind igraph_asymmetric_preference_game
func AsymmetricPreferenceGame(vertexCount int, preference Matrix, options AsymmetricPreferenceOptions) (AsymmetricPreferenceGraphResult, error) {
	if err := validateConstructorSize("vertex count", vertexCount); err != nil {
		return AsymmetricPreferenceGraphResult{}, err
	}
	outN, inN := preference.Dims()
	if outN < 1 || inN < 1 {
		return AsymmetricPreferenceGraphResult{}, fmt.Errorf("igraph: preference matrix must be non-empty")
	}
	if err := validateProbabilityMatrix(preference, outN, inN, false, "preference matrix"); err != nil {
		return AsymmetricPreferenceGraphResult{}, err
	}
	cp, e := newCMatrix(preference)
	if e != nil {
		return AsymmetricPreferenceGraphResult{}, e
	}
	defer cp.close()
	var cd *cMatrix
	if options.TypeDistribution != nil {
		if r, c := options.TypeDistribution.Dims(); r != outN || c != inN {
			return AsymmetricPreferenceGraphResult{}, fmt.Errorf("igraph: type distribution dimensions must match preference matrix")
		}
		sum := 0.0
		for i, row := range options.TypeDistribution.Rows() {
			for j, v := range row {
				if v < 0 || math.IsNaN(v) || math.IsInf(v, 0) {
					return AsymmetricPreferenceGraphResult{}, fmt.Errorf("igraph: type distribution[%d,%d] must be finite and non-negative", i, j)
				}
				sum += v
			}
		}
		if !(sum > 0) {
			return AsymmetricPreferenceGraphResult{}, fmt.Errorf("igraph: type distribution must have positive sum")
		}
		cd, e = newCMatrix(*options.TypeDistribution)
		if e != nil {
			return AsymmetricPreferenceGraphResult{}, e
		}
		defer cd.close()
	}
	var dp *C.igraph_matrix_t
	if cd != nil {
		dp = &cd.value
	}
	out, e := newIntVector(nil)
	if e != nil {
		return AsymmetricPreferenceGraphResult{}, e
	}
	defer out.close()
	in, e := newIntVector(nil)
	if e != nil {
		return AsymmetricPreferenceGraphResult{}, e
	}
	defer in.close()
	g, e := generateGraph("igraph_asymmetric_preference_game", options.Seed, func(gr *C.igraph_t) C.igraph_error_t {
		return C.go_igraph_asymmetric_preference_game(gr, C.igraph_int_t(vertexCount), C.igraph_int_t(outN), C.igraph_int_t(inN), dp, &cp.value, &out.value, &in.value, booltoint(options.Loops))
	})
	if e != nil {
		return AsymmetricPreferenceGraphResult{}, e
	}
	ov, e := out.slice()
	if e != nil {
		g.Close()
		return AsymmetricPreferenceGraphResult{}, e
	}
	iv, e := in.slice()
	if e != nil {
		g.Close()
		return AsymmetricPreferenceGraphResult{}, e
	}
	return AsymmetricPreferenceGraphResult{g, ov, iv}, nil
}

// SimpleInterconnectedIslandsGame samples equal islands joined by an exact number of edges per island pair.
//
//igraph:bind igraph_simple_interconnected_islands_game
func SimpleInterconnectedIslandsGame(count, size int, insideProbability float64, interEdges int, options IslandOptions) (IslandGraphResult, error) {
	if err := validateConstructorSize("island count", count); err != nil {
		return IslandGraphResult{}, err
	}
	if err := validateConstructorSize("island size", size); err != nil {
		return IslandGraphResult{}, err
	}
	if err := validateConstructorSize("inter-island edge count", interEdges); err != nil {
		return IslandGraphResult{}, err
	}
	if err := probability(insideProbability, "inside-island probability"); err != nil {
		return IslandGraphResult{}, err
	}
	max := 0
	if size != 0 {
		if size > int(^uint(0)>>1)/size {
			return IslandGraphResult{}, fmt.Errorf("igraph: island size squared overflows int")
		}
		max = size * size
	}
	if interEdges > max {
		return IslandGraphResult{}, fmt.Errorf("igraph: inter-island edge count %d exceeds %d", interEdges, max)
	}
	if count != 0 && size > int(^uint(0)>>1)/count {
		return IslandGraphResult{}, fmt.Errorf("igraph: vertex count overflows int")
	}
	g, e := generateGraph("igraph_simple_interconnected_islands_game", options.Seed, func(gr *C.igraph_t) C.igraph_error_t {
		return C.go_igraph_simple_interconnected_islands_game(gr, C.igraph_int_t(count), C.igraph_int_t(size), C.igraph_real_t(insideProbability), C.igraph_int_t(interEdges))
	})
	if e != nil {
		return IslandGraphResult{}, e
	}
	islands := make([]int, count*size)
	for i := range islands {
		islands[i] = i / size
	}
	return IslandGraphResult{g, islands}, nil
}
