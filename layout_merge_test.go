package igraph_test

import (
	"math"
	"reflect"
	"sync"
	"testing"

	"github.com/h8gi/go-igraph"
)

func TestMergeLayoutsDLA(t *testing.T) {
	firstGraph, _ := igraph.NewRing(3, false, false)
	secondGraph, _ := igraph.NewGraphFromEdges(2, []igraph.Edge{{From: 0, To: 1}}, false)
	defer firstGraph.Close()
	defer secondGraph.Close()
	firstCoordinates, _ := igraph.NewMatrixFromRows([][]float64{{-1, 0}, {1, 0}, {0, 1}})
	secondCoordinates, _ := igraph.NewMatrixFromRows([][]float64{{0, -1}, {0, 1}})
	wantFirst := firstCoordinates.Rows()
	wantSecond := secondCoordinates.Rows()
	seed := uint64(99)
	options := igraph.DLAMergeOptions{Seed: &seed}
	merged, err := igraph.MergeLayoutsDLA([]*igraph.Graph{firstGraph, secondGraph}, []igraph.Matrix{firstCoordinates, secondCoordinates}, options)
	if err != nil {
		t.Fatalf("MergeLayoutsDLA failed: %v", err)
	}
	replayed, err := igraph.MergeLayoutsDLA([]*igraph.Graph{firstGraph, secondGraph}, []igraph.Matrix{firstCoordinates, secondCoordinates}, options)
	if err != nil {
		t.Fatalf("MergeLayoutsDLA replay failed: %v", err)
	}
	assertAdvancedLayout(t, "DLA merge", merged, 5)
	if !reflect.DeepEqual(merged.Rows(), replayed.Rows()) {
		t.Fatal("seeded DLA merges differ")
	}
	if !reflect.DeepEqual(firstCoordinates.Rows(), wantFirst) || !reflect.DeepEqual(secondCoordinates.Rows(), wantSecond) {
		t.Fatal("DLA merge mutated input coordinates")
	}
	firstGraph.Close()
	secondGraph.Close()
	assertAdvancedLayout(t, "closed-source DLA merge", merged, 5)
}

func TestMergeLayoutsDLARepeatedAndEmpty(t *testing.T) {
	graph, _ := igraph.NewGraphFromEdges(2, []igraph.Edge{{From: 0, To: 1}}, false)
	defer graph.Close()
	coordinates, _ := igraph.NewMatrixFromRows([][]float64{{-1, 0}, {1, 0}})
	seed := uint64(3)
	merged, err := igraph.MergeLayoutsDLA([]*igraph.Graph{graph, graph}, []igraph.Matrix{coordinates, coordinates}, igraph.DLAMergeOptions{Seed: &seed})
	if err != nil {
		t.Fatalf("repeated-graph DLA merge failed: %v", err)
	}
	assertAdvancedLayout(t, "repeated DLA", merged, 4)
	empty, err := igraph.MergeLayoutsDLA(nil, nil, igraph.DLAMergeOptions{})
	if err != nil {
		t.Fatalf("empty DLA merge failed: %v", err)
	}
	assertAdvancedLayout(t, "empty DLA", empty, 0)
}

func TestMergeLayoutsDLAInvalid(t *testing.T) {
	graph, _ := igraph.NewGraphFromEdges(2, []igraph.Edge{{From: 0, To: 1}}, false)
	defer graph.Close()
	valid, _ := igraph.NewMatrixFromRows([][]float64{{0, 0}, {1, 0}})
	badRows, _ := igraph.NewMatrix(1, 2)
	badColumns, _ := igraph.NewMatrix(2, 3)
	badFinite, _ := igraph.NewMatrixFromRows([][]float64{{0, 0}, {math.NaN(), 1}})
	invalid := []func() error{
		func() error {
			_, err := igraph.MergeLayoutsDLA([]*igraph.Graph{graph}, nil, igraph.DLAMergeOptions{})
			return err
		},
		func() error {
			_, err := igraph.MergeLayoutsDLA([]*igraph.Graph{nil}, []igraph.Matrix{valid}, igraph.DLAMergeOptions{})
			return err
		},
		func() error {
			_, err := igraph.MergeLayoutsDLA([]*igraph.Graph{graph}, []igraph.Matrix{badRows}, igraph.DLAMergeOptions{})
			return err
		},
		func() error {
			_, err := igraph.MergeLayoutsDLA([]*igraph.Graph{graph}, []igraph.Matrix{badColumns}, igraph.DLAMergeOptions{})
			return err
		},
		func() error {
			_, err := igraph.MergeLayoutsDLA([]*igraph.Graph{graph}, []igraph.Matrix{badFinite}, igraph.DLAMergeOptions{})
			return err
		},
	}
	for index, call := range invalid {
		if err := call(); err == nil {
			t.Errorf("invalid DLA case %d succeeded", index)
		}
	}
	graph.Close()
	if _, err := igraph.MergeLayoutsDLA([]*igraph.Graph{graph}, []igraph.Matrix{valid}, igraph.DLAMergeOptions{}); err != igraph.ErrClosed {
		t.Fatalf("closed DLA graph error = %v", err)
	}
}

func TestMergeLayoutsDLAReversedConcurrency(t *testing.T) {
	left, _ := igraph.NewGraphFromEdges(2, []igraph.Edge{{From: 0, To: 1}}, false)
	right, _ := igraph.NewGraphFromEdges(2, []igraph.Edge{{From: 0, To: 1}}, false)
	defer left.Close()
	defer right.Close()
	coordinates, _ := igraph.NewMatrixFromRows([][]float64{{-1, 0}, {1, 0}})
	seed := uint64(5)
	var wait sync.WaitGroup
	for _, graphs := range [][]*igraph.Graph{{left, right}, {right, left}} {
		graphs := graphs
		wait.Add(1)
		go func() {
			defer wait.Done()
			if _, err := igraph.MergeLayoutsDLA(graphs, []igraph.Matrix{coordinates, coordinates}, igraph.DLAMergeOptions{Seed: &seed}); err != nil {
				t.Errorf("concurrent DLA merge failed: %v", err)
			}
		}()
	}
	wait.Wait()
}
