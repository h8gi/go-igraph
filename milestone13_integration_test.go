package igraph_test

import (
	"errors"
	"math"
	"reflect"
	"sync"
	"testing"

	igraph "github.com/h8gi/go-igraph"
)

func TestMilestone13IntegrationPipeline(t *testing.T) {
	edges := []igraph.Edge{
		{From: 0, To: 1}, {From: 0, To: 2}, {From: 1, To: 2}, {From: 1, To: 3},
		{From: 1, To: 4}, {From: 2, To: 3}, {From: 2, To: 4}, {From: 3, To: 4},
	}
	weights := []float64{2, 2, 3, 1, 1, 4, 4, 4}
	graph := newMilestone13Graph(t, 5, edges)

	dyads, err := graph.DyadCensus()
	if err != nil || dyads != (igraph.DyadCensusResult{Mutual: 8, Null: 2}) {
		t.Fatalf("DyadCensus = %#v, %v", dyads, err)
	}
	triads, err := graph.TriadCensus()
	if err != nil || sumInt64(triads) != 10 {
		t.Fatalf("TriadCensus total = %d, %v: %v", sumInt64(triads), err, triads)
	}
	triangles, err := graph.TrianglesList()
	if err != nil || len(triangles) != 5 {
		t.Fatalf("TrianglesList = %v, %v", triangles, err)
	}
	triangleCount, err := graph.TrianglesCount()
	if err != nil || triangleCount != int64(len(triangles)) {
		t.Fatalf("TrianglesCount = %d, %v; list length %d", triangleCount, err, len(triangles))
	}
	adjacent, err := graph.AdjacentTrianglesCount(igraph.AllVertices())
	if err != nil || sumInt64(adjacent) != 3*triangleCount {
		t.Fatalf("AdjacentTrianglesCount = %v, %v", adjacent, err)
	}

	motifOptions := igraph.MotifsRandesuOptions{Size: 3}
	histogram, err := graph.MotifsRandesu(motifOptions)
	if err != nil {
		t.Fatal(err)
	}
	motifCount, err := graph.MotifsRandesuNo(motifOptions)
	if err != nil || finiteHistogramTotal(histogram) != float64(motifCount) {
		t.Fatalf("RANDESU histogram total = %v, count = %d, error = %v", finiteHistogramTotal(histogram), motifCount, err)
	}
	sample, err := igraph.VertexRange(0, 5)
	if err != nil {
		t.Fatal(err)
	}
	estimate, err := graph.MotifsRandesuEstimate(igraph.MotifsRandesuEstimateOptions{
		Size: 3, SampleVertices: sample,
	})
	if err != nil || estimate != float64(motifCount) {
		t.Fatalf("MotifsRandesuEstimate = %v, %v; want %d", estimate, err, motifCount)
	}

	basis, err := graph.GraphletsCandidateBasis(weights)
	if err != nil || len(basis.Cliques) == 0 || len(basis.Cliques) != len(basis.Thresholds) {
		t.Fatalf("GraphletsCandidateBasis = %#v, %v", basis, err)
	}
	projected, err := graph.GraphletsProject(basis.Cliques, nil, weights, 1000)
	if err != nil || len(projected) != len(basis.Cliques) {
		t.Fatalf("GraphletsProject = %v, %v", projected, err)
	}
	decomposition, err := graph.Graphlets(weights, 1000)
	if err != nil || len(decomposition.Cliques) != len(decomposition.Mu) {
		t.Fatalf("Graphlets = %#v, %v", decomposition, err)
	}
	restarted, err := graph.GraphletsProject(decomposition.Cliques, decomposition.Mu, weights, 0)
	if err != nil || !reflect.DeepEqual(restarted, decomposition.Mu) {
		t.Fatalf("GraphletsProject startMu = %v, %v; want %v", restarted, err, decomposition.Mu)
	}

	if err := graph.Close(); err != nil {
		t.Fatal(err)
	}
	if triads == nil || triangles == nil || adjacent == nil || histogram == nil ||
		basis.Cliques == nil || basis.Thresholds == nil || projected == nil ||
		decomposition.Cliques == nil || decomposition.Mu == nil {
		t.Fatal("Milestone 13 returned nil Go-owned storage")
	}
	basis.Cliques[0][0] = 99
	if decomposition.Cliques[0][0] == 99 {
		t.Fatal("candidate and decomposition clique results share backing storage")
	}
	projected[0] = 99
	if decomposition.Mu[0] == 99 {
		t.Fatal("projection and decomposition coefficients share backing storage")
	}
}

func TestMilestone13ConcurrentReadsAndClose(t *testing.T) {
	edges := []igraph.Edge{
		{From: 0, To: 1}, {From: 0, To: 2}, {From: 1, To: 2}, {From: 1, To: 3},
		{From: 1, To: 4}, {From: 2, To: 3}, {From: 2, To: 4}, {From: 3, To: 4},
	}
	weights := []float64{2, 2, 3, 1, 1, 4, 4, 4}
	live := newMilestone13Graph(t, 5, edges)
	for err := range runMilestone13Calls(milestone13ReadCalls(live, weights), nil) {
		if err != nil {
			t.Errorf("live concurrent motif/graphlet call error = %v", err)
		}
	}

	closing := newMilestone13Graph(t, 5, edges)
	for err := range runMilestone13Calls(milestone13ReadCalls(closing, weights), func() {
		if err := closing.Close(); err != nil {
			t.Error(err)
		}
	}) {
		if err != nil && !errors.Is(err, igraph.ErrClosed) {
			t.Errorf("Close-race motif/graphlet call error = %v", err)
		}
	}
	if _, err := closing.MotifsRandesu(igraph.MotifsRandesuOptions{Size: 3}); !errors.Is(err, igraph.ErrClosed) {
		t.Errorf("post-close MotifsRandesu error = %v", err)
	}
}

func milestone13ReadCalls(graph *igraph.Graph, weights []float64) []func() error {
	seed := uint64(204)
	return []func() error{
		func() error { _, err := graph.DyadCensus(); return err },
		func() error { _, err := graph.TriadCensus(); return err },
		func() error { _, err := graph.AdjacentTrianglesCount(igraph.AllVertices()); return err },
		func() error { _, err := graph.TrianglesCount(); return err },
		func() error { _, err := graph.TrianglesList(); return err },
		func() error {
			_, err := graph.MotifsRandesu(igraph.MotifsRandesuOptions{Size: 3, Seed: &seed})
			return err
		},
		func() error {
			_, err := graph.MotifsRandesuEstimate(igraph.MotifsRandesuEstimateOptions{
				Size: 3, SampleSize: 3, Seed: &seed,
			})
			return err
		},
		func() error {
			_, err := graph.MotifsRandesuNo(igraph.MotifsRandesuOptions{Size: 3, Seed: &seed})
			return err
		},
		func() error { _, err := graph.Graphlets(weights, 2); return err },
		func() error { _, err := graph.GraphletsCandidateBasis(weights); return err },
		func() error {
			_, err := graph.GraphletsProject([][]int{{0, 1, 2}}, nil, weights, 2)
			return err
		},
	}
}

func runMilestone13Calls(calls []func() error, afterStart func()) <-chan error {
	start := make(chan struct{})
	errorsByCall := make(chan error, len(calls))
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
	if afterStart != nil {
		afterStart()
	}
	wait.Wait()
	close(errorsByCall)
	return errorsByCall
}

func newMilestone13Graph(t *testing.T, vertices int, edges []igraph.Edge) *igraph.Graph {
	t.Helper()
	graph, err := igraph.NewGraphFromEdges(vertices, edges, false)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = graph.Close() })
	return graph
}

func sumInt64(values []int64) int64 {
	var total int64
	for _, value := range values {
		total += value
	}
	return total
}

func finiteHistogramTotal(values []float64) float64 {
	var total float64
	for _, value := range values {
		if !math.IsNaN(value) {
			total += value
		}
	}
	return total
}
