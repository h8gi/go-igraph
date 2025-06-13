package igraph

// #cgo pkg-config: igraph libxml-2.0
// #include <stdio.h>
// #include <igraph/igraph.h>
import "C"
import (
	"errors"
	"os"
	"runtime"
	"unsafe"
)

const (
	IGRAPH_DIRECTED   = C.IGRAPH_DIRECTED
	IGRAPH_UNDIRECTED = C.IGRAPH_UNDIRECTED
)

type Graph struct {
	graph C.igraph_t
}

func (g *Graph) destroy() {
	C.igraph_destroy(&g.graph)
}

func NewGraph() *Graph {
	g := &Graph{}
	runtime.SetFinalizer(g, (*Graph).destroy)
	return g
}

func NewLattice(dim *Vector, nei int, directed bool, mutual bool, circular bool) *Graph {
	g := NewGraph()

	var dims C.igraph_vector_int_t
	C.igraph_vector_int_init(&dims, C.igraph_integer_t(dim.size))
	for i := 0; i < dim.size; i++ {
		val, _ := dim.Get(i)
		C.igraph_vector_int_set(&dims, C.igraph_integer_t(i), C.igraph_integer_t(val))
	}
	defer C.igraph_vector_int_destroy(&dims)

	C.igraph_lattice(&g.graph, &dims, C.igraph_integer_t(nei),
		C.igraph_bool_t(booltoint(directed)), C.igraph_bool_t(booltoint(mutual)), C.igraph_bool_t(booltoint(circular)))

	return g
}

func (g *Graph) WriteEdgeList(file *os.File) error {
	mode := C.CString("w")
	defer C.free(unsafe.Pointer(mode))
	fstruct := C.fdopen(C.int(file.Fd()), mode)
	if err := C.igraph_write_graph_edgelist(&g.graph, fstruct); err != 0 {
		return errors.New("Write failed")
	}
	C.fflush(fstruct)
	return nil
}

func (g *Graph) WriteGraphML(file *os.File, prefixattr bool) error {
	mode := C.CString("w")
	defer C.free(unsafe.Pointer(mode))
	fstruct := C.fdopen(C.int(file.Fd()), mode)
	if err := C.igraph_write_graph_graphml(&g.graph, fstruct, C.igraph_bool_t(booltoint(prefixattr))); err != 0 {
		return errors.New("Write failed")
	}
	C.fflush(fstruct)
	return nil

}
