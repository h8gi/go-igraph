package igraph

import (
	"errors"
	"testing"
)

func TestGraphAttributeListInitializationAndUpstreamFailuresCleanup(t *testing.T) {
	graph, err := NewGraph()
	if err != nil {
		t.Fatal(err)
	}
	defer graph.Close()

	injected := errors.New("injected failure")
	graph.mu.RLock()
	defer graph.mu.RUnlock()

	if _, err := graphAttributesLockedWithHooks(&graph.graph, graphAttributeListHooks{
		newNames: func() (*stringVector, error) { return nil, injected },
	}); !errors.Is(err, injected) {
		t.Fatalf("name initialization failure = %v", err)
	}

	for _, test := range []struct {
		name      string
		hooks     graphAttributeListHooks
		wantNames int
		wantTypes int
	}{
		{
			name: "type initialization",
			hooks: graphAttributeListHooks{
				newTypes: func() (*intVector, error) { return nil, injected },
			},
			wantNames: 1,
		},
		{
			name: "upstream list",
			hooks: graphAttributeListHooks{
				list: func() error { return injected },
			},
			wantNames: 1,
			wantTypes: 1,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			namesClosed := 0
			typesClosed := 0
			test.hooks.namesClose = func() { namesClosed++ }
			test.hooks.typesClose = func() { typesClosed++ }
			if _, err := graphAttributesLockedWithHooks(&graph.graph, test.hooks); !errors.Is(err, injected) {
				t.Fatalf("failure = %v", err)
			}
			if namesClosed != test.wantNames || typesClosed != test.wantTypes {
				t.Fatalf(
					"cleanup counts = names %d, types %d; want %d, %d",
					namesClosed,
					typesClosed,
					test.wantNames,
					test.wantTypes,
				)
			}
		})
	}
}

func TestGraphAttributeMutationFailureIsAtomic(t *testing.T) {
	graph, err := NewGraph()
	if err != nil {
		t.Fatal(err)
	}
	defer graph.Close()

	if err := graph.SetGraphNumericAttribute("number", 1); err != nil {
		t.Fatal(err)
	}
	if err := graph.SetGraphStringAttribute("string", "before"); err != nil {
		t.Fatal(err)
	}
	if err := graph.SetGraphBooleanAttribute("boolean", false); err != nil {
		t.Fatal(err)
	}

	injected := errors.New("injected upstream failure")
	if err := graph.setGraphNumericAttribute("number", 2, func() error { return injected }); !errors.Is(err, injected) {
		t.Fatalf("numeric failure = %v", err)
	}
	if err := graph.setGraphStringAttribute("string", "after", func() error { return injected }); !errors.Is(err, injected) {
		t.Fatalf("string failure = %v", err)
	}
	if err := graph.setGraphBooleanAttribute("boolean", true, func() error { return injected }); !errors.Is(err, injected) {
		t.Fatalf("Boolean failure = %v", err)
	}

	if got, err := graph.GraphNumericAttribute("number"); err != nil || got != 1 {
		t.Fatalf("number after failure = %v, %v", got, err)
	}
	if got, err := graph.GraphStringAttribute("string"); err != nil || got != "before" {
		t.Fatalf("string after failure = %q, %v", got, err)
	}
	if got, err := graph.GraphBooleanAttribute("boolean"); err != nil || got {
		t.Fatalf("boolean after failure = %v, %v", got, err)
	}
}
