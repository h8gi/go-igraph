package igraph_test

import (
	"errors"
	"reflect"
	"sync"
	"testing"

	igraph "github.com/h8gi/go-igraph"
)

func milestone26Graph(t *testing.T) *igraph.Graph {
	t.Helper()
	g, err := igraph.NewGraphFromEdges(6, []igraph.Edge{{From: 0, To: 1}, {From: 1, To: 2}, {From: 2, To: 3}, {From: 3, To: 4}, {From: 4, To: 5}, {From: 5, To: 0}, {From: 0, To: 3}, {From: 1, To: 4}}, false)
	if err != nil {
		t.Fatal(err)
	}
	return g
}

func closeMilestone26Graphs(t *testing.T, graphs []*igraph.Graph) {
	t.Helper()
	for _, graph := range graphs {
		if err := graph.Close(); err != nil {
			t.Fatal(err)
		}
	}
}

func milestone26Edges(t *testing.T, graphs []*igraph.Graph) [][]igraph.Edge {
	t.Helper()
	result := make([][]igraph.Edge, len(graphs))
	for i, graph := range graphs {
		edges, err := graph.Edges()
		if err != nil {
			t.Fatal(err)
		}
		result[i] = edges
	}
	return result
}

func TestMilestone26ConstructFitAnalyzeSampleWorkflow(t *testing.T) {
	graph := milestone26Graph(t)
	seed := uint64(2026)
	fitted, err := graph.FitHRG(igraph.HRGFitOptions{Steps: 100, Seed: &seed})
	if err != nil {
		t.Fatal(err)
	}
	replayed, err := graph.FitHRG(igraph.HRGFitOptions{Steps: 100, Seed: &seed})
	if err != nil || !reflect.DeepEqual(fitted, replayed) {
		t.Fatalf("fresh fit replay = %#v, %v", replayed, err)
	}

	dendrogram, probabilities, err := fitted.Dendrogram()
	if err != nil {
		t.Fatal(err)
	}
	roundTrip, err := igraph.NewHRGModelFromDendrogram(dendrogram, probabilities)
	if err != nil {
		t.Fatal(err)
	}
	if roundTrip.LeafCount() != fitted.LeafCount() || !reflect.DeepEqual(roundTrip.Probabilities(), fitted.Probabilities()) {
		t.Fatalf("round trip lost model alignment")
	}
	warm, err := graph.FitHRG(igraph.HRGFitOptions{Steps: 25, Seed: &seed, StartingModel: &roundTrip})
	if err != nil || warm.LeafCount() != 6 {
		t.Fatalf("warm fit = %#v, %v", warm, err)
	}

	analysis := igraph.HRGAnalysisOptions{Samples: 50, Seed: &seed, StartingModel: &warm}
	consensus, err := graph.ConsensusHRG(analysis)
	if err != nil {
		t.Fatal(err)
	}
	predictionOptions := igraph.HRGPredictionOptions{Samples: 50, Bins: 5, Seed: &seed, StartingModel: &warm}
	prediction, err := graph.PredictHRG(predictionOptions)
	if err != nil {
		t.Fatal(err)
	}
	if len(consensus.Parents) != 6+len(consensus.Weights) || len(prediction.Edges) != len(prediction.Probabilities) {
		t.Fatalf("analysis alignment = %d/%d, %d/%d", len(consensus.Parents), len(consensus.Weights), len(prediction.Edges), len(prediction.Probabilities))
	}

	samples, err := warm.Sample(3, &seed)
	if err != nil {
		t.Fatal(err)
	}
	defer closeMilestone26Graphs(t, samples)
	again, err := warm.Sample(3, &seed)
	if err != nil {
		t.Fatal(err)
	}
	defer closeMilestone26Graphs(t, again)
	if !reflect.DeepEqual(milestone26Edges(t, samples), milestone26Edges(t, again)) {
		t.Fatal("sample seed replay differs")
	}

	if err := graph.Close(); err != nil {
		t.Fatal(err)
	}
	if err := dendrogram.Close(); err != nil {
		t.Fatal(err)
	}
	if fitted.LeafCount() != 6 || len(consensus.Parents) == 0 || prediction.Probabilities == nil {
		t.Fatal("Go-owned results changed after source closure")
	}
	if err := samples[0].Close(); err != nil {
		t.Fatal(err)
	}
	if vertices, err := samples[1].VertexCount(); err != nil || vertices != 6 {
		t.Fatalf("sibling ownership = %d, %v", vertices, err)
	}
}

func TestMilestone26ConcurrentAnalysisAndCloseRace(t *testing.T) {
	graph := milestone26Graph(t)
	seed := uint64(9)
	var wg sync.WaitGroup
	for operation := range 6 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var err error
			switch operation % 3 {
			case 0:
				_, err = graph.FitHRG(igraph.HRGFitOptions{Steps: 20, Seed: &seed})
			case 1:
				_, err = graph.ConsensusHRG(igraph.HRGAnalysisOptions{Samples: 20, Seed: &seed})
			default:
				_, err = graph.PredictHRG(igraph.HRGPredictionOptions{Samples: 20, Bins: 4, Seed: &seed})
			}
			if err != nil && !errors.Is(err, igraph.ErrClosed) {
				t.Errorf("close race: %v", err)
			}
		}()
	}
	wg.Add(1)
	go func() { defer wg.Done(); _ = graph.Close() }()
	wg.Wait()
}
