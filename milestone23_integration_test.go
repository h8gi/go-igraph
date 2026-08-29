package igraph_test

import (
	"errors"
	"reflect"
	"sync"
	"testing"

	igraph "github.com/h8gi/go-igraph"
)

func TestMilestone23GraphAlgebraWorkflowAndOwnership(t *testing.T) {
	path := newMilestone23Graph(t, 3, []igraph.Edge{{From: 0, To: 1}, {From: 1, To: 2}}, false)
	chord := newMilestone23Graph(t, 3, []igraph.Edge{{From: 0, To: 2}}, false)

	union, err := igraph.UnionMany([]*igraph.Graph{path, chord, path}, nil)
	if err != nil {
		t.Fatal(err)
	}
	reversed, err := igraph.UnionMany([]*igraph.Graph{chord, path}, nil)
	if err != nil {
		t.Fatal(err)
	}
	disjoint, err := igraph.DisjointUnionMany([]*igraph.Graph{path, chord}, nil)
	if err != nil {
		t.Fatal(err)
	}
	intersection, err := igraph.IntersectionMany([]*igraph.Graph{path, path}, nil)
	if err != nil {
		t.Fatal(err)
	}
	join, err := path.Join(chord, nil)
	if err != nil {
		t.Fatal(err)
	}
	product, err := path.Product(chord, igraph.GraphProductCartesian)
	if err != nil {
		t.Fatal(err)
	}
	rooted, err := path.RootedProduct(chord, 0)
	if err != nil {
		t.Fatal(err)
	}
	power, err := union.Graph.GraphPower(2, false)
	if err != nil {
		t.Fatal(err)
	}
	selected, err := igraph.VertexIDs(0, 2)
	if err != nil {
		t.Fatal(err)
	}
	induced, err := power.Graph.InducedSubgraph(selected)
	if err != nil {
		t.Fatal(err)
	}
	mycielski, err := chord.Mycielskian(1)
	if err != nil {
		t.Fatal(err)
	}

	if err := path.Close(); err != nil {
		t.Fatal(err)
	}
	if err := chord.Close(); err != nil {
		t.Fatal(err)
	}
	if err := union.Graph.Close(); err != nil {
		t.Fatal(err)
	}
	if err := reversed.Graph.Close(); err != nil {
		t.Fatal(err)
	}

	assertMilestone23Shape(t, disjoint.Graph, 6, 3)
	assertMilestone23Shape(t, intersection.Graph, 3, 2)
	assertMilestone23Shape(t, join.Graph, 6, 12)
	assertMilestone23Shape(t, product.Graph, 9, 9)
	assertMilestone23Shape(t, rooted.Graph, 9, 5)
	assertMilestone23Shape(t, power.Graph, 3, 3)
	assertMilestone23Shape(t, induced.Graph, 2, 1)
	assertMilestone23Shape(t, mycielski.Graph, 7, 6)
	components, err := induced.Graph.ConnectedComponents(igraph.ConnectednessWeak)
	if err != nil || components.Count != 1 {
		t.Fatalf("induced components = %#v, %v", components, err)
	}

	for _, graph := range []*igraph.Graph{disjoint.Graph, intersection.Graph, join.Graph, product.Graph, rooted.Graph, power.Graph, induced.Graph, mycielski.Graph} {
		if err := graph.Close(); err != nil {
			t.Fatal(err)
		}
	}
	if !reflect.DeepEqual(union.Inputs[0].Vertices.OldToNew, []int{0, 1, 2}) ||
		!reflect.DeepEqual(union.Inputs[2].Vertices.OldToNew, []int{0, 1, 2}) {
		t.Fatalf("repeated operand mappings after closure = %#v", union.Inputs)
	}
	if !reflect.DeepEqual(reversed.Inputs[0].Vertices.OldToNew, []int{0, 1, 2}) {
		t.Fatalf("reversed operand mapping after closure = %#v", reversed.Inputs)
	}
	if !reflect.DeepEqual(product.Vertices[8], igraph.ProductVertexProvenance{LeftVertex: 2, RightVertex: 2}) {
		t.Fatalf("product provenance after closure = %#v", product.Vertices)
	}
	if !reflect.DeepEqual(rooted.LeftToResult[2], []int{6, 7, 8}) {
		t.Fatalf("rooted-product provenance after closure = %#v", rooted.LeftToResult)
	}
	if !reflect.DeepEqual(induced.Vertices.OldToNew, []int{0, igraph.RemovedID, 1}) {
		t.Fatalf("induced mapping after closure = %#v", induced.Vertices)
	}
	if !reflect.DeepEqual(mycielski.SourceToResult, [][]int{{0, 3}, {1, 4}, {2, 5}}) {
		t.Fatalf("Mycielski provenance after closure = %#v", mycielski.SourceToResult)
	}
}

func TestMilestone23AtomicTransformationsAndDegenerateGraphs(t *testing.T) {
	directed := newMilestone23Graph(t, 3, []igraph.Edge{{From: 0, To: 1}, {From: 1, To: 2}}, true)
	selector, err := igraph.EdgeIDs(0, 0)
	if err != nil {
		t.Fatal(err)
	}
	reversed, err := directed.ReverseEdgesInPlace(selector)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := directed.ContractVerticesInPlace([]int{7, 7, 9}, nil); err != nil {
		t.Fatal(err)
	}
	assertMilestone23Shape(t, directed, 2, 2)
	if !reflect.DeepEqual(reversed.Mapping.Edges.OldToNew, []int{0, 1}) {
		t.Fatalf("reverse mapping = %#v", reversed.Mapping)
	}
	if err := directed.Close(); err != nil {
		t.Fatal(err)
	}
	neighborhood := newMilestone23Graph(t, 4, []igraph.Edge{{From: 0, To: 1}, {From: 1, To: 2}, {From: 2, To: 3}}, false)
	closure, err := neighborhood.ConnectNeighborhoodInPlace(2, igraph.DirectionAll)
	if err != nil {
		t.Fatal(err)
	}
	assertMilestone23Shape(t, neighborhood, 4, 5)
	if !closure.EdgeMappingAvailable || !reflect.DeepEqual(closure.Mapping.Edges.OldToNew, []int{0, 1, 2}) {
		t.Fatalf("neighborhood mapping = %#v", closure)
	}
	degrees, err := neighborhood.Degree(igraph.AllVertices(), igraph.DegreeOptions{})
	if err != nil || !reflect.DeepEqual(degrees, []int{2, 3, 3, 2}) {
		t.Fatalf("neighborhood degrees = %v, %v", degrees, err)
	}
	_ = neighborhood.Close()

	empty := newMilestone23Graph(t, 0, nil, true)
	power, err := empty.GraphPower(0, true)
	if err != nil {
		t.Fatal(err)
	}
	mycielski, err := empty.Mycielskian(1)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := empty.ReverseEdgesInPlace(igraph.NoEdges()); err != nil {
		t.Fatal(err)
	}
	if _, err := empty.ContractVerticesInPlace([]int{}, nil); err != nil {
		t.Fatal(err)
	}
	assertMilestone23Shape(t, power.Graph, 0, 0)
	assertMilestone23Shape(t, mycielski.Graph, 1, 0)
	_ = power.Graph.Close()
	_ = mycielski.Graph.Close()
	_ = empty.Close()
	if _, err := empty.GraphPower(1, true); !errors.Is(err, igraph.ErrClosed) {
		t.Fatalf("GraphPower after close = %v", err)
	}
	if _, err := empty.Mycielskian(1); !errors.Is(err, igraph.ErrClosed) {
		t.Fatalf("Mycielskian after close = %v", err)
	}
}

func TestMilestone23ConcurrentRepeatedAndReversedOperands(t *testing.T) {
	left := newMilestone23Graph(t, 3, []igraph.Edge{{From: 0, To: 1}, {From: 1, To: 2}}, false)
	right := newMilestone23Graph(t, 3, []igraph.Edge{{From: 0, To: 2}}, false)
	var wait sync.WaitGroup
	errorsSeen := make(chan error, 64)
	for range 8 {
		wait.Add(1)
		go func() {
			defer wait.Done()
			result, err := igraph.UnionMany([]*igraph.Graph{left, right, left}, nil)
			if err != nil {
				errorsSeen <- err
			} else {
				_ = result.Graph.Close()
			}
			joined, err := right.Join(left, nil)
			if err != nil {
				errorsSeen <- err
			} else {
				_ = joined.Graph.Close()
			}
		}()
	}
	wait.Wait()
	close(errorsSeen)
	for err := range errorsSeen {
		t.Error(err)
	}

	wait.Add(3)
	go func() {
		defer wait.Done()
		result, err := left.Product(right, igraph.GraphProductStrong)
		if result.Graph != nil {
			_ = result.Graph.Close()
		}
		if err != nil && !errors.Is(err, igraph.ErrClosed) {
			t.Errorf("product close race = %v", err)
		}
	}()
	go func() { defer wait.Done(); _ = left.Close() }()
	go func() { defer wait.Done(); _ = right.Close() }()
	wait.Wait()
	if result, err := igraph.UnionMany([]*igraph.Graph{left, right, left}, nil); !errors.Is(err, igraph.ErrClosed) || result.Graph != nil {
		t.Fatalf("union after close = %#v, %v", result, err)
	}
}

func newMilestone23Graph(t *testing.T, vertices int, edges []igraph.Edge, directed bool) *igraph.Graph {
	t.Helper()
	graph, err := igraph.NewGraphFromEdges(vertices, edges, directed)
	if err != nil {
		t.Fatal(err)
	}
	return graph
}

func assertMilestone23Shape(t *testing.T, graph *igraph.Graph, wantVertices, wantEdges int) {
	t.Helper()
	vertices, err := graph.VertexCount()
	if err != nil {
		t.Fatal(err)
	}
	edges, err := graph.EdgeCount()
	if err != nil {
		t.Fatal(err)
	}
	if vertices != wantVertices || edges != wantEdges {
		t.Fatalf("graph shape = %d/%d, want %d/%d", vertices, edges, wantVertices, wantEdges)
	}
}
