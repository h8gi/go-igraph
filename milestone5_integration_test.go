package igraph

import (
	"errors"
	"reflect"
	"sort"
	"testing"
)

func TestMilestone5DirectedTransformationPipeline(t *testing.T) {
	source := testGraphFromEdges(t, 8, []Edge{
		{0, 1}, {0, 1}, {1, 2}, {2, 0}, {2, 2},
		{3, 4}, {4, 3}, {6, 7},
	}, true)

	deleted, err := source.DeleteEdges(mustEdgeIDs(t, 1))
	if err != nil {
		t.Fatal(err)
	}
	assertIDMapping(t, deleted.Vertices, []int{0, 1, 2, 3, 4, 5, 6, 7}, []int{0, 1, 2, 3, 4, 5, 6, 7})
	assertIDMapping(t, deleted.Edges, []int{0, RemovedID, 1, 2, 3, 4, 5, 6}, []int{0, 2, 3, 4, 5, 6, 7})

	induced, err := source.InducedSubgraph(mustVertexIDs(t, 2, 0, 1, 1))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = induced.Graph.Close() })
	assertIDMapping(t, induced.Vertices,
		[]int{0, 1, 2, RemovedID, RemovedID, RemovedID, RemovedID, RemovedID},
		[]int{0, 1, 2},
	)

	edgeSubgraph, err := source.EdgeSubgraph(mustEdgeIDs(t, 5, 4, 5), true)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = edgeSubgraph.Graph.Close() })
	assertIDMapping(t, edgeSubgraph.Vertices,
		[]int{RemovedID, RemovedID, RemovedID, 0, 1, RemovedID, RemovedID, RemovedID},
		[]int{3, 4},
	)

	components, err := source.Decompose(DecomposeOptions{})
	if err != nil {
		t.Fatal(err)
	}
	closeGraphs(t, components)
	assertComponentShapesUnordered(t, components, [][2]int{{3, 4}, {2, 2}, {1, 0}, {2, 1}})

	simplified, err := induced.Graph.SimplifyInPlace(SimplifyOptions{RemoveParallel: true, RemoveLoops: true})
	if err != nil {
		t.Fatal(err)
	}
	if !simplified.EdgeMappingAvailable {
		t.Fatal("simplification edge mapping is unavailable")
	}
	assertIDMapping(t, simplified.Mapping.Edges,
		[]int{0, 1, 2, RemovedID}, []int{0, 1, 2},
	)
	converted, err := induced.Graph.ConvertToUndirectedInPlace(UndirectedConversionCollapse, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !converted.EdgeMappingAvailable {
		t.Fatal("undirected conversion edge mapping is unavailable")
	}
	assertGraphShape(t, induced.Graph, 3, 3, false)

	pathOperand := testGraphFromEdges(t, 3, []Edge{{0, 1}, {1, 2}}, false)
	union, err := induced.Graph.Union(pathOperand, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = union.Graph.Close() })
	diagonal := testGraphFromEdges(t, 3, []Edge{{0, 2}}, false)
	difference, err := union.Graph.Difference(diagonal)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = difference.Graph.Close() })
	assertGraphShape(t, difference.Graph, 3, 2, false)

	if err := source.Close(); err != nil {
		t.Fatal(err)
	}
	assertGraphShape(t, induced.Graph, 3, 3, false)
	assertGraphShape(t, edgeSubgraph.Graph, 2, 2, true)
	for _, component := range components {
		if _, err := component.VertexCount(); err != nil {
			t.Errorf("component after source close error = %v", err)
		}
	}
	induced.Vertices.OldToNew[0] = 99
	if induced.Vertices.OldToNew[0] != 99 {
		t.Error("induced mapping is not Go-owned")
	}

	points, err := difference.Graph.ArticulationPoints()
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(points, []int{1}) {
		t.Errorf("ArticulationPoints() = %v, want [1]", points)
	}
	bridges, err := difference.Graph.Bridges()
	if err != nil {
		t.Fatal(err)
	}
	sort.Ints(bridges)
	if !reflect.DeepEqual(bridges, []int{0, 1}) {
		t.Errorf("Bridges() = %v, want [0 1]", bridges)
	}
	biconnected, err := difference.Graph.BiconnectedComponents()
	if err != nil {
		t.Fatal(err)
	}
	if biconnected.Count != 2 || !reflect.DeepEqual(biconnected.ArticulationPoints, []int{1}) {
		t.Errorf("BiconnectedComponents() = %#v", biconnected)
	}
}

func TestMilestone5UndirectedEmptyAllAndUnavailableMapping(t *testing.T) {
	source := testGraphFromEdges(t, 6, []Edge{
		{0, 1}, {1, 0}, {1, 1}, {2, 3},
	}, false)
	allDeleted, err := source.Clone()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = allDeleted.Close() })

	identity, err := source.DeleteVertices(NoVertices())
	if err != nil {
		t.Fatal(err)
	}
	wantVertices, _ := identityIDMapping(6)
	wantEdges, _ := identityIDMapping(4)
	if !reflect.DeepEqual(identity, (GraphIDMapping{Vertices: wantVertices, Edges: wantEdges})) {
		t.Errorf("empty deletion mapping = %#v", identity)
	}

	allSubgraph, err := source.InducedSubgraph(AllVertices())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = allSubgraph.Graph.Close() })
	emptyEdges, err := source.EdgeSubgraph(NoEdges(), true)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = emptyEdges.Graph.Close() })
	assertGraphShape(t, emptyEdges.Graph, 0, 0, false)
	if emptyEdges.Vertices.NewToOld == nil {
		t.Fatal("empty edge-subgraph inverse mapping is nil")
	}

	components, err := source.Decompose(DecomposeOptions{})
	if err != nil {
		t.Fatal(err)
	}
	closeGraphs(t, components)
	assertComponentShapesUnordered(t, components, [][2]int{{2, 3}, {2, 1}, {1, 0}, {1, 0}})

	if _, err := source.SimplifyInPlace(SimplifyOptions{RemoveParallel: true, RemoveLoops: true}); err != nil {
		t.Fatal(err)
	}
	converted, err := source.ConvertToDirectedInPlace(DirectedConversionMutual)
	if err != nil {
		t.Fatal(err)
	}
	if converted.EdgeMappingAvailable || converted.Mapping.Edges.OldToNew == nil || converted.Mapping.Edges.NewToOld == nil {
		t.Errorf("mutual conversion availability/mapping = %#v", converted)
	}
	composition, err := source.Compose(source, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = composition.Graph.Close() })
	if composition.Edges == nil {
		t.Fatal("composition provenance is nil")
	}

	deletedAllMapping, err := allDeleted.DeleteVertices(AllVertices())
	if err != nil {
		t.Fatal(err)
	}
	assertGraphShape(t, allDeleted, 0, 0, false)
	if deletedAllMapping.Vertices.NewToOld == nil || deletedAllMapping.Edges.NewToOld == nil {
		t.Fatal("all-deletion mapping contains nil inverse slice")
	}

	empty := testGraphFromEdges(t, 0, nil, false)
	emptyComponents, err := empty.Decompose(DecomposeOptions{})
	if err != nil || emptyComponents == nil || len(emptyComponents) != 0 {
		t.Errorf("empty Decompose() = %v, %v, want non-nil empty, nil", emptyComponents, err)
	}

	if err := source.Close(); err != nil {
		t.Fatal(err)
	}
	assertGraphShape(t, allSubgraph.Graph, 6, 4, false)
	if _, err := source.Bridges(); !errors.Is(err, ErrClosed) {
		t.Errorf("closed Bridges() error = %v, want %v", err, ErrClosed)
	}
	if err := composition.Graph.Close(); err != nil {
		t.Fatal(err)
	}
	if err := composition.Graph.Close(); err != nil {
		t.Fatal(err)
	}
}

func mustVertexIDs(t *testing.T, ids ...int) VertexSelector {
	t.Helper()
	selector, err := VertexIDs(ids...)
	if err != nil {
		t.Fatal(err)
	}
	return selector
}

func mustEdgeIDs(t *testing.T, ids ...int) EdgeSelector {
	t.Helper()
	selector, err := EdgeIDs(ids...)
	if err != nil {
		t.Fatal(err)
	}
	return selector
}
