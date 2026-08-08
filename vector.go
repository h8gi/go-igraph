package igraph

// #cgo pkg-config: igraph
// #include <igraph.h>
//
// static igraph_real_t go_igraph_vector_get(const igraph_vector_t *vector, igraph_int_t pos) {
//   return VECTOR(*vector)[pos];
// }
import "C"
import (
	"errors"
	"fmt"
	"runtime"
	"sync"
)

var ErrClosed = errors.New("igraph: resource is closed")

type Vector struct {
	mu     sync.Mutex
	vector C.igraph_vector_t
	size   int
	closed bool
}

//igraph:bind igraph_vector_init
func NewVector(size int) (*Vector, error) {
	if size < 0 {
		return nil, fmt.Errorf("igraph: vector size must be non-negative: %d", size)
	}

	v := &Vector{size: size}
	if code := C.igraph_vector_init(&v.vector, C.igraph_int_t(size)); code != C.IGRAPH_SUCCESS {
		return nil, igraphError("initialize vector", int(code))
	}
	runtime.SetFinalizer(v, (*Vector).finalize)
	return v, nil
}

func NewVectorFromSlice(values []float64) (*Vector, error) {
	v, err := NewVector(len(values))
	if err != nil {
		return nil, err
	}
	for i, value := range values {
		C.igraph_vector_set(&v.vector, C.igraph_int_t(i), C.double(value))
	}
	return v, nil
}

//igraph:bind igraph_vector_set
func (v *Vector) Set(pos int, value float64) error {
	if v == nil {
		return ErrClosed
	}
	v.mu.Lock()
	defer v.mu.Unlock()
	if v.closed {
		return ErrClosed
	}
	if pos < 0 || pos >= v.size {
		return fmt.Errorf("igraph: vector index %d out of range [0, %d)", pos, v.size)
	}
	C.igraph_vector_set(&v.vector, C.igraph_int_t(pos), C.double(value))
	return nil
}

//igraph:internal igraph_vector_size
func (v *Vector) Get(pos int) (float64, error) {
	if v == nil {
		return 0, ErrClosed
	}
	v.mu.Lock()
	defer v.mu.Unlock()
	if v.closed {
		return 0, ErrClosed
	}
	if pos < 0 || pos >= v.size {
		return 0, fmt.Errorf("igraph: vector index %d out of range [0, %d)", pos, v.size)
	}
	return float64(C.go_igraph_vector_get(&v.vector, C.igraph_int_t(pos))), nil
}

//igraph:internal igraph_vector_destroy
func (v *Vector) Close() error {
	if v == nil {
		return nil
	}
	v.mu.Lock()
	defer v.mu.Unlock()
	if v.closed {
		return nil
	}
	C.igraph_vector_destroy(&v.vector)
	v.closed = true
	runtime.SetFinalizer(v, nil)
	return nil
}

func (v *Vector) finalize() {
	_ = v.Close()
}
