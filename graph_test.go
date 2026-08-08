package igraph

import (
	"bufio"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestWriteEdgeList(t *testing.T) {
	g := testLattice(t, false)
	file := createTempOutput(t, "graph.edgelist")

	if err := g.WriteEdgeList(file); err != nil {
		t.Fatalf("WriteEdgeList() error = %v", err)
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		t.Fatalf("Seek() error = %v", err)
	}

	edges := make(map[string]bool)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var from, to int
		if _, err := fmt.Sscanf(scanner.Text(), "%d %d", &from, &to); err != nil {
			t.Fatalf("invalid edge %q: %v", scanner.Text(), err)
		}
		if from < 0 || from >= 4 || to < 0 || to >= 4 {
			t.Fatalf("edge outside expected vertex range: %d %d", from, to)
		}
		edges[fmt.Sprintf("%d-%d", from, to)] = true
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("Scan() error = %v", err)
	}
	if len(edges) != 4 {
		t.Fatalf("edge count = %d, want 4", len(edges))
	}
}

func TestWriteGraphML(t *testing.T) {
	for _, directed := range []bool{false, true} {
		t.Run(fmt.Sprintf("directed=%t", directed), func(t *testing.T) {
			g := testLattice(t, directed)
			file := createTempOutput(t, "graph.graphml")

			if err := g.WriteGraphML(file, false); err != nil {
				t.Fatalf("WriteGraphML() error = %v", err)
			}
			if _, err := file.Seek(0, io.SeekStart); err != nil {
				t.Fatalf("Seek() error = %v", err)
			}

			wantDefault := "undirected"
			if directed {
				wantDefault = "directed"
			}
			assertGraphML(t, file, wantDefault, 4, 4)
		})
	}
}

func TestGraphWritersRejectInvalidFiles(t *testing.T) {
	g := testLattice(t, false)

	if err := g.WriteEdgeList(nil); err == nil {
		t.Error("WriteEdgeList(nil) error = nil")
	}
	if err := g.WriteGraphML(nil, false); err == nil {
		t.Error("WriteGraphML(nil) error = nil")
	}

	closed := createTempOutput(t, "closed")
	if err := closed.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	if err := g.WriteEdgeList(closed); err == nil {
		t.Error("WriteEdgeList(closed) error = nil")
	}
}

func TestOpenFileStreamRejectsReadOnlyDescriptor(t *testing.T) {
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = reader.Close()
		_ = writer.Close()
	})
	if stream, err := openFileStream(reader); stream != nil || err == nil {
		t.Errorf("openFileStream(read-only pipe) = %v, %v, want nil, error", stream, err)
	}
}

func TestGraphClose(t *testing.T) {
	g := testLattice(t, false)
	if err := g.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	if err := g.Close(); err != nil {
		t.Fatalf("second Close() error = %v", err)
	}
	if err := g.WriteEdgeList(createTempOutput(t, "closed.edgelist")); !errors.Is(err, ErrClosed) {
		t.Errorf("WriteEdgeList() after Close error = %v, want %v", err, ErrClosed)
	}
	if err := g.WriteGraphML(createTempOutput(t, "closed.graphml"), false); !errors.Is(err, ErrClosed) {
		t.Errorf("WriteGraphML() after Close error = %v, want %v", err, ErrClosed)
	}
}

func TestNilGraph(t *testing.T) {
	var g *Graph
	if err := g.Close(); err != nil {
		t.Fatalf("nil Close() error = %v", err)
	}
	if err := g.WriteEdgeList(createTempOutput(t, "nil.edgelist")); !errors.Is(err, ErrClosed) {
		t.Errorf("nil WriteEdgeList() error = %v, want %v", err, ErrClosed)
	}
	if err := g.WriteGraphML(createTempOutput(t, "nil.graphml"), false); !errors.Is(err, ErrClosed) {
		t.Errorf("nil WriteGraphML() error = %v, want %v", err, ErrClosed)
	}
}

func TestGraphFinalizerFallback(t *testing.T) {
	g := testLattice(t, false)
	g.finalize()
	if err := g.WriteEdgeList(createTempOutput(t, "finalized.edgelist")); !errors.Is(err, ErrClosed) {
		t.Errorf("WriteEdgeList() after finalize error = %v, want %v", err, ErrClosed)
	}
}

func TestGraphConstructorsRejectInvalidInput(t *testing.T) {
	tests := []struct {
		name       string
		dimensions []int
		neighbors  int
	}{
		{name: "empty dimensions", dimensions: nil, neighbors: 1},
		{name: "negative dimension", dimensions: []int{2, -1}, neighbors: 1},
		{name: "negative neighbors", dimensions: []int{2, 2}, neighbors: -1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := NewLattice(tt.dimensions, tt.neighbors, false, false, false); err == nil {
				t.Error("NewLattice() error = nil")
			}
		})
	}
}

func TestNewGraph(t *testing.T) {
	g, err := NewGraph()
	if err != nil {
		t.Fatalf("NewGraph() error = %v", err)
	}
	t.Cleanup(func() { _ = g.Close() })
	file := createTempOutput(t, "empty.edgelist")
	if err := g.WriteEdgeList(file); err != nil {
		t.Fatalf("WriteEdgeList() error = %v", err)
	}
}

func TestNewGraphFromEdges(t *testing.T) {
	edges := []Edge{{From: 2, To: 0}, {From: 0, To: 1}, {From: 0, To: 1}}
	for _, directed := range []bool{false, true} {
		t.Run(fmt.Sprintf("directed=%t", directed), func(t *testing.T) {
			g, err := NewGraphFromEdges(4, edges, directed)
			if err != nil {
				t.Fatalf("NewGraphFromEdges() error = %v", err)
			}
			t.Cleanup(func() { _ = g.Close() })

			if got, _ := g.VertexCount(); got != 4 {
				t.Errorf("VertexCount() = %d, want 4", got)
			}
			if got, _ := g.EdgeCount(); got != len(edges) {
				t.Errorf("EdgeCount() = %d, want %d", got, len(edges))
			}
			if got, _ := g.IsDirected(); got != directed {
				t.Errorf("IsDirected() = %t, want %t", got, directed)
			}
			from, to, err := g.EdgeEndpoints(0)
			matches := from == 2 && to == 0
			if !directed {
				matches = matches || from == 0 && to == 2
			}
			if err != nil || !matches {
				t.Errorf("EdgeEndpoints(0) = (%d, %d), %v, want edge between 2 and 0", from, to, err)
			}
		})
	}
}

func TestNewGraphFromEdgesEmptyAndIsolatedVertices(t *testing.T) {
	for _, vertexCount := range []int{0, 3} {
		g, err := NewGraphFromEdges(vertexCount, nil, true)
		if err != nil {
			t.Fatalf("NewGraphFromEdges(%d, nil) error = %v", vertexCount, err)
		}
		if got, _ := g.VertexCount(); got != vertexCount {
			t.Errorf("VertexCount() = %d, want %d", got, vertexCount)
		}
		if got, _ := g.EdgeCount(); got != 0 {
			t.Errorf("EdgeCount() = %d, want 0", got)
		}
		if err := g.Close(); err != nil {
			t.Fatalf("Close() error = %v", err)
		}
	}
}

func TestNewGraphFromEdgesRejectsInvalidInput(t *testing.T) {
	tests := []struct {
		name        string
		vertexCount int
		edges       []Edge
	}{
		{name: "negative vertex count", vertexCount: -1},
		{name: "negative source", vertexCount: 2, edges: []Edge{{From: -1, To: 0}}},
		{name: "source too large", vertexCount: 2, edges: []Edge{{From: 2, To: 0}}},
		{name: "negative target", vertexCount: 2, edges: []Edge{{From: 0, To: -1}}},
		{name: "target too large", vertexCount: 2, edges: []Edge{{From: 0, To: 2}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if g, err := NewGraphFromEdges(tt.vertexCount, tt.edges, false); err == nil {
				_ = g.Close()
				t.Error("NewGraphFromEdges() error = nil")
			}
		})
	}
}

func TestGraphCloneHasIndependentOwnership(t *testing.T) {
	original, err := NewGraphFromEdges(3, []Edge{{From: 0, To: 1}}, true)
	if err != nil {
		t.Fatalf("NewGraphFromEdges() error = %v", err)
	}
	t.Cleanup(func() { _ = original.Close() })
	clone, err := original.Clone()
	if err != nil {
		t.Fatalf("Clone() error = %v", err)
	}
	t.Cleanup(func() { _ = clone.Close() })

	if err := clone.AddEdge(1, 2); err != nil {
		t.Fatalf("clone AddEdge() error = %v", err)
	}
	if got, _ := original.EdgeCount(); got != 1 {
		t.Errorf("original EdgeCount() = %d, want 1", got)
	}
	if got, _ := clone.EdgeCount(); got != 2 {
		t.Errorf("clone EdgeCount() = %d, want 2", got)
	}
	if err := original.Close(); err != nil {
		t.Fatalf("original Close() error = %v", err)
	}
	if got, err := clone.VertexCount(); err != nil || got != 3 {
		t.Errorf("clone VertexCount() after original close = %d, %v, want 3, nil", got, err)
	}
	if err := clone.Close(); err != nil {
		t.Fatalf("clone Close() error = %v", err)
	}
}

func TestGraphCloneRejectsClosedAndNilGraphs(t *testing.T) {
	g, err := NewGraph()
	if err != nil {
		t.Fatalf("NewGraph() error = %v", err)
	}
	if err := g.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	if clone, err := g.Clone(); clone != nil || !errors.Is(err, ErrClosed) {
		t.Errorf("closed Clone() = %v, %v, want nil, %v", clone, err, ErrClosed)
	}

	var nilGraph *Graph
	if clone, err := nilGraph.Clone(); clone != nil || !errors.Is(err, ErrClosed) {
		t.Errorf("nil Clone() = %v, %v, want nil, %v", clone, err, ErrClosed)
	}
}

func TestGraphInspection(t *testing.T) {
	tests := []struct {
		name     string
		directed bool
	}{
		{name: "undirected", directed: false},
		{name: "directed", directed: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := testLattice(t, tt.directed)

			if got, err := g.VertexCount(); err != nil || got != 4 {
				t.Errorf("VertexCount() = %d, %v, want 4, nil", got, err)
			}
			if got, err := g.EdgeCount(); err != nil || got != 4 {
				t.Errorf("EdgeCount() = %d, %v, want 4, nil", got, err)
			}
			if got, err := g.IsDirected(); err != nil || got != tt.directed {
				t.Errorf("IsDirected() = %t, %v, want %t, nil", got, err, tt.directed)
			}
			if got, err := g.IsEmpty(); err != nil || got {
				t.Errorf("IsEmpty() = %t, %v, want false, nil", got, err)
			}

			from, to, err := g.EdgeEndpoints(0)
			if err != nil {
				t.Fatalf("EdgeEndpoints(0) error = %v", err)
			}
			if from < 0 || from >= 4 || to < 0 || to >= 4 || from == to {
				t.Errorf("EdgeEndpoints(0) = (%d, %d), want distinct valid vertices", from, to)
			}
		})
	}
}

func TestEmptyGraphInspection(t *testing.T) {
	g, err := NewGraph()
	if err != nil {
		t.Fatalf("NewGraph() error = %v", err)
	}
	t.Cleanup(func() { _ = g.Close() })

	if got, err := g.VertexCount(); err != nil || got != 0 {
		t.Errorf("VertexCount() = %d, %v, want 0, nil", got, err)
	}
	if got, err := g.EdgeCount(); err != nil || got != 0 {
		t.Errorf("EdgeCount() = %d, %v, want 0, nil", got, err)
	}
	if got, err := g.IsDirected(); err != nil || got {
		t.Errorf("IsDirected() = %t, %v, want false, nil", got, err)
	}
	if got, err := g.IsEmpty(); err != nil || !got {
		t.Errorf("IsEmpty() = %t, %v, want true, nil", got, err)
	}
	if _, _, err := g.EdgeEndpoints(0); err == nil {
		t.Error("EdgeEndpoints(0) error = nil")
	}
}

func TestEdgelessGraphIsNotEmpty(t *testing.T) {
	g, err := NewLattice([]int{1}, 1, false, false, false)
	if err != nil {
		t.Fatalf("NewLattice() error = %v", err)
	}
	t.Cleanup(func() { _ = g.Close() })

	if got, err := g.EdgeCount(); err != nil || got != 0 {
		t.Errorf("EdgeCount() = %d, %v, want 0, nil", got, err)
	}
	if got, err := g.IsEmpty(); err != nil || got {
		t.Errorf("IsEmpty() = %t, %v, want false, nil", got, err)
	}
}

func TestGraphInspectionRejectsInvalidEdgeIDs(t *testing.T) {
	g := testLattice(t, false)
	for _, edgeID := range []int{-1, 4} {
		if _, _, err := g.EdgeEndpoints(edgeID); err == nil {
			t.Errorf("EdgeEndpoints(%d) error = nil", edgeID)
		}
	}
}

func TestGraphInspectionRejectsClosedGraph(t *testing.T) {
	g := testLattice(t, false)
	if err := g.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	if _, err := g.VertexCount(); !errors.Is(err, ErrClosed) {
		t.Errorf("VertexCount() error = %v, want %v", err, ErrClosed)
	}
	if _, err := g.EdgeCount(); !errors.Is(err, ErrClosed) {
		t.Errorf("EdgeCount() error = %v, want %v", err, ErrClosed)
	}
	if _, err := g.IsDirected(); !errors.Is(err, ErrClosed) {
		t.Errorf("IsDirected() error = %v, want %v", err, ErrClosed)
	}
	if _, err := g.IsEmpty(); !errors.Is(err, ErrClosed) {
		t.Errorf("IsEmpty() error = %v, want %v", err, ErrClosed)
	}
	if _, _, err := g.EdgeEndpoints(0); !errors.Is(err, ErrClosed) {
		t.Errorf("EdgeEndpoints(0) error = %v, want %v", err, ErrClosed)
	}
}

func TestNilGraphInspection(t *testing.T) {
	var g *Graph
	if _, err := g.VertexCount(); !errors.Is(err, ErrClosed) {
		t.Errorf("VertexCount() error = %v, want %v", err, ErrClosed)
	}
	if _, err := g.EdgeCount(); !errors.Is(err, ErrClosed) {
		t.Errorf("EdgeCount() error = %v, want %v", err, ErrClosed)
	}
	if _, err := g.IsDirected(); !errors.Is(err, ErrClosed) {
		t.Errorf("IsDirected() error = %v, want %v", err, ErrClosed)
	}
	if _, err := g.IsEmpty(); !errors.Is(err, ErrClosed) {
		t.Errorf("IsEmpty() error = %v, want %v", err, ErrClosed)
	}
	if _, _, err := g.EdgeEndpoints(0); !errors.Is(err, ErrClosed) {
		t.Errorf("EdgeEndpoints(0) error = %v, want %v", err, ErrClosed)
	}
}

func TestAddVertices(t *testing.T) {
	for _, directed := range []bool{false, true} {
		t.Run(fmt.Sprintf("directed=%t", directed), func(t *testing.T) {
			g := testLattice(t, directed)
			if err := g.AddVertices(2); err != nil {
				t.Fatalf("AddVertices(2) error = %v", err)
			}
			if got, err := g.VertexCount(); err != nil || got != 6 {
				t.Errorf("VertexCount() = %d, %v, want 6, nil", got, err)
			}
			if err := g.AddVertices(0); err != nil {
				t.Fatalf("AddVertices(0) error = %v", err)
			}
			if got, _ := g.VertexCount(); got != 6 {
				t.Errorf("VertexCount() after no-op = %d, want 6", got)
			}
		})
	}
}

func TestAddEdge(t *testing.T) {
	for _, directed := range []bool{false, true} {
		t.Run(fmt.Sprintf("directed=%t", directed), func(t *testing.T) {
			g := testLattice(t, directed)
			before, _ := g.EdgeCount()
			if err := g.AddEdge(3, 0); err != nil {
				t.Fatalf("AddEdge() error = %v", err)
			}
			if got, err := g.EdgeCount(); err != nil || got != before+1 {
				t.Errorf("EdgeCount() = %d, %v, want %d, nil", got, err, before+1)
			}
			from, to, err := g.EdgeEndpoints(before)
			endpointsMatch := from == 3 && to == 0
			if !directed {
				endpointsMatch = endpointsMatch || from == 0 && to == 3
			}
			if err != nil || !endpointsMatch {
				t.Errorf("EdgeEndpoints(%d) = (%d, %d), %v, want edge between 3 and 0", before, from, to, err)
			}
		})
	}
}

func TestAddEdgesAllowsLoopsAndParallelEdges(t *testing.T) {
	for _, directed := range []bool{false, true} {
		t.Run(fmt.Sprintf("directed=%t", directed), func(t *testing.T) {
			g := testLattice(t, directed)
			before, _ := g.EdgeCount()
			edges := []Edge{{From: 0, To: 0}, {From: 0, To: 1}, {From: 0, To: 1}}
			if err := g.AddEdges(edges); err != nil {
				t.Fatalf("AddEdges() error = %v", err)
			}
			if got, err := g.EdgeCount(); err != nil || got != before+len(edges) {
				t.Errorf("EdgeCount() = %d, %v, want %d, nil", got, err, before+len(edges))
			}
			for index, want := range edges {
				from, to, err := g.EdgeEndpoints(before + index)
				if err != nil || from != want.From || to != want.To {
					t.Errorf("EdgeEndpoints(%d) = (%d, %d), %v, want (%d, %d), nil", before+index, from, to, err, want.From, want.To)
				}
			}
		})
	}
}

func TestAddEdgesEmptyBatchIsNoOp(t *testing.T) {
	g := testLattice(t, false)
	before, _ := g.EdgeCount()
	if err := g.AddEdges(nil); err != nil {
		t.Fatalf("AddEdges(nil) error = %v", err)
	}
	if got, _ := g.EdgeCount(); got != before {
		t.Errorf("EdgeCount() = %d, want %d", got, before)
	}
}

func TestGraphMutationsRejectInvalidInputAtomically(t *testing.T) {
	g := testLattice(t, false)
	beforeVertices, _ := g.VertexCount()
	beforeEdges, _ := g.EdgeCount()

	if err := g.AddVertices(-1); err == nil {
		t.Error("AddVertices(-1) error = nil")
	}
	if got, _ := g.VertexCount(); got != beforeVertices {
		t.Errorf("VertexCount() after invalid addition = %d, want %d", got, beforeVertices)
	}
	for _, edge := range []Edge{{From: -1, To: 0}, {From: 0, To: beforeVertices}} {
		if err := g.AddEdge(edge.From, edge.To); err == nil {
			t.Errorf("AddEdge(%d, %d) error = nil", edge.From, edge.To)
		}
	}
	if err := g.AddEdges([]Edge{{From: 0, To: 1}, {From: 1, To: beforeVertices}}); err == nil {
		t.Error("AddEdges() with invalid endpoint error = nil")
	}
	if got, _ := g.EdgeCount(); got != beforeEdges {
		t.Errorf("EdgeCount() after invalid additions = %d, want %d", got, beforeEdges)
	}
}

func TestGraphMutationsRejectClosedAndNilGraphs(t *testing.T) {
	g := testLattice(t, false)
	if err := g.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	if err := g.AddVertices(1); !errors.Is(err, ErrClosed) {
		t.Errorf("AddVertices() error = %v, want %v", err, ErrClosed)
	}
	if err := g.AddEdge(0, 1); !errors.Is(err, ErrClosed) {
		t.Errorf("AddEdge() error = %v, want %v", err, ErrClosed)
	}
	if err := g.AddEdges([]Edge{{From: 0, To: 1}}); !errors.Is(err, ErrClosed) {
		t.Errorf("AddEdges() error = %v, want %v", err, ErrClosed)
	}

	var nilGraph *Graph
	if err := nilGraph.AddVertices(1); !errors.Is(err, ErrClosed) {
		t.Errorf("nil AddVertices() error = %v, want %v", err, ErrClosed)
	}
	if err := nilGraph.AddEdge(0, 1); !errors.Is(err, ErrClosed) {
		t.Errorf("nil AddEdge() error = %v, want %v", err, ErrClosed)
	}
	if err := nilGraph.AddEdges(nil); !errors.Is(err, ErrClosed) {
		t.Errorf("nil AddEdges() error = %v, want %v", err, ErrClosed)
	}
}

func testLattice(t *testing.T, directed bool) *Graph {
	t.Helper()
	g, err := NewLattice([]int{2, 2}, 1, directed, false, false)
	if err != nil {
		t.Fatalf("NewLattice() error = %v", err)
	}
	t.Cleanup(func() { _ = g.Close() })
	return g
}

func createTempOutput(t *testing.T, name string) *os.File {
	t.Helper()
	file, err := os.Create(filepath.Join(t.TempDir(), name))
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	t.Cleanup(func() { _ = file.Close() })
	return file
}

func assertGraphML(t *testing.T, reader io.Reader, wantDefault string, wantNodes, wantEdges int) {
	t.Helper()
	decoder := xml.NewDecoder(reader)
	var gotDefault string
	var nodes, edges int
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("invalid GraphML: %v", err)
		}
		start, ok := token.(xml.StartElement)
		if !ok {
			continue
		}
		switch start.Name.Local {
		case "graph":
			for _, attr := range start.Attr {
				if attr.Name.Local == "edgedefault" {
					gotDefault = attr.Value
				}
			}
		case "node":
			nodes++
		case "edge":
			edges++
		}
	}
	if gotDefault != wantDefault {
		t.Errorf("edgedefault = %q, want %q", gotDefault, wantDefault)
	}
	if nodes != wantNodes {
		t.Errorf("node count = %d, want %d", nodes, wantNodes)
	}
	if edges != wantEdges {
		t.Errorf("edge count = %d, want %d", edges, wantEdges)
	}
}
