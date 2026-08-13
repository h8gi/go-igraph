package igraph_test

import (
	"testing"

	igraph "github.com/h8gi/go-igraph"
)

func TestMilestone15RoutingAndReachabilityPipeline(t *testing.T) {
	graph, err := igraph.NewGraphFromEdges(5, []igraph.Edge{
		{From: 0, To: 1}, {From: 0, To: 2}, {From: 1, To: 3},
		{From: 2, To: 3}, {From: 3, To: 4},
	}, true)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = graph.Close() })

	paths, err := graph.KShortestPaths(0, 4, 2, igraph.PathOptions{Direction: igraph.DirectionOut})
	if err != nil || len(paths) != 2 {
		t.Fatalf("bounded alternatives = %#v, %v", paths, err)
	}
	bounded, err := graph.SimplePaths(0, igraph.AllVertices(), igraph.SimplePathOptions{
		Direction: igraph.DirectionOut, MaxResults: 3,
	})
	if err != nil || len(bounded.Paths) != 3 || !bounded.Truncated {
		t.Fatalf("bounded simple paths = %#v, %v", bounded, err)
	}

	cutoff, err := graph.CutoffDistances(igraph.AllVertices(), igraph.AllVertices(), 2, igraph.PathOptions{Direction: igraph.DirectionOut})
	if err != nil || cutoff.Rows() == nil {
		t.Fatalf("cutoff distances = %#v, %v", cutoff, err)
	}
	if _, err := graph.Eccentricities(igraph.AllVertices(), igraph.PathOptions{Direction: igraph.DirectionOut}); err != nil {
		t.Fatal(err)
	}
	if _, err := graph.Radius(igraph.PathOptions{Direction: igraph.DirectionOut}); err != nil {
		t.Fatal(err)
	}
	if _, err := graph.PathLengthHistogram(true); err != nil {
		t.Fatal(err)
	}

	reachable, err := graph.Reachability(igraph.DirectionOut)
	if err != nil || len(reachable.Reachable) != 5 {
		t.Fatalf("reachability = %#v, %v", reachable, err)
	}
	closure, err := graph.TransitiveClosure()
	if err != nil {
		t.Fatal(err)
	}
	neighborhoods, err := graph.NeighborhoodGraphs(igraph.AllVertices(), igraph.NeighborhoodOptions{Order: 1, Direction: igraph.DirectionOut})
	if err != nil {
		t.Fatal(err)
	}
	if err := graph.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := closure.EdgeCount(); err != nil {
		t.Fatalf("independent closure after source close: %v", err)
	}
	_ = closure.Close()
	for _, neighborhood := range neighborhoods {
		if _, err := neighborhood.Graph.VertexCount(); err != nil {
			t.Fatalf("independent neighborhood after source close: %v", err)
		}
		_ = neighborhood.Graph.Close()
	}
}

func TestMilestone15WidestVoronoiSpannerAndEulerianPipeline(t *testing.T) {
	graph, err := igraph.NewGraphFromEdges(4, []igraph.Edge{
		{From: 0, To: 1}, {From: 1, To: 3}, {From: 0, To: 2}, {From: 2, To: 3},
	}, false)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = graph.Close() })
	if path, err := graph.WidestPath(0, 3, []float64{5, 4, 3, 3}, igraph.DirectionAll); err != nil || !path.Found {
		t.Fatalf("widest path = %#v, %v", path, err)
	}
	if result, err := graph.Voronoi([]int{0, 3}, igraph.VoronoiOptions{Direction: igraph.DirectionAll, TieBreaker: igraph.VoronoiFirst}); err != nil || len(result.Membership) != 4 {
		t.Fatalf("Voronoi = %#v, %v", result, err)
	}
	seed := uint64(15)
	spanner, err := graph.Spanner(igraph.SpannerOptions{Stretch: 3, Seed: &seed})
	if err != nil || len(spanner.SourceEdges) == 0 {
		t.Fatalf("spanner = %#v, %v", spanner, err)
	}
	_ = spanner.Graph.Close()

	cycle, err := igraph.NewGraphFromEdges(3, []igraph.Edge{{From: 0, To: 1}, {From: 1, To: 2}, {From: 2, To: 0}}, true)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = cycle.Close() })
	status, err := cycle.EulerianStatus()
	if err != nil || !status.HasCycle {
		t.Fatalf("Eulerian status = %#v, %v", status, err)
	}
	result, err := cycle.EulerianCycle()
	if err != nil || !result.Found || len(result.Vertices) != len(result.Edges)+1 {
		t.Fatalf("Eulerian alignment = %#v, %v", result, err)
	}
}
