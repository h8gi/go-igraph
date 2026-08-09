package igraph

import (
	"errors"
	"reflect"
	"testing"
)

func TestDeleteEdgesDirectedLoopsParallelAndMappings(t *testing.T) {
	graph := deletionTestGraph(t, true)
	selector, err := EdgeIDs(1, 1, 2)
	if err != nil {
		t.Fatal(err)
	}
	mapping, err := graph.DeleteEdges(selector)
	if err != nil {
		t.Fatalf("DeleteEdges() error = %v", err)
	}
	assertIDMapping(t, mapping.Vertices, []int{0, 1, 2, 3}, []int{0, 1, 2, 3})
	assertIDMapping(t, mapping.Edges, []int{0, RemovedID, RemovedID, 1, 2}, []int{0, 3, 4})
	assertGraphState(t, graph, 4, true, []Edge{{0, 1}, {2, 0}, {3, 2}})
	if adjacent, err := graph.AreAdjacent(0, 1); err != nil || !adjacent {
		t.Errorf("AreAdjacent(0, 1) = %t, %v, want true, nil", adjacent, err)
	}
	if err := graph.AddEdge(1, 1); err != nil {
		t.Fatalf("graph unusable after DeleteEdges: %v", err)
	}
}

func TestDeleteEdgesPairDuplicateAndUndirected(t *testing.T) {
	graph := deletionTestGraph(t, false)
	selector, err := EdgePairs([]Edge{{0, 1}, {0, 1}}, false)
	if err != nil {
		t.Fatal(err)
	}
	mapping, err := graph.DeleteEdges(selector)
	if err != nil {
		t.Fatal(err)
	}
	assertIDMapping(t, mapping.Edges, []int{0, RemovedID, 1, 2, 3}, []int{0, 2, 3, 4})
	assertGraphState(t, graph, 4, false, []Edge{{0, 1}, {1, 1}, {0, 2}, {2, 3}})
}

func TestDeleteEdgesEmptyAndAll(t *testing.T) {
	graph := deletionTestGraph(t, true)
	mapping, err := graph.DeleteEdges(NoEdges())
	if err != nil {
		t.Fatal(err)
	}
	assertIDMapping(t, mapping.Vertices, []int{0, 1, 2, 3}, []int{0, 1, 2, 3})
	assertIDMapping(t, mapping.Edges, []int{0, 1, 2, 3, 4}, []int{0, 1, 2, 3, 4})
	assertGraphState(t, graph, 4, true, deletionEdges())

	mapping, err = graph.DeleteEdges(AllEdges())
	if err != nil {
		t.Fatal(err)
	}
	assertIDMapping(t, mapping.Edges, []int{RemovedID, RemovedID, RemovedID, RemovedID, RemovedID}, []int{})
	assertGraphState(t, graph, 4, true, []Edge{})
}

func TestDeleteVerticesDirectedLoopsParallelAndMappings(t *testing.T) {
	graph := deletionTestGraph(t, true)
	selector, err := VertexIDs(1, 1)
	if err != nil {
		t.Fatal(err)
	}
	mapping, err := graph.DeleteVertices(selector)
	if err != nil {
		t.Fatalf("DeleteVertices() error = %v", err)
	}
	assertIDMapping(t, mapping.Vertices, []int{0, RemovedID, 1, 2}, []int{0, 2, 3})
	assertIDMapping(t, mapping.Edges, []int{RemovedID, RemovedID, RemovedID, 0, 1}, []int{3, 4})
	assertGraphState(t, graph, 3, true, []Edge{{1, 0}, {2, 1}})
	if adjacent, err := graph.AreAdjacent(1, 0); err != nil || !adjacent {
		t.Errorf("AreAdjacent(1, 0) = %t, %v, want true, nil", adjacent, err)
	}
	if err := graph.AddEdge(0, 2); err != nil {
		t.Fatalf("graph unusable after DeleteVertices: %v", err)
	}
}

func TestDeleteVerticesUndirectedAndRange(t *testing.T) {
	graph := deletionTestGraph(t, false)
	selector, err := VertexRange(0, 1)
	if err != nil {
		t.Fatal(err)
	}
	mapping, err := graph.DeleteVertices(selector)
	if err != nil {
		t.Fatal(err)
	}
	assertIDMapping(t, mapping.Vertices, []int{RemovedID, 0, 1, 2}, []int{1, 2, 3})
	assertIDMapping(t, mapping.Edges, []int{RemovedID, RemovedID, 0, RemovedID, 1}, []int{2, 4})
	assertGraphState(t, graph, 3, false, []Edge{{0, 0}, {1, 2}})
}

func TestDeleteVerticesEmptyAndAll(t *testing.T) {
	graph := deletionTestGraph(t, true)
	mapping, err := graph.DeleteVertices(NoVertices())
	if err != nil {
		t.Fatal(err)
	}
	assertIDMapping(t, mapping.Vertices, []int{0, 1, 2, 3}, []int{0, 1, 2, 3})
	assertIDMapping(t, mapping.Edges, []int{0, 1, 2, 3, 4}, []int{0, 1, 2, 3, 4})
	assertGraphState(t, graph, 4, true, deletionEdges())

	mapping, err = graph.DeleteVertices(AllVertices())
	if err != nil {
		t.Fatal(err)
	}
	assertIDMapping(t, mapping.Vertices, []int{RemovedID, RemovedID, RemovedID, RemovedID}, []int{})
	assertIDMapping(t, mapping.Edges, []int{RemovedID, RemovedID, RemovedID, RemovedID, RemovedID}, []int{})
	assertGraphState(t, graph, 0, true, []Edge{})
}

func TestDeletionOnEmptyGraphReturnsNonNilMappings(t *testing.T) {
	graph, err := NewGraph()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = graph.Close() })
	vertexMapping, err := graph.DeleteVertices(AllVertices())
	if err != nil {
		t.Fatal(err)
	}
	edgeMapping, err := graph.DeleteEdges(AllEdges())
	if err != nil {
		t.Fatal(err)
	}
	for _, mapping := range []GraphIDMapping{vertexMapping, edgeMapping} {
		assertIDMapping(t, mapping.Vertices, []int{}, []int{})
		assertIDMapping(t, mapping.Edges, []int{}, []int{})
	}
}

func TestDeletionMappingsSurviveGraphClose(t *testing.T) {
	graph := deletionTestGraph(t, true)
	selector, _ := VertexIDs(1)
	mapping, err := graph.DeleteVertices(selector)
	if err != nil {
		t.Fatal(err)
	}
	if err := graph.Close(); err != nil {
		t.Fatal(err)
	}
	assertIDMapping(t, mapping.Vertices, []int{0, RemovedID, 1, 2}, []int{0, 2, 3})
	assertIDMapping(t, mapping.Edges, []int{RemovedID, RemovedID, RemovedID, 0, 1}, []int{3, 4})
	mapping.Vertices.OldToNew[0] = 99
}

func TestDeletionRejectsClosedNilAndInvalidSelectorsWithoutMutation(t *testing.T) {
	for _, test := range []struct {
		name   string
		delete func(*Graph) error
	}{
		{
			name: "vertex out of range",
			delete: func(graph *Graph) error {
				selector, _ := VertexIDs(4)
				_, err := graph.DeleteVertices(selector)
				return err
			},
		},
		{
			name: "invalid vertex kind",
			delete: func(graph *Graph) error {
				_, err := graph.DeleteVertices(VertexSelector{kind: vertexSelectorKind(255)})
				return err
			},
		},
		{
			name: "edge out of range",
			delete: func(graph *Graph) error {
				selector, _ := EdgeIDs(5)
				_, err := graph.DeleteEdges(selector)
				return err
			},
		},
		{
			name: "missing edge pair",
			delete: func(graph *Graph) error {
				selector, _ := EdgePairs([]Edge{{0, 3}}, true)
				_, err := graph.DeleteEdges(selector)
				return err
			},
		},
		{
			name: "invalid edge kind",
			delete: func(graph *Graph) error {
				_, err := graph.DeleteEdges(EdgeSelector{kind: edgeSelectorKind(255)})
				return err
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			graph := deletionTestGraph(t, true)
			before := captureGraphState(t, graph)
			if err := test.delete(graph); err == nil {
				t.Fatal("deletion error = nil")
			}
			assertCapturedGraphState(t, graph, before)
		})
	}

	closed := deletionTestGraph(t, true)
	if err := closed.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := closed.DeleteVertices(AllVertices()); !errors.Is(err, ErrClosed) {
		t.Errorf("closed DeleteVertices error = %v, want %v", err, ErrClosed)
	}
	if _, err := closed.DeleteEdges(AllEdges()); !errors.Is(err, ErrClosed) {
		t.Errorf("closed DeleteEdges error = %v, want %v", err, ErrClosed)
	}
	var nilGraph *Graph
	if _, err := nilGraph.DeleteVertices(AllVertices()); !errors.Is(err, ErrClosed) {
		t.Errorf("nil DeleteVertices error = %v, want %v", err, ErrClosed)
	}
	if _, err := nilGraph.DeleteEdges(AllEdges()); !errors.Is(err, ErrClosed) {
		t.Errorf("nil DeleteEdges error = %v, want %v", err, ErrClosed)
	}
}

func TestDeleteVerticesInjectedFailuresAreAtomic(t *testing.T) {
	stages := []deletionStage{
		deletionBeforeClone,
		deletionBeforeEdgeSnapshot,
		deletionBeforeSelectorInit,
		deletionBeforeFirstMappingInit,
		deletionBeforeSecondMappingInit,
		deletionAtMutation,
		deletionAfterMutation,
	}
	selector, _ := VertexIDs(1)
	for _, stage := range stages {
		t.Run(stageName(stage), func(t *testing.T) {
			graph := deletionTestGraph(t, true)
			before := captureGraphState(t, graph)
			injected := errors.New("injected failure")
			mapping, err := graph.deleteVertices(selector, failDeletionAt(stage, injected))
			if !errors.Is(err, injected) {
				t.Errorf("deleteVertices() error = %v, want injected", err)
			}
			if !reflect.DeepEqual(mapping, GraphIDMapping{}) {
				t.Errorf("deleteVertices() mapping = %#v, want zero", mapping)
			}
			assertCapturedGraphState(t, graph, before)
		})
	}
}

func TestDeleteEdgesInjectedFailuresAreAtomic(t *testing.T) {
	selector, _ := EdgeIDs(1)
	for _, stage := range []deletionStage{deletionBeforeClone, deletionBeforeSelectorInit, deletionAtMutation, deletionAfterMutation} {
		t.Run(stageName(stage), func(t *testing.T) {
			graph := deletionTestGraph(t, false)
			before := captureGraphState(t, graph)
			injected := errors.New("injected failure")
			mapping, err := graph.deleteEdges(selector, failDeletionAt(stage, injected))
			if !errors.Is(err, injected) {
				t.Errorf("deleteEdges() error = %v, want injected", err)
			}
			if !reflect.DeepEqual(mapping, GraphIDMapping{}) {
				t.Errorf("deleteEdges() mapping = %#v, want zero", mapping)
			}
			assertCapturedGraphState(t, graph, before)
		})
	}
}

func TestDeletionHelpersRejectInconsistentInternalResults(t *testing.T) {
	if mapping, err := deletionIDMapping(-1, nil); err == nil {
		t.Errorf("deletionIDMapping(-1) = %#v, nil, want error", mapping)
	}
	if mapping, err := deletionIDMapping(1, []int{-1}); err == nil {
		t.Errorf("deletionIDMapping(invalid ID) = %#v, nil, want error", mapping)
	}
	identity, err := identityIDMapping(1)
	if err != nil {
		t.Fatal(err)
	}
	for _, edge := range []Edge{{From: -1, To: 0}, {From: 0, To: 1}} {
		if mapping, err := vertexDeletionEdgeMapping([]Edge{edge}, identity); err == nil {
			t.Errorf("vertexDeletionEdgeMapping(%v) = %#v, nil, want error", edge, mapping)
		}
	}
	if edges, err := edgeSlice(nil); edges != nil || err == nil {
		t.Errorf("edgeSlice(nil) = %v, %v, want nil, error", edges, err)
	}
	if err := validateDeletionCount("edge", 2, 1); err == nil {
		t.Error("validateDeletionCount(mismatch) error = nil")
	}
	if err := validateReverseDeletionMapping([]int{0}, []int{1}); err == nil {
		t.Error("validateReverseDeletionMapping(mismatch) error = nil")
	}
}

func failDeletionAt(stage deletionStage, failure error) deletionFailureHook {
	return func(current deletionStage) error {
		if current == stage {
			return failure
		}
		return nil
	}
}

func stageName(stage deletionStage) string {
	return []string{
		"before-clone",
		"before-edge-snapshot",
		"before-selector-init",
		"before-first-map",
		"before-second-map",
		"mutation-call",
		"after-mutation",
	}[stage]
}

type capturedGraphState struct {
	vertices int
	directed bool
	edges    []Edge
}

func captureGraphState(t *testing.T, graph *Graph) capturedGraphState {
	t.Helper()
	vertices, err := graph.VertexCount()
	if err != nil {
		t.Fatal(err)
	}
	directed, err := graph.IsDirected()
	if err != nil {
		t.Fatal(err)
	}
	edges, err := graph.Edges()
	if err != nil {
		t.Fatal(err)
	}
	return capturedGraphState{vertices: vertices, directed: directed, edges: edges}
}

func assertCapturedGraphState(t *testing.T, graph *Graph, want capturedGraphState) {
	t.Helper()
	assertGraphState(t, graph, want.vertices, want.directed, want.edges)
}

func assertGraphState(t *testing.T, graph *Graph, vertices int, directed bool, edges []Edge) {
	t.Helper()
	got := captureGraphState(t, graph)
	if got.vertices != vertices || got.directed != directed || !reflect.DeepEqual(got.edges, edges) {
		t.Errorf("graph state = %#v, want vertices=%d directed=%t edges=%v", got, vertices, directed, edges)
	}
}

func assertIDMapping(t *testing.T, mapping IDMapping, oldToNew, newToOld []int) {
	t.Helper()
	if mapping.OldToNew == nil || mapping.NewToOld == nil {
		t.Errorf("mapping contains nil slice: %#v", mapping)
	}
	if !reflect.DeepEqual(mapping.OldToNew, oldToNew) || !reflect.DeepEqual(mapping.NewToOld, newToOld) {
		t.Errorf("mapping = %#v, want old-to-new=%v new-to-old=%v", mapping, oldToNew, newToOld)
	}
}

func deletionTestGraph(t *testing.T, directed bool) *Graph {
	t.Helper()
	graph, err := NewGraphFromEdges(4, deletionEdges(), directed)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = graph.Close() })
	return graph
}

func deletionEdges() []Edge {
	return []Edge{{0, 1}, {0, 1}, {1, 1}, {2, 0}, {3, 2}}
}
