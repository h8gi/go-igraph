package igraph_test

import (
	"fmt"
	"math"
	"reflect"
	"testing"

	"github.com/h8gi/go-igraph"
)

func TestGeometricRandomGame(t *testing.T) {
	seed := uint64(340)
	first, err := igraph.GeometricRandomGame(10, 0.3, igraph.GeometricGraphOptions{Seed: &seed, Torus: true})
	if err != nil {
		t.Fatal(err)
	}
	defer first.Graph.Close()
	second, err := igraph.GeometricRandomGame(10, 0.3, igraph.GeometricGraphOptions{Seed: &seed, Torus: true})
	if err != nil {
		t.Fatal(err)
	}
	defer second.Graph.Close()
	if !reflect.DeepEqual(first.Coordinates.Rows(), second.Coordinates.Rows()) || !reflect.DeepEqual(mustEdges(t, first.Graph), mustEdges(t, second.Graph)) {
		t.Fatal("same seed differed")
	}
	if r, c := first.Coordinates.Dims(); r != 10 || c != 2 {
		t.Fatalf("coordinates=%dx%d", r, c)
	}
	copyRows := first.Coordinates.Rows()
	first.Graph.Close()
	if !reflect.DeepEqual(copyRows, first.Coordinates.Rows()) {
		t.Fatal("coordinates depended on graph")
	}
	complete, err := igraph.GeometricRandomGame(4, 2, igraph.GeometricGraphOptions{Seed: &seed})
	if err != nil {
		t.Fatal(err)
	}
	defer complete.Graph.Close()
	if _, e := mustCounts(t, complete.Graph); e != 6 {
		t.Fatalf("complete edges=%d", e)
	}
	empty, err := igraph.GeometricRandomGame(0, 0, igraph.GeometricGraphOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer empty.Graph.Close()
	if r, c := empty.Coordinates.Dims(); r != 0 || c != 2 {
		t.Fatalf("empty coordinates=%dx%d", r, c)
	}
	if _, e := igraph.GeometricRandomGame(-1, 0, igraph.GeometricGraphOptions{}); e == nil {
		t.Fatal("accepted negative vertices")
	}
	if _, e := igraph.GeometricRandomGame(1, -1, igraph.GeometricGraphOptions{}); e == nil {
		t.Fatal("accepted negative radius")
	}
	if _, e := igraph.GeometricRandomGame(1, math.NaN(), igraph.GeometricGraphOptions{}); e == nil {
		t.Fatal("accepted NaN radius")
	}
	if _, e := igraph.GeometricRandomGame(1, math.Inf(1), igraph.GeometricGraphOptions{}); e == nil {
		t.Fatal("accepted infinite radius")
	}
}

func TestLatentSamplers(t *testing.T) {
	seed := uint64(340)
	dirichlet, err := igraph.SampleDirichlet(8, []float64{1, 2, 3}, igraph.LatentSampleOptions{Seed: &seed})
	if err != nil {
		t.Fatal(err)
	}
	again, err := igraph.SampleDirichlet(8, []float64{1, 2, 3}, igraph.LatentSampleOptions{Seed: &seed})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(dirichlet.Rows(), again.Rows()) {
		t.Fatal("Dirichlet same seed differed")
	}
	for i, row := range dirichlet.Rows() {
		sum := 0.0
		for _, v := range row {
			if v < 0 {
				t.Fatalf("row %d negative", i)
			}
			sum += v
		}
		if math.Abs(sum-1) > 1e-12 {
			t.Fatalf("row %d sum=%g", i, sum)
		}
	}
	surface, err := igraph.SampleSphereSurface(6, 3, 2, igraph.LatentSampleOptions{Seed: &seed, Positive: true})
	if err != nil {
		t.Fatal(err)
	}
	for i, row := range surface.Rows() {
		norm := 0.0
		for _, v := range row {
			if v < 0 {
				t.Fatalf("surface row %d negative", i)
			}
			norm += v * v
		}
		if math.Abs(math.Sqrt(norm)-2) > 1e-12 {
			t.Fatalf("surface norm=%g", math.Sqrt(norm))
		}
	}
	volume, err := igraph.SampleSphereVolume(20, 4, 2, igraph.LatentSampleOptions{Seed: &seed})
	if err != nil {
		t.Fatal(err)
	}
	for _, row := range volume.Rows() {
		norm := 0.0
		for _, v := range row {
			norm += v * v
		}
		if math.Sqrt(norm) > 2+1e-12 {
			t.Fatalf("volume norm=%g", math.Sqrt(norm))
		}
	}
	empty, err := igraph.SampleDirichlet(0, []float64{1, 1}, igraph.LatentSampleOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if r, c := empty.Dims(); r != 0 || c != 2 {
		t.Fatalf("empty=%dx%d", r, c)
	}
	emptySurface, err := igraph.SampleSphereSurface(0, 3, 1, igraph.LatentSampleOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if r, c := emptySurface.Dims(); r != 0 || c != 3 {
		t.Fatalf("empty surface=%dx%d", r, c)
	}
	emptyVolume, err := igraph.SampleSphereVolume(0, 2, 1, igraph.LatentSampleOptions{Positive: true})
	if err != nil {
		t.Fatal(err)
	}
	if r, c := emptyVolume.Dims(); r != 0 || c != 2 {
		t.Fatalf("empty volume=%dx%d", r, c)
	}
	for i, call := range []func() error{func() error { _, e := igraph.SampleDirichlet(1, nil, igraph.LatentSampleOptions{}); return e }, func() error {
		_, e := igraph.SampleDirichlet(1, []float64{1, 0}, igraph.LatentSampleOptions{})
		return e
	}, func() error {
		_, e := igraph.SampleDirichlet(1, []float64{1, math.Inf(1)}, igraph.LatentSampleOptions{})
		return e
	}, func() error { _, e := igraph.SampleSphereSurface(1, 1, 1, igraph.LatentSampleOptions{}); return e }, func() error { _, e := igraph.SampleSphereVolume(-1, 2, 1, igraph.LatentSampleOptions{}); return e }, func() error { _, e := igraph.SampleSphereVolume(1, 2, 0, igraph.LatentSampleOptions{}); return e }, func() error {
		_, e := igraph.SampleDirichlet(1, []float64{1, math.NaN()}, igraph.LatentSampleOptions{})
		return e
	}, func() error {
		_, e := igraph.SampleSphereSurface(1, 2, math.Inf(1), igraph.LatentSampleOptions{})
		return e
	}, func() error {
		_, e := igraph.SampleDirichlet(-1, []float64{1, 1}, igraph.LatentSampleOptions{})
		return e
	}, func() error {
		_, e := igraph.SampleDirichlet(1, []float64{-1, 1}, igraph.LatentSampleOptions{})
		return e
	}, func() error { _, e := igraph.SampleSphereSurface(-1, 2, 1, igraph.LatentSampleOptions{}); return e }, func() error {
		_, e := igraph.SampleSphereSurface(1, 2, math.NaN(), igraph.LatentSampleOptions{})
		return e
	}, func() error { _, e := igraph.SampleSphereVolume(1, 1, 1, igraph.LatentSampleOptions{}); return e }} {
		if call() == nil {
			t.Errorf("validation %d accepted", i)
		}
	}
}

func TestDotProductGame(t *testing.T) {
	seed := uint64(340)
	positions := matrix(t, [][]float64{{1, 0}, {0.5, 0.5}, {0, 1}})
	first, err := igraph.DotProductGame(positions, igraph.LatentGraphOptions{Seed: &seed, Directed: true})
	if err != nil {
		t.Fatal(err)
	}
	defer first.Close()
	second, err := igraph.DotProductGame(positions, igraph.LatentGraphOptions{Seed: &seed, Directed: true})
	if err != nil {
		t.Fatal(err)
	}
	defer second.Close()
	if !reflect.DeepEqual(mustEdges(t, first), mustEdges(t, second)) {
		t.Fatal("same seed differed")
	}
	if v, _ := first.VertexCount(); v != 3 {
		t.Fatalf("vertices=%d", v)
	}
	if d, _ := first.IsDirected(); !d {
		t.Fatal("expected directed")
	}
	empty, _ := igraph.NewMatrix(0, 3)
	graph, err := igraph.DotProductGame(empty, igraph.LatentGraphOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer graph.Close()
	if v, _ := graph.VertexCount(); v != 0 {
		t.Fatalf("empty vertices=%d", v)
	}
	if _, e := igraph.DotProductGame(matrix(t, [][]float64{{1, 1}, {1, 1}}), igraph.LatentGraphOptions{}); e == nil {
		t.Fatal("accepted dot product above one")
	}
	if _, e := igraph.DotProductGame(matrix(t, [][]float64{{1}, {-1}}), igraph.LatentGraphOptions{}); e == nil {
		t.Fatal("accepted negative dot product")
	}
	if _, e := igraph.DotProductGame(matrix(t, [][]float64{{math.NaN()}}), igraph.LatentGraphOptions{}); e == nil {
		t.Fatal("accepted non-finite position")
	}
	if _, e := igraph.DotProductGame(matrix(t, [][]float64{{math.Inf(1)}}), igraph.LatentGraphOptions{}); e == nil {
		t.Fatal("accepted infinite position")
	}
	zeroDimensions, _ := igraph.NewMatrix(2, 0)
	zeroGraph, err := igraph.DotProductGame(zeroDimensions, igraph.LatentGraphOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer zeroGraph.Close()
	if v, _ := zeroGraph.VertexCount(); v != 2 {
		t.Fatalf("zero-dimensional vertices=%d", v)
	}
}

func ExampleDotProductGame() {
	positions, err := igraph.SampleDirichlet(6, []float64{1, 1, 1}, igraph.LatentSampleOptions{})
	if err != nil {
		panic(err)
	}
	graph, err := igraph.DotProductGame(positions, igraph.LatentGraphOptions{})
	if err != nil {
		panic(err)
	}
	defer graph.Close()
	vertices, _ := graph.VertexCount()
	fmt.Println(vertices)
	// Output:
	// 6
}
