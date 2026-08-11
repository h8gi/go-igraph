package igraph

import (
	"errors"
	"sync"
	"testing"
)

func TestIsomorphic(t *testing.T) {
	left := testGraphFromEdges(t, 4, []Edge{{0, 1}, {1, 2}, {2, 3}, {3, 0}}, false)
	right := testGraphFromEdges(t, 4, []Edge{{0, 2}, {2, 1}, {1, 3}, {3, 0}}, false)
	nonmatch := testGraphFromEdges(t, 4, []Edge{{0, 1}, {1, 2}, {2, 3}}, false)

	found, err := left.Isomorphic(right)
	if err != nil || !found {
		t.Fatalf("Isomorphic(isomorphic) = %v, %v; want true, nil", found, err)
	}
	found, err = left.Isomorphic(nonmatch)
	if err != nil || found {
		t.Fatalf("Isomorphic(non-isomorphic) = %v, %v; want false, nil", found, err)
	}
	found, err = left.Isomorphic(left)
	if err != nil || !found {
		t.Fatalf("Isomorphic(self) = %v, %v; want true, nil", found, err)
	}
}

func TestIsomorphicMultigraphAndEmpty(t *testing.T) {
	left := testGraphFromEdges(t, 2, []Edge{{0, 0}, {0, 1}, {0, 1}}, false)
	right := testGraphFromEdges(t, 2, []Edge{{1, 1}, {1, 0}, {1, 0}}, false)
	found, err := left.Isomorphic(right)
	if err != nil || !found {
		t.Fatalf("Isomorphic(multigraph) = %v, %v; want true, nil", found, err)
	}

	empty1 := testGraphFromEdges(t, 0, nil, false)
	empty2 := testGraphFromEdges(t, 0, nil, false)
	found, err = empty1.Isomorphic(empty2)
	if err != nil || !found {
		t.Fatalf("Isomorphic(empty) = %v, %v; want true, nil", found, err)
	}
}

func TestContainsSubgraphIsomorphicTo(t *testing.T) {
	target := testGraphFromEdges(t, 5, []Edge{{0, 1}, {1, 2}, {2, 0}, {2, 3}, {3, 4}}, false)
	triangle := testGraphFromEdges(t, 3, []Edge{{0, 1}, {1, 2}, {2, 0}}, false)
	clique4 := testGraphFromEdges(t, 4, []Edge{{0, 1}, {0, 2}, {0, 3}, {1, 2}, {1, 3}, {2, 3}}, false)

	found, err := target.ContainsSubgraphIsomorphicTo(triangle)
	if err != nil || !found {
		t.Fatalf("ContainsSubgraphIsomorphicTo(match) = %v, %v; want true, nil", found, err)
	}
	found, err = target.ContainsSubgraphIsomorphicTo(clique4)
	if err != nil || found {
		t.Fatalf("ContainsSubgraphIsomorphicTo(non-match) = %v, %v; want false, nil", found, err)
	}
}

func TestIsomorphismInvalidGraphsAndDirectedness(t *testing.T) {
	graph := testGraphFromEdges(t, 2, []Edge{{0, 1}}, false)
	directed := testGraphFromEdges(t, 2, []Edge{{0, 1}}, true)
	closed := testGraphFromEdges(t, 0, nil, false)
	if err := closed.Close(); err != nil {
		t.Fatal(err)
	}

	for name, call := range map[string]func() error{
		"nil receiver":   func() error { _, err := (*Graph)(nil).Isomorphic(graph); return err },
		"nil operand":    func() error { _, err := graph.Isomorphic(nil); return err },
		"closed operand": func() error { _, err := graph.Isomorphic(closed); return err },
		"nil pattern":    func() error { _, err := graph.ContainsSubgraphIsomorphicTo(nil); return err },
		"closed target":  func() error { _, err := closed.ContainsSubgraphIsomorphicTo(graph); return err },
	} {
		t.Run(name, func(t *testing.T) {
			if err := call(); !errors.Is(err, ErrClosed) {
				t.Fatalf("error = %v, want %v", err, ErrClosed)
			}
		})
	}
	if _, err := graph.Isomorphic(directed); err == nil {
		t.Fatal("Isomorphic(directedness mismatch) error = nil")
	}
	if _, err := graph.ContainsSubgraphIsomorphicTo(directed); err == nil {
		t.Fatal("ContainsSubgraphIsomorphicTo(directedness mismatch) error = nil")
	}
}

func TestIsomorphismReversedConcurrentCalls(t *testing.T) {
	left := testGraphFromEdges(t, 4, []Edge{{0, 1}, {1, 2}, {2, 3}}, false)
	right := testGraphFromEdges(t, 4, []Edge{{3, 2}, {2, 1}, {1, 0}}, false)

	var wg sync.WaitGroup
	errorsCh := make(chan error, 200)
	for i := 0; i < 100; i++ {
		wg.Add(2)
		go func() { defer wg.Done(); _, err := left.Isomorphic(right); errorsCh <- err }()
		go func() { defer wg.Done(); _, err := right.Isomorphic(left); errorsCh <- err }()
	}
	wg.Wait()
	close(errorsCh)
	for err := range errorsCh {
		if err != nil {
			t.Fatalf("concurrent Isomorphic error = %v", err)
		}
	}
}
