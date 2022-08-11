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

// cgo argument has Go pointer to Go pointer
// GODEBUG=cgocheck=0
// The runtime/cgo.Handle type can be used to safely pass Go values between Go and C. See the runtime/cgo package documentation for details.
func VectorView(s []float64) *Vector {
	v := &Vector{size: len(s)}
	C.igraph_vector_view(&v.vector, (*C.double)(&s[0]), C.long(len(s)))
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
