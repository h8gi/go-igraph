package igraph

import (
	"errors"
	"math"
	"reflect"
	"sync"
	"testing"
	"time"
)

func TestCrossGraphLocalScanSnapshots(t *testing.T) {
	neighborhoods, err := NewGraphFromEdges(3, []Edge{{0, 1}, {1, 2}}, false)
	if err != nil {
		t.Fatal(err)
	}
	defer neighborhoods.Close()
	comparison, err := NewGraphFromEdges(3, []Edge{{0, 1}, {0, 2}}, false)
	if err != nil {
		t.Fatal(err)
	}
	defer comparison.Close()
	tests := []struct {
		radius  int
		weights []float64
		want    []float64
	}{
		{0, nil, []float64{1, 1, 0}},
		{0, []float64{2, 5}, []float64{2, 2, 0}},
		{1, nil, []float64{1, 2, 0}},
		{1, []float64{2, 5}, []float64{2, 7, 0}},
		{2, nil, []float64{2, 2, 2}},
	}
	for _, test := range tests {
		got, err := neighborhoods.CrossGraphLocalScan(comparison, LocalScanOptions{Radius: test.radius, Direction: DirectionAll, Weights: test.weights})
		if err != nil {
			t.Fatalf("radius %d: %v", test.radius, err)
		}
		if !reflect.DeepEqual(got, test.want) {
			t.Fatalf("radius %d = %v, want %v", test.radius, got, test.want)
		}
	}
}

func TestCrossGraphLocalScanSameGraph(t *testing.T) {
	g, err := NewGraphFromEdges(3, []Edge{{0, 1}, {1, 2}}, false)
	if err != nil {
		t.Fatal(err)
	}
	defer g.Close()
	options := LocalScanOptions{Radius: 2, Direction: DirectionAll, Weights: []float64{2, 5}}
	want, err := g.LocalScan(options)
	if err != nil {
		t.Fatal(err)
	}
	got, err := g.CrossGraphLocalScan(g, options)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("same graph = %v, want %v", got, want)
	}
}

func TestCrossGraphLocalScanValidationAndFailures(t *testing.T) {
	first, err := NewGraphFromEdges(2, []Edge{{0, 1}}, false)
	if err != nil {
		t.Fatal(err)
	}
	defer first.Close()
	second, err := NewGraphFromEdges(3, []Edge{{0, 1}}, false)
	if err != nil {
		t.Fatal(err)
	}
	defer second.Close()
	if _, err := first.CrossGraphLocalScan(second, LocalScanOptions{}); err == nil {
		t.Fatal("expected vertex count error")
	}
	directed, err := NewGraphFromEdges(2, []Edge{{0, 1}}, true)
	if err != nil {
		t.Fatal(err)
	}
	defer directed.Close()
	if _, err := first.CrossGraphLocalScan(directed, LocalScanOptions{}); err == nil {
		t.Fatal("expected directedness error")
	}
	if _, err := first.CrossGraphLocalScan(first, LocalScanOptions{Radius: -1}); err == nil {
		t.Fatal("expected radius error")
	}
	if _, err := first.CrossGraphLocalScan(first, LocalScanOptions{Direction: DirectionMode(99)}); err == nil {
		t.Fatal("expected direction error")
	}
	if _, err := first.CrossGraphLocalScan(first, LocalScanOptions{Weights: []float64{math.NaN()}}); err == nil {
		t.Fatal("expected weight error")
	}
	failure := errors.New("failure")
	if _, err := first.crossGraphLocalScan(first, LocalScanOptions{}, localScanHooks{newResult: func([]float64) (*realVector, error) { return nil, failure }}); !errors.Is(err, failure) {
		t.Fatalf("initialization error = %v", err)
	}
	if _, err := first.crossGraphLocalScan(first, LocalScanOptions{}, localScanHooks{run: func() error { return failure }}); !errors.Is(err, failure) {
		t.Fatalf("operation error = %v", err)
	}
	var nilGraph *Graph
	if _, err := nilGraph.CrossGraphLocalScan(first, LocalScanOptions{}); !errors.Is(err, ErrClosed) {
		t.Fatalf("nil receiver error = %v", err)
	}
	if _, err := first.CrossGraphLocalScan(nil, LocalScanOptions{}); !errors.Is(err, ErrClosed) {
		t.Fatalf("nil comparison error = %v", err)
	}
}

func TestCrossGraphLocalScanReversedCallsDoNotDeadlock(t *testing.T) {
	first, err := NewGraphFromEdges(3, []Edge{{0, 1}}, false)
	if err != nil {
		t.Fatal(err)
	}
	defer first.Close()
	second, err := NewGraphFromEdges(3, []Edge{{1, 2}}, false)
	if err != nil {
		t.Fatal(err)
	}
	defer second.Close()
	var wait sync.WaitGroup
	errorsChannel := make(chan error, 2)
	for _, pair := range [][2]*Graph{{first, second}, {second, first}} {
		wait.Add(1)
		go func(neighborhoods, comparison *Graph) {
			defer wait.Done()
			for index := 0; index < 20; index++ {
				if _, err := neighborhoods.CrossGraphLocalScan(comparison, LocalScanOptions{Radius: 1, Direction: DirectionAll}); err != nil {
					errorsChannel <- err
					return
				}
			}
		}(pair[0], pair[1])
	}
	done := make(chan struct{})
	go func() { wait.Wait(); close(done) }()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("reversed scans deadlocked")
	}
	close(errorsChannel)
	for err := range errorsChannel {
		t.Fatal(err)
	}
}

func TestCrossGraphLocalScanCloseWaitsForBorrow(t *testing.T) {
	first, err := NewGraphFromEdges(1, nil, false)
	if err != nil {
		t.Fatal(err)
	}
	defer first.Close()
	second, err := NewGraphFromEdges(1, nil, false)
	if err != nil {
		t.Fatal(err)
	}
	entered, release := make(chan struct{}), make(chan struct{})
	scanDone := make(chan error, 1)
	go func() {
		_, err := first.crossGraphLocalScan(second, LocalScanOptions{}, localScanHooks{run: func() error { close(entered); <-release; return nil }})
		scanDone <- err
	}()
	<-entered
	closeDone := make(chan error, 1)
	go func() { closeDone <- second.Close() }()
	select {
	case err := <-closeDone:
		t.Fatalf("Close returned during borrow: %v", err)
	case <-time.After(20 * time.Millisecond):
	}
	close(release)
	if err := <-scanDone; err != nil {
		t.Fatal(err)
	}
	if err := <-closeDone; err != nil {
		t.Fatal(err)
	}
	if _, err := first.CrossGraphLocalScan(second, LocalScanOptions{}); !errors.Is(err, ErrClosed) {
		t.Fatalf("closed comparison error = %v", err)
	}
}
