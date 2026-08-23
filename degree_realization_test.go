package igraph_test

import (
	"reflect"
	"testing"

	igraph "github.com/h8gi/go-igraph"
)

func TestRealizeDegreeSequence(t *testing.T) {
	methods := []igraph.DegreeSequenceRealizationMethod{
		igraph.RealizeSmallest,
		igraph.RealizeLargest,
		igraph.RealizeIndex,
	}
	for _, method := range methods {
		graph, err := igraph.RealizeDegreeSequence(
			[]int{2, 2, 2, 2}, nil,
			igraph.DegreeSequenceRealizationOptions{Method: method},
		)
		if err != nil {
			t.Fatalf("RealizeDegreeSequence(method %d) failed: %v", method, err)
		}
		defer graph.Close()
		assertDegrees(t, graph, igraph.DirectionAll, []int{2, 2, 2, 2})
	}

	directed, err := igraph.RealizeDegreeSequence(
		[]int{2, 1, 0}, []int{0, 1, 2}, igraph.DegreeSequenceRealizationOptions{},
	)
	if err != nil {
		t.Fatalf("directed RealizeDegreeSequence failed: %v", err)
	}
	defer directed.Close()
	assertDegrees(t, directed, igraph.DirectionOut, []int{2, 1, 0})
	assertDegrees(t, directed, igraph.DirectionIn, []int{0, 1, 2})

	for name, options := range map[string]igraph.DegreeSequenceRealizationOptions{
		"multi":           {EdgeTypes: igraph.EdgeTypeMulti},
		"loops and multi": {EdgeTypes: igraph.EdgeTypeLoopsAndMulti},
	} {
		t.Run(name, func(t *testing.T) {
			graph, err := igraph.RealizeDegreeSequence([]int{2, 2}, nil, options)
			if err != nil {
				t.Fatalf("RealizeDegreeSequence failed: %v", err)
			}
			defer graph.Close()
			assertDegrees(t, graph, igraph.DirectionAll, []int{2, 2})
		})
	}

	empty, err := igraph.RealizeDegreeSequence(nil, []int{}, igraph.DegreeSequenceRealizationOptions{})
	if err != nil {
		t.Fatalf("empty RealizeDegreeSequence failed: %v", err)
	}
	defer empty.Close()
	if vertices, err := empty.VertexCount(); err != nil || vertices != 0 {
		t.Errorf("empty vertex count = %d, %v, want 0, nil", vertices, err)
	}
}

func TestRealizeBipartiteDegreeSequence(t *testing.T) {
	result, err := igraph.RealizeBipartiteDegreeSequence(
		[]int{2, 1}, []int{1, 1, 1}, igraph.DegreeSequenceRealizationOptions{},
	)
	if err != nil {
		t.Fatalf("RealizeBipartiteDegreeSequence failed: %v", err)
	}
	defer result.Graph.Close()
	assertDegrees(t, result.Graph, igraph.DirectionAll, []int{2, 1, 1, 1, 1})
	if want := (igraph.BipartitePartition{false, false, true, true, true}); !reflect.DeepEqual(result.Partition, want) {
		t.Errorf("partition = %#v, want %#v", result.Partition, want)
	}

	multi, err := igraph.RealizeBipartiteDegreeSequence(
		[]int{2}, []int{2},
		igraph.DegreeSequenceRealizationOptions{EdgeTypes: igraph.EdgeTypeLoopsAndMulti, Method: igraph.RealizeIndex},
	)
	if err != nil {
		t.Fatalf("multi RealizeBipartiteDegreeSequence failed: %v", err)
	}
	defer multi.Graph.Close()
	assertDegrees(t, multi.Graph, igraph.DirectionAll, []int{2, 2})

	empty, err := igraph.RealizeBipartiteDegreeSequence(nil, nil, igraph.DegreeSequenceRealizationOptions{})
	if err != nil {
		t.Fatalf("empty RealizeBipartiteDegreeSequence failed: %v", err)
	}
	defer empty.Graph.Close()
	if empty.Partition == nil || len(empty.Partition) != 0 {
		t.Errorf("empty partition = %#v, want non-nil empty", empty.Partition)
	}
}

func TestDegreeSequenceRealizationValidation(t *testing.T) {
	ordinary := []struct {
		name string
		out  []int
		in   []int
		opts igraph.DegreeSequenceRealizationOptions
	}{
		{"negative", []int{-1}, nil, igraph.DegreeSequenceRealizationOptions{}},
		{"odd undirected sum", []int{1}, nil, igraph.DegreeSequenceRealizationOptions{}},
		{"directed length mismatch", []int{1, 0}, []int{1}, igraph.DegreeSequenceRealizationOptions{}},
		{"directed sum mismatch", []int{1, 0}, []int{0, 0}, igraph.DegreeSequenceRealizationOptions{}},
		{"directed multi", []int{1, 0}, []int{0, 1}, igraph.DegreeSequenceRealizationOptions{EdgeTypes: igraph.EdgeTypeMulti}},
		{"loops only", []int{2}, nil, igraph.DegreeSequenceRealizationOptions{EdgeTypes: igraph.EdgeTypeLoops}},
		{"invalid edge type", nil, nil, igraph.DegreeSequenceRealizationOptions{EdgeTypes: igraph.EdgeType(255)}},
		{"invalid method", nil, nil, igraph.DegreeSequenceRealizationOptions{Method: igraph.DegreeSequenceRealizationMethod(255)}},
		{"not graphical", []int{2, 0}, nil, igraph.DegreeSequenceRealizationOptions{}},
	}
	for _, test := range ordinary {
		t.Run(test.name, func(t *testing.T) {
			if graph, err := igraph.RealizeDegreeSequence(test.out, test.in, test.opts); err == nil {
				graph.Close()
				t.Error("RealizeDegreeSequence succeeded")
			}
		})
	}

	if result, err := igraph.RealizeBipartiteDegreeSequence(
		[]int{1}, []int{0}, igraph.DegreeSequenceRealizationOptions{},
	); err == nil {
		result.Graph.Close()
		t.Error("mismatched bipartite sums succeeded")
	}
}

func assertDegrees(t *testing.T, graph *igraph.Graph, direction igraph.DirectionMode, want []int) {
	t.Helper()
	got, err := graph.Degree(igraph.AllVertices(), igraph.DegreeOptions{Direction: direction, CountLoops: true})
	if err != nil {
		t.Fatalf("Degree failed: %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("degrees = %v, want %v", got, want)
	}
}
