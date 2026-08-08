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
//
// static igraph_error_t go_igraph_add_edges(
//     igraph_t *graph, const igraph_int_t *endpoints, igraph_int_t endpoint_count) {
//   igraph_vector_int_t edges;
//   igraph_error_t err = igraph_vector_int_init_array(&edges, endpoints, endpoint_count);
//   if (err != IGRAPH_SUCCESS) {
//     return err;
//   }
//   err = igraph_add_edges(graph, &edges, NULL);
//   igraph_vector_int_destroy(&edges);
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

// Edge identifies an edge by its endpoint vertex IDs.
type Edge struct {
	From int
	To   int
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

// VertexCount returns the number of vertices in the graph.
//
//igraph:bind igraph_vcount
func (g *Graph) VertexCount() (int, error) {
	if g == nil {
		return 0, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return 0, ErrClosed
	}
	return int(C.igraph_vcount(&g.graph)), nil
}

// EdgeCount returns the number of edges in the graph.
//
//igraph:bind igraph_ecount
func (g *Graph) EdgeCount() (int, error) {
	if g == nil {
		return 0, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return 0, ErrClosed
	}
	return int(C.igraph_ecount(&g.graph)), nil
}

// IsDirected reports whether the graph is directed.
//
//igraph:bind igraph_is_directed
func (g *Graph) IsDirected() (bool, error) {
	if g == nil {
		return false, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return false, ErrClosed
	}
	return C.igraph_is_directed(&g.graph) != booltoint(false), nil
}

// EdgeEndpoints returns the source and target vertices of an edge.
// For undirected graphs, the returned order is the order stored by igraph.
//
//igraph:bind igraph_edge
func (g *Graph) EdgeEndpoints(edgeID int) (from, to int, err error) {
	if g == nil {
		return 0, 0, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return 0, 0, ErrClosed
	}

	edgeCount := int(C.igraph_ecount(&g.graph))
	if edgeID < 0 || edgeID >= edgeCount {
		return 0, 0, fmt.Errorf("igraph: edge ID %d out of range [0, %d)", edgeID, edgeCount)
	}
	var cFrom, cTo C.igraph_int_t
	if code := C.igraph_edge(&g.graph, C.igraph_int_t(edgeID), &cFrom, &cTo); code != C.IGRAPH_SUCCESS {
		return 0, 0, igraphError("get edge endpoints", int(code))
	}
	return int(cFrom), int(cTo), nil
}

// IsEmpty reports whether the graph has no vertices.
func (g *Graph) IsEmpty() (bool, error) {
	if g == nil {
		return false, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return false, ErrClosed
	}
	return C.igraph_vcount(&g.graph) == 0, nil
}

// AddVertices appends count isolated vertices to the graph. A zero count is a
// no-op. Negative counts return an error without modifying the graph.
//
//igraph:bind igraph_add_vertices
func (g *Graph) AddVertices(count int) error {
	if g == nil {
		return ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return ErrClosed
	}
	if count < 0 {
		return fmt.Errorf("igraph: vertex count must be non-negative: %d", count)
	}
	if count == 0 {
		return nil
	}
	if code := C.igraph_add_vertices(&g.graph, C.igraph_int_t(count), nil); code != C.IGRAPH_SUCCESS {
		return igraphError("add vertices", int(code))
	}
	return nil
}

// AddEdge appends one edge. Self-loops and parallel edges are allowed.
//
//igraph:bind igraph_add_edge
func (g *Graph) AddEdge(from, to int) error {
	if g == nil {
		return ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return ErrClosed
	}
	if err := validateEdge(Edge{From: from, To: to}, int(C.igraph_vcount(&g.graph)), 0); err != nil {
		return err
	}
	if code := C.igraph_add_edge(&g.graph, C.igraph_int_t(from), C.igraph_int_t(to)); code != C.IGRAPH_SUCCESS {
		return igraphError("add edge", int(code))
	}
	return nil
}

// AddEdges appends a batch of edges atomically. Self-loops and parallel edges
// are allowed, and an empty batch is a no-op. All endpoints are validated
// before the graph is modified.
//
//igraph:bind igraph_add_edges
//igraph:internal igraph_vector_int_init_array
//igraph:internal igraph_vector_int_destroy
func (g *Graph) AddEdges(edges []Edge) error {
	if g == nil {
		return ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return ErrClosed
	}
	if len(edges) == 0 {
		return nil
	}

	vertexCount := int(C.igraph_vcount(&g.graph))
	endpoints := make([]C.igraph_int_t, 0, 2*len(edges))
	for index, edge := range edges {
		if err := validateEdge(edge, vertexCount, index); err != nil {
			return err
		}
		endpoints = append(endpoints, C.igraph_int_t(edge.From), C.igraph_int_t(edge.To))
	}
	code := C.go_igraph_add_edges(&g.graph, &endpoints[0], C.igraph_int_t(len(endpoints)))
	runtime.KeepAlive(endpoints)
	if code != C.IGRAPH_SUCCESS {
		return igraphError("add edges", int(code))
	}
	return nil
}

func validateEdge(edge Edge, vertexCount, index int) error {
	if edge.From < 0 || edge.From >= vertexCount {
		return fmt.Errorf("igraph: edge %d source vertex %d out of range [0, %d)", index, edge.From, vertexCount)
	}
	if edge.To < 0 || edge.To >= vertexCount {
		return fmt.Errorf("igraph: edge %d target vertex %d out of range [0, %d)", index, edge.To, vertexCount)
	}
	return nil
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
