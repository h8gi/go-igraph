package igraph

import (
	"errors"
	"testing"
)

func TestLaplacianFailurePaths(t *testing.T) {
	g, _ := NewGraphFromEdges(2, []Edge{{0, 1}}, false)
	defer g.Close()
	injected := errors.New("injected Laplacian failure")

	adapters := defaultLaplacianAdapters()
	adapters.newMatrix = func(Matrix) (*cMatrix, error) { return nil, injected }
	if _, err := g.laplacian(LaplacianOptions{}, &adapters); !errors.Is(err, injected) {
		t.Fatalf("matrix init = %v", err)
	}

	adapters = defaultLaplacianAdapters()
	adapters.newReal = func([]float64) (*realVector, error) { return nil, injected }
	if _, err := g.laplacian(LaplacianOptions{Weights: []float64{1}}, &adapters); !errors.Is(err, injected) {
		t.Fatalf("weight init = %v", err)
	}

	adapters = defaultLaplacianAdapters()
	adapters.call = func(*Graph, *cMatrix, laplacianDirection, cLaplacianNormalization, *realVector) int { return 4 }
	if _, err := g.laplacian(LaplacianOptions{}, &adapters); err == nil {
		t.Fatal("upstream error not propagated")
	}

	adapters = defaultLaplacianAdapters()
	adapters.convert = func(*cMatrix) (Matrix, error) { return Matrix{}, injected }
	if _, err := g.laplacian(LaplacianOptions{}, &adapters); !errors.Is(err, injected) {
		t.Fatalf("conversion = %v", err)
	}

	adapters = defaultLaplacianAdapters()
	adapters.convert = func(*cMatrix) (Matrix, error) { return NewMatrix(1, 1) }
	if _, err := g.laplacian(LaplacianOptions{}, &adapters); err == nil {
		t.Fatal("dimension mismatch accepted")
	}
}
