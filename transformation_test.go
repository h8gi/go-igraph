package igraph

import (
	"errors"
	"fmt"
	"reflect"
	"sort"
	"testing"
)

func TestSimplifyInPlaceOptions(t *testing.T) {
	edges := []Edge{
		{0, 1}, {0, 1}, {1, 0}, {1, 1}, {1, 1}, {2, 2}, {2, 3},
	}
	tests := []struct {
		name    string
		options SimplifyOptions
		want    []Edge
	}{
		{name: "identity", want: edges},
		{
			name: "parallel only", options: SimplifyOptions{RemoveParallel: true},
			want: []Edge{{0, 1}, {1, 0}, {1, 1}, {2, 2}, {2, 3}},
		},
		{
			name: "loops only", options: SimplifyOptions{RemoveLoops: true},
			want: []Edge{{0, 1}, {0, 1}, {1, 0}, {2, 3}},
		},
		{
			name: "parallel and loops", options: SimplifyOptions{RemoveParallel: true, RemoveLoops: true},
			want: []Edge{{0, 1}, {1, 0}, {2, 3}},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			graph := transformationGraph(t, 4, edges, true)
			if _, err := graph.SimplifyInPlace(test.options); err != nil {
				t.Fatal(err)
			}
			assertTransformationGraph(t, graph, 4, true, test.want)
		})
	}
}

func TestSimplifyInPlaceUndirectedAndAllLoops(t *testing.T) {
	graph := transformationGraph(t, 3, []Edge{{0, 1}, {1, 0}, {2, 2}, {2, 2}}, false)
	if _, err := graph.SimplifyInPlace(SimplifyOptions{RemoveParallel: true}); err != nil {
		t.Fatal(err)
	}
	assertTransformationGraph(t, graph, 3, false, []Edge{{0, 1}, {2, 2}})
	if _, err := graph.SimplifyInPlace(SimplifyOptions{RemoveLoops: true}); err != nil {
		t.Fatal(err)
	}
	assertTransformationGraph(t, graph, 3, false, []Edge{{0, 1}})

	loops := transformationGraph(t, 2, []Edge{{0, 0}, {0, 0}, {1, 1}}, true)
	if _, err := loops.SimplifyInPlace(SimplifyOptions{RemoveLoops: true}); err != nil {
		t.Fatal(err)
	}
	assertTransformationGraph(t, loops, 2, true, []Edge{})
}

func TestSimplifyInPlaceMapping(t *testing.T) {
	edges := []Edge{{0, 1}, {0, 1}, {1, 0}, {1, 1}, {1, 1}, {2, 2}, {2, 3}}
	graph := transformationGraph(t, 4, edges, true)
	result, err := graph.SimplifyInPlace(SimplifyOptions{RemoveParallel: true, RemoveLoops: true})
	if err != nil {
		t.Fatal(err)
	}
	if !result.EdgeMappingAvailable {
		t.Fatal("edge mapping is unavailable")
	}
	assertIDMapping(t, result.Mapping.Vertices, []int{0, 1, 2, 3}, []int{0, 1, 2, 3})
	assertIDMapping(
		t, result.Mapping.Edges,
		[]int{0, 0, 1, RemovedID, RemovedID, RemovedID, 2},
		[]int{0, 2, 6},
	)
	if err := graph.Close(); err != nil {
		t.Fatal(err)
	}
	result.Mapping.Edges.OldToNew[0] = 99
	if result.Mapping.Edges.OldToNew[0] != 99 {
		t.Error("mapping is not mutable after graph closure")
	}
}

func TestConvertToDirectedInPlaceModes(t *testing.T) {
	edges := []Edge{{0, 1}, {0, 1}, {1, 1}, {2, 0}}
	tests := []struct {
		name      string
		mode      DirectedConversionMode
		wantEdges []Edge
		wantCount int
	}{
		{
			name: "mutual", mode: DirectedConversionMutual,
			wantEdges: []Edge{
				{0, 1}, {1, 0}, {0, 1}, {1, 0},
				{1, 1}, {1, 1}, {2, 0}, {0, 2},
			},
		},
		{
			name: "acyclic", mode: DirectedConversionAcyclic,
			wantEdges: []Edge{{0, 1}, {0, 1}, {1, 1}, {0, 2}},
		},
		{name: "arbitrary", mode: DirectedConversionArbitrary, wantCount: len(edges)},
		{name: "random", mode: DirectedConversionRandom, wantCount: len(edges)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			graph := transformationGraph(t, 3, edges, false)
			result, err := graph.ConvertToDirectedInPlace(test.mode)
			if err != nil {
				t.Fatal(err)
			}
			if test.mode == DirectedConversionMutual {
				assertUnavailableEdgeMapping(t, result)
			} else {
				assertIDMapping(t, result.Mapping.Edges, []int{0, 1, 2, 3}, []int{0, 1, 2, 3})
			}
			if test.wantEdges != nil {
				assertTransformationGraph(t, graph, 3, true, test.wantEdges)
			} else {
				got, err := graph.Edges()
				if err != nil {
					t.Fatal(err)
				}
				if directed, _ := graph.IsDirected(); !directed || len(got) != test.wantCount {
					t.Errorf("directed/count = %t/%d, want true/%d; edges=%v", directed, len(got), test.wantCount, got)
				}
				if normalized := normalizedEdges(got, false); !reflect.DeepEqual(normalized, normalizedEdges(edges, false)) {
					t.Errorf("undirected endpoint multiset = %v, want %v", normalized, normalizedEdges(edges, false))
				}
			}
		})
	}
}

func TestConvertToUndirectedInPlaceModes(t *testing.T) {
	edges := []Edge{{0, 1}, {0, 1}, {1, 0}, {2, 1}, {2, 2}, {2, 2}}
	tests := []struct {
		name string
		mode UndirectedConversionMode
		want []Edge
	}{
		{
			name: "each", mode: UndirectedConversionEach,
			want: []Edge{{0, 1}, {0, 1}, {0, 1}, {1, 2}, {2, 2}, {2, 2}},
		},
		{
			name: "collapse", mode: UndirectedConversionCollapse,
			want: []Edge{{0, 1}, {1, 2}, {2, 2}},
		},
		{
			name: "mutual", mode: UndirectedConversionMutual,
			want: []Edge{{0, 1}, {2, 2}, {2, 2}},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			graph := transformationGraph(t, 3, edges, true)
			result, err := graph.ConvertToUndirectedInPlace(test.mode, nil)
			if err != nil {
				t.Fatal(err)
			}
			if !result.EdgeMappingAvailable {
				t.Fatal("edge mapping is unavailable")
			}
			switch test.mode {
			case UndirectedConversionEach:
				assertIDMapping(t, result.Mapping.Edges, []int{0, 1, 2, 3, 4, 5}, []int{0, 1, 2, 3, 4, 5})
			case UndirectedConversionCollapse:
				assertIDMapping(t, result.Mapping.Edges, []int{0, 0, 0, 1, 2, 2}, []int{0, 3, 4})
			case UndirectedConversionMutual:
				assertIDMapping(t, result.Mapping.Edges, []int{0, RemovedID, 0, RemovedID, 1, 2}, []int{0, 4, 5})
			}
			assertTransformationGraph(t, graph, 3, false, test.want)
		})
	}
}

func TestDirectionConversionIdentityEmptyAndEdgeless(t *testing.T) {
	simpleEdges := []Edge{{0, 1}, {1, 2}}
	simple := transformationGraph(t, 3, simpleEdges, false)
	if _, err := simple.SimplifyInPlace(SimplifyOptions{RemoveParallel: true, RemoveLoops: true}); err != nil {
		t.Fatal(err)
	}
	assertTransformationGraph(t, simple, 3, false, simpleEdges)
	identity, err := simple.SimplifyInPlace(SimplifyOptions{})
	if err != nil {
		t.Fatal(err)
	}
	assertIDMapping(t, identity.Mapping.Edges, []int{0, 1}, []int{0, 1})

	directedEdges := []Edge{{1, 0}, {1, 1}}
	directed := transformationGraph(t, 2, directedEdges, true)
	identity, err = directed.ConvertToDirectedInPlace(DirectedConversionMutual)
	if err != nil {
		t.Fatal(err)
	}
	if !identity.EdgeMappingAvailable {
		t.Error("already-directed no-op mapping is unavailable")
	}
	assertIDMapping(t, identity.Mapping.Edges, []int{0, 1}, []int{0, 1})
	assertGraphState(t, directed, 2, true, directedEdges)

	undirectedEdges := []Edge{{1, 0}, {1, 1}}
	undirected := transformationGraph(t, 2, undirectedEdges, false)
	undirectedBefore := captureGraphState(t, undirected)
	if _, err := undirected.ConvertToUndirectedInPlace(UndirectedConversionMutual, nil); err != nil {
		t.Fatal(err)
	}
	assertCapturedGraphState(t, undirected, undirectedBefore)

	for _, vertexCount := range []int{0, 3} {
		graph := transformationGraph(t, vertexCount, nil, false)
		if _, err := graph.SimplifyInPlace(SimplifyOptions{RemoveParallel: true, RemoveLoops: true}); err != nil {
			t.Fatal(err)
		}
		if _, err := graph.ConvertToDirectedInPlace(DirectedConversionAcyclic); err != nil {
			t.Fatal(err)
		}
		assertTransformationGraph(t, graph, vertexCount, true, []Edge{})
		if _, err := graph.ConvertToUndirectedInPlace(UndirectedConversionEach, nil); err != nil {
			t.Fatal(err)
		}
		assertTransformationGraph(t, graph, vertexCount, false, []Edge{})
	}

	empty := transformationGraph(t, 0, nil, false)
	emptyResult, err := empty.ConvertToDirectedInPlace(DirectedConversionMutual)
	if err != nil {
		t.Fatal(err)
	}
	if !emptyResult.EdgeMappingAvailable {
		t.Error("empty mutual conversion mapping is unavailable")
	}
	assertIDMapping(t, emptyResult.Mapping.Edges, []int{}, []int{})
}

func TestGraphTransformationsRejectInvalidModesAndClosedGraphs(t *testing.T) {
	graph := transformationGraph(t, 2, []Edge{{0, 1}}, false)
	before := captureGraphState(t, graph)
	if _, err := graph.ConvertToDirectedInPlace(DirectedConversionMode(99)); err == nil {
		t.Error("ConvertToDirectedInPlace(invalid) error = nil")
	}
	assertCapturedGraphState(t, graph, before)

	directed := transformationGraph(t, 2, []Edge{{0, 1}}, true)
	before = captureGraphState(t, directed)
	if _, err := directed.ConvertToUndirectedInPlace(UndirectedConversionMode(99), nil); err == nil {
		t.Error("ConvertToUndirectedInPlace(invalid) error = nil")
	}
	assertCapturedGraphState(t, directed, before)

	if err := graph.Close(); err != nil {
		t.Fatal(err)
	}
	assertTransformationClosed(t, graph)
	var nilGraph *Graph
	assertTransformationClosed(t, nilGraph)
}

func TestGraphTransformationAtomicReceiver(t *testing.T) {
	forced := errors.New("forced failure")
	edges := []Edge{{0, 1}, {0, 1}, {1, 1}, {2, 0}}
	tests := []struct {
		name  string
		stage graphTransformationStage
	}{
		{name: "clone initialization", stage: graphTransformationAtClone},
		{name: "upstream transformation", stage: graphTransformationAtTransform},
		{name: "post-transformation conversion", stage: graphTransformationAfterTransform},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			graph := transformationGraph(t, 3, edges, true)
			before := captureGraphState(t, graph)
			_, err := graph.simplifyInPlace(
				SimplifyOptions{RemoveParallel: true, RemoveLoops: true},
				func(stage graphTransformationStage) error {
					if stage == test.stage {
						return forced
					}
					return nil
				},
			)
			if !errors.Is(err, forced) {
				t.Errorf("error = %v, want %v", err, forced)
			}
			assertCapturedGraphState(t, graph, before)
		})
	}

	graph := transformationGraph(t, 3, edges, true)
	result, err := graph.simplifyInPlace(
		SimplifyOptions{RemoveParallel: true, RemoveLoops: true}, nil,
	)
	if err != nil {
		t.Fatal(err)
	}
	assertTransformationGraph(t, graph, 3, true, []Edge{{0, 1}, {2, 0}})
	assertIDMapping(t, result.Mapping.Edges, []int{0, 0, RemovedID, 1}, []int{0, 3})
}

func TestEdgeTransformationMappingValidation(t *testing.T) {
	tests := []struct {
		name string
		run  func() error
	}{
		{
			name: "identity edge count mismatch",
			run: func() error {
				_, err := identityEndpointEdgeMapping([]Edge{{0, 1}}, nil, true)
				return err
			},
		},
		{
			name: "identity endpoint mismatch",
			run: func() error {
				_, err := identityEndpointEdgeMapping([]Edge{{0, 1}}, []Edge{{1, 0}}, true)
				return err
			},
		},
		{
			name: "stable filter source without result",
			run: func() error {
				_, err := stableFilteredEdgeMapping([]Edge{{0, 1}}, nil, true, nil)
				return err
			},
		},
		{
			name: "stable filter result without source",
			run: func() error {
				_, err := stableFilteredEdgeMapping(nil, []Edge{{0, 1}}, true, nil)
				return err
			},
		},
		{
			name: "many-to-one duplicate result",
			run: func() error {
				_, err := manyToOneEdgeMapping(nil, []Edge{{0, 1}, {0, 1}}, true, nil)
				return err
			},
		},
		{
			name: "many-to-one source without result",
			run: func() error {
				_, err := manyToOneEdgeMapping([]Edge{{0, 1}}, []Edge{{1, 2}}, true, nil)
				return err
			},
		},
		{
			name: "many-to-one result without source",
			run: func() error {
				_, err := manyToOneEdgeMapping([]Edge{{0, 1}}, []Edge{{0, 1}, {1, 2}}, true, nil)
				return err
			},
		},
		{
			name: "mutual loop mismatch",
			run: func() error {
				_, err := mutualEdgeMapping([]Edge{{0, 0}}, nil)
				return err
			},
		},
		{
			name: "mutual reciprocal mismatch",
			run: func() error {
				_, err := mutualEdgeMapping([]Edge{{0, 1}, {1, 0}}, nil)
				return err
			},
		},
		{
			name: "mutual result without source",
			run: func() error {
				_, err := mutualEdgeMapping(nil, []Edge{{0, 1}})
				return err
			},
		},
		{
			name: "mapping out of range",
			run: func() error {
				_, err := completeEdgeMapping([]int{1}, 1)
				return err
			},
		},
		{
			name: "mapping result without source",
			run: func() error {
				_, err := completeEdgeMapping([]int{RemovedID}, 1)
				return err
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := test.run(); err == nil {
				t.Error("error = nil")
			}
		})
	}
}

func assertUnavailableEdgeMapping(t *testing.T, result GraphTransformationResult) {
	t.Helper()
	if result.EdgeMappingAvailable {
		t.Error("EdgeMappingAvailable = true, want false")
	}
	if result.Mapping.Edges.OldToNew == nil || result.Mapping.Edges.NewToOld == nil ||
		len(result.Mapping.Edges.OldToNew) != 0 || len(result.Mapping.Edges.NewToOld) != 0 {
		t.Errorf("unavailable edge mapping = %#v, want non-nil empty slices", result.Mapping.Edges)
	}
	if result.Mapping.Vertices.OldToNew == nil || result.Mapping.Vertices.NewToOld == nil {
		t.Errorf("vertex identity mapping = %#v, want non-nil slices", result.Mapping.Vertices)
	}
}

func transformationGraph(t *testing.T, vertexCount int, edges []Edge, directed bool) *Graph {
	t.Helper()
	graph, err := NewGraphFromEdges(vertexCount, edges, directed)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = graph.Close() })
	return graph
}

func assertTransformationClosed(t *testing.T, graph *Graph) {
	t.Helper()
	if _, err := graph.SimplifyInPlace(SimplifyOptions{RemoveLoops: true}); !errors.Is(err, ErrClosed) {
		t.Errorf("SimplifyInPlace() error = %v, want %v", err, ErrClosed)
	}
	if _, err := graph.ConvertToDirectedInPlace(DirectedConversionArbitrary); !errors.Is(err, ErrClosed) {
		t.Errorf("ConvertToDirectedInPlace() error = %v, want %v", err, ErrClosed)
	}
	if _, err := graph.ConvertToUndirectedInPlace(UndirectedConversionEach, nil); !errors.Is(err, ErrClosed) {
		t.Errorf("ConvertToUndirectedInPlace() error = %v, want %v", err, ErrClosed)
	}
}

func TestTransformationInvalidModes(t *testing.T) {
	graph := transformationGraph(t, 2, []Edge{{0, 1}}, false)
	if _, err := graph.ConvertToDirectedInPlace(DirectedConversionMode(99)); err == nil {
		t.Errorf("ConvertToDirectedInPlace(invalid) error = nil")
	}

	directed := transformationGraph(t, 2, []Edge{{0, 1}}, true)
	if _, err := directed.ConvertToUndirectedInPlace(UndirectedConversionMode(99), nil); err == nil {
		t.Errorf("ConvertToUndirectedInPlace(invalid) error = nil")
	}
}

func assertTransformationGraph(t *testing.T, graph *Graph, vertexCount int, directed bool, want []Edge) {
	t.Helper()
	gotVertexCount, err := graph.VertexCount()
	if err != nil {
		t.Fatal(err)
	}
	gotDirected, err := graph.IsDirected()
	if err != nil {
		t.Fatal(err)
	}
	gotEdges, err := graph.Edges()
	if err != nil {
		t.Fatal(err)
	}
	if gotVertexCount != vertexCount || gotDirected != directed {
		t.Errorf("vertices/directed = %d/%t, want %d/%t", gotVertexCount, gotDirected, vertexCount, directed)
	}
	if got, want := normalizedEdges(gotEdges, directed), normalizedEdges(want, directed); !reflect.DeepEqual(got, want) {
		t.Errorf("edges = %v, want %v", gotEdges, want)
	}
}

func normalizedEdges(edges []Edge, directed bool) []string {
	result := make([]string, len(edges))
	for index, edge := range edges {
		from, to := edge.From, edge.To
		if !directed && from > to {
			from, to = to, from
		}
		result[index] = fmt.Sprintf("%d:%d", from, to)
	}
	sort.Strings(result)
	return result
}
