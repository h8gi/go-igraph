package igraph

// #cgo pkg-config: igraph libxml-2.0
// #include <stdio.h>
// #include <igraph.h>
//
// static igraph_error_t go_igraph_lattice(
//     igraph_t *graph, const igraph_vector_t *dimensions, igraph_int_t nei,
//     igraph_bool_t directed, igraph_bool_t mutual, igraph_bool_t circular) {
//   igraph_int_t size = igraph_vector_size(dimensions);
//   igraph_vector_int_t integer_dimensions;
//   igraph_error_t err = igraph_vector_int_init(&integer_dimensions, size);
//   if (err != IGRAPH_SUCCESS) {
//     return err;
//   }
//   for (igraph_int_t i = 0; i < size; ++i) {
//     VECTOR(integer_dimensions)[i] = (igraph_int_t) VECTOR(*dimensions)[i];
//   }
//   igraph_vector_bool_t periodic;
//   err = igraph_vector_bool_init(&periodic, size);
//   if (err != IGRAPH_SUCCESS) {
//     igraph_vector_int_destroy(&integer_dimensions);
//     return err;
//   }
//   igraph_vector_bool_fill(&periodic, circular);
//   err = igraph_square_lattice(graph, &integer_dimensions, nei, directed, mutual, &periodic);
//   igraph_vector_bool_destroy(&periodic);
//   igraph_vector_int_destroy(&integer_dimensions);
//   return err;
// }
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
	C.go_igraph_lattice(&g.graph, &dim.vector, C.igraph_int_t(nei),
		booltoint(directed), booltoint(mutual), booltoint(circular))

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
	if err := C.igraph_write_graph_graphml(&g.graph, fstruct, booltoint(prefixattr)); err != 0 {
		return errors.New("Write failed")
	}
	C.fflush(fstruct)
	return nil

}
