package igraph

import (
	"errors"
	"os"
	"reflect"
	"sync"
	"testing"
)

func TestReadEdgeListTopologyOptionsAndFileBorrowing(t *testing.T) {
	file := graphReaderFile(t, "0 1\n1 1\n0 1\n")
	graph, err := ReadEdgeList(file, EdgeListReadOptions{VertexCount: 4, Directed: true})
	if err != nil {
		t.Fatal(err)
	}
	defer graph.Close()
	assertGraphShape(t, graph, 4, 3, true)
	assertEdgesEqual(t, graph, []Edge{{0, 1}, {1, 1}, {0, 1}})
	if offset, err := file.Seek(0, 1); err != nil || offset != 0 {
		t.Fatalf("file offset = %d, %v", offset, err)
	}
	if _, err := file.WriteAt([]byte("2 3\n"), 0); err != nil {
		t.Fatalf("caller file was closed: %v", err)
	}
}

func TestReadEdgeListStartsAtCurrentOffsetAndEmptyInput(t *testing.T) {
	const prefix = "ignored prefix\n"
	file := graphReaderFile(t, prefix+"2 3\n")
	if _, err := file.Seek(int64(len(prefix)), 0); err != nil {
		t.Fatal(err)
	}
	graph, err := ReadEdgeList(file, EdgeListReadOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer graph.Close()
	assertGraphShape(t, graph, 4, 1, false)
	if offset, _ := file.Seek(0, 1); offset != int64(len(prefix)) {
		t.Fatalf("file offset = %d", offset)
	}

	empty := graphReaderFile(t, "")
	emptyGraph, err := ReadEdgeList(empty, EdgeListReadOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer emptyGraph.Close()
	assertGraphShape(t, emptyGraph, 0, 0, false)
}

func TestReadGraphMLDirectednessAndTypedAttributes(t *testing.T) {
	file := graphReaderFile(t, `<?xml version="1.0" encoding="UTF-8"?>
<graphml xmlns="http://graphml.graphdrawing.org/xmlns">
  <key id="g_name" for="graph" attr.name="name" attr.type="string"/>
  <key id="v_score" for="node" attr.name="score" attr.type="double"/>
  <key id="v_active" for="node" attr.name="active" attr.type="boolean"/>
  <key id="e_label" for="edge" attr.name="label" attr.type="string"/>
  <graph id="selected" edgedefault="directed">
    <data key="g_name">selected</data>
    <node id="n0"><data key="v_score">1.5</data><data key="v_active">true</data></node>
    <node id="n1"><data key="v_score">2.5</data><data key="v_active">false</data></node>
    <edge source="n0" target="n1"><data key="e_label">route</data></edge>
  </graph>
</graphml>`)
	graph, err := ReadGraphML(file, 0)
	if err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	defer graph.Close()
	assertGraphShape(t, graph, 2, 1, true)
	if got, _ := graph.GraphStringAttribute("name"); got != "selected" {
		t.Fatalf("name = %q", got)
	}
	if got, _ := graph.VertexNumericAttributes("score"); !reflect.DeepEqual(got, []float64{1.5, 2.5}) {
		t.Fatalf("scores = %v", got)
	}
	if got, _ := graph.VertexBooleanAttributes("active"); !reflect.DeepEqual(got, []bool{true, false}) {
		t.Fatalf("active = %v", got)
	}
	if got, _ := graph.EdgeStringAttributes("label"); !reflect.DeepEqual(got, []string{"route"}) {
		t.Fatalf("labels = %v", got)
	}
}

func TestReadGraphMLSelectsGraphByIndex(t *testing.T) {
	file := graphReaderFile(t, `<?xml version="1.0" encoding="UTF-8"?>
<graphml xmlns="http://graphml.graphdrawing.org/xmlns">
  <graph id="first" edgedefault="undirected"><node id="a"/></graph>
  <graph id="second" edgedefault="directed"><node id="b"/><node id="c"/></graph>
</graphml>`)
	graph, err := ReadGraphML(file, 1)
	if err != nil {
		t.Fatal(err)
	}
	defer graph.Close()
	assertGraphShape(t, graph, 2, 0, true)
}

func TestReadGraphMLRejectsUnsupportedAttributeTypes(t *testing.T) {
	file := graphReaderFile(t, `<?xml version="1.0"?><graphml xmlns="http://graphml.graphdrawing.org/xmlns"><key id="k" for="node" attr.name="payload" attr.type="vector"/><graph edgedefault="undirected"><node id="a"><data key="k">1 2</data></node></graph></graphml>`)
	graph, err := ReadGraphML(file, 0)
	if err == nil || graph != nil {
		t.Fatalf("unsupported attribute = %v, %v", graph, err)
	}
}

func TestReadGMLDirectednessIdentityAndAttributes(t *testing.T) {
	file := graphReaderFile(t, `graph [
  directed 1
  title "demo"
  node [ id 10 label "A" score 1.5 ]
  node [ id 20 label "B" score 2.5 ]
  edge [ source 10 target 20 weight 3.5 label "route" ]
]`)
	graph, err := ReadGML(file)
	if err != nil {
		t.Fatal(err)
	}
	defer graph.Close()
	assertGraphShape(t, graph, 2, 1, true)
	assertEdgesEqual(t, graph, []Edge{{0, 1}})
	if got, _ := graph.GraphStringAttribute("title"); got != "demo" {
		t.Fatalf("title = %q", got)
	}
	if got, _ := graph.VertexStringAttributes("label"); !reflect.DeepEqual(got, []string{"A", "B"}) {
		t.Fatalf("labels = %v", got)
	}
	if got, _ := graph.VertexNumericAttributes("score"); !reflect.DeepEqual(got, []float64{1.5, 2.5}) {
		t.Fatalf("scores = %v", got)
	}
	if got, _ := graph.EdgeNumericAttributes("weight"); !reflect.DeepEqual(got, []float64{3.5}) {
		t.Fatalf("weights = %v", got)
	}
}

func TestGraphReadersRejectInvalidFilesAndOptions(t *testing.T) {
	if graph, err := ReadEdgeList(nil, EdgeListReadOptions{}); err == nil || graph != nil {
		t.Fatalf("nil edge list = %v, %v", graph, err)
	}
	file := graphReaderFile(t, "0 1\n")
	if graph, err := ReadEdgeList(file, EdgeListReadOptions{VertexCount: -1}); err == nil || graph != nil {
		t.Fatalf("negative vertex count = %v, %v", graph, err)
	}
	if graph, err := ReadGraphML(file, -1); err == nil || graph != nil {
		t.Fatalf("negative GraphML index = %v, %v", graph, err)
	}
	closed := graphReaderFile(t, "0 1\n")
	if err := closed.Close(); err != nil {
		t.Fatal(err)
	}
	if graph, err := ReadEdgeList(closed, EdgeListReadOptions{}); err == nil || graph != nil {
		t.Fatalf("closed file = %v, %v", graph, err)
	}
	readPipe, writePipe, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer readPipe.Close()
	defer writePipe.Close()
	if graph, err := ReadGML(readPipe); err == nil || graph != nil {
		t.Fatalf("pipe = %v, %v", graph, err)
	}
	for name, call := range map[string]func() (*Graph, error){
		"edge list": func() (*Graph, error) { return ReadEdgeList(graphReaderFile(t, "0 x\n"), EdgeListReadOptions{}) },
		"GraphML":   func() (*Graph, error) { return ReadGraphML(graphReaderFile(t, ""), 0) },
		"GML":       func() (*Graph, error) { return ReadGML(graphReaderFile(t, "")) },
	} {
		t.Run(name, func(t *testing.T) {
			if graph, err := call(); err == nil || graph != nil {
				t.Fatalf("malformed import = %v, %v", graph, err)
			}
		})
	}
	graphML := graphReaderFile(t, `<?xml version="1.0"?><graphml xmlns="http://graphml.graphdrawing.org/xmlns"><graph id="g" edgedefault="undirected"/></graphml>`)
	if graph, err := ReadGraphML(graphML, 1); err == nil || graph != nil {
		t.Fatalf("out-of-range GraphML index = %v, %v", graph, err)
	}
	if graph, err := readGraphFile(file, "test", nil); err == nil || graph != nil {
		t.Fatalf("nil reader = %v, %v", graph, err)
	}
}

func TestConcurrentGraphMLImportsAndIndependentClose(t *testing.T) {
	file := graphReaderFile(t, `<?xml version="1.0"?><graphml xmlns="http://graphml.graphdrawing.org/xmlns"><key id="k" for="node" attr.name="label" attr.type="string"/><graph id="g" edgedefault="undirected"><node id="a"><data key="k">A</data></node></graph></graphml>`)
	const workers = 12
	results := make(chan *Graph, workers)
	errorsCh := make(chan error, workers)
	var wait sync.WaitGroup
	for i := 0; i < workers; i++ {
		wait.Add(1)
		go func() {
			defer wait.Done()
			graph, err := ReadGraphML(file, 0)
			if err != nil {
				errorsCh <- err
				return
			}
			results <- graph
		}()
	}
	wait.Wait()
	close(results)
	close(errorsCh)
	for err := range errorsCh {
		t.Error(err)
	}
	graphs := make([]*Graph, 0, workers)
	for graph := range results {
		graphs = append(graphs, graph)
	}
	if len(graphs) != workers {
		t.Fatalf("graphs = %d, want %d", len(graphs), workers)
	}
	if err := graphs[0].Close(); err != nil {
		t.Fatal(err)
	}
	for _, graph := range graphs[1:] {
		if got, err := graph.VertexStringAttributes("label"); err != nil || !reflect.DeepEqual(got, []string{"A"}) {
			t.Errorf("independent graph = %v, %v", got, err)
		}
		_ = graph.Close()
	}
	if _, err := graphs[0].VertexAttributes(); !errors.Is(err, ErrClosed) {
		t.Fatalf("read after Close = %v", err)
	}
}

func graphReaderFile(t *testing.T, contents string) *os.File {
	t.Helper()
	file, err := os.CreateTemp(t.TempDir(), "reader-*")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = file.Close() })
	if _, err := file.WriteString(contents); err != nil {
		t.Fatal(err)
	}
	if _, err := file.Seek(0, 0); err != nil {
		t.Fatal(err)
	}
	return file
}
