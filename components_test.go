package igraph

import (
	"errors"
	"reflect"
	"testing"
)

func TestDirectedConnectedComponents(t *testing.T) {
	g, err := NewGraphFromEdges(5, []Edge{
		{From: 0, To: 1},
		{From: 1, To: 0},
		{From: 1, To: 2},
		{From: 2, To: 3},
		{From: 3, To: 2},
	}, true)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = g.Close() })

	weak, err := g.ConnectedComponents(ConnectednessWeak)
	if err != nil {
		t.Fatalf("ConnectedComponents(weak) error = %v", err)
	}
	assertComponents(t, weak, [][]int{{0, 1, 2, 3}, {4}})
	if got, err := g.IsConnected(ConnectednessWeak); err != nil || got {
		t.Errorf("IsConnected(weak) = %t, %v, want false, nil", got, err)
	}

	strong, err := g.ConnectedComponents(ConnectednessStrong)
	if err != nil {
		t.Fatalf("ConnectedComponents(strong) error = %v", err)
	}
	assertComponents(t, strong, [][]int{{0, 1}, {2, 3}, {4}})
	if strong.Membership[0] >= strong.Membership[2] {
		t.Errorf("strong component IDs = %v, want source component before target component", strong.Membership)
	}
	if got, err := g.IsConnected(ConnectednessStrong); err != nil || got {
		t.Errorf("IsConnected(strong) = %t, %v, want false, nil", got, err)
	}
}

func TestDirectedWeaklyButNotStronglyConnected(t *testing.T) {
	g, err := NewGraphFromEdges(3, []Edge{{From: 0, To: 1}, {From: 1, To: 2}}, true)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = g.Close() })

	weak, err := g.ConnectedComponents(ConnectednessWeak)
	if err != nil {
		t.Fatal(err)
	}
	assertComponents(t, weak, [][]int{{0, 1, 2}})
	strong, err := g.ConnectedComponents(ConnectednessStrong)
	if err != nil {
		t.Fatal(err)
	}
	assertComponents(t, strong, [][]int{{0}, {1}, {2}})

	if connected, err := g.IsConnected(ConnectednessWeak); err != nil || !connected {
		t.Errorf("IsConnected(weak) = %t, %v, want true, nil", connected, err)
	}
	if connected, err := g.IsConnected(ConnectednessStrong); err != nil || connected {
		t.Errorf("IsConnected(strong) = %t, %v, want false, nil", connected, err)
	}
}

func TestUndirectedConnectedComponents(t *testing.T) {
	tests := []struct {
		name       string
		edges      []Edge
		components [][]int
		connected  bool
	}{
		{
			name:       "connected",
			edges:      []Edge{{From: 0, To: 1}, {From: 1, To: 2}, {From: 2, To: 3}},
			components: [][]int{{0, 1, 2, 3}},
			connected:  true,
		},
		{
			name:       "disconnected",
			edges:      []Edge{{From: 0, To: 1}, {From: 2, To: 3}},
			components: [][]int{{0, 1}, {2, 3}},
			connected:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g, err := NewGraphFromEdges(4, tt.edges, false)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = g.Close() })

			var first ConnectedComponents
			for _, mode := range []ConnectednessMode{ConnectednessWeak, ConnectednessStrong} {
				result, err := g.ConnectedComponents(mode)
				if err != nil {
					t.Fatalf("ConnectedComponents(%d) error = %v", mode, err)
				}
				assertComponents(t, result, tt.components)
				connected, err := g.IsConnected(mode)
				if err != nil || connected != tt.connected {
					t.Errorf("IsConnected(%d) = %t, %v, want %t, nil", mode, connected, err, tt.connected)
				}
				if mode == ConnectednessWeak {
					first = result
				} else if !reflect.DeepEqual(result, first) {
					t.Errorf("strong result = %#v, want undirected weak result %#v", result, first)
				}
			}
		})
	}
}

func TestEmptyAndIsolatedConnectedComponents(t *testing.T) {
	for _, vertexCount := range []int{0, 3} {
		t.Run(testNameForVertexCount(vertexCount), func(t *testing.T) {
			g, err := NewGraphFromEdges(vertexCount, nil, true)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = g.Close() })

			for _, mode := range []ConnectednessMode{ConnectednessWeak, ConnectednessStrong} {
				result, err := g.ConnectedComponents(mode)
				if err != nil {
					t.Fatalf("ConnectedComponents(%d) error = %v", mode, err)
				}
				if vertexCount == 0 {
					if result.Membership == nil || result.Sizes == nil {
						t.Errorf("empty result contains nil slice: %#v", result)
					}
					assertComponents(t, result, nil)
				} else {
					assertComponents(t, result, [][]int{{0}, {1}, {2}})
				}
				if connected, err := g.IsConnected(mode); err != nil || connected {
					t.Errorf("IsConnected(%d) = %t, %v, want false, nil", mode, connected, err)
				}
			}
		})
	}
}

func TestConnectedComponentsOwnsResults(t *testing.T) {
	g, err := NewGraphFromEdges(3, []Edge{{From: 0, To: 1}}, false)
	if err != nil {
		t.Fatal(err)
	}
	result, err := g.ConnectedComponents(ConnectednessWeak)
	if err != nil {
		t.Fatal(err)
	}
	wantMembership := append([]int(nil), result.Membership...)
	wantSizes := append([]int(nil), result.Sizes...)
	if err := g.Close(); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(result.Membership, wantMembership) || !reflect.DeepEqual(result.Sizes, wantSizes) {
		t.Errorf("result changed after graph Close: %#v", result)
	}
	result.Membership[0] = result.Membership[1]
	result.Sizes[0] = 99
}

func TestConnectedComponentQueriesRejectInvalidModeAndClosedGraph(t *testing.T) {
	g, err := NewGraphFromEdges(1, nil, false)
	if err != nil {
		t.Fatal(err)
	}
	invalid := ConnectednessMode(99)
	if _, err := g.ConnectedComponents(invalid); err == nil {
		t.Error("ConnectedComponents(invalid) error = nil")
	}
	if _, err := g.IsConnected(invalid); err == nil {
		t.Error("IsConnected(invalid) error = nil")
	}
	if err := g.Close(); err != nil {
		t.Fatal(err)
	}
	assertComponentQueriesClosed(t, g)
	var nilGraph *Graph
	assertComponentQueriesClosed(t, nilGraph)
}

func TestValidateConnectedComponentsRejectsInvalidResults(t *testing.T) {
	tests := []struct {
		name        string
		result      ConnectedComponents
		vertexCount int
	}{
		{name: "membership length", result: ConnectedComponents{}, vertexCount: 1},
		{name: "size length", result: ConnectedComponents{Membership: []int{0}, Count: 1}, vertexCount: 1},
		{name: "negative component", result: ConnectedComponents{Membership: []int{-1}, Sizes: []int{1}, Count: 1}, vertexCount: 1},
		{name: "large component", result: ConnectedComponents{Membership: []int{1}, Sizes: []int{1}, Count: 1}, vertexCount: 1},
		{name: "incorrect size", result: ConnectedComponents{Membership: []int{0}, Sizes: []int{2}, Count: 1}, vertexCount: 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validateConnectedComponents(tt.result, tt.vertexCount); err == nil {
				t.Error("validateConnectedComponents() error = nil")
			}
		})
	}
}

func assertComponentQueriesClosed(t *testing.T, g *Graph) {
	t.Helper()
	if _, err := g.ConnectedComponents(ConnectednessWeak); !errors.Is(err, ErrClosed) {
		t.Errorf("ConnectedComponents() error = %v, want %v", err, ErrClosed)
	}
	if _, err := g.IsConnected(ConnectednessWeak); !errors.Is(err, ErrClosed) {
		t.Errorf("IsConnected() error = %v, want %v", err, ErrClosed)
	}
}

func assertComponents(t *testing.T, got ConnectedComponents, wantGroups [][]int) {
	t.Helper()
	if got.Count != len(wantGroups) {
		t.Fatalf("Count = %d, want %d; result = %#v", got.Count, len(wantGroups), got)
	}
	if len(got.Sizes) != got.Count {
		t.Errorf("len(Sizes) = %d, want Count %d", len(got.Sizes), got.Count)
	}
	wantVertexCount := 0
	for _, group := range wantGroups {
		wantVertexCount += len(group)
		if len(group) == 0 {
			t.Fatal("test has empty expected component")
		}
	}
	if len(got.Membership) != wantVertexCount {
		t.Fatalf("len(Membership) = %d, want %d", len(got.Membership), wantVertexCount)
	}
	for _, group := range wantGroups {
		componentID := got.Membership[group[0]]
		if componentID < 0 || componentID >= got.Count {
			t.Fatalf("Membership[%d] = %d out of range [0, %d)", group[0], componentID, got.Count)
		}
		if got.Sizes[componentID] != len(group) {
			t.Errorf("Sizes[%d] = %d, want %d", componentID, got.Sizes[componentID], len(group))
		}
		for _, vertexID := range group[1:] {
			if got.Membership[vertexID] != componentID {
				t.Errorf("vertices %d and %d have different components in %v", group[0], vertexID, got.Membership)
			}
		}
	}
	for i, left := range wantGroups {
		for _, right := range wantGroups[i+1:] {
			if got.Membership[left[0]] == got.Membership[right[0]] {
				t.Errorf("expected distinct groups %v and %v, membership = %v", left, right, got.Membership)
			}
		}
	}
}

func testNameForVertexCount(vertexCount int) string {
	if vertexCount == 0 {
		return "empty"
	}
	return "isolated"
}
