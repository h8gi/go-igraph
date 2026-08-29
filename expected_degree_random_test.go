package igraph_test

import (
	"errors"
	"math"
	"reflect"
	"strings"
	"testing"

	"github.com/h8gi/go-igraph"
)

func TestChungLuGame(t *testing.T) {
	seed := uint64(335)
	weights := []float64{1, 2, 1, 2, 1, 1}

	for _, variant := range []igraph.ChungLuVariant{
		igraph.ChungLuOriginal,
		igraph.ChungLuMaximumEntropy,
		igraph.ChungLuNorrosReittu,
	} {
		t.Run("undirected variant "+string(rune('0'+variant)), func(t *testing.T) {
			graph, err := igraph.ChungLuGame(weights, nil, igraph.ChungLuOptions{Seed: &seed, Variant: variant})
			if err != nil {
				t.Fatalf("ChungLuGame failed: %v", err)
			}
			defer graph.Close()
			vertices, _ := mustCounts(t, graph)
			if vertices != len(weights) {
				t.Fatalf("vertex count = %d, want %d", vertices, len(weights))
			}
			directed, err := graph.IsDirected()
			if err != nil || directed {
				t.Fatalf("IsDirected = %v, %v, want false", directed, err)
			}
			mustBeSimple(t, graph)
		})
	}

	t.Run("directed loops", func(t *testing.T) {
		graph, err := igraph.ChungLuGame(
			[]float64{1, 2, 1}, []float64{2, 1, 1},
			igraph.ChungLuOptions{Seed: &seed, Loops: true, Variant: igraph.ChungLuMaximumEntropy},
		)
		if err != nil {
			t.Fatalf("ChungLuGame failed: %v", err)
		}
		defer graph.Close()
		directed, err := graph.IsDirected()
		if err != nil || !directed {
			t.Fatalf("IsDirected = %v, %v, want true", directed, err)
		}
		multiplicities, err := graph.EdgeMultiplicities(igraph.AllEdges())
		if err != nil {
			t.Fatalf("EdgeMultiplicities failed: %v", err)
		}
		for edgeID, multiplicity := range multiplicities {
			if multiplicity != 1 {
				t.Errorf("edge %d multiplicity = %d, want 1", edgeID, multiplicity)
			}
		}
	})

	t.Run("zero weight vertex is isolated", func(t *testing.T) {
		graph, err := igraph.ChungLuGame(
			[]float64{0, 2, 2}, nil,
			igraph.ChungLuOptions{Seed: &seed, Loops: true, Variant: igraph.ChungLuMaximumEntropy},
		)
		if err != nil {
			t.Fatalf("ChungLuGame failed: %v", err)
		}
		defer graph.Close()
		selector, err := igraph.VertexIDs(0)
		if err != nil {
			t.Fatalf("VertexIDs failed: %v", err)
		}
		degrees, err := graph.Degree(selector, igraph.DegreeOptions{Direction: igraph.DirectionAll, CountLoops: true})
		if err != nil {
			t.Fatalf("Degree failed: %v", err)
		}
		if len(degrees) != 1 || degrees[0] != 0 {
			t.Fatalf("zero-weight vertex degree = %v, want [0]", degrees)
		}
	})

	t.Run("empty", func(t *testing.T) {
		graph, err := igraph.ChungLuGame(nil, nil, igraph.ChungLuOptions{})
		if err != nil {
			t.Fatalf("ChungLuGame failed: %v", err)
		}
		defer graph.Close()
		vertices, edges := mustCounts(t, graph)
		if vertices != 0 || edges != 0 {
			t.Fatalf("counts = (%d, %d), want (0, 0)", vertices, edges)
		}
	})

	t.Run("same seed", func(t *testing.T) {
		large := make([]float64, 30)
		for index := range large {
			large[index] = 2
		}
		options := igraph.ChungLuOptions{Seed: &seed, Variant: igraph.ChungLuMaximumEntropy}
		first, err := igraph.ChungLuGame(large, nil, options)
		if err != nil {
			t.Fatalf("first ChungLuGame failed: %v", err)
		}
		defer first.Close()
		second, err := igraph.ChungLuGame(large, nil, options)
		if err != nil {
			t.Fatalf("second ChungLuGame failed: %v", err)
		}
		defer second.Close()
		if !reflect.DeepEqual(mustEdges(t, first), mustEdges(t, second)) {
			t.Fatal("same seed produced different edge lists")
		}
	})

	t.Run("validation", func(t *testing.T) {
		tests := []struct {
			name string
			out  []float64
			in   []float64
			opts igraph.ChungLuOptions
		}{
			{name: "invalid variant", out: []float64{1}, opts: igraph.ChungLuOptions{Variant: igraph.ChungLuVariant(255)}},
			{name: "negative", out: []float64{-1}},
			{name: "NaN", out: []float64{math.NaN()}},
			{name: "infinity", out: []float64{math.Inf(1)}},
			{name: "sum overflow", out: []float64{math.MaxFloat64, math.MaxFloat64}},
			{name: "length mismatch", out: []float64{1, 1}, in: []float64{2}},
			{name: "invalid in weight", out: []float64{1}, in: []float64{math.NaN()}},
			{name: "sum mismatch", out: []float64{1, 1}, in: []float64{1, 2}},
			{name: "original probability", out: []float64{10, 10}},
			{name: "original loop probability", out: []float64{5, 0}, opts: igraph.ChungLuOptions{Loops: true}},
		}
		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				if graph, err := igraph.ChungLuGame(test.out, test.in, test.opts); err == nil || graph != nil {
					t.Fatalf("ChungLuGame = %v, %v, want nil error result", graph, err)
				}
			})
		}
	})

	t.Run("excluded large self-pair", func(t *testing.T) {
		graph, err := igraph.ChungLuGame(
			[]float64{100, 0}, []float64{100, 0},
			igraph.ChungLuOptions{Variant: igraph.ChungLuOriginal},
		)
		if err != nil {
			t.Fatalf("ChungLuGame failed: %v", err)
		}
		defer graph.Close()
		_, edges := mustCounts(t, graph)
		if edges != 0 {
			t.Fatalf("edge count = %d, want 0", edges)
		}
	})

}

func TestStaticFitnessGame(t *testing.T) {
	seed := uint64(335)

	t.Run("undirected fixed edge count", func(t *testing.T) {
		graph, err := igraph.StaticFitnessGame(12, []float64{1, 2, 3, 4, 5, 6}, nil, igraph.StaticFitnessOptions{Seed: &seed})
		if err != nil {
			t.Fatalf("StaticFitnessGame failed: %v", err)
		}
		defer graph.Close()
		vertices, edges := mustCounts(t, graph)
		if vertices != 6 || edges != 12 {
			t.Fatalf("counts = (%d, %d), want (6, 12)", vertices, edges)
		}
		mustBeSimple(t, graph)
	})

	t.Run("directed", func(t *testing.T) {
		graph, err := igraph.StaticFitnessGame(
			9, []float64{1, 2, 3, 4}, []float64{4, 3, 2, 1},
			igraph.StaticFitnessOptions{Seed: &seed},
		)
		if err != nil {
			t.Fatalf("StaticFitnessGame failed: %v", err)
		}
		defer graph.Close()
		vertices, edges := mustCounts(t, graph)
		if vertices != 4 || edges != 9 {
			t.Fatalf("counts = (%d, %d), want (4, 9)", vertices, edges)
		}
		directed, err := graph.IsDirected()
		if err != nil || !directed {
			t.Fatalf("IsDirected = %v, %v, want true", directed, err)
		}
	})

	t.Run("directed fitness alignment", func(t *testing.T) {
		graph, err := igraph.StaticFitnessGame(
			4, []float64{1, 0, 0}, []float64{0, 1, 0},
			igraph.StaticFitnessOptions{Seed: &seed, EdgeTypes: igraph.EdgeTypeMulti},
		)
		if err != nil {
			t.Fatalf("StaticFitnessGame failed: %v", err)
		}
		defer graph.Close()
		for _, edge := range mustEdges(t, graph) {
			if edge != (igraph.Edge{From: 0, To: 1}) {
				t.Fatalf("edge = %v, want 0 -> 1", edge)
			}
		}
	})

	t.Run("loops and parallel edges", func(t *testing.T) {
		graph, err := igraph.StaticFitnessGame(
			5, []float64{1}, nil,
			igraph.StaticFitnessOptions{Seed: &seed, EdgeTypes: igraph.EdgeTypeLoopsAndMulti},
		)
		if err != nil {
			t.Fatalf("StaticFitnessGame failed: %v", err)
		}
		defer graph.Close()
		edges := mustEdges(t, graph)
		if len(edges) != 5 {
			t.Fatalf("edge count = %d, want 5", len(edges))
		}
		for _, edge := range edges {
			if edge != (igraph.Edge{From: 0, To: 0}) {
				t.Fatalf("edge = %v, want loop at vertex 0", edge)
			}
		}
	})

	t.Run("parallel edges without loops", func(t *testing.T) {
		graph, err := igraph.StaticFitnessGame(
			5, []float64{1, 1}, nil,
			igraph.StaticFitnessOptions{Seed: &seed, EdgeTypes: igraph.EdgeTypeMulti},
		)
		if err != nil {
			t.Fatalf("StaticFitnessGame failed: %v", err)
		}
		defer graph.Close()
		_, edges := mustCounts(t, graph)
		if edges != 5 {
			t.Fatalf("edge count = %d, want 5", edges)
		}
	})

	t.Run("empty and zero edges", func(t *testing.T) {
		for _, fitness := range [][]float64{nil, {0, 0, 0}} {
			graph, err := igraph.StaticFitnessGame(0, fitness, nil, igraph.StaticFitnessOptions{})
			if err != nil {
				t.Fatalf("StaticFitnessGame failed: %v", err)
			}
			vertices, edges := mustCounts(t, graph)
			if vertices != len(fitness) || edges != 0 {
				t.Errorf("counts = (%d, %d), want (%d, 0)", vertices, edges, len(fitness))
			}
			graph.Close()
		}
	})

	t.Run("same seed", func(t *testing.T) {
		fitness := []float64{1, 2, 3, 4, 5, 6, 7, 8}
		options := igraph.StaticFitnessOptions{Seed: &seed, EdgeTypes: igraph.EdgeTypeLoopsAndMulti}
		first, err := igraph.StaticFitnessGame(30, fitness, nil, options)
		if err != nil {
			t.Fatalf("first StaticFitnessGame failed: %v", err)
		}
		defer first.Close()
		second, err := igraph.StaticFitnessGame(30, fitness, nil, options)
		if err != nil {
			t.Fatalf("second StaticFitnessGame failed: %v", err)
		}
		defer second.Close()
		if !reflect.DeepEqual(mustEdges(t, first), mustEdges(t, second)) {
			t.Fatal("same seed produced different edge lists")
		}
		otherSeed := seed + 1
		third, err := igraph.StaticFitnessGame(30, fitness, nil, igraph.StaticFitnessOptions{
			Seed: &otherSeed, EdgeTypes: igraph.EdgeTypeLoopsAndMulti,
		})
		if err != nil {
			t.Fatalf("different-seed StaticFitnessGame failed: %v", err)
		}
		defer third.Close()
		if reflect.DeepEqual(mustEdges(t, first), mustEdges(t, third)) {
			t.Fatal("different seeds produced the same edge list")
		}
	})

	t.Run("validation", func(t *testing.T) {
		tests := []struct {
			name  string
			edges int
			out   []float64
			in    []float64
			opts  igraph.StaticFitnessOptions
		}{
			{name: "negative edges", edges: -1, out: []float64{1}},
			{name: "invalid edge types", out: []float64{1}, opts: igraph.StaticFitnessOptions{EdgeTypes: igraph.EdgeType(255)}},
			{name: "negative fitness", out: []float64{-1}},
			{name: "NaN fitness", out: []float64{math.NaN()}},
			{name: "infinite fitness", out: []float64{math.Inf(1)}},
			{name: "length mismatch", out: []float64{1, 1}, in: []float64{1}},
			{name: "negative in-fitness", out: []float64{1}, in: []float64{-1}},
			{name: "no eligible pair", edges: 1, out: []float64{1}},
			{name: "too many simple edges", edges: 2, out: []float64{1, 1}},
		}
		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				if graph, err := igraph.StaticFitnessGame(test.edges, test.out, test.in, test.opts); err == nil || graph != nil {
					t.Fatalf("StaticFitnessGame = %v, %v, want nil error result", graph, err)
				}
			})
		}
	})

	t.Run("returned graph ownership", func(t *testing.T) {
		graph, err := igraph.StaticFitnessGame(1, []float64{1, 1}, nil, igraph.StaticFitnessOptions{})
		if err != nil {
			t.Fatalf("StaticFitnessGame failed: %v", err)
		}
		if err := graph.Close(); err != nil {
			t.Fatalf("Close failed: %v", err)
		}
		if err := graph.Close(); err != nil {
			t.Fatalf("second Close failed: %v", err)
		}
		if _, err := graph.VertexCount(); !errors.Is(err, igraph.ErrClosed) {
			t.Fatalf("VertexCount after Close error = %v, want ErrClosed", err)
		}
	})
}

func TestStaticPowerLawGame(t *testing.T) {
	seed := uint64(335)

	t.Run("undirected", func(t *testing.T) {
		graph, err := igraph.StaticPowerLawGame(30, 40, 2.5, igraph.StaticPowerLawOptions{Seed: &seed, FiniteSizeCorrection: true})
		if err != nil {
			t.Fatalf("StaticPowerLawGame failed: %v", err)
		}
		defer graph.Close()
		vertices, edges := mustCounts(t, graph)
		if vertices != 30 || edges != 40 {
			t.Fatalf("counts = (%d, %d), want (30, 40)", vertices, edges)
		}
		mustBeSimple(t, graph)
	})

	t.Run("directed", func(t *testing.T) {
		inExponent := 3.0
		graph, err := igraph.StaticPowerLawGame(20, 30, 2.5, igraph.StaticPowerLawOptions{Seed: &seed, InExponent: &inExponent})
		if err != nil {
			t.Fatalf("StaticPowerLawGame failed: %v", err)
		}
		defer graph.Close()
		directed, err := graph.IsDirected()
		if err != nil || !directed {
			t.Fatalf("IsDirected = %v, %v, want true", directed, err)
		}
		vertices, edges := mustCounts(t, graph)
		if vertices != 20 || edges != 30 {
			t.Fatalf("counts = (%d, %d), want (20, 30)", vertices, edges)
		}
	})

	t.Run("single vertex loops and parallel edges", func(t *testing.T) {
		graph, err := igraph.StaticPowerLawGame(1, 3, 2.5, igraph.StaticPowerLawOptions{
			Seed: &seed, EdgeTypes: igraph.EdgeTypeLoopsAndMulti,
		})
		if err != nil {
			t.Fatalf("StaticPowerLawGame failed: %v", err)
		}
		defer graph.Close()
		_, edges := mustCounts(t, graph)
		if edges != 3 {
			t.Fatalf("edge count = %d, want 3", edges)
		}
	})

	t.Run("infinite exponent and reproducibility", func(t *testing.T) {
		options := igraph.StaticPowerLawOptions{Seed: &seed}
		first, err := igraph.StaticPowerLawGame(20, 20, math.Inf(1), options)
		if err != nil {
			t.Fatalf("first StaticPowerLawGame failed: %v", err)
		}
		defer first.Close()
		second, err := igraph.StaticPowerLawGame(20, 20, math.Inf(1), options)
		if err != nil {
			t.Fatalf("second StaticPowerLawGame failed: %v", err)
		}
		defer second.Close()
		if !reflect.DeepEqual(mustEdges(t, first), mustEdges(t, second)) {
			t.Fatal("same seed produced different edge lists")
		}
	})

	t.Run("empty", func(t *testing.T) {
		graph, err := igraph.StaticPowerLawGame(0, 0, 2, igraph.StaticPowerLawOptions{})
		if err != nil {
			t.Fatalf("StaticPowerLawGame failed: %v", err)
		}
		defer graph.Close()
		vertices, edges := mustCounts(t, graph)
		if vertices != 0 || edges != 0 {
			t.Fatalf("counts = (%d, %d), want (0, 0)", vertices, edges)
		}
	})

	t.Run("validation", func(t *testing.T) {
		badIn := 1.5
		validIn := 2.5
		tests := []struct {
			name       string
			vertices   int
			edges      int
			out        float64
			in         *float64
			edgeTypes  igraph.EdgeType
			wantPhrase string
		}{
			{name: "negative vertices", vertices: -1, out: 2},
			{name: "negative edges", edges: -1, out: 2},
			{name: "invalid edge type", out: 2, edgeTypes: igraph.EdgeType(255)},
			{name: "low out exponent", out: 1.5},
			{name: "NaN out exponent", out: math.NaN()},
			{name: "negative infinity", out: math.Inf(-1)},
			{name: "low in exponent", out: 2, in: &badIn},
			{name: "positive edges on empty", edges: 1, out: 2, in: &validIn, wantPhrase: "at least one vertex"},
		}
		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				graph, err := igraph.StaticPowerLawGame(test.vertices, test.edges, test.out, igraph.StaticPowerLawOptions{
					InExponent: test.in, EdgeTypes: test.edgeTypes,
				})
				if err == nil || graph != nil {
					t.Fatalf("StaticPowerLawGame = %v, %v, want nil error result", graph, err)
				}
				if test.wantPhrase != "" && !strings.Contains(err.Error(), test.wantPhrase) {
					t.Fatalf("error = %q, want phrase %q", err, test.wantPhrase)
				}
			})
		}
	})
}
