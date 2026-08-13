package igraph

import (
	"errors"
	"math"
	"reflect"
	"testing"
)

func TestMotifInitializationUpstreamAndConversionFailures(t *testing.T) {
	graph, err := NewGraphFromEdges(3, []Edge{{From: 0, To: 1}}, true)
	if err != nil {
		t.Fatal(err)
	}
	defer graph.Close()
	failure := errors.New("injected motif failure")

	for _, adjacent := range []bool{false, true} {
		adapters := defaultMotifAdapters()
		adapters.initializeReal = func() (*realVector, error) { return nil, failure }
		if _, err := callRealMotif(graph, adjacent, &adapters); !errors.Is(err, failure) {
			t.Errorf("real initialization adjacent=%t error = %v", adjacent, err)
		}
	}
	listInitialization := defaultMotifAdapters()
	listInitialization.initializeInt = func() (*intVector, error) { return nil, failure }
	if _, err := graph.trianglesList(&listInitialization); !errors.Is(err, failure) {
		t.Errorf("list initialization error = %v", err)
	}

	for operation := 0; operation < 5; operation++ {
		adapters := defaultMotifAdapters()
		switch operation {
		case 0:
			adapters.dyadCall = func(*Graph) ([3]float64, int) { return [3]float64{}, 1 }
		case 1:
			adapters.triadCall = func(*Graph, *realVector) int { return 1 }
		case 2:
			adapters.adjacentCall = func(*Graph, *realVector, *cVertexSelector) int { return 1 }
		case 3:
			adapters.countCall = func(*Graph) (float64, int) { return 0, 1 }
		case 4:
			adapters.listCall = func(*Graph, *intVector) int { return 1 }
		}
		if err := callMotifOperation(graph, operation, &adapters); err == nil {
			t.Errorf("upstream operation %d returned nil error", operation)
		}
	}

	for _, adjacent := range []bool{false, true} {
		adapters := defaultMotifAdapters()
		adapters.convertReal = func(*realVector) ([]float64, error) { return nil, failure }
		if _, err := callRealMotif(graph, adjacent, &adapters); !errors.Is(err, failure) {
			t.Errorf("real conversion adjacent=%t error = %v", adjacent, err)
		}
	}
	listConversion := defaultMotifAdapters()
	listConversion.convertInt = func(*intVector) ([]int, error) { return nil, failure }
	if _, err := graph.trianglesList(&listConversion); !errors.Is(err, failure) {
		t.Errorf("list conversion error = %v", err)
	}
}

func TestMotifTemporaryResultsCloseOnEveryPostInitializationPath(t *testing.T) {
	graph, err := NewGraphFromEdges(3, nil, true)
	if err != nil {
		t.Fatal(err)
	}
	defer graph.Close()
	for _, operation := range []int{1, 2, 4} {
		closed := 0
		adapters := defaultMotifAdapters()
		adapters.closeReal = func(vector *realVector) { closed++; vector.close() }
		adapters.closeInt = func(vector *intVector) { closed++; vector.close() }
		switch operation {
		case 1:
			adapters.triadCall = func(*Graph, *realVector) int { return 1 }
		case 2:
			adapters.adjacentCall = func(*Graph, *realVector, *cVertexSelector) int { return 1 }
		case 4:
			adapters.listCall = func(*Graph, *intVector) int { return 1 }
		}
		_ = callMotifOperation(graph, operation, &adapters)
		if closed != 1 {
			t.Errorf("operation %d closed %d temporary results", operation, closed)
		}
	}
}

func TestCheckedMotifCountRejectsInvalidValues(t *testing.T) {
	invalid := []float64{-1, 0.5, math.NaN(), math.Inf(1), maximumExactMotifCount + 2}
	for _, value := range invalid {
		if _, err := checkedMotifCount(value, "test"); err == nil {
			t.Errorf("checkedMotifCount(%v) returned nil error", value)
		}
	}
	for _, value := range []float64{0, 1, maximumExactMotifCount} {
		if got, err := checkedMotifCount(value, "test"); err != nil || got != int64(value) {
			t.Errorf("checkedMotifCount(%v) = %d, %v", value, got, err)
		}
	}
	if _, err := checkedMotifCounts(make([]float64, 15), 16, "test"); err == nil {
		t.Error("checkedMotifCounts accepted wrong length")
	}
}

func TestMotifMalformedConvertedResults(t *testing.T) {
	graph, err := NewGraphFromEdges(3, nil, true)
	if err != nil {
		t.Fatal(err)
	}
	defer graph.Close()

	triad := defaultMotifAdapters()
	triad.convertReal = func(*realVector) ([]float64, error) { return make([]float64, 15), nil }
	if _, err := graph.triadCensus(&triad); err == nil {
		t.Error("triad census accepted malformed result")
	}
	adjacent := defaultMotifAdapters()
	adjacent.convertReal = func(*realVector) ([]float64, error) { return []float64{1}, nil }
	if _, err := graph.adjacentTrianglesCount(AllVertices(), &adjacent); err == nil {
		t.Error("adjacent triangle count accepted malformed result")
	}
	list := defaultMotifAdapters()
	list.convertInt = func(*intVector) ([]int, error) { return []int{0, 1}, nil }
	if _, err := graph.trianglesList(&list); err == nil {
		t.Error("triangle list accepted malformed result")
	}
	dyad := defaultMotifAdapters()
	dyad.dyadCall = func(*Graph) ([3]float64, int) { return [3]float64{1, math.NaN(), 1}, 0 }
	if _, err := graph.dyadCensus(&dyad); err == nil {
		t.Error("dyad census accepted malformed count")
	}
	count := defaultMotifAdapters()
	count.countCall = func(*Graph) (float64, int) { return math.Inf(1), 0 }
	if _, err := graph.trianglesCount(&count); err == nil {
		t.Error("triangle count accepted malformed count")
	}
}

func callRealMotif(graph *Graph, adjacent bool, adapters *motifAdapters) ([]int64, error) {
	if adjacent {
		return graph.adjacentTrianglesCount(AllVertices(), adapters)
	}
	return graph.triadCensus(adapters)
}

func callMotifOperation(graph *Graph, operation int, adapters *motifAdapters) error {
	switch operation {
	case 0:
		_, err := graph.dyadCensus(adapters)
		return err
	case 1:
		_, err := graph.triadCensus(adapters)
		return err
	case 2:
		_, err := graph.adjacentTrianglesCount(AllVertices(), adapters)
		return err
	case 3:
		_, err := graph.trianglesCount(adapters)
		return err
	default:
		_, err := graph.trianglesList(adapters)
		return err
	}
}

func TestAdjacentTriangleSelectorOrder(t *testing.T) {
	graph, err := NewGraphFromEdges(3, []Edge{{0, 1}, {1, 2}, {2, 0}}, false)
	if err != nil {
		t.Fatal(err)
	}
	defer graph.Close()
	selector, _ := VertexIDs(2, 0, 2)
	got, err := graph.AdjacentTrianglesCount(selector)
	if err != nil || !reflect.DeepEqual(got, []int64{1, 1, 1}) {
		t.Fatalf("AdjacentTrianglesCount = %v, %v", got, err)
	}
}
