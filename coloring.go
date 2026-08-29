package igraph

// #cgo pkg-config: igraph
// #include <igraph.h>
import "C"

import "fmt"

// ColoringHeuristic selects the deterministic ordering used by greedy vertex
// coloring. Its zero value, ColoringColoredNeighbors, is the default.
type ColoringHeuristic int

const (
	// ColoringColoredNeighbors repeatedly chooses a vertex with the most
	// already-colored neighbors.
	ColoringColoredNeighbors ColoringHeuristic = iota
	// ColoringDSatur repeatedly chooses a vertex with the highest saturation
	// degree, breaking ties by degree.
	ColoringDSatur
)

func (heuristic ColoringHeuristic) cValue() (C.igraph_coloring_greedy_t, error) {
	switch heuristic {
	case ColoringColoredNeighbors:
		return C.IGRAPH_COLORING_GREEDY_COLORED_NEIGHBORS, nil
	case ColoringDSatur:
		return C.IGRAPH_COLORING_GREEDY_DSATUR, nil
	default:
		return 0, fmt.Errorf("igraph: invalid coloring heuristic %d", heuristic)
	}
}

// BipartiteColoringResult reports whether a caller-provided partition is a
// valid bipartite coloring and the orientation of its directed cross-mode
// edges. DirectionOut means false-to-true, DirectionIn means true-to-false,
// and DirectionAll means undirected, mixed directions, or no directional
// evidence. Direction is also DirectionAll when Valid is false.
type BipartiteColoringResult struct {
	Valid     bool
	Direction DirectionMode
}

// GreedyVertexColoring returns a non-nil Go-owned color slice indexed by
// vertex ID. Colors are non-negative. The result is valid but is not
// necessarily a minimum coloring. Direction and parallel-edge multiplicity
// do not affect validity. Self-loops are ignored by pinned igraph.
//
//igraph:bind igraph_vertex_coloring_greedy
func (g *Graph) GreedyVertexColoring(heuristic ColoringHeuristic) ([]int, error) {
	return g.greedyVertexColoring(heuristic, nil)
}

// IsVertexColoring reports whether colors is a valid vertex coloring. colors
// is borrowed for the synchronous call, must have one non-negative entry per
// vertex, and is copied into temporary C-owned storage. Direction and loops
// are ignored; parallel edges do not change adjacency.
//
//igraph:bind igraph_is_vertex_coloring
func (g *Graph) IsVertexColoring(colors []int) (bool, error) {
	return g.validateColoring(colors, false, nil)
}

// IsEdgeColoring reports whether colors is a valid edge coloring. colors is
// borrowed for the synchronous call, must have one non-negative entry per
// edge, and is copied into temporary C-owned storage. Direction is ignored;
// parallel and other incident edges must differ, while a loop is not compared
// with itself.
//
//igraph:bind igraph_is_edge_coloring
func (g *Graph) IsEdgeColoring(colors []int) (bool, error) {
	return g.validateColoring(colors, true, nil)
}

// IsBipartiteColoring validates partition and reports its inferred direction.
// partition is borrowed for the synchronous call, must have one entry per
// vertex, and is copied into temporary C-owned storage. Pinned igraph ignores
// self-loops here, unlike Graph.Bipartite. Undirected or mixed-direction edges
// produce DirectionAll.
//
//igraph:bind igraph_is_bipartite_coloring
func (g *Graph) IsBipartiteColoring(partition BipartitePartition) (BipartiteColoringResult, error) {
	return g.isBipartiteColoring(partition, nil)
}

type coloringAdapters struct {
	newInt    func([]int) (*intVector, error)
	closeInt  func(*intVector)
	intSlice  func(*intVector) ([]int, error)
	greedy    func(*Graph, *intVector, ColoringHeuristic) int
	validate  func(*Graph, *intVector, bool) (bool, int)
	newBool   func([]bool) (*boolVector, error)
	closeBool func(*boolVector)
	bipartite func(*Graph, *boolVector) (bool, DirectionMode, int)
}

func defaultColoringAdapters() coloringAdapters {
	return coloringAdapters{
		newInt: newIntVector, closeInt: (*intVector).close, intSlice: (*intVector).slice,
		greedy: func(g *Graph, colors *intVector, heuristic ColoringHeuristic) int {
			value, _ := heuristic.cValue()
			return int(C.igraph_vertex_coloring_greedy(&g.graph, &colors.value, value))
		},
		validate: func(g *Graph, colors *intVector, edge bool) (bool, int) {
			var result C.igraph_bool_t
			var code C.igraph_error_t
			if edge {
				code = C.igraph_is_edge_coloring(&g.graph, &colors.value, &result)
			} else {
				code = C.igraph_is_vertex_coloring(&g.graph, &colors.value, &result)
			}
			return result != booltoint(false), int(code)
		},
		newBool: newBoolVector, closeBool: (*boolVector).close,
		bipartite: func(g *Graph, partition *boolVector) (bool, DirectionMode, int) {
			var result C.igraph_bool_t
			var mode C.igraph_neimode_t
			code := C.igraph_is_bipartite_coloring(&g.graph, &partition.value, &result, &mode)
			direction := DirectionAll
			if result != booltoint(false) {
				switch mode {
				case C.IGRAPH_OUT:
					direction = DirectionOut
				case C.IGRAPH_IN:
					direction = DirectionIn
				}
			}
			return result != booltoint(false), direction, int(code)
		},
	}
}

func resolvedColoringAdapters(adapters *coloringAdapters) coloringAdapters {
	if adapters == nil {
		return defaultColoringAdapters()
	}
	return *adapters
}

func (g *Graph) greedyVertexColoring(heuristic ColoringHeuristic, adapters *coloringAdapters) ([]int, error) {
	if g == nil {
		return nil, ErrClosed
	}
	if _, err := heuristic.cValue(); err != nil {
		return nil, err
	}
	g.mu.RLock()
	defer g.mu.RUnlock()
	if g.closed {
		return nil, ErrClosed
	}
	op := resolvedColoringAdapters(adapters)
	colors, err := op.newInt(nil)
	if err != nil {
		return nil, err
	}
	defer op.closeInt(colors)
	if code := op.greedy(g, colors, heuristic); code != int(C.IGRAPH_SUCCESS) {
		return nil, igraphError("greedily color vertices", code)
	}
	values, err := op.intSlice(colors)
	if err != nil {
		return nil, err
	}
	if values == nil {
		values = []int{}
	}
	return values, nil
}

func validateColors(colors []int, want int, kind string) error {
	if len(colors) != want {
		return fmt.Errorf("igraph: %s color count %d does not match %s count %d", kind, len(colors), kind, want)
	}
	for i, color := range colors {
		if color < 0 {
			return fmt.Errorf("igraph: %s color at index %d must be non-negative: %d", kind, i, color)
		}
		if _, err := intToIgraphInt(color, fmt.Sprintf("%s color at index %d", kind, i)); err != nil {
			return err
		}
	}
	return nil
}

func (g *Graph) validateColoring(colors []int, edge bool, adapters *coloringAdapters) (bool, error) {
	if g == nil {
		return false, ErrClosed
	}
	g.mu.RLock()
	defer g.mu.RUnlock()
	if g.closed {
		return false, ErrClosed
	}
	kind, count := "vertex", int(C.igraph_vcount(&g.graph))
	if edge {
		kind, count = "edge", int(C.igraph_ecount(&g.graph))
	}
	if err := validateColors(colors, count, kind); err != nil {
		return false, err
	}
	op := resolvedColoringAdapters(adapters)
	vector, err := op.newInt(colors)
	if err != nil {
		return false, err
	}
	defer op.closeInt(vector)
	valid, code := op.validate(g, vector, edge)
	if code != int(C.IGRAPH_SUCCESS) {
		return false, igraphError("validate "+kind+" coloring", code)
	}
	return valid, nil
}

func (g *Graph) isBipartiteColoring(partition BipartitePartition, adapters *coloringAdapters) (BipartiteColoringResult, error) {
	if g == nil {
		return BipartiteColoringResult{}, ErrClosed
	}
	g.mu.RLock()
	defer g.mu.RUnlock()
	if g.closed {
		return BipartiteColoringResult{}, ErrClosed
	}
	if err := validateBipartitePartitionLength(partition, int(C.igraph_vcount(&g.graph))); err != nil {
		return BipartiteColoringResult{}, err
	}
	op := resolvedColoringAdapters(adapters)
	vector, err := op.newBool([]bool(partition))
	if err != nil {
		return BipartiteColoringResult{}, err
	}
	defer op.closeBool(vector)
	valid, direction, code := op.bipartite(g, vector)
	if code != int(C.IGRAPH_SUCCESS) {
		return BipartiteColoringResult{}, igraphError("validate bipartite coloring", code)
	}
	if !valid {
		direction = DirectionAll
	}
	return BipartiteColoringResult{Valid: valid, Direction: direction}, nil
}
