package igraph_test

import (
	"errors"
	"math"
	"reflect"
	"testing"

	igraph "github.com/h8gi/go-igraph"
)

func TestGraphletsCandidateBasisWeightedAndUnweighted(t *testing.T) {
	graph := newGraphletGraph(t, 5, completeEdges(5), false)
	unweighted, err := graph.GraphletsCandidateBasis(nil)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(unweighted.Cliques, [][]int{{0, 1, 2, 3, 4}}) ||
		!reflect.DeepEqual(unweighted.Thresholds, []float64{1}) {
		t.Fatalf("unweighted candidate basis = %#v", unweighted)
	}
	weights := make([]float64, 10)
	for index := range weights {
		weights[index] = 1
	}
	weights[0] = 2
	weighted, err := graph.GraphletsCandidateBasis(weights)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(weighted.Cliques, [][]int{{0, 1, 2, 3, 4}, {0, 1}}) ||
		!reflect.DeepEqual(weighted.Thresholds, []float64{1, 2}) {
		t.Fatalf("weighted candidate basis = %#v", weighted)
	}
	if err := graph.Close(); err != nil {
		t.Fatal(err)
	}
	weighted.Cliques[0][0] = 99
	weighted.Thresholds[0] = 99

	directed := newGraphletGraph(t, 3, completeEdges(3), true)
	directedBasis, err := directed.GraphletsCandidateBasis(nil)
	if err != nil || !reflect.DeepEqual(directedBasis.Cliques, [][]int{{0, 1, 2}}) {
		t.Fatalf("directed candidate basis = %#v, %v", directedBasis, err)
	}
}

func TestGraphletsKnownProjection(t *testing.T) {
	edges := []igraph.Edge{
		{0, 1}, {0, 2}, {1, 2}, {1, 3},
		{1, 4}, {2, 3}, {2, 4}, {3, 4},
	}
	weights := []float64{2, 2, 3, 1, 1, 4, 4, 4}
	graph := newGraphletGraph(t, 5, edges, false)
	result, err := graph.Graphlets(weights, 1000)
	if err != nil {
		t.Fatal(err)
	}
	wantCliques := [][]int{{2, 3, 4}, {0, 1, 2}, {1, 2, 3, 4}, {1, 2}}
	if !reflect.DeepEqual(result.Cliques, wantCliques) {
		t.Fatalf("Graphlets cliques = %v, want %v", result.Cliques, wantCliques)
	}
	wantMu := []float64{1.13842, 0.92554, 0.86148, 0}
	for index := range wantMu {
		if math.Abs(result.Mu[index]-wantMu[index]) > 0.00001 {
			t.Errorf("Graphlets Mu[%d] = %.8f, want %.5f", index, result.Mu[index], wantMu[index])
		}
	}
	if err := graph.Close(); err != nil {
		t.Fatal(err)
	}
	result.Cliques[0][0] = 99
	result.Mu[0] = 99
}

func TestGraphletsProjectInitialMu(t *testing.T) {
	graph := newGraphletGraph(t, 3, completeEdges(3), false)
	cliques := [][]int{{2, 0, 1}, {0, 1}}
	started, err := graph.GraphletsProject(cliques, []float64{2, 3}, nil, 0)
	if err != nil || !reflect.DeepEqual(started, []float64{2, 3}) {
		t.Fatalf("GraphletsProject initial Mu = %v, %v", started, err)
	}
	defaulted, err := graph.GraphletsProject(cliques, []float64{}, nil, 0)
	if err != nil || !reflect.DeepEqual(defaulted, []float64{1, 1}) {
		t.Fatalf("GraphletsProject default Mu = %v, %v", defaulted, err)
	}
	projected, err := graph.GraphletsProject(cliques, []float64{2, 3}, []float64{1, 2, 3}, 2)
	if err != nil || len(projected) != 2 {
		t.Fatalf("GraphletsProject = %v, %v", projected, err)
	}
	for index, value := range projected {
		if math.IsNaN(value) || math.IsInf(value, 0) || value < 0 {
			t.Errorf("GraphletsProject[%d] = %v", index, value)
		}
	}
	if err := graph.Close(); err != nil {
		t.Fatal(err)
	}
	projected[0] = 99
}

func TestGraphletsEmptyGraph(t *testing.T) {
	graph := newGraphletGraph(t, 3, nil, false)
	result, err := graph.Graphlets(nil, 1)
	if err != nil || result.Cliques == nil || result.Mu == nil || len(result.Cliques) != 0 || len(result.Mu) != 0 {
		t.Fatalf("empty Graphlets = %#v, %v", result, err)
	}
	basis, err := graph.GraphletsCandidateBasis([]float64{})
	if err != nil || basis.Cliques == nil || basis.Thresholds == nil || len(basis.Cliques) != 0 || len(basis.Thresholds) != 0 {
		t.Fatalf("empty GraphletsCandidateBasis = %#v, %v", basis, err)
	}
	mu, err := graph.GraphletsProject(nil, nil, nil, 1)
	if err != nil || mu == nil || len(mu) != 0 {
		t.Fatalf("empty GraphletsProject = %#v, %v", mu, err)
	}
}

func TestGraphletsValidateInputs(t *testing.T) {
	graph := newGraphletGraph(t, 3, completeEdges(3), false)
	invalidWeights := [][]float64{
		{}, {1}, {1, 1, -1}, {1, 1, math.NaN()}, {1, 1, math.Inf(1)},
	}
	for index, weights := range invalidWeights {
		if _, err := graph.Graphlets(weights, 1); err == nil {
			t.Errorf("Graphlets accepted invalid weights %d", index)
		}
		if _, err := graph.GraphletsCandidateBasis(weights); err == nil {
			t.Errorf("GraphletsCandidateBasis accepted invalid weights %d", index)
		}
	}
	if _, err := graph.Graphlets(nil, -1); err == nil {
		t.Error("Graphlets accepted negative iterations")
	}
	if _, err := graph.GraphletsProject([][]int{{0, 1}}, nil, nil, -1); err == nil {
		t.Error("GraphletsProject accepted negative iterations")
	}
	invalidCliques := [][][]int{
		{{}}, {{0}}, {{-1, 1}}, {{0, 3}}, {{0, 0}}, {{0, 1}, {1, 0}},
	}
	for index, cliques := range invalidCliques {
		if _, err := graph.GraphletsProject(cliques, nil, nil, 1); err == nil {
			t.Errorf("GraphletsProject accepted invalid cliques %d: %v", index, cliques)
		}
	}
	path := newGraphletGraph(t, 3, []igraph.Edge{{0, 1}, {1, 2}}, false)
	if _, err := path.GraphletsProject([][]int{{0, 2}}, nil, nil, 1); err == nil {
		t.Error("GraphletsProject accepted a non-clique")
	}
	if _, err := graph.GraphletsProject([][]int{{0, 1}}, []float64{1, 2}, nil, 1); err == nil {
		t.Error("GraphletsProject accepted mismatched initial Mu")
	}
	for _, initial := range [][]float64{{-1}, {math.NaN()}, {math.Inf(1)}} {
		if _, err := graph.GraphletsProject([][]int{{0, 1}}, initial, nil, 1); err == nil {
			t.Errorf("GraphletsProject accepted initial Mu %v", initial)
		}
	}
}

func TestGraphletsRejectNonSimpleGraphs(t *testing.T) {
	tests := []struct {
		name     string
		edges    []igraph.Edge
		directed bool
	}{
		{name: "loop", edges: []igraph.Edge{{0, 0}}},
		{name: "parallel", edges: []igraph.Edge{{0, 1}, {0, 1}}},
		{name: "antiparallel", edges: []igraph.Edge{{0, 1}, {1, 0}}, directed: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			graph := newGraphletGraph(t, 2, test.edges, test.directed)
			if _, err := graph.Graphlets(nil, 1); err == nil {
				t.Error("Graphlets accepted non-simple graph")
			}
			if _, err := graph.GraphletsCandidateBasis(nil); err == nil {
				t.Error("GraphletsCandidateBasis accepted non-simple graph")
			}
			if _, err := graph.GraphletsProject(nil, nil, nil, 1); err == nil {
				t.Error("GraphletsProject accepted non-simple graph with an empty basis")
			}
		})
	}
}

func TestGraphletsClosedAndNil(t *testing.T) {
	var nilGraph *igraph.Graph
	if _, err := nilGraph.Graphlets(nil, 1); !errors.Is(err, igraph.ErrClosed) {
		t.Errorf("nil Graphlets error = %v", err)
	}
	if _, err := nilGraph.GraphletsCandidateBasis(nil); !errors.Is(err, igraph.ErrClosed) {
		t.Errorf("nil GraphletsCandidateBasis error = %v", err)
	}
	if _, err := nilGraph.GraphletsProject(nil, nil, nil, 1); !errors.Is(err, igraph.ErrClosed) {
		t.Errorf("nil GraphletsProject error = %v", err)
	}
	graph := newGraphletGraph(t, 2, []igraph.Edge{{0, 1}}, false)
	if err := graph.Close(); err != nil {
		t.Fatal(err)
	}
	calls := []func() error{
		func() error { _, err := graph.Graphlets(nil, 1); return err },
		func() error { _, err := graph.GraphletsCandidateBasis(nil); return err },
		func() error { _, err := graph.GraphletsProject(nil, nil, nil, 1); return err },
	}
	for index, call := range calls {
		if err := call(); !errors.Is(err, igraph.ErrClosed) {
			t.Errorf("closed graphlet call %d error = %v", index, err)
		}
	}
}

func newGraphletGraph(t *testing.T, vertices int, edges []igraph.Edge, directed bool) *igraph.Graph {
	t.Helper()
	graph, err := igraph.NewGraphFromEdges(vertices, edges, directed)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = graph.Close() })
	return graph
}

func completeEdges(vertices int) []igraph.Edge {
	edges := make([]igraph.Edge, 0, vertices*(vertices-1)/2)
	for from := 0; from < vertices; from++ {
		for to := from + 1; to < vertices; to++ {
			edges = append(edges, igraph.Edge{From: from, To: to})
		}
	}
	return edges
}
