package igraph_test

import (
	"math"
	"testing"

	igraph "github.com/h8gi/go-igraph"
)

func TestAdjacencySpectralEmbedding(t *testing.T) {
	t.Run("undirected embedding", func(t *testing.T) {
		g, err := igraph.NewRing(8, false, false)
		if err != nil {
			t.Fatalf("NewRing failed: %v", err)
		}
		defer g.Close()

		result, err := g.AdjacencySpectralEmbedding(2, igraph.SpectralEmbeddingOptions{})
		if err != nil {
			t.Fatalf("AdjacencySpectralEmbedding failed: %v", err)
		}
		if r, c := result.X.Dims(); r != 8 || c != 2 {
			t.Fatalf("got X dims (%d, %d), want (8, 2)", r, c)
		}
		if r, _ := result.Y.Dims(); r != 0 {
			t.Errorf("got %d Y rows for undirected graph, want 0", r)
		}
		if len(result.SingularValues) != 2 {
			t.Fatalf("got %d singular values, want 2", len(result.SingularValues))
		}
		// For undirected graphs the D values are eigenvalues selected by
		// magnitude; they may be negative.
		if math.Abs(result.SingularValues[0]) < math.Abs(result.SingularValues[1]) {
			t.Errorf("eigenvalues not in decreasing magnitude order: %v", result.SingularValues)
		}
		for _, value := range result.SingularValues {
			if math.IsNaN(value) {
				t.Errorf("invalid singular value: %v", value)
			}
		}
	})

	t.Run("directed embedding fills Y", func(t *testing.T) {
		g, err := igraph.NewRing(8, true, false)
		if err != nil {
			t.Fatalf("NewRing failed: %v", err)
		}
		defer g.Close()

		result, err := g.AdjacencySpectralEmbedding(2, igraph.SpectralEmbeddingOptions{})
		if err != nil {
			t.Fatalf("AdjacencySpectralEmbedding failed: %v", err)
		}
		if r, c := result.X.Dims(); r != 8 || c != 2 {
			t.Fatalf("got X dims (%d, %d), want (8, 2)", r, c)
		}
		if r, c := result.Y.Dims(); r != 8 || c != 2 {
			t.Fatalf("got Y dims (%d, %d), want (8, 2)", r, c)
		}
	})

	t.Run("scaled, weighted, and custom degree correction", func(t *testing.T) {
		g, err := igraph.NewRing(6, false, false)
		if err != nil {
			t.Fatalf("NewRing failed: %v", err)
		}
		defer g.Close()

		weights := []float64{1, 2, 1, 2, 1, 2}
		correction := []float64{0.4, 0.4, 0.4, 0.4, 0.4, 0.4}
		result, err := g.AdjacencySpectralEmbedding(3, igraph.SpectralEmbeddingOptions{
			Scaled:           true,
			Weights:          weights,
			DegreeCorrection: correction,
			Which:            igraph.EmbeddingLargestAlgebraic,
		})
		if err != nil {
			t.Fatalf("AdjacencySpectralEmbedding failed: %v", err)
		}
		if r, c := result.X.Dims(); r != 6 || c != 3 {
			t.Fatalf("got X dims (%d, %d), want (6, 3)", r, c)
		}
	})

	t.Run("repeated calls agree on singular values", func(t *testing.T) {
		g, err := igraph.NewRing(10, false, false)
		if err != nil {
			t.Fatalf("NewRing failed: %v", err)
		}
		defer g.Close()

		first, err := g.AdjacencySpectralEmbedding(2, igraph.SpectralEmbeddingOptions{})
		if err != nil {
			t.Fatalf("first call failed: %v", err)
		}
		second, err := g.AdjacencySpectralEmbedding(2, igraph.SpectralEmbeddingOptions{})
		if err != nil {
			t.Fatalf("second call failed: %v", err)
		}
		for i := range first.SingularValues {
			if math.Abs(first.SingularValues[i]-second.SingularValues[i]) > 1e-9 {
				t.Errorf("singular value %d differs: %v vs %v", i, first.SingularValues[i], second.SingularValues[i])
			}
		}
	})

	t.Run("invalid parameters", func(t *testing.T) {
		g, err := igraph.NewRing(5, false, false)
		if err != nil {
			t.Fatalf("NewRing failed: %v", err)
		}
		defer g.Close()

		if _, err := g.AdjacencySpectralEmbedding(0, igraph.SpectralEmbeddingOptions{}); err == nil {
			t.Error("expected error for zero dimension")
		}
		if _, err := g.AdjacencySpectralEmbedding(-1, igraph.SpectralEmbeddingOptions{}); err == nil {
			t.Error("expected error for negative dimension")
		}
		if _, err := g.AdjacencySpectralEmbedding(6, igraph.SpectralEmbeddingOptions{}); err == nil {
			t.Error("expected error for dimension exceeding vertex count")
		}
		if _, err := g.AdjacencySpectralEmbedding(2, igraph.SpectralEmbeddingOptions{Weights: []float64{1}}); err == nil {
			t.Error("expected error for mismatched Weights length")
		}
		if _, err := g.AdjacencySpectralEmbedding(2, igraph.SpectralEmbeddingOptions{DegreeCorrection: []float64{1}}); err == nil {
			t.Error("expected error for mismatched DegreeCorrection length")
		}
		if _, err := g.AdjacencySpectralEmbedding(2, igraph.SpectralEmbeddingOptions{DegreeCorrection: []float64{1, 1, math.NaN(), 1, 1}}); err == nil {
			t.Error("expected error for NaN DegreeCorrection")
		}
		if _, err := g.AdjacencySpectralEmbedding(2, igraph.SpectralEmbeddingOptions{Type: igraph.LaplacianEmbeddingDAD}); err == nil {
			t.Error("expected error for Type on adjacency embedding")
		}
		if _, err := g.AdjacencySpectralEmbedding(2, igraph.SpectralEmbeddingOptions{Which: igraph.SpectralEmbeddingWhich(99)}); err == nil {
			t.Error("expected error for invalid Which")
		}
		if _, err := g.AdjacencySpectralEmbedding(2, igraph.SpectralEmbeddingOptions{Solver: igraph.SpectralSolverOptions{MaxIterations: -1}}); err == nil {
			t.Error("expected error for negative solver iterations")
		}
	})

	t.Run("smallest algebraic eigenvalues", func(t *testing.T) {
		g, err := igraph.NewRing(6, false, false)
		if err != nil {
			t.Fatalf("NewRing failed: %v", err)
		}
		defer g.Close()

		result, err := g.AdjacencySpectralEmbedding(2, igraph.SpectralEmbeddingOptions{Which: igraph.EmbeddingSmallestAlgebraic})
		if err != nil {
			t.Fatalf("AdjacencySpectralEmbedding failed: %v", err)
		}
		if r, c := result.X.Dims(); r != 6 || c != 2 {
			t.Fatalf("got X dims (%d, %d), want (6, 2)", r, c)
		}
	})

	t.Run("single vertex graph", func(t *testing.T) {
		g, err := igraph.NewGraphFromEdges(1, nil, false)
		if err != nil {
			t.Fatalf("NewGraphFromEdges failed: %v", err)
		}
		defer g.Close()

		result, err := g.AdjacencySpectralEmbedding(1, igraph.SpectralEmbeddingOptions{})
		if err != nil {
			t.Fatalf("AdjacencySpectralEmbedding failed: %v", err)
		}
		if r, c := result.X.Dims(); r != 1 || c != 1 {
			t.Fatalf("got X dims (%d, %d), want (1, 1)", r, c)
		}
	})

	t.Run("empty graph", func(t *testing.T) {
		g, err := igraph.NewGraphFromEdges(0, nil, false)
		if err != nil {
			t.Fatalf("NewGraphFromEdges failed: %v", err)
		}
		defer g.Close()
		if _, err := g.AdjacencySpectralEmbedding(1, igraph.SpectralEmbeddingOptions{}); err == nil {
			t.Error("expected error for embedding of empty graph")
		}
	})

	t.Run("closed and nil graph", func(t *testing.T) {
		var nilG *igraph.Graph
		if _, err := nilG.AdjacencySpectralEmbedding(1, igraph.SpectralEmbeddingOptions{}); err != igraph.ErrClosed {
			t.Errorf("got %v, want ErrClosed for nil graph", err)
		}
		g, _ := igraph.NewGraphFromEdges(3, nil, false)
		g.Close()
		if _, err := g.AdjacencySpectralEmbedding(1, igraph.SpectralEmbeddingOptions{}); err != igraph.ErrClosed {
			t.Errorf("got %v, want ErrClosed for closed graph", err)
		}
	})
}

func TestLaplacianSpectralEmbedding(t *testing.T) {
	t.Run("undirected Laplacian types", func(t *testing.T) {
		g, err := igraph.NewRing(8, false, false)
		if err != nil {
			t.Fatalf("NewRing failed: %v", err)
		}
		defer g.Close()

		for _, embeddingType := range []igraph.LaplacianEmbeddingType{
			igraph.LaplacianEmbeddingDA,
			igraph.LaplacianEmbeddingIDAD,
			igraph.LaplacianEmbeddingDAD,
		} {
			result, err := g.LaplacianSpectralEmbedding(2, igraph.SpectralEmbeddingOptions{Type: embeddingType})
			if err != nil {
				t.Fatalf("LaplacianSpectralEmbedding type %d failed: %v", embeddingType, err)
			}
			if r, c := result.X.Dims(); r != 8 || c != 2 {
				t.Fatalf("type %d: got X dims (%d, %d), want (8, 2)", embeddingType, r, c)
			}
			if len(result.SingularValues) != 2 {
				t.Fatalf("type %d: got %d singular values, want 2", embeddingType, len(result.SingularValues))
			}
		}

		// OAP is restricted to directed graphs upstream.
		if _, err := g.LaplacianSpectralEmbedding(2, igraph.SpectralEmbeddingOptions{Type: igraph.LaplacianEmbeddingOAP}); err == nil {
			t.Error("expected error for OAP on undirected graph")
		}
	})

	t.Run("directed OAP embedding fills Y", func(t *testing.T) {
		g, err := igraph.NewRing(8, true, false)
		if err != nil {
			t.Fatalf("NewRing failed: %v", err)
		}
		defer g.Close()

		result, err := g.LaplacianSpectralEmbedding(2, igraph.SpectralEmbeddingOptions{Type: igraph.LaplacianEmbeddingOAP})
		if err != nil {
			t.Fatalf("LaplacianSpectralEmbedding failed: %v", err)
		}
		if r, c := result.Y.Dims(); r != 8 || c != 2 {
			t.Fatalf("got Y dims (%d, %d), want (8, 2)", r, c)
		}

		// The undirected-only Laplacian definitions are rejected upstream
		// for directed graphs.
		if _, err := g.LaplacianSpectralEmbedding(2, igraph.SpectralEmbeddingOptions{Type: igraph.LaplacianEmbeddingDA}); err == nil {
			t.Error("expected error for D-A on directed graph")
		}
	})

	t.Run("invalid parameters", func(t *testing.T) {
		g, err := igraph.NewRing(5, false, false)
		if err != nil {
			t.Fatalf("NewRing failed: %v", err)
		}
		defer g.Close()

		if _, err := g.LaplacianSpectralEmbedding(2, igraph.SpectralEmbeddingOptions{DegreeCorrection: []float64{1, 1, 1, 1, 1}}); err == nil {
			t.Error("expected error for DegreeCorrection on Laplacian embedding")
		}
		if _, err := g.LaplacianSpectralEmbedding(2, igraph.SpectralEmbeddingOptions{Type: igraph.LaplacianEmbeddingType(99)}); err == nil {
			t.Error("expected error for invalid Type")
		}
		if _, err := g.LaplacianSpectralEmbedding(0, igraph.SpectralEmbeddingOptions{}); err == nil {
			t.Error("expected error for zero dimension")
		}
	})

	t.Run("closed and nil graph", func(t *testing.T) {
		var nilG *igraph.Graph
		if _, err := nilG.LaplacianSpectralEmbedding(1, igraph.SpectralEmbeddingOptions{}); err != igraph.ErrClosed {
			t.Errorf("got %v, want ErrClosed for nil graph", err)
		}
		g, _ := igraph.NewGraphFromEdges(3, nil, false)
		g.Close()
		if _, err := g.LaplacianSpectralEmbedding(1, igraph.SpectralEmbeddingOptions{}); err != igraph.ErrClosed {
			t.Errorf("got %v, want ErrClosed for closed graph", err)
		}
	})
}

func TestDimSelect(t *testing.T) {
	t.Run("elbow detection", func(t *testing.T) {
		dim, err := igraph.DimSelect([]float64{100, 90, 3, 2, 1})
		if err != nil {
			t.Fatalf("DimSelect failed: %v", err)
		}
		if dim != 2 {
			t.Errorf("got dim %d, want 2", dim)
		}
	})

	t.Run("embedding pipeline", func(t *testing.T) {
		g, err := igraph.NewRing(10, false, false)
		if err != nil {
			t.Fatalf("NewRing failed: %v", err)
		}
		defer g.Close()

		result, err := g.AdjacencySpectralEmbedding(5, igraph.SpectralEmbeddingOptions{})
		if err != nil {
			t.Fatalf("AdjacencySpectralEmbedding failed: %v", err)
		}
		dim, err := igraph.DimSelect(result.SingularValues)
		if err != nil {
			t.Fatalf("DimSelect failed: %v", err)
		}
		if dim < 1 || dim > len(result.SingularValues) {
			t.Errorf("got dim %d, want within [1, %d]", dim, len(result.SingularValues))
		}
	})

	t.Run("invalid input", func(t *testing.T) {
		if _, err := igraph.DimSelect(nil); err == nil {
			t.Error("expected error for empty input")
		}
		if _, err := igraph.DimSelect([]float64{1, math.NaN()}); err == nil {
			t.Error("expected error for NaN input")
		}
		if _, err := igraph.DimSelect([]float64{1, math.Inf(1)}); err == nil {
			t.Error("expected error for Inf input")
		}
	})
}
