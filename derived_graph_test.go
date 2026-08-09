package igraph

import (
	"errors"
	"reflect"
	"sync"
	"testing"
	"time"
)

func TestIDMappingSemanticsAndOwnership(t *testing.T) {
	input := []int{2, RemovedID, 0, 2}
	mapping, err := newIDMapping(input, 4)
	if err != nil {
		t.Fatalf("newIDMapping() error = %v", err)
	}
	input[0] = RemovedID
	if want := []int{2, RemovedID, 0, 2}; !reflect.DeepEqual(mapping.OldToNew, want) {
		t.Errorf("OldToNew = %v, want %v", mapping.OldToNew, want)
	}
	if want := []int{2, RemovedID, 0, RemovedID}; !reflect.DeepEqual(mapping.NewToOld, want) {
		t.Errorf("NewToOld = %v, want %v", mapping.NewToOld, want)
	}

	empty, err := newIDMapping(nil, 0)
	if err != nil {
		t.Fatalf("empty newIDMapping() error = %v", err)
	}
	if empty.OldToNew == nil || empty.NewToOld == nil || len(empty.OldToNew) != 0 || len(empty.NewToOld) != 0 {
		t.Errorf("empty mapping = %#v, want non-nil empty slices", empty)
	}

	identity, err := identityIDMapping(3)
	if err != nil {
		t.Fatalf("identityIDMapping() error = %v", err)
	}
	wantIdentity := IDMapping{OldToNew: []int{0, 1, 2}, NewToOld: []int{0, 1, 2}}
	if !reflect.DeepEqual(identity, wantIdentity) {
		t.Errorf("identity mapping = %#v, want %#v", identity, wantIdentity)
	}
}

func TestIDMappingRejectsInvalidValues(t *testing.T) {
	for _, test := range []struct {
		name     string
		oldToNew []int
		newCount int
	}{
		{name: "negative count", newCount: -1},
		{name: "invalid negative sentinel", oldToNew: []int{-2}, newCount: 1},
		{name: "out of range", oldToNew: []int{1}, newCount: 1},
	} {
		t.Run(test.name, func(t *testing.T) {
			if mapping, err := newIDMapping(test.oldToNew, test.newCount); err == nil {
				t.Errorf("newIDMapping() = %#v, nil, want error", mapping)
			}
		})
	}
	if mapping, err := identityIDMapping(-1); err == nil {
		t.Errorf("identityIDMapping(-1) = %#v, nil, want error", mapping)
	}
}

func TestGraphListExtractionTransfersIndependentOwnership(t *testing.T) {
	sourceA := testGraphFromEdges(t, 3, []Edge{{0, 1}}, true)
	sourceB := testGraphFromEdges(t, 2, []Edge{{0, 1}}, false)
	list, err := newGraphListFromCopies([]*Graph{sourceA, sourceB})
	if err != nil {
		t.Fatalf("newGraphListFromCopies() error = %v", err)
	}
	graphs, err := list.takeGraphs()
	if err != nil {
		t.Fatalf("takeGraphs() error = %v", err)
	}
	if graphs == nil || len(graphs) != 2 {
		t.Fatalf("takeGraphs() = %v, want two graphs", graphs)
	}

	if err := sourceA.Close(); err != nil {
		t.Fatal(err)
	}
	if err := sourceB.Close(); err != nil {
		t.Fatal(err)
	}
	if got, err := graphs[0].VertexCount(); err != nil || got != 3 {
		t.Errorf("first VertexCount() after source close = %d, %v, want 3, nil", got, err)
	}
	if err := graphs[1].Close(); err != nil {
		t.Fatal(err)
	}
	if err := graphs[1].Close(); err != nil {
		t.Fatalf("repeated sibling Close() error = %v", err)
	}
	if got, err := graphs[0].EdgeCount(); err != nil || got != 1 {
		t.Errorf("first EdgeCount() after sibling close = %d, %v, want 1, nil", got, err)
	}
	if err := graphs[0].Close(); err != nil {
		t.Fatal(err)
	}
}

func TestGraphListExtractionCleansUpEveryFailurePosition(t *testing.T) {
	for failAt := 0; failAt < 3; failAt++ {
		t.Run(string(rune('0'+failAt)), func(t *testing.T) {
			sources := []*Graph{
				testGraphFromEdges(t, 1, nil, false),
				testGraphFromEdges(t, 2, []Edge{{0, 1}}, false),
				testGraphFromEdges(t, 3, []Edge{{0, 1}, {1, 2}}, true),
			}
			list, err := newGraphListFromCopies(sources)
			if err != nil {
				t.Fatal(err)
			}
			injected := errors.New("injected conversion failure")
			adopted := make([]*Graph, 0, failAt)
			graphs, err := list.takeGraphsWithHooks(graphListExtractionHooks{
				beforeAdopt: func(index int) error {
					if index == failAt {
						return injected
					}
					return nil
				},
				afterAdopt: func(_ int, graph *Graph) error {
					adopted = append(adopted, graph)
					return nil
				},
			})
			if graphs != nil || !errors.Is(err, injected) {
				t.Errorf("takeGraphsWithHooks() = %v, %v, want nil, injected error", graphs, err)
			}
			if list.initialized {
				t.Error("graph list remains initialized after extraction failure")
			}
			for index, graph := range adopted {
				if _, err := graph.VertexCount(); !errors.Is(err, ErrClosed) {
					t.Errorf("adopted graph %d error = %v, want %v", index, err, ErrClosed)
				}
			}
		})
	}
}

func TestGraphListInitializationFailureAndEmptyResult(t *testing.T) {
	list, err := initializeGraphList(&graphList{}, func() int { return 1 })
	if list != nil || err == nil {
		t.Errorf("initializeGraphList(failure) = %v, %v, want nil, error", list, err)
	}

	empty, err := newGraphList()
	if err != nil {
		t.Fatal(err)
	}
	graphs, err := empty.takeGraphs()
	if err != nil {
		t.Fatal(err)
	}
	if graphs == nil || len(graphs) != 0 {
		t.Errorf("empty takeGraphs() = %v, want non-nil empty slice", graphs)
	}
}

func TestGraphListUpstreamAndPostAdoptFailuresCleanUp(t *testing.T) {
	for _, stage := range []string{"upstream", "post-adopt"} {
		t.Run(stage, func(t *testing.T) {
			source := testGraphFromEdges(t, 2, []Edge{{0, 1}}, false)
			list, err := newGraphListFromCopies([]*Graph{source, source})
			if err != nil {
				t.Fatal(err)
			}
			injected := errors.New("injected failure")
			var adopted *Graph
			hooks := graphListExtractionHooks{}
			if stage == "upstream" {
				hooks.beforeRemove = func(int) error { return injected }
			} else {
				hooks.afterAdopt = func(_ int, graph *Graph) error {
					adopted = graph
					return injected
				}
			}
			graphs, err := list.takeGraphsWithHooks(hooks)
			if graphs != nil || !errors.Is(err, injected) {
				t.Errorf("takeGraphsWithHooks() = %v, %v, want nil, injected error", graphs, err)
			}
			if list.initialized {
				t.Error("graph list remains initialized after failure")
			}
			if adopted != nil {
				if _, err := adopted.VertexCount(); !errors.Is(err, ErrClosed) {
					t.Errorf("adopted graph error = %v, want %v", err, ErrClosed)
				}
			}
		})
	}
}

func TestWithGraphsLockedRejectsClosedNilAndDuplicateDoesNotDeadlock(t *testing.T) {
	graph := testGraphFromEdges(t, 1, nil, false)
	if err := withGraphsLocked([]*Graph{graph, graph}, func() error { return nil }); err != nil {
		t.Fatalf("withGraphsLocked(duplicate) error = %v", err)
	}
	if err := withGraphsLocked([]*Graph{nil}, func() error { return nil }); !errors.Is(err, ErrClosed) {
		t.Errorf("withGraphsLocked(nil) error = %v, want %v", err, ErrClosed)
	}
	if err := graph.Close(); err != nil {
		t.Fatal(err)
	}
	if err := withGraphsLocked([]*Graph{graph}, func() error { return nil }); !errors.Is(err, ErrClosed) {
		t.Errorf("withGraphsLocked(closed) error = %v, want %v", err, ErrClosed)
	}
}

func TestWithGraphsLockedOppositeOrdersDoNotDeadlock(t *testing.T) {
	a := testGraphFromEdges(t, 1, nil, false)
	b := testGraphFromEdges(t, 1, nil, false)
	var start sync.WaitGroup
	start.Add(2)
	done := make(chan error, 2)
	for _, graphs := range [][]*Graph{{a, b}, {b, a}} {
		go func(graphs []*Graph) {
			start.Done()
			start.Wait()
			done <- withGraphsLocked(graphs, func() error {
				time.Sleep(time.Millisecond)
				return nil
			})
		}(graphs)
	}
	for range 2 {
		select {
		case err := <-done:
			if err != nil {
				t.Errorf("withGraphsLocked() error = %v", err)
			}
		case <-time.After(2 * time.Second):
			t.Fatal("opposite lock orders deadlocked")
		}
	}
}

func testGraphFromEdges(t *testing.T, vertexCount int, edges []Edge, directed bool) *Graph {
	t.Helper()
	graph, err := NewGraphFromEdges(vertexCount, edges, directed)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = graph.Close() })
	return graph
}
