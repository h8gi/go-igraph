package igraph

import (
	"errors"
	"math"
	"reflect"
	"testing"
)

func TestLocalScanRadiusAndWeights(t *testing.T) {
	g, err := NewGraphFromEdges(3, []Edge{{0, 1}, {1, 2}}, false)
	if err != nil {
		t.Fatal(err)
	}
	defer g.Close()
	tests := []struct {
		radius  int
		weights []float64
		want    []float64
	}{
		{0, nil, []float64{1, 2, 1}},
		{0, []float64{2, 5}, []float64{2, 7, 5}},
		{1, nil, []float64{1, 2, 1}},
		{1, []float64{2, 5}, []float64{2, 7, 5}},
		{2, nil, []float64{2, 2, 2}},
	}
	for _, test := range tests {
		got, err := g.LocalScan(LocalScanOptions{Radius: test.radius, Direction: DirectionAll, Weights: test.weights})
		if err != nil {
			t.Fatalf("radius %d: %v", test.radius, err)
		}
		if !reflect.DeepEqual(got, test.want) {
			t.Fatalf("radius %d = %v, want %v", test.radius, got, test.want)
		}
	}
}

func TestLocalScanDirectedModes(t *testing.T) {
	g, err := NewGraphFromEdges(3, []Edge{{0, 1}, {0, 2}, {2, 1}}, true)
	if err != nil {
		t.Fatal(err)
	}
	defer g.Close()
	out, err := g.LocalScan(LocalScanOptions{Direction: DirectionOut})
	if err != nil {
		t.Fatal(err)
	}
	in, err := g.LocalScan(LocalScanOptions{Direction: DirectionIn})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(out, []float64{2, 0, 1}) {
		t.Fatalf("out = %v", out)
	}
	if !reflect.DeepEqual(in, []float64{0, 2, 1}) {
		t.Fatalf("in = %v", in)
	}
}

func TestLocalScanSubsetsOrderLoopsAndParallelEdges(t *testing.T) {
	g, err := NewGraphFromEdges(2, []Edge{{0, 0}, {0, 1}, {0, 1}}, false)
	if err != nil {
		t.Fatal(err)
	}
	defer g.Close()
	subsets := [][]int{{}, {0}, {0, 1}, {0}, {1}}
	got, err := g.LocalScanSubsets(subsets, SubsetLocalScanOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, []float64{0, 1, 3, 1, 0}) {
		t.Fatalf("subsets = %v", got)
	}
	weighted, err := g.LocalScanSubsets([][]int{{0}, {0, 1}}, SubsetLocalScanOptions{Weights: []float64{4, 2, 3}})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(weighted, []float64{4, 9}) {
		t.Fatalf("weighted = %v", weighted)
	}
	got[0] = 99
	again, err := g.LocalScanSubsets([][]int{{}}, SubsetLocalScanOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(again, []float64{0}) {
		t.Fatalf("result aliases call storage: %v", again)
	}
}

func TestLocalScanValidationAndFailures(t *testing.T) {
	g, err := NewGraphFromEdges(2, []Edge{{0, 1}}, true)
	if err != nil {
		t.Fatal(err)
	}
	failure := errors.New("failure")
	if _, err := g.LocalScan(LocalScanOptions{Radius: -1}); err == nil {
		t.Fatal("expected radius error")
	}
	if _, err := g.LocalScan(LocalScanOptions{Direction: DirectionMode(99)}); err == nil {
		t.Fatal("expected direction error")
	}
	if _, err := g.LocalScan(LocalScanOptions{Weights: []float64{math.NaN()}}); err == nil {
		t.Fatal("expected weight error")
	}
	if _, err := g.localScan(LocalScanOptions{}, localScanHooks{newResult: func([]float64) (*realVector, error) { return nil, failure }}); !errors.Is(err, failure) {
		t.Fatalf("result initialization error = %v", err)
	}
	if _, err := g.localScan(LocalScanOptions{}, localScanHooks{run: func() error { return failure }}); !errors.Is(err, failure) {
		t.Fatalf("operation error = %v", err)
	}
	if _, err := g.LocalScanSubsets([][]int{{-1}}, SubsetLocalScanOptions{}); err == nil {
		t.Fatal("expected invalid vertex error")
	}
	if _, err := g.LocalScanSubsets([][]int{{0, 0}}, SubsetLocalScanOptions{}); err == nil {
		t.Fatal("expected duplicate vertex error")
	}
	if _, err := g.localScanSubsets(nil, SubsetLocalScanOptions{}, localScanHooks{newList: func() (*intVectorList, error) { return nil, failure }}); !errors.Is(err, failure) {
		t.Fatalf("list initialization error = %v", err)
	}
	if _, err := g.localScanSubsets([][]int{{0}}, SubsetLocalScanOptions{}, localScanHooks{newVector: func([]int) (*intVector, error) { return nil, failure }}); !errors.Is(err, failure) {
		t.Fatalf("vector initialization error = %v", err)
	}
	if _, err := g.localScanSubsets([][]int{{0}}, SubsetLocalScanOptions{}, localScanHooks{append: func(*intVectorList, *intVector) error { return failure }}); !errors.Is(err, failure) {
		t.Fatalf("append error = %v", err)
	}
	if _, err := g.localScanSubsets(nil, SubsetLocalScanOptions{}, localScanHooks{newResult: func([]float64) (*realVector, error) { return nil, failure }}); !errors.Is(err, failure) {
		t.Fatalf("subset result initialization error = %v", err)
	}
	if _, err := g.localScanSubsets(nil, SubsetLocalScanOptions{}, localScanHooks{run: func() error { return failure }}); !errors.Is(err, failure) {
		t.Fatalf("subset operation error = %v", err)
	}
	if err := g.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := g.LocalScan(LocalScanOptions{}); !errors.Is(err, ErrClosed) {
		t.Fatalf("closed scan error = %v", err)
	}
	if _, err := g.LocalScanSubsets(nil, SubsetLocalScanOptions{}); !errors.Is(err, ErrClosed) {
		t.Fatalf("closed subset error = %v", err)
	}
	var nilGraph *Graph
	if _, err := nilGraph.LocalScan(LocalScanOptions{}); !errors.Is(err, ErrClosed) {
		t.Fatalf("nil scan error = %v", err)
	}
	if _, err := nilGraph.LocalScanSubsets(nil, SubsetLocalScanOptions{}); !errors.Is(err, ErrClosed) {
		t.Fatalf("nil subset error = %v", err)
	}
}

func TestLocalScanEmptyGraph(t *testing.T) {
	g, err := NewGraphFromEdges(0, nil, false)
	if err != nil {
		t.Fatal(err)
	}
	defer g.Close()
	got, err := g.LocalScan(LocalScanOptions{Radius: 2})
	if err != nil {
		t.Fatal(err)
	}
	if got == nil || len(got) != 0 {
		t.Fatalf("empty scan = %#v", got)
	}
	subsets, err := g.LocalScanSubsets([][]int{{}, {}}, SubsetLocalScanOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(subsets, []float64{0, 0}) {
		t.Fatalf("empty subsets = %#v", subsets)
	}
}
