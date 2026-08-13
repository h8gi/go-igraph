package igraph_test

import (
	"errors"
	"reflect"
	"sort"
	"testing"

	igraph "github.com/h8gi/go-igraph"
)

func TestDyadCensusKnownAnswers(t *testing.T) {
	tests := []struct {
		name     string
		vertices int
		edges    []igraph.Edge
		directed bool
		want     igraph.DyadCensusResult
	}{
		{name: "empty", directed: true, want: igraph.DyadCensusResult{}},
		{name: "isolates", vertices: 3, directed: true, want: igraph.DyadCensusResult{Null: 3}},
		{name: "directed mixed", vertices: 3, directed: true, edges: []igraph.Edge{
			{From: 0, To: 1}, {From: 1, To: 0}, {From: 1, To: 2},
		}, want: igraph.DyadCensusResult{Mutual: 1, Asymmetric: 1, Null: 1}},
		{name: "undirected edge is mutual", vertices: 2, directed: false,
			edges: []igraph.Edge{{From: 0, To: 1}}, want: igraph.DyadCensusResult{Mutual: 1}},
		{name: "loops and multiplicity ignored", vertices: 2, directed: true, edges: []igraph.Edge{
			{From: 0, To: 0}, {From: 0, To: 1}, {From: 0, To: 1},
		}, want: igraph.DyadCensusResult{Asymmetric: 1}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			graph := newMotifGraph(t, test.vertices, test.edges, test.directed)
			got, err := graph.DyadCensus()
			if err != nil || got != test.want {
				t.Fatalf("DyadCensus = %#v, %v; want %#v", got, err, test.want)
			}
		})
	}
}

func TestTriadCensusKnownAnswersAndOwnership(t *testing.T) {
	graph := newMotifGraph(t, 4, []igraph.Edge{
		{From: 0, To: 1}, {From: 1, To: 2}, {From: 2, To: 0},
	}, true)
	result, err := graph.TriadCensus()
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 16 {
		t.Fatalf("TriadCensus length = %d", len(result))
	}
	var total int64
	for _, count := range result {
		total += count
	}
	if total != 4 {
		t.Fatalf("TriadCensus total = %d, want C(4,3)=4: %v", total, result)
	}
	if err := graph.Close(); err != nil {
		t.Fatal(err)
	}
	result[0]++
}

func TestTriangleQueriesIgnoreDirectionMultiplicityAndLoops(t *testing.T) {
	edges := []igraph.Edge{
		{From: 0, To: 1}, {From: 0, To: 2}, {From: 1, To: 2},
		{From: 0, To: 1}, {From: 2, To: 2},
	}
	graph := newMotifGraph(t, 4, edges, true)
	count, err := graph.TrianglesCount()
	if err != nil || count != 1 {
		t.Fatalf("TrianglesCount = %d, %v", count, err)
	}
	triangles, err := graph.TrianglesList()
	if err != nil {
		t.Fatal(err)
	}
	if got := canonicalTriangles(triangles); !reflect.DeepEqual(got, [][3]int{{0, 1, 2}}) {
		t.Fatalf("TrianglesList = %v", triangles)
	}
	selector, err := igraph.VertexIDs(2, 0, 3, 2)
	if err != nil {
		t.Fatal(err)
	}
	adjacent, err := graph.AdjacentTrianglesCount(selector)
	if err != nil || !reflect.DeepEqual(adjacent, []int64{1, 1, 0, 1}) {
		t.Fatalf("AdjacentTrianglesCount = %v, %v", adjacent, err)
	}
	if err := graph.Close(); err != nil {
		t.Fatal(err)
	}
	triangles[0][0] = 99
	adjacent[0] = 99
}

func TestTriangleQueriesEmptyAndNoSelection(t *testing.T) {
	empty := newMotifGraph(t, 0, nil, false)
	if count, err := empty.TrianglesCount(); err != nil || count != 0 {
		t.Fatalf("empty TrianglesCount = %d, %v", count, err)
	}
	if triangles, err := empty.TrianglesList(); err != nil || triangles == nil || len(triangles) != 0 {
		t.Fatalf("empty TrianglesList = %#v, %v", triangles, err)
	}
	if counts, err := empty.AdjacentTrianglesCount(igraph.NoVertices()); err != nil || counts == nil || len(counts) != 0 {
		t.Fatalf("empty AdjacentTrianglesCount = %#v, %v", counts, err)
	}
}

func TestAdjacentTrianglesCountValidatesSelector(t *testing.T) {
	graph := newMotifGraph(t, 3, nil, false)
	selector, err := igraph.VertexIDs(3)
	if err != nil {
		t.Fatal(err)
	}
	if result, err := graph.AdjacentTrianglesCount(selector); err == nil || result != nil {
		t.Fatalf("invalid selector = %#v, %v", result, err)
	}
}

func TestMotifQueriesClosedAndNil(t *testing.T) {
	var nilGraph *igraph.Graph
	if _, err := nilGraph.DyadCensus(); !errors.Is(err, igraph.ErrClosed) {
		t.Errorf("nil DyadCensus error = %v", err)
	}
	graph := newMotifGraph(t, 3, nil, true)
	if err := graph.Close(); err != nil {
		t.Fatal(err)
	}
	calls := []func() error{
		func() error { _, err := graph.DyadCensus(); return err },
		func() error { _, err := graph.TriadCensus(); return err },
		func() error { _, err := graph.AdjacentTrianglesCount(igraph.AllVertices()); return err },
		func() error { _, err := graph.TrianglesCount(); return err },
		func() error { _, err := graph.TrianglesList(); return err },
	}
	for index, call := range calls {
		if err := call(); !errors.Is(err, igraph.ErrClosed) {
			t.Errorf("closed motif call %d error = %v", index, err)
		}
	}
}

func newMotifGraph(t *testing.T, vertices int, edges []igraph.Edge, directed bool) *igraph.Graph {
	t.Helper()
	graph, err := igraph.NewGraphFromEdges(vertices, edges, directed)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = graph.Close() })
	return graph
}

func canonicalTriangles(values [][3]int) [][3]int {
	result := append([][3]int{}, values...)
	for index := range result {
		sort.Ints(result[index][:])
	}
	sort.Slice(result, func(i, j int) bool {
		for column := 0; column < 3; column++ {
			if result[i][column] != result[j][column] {
				return result[i][column] < result[j][column]
			}
		}
		return false
	})
	return result
}
