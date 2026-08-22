package igraph_test

import (
	"errors"
	"math"
	"reflect"
	"testing"

	igraph "github.com/h8gi/go-igraph"
)

func TestAdjacencyMatrixDirectedWeightedParallelAndOwnership(t *testing.T) {
	graph, err := igraph.NewGraphFromEdges(3, []igraph.Edge{
		{From: 0, To: 1}, {From: 0, To: 1}, {From: 1, To: 2}, {From: 2, To: 2},
	}, true)
	if err != nil {
		t.Fatal(err)
	}
	weights := []float64{2, -0.5, 3, 4}
	matrix, err := graph.AdjacencyMatrix(weights, igraph.AdjacencyMatrixOptions{Loops: igraph.AdjacencyLoopsOnce})
	if err != nil {
		t.Fatal(err)
	}
	if want := [][]float64{{0, 1.5, 0}, {0, 0, 3}, {0, 0, 4}}; !reflect.DeepEqual(matrix.Rows(), want) {
		t.Fatalf("matrix = %v, want %v", matrix.Rows(), want)
	}
	if err := graph.Close(); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(matrix.Rows(), [][]float64{{0, 1.5, 0}, {0, 0, 3}, {0, 0, 4}}) {
		t.Fatalf("matrix after close = %v", matrix.Rows())
	}
}

func TestAdjacencyMatrixUndirectedTrianglesAndLoops(t *testing.T) {
	graph, err := igraph.NewGraphFromEdges(3, []igraph.Edge{
		{From: 0, To: 1}, {From: 0, To: 1}, {From: 1, To: 2}, {From: 2, To: 2},
	}, false)
	if err != nil {
		t.Fatal(err)
	}
	defer graph.Close()

	for _, test := range []struct {
		mode igraph.AdjacencyMatrixMode
		want [][]float64
	}{
		{igraph.AdjacencyMatrixBoth, [][]float64{{0, 2, 0}, {2, 0, 1}, {0, 1, 1}}},
		{igraph.AdjacencyMatrixUpper, [][]float64{{0, 2, 0}, {0, 0, 1}, {0, 0, 1}}},
		{igraph.AdjacencyMatrixLower, [][]float64{{0, 0, 0}, {2, 0, 0}, {0, 1, 1}}},
	} {
		matrix, err := graph.AdjacencyMatrix(nil, igraph.AdjacencyMatrixOptions{Mode: test.mode, Loops: igraph.AdjacencyLoopsOnce})
		if err != nil {
			t.Fatalf("mode %d: %v", test.mode, err)
		}
		if !reflect.DeepEqual(matrix.Rows(), test.want) {
			t.Errorf("mode %d matrix = %v, want %v", test.mode, matrix.Rows(), test.want)
		}
	}
	twice, err := graph.AdjacencyMatrix(nil, igraph.AdjacencyMatrixOptions{Loops: igraph.AdjacencyLoopsTwice})
	if err != nil {
		t.Fatal(err)
	}
	if value, _ := twice.At(2, 2); value != 2 {
		t.Fatalf("twice loop diagonal = %v", value)
	}
	without, err := graph.AdjacencyMatrix(nil, igraph.AdjacencyMatrixOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if value, _ := without.At(2, 2); value != 0 {
		t.Fatalf("no-loop diagonal = %v", value)
	}
}

func TestStochasticMatrixRowColumnWeightedAndIsolated(t *testing.T) {
	graph, err := igraph.NewGraphFromEdges(4, []igraph.Edge{
		{From: 0, To: 1}, {From: 0, To: 2}, {From: 1, To: 2},
	}, true)
	if err != nil {
		t.Fatal(err)
	}
	defer graph.Close()
	weights := []float64{2, 1, 3}

	rowWise, err := graph.StochasticMatrix(weights, igraph.StochasticMatrixOptions{})
	if err != nil {
		t.Fatal(err)
	}
	assertApproxMatrix(t, rowWise, [][]float64{{0, 2.0 / 3.0, 1.0 / 3.0, 0}, {0, 0, 1, 0}, {0, 0, 0, 0}, {0, 0, 0, 0}})

	columnWise, err := graph.StochasticMatrix(weights, igraph.StochasticMatrixOptions{ColumnWise: true})
	if err != nil {
		t.Fatal(err)
	}
	assertApproxMatrix(t, columnWise, [][]float64{{0, 1, 0.25, 0}, {0, 0, 0.75, 0}, {0, 0, 0, 0}, {0, 0, 0, 0}})

	zeroWeight, err := graph.StochasticMatrix([]float64{0, 0, 3}, igraph.StochasticMatrixOptions{})
	if err != nil {
		t.Fatal(err)
	}
	assertApproxMatrix(t, zeroWeight, [][]float64{{0, 0, 0, 0}, {0, 0, 1, 0}, {0, 0, 0, 0}, {0, 0, 0, 0}})
}

func TestAdjacencyConversionEmptyValidationAndClosed(t *testing.T) {
	empty, _ := igraph.NewGraph()
	adjacency, err := empty.AdjacencyMatrix(nil, igraph.AdjacencyMatrixOptions{})
	if err != nil {
		t.Fatal(err)
	}
	rows, columns := adjacency.Dims()
	if rows != 0 || columns != 0 {
		t.Fatalf("adjacency dims = %dx%d", rows, columns)
	}
	stochastic, err := empty.StochasticMatrix(nil, igraph.StochasticMatrixOptions{})
	if err != nil {
		t.Fatal(err)
	}
	rows, columns = stochastic.Dims()
	if rows != 0 || columns != 0 {
		t.Fatalf("stochastic dims = %dx%d", rows, columns)
	}
	_ = empty.Close()
	if _, err := empty.AdjacencyMatrix(nil, igraph.AdjacencyMatrixOptions{}); !errors.Is(err, igraph.ErrClosed) {
		t.Fatalf("adjacency after close = %v", err)
	}
	if _, err := empty.StochasticMatrix(nil, igraph.StochasticMatrixOptions{}); !errors.Is(err, igraph.ErrClosed) {
		t.Fatalf("stochastic after close = %v", err)
	}
	var nilGraph *igraph.Graph
	if _, err := nilGraph.AdjacencyMatrix(nil, igraph.AdjacencyMatrixOptions{}); !errors.Is(err, igraph.ErrClosed) {
		t.Fatalf("nil adjacency = %v", err)
	}

	graph, _ := igraph.NewGraphFromEdges(2, []igraph.Edge{{From: 0, To: 1}}, true)
	defer graph.Close()
	if _, err := graph.AdjacencyMatrix([]float64{}, igraph.AdjacencyMatrixOptions{}); err == nil {
		t.Fatal("short adjacency weights error nil")
	}
	if _, err := graph.AdjacencyMatrix([]float64{math.NaN()}, igraph.AdjacencyMatrixOptions{}); err == nil {
		t.Fatal("NaN adjacency weight error nil")
	}
	if _, err := graph.StochasticMatrix([]float64{-1}, igraph.StochasticMatrixOptions{}); err == nil {
		t.Fatal("negative stochastic weight error nil")
	}
	if _, err := graph.StochasticMatrix([]float64{math.Inf(1)}, igraph.StochasticMatrixOptions{}); err == nil {
		t.Fatal("infinite stochastic weight error nil")
	}
	if _, err := graph.AdjacencyMatrix(nil, igraph.AdjacencyMatrixOptions{Mode: igraph.AdjacencyMatrixMode(99)}); err == nil {
		t.Fatal("invalid adjacency mode error nil")
	}
	if _, err := graph.AdjacencyMatrix(nil, igraph.AdjacencyMatrixOptions{Loops: igraph.AdjacencyLoops(99)}); err == nil {
		t.Fatal("invalid loop mode error nil")
	}
}

func assertApproxMatrix(t *testing.T, got igraph.Matrix, want [][]float64) {
	t.Helper()
	rows := got.Rows()
	if len(rows) != len(want) {
		t.Fatalf("row count = %d, want %d", len(rows), len(want))
	}
	for row := range want {
		for column := range want[row] {
			if math.Abs(rows[row][column]-want[row][column]) > 1e-12 {
				t.Fatalf("matrix[%d][%d] = %v, want %v", row, column, rows[row][column], want[row][column])
			}
		}
	}
}
