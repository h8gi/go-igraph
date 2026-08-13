package igraph_test

import (
	"errors"
	"math"
	"reflect"
	"sync"
	"testing"

	igraph "github.com/h8gi/go-igraph"
)

func TestBiadjacencyRoundTripMultiplicityAndIDs(t *testing.T) {
	matrix, err := igraph.NewMatrixFromRows([][]float64{{0, 2, 1}, {1, 0, 0}})
	if err != nil {
		t.Fatal(err)
	}
	constructed, err := igraph.NewBiadjacency(matrix, igraph.BiadjacencyOptions{Multiple: true})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = constructed.Graph.Close() })
	if !reflect.DeepEqual(constructed.Partition, igraph.BipartitePartition{false, false, true, true, true}) {
		t.Fatalf("partition = %v", constructed.Partition)
	}
	edges, err := constructed.Graph.EdgeCount()
	if err != nil || edges != 4 {
		t.Fatalf("edges = %d, %v", edges, err)
	}
	roundTrip, err := constructed.Graph.Biadjacency(constructed.Partition, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(roundTrip.Matrix.Rows(), matrix.Rows()) {
		t.Fatalf("matrix = %v", roundTrip.Matrix.Rows())
	}
	if !reflect.DeepEqual(roundTrip.RowVertexIDs, []int{0, 1}) || !reflect.DeepEqual(roundTrip.ColumnVertexIDs, []int{2, 3, 4}) {
		t.Fatalf("row IDs = %v, column IDs = %v", roundTrip.RowVertexIDs, roundTrip.ColumnVertexIDs)
	}
	if err := constructed.Graph.Close(); err != nil {
		t.Fatal(err)
	}
	if got := roundTrip.Matrix.Rows(); !reflect.DeepEqual(got, matrix.Rows()) {
		t.Fatalf("matrix after close = %v", got)
	}
}

func TestBiadjacencySingleEdgeAndDirectedModes(t *testing.T) {
	matrix, _ := igraph.NewMatrixFromRows([][]float64{{0.5, 7}})
	for _, test := range []struct {
		mode igraph.DirectionMode
		want []igraph.Edge
	}{
		{igraph.DirectionOut, []igraph.Edge{{From: 0, To: 1}, {From: 0, To: 2}}},
		{igraph.DirectionIn, []igraph.Edge{{From: 1, To: 0}, {From: 2, To: 0}}},
		{igraph.DirectionAll, []igraph.Edge{{From: 0, To: 1}, {From: 1, To: 0}, {From: 0, To: 2}, {From: 2, To: 0}}},
	} {
		result, err := igraph.NewBiadjacency(matrix, igraph.BiadjacencyOptions{Directed: true, Direction: test.mode})
		if err != nil {
			t.Fatal(err)
		}
		edges, err := result.Graph.Edges()
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(edges, test.want) {
			t.Errorf("mode %d edges = %v, want %v", test.mode, edges, test.want)
		}
		_ = result.Graph.Close()
	}
}

func TestWeightedBiadjacencyRoundTripAndAlignment(t *testing.T) {
	matrix, _ := igraph.NewMatrixFromRows([][]float64{{0, -2, 3.5}, {4, 0, 0}})
	result, err := igraph.NewWeightedBiadjacency(matrix, igraph.WeightedBiadjacencyOptions{})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = result.Graph.Close() })
	if !reflect.DeepEqual(result.Weights, []float64{4, -2, 3.5}) {
		t.Fatalf("weights = %v", result.Weights)
	}
	roundTrip, err := result.Graph.Biadjacency(result.Partition, result.Weights)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(roundTrip.Matrix.Rows(), matrix.Rows()) {
		t.Fatalf("weighted matrix = %v", roundTrip.Matrix.Rows())
	}

	parallel, err := igraph.NewBipartite(igraph.BipartitePartition{false, true}, []igraph.Edge{{0, 1}, {0, 1}}, false)
	if err != nil {
		t.Fatal(err)
	}
	defer parallel.Graph.Close()
	summed, err := parallel.Graph.Biadjacency(parallel.Partition, []float64{2, -0.5})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(summed.Matrix.Rows(), [][]float64{{1.5}}) {
		t.Fatalf("summed = %v", summed.Matrix.Rows())
	}
}

func TestBiadjacencyEmptyRectangularMatrices(t *testing.T) {
	for _, dims := range [][2]int{{0, 0}, {0, 3}, {2, 0}} {
		matrix, err := igraph.NewMatrix(dims[0], dims[1])
		if err != nil {
			t.Fatal(err)
		}
		result, err := igraph.NewBiadjacency(matrix, igraph.BiadjacencyOptions{})
		if err != nil {
			t.Fatalf("dims %v: %v", dims, err)
		}
		rows, columns := result.Partition[:dims[0]], result.Partition[dims[0]:]
		if len(rows) != dims[0] || len(columns) != dims[1] {
			t.Fatalf("dims %v partition = %v", dims, result.Partition)
		}
		roundTrip, err := result.Graph.Biadjacency(result.Partition, nil)
		if err != nil {
			t.Fatal(err)
		}
		gotRows, gotColumns := roundTrip.Matrix.Dims()
		if gotRows != dims[0] || gotColumns != dims[1] {
			t.Errorf("round trip dims = %dx%d, want %dx%d", gotRows, gotColumns, dims[0], dims[1])
		}
		if roundTrip.RowVertexIDs == nil || roundTrip.ColumnVertexIDs == nil {
			t.Fatal("nil ID result")
		}
		_ = result.Graph.Close()
	}
}

func TestBiadjacencyValidationAndClosedGraph(t *testing.T) {
	invalid := []struct {
		values   [][]float64
		weighted bool
	}{
		{[][]float64{{-1}}, false},
		{[][]float64{{math.NaN()}}, false},
		{[][]float64{{math.Inf(1)}}, true},
	}
	for _, test := range invalid {
		matrix, _ := igraph.NewMatrixFromRows(test.values)
		var err error
		if test.weighted {
			_, err = igraph.NewWeightedBiadjacency(matrix, igraph.WeightedBiadjacencyOptions{})
		} else {
			_, err = igraph.NewBiadjacency(matrix, igraph.BiadjacencyOptions{})
		}
		if err == nil {
			t.Fatalf("matrix %v weighted %v: error nil", test.values, test.weighted)
		}
	}
	fractional, _ := igraph.NewMatrixFromRows([][]float64{{1.5}})
	if _, err := igraph.NewBiadjacency(fractional, igraph.BiadjacencyOptions{Multiple: true}); err == nil {
		t.Fatal("fractional multiplicity error nil")
	}
	one, _ := igraph.NewMatrixFromRows([][]float64{{1}})
	if _, err := igraph.NewBiadjacency(one, igraph.BiadjacencyOptions{Direction: igraph.DirectionMode(99)}); err == nil {
		t.Fatal("invalid unweighted direction error nil")
	}
	if _, err := igraph.NewWeightedBiadjacency(one, igraph.WeightedBiadjacencyOptions{Direction: igraph.DirectionMode(99)}); err == nil {
		t.Fatal("invalid weighted direction error nil")
	}

	graph, _ := igraph.NewGraphFromEdges(2, []igraph.Edge{{0, 1}}, false)
	if _, err := graph.Biadjacency(igraph.BipartitePartition{false, true}, []float64{}); err == nil {
		t.Fatal("empty weights error nil")
	}
	if _, err := graph.Biadjacency(igraph.BipartitePartition{false}, nil); err == nil {
		t.Fatal("short partition error nil")
	}
	if _, err := graph.Biadjacency(igraph.BipartitePartition{false, false}, nil); err == nil {
		t.Fatal("same-mode edge error nil")
	}
	if _, err := graph.Biadjacency(igraph.BipartitePartition{false, true}, []float64{math.NaN()}); err == nil {
		t.Fatal("NaN weight error nil")
	}
	_ = graph.Close()
	if _, err := graph.Biadjacency(igraph.BipartitePartition{false, true}, nil); !errors.Is(err, igraph.ErrClosed) {
		t.Fatalf("closed error = %v", err)
	}
}

func TestBiadjacencyConcurrentReads(t *testing.T) {
	graph, err := igraph.NewBipartite(
		igraph.BipartitePartition{false, false, true},
		[]igraph.Edge{{From: 0, To: 2}, {From: 1, To: 2}}, false,
	)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = graph.Graph.Close() })
	var group sync.WaitGroup
	for range 8 {
		group.Add(1)
		go func() {
			defer group.Done()
			result, err := graph.Graph.Biadjacency(graph.Partition, nil)
			if err != nil || !reflect.DeepEqual(result.Matrix.Rows(), [][]float64{{1}, {1}}) {
				t.Errorf("Biadjacency() = %v, %v", result.Matrix.Rows(), err)
			}
		}()
	}
	group.Wait()
}
