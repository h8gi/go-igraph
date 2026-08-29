package igraph

import (
	"errors"
	"reflect"
	"sync"
	"testing"
)

func TestCohesiveBlocksKnownAnswerAndOwnership(t *testing.T) {
	edges := []Edge{
		{0, 1}, {0, 2}, {0, 3}, {1, 2}, {1, 3}, {2, 3},
		{0, 4}, {0, 5}, {1, 4}, {1, 5}, {4, 5},
	}
	g, err := NewGraphFromEdges(6, edges, false)
	if err != nil {
		t.Fatal(err)
	}
	result, err := g.CohesiveBlocks()
	if err != nil {
		t.Fatal(err)
	}
	if err := g.Close(); err != nil {
		t.Fatal(err)
	}
	defer result.BlockTree.Close()

	wantBlocks := [][]int{{0, 1, 2, 3, 4, 5}, {0, 1, 2, 3}, {0, 1, 4, 5}}
	if !reflect.DeepEqual(result.Blocks, wantBlocks) {
		t.Errorf("Blocks = %v, want %v", result.Blocks, wantBlocks)
	}
	if !reflect.DeepEqual(result.Cohesion, []int{2, 3, 3}) || !reflect.DeepEqual(result.Parents, []int{-1, 0, 0}) {
		t.Errorf("Cohesion/Parents = %v/%v", result.Cohesion, result.Parents)
	}
	if vertices, err := result.BlockTree.VertexCount(); err != nil || vertices != 3 {
		t.Errorf("block tree vertex count = %d, %v", vertices, err)
	}
	if treeEdges, err := result.BlockTree.Edges(); err != nil || !reflect.DeepEqual(treeEdges, []Edge{{0, 1}, {0, 2}}) {
		t.Errorf("block tree edges = %v, %v", treeEdges, err)
	}
}

func TestCohesiveBlocksSmallAndDisconnectedGraphs(t *testing.T) {
	tests := []struct {
		name     string
		vertices int
		edges    []Edge
		blocks   [][]int
		cohesion []int
		parents  []int
	}{
		{name: "empty", blocks: [][]int{{}}, cohesion: []int{0}, parents: []int{-1}},
		{name: "singleton", vertices: 1, blocks: [][]int{{0}}, cohesion: []int{0}, parents: []int{-1}},
		{name: "disconnected", vertices: 3, edges: []Edge{{0, 1}}, blocks: [][]int{{0, 1, 2}, {0, 1}}, cohesion: []int{0, 1}, parents: []int{-1, 0}},
		{name: "single edge", vertices: 2, edges: []Edge{{0, 1}}, blocks: [][]int{{0, 1}}, cohesion: []int{1}, parents: []int{-1}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g, err := NewGraphFromEdges(tt.vertices, tt.edges, false)
			if err != nil {
				t.Fatal(err)
			}
			defer g.Close()
			result, err := g.CohesiveBlocks()
			if err != nil {
				t.Fatal(err)
			}
			defer result.BlockTree.Close()
			if !reflect.DeepEqual(result.Blocks, tt.blocks) || !reflect.DeepEqual(result.Cohesion, tt.cohesion) || !reflect.DeepEqual(result.Parents, tt.parents) {
				t.Errorf("CohesiveBlocks() = %#v", result)
			}
		})
	}
}

func TestCohesiveBlocksRejectsInvalidAndClosedGraphs(t *testing.T) {
	graphs := []struct {
		vertices int
		edges    []Edge
		directed bool
	}{
		{2, []Edge{{0, 1}}, true},
		{1, []Edge{{0, 0}}, false},
		{2, []Edge{{0, 1}, {0, 1}}, false},
	}
	for _, tt := range graphs {
		g, err := NewGraphFromEdges(tt.vertices, tt.edges, tt.directed)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := g.CohesiveBlocks(); err == nil {
			t.Error("CohesiveBlocks() error = nil")
		}
		g.Close()
	}
	closed, err := NewGraph()
	if err != nil {
		t.Fatal(err)
	}
	closed.Close()
	for _, g := range []*Graph{closed, nil} {
		if _, err := g.CohesiveBlocks(); !errors.Is(err, ErrClosed) {
			t.Errorf("CohesiveBlocks() error = %v, want %v", err, ErrClosed)
		}
	}
}

func TestCohesiveBlocksAllowsConcurrentReads(t *testing.T) {
	g, err := NewGraphFromEdges(4, []Edge{{0, 1}, {1, 2}, {2, 3}, {3, 0}}, false)
	if err != nil {
		t.Fatal(err)
	}
	defer g.Close()
	var wait sync.WaitGroup
	for range 8 {
		wait.Add(1)
		go func() {
			defer wait.Done()
			result, err := g.CohesiveBlocks()
			if err != nil {
				t.Error(err)
				return
			}
			if err := result.BlockTree.Close(); err != nil {
				t.Error(err)
			}
			if err := result.BlockTree.Close(); err != nil {
				t.Error(err)
			}
			if _, err := result.BlockTree.VertexCount(); !errors.Is(err, ErrClosed) {
				t.Errorf("closed block tree error = %v", err)
			}
		}()
	}
	wait.Wait()
}
