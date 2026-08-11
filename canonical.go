package igraph

/*
#cgo pkg-config: igraph
#include <igraph.h>
#include "isomorphism_cgo.h"
*/
import "C"

import (
	"fmt"
	"math/big"
	"unsafe"
)

// CanonicalGraphResult contains an independently owned canonical graph and
// the permutation from source vertex IDs to canonical vertex IDs. Graph must
// be closed by the caller.
type CanonicalGraphResult struct {
	Graph             *Graph
	SourceToCanonical []int
}

// CanonicalPermutation returns the canonical vertex labeling. The result is
// indexed by source vertex and contains its canonical vertex ID. VertexColors
// is optional; a non-nil slice is borrowed and copied for the call.
//
//igraph:bind igraph_canonical_permutation
func (g *Graph) CanonicalPermutation(vertexColors []int) ([]int, error) {
	var result []int
	err := withLockedGraphs([]*Graph{g}, func(graphs []*C.igraph_t) error {
		permutation, err := canonicalPermutationLocked(graphs[0], vertexColors)
		if err != nil {
			return err
		}
		result = permutation
		return nil
	})
	return result, err
}

// CanonicalGraph returns an independently owned canonical form and its
// source-to-canonical permutation. The result survives source closure and its
// Graph must be closed by the caller.
func (g *Graph) CanonicalGraph(vertexColors []int) (CanonicalGraphResult, error) {
	var result CanonicalGraphResult
	err := withLockedGraphs([]*Graph{g}, func(graphs []*C.igraph_t) error {
		permutation, err := canonicalPermutationLocked(graphs[0], vertexColors)
		if err != nil {
			return err
		}
		canonicalToSource, err := invertPermutation(permutation)
		if err != nil {
			return err
		}
		cPermutation, err := newIntVector(canonicalToSource)
		if err != nil {
			return err
		}
		defer cPermutation.close()
		var canonical C.igraph_t
		if code := C.go_igraph_permute_vertices(graphs[0], &canonical, &cPermutation.value); code != C.IGRAPH_SUCCESS {
			return igraphError("construct canonical graph", int(code))
		}
		result = CanonicalGraphResult{
			Graph:             adoptInitializedGraph(&canonical),
			SourceToCanonical: permutation,
		}
		return nil
	})
	return result, err
}

// AutomorphismGenerators returns a Go-owned set of zero-based source-to-source
// generator permutations. The set is not guaranteed to be minimal. A non-nil
// vertex color slice is borrowed and copied for the synchronous call.
//
//igraph:bind igraph_automorphism_group
func (g *Graph) AutomorphismGenerators(vertexColors []int) ([][]int, error) {
	result := make([][]int, 0)
	err := withLockedGraphs([]*Graph{g}, func(graphs []*C.igraph_t) error {
		colors, err := prepareCanonicalInput(graphs[0], vertexColors)
		if err != nil {
			return err
		}
		if colors != nil {
			defer colors.close()
		}
		generators, err := newIntVectorList()
		if err != nil {
			return err
		}
		defer generators.close()
		var colorPointer *C.igraph_vector_int_t
		if colors != nil {
			colorPointer = &colors.value
		}
		if code := C.go_igraph_automorphism_group(graphs[0], colorPointer, &generators.value); code != C.IGRAPH_SUCCESS {
			return igraphError("compute automorphism generators", int(code))
		}
		result, err = generators.slices()
		return err
	})
	return result, err
}

// AutomorphismGroupSize returns the exact automorphism group size. The result
// is a new Go-owned big.Int. A non-nil vertex color slice is borrowed and
// copied for the synchronous call.
//
//igraph:bind igraph_count_automorphisms_bliss
//igraph:internal igraph_free
func (g *Graph) AutomorphismGroupSize(vertexColors []int) (*big.Int, error) {
	var result *big.Int
	err := withLockedGraphs([]*Graph{g}, func(graphs []*C.igraph_t) error {
		colors, err := prepareCanonicalInput(graphs[0], vertexColors)
		if err != nil {
			return err
		}
		if colors != nil {
			defer colors.close()
		}
		var colorPointer *C.igraph_vector_int_t
		if colors != nil {
			colorPointer = &colors.value
		}
		var decimal *C.char
		if code := C.go_igraph_count_automorphisms_exact(graphs[0], colorPointer, &decimal); code != C.IGRAPH_SUCCESS {
			return igraphError("count automorphisms exactly", int(code))
		}
		if decimal == nil {
			return fmt.Errorf("igraph: exact automorphism count is missing")
		}
		defer C.go_igraph_free(unsafe.Pointer(decimal))
		value, ok := new(big.Int).SetString(C.GoString(decimal), 10)
		if !ok {
			return fmt.Errorf("igraph: invalid exact automorphism count")
		}
		result = value
		return nil
	})
	return result, err
}

//igraph:internal igraph_permute_vertices
func canonicalPermutationLocked(graph *C.igraph_t, vertexColors []int) ([]int, error) {
	colors, err := prepareCanonicalInput(graph, vertexColors)
	if err != nil {
		return nil, err
	}
	if colors != nil {
		defer colors.close()
	}
	permutation, err := newIntVector(nil)
	if err != nil {
		return nil, err
	}
	defer permutation.close()
	var colorPointer *C.igraph_vector_int_t
	if colors != nil {
		colorPointer = &colors.value
	}
	if code := C.go_igraph_canonical_permutation(graph, colorPointer, &permutation.value); code != C.IGRAPH_SUCCESS {
		return nil, igraphError("compute canonical permutation", int(code))
	}
	canonicalToSource, err := permutation.slice()
	if err != nil {
		return nil, err
	}
	return invertPermutation(canonicalToSource)
}

func invertPermutation(permutation []int) ([]int, error) {
	inverse := make([]int, len(permutation))
	seen := make([]bool, len(permutation))
	for source, target := range permutation {
		if target < 0 || target >= len(permutation) {
			return nil, fmt.Errorf("igraph: permutation value %d out of range [0, %d)", target, len(permutation))
		}
		if seen[target] {
			return nil, fmt.Errorf("igraph: permutation repeats value %d", target)
		}
		seen[target] = true
		inverse[target] = source
	}
	return inverse, nil
}

func prepareCanonicalInput(graph *C.igraph_t, vertexColors []int) (*intVector, error) {
	multiple, err := operatorGraphHasMultiple(graph)
	if err != nil {
		return nil, err
	}
	if multiple {
		return nil, fmt.Errorf("igraph: canonical and automorphism operations do not support parallel edges")
	}
	if vertexColors == nil {
		return nil, nil
	}
	vertexCount := int(C.igraph_vcount(graph))
	if len(vertexColors) != vertexCount {
		return nil, fmt.Errorf("igraph: vertex color length %d must match graph vertex count %d", len(vertexColors), vertexCount)
	}
	return newIntVector(vertexColors)
}
