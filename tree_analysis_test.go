package igraph

import (
	"errors"
	"math"
	"reflect"
	"testing"
)

func TestTreeAndForestAnalysis(t *testing.T) {
	tree, err := NewGraphFromEdges(4, []Edge{{0, 1}, {0, 2}, {2, 3}}, false)
	if err != nil {
		t.Fatal(err)
	}
	defer tree.Close()
	got, err := tree.IsTree(DirectionAll)
	if err != nil || !got.IsTree || got.Root < 0 {
		t.Fatalf("IsTree = %#v, %v", got, err)
	}
	forest, err := tree.IsForest(DirectionAll)
	if err != nil || !forest.IsForest || len(forest.Roots) != 1 {
		t.Fatalf("IsForest = %#v, %v", forest, err)
	}

	cycle, _ := NewGraphFromEdges(3, []Edge{{0, 1}, {1, 2}, {2, 0}}, false)
	defer cycle.Close()
	got, err = cycle.IsTree(DirectionAll)
	if err != nil || got.IsTree || got.Root != NoParent {
		t.Fatalf("cycle IsTree = %#v, %v", got, err)
	}
	forest, err = cycle.IsForest(DirectionAll)
	if err != nil || forest.IsForest || forest.Roots == nil || len(forest.Roots) != 0 {
		t.Fatalf("cycle IsForest = %#v, %v", forest, err)
	}

	directed, _ := NewGraphFromEdges(3, []Edge{{0, 1}, {0, 2}}, true)
	defer directed.Close()
	out, err := directed.IsTree(DirectionOut)
	if err != nil || !out.IsTree || out.Root != 0 {
		t.Fatalf("out tree = %#v, %v", out, err)
	}
	in, err := directed.IsTree(DirectionIn)
	if err != nil || in.IsTree {
		t.Fatalf("in tree = %#v, %v", in, err)
	}

	empty, _ := NewGraph()
	defer empty.Close()
	emptyTree, err := empty.IsTree(DirectionAll)
	if err != nil || emptyTree.IsTree || emptyTree.Root != NoParent {
		t.Fatalf("empty tree = %#v, %v", emptyTree, err)
	}
	emptyForest, err := empty.IsForest(DirectionAll)
	if err != nil || !emptyForest.IsForest || emptyForest.Roots == nil || len(emptyForest.Roots) != 0 {
		t.Fatalf("empty forest = %#v, %v", emptyForest, err)
	}
	singleton, _ := NewGraphFromEdges(1, nil, false)
	defer singleton.Close()
	singleTree, err := singleton.IsTree(DirectionAll)
	if err != nil || !singleTree.IsTree || singleTree.Root != 0 {
		t.Fatalf("singleton tree = %#v, %v", singleTree, err)
	}

	loop, _ := NewGraphFromEdges(2, []Edge{{0, 0}}, false)
	defer loop.Close()
	loopForest, err := loop.IsForest(DirectionAll)
	if err != nil || loopForest.IsForest {
		t.Fatalf("loop forest = %#v, %v", loopForest, err)
	}
}

func TestMinimumSpanningForest(t *testing.T) {
	g, _ := NewGraphFromEdges(5, []Edge{{0, 1}, {1, 2}, {0, 2}, {3, 4}}, false)
	defer g.Close()
	edges, err := g.MinimumSpanningForest([]float64{3, 2, 1, 4})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(edges, []int{2, 1, 3}) {
		t.Fatalf("weighted edges = %v", edges)
	}
	unweighted, err := g.MinimumSpanningForest(nil)
	if err != nil || len(unweighted) != 3 {
		t.Fatalf("unweighted = %v, %v", unweighted, err)
	}
	if _, err := g.MinimumSpanningForest([]float64{1}); err == nil {
		t.Fatal("wrong weight count accepted")
	}
	if _, err := g.MinimumSpanningForest([]float64{1, 2, math.NaN(), 4}); err == nil {
		t.Fatal("NaN accepted")
	}
	negative, err := g.MinimumSpanningForest([]float64{-3, -3, 1, -2})
	if err != nil || len(negative) != 3 {
		t.Fatalf("negative/tied weights = %v, %v", negative, err)
	}
	parallel, _ := NewGraphFromEdges(2, []Edge{{0, 1}, {0, 1}, {0, 0}}, true)
	defer parallel.Close()
	chosen, err := parallel.MinimumSpanningForest([]float64{2, -1, -20})
	if err != nil || !reflect.DeepEqual(chosen, []int{1}) {
		t.Fatalf("parallel MST = %v, %v", chosen, err)
	}
	edgeless, _ := NewGraphFromEdges(2, nil, false)
	defer edgeless.Close()
	none, err := edgeless.MinimumSpanningForest(nil)
	if err != nil || none == nil || len(none) != 0 {
		t.Fatalf("edgeless MST = %#v, %v", none, err)
	}
}

func TestUnfoldTree(t *testing.T) {
	g, _ := NewGraphFromEdges(4, []Edge{{0, 1}, {1, 2}, {2, 0}, {2, 3}}, false)
	defer g.Close()
	result, err := g.UnfoldTree([]int{0}, DirectionAll)
	if err != nil {
		t.Fatal(err)
	}
	defer result.Graph.Close()
	if len(result.SourceVertices) < 4 {
		t.Fatalf("mapping = %v", result.SourceVertices)
	}
	tree, err := result.Graph.IsTree(DirectionAll)
	if err != nil || !tree.IsTree {
		t.Fatalf("unfolded tree = %#v, %v", tree, err)
	}
	if _, err := g.UnfoldTree([]int{4}, DirectionAll); err == nil {
		t.Fatal("invalid root accepted")
	}
	g.Close()
	count, err := result.Graph.VertexCount()
	if err != nil || count != len(result.SourceVertices) {
		t.Fatalf("owned result after source close = %d, %v", count, err)
	}
}

func TestTreeAnalysisValidationAndClose(t *testing.T) {
	var nilGraph *Graph
	if _, err := nilGraph.IsTree(DirectionAll); !errors.Is(err, ErrClosed) {
		t.Fatalf("nil IsTree = %v", err)
	}
	if _, err := nilGraph.IsForest(DirectionAll); !errors.Is(err, ErrClosed) {
		t.Fatalf("nil IsForest = %v", err)
	}
	if _, err := nilGraph.MinimumSpanningForest(nil); !errors.Is(err, ErrClosed) {
		t.Fatalf("nil MST = %v", err)
	}
	if _, err := nilGraph.UnfoldTree(nil, DirectionAll); !errors.Is(err, ErrClosed) {
		t.Fatalf("nil unfold = %v", err)
	}
	g, _ := NewGraph()
	if _, err := g.IsTree(DirectionMode(99)); err == nil {
		t.Fatal("invalid tree direction accepted")
	}
	if _, err := g.IsForest(DirectionMode(99)); err == nil {
		t.Fatal("invalid forest direction accepted")
	}
	if _, err := g.UnfoldTree(nil, DirectionMode(99)); err == nil {
		t.Fatal("invalid unfold direction accepted")
	}
	g.Close()
	if _, err := g.IsTree(DirectionAll); !errors.Is(err, ErrClosed) {
		t.Fatalf("closed IsTree = %v", err)
	}
	if _, err := g.IsForest(DirectionAll); !errors.Is(err, ErrClosed) {
		t.Fatalf("closed IsForest = %v", err)
	}
	if _, err := g.MinimumSpanningForest(nil); !errors.Is(err, ErrClosed) {
		t.Fatalf("closed MST = %v", err)
	}
	if _, err := g.UnfoldTree(nil, DirectionAll); !errors.Is(err, ErrClosed) {
		t.Fatalf("closed unfold = %v", err)
	}
}
