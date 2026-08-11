package igraph_test

import (
	"errors"
	"sync"
	"testing"

	igraph "github.com/h8gi/go-igraph"
)

func TestMilestone12DirectedCycleAnalysisPipeline(t *testing.T) {
	edges := []igraph.Edge{
		{From: 0, To: 1}, {From: 1, To: 2}, {From: 2, To: 0},
		{From: 2, To: 3}, {From: 3, To: 4}, {From: 4, To: 2},
		{From: 4, To: 5},
	}
	graph := newMilestone12Graph(t, 6, edges, true)

	acyclic, err := graph.IsAcyclic()
	if err != nil || acyclic {
		t.Fatalf("IsAcyclic = %t, %v", acyclic, err)
	}
	dag, err := graph.IsDAG()
	if err != nil || dag {
		t.Fatalf("IsDAG = %t, %v", dag, err)
	}
	witness, err := graph.FindCycle(igraph.DirectionOut)
	if err != nil {
		t.Fatal(err)
	}
	assertMilestone12Cycle(t, graph, witness, true)
	girth, err := graph.Girth()
	if err != nil || girth.Length != 3 || len(girth.Vertices) != 3 {
		t.Fatalf("Girth = %#v, %v", girth, err)
	}

	bounded, err := graph.SimpleCycles(igraph.SimpleCycleOptions{
		Direction:  igraph.DirectionOut,
		MaxResults: 1,
	})
	if err != nil || len(bounded.Cycles) != 1 || !bounded.Truncated {
		t.Fatalf("bounded SimpleCycles = %#v, %v", bounded, err)
	}
	assertMilestone12Cycle(t, graph, bounded.Cycles[0], true)
	completeCycles, err := graph.SimpleCycles(igraph.SimpleCycleOptions{
		Direction:  igraph.DirectionOut,
		MaxResults: 2,
	})
	if err != nil || len(completeCycles.Cycles) != 2 || completeCycles.Truncated {
		t.Fatalf("complete SimpleCycles = %#v, %v", completeCycles, err)
	}
	for _, cycle := range completeCycles.Cycles {
		assertMilestone12Cycle(t, graph, cycle, true)
	}

	fundamental, err := graph.FundamentalCycleBasis(igraph.FundamentalCycleBasisOptions{})
	if err != nil {
		t.Fatal(err)
	}
	minimum, err := graph.MinimumCycleBasis(igraph.MinimumCycleBasisOptions{NaturalOrder: true})
	if err != nil {
		t.Fatal(err)
	}
	assertMilestone12Basis(t, fundamental, len(edges), 2)
	assertMilestone12Basis(t, minimum, len(edges), 2)

	feedbackEdges, err := graph.FeedbackEdgeSet(igraph.FeedbackEdgeOptions{})
	if err != nil {
		t.Fatal(err)
	}
	edgeReduced := cloneMilestone12Graph(t, graph)
	deleteMilestone12Edges(t, edgeReduced, feedbackEdges)
	assertMilestone12DAG(t, edgeReduced)

	feedbackVertices, err := graph.FeedbackVertexSet(nil)
	if err != nil {
		t.Fatal(err)
	}
	vertexReduced := cloneMilestone12Graph(t, graph)
	deleteMilestone12Vertices(t, vertexReduced, feedbackVertices)
	assertMilestone12DAG(t, vertexReduced)

	if err := graph.Close(); err != nil {
		t.Fatal(err)
	}
	if len(witness.Vertices) == 0 || len(completeCycles.Cycles) != 2 || len(fundamental) != 2 || len(minimum) != 2 || feedbackEdges == nil || feedbackVertices == nil {
		t.Fatal("Go-owned cycle-analysis results did not survive source closure")
	}
	before := completeCycles.Cycles[1].Vertices[0]
	completeCycles.Cycles[0].Vertices[0] = 99
	if completeCycles.Cycles[1].Vertices[0] != before {
		t.Fatal("simple-cycle results share mutable backing storage")
	}
}

func TestMilestone12UndirectedCycleAnalysisPipeline(t *testing.T) {
	edges := []igraph.Edge{
		{From: 0, To: 1}, {From: 1, To: 2}, {From: 2, To: 0},
		{From: 2, To: 3}, {From: 3, To: 4}, {From: 4, To: 2},
		{From: 4, To: 5},
	}
	graph := newMilestone12Graph(t, 6, edges, false)

	acyclic, err := graph.IsAcyclic()
	if err != nil || acyclic {
		t.Fatalf("IsAcyclic = %t, %v", acyclic, err)
	}
	dag, err := graph.IsDAG()
	if err != nil || dag {
		t.Fatalf("IsDAG = %t, %v", dag, err)
	}
	witness, err := graph.FindCycle(igraph.DirectionAll)
	if err != nil {
		t.Fatal(err)
	}
	assertMilestone12Cycle(t, graph, witness, false)
	bounded, err := graph.SimpleCycles(igraph.SimpleCycleOptions{
		Direction:  igraph.DirectionAll,
		MaxResults: 1,
	})
	if err != nil || len(bounded.Cycles) != 1 || !bounded.Truncated {
		t.Fatalf("bounded SimpleCycles = %#v, %v", bounded, err)
	}
	assertMilestone12Cycle(t, graph, bounded.Cycles[0], false)

	fundamental, err := graph.FundamentalCycleBasis(igraph.FundamentalCycleBasisOptions{})
	if err != nil {
		t.Fatal(err)
	}
	minimum, err := graph.MinimumCycleBasis(igraph.MinimumCycleBasisOptions{})
	if err != nil {
		t.Fatal(err)
	}
	assertMilestone12Basis(t, fundamental, len(edges), 2)
	assertMilestone12Basis(t, minimum, len(edges), 2)

	feedbackEdges, err := graph.FeedbackEdgeSet(igraph.FeedbackEdgeOptions{})
	if err != nil {
		t.Fatal(err)
	}
	edgeReduced := cloneMilestone12Graph(t, graph)
	deleteMilestone12Edges(t, edgeReduced, feedbackEdges)
	assertMilestone12Acyclic(t, edgeReduced)

	feedbackVertices, err := graph.FeedbackVertexSet(nil)
	if err != nil {
		t.Fatal(err)
	}
	vertexReduced := cloneMilestone12Graph(t, graph)
	deleteMilestone12Vertices(t, vertexReduced, feedbackVertices)
	assertMilestone12Acyclic(t, vertexReduced)
}

func TestMilestone12ConcurrentReadsAndClose(t *testing.T) {
	edges := []igraph.Edge{
		{From: 0, To: 1}, {From: 1, To: 2}, {From: 2, To: 3}, {From: 3, To: 4},
	}
	liveGraph := newMilestone12Graph(t, 5, edges, true)
	for err := range runMilestone12Calls(milestone12ReadCalls(liveGraph), nil) {
		if err != nil {
			t.Errorf("live concurrent cycle-analysis call error = %v", err)
		}
	}

	closingGraph := newMilestone12Graph(t, 5, edges, true)
	errorsByCall := runMilestone12Calls(milestone12ReadCalls(closingGraph), func() {
		if err := closingGraph.Close(); err != nil {
			t.Error(err)
		}
	})
	for err := range errorsByCall {
		if err != nil && !errors.Is(err, igraph.ErrClosed) {
			t.Errorf("Close-race cycle-analysis call error = %v", err)
		}
	}
	if _, err := closingGraph.FindCycle(igraph.DirectionOut); !errors.Is(err, igraph.ErrClosed) {
		t.Errorf("post-close FindCycle error = %v", err)
	}
}

func milestone12ReadCalls(graph *igraph.Graph) []func() error {
	return []func() error{
		func() error { _, err := graph.IsAcyclic(); return err },
		func() error { _, err := graph.IsDAG(); return err },
		func() error { _, err := graph.TopologicalSort(igraph.DirectionOut); return err },
		func() error { _, err := graph.FindCycle(igraph.DirectionOut); return err },
		func() error { _, err := graph.Girth(); return err },
		func() error {
			_, err := graph.SimpleCycles(igraph.SimpleCycleOptions{Direction: igraph.DirectionOut, MaxResults: 1})
			return err
		},
		func() error { _, err := graph.FundamentalCycleBasis(igraph.FundamentalCycleBasisOptions{}); return err },
		func() error { _, err := graph.MinimumCycleBasis(igraph.MinimumCycleBasisOptions{}); return err },
		func() error { _, err := graph.FeedbackEdgeSet(igraph.FeedbackEdgeOptions{}); return err },
		func() error { _, err := graph.FeedbackVertexSet(nil); return err },
	}
}

func runMilestone12Calls(calls []func() error, afterStart func()) <-chan error {
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

func newMilestone12Graph(t *testing.T, vertices int, edges []igraph.Edge, directed bool) *igraph.Graph {
	t.Helper()
	graph, err := igraph.NewGraphFromEdges(vertices, edges, directed)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = graph.Close() })
	return graph
}

func cloneMilestone12Graph(t *testing.T, graph *igraph.Graph) *igraph.Graph {
	t.Helper()
	clone, err := graph.Clone()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = clone.Close() })
	return clone
}

func deleteMilestone12Edges(t *testing.T, graph *igraph.Graph, ids []int) {
	t.Helper()
	selector, err := igraph.EdgeIDs(ids...)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := graph.DeleteEdges(selector); err != nil {
		t.Fatal(err)
	}
}

func deleteMilestone12Vertices(t *testing.T, graph *igraph.Graph, ids []int) {
	t.Helper()
	selector, err := igraph.VertexIDs(ids...)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := graph.DeleteVertices(selector); err != nil {
		t.Fatal(err)
	}
}

func assertMilestone12Acyclic(t *testing.T, graph *igraph.Graph) {
	t.Helper()
	acyclic, err := graph.IsAcyclic()
	if err != nil || !acyclic {
		t.Fatalf("reduced IsAcyclic = %t, %v", acyclic, err)
	}
}

func assertMilestone12DAG(t *testing.T, graph *igraph.Graph) {
	t.Helper()
	assertMilestone12Acyclic(t, graph)
	dag, err := graph.IsDAG()
	if err != nil || !dag {
		t.Fatalf("reduced IsDAG = %t, %v", dag, err)
	}
	order, err := graph.TopologicalSort(igraph.DirectionOut)
	if err != nil {
		t.Fatal(err)
	}
	vertices, err := graph.VertexCount()
	if err != nil || len(order) != vertices {
		t.Fatalf("topological order length = %d, vertices = %d, err = %v", len(order), vertices, err)
	}
	position := make([]int, vertices)
	seen := make([]bool, vertices)
	for index, vertex := range order {
		if vertex < 0 || vertex >= vertices {
			t.Fatalf("topological vertex %d out of range", vertex)
		}
		if seen[vertex] {
			t.Fatalf("topological vertex %d appears more than once", vertex)
		}
		seen[vertex] = true
		position[vertex] = index
	}
	edges, err := graph.EdgeCount()
	if err != nil {
		t.Fatal(err)
	}
	for edge := 0; edge < edges; edge++ {
		from, to, err := graph.EdgeEndpoints(edge)
		if err != nil {
			t.Fatal(err)
		}
		if position[from] >= position[to] {
			t.Fatalf("edge %d (%d -> %d) violates topological order %v", edge, from, to, order)
		}
	}
}

func assertMilestone12Cycle(t *testing.T, graph *igraph.Graph, cycle igraph.Cycle, directed bool) {
	t.Helper()
	if len(cycle.Vertices) == 0 || len(cycle.Vertices) != len(cycle.Edges) {
		t.Fatalf("invalid cycle shape: %#v", cycle)
	}
	for index, edge := range cycle.Edges {
		from, to, err := graph.EdgeEndpoints(edge)
		if err != nil {
			t.Fatal(err)
		}
		vertex := cycle.Vertices[index]
		next := cycle.Vertices[(index+1)%len(cycle.Vertices)]
		aligned := from == vertex && to == next
		if !directed {
			aligned = aligned || from == next && to == vertex
		}
		if !aligned {
			t.Fatalf("cycle edge %d (%d, %d) does not align with %d -> %d", edge, from, to, vertex, next)
		}
	}
}

func assertMilestone12Basis(t *testing.T, basis [][]int, edgeCount, rank int) {
	t.Helper()
	if len(basis) != rank {
		t.Fatalf("cycle basis size = %d, want rank %d: %v", len(basis), rank, basis)
	}
	rows := make([]uint64, len(basis))
	for row, cycle := range basis {
		if len(cycle) == 0 {
			t.Fatal("cycle basis contains an empty element")
		}
		for _, edge := range cycle {
			if edge < 0 || edge >= edgeCount {
				t.Fatalf("cycle-basis edge %d out of range [0, %d)", edge, edgeCount)
			}
			rows[row] ^= uint64(1) << edge
		}
	}
	pivots := make(map[int]uint64)
	independent := 0
	for _, row := range rows {
		for bit := edgeCount - 1; bit >= 0 && row != 0; bit-- {
			if row&(uint64(1)<<bit) == 0 {
				continue
			}
			if pivot, ok := pivots[bit]; ok {
				row ^= pivot
				continue
			}
			pivots[bit] = row
			independent++
			break
		}
	}
	if independent != rank {
		t.Fatalf("cycle basis GF(2) rank = %d, want %d: %v", independent, rank, basis)
	}
}
