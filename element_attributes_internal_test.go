package igraph

import (
	"errors"
	"reflect"
	"testing"
)

func TestElementAttributeVectorReadFailuresCleanup(t *testing.T) {
	graph, err := NewGraphFromEdges(2, []Edge{{From: 0, To: 1}}, false)
	if err != nil {
		t.Fatal(err)
	}
	defer graph.Close()
	if err := graph.SetVertexNumericAttributes("values", []float64{1, 2}); err != nil {
		t.Fatal(err)
	}

	injected := errors.New("injected failure")
	graph.mu.RLock()
	defer graph.mu.RUnlock()
	if _, err := numericElementAttributesLocked(
		&graph.graph,
		AttributeVertex,
		"values",
		numericElementReadHooks{
			newResult: func() (*realVector, error) { return nil, injected },
		},
	); !errors.Is(err, injected) {
		t.Fatalf("result initialization failure = %v", err)
	}

	closed := 0
	if _, err := numericElementAttributesLocked(
		&graph.graph,
		AttributeVertex,
		"values",
		numericElementReadHooks{
			read:        func() error { return injected },
			resultClose: func() { closed++ },
		},
	); !errors.Is(err, injected) {
		t.Fatalf("upstream read failure = %v", err)
	}
	if closed != 1 {
		t.Fatalf("result cleanup count = %d, want 1", closed)
	}
}

func TestElementAttributeMutationFailuresAreAtomic(t *testing.T) {
	graph, err := NewGraphFromEdges(2, []Edge{{From: 0, To: 1}}, false)
	if err != nil {
		t.Fatal(err)
	}
	defer graph.Close()
	if err := graph.SetVertexNumericAttributes("numbers", []float64{1, 2}); err != nil {
		t.Fatal(err)
	}
	if err := graph.SetVertexStringAttributes("strings", []string{"a", "b"}); err != nil {
		t.Fatal(err)
	}
	if err := graph.SetEdgeBooleanAttributes("booleans", []bool{false}); err != nil {
		t.Fatal(err)
	}

	injected := errors.New("injected upstream failure")
	if err := graph.setElementNumericAttributes(
		AttributeVertex,
		"numbers",
		[]float64{10, 20},
		func() error { return injected },
	); !errors.Is(err, injected) {
		t.Fatalf("numeric vector failure = %v", err)
	}
	if err := graph.setElementStringAttribute(
		AttributeVertex,
		"strings",
		1,
		"changed",
		func() error { return injected },
	); !errors.Is(err, injected) {
		t.Fatalf("string scalar failure = %v", err)
	}
	if err := graph.setElementBooleanAttributes(
		AttributeEdge,
		"booleans",
		[]bool{true},
		func() error { return injected },
	); !errors.Is(err, injected) {
		t.Fatalf("Boolean vector failure = %v", err)
	}

	numbers, err := graph.VertexNumericAttributes("numbers")
	if err != nil || !reflect.DeepEqual(numbers, []float64{1, 2}) {
		t.Fatalf("numbers after failure = %#v, %v", numbers, err)
	}
	strings, err := graph.VertexStringAttributes("strings")
	if err != nil || !reflect.DeepEqual(strings, []string{"a", "b"}) {
		t.Fatalf("strings after failure = %#v, %v", strings, err)
	}
	booleans, err := graph.EdgeBooleanAttributes("booleans")
	if err != nil || !reflect.DeepEqual(booleans, []bool{false}) {
		t.Fatalf("booleans after failure = %#v, %v", booleans, err)
	}
}

func TestElementAttributeInternalScopeValidation(t *testing.T) {
	graph, err := NewGraph()
	if err != nil {
		t.Fatal(err)
	}
	defer graph.Close()
	invalid := AttributeScope(255)
	if _, err := attributeMetadataLocked(&graph.graph, invalid); err == nil {
		t.Fatal("attributeMetadataLocked() unexpectedly accepted invalid scope")
	}
	if _, err := elementCountLocked(&graph.graph, invalid); err == nil {
		t.Fatal("elementCountLocked() unexpectedly accepted invalid scope")
	}
}
