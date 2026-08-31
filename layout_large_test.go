package igraph_test

import (
	"math"
	"reflect"
	"testing"

	"github.com/h8gi/go-igraph"
)

func TestLayoutDrL(t *testing.T) {
	graph, err := igraph.NewRing(6, false, false)
	if err != nil {
		t.Fatal(err)
	}
	defer graph.Close()
	seed := uint64(123)
	options := igraph.DrLOptions{Seed: &seed, Weights: []float64{1, 2, 1, 2, 1, 2}}
	first, err := graph.LayoutDrL(options)
	if err != nil {
		t.Fatalf("LayoutDrL failed: %v", err)
	}
	second, err := graph.LayoutDrL(options)
	if err != nil {
		t.Fatalf("LayoutDrL replay failed: %v", err)
	}
	assertAdvancedLayout(t, "DrL", first, 6)
	if !reflect.DeepEqual(first.Rows(), second.Rows()) {
		t.Fatal("seeded DrL layouts differ")
	}

	threeD, err := graph.LayoutDrL3D(igraph.DrLOptions{Seed: &seed, Preset: igraph.DrLCoarsen})
	if err != nil {
		t.Fatalf("LayoutDrL3D failed: %v", err)
	}
	rows, columns := threeD.Dims()
	if rows != 6 || columns != 3 {
		t.Fatalf("DrL3D dimensions = (%d, %d), want (6, 3)", rows, columns)
	}
	for _, row := range threeD.Rows() {
		for _, value := range row {
			if math.IsNaN(value) || math.IsInf(value, 0) {
				t.Fatalf("non-finite DrL3D coordinate: %v", value)
			}
		}
	}
}

func TestLayoutDrLInitialAndInvalid(t *testing.T) {
	graph, err := igraph.NewRing(4, false, false)
	if err != nil {
		t.Fatal(err)
	}
	initial, _ := igraph.NewMatrixFromRows([][]float64{{-1, 0}, {0, 1}, {1, 0}, {0, -1}})
	want := initial.Rows()
	result, err := graph.LayoutDrL(igraph.DrLOptions{Preset: igraph.DrLFinal, InitialCoordinates: &initial})
	if err != nil {
		t.Fatalf("seeded DrL failed: %v", err)
	}
	assertAdvancedLayout(t, "DrL final", result, 4)
	if !reflect.DeepEqual(initial.Rows(), want) {
		t.Fatal("DrL mutated initial coordinates")
	}

	badShape, _ := igraph.NewMatrix(4, 2)
	badFinite, _ := igraph.NewMatrixFromRows([][]float64{{0, 0}, {1, 0}, {2, math.NaN()}, {3, 0}})
	invalid := []func() error{
		func() error { _, err := graph.LayoutDrL(igraph.DrLOptions{Preset: igraph.DrLPreset(99)}); return err },
		func() error { _, err := graph.LayoutDrL(igraph.DrLOptions{Weights: []float64{1}}); return err },
		func() error {
			_, err := graph.LayoutDrL(igraph.DrLOptions{Weights: []float64{1, 1, -1, 1}})
			return err
		},
		func() error { _, err := graph.LayoutDrL(igraph.DrLOptions{InitialCoordinates: &badFinite}); return err },
		func() error {
			_, err := graph.LayoutDrL3D(igraph.DrLOptions{InitialCoordinates: &badShape})
			return err
		},
	}
	for index, call := range invalid {
		if err := call(); err == nil {
			t.Errorf("invalid DrL case %d succeeded", index)
		}
	}
	graph.Close()
	if _, err := graph.LayoutDrL(igraph.DrLOptions{}); err != igraph.ErrClosed {
		t.Fatalf("closed DrL error = %v", err)
	}
	var nilGraph *igraph.Graph
	if _, err := nilGraph.LayoutDrL3D(igraph.DrLOptions{}); err != igraph.ErrClosed {
		t.Fatalf("nil DrL3D error = %v", err)
	}
}

func TestLayoutDrLEmpty(t *testing.T) {
	graph, err := igraph.NewGraph()
	if err != nil {
		t.Fatal(err)
	}
	defer graph.Close()
	for dimensions, call := range map[int]func() (igraph.Matrix, error){2: func() (igraph.Matrix, error) { return graph.LayoutDrL(igraph.DrLOptions{}) }, 3: func() (igraph.Matrix, error) { return graph.LayoutDrL3D(igraph.DrLOptions{}) }} {
		matrix, err := call()
		if err != nil {
			t.Fatalf("empty %dD DrL failed: %v", dimensions, err)
		}
		rows, columns := matrix.Dims()
		if rows != 0 || columns != dimensions {
			t.Fatalf("empty DrL dimensions = (%d, %d), want (0, %d)", rows, columns, dimensions)
		}
	}
}

func TestLayoutLGL(t *testing.T) {
	graph, err := igraph.NewRing(8, false, false)
	if err != nil {
		t.Fatal(err)
	}
	defer graph.Close()
	seed := uint64(81)
	first, err := graph.LayoutLGL(igraph.LGLOptions{Seed: &seed, MaxIter: 5})
	if err != nil {
		t.Fatalf("LayoutLGL failed: %v", err)
	}
	second, err := graph.LayoutLGL(igraph.LGLOptions{Seed: &seed, MaxIter: 5})
	if err != nil {
		t.Fatalf("LayoutLGL replay failed: %v", err)
	}
	assertAdvancedLayout(t, "LGL", first, 8)
	if !reflect.DeepEqual(first.Rows(), second.Rows()) {
		t.Fatal("seeded LGL layouts differ")
	}
	root := 3
	if _, err := graph.LayoutLGL(igraph.LGLOptions{Root: &root, MaxIter: 2}); err != nil {
		t.Fatalf("rooted LGL failed: %v", err)
	}
}

func TestLayoutLGLEmptyInvalidAndClosed(t *testing.T) {
	empty, _ := igraph.NewGraph()
	matrix, err := empty.LayoutLGL(igraph.LGLOptions{})
	if err != nil {
		t.Fatalf("empty LGL failed: %v", err)
	}
	assertAdvancedLayout(t, "empty LGL", matrix, 0)
	empty.Close()
	if _, err := empty.LayoutLGL(igraph.LGLOptions{}); err != igraph.ErrClosed {
		t.Fatalf("closed LGL error = %v", err)
	}

	graph, _ := igraph.NewRing(3, false, false)
	defer graph.Close()
	badRoot := 3
	invalid := []igraph.LGLOptions{{MaxIter: -1}, {MaxDelta: -1}, {Area: math.Inf(1)}, {CoolingExponent: -1}, {RepulsionRadius: math.NaN()}, {CellSize: -1}, {Root: &badRoot}}
	for index, options := range invalid {
		if _, err := graph.LayoutLGL(options); err == nil {
			t.Errorf("invalid LGL case %d succeeded", index)
		}
	}
	var nilGraph *igraph.Graph
	if _, err := nilGraph.LayoutLGL(igraph.LGLOptions{}); err != igraph.ErrClosed {
		t.Fatalf("nil LGL error = %v", err)
	}
}
