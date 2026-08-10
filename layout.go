package igraph

/*
#include <stdint.h>
#include <igraph.h>
#include "layout_cgo.h"
*/
import "C"
import "fmt"

// DegMode controls edge direction interpretation in layout algorithms.
// It is an alias for DirectionMode (DegOut, DegIn, DegAll).
type DegMode = DirectionMode

const (
	DegOut = DirectionOut
	DegIn  = DirectionIn
	DegAll = DirectionAll
)

// LayoutRandomOptions specifies options for generating a random 2D layout.
type LayoutRandomOptions struct {
	// Seed is an optional seed for thread-safe reproducible execution.
	Seed *uint64
}

// SugiyamaOptions controls parameters for the Sugiyama layered layout.
type SugiyamaOptions struct {
	// HGap is the horizontal gap between vertices.
	HGap float64
	// VGap is the vertical gap between layers.
	VGap float64
	// MaxIter is the maximum number of iterations.
	MaxIter int
	// Weights specifies optional edge weights for crossing reduction.
	Weights []float64
}

// FruchtermanReingoldOptions controls parameters for the Fruchterman-Reingold 2D layout algorithm.
type FruchtermanReingoldOptions struct {
	// Seed is an optional seed for thread-safe reproducible execution.
	Seed *uint64
	// NIter is the number of iterations (default 500 if <= 0).
	NIter int
	// StartTemp is the starting temperature (default 0.0 uses upstream calculation).
	StartTemp float64
	// Weights specifies optional edge weights.
	Weights []float64
	// InitialCoordinates provides optional starting coordinates (matrix must have rows == V and cols == 2).
	InitialCoordinates *Matrix
	// MinX, MaxX, MinY, MaxY optionally bound vertex coordinates (slice lengths must match vertex count).
	MinX []float64
	MaxX []float64
	MinY []float64
	MaxY []float64
}

// KamadaKawaiOptions controls parameters for the Kamada-Kawai 2D layout algorithm.
type KamadaKawaiOptions struct {
	// Seed is an optional seed for thread-safe reproducible execution.
	Seed *uint64
	// MaxIter is the maximum number of iterations (default 500 * V if <= 0).
	MaxIter int
	// Epsilon is the convergence threshold.
	Epsilon float64
	// KKConst is the Kamada-Kawai constant (default 0.0 uses V).
	KKConst float64
	// Weights specifies optional edge weights.
	Weights []float64
	// InitialCoordinates provides optional starting coordinates (matrix must have rows == V and cols == 2).
	InitialCoordinates *Matrix
	// MinX, MaxX, MinY, MaxY optionally bound vertex coordinates (slice lengths must match vertex count).
	MinX []float64
	MaxX []float64
	MinY []float64
	MaxY []float64
}

// MDSOptions controls options for Multi-Dimensional Scaling (MDS) layout.
type MDSOptions struct {
	// Seed is an optional seed for thread-safe reproducible execution.
	Seed *uint64
}

// newOrderVertexSelector validates order slice and converts it to cVertexSelector.
// If order is nil, natural vertex ordering is used.
func newOrderVertexSelector(order []int, numVertices int) (*cVertexSelector, error) {
	if order == nil {
		return newCVertexSelector(AllVertices())
	}
	if len(order) != numVertices {
		return nil, fmt.Errorf("igraph: order slice length (%d) does not match vertex count (%d)", len(order), numVertices)
	}
	selector, err := VertexIDs(order...)
	if err != nil {
		return nil, err
	}
	if err := validateVertexSelector(selector, numVertices); err != nil {
		return nil, err
	}
	return newCVertexSelector(selector)
}

// LayoutCircle computes a 2D circular layout for the graph.
// If order is nil, natural vertex ordering is used.
// Public input slices are borrowed and returned values are Go-owned.
//
//igraph:bind igraph_layout_circle
func (g *Graph) LayoutCircle(order []int) (Matrix, error) {
	if g == nil {
		return Matrix{}, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return Matrix{}, ErrClosed
	}

	numVertices := int(C.igraph_vcount(&g.graph))
	cOrder, err := newOrderVertexSelector(order, numVertices)
	if err != nil {
		return Matrix{}, err
	}
	defer cOrder.close()

	cMat, err := newCMatrix(Matrix{})
	if err != nil {
		return Matrix{}, err
	}
	defer cMat.close()

	code := C.go_igraph_layout_circle(&g.graph, &cMat.value, cOrder.value)
	if code != C.IGRAPH_SUCCESS {
		return Matrix{}, igraphError("calculate circle layout", int(code))
	}
	return cMat.matrix()
}

// LayoutStar computes a 2D star layout with the given center vertex.
// If order is nil, natural vertex ordering is used.
// Public input slices are borrowed and returned values are Go-owned.
//
//igraph:bind igraph_layout_star
func (g *Graph) LayoutStar(center int, order []int) (Matrix, error) {
	if g == nil {
		return Matrix{}, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return Matrix{}, ErrClosed
	}

	numVertices := int(C.igraph_vcount(&g.graph))
	if numVertices == 0 || center < 0 || center >= numVertices {
		return Matrix{}, fmt.Errorf("igraph: center vertex ID %d is out of bounds for graph with %d vertices", center, numVertices)
	}

	var cOrderPtr *C.igraph_vector_int_t
	if order != nil {
		if len(order) != numVertices {
			return Matrix{}, fmt.Errorf("igraph: order slice length (%d) does not match vertex count (%d)", len(order), numVertices)
		}
		for i, id := range order {
			if id < 0 || id >= numVertices {
				return Matrix{}, fmt.Errorf("igraph: vertex ID at order index %d is %d, outside [0, %d)", i, id, numVertices)
			}
		}
		cOrder, err := newIntVector(order)
		if err != nil {
			return Matrix{}, err
		}
		defer cOrder.close()
		cOrderPtr = &cOrder.value
	}

	cMat, err := newCMatrix(Matrix{})
	if err != nil {
		return Matrix{}, err
	}
	defer cMat.close()

	cCenter, err := intToIgraphInt(center, "center vertex ID")
	if err != nil {
		return Matrix{}, err
	}

	code := C.go_igraph_layout_star(&g.graph, &cMat.value, cCenter, cOrderPtr)
	if code != C.IGRAPH_SUCCESS {
		return Matrix{}, igraphError("calculate star layout", int(code))
	}
	return cMat.matrix()
}

// LayoutGrid computes a 2D regular grid layout with the specified row width.
// Public input slices are borrowed and returned values are Go-owned.
//
//igraph:bind igraph_layout_grid
func (g *Graph) LayoutGrid(width int) (Matrix, error) {
	if g == nil {
		return Matrix{}, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return Matrix{}, ErrClosed
	}

	if width < 0 {
		return Matrix{}, fmt.Errorf("igraph: width must be non-negative: %d", width)
	}

	cMat, err := newCMatrix(Matrix{})
	if err != nil {
		return Matrix{}, err
	}
	defer cMat.close()

	cWidth, err := intToIgraphInt(width, "grid width")
	if err != nil {
		return Matrix{}, err
	}

	code := C.go_igraph_layout_grid(&g.graph, &cMat.value, cWidth)
	if code != C.IGRAPH_SUCCESS {
		return Matrix{}, igraphError("calculate grid layout", int(code))
	}
	return cMat.matrix()
}

// LayoutRandom computes a 2D random layout where coordinates are uniformly distributed.
// LayoutRandomOptions allows specifying a random seed for thread-safe reproducible execution.
// Public input slices are borrowed and returned values are Go-owned.
//
//igraph:bind igraph_layout_random
func (g *Graph) LayoutRandom(options LayoutRandomOptions) (Matrix, error) {
	if g == nil {
		return Matrix{}, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return Matrix{}, ErrClosed
	}

	cMat, err := newCMatrix(Matrix{})
	if err != nil {
		return Matrix{}, err
	}
	defer cMat.close()

	errRNG := withRNG(options.Seed, func() error {
		code := C.go_igraph_layout_random(&g.graph, &cMat.value)
		if code != C.IGRAPH_SUCCESS {
			return igraphError("calculate random layout", int(code))
		}
		return nil
	})
	if errRNG != nil {
		return Matrix{}, errRNG
	}
	return cMat.matrix()
}

// LayoutReingoldTilford computes a 2D tree layout using the Reingold-Tilford algorithm.
// If roots is nil, root vertices are selected automatically.
// Public input slices are borrowed and returned values are Go-owned.
//
//igraph:bind igraph_layout_reingold_tilford
func (g *Graph) LayoutReingoldTilford(mode DegMode, roots []int) (Matrix, error) {
	if g == nil {
		return Matrix{}, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return Matrix{}, ErrClosed
	}

	cMode, err := mode.cValue()
	if err != nil {
		return Matrix{}, err
	}

	numVertices := int(C.igraph_vcount(&g.graph))
	var cRootsPtr *C.igraph_vector_int_t
	if roots != nil {
		for i, id := range roots {
			if id < 0 || id >= numVertices {
				return Matrix{}, fmt.Errorf("igraph: root vertex ID at index %d is %d, outside [0, %d)", i, id, numVertices)
			}
		}
		cRoots, err := newIntVector(roots)
		if err != nil {
			return Matrix{}, err
		}
		defer cRoots.close()
		cRootsPtr = &cRoots.value
	}

	cMat, err := newCMatrix(Matrix{})
	if err != nil {
		return Matrix{}, err
	}
	defer cMat.close()

	code := C.go_igraph_layout_reingold_tilford(&g.graph, &cMat.value, cMode, cRootsPtr, nil)
	if code != C.IGRAPH_SUCCESS {
		return Matrix{}, igraphError("calculate Reingold-Tilford layout", int(code))
	}
	return cMat.matrix()
}

// LayoutReingoldTilfordCircular computes a 2D circular tree layout using the Reingold-Tilford algorithm.
// If roots is nil, root vertices are selected automatically.
// Public input slices are borrowed and returned values are Go-owned.
//
//igraph:bind igraph_layout_reingold_tilford_circular
func (g *Graph) LayoutReingoldTilfordCircular(mode DegMode, roots []int) (Matrix, error) {
	if g == nil {
		return Matrix{}, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return Matrix{}, ErrClosed
	}

	cMode, err := mode.cValue()
	if err != nil {
		return Matrix{}, err
	}

	numVertices := int(C.igraph_vcount(&g.graph))
	var cRootsPtr *C.igraph_vector_int_t
	if roots != nil {
		for i, id := range roots {
			if id < 0 || id >= numVertices {
				return Matrix{}, fmt.Errorf("igraph: root vertex ID at index %d is %d, outside [0, %d)", i, id, numVertices)
			}
		}
		cRoots, err := newIntVector(roots)
		if err != nil {
			return Matrix{}, err
		}
		defer cRoots.close()
		cRootsPtr = &cRoots.value
	}

	cMat, err := newCMatrix(Matrix{})
	if err != nil {
		return Matrix{}, err
	}
	defer cMat.close()

	code := C.go_igraph_layout_reingold_tilford_circular(&g.graph, &cMat.value, cMode, cRootsPtr, nil)
	if code != C.IGRAPH_SUCCESS {
		return Matrix{}, igraphError("calculate circular Reingold-Tilford layout", int(code))
	}
	return cMat.matrix()
}

// LayoutBipartite computes a 2D layout for bipartite graphs.
// types specifies the partition for each vertex (true for one type, false for the other).
// Public input slices are borrowed and returned values are Go-owned.
//
//igraph:bind igraph_layout_bipartite
func (g *Graph) LayoutBipartite(types []bool, hgap float64, vgap float64, maxIter int) (Matrix, error) {
	if g == nil {
		return Matrix{}, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return Matrix{}, ErrClosed
	}

	numVertices := int(C.igraph_vcount(&g.graph))
	if len(types) != numVertices {
		return Matrix{}, fmt.Errorf("igraph: types slice length (%d) does not match vertex count (%d)", len(types), numVertices)
	}

	if maxIter < 0 {
		return Matrix{}, fmt.Errorf("igraph: maxIter must be non-negative: %d", maxIter)
	}

	cTypes, err := newBoolVector(types)
	if err != nil {
		return Matrix{}, err
	}
	defer cTypes.close()

	cMat, err := newCMatrix(Matrix{})
	if err != nil {
		return Matrix{}, err
	}
	defer cMat.close()

	cMaxIter, err := intToIgraphInt(maxIter, "maxIter")
	if err != nil {
		return Matrix{}, err
	}

	code := C.go_igraph_layout_bipartite(
		&g.graph,
		&cTypes.value,
		&cMat.value,
		C.igraph_real_t(hgap),
		C.igraph_real_t(vgap),
		cMaxIter,
	)
	if code != C.IGRAPH_SUCCESS {
		return Matrix{}, igraphError("calculate bipartite layout", int(code))
	}
	return cMat.matrix()
}

// LayoutSugiyama computes a 2D Sugiyama layered layout for directed acyclic graphs.
// If layers is non-nil, it specifies explicit 0-based layer assignments per vertex.
// Public input slices are borrowed and returned values are Go-owned.
//
//igraph:bind igraph_layout_sugiyama
func (g *Graph) LayoutSugiyama(layers []int, options SugiyamaOptions) (Matrix, error) {
	if g == nil {
		return Matrix{}, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return Matrix{}, ErrClosed
	}

	numVertices := int(C.igraph_vcount(&g.graph))
	var cLayersPtr *C.igraph_vector_int_t
	if layers != nil {
		if len(layers) != numVertices {
			return Matrix{}, fmt.Errorf("igraph: layers slice length (%d) does not match vertex count (%d)", len(layers), numVertices)
		}
		for i, l := range layers {
			if l < 0 {
				return Matrix{}, fmt.Errorf("igraph: layer assignment at index %d must be non-negative: %d", i, l)
			}
		}
		cLayers, err := newIntVector(layers)
		if err != nil {
			return Matrix{}, err
		}
		defer cLayers.close()
		cLayersPtr = &cLayers.value
	}

	if options.MaxIter < 0 {
		return Matrix{}, fmt.Errorf("igraph: maxIter must be non-negative: %d", options.MaxIter)
	}

	weightsVec, err := newOptionalEdgeWeights(options.Weights, int(C.igraph_ecount(&g.graph)))
	if err != nil {
		return Matrix{}, err
	}
	if weightsVec != nil {
		defer weightsVec.close()
	}

	cMat, err := newCMatrix(Matrix{})
	if err != nil {
		return Matrix{}, err
	}
	defer cMat.close()

	cMaxIter, err := intToIgraphInt(options.MaxIter, "maxIter")
	if err != nil {
		return Matrix{}, err
	}

	code := C.go_igraph_layout_sugiyama(
		&g.graph,
		&cMat.value,
		nil,
		cLayersPtr,
		C.igraph_real_t(options.HGap),
		C.igraph_real_t(options.VGap),
		cMaxIter,
		edgeWeightPointer(weightsVec),
	)
	if code != C.IGRAPH_SUCCESS {
		return Matrix{}, igraphError("calculate Sugiyama layout", int(code))
	}
	return cMat.matrix()
}

// LayoutFruchtermanReingold computes a 2D force-directed layout using the Fruchterman-Reingold algorithm.
// Public input slices and matrices are borrowed/copied and returned values are Go-owned.
//
//igraph:bind igraph_layout_fruchterman_reingold
func (g *Graph) LayoutFruchtermanReingold(options FruchtermanReingoldOptions) (Matrix, error) {
	if g == nil {
		return Matrix{}, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return Matrix{}, ErrClosed
	}

	numVertices := int(C.igraph_vcount(&g.graph))
	numEdges := int(C.igraph_ecount(&g.graph))

	if options.NIter < 0 {
		return Matrix{}, fmt.Errorf("igraph: NIter must be non-negative: %d", options.NIter)
	}

	weightsVec, err := newOptionalEdgeWeights(options.Weights, numEdges)
	if err != nil {
		return Matrix{}, err
	}
	if weightsVec != nil {
		defer weightsVec.close()
	}

	minXVec, err := newOptionalVertexWeights(options.MinX, numVertices)
	if err != nil {
		return Matrix{}, err
	}
	if minXVec != nil {
		defer minXVec.close()
	}

	maxXVec, err := newOptionalVertexWeights(options.MaxX, numVertices)
	if err != nil {
		return Matrix{}, err
	}
	if maxXVec != nil {
		defer maxXVec.close()
	}

	minYVec, err := newOptionalVertexWeights(options.MinY, numVertices)
	if err != nil {
		return Matrix{}, err
	}
	if minYVec != nil {
		defer minYVec.close()
	}

	maxYVec, err := newOptionalVertexWeights(options.MaxY, numVertices)
	if err != nil {
		return Matrix{}, err
	}
	if maxYVec != nil {
		defer maxYVec.close()
	}

	var useSeed C.igraph_bool_t
	var cMat *cMatrix
	if options.InitialCoordinates != nil {
		rows, cols := options.InitialCoordinates.Dims()
		if rows != numVertices || cols != 2 {
			return Matrix{}, fmt.Errorf("igraph: initial coordinates matrix dimensions (%d, %d) do not match vertex count %d and dimension 2", rows, cols, numVertices)
		}
		cMat, err = newCMatrix(*options.InitialCoordinates)
		if err != nil {
			return Matrix{}, err
		}
		useSeed = booltoint(true)
	} else {
		cMat, err = newCMatrix(Matrix{})
		if err != nil {
			return Matrix{}, err
		}
		useSeed = booltoint(false)
	}
	defer cMat.close()

	cNIter, err := intToIgraphInt(options.NIter, "NIter")
	if err != nil {
		return Matrix{}, err
	}

	var runErr error
	errRNG := withRNG(options.Seed, func() error {
		code := C.go_igraph_layout_fruchterman_reingold(
			&g.graph,
			&cMat.value,
			useSeed,
			cNIter,
			C.igraph_real_t(options.StartTemp),
			C.IGRAPH_LAYOUT_AUTOGRID,
			edgeWeightPointer(weightsVec),
			edgeWeightPointer(minXVec),
			edgeWeightPointer(maxXVec),
			edgeWeightPointer(minYVec),
			edgeWeightPointer(maxYVec),
		)
		if code != C.IGRAPH_SUCCESS {
			runErr = igraphError("calculate Fruchterman-Reingold layout", int(code))
			return runErr
		}
		return nil
	})
	if errRNG != nil {
		return Matrix{}, errRNG
	}
	if runErr != nil {
		return Matrix{}, runErr
	}

	return cMat.matrix()
}

// LayoutKamadaKawai computes a 2D force-directed layout using the Kamada-Kawai algorithm.
// Public input slices and matrices are borrowed/copied and returned values are Go-owned.
//
//igraph:bind igraph_layout_kamada_kawai
func (g *Graph) LayoutKamadaKawai(options KamadaKawaiOptions) (Matrix, error) {
	if g == nil {
		return Matrix{}, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return Matrix{}, ErrClosed
	}

	numVertices := int(C.igraph_vcount(&g.graph))
	numEdges := int(C.igraph_ecount(&g.graph))

	if options.MaxIter < 0 {
		return Matrix{}, fmt.Errorf("igraph: MaxIter must be non-negative: %d", options.MaxIter)
	}

	maxIter := options.MaxIter
	if maxIter == 0 {
		maxIter = 500 * numVertices
		if maxIter == 0 {
			maxIter = 100
		}
	}

	epsilon := options.Epsilon
	if epsilon <= 0 {
		epsilon = 0.0001
	}

	kkconst := options.KKConst
	if kkconst <= 0 {
		kkconst = float64(numVertices)
		if kkconst == 0 {
			kkconst = 1
		}
	}

	weightsVec, err := newOptionalEdgeWeights(options.Weights, numEdges)
	if err != nil {
		return Matrix{}, err
	}
	if weightsVec != nil {
		defer weightsVec.close()
	}

	minXVec, err := newOptionalVertexWeights(options.MinX, numVertices)
	if err != nil {
		return Matrix{}, err
	}
	if minXVec != nil {
		defer minXVec.close()
	}

	maxXVec, err := newOptionalVertexWeights(options.MaxX, numVertices)
	if err != nil {
		return Matrix{}, err
	}
	if maxXVec != nil {
		defer maxXVec.close()
	}

	minYVec, err := newOptionalVertexWeights(options.MinY, numVertices)
	if err != nil {
		return Matrix{}, err
	}
	if minYVec != nil {
		defer minYVec.close()
	}

	maxYVec, err := newOptionalVertexWeights(options.MaxY, numVertices)
	if err != nil {
		return Matrix{}, err
	}
	if maxYVec != nil {
		defer maxYVec.close()
	}

	var useSeed C.igraph_bool_t
	var cMat *cMatrix
	if options.InitialCoordinates != nil {
		rows, cols := options.InitialCoordinates.Dims()
		if rows != numVertices || cols != 2 {
			return Matrix{}, fmt.Errorf("igraph: initial coordinates matrix dimensions (%d, %d) do not match vertex count %d and dimension 2", rows, cols, numVertices)
		}
		cMat, err = newCMatrix(*options.InitialCoordinates)
		if err != nil {
			return Matrix{}, err
		}
		useSeed = booltoint(true)
	} else {
		cMat, err = newCMatrix(Matrix{})
		if err != nil {
			return Matrix{}, err
		}
		useSeed = booltoint(false)
	}
	defer cMat.close()

	cMaxIter, err := intToIgraphInt(maxIter, "MaxIter")
	if err != nil {
		return Matrix{}, err
	}

	var runErr error
	errRNG := withRNG(options.Seed, func() error {
		code := C.go_igraph_layout_kamada_kawai(
			&g.graph,
			&cMat.value,
			useSeed,
			cMaxIter,
			C.igraph_real_t(epsilon),
			C.igraph_real_t(kkconst),
			edgeWeightPointer(weightsVec),
			edgeWeightPointer(minXVec),
			edgeWeightPointer(maxXVec),
			edgeWeightPointer(minYVec),
			edgeWeightPointer(maxYVec),
		)
		if code != C.IGRAPH_SUCCESS {
			runErr = igraphError("calculate Kamada-Kawai layout", int(code))
			return runErr
		}
		return nil
	})
	if errRNG != nil {
		return Matrix{}, errRNG
	}
	if runErr != nil {
		return Matrix{}, runErr
	}

	return cMat.matrix()
}

// LayoutMDS computes a layout using Multi-Dimensional Scaling based on vertex distance matrix.
// If distances is nil, shortest path distances between all vertex pairs are computed automatically.
// If distances is non-nil, its dimensions must be square and match vertex count.
// Symmetry handling is delegated to upstream, which accepts asymmetric matrices via its warning path.
// dim specifies layout dimension (must be 2 or 3; default 2 if <= 0).
// Public input matrices are borrowed and returned values are Go-owned.
//
//igraph:bind igraph_layout_mds
func (g *Graph) LayoutMDS(distances *Matrix, dim int, options MDSOptions) (Matrix, error) {
	if g == nil {
		return Matrix{}, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return Matrix{}, ErrClosed
	}

	numVertices := int(C.igraph_vcount(&g.graph))

	if dim <= 0 {
		dim = 2
	}
	if dim != 2 && dim != 3 {
		return Matrix{}, fmt.Errorf("igraph: layout dimension must be 2 or 3: %d", dim)
	}

	var cDistPtr *C.igraph_matrix_t
	if distances != nil {
		rows, cols := distances.Dims()
		if rows != numVertices || cols != numVertices {
			return Matrix{}, fmt.Errorf("igraph: distance matrix dimensions (%d, %d) do not match vertex count %d", rows, cols, numVertices)
		}
		cDist, err := newCMatrix(*distances)
		if err != nil {
			return Matrix{}, err
		}
		defer cDist.close()
		cDistPtr = &cDist.value
	}

	cMat, err := newCMatrix(Matrix{})
	if err != nil {
		return Matrix{}, err
	}
	defer cMat.close()

	cDim, err := intToIgraphInt(dim, "dimension")
	if err != nil {
		return Matrix{}, err
	}

	var runErr error
	errRNG := withRNG(options.Seed, func() error {
		code := C.go_igraph_layout_mds(
			&g.graph,
			&cMat.value,
			cDistPtr,
			cDim,
		)
		if code != C.IGRAPH_SUCCESS {
			runErr = igraphError("calculate MDS layout", int(code))
			return runErr
		}
		return nil
	})
	if errRNG != nil {
		return Matrix{}, errRNG
	}
	if runErr != nil {
		return Matrix{}, runErr
	}

	return cMat.matrix()
}
