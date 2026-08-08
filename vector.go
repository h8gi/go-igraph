package igraph

// #cgo pkg-config: igraph
// #include <igraph.h>
//
// static igraph_real_t go_igraph_vector_get(const igraph_vector_t *vector, igraph_int_t pos) {
//   return VECTOR(*vector)[pos];
// }
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

	C.igraph_vector_init(&v.vector, C.igraph_int_t(size))

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

// VectorView retains the historical API name but copies the input into
// C-managed storage. Keeping a view over Go slice memory in an igraph vector
// would violate cgo's pointer-passing rules once this function returns.
func VectorView(s []float64) *Vector {
	return NewVectorFromSlice(s)
}

func (v *Vector) Set(pos int, value float64) error {
	if pos < 0 || pos >= v.size {
		return fmt.Errorf("igraph: vector index %d out of range [0, %d)", pos, v.size)
	}
	C.igraph_vector_set(&v.vector, C.igraph_int_t(pos), C.double(value))
	return nil
}

func (v *Vector) Get(pos int) (float64, error) {
	if pos < 0 || pos >= v.size {
		return 0, fmt.Errorf("igraph: vector index %d out of range [0, %d)", pos, v.size)
	}
	return float64(C.go_igraph_vector_get(&v.vector, C.igraph_int_t(pos))), nil
}
