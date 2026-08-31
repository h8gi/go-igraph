package igraph

import "testing"

func TestDLAStableGraphLockValidation(t *testing.T) {
	if err := withLockedGraphs(nil, nil); err == nil {
		t.Fatal("withLockedGraphs accepted a nil operation")
	}
	if err := withGraphsLocked(nil, nil); err == nil {
		t.Fatal("withGraphsLocked accepted a nil operation")
	}
	called := false
	if err := withGraphsLocked(nil, func() error {
		called = true
		return nil
	}); err != nil || !called {
		t.Fatalf("empty stable lock = called %v, error %v", called, err)
	}
}

func TestSharedEdgePolicyBranches(t *testing.T) {
	want := []EdgeType{EdgeTypeSimple, EdgeTypeMulti, EdgeTypeLoops, EdgeTypeLoopsAndMulti}
	got := []EdgeType{
		edgeTypeFromFlags(false, false), edgeTypeFromFlags(false, true),
		edgeTypeFromFlags(true, false), edgeTypeFromFlags(true, true),
	}
	for index := range want {
		if got[index] != want[index] {
			t.Fatalf("edge policy %d = %v, want %v", index, got[index], want[index])
		}
	}
	for _, mode := range []RewireMode{RewireSimple, RewireLoops, RewireMulti, RewireLoopsAndMulti} {
		if _, err := mode.cValue(); err != nil {
			t.Fatalf("valid rewire mode %d failed: %v", mode, err)
		}
	}
	if _, err := RewireMode(99).cValue(); err == nil {
		t.Fatal("invalid rewire mode succeeded")
	}
}
