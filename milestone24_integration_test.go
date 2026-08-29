package igraph_test

import (
	"testing"

	igraph "github.com/h8gi/go-igraph"
)

func TestMilestone24ColoringWorkflow(t *testing.T) {
	base, err := igraph.NewGraphFromEdges(2, []igraph.Edge{{From: 0, To: 1}}, false)
	if err != nil {
		t.Fatal(err)
	}
	defer base.Close()
	mycielski, err := base.Mycielskian(2)
	if err != nil {
		t.Fatal(err)
	}
	defer mycielski.Graph.Close()
	colors, err := mycielski.Graph.GreedyVertexColoring(igraph.ColoringDSatur)
	if err != nil {
		t.Fatal(err)
	}
	valid, err := mycielski.Graph.IsVertexColoring(colors)
	if err != nil || !valid {
		t.Fatalf("Mycielski coloring = %v, %v", valid, err)
	}
	clique, err := mycielski.Graph.CliqueNumber()
	if err != nil {
		t.Fatal(err)
	}
	maxColor := -1
	for _, color := range colors {
		if color > maxColor {
			maxColor = color
		}
	}
	if len(colors) > 0 && maxColor+1 < clique {
		t.Fatalf("color upper bound %d below clique lower bound %d", maxColor+1, clique)
	}
	_ = mycielski.Graph.Close()
	if len(colors) != 11 {
		t.Fatalf("owned colors after Close = %v", colors)
	}

	bipartite, err := igraph.NewFullBipartite(2, 3, true, igraph.DirectionOut)
	if err != nil {
		t.Fatal(err)
	}
	defer bipartite.Graph.Close()
	bp, err := bipartite.Graph.IsBipartiteColoring(bipartite.Partition)
	if err != nil || !bp.Valid || bp.Direction != igraph.DirectionOut {
		t.Fatalf("bipartite validation = %#v, %v", bp, err)
	}
	edgeColors := []int{0, 1, 2, 1, 2, 0}
	if valid, err := bipartite.Graph.IsEdgeColoring(edgeColors); err != nil || !valid {
		t.Fatalf("edge coloring = %v, %v", valid, err)
	}
}

func TestMilestone24DegenerateGraphs(t *testing.T) {
	for _, fixture := range []struct {
		vertices int
		edges    []igraph.Edge
		directed bool
	}{
		{0, nil, false}, {1, nil, false}, {4, []igraph.Edge{{From: 0, To: 1}}, false}, {2, []igraph.Edge{{From: 0, To: 0}, {From: 0, To: 1}, {From: 0, To: 1}}, true},
	} {
		g, err := igraph.NewGraphFromEdges(fixture.vertices, fixture.edges, fixture.directed)
		if err != nil {
			t.Fatal(err)
		}
		colors, err := g.GreedyVertexColoring(igraph.ColoringColoredNeighbors)
		if err != nil {
			_ = g.Close()
			t.Fatal(err)
		}
		valid, err := g.IsVertexColoring(colors)
		_ = g.Close()
		if err != nil || !valid {
			t.Fatalf("degenerate colors %v = %v, %v", colors, valid, err)
		}
	}
	complete, err := igraph.NewGraphFromEdges(4, []igraph.Edge{{From: 0, To: 1}, {From: 0, To: 2}, {From: 0, To: 3}, {From: 1, To: 2}, {From: 1, To: 3}, {From: 2, To: 3}}, false)
	if err != nil {
		t.Fatal(err)
	}
	defer complete.Close()
	colors, err := complete.GreedyVertexColoring(igraph.ColoringDSatur)
	if err != nil {
		t.Fatal(err)
	}
	seen := map[int]bool{}
	for _, color := range colors {
		seen[color] = true
	}
	if len(seen) != 4 {
		t.Fatalf("K4 colors = %v", colors)
	}
}
