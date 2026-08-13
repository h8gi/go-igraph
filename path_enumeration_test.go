package igraph

import (
	"errors"
	"reflect"
	"testing"
)

func TestShortestPathsPreserveTargetsAndOwnership(t *testing.T) {
	g := newPathTestGraph(t)
	targets, _ := VertexIDs(3, 4, 3, 0)
	paths, err := g.ShortestPaths(0, targets, PathOptions{Direction: DirectionOut})
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) != 4 {
		t.Fatalf("path count = %d, want 4", len(paths))
	}
	if !reflect.DeepEqual(paths[0].Vertices, []int{0, 2, 3}) || !paths[0].Found {
		t.Errorf("path[0] = %#v", paths[0])
	}
	if paths[1].Found || paths[1].Vertices == nil || paths[1].Edges == nil {
		t.Errorf("unreachable path = %#v", paths[1])
	}
	if !reflect.DeepEqual(paths[0], paths[2]) {
		t.Errorf("duplicate target results differ: %#v %#v", paths[0], paths[2])
	}
	if !reflect.DeepEqual(paths[3], Path{Vertices: []int{0}, Edges: []int{}, Found: true}) {
		t.Errorf("zero path = %#v", paths[3])
	}
	if err := g.Close(); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(paths[0].Vertices, []int{0, 2, 3}) {
		t.Errorf("result changed after close: %#v", paths[0])
	}
}

func TestKShortestPathsBoundedWeightedAndEmpty(t *testing.T) {
	g, err := NewGraphFromEdges(4, []Edge{{0, 1}, {1, 3}, {0, 2}, {2, 3}, {0, 3}}, true)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = g.Close() })
	paths, err := g.KShortestPaths(0, 3, 2, PathOptions{Weights: []float64{1, 1, 2, 2, 5}})
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) != 2 {
		t.Fatalf("path count = %d, want 2", len(paths))
	}
	if !reflect.DeepEqual(paths[0].Vertices, []int{0, 1, 3}) || !reflect.DeepEqual(paths[1].Vertices, []int{0, 2, 3}) {
		t.Errorf("paths = %#v", paths)
	}
	unreachable, err := g.KShortestPaths(3, 0, 3, PathOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if unreachable == nil || len(unreachable) != 0 {
		t.Errorf("unreachable = %#v", unreachable)
	}
}

func TestSimplePathsBoundsAndTruncation(t *testing.T) {
	g, err := NewGraphFromEdges(4, []Edge{{0, 1}, {1, 3}, {0, 2}, {2, 3}, {0, 3}}, true)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = g.Close() })
	targets, _ := VertexIDs(3)
	min, maxEdges := 1, 2
	result, err := g.SimplePaths(0, targets, SimplePathOptions{Direction: DirectionOut, MinEdges: &min, MaxEdges: &maxEdges, MaxResults: 2})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Paths) != 2 || !result.Truncated {
		t.Fatalf("result = %#v", result)
	}
	for _, path := range result.Paths {
		if !path.Found || len(path.Edges) != len(path.Vertices)-1 {
			t.Errorf("unaligned path = %#v", path)
		}
		assertPathEdges(t, g, path, DirectionOut)
	}
}

func TestPathEnumerationRejectsInvalidAndClosed(t *testing.T) {
	g := newPathTestGraph(t)
	targets, _ := VertexIDs(3)
	badSelector, _ := VertexIDs(5)
	if _, err := g.ShortestPaths(-1, targets, PathOptions{}); err == nil {
		t.Error("negative source error = nil")
	}
	if _, err := g.ShortestPaths(0, badSelector, PathOptions{}); err == nil {
		t.Error("bad selector error = nil")
	}
	if _, err := g.KShortestPaths(0, 3, 0, PathOptions{}); err == nil {
		t.Error("zero k error = nil")
	}
	if _, err := g.KShortestPaths(0, 3, 1, PathOptions{Weights: []float64{-1, 1, 1, 1}}); err == nil {
		t.Error("negative k-shortest weight error = nil")
	}
	if _, err := g.SimplePaths(0, targets, SimplePathOptions{}); err == nil {
		t.Error("zero limit error = nil")
	}
	negative := -1
	if _, err := g.SimplePaths(0, targets, SimplePathOptions{MinEdges: &negative, MaxResults: 1}); err == nil {
		t.Error("negative minimum error = nil")
	}
	min, maxEdges := 3, 2
	if _, err := g.SimplePaths(0, targets, SimplePathOptions{MinEdges: &min, MaxEdges: &maxEdges, MaxResults: 1}); err == nil {
		t.Error("inverted bounds error = nil")
	}
	if err := g.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := g.ShortestPaths(0, targets, PathOptions{}); !errors.Is(err, ErrClosed) {
		t.Errorf("closed error = %v", err)
	}
	if _, err := g.KShortestPaths(0, 3, 1, PathOptions{}); !errors.Is(err, ErrClosed) {
		t.Errorf("closed error = %v", err)
	}
	if _, err := g.SimplePaths(0, targets, SimplePathOptions{MaxResults: 1}); !errors.Is(err, ErrClosed) {
		t.Errorf("closed error = %v", err)
	}
}
