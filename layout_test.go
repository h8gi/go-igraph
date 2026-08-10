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
