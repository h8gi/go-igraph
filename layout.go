package igraph

/*
#include <stdint.h>
#include <igraph.h>
#include "layout_cgo.h"
*/
import "C"
import (
	"fmt"
	"math"
)

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
	// HGap is the horizontal gap between vertices (default 1.0 if 0).
	HGap float64
	// VGap is the vertical gap between layers (default 1.0 if 0).
	VGap float64
	// MaxIter is the maximum number of iterations (default 100 if 0; negative is an error).
	MaxIter int
	// Weights specifies optional edge weights for crossing reduction.
	Weights []float64
}

// FruchtermanReingoldOptions controls parameters for the Fruchterman-Reingold 2D layout algorithm.
type FruchtermanReingoldOptions struct {
	// Seed is an optional seed for thread-safe reproducible execution.
	Seed *uint64
	// NIter is the number of iterations (default 500 if 0; negative is an error).
	NIter int
	// StartTemp is the starting temperature, the maximum vertex movement in
	// the first iteration (default sqrt(V) if 0; negative is an error).
	StartTemp float64
	// Weights specifies optional edge weights.
	Weights []float64
	// InitialCoordinates provides optional starting coordinates (matrix must
	// have rows == V and cols matching the layout dimension: 2 for
	// LayoutFruchtermanReingold, 3 for LayoutFruchtermanReingold3D).
	InitialCoordinates *Matrix
	// MinX, MaxX, MinY, MaxY optionally bound vertex coordinates (slice
	// lengths must match vertex count and must not contain NaN; ±Inf
	// disables the bound on that side).
	MinX []float64
	MaxX []float64
	MinY []float64
	MaxY []float64
	// MinZ, MaxZ optionally bound the z coordinate under the same rules;
	// they apply only to the 3D variant and must be nil for the 2D layout.
	MinZ []float64
	MaxZ []float64
}

// KamadaKawaiOptions controls parameters for the Kamada-Kawai 2D layout algorithm.
type KamadaKawaiOptions struct {
	// Seed is an optional seed for thread-safe reproducible execution.
	Seed *uint64
	// MaxIter is the maximum number of iterations (default 500 * V if 0; negative is an error).
	MaxIter int
	// Epsilon is the convergence threshold: iteration stops early once the
	// largest movement drops below it. Zero disables the early stop and runs
	// all MaxIter iterations; negative is an error.
	Epsilon float64
	// KKConst is the Kamada-Kawai vertex attraction constant (default V, or
	// 1 for an empty graph, if 0; negative is an error).
	KKConst float64
	// Weights specifies optional edge weights.
	Weights []float64
	// InitialCoordinates provides optional starting coordinates (matrix must
	// have rows == V and cols matching the layout dimension: 2 for
	// LayoutKamadaKawai, 3 for LayoutKamadaKawai3D).
	InitialCoordinates *Matrix
	// MinX, MaxX, MinY, MaxY optionally bound vertex coordinates (slice
	// lengths must match vertex count and must not contain NaN; ±Inf
	// disables the bound on that side).
	MinX []float64
	MaxX []float64
	MinY []float64
	MaxY []float64
	// MinZ, MaxZ optionally bound the z coordinate under the same rules;
	// they apply only to the 3D variant and must be nil for the 2D layout.
	MinZ []float64
	MaxZ []float64
}

// MDSOptions controls options for Multi-Dimensional Scaling (MDS) layout.
type MDSOptions struct {
	// Seed is an optional seed for thread-safe reproducible execution.
	Seed *uint64
}

// DavidsonHarelOptions controls the Davidson-Harel simulated-annealing layout.
type DavidsonHarelOptions struct {
	Seed                   *uint64
	MaxIter                int
	FineIter               int
	CoolFactor             float64
	InitialCoordinates     *Matrix
	WeightNodeDistance     *float64
	WeightBorder           *float64
	WeightEdgeLengths      *float64
	WeightEdgeCrossings    *float64
	WeightNodeEdgeDistance *float64
}

// GEMOptions controls the GEM force-directed layout.
type GEMOptions struct {
	Seed               *uint64
	MaxIter            int
	MaxTemperature     float64
	MinTemperature     float64
	InitialTemperature float64
	InitialCoordinates *Matrix
}

// GraphoptOptions controls the Graphopt physical-simulation layout.
type GraphoptOptions struct {
	Seed               *uint64
	NIter              int
	NodeCharge         *float64
	NodeMass           float64
	SpringLength       float64
	SpringConstant     float64
	MaxMovement        float64
	InitialCoordinates *Matrix
}

// DrLPreset selects a pinned igraph DrL parameter template.
type DrLPreset int

const (
	DrLDefault DrLPreset = iota
	DrLCoarsen
	DrLCoarsest
	DrLRefine
	DrLFinal
)

// DrLOptions controls the scalable DrL layout. Preset selects a Go-native
// parameter template; no C option type escapes the package.
type DrLOptions struct {
	Preset             DrLPreset
	Seed               *uint64
	Weights            []float64
	InitialCoordinates *Matrix
}

// LGLOptions controls the Large Graph Layout algorithm.
type LGLOptions struct {
	Seed            *uint64
	MaxIter         int
	MaxDelta        float64
	Area            float64
	CoolingExponent float64
	RepulsionRadius float64
	CellSize        float64
	Root            *int
}

// TreeRootChoice selects the heuristic used to choose one root per relevant
// component for tree layouts.
type TreeRootChoice int

const (
	TreeRootByDegree TreeRootChoice = iota
	TreeRootByEccentricity
)

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
	maxIter := options.MaxIter
	if maxIter == 0 {
		maxIter = 100
	}

	hGap := options.HGap
	if hGap == 0 {
		hGap = 1.0
	}
	vGap := options.VGap
	if vGap == 0 {
		vGap = 1.0
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

	cMaxIter, err := intToIgraphInt(maxIter, "maxIter")
	if err != nil {
		return Matrix{}, err
	}

	code := C.go_igraph_layout_sugiyama(
		&g.graph,
		&cMat.value,
		nil,
		cLayersPtr,
		C.igraph_real_t(hGap),
		C.igraph_real_t(vGap),
		cMaxIter,
		edgeWeightPointer(weightsVec),
	)
	if code != C.IGRAPH_SUCCESS {
		return Matrix{}, igraphError("calculate Sugiyama layout", int(code))
	}
	return cMat.matrix()
}

// newOptionalCoordinateBound validates and marshals one per-axis coordinate
// bound slice. NaN entries are rejected because upstream comparisons treat
// them as "no bound" and would silently ignore them; ±Inf is allowed and
// disables the bound on that side.
func newOptionalCoordinateBound(name string, values []float64, vertexCount int) (*realVector, error) {
	if len(values) == 0 {
		return nil, nil
	}
	if len(values) != vertexCount {
		return nil, fmt.Errorf("igraph: %s length %d does not match vertex count %d", name, len(values), vertexCount)
	}
	for index, value := range values {
		if math.IsNaN(value) {
			return nil, fmt.Errorf("igraph: %s at index %d must not be NaN", name, index)
		}
	}
	return newRealVector(values)
}

// forceLayoutBounds groups the per-axis coordinate bound slices shared by the
// force-directed layout options.
type forceLayoutBounds struct {
	minX, maxX, minY, maxY, minZ, maxZ []float64
}

// forceLayoutInputs marshals the inputs shared by the force-directed layout
// bindings: optional edge weights, per-axis coordinate bounds, and the
// coordinate matrix optionally seeded from initial coordinates.
type forceLayoutInputs struct {
	weights *realVector
	minX    *realVector
	maxX    *realVector
	minY    *realVector
	maxY    *realVector
	minZ    *realVector
	maxZ    *realVector
	coords  *cMatrix
	useSeed C.igraph_bool_t
}

func newForceLayoutInputs(weights []float64, bounds forceLayoutBounds, initial *Matrix, dim, numVertices, numEdges int) (*forceLayoutInputs, error) {
	if dim == 2 && (len(bounds.minZ) > 0 || len(bounds.maxZ) > 0) {
		return nil, fmt.Errorf("igraph: MinZ and MaxZ apply only to 3D layouts")
	}

	in := &forceLayoutInputs{}
	ok := false
	defer func() {
		if !ok {
			in.close()
		}
	}()

	var err error
	if in.weights, err = newOptionalEdgeWeights(weights, numEdges); err != nil {
		return nil, err
	}
	axes := []struct {
		name   string
		values []float64
		dst    **realVector
	}{
		{"MinX", bounds.minX, &in.minX},
		{"MaxX", bounds.maxX, &in.maxX},
		{"MinY", bounds.minY, &in.minY},
		{"MaxY", bounds.maxY, &in.maxY},
		{"MinZ", bounds.minZ, &in.minZ},
		{"MaxZ", bounds.maxZ, &in.maxZ},
	}
	for _, axis := range axes {
		if *axis.dst, err = newOptionalCoordinateBound(axis.name, axis.values, numVertices); err != nil {
			return nil, err
		}
	}
	if initial != nil {
		rows, cols := initial.Dims()
		if rows != numVertices || cols != dim {
			return nil, fmt.Errorf("igraph: initial coordinates matrix dimensions (%d, %d) do not match vertex count %d and dimension %d", rows, cols, numVertices, dim)
		}
		if in.coords, err = newCMatrix(*initial); err != nil {
			return nil, err
		}
		in.useSeed = booltoint(true)
	} else {
		if in.coords, err = newCMatrix(Matrix{}); err != nil {
			return nil, err
		}
		in.useSeed = booltoint(false)
	}
	ok = true
	return in, nil
}

func (in *forceLayoutInputs) close() {
	for _, vector := range []*realVector{in.weights, in.minX, in.maxX, in.minY, in.maxY, in.minZ, in.maxZ} {
		if vector != nil {
			vector.close()
		}
	}
	if in.coords != nil {
		in.coords.close()
	}
}

func (options FruchtermanReingoldOptions) bounds() forceLayoutBounds {
	return forceLayoutBounds{options.MinX, options.MaxX, options.MinY, options.MaxY, options.MinZ, options.MaxZ}
}

func (options KamadaKawaiOptions) bounds() forceLayoutBounds {
	return forceLayoutBounds{options.MinX, options.MaxX, options.MinY, options.MaxY, options.MinZ, options.MaxZ}
}

// LayoutFruchtermanReingold computes a 2D force-directed layout using the Fruchterman-Reingold algorithm.
// Public input slices and matrices are borrowed/copied and returned values are Go-owned.
//
//igraph:bind igraph_layout_fruchterman_reingold
func (g *Graph) LayoutFruchtermanReingold(options FruchtermanReingoldOptions) (Matrix, error) {
	return g.layoutFruchtermanReingold(options, 2)
}

// LayoutFruchtermanReingold3D computes a 3D force-directed layout using the Fruchterman-Reingold algorithm.
// It shares FruchtermanReingoldOptions with the 2D variant; InitialCoordinates must have
// 3 columns and MinZ/MaxZ bound the third axis.
// Public input slices and matrices are borrowed/copied and returned values are Go-owned.
//
//igraph:bind igraph_layout_fruchterman_reingold_3d
func (g *Graph) LayoutFruchtermanReingold3D(options FruchtermanReingoldOptions) (Matrix, error) {
	return g.layoutFruchtermanReingold(options, 3)
}

func (g *Graph) layoutFruchtermanReingold(options FruchtermanReingoldOptions, dim int) (Matrix, error) {
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
	nIter := options.NIter
	if nIter == 0 {
		nIter = 500
	}

	if options.StartTemp < 0 {
		return Matrix{}, fmt.Errorf("igraph: StartTemp must be non-negative: %v", options.StartTemp)
	}
	startTemp := options.StartTemp
	if startTemp == 0 {
		startTemp = math.Sqrt(float64(numVertices))
	}

	inputs, err := newForceLayoutInputs(options.Weights, options.bounds(), options.InitialCoordinates, dim, numVertices, numEdges)
	if err != nil {
		return Matrix{}, err
	}
	defer inputs.close()

	cNIter, err := intToIgraphInt(nIter, "NIter")
	if err != nil {
		return Matrix{}, err
	}

	operation := "calculate Fruchterman-Reingold layout"
	if dim == 3 {
		operation = "calculate 3D Fruchterman-Reingold layout"
	}

	var runErr error
	errRNG := withRNG(options.Seed, func() error {
		code := C.go_igraph_layout_fruchterman_reingold(
			&g.graph,
			&inputs.coords.value,
			inputs.useSeed,
			cNIter,
			C.igraph_real_t(startTemp),
			edgeWeightPointer(inputs.weights),
			edgeWeightPointer(inputs.minX),
			edgeWeightPointer(inputs.maxX),
			edgeWeightPointer(inputs.minY),
			edgeWeightPointer(inputs.maxY),
			edgeWeightPointer(inputs.minZ),
			edgeWeightPointer(inputs.maxZ),
			C.int(dim),
		)
		if code != C.IGRAPH_SUCCESS {
			runErr = igraphError(operation, int(code))
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

	return inputs.coords.matrix()
}

// LayoutKamadaKawai computes a 2D force-directed layout using the Kamada-Kawai algorithm.
// Public input slices and matrices are borrowed/copied and returned values are Go-owned.
//
//igraph:bind igraph_layout_kamada_kawai
func (g *Graph) LayoutKamadaKawai(options KamadaKawaiOptions) (Matrix, error) {
	return g.layoutKamadaKawai(options, 2)
}

// LayoutKamadaKawai3D computes a 3D force-directed layout using the Kamada-Kawai algorithm.
// It shares KamadaKawaiOptions with the 2D variant; InitialCoordinates must have
// 3 columns and MinZ/MaxZ bound the third axis.
// Public input slices and matrices are borrowed/copied and returned values are Go-owned.
//
//igraph:bind igraph_layout_kamada_kawai_3d
func (g *Graph) LayoutKamadaKawai3D(options KamadaKawaiOptions) (Matrix, error) {
	return g.layoutKamadaKawai(options, 3)
}

func (g *Graph) layoutKamadaKawai(options KamadaKawaiOptions, dim int) (Matrix, error) {
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

	if options.Epsilon < 0 {
		return Matrix{}, fmt.Errorf("igraph: Epsilon must be non-negative: %v", options.Epsilon)
	}

	if options.KKConst < 0 {
		return Matrix{}, fmt.Errorf("igraph: KKConst must be non-negative: %v", options.KKConst)
	}
	kkconst := options.KKConst
	if kkconst == 0 {
		kkconst = float64(numVertices)
		if kkconst == 0 {
			kkconst = 1
		}
	}

	inputs, err := newForceLayoutInputs(options.Weights, options.bounds(), options.InitialCoordinates, dim, numVertices, numEdges)
	if err != nil {
		return Matrix{}, err
	}
	defer inputs.close()

	cMaxIter, err := intToIgraphInt(maxIter, "MaxIter")
	if err != nil {
		return Matrix{}, err
	}

	operation := "calculate Kamada-Kawai layout"
	if dim == 3 {
		operation = "calculate 3D Kamada-Kawai layout"
	}

	var runErr error
	errRNG := withRNG(options.Seed, func() error {
		code := C.go_igraph_layout_kamada_kawai(
			&g.graph,
			&inputs.coords.value,
			inputs.useSeed,
			cMaxIter,
			C.igraph_real_t(options.Epsilon),
			C.igraph_real_t(kkconst),
			edgeWeightPointer(inputs.weights),
			edgeWeightPointer(inputs.minX),
			edgeWeightPointer(inputs.maxX),
			edgeWeightPointer(inputs.minY),
			edgeWeightPointer(inputs.maxY),
			edgeWeightPointer(inputs.minZ),
			edgeWeightPointer(inputs.maxZ),
			C.int(dim),
		)
		if code != C.IGRAPH_SUCCESS {
			runErr = igraphError(operation, int(code))
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

	return inputs.coords.matrix()
}

// LayoutMDS computes a layout using Multi-Dimensional Scaling based on vertex distance matrix.
// If distances is nil, shortest path distances between all vertex pairs are computed automatically.
// If distances is non-nil, its dimensions must be square and match vertex count, and it must be
// symmetric: upstream does not verify symmetry and leaves the result for asymmetric input
// unspecified, so asymmetric matrices are rejected here.
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
		for r := 0; r < rows; r++ {
			for c := r + 1; c < cols; c++ {
				upper, _ := distances.At(r, c)
				lower, _ := distances.At(c, r)
				if upper != lower {
					return Matrix{}, fmt.Errorf("igraph: distance matrix must be symmetric: element (%d, %d) is %v but element (%d, %d) is %v", r, c, upper, c, r, lower)
				}
			}
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

// LayoutRandom3D computes a 3D random layout where coordinates are uniformly distributed.
// LayoutRandomOptions allows specifying a random seed for thread-safe reproducible execution.
// Returned values are Go-owned.
//
//igraph:bind igraph_layout_random_3d
func (g *Graph) LayoutRandom3D(options LayoutRandomOptions) (Matrix, error) {
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
		code := C.go_igraph_layout_random_3d(&g.graph, &cMat.value)
		if code != C.IGRAPH_SUCCESS {
			return igraphError("calculate 3D random layout", int(code))
		}
		return nil
	})
	if errRNG != nil {
		return Matrix{}, errRNG
	}
	return cMat.matrix()
}

// LayoutGrid3D computes a 3D regular grid layout with the specified width and height;
// vertices fill each width x height layer before starting the next one. A width or
// height of zero lets upstream choose the extent automatically.
// Returned values are Go-owned.
//
//igraph:bind igraph_layout_grid_3d
func (g *Graph) LayoutGrid3D(width int, height int) (Matrix, error) {
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
	if height < 0 {
		return Matrix{}, fmt.Errorf("igraph: height must be non-negative: %d", height)
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
	cHeight, err := intToIgraphInt(height, "grid height")
	if err != nil {
		return Matrix{}, err
	}

	code := C.go_igraph_layout_grid_3d(&g.graph, &cMat.value, cWidth, cHeight)
	if code != C.IGRAPH_SUCCESS {
		return Matrix{}, igraphError("calculate 3D grid layout", int(code))
	}
	return cMat.matrix()
}

// LayoutSphere places vertices approximately uniformly on the surface of a unit sphere.
// The placement is deterministic. Returned values are Go-owned.
//
//igraph:bind igraph_layout_sphere
func (g *Graph) LayoutSphere() (Matrix, error) {
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

	code := C.go_igraph_layout_sphere(&g.graph, &cMat.value)
	if code != C.IGRAPH_SUCCESS {
		return Matrix{}, igraphError("calculate sphere layout", int(code))
	}
	return cMat.matrix()
}

func layoutOptionValue(value *float64, fallback float64) float64 {
	if value == nil {
		return fallback
	}
	return *value
}

func newLayoutSeedMatrix(initial *Matrix, vertices, dimensions int) (*cMatrix, C.igraph_bool_t, error) {
	if initial == nil && vertices == 0 {
		empty, err := NewMatrix(0, dimensions)
		if err != nil {
			return nil, booltoint(false), err
		}
		matrix, err := newCMatrix(empty)
		return matrix, booltoint(true), err
	}
	if initial == nil {
		matrix, err := newCMatrix(Matrix{})
		return matrix, booltoint(false), err
	}
	rows, columns := initial.Dims()
	if rows != vertices || columns != dimensions {
		return nil, booltoint(false), fmt.Errorf("igraph: initial coordinates matrix dimensions (%d, %d) do not match vertex count %d and dimension %d", rows, columns, vertices, dimensions)
	}
	for row, values := range initial.Rows() {
		for column, value := range values {
			if math.IsNaN(value) || math.IsInf(value, 0) {
				return nil, booltoint(false), fmt.Errorf("igraph: initial coordinate at row %d, column %d must be finite", row, column)
			}
		}
	}
	matrix, err := newCMatrix(*initial)
	return matrix, booltoint(true), err
}

func finiteNonNegativeLayoutOption(name string, value float64) error {
	if math.IsNaN(value) || math.IsInf(value, 0) || value < 0 {
		return fmt.Errorf("igraph: %s must be finite and non-negative: %v", name, value)
	}
	return nil
}

// LayoutDavidsonHarel computes a 2D simulated-annealing layout. Edge
// directions are ignored. InitialCoordinates is borrowed and copied; the
// returned V-by-2 matrix is Go-owned. Zero iteration and cooling fields select
// the pinned igraph defaults. Nil energy-weight fields select graph-dependent
// defaults; a non-nil pointer may explicitly select zero.
//
//igraph:bind igraph_layout_davidson_harel
func (g *Graph) LayoutDavidsonHarel(options DavidsonHarelOptions) (Matrix, error) {
	if g == nil {
		return Matrix{}, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return Matrix{}, ErrClosed
	}

	vertices := int(C.igraph_vcount(&g.graph))
	edges := int(C.igraph_ecount(&g.graph))
	if options.MaxIter < 0 || options.FineIter < 0 {
		return Matrix{}, fmt.Errorf("igraph: Davidson-Harel iteration counts must be non-negative")
	}
	maxIter := options.MaxIter
	if maxIter == 0 {
		maxIter = 10
	}
	fineIter := options.FineIter
	if fineIter == 0 {
		fineIter = 10
		if vertices > 1 && int(math.Ceil(math.Log2(float64(vertices)))) > fineIter {
			fineIter = int(math.Ceil(math.Log2(float64(vertices))))
		}
	}
	cool := options.CoolFactor
	if cool == 0 {
		cool = 0.75
	}
	if math.IsNaN(cool) || math.IsInf(cool, 0) || cool <= 0 || cool >= 1 {
		return Matrix{}, fmt.Errorf("igraph: CoolFactor must be finite and in (0, 1): %v", cool)
	}
	density := 0.0
	if vertices > 1 {
		density = float64(edges) / (float64(vertices) * float64(vertices-1) / 2)
		if density > 1 {
			density = 1
		}
	}
	weights := []struct {
		name  string
		value float64
	}{
		{"WeightNodeDistance", layoutOptionValue(options.WeightNodeDistance, 1)},
		{"WeightBorder", layoutOptionValue(options.WeightBorder, 0)},
		{"WeightEdgeLengths", layoutOptionValue(options.WeightEdgeLengths, density/10)},
		{"WeightEdgeCrossings", layoutOptionValue(options.WeightEdgeCrossings, 1-math.Sqrt(density))},
		{"WeightNodeEdgeDistance", layoutOptionValue(options.WeightNodeEdgeDistance, (1-density)/5)},
	}
	for _, weight := range weights {
		if err := finiteNonNegativeLayoutOption(weight.name, weight.value); err != nil {
			return Matrix{}, err
		}
	}
	coords, useSeed, err := newLayoutSeedMatrix(options.InitialCoordinates, vertices, 2)
	if err != nil {
		return Matrix{}, err
	}
	defer coords.close()
	cMaxIter, err := intToIgraphInt(maxIter, "MaxIter")
	if err != nil {
		return Matrix{}, err
	}
	cFineIter, err := intToIgraphInt(fineIter, "FineIter")
	if err != nil {
		return Matrix{}, err
	}
	err = withRNG(options.Seed, func() error {
		code := C.go_igraph_layout_davidson_harel(&g.graph, &coords.value, useSeed,
			cMaxIter, cFineIter, C.igraph_real_t(cool), C.igraph_real_t(weights[0].value),
			C.igraph_real_t(weights[1].value), C.igraph_real_t(weights[2].value),
			C.igraph_real_t(weights[3].value), C.igraph_real_t(weights[4].value))
		if code != C.IGRAPH_SUCCESS {
			return igraphError("calculate Davidson-Harel layout", int(code))
		}
		return nil
	})
	if err != nil {
		return Matrix{}, err
	}
	return coords.matrix()
}

// LayoutGEM computes a 2D GEM layout. Edge directions are ignored.
// InitialCoordinates is borrowed and copied; the returned matrix is Go-owned.
// Zero-valued scalar fields select the pinned igraph defaults.
//
//igraph:bind igraph_layout_gem
func (g *Graph) LayoutGEM(options GEMOptions) (Matrix, error) {
	if g == nil {
		return Matrix{}, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return Matrix{}, ErrClosed
	}
	vertices := int(C.igraph_vcount(&g.graph))
	if options.MaxIter < 0 {
		return Matrix{}, fmt.Errorf("igraph: MaxIter must be non-negative: %d", options.MaxIter)
	}
	maxIter := options.MaxIter
	if maxIter == 0 {
		maxIter = 40 * vertices * vertices
	}
	tempMax := options.MaxTemperature
	if tempMax == 0 {
		tempMax = math.Max(1, float64(vertices))
	}
	tempMin := options.MinTemperature
	if tempMin == 0 {
		tempMin = 0.1
	}
	tempInit := options.InitialTemperature
	if tempInit == 0 {
		tempInit = math.Sqrt(math.Max(1, float64(vertices)))
	}
	for name, value := range map[string]float64{"MaxTemperature": tempMax, "MinTemperature": tempMin, "InitialTemperature": tempInit} {
		if math.IsNaN(value) || math.IsInf(value, 0) || value <= 0 {
			return Matrix{}, fmt.Errorf("igraph: %s must be finite and positive: %v", name, value)
		}
	}
	if tempMin > tempInit || tempInit > tempMax {
		return Matrix{}, fmt.Errorf("igraph: GEM temperatures must satisfy MinTemperature <= InitialTemperature <= MaxTemperature")
	}
	coords, useSeed, err := newLayoutSeedMatrix(options.InitialCoordinates, vertices, 2)
	if err != nil {
		return Matrix{}, err
	}
	defer coords.close()
	cMaxIter, err := intToIgraphInt(maxIter, "MaxIter")
	if err != nil {
		return Matrix{}, err
	}
	err = withRNG(options.Seed, func() error {
		code := C.go_igraph_layout_gem(&g.graph, &coords.value, useSeed, cMaxIter,
			C.igraph_real_t(tempMax), C.igraph_real_t(tempMin), C.igraph_real_t(tempInit))
		if code != C.IGRAPH_SUCCESS {
			return igraphError("calculate GEM layout", int(code))
		}
		return nil
	})
	if err != nil {
		return Matrix{}, err
	}
	return coords.matrix()
}

// LayoutGraphopt computes a 2D Graphopt layout. InitialCoordinates is borrowed
// and copied; the returned matrix is Go-owned. Zero scalar fields select the
// original Graphopt defaults. Set NodeCharge to a non-nil pointer to explicitly
// choose a charge, including zero. Random initialization follows Seed.
//
//igraph:bind igraph_layout_graphopt
func (g *Graph) LayoutGraphopt(options GraphoptOptions) (Matrix, error) {
	if g == nil {
		return Matrix{}, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return Matrix{}, ErrClosed
	}
	vertices := int(C.igraph_vcount(&g.graph))
	if options.NIter < 0 {
		return Matrix{}, fmt.Errorf("igraph: NIter must be non-negative: %d", options.NIter)
	}
	nIter := options.NIter
	if nIter == 0 {
		nIter = 500
	}
	charge := layoutOptionValue(options.NodeCharge, 0.001)
	mass := options.NodeMass
	if mass == 0 {
		mass = 30
	}
	springConstant := options.SpringConstant
	if springConstant == 0 {
		springConstant = 1
	}
	maxMovement := options.MaxMovement
	if maxMovement == 0 {
		maxMovement = 5
	}
	for name, value := range map[string]float64{"NodeCharge": charge, "SpringLength": options.SpringLength, "SpringConstant": springConstant, "MaxMovement": maxMovement} {
		if err := finiteNonNegativeLayoutOption(name, value); err != nil {
			return Matrix{}, err
		}
	}
	if math.IsNaN(mass) || math.IsInf(mass, 0) || mass <= 0 {
		return Matrix{}, fmt.Errorf("igraph: NodeMass must be finite and positive: %v", mass)
	}
	coords, useSeed, err := newLayoutSeedMatrix(options.InitialCoordinates, vertices, 2)
	if err != nil {
		return Matrix{}, err
	}
	defer coords.close()
	cNIter, err := intToIgraphInt(nIter, "NIter")
	if err != nil {
		return Matrix{}, err
	}
	err = withRNG(options.Seed, func() error {
		code := C.go_igraph_layout_graphopt(&g.graph, &coords.value, cNIter,
			C.igraph_real_t(charge), C.igraph_real_t(mass), C.igraph_real_t(options.SpringLength),
			C.igraph_real_t(springConstant), C.igraph_real_t(maxMovement), useSeed)
		if code != C.IGRAPH_SUCCESS {
			return igraphError("calculate Graphopt layout", int(code))
		}
		return nil
	})
	if err != nil {
		return Matrix{}, err
	}
	return coords.matrix()
}

// LayoutDrL computes a scalable 2D DrL layout. Weights and initial coordinates
// are borrowed and copied; the returned V-by-2 matrix is Go-owned.
//
//igraph:bind igraph_layout_drl
func (g *Graph) LayoutDrL(options DrLOptions) (Matrix, error) {
	return g.layoutDrL(options, 2)
}

// LayoutDrL3D computes a scalable 3D DrL layout. It shares DrLOptions with the
// 2D variant; initial coordinates must have three columns.
//
//igraph:bind igraph_layout_drl_3d
func (g *Graph) LayoutDrL3D(options DrLOptions) (Matrix, error) {
	return g.layoutDrL(options, 3)
}

//igraph:internal igraph_layout_drl_options_init
func (g *Graph) layoutDrL(options DrLOptions, dimensions int) (Matrix, error) {
	if g == nil {
		return Matrix{}, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return Matrix{}, ErrClosed
	}
	if options.Preset < DrLDefault || options.Preset > DrLFinal {
		return Matrix{}, fmt.Errorf("igraph: invalid DrL preset: %d", options.Preset)
	}
	vertices := int(C.igraph_vcount(&g.graph))
	edges := int(C.igraph_ecount(&g.graph))
	for index, weight := range options.Weights {
		if math.IsNaN(weight) || math.IsInf(weight, 0) || weight < 0 {
			return Matrix{}, fmt.Errorf("igraph: DrL weight at index %d must be finite and non-negative: %v", index, weight)
		}
	}
	weights, err := newOptionalEdgeWeights(options.Weights, edges)
	if err != nil {
		return Matrix{}, err
	}
	if weights != nil {
		defer weights.close()
	}
	coords, useSeed, err := newLayoutSeedMatrix(options.InitialCoordinates, vertices, dimensions)
	if err != nil {
		return Matrix{}, err
	}
	defer coords.close()
	err = withRNG(options.Seed, func() error {
		code := C.go_igraph_layout_drl(
			&g.graph, &coords.value, useSeed, C.int(options.Preset),
			edgeWeightPointer(weights), C.int(dimensions),
		)
		if code != C.IGRAPH_SUCCESS {
			return igraphError("calculate DrL layout", int(code))
		}
		return nil
	})
	if err != nil {
		return Matrix{}, err
	}
	return coords.matrix()
}

// LayoutLGL computes a 2D Large Graph Layout. LGL ignores edge direction and
// is intended for connected graphs. If Root is nil, upstream chooses a random
// root under the optional reproducible Seed contract. The returned matrix is
// Go-owned.
//
//igraph:bind igraph_layout_lgl
func (g *Graph) LayoutLGL(options LGLOptions) (Matrix, error) {
	if g == nil {
		return Matrix{}, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return Matrix{}, ErrClosed
	}
	vertices := int(C.igraph_vcount(&g.graph))
	if vertices == 0 {
		return NewMatrix(0, 2)
	}
	if options.MaxIter < 0 {
		return Matrix{}, fmt.Errorf("igraph: MaxIter must be non-negative: %d", options.MaxIter)
	}
	maxIter := options.MaxIter
	if maxIter == 0 {
		maxIter = 150
	}
	maxDelta := options.MaxDelta
	if maxDelta == 0 {
		maxDelta = float64(vertices)
	}
	area := options.Area
	if area == 0 {
		area = float64(vertices) * float64(vertices)
	}
	cooling := options.CoolingExponent
	if cooling == 0 {
		cooling = 1.5
	}
	repulsion := options.RepulsionRadius
	if repulsion == 0 {
		repulsion = area * float64(vertices)
	}
	cellSize := options.CellSize
	if cellSize == 0 {
		cellSize = math.Pow(area, 0.25)
	}
	for name, value := range map[string]float64{
		"MaxDelta": maxDelta, "Area": area, "CoolingExponent": cooling,
		"RepulsionRadius": repulsion, "CellSize": cellSize,
	} {
		if math.IsNaN(value) || math.IsInf(value, 0) || value <= 0 {
			return Matrix{}, fmt.Errorf("igraph: %s must be finite and positive: %v", name, value)
		}
	}
	root := -1
	if options.Root != nil {
		root = *options.Root
		if root < 0 || root >= vertices {
			return Matrix{}, fmt.Errorf("igraph: LGL root vertex ID %d is out of bounds for graph with %d vertices", root, vertices)
		}
	}
	cMaxIter, err := intToIgraphInt(maxIter, "MaxIter")
	if err != nil {
		return Matrix{}, err
	}
	cRoot, err := intToIgraphInt(root, "Root")
	if err != nil {
		return Matrix{}, err
	}
	coords, err := newCMatrix(Matrix{})
	if err != nil {
		return Matrix{}, err
	}
	defer coords.close()
	err = withRNG(options.Seed, func() error {
		code := C.go_igraph_layout_lgl(
			&g.graph, &coords.value, cMaxIter, C.igraph_real_t(maxDelta),
			C.igraph_real_t(area), C.igraph_real_t(cooling), C.igraph_real_t(repulsion),
			C.igraph_real_t(cellSize), cRoot,
		)
		if code != C.IGRAPH_SUCCESS {
			return igraphError("calculate LGL layout", int(code))
		}
		return nil
	})
	if err != nil {
		return Matrix{}, err
	}
	return coords.matrix()
}

// AlignLayout translates and rotates borrowed vertex coordinates into an
// axis-aligned orientation. Any positive coordinate dimension is supported.
// The input is never mutated and the returned matrix is independently Go-owned.
//
//igraph:bind igraph_layout_align
func (g *Graph) AlignLayout(coordinates Matrix) (Matrix, error) {
	if g == nil {
		return Matrix{}, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return Matrix{}, ErrClosed
	}
	vertices := int(C.igraph_vcount(&g.graph))
	rows, columns := coordinates.Dims()
	if rows != vertices {
		return Matrix{}, fmt.Errorf("igraph: coordinate row count %d does not match vertex count %d", rows, vertices)
	}
	if vertices > 0 && columns == 0 {
		return Matrix{}, fmt.Errorf("igraph: layout coordinates must have at least one dimension")
	}
	for row, values := range coordinates.Rows() {
		for column, value := range values {
			if math.IsNaN(value) || math.IsInf(value, 0) {
				return Matrix{}, fmt.Errorf("igraph: coordinate at row %d, column %d must be finite", row, column)
			}
		}
	}
	cLayout, err := newCMatrix(coordinates)
	if err != nil {
		return Matrix{}, err
	}
	defer cLayout.close()
	if code := C.go_igraph_layout_align(&g.graph, &cLayout.value); code != C.IGRAPH_SUCCESS {
		return Matrix{}, igraphError("align layout", int(code))
	}
	return cLayout.matrix()
}

// TreeLayoutRoots returns a minimal Go-owned root set suitable for the
// Reingold-Tilford layouts. Undirected graphs ignore mode. Degree chooses the
// highest-degree candidate; eccentricity chooses the lowest-eccentricity one.
// Roots follow the pinned upstream component/priority order; equal-score ties
// retain upstream's deterministic vertex ordering.
//
//igraph:bind igraph_roots_for_tree_layout
func (g *Graph) TreeLayoutRoots(mode DegMode, choice TreeRootChoice) ([]int, error) {
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
	if choice < TreeRootByDegree || choice > TreeRootByEccentricity {
		return nil, fmt.Errorf("igraph: invalid tree root choice: %d", choice)
	}
	roots, err := newIntVector(nil)
	if err != nil {
		return nil, err
	}
	defer roots.close()
	if code := C.go_igraph_roots_for_tree_layout(&g.graph, cMode, &roots.value, C.int(choice)); code != C.IGRAPH_SUCCESS {
		return nil, igraphError("choose tree layout roots", int(code))
	}
	result, err := roots.slice()
	if err != nil {
		return nil, err
	}
	if result == nil {
		result = []int{}
	}
	return result, nil
}
