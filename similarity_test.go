package igraph

import (
	"errors"
	"math"
	"reflect"
	"sync"
	"testing"
)

func TestNeighborhoodSimilarity(t *testing.T) {
	graph := newSimilarityTestGraph(t)
	defer graph.Close()

	rows, err := VertexIDs(4, 0, 4)
	if err != nil {
		t.Fatal(err)
	}
	columns, err := VertexIDs(0, 1, 2)
	if err != nil {
		t.Fatal(err)
	}

	jaccard, err := graph.NeighborhoodSimilarity(rows, columns, NeighborhoodSimilarityOptions{
		Metric: SimilarityJaccard, Direction: DirectionOut,
	})
	if err != nil {
		t.Fatalf("NeighborhoodSimilarity(Jaccard) error = %v", err)
	}
	assertMatrixRows(t, jaccard, [][]float64{
		{0.5, 0.5, 0},
		{1, 1, 0},
		{0.5, 0.5, 0},
	})

	dice, err := graph.NeighborhoodSimilarity(rows, columns, NeighborhoodSimilarityOptions{
		Metric: SimilarityDice, Direction: DirectionOut,
	})
	if err != nil {
		t.Fatalf("NeighborhoodSimilarity(Dice) error = %v", err)
	}
	assertMatrixRows(t, dice, [][]float64{
		{2.0 / 3.0, 2.0 / 3.0, 0},
		{1, 1, 0},
		{2.0 / 3.0, 2.0 / 3.0, 0},
	})

	rowTwo, _ := VertexIDs(2)
	columnZero, _ := VertexIDs(0)
	withoutLoops, err := graph.NeighborhoodSimilarity(rowTwo, columnZero, NeighborhoodSimilarityOptions{
		Metric: SimilarityJaccard, Direction: DirectionOut,
	})
	if err != nil {
		t.Fatal(err)
	}
	withLoops, err := graph.NeighborhoodSimilarity(rowTwo, columnZero, NeighborhoodSimilarityOptions{
		Metric: SimilarityJaccard, Direction: DirectionOut, IncludeLoops: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	assertMatrixRows(t, withoutLoops, [][]float64{{0}})
	assertMatrixRows(t, withLoops, [][]float64{{1.0 / 3.0}})

	empty, err := graph.NeighborhoodSimilarity(NoVertices(), NoVertices(), NeighborhoodSimilarityOptions{})
	if err != nil {
		t.Fatalf("NeighborhoodSimilarity(empty) error = %v", err)
	}
	if rows, columns := empty.Dims(); rows != 0 || columns != 0 {
		t.Fatalf("NeighborhoodSimilarity(empty) dimensions = %d by %d", rows, columns)
	}
}

func TestSelectedToAllSimilarities(t *testing.T) {
	graph := newSimilarityTestGraph(t)
	defer graph.Close()

	selected, err := VertexIDs(0, 4, 0)
	if err != nil {
		t.Fatal(err)
	}
	inverse, err := graph.InverseLogWeightedSimilarity(selected, DirectionOut)
	if err != nil {
		t.Fatalf("InverseLogWeightedSimilarity() error = %v", err)
	}
	wCommonTwo := 1 / math.Log(3)
	wCommonThree := 1 / math.Log(2)
	assertMatrixRowsApprox(t, inverse, [][]float64{
		{0, wCommonTwo + wCommonThree, 0, 0, wCommonTwo},
		{wCommonTwo, wCommonTwo, 0, 0, 0},
		{0, wCommonTwo + wCommonThree, 0, 0, wCommonTwo},
	}, 1e-12)

	cocitationRows, _ := VertexIDs(2, 3, 2)
	cocitation, err := graph.CitationCoupling(cocitationRows, CouplingCocitation)
	if err != nil {
		t.Fatalf("CitationCoupling(cocitation) error = %v", err)
	}
	assertMatrixRows(t, cocitation, [][]float64{
		{0, 0, 0, 2, 0},
		{0, 0, 2, 0, 0},
		{0, 0, 0, 2, 0},
	})

	bibliographicRows, _ := VertexIDs(0, 4)
	bibliographic, err := graph.CitationCoupling(bibliographicRows, CouplingBibliographic)
	if err != nil {
		t.Fatalf("CitationCoupling(bibliographic) error = %v", err)
	}
	assertMatrixRows(t, bibliographic, [][]float64{
		{0, 2, 0, 0, 1},
		{1, 1, 0, 0, 0},
	})

	empty, err := graph.CitationCoupling(NoVertices(), CouplingCocitation)
	if err != nil {
		t.Fatalf("CitationCoupling(empty) error = %v", err)
	}
	if rows, columns := empty.Dims(); rows != 0 || columns != 5 {
		t.Fatalf("CitationCoupling(empty) dimensions = %d by %d, want 0 by 5", rows, columns)
	}
}

func TestSimilarityUndirectedModesAndCoupling(t *testing.T) {
	graph, err := NewGraphFromEdges(3, []Edge{
		{From: 0, To: 2},
		{From: 1, To: 2},
	}, false)
	if err != nil {
		t.Fatal(err)
	}
	defer graph.Close()

	var reference Matrix
	for index, direction := range []DirectionMode{DirectionOut, DirectionIn, DirectionAll} {
		result, err := graph.NeighborhoodSimilarity(
			AllVertices(), AllVertices(),
			NeighborhoodSimilarityOptions{Metric: SimilarityJaccard, Direction: direction},
		)
		if err != nil {
			t.Fatalf("NeighborhoodSimilarity(%v) error = %v", direction, err)
		}
		if index == 0 {
			reference = result
		} else if !reflect.DeepEqual(result.Rows(), reference.Rows()) {
			t.Errorf("undirected direction %v changed similarity", direction)
		}
	}

	cocitation, err := graph.CitationCoupling(AllVertices(), CouplingCocitation)
	if err != nil {
		t.Fatal(err)
	}
	bibliographic, err := graph.CitationCoupling(AllVertices(), CouplingBibliographic)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(cocitation.Rows(), bibliographic.Rows()) {
		t.Errorf("undirected coupling kinds differ: %v vs %v", cocitation.Rows(), bibliographic.Rows())
	}
}

func TestSimilarityValidationAndOwnership(t *testing.T) {
	graph := newSimilarityTestGraph(t)
	invalidVertex, _ := VertexIDs(5)
	invalidKind := VertexSelector{kind: vertexSelectorKind(99)}

	for _, selector := range []VertexSelector{invalidVertex, invalidKind} {
		if _, err := graph.NeighborhoodSimilarity(selector, AllVertices(), NeighborhoodSimilarityOptions{}); err == nil {
			t.Errorf("NeighborhoodSimilarity(%v) error = nil", selector)
		}
		if _, err := graph.InverseLogWeightedSimilarity(selector, DirectionOut); err == nil {
			t.Errorf("InverseLogWeightedSimilarity(%v) error = nil", selector)
		}
		if _, err := graph.CitationCoupling(selector, CouplingCocitation); err == nil {
			t.Errorf("CitationCoupling(%v) error = nil", selector)
		}
	}
	if _, err := graph.NeighborhoodSimilarity(AllVertices(), invalidVertex, NeighborhoodSimilarityOptions{}); err == nil {
		t.Error("NeighborhoodSimilarity(invalid column selector) error = nil")
	}
	if _, err := graph.NeighborhoodSimilarity(AllVertices(), AllVertices(), NeighborhoodSimilarityOptions{Metric: NeighborhoodSimilarityMetric(99)}); err == nil {
		t.Error("NeighborhoodSimilarity(invalid metric) error = nil")
	}
	if _, err := graph.NeighborhoodSimilarity(AllVertices(), AllVertices(), NeighborhoodSimilarityOptions{Direction: DirectionMode(99)}); err == nil {
		t.Error("NeighborhoodSimilarity(invalid direction) error = nil")
	}
	if _, err := graph.InverseLogWeightedSimilarity(AllVertices(), DirectionMode(99)); err == nil {
		t.Error("InverseLogWeightedSimilarity(invalid direction) error = nil")
	}
	if _, err := graph.CitationCoupling(AllVertices(), CitationCouplingKind(99)); err == nil {
		t.Error("CitationCoupling(invalid kind) error = nil")
	}
	if _, err := graph.selectedToAllSimilarity(AllVertices(), "test", nil, nil, similarityHooks{}); err == nil {
		t.Error("selectedToAllSimilarity(nil operation) error = nil")
	}
	selected, _ := VertexIDs(0)
	result, err := graph.CitationCoupling(selected, CouplingBibliographic)
	if err != nil {
		t.Fatal(err)
	}
	before := result.Rows()
	copyRows := result.Rows()
	copyRows[0][1] = -1
	if !reflect.DeepEqual(result.Rows(), before) {
		t.Fatal("mutating Rows result changed the similarity matrix")
	}
	if err := graph.Close(); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(result.Rows(), before) {
		t.Fatal("closing graph changed the similarity matrix")
	}

	for _, call := range []func() error{
		func() error {
			_, err := graph.NeighborhoodSimilarity(AllVertices(), AllVertices(), NeighborhoodSimilarityOptions{})
			return err
		},
		func() error { _, err := graph.InverseLogWeightedSimilarity(AllVertices(), DirectionOut); return err },
		func() error { _, err := graph.CitationCoupling(AllVertices(), CouplingCocitation); return err },
	} {
		if err := call(); !errors.Is(err, ErrClosed) {
			t.Errorf("closed graph error = %v, want %v", err, ErrClosed)
		}
	}
	var nilGraph *Graph
	if _, err := nilGraph.NeighborhoodSimilarity(AllVertices(), AllVertices(), NeighborhoodSimilarityOptions{}); !errors.Is(err, ErrClosed) {
		t.Errorf("nil graph error = %v, want %v", err, ErrClosed)
	}
	if _, err := nilGraph.InverseLogWeightedSimilarity(AllVertices(), DirectionOut); !errors.Is(err, ErrClosed) {
		t.Errorf("nil inverse graph error = %v, want %v", err, ErrClosed)
	}
	if _, err := nilGraph.CitationCoupling(AllVertices(), CouplingCocitation); !errors.Is(err, ErrClosed) {
		t.Errorf("nil coupling graph error = %v, want %v", err, ErrClosed)
	}
}

func TestSimilarityEmptyGraphAndInternalDimensions(t *testing.T) {
	graph, err := NewGraph()
	if err != nil {
		t.Fatal(err)
	}
	defer graph.Close()

	for name, call := range map[string]func() (Matrix, error){
		"neighborhood": func() (Matrix, error) {
			return graph.NeighborhoodSimilarity(AllVertices(), AllVertices(), NeighborhoodSimilarityOptions{})
		},
		"inverse": func() (Matrix, error) {
			return graph.InverseLogWeightedSimilarity(AllVertices(), DirectionAll)
		},
		"coupling": func() (Matrix, error) {
			return graph.CitationCoupling(AllVertices(), CouplingBibliographic)
		},
	} {
		result, err := call()
		if err != nil {
			t.Errorf("%s empty graph error = %v", name, err)
			continue
		}
		if rows, columns := result.Dims(); rows != 0 || columns != 0 {
			t.Errorf("%s empty graph dimensions = %d by %d", name, rows, columns)
		}
	}

	matrix, err := NewMatrix(1, 1)
	if err != nil {
		t.Fatal(err)
	}
	cMatrix, err := newCMatrix(matrix)
	if err != nil {
		t.Fatal(err)
	}
	defer cMatrix.close()
	if _, err := checkedSimilarityMatrix(cMatrix, 2, 1); err == nil {
		t.Error("checkedSimilarityMatrix(mismatched dimensions) error = nil")
	}
}

func TestSimilarityInitializationAndOperationFailures(t *testing.T) {
	graph := newSimilarityTestGraph(t)
	defer graph.Close()

	initializationError := errors.New("initialize similarity result")
	initializationHooks := similarityHooks{
		newResult: func() (*cMatrix, error) { return nil, initializationError },
	}
	if _, err := graph.neighborhoodSimilarity(
		AllVertices(), AllVertices(), NeighborhoodSimilarityOptions{}, initializationHooks,
	); !errors.Is(err, initializationError) {
		t.Errorf("neighborhood initialization error = %v, want %v", err, initializationError)
	}
	if _, err := graph.inverseLogWeightedSimilarity(
		AllVertices(), DirectionOut, initializationHooks,
	); !errors.Is(err, initializationError) {
		t.Errorf("selected-to-all initialization error = %v, want %v", err, initializationError)
	}

	operationError := errors.New("similarity operation failed")
	operationHooks := similarityHooks{run: func() error { return operationError }}
	if _, err := graph.neighborhoodSimilarity(
		AllVertices(), AllVertices(), NeighborhoodSimilarityOptions{}, operationHooks,
	); !errors.Is(err, operationError) {
		t.Errorf("neighborhood operation error = %v, want %v", err, operationError)
	}
	if _, err := graph.inverseLogWeightedSimilarity(
		AllVertices(), DirectionOut, operationHooks,
	); !errors.Is(err, operationError) {
		t.Errorf("selected-to-all operation error = %v, want %v", err, operationError)
	}
}

func TestSimilarityConcurrentReads(t *testing.T) {
	graph := newSimilarityTestGraph(t)
	defer graph.Close()

	var wait sync.WaitGroup
	errorsByWorker := make(chan error, 36)
	for worker := 0; worker < 4; worker++ {
		wait.Add(1)
		go func() {
			defer wait.Done()
			for iteration := 0; iteration < 3; iteration++ {
				if _, err := graph.NeighborhoodSimilarity(AllVertices(), AllVertices(), NeighborhoodSimilarityOptions{}); err != nil {
					errorsByWorker <- err
				}
				if _, err := graph.InverseLogWeightedSimilarity(AllVertices(), DirectionOut); err != nil {
					errorsByWorker <- err
				}
				if _, err := graph.CitationCoupling(AllVertices(), CouplingCocitation); err != nil {
					errorsByWorker <- err
				}
			}
		}()
	}
	wait.Wait()
	close(errorsByWorker)
	for err := range errorsByWorker {
		t.Errorf("concurrent similarity error = %v", err)
	}
}

func newSimilarityTestGraph(t *testing.T) *Graph {
	t.Helper()
	graph, err := NewGraphFromEdges(5, []Edge{
		{From: 0, To: 2},
		{From: 0, To: 3},
		{From: 1, To: 2},
		{From: 1, To: 3},
		{From: 4, To: 2},
	}, true)
	if err != nil {
		t.Fatal(err)
	}
	return graph
}

func assertMatrixRowsApprox(t *testing.T, matrix Matrix, want [][]float64, tolerance float64) {
	t.Helper()
	got := matrix.Rows()
	if len(got) != len(want) {
		t.Fatalf("matrix row count = %d, want %d", len(got), len(want))
	}
	for row := range want {
		if len(got[row]) != len(want[row]) {
			t.Fatalf("matrix row %d length = %d, want %d", row, len(got[row]), len(want[row]))
		}
		for column := range want[row] {
			if math.Abs(got[row][column]-want[row][column]) > tolerance {
				t.Errorf("matrix[%d][%d] = %v, want %v", row, column, got[row][column], want[row][column])
			}
		}
	}
}
