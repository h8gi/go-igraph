package igraph

// #cgo pkg-config: igraph
// #include <igraph.h>
// #include "sequence_families_cgo.h"
import "C"

import "fmt"

// NewHexagonalLattice constructs a hexagonal lattice from one to three shape
// dimensions. A zero dimension produces an empty graph. Coordinates are
// ordered lexicographically with the second coordinate most significant.
// Directed edges point from lower to higher vertex IDs unless mutual is true.
// Dimensions is borrowed only for the call.
//
//igraph:bind igraph_hexagonal_lattice
func NewHexagonalLattice(dimensions []int, directed, mutual bool) (*Graph, error) {
	return newPlanarLattice(dimensions, directed, mutual, false)
}

// NewTriangularLattice constructs the planar dual shape corresponding to
// NewHexagonalLattice. Its dimension, coordinate, direction, ownership, and
// empty-graph semantics are identical.
//
//igraph:bind igraph_triangular_lattice
func NewTriangularLattice(dimensions []int, directed, mutual bool) (*Graph, error) {
	return newPlanarLattice(dimensions, directed, mutual, true)
}

func newPlanarLattice(dimensions []int, directed, mutual, triangular bool) (*Graph, error) {
	if len(dimensions) < 1 || len(dimensions) > 3 {
		return nil, fmt.Errorf("igraph: lattice requires one to three dimensions: %d", len(dimensions))
	}
	for index, dimension := range dimensions {
		if dimension < 0 {
			return nil, fmt.Errorf("igraph: lattice dimension %d must be non-negative: %d", index, dimension)
		}
	}
	vector, err := newIntVector(dimensions)
	if err != nil {
		return nil, err
	}
	defer vector.close()
	var graph C.igraph_t
	var code C.igraph_error_t
	if triangular {
		code = C.go_igraph_triangular_lattice(&graph, &vector.value, booltoint(directed), booltoint(mutual))
	} else {
		code = C.go_igraph_hexagonal_lattice(&graph, &vector.value, booltoint(directed), booltoint(mutual))
	}
	if code != C.IGRAPH_SUCCESS {
		return nil, igraphError("construct planar lattice", int(code))
	}
	return adoptInitializedGraph(&graph), nil
}

func checkedPower(base, exponent int, name string) (int, error) {
	maximum := int(^uint(0) >> 1)
	result := 1
	for exponent > 0 {
		if exponent&1 != 0 {
			if base != 0 && result > maximum/base {
				return 0, fmt.Errorf("igraph: %s size overflows int", name)
			}
			result *= base
		}
		exponent >>= 1
		if exponent > 0 {
			if base != 0 && base > maximum/base {
				return 0, fmt.Errorf("igraph: %s size overflows int", name)
			}
			base *= base
		}
	}
	return result, nil
}

// NewDeBruijn constructs the directed de Bruijn graph for strings of length
// stringLength over alphabetSize symbols. Vertex IDs encode strings in
// lexicographic base-alphabetSize order. Length zero produces one vertex;
// alphabet size zero with positive length produces an empty graph.
//
//igraph:bind igraph_de_bruijn
func NewDeBruijn(alphabetSize, stringLength int) (*Graph, error) {
	if alphabetSize < 0 || stringLength < 0 {
		return nil, fmt.Errorf("igraph: de Bruijn parameters must be non-negative")
	}
	if err := validateConstructorSize("de Bruijn alphabet size", alphabetSize); err != nil {
		return nil, err
	}
	if err := validateConstructorSize("de Bruijn string length", stringLength); err != nil {
		return nil, err
	}
	vertices, err := checkedPower(alphabetSize, stringLength, "de Bruijn graph")
	if err != nil {
		return nil, err
	}
	if alphabetSize != 0 && vertices > int(^uint(0)>>1)/alphabetSize/2 {
		return nil, fmt.Errorf("igraph: de Bruijn edge capacity overflows int")
	}
	var graph C.igraph_t
	code := C.go_igraph_de_bruijn(&graph, C.igraph_int_t(alphabetSize), C.igraph_int_t(stringLength))
	if code != C.IGRAPH_SUCCESS {
		return nil, igraphError("construct de Bruijn graph", int(code))
	}
	return adoptInitializedGraph(&graph), nil
}

// NewKautz constructs the directed Kautz graph whose alphabet has degree+1
// symbols and whose labels have order+1 characters. Adjacent characters differ.
// Vertex IDs follow pinned igraph's lexicographic label ordering.
//
//igraph:bind igraph_kautz
func NewKautz(degree, order int) (*Graph, error) {
	if degree < 0 || order < 0 {
		return nil, fmt.Errorf("igraph: Kautz parameters must be non-negative")
	}
	if err := validateConstructorSize("Kautz degree", degree); err != nil {
		return nil, err
	}
	if err := validateConstructorSize("Kautz order", order); err != nil {
		return nil, err
	}
	power, err := checkedPower(degree, order, "Kautz graph")
	if err != nil {
		return nil, err
	}
	maximum := int(^uint(0) >> 1)
	if degree == maximum || power > maximum/(degree+1) {
		return nil, fmt.Errorf("igraph: Kautz vertex count overflows int")
	}
	vertices := power * (degree + 1)
	if degree != 0 && vertices > maximum/degree/2 {
		return nil, fmt.Errorf("igraph: Kautz edge capacity overflows int")
	}
	var graph C.igraph_t
	code := C.go_igraph_kautz(&graph, C.igraph_int_t(degree), C.igraph_int_t(order))
	if code != C.IGRAPH_SUCCESS {
		return nil, igraphError("construct Kautz graph", int(code))
	}
	return adoptInitializedGraph(&graph), nil
}

// NewLCF constructs an undirected Hamiltonian graph from typed LCF notation.
// Vertices first form the cycle 0..vertexCount-1. Each shift is repeated in
// caller order repeats times and interpreted modulo vertexCount. Duplicate
// edges and loops are simplified. Shifts is borrowed only for the call.
//
//igraph:bind igraph_lcf
func NewLCF(vertexCount int, shifts []int, repeats int) (*Graph, error) {
	if err := validateConstructorSize("LCF vertex count", vertexCount); err != nil {
		return nil, err
	}
	if repeats < 0 {
		return nil, fmt.Errorf("igraph: LCF repeats must be non-negative: %d", repeats)
	}
	if err := validateConstructorSize("LCF repeats", repeats); err != nil {
		return nil, err
	}
	if vertexCount == 0 && len(shifts) != 0 && repeats != 0 {
		return nil, fmt.Errorf("igraph: empty LCF graph cannot contain shifts")
	}
	maximum := int(^uint(0) >> 1)
	if repeats != 0 && len(shifts) > (maximum-vertexCount)/repeats {
		return nil, fmt.Errorf("igraph: LCF edge capacity overflows int")
	}
	vector, err := newIntVector(shifts)
	if err != nil {
		return nil, err
	}
	defer vector.close()
	var graph C.igraph_t
	code := C.go_igraph_lcf(&graph, C.igraph_int_t(vertexCount), &vector.value, C.igraph_int_t(repeats))
	if code != C.IGRAPH_SUCCESS {
		return nil, igraphError("construct LCF graph", int(code))
	}
	return adoptInitializedGraph(&graph), nil
}
