package igraph

import (
	"errors"
	"testing"
)

func TestExpectedDegreeRandomInitializationAndUpstreamFailures(t *testing.T) {
	forced := errors.New("forced expected-degree random failure")
	base := func() expectedDegreeRandomAdapters {
		return expectedDegreeRandomAdapters{
			newReal:   func([]float64) (*realVector, error) { return &realVector{}, nil },
			closeReal: func(*realVector) {},
			chungLu: func(*realVector, *realVector, ChungLuOptions) expectedDegreeGraphCallResult {
				return expectedDegreeGraphCallResult{code: 1}
			},
			staticFitness: func(int, *realVector, *realVector, StaticFitnessOptions) expectedDegreeGraphCallResult {
				return expectedDegreeGraphCallResult{code: 1}
			},
			staticPower: func(int, int, float64, StaticPowerLawOptions) expectedDegreeGraphCallResult {
				return expectedDegreeGraphCallResult{code: 1}
			},
		}
	}

	t.Run("first vector initialization", func(t *testing.T) {
		for _, call := range []func(*expectedDegreeRandomAdapters) error{
			func(adapters *expectedDegreeRandomAdapters) error {
				_, err := chungLuGame([]float64{1}, nil, ChungLuOptions{}, adapters)
				return err
			},
			func(adapters *expectedDegreeRandomAdapters) error {
				_, err := staticFitnessGame(0, []float64{1}, nil, StaticFitnessOptions{}, adapters)
				return err
			},
		} {
			adapters := base()
			adapters.newReal = func([]float64) (*realVector, error) { return nil, forced }
			if err := call(&adapters); !errors.Is(err, forced) {
				t.Fatalf("error = %v, want injected failure", err)
			}
		}
	})

	t.Run("second vector initialization closes first", func(t *testing.T) {
		for _, call := range []func(*expectedDegreeRandomAdapters) error{
			func(adapters *expectedDegreeRandomAdapters) error {
				_, err := chungLuGame([]float64{1}, []float64{1}, ChungLuOptions{}, adapters)
				return err
			},
			func(adapters *expectedDegreeRandomAdapters) error {
				_, err := staticFitnessGame(0, []float64{1}, []float64{1}, StaticFitnessOptions{}, adapters)
				return err
			},
		} {
			adapters := base()
			calls, closes := 0, 0
			adapters.newReal = func([]float64) (*realVector, error) {
				calls++
				if calls == 2 {
					return nil, forced
				}
				return &realVector{}, nil
			}
			adapters.closeReal = func(*realVector) { closes++ }
			if err := call(&adapters); !errors.Is(err, forced) {
				t.Fatalf("error = %v, want injected failure", err)
			}
			if closes != 1 {
				t.Fatalf("closed vectors = %d, want 1", closes)
			}
		}
	})

	t.Run("upstream", func(t *testing.T) {
		tests := []struct {
			name string
			call func(*expectedDegreeRandomAdapters) (*Graph, error)
		}{
			{name: "Chung-Lu", call: func(adapters *expectedDegreeRandomAdapters) (*Graph, error) {
				return chungLuGame([]float64{1}, nil, ChungLuOptions{}, adapters)
			}},
			{name: "static fitness", call: func(adapters *expectedDegreeRandomAdapters) (*Graph, error) {
				return staticFitnessGame(0, []float64{1}, nil, StaticFitnessOptions{}, adapters)
			}},
			{name: "static power law", call: func(adapters *expectedDegreeRandomAdapters) (*Graph, error) {
				return staticPowerLawGame(1, 0, 2, StaticPowerLawOptions{}, adapters)
			}},
		}
		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				adapters := base()
				graph, err := test.call(&adapters)
				if err == nil || graph != nil {
					t.Fatalf("call = %v, %v, want nil error result", graph, err)
				}
			})
		}
	})
}

func TestExpectedDegreeRandomValidationHelpers(t *testing.T) {
	if capacity := staticFitnessPairCapacity([]float64{1, 1, 0}, nil, false); capacity != 1 {
		t.Fatalf("undirected capacity = %d, want 1", capacity)
	}
	if capacity := staticFitnessPairCapacity([]float64{1, 1}, nil, true); capacity != 3 {
		t.Fatalf("undirected loop capacity = %d, want 3", capacity)
	}
	if capacity := staticFitnessPairCapacity([]float64{1, 1}, []float64{1, 0}, false); capacity != 1 {
		t.Fatalf("directed capacity = %d, want 1", capacity)
	}
	if got := saturatedProduct(int(^uint(0)>>1), 2); got != int(^uint(0)>>1) {
		t.Fatalf("saturated product = %d", got)
	}
}
