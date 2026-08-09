package igraph

import (
	"math"
	"testing"
)

func TestMilestone4DirectedWeightedSelectorIntegration(t *testing.T) {
	graph, err := NewGraphFromEdges(5, []Edge{
		{0, 1}, {0, 2}, {1, 2}, {2, 0}, {2, 3}, {3, 4}, {4, 3},
	}, true)
	if err != nil {
		t.Fatal(err)
	}
	weights := []float64{2, 5, 1, 3, 2, 1, 1}
	vertices, _ := VertexIDs(2, 0, 2, 4)
	cutoff := 5.0
	closeness, err := graph.Closeness(vertices, DistanceCentralityOptions{
		Direction: DirectionOut, Weights: weights, Normalized: true, Cutoff: &cutoff,
	})
	if err != nil {
		t.Fatal(err)
	}
	harmonic, err := graph.HarmonicCentrality(vertices, DistanceCentralityOptions{
		Direction: DirectionOut, Weights: weights, Normalized: true, Cutoff: &cutoff,
	})
	if err != nil {
		t.Fatal(err)
	}
	if closeness.Scores == nil || closeness.ReachableCounts == nil || harmonic == nil {
		t.Fatal("distance centrality returned a nil collection")
	}
	assertFloat(t, closeness.Scores[0], closeness.Scores[2])
	assertFloat(t, harmonic[0], harmonic[2])
	if closeness.ReachableCounts[0] != closeness.ReachableCounts[2] {
		t.Errorf("duplicate closeness reachability differs: %v", closeness.ReachableCounts)
	}

	vertexBetweenness, err := graph.VertexBetweenness(vertices, BetweennessOptions{
		Weights: weights, DirectedPaths: true, Normalized: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	edges, _ := EdgeIDs(4, 0, 4, 6)
	edgeBetweenness, err := graph.EdgeBetweenness(edges, BetweennessOptions{
		Weights: weights, DirectedPaths: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	assertFloat(t, vertexBetweenness[0], vertexBetweenness[2])
	assertFloat(t, edgeBetweenness[0], edgeBetweenness[2])

	pageRank, err := graph.PageRank(vertices, PageRankOptions{
		Weights: weights, Direction: PageRankRespectDirections,
	})
	if err != nil {
		t.Fatal(err)
	}
	allPageRank, err := graph.PageRank(AllVertices(), PageRankOptions{Weights: weights})
	if err != nil {
		t.Fatal(err)
	}
	assertFloat(t, pageRank.Scores[0], pageRank.Scores[2])
	assertProbabilityScores(t, allPageRank.Scores)

	degreeCentralization, err := graph.DegreeCentralization(DegreeCentralizationOptions{Direction: DirectionIn})
	if err != nil {
		t.Fatal(err)
	}
	degrees, err := graph.Degree(AllVertices(), DegreeOptions{Direction: DirectionIn})
	if err != nil {
		t.Fatal(err)
	}
	for index, degree := range degrees {
		assertFloat(t, degreeCentralization.Scores[index], float64(degree))
	}
	betweennessCentralization, err := graph.BetweennessCentralization(BetweennessCentralizationOptions{DirectedPaths: true})
	if err != nil {
		t.Fatal(err)
	}
	allBetweenness, err := graph.VertexBetweenness(AllVertices(), BetweennessOptions{DirectedPaths: true})
	if err != nil {
		t.Fatal(err)
	}
	assertFloatSlice(t, betweennessCentralization.Scores, allBetweenness)

	if err := graph.Close(); err != nil {
		t.Fatal(err)
	}
	assertFloat(t, closeness.Scores[0], closeness.Scores[2])
	assertFloat(t, edgeBetweenness[0], edgeBetweenness[2])
	assertProbabilityScores(t, allPageRank.Scores)
}

func TestMilestone4DisconnectedAndDegenerateIntegration(t *testing.T) {
	disconnected, err := NewGraphFromEdges(4, []Edge{{0, 1}, {1, 2}}, false)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = disconnected.Close() })
	closeness, err := disconnected.Closeness(AllVertices(), DistanceCentralityOptions{})
	if err != nil {
		t.Fatal(err)
	}
	harmonic, err := disconnected.HarmonicCentrality(AllVertices(), DistanceCentralityOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if !math.IsNaN(closeness.Scores[3]) || harmonic[3] != 0 {
		t.Errorf("isolated vertex scores: closeness=%v harmonic=%v", closeness.Scores, harmonic)
	}
	centralization, err := disconnected.ClosenessCentralization(ClosenessCentralizationOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if !math.IsNaN(centralization.Value) || !math.IsNaN(centralization.Scores[3]) {
		t.Errorf("disconnected centralization = %#v", centralization)
	}
	pageRank, err := disconnected.PageRank(AllVertices(), PageRankOptions{})
	if err != nil {
		t.Fatal(err)
	}
	assertProbabilityScores(t, pageRank.Scores)

	single, err := NewGraphFromEdges(1, nil, false)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = single.Close() })
	eigenvector, err := single.EigenvectorCentrality(EigenvectorCentralityOptions{})
	if err != nil {
		t.Fatal(err)
	}
	hits, err := single.HITS(HITSOptions{})
	if err != nil {
		t.Fatal(err)
	}
	singlePageRank, err := single.PageRank(AllVertices(), PageRankOptions{})
	if err != nil {
		t.Fatal(err)
	}
	assertFloatSlice(t, eigenvector.Scores, []float64{1})
	assertFloatSlice(t, hits.HubScores, []float64{1})
	assertFloatSlice(t, hits.AuthorityScores, []float64{1})
	assertFloatSlice(t, singlePageRank.Scores, []float64{1})
}
