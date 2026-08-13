package igraph_test

import (
	"reflect"
	"sync"
	"testing"

	igraph "github.com/h8gi/go-igraph"
)

func TestNewBipartiteGNPBoundariesAndDirections(t *testing.T) {
	tests := []struct {
		name      string
		p         float64
		options   igraph.BipartiteRandomOptions
		wantEdges int
	}{
		{name: "empty undirected", p: 0, wantEdges: 0},
		{name: "complete undirected", p: 1, wantEdges: 6},
		{name: "complete directed out", p: 1, options: igraph.BipartiteRandomOptions{Directed: true, Direction: igraph.DirectionOut}, wantEdges: 6},
		{name: "complete directed in", p: 1, options: igraph.BipartiteRandomOptions{Directed: true, Direction: igraph.DirectionIn}, wantEdges: 6},
		{name: "complete directed all", p: 1, options: igraph.BipartiteRandomOptions{Directed: true, Direction: igraph.DirectionAll}, wantEdges: 12},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := igraph.NewBipartiteGNP(2, 3, test.p, test.options)
			if err != nil {
				t.Fatal(err)
			}
			defer result.Graph.Close()
			assertRandomBipartiteResult(t, result, 2, 3, test.wantEdges, test.options.Directed)
			assertRandomBipartiteOrientation(t, result, test.options)
		})
	}

	for _, sizes := range [][2]int{{0, 0}, {0, 3}, {2, 0}} {
		result, err := igraph.NewBipartiteGNP(sizes[0], sizes[1], 1, igraph.BipartiteRandomOptions{})
		if err != nil {
			t.Fatalf("zero-sized modes %v: %v", sizes, err)
		}
		assertRandomBipartiteResult(t, result, sizes[0], sizes[1], 0, false)
		result.Graph.Close()
	}
}

func TestNewBipartiteGNMBoundaries(t *testing.T) {
	for _, test := range []struct {
		name      string
		n1, n2, m int
		options   igraph.BipartiteRandomOptions
	}{
		{name: "zero", n1: 2, n2: 3, m: 0},
		{name: "undirected maximum", n1: 2, n2: 3, m: 6},
		{name: "mutual directed maximum", n1: 2, n2: 3, m: 12, options: igraph.BipartiteRandomOptions{Directed: true, Direction: igraph.DirectionAll}},
		{name: "empty modes", n1: 0, n2: 0, m: 0},
	} {
		t.Run(test.name, func(t *testing.T) {
			result, err := igraph.NewBipartiteGNM(test.n1, test.n2, test.m, test.options)
			if err != nil {
				t.Fatal(err)
			}
			defer result.Graph.Close()
			assertRandomBipartiteResult(t, result, test.n1, test.n2, test.m, test.options.Directed)
		})
	}
	if _, err := igraph.NewBipartiteGNM(2, 3, 7, igraph.BipartiteRandomOptions{}); err == nil {
		t.Error("expected simple edge capacity error")
	}
	if _, err := igraph.NewBipartiteGNM(0, 3, 1, igraph.BipartiteRandomOptions{}); err == nil {
		t.Error("expected edge capacity error for empty mode")
	}
}

func TestNewBipartiteIEAParallelEdges(t *testing.T) {
	seed := uint64(17)
	result, err := igraph.NewBipartiteIEA(2, 2, 5, igraph.BipartiteRandomOptions{Seed: &seed})
	if err != nil {
		t.Fatal(err)
	}
	defer result.Graph.Close()
	assertRandomBipartiteResult(t, result, 2, 2, 5, false)
	edges, err := result.Graph.Edges()
	if err != nil {
		t.Fatal(err)
	}
	seen := make(map[igraph.Edge]bool)
	duplicate := false
	for _, edge := range edges {
		if seen[edge] {
			duplicate = true
		}
		seen[edge] = true
	}
	if !duplicate {
		t.Error("five IEA edges over four endpoint pairs must include a parallel edge")
	}
	if _, err := igraph.NewBipartiteIEA(0, 2, 1, igraph.BipartiteRandomOptions{}); err == nil {
		t.Error("expected non-empty mode error")
	}
	empty, err := igraph.NewBipartiteIEA(0, 2, 0, igraph.BipartiteRandomOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer empty.Graph.Close()
	assertRandomBipartiteResult(t, empty, 0, 2, 0, false)
}

func TestRandomBipartiteDeterministicReplay(t *testing.T) {
	tests := []struct {
		name string
		call func(*uint64) (igraph.BipartiteGraphResult, error)
	}{
		{name: "gnp", call: func(seed *uint64) (igraph.BipartiteGraphResult, error) {
			return igraph.NewBipartiteGNP(20, 15, .2, igraph.BipartiteRandomOptions{Directed: true, Direction: igraph.DirectionAll, Seed: seed})
		}},
		{name: "gnm", call: func(seed *uint64) (igraph.BipartiteGraphResult, error) {
			return igraph.NewBipartiteGNM(20, 15, 50, igraph.BipartiteRandomOptions{Seed: seed})
		}},
		{name: "iea", call: func(seed *uint64) (igraph.BipartiteGraphResult, error) {
			return igraph.NewBipartiteIEA(20, 15, 50, igraph.BipartiteRandomOptions{Directed: true, Direction: igraph.DirectionIn, Seed: seed})
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			seed := uint64(12345)
			first, err := test.call(&seed)
			if err != nil {
				t.Fatal(err)
			}
			defer first.Graph.Close()
			second, err := test.call(&seed)
			if err != nil {
				t.Fatal(err)
			}
			defer second.Graph.Close()
			firstEdges, _ := first.Graph.Edges()
			secondEdges, _ := second.Graph.Edges()
			if !reflect.DeepEqual(first.Partition, second.Partition) || !reflect.DeepEqual(firstEdges, secondEdges) {
				t.Errorf("same seed produced different results")
			}
		})
	}
}

func TestRandomBipartiteConcurrentSeededCalls(t *testing.T) {
	seed := uint64(99)
	reference, err := igraph.NewBipartiteGNP(30, 20, .25, igraph.BipartiteRandomOptions{Seed: &seed})
	if err != nil {
		t.Fatal(err)
	}
	want, _ := reference.Graph.Edges()
	reference.Graph.Close()

	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			result, err := igraph.NewBipartiteGNP(30, 20, .25, igraph.BipartiteRandomOptions{Seed: &seed})
			if err != nil {
				t.Errorf("concurrent call: %v", err)
				return
			}
			defer result.Graph.Close()
			got, err := result.Graph.Edges()
			if err != nil || !reflect.DeepEqual(got, want) {
				t.Errorf("concurrent seeded result differs: %v", err)
			}
		}()
	}
	wg.Wait()
}

func TestRandomBipartiteValidation(t *testing.T) {
	badOptions := igraph.BipartiteRandomOptions{Direction: igraph.DirectionMode(99)}
	for _, call := range []func() error{
		func() error { _, err := igraph.NewBipartiteGNP(-1, 1, .5, igraph.BipartiteRandomOptions{}); return err },
		func() error { _, err := igraph.NewBipartiteGNP(1, -1, .5, igraph.BipartiteRandomOptions{}); return err },
		func() error { _, err := igraph.NewBipartiteGNP(1, 1, -1, igraph.BipartiteRandomOptions{}); return err },
		func() error { _, err := igraph.NewBipartiteGNP(1, 1, 2, igraph.BipartiteRandomOptions{}); return err },
		func() error { _, err := igraph.NewBipartiteGNP(1, 1, .5, badOptions); return err },
		func() error { _, err := igraph.NewBipartiteGNM(1, 1, -1, igraph.BipartiteRandomOptions{}); return err },
		func() error { _, err := igraph.NewBipartiteIEA(1, 1, -1, igraph.BipartiteRandomOptions{}); return err },
	} {
		if err := call(); err == nil {
			t.Error("expected validation error")
		}
	}
}

func TestRandomBipartiteResultOwnership(t *testing.T) {
	result, err := igraph.NewBipartiteGNP(2, 2, 0, igraph.BipartiteRandomOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if result.Partition == nil {
		t.Fatal("partition must be non-nil")
	}
	if err := result.Graph.Close(); err != nil {
		t.Fatal(err)
	}
	result.Partition[0] = true
	if !result.Partition[0] {
		t.Error("partition must remain mutable after graph close")
	}
}

func assertRandomBipartiteResult(t *testing.T, result igraph.BipartiteGraphResult, n1, n2, edges int, directed bool) {
	t.Helper()
	if result.Partition == nil || len(result.Partition) != n1+n2 {
		t.Fatalf("partition = %v, want non-nil length %d", result.Partition, n1+n2)
	}
	for i, mode := range result.Partition {
		if mode != (i >= n1) {
			t.Errorf("partition[%d] = %t", i, mode)
		}
	}
	if got, err := result.Graph.VertexCount(); err != nil || got != n1+n2 {
		t.Errorf("VertexCount = %d, %v", got, err)
	}
	if got, err := result.Graph.EdgeCount(); err != nil || got != edges {
		t.Errorf("EdgeCount = %d, %v, want %d", got, err, edges)
	}
	if got, err := result.Graph.IsDirected(); err != nil || got != directed {
		t.Errorf("IsDirected = %t, %v", got, err)
	}
	ok, err := result.Graph.IsBipartitePartition(result.Partition)
	if err != nil || !ok {
		t.Errorf("IsBipartitePartition = %t, %v", ok, err)
	}
}

func assertRandomBipartiteOrientation(t *testing.T, result igraph.BipartiteGraphResult, options igraph.BipartiteRandomOptions) {
	t.Helper()
	if !options.Directed || options.Direction == igraph.DirectionAll {
		return
	}
	edges, err := result.Graph.Edges()
	if err != nil {
		t.Fatal(err)
	}
	for _, edge := range edges {
		fromMode, toMode := result.Partition[edge.From], result.Partition[edge.To]
		if options.Direction == igraph.DirectionOut && (fromMode || !toMode) {
			t.Errorf("out edge has wrong orientation: %v", edge)
		}
		if options.Direction == igraph.DirectionIn && (!fromMode || toMode) {
			t.Errorf("in edge has wrong orientation: %v", edge)
		}
	}
}
