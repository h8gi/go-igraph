package igraph

import (
	"errors"
	"fmt"
	"reflect"
	"sort"
	"testing"
)

func TestCutStructureFixtures(t *testing.T) {
	tests := []struct {
		name           string
		vertexCount    int
		edges          []Edge
		directed       bool
		points         []int
		bridges        []int
		componentEdges [][]int
	}{
		{
			name:        "path",
			vertexCount: 4,
			edges:       []Edge{{From: 0, To: 1}, {From: 1, To: 2}, {From: 2, To: 3}},
			points:      []int{1, 2}, bridges: []int{0, 1, 2},
			componentEdges: [][]int{{0}, {1}, {2}},
		},
		{
			name:           "cycle",
			vertexCount:    4,
			edges:          []Edge{{From: 0, To: 1}, {From: 1, To: 2}, {From: 2, To: 3}, {From: 3, To: 0}},
			componentEdges: [][]int{{0, 1, 2, 3}},
		},
		{
			name:        "tree",
			vertexCount: 5,
			edges:       []Edge{{From: 0, To: 1}, {From: 0, To: 2}, {From: 0, To: 3}, {From: 3, To: 4}},
			points:      []int{0, 3}, bridges: []int{0, 1, 2, 3},
			componentEdges: [][]int{{0}, {1}, {2}, {3}},
		},
		{
			name:        "complete",
			vertexCount: 4,
			edges: []Edge{
				{From: 0, To: 1}, {From: 0, To: 2}, {From: 0, To: 3},
				{From: 1, To: 2}, {From: 1, To: 3}, {From: 2, To: 3},
			},
			componentEdges: [][]int{{0, 1, 2, 3, 4, 5}},
		},
		{
			name:        "disconnected union",
			vertexCount: 6,
			edges: []Edge{
				{From: 0, To: 1}, {From: 1, To: 2}, {From: 2, To: 0},
				{From: 3, To: 4},
			},
			bridges: []int{3}, componentEdges: [][]int{{0, 1, 2}, {3}},
		},
		{
			name:        "loops and parallel edges",
			vertexCount: 3,
			edges: []Edge{
				{From: 0, To: 1}, {From: 0, To: 1},
				{From: 1, To: 2}, {From: 2, To: 2},
			},
			points: []int{1}, bridges: []int{2},
			componentEdges: [][]int{{0, 1}, {2}},
		},
		{
			name:        "directed treated as undirected",
			vertexCount: 4,
			edges:       []Edge{{From: 1, To: 0}, {From: 1, To: 2}, {From: 3, To: 2}},
			directed:    true,
			points:      []int{1, 2}, bridges: []int{0, 1, 2},
			componentEdges: [][]int{{0}, {1}, {2}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g, err := NewGraphFromEdges(tt.vertexCount, tt.edges, tt.directed)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = g.Close() })

			points, err := g.ArticulationPoints()
			if err != nil {
				t.Fatalf("ArticulationPoints() error = %v", err)
			}
			assertIntSet(t, points, tt.points)
			bridges, err := g.Bridges()
			if err != nil {
				t.Fatalf("Bridges() error = %v", err)
			}
			assertIntSet(t, bridges, tt.bridges)

			result, err := g.BiconnectedComponents()
			if err != nil {
				t.Fatalf("BiconnectedComponents() error = %v", err)
			}
			assertBiconnectedResult(t, result, tt.vertexCount, tt.edges, tt.componentEdges, tt.points)
		})
	}
}

func TestCutStructureDegenerateGraphs(t *testing.T) {
	for _, vertexCount := range []int{0, 1, 3} {
		t.Run(testNameForVertexCount(vertexCount), func(t *testing.T) {
			g, err := NewGraphFromEdges(vertexCount, nil, vertexCount%2 == 1)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = g.Close() })
			points, err := g.ArticulationPoints()
			if err != nil || points == nil || len(points) != 0 {
				t.Errorf("ArticulationPoints() = %v, %v, want non-nil empty, nil", points, err)
			}
			bridges, err := g.Bridges()
			if err != nil || bridges == nil || len(bridges) != 0 {
				t.Errorf("Bridges() = %v, %v, want non-nil empty, nil", bridges, err)
			}
			result, err := g.BiconnectedComponents()
			if err != nil {
				t.Fatal(err)
			}
			if result.Count != 0 || result.ComponentEdges == nil || result.ComponentVertices == nil || result.ArticulationPoints == nil {
				t.Errorf("BiconnectedComponents() = %#v, want zero with non-nil slices", result)
			}
		})
	}
}

func TestCutStructureResultsSurviveClose(t *testing.T) {
	g, err := NewGraphFromEdges(3, []Edge{{From: 0, To: 1}, {From: 1, To: 2}}, false)
	if err != nil {
		t.Fatal(err)
	}
	points, err := g.ArticulationPoints()
	if err != nil {
		t.Fatal(err)
	}
	bridges, err := g.Bridges()
	if err != nil {
		t.Fatal(err)
	}
	components, err := g.BiconnectedComponents()
	if err != nil {
		t.Fatal(err)
	}
	if err := g.Close(); err != nil {
		t.Fatal(err)
	}
	assertIntSet(t, points, []int{1})
	assertIntSet(t, bridges, []int{0, 1})
	assertIntSet(t, components.ArticulationPoints, []int{1})
	components.ComponentEdges[0][0] = 99
}

func TestCutStructureRejectsClosedGraph(t *testing.T) {
	g, err := NewGraph()
	if err != nil {
		t.Fatal(err)
	}
	if err := g.Close(); err != nil {
		t.Fatal(err)
	}
	for _, graph := range []*Graph{g, nil} {
		if _, err := graph.ArticulationPoints(); !errors.Is(err, ErrClosed) {
			t.Errorf("ArticulationPoints() error = %v, want %v", err, ErrClosed)
		}
		if _, err := graph.Bridges(); !errors.Is(err, ErrClosed) {
			t.Errorf("Bridges() error = %v, want %v", err, ErrClosed)
		}
		if _, err := graph.BiconnectedComponents(); !errors.Is(err, ErrClosed) {
			t.Errorf("BiconnectedComponents() error = %v, want %v", err, ErrClosed)
		}
	}
}

func TestValidateBiconnectedComponentsRejectsInvalidResults(t *testing.T) {
	tests := []BiconnectedComponents{
		{ComponentVertices: [][]int{}, ArticulationPoints: []int{}},
		{ComponentEdges: [][]int{}, ArticulationPoints: []int{}},
		{ComponentEdges: [][]int{}, ComponentVertices: [][]int{}},
		{Count: 1, ComponentEdges: [][]int{}, ComponentVertices: [][]int{}, ArticulationPoints: []int{}},
		{Count: 1, ComponentEdges: [][]int{{}}, ComponentVertices: [][]int{nil}, ArticulationPoints: []int{}},
		{Count: 1, ComponentEdges: [][]int{{1}}, ComponentVertices: [][]int{{0}}, ArticulationPoints: []int{}},
		{Count: 1, ComponentEdges: [][]int{{0}}, ComponentVertices: [][]int{{1}}, ArticulationPoints: []int{}},
		{ComponentEdges: [][]int{}, ComponentVertices: [][]int{}, ArticulationPoints: []int{1}},
	}
	for i, result := range tests {
		if err := validateBiconnectedComponents(result, 1, 1); err == nil {
			t.Errorf("case %d: validateBiconnectedComponents(%#v) error = nil", i, result)
		}
	}
}

func TestCollectCutStructureIDsPropagatesFailuresAndCleansUp(t *testing.T) {
	forced := errors.New("forced failure")
	tests := []struct {
		name       string
		failInit   bool
		failQuery  bool
		failSlice  bool
		wantClosed int
	}{
		{name: "initialization", failInit: true},
		{name: "upstream query", failQuery: true, wantClosed: 1},
		{name: "result conversion", failSlice: true, wantClosed: 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			closed := 0
			queried := false
			result, err := collectCutStructureIDs(cutStructureOperations{
				newVector: func() (*intVector, error) {
					if tt.failInit {
						return nil, forced
					}
					return &intVector{}, nil
				},
				close: func(*intVector) { closed++ },
				query: func(*intVector) error {
					queried = true
					if tt.failQuery {
						return forced
					}
					return nil
				},
				slice: func(*intVector) ([]int, error) {
					if tt.failSlice {
						return nil, forced
					}
					return []int{}, nil
				},
			})
			if result != nil || !errors.Is(err, forced) {
				t.Errorf("collectCutStructureIDs() = %v, %v, want nil, %v", result, err, forced)
			}
			if closed != tt.wantClosed {
				t.Errorf("close count = %d, want %d", closed, tt.wantClosed)
			}
			if queried != !tt.failInit {
				t.Errorf("query called = %t with initialization failure = %t", queried, tt.failInit)
			}
		})
	}
}

func TestCollectBiconnectedComponentsCleansUpAfterFailures(t *testing.T) {
	forced := errors.New("forced failure")
	tests := []struct {
		name              string
		failListInitCall  int
		failVectorInit    bool
		failQuery         bool
		failListSliceCall int
		failVectorSlice   bool
		wantClosedLists   int
		wantClosedVectors int
	}{
		{name: "first list initialization", failListInitCall: 1},
		{name: "second list initialization", failListInitCall: 2, wantClosedLists: 1},
		{name: "vector initialization", failVectorInit: true, wantClosedLists: 2},
		{name: "upstream query", failQuery: true, wantClosedLists: 2, wantClosedVectors: 1},
		{name: "first nested conversion", failListSliceCall: 1, wantClosedLists: 2, wantClosedVectors: 1},
		{name: "partial nested conversion", failListSliceCall: 2, wantClosedLists: 2, wantClosedVectors: 1},
		{name: "articulation conversion", failVectorSlice: true, wantClosedLists: 2, wantClosedVectors: 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			listInitCalls := 0
			listSliceCalls := 0
			closedLists := 0
			closedVectors := 0
			operations := biconnectedOperations{
				newList: func() (*intVectorList, error) {
					listInitCalls++
					if listInitCalls == tt.failListInitCall {
						return nil, forced
					}
					return &intVectorList{}, nil
				},
				newVector: func() (*intVector, error) {
					if tt.failVectorInit {
						return nil, forced
					}
					return &intVector{}, nil
				},
				closeList:   func(*intVectorList) { closedLists++ },
				closeVector: func(*intVector) { closedVectors++ },
				query: func(*intVectorList, *intVectorList, *intVector) (int, error) {
					if tt.failQuery {
						return 0, forced
					}
					return 0, nil
				},
				listSlices: func(*intVectorList) ([][]int, error) {
					listSliceCalls++
					if listSliceCalls == tt.failListSliceCall {
						return nil, forced
					}
					return [][]int{}, nil
				},
				vectorSlice: func(*intVector) ([]int, error) {
					if tt.failVectorSlice {
						return nil, forced
					}
					return []int{}, nil
				},
			}
			result, err := collectBiconnectedComponents(0, 0, operations)
			if !errors.Is(err, forced) {
				t.Errorf("error = %v, want %v", err, forced)
			}
			if !reflect.DeepEqual(result, BiconnectedComponents{}) {
				t.Errorf("result = %#v, want zero value", result)
			}
			if closedLists != tt.wantClosedLists || closedVectors != tt.wantClosedVectors {
				t.Errorf("closed lists/vectors = %d/%d, want %d/%d", closedLists, closedVectors, tt.wantClosedLists, tt.wantClosedVectors)
			}
		})
	}
}

func assertBiconnectedResult(t *testing.T, got BiconnectedComponents, vertexCount int, edges []Edge, wantEdges [][]int, wantPoints []int) {
	t.Helper()
	if got.Count != len(got.ComponentEdges) || got.Count != len(got.ComponentVertices) {
		t.Fatalf("inconsistent component count: %#v", got)
	}
	assertNestedIntSets(t, got.ComponentEdges, wantEdges)
	assertIntSet(t, got.ArticulationPoints, wantPoints)
	componentMemberships := make([]int, vertexCount)
	for componentID, componentEdges := range got.ComponentEdges {
		wantVertices := map[int]bool{}
		for _, edgeID := range componentEdges {
			if edgeID < 0 || edgeID >= len(edges) {
				t.Fatalf("component %d has invalid edge ID %d", componentID, edgeID)
			}
			wantVertices[edges[edgeID].From] = true
			wantVertices[edges[edgeID].To] = true
		}
		gotVertices := append([]int(nil), got.ComponentVertices[componentID]...)
		sort.Ints(gotVertices)
		wantVertexSlice := make([]int, 0, len(wantVertices))
		for vertexID := range wantVertices {
			if vertexID < 0 || vertexID >= vertexCount {
				t.Fatalf("component %d has invalid vertex ID %d", componentID, vertexID)
			}
			wantVertexSlice = append(wantVertexSlice, vertexID)
		}
		sort.Ints(wantVertexSlice)
		if !reflect.DeepEqual(gotVertices, wantVertexSlice) {
			t.Errorf("component %d vertices = %v, want endpoints %v", componentID, gotVertices, wantVertexSlice)
		}
		for _, vertexID := range gotVertices {
			componentMemberships[vertexID]++
		}
	}
	pointsFromMembership := []int{}
	for vertexID, memberships := range componentMemberships {
		if memberships > 1 {
			pointsFromMembership = append(pointsFromMembership, vertexID)
		}
	}
	assertIntSet(t, got.ArticulationPoints, pointsFromMembership)
}

func assertIntSet(t *testing.T, got, want []int) {
	t.Helper()
	got = append([]int(nil), got...)
	want = append([]int(nil), want...)
	sort.Ints(got)
	sort.Ints(want)
	if !reflect.DeepEqual(got, want) {
		t.Errorf("IDs = %v, want %v", got, want)
	}
}

func assertNestedIntSets(t *testing.T, got, want [][]int) {
	t.Helper()
	normalize := func(values [][]int) []string {
		result := make([]string, len(values))
		for i, value := range values {
			value = append([]int(nil), value...)
			sort.Ints(value)
			result[i] = fmt.Sprint(value)
		}
		sort.Strings(result)
		return result
	}
	if got, want := normalize(got), normalize(want); !reflect.DeepEqual(got, want) {
		t.Errorf("components = %v, want %v", got, want)
	}
}
