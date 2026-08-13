package igraph

import (
	"errors"
	"math"
	"testing"
)

func TestGraphletInitializationAndUpstreamFailures(t *testing.T) {
	graph, err := NewGraphFromEdges(3, []Edge{{0, 1}, {1, 2}, {2, 0}}, false)
	if err != nil {
		t.Fatal(err)
	}
	defer graph.Close()
	failure := errors.New("injected graphlet failure")

	for operation := 0; operation < 3; operation++ {
		adapters := defaultGraphletAdapters()
		adapters.createReal = func([]float64) (*realVector, error) { return nil, failure }
		if err := callGraphletOperation(graph, operation, &adapters); !errors.Is(err, failure) {
			t.Errorf("weight initialization operation %d error = %v", operation, err)
		}
	}

	for operation := 0; operation < 3; operation++ {
		adapters := defaultGraphletAdapters()
		adapters.simpleCall = func(*Graph) (bool, int) { return false, 1 }
		if err := callGraphletOperation(graph, operation, &adapters); err == nil {
			t.Errorf("shape check operation %d returned nil upstream error", operation)
		}
	}

	for operation := 0; operation < 2; operation++ {
		adapters := defaultGraphletAdapters()
		adapters.initializeList = func() (*intVectorList, error) { return nil, failure }
		if err := callGraphletOperation(graph, operation, &adapters); !errors.Is(err, failure) {
			t.Errorf("list initialization operation %d error = %v", operation, err)
		}
	}

	for operation := 0; operation < 2; operation++ {
		closed := 0
		adapters := defaultGraphletAdapters()
		adapters.initializeReal = func() (*realVector, error) { return nil, failure }
		adapters.closeList = func(list *intVectorList) { closed++; list.close() }
		if err := callGraphletOperation(graph, operation, &adapters); !errors.Is(err, failure) {
			t.Errorf("value initialization operation %d error = %v", operation, err)
		}
		if closed != 1 {
			t.Errorf("value initialization operation %d closed %d lists", operation, closed)
		}
	}

	listInitialization := defaultGraphletAdapters()
	listInitialization.createList = func([][]int) (*intVectorList, error) { return nil, failure }
	if err := callGraphletOperation(graph, 2, &listInitialization); !errors.Is(err, failure) {
		t.Errorf("project list initialization error = %v", err)
	}
	muInitialization := defaultGraphletAdapters()
	createCalls, realClosed, listClosed := 0, 0, 0
	muInitialization.createReal = func(values []float64) (*realVector, error) {
		createCalls++
		if createCalls == 2 {
			return nil, failure
		}
		return newRealVector(values)
	}
	muInitialization.closeReal = func(vector *realVector) { realClosed++; vector.close() }
	muInitialization.closeList = func(list *intVectorList) { listClosed++; list.close() }
	if err := callGraphletOperation(graph, 2, &muInitialization); !errors.Is(err, failure) {
		t.Errorf("project Mu initialization error = %v", err)
	}
	if realClosed != 1 || listClosed != 1 {
		t.Errorf("project Mu initialization closes = real %d, list %d", realClosed, listClosed)
	}

	for operation := 0; operation < 3; operation++ {
		adapters := defaultGraphletAdapters()
		switch operation {
		case 0:
			adapters.graphletsCall = func(*Graph, *realVector, *intVectorList, *realVector, int) int { return 1 }
		case 1:
			adapters.candidateCall = func(*Graph, *realVector, *intVectorList, *realVector) int { return 1 }
		case 2:
			adapters.projectCall = func(*Graph, *realVector, *intVectorList, *realVector, bool, int) int { return 1 }
		}
		if err := callGraphletOperation(graph, operation, &adapters); err == nil {
			t.Errorf("upstream operation %d returned nil error", operation)
		}
	}
}

func TestGraphletConversionAndMalformedFailures(t *testing.T) {
	graph, err := NewGraphFromEdges(3, []Edge{{0, 1}, {1, 2}, {2, 0}}, false)
	if err != nil {
		t.Fatal(err)
	}
	defer graph.Close()
	failure := errors.New("injected graphlet conversion failure")

	for operation := 0; operation < 2; operation++ {
		listConversion := defaultGraphletAdapters()
		listConversion.convertList = func(*intVectorList) ([][]int, error) { return nil, failure }
		if err := callGraphletOperation(graph, operation, &listConversion); !errors.Is(err, failure) {
			t.Errorf("list conversion operation %d error = %v", operation, err)
		}
		realConversion := defaultGraphletAdapters()
		realConversion.convertReal = func(*realVector) ([]float64, error) { return nil, failure }
		if err := callGraphletOperation(graph, operation, &realConversion); !errors.Is(err, failure) {
			t.Errorf("real conversion operation %d error = %v", operation, err)
		}
	}
	projectConversion := defaultGraphletAdapters()
	projectConversion.convertReal = func(*realVector) ([]float64, error) { return nil, failure }
	if err := callGraphletOperation(graph, 2, &projectConversion); !errors.Is(err, failure) {
		t.Errorf("project conversion error = %v", err)
	}

	malformedCliques := defaultGraphletAdapters()
	malformedCliques.convertList = func(*intVectorList) ([][]int, error) { return [][]int{{0}}, nil }
	malformedCliques.convertReal = func(*realVector) ([]float64, error) { return []float64{1}, nil }
	if err := callGraphletOperation(graph, 0, &malformedCliques); err == nil {
		t.Error("Graphlets accepted malformed clique")
	}
	misaligned := defaultGraphletAdapters()
	misaligned.convertList = func(*intVectorList) ([][]int, error) { return [][]int{{0, 1}}, nil }
	misaligned.convertReal = func(*realVector) ([]float64, error) { return nil, nil }
	if err := callGraphletOperation(graph, 1, &misaligned); err == nil {
		t.Error("GraphletsCandidateBasis accepted misaligned values")
	}
	malformedValue := defaultGraphletAdapters()
	malformedValue.convertList = func(*intVectorList) ([][]int, error) { return [][]int{{0, 1}}, nil }
	malformedValue.convertReal = func(*realVector) ([]float64, error) { return []float64{math.NaN()}, nil }
	if err := callGraphletOperation(graph, 0, &malformedValue); err == nil {
		t.Error("Graphlets accepted malformed coefficient")
	}
	projectValue := defaultGraphletAdapters()
	projectValue.convertReal = func(*realVector) ([]float64, error) { return []float64{math.Inf(1)}, nil }
	if err := callGraphletOperation(graph, 2, &projectValue); err == nil {
		t.Error("GraphletsProject accepted malformed coefficient")
	}
}

func TestGraphletTemporaryResourcesCloseAfterUpstreamFailure(t *testing.T) {
	graph, err := NewGraphFromEdges(3, []Edge{{0, 1}, {1, 2}, {2, 0}}, false)
	if err != nil {
		t.Fatal(err)
	}
	defer graph.Close()

	for operation := 0; operation < 3; operation++ {
		realClosed, listClosed := 0, 0
		adapters := defaultGraphletAdapters()
		adapters.closeReal = func(vector *realVector) { realClosed++; vector.close() }
		adapters.closeList = func(list *intVectorList) { listClosed++; list.close() }
		switch operation {
		case 0:
			adapters.graphletsCall = func(*Graph, *realVector, *intVectorList, *realVector, int) int { return 1 }
		case 1:
			adapters.candidateCall = func(*Graph, *realVector, *intVectorList, *realVector) int { return 1 }
		case 2:
			adapters.projectCall = func(*Graph, *realVector, *intVectorList, *realVector, bool, int) int { return 1 }
		}
		_ = callGraphletOperation(graph, operation, &adapters)
		if realClosed != 2 || listClosed != 1 {
			t.Errorf("operation %d closes = real %d, list %d", operation, realClosed, listClosed)
		}
	}
}

func callGraphletOperation(graph *Graph, operation int, adapters *graphletAdapters) error {
	switch operation {
	case 0:
		_, err := graph.graphlets(nil, 1, adapters)
		return err
	case 1:
		_, err := graph.graphletsCandidateBasis(nil, adapters)
		return err
	default:
		_, err := graph.graphletsProject([][]int{{0, 1, 2}}, []float64{1}, nil, 1, adapters)
		return err
	}
}
