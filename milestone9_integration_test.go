package igraph_test

import (
	"math"
	"sync"
	"testing"

	igraph "github.com/h8gi/go-igraph"
)

// TestMilestone9IntegrationPipeline combines reproducible graph generation,
// deterministic and force-directed layout computation, and spectral-embedding
// based downstream analysis.
func TestMilestone9IntegrationPipeline(t *testing.T) {
	seed := uint64(2026)

	// Step 1: Generate a reproducible scale-free graph.
	g, err := igraph.BarabasiGame(20, 2, 1.0, 1.0, false, igraph.BarabasiOptions{Seed: &seed})
	if err != nil {
		t.Fatalf("BarabasiGame failed: %v", err)
	}
	defer g.Close()
	vCount, eCount := mustCounts(t, g)
	if vCount != 20 || eCount == 0 {
		t.Fatalf("expected 20 vertices and non-zero edges, got %d and %d", vCount, eCount)
	}

	// Step 2: Deterministic circle layout seeds the force-directed run.
	circle, err := g.LayoutCircle(nil)
	if err != nil {
		t.Fatalf("LayoutCircle failed: %v", err)
	}
	if r, c := circle.Dims(); r != vCount || c != 2 {
		t.Fatalf("got circle dims (%d, %d), want (%d, 2)", r, c, vCount)
	}

	// Step 3: Bounded Fruchterman-Reingold refinement from the circle layout.
	bound := make([]float64, vCount)
	negBound := make([]float64, vCount)
	for i := range bound {
		bound[i] = 10
		negBound[i] = -10
	}
	frOptions := igraph.FruchtermanReingoldOptions{
		Seed:               &seed,
		NIter:              100,
		InitialCoordinates: &circle,
		MinX:               negBound, MaxX: bound,
		MinY: negBound, MaxY: bound,
	}
	fr, err := g.LayoutFruchtermanReingold(frOptions)
	if err != nil {
		t.Fatalf("LayoutFruchtermanReingold failed: %v", err)
	}
	for r := 0; r < vCount; r++ {
		for c := 0; c < 2; c++ {
			v, _ := fr.At(r, c)
			if v < -10 || v > 10 || math.IsNaN(v) {
				t.Errorf("coordinate (%d, %d) = %v escapes bounds [-10, 10]", r, c, v)
			}
		}
	}

	// Step 4: The seeded pipeline is reproducible end to end.
	frAgain, err := g.LayoutFruchtermanReingold(frOptions)
	if err != nil {
		t.Fatalf("LayoutFruchtermanReingold rerun failed: %v", err)
	}
	assertEqualMatrices(t, "FruchtermanReingold", fr, frAgain)

	// Step 5: A 3D Kamada-Kawai layout of the same graph.
	kk3d, err := g.LayoutKamadaKawai3D(igraph.KamadaKawaiOptions{Seed: &seed, MaxIter: 100})
	if err != nil {
		t.Fatalf("LayoutKamadaKawai3D failed: %v", err)
	}
	if r, c := kk3d.Dims(); r != vCount || c != 3 {
		t.Fatalf("got 3D dims (%d, %d), want (%d, 3)", r, c, vCount)
	}

	// Step 6: Spectral embedding and dimensionality selection drive the
	// downstream analysis.
	embedding, err := g.AdjacencySpectralEmbedding(4, igraph.SpectralEmbeddingOptions{Seed: &seed})
	if err != nil {
		t.Fatalf("AdjacencySpectralEmbedding failed: %v", err)
	}
	if r, c := embedding.X.Dims(); r != vCount || c != 4 {
		t.Fatalf("got embedding dims (%d, %d), want (%d, 4)", r, c, vCount)
	}
	magnitudes := make([]float64, len(embedding.SingularValues))
	for i, value := range embedding.SingularValues {
		magnitudes[i] = math.Abs(value)
	}
	dim, err := igraph.DimSelect(magnitudes)
	if err != nil {
		t.Fatalf("DimSelect failed: %v", err)
	}
	if dim < 1 || dim > len(magnitudes) {
		t.Fatalf("got dim %d, want within [1, %d]", dim, len(magnitudes))
	}

	// Step 7: Laplacian embedding at the selected dimensionality.
	laplacian, err := g.LaplacianSpectralEmbedding(dim, igraph.SpectralEmbeddingOptions{Seed: &seed, Type: igraph.LaplacianEmbeddingDAD})
	if err != nil {
		t.Fatalf("LaplacianSpectralEmbedding failed: %v", err)
	}
	if r, c := laplacian.X.Dims(); r != vCount || c != dim {
		t.Fatalf("got Laplacian embedding dims (%d, %d), want (%d, %d)", r, c, vCount, dim)
	}
}

// TestMilestone9ConcurrentSeedIsolation verifies that concurrently executed
// seeded layout and embedding calls produce exactly the results of their
// serial counterparts, demonstrating thread safety and seed isolation.
func TestMilestone9ConcurrentSeedIsolation(t *testing.T) {
	g, err := igraph.NewRing(12, false, false)
	if err != nil {
		t.Fatalf("NewRing failed: %v", err)
	}
	defer g.Close()

	type run struct {
		name    string
		compute func(seed uint64) (igraph.Matrix, error)
	}
	runs := []run{
		{"FruchtermanReingold", func(seed uint64) (igraph.Matrix, error) {
			return g.LayoutFruchtermanReingold(igraph.FruchtermanReingoldOptions{Seed: &seed, NIter: 50})
		}},
		{"KamadaKawai3D", func(seed uint64) (igraph.Matrix, error) {
			return g.LayoutKamadaKawai3D(igraph.KamadaKawaiOptions{Seed: &seed, MaxIter: 50})
		}},
		{"Random3D", func(seed uint64) (igraph.Matrix, error) {
			return g.LayoutRandom3D(igraph.LayoutRandomOptions{Seed: &seed})
		}},
		{"AdjacencySpectralEmbedding", func(seed uint64) (igraph.Matrix, error) {
			result, err := g.AdjacencySpectralEmbedding(3, igraph.SpectralEmbeddingOptions{Seed: &seed})
			if err != nil {
				return igraph.Matrix{}, err
			}
			return result.X, nil
		}},
	}

	// Serial reference results, one distinct seed per run.
	references := make([]igraph.Matrix, len(runs))
	for i, r := range runs {
		reference, err := r.compute(uint64(100 + i))
		if err != nil {
			t.Fatalf("serial %s failed: %v", r.name, err)
		}
		references[i] = reference
	}

	// Re-run everything concurrently, several times per seed, and require
	// exact agreement with the serial references.
	var wg sync.WaitGroup
	for round := 0; round < 3; round++ {
		for i, r := range runs {
			wg.Add(1)
			go func(index int, r run) {
				defer wg.Done()
				result, err := r.compute(uint64(100 + index))
				if err != nil {
					t.Errorf("concurrent %s failed: %v", r.name, err)
					return
				}
				assertEqualMatrices(t, r.name, references[index], result)
			}(i, r)
		}
	}
	wg.Wait()
}

func assertEqualMatrices(t *testing.T, name string, want, got igraph.Matrix) {
	t.Helper()
	wantRows, wantCols := want.Dims()
	gotRows, gotCols := got.Dims()
	if wantRows != gotRows || wantCols != gotCols {
		t.Errorf("%s: got dims (%d, %d), want (%d, %d)", name, gotRows, gotCols, wantRows, wantCols)
		return
	}
	for r := 0; r < wantRows; r++ {
		for c := 0; c < wantCols; c++ {
			wantValue, _ := want.At(r, c)
			gotValue, _ := got.At(r, c)
			if wantValue != gotValue {
				t.Errorf("%s: mismatch at (%d, %d): got %v, want %v", name, r, c, gotValue, wantValue)
				return
			}
		}
	}
}
