package igraph

import (
	"errors"
	"math"
	"reflect"
	"testing"
)

func TestWidestPathsAndWidths(t *testing.T) {
	graph := newRoutingTestGraph(t)
	weights := []float64{5, 4, 3, 3}
	path, err := graph.WidestPath(0, 3, weights, DirectionOut)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(path, Path{Vertices: []int{0, 1, 3}, Edges: []int{0, 1}, Found: true}) {
		t.Errorf("widest path = %#v", path)
	}
	targets, _ := VertexIDs(3, 2, 3)
	paths, err := graph.WidestPaths(0, targets, weights, DirectionOut)
	if err != nil || len(paths) != 3 || !reflect.DeepEqual(paths[0], paths[2]) {
		t.Errorf("widest paths = %#v, %v", paths, err)
	}
	sources, _ := VertexIDs(0, 1)
	widths, err := graph.WidestPathWidths(sources, targets, weights, DirectionOut)
	if err != nil {
		t.Fatal(err)
	}
	assertMatrixRows(t, widths, [][]float64{{4, 3, 4}, {4, math.Inf(-1), 4}})
	uniqueTargets, _ := VertexIDs(3, 2)
	if _, err := graph.WidestPathWidths(sources, uniqueTargets, weights, DirectionOut); err != nil {
		t.Errorf("unique-target widths error = %v", err)
	}
	if incoming, err := graph.WidestPath(3, 0, weights, DirectionIn); err != nil || !incoming.Found {
		t.Errorf("incoming widest path = %#v, %v", incoming, err)
	}

	unreachable, err := graph.WidestPath(3, 0, weights, DirectionOut)
	if err != nil || unreachable.Found || unreachable.Vertices == nil || unreachable.Edges == nil {
		t.Errorf("unreachable widest path = %#v, %v", unreachable, err)
	}
}

func TestVoronoiAndSpanner(t *testing.T) {
	path, err := NewGraphFromEdges(4, []Edge{{0, 1}, {1, 2}, {2, 3}}, false)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = path.Close() })
	voronoi, err := path.Voronoi([]int{0, 3}, VoronoiOptions{Direction: DirectionAll, TieBreaker: VoronoiFirst})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(voronoi.Membership, []int{0, 0, 1, 1}) || !reflect.DeepEqual(voronoi.Distances, []float64{0, 1, 1, 0}) {
		t.Errorf("Voronoi = %#v", voronoi)
	}
	seed := uint64(7)
	spanner, err := path.Spanner(SpannerOptions{Stretch: 3, Seed: &seed})
	if err != nil {
		t.Fatal(err)
	}
	if spanner.Graph == nil || spanner.SourceEdges == nil {
		t.Fatalf("spanner = %#v", spanner)
	}
	if err := path.Close(); err != nil {
		t.Fatal(err)
	}
	edges, err := spanner.Graph.EdgeCount()
	if err != nil || edges != len(spanner.SourceEdges) {
		t.Errorf("spanner edges = %d, provenance = %v, error = %v", edges, spanner.SourceEdges, err)
	}
	if err := spanner.Graph.Close(); err != nil {
		t.Fatal(err)
	}
	weighted, err := newRoutingTestGraph(t).Spanner(SpannerOptions{
		Stretch: 3, Weights: []float64{5, 4, 3, 3}, Seed: &seed,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := weighted.Graph.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestRoutingRejectsInvalidAndClosed(t *testing.T) {
	graph := newRoutingTestGraph(t)
	weights := []float64{5, 4, 3, 3}
	if _, err := graph.WidestPath(-1, 0, weights, DirectionOut); err == nil {
		t.Error("invalid widest source accepted")
	}
	if _, err := graph.KShortestPaths(-1, 0, 1, PathOptions{}); err == nil {
		t.Error("invalid k-shortest source accepted")
	}
	if _, err := graph.KShortestPaths(0, 4, 1, PathOptions{}); err == nil {
		t.Error("invalid k-shortest target accepted")
	}
	if _, err := graph.KShortestPaths(0, 3, 0, PathOptions{}); err == nil {
		t.Error("invalid k-shortest count accepted")
	}
	if _, err := graph.SimplePaths(-1, AllVertices(), SimplePathOptions{MaxResults: 1}); err == nil {
		t.Error("invalid simple-path source accepted")
	}
	if _, err := graph.WidestPath(0, 4, weights, DirectionOut); err == nil {
		t.Error("invalid widest target accepted")
	}
	if _, err := graph.WidestPath(0, -1, weights, DirectionOut); err == nil {
		t.Error("negative widest target accepted")
	}
	if _, err := graph.WidestPath(0, 3, weights, DirectionMode(99)); err == nil {
		t.Error("invalid widest direction accepted")
	}
	if _, err := graph.WidestPath(0, 3, nil, DirectionOut); err == nil {
		t.Error("nil widest weights accepted")
	}
	if _, err := graph.WidestPath(0, 3, []float64{1, 1, 1, math.NaN()}, DirectionOut); err == nil {
		t.Error("NaN widest weight accepted")
	}
	if _, err := graph.WidestPaths(0, NoVertices(), weights, DirectionOut); err != nil {
		t.Errorf("empty widest targets error = %v", err)
	}
	if emptyWidths, err := graph.WidestPathWidths(AllVertices(), NoVertices(), weights, DirectionOut); err != nil {
		t.Errorf("empty width targets error = %v", err)
	} else if rows, columns := emptyWidths.Dims(); rows != 4 || columns != 0 {
		t.Errorf("empty width target dimensions = (%d, %d)", rows, columns)
	}
	if _, err := graph.WidestPathWidths(AllVertices(), AllVertices(), []float64{1}, DirectionOut); err == nil {
		t.Error("invalid width weights accepted")
	}
	if _, err := graph.WidestPathWidths(AllVertices(), AllVertices(), weights, DirectionMode(99)); err == nil {
		t.Error("invalid width direction accepted")
	}
	invalidVertices, _ := VertexIDs(4)
	if _, err := graph.WidestPathWidths(invalidVertices, AllVertices(), weights, DirectionOut); err == nil {
		t.Error("invalid width source selector accepted")
	}
	if _, err := graph.WidestPathWidths(AllVertices(), invalidVertices, weights, DirectionOut); err == nil {
		t.Error("invalid width target selector accepted")
	}
	if _, err := graph.Voronoi([]int{4}, VoronoiOptions{}); err == nil {
		t.Error("invalid generator accepted")
	}
	if _, err := graph.Voronoi([]int{0}, VoronoiOptions{Weights: []float64{1, -1, 1, 1}}); err == nil {
		t.Error("negative Voronoi weight accepted")
	}
	if _, err := graph.Voronoi([]int{0}, VoronoiOptions{TieBreaker: VoronoiTieBreaker(99)}); err == nil {
		t.Error("invalid tie breaker accepted")
	}
	if _, err := graph.Voronoi([]int{0}, VoronoiOptions{TieBreaker: VoronoiLast}); err != nil {
		t.Errorf("last tie breaker error = %v", err)
	}
	if _, err := graph.Voronoi([]int{0}, VoronoiOptions{Direction: DirectionMode(99)}); err == nil {
		t.Error("invalid Voronoi direction accepted")
	}
	if _, err := graph.Voronoi([]int{0}, VoronoiOptions{Weights: []float64{1, 1, 1, 1}}); err != nil {
		t.Errorf("weighted Voronoi error = %v", err)
	}
	seed := uint64(3)
	if _, err := graph.Voronoi([]int{0}, VoronoiOptions{TieBreaker: VoronoiRandom, Seed: &seed}); err != nil {
		t.Errorf("random tie breaker error = %v", err)
	}
	for _, stretch := range []float64{0, math.NaN(), math.Inf(1)} {
		if _, err := graph.Spanner(SpannerOptions{Stretch: stretch}); err == nil {
			t.Errorf("invalid stretch %v accepted", stretch)
		}
	}
	if _, err := graph.Spanner(SpannerOptions{Stretch: 1, Weights: []float64{1, -1, 1, 1}}); err == nil {
		t.Error("negative spanner weight accepted")
	}
	if _, err := graph.Spanner(SpannerOptions{Stretch: 1, Weights: []float64{1}}); err == nil {
		t.Error("invalid spanner weight length accepted")
	}
	if err := graph.Close(); err != nil {
		t.Fatal(err)
	}
	var nilGraph *Graph
	for _, closed := range []*Graph{graph, nilGraph} {
		_, pathErr := closed.WidestPath(0, 0, weights, DirectionOut)
		_, widthsErr := closed.WidestPathWidths(AllVertices(), AllVertices(), weights, DirectionOut)
		_, voronoiErr := closed.Voronoi(nil, VoronoiOptions{})
		_, spannerErr := closed.Spanner(SpannerOptions{Stretch: 1})
		_, shortestErr := closed.ShortestPaths(0, AllVertices(), PathOptions{})
		_, kErr := closed.KShortestPaths(0, 0, 1, PathOptions{})
		_, simpleErr := closed.SimplePaths(0, AllVertices(), SimplePathOptions{MaxResults: 1})
		for i, err := range []error{pathErr, widthsErr, voronoiErr, spannerErr, shortestErr, kErr, simpleErr} {
			if !errors.Is(err, ErrClosed) {
				t.Errorf("closed routing check %d error = %v", i, err)
			}
		}
	}
}

func TestRoutingEmptyGraphResults(t *testing.T) {
	graph, err := NewGraph()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = graph.Close() })
	widths, err := graph.WidestPathWidths(AllVertices(), AllVertices(), []float64{}, DirectionOut)
	if err != nil {
		t.Fatal(err)
	}
	if rows, columns := widths.Dims(); rows != 0 || columns != 0 {
		t.Errorf("empty widths dimensions = (%d, %d)", rows, columns)
	}
	if paths, err := graph.WidestPaths(0, NoVertices(), []float64{}, DirectionOut); err == nil || paths != nil {
		// Empty graphs have no valid source vertex.
		if err == nil {
			t.Error("widest paths accepted a source in an empty graph")
		}
	}
	voronoi, err := graph.Voronoi(nil, VoronoiOptions{})
	if err != nil || voronoi.Membership == nil || voronoi.Distances == nil {
		t.Errorf("empty Voronoi = %#v, %v", voronoi, err)
	}
	spanner, err := graph.Spanner(SpannerOptions{Stretch: 1})
	if err != nil {
		t.Fatal(err)
	}
	if spanner.SourceEdges == nil {
		t.Error("empty spanner provenance is nil")
	}
	if err := spanner.Graph.Close(); err != nil {
		t.Fatal(err)
	}
}

func newRoutingTestGraph(t *testing.T) *Graph {
	t.Helper()
	graph, err := NewGraphFromEdges(4, []Edge{{0, 1}, {1, 3}, {0, 2}, {2, 3}}, true)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = graph.Close() })
	return graph
}
