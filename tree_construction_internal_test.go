package igraph

import (
	"errors"
	"strings"
	"testing"
)

func TestTreeConstructionAdapterFailures(t *testing.T) {
	failure := errors.New("simulated vector failure")
	for name, construct := range map[string]func(*treeConstructionAdapters) error{
		"from Prufer": func(adapters *treeConstructionAdapters) error {
			_, err := newTreeFromPrufer(nil, adapters)
			return err
		},
		"from parents": func(adapters *treeConstructionAdapters) error {
			_, err := newTreeFromParents(nil, TreeOut, adapters)
			return err
		},
		"symmetric": func(adapters *treeConstructionAdapters) error {
			_, err := newSymmetricTree(nil, TreeOut, adapters)
			return err
		},
	} {
		t.Run(name+" initialization", func(t *testing.T) {
			adapters := defaultTreeConstructionAdapters()
			adapters.newInt = func([]int) (*intVector, error) { return nil, failure }
			if err := construct(&adapters); !errors.Is(err, failure) {
				t.Errorf("error = %v, want injected failure", err)
			}
		})
	}

	graph, err := NewTreeFromPrufer(nil)
	if err != nil {
		t.Fatalf("NewTreeFromPrufer failed: %v", err)
	}
	defer graph.Close()
	initialize := defaultTreeConstructionAdapters()
	initialize.newInt = func([]int) (*intVector, error) { return nil, failure }
	if _, err := graph.pruferSequence(&initialize); !errors.Is(err, failure) {
		t.Errorf("Prufer initialization error = %v", err)
	}

	convert := defaultTreeConstructionAdapters()
	convert.vectorSlice = func(*intVector) ([]int, error) { return nil, failure }
	if result, err := graph.pruferSequence(&convert); !errors.Is(err, failure) || result != nil {
		t.Errorf("Prufer conversion = %#v, %v", result, err)
	}
}

func TestTreeConstructionUpstreamFailures(t *testing.T) {
	failed := treeConstructionCallResult{code: 1}
	tests := []struct {
		name      string
		operation string
		invoke    func(*treeConstructionAdapters) error
		configure func(*treeConstructionAdapters)
	}{
		{"Prufer constructor", "construct tree from Prüfer sequence", func(a *treeConstructionAdapters) error { _, err := newTreeFromPrufer(nil, a); return err }, func(a *treeConstructionAdapters) {
			a.fromPrufer = func(*intVector) treeConstructionCallResult { return failed }
		}},
		{"Prufer conversion", "convert tree to Prüfer sequence", func(a *treeConstructionAdapters) error {
			g, _ := NewTreeFromPrufer(nil)
			defer g.Close()
			_, err := g.pruferSequence(a)
			return err
		}, func(a *treeConstructionAdapters) { a.toPrufer = func(*Graph, *intVector) int { return 1 } }},
		{"parent constructor", "construct tree from parent vector", func(a *treeConstructionAdapters) error { _, err := newTreeFromParents(nil, TreeOut, a); return err }, func(a *treeConstructionAdapters) {
			a.fromParents = func(*intVector, TreeMode) treeConstructionCallResult { return failed }
		}},
		{"symmetric constructor", "construct symmetric tree", func(a *treeConstructionAdapters) error { _, err := newSymmetricTree(nil, TreeOut, a); return err }, func(a *treeConstructionAdapters) {
			a.symmetricTree = func(*intVector, TreeMode) treeConstructionCallResult { return failed }
		}},
		{"regular constructor", "construct regular tree", func(a *treeConstructionAdapters) error { _, err := newRegularTree(1, 2, TreeOut, a); return err }, func(a *treeConstructionAdapters) {
			a.regularTree = func(int, int, TreeMode) treeConstructionCallResult { return failed }
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			adapters := defaultTreeConstructionAdapters()
			test.configure(&adapters)
			if err := test.invoke(&adapters); err == nil || !strings.Contains(err.Error(), test.operation) {
				t.Errorf("error = %v", err)
			}
		})
	}
}

func TestTreeConstructionTemporaryVectorsAreClosed(t *testing.T) {
	tests := []struct {
		name   string
		invoke func(*treeConstructionAdapters) error
	}{
		{"Prufer constructor", func(a *treeConstructionAdapters) error { _, err := newTreeFromPrufer(nil, a); return err }},
		{"parent constructor", func(a *treeConstructionAdapters) error { _, err := newTreeFromParents(nil, TreeOut, a); return err }},
		{"symmetric constructor", func(a *treeConstructionAdapters) error { _, err := newSymmetricTree(nil, TreeOut, a); return err }},
		{"Prufer conversion", func(a *treeConstructionAdapters) error {
			g, _ := NewTreeFromPrufer(nil)
			defer g.Close()
			_, err := g.pruferSequence(a)
			return err
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			adapters := defaultTreeConstructionAdapters()
			closed := 0
			adapters.closeInt = func(vector *intVector) {
				closed++
				vector.close()
			}
			if err := test.invoke(&adapters); err != nil {
				t.Fatalf("operation failed: %v", err)
			}
			if closed != 1 {
				t.Errorf("closed vectors = %d, want 1", closed)
			}
		})
	}
}
