package igraph_test

import (
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

		for r := 0; r < 10; r++ {
			for c := 0; c < 2; c++ {
				v1, _ := coords1.At(r, c)
				v2, _ := coords2.At(r, c)
				if v1 != v2 {
					t.Errorf("mismatch at (%d, %d): %v vs %v", r, c, v1, v2)
				}
			}
		}
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

		for r := 0; r < 10; r++ {
			for c := 0; c < 2; c++ {
				v1, _ := c1.At(r, c)
				v2, _ := c2.At(r, c)
				if v1 != v2 {
					t.Errorf("mismatch at (%d, %d): %v vs %v", r, c, v1, v2)
				}
			}
		}

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
		if _, err := g.LayoutFruchtermanReingold(igraph.FruchtermanReingoldOptions{MinX: []float64{1.0}}); err == nil {
			t.Error("expected error for mismatched MinX length")
		}
		badInit, _ := igraph.NewMatrixFromRows([][]float64{{0, 0}, {1, 1}})
		if _, err := g.LayoutFruchtermanReingold(igraph.FruchtermanReingoldOptions{InitialCoordinates: &badInit}); err == nil {
			t.Error("expected error for mismatched InitialCoordinates dimensions")
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
		weights := []float64{1.0, 1.0, 1.0, 1.0, 1.0}

		coordsInit, err := g.LayoutKamadaKawai(igraph.KamadaKawaiOptions{
			Seed:               &seed,
			MaxIter:            50,
			Weights:            weights,
			InitialCoordinates: &initCoords,
			MinX:               minX,
			MaxX:               maxX,
		})
		if err != nil {
			t.Fatalf("LayoutKamadaKawai with initial coords failed: %v", err)
		}
		if r, c := coordsInit.Dims(); r != 5 || c != 2 {
			t.Fatalf("got dims (%d, %d), want (5, 2)", r, c)
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

		for r := 0; r < 10; r++ {
			for c := 0; c < 2; c++ {
				v1, _ := c1.At(r, c)
				v2, _ := c2.At(r, c)
				if v1 != v2 {
					t.Errorf("mismatch at (%d, %d): %v vs %v", r, c, v1, v2)
				}
			}
		}

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
		if _, err := g.LayoutKamadaKawai(igraph.KamadaKawaiOptions{MinX: []float64{1.0}}); err == nil {
			t.Error("expected error for mismatched MinX length")
		}
		badInit, _ := igraph.NewMatrixFromRows([][]float64{{0, 0}, {1, 1}})
		if _, err := g.LayoutKamadaKawai(igraph.KamadaKawaiOptions{InitialCoordinates: &badInit}); err == nil {
			t.Error("expected error for mismatched InitialCoordinates dimensions")
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
