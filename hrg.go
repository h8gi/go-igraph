package igraph

// #cgo pkg-config: igraph
// #include <igraph.h>
// #include "algorithm_cgo.h"
// #include "igraph_error_cgo.h"
// static void go_hrg_set_int(igraph_vector_int_t *v, igraph_int_t i, igraph_int_t x) { VECTOR(*v)[i] = x; }
// static void go_hrg_set_real(igraph_vector_t *v, igraph_int_t i, igraph_real_t x) { VECTOR(*v)[i] = x; }
// static igraph_real_t go_hrg_get_real(const igraph_vector_t *v, igraph_int_t i) { return VECTOR(*v)[i]; }
// static igraph_error_t go_hrg_init(igraph_hrg_t *h, igraph_int_t n) { GO_IGRAPH_CALL(igraph_hrg_init(h, n)); }
// static igraph_error_t go_hrg_resize(igraph_hrg_t *h, igraph_int_t n) { GO_IGRAPH_CALL(igraph_hrg_resize(h, n)); }
// static igraph_error_t go_hrg_create(igraph_hrg_t *h, const igraph_t *g, const igraph_vector_t *p) { GO_IGRAPH_CALL(igraph_hrg_create(h, g, p)); }
// static igraph_error_t go_from_hrg(igraph_t *g, const igraph_hrg_t *h, igraph_vector_t *p) { GO_IGRAPH_CALL(igraph_from_hrg_dendrogram(g, h, p)); }
import "C"

import (
	"errors"
	"fmt"
	"math"
)

// HRGModel is an immutable, fully Go-owned hierarchical random graph model.
// A model with n leaves has n-1 internal nodes. Leaves are numbered 0 through
// n-1; internal nodes are encoded -1, -2, ... and -1 is the root. The four
// model slices are aligned by internal-node index.
type HRGModel struct {
	left, right     []int
	probabilities   []float64
	edges, vertices []int
}

// NewHRGModel validates and copies an HRG model. left, right, probabilities,
// and edges are borrowed only during this call. The zero value and four empty
// slices describe the valid one-leaf model.
func NewHRGModel(left, right []int, probabilities []float64, edges []int) (HRGModel, error) {
	nodes := len(left)
	if len(right) != nodes || len(probabilities) != nodes || len(edges) != nodes {
		return HRGModel{}, fmt.Errorf("igraph: HRG model slices have lengths %d, %d, %d, and %d", nodes, len(right), len(probabilities), len(edges))
	}
	if _, err := intToIgraphInt(nodes+1, "HRG leaf count"); err != nil {
		return HRGModel{}, err
	}
	m := HRGModel{left: append([]int{}, left...), right: append([]int{}, right...), probabilities: append([]float64{}, probabilities...), edges: append([]int{}, edges...), vertices: make([]int, nodes)}
	if nodes == 0 {
		return m, nil
	}
	parents := make([]int, nodes)
	leaves := make([]bool, nodes+1)
	for i := range parents {
		parents[i] = -1
	}
	for i := 0; i < nodes; i++ {
		if math.IsNaN(m.probabilities[i]) || math.IsInf(m.probabilities[i], 0) || m.probabilities[i] < 0 || m.probabilities[i] > 1 {
			return HRGModel{}, fmt.Errorf("igraph: HRG probability %d must be finite and in [0, 1]", i)
		}
		if m.edges[i] < 0 {
			return HRGModel{}, fmt.Errorf("igraph: HRG edge count %d is negative", i)
		}
		if _, err := intToIgraphInt(m.edges[i], fmt.Sprintf("HRG edge count %d", i)); err != nil {
			return HRGModel{}, err
		}
		for _, child := range []int{m.left[i], m.right[i]} {
			if child >= 0 {
				if child > nodes || leaves[child] {
					return HRGModel{}, fmt.Errorf("igraph: HRG node %d has invalid or repeated leaf child %d", i, child)
				}
				leaves[child] = true
			} else {
				ci := -child - 1
				if ci <= i || ci >= nodes || parents[ci] != -1 {
					return HRGModel{}, fmt.Errorf("igraph: HRG node %d has invalid or repeated internal child %d", i, child)
				}
				parents[ci] = i
			}
		}
	}
	// There are exactly 2*n child slots and exactly 2*n possible non-root
	// nodes. The range and uniqueness checks above therefore also prove that
	// every leaf and non-root internal node is reachable.
	for i := nodes - 1; i >= 0; i-- {
		count := 0
		for _, child := range []int{m.left[i], m.right[i]} {
			if child >= 0 {
				count++
			} else {
				count += m.vertices[-child-1]
			}
		}
		m.vertices[i] = count
	}
	return m, nil
}

func (m HRGModel) LeafCount() int           { return len(m.left) + 1 }
func (m HRGModel) LeftChildren() []int      { return append([]int{}, m.left...) }
func (m HRGModel) RightChildren() []int     { return append([]int{}, m.right...) }
func (m HRGModel) Probabilities() []float64 { return append([]float64{}, m.probabilities...) }
func (m HRGModel) EdgeCounts() []int        { return append([]int{}, m.edges...) }

type cHRG struct {
	value       C.igraph_hrg_t
	initialized bool
}

//igraph:internal igraph_hrg_init
func newCHRG(m HRGModel) (*cHRG, error) {
	return newCHRGWithInitializer(m, func(h *cHRG, size int) int {
		return int(C.go_hrg_init(&h.value, C.igraph_int_t(size)))
	})
}

func newCHRGWithInitializer(m HRGModel, initialize func(*cHRG, int) int) (*cHRG, error) {
	validated, err := NewHRGModel(m.left, m.right, m.probabilities, m.edges)
	if err != nil {
		return nil, err
	}
	h := &cHRG{}
	if code := initialize(h, validated.LeafCount()); code != int(C.IGRAPH_SUCCESS) {
		return nil, igraphError("initialize HRG model", code)
	}
	h.initialized = true
	for i := range validated.left {
		ci := C.igraph_int_t(i)
		C.go_hrg_set_int(&h.value.left, ci, C.igraph_int_t(validated.left[i]))
		C.go_hrg_set_int(&h.value.right, ci, C.igraph_int_t(validated.right[i]))
		C.go_hrg_set_real(&h.value.prob, ci, C.igraph_real_t(validated.probabilities[i]))
		C.go_hrg_set_int(&h.value.edges, ci, C.igraph_int_t(validated.edges[i]))
		C.go_hrg_set_int(&h.value.vertices, ci, C.igraph_int_t(validated.vertices[i]))
	}
	return h, nil
}

//igraph:internal igraph_hrg_destroy
func (h *cHRG) close() {
	if h != nil && h.initialized {
		C.igraph_hrg_destroy(&h.value)
		h.initialized = false
	}
}

//igraph:internal igraph_hrg_resize
func (h *cHRG) resize(size int) error {
	if size < 1 {
		return fmt.Errorf("igraph: HRG leaf count must be positive: %d", size)
	}
	cSize, err := intToIgraphInt(size, "HRG leaf count")
	if err != nil {
		return err
	}
	if code := C.go_hrg_resize(&h.value, cSize); code != C.IGRAPH_SUCCESS {
		return igraphError("resize HRG model", int(code))
	}
	return nil
}

//igraph:internal igraph_hrg_size
func (h *cHRG) model() (HRGModel, error) {
	return h.modelWithReaders(hrgModelReaders{
		size:  func(h *cHRG) (int, error) { return igraphIntToInt(C.igraph_hrg_size(&h.value), "HRG leaf count") },
		left:  func(h *cHRG) ([]int, error) { return intVectorSlice(&h.value.left) },
		right: func(h *cHRG) ([]int, error) { return intVectorSlice(&h.value.right) },
		edges: func(h *cHRG) ([]int, error) { return intVectorSlice(&h.value.edges) },
		probabilities: func(h *cHRG, size int) ([]float64, error) {
			values := make([]float64, size)
			for i := range values {
				values[i] = float64(C.go_hrg_get_real(&h.value.prob, C.igraph_int_t(i)))
			}
			return values, nil
		},
	})
}

type hrgModelReaders struct {
	size               func(*cHRG) (int, error)
	left, right, edges func(*cHRG) ([]int, error)
	probabilities      func(*cHRG, int) ([]float64, error)
}

func (h *cHRG) modelWithReaders(readers hrgModelReaders) (HRGModel, error) {
	n, err := readers.size(h)
	if err != nil || n < 1 {
		if err == nil {
			err = errors.New("igraph: HRG returned invalid leaf count")
		}
		return HRGModel{}, err
	}
	left, err := readers.left(h)
	if err != nil {
		return HRGModel{}, err
	}
	right, err := readers.right(h)
	if err != nil {
		return HRGModel{}, err
	}
	prob, err := readers.probabilities(h, n-1)
	if err != nil {
		return HRGModel{}, err
	}
	edges, err := readers.edges(h)
	if err != nil {
		return HRGModel{}, err
	}
	return NewHRGModel(left, right, prob, edges)
}

// NewHRGModelFromDendrogram converts a borrowed directed binary-tree graph and
// vertex-aligned probabilities to a Go-owned model. Leaf entries must be NaN;
// internal entries must be finite and in [0,1]. Pinned igraph 1.0.1 rejects a
// singleton dendrogram; models with at least two leaves are supported.
//
//igraph:bind igraph_hrg_create
func NewHRGModelFromDendrogram(graph *Graph, probabilities []float64) (HRGModel, error) {
	if graph == nil {
		return HRGModel{}, ErrClosed
	}
	graph.mu.RLock()
	defer graph.mu.RUnlock()
	if graph.closed {
		return HRGModel{}, ErrClosed
	}
	vertices := int(C.igraph_vcount(&graph.graph))
	if len(probabilities) != vertices {
		return HRGModel{}, fmt.Errorf("igraph: dendrogram has %d vertices but %d probabilities", vertices, len(probabilities))
	}
	internal := make([]float64, 0, (vertices-1)/2)
	for _, probability := range probabilities {
		if !math.IsNaN(probability) {
			internal = append(internal, probability)
		}
	}
	if len(internal) != (vertices-1)/2 {
		return HRGModel{}, errors.New("igraph: dendrogram probabilities do not identify every internal node")
	}
	p, err := newRealVector(internal)
	if err != nil {
		return HRGModel{}, err
	}
	defer p.close()
	h := &cHRG{}
	if code := C.go_hrg_init(&h.value, 0); code != C.IGRAPH_SUCCESS {
		return HRGModel{}, igraphError("initialize HRG model", int(code))
	}
	h.initialized = true
	defer h.close()
	if code := C.go_hrg_create(&h.value, &graph.graph, &p.value); code != C.IGRAPH_SUCCESS {
		return HRGModel{}, igraphError("create HRG model from dendrogram", int(code))
	}
	model, err := h.model()
	if err != nil {
		return HRGModel{}, err
	}
	// igraph 1.0.1 reconstructs topology and subtree counts but does not copy
	// the new internal-only probability argument into the model reliably.
	// Preserve the validated caller values explicitly.
	return NewHRGModel(model.left, model.right, internal, model.edges)
}

// Dendrogram returns an independently owned directed binary-tree graph and a
// vertex-ID-aligned Go-owned probability slice. Leaf probabilities are NaN.
//
//igraph:bind igraph_from_hrg_dendrogram
func (m HRGModel) Dendrogram() (*Graph, []float64, error) {
	h, err := newCHRG(m)
	if err != nil {
		return nil, nil, err
	}
	defer h.close()
	p, err := newRealVector(nil)
	if err != nil {
		return nil, nil, err
	}
	defer p.close()
	var graph C.igraph_t
	if code := C.go_from_hrg(&graph, &h.value, &p.value); code != C.IGRAPH_SUCCESS {
		return nil, nil, igraphError("convert HRG model to dendrogram", int(code))
	}
	initialized := true
	defer func() {
		if initialized {
			C.igraph_destroy(&graph)
		}
	}()
	values, err := p.slice()
	if err != nil {
		return nil, nil, err
	}
	result := adoptInitializedGraph(&graph)
	initialized = false
	return result, values, nil
}
