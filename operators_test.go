package igraph

import (
	"errors"
	"reflect"
	"sync"
	"testing"
	"time"
)

func TestDisjointUnionPreservesOrderMultiplicityAndMappings(t *testing.T) {
	left := testGraphFromEdges(t, 3, []Edge{{0, 1}, {0, 1}, {2, 2}}, true)
	right := testGraphFromEdges(t, 2, []Edge{{1, 0}, {1, 1}}, true)
	result, err := left.DisjointUnion(right, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = result.Graph.Close() })
	assertGraphShape(t, result.Graph, 5, 5, true)
	assertEdgesEqual(t, result.Graph, []Edge{{0, 1}, {0, 1}, {2, 2}, {4, 3}, {4, 4}})
	wantLeft := GraphIDMapping{
		Vertices: IDMapping{OldToNew: []int{0, 1, 2}, NewToOld: []int{0, 1, 2, RemovedID, RemovedID}},
		Edges:    IDMapping{OldToNew: []int{0, 1, 2}, NewToOld: []int{0, 1, 2, RemovedID, RemovedID}},
	}
	wantRight := GraphIDMapping{
		Vertices: IDMapping{OldToNew: []int{3, 4}, NewToOld: []int{RemovedID, RemovedID, RemovedID, 0, 1}},
		Edges:    IDMapping{OldToNew: []int{3, 4}, NewToOld: []int{RemovedID, RemovedID, RemovedID, 0, 1}},
	}
	if !reflect.DeepEqual(result.Left, wantLeft) || !reflect.DeepEqual(result.Right, wantRight) {
		t.Errorf("mappings = left %#v right %#v, want left %#v right %#v", result.Left, result.Right, wantLeft, wantRight)
	}
}

func TestUnionAndIntersectionMappingsLoopsParallelAndOperandOrder(t *testing.T) {
	left := testGraphFromEdges(t, 4, []Edge{{0, 1}, {0, 1}, {1, 1}, {3, 0}}, true)
	right := testGraphFromEdges(t, 3, []Edge{{0, 1}, {1, 1}, {1, 1}, {2, 0}}, true)

	union, err := left.Union(right, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = union.Graph.Close() })
	assertGraphShape(t, union.Graph, 4, 6, true)
	assertEdgesEqual(t, union.Graph, []Edge{{0, 1}, {0, 1}, {1, 1}, {1, 1}, {2, 0}, {3, 0}})
	assertOperatorMappingConsistent(t, left, union.Graph, union.Left)
	assertOperatorMappingConsistent(t, right, union.Graph, union.Right)

	intersection, err := left.Intersection(right, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = intersection.Graph.Close() })
	assertGraphShape(t, intersection.Graph, 4, 2, true)
	assertEdgesEqual(t, intersection.Graph, []Edge{{0, 1}, {1, 1}})
	assertOperatorMappingConsistent(t, left, intersection.Graph, intersection.Left)
	assertOperatorMappingConsistent(t, right, intersection.Graph, intersection.Right)
	if !reflect.DeepEqual(intersection.Left.Edges.NewToOld, []int{0, 2}) ||
		!reflect.DeepEqual(intersection.Right.Edges.NewToOld, []int{0, 1}) {
		t.Errorf("intersection exact inverse maps = %v / %v", intersection.Left.Edges.NewToOld, intersection.Right.Edges.NewToOld)
	}

	reversed, err := right.Union(left, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = reversed.Graph.Close() })
	assertEdgesEqual(t, reversed.Graph, []Edge{{0, 1}, {0, 1}, {1, 1}, {1, 1}, {2, 0}, {3, 0}})
	assertOperatorMappingConsistent(t, right, reversed.Graph, reversed.Left)
	assertOperatorMappingConsistent(t, left, reversed.Graph, reversed.Right)
}

func TestManyUnionAndIntersectionMappingsRepeatedOperands(t *testing.T) {
	first := testGraphFromEdges(t, 4, []Edge{{0, 1}, {0, 1}, {1, 1}, {3, 0}}, true)
	second := testGraphFromEdges(t, 3, []Edge{{0, 1}, {1, 1}, {1, 1}, {2, 0}}, true)
	defer first.Close()
	defer second.Close()

	union, err := UnionMany([]*Graph{first, second, first}, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer union.Graph.Close()
	if len(union.Inputs) != 3 {
		t.Fatalf("union mapping count = %d, want 3", len(union.Inputs))
	}
	assertGraphShape(t, union.Graph, 4, 6, true)
	assertEdgesEqual(t, union.Graph, []Edge{{3, 0}, {2, 0}, {1, 1}, {1, 1}, {0, 1}, {0, 1}})
	for index, source := range []*Graph{first, second, first} {
		assertOperatorMappingConsistent(t, source, union.Graph, union.Inputs[index])
	}
	if !reflect.DeepEqual(union.Inputs[0], union.Inputs[2]) {
		t.Fatalf("repeated-input mappings differ: %#v / %#v", union.Inputs[0], union.Inputs[2])
	}

	intersection, err := IntersectionMany([]*Graph{first, second, first}, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer intersection.Graph.Close()
	assertGraphShape(t, intersection.Graph, 4, 2, true)
	assertEdgesEqual(t, intersection.Graph, []Edge{{1, 1}, {0, 1}})
	for index, source := range []*Graph{first, second, first} {
		assertOperatorMappingConsistent(t, source, intersection.Graph, intersection.Inputs[index])
	}
	if !reflect.DeepEqual(intersection.Inputs[0], intersection.Inputs[2]) {
		t.Fatalf("repeated-input intersection mappings differ: %#v / %#v", intersection.Inputs[0], intersection.Inputs[2])
	}
}

func TestDisjointUnionManyOrderMappingsAndEmpty(t *testing.T) {
	first := testGraphFromEdges(t, 2, []Edge{{0, 1}}, false)
	second := testGraphFromEdges(t, 1, []Edge{{0, 0}}, false)
	defer first.Close()
	defer second.Close()

	result, err := DisjointUnionMany([]*Graph{first, second, first}, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer result.Graph.Close()
	assertGraphShape(t, result.Graph, 5, 3, false)
	assertEdgesEqual(t, result.Graph, []Edge{{0, 1}, {2, 2}, {3, 4}})
	want := []GraphIDMapping{
		{Vertices: IDMapping{OldToNew: []int{0, 1}, NewToOld: []int{0, 1, -1, -1, -1}}, Edges: IDMapping{OldToNew: []int{0}, NewToOld: []int{0, -1, -1}}},
		{Vertices: IDMapping{OldToNew: []int{2}, NewToOld: []int{-1, -1, 0, -1, -1}}, Edges: IDMapping{OldToNew: []int{1}, NewToOld: []int{-1, 0, -1}}},
		{Vertices: IDMapping{OldToNew: []int{3, 4}, NewToOld: []int{-1, -1, -1, 0, 1}}, Edges: IDMapping{OldToNew: []int{2}, NewToOld: []int{-1, -1, 0}}},
	}
	if !reflect.DeepEqual(result.Inputs, want) {
		t.Fatalf("disjoint mappings = %#v, want %#v", result.Inputs, want)
	}

	for name, call := range map[string]func([]*Graph, *GraphOperatorAttributePolicy) (ManyGraphOperatorResult, error){
		"disjoint":     DisjointUnionMany,
		"union":        UnionMany,
		"intersection": IntersectionMany,
	} {
		empty, err := call(nil, nil)
		if err != nil {
			t.Fatalf("empty %s error = %v", name, err)
		}
		if empty.Inputs == nil || len(empty.Inputs) != 0 {
			t.Errorf("empty %s mappings = %#v, want non-nil empty", name, empty.Inputs)
		}
		assertGraphShape(t, empty.Graph, 0, 0, true)
		_ = empty.Graph.Close()
	}
}

func TestManyOperatorsAttributesOwnershipAndValidation(t *testing.T) {
	first := testGraphFromEdges(t, 2, []Edge{{0, 1}}, true)
	second := testGraphFromEdges(t, 2, []Edge{{0, 1}}, true)
	for index, graph := range []*Graph{first, second} {
		base := float64(index + 1)
		if err := graph.SetGraphNumericAttribute("score", base); err != nil {
			t.Fatal(err)
		}
		if err := graph.SetVertexNumericAttributes("score", []float64{base, base + 2}); err != nil {
			t.Fatal(err)
		}
		if err := graph.SetEdgeNumericAttributes("score", []float64{base}); err != nil {
			t.Fatal(err)
		}
	}
	if result, err := UnionMany([]*Graph{first, second}, nil); err == nil || result.Graph != nil {
		t.Fatalf("UnionMany without conflict policy = %#v, %v", result, err)
	}
	policy := &GraphOperatorAttributePolicy{
		Graph:    AttributeCombinationPolicy{Default: AttributeCombineSum},
		Vertices: AttributeCombinationPolicy{Default: AttributeCombineSum},
		Edges:    AttributeCombinationPolicy{Default: AttributeCombineSum},
	}
	result, err := UnionMany([]*Graph{first, second}, policy)
	if err != nil {
		t.Fatal(err)
	}
	if got, _ := result.Graph.GraphNumericAttribute("score"); got != 3 {
		t.Errorf("graph score = %v, want 3", got)
	}
	if got, _ := result.Graph.VertexNumericAttributes("score"); !reflect.DeepEqual(got, []float64{3, 7}) {
		t.Errorf("vertex scores = %v, want [3 7]", got)
	}
	if got, _ := result.Graph.EdgeNumericAttributes("score"); !reflect.DeepEqual(got, []float64{3}) {
		t.Errorf("edge scores = %v, want [3]", got)
	}
	if err := first.Close(); err != nil {
		t.Fatal(err)
	}
	if err := second.Close(); err != nil {
		t.Fatal(err)
	}
	if got, _ := result.Graph.EdgeNumericAttributes("score"); !reflect.DeepEqual(got, []float64{3}) {
		t.Errorf("result attribute after source closure = %v", got)
	}
	if err := result.Graph.Close(); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(result.Inputs[0].Edges.OldToNew, []int{0}) {
		t.Errorf("mapping after all graph closure = %#v", result.Inputs)
	}

	open := testGraphFromEdges(t, 1, nil, false)
	defer open.Close()
	closed := testGraphFromEdges(t, 1, nil, false)
	_ = closed.Close()
	var nilGraph *Graph
	for name, call := range map[string]func([]*Graph, *GraphOperatorAttributePolicy) (ManyGraphOperatorResult, error){
		"disjoint":     DisjointUnionMany,
		"union":        UnionMany,
		"intersection": IntersectionMany,
	} {
		if got, err := call([]*Graph{open, nilGraph}, nil); !errors.Is(err, ErrClosed) || got.Graph != nil {
			t.Errorf("%s nil input = %#v, %v", name, got, err)
		}
		if got, err := call([]*Graph{open, closed}, nil); !errors.Is(err, ErrClosed) || got.Graph != nil {
			t.Errorf("%s closed input = %#v, %v", name, got, err)
		}
	}
	directed := testGraphFromEdges(t, 1, nil, true)
	defer directed.Close()
	if got, err := UnionMany([]*Graph{open, directed}, nil); err == nil || got.Graph != nil {
		t.Errorf("mixed directedness UnionMany = %#v, %v", got, err)
	}
}

func TestDisjointUnionManyConcatenatesElementAttributes(t *testing.T) {
	first := testGraphFromEdges(t, 2, []Edge{{0, 1}}, true)
	second := testGraphFromEdges(t, 1, []Edge{{0, 0}}, true)
	defer first.Close()
	defer second.Close()
	if err := first.SetGraphStringAttribute("name", "first"); err != nil {
		t.Fatal(err)
	}
	if err := second.SetGraphStringAttribute("name", "second"); err != nil {
		t.Fatal(err)
	}
	if err := first.SetVertexStringAttributes("label", []string{"a", "b"}); err != nil {
		t.Fatal(err)
	}
	if err := second.SetVertexStringAttributes("label", []string{"c"}); err != nil {
		t.Fatal(err)
	}
	if err := first.SetEdgeBooleanAttributes("active", []bool{true}); err != nil {
		t.Fatal(err)
	}
	if err := second.SetEdgeBooleanAttributes("active", []bool{false}); err != nil {
		t.Fatal(err)
	}
	result, err := DisjointUnionMany([]*Graph{first, second}, &GraphOperatorAttributePolicy{
		Graph: AttributeCombinationPolicy{Default: AttributeCombineConcat},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer result.Graph.Close()
	if got, _ := result.Graph.GraphStringAttribute("name"); got != "firstsecond" {
		t.Errorf("graph name = %q, want firstsecond", got)
	}
	if got, _ := result.Graph.VertexStringAttributes("label"); !reflect.DeepEqual(got, []string{"a", "b", "c"}) {
		t.Errorf("vertex labels = %v", got)
	}
	if got, _ := result.Graph.EdgeBooleanAttributes("active"); !reflect.DeepEqual(got, []bool{true, false}) {
		t.Errorf("edge active = %v", got)
	}
}

func TestManyOperatorInternalValidation(t *testing.T) {
	graph := testGraphFromEdges(t, 1, nil, true)
	defer graph.Close()
	if result, err := manyGraphOperator([]*Graph{graph}, manyOperatorMode(255), nil); err == nil || result.Graph != nil {
		t.Errorf("invalid many operator = %#v, %v", result, err)
	}
	if err := restoreManyOperatorAttributes(graph, []graphAttributeSnapshot{{}}, nil, nil); err == nil {
		t.Error("misaligned attribute snapshots and mappings were accepted")
	}
}

func TestManyOperatorLockOrderingHandlesReversedAndRepeatedOperands(t *testing.T) {
	left := testGraphFromEdges(t, 3, []Edge{{0, 1}, {1, 2}}, true)
	right := testGraphFromEdges(t, 3, []Edge{{0, 2}, {2, 1}}, true)
	defer left.Close()
	defer right.Close()
	var wait sync.WaitGroup
	errorsCh := make(chan error, 40)
	for range 20 {
		wait.Add(2)
		go func() {
			defer wait.Done()
			result, err := UnionMany([]*Graph{left, right, left}, nil)
			if result.Graph != nil {
				_ = result.Graph.Close()
			}
			errorsCh <- err
		}()
		go func() {
			defer wait.Done()
			result, err := IntersectionMany([]*Graph{right, left, right}, nil)
			if result.Graph != nil {
				_ = result.Graph.Close()
			}
			errorsCh <- err
		}()
	}
	done := make(chan struct{})
	go func() { wait.Wait(); close(done) }()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("reversed/repeated many-operator calls deadlocked")
	}
	close(errorsCh)
	for err := range errorsCh {
		if err != nil {
			t.Errorf("concurrent many-operator error = %v", err)
		}
	}
}

func TestManyOperatorsConcurrentUseAndClose(t *testing.T) {
	first := testGraphFromEdges(t, 3, []Edge{{0, 1}, {1, 2}}, true)
	second := testGraphFromEdges(t, 3, []Edge{{0, 2}, {2, 1}}, true)
	var wait sync.WaitGroup
	for range 8 {
		wait.Add(1)
		go func() {
			defer wait.Done()
			result, err := UnionMany([]*Graph{first, second, first}, nil)
			if result.Graph != nil {
				_ = result.Graph.Close()
			}
			if err != nil && !errors.Is(err, ErrClosed) {
				t.Errorf("UnionMany close race error = %v", err)
			}
		}()
	}
	wait.Add(1)
	go func() { defer wait.Done(); _ = second.Close() }()
	wait.Wait()
	_ = first.Close()
}

func TestDifferenceIsOrderedAndComplementOptions(t *testing.T) {
	left := testGraphFromEdges(t, 3, []Edge{{0, 1}, {0, 1}, {1, 1}, {2, 0}}, true)
	right := testGraphFromEdges(t, 3, []Edge{{0, 1}, {1, 1}, {1, 1}}, true)
	difference, err := left.Difference(right)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = difference.Graph.Close() })
	assertGraphShape(t, difference.Graph, 3, 2, true)
	assertEdgesEqual(t, difference.Graph, []Edge{{0, 1}, {2, 0}})
	wantIdentity, _ := identityIDMapping(3)
	if !reflect.DeepEqual(difference.Left.Vertices, wantIdentity) {
		t.Errorf("difference vertices = %#v, want %#v", difference.Left.Vertices, wantIdentity)
	}
	wantEdges := IDMapping{
		OldToNew: []int{0, RemovedID, RemovedID, 1},
		NewToOld: []int{0, 3},
	}
	if !reflect.DeepEqual(difference.Left.Edges, wantEdges) {
		t.Errorf("difference edges = %#v, want %#v", difference.Left.Edges, wantEdges)
	}
	reversed, err := right.Difference(left)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = reversed.Graph.Close() })
	assertGraphShape(t, reversed.Graph, 3, 1, true)
	assertEdgesEqual(t, reversed.Graph, []Edge{{1, 1}})
	wantReversedEdges := IDMapping{
		OldToNew: []int{RemovedID, 0, RemovedID},
		NewToOld: []int{1},
	}
	if !reflect.DeepEqual(reversed.Left.Edges, wantReversedEdges) {
		t.Errorf("reversed difference edges = %#v, want %#v", reversed.Left.Edges, wantReversedEdges)
	}

	undirected := testGraphFromEdges(t, 3, []Edge{{0, 1}, {2, 2}}, false)
	withoutLoops, err := undirected.Complement(false)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = withoutLoops.Graph.Close() })
	assertEdgesEqual(t, withoutLoops.Graph, []Edge{{0, 2}, {1, 2}})
	withLoops, err := undirected.Complement(true)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = withLoops.Graph.Close() })
	assertEdgesEqual(t, withLoops.Graph, []Edge{{0, 2}, {0, 0}, {1, 2}, {1, 1}})

	directedParallel := testGraphFromEdges(t, 2, []Edge{{0, 1}, {0, 1}}, true)
	if result, err := directedParallel.Complement(true); err == nil || result.Graph != nil {
		t.Errorf("parallel Complement() = %#v, %v, want zero, error", result, err)
	}
	directed := testGraphFromEdges(t, 2, []Edge{{0, 1}}, true)
	directedComplement, err := directed.Complement(true)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = directedComplement.Graph.Close() })
	assertEdgesEqual(t, directedComplement.Graph, []Edge{{0, 0}, {1, 1}, {1, 0}})
}

func TestComposeReturnsPerResultEdgeProvenance(t *testing.T) {
	left := testGraphFromEdges(t, 4, []Edge{{0, 1}, {0, 2}, {0, 2}, {2, 2}}, true)
	right := testGraphFromEdges(t, 4, []Edge{{1, 3}, {2, 3}, {2, 3}, {2, 2}}, true)
	result, err := left.Compose(right, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = result.Graph.Close() })
	assertGraphShape(t, result.Graph, 4, 10, true)
	if len(result.Edges) != 10 {
		t.Fatalf("composition provenance length = %d, want 10", len(result.Edges))
	}
	resultEdges, err := result.Graph.Edges()
	if err != nil {
		t.Fatal(err)
	}
	leftEdges, _ := left.Edges()
	rightEdges, _ := right.Edges()
	for edgeID, provenance := range result.Edges {
		leftEdge := leftEdges[provenance.LeftEdge]
		rightEdge := rightEdges[provenance.RightEdge]
		if leftEdge.To != rightEdge.From {
			t.Errorf("edge %d provenance does not join: %#v then %#v", edgeID, leftEdge, rightEdge)
		}
		if want := (Edge{From: leftEdge.From, To: rightEdge.To}); resultEdges[edgeID] != want {
			t.Errorf("edge %d = %#v, provenance produces %#v", edgeID, resultEdges[edgeID], want)
		}
	}
	wantIdentity, _ := identityIDMapping(4)
	if !reflect.DeepEqual(result.LeftVertices, wantIdentity) || !reflect.DeepEqual(result.RightVertices, wantIdentity) {
		t.Errorf("composition vertex mappings = %#v / %#v", result.LeftVertices, result.RightVertices)
	}
}

func TestOperatorsEmptySelfOwnershipAndClosed(t *testing.T) {
	empty := testGraphFromEdges(t, 0, nil, false)
	self := testGraphFromEdges(t, 2, []Edge{{0, 1}, {0, 1}, {1, 1}}, false)

	disjoint, err := empty.DisjointUnion(empty, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = disjoint.Graph.Close() })
	assertGraphShape(t, disjoint.Graph, 0, 0, false)
	assertNonNilBinaryMappings(t, disjoint)
	for name, call := range map[string]func() (*Graph, error){
		"union": func() (*Graph, error) {
			result, err := empty.Union(empty, nil)
			return result.Graph, err
		},
		"intersection": func() (*Graph, error) {
			result, err := empty.Intersection(empty, nil)
			return result.Graph, err
		},
		"difference": func() (*Graph, error) {
			result, err := empty.Difference(empty)
			return result.Graph, err
		},
		"composition": func() (*Graph, error) {
			result, err := empty.Compose(empty, nil)
			if result.Edges == nil {
				t.Error("empty composition provenance is nil")
			}
			return result.Graph, err
		},
		"complement": func() (*Graph, error) {
			result, err := empty.Complement(true)
			return result.Graph, err
		},
	} {
		graph, err := call()
		if err != nil {
			t.Errorf("empty %s error = %v", name, err)
			continue
		}
		t.Cleanup(func() { _ = graph.Close() })
		assertGraphShape(t, graph, 0, 0, false)
	}

	union, err := self.Union(self, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = union.Graph.Close() })
	assertGraphShape(t, union.Graph, 2, 3, false)
	intersection, err := self.Intersection(self, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = intersection.Graph.Close() })
	assertGraphShape(t, intersection.Graph, 2, 3, false)
	difference, err := self.Difference(self)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = difference.Graph.Close() })
	assertGraphShape(t, difference.Graph, 2, 0, false)
	composition, err := self.Compose(self, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = composition.Graph.Close() })
	if composition.Edges == nil {
		t.Fatal("composition provenance is nil")
	}

	if err := self.Close(); err != nil {
		t.Fatal(err)
	}
	assertGraphShape(t, union.Graph, 2, 3, false)
	union.Left.Edges.OldToNew[0] = 99
	for name, call := range map[string]func() error{
		"disjoint":   func() error { _, err := self.DisjointUnion(empty, nil); return err },
		"union":      func() error { _, err := self.Union(empty, nil); return err },
		"intersect":  func() error { _, err := self.Intersection(empty, nil); return err },
		"difference": func() error { _, err := self.Difference(empty); return err },
		"compose":    func() error { _, err := self.Compose(empty, nil); return err },
		"complement": func() error { _, err := self.Complement(false); return err },
	} {
		if err := call(); !errors.Is(err, ErrClosed) {
			t.Errorf("closed %s error = %v, want %v", name, err, ErrClosed)
		}
	}
	var nilGraph *Graph
	if _, err := nilGraph.Union(empty, nil); !errors.Is(err, ErrClosed) {
		t.Errorf("nil Union error = %v", err)
	}
}

func TestComposeDifferentVertexCounts(t *testing.T) {
	left := testGraphFromEdges(t, 2, []Edge{{0, 1}}, true)
	right := testGraphFromEdges(t, 3, []Edge{{1, 2}}, true)
	result, err := left.Compose(right, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = result.Graph.Close() })
	assertGraphShape(t, result.Graph, 3, 1, true)
	assertEdgesEqual(t, result.Graph, []Edge{{0, 2}})
	wantLeft := IDMapping{OldToNew: []int{0, 1}, NewToOld: []int{0, 1, RemovedID}}
	wantRight, _ := identityIDMapping(3)
	if !reflect.DeepEqual(result.LeftVertices, wantLeft) || !reflect.DeepEqual(result.RightVertices, wantRight) {
		t.Errorf("composition mappings = %#v / %#v", result.LeftVertices, result.RightVertices)
	}
}

func TestComposeUndirectedLoopsPreservesExactProvenance(t *testing.T) {
	left := testGraphFromEdges(t, 3, []Edge{{0, 1}, {1, 1}}, false)
	right := testGraphFromEdges(t, 3, []Edge{{1, 2}, {1, 1}}, false)
	result, err := left.Compose(right, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = result.Graph.Close() })
	resultEdges, _ := result.Graph.Edges()
	leftEdges, _ := left.Edges()
	rightEdges, _ := right.Edges()
	if len(resultEdges) == 0 || len(result.Edges) != len(resultEdges) {
		t.Fatalf("undirected result/provenance lengths = %d/%d", len(resultEdges), len(result.Edges))
	}
	usedLoop := false
	for edgeID, provenance := range result.Edges {
		leftEdge := leftEdges[provenance.LeftEdge]
		rightEdge := rightEdges[provenance.RightEdge]
		if leftEdge.From == leftEdge.To || rightEdge.From == rightEdge.To {
			usedLoop = true
		}
		if !undirectedCompositionMatches(leftEdge, rightEdge, resultEdges[edgeID]) {
			t.Errorf("edge %d = %#v is inconsistent with provenance %#v then %#v", edgeID, resultEdges[edgeID], leftEdge, rightEdge)
		}
	}
	if !usedLoop {
		t.Error("undirected composition provenance did not exercise a loop")
	}
}

func TestOperatorsRejectMixedDirectednessWithoutResult(t *testing.T) {
	directed := testGraphFromEdges(t, 2, []Edge{{0, 1}}, true)
	undirected := testGraphFromEdges(t, 2, []Edge{{0, 1}}, false)
	if result, err := directed.DisjointUnion(undirected, nil); err == nil || result.Graph != nil {
		t.Errorf("mixed DisjointUnion = %#v, %v", result, err)
	}
	if result, err := directed.Union(undirected, nil); err == nil || result.Graph != nil {
		t.Errorf("mixed Union = %#v, %v", result, err)
	}
	if result, err := directed.Intersection(undirected, nil); err == nil || result.Graph != nil {
		t.Errorf("mixed Intersection = %#v, %v", result, err)
	}
	if result, err := directed.Difference(undirected); err == nil || result.Graph != nil {
		t.Errorf("mixed Difference = %#v, %v", result, err)
	}
	if result, err := directed.Compose(undirected, nil); err == nil || result.Graph != nil {
		t.Errorf("mixed Compose = %#v, %v", result, err)
	}
}

func TestComplementParallelEdgesRejectionAndLoops(t *testing.T) {
	multigraph := testGraphFromEdges(t, 2, []Edge{{0, 1}, {0, 1}}, false)
	if res, err := multigraph.Complement(false); err == nil || res.Graph != nil {
		t.Errorf("Complement on multigraph = %#v, %v, want error", res, err)
	}

	simple := testGraphFromEdges(t, 3, []Edge{{0, 1}}, false)
	resLoops, err := simple.Complement(true)
	if err != nil {
		t.Fatalf("Complement(true) error = %v", err)
	}
	_ = resLoops.Graph.Close()

	resNoLoops, err := simple.Complement(false)
	if err != nil {
		t.Fatalf("Complement(false) error = %v", err)
	}
	_ = resNoLoops.Graph.Close()
}

func TestOperatorLockOrderingHandlesReversedAndRepeatedOperands(t *testing.T) {
	left := testGraphFromEdges(t, 3, []Edge{{0, 1}, {1, 2}}, true)
	right := testGraphFromEdges(t, 3, []Edge{{0, 2}, {2, 1}}, true)
	for _, graph := range []*Graph{left, right} {
		if err := graph.SetVertexNumericAttributes("score", []float64{1, 2, 3}); err != nil {
			t.Fatal(err)
		}
		if err := graph.SetEdgeNumericAttributes("weight", []float64{4, 5}); err != nil {
			t.Fatal(err)
		}
	}
	policy := &GraphOperatorAttributePolicy{
		Vertices: AttributeCombinationPolicy{Default: AttributeCombineFirst},
		Edges:    AttributeCombinationPolicy{Default: AttributeCombineSum},
	}
	var wait sync.WaitGroup
	errorsCh := make(chan error, 60)
	for index := 0; index < 20; index++ {
		wait.Add(3)
		go func() {
			defer wait.Done()
			result, err := left.Union(right, policy)
			if result.Graph != nil {
				_ = result.Graph.Close()
			}
			errorsCh <- err
		}()
		go func() {
			defer wait.Done()
			result, err := right.Intersection(left, policy)
			if result.Graph != nil {
				_ = result.Graph.Close()
			}
			errorsCh <- err
		}()
		go func() {
			defer wait.Done()
			result, err := left.Compose(left, policy)
			if result.Graph != nil {
				_ = result.Graph.Close()
			}
			errorsCh <- err
		}()
	}
	done := make(chan struct{})
	go func() { wait.Wait(); close(done) }()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("reversed/repeated operator calls deadlocked")
	}
	close(errorsCh)
	for err := range errorsCh {
		if err != nil {
			t.Errorf("concurrent operator error = %v", err)
		}
	}
}

func TestCollectMappedMergeCleansInitializationQueryAndConversionFailures(t *testing.T) {
	forced := errors.New("forced failure")
	for _, stage := range []string{
		"first init", "second init", "query", "query with graph", "nil graph",
		"left conversion", "right conversion", "mapping length", "mapping value",
	} {
		t.Run(stage, func(t *testing.T) {
			initCalls := 0
			sliceCalls := 0
			closedVectors := 0
			closedGraphs := 0
			operations := mappedMergeOperations{
				newVector: func() (*intVector, error) {
					initCalls++
					if (stage == "first init" && initCalls == 1) || (stage == "second init" && initCalls == 2) {
						return nil, forced
					}
					return &intVector{}, nil
				},
				closeVector: func(*intVector) { closedVectors++ },
				query: func(*intVector, *intVector) (*Graph, int, int, error) {
					if stage == "query" {
						return nil, 0, 0, forced
					}
					if stage == "query with graph" {
						return &Graph{}, 1, 1, forced
					}
					if stage == "nil graph" {
						return nil, 0, 0, nil
					}
					return &Graph{}, 1, 1, nil
				},
				vectorSlice: func(*intVector) ([]int, error) {
					sliceCalls++
					if (stage == "left conversion" && sliceCalls == 1) || (stage == "right conversion" && sliceCalls == 2) {
						return nil, forced
					}
					if stage == "mapping length" && sliceCalls == 1 {
						return []int{}, nil
					}
					if stage == "mapping value" && sliceCalls == 1 {
						return []int{2}, nil
					}
					return []int{0}, nil
				},
				closeGraph: func(*Graph) { closedGraphs++ },
			}
			result, err := collectMappedMerge(1, 1, 1, 1, mappedMergeUnion, operations)
			if result.Graph != nil || err == nil {
				t.Errorf("collectMappedMerge() = %#v, %v, want zero, error", result, err)
			}
			wantVectors := 2
			if stage == "first init" {
				wantVectors = 0
			} else if stage == "second init" {
				wantVectors = 1
			}
			wantGraphs := 1
			if stage == "first init" || stage == "second init" || stage == "query" || stage == "nil graph" {
				wantGraphs = 0
			}
			if closedVectors != wantVectors || closedGraphs != wantGraphs {
				t.Errorf("closed vectors/graphs = %d/%d, want %d/%d", closedVectors, closedGraphs, wantVectors, wantGraphs)
			}
		})
	}

	closed := 0
	result, err := collectMappedMerge(1, 1, 1, 1, mappedMergeIntersection, mappedMergeOperations{
		newVector:   func() (*intVector, error) { return &intVector{}, nil },
		closeVector: func(*intVector) {},
		query:       func(*intVector, *intVector) (*Graph, int, int, error) { return &Graph{}, 1, 1, nil },
		vectorSlice: func(*intVector) ([]int, error) { return []int{0}, nil },
		closeGraph:  func(*Graph) { closed++ },
	})
	if err != nil || result.Graph == nil || closed != 0 {
		t.Errorf("intersection collector success = %#v, %v, closed=%d", result, err, closed)
	}
	result, err = collectMappedMerge(1, 1, 1, 1, mappedMergeIntersection, mappedMergeOperations{
		newVector:   func() (*intVector, error) { return &intVector{}, nil },
		closeVector: func(*intVector) {},
		query:       func(*intVector, *intVector) (*Graph, int, int, error) { return &Graph{}, 1, 1, nil },
		vectorSlice: func(*intVector) ([]int, error) { return []int{}, nil },
		closeGraph:  func(*Graph) { closed++ },
	})
	if result.Graph != nil || err == nil || closed != 1 {
		t.Errorf("intersection length failure = %#v, %v, closed=%d", result, err, closed)
	}
}

func TestCollectCompositionCleansFailuresAndValidatesProvenance(t *testing.T) {
	forced := errors.New("forced failure")
	for _, stage := range []string{
		"first init", "second init", "query", "query with graph", "nil graph", "left conversion",
		"right conversion", "length", "left provenance", "right provenance",
	} {
		t.Run(stage, func(t *testing.T) {
			initCalls := 0
			sliceCalls := 0
			closedVectors := 0
			closedGraphs := 0
			operations := compositionOperations{
				newVector: func() (*intVector, error) {
					initCalls++
					if (stage == "first init" && initCalls == 1) || (stage == "second init" && initCalls == 2) {
						return nil, forced
					}
					return &intVector{}, nil
				},
				closeVector: func(*intVector) { closedVectors++ },
				query: func(*intVector, *intVector) (*Graph, int, int, error) {
					if stage == "query" {
						return nil, 0, 0, forced
					}
					if stage == "query with graph" {
						return &Graph{}, 1, 1, forced
					}
					if stage == "nil graph" {
						return nil, 0, 0, nil
					}
					return &Graph{}, 1, 1, nil
				},
				vectorSlice: func(*intVector) ([]int, error) {
					sliceCalls++
					if (stage == "left conversion" && sliceCalls == 1) || (stage == "right conversion" && sliceCalls == 2) {
						return nil, forced
					}
					if stage == "length" && sliceCalls == 1 {
						return []int{}, nil
					}
					if stage == "left provenance" && sliceCalls == 1 {
						return []int{1}, nil
					}
					if stage == "right provenance" && sliceCalls == 2 {
						return []int{1}, nil
					}
					return []int{0}, nil
				},
				closeGraph: func(*Graph) { closedGraphs++ },
			}
			result, err := collectComposition(1, 1, 1, 1, operations)
			if result.Graph != nil || err == nil {
				t.Errorf("collectComposition() = %#v, %v, want zero, error", result, err)
			}
			wantVectors := 2
			if stage == "first init" {
				wantVectors = 0
			} else if stage == "second init" {
				wantVectors = 1
			}
			wantGraphs := 1
			if stage == "first init" || stage == "second init" || stage == "query" || stage == "nil graph" {
				wantGraphs = 0
			}
			if closedVectors != wantVectors || closedGraphs != wantGraphs {
				t.Errorf("closed vectors/graphs = %d/%d, want %d/%d", closedVectors, closedGraphs, wantVectors, wantGraphs)
			}
		})
	}
}

func TestOperatorCollectionHelpersRejectNilAndCloseFailures(t *testing.T) {
	if graph, err := collectOwnedOperatorGraph(func() (*Graph, error) { return nil, nil }); graph != nil || err == nil {
		t.Errorf("nil collectOwnedOperatorGraph() = %v, %v", graph, err)
	}
	forced := errors.New("forced failure")
	graph := testGraphFromEdges(t, 1, nil, false)
	if result, err := collectOwnedOperatorGraph(func() (*Graph, error) { return graph, forced }); result != nil || !errors.Is(err, forced) {
		t.Errorf("failed collectOwnedOperatorGraph() = %v, %v", result, err)
	}
	if _, err := graph.VertexCount(); !errors.Is(err, ErrClosed) {
		t.Errorf("failed query graph error = %v, want %v", err, ErrClosed)
	}

	graph = testGraphFromEdges(t, 1, nil, false)
	result, err := collectBinaryGraphResult(
		func() (*Graph, error) { return graph, nil },
		func(*Graph) (GraphIDMapping, GraphIDMapping, error) {
			return GraphIDMapping{}, GraphIDMapping{}, forced
		},
	)
	if result.Graph != nil || !errors.Is(err, forced) {
		t.Errorf("mapping failure = %#v, %v", result, err)
	}
	if _, err := graph.VertexCount(); !errors.Is(err, ErrClosed) {
		t.Errorf("mapping failure graph error = %v, want %v", err, ErrClosed)
	}
	if _, err := mappingFromExactInverse(1, []int{1}); err == nil {
		t.Error("out-of-range inverse provenance returned nil error")
	}
	if _, err := mappingFromExactInverse(1, []int{0, 0}); err == nil {
		t.Error("repeated inverse provenance returned nil error")
	}
}

func TestCollectDifferenceCleansFailuresAndStructuralMappingIsDeterministic(t *testing.T) {
	forced := errors.New("forced failure")
	for _, stage := range []string{
		"query", "query with graph", "nil graph", "result edges", "vertices", "edges",
	} {
		t.Run(stage, func(t *testing.T) {
			closed := 0
			operations := differenceOperations{
				query:         func() (*Graph, error) { return &Graph{}, nil },
				resultEdges:   func(*Graph) ([]Edge, error) { return []Edge{{0, 1}}, nil },
				vertexMapping: identityIDMapping,
				edgeMapping:   structuralDifferenceEdgeMapping,
				closeGraph:    func(*Graph) { closed++ },
			}
			switch stage {
			case "query":
				operations.query = func() (*Graph, error) { return nil, forced }
			case "query with graph":
				operations.query = func() (*Graph, error) { return &Graph{}, forced }
			case "nil graph":
				operations.query = func() (*Graph, error) { return nil, nil }
			case "result edges":
				operations.resultEdges = func(*Graph) ([]Edge, error) { return nil, forced }
			case "vertices":
				operations.vertexMapping = func(int) (IDMapping, error) { return IDMapping{}, forced }
			case "edges":
				operations.edgeMapping = func([]Edge, []Edge, bool) (IDMapping, error) { return IDMapping{}, forced }
			}
			result, err := collectDifference(2, []Edge{{0, 1}}, true, operations)
			if result.Graph != nil || err == nil {
				t.Errorf("collectDifference() = %#v, %v, want zero, error", result, err)
			}
			wantClosed := 1
			if stage == "query" || stage == "nil graph" {
				wantClosed = 0
			}
			if closed != wantClosed {
				t.Errorf("close count = %d, want %d", closed, wantClosed)
			}
		})
	}

	mapping, err := structuralDifferenceEdgeMapping(
		[]Edge{{3, 1}, {1, 3}, {0, 2}},
		[]Edge{{1, 3}, {0, 2}},
		false,
	)
	if err != nil {
		t.Fatal(err)
	}
	want := IDMapping{OldToNew: []int{0, RemovedID, 1}, NewToOld: []int{0, 2}}
	if !reflect.DeepEqual(mapping, want) {
		t.Errorf("parallel structural mapping = %#v, want %#v", mapping, want)
	}
	if _, err := structuralDifferenceEdgeMapping([]Edge{{0, 1}}, []Edge{{1, 1}}, true); err == nil {
		t.Error("result edge absent from source returned nil error")
	}
}

func assertOperatorMappingConsistent(t *testing.T, source, result *Graph, mapping GraphIDMapping) {
	t.Helper()
	sourceEdges, err := source.Edges()
	if err != nil {
		t.Fatal(err)
	}
	resultEdges, err := result.Edges()
	if err != nil {
		t.Fatal(err)
	}
	for oldID, newID := range mapping.Edges.OldToNew {
		if newID == RemovedID {
			continue
		}
		if newID < 0 || newID >= len(resultEdges) || sourceEdges[oldID] != resultEdges[newID] {
			t.Errorf("edge mapping %d -> %d does not preserve endpoints", oldID, newID)
		}
	}
}

func assertNonNilBinaryMappings(t *testing.T, result BinaryGraphOperatorResult) {
	t.Helper()
	for name, mapping := range map[string]IDMapping{
		"left vertices":  result.Left.Vertices,
		"left edges":     result.Left.Edges,
		"right vertices": result.Right.Vertices,
		"right edges":    result.Right.Edges,
	} {
		if mapping.OldToNew == nil || mapping.NewToOld == nil {
			t.Errorf("%s mapping contains nil slice: %#v", name, mapping)
		}
	}
}

func undirectedCompositionMatches(left, right, result Edge) bool {
	leftDirections := []Edge{left}
	if left.From != left.To {
		leftDirections = append(leftDirections, Edge{From: left.To, To: left.From})
	}
	rightDirections := []Edge{right}
	if right.From != right.To {
		rightDirections = append(rightDirections, Edge{From: right.To, To: right.From})
	}
	resultKey := endpointKey(result, false)
	for _, leftDirection := range leftDirections {
		for _, rightDirection := range rightDirections {
			if leftDirection.To != rightDirection.From {
				continue
			}
			candidate := endpointKey(Edge{From: leftDirection.From, To: rightDirection.To}, false)
			if candidate == resultKey {
				return true
			}
		}
	}
	return false
}

func TestComplementRejectsParallelEdges(t *testing.T) {
	g, err := NewGraphFromEdges(2, []Edge{{0, 1}, {0, 1}}, false)
	if err != nil {
		t.Fatal(err)
	}
	defer g.Close()

	if _, err := g.Complement(false); err == nil {
		t.Error("expected error when calling Complement on graph with parallel edges")
	}
}

func TestJoinOrderingMappingsAttributesAndMultiplicity(t *testing.T) {
	left := testGraphFromEdges(t, 2, []Edge{{0, 1}, {0, 1}}, false)
	right := testGraphFromEdges(t, 2, []Edge{{0, 0}}, false)
	defer left.Close()
	defer right.Close()
	if err := left.SetVertexStringAttributes("label", []string{"a", "b"}); err != nil {
		t.Fatal(err)
	}
	if err := right.SetVertexStringAttributes("label", []string{"c", "d"}); err != nil {
		t.Fatal(err)
	}
	if err := left.SetEdgeNumericAttributes("weight", []float64{2, 3}); err != nil {
		t.Fatal(err)
	}
	if err := right.SetEdgeNumericAttributes("weight", []float64{5}); err != nil {
		t.Fatal(err)
	}

	result, err := left.Join(right, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer result.Graph.Close()
	assertGraphShape(t, result.Graph, 4, 7, false)
	assertOperatorMappingConsistent(t, left, result.Graph, result.Left)
	assertOffsetOperatorMappingConsistent(t, right, result.Graph, result.Right, 2)
	if !reflect.DeepEqual(result.Left.Vertices.OldToNew, []int{0, 1}) ||
		!reflect.DeepEqual(result.Right.Vertices.OldToNew, []int{2, 3}) {
		t.Errorf("join vertex mappings = %#v / %#v", result.Left.Vertices, result.Right.Vertices)
	}
	labels, err := result.Graph.VertexStringAttributes("label")
	if err != nil || !reflect.DeepEqual(labels, []string{"a", "b", "c", "d"}) {
		t.Errorf("join labels = %v, %v", labels, err)
	}
	weights, err := result.Graph.EdgeNumericAttributes("weight")
	if err != nil {
		t.Fatal(err)
	}
	for sourceID, resultID := range result.Left.Edges.OldToNew {
		if weights[resultID] != []float64{2, 3}[sourceID] {
			t.Errorf("left edge weight %d = %v", sourceID, weights[resultID])
		}
	}
	if weights[result.Right.Edges.OldToNew[0]] != 5 {
		t.Errorf("right edge weight = %v", weights[result.Right.Edges.OldToNew[0]])
	}
	assertNonNilBinaryMappings(t, result)
}

func TestJoinRejectsUnresolvedAttributeConflict(t *testing.T) {
	left := testGraphFromEdges(t, 1, nil, false)
	right := testGraphFromEdges(t, 1, nil, false)
	defer left.Close()
	defer right.Close()
	if err := left.SetGraphNumericAttribute("score", 1); err != nil {
		t.Fatal(err)
	}
	if err := right.SetGraphNumericAttribute("score", 2); err != nil {
		t.Fatal(err)
	}
	if result, err := left.Join(right, nil); err == nil || result.Graph != nil {
		t.Errorf("Join without graph-attribute policy = %#v, %v", result, err)
	}
}

func TestDirectedJoinAddsBothDirections(t *testing.T) {
	left := testGraphFromEdges(t, 1, nil, true)
	right := testGraphFromEdges(t, 2, nil, true)
	defer left.Close()
	defer right.Close()
	result, err := left.Join(right, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer result.Graph.Close()
	assertGraphShape(t, result.Graph, 3, 4, true)
	assertEdgeMultiset(t, result.Graph, []Edge{{0, 1}, {1, 0}, {0, 2}, {2, 0}}, true)
}

func TestGraphProductModesVertexProvenanceAndOwnership(t *testing.T) {
	left := testGraphFromEdges(t, 2, []Edge{{0, 1}}, true)
	right := testGraphFromEdges(t, 3, []Edge{{0, 1}, {1, 2}}, true)
	if err := left.SetGraphStringAttribute("lost", "left"); err != nil {
		t.Fatal(err)
	}
	if err := right.SetVertexNumericAttributes("lost", []float64{1, 2, 3}); err != nil {
		t.Fatal(err)
	}
	wantEdges := map[GraphProductMode]int{
		GraphProductCartesian:     7,
		GraphProductLexicographic: 13,
		GraphProductStrong:        9,
		GraphProductTensor:        2,
		GraphProductModular:       6,
	}
	var retained GraphProductResult
	for mode, edgeCount := range wantEdges {
		result, err := left.Product(right, mode)
		if err != nil {
			t.Fatalf("Product(%d): %v", mode, err)
		}
		assertGraphShape(t, result.Graph, 6, edgeCount, true)
		wantVertices := []ProductVertexProvenance{{0, 0}, {0, 1}, {0, 2}, {1, 0}, {1, 1}, {1, 2}}
		if !reflect.DeepEqual(result.Vertices, wantVertices) {
			t.Errorf("Product(%d) provenance = %#v", mode, result.Vertices)
		}
		if !reflect.DeepEqual(result.LeftToResult, [][]int{{0, 1, 2}, {3, 4, 5}}) ||
			!reflect.DeepEqual(result.RightToResult, [][]int{{0, 3}, {1, 4}, {2, 5}}) {
			t.Errorf("Product(%d) source maps = %#v / %#v", mode, result.LeftToResult, result.RightToResult)
		}
		if attributes, err := result.Graph.GraphAttributes(); err != nil || len(attributes) != 0 {
			t.Errorf("Product(%d) graph attributes = %v, %v", mode, attributes, err)
		}
		if mode == GraphProductCartesian {
			retained = result
		} else {
			_ = result.Graph.Close()
		}
	}
	if err := left.Close(); err != nil {
		t.Fatal(err)
	}
	if err := right.Close(); err != nil {
		t.Fatal(err)
	}
	assertGraphShape(t, retained.Graph, 6, 7, true)
	if err := retained.Graph.Close(); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(retained.RightToResult[1], []int{1, 4}) {
		t.Errorf("provenance after closure = %#v", retained)
	}
}

func TestRootedProductOrderingRootAndEmptyOperands(t *testing.T) {
	left := testGraphFromEdges(t, 2, []Edge{{0, 1}}, false)
	right := testGraphFromEdges(t, 3, []Edge{{0, 1}, {1, 2}}, false)
	defer left.Close()
	defer right.Close()
	result, err := left.RootedProduct(right, 1)
	if err != nil {
		t.Fatal(err)
	}
	defer result.Graph.Close()
	assertGraphShape(t, result.Graph, 6, 5, false)
	assertEdgeMultiset(t, result.Graph, []Edge{{0, 1}, {1, 2}, {3, 4}, {4, 5}, {1, 4}}, false)

	empty := testGraphFromEdges(t, 0, nil, false)
	defer empty.Close()
	emptyLeft, err := empty.Product(right, GraphProductCartesian)
	if err != nil {
		t.Fatal(err)
	}
	defer emptyLeft.Graph.Close()
	assertGraphShape(t, emptyLeft.Graph, 0, 0, false)
	if emptyLeft.Vertices == nil || emptyLeft.LeftToResult == nil || emptyLeft.RightToResult == nil {
		t.Errorf("empty product provenance contains nil slice: %#v", emptyLeft)
	}

	singleton := testGraphFromEdges(t, 1, nil, false)
	defer singleton.Close()
	rootFromSecond, err := singleton.RootedProduct(right, 2)
	if err != nil {
		t.Fatalf("root valid only in second operand: %v", err)
	}
	defer rootFromSecond.Graph.Close()
	assertGraphShape(t, rootFromSecond.Graph, 3, 2, false)
}

func TestProductAndJoinValidationAndClosedGraphs(t *testing.T) {
	undirected := testGraphFromEdges(t, 2, []Edge{{0, 1}}, false)
	directed := testGraphFromEdges(t, 2, []Edge{{0, 1}}, true)
	nonSimple := testGraphFromEdges(t, 2, []Edge{{0, 1}, {0, 1}}, false)
	closed := testGraphFromEdges(t, 1, nil, false)
	_ = closed.Close()
	defer undirected.Close()
	defer directed.Close()
	defer nonSimple.Close()

	if got, err := undirected.Product(directed, GraphProductCartesian); err == nil || got.Graph != nil {
		t.Errorf("mixed product = %#v, %v", got, err)
	}
	if got, err := undirected.Join(directed, nil); err == nil || got.Graph != nil {
		t.Errorf("mixed join = %#v, %v", got, err)
	}
	if got, err := undirected.Product(undirected, GraphProductMode(99)); err == nil || got.Graph != nil {
		t.Errorf("invalid product mode = %#v, %v", got, err)
	}
	for _, root := range []int{-1, 2} {
		if got, err := undirected.RootedProduct(undirected, root); err == nil || got.Graph != nil {
			t.Errorf("invalid root %d = %#v, %v", root, got, err)
		}
	}
	if got, err := nonSimple.Product(undirected, GraphProductModular); err == nil || got.Graph != nil {
		t.Errorf("non-simple modular product = %#v, %v", got, err)
	}
	if got, err := undirected.Product(nonSimple, GraphProductModular); err == nil || got.Graph != nil {
		t.Errorf("non-simple right modular product = %#v, %v", got, err)
	}
	if got, err := closed.Product(undirected, GraphProductCartesian); !errors.Is(err, ErrClosed) || got.Graph != nil {
		t.Errorf("closed product = %#v, %v", got, err)
	}
	if got, err := undirected.Join(closed, nil); !errors.Is(err, ErrClosed) || got.Graph != nil {
		t.Errorf("closed join = %#v, %v", got, err)
	}
	if got, err := undirected.RootedProduct(closed, 0); !errors.Is(err, ErrClosed) || got.Graph != nil {
		t.Errorf("closed rooted product = %#v, %v", got, err)
	}
}

func TestStructuralSubsetEdgeMappingRejectsMissingEdge(t *testing.T) {
	if mapping, err := structuralSubsetEdgeMapping([]Edge{{0, 1}}, nil, true); err == nil || mapping.OldToNew != nil {
		t.Errorf("missing subset edge = %#v, %v", mapping, err)
	}
}

func TestProductAndJoinConcurrentReversedOperands(t *testing.T) {
	left := testGraphFromEdges(t, 3, []Edge{{0, 1}, {1, 2}}, false)
	right := testGraphFromEdges(t, 2, []Edge{{0, 1}}, false)
	defer left.Close()
	defer right.Close()
	selfProduct, err := left.Product(left, GraphProductTensor)
	if err != nil {
		t.Fatalf("repeated product operand: %v", err)
	}
	_ = selfProduct.Graph.Close()
	selfJoin, err := right.Join(right, nil)
	if err != nil {
		t.Fatalf("repeated join operand: %v", err)
	}
	_ = selfJoin.Graph.Close()
	var wait sync.WaitGroup
	for index := 0; index < 12; index++ {
		wait.Add(2)
		go func() {
			defer wait.Done()
			result, err := left.Product(right, GraphProductStrong)
			if err != nil {
				t.Errorf("forward product: %v", err)
				return
			}
			_ = result.Graph.Close()
		}()
		go func() {
			defer wait.Done()
			result, err := right.Join(left, nil)
			if err != nil {
				t.Errorf("reverse join: %v", err)
				return
			}
			_ = result.Graph.Close()
		}()
	}
	wait.Wait()
}

func TestProductConcurrentUseAndClose(t *testing.T) {
	left := testGraphFromEdges(t, 8, []Edge{{0, 1}, {1, 2}, {2, 3}, {3, 4}}, false)
	right := testGraphFromEdges(t, 8, []Edge{{4, 5}, {5, 6}, {6, 7}}, false)
	var wait sync.WaitGroup
	for range 8 {
		wait.Add(1)
		go func() {
			defer wait.Done()
			result, err := left.Product(right, GraphProductCartesian)
			if result.Graph != nil {
				_ = result.Graph.Close()
			}
			if err != nil && !errors.Is(err, ErrClosed) {
				t.Errorf("Product close race error = %v", err)
			}
		}()
	}
	wait.Add(1)
	go func() { defer wait.Done(); _ = right.Close() }()
	wait.Wait()
	_ = left.Close()
}

func assertOffsetOperatorMappingConsistent(t *testing.T, source, result *Graph, mapping GraphIDMapping, vertexOffset int) {
	t.Helper()
	sourceEdges, err := source.Edges()
	if err != nil {
		t.Fatal(err)
	}
	resultEdges, err := result.Edges()
	if err != nil {
		t.Fatal(err)
	}
	for oldID, newID := range mapping.Edges.OldToNew {
		want := sourceEdges[oldID]
		want.From += vertexOffset
		want.To += vertexOffset
		if newID < 0 || newID >= len(resultEdges) || endpointKey(want, false) != endpointKey(resultEdges[newID], false) {
			t.Errorf("offset edge mapping %d -> %d does not preserve endpoints", oldID, newID)
		}
	}
}

func assertEdgeMultiset(t *testing.T, graph *Graph, want []Edge, directed bool) {
	t.Helper()
	got, err := graph.Edges()
	if err != nil {
		t.Fatal(err)
	}
	counts := func(edges []Edge) map[edgeEndpointKey]int {
		result := make(map[edgeEndpointKey]int)
		for _, edge := range edges {
			result[endpointKey(edge, directed)]++
		}
		return result
	}
	if !reflect.DeepEqual(counts(got), counts(want)) {
		t.Errorf("edge multiset = %v, want %v", got, want)
	}
}
