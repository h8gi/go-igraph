package igraph

import (
	"errors"
	"reflect"
	"testing"
)

func TestEdgeDiagnosticsKnownAnswers(t *testing.T) {
	graph := mustDiagnosticGraph(t, true, 3, []Edge{
		{From: 0, To: 0},
		{From: 0, To: 1},
		{From: 0, To: 1},
		{From: 1, To: 0},
		{From: 2, To: 2},
		{From: 1, To: 2},
	})

	if got, err := graph.HasLoopEdges(); err != nil || !got {
		t.Fatalf("HasLoopEdges() = %v, %v, want true, nil", got, err)
	}
	if got, err := graph.LoopEdgeCount(); err != nil || got != 2 {
		t.Fatalf("LoopEdgeCount() = %d, %v, want 2, nil", got, err)
	}
	if got, err := graph.HasMultipleEdges(); err != nil || !got {
		t.Fatalf("HasMultipleEdges() = %v, %v, want true, nil", got, err)
	}
	if got, err := graph.HasMutualEdges(false); err != nil || !got {
		t.Fatalf("HasMutualEdges(false) = %v, %v, want true, nil", got, err)
	}

	assertBoolSliceResult(t, "LoopEdgeFlags", callBoolSlice(func() ([]bool, error) { return graph.LoopEdgeFlags(AllEdges()) }),
		[]bool{true, false, false, false, true, false})
	assertIntSliceResult(t, "EdgeMultiplicities", callIntSlice(func() ([]int, error) { return graph.EdgeMultiplicities(AllEdges()) }),
		[]int{1, 2, 2, 1, 1, 1})
	assertBoolSliceResult(t, "MultipleEdgeFlags", callBoolSlice(func() ([]bool, error) { return graph.MultipleEdgeFlags(AllEdges()) }),
		[]bool{false, false, true, false, false, false})
	assertBoolSliceResult(t, "MutualEdgeFlags(false)", callBoolSlice(func() ([]bool, error) { return graph.MutualEdgeFlags(AllEdges(), false) }),
		[]bool{false, true, true, true, false, false})
	assertBoolSliceResult(t, "MutualEdgeFlags(true)", callBoolSlice(func() ([]bool, error) { return graph.MutualEdgeFlags(AllEdges(), true) }),
		[]bool{true, true, true, true, true, false})

	selected, err := EdgeIDs(4, 2, 4, 0)
	if err != nil {
		t.Fatal(err)
	}
	assertBoolSliceResult(t, "selected loops", callBoolSlice(func() ([]bool, error) { return graph.LoopEdgeFlags(selected) }),
		[]bool{true, false, true, true})
	assertIntSliceResult(t, "selected multiplicities", callIntSlice(func() ([]int, error) { return graph.EdgeMultiplicities(selected) }),
		[]int{1, 2, 1, 1})

	if err := graph.Close(); err != nil {
		t.Fatal(err)
	}
	assertBoolSliceError(t, callBoolSlice(func() ([]bool, error) { return graph.LoopEdgeFlags(AllEdges()) }), ErrClosed)
	assertIntSliceError(t, callIntSlice(func() ([]int, error) { return graph.EdgeMultiplicities(AllEdges()) }), ErrClosed)
	if _, err := graph.HasLoopEdges(); !errors.Is(err, ErrClosed) {
		t.Fatalf("HasLoopEdges() after Close error = %v, want ErrClosed", err)
	}
	if _, err := graph.LoopEdgeCount(); !errors.Is(err, ErrClosed) {
		t.Fatalf("LoopEdgeCount() after Close error = %v, want ErrClosed", err)
	}
	if _, err := graph.HasMultipleEdges(); !errors.Is(err, ErrClosed) {
		t.Fatalf("HasMultipleEdges() after Close error = %v, want ErrClosed", err)
	}
	if _, err := graph.HasMutualEdges(false); !errors.Is(err, ErrClosed) {
		t.Fatalf("HasMutualEdges() after Close error = %v, want ErrClosed", err)
	}
}

func TestEdgeDiagnosticsEmptyAndUndirected(t *testing.T) {
	empty := mustDiagnosticGraph(t, true, 0, nil)
	defer empty.Close()
	if got, err := empty.HasLoopEdges(); err != nil || got {
		t.Fatalf("empty HasLoopEdges() = %v, %v, want false, nil", got, err)
	}
	if got, err := empty.LoopEdgeCount(); err != nil || got != 0 {
		t.Fatalf("empty LoopEdgeCount() = %d, %v, want 0, nil", got, err)
	}
	if got, err := empty.HasMultipleEdges(); err != nil || got {
		t.Fatalf("empty HasMultipleEdges() = %v, %v, want false, nil", got, err)
	}
	if got, err := empty.HasMutualEdges(true); err != nil || got {
		t.Fatalf("empty HasMutualEdges() = %v, %v, want false, nil", got, err)
	}
	assertBoolSliceResult(t, "empty loops", callBoolSlice(func() ([]bool, error) { return empty.LoopEdgeFlags(NoEdges()) }), []bool{})
	assertIntSliceResult(t, "empty multiplicities", callIntSlice(func() ([]int, error) { return empty.EdgeMultiplicities(NoEdges()) }), []int{})

	undirected := mustDiagnosticGraph(t, false, 2, []Edge{{From: 0, To: 1}})
	defer undirected.Close()
	if got, err := undirected.HasMutualEdges(false); err != nil || !got {
		t.Fatalf("undirected HasMutualEdges(false) = %v, %v, want true, nil", got, err)
	}
	assertBoolSliceResult(t, "undirected mutual", callBoolSlice(func() ([]bool, error) { return undirected.MutualEdgeFlags(AllEdges(), false) }), []bool{true})
}

func TestEdgeDiagnosticsRejectInvalidSelectors(t *testing.T) {
	graph := mustDiagnosticGraph(t, false, 2, []Edge{{From: 0, To: 1}})
	defer graph.Close()
	invalid := EdgeSelector{kind: edgeSelectorIDs, ids: []int{1}}
	assertBoolSliceError(t, callBoolSlice(func() ([]bool, error) { return graph.LoopEdgeFlags(invalid) }), nil)
	assertIntSliceError(t, callIntSlice(func() ([]int, error) { return graph.EdgeMultiplicities(invalid) }), nil)
	assertBoolSliceError(t, callBoolSlice(func() ([]bool, error) { return graph.MultipleEdgeFlags(invalid) }), nil)
	assertBoolSliceError(t, callBoolSlice(func() ([]bool, error) { return graph.MutualEdgeFlags(invalid, false) }), nil)
}

type boolSliceCall struct {
	value []bool
	err   error
}

type intSliceCall struct {
	value []int
	err   error
}

func callBoolSlice(call func() ([]bool, error)) boolSliceCall {
	value, err := call()
	return boolSliceCall{value: value, err: err}
}

func callIntSlice(call func() ([]int, error)) intSliceCall {
	value, err := call()
	return intSliceCall{value: value, err: err}
}

func assertBoolSliceResult(t *testing.T, name string, got boolSliceCall, want []bool) {
	t.Helper()
	if got.err != nil || !reflect.DeepEqual(got.value, want) {
		t.Fatalf("%s = %v, %v, want %v, nil", name, got.value, got.err, want)
	}
}

func assertIntSliceResult(t *testing.T, name string, got intSliceCall, want []int) {
	t.Helper()
	if got.err != nil || !reflect.DeepEqual(got.value, want) {
		t.Fatalf("%s = %v, %v, want %v, nil", name, got.value, got.err, want)
	}
}

func assertBoolSliceError(t *testing.T, got boolSliceCall, target error) {
	t.Helper()
	if got.value != nil || got.err == nil || target != nil && !errors.Is(got.err, target) {
		t.Fatalf("result = %v, %v, want nil, error matching %v", got.value, got.err, target)
	}
}

func assertIntSliceError(t *testing.T, got intSliceCall, target error) {
	t.Helper()
	if got.value != nil || got.err == nil || target != nil && !errors.Is(got.err, target) {
		t.Fatalf("result = %v, %v, want nil, error matching %v", got.value, got.err, target)
	}
}

func mustDiagnosticGraph(t *testing.T, directed bool, vertices int, edges []Edge) *Graph {
	t.Helper()
	graph, err := NewGraphFromEdges(vertices, edges, directed)
	if err != nil {
		t.Fatal(err)
	}
	return graph
}
