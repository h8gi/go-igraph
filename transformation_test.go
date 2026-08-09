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
			if err := graph.SimplifyInPlace(test.options); err != nil {
				t.Fatal(err)
			}
			assertTransformationGraph(t, graph, 4, true, test.want)
		})
	}
}

func TestSimplifyInPlaceUndirectedAndAllLoops(t *testing.T) {
	graph := transformationGraph(t, 3, []Edge{{0, 1}, {1, 0}, {2, 2}, {2, 2}}, false)
	if err := graph.SimplifyInPlace(SimplifyOptions{RemoveParallel: true}); err != nil {
		t.Fatal(err)
	}
	assertTransformationGraph(t, graph, 3, false, []Edge{{0, 1}, {2, 2}})
	if err := graph.SimplifyInPlace(SimplifyOptions{RemoveLoops: true}); err != nil {
		t.Fatal(err)
	}
	assertTransformationGraph(t, graph, 3, false, []Edge{{0, 1}})

	loops := transformationGraph(t, 2, []Edge{{0, 0}, {0, 0}, {1, 1}}, true)
	if err := loops.SimplifyInPlace(SimplifyOptions{RemoveLoops: true}); err != nil {
		t.Fatal(err)
	}
	assertTransformationGraph(t, loops, 2, true, []Edge{})
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
			if err := graph.ConvertToDirectedInPlace(test.mode); err != nil {
				t.Fatal(err)
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
			if err := graph.ConvertToUndirectedInPlace(test.mode); err != nil {
				t.Fatal(err)
			}
			assertTransformationGraph(t, graph, 3, false, test.want)
		})
	}
}

func TestDirectionConversionIdentityEmptyAndEdgeless(t *testing.T) {
	simpleEdges := []Edge{{0, 1}, {1, 2}}
	simple := transformationGraph(t, 3, simpleEdges, false)
	if err := simple.SimplifyInPlace(SimplifyOptions{RemoveParallel: true, RemoveLoops: true}); err != nil {
		t.Fatal(err)
	}
	assertTransformationGraph(t, simple, 3, false, simpleEdges)

	directedEdges := []Edge{{1, 0}, {1, 1}}
	directed := transformationGraph(t, 2, directedEdges, true)
	if err := directed.ConvertToDirectedInPlace(DirectedConversionMutual); err != nil {
		t.Fatal(err)
	}
	assertGraphState(t, directed, 2, true, directedEdges)

	undirectedEdges := []Edge{{1, 0}, {1, 1}}
	undirected := transformationGraph(t, 2, undirectedEdges, false)
	undirectedBefore := captureGraphState(t, undirected)
	if err := undirected.ConvertToUndirectedInPlace(UndirectedConversionMutual); err != nil {
		t.Fatal(err)
	}
	assertCapturedGraphState(t, undirected, undirectedBefore)

	for _, vertexCount := range []int{0, 3} {
		graph := transformationGraph(t, vertexCount, nil, false)
		if err := graph.SimplifyInPlace(SimplifyOptions{RemoveParallel: true, RemoveLoops: true}); err != nil {
			t.Fatal(err)
		}
		if err := graph.ConvertToDirectedInPlace(DirectedConversionAcyclic); err != nil {
			t.Fatal(err)
		}
		assertTransformationGraph(t, graph, vertexCount, true, []Edge{})
		if err := graph.ConvertToUndirectedInPlace(UndirectedConversionEach); err != nil {
			t.Fatal(err)
		}
		assertTransformationGraph(t, graph, vertexCount, false, []Edge{})
	}
}

func TestGraphTransformationsRejectInvalidModesAndClosedGraphs(t *testing.T) {
	graph := transformationGraph(t, 2, []Edge{{0, 1}}, false)
	before := captureGraphState(t, graph)
	if err := graph.ConvertToDirectedInPlace(DirectedConversionMode(99)); err == nil {
		t.Error("ConvertToDirectedInPlace(invalid) error = nil")
	}
	assertCapturedGraphState(t, graph, before)

	directed := transformationGraph(t, 2, []Edge{{0, 1}}, true)
	before = captureGraphState(t, directed)
	if err := directed.ConvertToUndirectedInPlace(UndirectedConversionMode(99)); err == nil {
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

func TestExecuteAtomicTransformationFailureCleanup(t *testing.T) {
	forced := errors.New("forced failure")
	tests := []struct {
		name        string
		cloneError  bool
		mutateError bool
		wantDestroy int
		wantCommit  int
	}{
		{name: "clone initialization", cloneError: true},
		{name: "upstream transformation", mutateError: true, wantDestroy: 1},
		{name: "success", wantCommit: 1},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			destroyed := 0
			committed := 0
			source := "original"
			replacement := "clone"
			err := executeAtomicTransformation(
				func() error {
					if test.cloneError {
						return forced
					}
					return nil
				},
				func() error {
					if test.mutateError {
						return forced
					}
					return nil
				},
				func() {
					destroyed++
					replacement = "destroyed"
				},
				func() {
					committed++
					source = replacement
				},
			)
			if test.cloneError || test.mutateError {
				if !errors.Is(err, forced) {
					t.Errorf("error = %v, want %v", err, forced)
				}
			} else if err != nil {
				t.Errorf("error = %v, want nil", err)
			}
			if destroyed != test.wantDestroy || committed != test.wantCommit {
				t.Errorf("destroy/commit = %d/%d, want %d/%d", destroyed, committed, test.wantDestroy, test.wantCommit)
			}
			wantSource := "original"
			if test.wantCommit == 1 {
				wantSource = "clone"
			}
			if source != wantSource {
				t.Errorf("source = %q, want %q", source, wantSource)
			}
		})
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
	if err := graph.SimplifyInPlace(SimplifyOptions{RemoveLoops: true}); !errors.Is(err, ErrClosed) {
		t.Errorf("SimplifyInPlace() error = %v, want %v", err, ErrClosed)
	}
	if err := graph.ConvertToDirectedInPlace(DirectedConversionArbitrary); !errors.Is(err, ErrClosed) {
		t.Errorf("ConvertToDirectedInPlace() error = %v, want %v", err, ErrClosed)
	}
	if err := graph.ConvertToUndirectedInPlace(UndirectedConversionEach); !errors.Is(err, ErrClosed) {
		t.Errorf("ConvertToUndirectedInPlace() error = %v, want %v", err, ErrClosed)
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
