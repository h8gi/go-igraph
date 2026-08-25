package igraph

import (
	"errors"
	"reflect"
	"sync"
	"testing"
)

func TestBondPercolationKnownAnswer(t *testing.T) {
	for _, directed := range []bool{false, true} {
		g, err := NewGraphFromEdges(4, []Edge{{From: 0, To: 1}, {From: 1, To: 2}, {From: 2, To: 3}}, directed)
		if err != nil {
			t.Fatal(err)
		}
		result, err := g.BondPercolation([]int{1, 0, 2})
		if err != nil {
			t.Fatal(err)
		}
		if err := g.Close(); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(result.GiantComponentSizes, []int{2, 3, 4}) {
			t.Errorf("GiantComponentSizes = %v", result.GiantComponentSizes)
		}
		if !reflect.DeepEqual(result.ActiveVertexCounts, []int{2, 3, 4}) {
			t.Errorf("ActiveVertexCounts = %v", result.ActiveVertexCounts)
		}
	}
}

func TestSitePercolationKnownAnswer(t *testing.T) {
	g, err := NewGraphFromEdges(4, []Edge{{From: 0, To: 1}, {From: 1, To: 2}, {From: 2, To: 3}}, false)
	if err != nil {
		t.Fatal(err)
	}
	result, err := g.SitePercolation([]int{1, 3, 0, 2})
	if err != nil {
		t.Fatal(err)
	}
	if err := g.Close(); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(result.GiantComponentSizes, []int{1, 1, 2, 4}) {
		t.Errorf("GiantComponentSizes = %v", result.GiantComponentSizes)
	}
	if !reflect.DeepEqual(result.ActiveEdgeCounts, []int{0, 0, 1, 3}) {
		t.Errorf("ActiveEdgeCounts = %v", result.ActiveEdgeCounts)
	}
}

func TestPercolationLoopsParallelEdgesAndIsolates(t *testing.T) {
	g, err := NewGraphFromEdges(3, []Edge{{From: 0, To: 0}, {From: 0, To: 1}, {From: 0, To: 1}}, false)
	if err != nil {
		t.Fatal(err)
	}
	defer g.Close()
	bond, err := g.BondPercolation([]int{0, 1, 2})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(bond.GiantComponentSizes, []int{1, 2, 2}) || !reflect.DeepEqual(bond.ActiveVertexCounts, []int{1, 2, 2}) {
		t.Errorf("BondPercolation() = %#v", bond)
	}
	site, err := g.SitePercolation([]int{2, 0, 1})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(site.GiantComponentSizes, []int{1, 1, 2}) || !reflect.DeepEqual(site.ActiveEdgeCounts, []int{0, 2, 4}) {
		t.Errorf("SitePercolation() = %#v", site)
	}
}

func TestEdgeListPercolationKnownAnswer(t *testing.T) {
	result, err := EdgeListPercolation([]Edge{
		{From: 0, To: 1},
		{From: 2, To: 3},
		{From: 1, To: 2},
		{From: 0, To: 0},
		{From: 0, To: 1},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(result.GiantComponentSizes, []int{2, 2, 4, 4, 4}) {
		t.Errorf("GiantComponentSizes = %v", result.GiantComponentSizes)
	}
	if !reflect.DeepEqual(result.ActiveVertexCounts, []int{2, 4, 4, 4, 4}) {
		t.Errorf("ActiveVertexCounts = %v", result.ActiveVertexCounts)
	}
}

func TestPercolationEmptyInputs(t *testing.T) {
	g, err := NewGraph()
	if err != nil {
		t.Fatal(err)
	}
	defer g.Close()
	bond, err := g.BondPercolation(nil)
	if err != nil || bond.GiantComponentSizes == nil || bond.ActiveVertexCounts == nil || len(bond.GiantComponentSizes) != 0 || len(bond.ActiveVertexCounts) != 0 {
		t.Errorf("BondPercolation(nil) = %#v, %v", bond, err)
	}
	site, err := g.SitePercolation(nil)
	if err != nil || site.GiantComponentSizes == nil || site.ActiveEdgeCounts == nil || len(site.GiantComponentSizes) != 0 || len(site.ActiveEdgeCounts) != 0 {
		t.Errorf("SitePercolation(nil) = %#v, %v", site, err)
	}
	edgeList, err := EdgeListPercolation(nil)
	if err != nil || edgeList.GiantComponentSizes == nil || edgeList.ActiveVertexCounts == nil || len(edgeList.GiantComponentSizes) != 0 || len(edgeList.ActiveVertexCounts) != 0 {
		t.Errorf("EdgeListPercolation(nil) = %#v, %v", edgeList, err)
	}
}

func TestPercolationRejectsInvalidInputAndClosedGraphs(t *testing.T) {
	g, err := NewGraphFromEdges(3, []Edge{{From: 0, To: 1}, {From: 1, To: 2}}, false)
	if err != nil {
		t.Fatal(err)
	}
	for _, order := range [][]int{nil, {0}, {0, 0}, {0, 2}, {-1, 0}} {
		if _, err := g.BondPercolation(order); err == nil {
			t.Errorf("BondPercolation(%v) error = nil", order)
		}
	}
	for _, order := range [][]int{nil, {0, 1}, {0, 1, 1}, {0, 1, 3}, {-1, 1, 2}} {
		if _, err := g.SitePercolation(order); err == nil {
			t.Errorf("SitePercolation(%v) error = nil", order)
		}
	}
	for _, edges := range [][]Edge{{{From: -1, To: 0}}, {{From: 0, To: -1}}} {
		if _, err := EdgeListPercolation(edges); err == nil {
			t.Errorf("EdgeListPercolation(%v) error = nil", edges)
		}
	}
	if err := g.Close(); err != nil {
		t.Fatal(err)
	}
	for _, graph := range []*Graph{g, nil} {
		if _, err := graph.BondPercolation([]int{0, 1}); !errors.Is(err, ErrClosed) {
			t.Errorf("BondPercolation() error = %v, want %v", err, ErrClosed)
		}
		if _, err := graph.SitePercolation([]int{0, 1, 2}); !errors.Is(err, ErrClosed) {
			t.Errorf("SitePercolation() error = %v, want %v", err, ErrClosed)
		}
	}
}

func TestPercolationAllowsConcurrentReads(t *testing.T) {
	g, err := NewGraphFromEdges(3, []Edge{{From: 0, To: 1}, {From: 1, To: 2}}, false)
	if err != nil {
		t.Fatal(err)
	}
	defer g.Close()
	var wait sync.WaitGroup
	for range 8 {
		wait.Add(1)
		go func() {
			defer wait.Done()
			if result, err := g.BondPercolation([]int{0, 1}); err != nil || result.GiantComponentSizes[1] != 3 {
				t.Errorf("BondPercolation() = %#v, %v", result, err)
			}
			if result, err := g.SitePercolation([]int{0, 1, 2}); err != nil || result.GiantComponentSizes[2] != 3 {
				t.Errorf("SitePercolation() = %#v, %v", result, err)
			}
		}()
	}
	wait.Wait()
}
