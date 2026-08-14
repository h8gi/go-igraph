package igraph_test

import (
	"errors"
	"math"
	"reflect"
	"sync"
	"testing"

	igraph "github.com/h8gi/go-igraph"
)

func newAttributedTestGraph(t *testing.T) *igraph.Graph {
	t.Helper()
	graph, err := igraph.NewGraphFromEdges(3, []igraph.Edge{
		{From: 0, To: 0},
		{From: 0, To: 1},
		{From: 0, To: 1},
	}, false)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = graph.Close() })
	return graph
}

func TestVertexAttributesTypedRoundTrip(t *testing.T) {
	graph := newAttributedTestGraph(t)
	numbers := []float64{1, 2, 3}
	strings := []string{"", "one", "two"}
	booleans := []bool{true, false, true}
	if err := graph.SetVertexNumericAttributes("z-number", numbers); err != nil {
		t.Fatal(err)
	}
	if err := graph.SetVertexStringAttributes("m-string", strings); err != nil {
		t.Fatal(err)
	}
	if err := graph.SetVertexBooleanAttributes("a-boolean", booleans); err != nil {
		t.Fatal(err)
	}
	numbers[0] = 99
	strings[0] = "changed"
	booleans[0] = false

	wantMetadata := []igraph.AttributeMetadata{
		{Name: "a-boolean", Scope: igraph.AttributeVertex, Type: igraph.AttributeBoolean},
		{Name: "m-string", Scope: igraph.AttributeVertex, Type: igraph.AttributeString},
		{Name: "z-number", Scope: igraph.AttributeVertex, Type: igraph.AttributeNumeric},
	}
	metadata, err := graph.VertexAttributes()
	if err != nil || !reflect.DeepEqual(metadata, wantMetadata) {
		t.Fatalf("VertexAttributes() = %#v, %v; want %#v", metadata, err, wantMetadata)
	}
	assertFloatSlice(t, []float64{1, 2, 3}, graph.VertexNumericAttributes, "z-number")
	assertStringSlice(t, []string{"", "one", "two"}, graph.VertexStringAttributes, "m-string")
	assertBoolSlice(t, []bool{true, false, true}, graph.VertexBooleanAttributes, "a-boolean")

	if got, err := graph.VertexNumericAttribute("z-number", 2); err != nil || got != 3 {
		t.Fatalf("VertexNumericAttribute() = %v, %v", got, err)
	}
	if got, err := graph.VertexStringAttribute("m-string", 0); err != nil || got != "" {
		t.Fatalf("VertexStringAttribute() = %q, %v", got, err)
	}
	if got, err := graph.VertexBooleanAttribute("a-boolean", 1); err != nil || got {
		t.Fatalf("VertexBooleanAttribute() = %v, %v", got, err)
	}

	if err := graph.SetVertexNumericAttribute("z-number", 1, 20); err != nil {
		t.Fatal(err)
	}
	if err := graph.SetVertexStringAttribute("m-string", 2, "updated"); err != nil {
		t.Fatal(err)
	}
	if err := graph.SetVertexBooleanAttribute("a-boolean", 1, true); err != nil {
		t.Fatal(err)
	}
	assertFloatSlice(t, []float64{1, 20, 3}, graph.VertexNumericAttributes, "z-number")
	assertStringSlice(t, []string{"", "one", "updated"}, graph.VertexStringAttributes, "m-string")
	assertBoolSlice(t, []bool{true, true, true}, graph.VertexBooleanAttributes, "a-boolean")
}

func TestEdgeAttributesTypedRoundTripWithLoopsAndParallelEdges(t *testing.T) {
	graph := newAttributedTestGraph(t)
	numbers := []float64{10, 20, 30}
	strings := []string{"loop", "parallel-a", "parallel-b"}
	booleans := []bool{true, false, true}
	if err := graph.SetEdgeNumericAttributes("z-number", numbers); err != nil {
		t.Fatal(err)
	}
	if err := graph.SetEdgeStringAttributes("m-string", strings); err != nil {
		t.Fatal(err)
	}
	if err := graph.SetEdgeBooleanAttributes("a-boolean", booleans); err != nil {
		t.Fatal(err)
	}

	wantMetadata := []igraph.AttributeMetadata{
		{Name: "a-boolean", Scope: igraph.AttributeEdge, Type: igraph.AttributeBoolean},
		{Name: "m-string", Scope: igraph.AttributeEdge, Type: igraph.AttributeString},
		{Name: "z-number", Scope: igraph.AttributeEdge, Type: igraph.AttributeNumeric},
	}
	metadata, err := graph.EdgeAttributes()
	if err != nil || !reflect.DeepEqual(metadata, wantMetadata) {
		t.Fatalf("EdgeAttributes() = %#v, %v; want %#v", metadata, err, wantMetadata)
	}
	assertFloatSlice(t, numbers, graph.EdgeNumericAttributes, "z-number")
	assertStringSlice(t, strings, graph.EdgeStringAttributes, "m-string")
	assertBoolSlice(t, booleans, graph.EdgeBooleanAttributes, "a-boolean")

	if got, err := graph.EdgeStringAttribute("m-string", 0); err != nil || got != "loop" {
		t.Fatalf("loop string = %q, %v", got, err)
	}
	if got, err := graph.EdgeNumericAttribute("z-number", 1); err != nil || got != 20 {
		t.Fatalf("parallel edge 1 number = %v, %v", got, err)
	}
	if got, err := graph.EdgeNumericAttribute("z-number", 2); err != nil || got != 30 {
		t.Fatalf("parallel edge 2 number = %v, %v", got, err)
	}
	if err := graph.SetEdgeNumericAttribute("z-number", 2, 300); err != nil {
		t.Fatal(err)
	}
	if err := graph.SetEdgeStringAttribute("m-string", 1, "changed"); err != nil {
		t.Fatal(err)
	}
	if err := graph.SetEdgeBooleanAttribute("a-boolean", 1, true); err != nil {
		t.Fatal(err)
	}
	assertFloatSlice(t, []float64{10, 20, 300}, graph.EdgeNumericAttributes, "z-number")
	assertStringSlice(t, []string{"loop", "changed", "parallel-b"}, graph.EdgeStringAttributes, "m-string")
	assertBoolSlice(t, []bool{true, true, true}, graph.EdgeBooleanAttributes, "a-boolean")
}

func TestElementAttributesRejectInvalidInputsAtomically(t *testing.T) {
	graph := newAttributedTestGraph(t)
	if err := graph.SetVertexNumericAttributes("vertex-number", []float64{1, 2, 3}); err != nil {
		t.Fatal(err)
	}
	if err := graph.SetEdgeStringAttributes("edge-string", []string{"a", "b", "c"}); err != nil {
		t.Fatal(err)
	}

	for _, values := range [][]float64{nil, {}, {1, 2}} {
		if err := graph.SetVertexNumericAttributes("bad-length", values); err == nil {
			t.Errorf("SetVertexNumericAttributes(%v) unexpectedly succeeded", values)
		}
	}
	for _, values := range [][]string{nil, {}, {"a", "b"}} {
		if err := graph.SetEdgeStringAttributes("bad-length", values); err == nil {
			t.Errorf("SetEdgeStringAttributes(%v) unexpectedly succeeded", values)
		}
	}
	for _, values := range [][]bool{nil, {}, {true}} {
		if err := graph.SetVertexBooleanAttributes("bad-length", values); err == nil {
			t.Errorf("SetVertexBooleanAttributes(%v) unexpectedly succeeded", values)
		}
	}

	for _, id := range []int{-1, 3} {
		if _, err := graph.VertexNumericAttribute("vertex-number", id); err == nil {
			t.Errorf("VertexNumericAttribute ID %d unexpectedly succeeded", id)
		}
		if err := graph.SetVertexNumericAttribute("vertex-number", id, 1); err == nil {
			t.Errorf("SetVertexNumericAttribute ID %d unexpectedly succeeded", id)
		}
		if _, err := graph.EdgeStringAttribute("edge-string", id); err == nil {
			t.Errorf("EdgeStringAttribute ID %d unexpectedly succeeded", id)
		}
	}

	if err := graph.SetVertexNumericAttribute("missing", 0, 1); !errors.Is(err, igraph.ErrAttributeNotFound) {
		t.Errorf("scalar creation error = %v, want ErrAttributeNotFound", err)
	}
	if err := graph.SetVertexStringAttributes("vertex-number", []string{"a", "b", "c"}); !errors.Is(err, igraph.ErrAttributeTypeMismatch) {
		t.Errorf("cross-type overwrite error = %v", err)
	}
	if _, err := graph.VertexStringAttributes("vertex-number"); !errors.Is(err, igraph.ErrAttributeTypeMismatch) {
		t.Errorf("wrong-type getter error = %v", err)
	}
	if _, err := graph.EdgeNumericAttributes("missing"); !errors.Is(err, igraph.ErrAttributeNotFound) {
		t.Errorf("missing getter error = %v", err)
	}

	for _, value := range []float64{math.NaN(), math.Inf(1), math.Inf(-1)} {
		if err := graph.SetVertexNumericAttribute("vertex-number", 0, value); err == nil {
			t.Errorf("non-finite scalar %v unexpectedly succeeded", value)
		}
		if err := graph.SetVertexNumericAttributes("vertex-number", []float64{1, value, 3}); err == nil {
			t.Errorf("non-finite vector %v unexpectedly succeeded", value)
		}
	}
	invalidUTF8 := string([]byte{0xff})
	for _, value := range []string{"bad\x00value", invalidUTF8} {
		if err := graph.SetEdgeStringAttribute("edge-string", 0, value); err == nil {
			t.Errorf("invalid scalar string %q unexpectedly succeeded", value)
		}
		if err := graph.SetEdgeStringAttributes("edge-string", []string{"a", value, "c"}); err == nil {
			t.Errorf("invalid string vector %q unexpectedly succeeded", value)
		}
	}
	for _, name := range []string{"", "bad\x00name", invalidUTF8} {
		if err := graph.SetVertexNumericAttributes(name, []float64{1, 2, 3}); err == nil {
			t.Errorf("invalid name %q unexpectedly succeeded", name)
		}
	}

	assertFloatSlice(t, []float64{1, 2, 3}, graph.VertexNumericAttributes, "vertex-number")
	assertStringSlice(t, []string{"a", "b", "c"}, graph.EdgeStringAttributes, "edge-string")
}

func TestElementAttributesEmptyNilAndEmptySlices(t *testing.T) {
	graph, err := igraph.NewGraph()
	if err != nil {
		t.Fatal(err)
	}
	defer graph.Close()
	if err := graph.SetVertexNumericAttributes("nil", nil); err != nil {
		t.Fatal(err)
	}
	if err := graph.SetVertexBooleanAttributes("empty", []bool{}); err != nil {
		t.Fatal(err)
	}
	if err := graph.SetEdgeStringAttributes("nil", nil); err != nil {
		t.Fatal(err)
	}
	if err := graph.SetEdgeNumericAttributes("empty", []float64{}); err != nil {
		t.Fatal(err)
	}
	assertFloatSlice(t, []float64{}, graph.VertexNumericAttributes, "nil")
	assertBoolSlice(t, []bool{}, graph.VertexBooleanAttributes, "empty")
	assertStringSlice(t, []string{}, graph.EdgeStringAttributes, "nil")
	assertFloatSlice(t, []float64{}, graph.EdgeNumericAttributes, "empty")
}

func TestElementAttributeDefaultsAfterTopologyGrowth(t *testing.T) {
	graph, err := igraph.NewGraphFromEdges(1, nil, false)
	if err != nil {
		t.Fatal(err)
	}
	defer graph.Close()
	if err := graph.SetVertexNumericAttributes("number", []float64{1}); err != nil {
		t.Fatal(err)
	}
	if err := graph.SetVertexStringAttributes("string", []string{"present"}); err != nil {
		t.Fatal(err)
	}
	if err := graph.SetVertexBooleanAttributes("boolean", []bool{true}); err != nil {
		t.Fatal(err)
	}
	if err := graph.SetEdgeNumericAttributes("number", nil); err != nil {
		t.Fatal(err)
	}
	if err := graph.SetEdgeStringAttributes("string", nil); err != nil {
		t.Fatal(err)
	}
	if err := graph.SetEdgeBooleanAttributes("boolean", nil); err != nil {
		t.Fatal(err)
	}
	if err := graph.AddVertices(1); err != nil {
		t.Fatal(err)
	}
	if err := graph.AddEdge(0, 1); err != nil {
		t.Fatal(err)
	}

	vertexNumbers, err := graph.VertexNumericAttributes("number")
	if err != nil || len(vertexNumbers) != 2 || vertexNumbers[0] != 1 || !math.IsNaN(vertexNumbers[1]) {
		t.Fatalf("grown vertex numbers = %#v, %v", vertexNumbers, err)
	}
	assertStringSlice(t, []string{"present", ""}, graph.VertexStringAttributes, "string")
	assertBoolSlice(t, []bool{true, false}, graph.VertexBooleanAttributes, "boolean")
	edgeNumbers, err := graph.EdgeNumericAttributes("number")
	if err != nil || len(edgeNumbers) != 1 || !math.IsNaN(edgeNumbers[0]) {
		t.Fatalf("grown edge numbers = %#v, %v", edgeNumbers, err)
	}
	assertStringSlice(t, []string{""}, graph.EdgeStringAttributes, "string")
	assertBoolSlice(t, []bool{false}, graph.EdgeBooleanAttributes, "boolean")
}

func TestElementAttributeRemovalIsScopeIsolated(t *testing.T) {
	graph := newAttributedTestGraph(t)
	if err := graph.SetGraphStringAttribute("same", "graph"); err != nil {
		t.Fatal(err)
	}
	if err := graph.SetVertexStringAttributes("same", []string{"v0", "v1", "v2"}); err != nil {
		t.Fatal(err)
	}
	if err := graph.SetVertexBooleanAttributes("other", []bool{true, true, true}); err != nil {
		t.Fatal(err)
	}
	if err := graph.SetEdgeStringAttributes("same", []string{"e0", "e1", "e2"}); err != nil {
		t.Fatal(err)
	}
	if err := graph.RemoveVertexAttribute("same"); err != nil {
		t.Fatal(err)
	}
	if _, err := graph.VertexStringAttributes("same"); !errors.Is(err, igraph.ErrAttributeNotFound) {
		t.Fatalf("removed vertex attribute error = %v", err)
	}
	if got, err := graph.GraphStringAttribute("same"); err != nil || got != "graph" {
		t.Fatalf("graph attribute after vertex remove = %q, %v", got, err)
	}
	assertStringSlice(t, []string{"e0", "e1", "e2"}, graph.EdgeStringAttributes, "same")
	if err := graph.RemoveAllVertexAttributes(); err != nil {
		t.Fatal(err)
	}
	if metadata, err := graph.VertexAttributes(); err != nil || metadata == nil || len(metadata) != 0 {
		t.Fatalf("vertex metadata after remove all = %#v, %v", metadata, err)
	}
	if err := graph.RemoveAllEdgeAttributes(); err != nil {
		t.Fatal(err)
	}
	if metadata, err := graph.EdgeAttributes(); err != nil || metadata == nil || len(metadata) != 0 {
		t.Fatalf("edge metadata after remove all = %#v, %v", metadata, err)
	}
	if err := graph.RemoveEdgeAttribute("missing"); !errors.Is(err, igraph.ErrAttributeNotFound) {
		t.Fatalf("missing remove error = %v", err)
	}
}

func TestElementAttributeResultsAreGoOwnedAfterClose(t *testing.T) {
	graph := newAttributedTestGraph(t)
	if err := graph.SetVertexStringAttributes("names", []string{"a", "b", "c"}); err != nil {
		t.Fatal(err)
	}
	if err := graph.SetEdgeNumericAttributes("weights", []float64{1, 2, 3}); err != nil {
		t.Fatal(err)
	}
	vertexMetadata, err := graph.VertexAttributes()
	if err != nil {
		t.Fatal(err)
	}
	names, err := graph.VertexStringAttributes("names")
	if err != nil {
		t.Fatal(err)
	}
	weights, err := graph.EdgeNumericAttributes("weights")
	if err != nil {
		t.Fatal(err)
	}
	graph.Close()
	if vertexMetadata[0].Name != "names" || !reflect.DeepEqual(names, []string{"a", "b", "c"}) || !reflect.DeepEqual(weights, []float64{1, 2, 3}) {
		t.Fatalf("results changed after Close: %#v, %#v, %#v", vertexMetadata, names, weights)
	}
	assertElementAttributeMethodsClosed(t, graph)
	assertElementAttributeMethodsClosed(t, nil)
}

func assertElementAttributeMethodsClosed(t *testing.T, graph *igraph.Graph) {
	t.Helper()
	var checks []error
	_, err := graph.VertexAttributes()
	checks = append(checks, err)
	_, err = graph.EdgeAttributes()
	checks = append(checks, err)
	_, err = graph.VertexNumericAttribute("x", 0)
	checks = append(checks, err)
	_, err = graph.EdgeStringAttribute("x", 0)
	checks = append(checks, err)
	_, err = graph.VertexBooleanAttributes("x")
	checks = append(checks, err)
	_, err = graph.EdgeNumericAttributes("x")
	checks = append(checks, err)
	_, err = graph.VertexStringAttributes("x")
	checks = append(checks, err)
	_, err = graph.EdgeBooleanAttribute("x", 0)
	checks = append(checks, err)
	checks = append(checks,
		graph.SetVertexNumericAttribute("x", 0, 1),
		graph.SetEdgeStringAttribute("x", 0, "x"),
		graph.SetVertexBooleanAttribute("x", 0, true),
		graph.SetEdgeNumericAttributes("x", nil),
		graph.SetVertexStringAttributes("x", nil),
		graph.SetEdgeBooleanAttributes("x", nil),
		graph.RemoveVertexAttribute("x"),
		graph.RemoveEdgeAttribute("x"),
		graph.RemoveAllVertexAttributes(),
		graph.RemoveAllEdgeAttributes(),
	)
	for i, err := range checks {
		if !errors.Is(err, igraph.ErrClosed) {
			t.Errorf("closed check %d error = %v, want ErrClosed", i, err)
		}
	}
}

func TestElementAttributesConcurrentUseAndRepeatedClose(t *testing.T) {
	graph := newAttributedTestGraph(t)
	if err := graph.SetVertexNumericAttributes("values", []float64{0, 0, 0}); err != nil {
		t.Fatal(err)
	}
	if err := graph.SetEdgeBooleanAttributes("flags", []bool{false, false, false}); err != nil {
		t.Fatal(err)
	}
	var workers sync.WaitGroup
	for worker := 0; worker < 8; worker++ {
		workers.Add(1)
		go func(worker int) {
			defer workers.Done()
			for i := 0; i < 100; i++ {
				var err error
				switch worker % 4 {
				case 0:
					err = graph.SetVertexNumericAttribute("values", worker%3, float64(i))
				case 1:
					_, err = graph.VertexNumericAttributes("values")
				case 2:
					err = graph.SetEdgeBooleanAttribute("flags", worker%3, i%2 == 0)
				case 3:
					_, err = graph.EdgeBooleanAttributes("flags")
				}
				if err != nil && !errors.Is(err, igraph.ErrClosed) {
					t.Errorf("worker error = %v", err)
					return
				}
			}
		}(worker)
	}
	graph.Close()
	graph.Close()
	workers.Wait()
}

func assertFloatSlice(
	t *testing.T,
	want []float64,
	getter func(string) ([]float64, error),
	name string,
) {
	t.Helper()
	got, err := getter(name)
	if err != nil || got == nil || !reflect.DeepEqual(got, want) {
		t.Fatalf("float values = %#v, %v; want %#v", got, err, want)
	}
}

func assertStringSlice(
	t *testing.T,
	want []string,
	getter func(string) ([]string, error),
	name string,
) {
	t.Helper()
	got, err := getter(name)
	if err != nil || got == nil || !reflect.DeepEqual(got, want) {
		t.Fatalf("string values = %#v, %v; want %#v", got, err, want)
	}
}

func assertBoolSlice(
	t *testing.T,
	want []bool,
	getter func(string) ([]bool, error),
	name string,
) {
	t.Helper()
	got, err := getter(name)
	if err != nil || got == nil || !reflect.DeepEqual(got, want) {
		t.Fatalf("Boolean values = %#v, %v; want %#v", got, err, want)
	}
}
