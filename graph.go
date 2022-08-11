package igraph

// #cgo pkg-config: igraph libxml-2.0
// #include <stdio.h>
// #include <igraph.h>
import "C"
import (
	"errors"
	"os"
	"runtime"
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
	C.igraph_lattice(&g.graph, &dim.vector, C.int(nei),
		booltoint(directed), booltoint(mutual), booltoint(circular))

	return g
}

func (g *Graph) WriteEdgeList(file *os.File) error {
	fstruct := C.fdopen(C.int(file.Fd()), C.CString("w"))
	if err := C.igraph_write_graph_edgelist(&g.graph, fstruct); err != 0 {
		return errors.New("Write failed")
	}
	C.fflush(fstruct)
	return nil
}

func (g *Graph) WriteGraphML(file *os.File, prefixattr bool) error {
	fstruct := C.fdopen(C.int(file.Fd()), C.CString("w"))
	if err := C.igraph_write_graph_graphml(&g.graph, fstruct, booltoint(prefixattr)); err != 0 {
		return errors.New("Write failed")
	}
	C.fflush(fstruct)
	return nil

}
