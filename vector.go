package igraph

// #cgo pkg-config: igraph
// #include <igraph.h>
import "C"
import (
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

func (v *Vector) Set(pos int, value float64) {
	C.igraph_vector_set(&v.vector, C.long(pos), C.double(value))
}
