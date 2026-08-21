package igraph_test

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	igraph "github.com/h8gi/go-igraph"
)

func TestMilestone17AttributedInterchangePipeline(t *testing.T) {
	left := importMilestone17Graph(t, "left", []igraph.Edge{{From: 0, To: 1}, {From: 1, To: 2}}, []float64{2, 4})
	right := importMilestone17Graph(t, "right", []igraph.Edge{{From: 0, To: 1}, {From: 2, To: 0}}, []float64{3, 8})

	metadata, err := left.VertexAttributes()
	if err != nil || len(metadata) == 0 {
		t.Fatalf("vertex metadata = %v, %v", metadata, err)
	}
	if err := left.SetVertexNumericAttribute("score", 1, 12); err != nil {
		t.Fatal(err)
	}
	if err := right.SetVertexNumericAttribute("score", 1, 18); err != nil {
		t.Fatal(err)
	}

	policy := &igraph.GraphOperatorAttributePolicy{
		Graph: igraph.AttributeCombinationPolicy{Default: igraph.AttributeCombineFirst},
		Vertices: igraph.AttributeCombinationPolicy{
			Default:   igraph.AttributeCombineFirst,
			Overrides: map[string]igraph.AttributeCombination{"score": igraph.AttributeCombineMean},
		},
		Edges: igraph.AttributeCombinationPolicy{
			Default:   igraph.AttributeCombineFirst,
			Overrides: map[string]igraph.AttributeCombination{"weight": igraph.AttributeCombineSum},
		},
	}
	result, err := left.Union(right, policy)
	if err != nil {
		t.Fatal(err)
	}
	defer result.Graph.Close()

	assertMilestone17Mapping(t, left, result.Graph, result.Left)
	assertMilestone17Mapping(t, right, result.Graph, result.Right)
	scores, err := result.Graph.VertexNumericAttributes("score")
	if err != nil || !reflect.DeepEqual(scores, []float64{10, 15, 30}) {
		t.Fatalf("combined scores = %v, %v", scores, err)
	}
	weights, err := result.Graph.EdgeNumericAttributes("weight")
	if err != nil {
		t.Fatal(err)
	}
	sharedEdge := result.Left.Edges.OldToNew[0]
	if sharedEdge != result.Right.Edges.OldToNew[0] || weights[sharedEdge] != 5 {
		t.Fatalf("shared edge mappings = %v/%v, weights = %v", result.Left.Edges.OldToNew, result.Right.Edges.OldToNew, weights)
	}

	if err := left.Close(); err != nil {
		t.Fatal(err)
	}
	if err := right.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := left.VertexAttributes(); !errors.Is(err, igraph.ErrClosed) {
		t.Fatalf("post-close attribute error = %v, want ErrClosed", err)
	}
	if got, err := result.Graph.VertexNumericAttributes("score"); err != nil || !reflect.DeepEqual(got, scores) {
		t.Fatalf("derived values after operand closure = %v, %v", got, err)
	}

	path := filepath.Join(t.TempDir(), "combined.graphml")
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := result.Graph.WriteGraphML(file, false); err != nil {
		file.Close()
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	if err := result.Graph.Close(); err != nil {
		t.Fatal(err)
	}

	file, err = os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	reimported, err := igraph.ReadGraphML(file, 0)
	closeErr := file.Close()
	if err != nil {
		t.Fatal(err)
	}
	if closeErr != nil {
		t.Fatal(closeErr)
	}
	defer reimported.Close()
	if got, err := reimported.VertexNumericAttributes("score"); err != nil || !reflect.DeepEqual(got, scores) {
		t.Fatalf("round-trip scores = %v, %v", got, err)
	}
	if got, err := reimported.EdgeNumericAttributes("weight"); err != nil || !reflect.DeepEqual(got, weights) {
		t.Fatalf("round-trip weights = %v, %v", got, err)
	}
}

func importMilestone17Graph(t *testing.T, name string, edges []igraph.Edge, weights []float64) *igraph.Graph {
	t.Helper()
	source, err := igraph.NewGraphFromEdges(3, edges, true)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = source.Close() })
	if err := source.SetGraphStringAttribute("source", name); err != nil {
		t.Fatal(err)
	}
	if err := source.SetVertexNumericAttributes("score", []float64{10, 20, 30}); err != nil {
		t.Fatal(err)
	}
	if err := source.SetVertexBooleanAttributes("active", []bool{true, false, true}); err != nil {
		t.Fatal(err)
	}
	if err := source.SetEdgeNumericAttributes("weight", weights); err != nil {
		t.Fatal(err)
	}

	file, err := os.CreateTemp(t.TempDir(), name+"-*.graphml")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = file.Close() })
	if err := source.WriteGraphML(file, false); err != nil {
		t.Fatal(err)
	}
	if err := source.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := file.Seek(0, 0); err != nil {
		t.Fatal(err)
	}
	imported, err := igraph.ReadGraphML(file, 0)
	if err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = imported.Close() })
	return imported
}

func assertMilestone17Mapping(t *testing.T, source, result *igraph.Graph, mapping igraph.GraphIDMapping) {
	t.Helper()
	sourceEdges, err := source.Edges()
	if err != nil {
		t.Fatal(err)
	}
	resultEdges, err := result.Edges()
	if err != nil {
		t.Fatal(err)
	}
	vertices, err := source.VertexCount()
	if err != nil {
		t.Fatal(err)
	}
	if len(mapping.Vertices.OldToNew) != vertices || len(mapping.Edges.OldToNew) != len(sourceEdges) {
		t.Fatalf("mapping lengths = vertices %d, edges %d", len(mapping.Vertices.OldToNew), len(mapping.Edges.OldToNew))
	}
	for id, mapped := range mapping.Vertices.OldToNew {
		if mapped != id {
			t.Fatalf("vertex %d maps to %d", id, mapped)
		}
	}
	for id, mapped := range mapping.Edges.OldToNew {
		want := igraph.Edge{From: mapping.Vertices.OldToNew[sourceEdges[id].From], To: mapping.Vertices.OldToNew[sourceEdges[id].To]}
		if mapped < 0 || mapped >= len(resultEdges) || resultEdges[mapped] != want {
			t.Fatalf("edge %d mapping = %d (%v), want endpoint %v", id, mapped, resultEdges, want)
		}
	}
}
