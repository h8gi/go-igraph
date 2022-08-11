package igraph

// #cgo pkg-config: igraph
// #include <igraph.h>
import "C"
import (
	"fmt"
	"runtime"
)

type Vector struct {
	vector C.igraph_vector_t
	size   int
}

func (v *Vector) destroy() {
	C.igraph_vector_destroy(&v.vector)
}

func NewVector(size int) *Vector {
	v := &Vector{size: size}
	runtime.SetFinalizer(v, (*Vector).destroy)

	C.igraph_vector_init(&v.vector, C.long(size))

	return v
}

func NewVectorFromSlice(s []float64) *Vector {
	// Should I use igraph_vector_view?
	v := NewVector(len(s))
	for i, f := range s {
		v.Set(i, f)
	}
	return v
}

func (v *Vector) Set(pos int, value float64) error {
	if pos > v.size-1 {
		return fmt.Errorf("Illegal access: size %d", v.size)
	}
	C.igraph_vector_set(&v.vector, C.long(pos), C.double(value))
	return nil
}

func (v *Vector) Get(pos int) (float64, error) {
	if pos > v.size-1 {
		return 0, fmt.Errorf("Illegal access: size %d", v.size)
	}
	return float64(C.igraph_vector_e(&v.vector, C.long(pos))), nil
}
