package igraph_test

import (
	"errors"
	"math"
	"reflect"
	"sync"
	"testing"

	igraph "github.com/h8gi/go-igraph"
)

func TestGraphAttributesTypedRoundTrip(t *testing.T) {
	graph, err := igraph.NewGraph()
	if err != nil {
		t.Fatal(err)
	}
	defer graph.Close()

	empty, err := graph.GraphAttributes()
	if err != nil {
		t.Fatal(err)
	}
	if empty == nil || len(empty) != 0 {
		t.Fatalf("GraphAttributes() = %#v, want non-nil empty slice", empty)
	}

	nameBytes := []byte("weight")
	valueBytes := []byte("owned")
	if err := graph.SetGraphNumericAttribute(string(nameBytes), 1.5); err != nil {
		t.Fatal(err)
	}
	if err := graph.SetGraphBooleanAttribute("active", true); err != nil {
		t.Fatal(err)
	}
	if err := graph.SetGraphStringAttribute("label", string(valueBytes)); err != nil {
		t.Fatal(err)
	}
	if err := graph.SetGraphStringAttribute("empty", ""); err != nil {
		t.Fatal(err)
	}
	nameBytes[0] = 'X'
	valueBytes[0] = 'X'

	metadata, err := graph.GraphAttributes()
	if err != nil {
		t.Fatal(err)
	}
	wantMetadata := []igraph.AttributeMetadata{
		{Name: "active", Scope: igraph.AttributeGraph, Type: igraph.AttributeBoolean},
		{Name: "empty", Scope: igraph.AttributeGraph, Type: igraph.AttributeString},
		{Name: "label", Scope: igraph.AttributeGraph, Type: igraph.AttributeString},
		{Name: "weight", Scope: igraph.AttributeGraph, Type: igraph.AttributeNumeric},
	}
	if !reflect.DeepEqual(metadata, wantMetadata) {
		t.Fatalf("GraphAttributes() = %#v, want %#v", metadata, wantMetadata)
	}

	if got, err := graph.GraphNumericAttribute("weight"); err != nil || got != 1.5 {
		t.Fatalf("GraphNumericAttribute(weight) = %v, %v", got, err)
	}
	if got, err := graph.GraphBooleanAttribute("active"); err != nil || !got {
		t.Fatalf("GraphBooleanAttribute(active) = %v, %v", got, err)
	}
	if got, err := graph.GraphStringAttribute("label"); err != nil || got != "owned" {
		t.Fatalf("GraphStringAttribute(label) = %q, %v", got, err)
	}
	if got, err := graph.GraphStringAttribute("empty"); err != nil || got != "" {
		t.Fatalf("GraphStringAttribute(empty) = %q, %v", got, err)
	}

	if err := graph.SetGraphNumericAttribute("weight", 2.5); err != nil {
		t.Fatal(err)
	}
	if got, err := graph.GraphNumericAttribute("weight"); err != nil || got != 2.5 {
		t.Fatalf("overwritten weight = %v, %v", got, err)
	}
	if err := graph.SetGraphStringAttribute("weight", "wrong"); !errors.Is(err, igraph.ErrAttributeTypeMismatch) {
		t.Fatalf("wrong-type overwrite error = %v", err)
	}
	if _, err := graph.GraphStringAttribute("weight"); !errors.Is(err, igraph.ErrAttributeTypeMismatch) {
		t.Fatalf("wrong-type getter error = %v", err)
	}
	if _, err := graph.GraphNumericAttribute("missing"); !errors.Is(err, igraph.ErrAttributeNotFound) {
		t.Fatalf("missing getter error = %v", err)
	}
	if err := graph.RemoveGraphAttribute("missing"); !errors.Is(err, igraph.ErrAttributeNotFound) {
		t.Fatalf("missing remove error = %v", err)
	}

	if err := graph.RemoveGraphAttribute("active"); err != nil {
		t.Fatal(err)
	}
	if _, err := graph.GraphBooleanAttribute("active"); !errors.Is(err, igraph.ErrAttributeNotFound) {
		t.Fatalf("removed getter error = %v", err)
	}
	if err := graph.RemoveAllGraphAttributes(); err != nil {
		t.Fatal(err)
	}
	remaining, err := graph.GraphAttributes()
	if err != nil {
		t.Fatal(err)
	}
	if remaining == nil || len(remaining) != 0 {
		t.Fatalf("remaining attributes = %#v", remaining)
	}
}

func TestGraphAttributesRejectInvalidInputAtomically(t *testing.T) {
	graph, err := igraph.NewGraph()
	if err != nil {
		t.Fatal(err)
	}
	defer graph.Close()
	if err := graph.SetGraphNumericAttribute("kept", 7); err != nil {
		t.Fatal(err)
	}

	invalidUTF8 := string([]byte{0xff})
	for _, name := range []string{"", "bad\x00name", invalidUTF8} {
		if err := graph.SetGraphNumericAttribute(name, 1); err == nil {
			t.Errorf("SetGraphNumericAttribute(%q) unexpectedly succeeded", name)
		}
		if _, err := graph.GraphNumericAttribute(name); err == nil {
			t.Errorf("GraphNumericAttribute(%q) unexpectedly succeeded", name)
		}
		if err := graph.RemoveGraphAttribute(name); err == nil {
			t.Errorf("RemoveGraphAttribute(%q) unexpectedly succeeded", name)
		}
	}
	for _, value := range []float64{math.NaN(), math.Inf(1), math.Inf(-1)} {
		if err := graph.SetGraphNumericAttribute("invalid", value); err == nil {
			t.Errorf("SetGraphNumericAttribute(invalid, %v) unexpectedly succeeded", value)
		}
	}
	for _, value := range []string{"bad\x00value", invalidUTF8} {
		if err := graph.SetGraphStringAttribute("invalid", value); err == nil {
			t.Errorf("SetGraphStringAttribute(invalid, %q) unexpectedly succeeded", value)
		}
	}

	metadata, err := graph.GraphAttributes()
	if err != nil {
		t.Fatal(err)
	}
	want := []igraph.AttributeMetadata{{Name: "kept", Scope: igraph.AttributeGraph, Type: igraph.AttributeNumeric}}
	if !reflect.DeepEqual(metadata, want) {
		t.Fatalf("attributes after rejected mutations = %#v, want %#v", metadata, want)
	}
}

func TestGraphAttributesClosedAndGoOwned(t *testing.T) {
	graph, err := igraph.NewGraph()
	if err != nil {
		t.Fatal(err)
	}
	if err := graph.SetGraphStringAttribute("name", "value"); err != nil {
		t.Fatal(err)
	}
	metadata, err := graph.GraphAttributes()
	if err != nil {
		t.Fatal(err)
	}
	value, err := graph.GraphStringAttribute("name")
	if err != nil {
		t.Fatal(err)
	}
	graph.Close()

	if metadata[0].Name != "name" || value != "value" {
		t.Fatalf("Go-owned results changed after Close: %#v, %q", metadata, value)
	}
	assertGraphAttributeMethodsClosed(t, graph)
	assertGraphAttributeMethodsClosed(t, nil)
}

func assertGraphAttributeMethodsClosed(t *testing.T, graph *igraph.Graph) {
	t.Helper()
	checks := []error{}
	_, err := graph.GraphAttributes()
	checks = append(checks, err)
	_, err = graph.GraphNumericAttribute("x")
	checks = append(checks, err)
	_, err = graph.GraphStringAttribute("x")
	checks = append(checks, err)
	_, err = graph.GraphBooleanAttribute("x")
	checks = append(checks, err)
	checks = append(checks,
		graph.SetGraphNumericAttribute("x", 1),
		graph.SetGraphStringAttribute("x", "x"),
		graph.SetGraphBooleanAttribute("x", true),
		graph.RemoveGraphAttribute("x"),
		graph.RemoveAllGraphAttributes(),
	)
	for i, err := range checks {
		if !errors.Is(err, igraph.ErrClosed) {
			t.Errorf("closed check %d error = %v, want ErrClosed", i, err)
		}
	}
}

func TestGraphAttributesConcurrentUseAndClose(t *testing.T) {
	graph, err := igraph.NewGraph()
	if err != nil {
		t.Fatal(err)
	}
	if err := graph.SetGraphNumericAttribute("counter", 0); err != nil {
		t.Fatal(err)
	}

	var workers sync.WaitGroup
	for worker := 0; worker < 8; worker++ {
		workers.Add(1)
		go func(worker int) {
			defer workers.Done()
			for i := 0; i < 100; i++ {
				if worker%2 == 0 {
					err := graph.SetGraphNumericAttribute("counter", float64(i))
					if err != nil && !errors.Is(err, igraph.ErrClosed) {
						t.Errorf("writer error = %v", err)
						return
					}
				} else {
					_, err := graph.GraphNumericAttribute("counter")
					if err != nil && !errors.Is(err, igraph.ErrClosed) {
						t.Errorf("reader error = %v", err)
						return
					}
				}
			}
		}(worker)
	}
	graph.Close()
	workers.Wait()
}
