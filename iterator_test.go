package igraph

import (
	"errors"
	"reflect"
	"sync"
	"testing"
)

func TestSelectedIDsFollowSelectorOrder(t *testing.T) {
	graph, err := NewGraphFromEdges(4, []Edge{{0, 1}, {1, 2}, {2, 3}}, true)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = graph.Close() })
	vertices, _ := VertexIDs(3, 1, 3)
	if got, err := graph.SelectedVertexIDs(vertices); err != nil || !reflect.DeepEqual(got, []int{3, 1, 3}) {
		t.Errorf("SelectedVertexIDs() = %v, %v", got, err)
	}
	edges, _ := EdgeIDs(2, 0, 2)
	if got, err := graph.SelectedEdgeIDs(edges); err != nil || !reflect.DeepEqual(got, []int{2, 0, 2}) {
		t.Errorf("SelectedEdgeIDs() = %v, %v", got, err)
	}
}

func TestSelectedIDsEmpty(t *testing.T) {
	graph, err := NewGraphFromEdges(2, []Edge{{0, 1}}, false)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = graph.Close() })
	emptyVertices, err := VertexIDs()
	if err != nil {
		t.Fatal(err)
	}
	emptyEdges, err := EdgeIDs()
	if err != nil {
		t.Fatal(err)
	}
	emptyPairs, err := EdgePairs(nil, true)
	if err != nil {
		t.Fatal(err)
	}
	for name, selector := range map[string]VertexSelector{
		"none": NoVertices(),
		"IDs":  emptyVertices,
	} {
		if got, err := graph.SelectedVertexIDs(selector); err != nil || got == nil || len(got) != 0 {
			t.Errorf("SelectedVertexIDs(%s) = %#v, %v", name, got, err)
		}
	}
	for name, selector := range map[string]EdgeSelector{
		"none":  NoEdges(),
		"IDs":   emptyEdges,
		"pairs": emptyPairs,
	} {
		if got, err := graph.SelectedEdgeIDs(selector); err != nil || got == nil || len(got) != 0 {
			t.Errorf("SelectedEdgeIDs(%s) = %#v, %v", name, got, err)
		}
	}
}

func TestSelectedIDsRemainValidAfterEarlyTerminationAndClose(t *testing.T) {
	graph, err := NewGraphFromEdges(3, []Edge{{0, 1}, {1, 2}}, false)
	if err != nil {
		t.Fatal(err)
	}
	vertices, err := graph.SelectedVertexIDs(AllVertices())
	if err != nil {
		t.Fatal(err)
	}
	edges, err := graph.SelectedEdgeIDs(AllEdges())
	if err != nil {
		t.Fatal(err)
	}
	for range vertices {
		break
	}
	for range edges {
		break
	}
	if err := graph.Close(); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(vertices, []int{0, 1, 2}) || !reflect.DeepEqual(edges, []int{0, 1}) {
		t.Errorf("results after graph close = %v, %v", vertices, edges)
	}
}

func TestSelectedIDsRejectErrorsAndClosedGraphs(t *testing.T) {
	graph, err := NewGraphFromEdges(2, []Edge{{0, 1}}, true)
	if err != nil {
		t.Fatal(err)
	}
	badVertices, _ := VertexIDs(2)
	if _, err := graph.SelectedVertexIDs(badVertices); err == nil {
		t.Error("invalid vertex selector error = nil")
	}
	badEdges, _ := EdgeIDs(1)
	if _, err := graph.SelectedEdgeIDs(badEdges); err == nil {
		t.Error("invalid edge selector error = nil")
	}
	if err := graph.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := graph.SelectedVertexIDs(AllVertices()); !errors.Is(err, ErrClosed) {
		t.Errorf("closed vertex selection error = %v", err)
	}
	if _, err := graph.SelectedEdgeIDs(AllEdges()); !errors.Is(err, ErrClosed) {
		t.Errorf("closed edge selection error = %v", err)
	}
}

func TestCIteratorConstructorsPropagateInvalidSelectors(t *testing.T) {
	graph, err := NewGraph()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = graph.Close() })
	invalidVertices := VertexSelector{kind: vertexSelectorKind(255)}
	if ids, err := materializeVertexIDs(&graph.graph, invalidVertices); ids != nil || err == nil {
		t.Errorf("materializeVertexIDs() = %v, %v, want nil, error", ids, err)
	}
	invalidEdges := EdgeSelector{kind: edgeSelectorKind(255)}
	if ids, err := materializeEdgeIDs(&graph.graph, invalidEdges); ids != nil || err == nil {
		t.Errorf("materializeEdgeIDs() = %v, %v, want nil, error", ids, err)
	}
}

func TestSelectedIDsSerializeWithGraphMutation(t *testing.T) {
	graph, err := NewGraphFromEdges(2, []Edge{{0, 1}}, false)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = graph.Close() })

	var wait sync.WaitGroup
	wait.Add(2)
	errorsChannel := make(chan error, 2)
	go func() {
		defer wait.Done()
		_, err := graph.SelectedVertexIDs(AllVertices())
		errorsChannel <- err
	}()
	go func() {
		defer wait.Done()
		errorsChannel <- graph.AddVertices(1)
	}()
	wait.Wait()
	close(errorsChannel)
	for err := range errorsChannel {
		if err != nil {
			t.Errorf("concurrent operation error = %v", err)
		}
	}
	if count, err := graph.VertexCount(); err != nil || count != 3 {
		t.Errorf("VertexCount() = %d, %v, want 3, nil", count, err)
	}
}
