package igraph_test

import (
	"errors"
	"fmt"
	"math"
	"reflect"
	"sync"
	"testing"

	"github.com/h8gi/go-igraph"
)

func TestCorrelatedGame(t *testing.T) {
	seed := uint64(339)
	source, err := igraph.NewGraphFromEdges(4, []igraph.Edge{{From: 0, To: 1}, {From: 1, To: 2}}, false)
	if err != nil {
		t.Fatal(err)
	}
	identity, err := source.CorrelatedGame(1, 0.4, igraph.CorrelatedGraphOptions{Seed: &seed})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(mustEdges(t, source), mustEdges(t, identity)) {
		t.Fatal("correlation one did not preserve edges")
	}
	source.Close()
	if v, _ := identity.VertexCount(); v != 4 {
		t.Fatalf("result unusable after source close: %d", v)
	}
	identity.Close()

	permutedSource, err := igraph.NewGraphFromEdges(3, []igraph.Edge{{From: 0, To: 1}}, false)
	if err != nil {
		t.Fatal(err)
	}
	defer permutedSource.Close()
	permuted, err := permutedSource.CorrelatedGame(1, 0.5, igraph.CorrelatedGraphOptions{Seed: &seed, Permutation: []int{2, 0, 1}})
	if err != nil {
		t.Fatal(err)
	}
	defer permuted.Close()
	if !reflect.DeepEqual(mustEdges(t, permuted), []igraph.Edge{{From: 1, To: 2}}) {
		t.Fatalf("permuted edges=%v", mustEdges(t, permuted))
	}
	random1, err := permutedSource.CorrelatedGame(0, 0.5, igraph.CorrelatedGraphOptions{Seed: &seed})
	if err != nil {
		t.Fatal(err)
	}
	defer random1.Close()
	random2, err := permutedSource.CorrelatedGame(0, 0.5, igraph.CorrelatedGraphOptions{Seed: &seed})
	if err != nil {
		t.Fatal(err)
	}
	defer random2.Close()
	if !reflect.DeepEqual(mustEdges(t, random1), mustEdges(t, random2)) {
		t.Fatal("same seed differed")
	}

	nonSimple, err := igraph.NewGraphFromEdges(2, []igraph.Edge{{From: 0, To: 1}, {From: 0, To: 1}}, false)
	if err != nil {
		t.Fatal(err)
	}
	defer nonSimple.Close()
	if _, err := nonSimple.CorrelatedGame(0.5, 0.5, igraph.CorrelatedGraphOptions{}); err == nil {
		t.Fatal("accepted nonsimple source")
	}
	closed, _ := igraph.NewGraph()
	closed.Close()
	if _, err := closed.CorrelatedGame(0.5, 0.5, igraph.CorrelatedGraphOptions{}); !errors.Is(err, igraph.ErrClosed) {
		t.Fatalf("closed error=%v", err)
	}
	var nilGraph *igraph.Graph
	if _, err := nilGraph.CorrelatedGame(0.5, 0.5, igraph.CorrelatedGraphOptions{}); !errors.Is(err, igraph.ErrClosed) {
		t.Fatalf("nil error=%v", err)
	}
	if _, err := permutedSource.CorrelatedGame(math.NaN(), 0.5, igraph.CorrelatedGraphOptions{}); err == nil {
		t.Fatal("accepted NaN correlation")
	}
	if _, err := permutedSource.CorrelatedGame(0.5, 0, igraph.CorrelatedGraphOptions{}); err == nil {
		t.Fatal("accepted zero probability")
	}
	for i, p := range [][]int{{0}, {0, 3, 1}, {0, 0, 2}} {
		if _, err := permutedSource.CorrelatedGame(0.5, 0.5, igraph.CorrelatedGraphOptions{Permutation: p}); err == nil {
			t.Errorf("invalid permutation %d accepted", i)
		}
	}
}

func TestCorrelatedPairGame(t *testing.T) {
	seed := uint64(339)
	options := igraph.CorrelatedGraphOptions{Seed: &seed, Permutation: []int{0, 1, 2, 3, 4, 5, 6, 7}}
	first, err := igraph.CorrelatedPairGame(8, 1, 0.4, true, options)
	if err != nil {
		t.Fatal(err)
	}
	defer first.First.Close()
	defer first.Second.Close()
	if !reflect.DeepEqual(mustEdges(t, first.First), mustEdges(t, first.Second)) {
		t.Fatal("correlation one pair differs")
	}
	if d, _ := first.First.IsDirected(); !d {
		t.Fatal("pair is not directed")
	}
	second, err := igraph.CorrelatedPairGame(8, 1, 0.4, true, options)
	if err != nil {
		t.Fatal(err)
	}
	defer second.First.Close()
	defer second.Second.Close()
	if !reflect.DeepEqual(mustEdges(t, first.First), mustEdges(t, second.First)) {
		t.Fatal("same seed differed")
	}
	first.First.Close()
	if _, err := first.Second.VertexCount(); err != nil {
		t.Fatalf("second depended on first: %v", err)
	}
	empty, err := igraph.CorrelatedPairGame(0, 0, 0.5, false, igraph.CorrelatedGraphOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer empty.First.Close()
	defer empty.Second.Close()
	if v, e := mustCounts(t, empty.First); v != 0 || e != 0 {
		t.Fatalf("empty=%d/%d", v, e)
	}
	for i, call := range []func() error{func() error {
		_, e := igraph.CorrelatedPairGame(-1, 0, 0.5, false, igraph.CorrelatedGraphOptions{})
		return e
	}, func() error {
		_, e := igraph.CorrelatedPairGame(2, -0.1, 0.5, false, igraph.CorrelatedGraphOptions{})
		return e
	}, func() error {
		_, e := igraph.CorrelatedPairGame(2, 0.5, 1, false, igraph.CorrelatedGraphOptions{})
		return e
	}, func() error {
		_, e := igraph.CorrelatedPairGame(2, 0.5, 0.5, false, igraph.CorrelatedGraphOptions{Permutation: []int{0, 0}})
		return e
	}, func() error {
		_, e := igraph.CorrelatedPairGame(2, math.Inf(1), 0.5, false, igraph.CorrelatedGraphOptions{})
		return e
	}, func() error {
		_, e := igraph.CorrelatedPairGame(2, 0.5, math.NaN(), false, igraph.CorrelatedGraphOptions{})
		return e
	}, func() error {
		_, e := igraph.CorrelatedPairGame(2, 0.5, 0.5, false, igraph.CorrelatedGraphOptions{Permutation: []int{0}})
		return e
	}, func() error {
		_, e := igraph.CorrelatedPairGame(2, 0.5, 0.5, false, igraph.CorrelatedGraphOptions{Permutation: []int{0, 2}})
		return e
	}} {
		if call() == nil {
			t.Errorf("validation %d accepted", i)
		}
	}
}

func TestCorrelatedGameConcurrentClose(t *testing.T) {
	seed := uint64(339)
	source, err := igraph.NewGraphFromEdges(4, []igraph.Edge{{From: 0, To: 1}}, false)
	if err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	wg.Add(1)
	started := make(chan struct{})
	done := make(chan error, 1)
	go func() {
		defer wg.Done()
		close(started)
		graph, err := source.CorrelatedGame(0.5, 0.3, igraph.CorrelatedGraphOptions{Seed: &seed})
		if graph != nil {
			graph.Close()
		}
		done <- err
	}()
	<-started
	source.Close()
	wg.Wait()
	if err := <-done; err != nil && !errors.Is(err, igraph.ErrClosed) {
		t.Fatalf("race error=%v", err)
	}
	if _, err := source.CorrelatedGame(0.5, 0.3, igraph.CorrelatedGraphOptions{}); !errors.Is(err, igraph.ErrClosed) {
		t.Fatalf("use after close=%v", err)
	}
}

func ExampleCorrelatedPairGame() {
	seed := uint64(3)
	pair, err := igraph.CorrelatedPairGame(5, 0.8, 0.3, false, igraph.CorrelatedGraphOptions{Seed: &seed})
	if err != nil {
		panic(err)
	}
	defer pair.First.Close()
	defer pair.Second.Close()
	vertices, _ := pair.First.VertexCount()
	fmt.Println(vertices)
	// Output:
	// 5
}
