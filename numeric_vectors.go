package igraph

// #cgo pkg-config: igraph
// #include <igraph.h>
// #include "algorithm_cgo.h"
//
// static void go_igraph_vector_int_set(
//     igraph_vector_int_t *vector, igraph_int_t pos, igraph_int_t value) {
//   VECTOR(*vector)[pos] = value;
// }
//
// static igraph_int_t go_igraph_vector_int_get(
//     const igraph_vector_int_t *vector, igraph_int_t pos) {
//   return VECTOR(*vector)[pos];
// }
//
// static void go_igraph_vector_set_value(
//     igraph_vector_t *vector, igraph_int_t pos, igraph_real_t value) {
//   VECTOR(*vector)[pos] = value;
// }
//
// static igraph_real_t go_igraph_vector_get_value(
//     const igraph_vector_t *vector, igraph_int_t pos) {
//   return VECTOR(*vector)[pos];
// }
import "C"

import "fmt"

// intVector owns an initialized igraph_vector_int_t. Call close on every
// successful construction. Go input storage is borrowed only for the duration
// of construction; the C vector owns a copy. Nil and empty inputs both create
// a valid zero-length vector.
type intVector struct {
	value C.igraph_vector_int_t
}

//igraph:internal igraph_vector_int_init
func newIntVector(values []int) (*intVector, error) {
	return newIntVectorWithInitializer(values, func(vector *intVector, size int) int {
		return int(C.go_igraph_vector_int_init(&vector.value, C.igraph_int_t(size)))
	})
}

func newIntVectorWithInitializer(
	values []int,
	initialize func(*intVector, int) int,
) (*intVector, error) {
	_, err := intToIgraphInt(len(values), "integer vector length")
	if err != nil {
		return nil, err
	}
	converted := make([]C.igraph_int_t, len(values))
	for i, value := range values {
		converted[i], err = intToIgraphInt(value, fmt.Sprintf("integer vector value at index %d", i))
		if err != nil {
			return nil, err
		}
	}

	vector := &intVector{}
	if code := initialize(vector, len(values)); code != int(C.IGRAPH_SUCCESS) {
		return nil, igraphError("initialize integer vector", int(code))
	}
	for i, value := range converted {
		C.go_igraph_vector_int_set(&vector.value, C.igraph_int_t(i), value)
	}
	return vector, nil
}

// slice returns Go-owned storage that remains valid after close. Empty vectors
// return a non-nil empty slice so callers never retain C-backed memory.
//
//igraph:internal igraph_vector_int_size
func (v *intVector) slice() ([]int, error) {
	return intVectorSlice(&v.value)
}

func intVectorSlice(vector *C.igraph_vector_int_t) ([]int, error) {
	size, err := igraphIntToInt(C.igraph_vector_int_size(vector), "integer vector length")
	if err != nil {
		return nil, err
	}
	result := make([]int, size)
	for i := range result {
		result[i], err = igraphIntToInt(
			C.go_igraph_vector_int_get(vector, C.igraph_int_t(i)),
			fmt.Sprintf("integer vector value at index %d", i),
		)
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}

//igraph:internal igraph_vector_int_destroy
func (v *intVector) close() {
	C.igraph_vector_int_destroy(&v.value)
}

// realVector owns an initialized igraph_vector_t. Its ownership and nil/empty
// behavior match intVector.
type realVector struct {
	value C.igraph_vector_t
}

func newRealVectorSize(size int) (*realVector, error) {
	return newRealVectorSizeWithInitializer(size, func(vector *realVector, size int) int {
		return int(C.go_igraph_vector_init(&vector.value, C.igraph_int_t(size)))
	})
}

func newRealVectorSizeWithInitializer(
	size int,
	initialize func(*realVector, int) int,
) (*realVector, error) {
	if size < 0 {
		return nil, fmt.Errorf("igraph: real vector length must be non-negative: %d", size)
	}
	_, err := intToIgraphInt(size, "real vector length")
	if err != nil {
		return nil, err
	}
	vector := &realVector{}
	if code := initialize(vector, size); code != int(C.IGRAPH_SUCCESS) {
		return nil, igraphError("initialize real vector", code)
	}
	return vector, nil
}

func newRealVector(values []float64) (*realVector, error) {
	vector, err := newRealVectorSize(len(values))
	if err != nil {
		return nil, err
	}
	for i, value := range values {
		C.go_igraph_vector_set_value(&vector.value, C.igraph_int_t(i), C.igraph_real_t(value))
	}
	return vector, nil
}

// slice returns a Go-owned, non-nil slice that remains valid after close.
//
//igraph:internal igraph_vector_size
func (v *realVector) slice() ([]float64, error) {
	size, err := igraphIntToInt(C.igraph_vector_size(&v.value), "real vector length")
	if err != nil {
		return nil, err
	}
	result := make([]float64, size)
	for i := range result {
		result[i] = float64(C.go_igraph_vector_get_value(&v.value, C.igraph_int_t(i)))
	}
	return result, nil
}

//igraph:internal igraph_vector_destroy
func (v *realVector) close() {
	C.igraph_vector_destroy(&v.value)
}

func intToIgraphInt(value int, description string) (C.igraph_int_t, error) {
	converted := C.igraph_int_t(value)
	if int(converted) != value {
		return 0, fmt.Errorf("igraph: %s is out of range: %d", description, value)
	}
	return converted, nil
}

func igraphIntToInt(value C.igraph_int_t, description string) (int, error) {
	converted := int(value)
	if C.igraph_int_t(converted) != value {
		return 0, fmt.Errorf("igraph: %s is out of Go int range", description)
	}
	return converted, nil
}
