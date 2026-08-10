package igraph_test

import (
	"testing"

	"github.com/h8gi/go-igraph"
)

func TestIsGraphical(t *testing.T) {
	t.Run("undirected simple graphical sequence", func(t *testing.T) {
		ok, err := igraph.IsGraphical([]int{2, 2, 2}, nil, igraph.EdgeTypeSimple)
		if err != nil {
			t.Fatalf("IsGraphical failed: %v", err)
		}
		if !ok {
			t.Errorf("expected [2, 2, 2] to be graphical")
		}
	})

	t.Run("undirected non-graphical simple sequence", func(t *testing.T) {
		ok, err := igraph.IsGraphical([]int{3, 1, 1}, nil, igraph.EdgeTypeSimple)
		if err != nil {
			t.Fatalf("IsGraphical failed: %v", err)
		}
		if ok {
			t.Errorf("expected [3, 1, 1] to not be graphical for simple edge type")
		}
	})

	t.Run("directed graphical sequence", func(t *testing.T) {
		ok, err := igraph.IsGraphical([]int{1, 1, 1}, []int{1, 1, 1}, igraph.EdgeTypeSimple)
		if err != nil {
			t.Fatalf("IsGraphical failed: %v", err)
		}
		if !ok {
			t.Errorf("expected directed [1, 1, 1] to be graphical")
		}
	})

	t.Run("empty sequences", func(t *testing.T) {
		for name, inDeg := range map[string][]int{"nil inDeg": nil, "empty inDeg": {}} {
			ok, err := igraph.IsGraphical([]int{}, inDeg, igraph.EdgeTypeSimple)
			if err != nil {
				t.Fatalf("IsGraphical failed for %s: %v", name, err)
			}
			if !ok {
				t.Errorf("expected an empty sequence to be graphical for %s", name)
			}
		}
	})

	t.Run("edge type variations", func(t *testing.T) {
		types := []igraph.EdgeType{
			igraph.EdgeTypeSimple,
			igraph.EdgeTypeLoops,
			igraph.EdgeTypeMulti,
			igraph.EdgeTypeLoopsAndMulti,
		}
		for _, et := range types {
			ok, err := igraph.IsGraphical([]int{2, 2, 2}, nil, et)
			if err != nil {
				t.Fatalf("IsGraphical failed for EdgeType %d: %v", et, err)
			}
			if !ok {
				t.Errorf("expected [2, 2, 2] to be graphical for EdgeType %d", et)
			}
		}
	})

	t.Run("invalid edge type", func(t *testing.T) {
		if _, err := igraph.IsGraphical([]int{2, 2, 2}, nil, igraph.EdgeType(99)); err == nil {
			t.Errorf("expected error for invalid EdgeType")
		}
	})

	t.Run("mismatched directed lengths", func(t *testing.T) {
		if _, err := igraph.IsGraphical([]int{1, 1}, []int{1}, igraph.EdgeTypeSimple); err == nil {
			t.Errorf("expected error for mismatched slice lengths")
		}
	})
}

func TestIsBigraphical(t *testing.T) {
	t.Run("valid bipartite degree sequence", func(t *testing.T) {
		ok, err := igraph.IsBigraphical([]int{2, 2}, []int{2, 2}, igraph.EdgeTypeSimple)
		if err != nil {
			t.Fatalf("IsBigraphical failed: %v", err)
		}
		if !ok {
			t.Errorf("expected bipartite sequence to be bigraphical")
		}
	})

	t.Run("invalid bipartite degree sequence", func(t *testing.T) {
		ok, err := igraph.IsBigraphical([]int{3, 3}, []int{1, 1}, igraph.EdgeTypeSimple)
		if err != nil {
			t.Fatalf("IsBigraphical failed: %v", err)
		}
		if ok {
			t.Errorf("expected mismatched bipartite sequence to not be bigraphical")
		}
	})

	t.Run("empty sequences", func(t *testing.T) {
		ok, err := igraph.IsBigraphical([]int{}, []int{}, igraph.EdgeTypeSimple)
		if err != nil {
			t.Fatalf("IsBigraphical failed: %v", err)
		}
		if !ok {
			t.Errorf("expected empty sequences to be bigraphical")
		}
	})

	t.Run("loops component is ignored", func(t *testing.T) {
		// Bipartite graphs cannot contain self-loops, so the loops component
		// of the edge type must not change the answer. [2] and [2] needs a
		// double edge, so it is bigraphical only when multi-edges are allowed.
		cases := []struct {
			edgeTypes  igraph.EdgeType
			equivalent igraph.EdgeType
		}{
			{edgeTypes: igraph.EdgeTypeLoops, equivalent: igraph.EdgeTypeSimple},
			{edgeTypes: igraph.EdgeTypeLoopsAndMulti, equivalent: igraph.EdgeTypeMulti},
		}
		for _, tc := range cases {
			got, err := igraph.IsBigraphical([]int{2}, []int{2}, tc.edgeTypes)
			if err != nil {
				t.Fatalf("IsBigraphical failed for EdgeType %d: %v", tc.edgeTypes, err)
			}
			want, err := igraph.IsBigraphical([]int{2}, []int{2}, tc.equivalent)
			if err != nil {
				t.Fatalf("IsBigraphical failed for EdgeType %d: %v", tc.equivalent, err)
			}
			if got != want {
				t.Errorf("expected EdgeType %d to behave like EdgeType %d, got %v vs %v", tc.edgeTypes, tc.equivalent, got, want)
			}
		}
	})

	t.Run("invalid edge type", func(t *testing.T) {
		if _, err := igraph.IsBigraphical([]int{1}, []int{1}, igraph.EdgeType(99)); err == nil {
			t.Errorf("expected error for invalid EdgeType")
		}
	})
}
