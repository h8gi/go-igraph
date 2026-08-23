package igraph

import (
	"errors"
	"math"
	"testing"
)

func TestLaplacianKnownAnswersAndOwnership(t *testing.T) {
	g, err := NewGraphFromEdges(3, []Edge{{0, 1}, {1, 2}}, false)
	if err != nil {
		t.Fatal(err)
	}
	got, err := g.Laplacian(LaplacianOptions{})
	if err != nil {
		t.Fatal(err)
	}
	assertMatrixRowsClose(t, got, [][]float64{{1, -1, 0}, {-1, 2, -1}, {0, -1, 1}})
	weighted, err := g.Laplacian(LaplacianOptions{Weights: []float64{2, 3}})
	if err != nil {
		t.Fatal(err)
	}
	assertMatrixRowsClose(t, weighted, [][]float64{{2, -2, 0}, {-2, 5, -3}, {0, -3, 3}})
	g.Close()
	assertMatrixRowsClose(t, got, [][]float64{{1, -1, 0}, {-1, 2, -1}, {0, -1, 1}})
}

func TestLaplacianNormalizationKnownAnswers(t *testing.T) {
	g, _ := NewGraphFromEdges(4, []Edge{{0, 1}, {1, 2}}, false)
	defer g.Close()
	s := 1 / math.Sqrt2
	cases := []struct {
		normalization LaplacianNormalization
		want          [][]float64
	}{
		{LaplacianSymmetric, [][]float64{{1, -s, 0, 0}, {-s, 1, -s, 0}, {0, -s, 1, 0}, {0, 0, 0, 0}}},
		{LaplacianLeft, [][]float64{{1, -1, 0, 0}, {-0.5, 1, -0.5, 0}, {0, -1, 1, 0}, {0, 0, 0, 0}}},
		{LaplacianRight, [][]float64{{1, -0.5, 0, 0}, {-1, 1, -1, 0}, {0, -0.5, 1, 0}, {0, 0, 0, 0}}},
	}
	for _, tc := range cases {
		got, err := g.Laplacian(LaplacianOptions{Normalization: tc.normalization})
		if err != nil {
			t.Fatalf("normalization %d: %v", tc.normalization, err)
		}
		assertMatrixRowsClose(t, got, tc.want)
	}
}

func TestDirectedLaplacianModes(t *testing.T) {
	g, _ := NewGraphFromEdges(3, []Edge{{0, 1}, {1, 2}}, true)
	defer g.Close()
	cases := []struct {
		mode DirectionMode
		want [][]float64
	}{
		{DirectionOut, [][]float64{{1, -1, 0}, {0, 1, -1}, {0, 0, 0}}},
		{DirectionIn, [][]float64{{0, -1, 0}, {0, 1, -1}, {0, 0, 1}}},
		{DirectionAll, [][]float64{{1, -1, 0}, {-1, 2, -1}, {0, -1, 1}}},
	}
	for _, tc := range cases {
		got, err := g.Laplacian(LaplacianOptions{Direction: tc.mode})
		if err != nil {
			t.Fatalf("mode %d: %v", tc.mode, err)
		}
		assertMatrixRowsClose(t, got, tc.want)
	}
}

func TestLaplacianLoopsParallelAndEmpty(t *testing.T) {
	g, _ := NewGraphFromEdges(2, []Edge{{0, 0}, {0, 1}, {0, 1}}, false)
	defer g.Close()
	unweighted, err := g.Laplacian(LaplacianOptions{})
	if err != nil {
		t.Fatal(err)
	}
	assertMatrixRowsClose(t, unweighted, [][]float64{{2, -2}, {-2, 2}})
	weighted, err := g.Laplacian(LaplacianOptions{Weights: []float64{5, 2, 3}})
	if err != nil {
		t.Fatal(err)
	}
	assertMatrixRowsClose(t, weighted, [][]float64{{5, -5}, {-5, 5}})
	empty, _ := NewGraph()
	defer empty.Close()
	matrix, err := empty.Laplacian(LaplacianOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if rows, columns := matrix.Dims(); rows != 0 || columns != 0 || matrix.Rows() == nil {
		t.Fatalf("empty dimensions = %d by %d, rows=%#v", rows, columns, matrix.Rows())
	}
}

func TestLaplacianValidationAndClose(t *testing.T) {
	g, _ := NewGraphFromEdges(1, []Edge{{0, 0}}, false)
	for _, options := range []LaplacianOptions{
		{Direction: DirectionMode(99)}, {Normalization: LaplacianNormalization(99)},
		{Weights: []float64{}}, {Weights: []float64{-1}}, {Weights: []float64{math.NaN()}}, {Weights: []float64{math.Inf(1)}},
	} {
		if _, err := g.Laplacian(options); err == nil {
			t.Fatalf("options %#v accepted", options)
		}
	}
	var nilGraph *Graph
	if _, err := nilGraph.Laplacian(LaplacianOptions{}); !errors.Is(err, ErrClosed) {
		t.Fatalf("nil graph = %v", err)
	}
	g.Close()
	if _, err := g.Laplacian(LaplacianOptions{}); !errors.Is(err, ErrClosed) {
		t.Fatalf("closed graph = %v", err)
	}
}

func assertMatrixRowsClose(t *testing.T, got Matrix, want [][]float64) {
	t.Helper()
	rows := got.Rows()
	if len(rows) != len(want) {
		t.Fatalf("row count = %d, want %d: %v", len(rows), len(want), rows)
	}
	for i := range want {
		if len(rows[i]) != len(want[i]) {
			t.Fatalf("column count in row %d = %d, want %d", i, len(rows[i]), len(want[i]))
		}
		for j := range want[i] {
			if math.Abs(rows[i][j]-want[i][j]) > 1e-12 {
				t.Fatalf("matrix[%d][%d] = %.15g, want %.15g; matrix=%v", i, j, rows[i][j], want[i][j], rows)
			}
		}
	}
}
