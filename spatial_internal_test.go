package igraph

import (
	"errors"
	"math"
	"testing"
)

func TestValidateNearestNeighborOptions(t *testing.T) {
	maximum, cutoff := 0, 0.0
	validated, err := validateNearestNeighborOptions(NearestNeighborOptions{
		Metric: SpatialManhattan, MaxNeighbors: &maximum, Cutoff: &cutoff, Directed: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if validated.metric != SpatialManhattan || validated.maxNeighbors != 0 || validated.cutoff != 0 || !validated.directed {
		t.Fatalf("validated options = %#v", validated)
	}
	defaults, err := validateNearestNeighborOptions(NearestNeighborOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if defaults.metric != SpatialEuclidean || defaults.maxNeighbors != -1 || defaults.cutoff != -1 || defaults.directed {
		t.Fatalf("validated defaults = %#v", defaults)
	}

	negativeMaximum := -1
	negativeCutoff := -1.0
	for name, options := range map[string]NearestNeighborOptions{
		"metric":          {Metric: SpatialMetric(99)},
		"neighbors":       {MaxNeighbors: &negativeMaximum},
		"cutoff":          {Cutoff: &negativeCutoff},
		"nan cutoff":      {Cutoff: float64Pointer(math.NaN())},
		"infinite cutoff": {Cutoff: float64Pointer(math.Inf(1))},
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := validateNearestNeighborOptions(options); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}

func TestSpatialPointValidationAndCopy(t *testing.T) {
	points, _ := NewMatrixFromRows([][]float64{{0, 1}, {2, 3}})
	cPoints, err := newSpatialPoints(points, spatialPointRequirements{
		operation: "test geometry", exactDimensions: 2, distinct: true,
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer cPoints.close()
	copy, err := cPoints.matrix()
	if err != nil {
		t.Fatal(err)
	}
	want, got := points.Rows(), copy.Rows()
	for row := range want {
		for column := range want[row] {
			if want[row][column] != got[row][column] {
				t.Fatalf("copied point (%d, %d) = %v, want %v", row, column, got[row][column], want[row][column])
			}
		}
	}

	empty := Matrix{}
	if err := validateSpatialPoints(empty, spatialPointRequirements{exactDimensions: 2, minDimensions: 1}); err != nil {
		t.Fatalf("empty point set error = %v", err)
	}
}

func TestSpatialPointValidationErrors(t *testing.T) {
	duplicate, _ := NewMatrixFromRows([][]float64{{1, 2}, {1, 2}})
	oneDimensional, _ := NewMatrixFromRows([][]float64{{1}})
	nonFinite, _ := NewMatrixFromRows([][]float64{{0, math.Inf(-1)}})
	corrupt := Matrix{rows: 1, columns: 2, values: []float64{1}}

	for name, test := range map[string]struct {
		points       Matrix
		requirements spatialPointRequirements
	}{
		"duplicate":          {duplicate, spatialPointRequirements{operation: "test", distinct: true}},
		"exact dimensions":   {oneDimensional, spatialPointRequirements{operation: "test", exactDimensions: 2}},
		"minimum dimensions": {oneDimensional, spatialPointRequirements{operation: "test", minDimensions: 2}},
		"non-finite":         {nonFinite, spatialPointRequirements{}},
		"corrupt":            {corrupt, spatialPointRequirements{}},
	} {
		t.Run(name, func(t *testing.T) {
			if err := validateSpatialPoints(test.points, test.requirements); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}

	failure := errors.New("injected matrix initialization failure")
	valid, _ := NewMatrixFromRows([][]float64{{0, 0}})
	if _, err := newSpatialPoints(valid, spatialPointRequirements{}, func(Matrix) (*cMatrix, error) {
		return nil, failure
	}); !errors.Is(err, failure) {
		t.Fatalf("matrix initialization error = %v", err)
	}
}

func TestConvertAndAdoptSpatialGraphCleanup(t *testing.T) {
	failure := errors.New("injected conversion failure")
	for name, convert := range map[string]func() ([]float64, error){
		"conversion": func() ([]float64, error) { return nil, failure },
		"alignment":  func() ([]float64, error) { return []float64{1}, nil },
	} {
		t.Run(name, func(t *testing.T) {
			destroyed, adopted := false, false
			_, err := convertAndAdoptSpatialGraph(nil, spatialGraphValueAdapters{
				edgeCount: func() int { return 2 },
				convert:   convert,
				destroy:   func() { destroyed = true },
				adopt: func() *Graph {
					adopted = true
					return &Graph{}
				},
			})
			if err == nil || !destroyed || adopted {
				t.Fatalf("error = %v, destroyed = %t, adopted = %t", err, destroyed, adopted)
			}
		})
	}

	result, err := convertAndAdoptSpatialGraph(nil, spatialGraphValueAdapters{
		edgeCount: func() int { return 0 },
		convert:   func() ([]float64, error) { return nil, nil },
		adopt:     func() *Graph { return &Graph{} },
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.graph == nil || result.values == nil || len(result.values) != 0 {
		t.Fatalf("successful spatial result = %#v", result)
	}
}

func float64Pointer(value float64) *float64 { return &value }
