package igraph

import (
	"errors"
	"strings"
	"testing"
)

func TestFeedbackFailureAdapters(t *testing.T) {
	g := newCycleTestGraph(t, 3, []Edge{{0, 1}, {1, 2}, {2, 0}}, true)
	defer g.Close()
	failure := errors.New("simulated feedback failure")

	for _, vertex := range []bool{false, true} {
		resultInit := defaultFeedbackAdapters()
		resultInit.initializeResult = func() (*intVector, error) { return nil, failure }
		if _, err := callFeedbackWithAdapters(g, vertex, &resultInit); !errors.Is(err, failure) {
			t.Errorf("result initialization vertex=%v error = %v", vertex, err)
		}

		weightInit := defaultFeedbackAdapters()
		closes := 0
		weightInit.closeResult = func(vector *intVector) {
			closes++
			vector.close()
		}
		weightInit.initializeWeights = func([]float64, int, string) (*realVector, error) { return nil, failure }
		if _, err := callFeedbackWithWeightsAndAdapters(g, vertex, &weightInit); !errors.Is(err, failure) {
			t.Errorf("weight initialization vertex=%v error = %v", vertex, err)
		}
		if closes != 1 {
			t.Errorf("weight initialization vertex=%v closed %d results, want 1", vertex, closes)
		}

		upstream := defaultFeedbackAdapters()
		if vertex {
			upstream.vertexCall = func(*Graph, *intVector, *realVector) int { return 1 }
		} else {
			upstream.edgeCall = func(*Graph, *intVector, *realVector, FeedbackEdgeStrategy) int { return 1 }
		}
		if _, err := callFeedbackWithAdapters(g, vertex, &upstream); err == nil || !strings.Contains(err.Error(), "calculate feedback") {
			t.Errorf("upstream vertex=%v error = %v", vertex, err)
		}

		conversion := defaultFeedbackAdapters()
		conversion.convert = func(*intVector) ([]int, error) { return nil, failure }
		if _, err := callFeedbackWithAdapters(g, vertex, &conversion); !errors.Is(err, failure) {
			t.Errorf("conversion vertex=%v error = %v", vertex, err)
		}
	}
}

func TestFeedbackWeightedCleanup(t *testing.T) {
	g := newCycleTestGraph(t, 3, []Edge{{0, 0}, {1, 1}, {2, 2}}, true)
	defer g.Close()
	for _, vertex := range []bool{false, true} {
		adapters := defaultFeedbackAdapters()
		resultCloses, weightCloses := 0, 0
		adapters.closeResult = func(vector *intVector) {
			resultCloses++
			vector.close()
		}
		adapters.closeWeights = func(vector *realVector) {
			weightCloses++
			vector.close()
		}
		if vertex {
			adapters.vertexCall = func(*Graph, *intVector, *realVector) int { return 1 }
		} else {
			adapters.edgeCall = func(*Graph, *intVector, *realVector, FeedbackEdgeStrategy) int { return 1 }
		}
		if _, err := callFeedbackWithWeightsAndAdapters(g, vertex, &adapters); err == nil {
			t.Errorf("weighted upstream vertex=%v succeeded", vertex)
		}
		if resultCloses != 1 || weightCloses != 1 {
			t.Errorf("weighted cleanup vertex=%v result=%d weight=%d", vertex, resultCloses, weightCloses)
		}
	}
}

func callFeedbackWithAdapters(g *Graph, vertex bool, adapters *feedbackAdapters) ([]int, error) {
	if vertex {
		return g.feedbackVertexSet(nil, adapters)
	}
	return g.feedbackEdgeSet(FeedbackEdgeOptions{}, adapters)
}

func callFeedbackWithWeightsAndAdapters(g *Graph, vertex bool, adapters *feedbackAdapters) ([]int, error) {
	if vertex {
		return g.feedbackVertexSet([]float64{1, 1, 1}, adapters)
	}
	return g.feedbackEdgeSet(FeedbackEdgeOptions{Weights: []float64{1, 1, 1}}, adapters)
}
