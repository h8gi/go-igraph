package igraph_test

import (
	"errors"
	"reflect"
	"sync"
	"testing"

	igraph "github.com/h8gi/go-igraph"
)

func TestMilestone14IntegrationPipeline(t *testing.T) {
	matrix, err := igraph.NewMatrixFromRows([][]float64{
		{5, 1},
		{4, 6},
		{0, 3},
	})
	if err != nil {
		t.Fatal(err)
	}
	affiliation, err := igraph.NewWeightedBiadjacency(matrix, igraph.WeightedBiadjacencyOptions{})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = affiliation.Graph.Close() })

	recovered, err := affiliation.Graph.Biadjacency(affiliation.Partition, affiliation.Weights)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(recovered.RowVertexIDs, []int{0, 1, 2}) ||
		!reflect.DeepEqual(recovered.ColumnVertexIDs, []int{3, 4}) ||
		!reflect.DeepEqual(recovered.Matrix.Rows(), matrix.Rows()) {
		t.Fatalf("biadjacency round trip = rows %v, columns %v, matrix %v", recovered.RowVertexIDs, recovered.ColumnVertexIDs, recovered.Matrix.Rows())
	}

	sizes, err := affiliation.Graph.BipartiteProjectionSizes(affiliation.Partition)
	if err != nil || sizes.False != (igraph.BipartiteProjectionSize{Vertices: 3, Edges: 3}) || sizes.True != (igraph.BipartiteProjectionSize{Vertices: 2, Edges: 1}) {
		t.Fatalf("projection sizes = %#v, %v", sizes, err)
	}
	projections, err := affiliation.Graph.BipartiteProjections(affiliation.Partition)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = projections.False.Graph.Close() })
	t.Cleanup(func() { _ = projections.True.Graph.Close() })
	if !reflect.DeepEqual(projections.False.SourceVertexIDs, []int{0, 1, 2}) ||
		!reflect.DeepEqual(projections.False.Multiplicities, []int{2, 1, 1}) ||
		!reflect.DeepEqual(projections.True.SourceVertexIDs, []int{3, 4}) ||
		!reflect.DeepEqual(projections.True.Multiplicities, []int{2}) {
		t.Fatalf("projection provenance = false(%v, %v), true(%v, %v)", projections.False.SourceVertexIDs, projections.False.Multiplicities, projections.True.SourceVertexIDs, projections.True.Multiplicities)
	}

	unweighted, err := affiliation.Graph.MaximumBipartiteMatching(affiliation.Partition, igraph.BipartiteMatchingOptions{})
	if err != nil || unweighted.Size != 2 || len(unweighted.Pairs) != 2 {
		t.Fatalf("unweighted matching = %#v, %v", unweighted, err)
	}
	weighted, err := affiliation.Graph.MaximumBipartiteMatching(affiliation.Partition, igraph.BipartiteMatchingOptions{Weights: affiliation.Weights})
	if err != nil || weighted.Size != 2 || weighted.Weight != 11 || !reflect.DeepEqual(weighted.Pairs, []igraph.MatchedPair{{FalseVertex: 0, TrueVertex: 3}, {FalseVertex: 1, TrueVertex: 4}}) {
		t.Fatalf("weighted matching = %#v, %v", weighted, err)
	}

	layout, err := affiliation.Graph.LayoutBipartite([]bool(affiliation.Partition), 1, 1, 100)
	rows, columns := layout.Dims()
	if err != nil || rows != 5 || columns != 2 {
		t.Fatalf("bipartite layout = %v (%dx%d), %v", layout.Rows(), rows, columns, err)
	}

	seed := uint64(217)
	random, err := igraph.NewBipartiteGNM(3, 2, 4, igraph.BipartiteRandomOptions{Seed: &seed})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = random.Graph.Close() })
	randomReplay, err := igraph.NewBipartiteGNM(3, 2, 4, igraph.BipartiteRandomOptions{Seed: &seed})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = randomReplay.Graph.Close() })
	randomEdges, _ := random.Graph.Edges()
	replayEdges, _ := randomReplay.Graph.Edges()
	if !reflect.DeepEqual(randomEdges, replayEdges) || !reflect.DeepEqual(random.Partition, randomReplay.Partition) {
		t.Fatal("seeded bipartite generation did not replay exactly")
	}

	if err := affiliation.Graph.Close(); err != nil {
		t.Fatal(err)
	}
	if err := affiliation.Graph.Close(); err != nil {
		t.Fatalf("repeated source Close = %v", err)
	}
	if _, err := affiliation.Graph.Biadjacency(affiliation.Partition, affiliation.Weights); !errors.Is(err, igraph.ErrClosed) {
		t.Fatalf("post-close Biadjacency error = %v", err)
	}
	if count, err := projections.False.Graph.VertexCount(); err != nil || count != 3 {
		t.Fatalf("false projection after source close = %d, %v", count, err)
	}
	if err := projections.False.Graph.Close(); err != nil {
		t.Fatal(err)
	}
	if count, err := projections.True.Graph.VertexCount(); err != nil || count != 2 {
		t.Fatalf("true projection after sibling close = %d, %v", count, err)
	}
	if recovered.RowVertexIDs == nil || recovered.ColumnVertexIDs == nil || affiliation.Partition == nil || affiliation.Weights == nil || unweighted.Pairs == nil || weighted.Pairs == nil {
		t.Fatal("milestone 14 returned nil Go-owned storage")
	}

	projections.False.Graph.Close()
	projections.True.Graph.Close()
	random.Graph.Close()
	randomReplay.Graph.Close()
}

func TestMilestone14ConcurrentReadsAndClose(t *testing.T) {
	live := newMilestone14Graph(t)
	for err := range runMilestone14Calls(live) {
		if err != nil {
			t.Errorf("live concurrent bipartite call = %v", err)
		}
	}

	closing := newMilestone14Graph(t)
	calls := milestone14ReadCalls(closing)
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
	if err := closing.Graph.Close(); err != nil {
		t.Fatal(err)
	}
	wait.Wait()
	close(errorsByCall)
	for err := range errorsByCall {
		if err != nil && !errors.Is(err, igraph.ErrClosed) {
			t.Errorf("close-race bipartite call = %v", err)
		}
	}
}

func TestMilestone14ConcurrentSeedIsolation(t *testing.T) {
	seed := uint64(14)
	reference, err := igraph.NewBipartiteIEA(12, 9, 50, igraph.BipartiteRandomOptions{Directed: true, Direction: igraph.DirectionAll, Seed: &seed})
	if err != nil {
		t.Fatal(err)
	}
	want, _ := reference.Graph.Edges()
	reference.Graph.Close()

	var wait sync.WaitGroup
	for index := 0; index < 8; index++ {
		wait.Add(1)
		go func() {
			defer wait.Done()
			result, err := igraph.NewBipartiteIEA(12, 9, 50, igraph.BipartiteRandomOptions{Directed: true, Direction: igraph.DirectionAll, Seed: &seed})
			if err != nil {
				t.Errorf("seeded generation = %v", err)
				return
			}
			defer result.Graph.Close()
			got, err := result.Graph.Edges()
			if err != nil || !reflect.DeepEqual(got, want) {
				t.Errorf("isolated seeded result differs: %v", err)
			}
		}()
	}
	wait.Wait()
}

func newMilestone14Graph(t *testing.T) igraph.WeightedBipartiteGraphResult {
	t.Helper()
	matrix, err := igraph.NewMatrixFromRows([][]float64{{5, 1}, {4, 6}, {0, 3}})
	if err != nil {
		t.Fatal(err)
	}
	result, err := igraph.NewWeightedBiadjacency(matrix, igraph.WeightedBiadjacencyOptions{})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = result.Graph.Close() })
	return result
}

func milestone14ReadCalls(result igraph.WeightedBipartiteGraphResult) []func() error {
	return []func() error{
		func() error { _, err := result.Graph.Bipartite(); return err },
		func() error { _, err := result.Graph.IsBipartitePartition(result.Partition); return err },
		func() error { _, err := result.Graph.Biadjacency(result.Partition, result.Weights); return err },
		func() error { _, err := result.Graph.BipartiteProjectionSizes(result.Partition); return err },
		func() error {
			projection, err := result.Graph.BipartiteProjection(result.Partition, igraph.BipartiteModeFalse)
			if err == nil {
				err = projection.Graph.Close()
			}
			return err
		},
		func() error {
			_, err := result.Graph.MaximumBipartiteMatching(result.Partition, igraph.BipartiteMatchingOptions{})
			return err
		},
		func() error {
			_, err := result.Graph.MaximumBipartiteMatching(result.Partition, igraph.BipartiteMatchingOptions{Weights: result.Weights})
			return err
		},
		func() error { _, err := result.Graph.LayoutBipartite([]bool(result.Partition), 1, 1, 100); return err },
	}
}

func runMilestone14Calls(result igraph.WeightedBipartiteGraphResult) <-chan error {
	calls := milestone14ReadCalls(result)
	errorsByCall := make(chan error, len(calls))
	var wait sync.WaitGroup
	for _, call := range calls {
		wait.Add(1)
		go func(call func() error) {
			defer wait.Done()
			errorsByCall <- call()
		}(call)
	}
	wait.Wait()
	close(errorsByCall)
	return errorsByCall
}
