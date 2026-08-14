package igraph

// #cgo pkg-config: igraph
// #include <igraph.h>
// #include "spatial_cgo.h"
import "C"

import (
	"fmt"
	"math"
)

// ConvexHullResult contains the source point-row indices and coordinates of a
// two-dimensional convex hull in traversal order. PointIndices[i] identifies
// Coordinates row i. Both values are non-nil and Go-owned and retain no input
// or C storage.
type ConvexHullResult struct {
	PointIndices []int
	Coordinates  Matrix
}

// ConvexHull2D computes the convex hull of a two-dimensional point set. Row i
// of points identifies point i. The input Matrix is borrowed only for the
// synchronous call and copied into temporary C storage. Empty input returns a
// non-nil empty index slice and a 0-by-2 coordinate matrix. Collinear and
// duplicate points follow pinned igraph 1.0.1 hull selection semantics.
//
//igraph:bind igraph_convex_hull_2d
func ConvexHull2D(points Matrix) (ConvexHullResult, error) {
	return convexHull2D(points, nil)
}

// SpatialMetric selects the distance metric used by spatial graph operations.
// Its zero value is SpatialEuclidean.
type SpatialMetric uint8

const (
	// SpatialEuclidean selects ordinary straight-line (L2) distance.
	SpatialEuclidean SpatialMetric = iota
	// SpatialManhattan selects coordinate-wise absolute (L1) distance.
	SpatialManhattan
)

func (metric SpatialMetric) cValue() (C.igraph_metric_t, error) {
	switch metric {
	case SpatialEuclidean:
		return C.IGRAPH_METRIC_EUCLIDEAN, nil
	case SpatialManhattan:
		return C.IGRAPH_METRIC_MANHATTAN, nil
	default:
		return 0, fmt.Errorf("igraph: invalid spatial metric: %d", metric)
	}
}

// NearestNeighborOptions controls spatial nearest-neighbor graph construction.
// Metric defaults to SpatialEuclidean. MaxNeighbors and Cutoff are borrowed
// only for a synchronous call and are never retained. A nil bound is unlimited;
// an explicit maximum-neighbor count or cutoff must be non-negative, and the
// cutoff must be finite. Pinned igraph 1.0.1 includes only points strictly
// closer than Cutoff; a point at exactly the cutoff is excluded. Directed
// controls whether neighbor relationships retain their point-to-neighbor
// orientation.
type NearestNeighborOptions struct {
	Metric       SpatialMetric
	MaxNeighbors *int
	Cutoff       *float64
	Directed     bool
}

// NewNearestNeighborGraph constructs a spatial nearest-neighbor graph. Point
// matrix row i becomes vertex i. In a directed result, every vertex has an edge
// to at most MaxNeighbors nearest points strictly closer than Cutoff. In an
// undirected result, an edge is present when either endpoint selects the other,
// and reciprocal selections are collapsed. Self-loops and parallel edges are
// not created.
// When a maximum-neighbor boundary contains equal-distance points, which tied
// points are selected is not a compatibility promise.
//
// The point matrix and option pointers are borrowed only for the synchronous
// call and copied or read before return. The returned graph is independently
// owned and must be closed by the caller. Empty point input returns an empty
// graph, including when represented by a 0-by-0 Matrix. This binds an
// experimental API in pinned igraph 1.0.1.
//
//igraph:bind igraph_nearest_neighbor_graph
func NewNearestNeighborGraph(points Matrix, options NearestNeighborOptions) (*Graph, error) {
	return newNearestNeighborGraph(points, options, nil)
}

type validatedNearestNeighborOptions struct {
	metric       SpatialMetric
	maxNeighbors int
	cutoff       float64
	directed     bool
}

func validateNearestNeighborOptions(options NearestNeighborOptions) (validatedNearestNeighborOptions, error) {
	if _, err := options.Metric.cValue(); err != nil {
		return validatedNearestNeighborOptions{}, err
	}
	maxNeighbors := -1
	if options.MaxNeighbors != nil {
		if *options.MaxNeighbors < 0 {
			return validatedNearestNeighborOptions{}, fmt.Errorf(
				"igraph: spatial maximum neighbor count must be non-negative: %d",
				*options.MaxNeighbors,
			)
		}
		if _, err := intToIgraphInt(*options.MaxNeighbors, "spatial maximum neighbor count"); err != nil {
			return validatedNearestNeighborOptions{}, err
		}
		maxNeighbors = *options.MaxNeighbors
	}
	cutoff := -1.0
	if options.Cutoff != nil {
		if math.IsNaN(*options.Cutoff) || math.IsInf(*options.Cutoff, 0) || *options.Cutoff < 0 {
			return validatedNearestNeighborOptions{}, fmt.Errorf(
				"igraph: spatial cutoff must be finite and non-negative: %v",
				*options.Cutoff,
			)
		}
		cutoff = *options.Cutoff
	}
	return validatedNearestNeighborOptions{
		metric:       options.Metric,
		maxNeighbors: maxNeighbors,
		cutoff:       cutoff,
		directed:     options.Directed,
	}, nil
}

type nearestNeighborGraphCallResult struct {
	graph C.igraph_t
	code  int
}

type nearestNeighborGraphAdapters struct {
	newMatrix func(Matrix) (*cMatrix, error)
	call      func(*cMatrix, validatedNearestNeighborOptions) nearestNeighborGraphCallResult
}

func defaultNearestNeighborGraphAdapters() nearestNeighborGraphAdapters {
	return nearestNeighborGraphAdapters{
		newMatrix: newCMatrix,
		call: func(points *cMatrix, options validatedNearestNeighborOptions) nearestNeighborGraphCallResult {
			var graph C.igraph_t
			metric, _ := options.metric.cValue()
			code := C.go_igraph_nearest_neighbor_graph(
				&graph,
				&points.value,
				metric,
				C.igraph_int_t(options.maxNeighbors),
				C.igraph_real_t(options.cutoff),
				booltoint(options.directed),
			)
			return nearestNeighborGraphCallResult{graph: graph, code: int(code)}
		},
	}
}

func newNearestNeighborGraph(points Matrix, options NearestNeighborOptions, adapters *nearestNeighborGraphAdapters) (*Graph, error) {
	validated, err := validateNearestNeighborOptions(options)
	if err != nil {
		return nil, err
	}
	resolved := defaultNearestNeighborGraphAdapters()
	if adapters != nil {
		resolved = *adapters
	}
	cPoints, err := newSpatialPoints(points, spatialPointRequirements{
		operation: "nearest-neighbor graph", minDimensions: 1,
	}, resolved.newMatrix)
	if err != nil {
		return nil, err
	}
	defer cPoints.close()
	call := resolved.call(cPoints, validated)
	if call.code != int(C.IGRAPH_SUCCESS) {
		return nil, igraphError("construct nearest-neighbor graph", call.code)
	}
	return adoptInitializedGraph(&call.graph), nil
}

type convexHullAdapters struct {
	newMatrix     func(Matrix) (*cMatrix, error)
	newInt        func([]int) (*intVector, error)
	call          func(*cMatrix, *intVector, *cMatrix) int
	convertInt    func(*intVector) ([]int, error)
	convertMatrix func(*cMatrix) (Matrix, error)
}

func defaultConvexHullAdapters() convexHullAdapters {
	return convexHullAdapters{
		newMatrix: newCMatrix,
		newInt:    newIntVector,
		call: func(points *cMatrix, indices *intVector, coordinates *cMatrix) int {
			return int(C.go_igraph_convex_hull_2d(&points.value, &indices.value, &coordinates.value))
		},
		convertInt:    (*intVector).slice,
		convertMatrix: (*cMatrix).matrix,
	}
}

func convexHull2D(points Matrix, adapters *convexHullAdapters) (ConvexHullResult, error) {
	resolved := defaultConvexHullAdapters()
	if adapters != nil {
		resolved = *adapters
	}
	if points.rows == 0 && points.columns == 0 {
		var err error
		points, err = NewMatrix(0, 2)
		if err != nil {
			return ConvexHullResult{}, err
		}
	}
	cPoints, err := newSpatialPoints(points, spatialPointRequirements{
		operation: "2D convex hull", exactDimensions: 2,
	}, resolved.newMatrix)
	if err != nil {
		return ConvexHullResult{}, err
	}
	defer cPoints.close()
	indices, err := resolved.newInt(nil)
	if err != nil {
		return ConvexHullResult{}, err
	}
	defer indices.close()
	coordinates, err := resolved.newMatrix(Matrix{})
	if err != nil {
		return ConvexHullResult{}, err
	}
	defer coordinates.close()
	if code := resolved.call(cPoints, indices, coordinates); code != int(C.IGRAPH_SUCCESS) {
		return ConvexHullResult{}, igraphError("calculate 2D convex hull", code)
	}
	pointIndices, err := resolved.convertInt(indices)
	if err != nil {
		return ConvexHullResult{}, err
	}
	pointIndices = append([]int{}, pointIndices...)
	hullCoordinates, err := resolved.convertMatrix(coordinates)
	if err != nil {
		return ConvexHullResult{}, err
	}
	rows, columns := hullCoordinates.Dims()
	if columns != 2 || rows != len(pointIndices) {
		return ConvexHullResult{}, fmt.Errorf(
			"igraph: convex hull result dimensions (%d, %d) do not align with %d point indices",
			rows, columns, len(pointIndices),
		)
	}
	return ConvexHullResult{PointIndices: pointIndices, Coordinates: hullCoordinates}, nil
}

type spatialEdgeLengthAdapters struct {
	newMatrix func(Matrix) (*cMatrix, error)
	newReal   func([]float64) (*realVector, error)
	call      func(*Graph, *realVector, *cMatrix, SpatialMetric) int
	convert   func(*realVector) ([]float64, error)
}

func defaultSpatialEdgeLengthAdapters() spatialEdgeLengthAdapters {
	return spatialEdgeLengthAdapters{
		newMatrix: newCMatrix,
		newReal:   newRealVector,
		call: func(graph *Graph, lengths *realVector, points *cMatrix, metric SpatialMetric) int {
			cMetric, _ := metric.cValue()
			return int(C.go_igraph_spatial_edge_lengths(&graph.graph, &lengths.value, &points.value, cMetric))
		},
		convert: (*realVector).slice,
	}
}

// SpatialEdgeLengths computes one distance per edge in edge-ID order. Row i of
// points contains the coordinates of vertex i and arbitrary positive
// dimensionality is supported. Loops have length zero and parallel edges have
// repeated lengths. The point matrix is borrowed and copied for the synchronous
// call; the returned non-nil slice is Go-owned and survives graph closure.
// This binds an experimental API in pinned igraph 1.0.1.
//
//igraph:bind igraph_spatial_edge_lengths
func (g *Graph) SpatialEdgeLengths(points Matrix, metric SpatialMetric) ([]float64, error) {
	return g.spatialEdgeLengths(points, metric, nil)
}

func (g *Graph) spatialEdgeLengths(points Matrix, metric SpatialMetric, adapters *spatialEdgeLengthAdapters) ([]float64, error) {
	if g == nil {
		return nil, ErrClosed
	}
	g.mu.RLock()
	defer g.mu.RUnlock()
	if g.closed {
		return nil, ErrClosed
	}
	if _, err := metric.cValue(); err != nil {
		return nil, err
	}
	vertexCount := int(C.igraph_vcount(&g.graph))
	if points.rows != vertexCount {
		return nil, fmt.Errorf(
			"igraph: spatial point count %d does not match vertex count %d",
			points.rows, vertexCount,
		)
	}
	resolved := defaultSpatialEdgeLengthAdapters()
	if adapters != nil {
		resolved = *adapters
	}
	cPoints, err := newSpatialPoints(points, spatialPointRequirements{
		operation: "spatial edge lengths", minDimensions: 1,
	}, resolved.newMatrix)
	if err != nil {
		return nil, err
	}
	defer cPoints.close()
	lengths, err := resolved.newReal(nil)
	if err != nil {
		return nil, err
	}
	defer lengths.close()
	if code := resolved.call(g, lengths, cPoints, metric); code != int(C.IGRAPH_SUCCESS) {
		return nil, igraphError("calculate spatial edge lengths", code)
	}
	values, err := resolved.convert(lengths)
	if err != nil {
		return nil, err
	}
	values = append([]float64{}, values...)
	edgeCount := int(C.igraph_ecount(&g.graph))
	if len(values) != edgeCount {
		return nil, fmt.Errorf(
			"igraph: spatial edge length count %d does not match edge count %d",
			len(values), edgeCount,
		)
	}
	return values, nil
}

type spatialPointRequirements struct {
	operation       string
	exactDimensions int
	minDimensions   int
	distinct        bool
}

// newSpatialPoints validates a Go-owned point matrix and copies it into a
// temporary C matrix. Empty 0-by-0 point sets are accepted for every operation;
// dimension requirements apply when at least one point is present.
func newSpatialPoints(
	points Matrix,
	requirements spatialPointRequirements,
	create func(Matrix) (*cMatrix, error),
) (*cMatrix, error) {
	if err := validateSpatialPoints(points, requirements); err != nil {
		return nil, err
	}
	if create == nil {
		create = newCMatrix
	}
	return create(points)
}

func validateSpatialPoints(points Matrix, requirements spatialPointRequirements) error {
	size, err := matrixSize(points.rows, points.columns)
	if err != nil {
		return err
	}
	if len(points.values) != size {
		return fmt.Errorf(
			"igraph: spatial point matrix has %d values, want %d for %d by %d dimensions",
			len(points.values), size, points.rows, points.columns,
		)
	}
	operation := requirements.operation
	if operation == "" {
		operation = "spatial operation"
	}
	if points.rows > 0 {
		if requirements.exactDimensions > 0 && points.columns != requirements.exactDimensions {
			return fmt.Errorf(
				"igraph: %s requires %d-dimensional points, got %d",
				operation, requirements.exactDimensions, points.columns,
			)
		}
		if requirements.minDimensions > 0 && points.columns < requirements.minDimensions {
			return fmt.Errorf(
				"igraph: %s requires at least %d spatial dimension(s), got %d",
				operation, requirements.minDimensions, points.columns,
			)
		}
	}
	for index, value := range points.values {
		if math.IsNaN(value) || math.IsInf(value, 0) {
			row, column := 0, 0
			if points.columns > 0 {
				row, column = index/points.columns, index%points.columns
			}
			return fmt.Errorf(
				"igraph: spatial point coordinate (%d, %d) must be finite: %v",
				row, column, value,
			)
		}
	}
	if requirements.distinct {
		for left := 0; left < points.rows; left++ {
			for right := left + 1; right < points.rows; right++ {
				equal := true
				for column := 0; column < points.columns; column++ {
					if points.values[left*points.columns+column] != points.values[right*points.columns+column] {
						equal = false
						break
					}
				}
				if equal {
					return fmt.Errorf(
						"igraph: %s requires distinct points; rows %d and %d are equal",
						operation, left, right,
					)
				}
			}
		}
	}
	return nil
}

type spatialGraphValues struct {
	graph  *Graph
	values []float64
}

type spatialGraphValueAdapters struct {
	edgeCount func() int
	convert   func() ([]float64, error)
	destroy   func()
	adopt     func() *Graph
}

// convertAndAdoptSpatialGraph transfers a successfully initialized graph only
// after its edge-aligned values have been converted and validated. Conversion
// failure or misalignment destroys the graph before returning.
func convertAndAdoptSpatialGraph(
	graph *C.igraph_t,
	adapters spatialGraphValueAdapters,
) (spatialGraphValues, error) {
	if adapters.edgeCount == nil {
		adapters.edgeCount = func() int { return int(C.igraph_ecount(graph)) }
	}
	if adapters.destroy == nil {
		adapters.destroy = func() { C.igraph_destroy(graph) }
	}
	if adapters.adopt == nil {
		adapters.adopt = func() *Graph { return adoptInitializedGraph(graph) }
	}
	values, err := adapters.convert()
	if err != nil {
		adapters.destroy()
		return spatialGraphValues{}, err
	}
	values = append([]float64{}, values...)
	edgeCount := adapters.edgeCount()
	if len(values) != edgeCount {
		adapters.destroy()
		return spatialGraphValues{}, fmt.Errorf(
			"igraph: spatial edge value count %d does not match graph edge count %d",
			len(values), edgeCount,
		)
	}
	return spatialGraphValues{graph: adapters.adopt(), values: values}, nil
}
