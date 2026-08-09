package igraph

import (
	"errors"
	"reflect"
	"sort"
	"testing"
)

func TestInducedSubgraphMappingsAndOwnership(t *testing.T) {
	source := testGraphFromEdges(t, 5, []Edge{
		{0, 1}, {1, 2}, {2, 0}, {1, 1}, {0, 1}, {3, 1}, {4, 4},
	}, true)
	selector, err := VertexIDs(3, 1, 0, 1)
	if err != nil {
		t.Fatal(err)
	}
	result, err := source.InducedSubgraph(selector)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = result.Graph.Close() })
	assertGraphShape(t, result.Graph, 3, 4, true)
	assertEdgesEqual(t, result.Graph, []Edge{{0, 1}, {1, 1}, {0, 1}, {2, 1}})
	wantMapping := IDMapping{
		OldToNew: []int{0, 1, RemovedID, 2, RemovedID},
		NewToOld: []int{0, 1, 3},
	}
	if !reflect.DeepEqual(result.Vertices, wantMapping) {
		t.Errorf("Vertices = %#v, want %#v", result.Vertices, wantMapping)
	}

	if err := source.Close(); err != nil {
		t.Fatal(err)
	}
	assertGraphShape(t, result.Graph, 3, 4, true)
	result.Vertices.OldToNew[0] = 99
}

func TestInducedSubgraphEmptyAndIdentity(t *testing.T) {
	source := testGraphFromEdges(t, 3, []Edge{{0, 1}, {1, 2}}, false)
	empty, err := source.InducedSubgraph(NoVertices())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = empty.Graph.Close() })
	assertGraphShape(t, empty.Graph, 0, 0, false)
	if want := (IDMapping{OldToNew: []int{RemovedID, RemovedID, RemovedID}, NewToOld: []int{}}); !reflect.DeepEqual(empty.Vertices, want) {
		t.Errorf("empty mapping = %#v, want %#v", empty.Vertices, want)
	}
	if empty.Vertices.OldToNew == nil || empty.Vertices.NewToOld == nil {
		t.Fatal("empty induced-subgraph mapping contains a nil slice")
	}

	identity, err := source.InducedSubgraph(AllVertices())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = identity.Graph.Close() })
	wantIdentity, _ := identityIDMapping(3)
	if !reflect.DeepEqual(identity.Vertices, wantIdentity) {
		t.Errorf("identity mapping = %#v, want %#v", identity.Vertices, wantIdentity)
	}
	assertEdgesEqual(t, identity.Graph, []Edge{{0, 1}, {1, 2}})
}

func TestEdgeSubgraphSelectionAndIsolatedVertices(t *testing.T) {
	source := testGraphFromEdges(t, 5, []Edge{
		{0, 1}, {1, 2}, {2, 0}, {1, 1}, {0, 1}, {3, 1}, {4, 4},
	}, true)
	selector, err := EdgeIDs(5, 0, 5, 3)
	if err != nil {
		t.Fatal(err)
	}

	retained, err := source.EdgeSubgraph(selector, false)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = retained.Graph.Close() })
	assertGraphShape(t, retained.Graph, 5, 3, true)
	assertEdgesEqual(t, retained.Graph, []Edge{{0, 1}, {1, 1}, {3, 1}})
	wantIdentity, _ := identityIDMapping(5)
	if !reflect.DeepEqual(retained.Vertices, wantIdentity) {
		t.Errorf("retained mapping = %#v, want %#v", retained.Vertices, wantIdentity)
	}

	compacted, err := source.EdgeSubgraph(selector, true)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = compacted.Graph.Close() })
	assertGraphShape(t, compacted.Graph, 3, 3, true)
	assertEdgesEqual(t, compacted.Graph, []Edge{{0, 1}, {1, 1}, {2, 1}})
	wantCompacted := IDMapping{
		OldToNew: []int{0, 1, RemovedID, 2, RemovedID},
		NewToOld: []int{0, 1, 3},
	}
	if !reflect.DeepEqual(compacted.Vertices, wantCompacted) {
		t.Errorf("compacted mapping = %#v, want %#v", compacted.Vertices, wantCompacted)
	}

	if err := source.Close(); err != nil {
		t.Fatal(err)
	}
	assertGraphShape(t, compacted.Graph, 3, 3, true)
	compacted.Vertices.OldToNew[0] = 99
	retained.Vertices.NewToOld[0] = 99
}

func TestEdgeSubgraphEmptySelection(t *testing.T) {
	source := testGraphFromEdges(t, 3, []Edge{{0, 1}}, false)
	for _, deleteVertices := range []bool{false, true} {
		result, err := source.EdgeSubgraph(NoEdges(), deleteVertices)
		if err != nil {
			t.Fatalf("EdgeSubgraph(delete=%t) error = %v", deleteVertices, err)
		}
		t.Cleanup(func() { _ = result.Graph.Close() })
		wantVertices := 3
		if deleteVertices {
			wantVertices = 0
		}
		assertGraphShape(t, result.Graph, wantVertices, 0, false)
		wantMapping, _ := identityIDMapping(3)
		if deleteVertices {
			wantMapping = IDMapping{
				OldToNew: []int{RemovedID, RemovedID, RemovedID},
				NewToOld: []int{},
			}
		}
		if !reflect.DeepEqual(result.Vertices, wantMapping) {
			t.Errorf("mapping(delete=%t) = %#v, want %#v", deleteVertices, result.Vertices, wantMapping)
		}
		if result.Vertices.OldToNew == nil || result.Vertices.NewToOld == nil {
			t.Fatalf("mapping(delete=%t) contains a nil slice", deleteVertices)
		}
	}
}

func TestEdgeSubgraphAllEdgesLoopsParallelAndUndirected(t *testing.T) {
	source := testGraphFromEdges(t, 5, []Edge{
		{3, 1}, {1, 3}, {3, 3}, {0, 1}, {0, 1},
	}, false)
	result, err := source.EdgeSubgraph(AllEdges(), true)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = result.Graph.Close() })
	assertGraphShape(t, result.Graph, 3, 5, false)
	assertEdgesEqual(t, result.Graph, []Edge{{1, 2}, {1, 2}, {2, 2}, {0, 1}, {0, 1}})
	wantMapping := IDMapping{
		OldToNew: []int{0, 1, RemovedID, 2, RemovedID},
		NewToOld: []int{0, 1, 3},
	}
	if !reflect.DeepEqual(result.Vertices, wantMapping) {
		t.Errorf("Vertices = %#v, want %#v", result.Vertices, wantMapping)
	}
}

func TestEdgeSubgraphSingleUndirectedEdgeUsesExactSourceMapping(t *testing.T) {
	source := testGraphFromEdges(t, 5, []Edge{{3, 1}}, false)
	result, err := source.EdgeSubgraph(AllEdges(), true)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = result.Graph.Close() })
	assertGraphShape(t, result.Graph, 2, 1, false)
	assertEdgesEqual(t, result.Graph, []Edge{{0, 1}})
	wantMapping := IDMapping{
		OldToNew: []int{RemovedID, 0, RemovedID, 1, RemovedID},
		NewToOld: []int{1, 3},
	}
	if !reflect.DeepEqual(result.Vertices, wantMapping) {
		t.Errorf("Vertices = %#v, want exact upstream mapping %#v", result.Vertices, wantMapping)
	}
}

func TestSubgraphSelectorsRejectInvalidAndClosedGraph(t *testing.T) {
	graph := testGraphFromEdges(t, 2, []Edge{{0, 1}}, false)
	invalidVertices, _ := VertexIDs(2)
	if result, err := graph.InducedSubgraph(invalidVertices); err == nil || result.Graph != nil {
		t.Errorf("InducedSubgraph(invalid) = %#v, %v, want zero, error", result, err)
	}
	invalidEdges, _ := EdgeIDs(1)
	if result, err := graph.EdgeSubgraph(invalidEdges, false); err == nil || result.Graph != nil {
		t.Errorf("EdgeSubgraph(invalid) = %#v, %v, want zero, error", result, err)
	}
	missingPair, _ := EdgePairs([]Edge{{0, 0}}, false)
	if result, err := graph.EdgeSubgraph(missingPair, false); err == nil || result.Graph != nil {
		t.Errorf("EdgeSubgraph(missing pair) = %#v, %v, want zero, error", result, err)
	}
	if err := graph.Close(); err != nil {
		t.Fatal(err)
	}
	assertSubgraphsClosed(t, graph)
	assertSubgraphsClosed(t, nil)
}

func TestDecomposeFilteringLimitsAndOrdering(t *testing.T) {
	graph := testGraphFromEdges(t, 6, []Edge{
		{0, 1}, {1, 2}, {2, 0}, {3, 4},
	}, false)

	all, err := graph.Decompose(DecomposeOptions{})
	if err != nil {
		t.Fatal(err)
	}
	closeGraphs(t, all)
	assertComponentShapes(t, all, [][2]int{{3, 3}, {2, 1}, {1, 0}})

	filtered, err := graph.Decompose(DecomposeOptions{MinimumVertices: 2})
	if err != nil {
		t.Fatal(err)
	}
	closeGraphs(t, filtered)
	assertComponentShapes(t, filtered, [][2]int{{3, 3}, {2, 1}})

	limited, err := graph.Decompose(DecomposeOptions{MinimumVertices: 2, MaximumComponents: 1})
	if err != nil {
		t.Fatal(err)
	}
	closeGraphs(t, limited)
	assertComponentShapes(t, limited, [][2]int{{3, 3}})

	filteredAll, err := graph.Decompose(DecomposeOptions{MinimumVertices: 4})
	if err != nil {
		t.Fatal(err)
	}
	if filteredAll == nil || len(filteredAll) != 0 {
		t.Errorf("fully filtered Decompose() = %v, want non-nil empty", filteredAll)
	}
}

func TestDecomposeStrongDirectedAndIndependentOwnership(t *testing.T) {
	graph := testGraphFromEdges(t, 5, []Edge{
		{0, 1}, {1, 0}, {1, 2}, {2, 3}, {3, 2},
	}, true)
	weak, err := graph.Decompose(DecomposeOptions{Connectedness: ConnectednessWeak})
	if err != nil {
		t.Fatal(err)
	}
	closeGraphs(t, weak)
	assertComponentShapesUnordered(t, weak, [][2]int{{4, 5}, {1, 0}})

	components, err := graph.Decompose(DecomposeOptions{Connectedness: ConnectednessStrong})
	if err != nil {
		t.Fatal(err)
	}
	closeGraphs(t, components)
	assertComponentShapesUnordered(t, components, [][2]int{{2, 2}, {2, 2}, {1, 0}})
	if err := graph.Close(); err != nil {
		t.Fatal(err)
	}
	if err := components[0].Close(); err != nil {
		t.Fatal(err)
	}
	if err := components[0].Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := components[1].VertexCount(); err != nil {
		t.Errorf("sibling after source/first close error = %v", err)
	}
}

func TestDecomposeEmptyInvalidAndClosed(t *testing.T) {
	empty := testGraphFromEdges(t, 0, nil, false)
	components, err := empty.Decompose(DecomposeOptions{})
	if err != nil || components == nil || len(components) != 0 {
		t.Errorf("empty Decompose() = %v, %v, want non-nil empty, nil", components, err)
	}
	invalid := []DecomposeOptions{
		{Connectedness: ConnectednessMode(99)},
		{MinimumVertices: -1},
		{MaximumComponents: -1},
	}
	for _, options := range invalid {
		if result, err := empty.Decompose(options); err == nil || result != nil {
			t.Errorf("Decompose(%#v) = %v, %v, want nil, error", options, result, err)
		}
	}
	if err := empty.Close(); err != nil {
		t.Fatal(err)
	}
	if result, err := empty.Decompose(DecomposeOptions{}); !errors.Is(err, ErrClosed) || result != nil {
		t.Errorf("closed Decompose() = %v, %v", result, err)
	}
	var nilGraph *Graph
	if result, err := nilGraph.Decompose(DecomposeOptions{}); !errors.Is(err, ErrClosed) || result != nil {
		t.Errorf("nil Decompose() = %v, %v", result, err)
	}
}

func TestCollectInducedSubgraphCleansEveryFailure(t *testing.T) {
	forced := errors.New("forced failure")
	tests := []struct {
		name           string
		failInitAt     int
		queryError     bool
		queryWithGraph bool
		failSliceAt    int
		badMapping     bool
		wantVectors    int
		wantGraphs     int
	}{
		{name: "first vector initialization", failInitAt: 1},
		{name: "second vector initialization", failInitAt: 2, wantVectors: 1},
		{name: "upstream query", queryError: true, wantVectors: 2},
		{name: "query with graph and error", queryError: true, queryWithGraph: true, wantVectors: 2, wantGraphs: 1},
		{name: "forward conversion", failSliceAt: 1, wantVectors: 2, wantGraphs: 1},
		{name: "inverse conversion", failSliceAt: 2, wantVectors: 2, wantGraphs: 1},
		{name: "inconsistent mapping", badMapping: true, wantVectors: 2, wantGraphs: 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			initCalls := 0
			sliceCalls := 0
			closedVectors := 0
			closedGraphs := 0
			result, err := collectInducedSubgraph(1, inducedSubgraphOperations{
				newVector: func() (*intVector, error) {
					initCalls++
					if initCalls == tt.failInitAt {
						return nil, forced
					}
					return &intVector{}, nil
				},
				closeVector: func(*intVector) { closedVectors++ },
				query: func(*intVector, *intVector) (*Graph, int, error) {
					var graph *Graph
					if !tt.queryError || tt.queryWithGraph {
						graph = &Graph{}
					}
					if tt.queryError {
						return graph, 1, forced
					}
					return graph, 1, nil
				},
				vectorSlice: func(*intVector) ([]int, error) {
					sliceCalls++
					if sliceCalls == tt.failSliceAt {
						return nil, forced
					}
					if tt.badMapping && sliceCalls == 2 {
						return []int{RemovedID}, nil
					}
					return []int{0}, nil
				},
				closeGraph: func(*Graph) { closedGraphs++ },
			})
			if result.Graph != nil || err == nil {
				t.Errorf("collectInducedSubgraph() = %#v, %v, want zero, error", result, err)
			}
			if tt.failInitAt > 0 || tt.queryError || tt.failSliceAt > 0 {
				if !errors.Is(err, forced) {
					t.Errorf("error = %v, want %v", err, forced)
				}
			}
			if closedVectors != tt.wantVectors || closedGraphs != tt.wantGraphs {
				t.Errorf("closed vectors/graphs = %d/%d, want %d/%d", closedVectors, closedGraphs, tt.wantVectors, tt.wantGraphs)
			}
		})
	}
}

func TestCollectEdgeSubgraphCleansEveryFailure(t *testing.T) {
	forced := errors.New("forced failure")
	mapping, _ := identityIDMapping(1)
	for _, stage := range []string{
		"query", "nil graph", "identity mapping", "isolated selector",
		"vertex deletion", "source mapping count", "vertex count", "mapping count",
	} {
		t.Run(stage, func(t *testing.T) {
			closed := 0
			deleteVertices := stage != "identity mapping"
			operations := edgeSubgraphOperations{
				query:           func() (*Graph, error) { return &Graph{}, nil },
				identityMapping: func(int) (IDMapping, error) { return mapping, nil },
				isolatedVertices: func(*Graph) (VertexSelector, error) {
					return NoVertices(), nil
				},
				deleteVertices: func(*Graph, VertexSelector) (GraphIDMapping, error) {
					return GraphIDMapping{Vertices: mapping}, nil
				},
				vertexCount: func(*Graph) (int, error) { return 1, nil },
				closeGraph:  func(*Graph) { closed++ },
			}
			switch stage {
			case "query":
				operations.query = func() (*Graph, error) { return &Graph{}, forced }
			case "nil graph":
				operations.query = func() (*Graph, error) { return nil, nil }
			case "identity mapping":
				operations.identityMapping = func(int) (IDMapping, error) { return IDMapping{}, forced }
			case "isolated selector":
				operations.isolatedVertices = func(*Graph) (VertexSelector, error) { return VertexSelector{}, forced }
			case "vertex deletion":
				operations.deleteVertices = func(*Graph, VertexSelector) (GraphIDMapping, error) {
					return GraphIDMapping{}, forced
				}
			case "source mapping count":
				operations.deleteVertices = func(*Graph, VertexSelector) (GraphIDMapping, error) {
					return GraphIDMapping{Vertices: IDMapping{OldToNew: []int{}, NewToOld: []int{0}}}, nil
				}
			case "vertex count":
				operations.vertexCount = func(*Graph) (int, error) { return 0, forced }
			case "mapping count":
				operations.vertexCount = func(*Graph) (int, error) { return 2, nil }
			}

			result, err := collectEdgeSubgraph(1, deleteVertices, operations)
			if result.Graph != nil || err == nil {
				t.Errorf("collectEdgeSubgraph() = %#v, %v, want zero, error", result, err)
			}
			wantClosed := 1
			if stage == "nil graph" {
				wantClosed = 0
			}
			if closed != wantClosed {
				t.Errorf("close count = %d, want %d", closed, wantClosed)
			}
		})
	}
}

func TestCollectDecompositionFailuresAndPartialExtraction(t *testing.T) {
	forced := errors.New("forced failure")
	for _, stage := range []string{"initialize", "query", "take"} {
		t.Run(stage, func(t *testing.T) {
			closed := 0
			graphs, err := collectDecomposition(decompositionOperations{
				newList: func() (*graphList, error) {
					if stage == "initialize" {
						return nil, forced
					}
					return &graphList{}, nil
				},
				closeList: func(*graphList) { closed++ },
				query: func(*graphList) error {
					if stage == "query" {
						return forced
					}
					return nil
				},
				takeGraphs: func(*graphList) ([]*Graph, error) {
					if stage == "take" {
						return nil, forced
					}
					return []*Graph{}, nil
				},
			})
			if graphs != nil || !errors.Is(err, forced) {
				t.Errorf("collectDecomposition() = %v, %v", graphs, err)
			}
			wantClosed := 1
			if stage == "initialize" {
				wantClosed = 0
			}
			if closed != wantClosed {
				t.Errorf("close count = %d, want %d", closed, wantClosed)
			}
		})
	}

	t.Run("query after population", func(t *testing.T) {
		sources := []*Graph{
			testGraphFromEdges(t, 1, nil, false),
			testGraphFromEdges(t, 2, []Edge{{0, 1}}, false),
		}
		list, err := newGraphListFromCopies(sources)
		if err != nil {
			t.Fatal(err)
		}
		takeCalled := false
		graphs, err := collectDecomposition(decompositionOperations{
			newList:   func() (*graphList, error) { return list, nil },
			closeList: func(list *graphList) { list.close() },
			query:     func(*graphList) error { return forced },
			takeGraphs: func(*graphList) ([]*Graph, error) {
				takeCalled = true
				return nil, nil
			},
		})
		if graphs != nil || !errors.Is(err, forced) || takeCalled || list.initialized {
			t.Errorf("populated query failure = %v, %v, take=%t, initialized=%t", graphs, err, takeCalled, list.initialized)
		}
	})

	for failAt := 0; failAt < 3; failAt++ {
		t.Run("partial extraction", func(t *testing.T) {
			sources := []*Graph{
				testGraphFromEdges(t, 1, nil, false),
				testGraphFromEdges(t, 2, []Edge{{0, 1}}, false),
				testGraphFromEdges(t, 3, []Edge{{0, 1}, {1, 2}}, false),
			}
			list, err := newGraphListFromCopies(sources)
			if err != nil {
				t.Fatal(err)
			}
			var adopted []*Graph
			graphs, err := collectDecomposition(decompositionOperations{
				newList:   func() (*graphList, error) { return list, nil },
				closeList: func(list *graphList) { list.close() },
				query:     func(*graphList) error { return nil },
				takeGraphs: func(list *graphList) ([]*Graph, error) {
					return list.takeGraphsWithHooks(graphListExtractionHooks{
						beforeAdopt: func(index int) error {
							if index == failAt {
								return forced
							}
							return nil
						},
						afterAdopt: func(_ int, graph *Graph) error {
							adopted = append(adopted, graph)
							return nil
						},
					})
				},
			})
			if graphs != nil || !errors.Is(err, forced) {
				t.Errorf("partial collectDecomposition() = %v, %v", graphs, err)
			}
			for _, graph := range adopted {
				if _, err := graph.VertexCount(); !errors.Is(err, ErrClosed) {
					t.Errorf("adopted graph error = %v, want %v", err, ErrClosed)
				}
			}
		})
	}
}

func assertSubgraphsClosed(t *testing.T, graph *Graph) {
	t.Helper()
	if result, err := graph.InducedSubgraph(AllVertices()); !errors.Is(err, ErrClosed) || result.Graph != nil {
		t.Errorf("closed InducedSubgraph() = %#v, %v", result, err)
	}
	if result, err := graph.EdgeSubgraph(AllEdges(), false); !errors.Is(err, ErrClosed) || result.Graph != nil {
		t.Errorf("closed EdgeSubgraph() = %#v, %v", result, err)
	}
}

func assertEdgesEqual(t *testing.T, graph *Graph, want []Edge) {
	t.Helper()
	got, err := graph.Edges()
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Edges() = %v, want %v", got, want)
	}
}

func assertComponentShapes(t *testing.T, graphs []*Graph, want [][2]int) {
	t.Helper()
	if len(graphs) != len(want) {
		t.Fatalf("component count = %d, want %d", len(graphs), len(want))
	}
	for index, graph := range graphs {
		assertGraphShape(t, graph, want[index][0], want[index][1], false)
	}
}

func assertComponentShapesUnordered(t *testing.T, graphs []*Graph, want [][2]int) {
	t.Helper()
	got := make([][2]int, len(graphs))
	for index, graph := range graphs {
		vertices, err := graph.VertexCount()
		if err != nil {
			t.Fatal(err)
		}
		edges, err := graph.EdgeCount()
		if err != nil {
			t.Fatal(err)
		}
		got[index] = [2]int{vertices, edges}
	}
	less := func(values [][2]int, i, j int) bool {
		if values[i][0] != values[j][0] {
			return values[i][0] < values[j][0]
		}
		return values[i][1] < values[j][1]
	}
	sort.Slice(got, func(i, j int) bool { return less(got, i, j) })
	sort.Slice(want, func(i, j int) bool { return less(want, i, j) })
	if !reflect.DeepEqual(got, want) {
		t.Errorf("component shapes = %v, want %v", got, want)
	}
}

func closeGraphs(t *testing.T, graphs []*Graph) {
	t.Helper()
	for _, graph := range graphs {
		graph := graph
		t.Cleanup(func() { _ = graph.Close() })
	}
}
