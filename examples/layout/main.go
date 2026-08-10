// Command layout demonstrates the Milestone 9 layout and embedding APIs:
// deterministic layouts, reproducible force-directed layouts with fixed
// seeds, 3D layouts, and spectral embeddings with dimensionality selection.
package main

import (
	"fmt"
	"log"
	"math"

	igraph "github.com/h8gi/go-igraph"
)

func main() {
	const numVertices = 12
	seed := uint64(2026)

	fmt.Println("=== Milestone 9: Layouts & Embeddings Demo ===")

	// 1. Generate a reproducible scale-free network.
	g, err := igraph.BarabasiGame(numVertices, 2, 1.0, 1.0, false, igraph.BarabasiOptions{Seed: &seed})
	if err != nil {
		log.Fatalf("BarabasiGame failed: %v", err)
	}
	defer g.Close()
	fmt.Printf("1. Generated a %d-vertex Barabási-Albert graph.\n", numVertices)

	// 2. Deterministic layouts need no seed: identical calls give identical
	// coordinates.
	circle, err := g.LayoutCircle(nil)
	if err != nil {
		log.Fatalf("LayoutCircle failed: %v", err)
	}
	x0, _ := circle.At(0, 0)
	y0, _ := circle.At(0, 1)
	fmt.Printf("2. Deterministic circle layout places vertex 0 at (%.2f, %.2f).\n", x0, y0)

	// 3. A seeded Fruchterman-Reingold run refines the circle layout inside
	// per-axis bounds; the same seed always reproduces the same layout.
	bound := make([]float64, numVertices)
	negBound := make([]float64, numVertices)
	for i := range bound {
		bound[i] = 5
		negBound[i] = -5
	}
	fr, err := g.LayoutFruchtermanReingold(igraph.FruchtermanReingoldOptions{
		Seed:               &seed,
		InitialCoordinates: &circle,
		MinX:               negBound, MaxX: bound,
		MinY: negBound, MaxY: bound,
	})
	if err != nil {
		log.Fatalf("LayoutFruchtermanReingold failed: %v", err)
	}
	fmt.Printf("3. Seeded Fruchterman-Reingold layout stays within [-5, 5]: %s\n", boundsSummary(fr))

	// 4. 3D layouts share the vertex-to-row convention with 3 columns.
	sphere, err := g.LayoutSphere()
	if err != nil {
		log.Fatalf("LayoutSphere failed: %v", err)
	}
	rows, cols := sphere.Dims()
	fmt.Printf("4. Sphere layout returns a %dx%d coordinate matrix on the unit sphere.\n", rows, cols)

	kk3d, err := g.LayoutKamadaKawai3D(igraph.KamadaKawaiOptions{Seed: &seed, MaxIter: 200})
	if err != nil {
		log.Fatalf("LayoutKamadaKawai3D failed: %v", err)
	}
	fmt.Printf("   3D Kamada-Kawai layout spans %s\n", boundsSummary(kk3d))

	// 5. Spectral embedding with data-driven dimensionality selection.
	embedding, err := g.AdjacencySpectralEmbedding(4, igraph.SpectralEmbeddingOptions{Seed: &seed})
	if err != nil {
		log.Fatalf("AdjacencySpectralEmbedding failed: %v", err)
	}
	magnitudes := make([]float64, len(embedding.SingularValues))
	for i, value := range embedding.SingularValues {
		magnitudes[i] = math.Abs(value)
	}
	dim, err := igraph.DimSelect(magnitudes)
	if err != nil {
		log.Fatalf("DimSelect failed: %v", err)
	}
	fmt.Printf("5. Adjacency spectral embedding: eigenvalue magnitudes %.3v, elbow at dimension %d.\n", magnitudes, dim)

	laplacian, err := g.LaplacianSpectralEmbedding(dim, igraph.SpectralEmbeddingOptions{
		Seed: &seed,
		Type: igraph.LaplacianEmbeddingDAD,
	})
	if err != nil {
		log.Fatalf("LaplacianSpectralEmbedding failed: %v", err)
	}
	rows, cols = laplacian.X.Dims()
	fmt.Printf("   Laplacian (DAD) embedding at the selected dimensionality is %dx%d.\n", rows, cols)

	fmt.Println("Done: all layouts and embeddings are reproducible with the same seeds.")
}

// boundsSummary reports the coordinate ranges of a layout matrix.
func boundsSummary(m igraph.Matrix) string {
	rows, cols := m.Dims()
	low, high := math.Inf(1), math.Inf(-1)
	for r := 0; r < rows; r++ {
		for c := 0; c < cols; c++ {
			v, _ := m.At(r, c)
			low = math.Min(low, v)
			high = math.Max(high, v)
		}
	}
	return fmt.Sprintf("coordinates in [%.2f, %.2f]", low, high)
}
