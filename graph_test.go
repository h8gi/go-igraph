package igraph

import (
	"bufio"
	"encoding/xml"
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

func testLattice(t *testing.T, directed bool) *Graph {
	t.Helper()
	dimensions := NewVectorFromSlice([]float64{2, 2})
	return NewLattice(dimensions, 1, directed, false, false)
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
