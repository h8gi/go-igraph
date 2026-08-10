package igraph_test

import (
	"math"
	"sync"
	"testing"

	"github.com/h8gi/go-igraph"
)

func TestLayoutCircle(t *testing.T) {
	t.Run("empty graph", func(t *testing.T) {
		g, err := igraph.NewGraph()
		if err != nil {
			t.Fatalf("NewGraph failed: %v", err)
		}
		defer g.Close()

		coords, err := g.LayoutCircle(nil)
		if err != nil {
			t.Fatalf("LayoutCircle failed: %v", err)
		}
		rows, cols := coords.Dims()
		if rows != 0 || cols != 2 {
			t.Errorf("got dims (%d, %d), want (0, 2)", rows, cols)
		}
	})

	t.Run("single vertex graph", func(t *testing.T) {
		g, err := igraph.NewGraphFromEdges(1, nil, false)
		if err != nil {
			t.Fatalf("NewGraphFromEdges failed: %v", err)
		}
		defer g.Close()

		coords, err := g.LayoutCircle(nil)
		if err != nil {
			t.Fatalf("LayoutCircle failed: %v", err)
		}
		rows, cols := coords.Dims()
		if rows != 1 || cols != 2 {
			t.Errorf("got dims (%d, %d), want (1, 2)", rows, cols)
		}
	})

	t.Run("ring graph default and custom order", func(t *testing.T) {
		g, err := igraph.NewRing(5, false, false)
		if err != nil {
			t.Fatalf("NewRing failed: %v", err)
		}
		defer g.Close()

		coordsDefault, err := g.LayoutCircle(nil)
		if err != nil {
			t.Fatalf("LayoutCircle(nil) failed: %v", err)
		}
		rows, cols := coordsDefault.Dims()
		if rows != 5 || cols != 2 {
			t.Fatalf("got dims (%d, %d), want (5, 2)", rows, cols)
		}

		customOrder := []int{0, 2, 4, 1, 3}
		coordsCustom, err := g.LayoutCircle(customOrder)
		if err != nil {
			t.Fatalf("LayoutCircle(customOrder) failed: %v", err)
		}
		rowsC, colsC := coordsCustom.Dims()
		if rowsC != 5 || colsC != 2 {
			t.Fatalf("got dims (%d, %d), want (5, 2)", rowsC, colsC)
		}

		// The positions should reflect custom ordering
		valDef0, _ := coordsDefault.At(0, 0)
		valCust0, _ := coordsCustom.At(0, 0)
		if valDef0 != valCust0 {
			t.Errorf("vertex 0 (first in order) position differs: %v vs %v", valDef0, valCust0)
		}
	})

	t.Run("invalid order errors", func(t *testing.T) {
		g, err := igraph.NewRing(4, false, false)
		if err != nil {
			t.Fatalf("NewRing failed: %v", err)
		}
		defer g.Close()

		if _, err := g.LayoutCircle([]int{0, 1}); err == nil {
			t.Error("expected error for mismatched order length")
		}
		if _, err := g.LayoutCircle([]int{0, 1, 2, 99}); err == nil {
			t.Error("expected error for out of bounds vertex ID in order")
		}
		if _, err := g.LayoutCircle([]int{0, 1, 2, -1}); err == nil {
			t.Error("expected error for negative vertex ID in order")
		}
	})

	t.Run("closed and nil graph", func(t *testing.T) {
		var nilG *igraph.Graph
		if _, err := nilG.LayoutCircle(nil); err != igraph.ErrClosed {
			t.Errorf("got %v, want ErrClosed for nil graph", err)
		}

		g, _ := igraph.NewGraphFromEdges(3, nil, false)
		g.Close()
		if _, err := g.LayoutCircle(nil); err != igraph.ErrClosed {
			t.Errorf("got %v, want ErrClosed for closed graph", err)
		}
	})
}

func TestLayoutStar(t *testing.T) {
	t.Run("empty graph error", func(t *testing.T) {
		g, err := igraph.NewGraph()
		if err != nil {
			t.Fatalf("NewGraph failed: %v", err)
		}
		defer g.Close()

		if _, err := g.LayoutStar(0, nil); err == nil {
			t.Error("expected error for center vertex in empty graph")
		}
	})

	t.Run("star graph center layout with custom order and nil order", func(t *testing.T) {
		g, err := igraph.NewStar(5, 0, igraph.StarUndirected)
		if err != nil {
			t.Fatalf("NewStar failed: %v", err)
		}
		defer g.Close()

		coords, err := g.LayoutStar(0, nil)
		if err != nil {
			t.Fatalf("LayoutStar failed: %v", err)
		}
		rows, cols := coords.Dims()
		if rows != 5 || cols != 2 {
			t.Fatalf("got dims (%d, %d), want (5, 2)", rows, cols)
		}

		// Center vertex 0 should be at (0, 0)
		x0, _ := coords.At(0, 0)
		y0, _ := coords.At(0, 1)
		if x0 != 0 || y0 != 0 {
			t.Errorf("center vertex coords = (%v, %v), want (0, 0)", x0, y0)
		}

		customOrder := []int{0, 1, 2, 3, 4}
		coordsOrder, err := g.LayoutStar(0, customOrder)
		if err != nil {
			t.Fatalf("LayoutStar with custom order failed: %v", err)
		}
		if r, c := coordsOrder.Dims(); r != 5 || c != 2 {
			t.Fatalf("got dims (%d, %d), want (5, 2)", r, c)
		}
	})

	t.Run("invalid center or order", func(t *testing.T) {
		g, err := igraph.NewStar(4, 0, igraph.StarUndirected)
		if err != nil {
			t.Fatalf("NewStar failed: %v", err)
		}
		defer g.Close()

		if _, err := g.LayoutStar(-1, nil); err == nil {
			t.Error("expected error for negative center ID")
		}
		if _, err := g.LayoutStar(10, nil); err == nil {
			t.Error("expected error for out of bounds center ID")
		}
		if _, err := g.LayoutStar(0, []int{0, 1}); err == nil {
			t.Error("expected error for mismatched order length")
		}
		if _, err := g.LayoutStar(0, []int{0, 1, 2, 99}); err == nil {
			t.Error("expected error for out of bounds order ID")
		}
		if _, err := g.LayoutStar(0, []int{0, 1, -1, 3}); err == nil {
			t.Error("expected error for negative order ID")
		}
	})

	t.Run("closed and nil graph", func(t *testing.T) {
		var nilG *igraph.Graph
		if _, err := nilG.LayoutStar(0, nil); err != igraph.ErrClosed {
			t.Errorf("got %v, want ErrClosed for nil graph", err)
		}

		g, _ := igraph.NewStar(3, 0, igraph.StarUndirected)
		g.Close()
		if _, err := g.LayoutStar(0, nil); err != igraph.ErrClosed {
			t.Errorf("got %v, want ErrClosed for closed graph", err)
		}
	})
}

func TestLayoutGrid(t *testing.T) {
	t.Run("empty graph", func(t *testing.T) {
		g, err := igraph.NewGraph()
		if err != nil {
			t.Fatalf("NewGraph failed: %v", err)
		}
		defer g.Close()

		coords, err := g.LayoutGrid(2)
		if err != nil {
			t.Fatalf("LayoutGrid failed: %v", err)
		}
		rows, cols := coords.Dims()
		if rows != 0 || cols != 2 {
			t.Errorf("got dims (%d, %d), want (0, 2)", rows, cols)
		}
	})

	t.Run("grid layout width 0 and positive width", func(t *testing.T) {
		g, err := igraph.NewGraphFromEdges(9, nil, false)
		if err != nil {
			t.Fatalf("NewGraphFromEdges failed: %v", err)
		}
		defer g.Close()

		coords, err := g.LayoutGrid(3)
		if err != nil {
			t.Fatalf("LayoutGrid failed: %v", err)
		}
		rows, cols := coords.Dims()
		if rows != 9 || cols != 2 {
			t.Fatalf("got dims (%d, %d), want (9, 2)", rows, cols)
		}

		// width 0 lets upstream compute grid width automatically
		coordsAuto, err := g.LayoutGrid(0)
		if err != nil {
			t.Fatalf("LayoutGrid(0) failed: %v", err)
		}
		if r, c := coordsAuto.Dims(); r != 9 || c != 2 {
			t.Fatalf("got dims (%d, %d), want (9, 2)", r, c)
		}
	})

	t.Run("negative width error", func(t *testing.T) {
		g, err := igraph.NewGraphFromEdges(4, nil, false)
		if err != nil {
			t.Fatalf("NewGraphFromEdges failed: %v", err)
		}
		defer g.Close()

		if _, err := g.LayoutGrid(-1); err == nil {
			t.Error("expected error for negative width")
		}
	})

	t.Run("closed and nil graph", func(t *testing.T) {
		var nilG *igraph.Graph
		if _, err := nilG.LayoutGrid(2); err != igraph.ErrClosed {
			t.Errorf("got %v, want ErrClosed for nil graph", err)
		}

		g, _ := igraph.NewGraphFromEdges(3, nil, false)
		g.Close()
		if _, err := g.LayoutGrid(2); err != igraph.ErrClosed {
			t.Errorf("got %v, want ErrClosed for closed graph", err)
		}
	})
}

func TestLayoutRandom(t *testing.T) {
	t.Run("empty graph", func(t *testing.T) {
		g, err := igraph.NewGraph()
		if err != nil {
			t.Fatalf("NewGraph failed: %v", err)
		}
		defer g.Close()

		coords, err := g.LayoutRandom(igraph.LayoutRandomOptions{})
		if err != nil {
			t.Fatalf("LayoutRandom failed: %v", err)
		}
		rows, cols := coords.Dims()
		if rows != 0 || cols != 2 {
			t.Errorf("got dims (%d, %d), want (0, 2)", rows, cols)
		}
	})

	t.Run("unseeded random layout", func(t *testing.T) {
		g, err := igraph.NewRing(5, false, false)
		if err != nil {
			t.Fatalf("NewRing failed: %v", err)
		}
		defer g.Close()

		coords, err := g.LayoutRandom(igraph.LayoutRandomOptions{})
		if err != nil {
			t.Fatalf("LayoutRandom unseeded failed: %v", err)
		}
		if r, c := coords.Dims(); r != 5 || c != 2 {
			t.Fatalf("got dims (%d, %d), want (5, 2)", r, c)
		}
	})

	t.Run("reproducibility with seed", func(t *testing.T) {
		g, err := igraph.NewRing(10, false, false)
		if err != nil {
			t.Fatalf("NewRing failed: %v", err)
		}
		defer g.Close()

		seed := uint64(42)
		coords1, err := g.LayoutRandom(igraph.LayoutRandomOptions{Seed: &seed})
		if err != nil {
			t.Fatalf("LayoutRandom 1 failed: %v", err)
		}

		coords2, err := g.LayoutRandom(igraph.LayoutRandomOptions{Seed: &seed})
		if err != nil {
			t.Fatalf("LayoutRandom 2 failed: %v", err)
		}

		r1, c1 := coords1.Dims()
		r2, c2 := coords2.Dims()
		if r1 != 10 || c1 != 2 || r2 != 10 || c2 != 2 {
			t.Fatalf("unexpected dimensions: (%d, %d) vs (%d, %d)", r1, c1, r2, c2)
		}

		assertEqualMatrices(t, "LayoutRandom", coords1, coords2)
	})

	t.Run("concurrent seed isolation", func(t *testing.T) {
		g, err := igraph.NewRing(10, false, false)
		if err != nil {
			t.Fatalf("NewRing failed: %v", err)
		}
		defer g.Close()

		var wg sync.WaitGroup
		for i := 0; i < 5; i++ {
			wg.Add(1)
			go func(seedVal uint64) {
				defer wg.Done()
				seed := seedVal
				_, err := g.LayoutRandom(igraph.LayoutRandomOptions{Seed: &seed})
				if err != nil {
					t.Errorf("concurrent LayoutRandom failed: %v", err)
				}
			}(uint64(100 + i))
		}
		wg.Wait()
	})

	t.Run("closed and nil graph", func(t *testing.T) {
		var nilG *igraph.Graph
		if _, err := nilG.LayoutRandom(igraph.LayoutRandomOptions{}); err != igraph.ErrClosed {
			t.Errorf("got %v, want ErrClosed for nil graph", err)
		}

		g, _ := igraph.NewGraphFromEdges(3, nil, false)
		g.Close()
		if _, err := g.LayoutRandom(igraph.LayoutRandomOptions{}); err != igraph.ErrClosed {
			t.Errorf("got %v, want ErrClosed for closed graph", err)
		}
	})
}

func TestLayoutReingoldTilford(t *testing.T) {
	t.Run("tree graph layout with nil and custom roots", func(t *testing.T) {
		g, err := igraph.NewKaryTree(7, 2, igraph.TreeOut)
		if err != nil {
			t.Fatalf("NewKaryTree failed: %v", err)
		}
		defer g.Close()

		coords, err := g.LayoutReingoldTilford(igraph.DegOut, nil)
		if err != nil {
			t.Fatalf("LayoutReingoldTilford(nil) failed: %v", err)
		}
		if r, c := coords.Dims(); r != 7 || c != 2 {
			t.Fatalf("got dims (%d, %d), want (7, 2)", r, c)
		}

		coordsRoots, err := g.LayoutReingoldTilford(igraph.DegOut, []int{0})
		if err != nil {
			t.Fatalf("LayoutReingoldTilford([]int{0}) failed: %v", err)
		}
		if r, c := coordsRoots.Dims(); r != 7 || c != 2 {
			t.Fatalf("got dims (%d, %d), want (7, 2)", r, c)
		}
	})

	t.Run("invalid root ID or mode", func(t *testing.T) {
		g, err := igraph.NewKaryTree(5, 2, igraph.TreeOut)
		if err != nil {
			t.Fatalf("NewKaryTree failed: %v", err)
		}
		defer g.Close()

		if _, err := g.LayoutReingoldTilford(igraph.DegOut, []int{99}); err == nil {
			t.Error("expected error for out of bounds root ID")
		}
		if _, err := g.LayoutReingoldTilford(igraph.DegOut, []int{-1}); err == nil {
			t.Error("expected error for negative root ID")
		}
		if _, err := g.LayoutReingoldTilford(igraph.DegMode(99), nil); err == nil {
			t.Error("expected error for invalid DegMode")
		}
	})

	t.Run("closed and nil graph", func(t *testing.T) {
		var nilG *igraph.Graph
		if _, err := nilG.LayoutReingoldTilford(igraph.DegOut, nil); err != igraph.ErrClosed {
			t.Errorf("got %v, want ErrClosed for nil graph", err)
		}

		g, _ := igraph.NewGraphFromEdges(3, nil, false)
		g.Close()
		if _, err := g.LayoutReingoldTilford(igraph.DegOut, nil); err != igraph.ErrClosed {
			t.Errorf("got %v, want ErrClosed for closed graph", err)
		}
	})
}

func TestLayoutReingoldTilfordCircular(t *testing.T) {
	t.Run("tree graph circular layout with nil and custom roots", func(t *testing.T) {
		g, err := igraph.NewKaryTree(7, 2, igraph.TreeOut)
		if err != nil {
			t.Fatalf("NewKaryTree failed: %v", err)
		}
		defer g.Close()

		coords, err := g.LayoutReingoldTilfordCircular(igraph.DegOut, nil)
		if err != nil {
			t.Fatalf("LayoutReingoldTilfordCircular(nil) failed: %v", err)
		}
		if r, c := coords.Dims(); r != 7 || c != 2 {
			t.Fatalf("got dims (%d, %d), want (7, 2)", r, c)
		}

		coordsRoots, err := g.LayoutReingoldTilfordCircular(igraph.DegOut, []int{0})
		if err != nil {
			t.Fatalf("LayoutReingoldTilfordCircular([]int{0}) failed: %v", err)
		}
		if r, c := coordsRoots.Dims(); r != 7 || c != 2 {
			t.Fatalf("got dims (%d, %d), want (7, 2)", r, c)
		}
	})

	t.Run("invalid root ID or mode", func(t *testing.T) {
		g, err := igraph.NewKaryTree(5, 2, igraph.TreeOut)
		if err != nil {
			t.Fatalf("NewKaryTree failed: %v", err)
		}
		defer g.Close()

		if _, err := g.LayoutReingoldTilfordCircular(igraph.DegOut, []int{99}); err == nil {
			t.Error("expected error for out of bounds root ID")
		}
		if _, err := g.LayoutReingoldTilfordCircular(igraph.DegOut, []int{-1}); err == nil {
			t.Error("expected error for negative root ID")
		}
		if _, err := g.LayoutReingoldTilfordCircular(igraph.DegMode(99), nil); err == nil {
			t.Error("expected error for invalid DegMode")
		}
	})

	t.Run("closed and nil graph", func(t *testing.T) {
		var nilG *igraph.Graph
		if _, err := nilG.LayoutReingoldTilfordCircular(igraph.DegOut, nil); err != igraph.ErrClosed {
			t.Errorf("got %v, want ErrClosed for nil graph", err)
		}

		g, _ := igraph.NewGraphFromEdges(3, nil, false)
		g.Close()
		if _, err := g.LayoutReingoldTilfordCircular(igraph.DegOut, nil); err != igraph.ErrClosed {
			t.Errorf("got %v, want ErrClosed for closed graph", err)
		}
	})
}

func TestLayoutBipartite(t *testing.T) {
	t.Run("bipartite graph layout and empty graph", func(t *testing.T) {
		g, err := igraph.NewGraphFromEdges(4, []igraph.Edge{
			{From: 0, To: 2},
			{From: 0, To: 3},
			{From: 1, To: 2},
			{From: 1, To: 3},
		}, false)
		if err != nil {
			t.Fatalf("NewGraphFromEdges failed: %v", err)
		}
		defer g.Close()

		types := []bool{false, false, true, true}
		coords, err := g.LayoutBipartite(types, 1.0, 1.0, 100)
		if err != nil {
			t.Fatalf("LayoutBipartite failed: %v", err)
		}
		if r, c := coords.Dims(); r != 4 || c != 2 {
			t.Fatalf("got dims (%d, %d), want (4, 2)", r, c)
		}

		emptyG, _ := igraph.NewGraph()
		defer emptyG.Close()
		coordsEmpty, err := emptyG.LayoutBipartite([]bool{}, 1.0, 1.0, 100)
		if err != nil {
			t.Fatalf("LayoutBipartite empty failed: %v", err)
		}
		if r, c := coordsEmpty.Dims(); r != 0 || c != 2 {
			t.Fatalf("got dims (%d, %d), want (0, 2)", r, c)
		}
	})

	t.Run("invalid parameters", func(t *testing.T) {
		g, err := igraph.NewGraphFromEdges(4, nil, false)
		if err != nil {
			t.Fatalf("NewGraphFromEdges failed: %v", err)
		}
		defer g.Close()

		if _, err := g.LayoutBipartite([]bool{true}, 1.0, 1.0, 100); err == nil {
			t.Error("expected error for mismatched types length")
		}
		if _, err := g.LayoutBipartite([]bool{true, false, true, false}, 1.0, 1.0, -1); err == nil {
			t.Error("expected error for negative maxIter")
		}
	})

	t.Run("closed and nil graph", func(t *testing.T) {
		var nilG *igraph.Graph
		if _, err := nilG.LayoutBipartite([]bool{}, 1.0, 1.0, 100); err != igraph.ErrClosed {
			t.Errorf("got %v, want ErrClosed for nil graph", err)
		}

		g, _ := igraph.NewGraphFromEdges(3, nil, false)
		g.Close()
		if _, err := g.LayoutBipartite([]bool{true, false, true}, 1.0, 1.0, 100); err != igraph.ErrClosed {
			t.Errorf("got %v, want ErrClosed for closed graph", err)
		}
	})
}

func TestLayoutSugiyama(t *testing.T) {
	t.Run("sugiyama layout unweighted and weighted", func(t *testing.T) {
		g, err := igraph.NewGraphFromEdges(3, []igraph.Edge{
			{From: 0, To: 1},
			{From: 1, To: 2},
			{From: 0, To: 2},
		}, true)
		if err != nil {
			t.Fatalf("NewGraphFromEdges failed: %v", err)
		}
		defer g.Close()

		coords, err := g.LayoutSugiyama(nil, igraph.SugiyamaOptions{})
		if err != nil {
			t.Fatalf("LayoutSugiyama failed: %v", err)
		}
		if r, c := coords.Dims(); r != 3 || c != 2 {
			t.Fatalf("got dims (%d, %d), want (3, 2)", r, c)
		}

		layers := []int{0, 1, 2}
		weights := []float64{1.0, 2.0, 1.5}
		coordsCustom, err := g.LayoutSugiyama(layers, igraph.SugiyamaOptions{
			HGap:    2.0,
			VGap:    2.0,
			MaxIter: 50,
			Weights: weights,
		})
		if err != nil {
			t.Fatalf("LayoutSugiyama with options failed: %v", err)
		}
		if r, c := coordsCustom.Dims(); r != 3 || c != 2 {
			t.Fatalf("got dims (%d, %d), want (3, 2)", r, c)
		}
	})

	t.Run("invalid parameters", func(t *testing.T) {
		g, err := igraph.NewGraphFromEdges(3, []igraph.Edge{
			{From: 0, To: 1},
		}, true)
		if err != nil {
			t.Fatalf("NewGraphFromEdges failed: %v", err)
		}
		defer g.Close()

		if _, err := g.LayoutSugiyama([]int{0}, igraph.SugiyamaOptions{}); err == nil {
			t.Error("expected error for mismatched layers length")
		}
		if _, err := g.LayoutSugiyama([]int{0, -1, 1}, igraph.SugiyamaOptions{}); err == nil {
			t.Error("expected error for negative layer value")
		}
		if _, err := g.LayoutSugiyama(nil, igraph.SugiyamaOptions{MaxIter: -1}); err == nil {
			t.Error("expected error for negative maxIter")
		}
		if _, err := g.LayoutSugiyama(nil, igraph.SugiyamaOptions{Weights: []float64{1.0, 2.0}}); err == nil {
			t.Error("expected error for mismatched weights length")
		}
	})

	t.Run("closed and nil graph", func(t *testing.T) {
		var nilG *igraph.Graph
		if _, err := nilG.LayoutSugiyama(nil, igraph.SugiyamaOptions{}); err != igraph.ErrClosed {
			t.Errorf("got %v, want ErrClosed for nil graph", err)
		}

		g, _ := igraph.NewGraphFromEdges(3, nil, false)
		g.Close()
		if _, err := g.LayoutSugiyama(nil, igraph.SugiyamaOptions{}); err != igraph.ErrClosed {
			t.Errorf("got %v, want ErrClosed for closed graph", err)
		}
	})
}

func TestLayoutFruchtermanReingold(t *testing.T) {
	t.Run("unweighted and weighted with initial coordinates and bounds", func(t *testing.T) {
		g, err := igraph.NewRing(5, false, false)
		if err != nil {
			t.Fatalf("NewRing failed: %v", err)
		}
		defer g.Close()

		seed := uint64(123)
		coords, err := g.LayoutFruchtermanReingold(igraph.FruchtermanReingoldOptions{
			Seed:  &seed,
			NIter: 100,
		})
		if err != nil {
			t.Fatalf("LayoutFruchtermanReingold failed: %v", err)
		}
		if r, c := coords.Dims(); r != 5 || c != 2 {
			t.Fatalf("got dims (%d, %d), want (5, 2)", r, c)
		}

		// Initial coordinates reuse
		initCoords, _ := igraph.NewMatrixFromRows([][]float64{
			{0, 0}, {1, 1}, {2, 2}, {3, 3}, {4, 4},
		})
		minX := []float64{-10, -10, -10, -10, -10}
		maxX := []float64{10, 10, 10, 10, 10}
		minY := []float64{-10, -10, -10, -10, -10}
		maxY := []float64{10, 10, 10, 10, 10}
		weights := []float64{1.0, 1.0, 1.0, 1.0, 1.0}

		coordsInit, err := g.LayoutFruchtermanReingold(igraph.FruchtermanReingoldOptions{
			Seed:               &seed,
			NIter:              50,
			Weights:            weights,
			InitialCoordinates: &initCoords,
			MinX:               minX,
			MaxX:               maxX,
			MinY:               minY,
			MaxY:               maxY,
		})
		if err != nil {
			t.Fatalf("LayoutFruchtermanReingold with initial coords failed: %v", err)
		}
		if r, c := coordsInit.Dims(); r != 5 || c != 2 {
			t.Fatalf("got dims (%d, %d), want (5, 2)", r, c)
		}
		for r := 0; r < 5; r++ {
			for c := 0; c < 2; c++ {
				v, _ := coordsInit.At(r, c)
				if v < -10 || v > 10 {
					t.Errorf("coordinate (%d, %d) = %v escapes bounds [-10, 10]", r, c, v)
				}
			}
		}

		// Defaults (NIter 500, StartTemp sqrt(V)) must actually run the
		// algorithm: starting from identical coordinates, the layout must
		// move away from its initial placement.
		coordsDefault, err := g.LayoutFruchtermanReingold(igraph.FruchtermanReingoldOptions{
			Seed:               &seed,
			InitialCoordinates: &initCoords,
		})
		if err != nil {
			t.Fatalf("LayoutFruchtermanReingold with default options failed: %v", err)
		}
		moved := false
		for r := 0; r < 5; r++ {
			for c := 0; c < 2; c++ {
				before, _ := initCoords.At(r, c)
				after, _ := coordsDefault.At(r, c)
				if before != after {
					moved = true
				}
			}
		}
		if !moved {
			t.Error("default options left every vertex at its initial position; the algorithm did not run")
		}
	})

	t.Run("seed reproducibility and concurrent isolation", func(t *testing.T) {
		g, err := igraph.NewRing(10, false, false)
		if err != nil {
			t.Fatalf("NewRing failed: %v", err)
		}
		defer g.Close()

		seed := uint64(999)
		c1, err := g.LayoutFruchtermanReingold(igraph.FruchtermanReingoldOptions{Seed: &seed, NIter: 50})
		if err != nil {
			t.Fatalf("Layout 1 failed: %v", err)
		}
		c2, err := g.LayoutFruchtermanReingold(igraph.FruchtermanReingoldOptions{Seed: &seed, NIter: 50})
		if err != nil {
			t.Fatalf("Layout 2 failed: %v", err)
		}

		assertEqualMatrices(t, "seeded layout", c1, c2)

		var wg sync.WaitGroup
		for i := 0; i < 5; i++ {
			wg.Add(1)
			go func(seedVal uint64) {
				defer wg.Done()
				s := seedVal
				_, err := g.LayoutFruchtermanReingold(igraph.FruchtermanReingoldOptions{Seed: &s, NIter: 20})
				if err != nil {
					t.Errorf("concurrent LayoutFruchtermanReingold failed: %v", err)
				}
			}(uint64(50 + i))
		}
		wg.Wait()
	})

	t.Run("invalid parameters", func(t *testing.T) {
		g, err := igraph.NewRing(4, false, false)
		if err != nil {
			t.Fatalf("NewRing failed: %v", err)
		}
		defer g.Close()

		if _, err := g.LayoutFruchtermanReingold(igraph.FruchtermanReingoldOptions{NIter: -1}); err == nil {
			t.Error("expected error for negative NIter")
		}
		if _, err := g.LayoutFruchtermanReingold(igraph.FruchtermanReingoldOptions{Weights: []float64{1.0}}); err == nil {
			t.Error("expected error for mismatched Weights length")
		}
		if _, err := g.LayoutFruchtermanReingold(igraph.FruchtermanReingoldOptions{Weights: []float64{math.NaN(), 1, 1, 1}}); err == nil {
			t.Error("expected error for NaN Weight")
		}
		if _, err := g.LayoutFruchtermanReingold(igraph.FruchtermanReingoldOptions{MinX: []float64{1.0}}); err == nil {
			t.Error("expected error for mismatched MinX length")
		}
		if _, err := g.LayoutFruchtermanReingold(igraph.FruchtermanReingoldOptions{MaxX: []float64{1.0}}); err == nil {
			t.Error("expected error for mismatched MaxX length")
		}
		if _, err := g.LayoutFruchtermanReingold(igraph.FruchtermanReingoldOptions{MinY: []float64{1.0}}); err == nil {
			t.Error("expected error for mismatched MinY length")
		}
		if _, err := g.LayoutFruchtermanReingold(igraph.FruchtermanReingoldOptions{MaxY: []float64{1.0}}); err == nil {
			t.Error("expected error for mismatched MaxY length")
		}
		badInit, _ := igraph.NewMatrixFromRows([][]float64{{0, 0}, {1, 1}})
		if _, err := g.LayoutFruchtermanReingold(igraph.FruchtermanReingoldOptions{InitialCoordinates: &badInit}); err == nil {
			t.Error("expected error for mismatched InitialCoordinates dimensions")
		}
		if _, err := g.LayoutFruchtermanReingold(igraph.FruchtermanReingoldOptions{
			MinX: []float64{10, 10, 10, 10},
			MaxX: []float64{-10, -10, -10, -10},
		}); err == nil {
			t.Error("expected error for MinX greater than MaxX")
		}
		if _, err := g.LayoutFruchtermanReingold(igraph.FruchtermanReingoldOptions{
			MinX: []float64{math.NaN(), 0, 0, 0},
		}); err == nil {
			t.Error("expected error for NaN MinX")
		}
		if _, err := g.LayoutFruchtermanReingold(igraph.FruchtermanReingoldOptions{StartTemp: -1}); err == nil {
			t.Error("expected error for negative StartTemp")
		}
	})

	t.Run("closed and nil graph", func(t *testing.T) {
		var nilG *igraph.Graph
		if _, err := nilG.LayoutFruchtermanReingold(igraph.FruchtermanReingoldOptions{}); err != igraph.ErrClosed {
			t.Errorf("got %v, want ErrClosed for nil graph", err)
		}

		g, _ := igraph.NewGraphFromEdges(3, nil, false)
		g.Close()
		if _, err := g.LayoutFruchtermanReingold(igraph.FruchtermanReingoldOptions{}); err != igraph.ErrClosed {
			t.Errorf("got %v, want ErrClosed for closed graph", err)
		}
	})
}

func TestLayoutKamadaKawai(t *testing.T) {
	t.Run("unweighted and weighted with initial coordinates and bounds", func(t *testing.T) {
		g, err := igraph.NewRing(5, false, false)
		if err != nil {
			t.Fatalf("NewRing failed: %v", err)
		}
		defer g.Close()

		seed := uint64(123)
		coords, err := g.LayoutKamadaKawai(igraph.KamadaKawaiOptions{
			Seed:    &seed,
			MaxIter: 100,
		})
		if err != nil {
			t.Fatalf("LayoutKamadaKawai failed: %v", err)
		}
		if r, c := coords.Dims(); r != 5 || c != 2 {
			t.Fatalf("got dims (%d, %d), want (5, 2)", r, c)
		}

		initCoords, _ := igraph.NewMatrixFromRows([][]float64{
			{0, 0}, {1, 1}, {2, 2}, {3, 3}, {4, 4},
		})
		minX := []float64{-10, -10, -10, -10, -10}
		maxX := []float64{10, 10, 10, 10, 10}
		minY := []float64{-10, -10, -10, -10, -10}
		maxY := []float64{10, 10, 10, 10, 10}
		weights := []float64{1.0, 1.0, 1.0, 1.0, 1.0}

		coordsInit, err := g.LayoutKamadaKawai(igraph.KamadaKawaiOptions{
			Seed:               &seed,
			MaxIter:            50,
			Epsilon:            0.001,
			KKConst:            1.0,
			Weights:            weights,
			InitialCoordinates: &initCoords,
			MinX:               minX,
			MaxX:               maxX,
			MinY:               minY,
			MaxY:               maxY,
		})
		if err != nil {
			t.Fatalf("LayoutKamadaKawai with initial coords failed: %v", err)
		}
		if r, c := coordsInit.Dims(); r != 5 || c != 2 {
			t.Fatalf("got dims (%d, %d), want (5, 2)", r, c)
		}
		for r := 0; r < 5; r++ {
			for c := 0; c < 2; c++ {
				v, _ := coordsInit.At(r, c)
				if v < -10 || v > 10 {
					t.Errorf("coordinate (%d, %d) = %v escapes bounds [-10, 10]", r, c, v)
				}
			}
		}
	})

	t.Run("seed reproducibility and concurrent isolation", func(t *testing.T) {
		g, err := igraph.NewRing(10, false, false)
		if err != nil {
			t.Fatalf("NewRing failed: %v", err)
		}
		defer g.Close()

		seed := uint64(888)
		c1, err := g.LayoutKamadaKawai(igraph.KamadaKawaiOptions{Seed: &seed, MaxIter: 50})
		if err != nil {
			t.Fatalf("Layout 1 failed: %v", err)
		}
		c2, err := g.LayoutKamadaKawai(igraph.KamadaKawaiOptions{Seed: &seed, MaxIter: 50})
		if err != nil {
			t.Fatalf("Layout 2 failed: %v", err)
		}

		assertEqualMatrices(t, "seeded layout", c1, c2)

		var wg sync.WaitGroup
		for i := 0; i < 5; i++ {
			wg.Add(1)
			go func(seedVal uint64) {
				defer wg.Done()
				s := seedVal
				_, err := g.LayoutKamadaKawai(igraph.KamadaKawaiOptions{Seed: &s, MaxIter: 20})
				if err != nil {
					t.Errorf("concurrent LayoutKamadaKawai failed: %v", err)
				}
			}(uint64(50 + i))
		}
		wg.Wait()
	})

	t.Run("invalid parameters", func(t *testing.T) {
		g, err := igraph.NewRing(4, false, false)
		if err != nil {
			t.Fatalf("NewRing failed: %v", err)
		}
		defer g.Close()

		if _, err := g.LayoutKamadaKawai(igraph.KamadaKawaiOptions{MaxIter: -1}); err == nil {
			t.Error("expected error for negative MaxIter")
		}
		if _, err := g.LayoutKamadaKawai(igraph.KamadaKawaiOptions{Weights: []float64{1.0}}); err == nil {
			t.Error("expected error for mismatched Weights length")
		}
		if _, err := g.LayoutKamadaKawai(igraph.KamadaKawaiOptions{Weights: []float64{math.NaN(), 1, 1, 1}}); err == nil {
			t.Error("expected error for NaN Weight")
		}
		if _, err := g.LayoutKamadaKawai(igraph.KamadaKawaiOptions{MinX: []float64{1.0}}); err == nil {
			t.Error("expected error for mismatched MinX length")
		}
		if _, err := g.LayoutKamadaKawai(igraph.KamadaKawaiOptions{MaxX: []float64{1.0}}); err == nil {
			t.Error("expected error for mismatched MaxX length")
		}
		if _, err := g.LayoutKamadaKawai(igraph.KamadaKawaiOptions{MinY: []float64{1.0}}); err == nil {
			t.Error("expected error for mismatched MinY length")
		}
		if _, err := g.LayoutKamadaKawai(igraph.KamadaKawaiOptions{MaxY: []float64{1.0}}); err == nil {
			t.Error("expected error for mismatched MaxY length")
		}
		badInit, _ := igraph.NewMatrixFromRows([][]float64{{0, 0}, {1, 1}})
		if _, err := g.LayoutKamadaKawai(igraph.KamadaKawaiOptions{InitialCoordinates: &badInit}); err == nil {
			t.Error("expected error for mismatched InitialCoordinates dimensions")
		}
		if _, err := g.LayoutKamadaKawai(igraph.KamadaKawaiOptions{
			MinX: []float64{10, 10, 10, 10},
			MaxX: []float64{-10, -10, -10, -10},
		}); err == nil {
			t.Error("expected error for MinX greater than MaxX")
		}
		if _, err := g.LayoutKamadaKawai(igraph.KamadaKawaiOptions{
			MaxY: []float64{0, 0, math.NaN(), 0},
		}); err == nil {
			t.Error("expected error for NaN MaxY")
		}
		if _, err := g.LayoutKamadaKawai(igraph.KamadaKawaiOptions{Epsilon: -0.5}); err == nil {
			t.Error("expected error for negative Epsilon")
		}
		if _, err := g.LayoutKamadaKawai(igraph.KamadaKawaiOptions{KKConst: -5}); err == nil {
			t.Error("expected error for negative KKConst")
		}
	})

	t.Run("empty graph", func(t *testing.T) {
		g, err := igraph.NewGraphFromEdges(0, nil, false)
		if err != nil {
			t.Fatalf("NewGraphFromEdges failed: %v", err)
		}
		defer g.Close()

		coords, err := g.LayoutKamadaKawai(igraph.KamadaKawaiOptions{})
		if err != nil {
			t.Fatalf("LayoutKamadaKawai on empty graph failed: %v", err)
		}
		if r, _ := coords.Dims(); r != 0 {
			t.Errorf("got %d rows, want 0", r)
		}
	})

	t.Run("closed and nil graph", func(t *testing.T) {
		var nilG *igraph.Graph
		if _, err := nilG.LayoutKamadaKawai(igraph.KamadaKawaiOptions{}); err != igraph.ErrClosed {
			t.Errorf("got %v, want ErrClosed for nil graph", err)
		}

		g, _ := igraph.NewGraphFromEdges(3, nil, false)
		g.Close()
		if _, err := g.LayoutKamadaKawai(igraph.KamadaKawaiOptions{}); err != igraph.ErrClosed {
			t.Errorf("got %v, want ErrClosed for closed graph", err)
		}
	})
}

func TestLayoutMDS(t *testing.T) {
	t.Run("auto distances and custom distance matrix 2D and 3D", func(t *testing.T) {
		g, err := igraph.NewRing(4, false, false)
		if err != nil {
			t.Fatalf("NewRing failed: %v", err)
		}
		defer g.Close()

		coords2D, err := g.LayoutMDS(nil, 2, igraph.MDSOptions{})
		if err != nil {
			t.Fatalf("LayoutMDS 2D auto failed: %v", err)
		}
		if r, c := coords2D.Dims(); r != 4 || c != 2 {
			t.Fatalf("got dims (%d, %d), want (4, 2)", r, c)
		}

		coords3D, err := g.LayoutMDS(nil, 3, igraph.MDSOptions{})
		if err != nil {
			t.Fatalf("LayoutMDS 3D auto failed: %v", err)
		}
		if r, c := coords3D.Dims(); r != 4 || c != 3 {
			t.Fatalf("got dims (%d, %d), want (4, 3)", r, c)
		}

		// Custom distance matrix
		distMat, _ := igraph.NewMatrixFromRows([][]float64{
			{0, 1, 2, 1},
			{1, 0, 1, 2},
			{2, 1, 0, 1},
			{1, 2, 1, 0},
		})
		coordsCustom, err := g.LayoutMDS(&distMat, 2, igraph.MDSOptions{})
		if err != nil {
			t.Fatalf("LayoutMDS custom distances failed: %v", err)
		}
		if r, c := coordsCustom.Dims(); r != 4 || c != 2 {
			t.Fatalf("got dims (%d, %d), want (4, 2)", r, c)
		}

		// dim <= 0 defaults to 2
		coordsDefault, err := g.LayoutMDS(nil, 0, igraph.MDSOptions{})
		if err != nil {
			t.Fatalf("LayoutMDS default dim failed: %v", err)
		}
		if r, c := coordsDefault.Dims(); r != 4 || c != 2 {
			t.Fatalf("got dims (%d, %d), want (4, 2)", r, c)
		}
	})

	t.Run("invalid parameters", func(t *testing.T) {
		g, err := igraph.NewRing(4, false, false)
		if err != nil {
			t.Fatalf("NewRing failed: %v", err)
		}
		defer g.Close()

		if _, err := g.LayoutMDS(nil, 1, igraph.MDSOptions{}); err == nil {
			t.Error("expected error for dimension 1")
		}
		if _, err := g.LayoutMDS(nil, 4, igraph.MDSOptions{}); err == nil {
			t.Error("expected error for dimension 4")
		}

		badDist, _ := igraph.NewMatrixFromRows([][]float64{{0, 1}, {1, 0}})
		if _, err := g.LayoutMDS(&badDist, 2, igraph.MDSOptions{}); err == nil {
			t.Error("expected error for mismatched distance matrix dimensions")
		}

		// Upstream leaves the result for asymmetric distances unspecified,
		// so the binding rejects them at the boundary.
		asymDist, _ := igraph.NewMatrixFromRows([][]float64{
			{0, 1, 2, 3},
			{9, 0, 1, 2},
			{2, 1, 0, 1},
			{3, 2, 1, 0},
		})
		if _, err := g.LayoutMDS(&asymDist, 2, igraph.MDSOptions{}); err == nil {
			t.Error("expected error for asymmetric distance matrix")
		}
	})

	t.Run("closed and nil graph", func(t *testing.T) {
		var nilG *igraph.Graph
		if _, err := nilG.LayoutMDS(nil, 2, igraph.MDSOptions{}); err != igraph.ErrClosed {
			t.Errorf("got %v, want ErrClosed for nil graph", err)
		}

		g, _ := igraph.NewGraphFromEdges(3, nil, false)
		g.Close()
		if _, err := g.LayoutMDS(nil, 2, igraph.MDSOptions{}); err != igraph.ErrClosed {
			t.Errorf("got %v, want ErrClosed for closed graph", err)
		}
	})
}

func TestLayoutRandom3D(t *testing.T) {
	t.Run("dimensions and seed reproducibility", func(t *testing.T) {
		g, err := igraph.NewRing(6, false, false)
		if err != nil {
			t.Fatalf("NewRing failed: %v", err)
		}
		defer g.Close()

		seed := uint64(42)
		c1, err := g.LayoutRandom3D(igraph.LayoutRandomOptions{Seed: &seed})
		if err != nil {
			t.Fatalf("LayoutRandom3D failed: %v", err)
		}
		if r, c := c1.Dims(); r != 6 || c != 3 {
			t.Fatalf("got dims (%d, %d), want (6, 3)", r, c)
		}
		c2, err := g.LayoutRandom3D(igraph.LayoutRandomOptions{Seed: &seed})
		if err != nil {
			t.Fatalf("LayoutRandom3D second call failed: %v", err)
		}
		assertEqualMatrices(t, "seeded 3D layout", c1, c2)
	})

	t.Run("closed and nil graph", func(t *testing.T) {
		var nilG *igraph.Graph
		if _, err := nilG.LayoutRandom3D(igraph.LayoutRandomOptions{}); err != igraph.ErrClosed {
			t.Errorf("got %v, want ErrClosed for nil graph", err)
		}
		g, _ := igraph.NewGraphFromEdges(3, nil, false)
		g.Close()
		if _, err := g.LayoutRandom3D(igraph.LayoutRandomOptions{}); err != igraph.ErrClosed {
			t.Errorf("got %v, want ErrClosed for closed graph", err)
		}
	})
}

func TestLayoutGrid3D(t *testing.T) {
	t.Run("explicit and automatic extents", func(t *testing.T) {
		g, err := igraph.NewRing(8, false, false)
		if err != nil {
			t.Fatalf("NewRing failed: %v", err)
		}
		defer g.Close()

		coords, err := g.LayoutGrid3D(2, 2)
		if err != nil {
			t.Fatalf("LayoutGrid3D failed: %v", err)
		}
		if r, c := coords.Dims(); r != 8 || c != 3 {
			t.Fatalf("got dims (%d, %d), want (8, 3)", r, c)
		}
		// Two full 2x2 layers: vertex 0 and vertex 4 share x/y but differ in z.
		z0, _ := coords.At(0, 2)
		z4, _ := coords.At(4, 2)
		if z0 == z4 {
			t.Errorf("vertices 0 and 4 should be on different layers, both at z=%v", z0)
		}

		auto, err := g.LayoutGrid3D(0, 0)
		if err != nil {
			t.Fatalf("LayoutGrid3D auto failed: %v", err)
		}
		if r, c := auto.Dims(); r != 8 || c != 3 {
			t.Fatalf("got dims (%d, %d), want (8, 3)", r, c)
		}
	})

	t.Run("invalid parameters", func(t *testing.T) {
		g, err := igraph.NewRing(4, false, false)
		if err != nil {
			t.Fatalf("NewRing failed: %v", err)
		}
		defer g.Close()

		if _, err := g.LayoutGrid3D(-1, 2); err == nil {
			t.Error("expected error for negative width")
		}
		if _, err := g.LayoutGrid3D(2, -1); err == nil {
			t.Error("expected error for negative height")
		}
	})

	t.Run("closed and nil graph", func(t *testing.T) {
		var nilG *igraph.Graph
		if _, err := nilG.LayoutGrid3D(2, 2); err != igraph.ErrClosed {
			t.Errorf("got %v, want ErrClosed for nil graph", err)
		}
		g, _ := igraph.NewGraphFromEdges(3, nil, false)
		g.Close()
		if _, err := g.LayoutGrid3D(2, 2); err != igraph.ErrClosed {
			t.Errorf("got %v, want ErrClosed for closed graph", err)
		}
	})
}

func TestLayoutSphere(t *testing.T) {
	t.Run("unit sphere placement", func(t *testing.T) {
		g, err := igraph.NewRing(10, false, false)
		if err != nil {
			t.Fatalf("NewRing failed: %v", err)
		}
		defer g.Close()

		coords, err := g.LayoutSphere()
		if err != nil {
			t.Fatalf("LayoutSphere failed: %v", err)
		}
		if r, c := coords.Dims(); r != 10 || c != 3 {
			t.Fatalf("got dims (%d, %d), want (10, 3)", r, c)
		}
		for r := 0; r < 10; r++ {
			x, _ := coords.At(r, 0)
			y, _ := coords.At(r, 1)
			z, _ := coords.At(r, 2)
			radius := math.Sqrt(x*x + y*y + z*z)
			if math.Abs(radius-1) > 1e-9 {
				t.Errorf("vertex %d at radius %v, want 1", r, radius)
			}
		}

		again, err := g.LayoutSphere()
		if err != nil {
			t.Fatalf("LayoutSphere second call failed: %v", err)
		}
		assertEqualMatrices(t, "LayoutSphere", coords, again)
	})

	t.Run("closed and nil graph", func(t *testing.T) {
		var nilG *igraph.Graph
		if _, err := nilG.LayoutSphere(); err != igraph.ErrClosed {
			t.Errorf("got %v, want ErrClosed for nil graph", err)
		}
		g, _ := igraph.NewGraphFromEdges(3, nil, false)
		g.Close()
		if _, err := g.LayoutSphere(); err != igraph.ErrClosed {
			t.Errorf("got %v, want ErrClosed for closed graph", err)
		}
	})
}

func TestLayoutFruchtermanReingold3D(t *testing.T) {
	t.Run("dimensions, bounds, and initial coordinates", func(t *testing.T) {
		g, err := igraph.NewRing(5, false, false)
		if err != nil {
			t.Fatalf("NewRing failed: %v", err)
		}
		defer g.Close()

		seed := uint64(7)
		coords, err := g.LayoutFruchtermanReingold3D(igraph.FruchtermanReingoldOptions{Seed: &seed, NIter: 50})
		if err != nil {
			t.Fatalf("LayoutFruchtermanReingold3D failed: %v", err)
		}
		if r, c := coords.Dims(); r != 5 || c != 3 {
			t.Fatalf("got dims (%d, %d), want (5, 3)", r, c)
		}

		initCoords, _ := igraph.NewMatrixFromRows([][]float64{
			{0, 0, 0}, {1, 1, 1}, {2, 2, 2}, {3, 3, 3}, {4, 4, 4},
		})
		bound := []float64{5, 5, 5, 5, 5}
		negBound := []float64{-5, -5, -5, -5, -5}
		coordsBounded, err := g.LayoutFruchtermanReingold3D(igraph.FruchtermanReingoldOptions{
			Seed:               &seed,
			NIter:              50,
			Weights:            []float64{1, 1, 1, 1, 1},
			InitialCoordinates: &initCoords,
			MinX:               negBound, MaxX: bound,
			MinY: negBound, MaxY: bound,
			MinZ: negBound, MaxZ: bound,
		})
		if err != nil {
			t.Fatalf("LayoutFruchtermanReingold3D with bounds failed: %v", err)
		}
		for r := 0; r < 5; r++ {
			z, _ := coordsBounded.At(r, 2)
			if z < -5 || z > 5 {
				t.Errorf("vertex %d z=%v outside [-5, 5]", r, z)
			}
		}
	})

	t.Run("seed reproducibility", func(t *testing.T) {
		g, err := igraph.NewRing(6, false, false)
		if err != nil {
			t.Fatalf("NewRing failed: %v", err)
		}
		defer g.Close()

		seed := uint64(99)
		c1, err := g.LayoutFruchtermanReingold3D(igraph.FruchtermanReingoldOptions{Seed: &seed, NIter: 30})
		if err != nil {
			t.Fatalf("layout 1 failed: %v", err)
		}
		c2, err := g.LayoutFruchtermanReingold3D(igraph.FruchtermanReingoldOptions{Seed: &seed, NIter: 30})
		if err != nil {
			t.Fatalf("layout 2 failed: %v", err)
		}
		assertEqualMatrices(t, "seeded 3D layout", c1, c2)
	})

	t.Run("invalid parameters", func(t *testing.T) {
		g, err := igraph.NewRing(4, false, false)
		if err != nil {
			t.Fatalf("NewRing failed: %v", err)
		}
		defer g.Close()

		if _, err := g.LayoutFruchtermanReingold3D(igraph.FruchtermanReingoldOptions{MinZ: []float64{1.0}}); err == nil {
			t.Error("expected error for mismatched MinZ length")
		}
		if _, err := g.LayoutFruchtermanReingold3D(igraph.FruchtermanReingoldOptions{MaxZ: []float64{0, 0, math.NaN(), 0}}); err == nil {
			t.Error("expected error for NaN MaxZ")
		}
		badInit, _ := igraph.NewMatrixFromRows([][]float64{{0, 0}, {1, 1}, {2, 2}, {3, 3}})
		if _, err := g.LayoutFruchtermanReingold3D(igraph.FruchtermanReingoldOptions{InitialCoordinates: &badInit}); err == nil {
			t.Error("expected error for 2-column InitialCoordinates in 3D layout")
		}
		// Non-positive weights pass Go-side validation and are rejected by
		// the upstream C implementation.
		if _, err := g.LayoutFruchtermanReingold3D(igraph.FruchtermanReingoldOptions{Weights: []float64{-1, 1, 1, 1}}); err == nil {
			t.Error("expected upstream error for negative weight")
		}
		// The 2D variant must reject z-axis bounds; an empty slice means no
		// bound and is accepted.
		if _, err := g.LayoutFruchtermanReingold(igraph.FruchtermanReingoldOptions{MinZ: []float64{0, 0, 0, 0}}); err == nil {
			t.Error("expected error for MinZ on 2D layout")
		}
		if _, err := g.LayoutFruchtermanReingold(igraph.FruchtermanReingoldOptions{MaxZ: []float64{0, 0, 0, 0}}); err == nil {
			t.Error("expected error for MaxZ on 2D layout")
		}
		if _, err := g.LayoutFruchtermanReingold(igraph.FruchtermanReingoldOptions{NIter: 5, MinZ: []float64{}}); err != nil {
			t.Errorf("empty MinZ on 2D layout should be accepted: %v", err)
		}
	})

	t.Run("closed and nil graph", func(t *testing.T) {
		var nilG *igraph.Graph
		if _, err := nilG.LayoutFruchtermanReingold3D(igraph.FruchtermanReingoldOptions{}); err != igraph.ErrClosed {
			t.Errorf("got %v, want ErrClosed for nil graph", err)
		}
		g, _ := igraph.NewGraphFromEdges(3, nil, false)
		g.Close()
		if _, err := g.LayoutFruchtermanReingold3D(igraph.FruchtermanReingoldOptions{}); err != igraph.ErrClosed {
			t.Errorf("got %v, want ErrClosed for closed graph", err)
		}
	})
}

func TestLayoutKamadaKawai3D(t *testing.T) {
	t.Run("dimensions and empty graph", func(t *testing.T) {
		g, err := igraph.NewRing(5, false, false)
		if err != nil {
			t.Fatalf("NewRing failed: %v", err)
		}
		defer g.Close()

		seed := uint64(11)
		coords, err := g.LayoutKamadaKawai3D(igraph.KamadaKawaiOptions{Seed: &seed, MaxIter: 50})
		if err != nil {
			t.Fatalf("LayoutKamadaKawai3D failed: %v", err)
		}
		if r, c := coords.Dims(); r != 5 || c != 3 {
			t.Fatalf("got dims (%d, %d), want (5, 3)", r, c)
		}

		empty, err := igraph.NewGraphFromEdges(0, nil, false)
		if err != nil {
			t.Fatalf("NewGraphFromEdges failed: %v", err)
		}
		defer empty.Close()
		emptyCoords, err := empty.LayoutKamadaKawai3D(igraph.KamadaKawaiOptions{})
		if err != nil {
			t.Fatalf("LayoutKamadaKawai3D on empty graph failed: %v", err)
		}
		if r, _ := emptyCoords.Dims(); r != 0 {
			t.Errorf("got %d rows, want 0", r)
		}
	})

	t.Run("invalid parameters", func(t *testing.T) {
		g, err := igraph.NewRing(4, false, false)
		if err != nil {
			t.Fatalf("NewRing failed: %v", err)
		}
		defer g.Close()

		if _, err := g.LayoutKamadaKawai3D(igraph.KamadaKawaiOptions{MinZ: []float64{1.0}}); err == nil {
			t.Error("expected error for mismatched MinZ length")
		}
		badInit, _ := igraph.NewMatrixFromRows([][]float64{{0, 0}, {1, 1}, {2, 2}, {3, 3}})
		if _, err := g.LayoutKamadaKawai3D(igraph.KamadaKawaiOptions{InitialCoordinates: &badInit}); err == nil {
			t.Error("expected error for 2-column InitialCoordinates in 3D layout")
		}
		// Non-positive weights pass Go-side validation and are rejected by
		// the upstream C implementation.
		if _, err := g.LayoutKamadaKawai3D(igraph.KamadaKawaiOptions{Weights: []float64{0, 1, 1, 1}}); err == nil {
			t.Error("expected upstream error for non-positive weight")
		}
		if _, err := g.LayoutKamadaKawai(igraph.KamadaKawaiOptions{MinZ: []float64{0, 0, 0, 0}}); err == nil {
			t.Error("expected error for MinZ on 2D layout")
		}
	})

	t.Run("closed and nil graph", func(t *testing.T) {
		var nilG *igraph.Graph
		if _, err := nilG.LayoutKamadaKawai3D(igraph.KamadaKawaiOptions{}); err != igraph.ErrClosed {
			t.Errorf("got %v, want ErrClosed for nil graph", err)
		}
		g, _ := igraph.NewGraphFromEdges(3, nil, false)
		g.Close()
		if _, err := g.LayoutKamadaKawai3D(igraph.KamadaKawaiOptions{}); err != igraph.ErrClosed {
			t.Errorf("got %v, want ErrClosed for closed graph", err)
		}
	})
}
