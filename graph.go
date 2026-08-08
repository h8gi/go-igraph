package igraph

// #cgo pkg-config: igraph libxml-2.0
// #include <stdio.h>
// #include <unistd.h>
// #include <igraph.h>
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

// DirectionMode controls how edge directions are interpreted by topology
// queries. On undirected graphs all modes are equivalent.
type DirectionMode uint8

const (
	DirectionOut DirectionMode = iota
	DirectionIn
	DirectionAll
)

func (mode DirectionMode) cValue() (C.igraph_neimode_t, error) {
	switch mode {
	case DirectionOut:
		return C.IGRAPH_OUT, nil
	case DirectionIn:
		return C.IGRAPH_IN, nil
	case DirectionAll:
		return C.IGRAPH_ALL, nil
	default:
		return 0, fmt.Errorf("igraph: invalid direction mode: %d", mode)
	}
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

// NewGraphFromEdges constructs a graph with exactly vertexCount vertices.
// Vertex IDs are zero-based and must be less than vertexCount. The vertex
// count is never inferred from edges, so isolated vertices and empty graphs
// are represented by the explicit vertexCount argument. Self-loops and
// parallel edges are allowed.
//
//igraph:bind igraph_create
func NewGraphFromEdges(vertexCount int, edges []Edge, directed bool) (*Graph, error) {
	if vertexCount < 0 {
		return nil, fmt.Errorf("igraph: vertex count must be non-negative: %d", vertexCount)
	}
	if int(C.igraph_int_t(vertexCount)) != vertexCount {
		return nil, fmt.Errorf("igraph: vertex count is too large: %d", vertexCount)
	}
	if len(edges) > int(^uint(0)>>1)/2 {
		return nil, errors.New("igraph: edge list is too large")
	}

	endpoints := make([]int, 0, 2*len(edges))
	for index, edge := range edges {
		if err := validateEdge(edge, vertexCount, index); err != nil {
			return nil, err
		}
		endpoints = append(endpoints, edge.From, edge.To)
	}
	cEndpoints, err := newIntVector(endpoints)
	if err != nil {
		return nil, err
	}
	defer cEndpoints.close()

	g := &Graph{}
	code := C.igraph_create(
		&g.graph,
		&cEndpoints.value,
		C.igraph_int_t(vertexCount),
		booltoint(directed),
	)
	if code != C.IGRAPH_SUCCESS {
		return nil, igraphError("create graph from edges", int(code))
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

	for i, dimension := range dimensions {
		if dimension < 0 {
			return nil, fmt.Errorf("igraph: lattice dimension %d must be non-negative: %d", i, dimension)
		}
	}
	cDimensions, err := newIntVector(dimensions)
	if err != nil {
		return nil, err
	}
	defer cDimensions.close()
	periodicValues := make([]bool, len(dimensions))
	for i := range periodicValues {
		periodicValues[i] = circular
	}
	periodic, err := newBoolVector(periodicValues)
	if err != nil {
		return nil, err
	}
	defer periodic.close()

	g := &Graph{}
	code := C.igraph_square_lattice(
		&g.graph,
		&cDimensions.value,
		C.igraph_int_t(neighbors),
		booltoint(directed),
		booltoint(mutual),
		&periodic.value,
	)
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

// Clone returns an independently owned copy of the graph. Closing or mutating
// either graph does not affect the other.
//
//igraph:bind igraph_copy
func (g *Graph) Clone() (*Graph, error) {
	if g == nil {
		return nil, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return nil, ErrClosed
	}

	clone := &Graph{}
	if code := C.igraph_copy(&clone.graph, &g.graph); code != C.IGRAPH_SUCCESS {
		return nil, igraphError("clone graph", int(code))
	}
	runtime.SetFinalizer(clone, (*Graph).finalize)
	return clone, nil
}

// Neighbors returns adjacent vertex IDs in the requested direction. Parallel
// edges produce repeated IDs.
//
//igraph:bind igraph_neighbors
//igraph:internal igraph_vector_int_init
//igraph:internal igraph_vector_int_size
//igraph:internal igraph_vector_int_destroy
func (g *Graph) Neighbors(vertex int, mode DirectionMode) ([]int, error) {
	if g == nil {
		return nil, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return nil, ErrClosed
	}
	cMode, err := mode.cValue()
	if err != nil {
		return nil, err
	}
	if err := validateVertexID(vertex, int(C.igraph_vcount(&g.graph))); err != nil {
		return nil, err
	}
	result, err := newIntVector(nil)
	if err != nil {
		return nil, err
	}
	defer result.close()
	if code := C.igraph_neighbors(&g.graph, &result.value, C.igraph_int_t(vertex), cMode, C.IGRAPH_LOOPS_ONCE, booltoint(true)); code != C.IGRAPH_SUCCESS {
		return nil, igraphError("get neighbors", int(code))
	}
	return result.slice()
}

// IncidentEdges returns incident edge IDs in the requested direction. Loops
// are included once with DirectionAll.
//
//igraph:bind igraph_incident
func (g *Graph) IncidentEdges(vertex int, mode DirectionMode) ([]int, error) {
	if g == nil {
		return nil, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return nil, ErrClosed
	}
	cMode, err := mode.cValue()
	if err != nil {
		return nil, err
	}
	if err := validateVertexID(vertex, int(C.igraph_vcount(&g.graph))); err != nil {
		return nil, err
	}
	result, err := newIntVector(nil)
	if err != nil {
		return nil, err
	}
	defer result.close()
	if code := C.igraph_incident(&g.graph, &result.value, C.igraph_int_t(vertex), cMode, C.IGRAPH_LOOPS_ONCE); code != C.IGRAPH_SUCCESS {
		return nil, igraphError("get incident edges", int(code))
	}
	return result.slice()
}

// AreAdjacent reports whether an edge exists from the first vertex to the
// second. Undirected graphs ignore endpoint order.
//
//igraph:bind igraph_are_adjacent
func (g *Graph) AreAdjacent(from, to int) (bool, error) {
	if g == nil {
		return false, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return false, ErrClosed
	}
	vertexCount := int(C.igraph_vcount(&g.graph))
	if err := validateVertexID(from, vertexCount); err != nil {
		return false, err
	}
	if err := validateVertexID(to, vertexCount); err != nil {
		return false, err
	}
	var result C.igraph_bool_t
	if code := C.igraph_are_adjacent(&g.graph, C.igraph_int_t(from), C.igraph_int_t(to), &result); code != C.IGRAPH_SUCCESS {
		return false, igraphError("check adjacency", int(code))
	}
	return result != booltoint(false), nil
}

// EdgeID finds one edge between two vertices. Endpoint order is significant
// when directed is true. found is false when no such edge exists.
//
//igraph:bind igraph_get_eid
func (g *Graph) EdgeID(from, to int, directed bool) (edgeID int, found bool, err error) {
	if g == nil {
		return 0, false, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return 0, false, ErrClosed
	}
	vertexCount := int(C.igraph_vcount(&g.graph))
	if err := validateVertexID(from, vertexCount); err != nil {
		return 0, false, err
	}
	if err := validateVertexID(to, vertexCount); err != nil {
		return 0, false, err
	}
	var result C.igraph_integer_t
	if code := C.igraph_get_eid(&g.graph, &result, C.igraph_int_t(from), C.igraph_int_t(to), booltoint(directed), booltoint(false)); code != C.IGRAPH_SUCCESS {
		return 0, false, igraphError("find edge", int(code))
	}
	if result < 0 {
		return 0, false, nil
	}
	return int(result), true, nil
}

// Edges returns all edges in edge ID order.
//
//igraph:bind igraph_get_edgelist
func (g *Graph) Edges() ([]Edge, error) {
	if g == nil {
		return nil, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return nil, ErrClosed
	}
	result, err := newIntVector(nil)
	if err != nil {
		return nil, err
	}
	defer result.close()
	if code := C.igraph_get_edgelist(&g.graph, &result.value, booltoint(false)); code != C.IGRAPH_SUCCESS {
		return nil, igraphError("get edge list", int(code))
	}
	values, err := result.slice()
	if err != nil {
		return nil, err
	}
	edges := make([]Edge, len(values)/2)
	for i := range edges {
		edges[i] = Edge{From: values[2*i], To: values[2*i+1]}
	}
	return edges, nil
}

func validateVertexID(vertex, vertexCount int) error {
	if vertex < 0 || vertex >= vertexCount {
		return fmt.Errorf("igraph: vertex ID %d out of range [0, %d)", vertex, vertexCount)
	}
	return nil
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
	endpoints := make([]int, 0, 2*len(edges))
	for index, edge := range edges {
		if err := validateEdge(edge, vertexCount, index); err != nil {
			return err
		}
		endpoints = append(endpoints, edge.From, edge.To)
	}
	cEndpoints, err := newIntVector(endpoints)
	if err != nil {
		return err
	}
	defer cEndpoints.close()
	code := C.igraph_add_edges(&g.graph, &cEndpoints.value, nil)
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
