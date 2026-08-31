package igraph_test

import (
	"fmt"
	"log"
	"math"

	igraph "github.com/h8gi/go-igraph"
)

func ExampleGraph_LayoutCircle() {
	graph, err := igraph.NewRing(4, false, false)
	if err != nil {
		log.Fatal(err)
	}
	defer graph.Close()

	// Deterministic layouts need no seed: identical calls always return
	// identical coordinates, with row i holding the position of vertex i.
	coords, err := graph.LayoutCircle(nil)
	if err != nil {
		log.Fatal(err)
	}
	for v := 0; v < 4; v++ {
		x, _ := coords.At(v, 0)
		y, _ := coords.At(v, 1)
		fmt.Printf("vertex %d: (%.2f, %.2f)\n", v, x, y)
	}
	// Output:
	// vertex 0: (1.00, 0.00)
	// vertex 1: (0.00, 1.00)
	// vertex 2: (-1.00, 0.00)
	// vertex 3: (-0.00, -1.00)
}

func ExampleGraph_LayoutFruchtermanReingold() {
	graph, err := igraph.NewRing(6, false, false)
	if err != nil {
		log.Fatal(err)
	}
	defer graph.Close()

	// A fixed Seed makes the force-directed layout reproducible; per-axis
	// bounds clamp every coordinate.
	seed := uint64(42)
	bound := []float64{3, 3, 3, 3, 3, 3}
	negBound := []float64{-3, -3, -3, -3, -3, -3}
	first, err := graph.LayoutFruchtermanReingold(igraph.FruchtermanReingoldOptions{
		Seed: &seed,
		MinX: negBound, MaxX: bound,
		MinY: negBound, MaxY: bound,
	})
	if err != nil {
		log.Fatal(err)
	}
	second, err := graph.LayoutFruchtermanReingold(igraph.FruchtermanReingoldOptions{
		Seed: &seed,
		MinX: negBound, MaxX: bound,
		MinY: negBound, MaxY: bound,
	})
	if err != nil {
		log.Fatal(err)
	}

	rows, cols := first.Dims()
	identical, inBounds := true, true
	for r := 0; r < rows; r++ {
		for c := 0; c < cols; c++ {
			v1, _ := first.At(r, c)
			v2, _ := second.At(r, c)
			identical = identical && v1 == v2
			inBounds = inBounds && v1 >= -3 && v1 <= 3
		}
	}
	fmt.Printf("dimensions: %dx%d\n", rows, cols)
	fmt.Printf("same seed reproduces the layout: %t\n", identical)
	fmt.Printf("all coordinates within [-3, 3]: %t\n", inBounds)
	// Output:
	// dimensions: 6x2
	// same seed reproduces the layout: true
	// all coordinates within [-3, 3]: true
}

func ExampleGraph_LayoutSphere() {
	graph, err := igraph.NewRing(8, false, false)
	if err != nil {
		log.Fatal(err)
	}
	defer graph.Close()

	coords, err := graph.LayoutSphere()
	if err != nil {
		log.Fatal(err)
	}
	rows, cols := coords.Dims()
	x, _ := coords.At(0, 0)
	y, _ := coords.At(0, 1)
	z, _ := coords.At(0, 2)
	fmt.Printf("dimensions: %dx%d\n", rows, cols)
	fmt.Printf("vertex 0 radius: %.2f\n", math.Sqrt(x*x+y*y+z*z))
	// Output:
	// dimensions: 8x3
	// vertex 0 radius: 1.00
}

func ExampleGraph_AdjacencySpectralEmbedding() {
	graph, err := igraph.NewRing(6, false, false)
	if err != nil {
		log.Fatal(err)
	}
	defer graph.Close()

	seed := uint64(7)
	result, err := graph.AdjacencySpectralEmbedding(2, igraph.SpectralEmbeddingOptions{Seed: &seed})
	if err != nil {
		log.Fatal(err)
	}
	rows, cols := result.X.Dims()
	fmt.Printf("embedding dimensions: %dx%d\n", rows, cols)
	fmt.Printf("number of eigenvalues: %d\n", len(result.SingularValues))
	// Output:
	// embedding dimensions: 6x2
	// number of eigenvalues: 2
}

func ExampleGraph_LayoutDrL() {
	graph, err := igraph.NewRing(6, false, false)
	if err != nil {
		log.Fatal(err)
	}
	defer graph.Close()
	seed := uint64(28)
	first, err := graph.LayoutDrL(igraph.DrLOptions{Seed: &seed})
	if err != nil {
		log.Fatal(err)
	}
	second, err := graph.LayoutDrL(igraph.DrLOptions{Seed: &seed})
	if err != nil {
		log.Fatal(err)
	}
	rows, columns := first.Dims()
	same := true
	for row := 0; row < rows; row++ {
		for column := 0; column < columns; column++ {
			left, _ := first.At(row, column)
			right, _ := second.At(row, column)
			same = same && left == right
		}
	}
	fmt.Printf("DrL dimensions: %dx%d\n", rows, columns)
	fmt.Printf("same seed reproduces DrL: %t\n", same)
	// Output:
	// DrL dimensions: 6x2
	// same seed reproduces DrL: true
}

func ExampleGraph_TreeLayoutRoots() {
	graph, err := igraph.NewGraphFromEdges(5, []igraph.Edge{
		{From: 0, To: 1}, {From: 1, To: 2}, {From: 2, To: 3}, {From: 3, To: 4},
	}, false)
	if err != nil {
		log.Fatal(err)
	}
	defer graph.Close()
	roots, err := graph.TreeLayoutRoots(igraph.DegAll, igraph.TreeRootByEccentricity)
	if err != nil {
		log.Fatal(err)
	}
	coordinates, err := graph.LayoutReingoldTilford(igraph.DegAll, roots)
	if err != nil {
		log.Fatal(err)
	}
	aligned, err := graph.AlignLayout(coordinates)
	if err != nil {
		log.Fatal(err)
	}
	rows, columns := aligned.Dims()
	fmt.Printf("selected roots: %v\n", roots)
	fmt.Printf("aligned tree dimensions: %dx%d\n", rows, columns)
	// Output:
	// selected roots: [2]
	// aligned tree dimensions: 5x2
}

func ExampleMergeLayoutsDLA() {
	left, err := igraph.NewGraphFromEdges(2, []igraph.Edge{{From: 0, To: 1}}, false)
	if err != nil {
		log.Fatal(err)
	}
	defer left.Close()
	right, err := igraph.NewGraphFromEdges(2, []igraph.Edge{{From: 0, To: 1}}, false)
	if err != nil {
		log.Fatal(err)
	}
	defer right.Close()
	coordinates, _ := igraph.NewMatrixFromRows([][]float64{{-1, 0}, {1, 0}})
	seed := uint64(28)
	merged, err := igraph.MergeLayoutsDLA(
		[]*igraph.Graph{left, right},
		[]igraph.Matrix{coordinates, coordinates},
		igraph.DLAMergeOptions{Seed: &seed},
	)
	if err != nil {
		log.Fatal(err)
	}
	rows, columns := merged.Dims()
	fmt.Printf("merged dimensions: %dx%d\n", rows, columns)
	// Output:
	// merged dimensions: 4x2
}
