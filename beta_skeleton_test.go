package igraph_test

import (
	"errors"
	"math"
	"testing"

	igraph "github.com/h8gi/go-igraph"
)

func TestBetaOneSkeletonsMatchGabrielGraph(t *testing.T) {
	points, _ := igraph.NewMatrixFromRows([][]float64{
		{0, 0}, {3, 0}, {4, 2}, {2, 4}, {0, 3}, {1.4, 1.1},
	})
	gabriel, err := igraph.NewGabrielGraph(points)
	if err != nil {
		t.Fatal(err)
	}
	defer gabriel.Close()
	for name, construct := range map[string]func() (*igraph.Graph, error){
		"lune":   func() (*igraph.Graph, error) { return igraph.NewLuneBetaSkeleton(points, 1) },
		"circle": func() (*igraph.Graph, error) { return igraph.NewCircleBetaSkeleton(points, 1) },
	} {
		t.Run(name, func(t *testing.T) {
			graph, err := construct()
			if err != nil {
				t.Fatal(err)
			}
			defer graph.Close()
			assertSameEdges(t, graph, gabriel)
			if directed, err := graph.IsDirected(); err != nil || directed {
				t.Fatalf("IsDirected = %t, %v, want false, nil", directed, err)
			}
		})
	}
}

func TestBetaSkeletonDimensionContracts(t *testing.T) {
	points2D, _ := igraph.NewMatrixFromRows([][]float64{{0, 0}, {2, 0}, {0, 2}})
	lune, err := igraph.NewLuneBetaSkeleton(points2D, 0.5)
	if err != nil {
		t.Fatal(err)
	}
	defer lune.Close()
	circle, err := igraph.NewCircleBetaSkeleton(points2D, 0.5)
	if err != nil {
		t.Fatal(err)
	}
	defer circle.Close()
	assertSameEdges(t, lune, circle)

	points3D, _ := igraph.NewMatrixFromRows([][]float64{
		{0, 0, 0}, {2, 0, 0}, {0, 2, 0}, {0, 0, 2}, {0.5, 0.4, 0.3},
	})
	threeDimensional, err := igraph.NewLuneBetaSkeleton(points3D, 2)
	if err != nil {
		t.Fatal(err)
	}
	defer threeDimensional.Close()
	if vertices, err := threeDimensional.VertexCount(); err != nil || vertices != 5 {
		t.Fatalf("VertexCount = %d, %v, want 5, nil", vertices, err)
	}
	if graph, err := igraph.NewLuneBetaSkeleton(points3D, 0.5); err == nil {
		_ = graph.Close()
		t.Fatal("expected 2D validation error for lune beta below 1")
	}
	if graph, err := igraph.NewCircleBetaSkeleton(points3D, 2); err == nil {
		_ = graph.Close()
		t.Fatal("expected 2D validation error for circle skeleton")
	}
}

func TestBetaWeightedGabrielAlignmentAndCutoff(t *testing.T) {
	points, _ := igraph.NewMatrixFromRows([][]float64{
		{0, 0}, {3, 0}, {4, 2}, {2, 4}, {0, 3}, {1.4, 1.1},
	})
	unlimited, err := igraph.NewBetaWeightedGabrielGraph(points, igraph.BetaWeightedGabrielOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer unlimited.Graph.Close()
	gabriel, err := igraph.NewGabrielGraph(points)
	if err != nil {
		t.Fatal(err)
	}
	defer gabriel.Close()
	assertSameEdges(t, unlimited.Graph, gabriel)
	if edges, _ := unlimited.Graph.EdgeCount(); len(unlimited.ThresholdBetas) != edges {
		t.Fatalf("threshold count = %d, want %d", len(unlimited.ThresholdBetas), edges)
	}
	for edgeID, beta := range unlimited.ThresholdBetas {
		if beta < 1 || math.IsNaN(beta) {
			t.Fatalf("threshold %d = %v, want >= 1 or +Inf", edgeID, beta)
		}
	}

	maximum := 1.25
	bounded, err := igraph.NewBetaWeightedGabrielGraph(points, igraph.BetaWeightedGabrielOptions{MaxBeta: &maximum})
	if err != nil {
		t.Fatal(err)
	}
	defer bounded.Graph.Close()
	assertSameEdges(t, bounded.Graph, gabriel)
	for edgeID, beta := range bounded.ThresholdBetas {
		if !math.IsInf(beta, 1) && (beta < 1 || beta >= maximum) {
			t.Fatalf("bounded threshold %d = %v, want [1, %v) or +Inf", edgeID, beta, maximum)
		}
	}

	thresholds := append([]float64{}, bounded.ThresholdBetas...)
	if err := bounded.Graph.Close(); err != nil {
		t.Fatal(err)
	}
	for index := range thresholds {
		if thresholds[index] != bounded.ThresholdBetas[index] {
			t.Fatal("threshold slice changed after graph closure")
		}
	}
}

func TestBetaSpatialEmptySingleAndOwnership(t *testing.T) {
	for _, construct := range []func() (*igraph.Graph, error){
		func() (*igraph.Graph, error) { return igraph.NewLuneBetaSkeleton(igraph.Matrix{}, 0.5) },
		func() (*igraph.Graph, error) { return igraph.NewLuneBetaSkeleton(igraph.Matrix{}, 2) },
		func() (*igraph.Graph, error) { return igraph.NewCircleBetaSkeleton(igraph.Matrix{}, 2) },
	} {
		graph, err := construct()
		if err != nil {
			t.Fatal(err)
		}
		if vertices, _ := graph.VertexCount(); vertices != 0 {
			t.Fatalf("empty vertex count = %d", vertices)
		}
		if err := graph.Close(); err != nil {
			t.Fatal(err)
		}
		if err := graph.Close(); err != nil {
			t.Fatalf("repeated Close error = %v", err)
		}
		if _, err := graph.VertexCount(); !errors.Is(err, igraph.ErrClosed) {
			t.Fatalf("VertexCount after Close error = %v, want ErrClosed", err)
		}
	}

	empty, err := igraph.NewBetaWeightedGabrielGraph(igraph.Matrix{}, igraph.BetaWeightedGabrielOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer empty.Graph.Close()
	if empty.ThresholdBetas == nil || len(empty.ThresholdBetas) != 0 {
		t.Fatalf("empty thresholds = %#v, want non-nil empty", empty.ThresholdBetas)
	}

	singlePoint, _ := igraph.NewMatrixFromRows([][]float64{{1, 2}})
	single, err := igraph.NewBetaWeightedGabrielGraph(singlePoint, igraph.BetaWeightedGabrielOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer single.Graph.Close()
	if vertices, _ := single.Graph.VertexCount(); vertices != 1 {
		t.Fatalf("single vertex count = %d", vertices)
	}
	if single.ThresholdBetas == nil || len(single.ThresholdBetas) != 0 {
		t.Fatalf("single thresholds = %#v", single.ThresholdBetas)
	}
}

func TestBetaSpatialRejectsInvalidInputs(t *testing.T) {
	points2D, _ := igraph.NewMatrixFromRows([][]float64{{0, 0}, {1, 0}})
	for _, beta := range []float64{0, -1, math.NaN(), math.Inf(1)} {
		if graph, err := igraph.NewLuneBetaSkeleton(points2D, beta); err == nil {
			_ = graph.Close()
			t.Fatalf("lune beta %v: expected validation error", beta)
		}
		if graph, err := igraph.NewCircleBetaSkeleton(points2D, beta); err == nil {
			_ = graph.Close()
			t.Fatalf("circle beta %v: expected validation error", beta)
		}
	}

	duplicate, _ := igraph.NewMatrixFromRows([][]float64{{0, 0}, {0, 0}})
	nonFinite, _ := igraph.NewMatrixFromRows([][]float64{{0, 0}, {math.Inf(1), 0}})
	for _, points := range []igraph.Matrix{duplicate, nonFinite} {
		if graph, err := igraph.NewLuneBetaSkeleton(points, 1); err == nil {
			_ = graph.Close()
			t.Fatal("expected lune point validation error")
		}
		if result, err := igraph.NewBetaWeightedGabrielGraph(points, igraph.BetaWeightedGabrielOptions{}); err == nil {
			_ = result.Graph.Close()
			t.Fatal("expected weighted Gabriel point validation error")
		}
	}

	for _, maximum := range []float64{0, -1, math.NaN(), math.Inf(1)} {
		if result, err := igraph.NewBetaWeightedGabrielGraph(points2D, igraph.BetaWeightedGabrielOptions{MaxBeta: &maximum}); err == nil {
			_ = result.Graph.Close()
			t.Fatalf("maximum beta %v: expected validation error", maximum)
		}
	}

	collinear, _ := igraph.NewMatrixFromRows([][]float64{{0, 0}, {1, 1}, {2, 2}})
	for name, construct := range map[string]func() (*igraph.Graph, error){
		"lune":   func() (*igraph.Graph, error) { return igraph.NewLuneBetaSkeleton(collinear, 2) },
		"circle": func() (*igraph.Graph, error) { return igraph.NewCircleBetaSkeleton(collinear, 2) },
	} {
		t.Run("degenerate "+name, func(t *testing.T) {
			if graph, err := construct(); err == nil {
				_ = graph.Close()
				t.Fatal("expected upstream degeneracy error")
			}
		})
	}
	if result, err := igraph.NewBetaWeightedGabrielGraph(collinear, igraph.BetaWeightedGabrielOptions{}); err == nil {
		_ = result.Graph.Close()
		t.Fatal("expected weighted Gabriel degeneracy error")
	}
}

func assertSameEdges(t *testing.T, left, right *igraph.Graph) {
	t.Helper()
	leftEdges, err := left.Edges()
	if err != nil {
		t.Fatal(err)
	}
	rightEdges, err := right.Edges()
	if err != nil {
		t.Fatal(err)
	}
	if len(leftEdges) != len(rightEdges) {
		t.Fatalf("edge counts differ: %d and %d", len(leftEdges), len(rightEdges))
	}
	for _, edge := range leftEdges {
		adjacent, err := right.AreAdjacent(edge.From, edge.To)
		if err != nil || !adjacent {
			t.Fatalf("edge (%d, %d) missing: %v", edge.From, edge.To, err)
		}
	}
}
