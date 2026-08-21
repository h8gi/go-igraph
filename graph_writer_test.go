package igraph

import (
	"errors"
	"math"
	"os"
	"reflect"
	"sort"
	"sync"
	"testing"
)

func normalizeEdges(edges []Edge, directed bool) []Edge {
	res := make([]Edge, len(edges))
	for i, e := range edges {
		if !directed && e.From > e.To {
			res[i] = Edge{From: e.To, To: e.From}
		} else {
			res[i] = e
		}
	}
	return res
}

func TestWriteEdgeListRoundTrip(t *testing.T) {
	for _, directed := range []bool{false, true} {
		edges := []Edge{{0, 1}, {1, 2}, {2, 0}, {0, 0}}
		g, err := NewGraphFromEdges(4, edges, directed)
		if err != nil {
			t.Fatal(err)
		}
		defer g.Close()

		file := graphWriterTempFile(t)
		if err := g.WriteEdgeList(file); err != nil {
			t.Fatal(err)
		}

		if _, err := file.Seek(0, 0); err != nil {
			t.Fatal(err)
		}
		readBack, err := ReadEdgeList(file, EdgeListReadOptions{VertexCount: 4, Directed: directed})
		if err != nil {
			t.Fatal(err)
		}
		defer readBack.Close()

		assertGraphShape(t, readBack, 4, len(edges), directed)
		readEdges, err := readBack.Edges()
		if err != nil {
			t.Fatal(err)
		}

		normRead := normalizeEdges(readEdges, directed)
		normOrig := normalizeEdges(edges, directed)
		sort.Slice(normRead, func(i, j int) bool {
			if normRead[i].From == normRead[j].From {
				return normRead[i].To < normRead[j].To
			}
			return normRead[i].From < normRead[j].From
		})
		sort.Slice(normOrig, func(i, j int) bool {
			if normOrig[i].From == normOrig[j].From {
				return normOrig[i].To < normOrig[j].To
			}
			return normOrig[i].From < normOrig[j].From
		})
		if !reflect.DeepEqual(normRead, normOrig) {
			t.Fatalf("edges (directed=%v) = %v, want %v", directed, normRead, normOrig)
		}
	}
}

func TestWriteEdgeListEmptyGraph(t *testing.T) {
	g, err := NewGraph()
	if err != nil {
		t.Fatal(err)
	}
	defer g.Close()

	file := graphWriterTempFile(t)
	if err := g.WriteEdgeList(file); err != nil {
		t.Fatal(err)
	}

	if _, err := file.Seek(0, 0); err != nil {
		t.Fatal(err)
	}
	readBack, err := ReadEdgeList(file, EdgeListReadOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer readBack.Close()
	assertGraphShape(t, readBack, 0, 0, false)
}

func TestWriteGraphMLRoundTripTypedAttributes(t *testing.T) {
	for _, directed := range []bool{true, false} {
		edges := []Edge{{0, 1}, {1, 2}}
		g, err := NewGraphFromEdges(3, edges, directed)
		if err != nil {
			t.Fatal(err)
		}
		defer g.Close()

		if err := g.SetGraphNumericAttribute("g_num", -3.14); err != nil {
			t.Fatal(err)
		}
		if err := g.SetGraphStringAttribute("g_str", "hello graph"); err != nil {
			t.Fatal(err)
		}
		if err := g.SetGraphBooleanAttribute("g_bool", true); err != nil {
			t.Fatal(err)
		}

		if err := g.SetVertexNumericAttributes("v_num", []float64{0.5, -1.25, 42.0}); err != nil {
			t.Fatal(err)
		}
		if err := g.SetVertexStringAttributes("v_str", []string{"v0", "", "v2"}); err != nil {
			t.Fatal(err)
		}
		if err := g.SetVertexBooleanAttributes("v_bool", []bool{true, false, true}); err != nil {
			t.Fatal(err)
		}

		if err := g.SetEdgeNumericAttributes("e_num", []float64{10.5, -20.25}); err != nil {
			t.Fatal(err)
		}
		if err := g.SetEdgeStringAttributes("e_str", []string{"first", ""}); err != nil {
			t.Fatal(err)
		}
		if err := g.SetEdgeBooleanAttributes("e_bool", []bool{false, true}); err != nil {
			t.Fatal(err)
		}

		file := graphWriterTempFile(t)
		if err := g.WriteGraphML(file, false); err != nil {
			t.Fatal(err)
		}

		if _, err := file.Seek(0, 0); err != nil {
			t.Fatal(err)
		}
		readBack, err := ReadGraphML(file, 0)
		if err != nil {
			t.Fatal(err)
		}
		defer readBack.Close()

		assertGraphShape(t, readBack, 3, 2, directed)
		assertEdgesEqual(t, readBack, edges)

		if val, err := readBack.GraphNumericAttribute("g_num"); err != nil || val != -3.14 {
			t.Fatalf("g_num = %f, %v", val, err)
		}
		if val, err := readBack.GraphStringAttribute("g_str"); err != nil || val != "hello graph" {
			t.Fatalf("g_str = %q, %v", val, err)
		}
		if val, err := readBack.GraphBooleanAttribute("g_bool"); err != nil || val != true {
			t.Fatalf("g_bool = %v, %v", val, err)
		}

		if val, err := readBack.VertexNumericAttributes("v_num"); err != nil || !reflect.DeepEqual(val, []float64{0.5, -1.25, 42.0}) {
			t.Fatalf("v_num = %v, %v", val, err)
		}
		if val, err := readBack.VertexStringAttributes("v_str"); err != nil || !reflect.DeepEqual(val, []string{"v0", "", "v2"}) {
			t.Fatalf("v_str = %v, %v", val, err)
		}
		if val, err := readBack.VertexBooleanAttributes("v_bool"); err != nil || !reflect.DeepEqual(val, []bool{true, false, true}) {
			t.Fatalf("v_bool = %v, %v", val, err)
		}

		if val, err := readBack.EdgeNumericAttributes("e_num"); err != nil || !reflect.DeepEqual(val, []float64{10.5, -20.25}) {
			t.Fatalf("e_num = %v, %v", val, err)
		}
		if val, err := readBack.EdgeStringAttributes("e_str"); err != nil || !reflect.DeepEqual(val, []string{"first", ""}) {
			t.Fatalf("e_str = %v, %v", val, err)
		}
		if val, err := readBack.EdgeBooleanAttributes("e_bool"); err != nil || !reflect.DeepEqual(val, []bool{false, true}) {
			t.Fatalf("e_bool = %v, %v", val, err)
		}
	}
}

func TestWriteGMLRoundTripAndLossyBooleanConversion(t *testing.T) {
	for _, directed := range []bool{true, false} {
		edges := []Edge{{0, 1}, {1, 2}}
		g, err := NewGraphFromEdges(3, edges, directed)
		if err != nil {
			t.Fatal(err)
		}
		defer g.Close()

		if err := g.SetGraphNumericAttribute("gnum", -3.14); err != nil {
			t.Fatal(err)
		}
		if err := g.SetGraphStringAttribute("gstr", "graphlabel"); err != nil {
			t.Fatal(err)
		}
		if err := g.SetGraphBooleanAttribute("gbool", true); err != nil {
			t.Fatal(err)
		}

		if err := g.SetVertexNumericAttributes("vnum", []float64{0.5, -1.25, 42.0}); err != nil {
			t.Fatal(err)
		}
		if err := g.SetVertexStringAttributes("vstr", []string{"A", "", "C"}); err != nil {
			t.Fatal(err)
		}
		if err := g.SetVertexBooleanAttributes("vbool", []bool{true, false, true}); err != nil {
			t.Fatal(err)
		}

		if err := g.SetEdgeNumericAttributes("enum", []float64{0.5, -20.25}); err != nil {
			t.Fatal(err)
		}
		if err := g.SetEdgeStringAttributes("estr", []string{"e1", ""}); err != nil {
			t.Fatal(err)
		}
		if err := g.SetEdgeBooleanAttributes("ebool", []bool{false, true}); err != nil {
			t.Fatal(err)
		}

		file := graphWriterTempFile(t)
		if err := g.WriteGML(file, GMLWriteOptions{Creator: "testsuite"}); err != nil {
			t.Fatal(err)
		}

		if _, err := file.Seek(0, 0); err != nil {
			t.Fatal(err)
		}
		readBack, err := ReadGML(file)
		if err != nil {
			t.Fatal(err)
		}
		defer readBack.Close()

		assertGraphShape(t, readBack, 3, 2, directed)
		assertEdgesEqual(t, readBack, edges)

		if val, err := readBack.GraphNumericAttribute("gnum"); err != nil || val != -3.14 {
			t.Fatalf("gnum = %f, %v", val, err)
		}
		if val, err := readBack.GraphStringAttribute("gstr"); err != nil || val != "graphlabel" {
			t.Fatalf("gstr = %q, %v", val, err)
		}
		// GML converts boolean attributes deterministically to numeric 1.0 / 0.0
		if val, err := readBack.GraphNumericAttribute("gbool"); err != nil || val != 1.0 {
			t.Fatalf("gbool converted numeric = %f, %v", val, err)
		}

		if val, err := readBack.VertexNumericAttributes("vnum"); err != nil || !reflect.DeepEqual(val, []float64{0.5, -1.25, 42.0}) {
			t.Fatalf("vnum = %v, %v", val, err)
		}
		if val, err := readBack.VertexStringAttributes("vstr"); err != nil || !reflect.DeepEqual(val, []string{"A", "", "C"}) {
			t.Fatalf("vstr = %v, %v", val, err)
		}
		if val, err := readBack.VertexNumericAttributes("vbool"); err != nil || !reflect.DeepEqual(val, []float64{1, 0, 1}) {
			t.Fatalf("vbool converted numeric = %v, %v", val, err)
		}

		if val, err := readBack.EdgeNumericAttributes("enum"); err != nil || !reflect.DeepEqual(val, []float64{0.5, -20.25}) {
			t.Fatalf("enum = %v, %v", val, err)
		}
		if val, err := readBack.EdgeStringAttributes("estr"); err != nil || !reflect.DeepEqual(val, []string{"e1", ""}) {
			t.Fatalf("estr = %v, %v", val, err)
		}
		if val, err := readBack.EdgeNumericAttributes("ebool"); err != nil || !reflect.DeepEqual(val, []float64{0, 1}) {
			t.Fatalf("ebool converted numeric = %v, %v", val, err)
		}
	}
}

func TestGraphWritersRoundTripNaN(t *testing.T) {
	graphMLInput := `<?xml version="1.0" encoding="UTF-8"?>
<graphml xmlns="http://graphml.graphdrawing.org/xmlns">
  <key id="score" for="node" attr.name="score" attr.type="double"/>
  <graph id="G" edgedefault="directed">
    <node id="n0"><data key="score">1.25</data></node>
    <node id="n1"/>
    <node id="n2"><data key="score">-2.5</data></node>
  </graph>
</graphml>`
	gmlInput := `Version 1
graph [
  directed 1
  node [ id 0 score 1.25 ]
  node [ id 1 ]
  node [ id 2 score -2.5 ]
]`
	tests := []struct {
		name        string
		input       string
		importGraph func(*os.File) (*Graph, error)
		write       func(*Graph, *os.File) error
		read        func(*os.File) (*Graph, error)
	}{
		{
			name:        "GraphML",
			input:       graphMLInput,
			importGraph: func(file *os.File) (*Graph, error) { return ReadGraphML(file, 0) },
			write:       func(g *Graph, file *os.File) error { return g.WriteGraphML(file, false) },
			read:        func(file *os.File) (*Graph, error) { return ReadGraphML(file, 0) },
		},
		{
			name:        "GML",
			input:       gmlInput,
			importGraph: ReadGML,
			write:       func(g *Graph, file *os.File) error { return g.WriteGML(file, GMLWriteOptions{}) },
			read:        ReadGML,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			source := graphReaderFile(t, test.input)
			g, err := test.importGraph(source)
			if err != nil {
				t.Fatal(err)
			}
			defer g.Close()
			want := []float64{1.25, math.NaN(), -2.5}
			imported, err := g.VertexNumericAttributes("score")
			if err != nil || len(imported) != len(want) || !math.IsNaN(imported[1]) {
				t.Fatalf("imported score = %v, %v", imported, err)
			}

			file := graphWriterTempFile(t)
			if err := test.write(g, file); err != nil {
				t.Fatal(err)
			}
			if _, err := file.Seek(0, 0); err != nil {
				t.Fatal(err)
			}
			readBack, err := test.read(file)
			if err != nil {
				t.Fatal(err)
			}
			defer readBack.Close()

			got, err := readBack.VertexNumericAttributes("score")
			if err != nil {
				t.Fatal(err)
			}
			if len(got) != len(want) || got[0] != want[0] || !math.IsNaN(got[1]) || got[2] != want[2] {
				t.Fatalf("score = %v, want finite values with NaN at index 1", got)
			}
		})
	}
}

func TestWriteGMLAllowsNamesOutsideReservedScope(t *testing.T) {
	g, err := NewGraphFromEdges(2, []Edge{{0, 1}}, true)
	if err != nil {
		t.Fatal(err)
	}
	defer g.Close()

	if err := g.SetGraphNumericAttribute("id", 1); err != nil {
		t.Fatal(err)
	}
	if err := g.SetVertexNumericAttributes("source", []float64{2, 3}); err != nil {
		t.Fatal(err)
	}
	if err := g.SetEdgeNumericAttributes("id", []float64{4}); err != nil {
		t.Fatal(err)
	}
	if err := g.SetGraphNumericAttribute("label", 5); err != nil {
		t.Fatal(err)
	}
	if err := g.SetVertexNumericAttributes("label", []float64{6, 7}); err != nil {
		t.Fatal(err)
	}
	if err := g.SetEdgeNumericAttributes("label", []float64{8}); err != nil {
		t.Fatal(err)
	}

	file := graphWriterTempFile(t)
	if err := g.WriteGML(file, GMLWriteOptions{}); err != nil {
		t.Fatal(err)
	}
	if _, err := file.Seek(0, 0); err != nil {
		t.Fatal(err)
	}
	readBack, err := ReadGML(file)
	if err != nil {
		t.Fatal(err)
	}
	defer readBack.Close()

	if value, err := readBack.GraphNumericAttribute("id"); err != nil || value != 1 {
		t.Fatalf("graph id = %v, %v", value, err)
	}
	if value, err := readBack.VertexNumericAttributes("source"); err != nil || !reflect.DeepEqual(value, []float64{2, 3}) {
		t.Fatalf("vertex source = %v, %v", value, err)
	}
	if value, err := readBack.EdgeNumericAttributes("id"); err != nil || !reflect.DeepEqual(value, []float64{4}) {
		t.Fatalf("edge id = %v, %v", value, err)
	}
	if value, err := readBack.GraphNumericAttribute("label"); err != nil || value != 5 {
		t.Fatalf("graph label = %v, %v", value, err)
	}
	if value, err := readBack.VertexNumericAttributes("label"); err != nil || !reflect.DeepEqual(value, []float64{6, 7}) {
		t.Fatalf("vertex label = %v, %v", value, err)
	}
	if value, err := readBack.EdgeNumericAttributes("label"); err != nil || !reflect.DeepEqual(value, []float64{8}) {
		t.Fatalf("edge label = %v, %v", value, err)
	}
}

func TestWritersRejectInvalidInputsAndOptions(t *testing.T) {
	g, err := NewGraph()
	if err != nil {
		t.Fatal(err)
	}
	defer g.Close()

	// nil file
	if err := g.WriteEdgeList(nil); err == nil {
		t.Fatal("expected error for nil file in WriteEdgeList")
	}
	if err := g.WriteGraphML(nil, false); err == nil {
		t.Fatal("expected error for nil file in WriteGraphML")
	}
	if err := g.WriteGML(nil, GMLWriteOptions{}); err == nil {
		t.Fatal("expected error for nil file in WriteGML")
	}

	// nil receiver
	var nilGraph *Graph
	file := graphWriterTempFile(t)
	if err := nilGraph.WriteEdgeList(file); !errors.Is(err, ErrClosed) {
		t.Fatalf("nil receiver WriteEdgeList = %v, want ErrClosed", err)
	}
	if err := nilGraph.WriteGraphML(file, false); !errors.Is(err, ErrClosed) {
		t.Fatalf("nil receiver WriteGraphML = %v, want ErrClosed", err)
	}
	if err := nilGraph.WriteGML(file, GMLWriteOptions{}); !errors.Is(err, ErrClosed) {
		t.Fatalf("nil receiver WriteGML = %v, want ErrClosed", err)
	}

	// closed graph
	closedGraph, err := NewGraph()
	if err != nil {
		t.Fatal(err)
	}
	if err := closedGraph.Close(); err != nil {
		t.Fatal(err)
	}
	if err := closedGraph.WriteEdgeList(file); !errors.Is(err, ErrClosed) {
		t.Fatalf("closed graph WriteEdgeList = %v, want ErrClosed", err)
	}
	if err := closedGraph.WriteGraphML(file, false); !errors.Is(err, ErrClosed) {
		t.Fatalf("closed graph WriteGraphML = %v, want ErrClosed", err)
	}
	if err := closedGraph.WriteGML(file, GMLWriteOptions{}); !errors.Is(err, ErrClosed) {
		t.Fatalf("closed graph WriteGML = %v, want ErrClosed", err)
	}

	// invalid GML options: NUL byte, quotes, and newlines in creator
	for _, invalidCreator := range []string{"invalid\x00creator", `creator"with"quotes`, "creator\nwith\nnewlines"} {
		if err := g.WriteGML(file, GMLWriteOptions{Creator: invalidCreator}); err == nil {
			t.Fatalf("expected error for invalid GML Creator %q", invalidCreator)
		}
	}

	// GML attribute names starting with digits, containing symbols, or reserved
	// in their specific scope are rejected.
	invalidAttributes := []struct {
		name string
		set  func(*Graph) error
	}{
		{name: "g_num", set: func(g *Graph) error { return g.SetGraphNumericAttribute("g_num", 1) }},
		{name: "1stAttr", set: func(g *Graph) error { return g.SetGraphNumericAttribute("1stAttr", 1) }},
		{name: "graph directed", set: func(g *Graph) error { return g.SetGraphNumericAttribute("directed", 1) }},
		{name: "graph node", set: func(g *Graph) error { return g.SetGraphNumericAttribute("node", 1) }},
		{name: "graph edge", set: func(g *Graph) error { return g.SetGraphNumericAttribute("edge", 1) }},
		{name: "duplicate vertex id", set: func(g *Graph) error { return g.SetVertexNumericAttributes("id", []float64{1, 1}) }},
		{name: "string vertex id", set: func(g *Graph) error { return g.SetVertexStringAttributes("id", []string{"a", "b"}) }},
		{name: "edge source", set: func(g *Graph) error { return g.SetEdgeNumericAttributes("source", []float64{1}) }},
		{name: "edge target", set: func(g *Graph) error { return g.SetEdgeNumericAttributes("target", []float64{1}) }},
	}
	for _, invalid := range invalidAttributes {
		gInv, err := NewGraphFromEdges(2, []Edge{{0, 1}}, true)
		if err != nil {
			t.Fatal(err)
		}
		if err := invalid.set(gInv); err != nil {
			gInv.Close()
			t.Fatal(err)
		}
		if err := gInv.WriteGML(file, GMLWriteOptions{}); err == nil {
			gInv.Close()
			t.Fatalf("expected error for invalid GML attribute %q", invalid.name)
		}
		gInv.Close()
	}

	// write failure to read-only fd (e.g. read end of pipe)
	rPipe, wPipe, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer rPipe.Close()
	defer wPipe.Close()
	if err := g.WriteEdgeList(rPipe); err == nil {
		t.Fatal("expected write error on read-only pipe for WriteEdgeList")
	}
	if err := g.WriteGraphML(rPipe, false); err == nil {
		t.Fatal("expected write error on read-only pipe for WriteGraphML")
	}
	if err := g.WriteGML(rPipe, GMLWriteOptions{}); err == nil {
		t.Fatal("expected write error on read-only pipe for WriteGML")
	}
}

func TestConcurrentWritersAndReaders(t *testing.T) {
	g, err := NewGraphFromEdges(4, []Edge{{0, 1}, {1, 2}, {2, 3}}, true)
	if err != nil {
		t.Fatal(err)
	}
	defer g.Close()

	if err := g.SetGraphNumericAttribute("gattr", 42.0); err != nil {
		t.Fatal(err)
	}
	if err := g.SetVertexStringAttributes("vname", []string{"a", "b", "c", "d"}); err != nil {
		t.Fatal(err)
	}

	const workers = 10
	var wg sync.WaitGroup
	errs := make(chan error, workers*4)

	for i := 0; i < workers; i++ {
		wg.Add(4)
		go func() {
			defer wg.Done()
			f := graphWriterTempFile(t)
			if err := g.WriteEdgeList(f); err != nil {
				errs <- err
			}
		}()
		go func() {
			defer wg.Done()
			f := graphWriterTempFile(t)
			if err := g.WriteGraphML(f, true); err != nil {
				errs <- err
			}
		}()
		go func() {
			defer wg.Done()
			f := graphWriterTempFile(t)
			if err := g.WriteGML(f, GMLWriteOptions{Creator: "concurrent"}); err != nil {
				errs <- err
			}
		}()
		go func() {
			defer wg.Done()
			if _, err := g.VertexStringAttributes("vname"); err != nil {
				errs <- err
			}
			if _, err := g.GraphNumericAttribute("gattr"); err != nil {
				errs <- err
			}
		}()
	}

	wg.Wait()
	close(errs)
	for err := range errs {
		t.Errorf("concurrent error: %v", err)
	}
}

func graphWriterTempFile(t *testing.T) *os.File {
	t.Helper()
	file, err := os.CreateTemp(t.TempDir(), "writer-*")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = file.Close() })
	return file
}
