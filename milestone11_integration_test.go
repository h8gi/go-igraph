package igraph_test

import (
	"errors"
	"reflect"
	"sync"
	"testing"

	igraph "github.com/h8gi/go-igraph"
)

func TestMilestone11CliqueAnalysisPipeline(t *testing.T) {
	graph, err := igraph.NewGraphFromEdges(5, []igraph.Edge{
		{From: 0, To: 1}, {From: 0, To: 2}, {From: 1, To: 2}, {From: 3, To: 4},
	}, false)
	if err != nil {
		t.Fatal(err)
	}

	cliqueNumber, err := graph.CliqueNumber()
	if err != nil || cliqueNumber != 3 {
		t.Fatalf("CliqueNumber = %d, %v", cliqueNumber, err)
	}
	independenceNumber, err := graph.IndependenceNumber()
	if err != nil || independenceNumber != 2 {
		t.Fatalf("IndependenceNumber = %d, %v", independenceNumber, err)
	}
	largest, err := graph.LargestCliques(1)
	if err != nil || largest.Truncated || !reflect.DeepEqual(largest.Sets, [][]int{{0, 1, 2}}) {
		t.Fatalf("LargestCliques = %#v, %v", largest, err)
	}
	maximal, err := graph.MaximalCliques(igraph.VertexSetEnumerationOptions{MaxResults: 2})
	if err != nil || maximal.Truncated || len(maximal.Sets) != 2 {
		t.Fatalf("MaximalCliques = %#v, %v", maximal, err)
	}
	fromVertices, err := graph.MaximalCliquesFromVertices([]int{0, 3}, igraph.VertexSetEnumerationOptions{MaxResults: 2})
	if err != nil || fromVertices.Truncated || !reflect.DeepEqual(fromVertices.Sets, [][]int{{3, 4}}) {
		t.Fatalf("MaximalCliquesFromVertices = %#v, %v", fromVertices, err)
	}
	histogram, err := graph.CliqueSizeHistogram(igraph.VertexSetRange{})
	if err != nil || !reflect.DeepEqual(histogram, []int{5, 4, 1}) {
		t.Fatalf("CliqueSizeHistogram = %v, %v", histogram, err)
	}
	maximalCount, err := graph.MaximalCliqueCount(igraph.VertexSetRange{})
	if err != nil || maximalCount != 2 {
		t.Fatalf("MaximalCliqueCount = %d, %v", maximalCount, err)
	}
	maximalHistogram, err := graph.MaximalCliqueSizeHistogram(igraph.VertexSetRange{})
	if err != nil || !reflect.DeepEqual(maximalHistogram, []int{0, 1, 1}) {
		t.Fatalf("MaximalCliqueSizeHistogram = %v, %v", maximalHistogram, err)
	}
	weights := []int{1, 1, 1, 10, 10}
	weightedNumber, err := graph.WeightedCliqueNumber(weights)
	if err != nil || weightedNumber != 20 {
		t.Fatalf("WeightedCliqueNumber = %d, %v", weightedNumber, err)
	}
	weighted, err := graph.MaximumWeightCliques(weights, 1)
	if err != nil || weighted.Truncated || !reflect.DeepEqual(weighted.Sets, [][]int{{3, 4}}) {
		t.Fatalf("MaximumWeightCliques = %#v, %v", weighted, err)
	}
	independent, err := graph.LargestIndependentVertexSets(6)
	if err != nil || independent.Truncated || len(independent.Sets) != 6 {
		t.Fatalf("LargestIndependentVertexSets = %#v, %v", independent, err)
	}

	if err := graph.Close(); err != nil {
		t.Fatal(err)
	}
	largest.Sets[0][0] = 99
	if weighted.Sets[0][0] == 99 || independent.Sets[0][0] == 99 {
		t.Error("returned nested slices share backing storage")
	}
	if !reflect.DeepEqual(histogram, []int{5, 4, 1}) || !reflect.DeepEqual(maximalHistogram, []int{0, 1, 1}) {
		t.Error("histograms changed after source closure")
	}
}

func TestMilestone11ConcurrentReadsAndClose(t *testing.T) {
	graph, err := igraph.NewGraphFromEdges(6, []igraph.Edge{
		{From: 0, To: 1}, {From: 0, To: 2}, {From: 1, To: 2},
		{From: 3, To: 4}, {From: 4, To: 5},
	}, false)
	if err != nil {
		t.Fatal(err)
	}
	weights := []int{1, 2, 3, 4, 5, 6}
	start := make(chan struct{})
	errorsByCall := make(chan error, 8)
	calls := []func() error{
		func() error { _, err := graph.CliqueNumber(); return err },
		func() error { _, err := graph.IndependenceNumber(); return err },
		func() error { _, err := graph.Cliques(igraph.VertexSetEnumerationOptions{MaxResults: 4}); return err },
		func() error {
			_, err := graph.MaximalCliques(igraph.VertexSetEnumerationOptions{MaxResults: 4})
			return err
		},
		func() error { _, err := graph.CliqueSizeHistogram(igraph.VertexSetRange{}); return err },
		func() error { _, err := graph.WeightedCliqueNumber(weights); return err },
		func() error { _, err := graph.MaximumWeightCliques(weights, 4); return err },
		func() error { _, err := graph.LargestIndependentVertexSets(4); return err },
	}
	var wait sync.WaitGroup
	for _, call := range calls {
		wait.Add(1)
		go func(call func() error) {
			defer wait.Done()
			<-start
			errorsByCall <- call()
		}(call)
	}
	close(start)
	if err := graph.Close(); err != nil {
		t.Fatal(err)
	}
	wait.Wait()
	close(errorsByCall)
	for err := range errorsByCall {
		if err != nil && !errors.Is(err, igraph.ErrClosed) {
			t.Errorf("concurrent call error = %v", err)
		}
	}
	if _, err := graph.CliqueNumber(); !errors.Is(err, igraph.ErrClosed) {
		t.Errorf("post-close CliqueNumber error = %v", err)
	}
}
