package igraph_test

import (
	"errors"
	"math"
	"sync"
	"testing"

	igraph "github.com/h8gi/go-igraph"
)

func milestone27Graph(t *testing.T) (*igraph.Graph, []float64) {
	t.Helper()
	graph, err := igraph.NewGraphFromEdges(4, []igraph.Edge{
		{From: 0, To: 1}, {From: 1, To: 2}, {From: 2, To: 0}, {From: 2, To: 3},
	}, false)
	if err != nil {
		t.Fatal(err)
	}
	return graph, []float64{1, 2, 3, 4}
}

func TestMilestone27CentralityAndLocalClusteringWorkflow(t *testing.T) {
	graph, weights := milestone27Graph(t)
	vertices, _ := igraph.VertexIDs(2, 1, 2, 0)
	edges, _ := igraph.EdgeIDs(3, 2, 3, 0)
	sources, _ := igraph.VertexIDs(0, 0)
	targets, _ := igraph.VertexIDs(3, 3)

	vertexSubset, err := graph.VertexBetweennessSubset(vertices, sources, targets, igraph.SubsetBetweennessOptions{})
	if err != nil {
		t.Fatal(err)
	}
	edgeSubset, err := graph.EdgeBetweennessSubset(edges, sources, targets, igraph.SubsetBetweennessOptions{})
	if err != nil {
		t.Fatal(err)
	}
	assertIntegrationFloats(t, vertexSubset, []float64{0.5, 0, 0.5, 0})
	assertIntegrationFloats(t, edgeSubset, []float64{0.5, 0.5, 0.5, 0})

	constraint, err := graph.BurtConstraint(vertices, weights)
	if err != nil || len(constraint) != 4 || constraint[0] != constraint[2] {
		t.Fatalf("constraint alignment = %#v, %v", constraint, err)
	}
	barrat, err := graph.BarratTransitivity(vertices, weights, igraph.TransitivityZero)
	if err != nil || len(barrat) != 4 || barrat[0] != barrat[2] {
		t.Fatalf("Barrat alignment = %#v, %v", barrat, err)
	}
	convergence, err := graph.EdgeConvergenceDegree()
	if err != nil {
		t.Fatal(err)
	}
	if len(convergence.Convergence) != 4 || len(convergence.InputSetSizes) != 4 || len(convergence.OutputSetSizes) != 4 {
		t.Fatalf("convergence alignment = %#v", convergence)
	}
	for edgeID, value := range convergence.Convergence {
		input, output := convergence.InputSetSizes[edgeID], convergence.OutputSetSizes[edgeID]
		if input+output != 0 && math.Abs(value-math.Abs((input-output)/(input+output))) > 1e-12 {
			t.Errorf("convergence invariant edge %d = %v, %v/%v", edgeID, value, input, output)
		}
	}
	clustering, err := graph.EdgeClustering(edges, igraph.EdgeClusteringOptions{CycleSize: 3, Normalize: true})
	if err != nil || len(clustering) != 4 || !(clustering[0] == clustering[2] || math.IsNaN(clustering[0]) && math.IsNaN(clustering[2])) {
		t.Fatalf("edge clustering alignment = %#v, %v", clustering, err)
	}

	if err := graph.Close(); err != nil {
		t.Fatal(err)
	}
	assertIntegrationFloats(t, vertexSubset, []float64{0.5, 0, 0.5, 0})
	assertIntegrationFloats(t, edgeSubset, []float64{0.5, 0.5, 0.5, 0})
	if constraint == nil || barrat == nil || clustering == nil || convergence.Convergence == nil {
		t.Fatal("Go-owned results became nil after graph closure")
	}
}

func TestMilestone27DegenerateAndCloseRace(t *testing.T) {
	empty, err := igraph.NewGraph()
	if err != nil {
		t.Fatal(err)
	}
	vertexSubset, err := empty.VertexBetweennessSubset(igraph.AllVertices(), igraph.AllVertices(), igraph.AllVertices(), igraph.SubsetBetweennessOptions{})
	if err != nil || vertexSubset == nil || len(vertexSubset) != 0 {
		t.Fatalf("empty subset = %#v, %v", vertexSubset, err)
	}
	barrat, err := empty.BarratTransitivity(igraph.AllVertices(), []float64{}, igraph.TransitivityNaN)
	if err != nil || barrat == nil || len(barrat) != 0 {
		t.Fatalf("empty Barrat = %#v, %v", barrat, err)
	}
	_ = empty.Close()

	graph, weights := milestone27Graph(t)
	var group sync.WaitGroup
	for operation := 0; operation < 10; operation++ {
		operation := operation
		group.Add(1)
		go func() {
			defer group.Done()
			var err error
			switch operation % 5 {
			case 0:
				_, err = graph.VertexBetweennessSubset(igraph.AllVertices(), igraph.AllVertices(), igraph.AllVertices(), igraph.SubsetBetweennessOptions{})
			case 1:
				_, err = graph.BurtConstraint(igraph.AllVertices(), weights)
			case 2:
				_, err = graph.EdgeConvergenceDegree()
			case 3:
				_, err = graph.BarratTransitivity(igraph.AllVertices(), weights, igraph.TransitivityNaN)
			default:
				_, err = graph.EdgeClustering(igraph.AllEdges(), igraph.EdgeClusteringOptions{CycleSize: 3})
			}
			if err != nil && !errors.Is(err, igraph.ErrClosed) {
				t.Errorf("close race operation %d: %v", operation, err)
			}
		}()
	}
	group.Add(1)
	go func() {
		defer group.Done()
		_ = graph.Close()
	}()
	group.Wait()
}

func assertIntegrationFloats(t *testing.T, got, want []float64) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("length = %d, want %d", len(got), len(want))
	}
	for index := range want {
		if math.Abs(got[index]-want[index]) > 1e-12 {
			t.Errorf("value %d = %v, want %v", index, got[index], want[index])
		}
	}
}
