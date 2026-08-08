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

// ErrClosed is returned when an operation requires a live C resource but its
// receiver is nil or has already been closed.
var ErrClosed = errors.New("igraph: resource is closed")

// Vector is a mutable real vector that owns a C resource. Call Close when the
// vector is no longer needed. Its methods are safe for concurrent use.
type Vector struct {
	mu     sync.Mutex
	vector *realVector
	size   int
	closed bool
}

// NewVector allocates a zero-filled vector with size elements.
//
//igraph:bind igraph_vector_init
func NewVector(size int) (*Vector, error) {
	if size < 0 {
		return nil, fmt.Errorf("igraph: vector size must be non-negative: %d", size)
	}

	vector, err := newRealVectorSize(size)
	if err != nil {
		return nil, err
	}
	v := &Vector{vector: vector, size: size}
	runtime.SetFinalizer(v, (*Vector).finalize)
	return v, nil
}

// NewVectorFromSlice copies values into a newly owned vector. Nil and empty
// inputs both create a valid zero-length vector.
func NewVectorFromSlice(values []float64) (*Vector, error) {
	vector, err := newRealVector(values)
	if err != nil {
		return nil, err
	}
	v := &Vector{vector: vector, size: len(values)}
	runtime.SetFinalizer(v, (*Vector).finalize)
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
	C.igraph_vector_set(&v.vector.value, C.igraph_int_t(pos), C.double(value))
	return nil
}

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
	return float64(C.go_igraph_vector_get(&v.vector.value, C.igraph_int_t(pos))), nil
}

// Close releases the vector's C resource. It is safe to call more than once or
// on a nil receiver.
func (v *Vector) Close() error {
	if v == nil {
		return nil
	}
	v.mu.Lock()
	defer v.mu.Unlock()
	if v.closed {
		return nil
	}
	v.vector.close()
	v.vector = nil
	v.closed = true
	runtime.SetFinalizer(v, nil)
	return nil
}

func (v *Vector) finalize() {
	_ = v.Close()
}
