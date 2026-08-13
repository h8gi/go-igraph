package igraph

// #cgo pkg-config: igraph
// #include <igraph.h>
import "C"

import "fmt"

// BipartitePartition assigns every vertex to one of a graph's two modes.
// Indexes are vertex IDs. False and true name the modes without assigning
// domain-specific meaning to either one. Inputs are borrowed only for a
// synchronous call; returned partitions are non-nil Go-owned copies.
type BipartitePartition []bool

// BipartiteResult reports whether a graph is bipartite. Partition is non-nil
// and has one entry per vertex when IsBipartite is true. It is empty when the
// graph is not bipartite. For disconnected graphs, the orientation of each
// component is chosen by pinned igraph and is not a compatibility promise.
type BipartiteResult struct {
	IsBipartite bool
	Partition   BipartitePartition
}

// BipartiteGraphResult contains an independently owned graph and its explicit
// partition. Graph must be closed by the caller. Partition is a non-nil
// Go-owned value that remains valid after Graph is closed.
type BipartiteGraphResult struct {
	Graph     *Graph
	Partition BipartitePartition
}

// Bipartite reports whether g is bipartite and, when it is, returns one valid
// partition. Direction is ignored. A self-loop makes a graph non-bipartite.
// The returned partition remains valid after g is closed.
//
//igraph:bind igraph_is_bipartite
func (g *Graph) Bipartite() (BipartiteResult, error) {
	return g.bipartite(nil)
}

// IsBipartitePartition reports whether partition assigns every edge endpoint
// to opposite modes. Direction is ignored. Partition is borrowed only for the
// call and must contain one entry per vertex.
func (g *Graph) IsBipartitePartition(partition BipartitePartition) (bool, error) {
	if g == nil {
		return false, ErrClosed
	}
	g.mu.RLock()
	defer g.mu.RUnlock()
	if g.closed {
		return false, ErrClosed
	}
	if err := validateBipartitePartitionLength(partition, int(C.igraph_vcount(&g.graph))); err != nil {
		return false, err
	}
	edgeCount := int(C.igraph_ecount(&g.graph))
	for edgeID := 0; edgeID < edgeCount; edgeID++ {
		var from, to C.igraph_int_t
		if code := C.igraph_edge(&g.graph, C.igraph_int_t(edgeID), &from, &to); code != C.IGRAPH_SUCCESS {
			return false, igraphError("inspect edge for bipartite partition", int(code))
		}
		if partition[int(from)] == partition[int(to)] {
			return false, nil
		}
	}
	return true, nil
}

// NewBipartite constructs a graph with the supplied explicit partition and
// edges. Every edge must join opposite modes. Self-loops are therefore invalid;
// parallel edges are allowed. Partition and edges are borrowed only for the
// call. The returned graph and partition are independently Go-owned.
//
//igraph:bind igraph_create_bipartite
func NewBipartite(partition BipartitePartition, edges []Edge, directed bool) (BipartiteGraphResult, error) {
	return newBipartite(partition, edges, directed, nil)
}

// NewFullBipartite constructs a complete bipartite graph with falseModeSize
// false-mode vertices followed by trueModeSize true-mode vertices. For a
// directed graph, DirectionOut creates false-to-true edges, DirectionIn creates
// true-to-false edges, and DirectionAll creates both. Direction is ignored for
// an undirected graph. The returned graph and partition are independently
// Go-owned.
//
//igraph:bind igraph_full_bipartite
func NewFullBipartite(falseModeSize, trueModeSize int, directed bool, direction DirectionMode) (BipartiteGraphResult, error) {
	return newFullBipartite(falseModeSize, trueModeSize, directed, direction, nil)
}

func validateBipartitePartitionLength(partition BipartitePartition, vertexCount int) error {
	if len(partition) != vertexCount {
		return fmt.Errorf("igraph: bipartite partition length %d does not match vertex count %d", len(partition), vertexCount)
	}
	return nil
}

type bipartiteAdapters struct {
	newBool     func([]bool) (*boolVector, error)
	convertBool func(*boolVector) ([]bool, error)
	closeBool   func(*boolVector)
	check       func(*Graph, *boolVector) (bool, int)
	create      func(*boolVector, *intVector, bool) bipartiteGraphCallResult
	full        func(int, int, bool, DirectionMode, *boolVector) bipartiteGraphCallResult
}

type bipartiteGraphCallResult struct {
	graph C.igraph_t
	code  int
}

func defaultBipartiteAdapters() bipartiteAdapters {
	return bipartiteAdapters{
		newBool:     newBoolVector,
		convertBool: (*boolVector).slice,
		closeBool:   (*boolVector).close,
		check: func(g *Graph, partition *boolVector) (bool, int) {
			var result C.igraph_bool_t
			returnValue := C.igraph_is_bipartite(&g.graph, &result, &partition.value)
			return result != booltoint(false), int(returnValue)
		},
		create: func(partition *boolVector, edges *intVector, directed bool) bipartiteGraphCallResult {
			var graph C.igraph_t
			code := C.igraph_create_bipartite(&graph, &partition.value, &edges.value, booltoint(directed))
			return bipartiteGraphCallResult{graph: graph, code: int(code)}
		},
		full: func(n1, n2 int, directed bool, direction DirectionMode, partition *boolVector) bipartiteGraphCallResult {
			var graph C.igraph_t
			mode, err := direction.cValue()
			if err != nil {
				return bipartiteGraphCallResult{code: int(C.IGRAPH_EINVAL)}
			}
			code := C.igraph_full_bipartite(&graph, &partition.value, C.igraph_int_t(n1), C.igraph_int_t(n2), booltoint(directed), mode)
			return bipartiteGraphCallResult{graph: graph, code: int(code)}
		},
	}
}

func resolvedBipartiteAdapters(adapters *bipartiteAdapters) bipartiteAdapters {
	if adapters == nil {
		return defaultBipartiteAdapters()
	}
	return *adapters
}

func (g *Graph) bipartite(adapters *bipartiteAdapters) (BipartiteResult, error) {
	if g == nil {
		return BipartiteResult{}, ErrClosed
	}
	g.mu.RLock()
	defer g.mu.RUnlock()
	if g.closed {
		return BipartiteResult{}, ErrClosed
	}
	resolved := resolvedBipartiteAdapters(adapters)
	partition, err := resolved.newBool(nil)
	if err != nil {
		return BipartiteResult{}, err
	}
	defer resolved.closeBool(partition)
	isBipartite, code := resolved.check(g, partition)
	if code != int(C.IGRAPH_SUCCESS) {
		return BipartiteResult{}, igraphError("check whether graph is bipartite", code)
	}
	if !isBipartite {
		return BipartiteResult{Partition: BipartitePartition{}}, nil
	}
	values, err := resolved.convertBool(partition)
	if err != nil {
		return BipartiteResult{}, err
	}
	return BipartiteResult{IsBipartite: true, Partition: BipartitePartition(values)}, nil
}

func newBipartite(partition BipartitePartition, edges []Edge, directed bool, adapters *bipartiteAdapters) (BipartiteGraphResult, error) {
	if err := validateConstructorSize("bipartite vertex count", len(partition)); err != nil {
		return BipartiteGraphResult{}, err
	}
	if len(edges) > int(^uint(0)>>1)/2 {
		return BipartiteGraphResult{}, fmt.Errorf("igraph: bipartite edge list is too large")
	}
	endpoints := make([]int, 0, 2*len(edges))
	for index, edge := range edges {
		if err := validateEdge(edge, len(partition), index); err != nil {
			return BipartiteGraphResult{}, err
		}
		if partition[edge.From] == partition[edge.To] {
			return BipartiteGraphResult{}, fmt.Errorf("igraph: edge %d connects vertices %d and %d in the same bipartite mode", index, edge.From, edge.To)
		}
		endpoints = append(endpoints, edge.From, edge.To)
	}
	resolved := resolvedBipartiteAdapters(adapters)
	types, err := resolved.newBool([]bool(partition))
	if err != nil {
		return BipartiteGraphResult{}, err
	}
	defer resolved.closeBool(types)
	cEdges, err := newIntVector(endpoints)
	if err != nil {
		return BipartiteGraphResult{}, err
	}
	defer cEdges.close()
	call := resolved.create(types, cEdges, directed)
	if call.code != int(C.IGRAPH_SUCCESS) {
		return BipartiteGraphResult{}, igraphError("construct bipartite graph", call.code)
	}
	return BipartiteGraphResult{Graph: adoptInitializedGraph(&call.graph), Partition: append(BipartitePartition{}, partition...)}, nil
}

func newFullBipartite(n1, n2 int, directed bool, direction DirectionMode, adapters *bipartiteAdapters) (BipartiteGraphResult, error) {
	if err := validateConstructorSize("false-mode vertex count", n1); err != nil {
		return BipartiteGraphResult{}, err
	}
	if err := validateConstructorSize("true-mode vertex count", n2); err != nil {
		return BipartiteGraphResult{}, err
	}
	if _, err := direction.cValue(); err != nil {
		return BipartiteGraphResult{}, err
	}
	resolved := resolvedBipartiteAdapters(adapters)
	types, err := resolved.newBool(nil)
	if err != nil {
		return BipartiteGraphResult{}, err
	}
	defer resolved.closeBool(types)
	call := resolved.full(n1, n2, directed, direction, types)
	if call.code != int(C.IGRAPH_SUCCESS) {
		return BipartiteGraphResult{}, igraphError("construct full bipartite graph", call.code)
	}
	partition, err := resolved.convertBool(types)
	if err != nil {
		C.igraph_destroy(&call.graph)
		return BipartiteGraphResult{}, err
	}
	return BipartiteGraphResult{Graph: adoptInitializedGraph(&call.graph), Partition: BipartitePartition(partition)}, nil
}
