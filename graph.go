package igraph

// #cgo pkg-config: igraph libxml-2.0
// #include <stdio.h>
// #include <unistd.h>
// #include <igraph.h>
//
// static igraph_error_t go_igraph_lattice(
//     igraph_t *graph, const igraph_int_t *dimensions, igraph_int_t dimension_count,
//     igraph_int_t nei, igraph_bool_t directed, igraph_bool_t mutual,
//     igraph_bool_t circular) {
//   igraph_vector_int_t integer_dimensions;
//   igraph_error_t err = igraph_vector_int_init(&integer_dimensions, dimension_count);
//   if (err != IGRAPH_SUCCESS) {
//     return err;
//   }
//   for (igraph_int_t i = 0; i < dimension_count; ++i) {
//     VECTOR(integer_dimensions)[i] = dimensions[i];
//   }
//   igraph_vector_bool_t periodic;
//   err = igraph_vector_bool_init(&periodic, dimension_count);
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
	"fmt"
	"os"
	"runtime"
	"sync"
	"unsafe"
)

const (
	IGRAPH_DIRECTED   = C.IGRAPH_DIRECTED
	IGRAPH_UNDIRECTED = C.IGRAPH_UNDIRECTED
)

type Graph struct {
	mu     sync.Mutex
	graph  C.igraph_t
	closed bool
}

//igraph:bind igraph_empty
func NewGraph() (*Graph, error) {
	g := &Graph{}
	if code := C.igraph_empty(&g.graph, 0, booltoint(false)); code != C.IGRAPH_SUCCESS {
		return nil, igraphError("initialize graph", int(code))
	}
	runtime.SetFinalizer(g, (*Graph).finalize)
	return g, nil
}

//igraph:bind igraph_square_lattice
func NewLattice(dimensions []int, neighbors int, directed, mutual, circular bool) (*Graph, error) {
	if len(dimensions) == 0 {
		return nil, errors.New("igraph: lattice dimensions must not be empty")
	}
	if neighbors < 0 {
		return nil, fmt.Errorf("igraph: lattice neighbor distance must be non-negative: %d", neighbors)
	}

	cDimensions := make([]C.igraph_int_t, len(dimensions))
	for i, dimension := range dimensions {
		if dimension < 0 {
			return nil, fmt.Errorf("igraph: lattice dimension %d must be non-negative: %d", i, dimension)
		}
		cDimensions[i] = C.igraph_int_t(dimension)
	}

	g := &Graph{}
	code := C.go_igraph_lattice(
		&g.graph,
		&cDimensions[0],
		C.igraph_int_t(len(cDimensions)),
		C.igraph_int_t(neighbors),
		booltoint(directed),
		booltoint(mutual),
		booltoint(circular),
	)
	runtime.KeepAlive(cDimensions)
	if code != C.IGRAPH_SUCCESS {
		return nil, igraphError("initialize lattice", int(code))
	}
	runtime.SetFinalizer(g, (*Graph).finalize)
	return g, nil
}

//igraph:bind igraph_write_graph_edgelist
func (g *Graph) WriteEdgeList(file *os.File) error {
	if g == nil {
		return ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return ErrClosed
	}

	fstruct, err := openFileStream(file)
	if err != nil {
		return err
	}
	defer C.fclose(fstruct)
	if code := C.igraph_write_graph_edgelist(&g.graph, fstruct); code != C.IGRAPH_SUCCESS {
		return igraphError("write edge list", int(code))
	}
	if C.fflush(fstruct) != 0 {
		return errors.New("igraph: failed to flush edge list")
	}
	return nil
}

//igraph:bind igraph_write_graph_graphml
func (g *Graph) WriteGraphML(file *os.File, prefixattr bool) error {
	if g == nil {
		return ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return ErrClosed
	}

	fstruct, err := openFileStream(file)
	if err != nil {
		return err
	}
	defer C.fclose(fstruct)
	if code := C.igraph_write_graph_graphml(&g.graph, fstruct, booltoint(prefixattr)); code != C.IGRAPH_SUCCESS {
		return igraphError("write GraphML", int(code))
	}
	if C.fflush(fstruct) != 0 {
		return errors.New("igraph: failed to flush GraphML")
	}
	return nil
}

//igraph:internal igraph_destroy
func (g *Graph) Close() error {
	if g == nil {
		return nil
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return nil
	}
	C.igraph_destroy(&g.graph)
	g.closed = true
	runtime.SetFinalizer(g, nil)
	return nil
}

func (g *Graph) finalize() {
	_ = g.Close()
}

func openFileStream(file *os.File) (*C.FILE, error) {
	if file == nil {
		return nil, errors.New("igraph: output file is nil")
	}

	fd := C.dup(C.int(file.Fd()))
	if fd < 0 {
		return nil, errors.New("igraph: failed to duplicate output file descriptor")
	}

	mode := C.CString("w")
	defer C.free(unsafe.Pointer(mode))
	fstruct := C.fdopen(fd, mode)
	if fstruct == nil {
		C.close(fd)
		return nil, errors.New("igraph: failed to open output stream")
	}
	return fstruct, nil
}
