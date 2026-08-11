package igraph

import (
	"errors"
	"math"
	"reflect"
	"strconv"
	"testing"
)

func TestWeightedCliquesAndMaximum(t *testing.T) {
	graph := testCliqueGraph(t)
	defer graph.Close()
	weights := []int{1, 1, 1, 10, 10}
	number, err := graph.WeightedCliqueNumber(weights)
	if err != nil || number != 20 {
		t.Fatalf("WeightedCliqueNumber = %d, %v", number, err)
	}
	maximum, err := graph.MaximumWeightCliques(weights, 1)
	if err != nil || maximum.Truncated || !reflect.DeepEqual(maximum.Sets, [][]int{{3, 4}}) {
		t.Fatalf("MaximumWeightCliques = %#v, %v", maximum, err)
	}
	minimum, upper := 3, 3
	result, err := graph.WeightedCliques(weights, WeightedCliqueOptions{
		Range: WeightRange{Minimum: &minimum, Maximum: &upper}, MaxResults: 1,
	})
	if err != nil || result.Truncated || !reflect.DeepEqual(result.Sets, [][]int{{0, 1, 2}}) {
		t.Errorf("weighted range = %#v, %v", result, err)
	}
}

func TestWeightedCliquesMaximalOnlyAndTruncation(t *testing.T) {
	graph := testCliqueGraph(t)
	defer graph.Close()
	weights := []int{1, 1, 1, 1, 1}
	all, err := graph.WeightedCliques(weights, WeightedCliqueOptions{MaxResults: 10})
	if err != nil || all.Truncated || len(all.Sets) != 10 {
		t.Fatalf("all weighted cliques = %#v, %v", all, err)
	}
	maximal, err := graph.WeightedCliques(weights, WeightedCliqueOptions{MaxResults: 2, MaximalOnly: true})
	if err != nil || maximal.Truncated || !reflect.DeepEqual(sortedVertexSets(maximal.Sets), [][]int{{3, 4}, {0, 1, 2}}) {
		t.Errorf("maximal weighted cliques = %#v, %v", maximal, err)
	}
	limited, err := graph.WeightedCliques(weights, WeightedCliqueOptions{MaxResults: 1, MaximalOnly: true})
	if err != nil || !limited.Truncated || len(limited.Sets) != 1 {
		t.Errorf("limited weighted cliques = %#v, %v", limited, err)
	}
}

func TestMaximumWeightCliqueTies(t *testing.T) {
	graph := testCliqueGraph(t)
	defer graph.Close()
	weights := []int{1, 1, 1, 1, 2}
	limited, err := graph.MaximumWeightCliques(weights, 1)
	if err != nil || !limited.Truncated || len(limited.Sets) != 1 {
		t.Fatalf("limited maximum-weight ties = %#v, %v", limited, err)
	}
	all, err := graph.MaximumWeightCliques(weights, 2)
	if err != nil || all.Truncated || !reflect.DeepEqual(sortedVertexSets(all.Sets), [][]int{{3, 4}, {0, 1, 2}}) {
		t.Errorf("maximum-weight ties = %#v, %v", all, err)
	}
}

func TestWeightedCliqueValidationAndGraphShapes(t *testing.T) {
	var nilGraph *Graph
	if _, err := nilGraph.WeightedCliques(nil, WeightedCliqueOptions{MaxResults: 1}); !errors.Is(err, ErrClosed) {
		t.Errorf("nil WeightedCliques error = %v", err)
	}
	if _, err := nilGraph.WeightedCliqueNumber(nil); !errors.Is(err, ErrClosed) {
		t.Errorf("nil WeightedCliqueNumber error = %v", err)
	}
	if _, err := nilGraph.MaximumWeightCliques(nil, 1); !errors.Is(err, ErrClosed) {
		t.Errorf("nil MaximumWeightCliques error = %v", err)
	}

	graph, err := NewGraphFromEdges(3, []Edge{{0, 0}, {0, 1}, {0, 1}, {1, 2}, {2, 0}}, true)
	if err != nil {
		t.Fatal(err)
	}
	weights := []int{1, 2, 3}
	maximum, err := graph.MaximumWeightCliques(weights, 2)
	if err != nil || maximum.Truncated || !reflect.DeepEqual(maximum.Sets, [][]int{{0, 1, 2}}) {
		t.Errorf("directed multigraph maximum = %#v, %v", maximum, err)
	}
	invalidWeights := [][]int{{1, 2}, {1, 0, 2}, {1, -1, 2}}
	if strconv.IntSize == 64 {
		invalidWeights = append(invalidWeights,
			[]int{(1 << 53) + 1, 1, 1},
			[]int{1 << 53, 1, 1},
		)
	}
	for _, invalid := range invalidWeights {
		if _, err := graph.WeightedCliqueNumber(invalid); err == nil {
			t.Errorf("weights %v unexpectedly valid", invalid)
		}
	}
	minimum, maximumBound := 3, 2
	if _, err := graph.WeightedCliques(weights, WeightedCliqueOptions{
		Range: WeightRange{Minimum: &minimum, Maximum: &maximumBound}, MaxResults: 1,
	}); err == nil {
		t.Error("expected invalid weight range")
	}
	if _, err := graph.WeightedCliques(weights, WeightedCliqueOptions{}); err == nil {
		t.Error("expected invalid result limit")
	}
	zero := 0
	if _, err := graph.WeightedCliques(weights, WeightedCliqueOptions{
		Range: WeightRange{Minimum: &zero}, MaxResults: 1,
	}); err == nil {
		t.Error("expected zero weight bound error")
	}
	if err := graph.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := graph.WeightedCliques(weights, WeightedCliqueOptions{MaxResults: 1}); !errors.Is(err, ErrClosed) {
		t.Errorf("closed WeightedCliques error = %v", err)
	}
	if _, err := graph.WeightedCliqueNumber(weights); !errors.Is(err, ErrClosed) {
		t.Errorf("closed WeightedCliqueNumber error = %v", err)
	}
	if _, err := graph.MaximumWeightCliques(weights, 1); !errors.Is(err, ErrClosed) {
		t.Errorf("closed MaximumWeightCliques error = %v", err)
	}
}

func TestWeightedCliqueEmptyGraphAndConversion(t *testing.T) {
	graph, err := NewGraphFromEdges(0, nil, false)
	if err != nil {
		t.Fatal(err)
	}
	defer graph.Close()
	number, err := graph.WeightedCliqueNumber([]int{})
	if err != nil || number != 0 {
		t.Errorf("empty weighted number = %d, %v", number, err)
	}
	enumerated, err := graph.WeightedCliques([]int{}, WeightedCliqueOptions{MaxResults: 1})
	if err != nil || enumerated.Sets == nil || len(enumerated.Sets) != 0 || enumerated.Truncated {
		t.Errorf("empty weighted cliques = %#v, %v", enumerated, err)
	}
	result, err := graph.MaximumWeightCliques([]int{}, 1)
	if err != nil || result.Sets == nil || len(result.Sets) != 0 || result.Truncated {
		t.Errorf("empty maximum weighted cliques = %#v, %v", result, err)
	}
	for _, value := range []float64{-1, 0.5, math.NaN(), math.Inf(1), math.MaxFloat64} {
		if _, err := checkedCliqueWeight(value, "test weight"); err == nil {
			t.Errorf("value %g unexpectedly valid", value)
		}
	}
}
