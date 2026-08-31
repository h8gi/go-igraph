package igraph_test

import (
	"math"
	"reflect"
	"sync"
	"testing"

	"github.com/h8gi/go-igraph"
)

func layoutFloat(value float64) *float64 { return &value }

func assertAdvancedLayout(t *testing.T, name string, matrix igraph.Matrix, vertices int) {
	t.Helper()
	rows, columns := matrix.Dims()
	if rows != vertices || columns != 2 {
		t.Fatalf("%s dimensions = (%d, %d), want (%d, 2)", name, rows, columns, vertices)
	}
	for row, values := range matrix.Rows() {
		for column, value := range values {
			if math.IsNaN(value) || math.IsInf(value, 0) {
				t.Fatalf("%s coordinate (%d, %d) is not finite: %v", name, row, column, value)
			}
		}
	}
}

func TestClassicIterativeLayoutsConcurrentAndCloseRace(t *testing.T) {
	graph, err := igraph.NewRing(6, false, false)
	if err != nil {
		t.Fatal(err)
	}
	defer graph.Close()
	seed := uint64(7)
	calls := []func() error{
		func() error {
			_, err := graph.LayoutDavidsonHarel(igraph.DavidsonHarelOptions{Seed: &seed, MaxIter: 1, FineIter: 1})
			return err
		},
		func() error { _, err := graph.LayoutGEM(igraph.GEMOptions{Seed: &seed, MaxIter: 10}); return err },
		func() error {
			_, err := graph.LayoutGraphopt(igraph.GraphoptOptions{Seed: &seed, NIter: 2})
			return err
		},
	}
	var wait sync.WaitGroup
	for _, call := range calls {
		call := call
		wait.Add(1)
		go func() {
			defer wait.Done()
			if err := call(); err != nil {
				t.Errorf("concurrent layout failed: %v", err)
			}
		}()
	}
	wait.Wait()

	racing, err := igraph.NewRing(20, false, false)
	if err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() {
		_, err := racing.LayoutGEM(igraph.GEMOptions{MaxIter: 200})
		done <- err
	}()
	racing.Close()
	if err := <-done; err != nil && err != igraph.ErrClosed {
		t.Fatalf("close-race layout error = %v", err)
	}
}

func TestClassicIterativeLayouts(t *testing.T) {
	graph, err := igraph.NewRing(5, false, false)
	if err != nil {
		t.Fatal(err)
	}
	defer graph.Close()
	seed := uint64(42)

	tests := []struct {
		name string
		run  func() (igraph.Matrix, error)
	}{
		{"Davidson-Harel", func() (igraph.Matrix, error) {
			return graph.LayoutDavidsonHarel(igraph.DavidsonHarelOptions{Seed: &seed, MaxIter: 1, FineIter: 1})
		}},
		{"GEM", func() (igraph.Matrix, error) {
			return graph.LayoutGEM(igraph.GEMOptions{Seed: &seed, MaxIter: 20})
		}},
		{"Graphopt", func() (igraph.Matrix, error) {
			return graph.LayoutGraphopt(igraph.GraphoptOptions{Seed: &seed, NIter: 5})
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			first, err := test.run()
			if err != nil {
				t.Fatalf("first call failed: %v", err)
			}
			second, err := test.run()
			if err != nil {
				t.Fatalf("second call failed: %v", err)
			}
			assertAdvancedLayout(t, test.name, first, 5)
			if !reflect.DeepEqual(first.Rows(), second.Rows()) {
				t.Fatalf("seeded %s layouts differ", test.name)
			}
		})
	}
}

func TestClassicIterativeLayoutsInitialCoordinatesAndOwnership(t *testing.T) {
	graph, err := igraph.NewRing(4, true, false)
	if err != nil {
		t.Fatal(err)
	}
	initial, err := igraph.NewMatrixFromRows([][]float64{{-1, 0}, {0, 1}, {1, 0}, {0, -1}})
	if err != nil {
		t.Fatal(err)
	}
	wantInitial := initial.Rows()

	results := []struct {
		name string
		m    igraph.Matrix
		err  error
	}{
		func() struct {
			name string
			m    igraph.Matrix
			err  error
		} {
			m, err := graph.LayoutDavidsonHarel(igraph.DavidsonHarelOptions{InitialCoordinates: &initial, MaxIter: 1, FineIter: 1})
			return struct {
				name string
				m    igraph.Matrix
				err  error
			}{"Davidson-Harel", m, err}
		}(),
		func() struct {
			name string
			m    igraph.Matrix
			err  error
		} {
			m, err := graph.LayoutGEM(igraph.GEMOptions{InitialCoordinates: &initial, MaxIter: 10})
			return struct {
				name string
				m    igraph.Matrix
				err  error
			}{"GEM", m, err}
		}(),
		func() struct {
			name string
			m    igraph.Matrix
			err  error
		} {
			m, err := graph.LayoutGraphopt(igraph.GraphoptOptions{InitialCoordinates: &initial, NIter: 2})
			return struct {
				name string
				m    igraph.Matrix
				err  error
			}{"Graphopt", m, err}
		}(),
	}
	if !reflect.DeepEqual(initial.Rows(), wantInitial) {
		t.Fatal("initial coordinates were mutated")
	}
	graph.Close()
	for _, result := range results {
		if result.err != nil {
			t.Fatalf("%s failed: %v", result.name, result.err)
		}
		assertAdvancedLayout(t, result.name, result.m, 4)
	}
	if _, err := graph.LayoutGEM(igraph.GEMOptions{}); err != igraph.ErrClosed {
		t.Fatalf("post-Close error = %v, want ErrClosed", err)
	}
	if _, err := graph.LayoutDavidsonHarel(igraph.DavidsonHarelOptions{}); err != igraph.ErrClosed {
		t.Fatalf("post-Close Davidson-Harel error = %v, want ErrClosed", err)
	}
	if _, err := graph.LayoutGraphopt(igraph.GraphoptOptions{}); err != igraph.ErrClosed {
		t.Fatalf("post-Close Graphopt error = %v, want ErrClosed", err)
	}
}

func TestClassicIterativeLayoutsEmptyAndInvalid(t *testing.T) {
	empty, err := igraph.NewGraph()
	if err != nil {
		t.Fatal(err)
	}
	defer empty.Close()
	for name, call := range map[string]func() (igraph.Matrix, error){
		"Davidson-Harel": func() (igraph.Matrix, error) { return empty.LayoutDavidsonHarel(igraph.DavidsonHarelOptions{}) },
		"GEM":            func() (igraph.Matrix, error) { return empty.LayoutGEM(igraph.GEMOptions{}) },
		"Graphopt":       func() (igraph.Matrix, error) { return empty.LayoutGraphopt(igraph.GraphoptOptions{}) },
	} {
		matrix, err := call()
		if err != nil {
			t.Fatalf("%s empty failed: %v", name, err)
		}
		assertAdvancedLayout(t, name, matrix, 0)
	}

	graph, err := igraph.NewRing(3, false, false)
	if err != nil {
		t.Fatal(err)
	}
	defer graph.Close()
	badShape, _ := igraph.NewMatrix(2, 2)
	badFinite, _ := igraph.NewMatrixFromRows([][]float64{{0, 0}, {1, math.NaN()}, {2, 0}})
	invalid := []func() error{
		func() error {
			_, err := graph.LayoutDavidsonHarel(igraph.DavidsonHarelOptions{MaxIter: -1})
			return err
		},
		func() error {
			_, err := graph.LayoutDavidsonHarel(igraph.DavidsonHarelOptions{CoolFactor: 1})
			return err
		},
		func() error {
			_, err := graph.LayoutDavidsonHarel(igraph.DavidsonHarelOptions{WeightBorder: layoutFloat(math.Inf(1))})
			return err
		},
		func() error {
			_, err := graph.LayoutGEM(igraph.GEMOptions{MinTemperature: 2, InitialTemperature: 1, MaxTemperature: 3})
			return err
		},
		func() error { _, err := graph.LayoutGEM(igraph.GEMOptions{MaxIter: -1}); return err },
		func() error { _, err := graph.LayoutGEM(igraph.GEMOptions{MaxTemperature: math.NaN()}); return err },
		func() error { _, err := graph.LayoutGEM(igraph.GEMOptions{InitialCoordinates: &badShape}); return err },
		func() error { _, err := graph.LayoutGraphopt(igraph.GraphoptOptions{NodeMass: -1}); return err },
		func() error { _, err := graph.LayoutGraphopt(igraph.GraphoptOptions{NIter: -1}); return err },
		func() error {
			_, err := graph.LayoutGraphopt(igraph.GraphoptOptions{NodeCharge: layoutFloat(math.Inf(1))})
			return err
		},
		func() error {
			_, err := graph.LayoutGraphopt(igraph.GraphoptOptions{InitialCoordinates: &badFinite})
			return err
		},
	}
	for index, call := range invalid {
		if err := call(); err == nil {
			t.Errorf("invalid case %d succeeded", index)
		}
	}
	var nilGraph *igraph.Graph
	if _, err := nilGraph.LayoutDavidsonHarel(igraph.DavidsonHarelOptions{}); err != igraph.ErrClosed {
		t.Fatalf("nil graph error = %v, want ErrClosed", err)
	}
	if _, err := nilGraph.LayoutGEM(igraph.GEMOptions{}); err != igraph.ErrClosed {
		t.Fatalf("nil GEM graph error = %v, want ErrClosed", err)
	}
	if _, err := nilGraph.LayoutGraphopt(igraph.GraphoptOptions{}); err != igraph.ErrClosed {
		t.Fatalf("nil Graphopt graph error = %v, want ErrClosed", err)
	}
}
