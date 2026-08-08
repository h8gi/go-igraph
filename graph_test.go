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
