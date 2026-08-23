package igraph

import (
	"errors"
	"testing"
)

func TestEdgeDiagnosticBoolFailurePaths(t *testing.T) {
	graph := mustDiagnosticGraph(t, false, 2, []Edge{{From: 0, To: 1}})
	defer graph.Close()
	failure := errors.New("injected edge diagnostic failure")

	initialization := defaultEdgeDiagnosticAdapters()
	initialization.newBool = func([]bool) (*boolVector, error) { return nil, failure }
	if result, err := graph.loopEdgeFlags(AllEdges(), &initialization); result != nil || !errors.Is(err, failure) {
		t.Fatalf("initialization result = %v, %v, want nil, injected error", result, err)
	}

	upstream := defaultEdgeDiagnosticAdapters()
	closed := 0
	upstream.closeBool = func(vector *boolVector) { closed++; vector.close() }
	upstream.multiple = func(*Graph, *boolVector, *cEdgeSelector) int { return 1 }
	if result, err := graph.multipleEdgeFlags(AllEdges(), &upstream); result != nil || err == nil {
		t.Fatalf("upstream result = %v, %v, want nil, error", result, err)
	}
	if closed != 1 {
		t.Fatalf("upstream close count = %d, want 1", closed)
	}

	conversion := defaultEdgeDiagnosticAdapters()
	conversion.boolSlice = func(*boolVector) ([]bool, error) { return nil, failure }
	if result, err := graph.mutualEdgeFlags(AllEdges(), false, &conversion); result != nil || !errors.Is(err, failure) {
		t.Fatalf("conversion result = %v, %v, want nil, injected error", result, err)
	}

	early := defaultEdgeDiagnosticAdapters()
	early.newBool = func([]bool) (*boolVector, error) { panic("empty selection initialized a vector") }
	result, err := graph.loopEdgeFlags(NoEdges(), &early)
	if err != nil || result == nil || len(result) != 0 {
		t.Fatalf("empty result = %v, %v, want non-nil empty, nil", result, err)
	}
}

func TestEdgeDiagnosticMultiplicityFailurePaths(t *testing.T) {
	graph := mustDiagnosticGraph(t, false, 2, []Edge{{From: 0, To: 1}})
	defer graph.Close()
	failure := errors.New("injected multiplicity failure")

	initialization := defaultEdgeDiagnosticAdapters()
	initialization.newInt = func([]int) (*intVector, error) { return nil, failure }
	if result, err := graph.edgeMultiplicities(AllEdges(), &initialization); result != nil || !errors.Is(err, failure) {
		t.Fatalf("initialization result = %v, %v, want nil, injected error", result, err)
	}

	upstream := defaultEdgeDiagnosticAdapters()
	closed := 0
	upstream.closeInt = func(vector *intVector) { closed++; vector.close() }
	upstream.countMultiple = func(*Graph, *intVector, *cEdgeSelector) int { return 1 }
	if result, err := graph.edgeMultiplicities(AllEdges(), &upstream); result != nil || err == nil {
		t.Fatalf("upstream result = %v, %v, want nil, error", result, err)
	}
	if closed != 1 {
		t.Fatalf("upstream close count = %d, want 1", closed)
	}

	conversion := defaultEdgeDiagnosticAdapters()
	conversion.intSlice = func(*intVector) ([]int, error) { return nil, failure }
	if result, err := graph.edgeMultiplicities(AllEdges(), &conversion); result != nil || !errors.Is(err, failure) {
		t.Fatalf("conversion result = %v, %v, want nil, injected error", result, err)
	}

	early := defaultEdgeDiagnosticAdapters()
	early.newInt = func([]int) (*intVector, error) { panic("empty selection initialized a vector") }
	result, err := graph.edgeMultiplicities(NoEdges(), &early)
	if err != nil || result == nil || len(result) != 0 {
		t.Fatalf("empty result = %v, %v, want non-nil empty, nil", result, err)
	}
}
