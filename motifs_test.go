package igraph_test

import (
	"errors"
	"fmt"
	"math"
	"reflect"
	"sort"
	"sync"
	"testing"

	igraph "github.com/h8gi/go-igraph"
)

func TestDyadCensusKnownAnswers(t *testing.T) {
	tests := []struct {
		name     string
		vertices int
		edges    []igraph.Edge
		directed bool
		want     igraph.DyadCensusResult
	}{
		{name: "empty", directed: true, want: igraph.DyadCensusResult{}},
		{name: "isolates", vertices: 3, directed: true, want: igraph.DyadCensusResult{Null: 3}},
		{name: "directed mixed", vertices: 3, directed: true, edges: []igraph.Edge{
			{From: 0, To: 1}, {From: 1, To: 0}, {From: 1, To: 2},
		}, want: igraph.DyadCensusResult{Mutual: 1, Asymmetric: 1, Null: 1}},
		{name: "undirected edge is mutual", vertices: 2, directed: false,
			edges: []igraph.Edge{{From: 0, To: 1}}, want: igraph.DyadCensusResult{Mutual: 1}},
		{name: "loops and multiplicity ignored", vertices: 2, directed: true, edges: []igraph.Edge{
			{From: 0, To: 0}, {From: 0, To: 1}, {From: 0, To: 1},
		}, want: igraph.DyadCensusResult{Asymmetric: 1}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			graph := newMotifGraph(t, test.vertices, test.edges, test.directed)
			got, err := graph.DyadCensus()
			if err != nil || got != test.want {
				t.Fatalf("DyadCensus = %#v, %v; want %#v", got, err, test.want)
			}
		})
	}
}

func TestMotifsRandesuKnownAnswersAndOwnership(t *testing.T) {
	graph := newMotifGraph(t, 4, []igraph.Edge{
		{From: 0, To: 1}, {From: 0, To: 2}, {From: 0, To: 3},
		{From: 1, To: 2}, {From: 1, To: 3}, {From: 2, To: 3},
	}, false)
	options := igraph.MotifsRandesuOptions{Size: 3}
	histogram, err := graph.MotifsRandesu(options)
	if err != nil {
		t.Fatal(err)
	}
	if histogram == nil {
		t.Fatal("MotifsRandesu returned nil histogram")
	}
	var total float64
	for _, count := range histogram {
		if !math.IsNaN(count) {
			total += count
		}
	}
	if total != 4 {
		t.Fatalf("MotifsRandesu total = %v, want 4: %v", total, histogram)
	}
	count, err := graph.MotifsRandesuNo(options)
	if err != nil || count != 4 {
		t.Fatalf("MotifsRandesuNo = %d, %v; want 4", count, err)
	}
	if err := graph.Close(); err != nil {
		t.Fatal(err)
	}
	histogram[0] = 99
}

func TestMotifsRandesuHistogramShapes(t *testing.T) {
	for _, directed := range []bool{false, true} {
		graph := newMotifGraph(t, 4, nil, directed)
		for _, size := range []int{3, 4} {
			histogram, err := graph.MotifsRandesu(igraph.MotifsRandesuOptions{Size: size})
			if err != nil {
				t.Fatalf("directed=%t size=%d: %v", directed, size, err)
			}
			expected := map[bool]map[int]int{
				false: {3: 4, 4: 11},
				true:  {3: 16, 4: 218},
			}[directed][size]
			if len(histogram) != expected {
				t.Errorf("directed=%t size=%d histogram length = %d, want %d", directed, size, len(histogram), expected)
			}
		}
	}
}

func TestMotifsRandesuEstimateSamplingModes(t *testing.T) {
	graph := newMotifGraph(t, 5, []igraph.Edge{
		{0, 1}, {0, 2}, {0, 3}, {0, 4}, {1, 2},
		{1, 3}, {1, 4}, {2, 3}, {2, 4}, {3, 4},
	}, false)
	seed := uint64(17)
	random, err := graph.MotifsRandesuEstimate(igraph.MotifsRandesuEstimateOptions{
		Size: 3, SampleSize: 3, Seed: &seed,
	})
	if err != nil || random != 5.0/3.0 {
		t.Fatalf("random MotifsRandesuEstimate = %v, %v; want %v", random, err, 5.0/3.0)
	}
	vertices, err := igraph.VertexIDs(4, 1, 3)
	if err != nil {
		t.Fatal(err)
	}
	explicit, err := graph.MotifsRandesuEstimate(igraph.MotifsRandesuEstimateOptions{
		Size: 3, SampleVertices: vertices,
	})
	if err != nil || explicit != 5 {
		t.Fatalf("explicit MotifsRandesuEstimate = %v, %v; want 5", explicit, err)
	}
}

func TestMotifsRandesuValidatesOptions(t *testing.T) {
	graph := newMotifGraph(t, 4, nil, true)
	invalidCounts := []igraph.MotifsRandesuOptions{
		{Size: 2},
		{Size: 5},
		{Size: 3, CutProb: []float64{0, 0}},
		{Size: 3, CutProb: []float64{0, math.NaN(), 0}},
		{Size: 3, CutProb: []float64{0, -0.1, 0}},
		{Size: 3, CutProb: []float64{0, 1.1, 0}},
	}
	for index, options := range invalidCounts {
		if result, err := graph.MotifsRandesu(options); err == nil || result != nil {
			t.Errorf("MotifsRandesu invalid option %d = %#v, %v", index, result, err)
		}
		if result, err := graph.MotifsRandesuNo(options); err == nil || result != 0 {
			t.Errorf("MotifsRandesuNo invalid option %d = %d, %v", index, result, err)
		}
	}

	duplicate, _ := igraph.VertexIDs(0, 0)
	outOfRange, _ := igraph.VertexIDs(4)
	invalidSamples := []igraph.MotifsRandesuEstimateOptions{
		{Size: 3},
		{Size: 3, SampleSize: -1},
		{Size: 3, SampleSize: 5},
		{Size: 3, SampleSize: 2, SampleVertices: igraph.NoVertices()},
		{Size: 3, SampleVertices: igraph.NoVertices()},
		{Size: 3, SampleVertices: duplicate},
		{Size: 3, SampleVertices: outOfRange},
	}
	for index, options := range invalidSamples {
		if result, err := graph.MotifsRandesuEstimate(options); err == nil || result != 0 {
			t.Errorf("MotifsRandesuEstimate invalid option %d = %v, %v", index, result, err)
		}
	}
}

func TestMotifsRandesuSeedIsolation(t *testing.T) {
	edges := []igraph.Edge{
		{0, 1}, {1, 2}, {2, 0}, {2, 3}, {3, 4}, {4, 2},
		{1, 4}, {4, 5}, {5, 1}, {0, 5}, {5, 6}, {6, 0},
	}
	graphs := []*igraph.Graph{
		newMotifGraph(t, 7, edges, true),
		newMotifGraph(t, 7, edges, true),
	}
	seeds := []uint64{41, 99}
	want := make([][]float64, len(seeds))
	for index := range seeds {
		var err error
		want[index], err = graphs[index].MotifsRandesu(igraph.MotifsRandesuOptions{
			Size: 3, CutProb: []float64{0.2, 0.4, 0.6}, Seed: &seeds[index],
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	var wait sync.WaitGroup
	errors := make(chan error, 20)
	for run := 0; run < 20; run++ {
		index := run % len(seeds)
		wait.Add(1)
		go func() {
			defer wait.Done()
			got, err := graphs[index].MotifsRandesu(igraph.MotifsRandesuOptions{
				Size: 3, CutProb: []float64{0.2, 0.4, 0.6}, Seed: &seeds[index],
			})
			if err != nil {
				errors <- err
				return
			}
			if !equalFloatSlicesWithNaN(got, want[index]) {
				errors <- errorsNewMotifMismatch(got, want[index])
			}
		}()
	}
	wait.Wait()
	close(errors)
	for err := range errors {
		t.Error(err)
	}
}

func equalFloatSlicesWithNaN(left, right []float64) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] && !(math.IsNaN(left[index]) && math.IsNaN(right[index])) {
			return false
		}
	}
	return true
}

func errorsNewMotifMismatch(got, want []float64) error {
	return fmt.Errorf("seeded RANDESU result mismatch: got %v, want %v", got, want)
}

func TestTriadCensusKnownAnswersAndOwnership(t *testing.T) {
	graph := newMotifGraph(t, 4, []igraph.Edge{
		{From: 0, To: 1}, {From: 1, To: 2}, {From: 2, To: 0},
	}, true)
	result, err := graph.TriadCensus()
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 16 {
		t.Fatalf("TriadCensus length = %d", len(result))
	}
	var total int64
	for _, count := range result {
		total += count
	}
	if total != 4 {
		t.Fatalf("TriadCensus total = %d, want C(4,3)=4: %v", total, result)
	}
	if err := graph.Close(); err != nil {
		t.Fatal(err)
	}
	result[0]++
}

func TestTriangleQueriesIgnoreDirectionMultiplicityAndLoops(t *testing.T) {
	edges := []igraph.Edge{
		{From: 0, To: 1}, {From: 0, To: 2}, {From: 1, To: 2},
		{From: 0, To: 1}, {From: 2, To: 2},
	}
	graph := newMotifGraph(t, 4, edges, true)
	count, err := graph.TrianglesCount()
	if err != nil || count != 1 {
		t.Fatalf("TrianglesCount = %d, %v", count, err)
	}
	triangles, err := graph.TrianglesList()
	if err != nil {
		t.Fatal(err)
	}
	if got := canonicalTriangles(triangles); !reflect.DeepEqual(got, [][3]int{{0, 1, 2}}) {
		t.Fatalf("TrianglesList = %v", triangles)
	}
	selector, err := igraph.VertexIDs(2, 0, 3, 2)
	if err != nil {
		t.Fatal(err)
	}
	adjacent, err := graph.AdjacentTrianglesCount(selector)
	if err != nil || !reflect.DeepEqual(adjacent, []int64{1, 1, 0, 1}) {
		t.Fatalf("AdjacentTrianglesCount = %v, %v", adjacent, err)
	}
	if err := graph.Close(); err != nil {
		t.Fatal(err)
	}
	triangles[0][0] = 99
	adjacent[0] = 99
}

func TestTriangleQueriesEmptyAndNoSelection(t *testing.T) {
	empty := newMotifGraph(t, 0, nil, false)
	if count, err := empty.TrianglesCount(); err != nil || count != 0 {
		t.Fatalf("empty TrianglesCount = %d, %v", count, err)
	}
	if triangles, err := empty.TrianglesList(); err != nil || triangles == nil || len(triangles) != 0 {
		t.Fatalf("empty TrianglesList = %#v, %v", triangles, err)
	}
	if counts, err := empty.AdjacentTrianglesCount(igraph.NoVertices()); err != nil || counts == nil || len(counts) != 0 {
		t.Fatalf("empty AdjacentTrianglesCount = %#v, %v", counts, err)
	}
}

func TestAdjacentTrianglesCountValidatesSelector(t *testing.T) {
	graph := newMotifGraph(t, 3, nil, false)
	selector, err := igraph.VertexIDs(3)
	if err != nil {
		t.Fatal(err)
	}
	if result, err := graph.AdjacentTrianglesCount(selector); err == nil || result != nil {
		t.Fatalf("invalid selector = %#v, %v", result, err)
	}
}

func TestMotifQueriesClosedAndNil(t *testing.T) {
	var nilGraph *igraph.Graph
	if _, err := nilGraph.DyadCensus(); !errors.Is(err, igraph.ErrClosed) {
		t.Errorf("nil DyadCensus error = %v", err)
	}
	if _, err := nilGraph.MotifsRandesu(igraph.MotifsRandesuOptions{Size: 3}); !errors.Is(err, igraph.ErrClosed) {
		t.Errorf("nil MotifsRandesu error = %v", err)
	}
	if _, err := nilGraph.MotifsRandesuEstimate(igraph.MotifsRandesuEstimateOptions{Size: 3, SampleSize: 1}); !errors.Is(err, igraph.ErrClosed) {
		t.Errorf("nil MotifsRandesuEstimate error = %v", err)
	}
	if _, err := nilGraph.MotifsRandesuNo(igraph.MotifsRandesuOptions{Size: 3}); !errors.Is(err, igraph.ErrClosed) {
		t.Errorf("nil MotifsRandesuNo error = %v", err)
	}
	graph := newMotifGraph(t, 3, nil, true)
	if err := graph.Close(); err != nil {
		t.Fatal(err)
	}
	calls := []func() error{
		func() error { _, err := graph.DyadCensus(); return err },
		func() error { _, err := graph.TriadCensus(); return err },
		func() error { _, err := graph.AdjacentTrianglesCount(igraph.AllVertices()); return err },
		func() error { _, err := graph.TrianglesCount(); return err },
		func() error { _, err := graph.TrianglesList(); return err },
		func() error { _, err := graph.MotifsRandesu(igraph.MotifsRandesuOptions{Size: 3}); return err },
		func() error {
			_, err := graph.MotifsRandesuEstimate(igraph.MotifsRandesuEstimateOptions{Size: 3, SampleSize: 1})
			return err
		},
		func() error { _, err := graph.MotifsRandesuNo(igraph.MotifsRandesuOptions{Size: 3}); return err },
	}
	for index, call := range calls {
		if err := call(); !errors.Is(err, igraph.ErrClosed) {
			t.Errorf("closed motif call %d error = %v", index, err)
		}
	}
}

func newMotifGraph(t *testing.T, vertices int, edges []igraph.Edge, directed bool) *igraph.Graph {
	t.Helper()
	graph, err := igraph.NewGraphFromEdges(vertices, edges, directed)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = graph.Close() })
	return graph
}

func canonicalTriangles(values [][3]int) [][3]int {
	result := append([][3]int{}, values...)
	for index := range result {
		sort.Ints(result[index][:])
	}
	sort.Slice(result, func(i, j int) bool {
		for column := 0; column < 3; column++ {
			if result[i][column] != result[j][column] {
				return result[i][column] < result[j][column]
			}
		}
		return false
	})
	return result
}
