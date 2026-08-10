package igraph_test

import (
	"math"
	"reflect"
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
