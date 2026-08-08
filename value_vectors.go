package igraph

// #cgo pkg-config: igraph
// #include <stdlib.h>
// #include <igraph.h>
//
// static void go_igraph_vector_bool_set_value(
//     igraph_vector_bool_t *vector, igraph_int_t pos, igraph_bool_t value) {
//   VECTOR(*vector)[pos] = value;
// }
//
// static igraph_bool_t go_igraph_vector_bool_get_value(
//     const igraph_vector_bool_t *vector, igraph_int_t pos) {
//   return VECTOR(*vector)[pos];
// }
import "C"

import (
	"fmt"
	"strings"
	"unicode/utf8"
	"unsafe"
)

// boolVector owns an initialized igraph_vector_bool_t. Go input storage is
// borrowed only during construction and copied into C-owned memory. Nil and
// empty inputs both create a valid zero-length vector.
type boolVector struct {
	value C.igraph_vector_bool_t
}

//igraph:internal igraph_vector_bool_init
func newBoolVector(values []bool) (*boolVector, error) {
	size, err := intToIgraphInt(len(values), "boolean vector length")
	if err != nil {
		return nil, err
	}
	vector := &boolVector{}
	if code := C.igraph_vector_bool_init(&vector.value, size); code != C.IGRAPH_SUCCESS {
		return nil, igraphError("initialize boolean vector", int(code))
	}
	for i, value := range values {
		C.go_igraph_vector_bool_set_value(&vector.value, C.igraph_int_t(i), booltoint(value))
	}
	return vector, nil
}

// slice returns Go-owned storage that remains valid after close. Empty vectors
// return a non-nil empty slice.
//
//igraph:internal igraph_vector_bool_size
func (v *boolVector) slice() ([]bool, error) {
	size, err := igraphIntToInt(C.igraph_vector_bool_size(&v.value), "boolean vector length")
	if err != nil {
		return nil, err
	}
	result := make([]bool, size)
	for i := range result {
		result[i] = C.go_igraph_vector_bool_get_value(&v.value, C.igraph_int_t(i)) != booltoint(false)
	}
	return result, nil
}

//igraph:internal igraph_vector_bool_destroy
func (v *boolVector) close() {
	C.igraph_vector_bool_destroy(&v.value)
}

// stringVector owns an initialized igraph_strvector_t. Each valid Go string is
// copied by igraph during construction; no Go pointer or temporary C string is
// retained. Nil and empty inputs both create a valid zero-length vector.
type stringVector struct {
	value C.igraph_strvector_t
}

//igraph:internal igraph_strvector_init
//igraph:internal igraph_strvector_set
func newStringVector(values []string) (*stringVector, error) {
	for i, value := range values {
		if !utf8.ValidString(value) {
			return nil, fmt.Errorf("igraph: string vector value at index %d is not valid UTF-8", i)
		}
		if strings.IndexByte(value, 0) >= 0 {
			return nil, fmt.Errorf("igraph: string vector value at index %d contains an embedded NUL byte", i)
		}
	}
	size, err := intToIgraphInt(len(values), "string vector length")
	if err != nil {
		return nil, err
	}
	vector := &stringVector{}
	if code := C.igraph_strvector_init(&vector.value, size); code != C.IGRAPH_SUCCESS {
		return nil, igraphError("initialize string vector", int(code))
	}
	for i, value := range values {
		cValue := C.CString(value)
		code := C.igraph_strvector_set(&vector.value, C.igraph_int_t(i), cValue)
		C.free(unsafe.Pointer(cValue))
		if code != C.IGRAPH_SUCCESS {
			vector.close()
			return nil, igraphError("set string vector value", int(code))
		}
	}
	return vector, nil
}

// slice returns Go-owned strings that remain valid after close. igraph strings
// are UTF-8, NUL-terminated values, and empty vectors return a non-nil slice.
//
//igraph:internal igraph_strvector_size
//igraph:internal igraph_strvector_get
func (v *stringVector) slice() ([]string, error) {
	size, err := igraphIntToInt(C.igraph_strvector_size(&v.value), "string vector length")
	if err != nil {
		return nil, err
	}
	result := make([]string, size)
	for i := range result {
		result[i] = C.GoString(C.igraph_strvector_get(&v.value, C.igraph_int_t(i)))
	}
	return result, nil
}

//igraph:internal igraph_strvector_destroy
func (v *stringVector) close() {
	C.igraph_strvector_destroy(&v.value)
}
