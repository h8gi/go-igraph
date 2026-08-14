package igraph

// #cgo pkg-config: igraph
// #include <igraph.h>
import "C"

import (
	"fmt"
	"math"
)

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
// cutoff must be finite. Directed controls whether neighbor relationships retain
// their point-to-neighbor orientation.
type NearestNeighborOptions struct {
	Metric       SpatialMetric
	MaxNeighbors *int
	Cutoff       *float64
	Directed     bool
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
