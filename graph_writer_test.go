package igraph

import (
	"errors"
	"os"
	"reflect"
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
		sortEdges := func(s []Edge) {
			for i := 0; i < len(s); i++ {
				for j := i + 1; j < len(s); j++ {
					if s[i].From > s[j].From || (s[i].From == s[j].From && s[i].To > s[j].To) {
						s[i], s[j] = s[j], s[i]
					}
				}
			}
		}
		sortEdges(normRead)
		sortEdges(normOrig)
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

	// invalid GML options: NUL byte in creator
	if err := g.WriteGML(file, GMLWriteOptions{Creator: "invalid\x00creator"}); err == nil {
		t.Fatal("expected error for NUL byte in GML Creator")
	}

	// GML attribute names containing non-alphanumeric characters are rejected
	gWithInvalidAttr, err := NewGraph()
	if err != nil {
		t.Fatal(err)
	}
	defer gWithInvalidAttr.Close()
	if err := gWithInvalidAttr.SetGraphNumericAttribute("g_num", 1.0); err != nil {
		t.Fatal(err)
	}
	if err := gWithInvalidAttr.WriteGML(file, GMLWriteOptions{}); err == nil {
		t.Fatal("expected error for attribute name containing underscore in WriteGML")
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
