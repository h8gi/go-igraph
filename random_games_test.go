package igraph_test

import (
	"math"
	"reflect"
	"strings"
	"testing"

	"github.com/h8gi/go-igraph"
)

func mustCounts(t *testing.T, g *igraph.Graph) (vertices, edges int) {
	t.Helper()
	vertices, err := g.VertexCount()
	if err != nil {
		t.Fatalf("VertexCount failed: %v", err)
	}
	edges, err = g.EdgeCount()
	if err != nil {
		t.Fatalf("EdgeCount failed: %v", err)
	}
	return vertices, edges
}

func mustEdges(t *testing.T, g *igraph.Graph) []igraph.Edge {
	t.Helper()
	edges, err := g.Edges()
	if err != nil {
		t.Fatalf("Edges failed: %v", err)
	}
	return edges
}

func mustBeSimple(t *testing.T, g *igraph.Graph) {
	t.Helper()
	directed, err := g.IsDirected()
	if err != nil {
		t.Fatalf("IsDirected failed: %v", err)
	}
	seen := make(map[igraph.Edge]bool)
	for _, e := range mustEdges(t, g) {
		if e.From == e.To {
			t.Errorf("expected no self-loops, got loop at vertex %d", e.From)
		}
		key := e
		if !directed && key.From > key.To {
			key = igraph.Edge{From: e.To, To: e.From}
		}
		if seen[key] {
			t.Errorf("expected no parallel edges, got duplicate edge %v", key)
		}
		seen[key] = true
	}
}

func TestErdosRenyiGNM(t *testing.T) {
	seed := uint64(42)

	t.Run("undirected simple", func(t *testing.T) {
		g, err := igraph.ErdosRenyiGNM(10, 20, false, false, igraph.ErdosRenyiOptions{Seed: &seed})
		if err != nil {
			t.Fatalf("ErdosRenyiGNM failed: %v", err)
		}
		defer g.Close()

		vertices, edges := mustCounts(t, g)
		if vertices != 10 || edges != 20 {
			t.Errorf("expected 10 vertices and 20 edges, got %d and %d", vertices, edges)
		}
		directed, err := g.IsDirected()
		if err != nil {
			t.Fatalf("IsDirected failed: %v", err)
		}
		if directed {
			t.Errorf("expected an undirected graph")
		}
		mustBeSimple(t, g)
	})

	t.Run("directed", func(t *testing.T) {
		g, err := igraph.ErdosRenyiGNM(6, 12, true, false, igraph.ErdosRenyiOptions{Seed: &seed})
		if err != nil {
			t.Fatalf("ErdosRenyiGNM failed: %v", err)
		}
		defer g.Close()

		vertices, edges := mustCounts(t, g)
		if vertices != 6 || edges != 12 {
			t.Errorf("expected 6 vertices and 12 edges, got %d and %d", vertices, edges)
		}
		directed, err := g.IsDirected()
		if err != nil {
			t.Fatalf("IsDirected failed: %v", err)
		}
		if !directed {
			t.Errorf("expected a directed graph")
		}
		mustBeSimple(t, g)
	})

	t.Run("single vertex loop", func(t *testing.T) {
		// The only possible edge on one vertex with loops is the self-loop.
		g, err := igraph.ErdosRenyiGNM(1, 1, false, true, igraph.ErdosRenyiOptions{Seed: &seed})
		if err != nil {
			t.Fatalf("ErdosRenyiGNM failed: %v", err)
		}
		defer g.Close()

		edges := mustEdges(t, g)
		if len(edges) != 1 || edges[0].From != 0 || edges[0].To != 0 {
			t.Errorf("expected exactly one self-loop at vertex 0, got %v", edges)
		}
	})

	t.Run("zero vertices", func(t *testing.T) {
		g, err := igraph.ErdosRenyiGNM(0, 0, false, false, igraph.ErdosRenyiOptions{})
		if err != nil {
			t.Fatalf("ErdosRenyiGNM failed: %v", err)
		}
		defer g.Close()

		vertices, edges := mustCounts(t, g)
		if vertices != 0 || edges != 0 {
			t.Errorf("expected an empty graph, got %d vertices and %d edges", vertices, edges)
		}
	})

	t.Run("zero edges", func(t *testing.T) {
		g, err := igraph.ErdosRenyiGNM(5, 0, false, false, igraph.ErdosRenyiOptions{})
		if err != nil {
			t.Fatalf("ErdosRenyiGNM failed: %v", err)
		}
		defer g.Close()

		vertices, edges := mustCounts(t, g)
		if vertices != 5 || edges != 0 {
			t.Errorf("expected 5 vertices and 0 edges, got %d and %d", vertices, edges)
		}
	})

	t.Run("seed reproducibility", func(t *testing.T) {
		first, err := igraph.ErdosRenyiGNM(30, 60, false, false, igraph.ErdosRenyiOptions{Seed: &seed})
		if err != nil {
			t.Fatalf("ErdosRenyiGNM failed: %v", err)
		}
		defer first.Close()
		second, err := igraph.ErdosRenyiGNM(30, 60, false, false, igraph.ErdosRenyiOptions{Seed: &seed})
		if err != nil {
			t.Fatalf("ErdosRenyiGNM failed on second run: %v", err)
		}
		defer second.Close()

		if !reflect.DeepEqual(mustEdges(t, first), mustEdges(t, second)) {
			t.Errorf("expected identical edge lists for the same seed")
		}
	})

	t.Run("invalid parameters", func(t *testing.T) {
		if _, err := igraph.ErdosRenyiGNM(-1, 0, false, false, igraph.ErdosRenyiOptions{}); err == nil {
			t.Errorf("expected an error for negative vertex count")
		}
		if _, err := igraph.ErdosRenyiGNM(0, -1, false, false, igraph.ErdosRenyiOptions{}); err == nil {
			t.Errorf("expected an error for negative edge count")
		}
	})

	t.Run("upstream error for impossible edge count", func(t *testing.T) {
		// A simple undirected graph on 3 vertices holds at most 3 edges.
		if _, err := igraph.ErdosRenyiGNM(3, 100, false, false, igraph.ErdosRenyiOptions{}); err == nil {
			t.Errorf("expected an upstream error for an impossible edge count")
		}
	})
}

func TestErdosRenyiGNP(t *testing.T) {
	seed := uint64(42)

	t.Run("probability zero", func(t *testing.T) {
		g, err := igraph.ErdosRenyiGNP(10, 0, false, false, igraph.ErdosRenyiOptions{})
		if err != nil {
			t.Fatalf("ErdosRenyiGNP failed: %v", err)
		}
		defer g.Close()

		vertices, edges := mustCounts(t, g)
		if vertices != 10 || edges != 0 {
			t.Errorf("expected 10 vertices and 0 edges, got %d and %d", vertices, edges)
		}
	})

	t.Run("probability one", func(t *testing.T) {
		cases := []struct {
			name     string
			directed bool
			loops    bool
			want     int
		}{
			{name: "undirected simple", directed: false, loops: false, want: 5 * 4 / 2},
			{name: "undirected loops", directed: false, loops: true, want: 5*4/2 + 5},
			{name: "directed simple", directed: true, loops: false, want: 5 * 4},
			{name: "directed loops", directed: true, loops: true, want: 5 * 5},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				g, err := igraph.ErdosRenyiGNP(5, 1, tc.directed, tc.loops, igraph.ErdosRenyiOptions{})
				if err != nil {
					t.Fatalf("ErdosRenyiGNP failed: %v", err)
				}
				defer g.Close()

				_, edges := mustCounts(t, g)
				if edges != tc.want {
					t.Errorf("expected %d edges, got %d", tc.want, edges)
				}
			})
		}
	})

	t.Run("zero vertices", func(t *testing.T) {
		g, err := igraph.ErdosRenyiGNP(0, 0.5, false, false, igraph.ErdosRenyiOptions{})
		if err != nil {
			t.Fatalf("ErdosRenyiGNP failed: %v", err)
		}
		defer g.Close()

		vertices, edges := mustCounts(t, g)
		if vertices != 0 || edges != 0 {
			t.Errorf("expected an empty graph, got %d vertices and %d edges", vertices, edges)
		}
	})

	t.Run("seed reproducibility", func(t *testing.T) {
		first, err := igraph.ErdosRenyiGNP(30, 0.2, true, false, igraph.ErdosRenyiOptions{Seed: &seed})
		if err != nil {
			t.Fatalf("ErdosRenyiGNP failed: %v", err)
		}
		defer first.Close()
		second, err := igraph.ErdosRenyiGNP(30, 0.2, true, false, igraph.ErdosRenyiOptions{Seed: &seed})
		if err != nil {
			t.Fatalf("ErdosRenyiGNP failed on second run: %v", err)
		}
		defer second.Close()

		if !reflect.DeepEqual(mustEdges(t, first), mustEdges(t, second)) {
			t.Errorf("expected identical edge lists for the same seed")
		}
	})

	t.Run("invalid parameters", func(t *testing.T) {
		if _, err := igraph.ErdosRenyiGNP(-1, 0.5, false, false, igraph.ErdosRenyiOptions{}); err == nil {
			t.Errorf("expected an error for negative vertex count")
		}
		invalid := []float64{-0.1, 1.1, math.NaN()}
		for _, p := range invalid {
			if _, err := igraph.ErdosRenyiGNP(5, p, false, false, igraph.ErdosRenyiOptions{}); err == nil {
				t.Errorf("expected an error for probability %v", p)
			}
		}
	})
}

func TestKRegularGame(t *testing.T) {
	seed := uint64(42)

	t.Run("undirected simple", func(t *testing.T) {
		g, err := igraph.KRegularGame(10, 3, false, false, igraph.KRegularOptions{Seed: &seed})
		if err != nil {
			t.Fatalf("KRegularGame failed: %v", err)
		}
		defer g.Close()

		vertices, edges := mustCounts(t, g)
		if vertices != 10 || edges != 10*3/2 {
			t.Errorf("expected 10 vertices and 15 edges, got %d and %d", vertices, edges)
		}
		degrees, err := g.Degree(igraph.AllVertices(), igraph.DegreeOptions{Direction: igraph.DirectionAll, CountLoops: true})
		if err != nil {
			t.Fatalf("Degree failed: %v", err)
		}
		for vertex, degree := range degrees {
			if degree != 3 {
				t.Errorf("expected degree 3 at vertex %d, got %d", vertex, degree)
			}
		}
		mustBeSimple(t, g)
	})

	t.Run("directed", func(t *testing.T) {
		g, err := igraph.KRegularGame(6, 2, true, false, igraph.KRegularOptions{Seed: &seed})
		if err != nil {
			t.Fatalf("KRegularGame failed: %v", err)
		}
		defer g.Close()

		for _, direction := range []igraph.DirectionMode{igraph.DirectionOut, igraph.DirectionIn} {
			degrees, err := g.Degree(igraph.AllVertices(), igraph.DegreeOptions{Direction: direction, CountLoops: true})
			if err != nil {
				t.Fatalf("Degree failed: %v", err)
			}
			for vertex, degree := range degrees {
				if degree != 2 {
					t.Errorf("expected degree 2 at vertex %d for direction %d, got %d", vertex, direction, degree)
				}
			}
		}
	})

	t.Run("degree zero", func(t *testing.T) {
		g, err := igraph.KRegularGame(4, 0, false, false, igraph.KRegularOptions{})
		if err != nil {
			t.Fatalf("KRegularGame failed: %v", err)
		}
		defer g.Close()

		vertices, edges := mustCounts(t, g)
		if vertices != 4 || edges != 0 {
			t.Errorf("expected 4 vertices and 0 edges, got %d and %d", vertices, edges)
		}
	})

	t.Run("zero vertices", func(t *testing.T) {
		g, err := igraph.KRegularGame(0, 0, false, false, igraph.KRegularOptions{})
		if err != nil {
			t.Fatalf("KRegularGame failed: %v", err)
		}
		defer g.Close()

		vertices, edges := mustCounts(t, g)
		if vertices != 0 || edges != 0 {
			t.Errorf("expected an empty graph, got %d vertices and %d edges", vertices, edges)
		}
	})

	t.Run("multiple edges", func(t *testing.T) {
		g, err := igraph.KRegularGame(2, 3, false, true, igraph.KRegularOptions{Seed: &seed})
		if err != nil {
			t.Fatalf("KRegularGame failed: %v", err)
		}
		defer g.Close()

		degrees, err := g.Degree(igraph.AllVertices(), igraph.DegreeOptions{Direction: igraph.DirectionAll, CountLoops: true})
		if err != nil {
			t.Fatalf("Degree failed: %v", err)
		}
		for vertex, degree := range degrees {
			if degree != 3 {
				t.Errorf("expected degree 3 at vertex %d, got %d", vertex, degree)
			}
		}
	})

	t.Run("seed reproducibility", func(t *testing.T) {
		first, err := igraph.KRegularGame(20, 4, false, false, igraph.KRegularOptions{Seed: &seed})
		if err != nil {
			t.Fatalf("KRegularGame failed: %v", err)
		}
		defer first.Close()
		second, err := igraph.KRegularGame(20, 4, false, false, igraph.KRegularOptions{Seed: &seed})
		if err != nil {
			t.Fatalf("KRegularGame failed on second run: %v", err)
		}
		defer second.Close()

		if !reflect.DeepEqual(mustEdges(t, first), mustEdges(t, second)) {
			t.Errorf("expected identical edge lists for the same seed")
		}
	})

	t.Run("invalid parameters", func(t *testing.T) {
		if _, err := igraph.KRegularGame(-1, 0, false, false, igraph.KRegularOptions{}); err == nil {
			t.Errorf("expected an error for negative vertex count")
		}
		if _, err := igraph.KRegularGame(4, -1, false, false, igraph.KRegularOptions{}); err == nil {
			t.Errorf("expected an error for negative degree")
		}
		if _, err := igraph.KRegularGame(4, 4, false, false, igraph.KRegularOptions{}); err == nil {
			t.Errorf("expected an error for a simple graph degree not smaller than the vertex count")
		}
		if _, err := igraph.KRegularGame(3, 3, false, true, igraph.KRegularOptions{}); err == nil {
			t.Errorf("expected an error for an odd undirected degree sum")
		}
	})
}

func TestRandomTreeGame(t *testing.T) {
	seed := uint64(42)

	t.Run("undirected methods", func(t *testing.T) {
		for _, method := range []igraph.TreeGameMethod{igraph.TreeGamePrufer, igraph.TreeGameLERW} {
			g, err := igraph.RandomTreeGame(10, false, method, igraph.TreeGameOptions{Seed: &seed})
			if err != nil {
				t.Fatalf("RandomTreeGame failed for method %d: %v", method, err)
			}
			defer g.Close()

			vertices, edges := mustCounts(t, g)
			if vertices != 10 || edges != 9 {
				t.Errorf("expected a tree with 10 vertices and 9 edges, got %d and %d", vertices, edges)
			}
			connected, err := g.IsConnected(igraph.ConnectednessWeak)
			if err != nil {
				t.Fatalf("IsConnected failed: %v", err)
			}
			if !connected {
				t.Errorf("expected a connected tree for method %d", method)
			}
		}
	})

	t.Run("directed LERW", func(t *testing.T) {
		g, err := igraph.RandomTreeGame(10, true, igraph.TreeGameLERW, igraph.TreeGameOptions{Seed: &seed})
		if err != nil {
			t.Fatalf("RandomTreeGame failed: %v", err)
		}
		defer g.Close()

		directed, err := g.IsDirected()
		if err != nil {
			t.Fatalf("IsDirected failed: %v", err)
		}
		if !directed {
			t.Errorf("expected a directed tree")
		}
		vertices, edges := mustCounts(t, g)
		if vertices != 10 || edges != 9 {
			t.Errorf("expected a tree with 10 vertices and 9 edges, got %d and %d", vertices, edges)
		}
		connected, err := g.IsConnected(igraph.ConnectednessWeak)
		if err != nil {
			t.Fatalf("IsConnected failed: %v", err)
		}
		if !connected {
			t.Errorf("expected a weakly connected tree")
		}
	})

	t.Run("small trees", func(t *testing.T) {
		for n := 0; n <= 1; n++ {
			g, err := igraph.RandomTreeGame(n, false, igraph.TreeGameLERW, igraph.TreeGameOptions{})
			if err != nil {
				t.Fatalf("RandomTreeGame failed for %d vertices: %v", n, err)
			}
			defer g.Close()

			vertices, edges := mustCounts(t, g)
			if vertices != n || edges != 0 {
				t.Errorf("expected %d vertices and 0 edges, got %d and %d", n, vertices, edges)
			}
		}
	})

	t.Run("seed reproducibility", func(t *testing.T) {
		for _, method := range []igraph.TreeGameMethod{igraph.TreeGamePrufer, igraph.TreeGameLERW} {
			first, err := igraph.RandomTreeGame(20, false, method, igraph.TreeGameOptions{Seed: &seed})
			if err != nil {
				t.Fatalf("RandomTreeGame failed for method %d: %v", method, err)
			}
			defer first.Close()
			second, err := igraph.RandomTreeGame(20, false, method, igraph.TreeGameOptions{Seed: &seed})
			if err != nil {
				t.Fatalf("RandomTreeGame failed on second run for method %d: %v", method, err)
			}
			defer second.Close()

			if !reflect.DeepEqual(mustEdges(t, first), mustEdges(t, second)) {
				t.Errorf("expected identical edge lists for the same seed with method %d", method)
			}
		}
	})

	t.Run("invalid parameters", func(t *testing.T) {
		if _, err := igraph.RandomTreeGame(-1, false, igraph.TreeGameLERW, igraph.TreeGameOptions{}); err == nil {
			t.Errorf("expected an error for negative vertex count")
		}
		if _, err := igraph.RandomTreeGame(5, true, igraph.TreeGamePrufer, igraph.TreeGameOptions{}); err == nil {
			t.Errorf("expected an error for a directed Prüfer tree")
		}
		if _, err := igraph.RandomTreeGame(5, false, igraph.TreeGameMethod(99), igraph.TreeGameOptions{}); err == nil {
			t.Errorf("expected an error for an invalid method")
		}
	})
}

func TestDegreeSequenceGame(t *testing.T) {
	seed := uint64(12345)

	t.Run("undirected valid methods", func(t *testing.T) {
		outDeg := []int{2, 2, 2, 2} // 4-cycle or 2 2-cycles
		methods := []igraph.DegSeqMethod{
			igraph.DegSeqConfiguration,
			igraph.DegSeqVL,
			igraph.DegSeqSimpleNoMultiple,
			igraph.DegSeqSimpleNoMultipleUniform,
			igraph.DegSeqEdgeSwitchingSimple,
			igraph.DegSeqSimple,
		}

		for _, method := range methods {
			g, err := igraph.DegreeSequenceGame(outDeg, nil, method, igraph.DegSeqOptions{Seed: &seed})
			if err != nil {
				t.Fatalf("DegreeSequenceGame failed for method %d: %v", method, err)
			}
			defer g.Close()

			vertices, edges := mustCounts(t, g)
			if vertices != 4 || edges != 4 {
				t.Errorf("expected 4 vertices and 4 edges for method %d, got %d and %d", method, vertices, edges)
			}

			directed, err := g.IsDirected()
			if err != nil {
				t.Fatalf("IsDirected failed: %v", err)
			}
			if directed {
				t.Errorf("expected undirected graph for method %d", method)
			}
		}
	})

	t.Run("directed valid methods", func(t *testing.T) {
		outDeg := []int{1, 1, 1}
		inDeg := []int{1, 1, 1} // 3-cycle

		methods := []igraph.DegSeqMethod{
			igraph.DegSeqConfiguration,
			igraph.DegSeqSimpleNoMultiple,
			igraph.DegSeqSimpleNoMultipleUniform,
			igraph.DegSeqEdgeSwitchingSimple,
		}

		for _, method := range methods {
			g, err := igraph.DegreeSequenceGame(outDeg, inDeg, method, igraph.DegSeqOptions{Seed: &seed})
			if err != nil {
				t.Fatalf("DegreeSequenceGame failed for directed method %d: %v", method, err)
			}
			defer g.Close()

			vertices, edges := mustCounts(t, g)
			if vertices != 3 || edges != 3 {
				t.Errorf("expected 3 vertices and 3 edges for method %d, got %d and %d", method, vertices, edges)
			}

			directed, err := g.IsDirected()
			if err != nil {
				t.Fatalf("IsDirected failed: %v", err)
			}
			if !directed {
				t.Errorf("expected directed graph for method %d", method)
			}
		}
	})

	t.Run("empty sequence", func(t *testing.T) {
		// A nil or empty inDeg selects the undirected form.
		for name, inDeg := range map[string][]int{"nil inDeg": nil, "empty inDeg": {}} {
			g, err := igraph.DegreeSequenceGame([]int{}, inDeg, igraph.DegSeqConfiguration, igraph.DegSeqOptions{})
			if err != nil {
				t.Fatalf("DegreeSequenceGame failed for empty sequence with %s: %v", name, err)
			}
			defer g.Close()

			vertices, edges := mustCounts(t, g)
			if vertices != 0 || edges != 0 {
				t.Errorf("expected 0 vertices and 0 edges with %s, got %d and %d", name, vertices, edges)
			}
			directed, err := g.IsDirected()
			if err != nil {
				t.Fatalf("IsDirected failed: %v", err)
			}
			if directed {
				t.Errorf("expected an undirected graph with %s", name)
			}
		}
	})

	t.Run("seed reproducibility", func(t *testing.T) {
		outDeg := []int{3, 3, 3, 3}
		g1, err := igraph.DegreeSequenceGame(outDeg, nil, igraph.DegSeqConfiguration, igraph.DegSeqOptions{Seed: &seed})
		if err != nil {
			t.Fatalf("DegreeSequenceGame failed: %v", err)
		}
		defer g1.Close()

		g2, err := igraph.DegreeSequenceGame(outDeg, nil, igraph.DegSeqConfiguration, igraph.DegSeqOptions{Seed: &seed})
		if err != nil {
			t.Fatalf("DegreeSequenceGame failed: %v", err)
		}
		defer g2.Close()

		if !reflect.DeepEqual(mustEdges(t, g1), mustEdges(t, g2)) {
			t.Errorf("expected identical edge lists for same seed")
		}
	})

	t.Run("invalid parameters", func(t *testing.T) {
		// Negative degree
		if _, err := igraph.DegreeSequenceGame([]int{2, -1, 1}, nil, igraph.DegSeqConfiguration, igraph.DegSeqOptions{}); err == nil {
			t.Errorf("expected error for negative degree")
		}

		// Undirected odd degree sum yields the descriptive Go error for
		// every method, including the default configuration model.
		for _, method := range []igraph.DegSeqMethod{igraph.DegSeqConfiguration, igraph.DegSeqVL, igraph.DegSeqSimpleNoMultiple} {
			_, err := igraph.DegreeSequenceGame([]int{1, 1, 1}, nil, method, igraph.DegSeqOptions{})
			if err == nil {
				t.Errorf("expected error for undirected odd degree sum with method %d", method)
			} else if !strings.Contains(err.Error(), "must be even") {
				t.Errorf("expected descriptive even-sum error for method %d, got: %v", method, err)
			}
		}

		// Directed mismatched slice lengths
		if _, err := igraph.DegreeSequenceGame([]int{1, 1}, []int{1, 1, 1}, igraph.DegSeqConfiguration, igraph.DegSeqOptions{}); err == nil {
			t.Errorf("expected error for mismatched in/out slice lengths")
		}

		// Directed mismatched sums
		if _, err := igraph.DegreeSequenceGame([]int{1, 2}, []int{1, 1}, igraph.DegSeqConfiguration, igraph.DegSeqOptions{}); err == nil {
			t.Errorf("expected error for mismatched in/out degree sums")
		}

		// Directed with VL method
		if _, err := igraph.DegreeSequenceGame([]int{1, 1}, []int{1, 1}, igraph.DegSeqVL, igraph.DegSeqOptions{}); err == nil {
			t.Errorf("expected error for directed graph with VL method")
		}

		// Invalid method
		if _, err := igraph.DegreeSequenceGame([]int{2, 2, 2, 2}, nil, igraph.DegSeqMethod(99), igraph.DegSeqOptions{}); err == nil {
			t.Errorf("expected error for invalid DegSeqMethod")
		}
	})

	t.Run("upstream error for non-graphical sequence", func(t *testing.T) {
		// [3, 3, 1, 1] has an even sum, so it passes Go-side validation, but
		// it violates the Erdős-Gallai condition and no simple graph can
		// realize it: the upstream sampler must report the failure.
		if _, err := igraph.DegreeSequenceGame([]int{3, 3, 1, 1}, nil, igraph.DegSeqSimpleNoMultiple, igraph.DegSeqOptions{}); err == nil {
			t.Errorf("expected an upstream error for a non-graphical simple sequence")
		}
	})
}

func TestBarabasiGame(t *testing.T) {
	seed := uint64(42)

	t.Run("undirected simple", func(t *testing.T) {
		g, err := igraph.BarabasiGame(10, 2, 1.0, 1.0, false, igraph.BarabasiOptions{Seed: &seed})
		if err != nil {
			t.Fatalf("BarabasiGame failed: %v", err)
		}
		defer g.Close()

		vertices, edges := mustCounts(t, g)
		if vertices != 10 {
			t.Errorf("expected 10 vertices, got %d", vertices)
		}
		if edges == 0 {
			t.Errorf("expected non-zero edge count, got 0")
		}
		directed, err := g.IsDirected()
		if err != nil {
			t.Fatalf("IsDirected failed: %v", err)
		}
		if directed {
			t.Errorf("expected undirected graph")
		}
	})

	t.Run("directed and algorithms", func(t *testing.T) {
		algos := []igraph.BarabasiAlgorithm{
			igraph.BarabasiBag,
			igraph.BarabasiPSumTree,
			igraph.BarabasiPSumTreeMultiple,
		}
		for _, algo := range algos {
			g, err := igraph.BarabasiGame(10, 2, 1.0, 1.0, true, igraph.BarabasiOptions{
				Seed:      &seed,
				Algorithm: algo,
			})
			if err != nil {
				t.Fatalf("BarabasiGame failed for algo %d: %v", algo, err)
			}
			defer g.Close()

			vertices, _ := mustCounts(t, g)
			if vertices != 10 {
				t.Errorf("expected 10 vertices for algo %d, got %d", algo, vertices)
			}
			directed, err := g.IsDirected()
			if err != nil {
				t.Fatalf("IsDirected failed: %v", err)
			}
			if !directed {
				t.Errorf("expected directed graph for algo %d", algo)
			}
		}
	})

	t.Run("custom OutSeq", func(t *testing.T) {
		outSeq := []int{0, 1, 2, 1, 2}
		g, err := igraph.BarabasiGame(5, 0, 1.0, 1.0, false, igraph.BarabasiOptions{
			Seed:   &seed,
			OutSeq: outSeq,
		})
		if err != nil {
			t.Fatalf("BarabasiGame with OutSeq failed: %v", err)
		}
		defer g.Close()

		vertices, _ := mustCounts(t, g)
		if vertices != 5 {
			t.Errorf("expected 5 vertices, got %d", vertices)
		}
	})

	t.Run("StartFrom initial graph", func(t *testing.T) {
		start, err := igraph.NewFull(3, false, false)
		if err != nil {
			t.Fatalf("FullGraph failed: %v", err)
		}
		defer start.Close()

		g, err := igraph.BarabasiGame(6, 2, 1.0, 1.0, false, igraph.BarabasiOptions{
			Seed:      &seed,
			StartFrom: start,
		})
		if err != nil {
			t.Fatalf("BarabasiGame with StartFrom failed: %v", err)
		}
		defer g.Close()

		vertices, _ := mustCounts(t, g)
		if vertices != 6 {
			t.Errorf("expected 6 vertices, got %d", vertices)
		}
	})

	t.Run("seed reproducibility", func(t *testing.T) {
		first, err := igraph.BarabasiGame(20, 2, 1.0, 1.0, false, igraph.BarabasiOptions{Seed: &seed})
		if err != nil {
			t.Fatalf("BarabasiGame failed: %v", err)
		}
		defer first.Close()

		second, err := igraph.BarabasiGame(20, 2, 1.0, 1.0, false, igraph.BarabasiOptions{Seed: &seed})
		if err != nil {
			t.Fatalf("BarabasiGame failed on second run: %v", err)
		}
		defer second.Close()

		if !reflect.DeepEqual(mustEdges(t, first), mustEdges(t, second)) {
			t.Errorf("expected identical edge lists for the same seed")
		}
	})

	t.Run("invalid parameters", func(t *testing.T) {
		if _, err := igraph.BarabasiGame(-1, 2, 1.0, 1.0, false, igraph.BarabasiOptions{}); err == nil {
			t.Errorf("expected error for negative n")
		}
		if _, err := igraph.BarabasiGame(10, -1, 1.0, 1.0, false, igraph.BarabasiOptions{}); err == nil {
			t.Errorf("expected error for negative m")
		}
		if _, err := igraph.BarabasiGame(10, 2, math.NaN(), 1.0, false, igraph.BarabasiOptions{}); err == nil {
			t.Errorf("expected error for NaN power")
		}
		if _, err := igraph.BarabasiGame(10, 2, 1.0, -0.5, false, igraph.BarabasiOptions{}); err == nil {
			t.Errorf("expected error for negative zeroAppeal")
		}
		if _, err := igraph.BarabasiGame(10, 2, 1.0, math.NaN(), false, igraph.BarabasiOptions{}); err == nil {
			t.Errorf("expected error for NaN zeroAppeal")
		}
		if _, err := igraph.BarabasiGame(10, 2, 1.0, 1.0, false, igraph.BarabasiOptions{Algorithm: igraph.BarabasiAlgorithm(99)}); err == nil {
			t.Errorf("expected error for invalid algorithm")
		}
		if _, err := igraph.BarabasiGame(5, 2, 1.0, 1.0, false, igraph.BarabasiOptions{OutSeq: []int{1, 2}}); err == nil {
			t.Errorf("expected error for mismatched OutSeq length")
		}
		if _, err := igraph.BarabasiGame(5, 2, 1.0, 1.0, false, igraph.BarabasiOptions{OutSeq: []int{}}); err == nil {
			t.Errorf("expected error for empty non-nil OutSeq with n > 0")
		}
		if _, err := igraph.BarabasiGame(2, 2, 1.0, 1.0, false, igraph.BarabasiOptions{OutSeq: []int{-1, 1}}); err == nil {
			t.Errorf("expected error for negative element in OutSeq")
		}

		closed, _ := igraph.NewFull(2, false, false)
		closed.Close()
		if _, err := igraph.BarabasiGame(5, 2, 1.0, 1.0, false, igraph.BarabasiOptions{StartFrom: closed}); err == nil {
			t.Errorf("expected error for closed StartFrom graph")
		}
	})
}

func TestWattsStrogatzGame(t *testing.T) {
	seed := uint64(42)

	t.Run("undirected simple small-world", func(t *testing.T) {
		g, err := igraph.WattsStrogatzGame(1, 20, 2, 0.1, false, false, igraph.WattsStrogatzOptions{Seed: &seed})
		if err != nil {
			t.Fatalf("WattsStrogatzGame failed: %v", err)
		}
		defer g.Close()

		vertices, edges := mustCounts(t, g)
		if vertices != 20 {
			t.Errorf("expected 20 vertices, got %d", vertices)
		}
		if edges != 20*2 {
			t.Errorf("expected 40 edges, got %d", edges)
		}
	})

	t.Run("loops and multiple edge types", func(t *testing.T) {
		g, err := igraph.WattsStrogatzGame(1, 10, 1, 0.5, true, true, igraph.WattsStrogatzOptions{Seed: &seed})
		if err != nil {
			t.Fatalf("WattsStrogatzGame failed: %v", err)
		}
		defer g.Close()

		vertices, _ := mustCounts(t, g)
		if vertices != 10 {
			t.Errorf("expected 10 vertices, got %d", vertices)
		}
	})

	t.Run("seed reproducibility", func(t *testing.T) {
		first, err := igraph.WattsStrogatzGame(1, 20, 2, 0.2, false, false, igraph.WattsStrogatzOptions{Seed: &seed})
		if err != nil {
			t.Fatalf("WattsStrogatzGame failed: %v", err)
		}
		defer first.Close()

		second, err := igraph.WattsStrogatzGame(1, 20, 2, 0.2, false, false, igraph.WattsStrogatzOptions{Seed: &seed})
		if err != nil {
			t.Fatalf("WattsStrogatzGame failed on second run: %v", err)
		}
		defer second.Close()

		if !reflect.DeepEqual(mustEdges(t, first), mustEdges(t, second)) {
			t.Errorf("expected identical edge lists for the same seed")
		}
	})

	t.Run("invalid parameters", func(t *testing.T) {
		if _, err := igraph.WattsStrogatzGame(0, 10, 2, 0.1, false, false, igraph.WattsStrogatzOptions{}); err == nil {
			t.Errorf("expected error for dim < 1")
		}
		if _, err := igraph.WattsStrogatzGame(1, 0, 2, 0.1, false, false, igraph.WattsStrogatzOptions{}); err == nil {
			t.Errorf("expected error for size < 1")
		}
		if _, err := igraph.WattsStrogatzGame(1, 10, -1, 0.1, false, false, igraph.WattsStrogatzOptions{}); err == nil {
			t.Errorf("expected error for nei < 0")
		}
		invalidP := []float64{-0.1, 1.1, math.NaN()}
		for _, p := range invalidP {
			if _, err := igraph.WattsStrogatzGame(1, 10, 2, p, false, false, igraph.WattsStrogatzOptions{}); err == nil {
				t.Errorf("expected error for invalid probability p = %g", p)
			}
		}
	})
}

func TestSBMGame(t *testing.T) {
	seed := uint64(42)

	t.Run("undirected simple SBM", func(t *testing.T) {
		prefMat, err := igraph.NewMatrixFromRows([][]float64{
			{0.8, 0.1},
			{0.1, 0.8},
		})
		if err != nil {
			t.Fatalf("NewMatrixFromRows failed: %v", err)
		}

		g, err := igraph.SBMGame(10, prefMat, []int{5, 5}, false, false, igraph.SBMOptions{Seed: &seed})
		if err != nil {
			t.Fatalf("SBMGame failed: %v", err)
		}
		defer g.Close()

		vertices, _ := mustCounts(t, g)
		if vertices != 10 {
			t.Errorf("expected 10 vertices, got %d", vertices)
		}
		directed, err := g.IsDirected()
		if err != nil {
			t.Fatalf("IsDirected failed: %v", err)
		}
		if directed {
			t.Errorf("expected undirected graph")
		}
	})

	t.Run("directed SBM with loops", func(t *testing.T) {
		prefMat, err := igraph.NewMatrixFromRows([][]float64{
			{0.5, 0.2},
			{0.1, 0.5},
		})
		if err != nil {
			t.Fatalf("NewMatrixFromRows failed: %v", err)
		}

		g, err := igraph.SBMGame(6, prefMat, []int{3, 3}, true, true, igraph.SBMOptions{Seed: &seed})
		if err != nil {
			t.Fatalf("SBMGame failed: %v", err)
		}
		defer g.Close()

		vertices, _ := mustCounts(t, g)
		if vertices != 6 {
			t.Errorf("expected 6 vertices, got %d", vertices)
		}
		directed, err := g.IsDirected()
		if err != nil {
			t.Fatalf("IsDirected failed: %v", err)
		}
		if !directed {
			t.Errorf("expected directed graph")
		}
	})

	t.Run("seed reproducibility", func(t *testing.T) {
		prefMat, _ := igraph.NewMatrixFromRows([][]float64{
			{0.7, 0.2},
			{0.2, 0.7},
		})

		first, err := igraph.SBMGame(10, prefMat, []int{5, 5}, false, false, igraph.SBMOptions{Seed: &seed})
		if err != nil {
			t.Fatalf("SBMGame failed: %v", err)
		}
		defer first.Close()

		second, err := igraph.SBMGame(10, prefMat, []int{5, 5}, false, false, igraph.SBMOptions{Seed: &seed})
		if err != nil {
			t.Fatalf("SBMGame failed on second run: %v", err)
		}
		defer second.Close()

		if !reflect.DeepEqual(mustEdges(t, first), mustEdges(t, second)) {
			t.Errorf("expected identical edge lists for the same seed")
		}
	})

	t.Run("invalid parameters", func(t *testing.T) {
		mat2x2, _ := igraph.NewMatrixFromRows([][]float64{{0.5, 0.5}, {0.5, 0.5}})

		if _, err := igraph.SBMGame(-1, mat2x2, []int{1, 1}, false, false, igraph.SBMOptions{}); err == nil {
			t.Errorf("expected error for negative n")
		}
		if _, err := igraph.SBMGame(10, mat2x2, []int{5, -1}, false, false, igraph.SBMOptions{}); err == nil {
			t.Errorf("expected error for negative block size")
		}
		if _, err := igraph.SBMGame(10, mat2x2, []int{5, 4}, false, false, igraph.SBMOptions{}); err == nil {
			t.Errorf("expected error for blockSizes sum != n")
		}
		if _, err := igraph.SBMGame(10, mat2x2, []int{3, 3, 4}, false, false, igraph.SBMOptions{}); err == nil {
			t.Errorf("expected error for matrix size mismatch with blockSizes count")
		}

		badMat, _ := igraph.NewMatrixFromRows([][]float64{{-0.1, 0.5}, {0.5, 0.5}})
		if _, err := igraph.SBMGame(10, badMat, []int{5, 5}, false, false, igraph.SBMOptions{}); err == nil {
			t.Errorf("expected error for negative matrix probability")
		}
		badMat2, _ := igraph.NewMatrixFromRows([][]float64{{1.5, 0.5}, {0.5, 0.5}})
		if _, err := igraph.SBMGame(10, badMat2, []int{5, 5}, false, false, igraph.SBMOptions{}); err == nil {
			t.Errorf("expected error for matrix probability > 1")
		}
		badMatNaN, _ := igraph.NewMatrixFromRows([][]float64{{math.NaN(), 0.5}, {0.5, 0.5}})
		if _, err := igraph.SBMGame(10, badMatNaN, []int{5, 5}, false, false, igraph.SBMOptions{}); err == nil {
			t.Errorf("expected error for NaN matrix element")
		}
		if _, err := igraph.SBMGame(10, mat2x2, []int{math.MaxInt, 1}, false, false, igraph.SBMOptions{}); err == nil {
			t.Errorf("expected error for blockSizes sum overflow")
		}
	})
}

func TestRewire(t *testing.T) {
	seed := uint64(42)

	t.Run("in-place rewiring preserves degree sequence", func(t *testing.T) {
		g, err := igraph.NewFull(6, false, false)
		if err != nil {
			t.Fatalf("NewFull failed: %v", err)
		}
		defer g.Close()

		degBefore, err := g.Degree(igraph.AllVertices(), igraph.DegreeOptions{Direction: igraph.DirectionAll})
		if err != nil {
			t.Fatalf("Degree failed: %v", err)
		}

		if err := g.Rewire(20, igraph.RewireSimple, igraph.RewireOptions{Seed: &seed}); err != nil {
			t.Fatalf("Rewire failed: %v", err)
		}

		degAfter, err := g.Degree(igraph.AllVertices(), igraph.DegreeOptions{Direction: igraph.DirectionAll})
		if err != nil {
			t.Fatalf("Degree failed after rewire: %v", err)
		}

		if !reflect.DeepEqual(degBefore, degAfter) {
			t.Errorf("expected degree sequence to be preserved, got %v vs %v", degAfter, degBefore)
		}
	})

	t.Run("atomic failure leaves receiver unchanged", func(t *testing.T) {
		g, err := igraph.NewFull(4, false, false)
		if err != nil {
			t.Fatalf("NewFull failed: %v", err)
		}
		defer g.Close()

		edgesBefore := mustEdges(t, g)

		if err := g.Rewire(-1, igraph.RewireSimple, igraph.RewireOptions{}); err == nil {
			t.Errorf("expected error for negative trial count")
		}

		edgesAfter := mustEdges(t, g)
		if !reflect.DeepEqual(edgesBefore, edgesAfter) {
			t.Errorf("expected graph to remain unchanged on validation failure")
		}
	})

	t.Run("invalid mode and closed graph", func(t *testing.T) {
		g, err := igraph.NewFull(4, false, false)
		if err != nil {
			t.Fatalf("NewFull failed: %v", err)
		}

		if err := g.Rewire(10, igraph.RewireMode(99), igraph.RewireOptions{}); err == nil {
			t.Errorf("expected error for invalid RewireMode")
		}

		g.Close()
		if err := g.Rewire(10, igraph.RewireSimple, igraph.RewireOptions{}); err == nil {
			t.Errorf("expected ErrClosed for closed graph")
		}
	})
}

func TestRewireEdges(t *testing.T) {
	seed := uint64(42)

	t.Run("independent rewired graph", func(t *testing.T) {
		g, err := igraph.NewFull(6, false, false)
		if err != nil {
			t.Fatalf("NewFull failed: %v", err)
		}
		defer g.Close()

		rewired, err := g.RewireEdges(0.5, false, false, igraph.RewireOptions{Seed: &seed})
		if err != nil {
			t.Fatalf("RewireEdges failed: %v", err)
		}
		defer rewired.Close()

		v1, _ := mustCounts(t, g)
		v2, _ := mustCounts(t, rewired)
		if v1 != v2 {
			t.Errorf("expected vertex count %d, got %d", v1, v2)
		}
	})

	t.Run("seed reproducibility", func(t *testing.T) {
		g, err := igraph.NewFull(8, false, false)
		if err != nil {
			t.Fatalf("NewFull failed: %v", err)
		}
		defer g.Close()

		first, err := g.RewireEdges(0.3, false, false, igraph.RewireOptions{Seed: &seed})
		if err != nil {
			t.Fatalf("RewireEdges failed: %v", err)
		}
		defer first.Close()

		second, err := g.RewireEdges(0.3, false, false, igraph.RewireOptions{Seed: &seed})
		if err != nil {
			t.Fatalf("RewireEdges failed on second run: %v", err)
		}
		defer second.Close()

		if !reflect.DeepEqual(mustEdges(t, first), mustEdges(t, second)) {
			t.Errorf("expected identical edge lists for the same seed")
		}
	})

	t.Run("invalid parameters", func(t *testing.T) {
		g, err := igraph.NewFull(4, false, false)
		if err != nil {
			t.Fatalf("NewFull failed: %v", err)
		}
		defer g.Close()

		for _, prob := range []float64{-0.1, 1.1, math.NaN()} {
			if _, err := g.RewireEdges(prob, false, false, igraph.RewireOptions{}); err == nil {
				t.Errorf("expected error for invalid probability %g", prob)
			}
		}

		closed, _ := igraph.NewFull(4, false, false)
		closed.Close()
		if _, err := closed.RewireEdges(0.5, false, false, igraph.RewireOptions{}); err == nil {
			t.Errorf("expected ErrClosed for closed graph")
		}
	})
}

func TestRandomWalk(t *testing.T) {
	seed := uint64(42)

	t.Run("unweighted and weighted walk", func(t *testing.T) {
		g, err := igraph.NewFull(6, false, false)
		if err != nil {
			t.Fatalf("NewFull failed: %v", err)
		}
		defer g.Close()

		vPath, ePath, err := g.RandomWalk(0, 5, igraph.DirectionOut, nil, igraph.RandomWalkOptions{Seed: &seed})
		if err != nil {
			t.Fatalf("RandomWalk failed: %v", err)
		}
		if len(vPath) != 6 || len(ePath) != 5 {
			t.Errorf("expected 6 vertices and 5 edges in walk path, got %d and %d", len(vPath), len(ePath))
		}

		weights := make([]float64, 6*5/2)
		for i := range weights {
			weights[i] = 1.0
		}
		vPathW, ePathW, err := g.RandomWalk(0, 5, igraph.DirectionOut, weights, igraph.RandomWalkOptions{Seed: &seed})
		if err != nil {
			t.Fatalf("RandomWalk with weights failed: %v", err)
		}
		if len(vPathW) != 6 || len(ePathW) != 5 {
			t.Errorf("expected 6 vertices and 5 edges in weighted walk path, got %d and %d", len(vPathW), len(ePathW))
		}
	})

	t.Run("stuck mode on isolated vertex", func(t *testing.T) {
		g, err := igraph.ErdosRenyiGNM(5, 0, true, false, igraph.ErdosRenyiOptions{})
		if err != nil {
			t.Fatalf("ErdosRenyiGNM failed: %v", err)
		}
		defer g.Close()

		// Stuck error mode
		_, _, err = g.RandomWalk(0, 5, igraph.DirectionOut, nil, igraph.RandomWalkOptions{
			StuckMode: igraph.RandomWalkStuckError,
		})
		if err == nil {
			t.Errorf("expected error when random walk gets stuck in StuckError mode")
		}

		// Stuck return mode returns partial path
		vPath, ePath, err := g.RandomWalk(0, 5, igraph.DirectionOut, nil, igraph.RandomWalkOptions{
			StuckMode: igraph.RandomWalkStuckReturn,
		})
		if err != nil {
			t.Fatalf("RandomWalk with StuckReturn failed: %v", err)
		}
		if len(vPath) != 1 || len(ePath) != 0 {
			t.Errorf("expected path of length 1 (start vertex only) when stuck, got %v and %v", vPath, ePath)
		}
	})

	t.Run("seed reproducibility", func(t *testing.T) {
		g, err := igraph.NewFull(10, false, false)
		if err != nil {
			t.Fatalf("NewFull failed: %v", err)
		}
		defer g.Close()

		v1, e1, err := g.RandomWalk(0, 8, igraph.DirectionOut, nil, igraph.RandomWalkOptions{Seed: &seed})
		if err != nil {
			t.Fatalf("RandomWalk failed: %v", err)
		}
		v2, e2, err := g.RandomWalk(0, 8, igraph.DirectionOut, nil, igraph.RandomWalkOptions{Seed: &seed})
		if err != nil {
			t.Fatalf("RandomWalk failed on second run: %v", err)
		}

		if !reflect.DeepEqual(v1, v2) || !reflect.DeepEqual(e1, e2) {
			t.Errorf("expected identical walk paths for the same seed")
		}
	})

	t.Run("invalid parameters", func(t *testing.T) {
		g, err := igraph.NewFull(5, false, false)
		if err != nil {
			t.Fatalf("NewFull failed: %v", err)
		}
		defer g.Close()

		if _, _, err := g.RandomWalk(-1, 5, igraph.DirectionOut, nil, igraph.RandomWalkOptions{}); err == nil {
			t.Errorf("expected error for start < 0")
		}
		if _, _, err := g.RandomWalk(10, 5, igraph.DirectionOut, nil, igraph.RandomWalkOptions{}); err == nil {
			t.Errorf("expected error for start >= vcount")
		}
		if _, _, err := g.RandomWalk(0, -1, igraph.DirectionOut, nil, igraph.RandomWalkOptions{}); err == nil {
			t.Errorf("expected error for steps < 0")
		}
		if _, _, err := g.RandomWalk(0, 5, igraph.DirectionMode(99), nil, igraph.RandomWalkOptions{}); err == nil {
			t.Errorf("expected error for invalid DirectionMode")
		}
		if _, _, err := g.RandomWalk(0, 5, igraph.DirectionOut, nil, igraph.RandomWalkOptions{StuckMode: igraph.RandomWalkStuckMode(99)}); err == nil {
			t.Errorf("expected error for invalid RandomWalkStuckMode")
		}
		if _, _, err := g.RandomWalk(0, 5, igraph.DirectionOut, []float64{1.0}, igraph.RandomWalkOptions{}); err == nil {
			t.Errorf("expected error for weights length mismatch")
		}
		if _, _, err := g.RandomWalk(0, 5, igraph.DirectionOut, []float64{-1, 1, 1, 1, 1, 1, 1, 1, 1, 1}, igraph.RandomWalkOptions{}); err == nil {
			t.Errorf("expected error for negative weight value")
		}

		closed, _ := igraph.NewFull(4, false, false)
		closed.Close()
		if _, _, err := closed.RandomWalk(0, 5, igraph.DirectionOut, nil, igraph.RandomWalkOptions{}); err == nil {
			t.Errorf("expected ErrClosed for closed graph")
		}
	})
}

func TestRandomSpanningTree(t *testing.T) {
	seed := uint64(42)

	t.Run("unweighted and root option", func(t *testing.T) {
		g, err := igraph.NewFull(6, false, false)
		if err != nil {
			t.Fatalf("NewFull failed: %v", err)
		}
		defer g.Close()

		treeEdges, err := g.RandomSpanningTree(igraph.SpanningTreeOptions{Seed: &seed})
		if err != nil {
			t.Fatalf("RandomSpanningTree failed: %v", err)
		}
		if len(treeEdges) != 5 {
			t.Errorf("expected 5 edges in spanning tree of 6 vertices, got %d", len(treeEdges))
		}

		root := 2
		treeEdgesRoot, err := g.RandomSpanningTree(igraph.SpanningTreeOptions{
			Seed: &seed,
			Root: &root,
		})
		if err != nil {
			t.Fatalf("RandomSpanningTree with Root failed: %v", err)
		}
		if len(treeEdgesRoot) != 5 {
			t.Errorf("expected 5 edges in spanning tree with Root, got %d", len(treeEdgesRoot))
		}
	})

	t.Run("seed reproducibility", func(t *testing.T) {
		g, err := igraph.NewFull(10, false, false)
		if err != nil {
			t.Fatalf("NewFull failed: %v", err)
		}
		defer g.Close()

		first, err := g.RandomSpanningTree(igraph.SpanningTreeOptions{Seed: &seed})
		if err != nil {
			t.Fatalf("RandomSpanningTree failed: %v", err)
		}
		second, err := g.RandomSpanningTree(igraph.SpanningTreeOptions{Seed: &seed})
		if err != nil {
			t.Fatalf("RandomSpanningTree failed on second run: %v", err)
		}

		if !reflect.DeepEqual(first, second) {
			t.Errorf("expected identical spanning tree edge lists for the same seed")
		}
	})

	t.Run("invalid parameters", func(t *testing.T) {
		g, err := igraph.NewFull(5, false, false)
		if err != nil {
			t.Fatalf("NewFull failed: %v", err)
		}
		defer g.Close()

		badRoot := 99
		if _, err := g.RandomSpanningTree(igraph.SpanningTreeOptions{Root: &badRoot}); err == nil {
			t.Errorf("expected error for root out of bounds")
		}

		closed, _ := igraph.NewFull(4, false, false)
		closed.Close()
		if _, err := closed.RandomSpanningTree(igraph.SpanningTreeOptions{}); err == nil {
			t.Errorf("expected ErrClosed for closed graph")
		}
	})
}
