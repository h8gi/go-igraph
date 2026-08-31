package igraph

import (
	"errors"
	"math"
	"reflect"
	"sync"
	"testing"
)

func TestHRGConsensusSeedReplayAndWarmStart(t *testing.T) {
	g := newHRGFitGraph(t)
	defer g.Close()
	seed := uint64(42)
	options := HRGAnalysisOptions{Samples: 100, Seed: &seed}
	first, err := g.ConsensusHRG(options)
	if err != nil {
		t.Fatal(err)
	}
	second, err := g.ConsensusHRG(options)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("seed replay differs: %#v != %#v", first, second)
	}
	if len(first.Parents) < 6 || first.Model.LeafCount() != 6 {
		t.Fatalf("consensus = %#v", first)
	}
	before := first.Model.Probabilities()
	warmOptions := options
	warmOptions.StartingModel = &first.Model
	warm, err := g.ConsensusHRG(warmOptions)
	if err != nil {
		t.Fatal(err)
	}
	if warm.Model.LeafCount() != 6 || !reflect.DeepEqual(first.Model.Probabilities(), before) {
		t.Fatal("warm consensus changed starting model")
	}
	if err := g.Close(); err != nil {
		t.Fatal(err)
	}
	if len(first.Parents) == 0 {
		t.Fatal("Go-owned consensus changed after close")
	}
}

func TestHRGPredictionSeedReplayAndAlignment(t *testing.T) {
	g := newHRGFitGraph(t)
	defer g.Close()
	seed := uint64(42)
	options := HRGPredictionOptions{Samples: 100, Bins: 10, Seed: &seed}
	first, err := g.PredictHRG(options)
	if err != nil {
		t.Fatal(err)
	}
	second, err := g.PredictHRG(options)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("seed replay differs: %#v != %#v", first, second)
	}
	if len(first.Edges) == 0 || len(first.Edges) != len(first.Probabilities) || first.Model.LeafCount() != 6 {
		t.Fatalf("prediction = %#v", first)
	}
	existing, err := g.Edges()
	if err != nil {
		t.Fatal(err)
	}
	present := map[Edge]bool{}
	for _, edge := range existing {
		if edge.From > edge.To {
			edge.From, edge.To = edge.To, edge.From
		}
		present[edge] = true
	}
	for _, edge := range first.Edges {
		if edge.From > edge.To {
			edge.From, edge.To = edge.To, edge.From
		}
		if present[edge] {
			t.Fatalf("predicted existing edge %#v", edge)
		}
	}
}

func TestHRGAnalysisValidation(t *testing.T) {
	if _, err := (*Graph)(nil).ConsensusHRG(HRGAnalysisOptions{Samples: 1}); !errors.Is(err, ErrClosed) {
		t.Fatalf("nil consensus error = %v", err)
	}
	if _, err := (*Graph)(nil).PredictHRG(HRGPredictionOptions{Samples: 1, Bins: 1}); !errors.Is(err, ErrClosed) {
		t.Fatalf("nil prediction error = %v", err)
	}
	g := newHRGFitGraph(t)
	defer g.Close()
	if _, err := g.ConsensusHRG(HRGAnalysisOptions{}); err == nil {
		t.Fatal("expected sample error")
	}
	if _, err := g.PredictHRG(HRGPredictionOptions{Samples: 1}); err == nil {
		t.Fatal("expected bin error")
	}
	wrong := testHRGModel(t)
	if _, err := g.ConsensusHRG(HRGAnalysisOptions{Samples: 1, StartingModel: &wrong}); err == nil {
		t.Fatal("expected alignment error")
	}
	if _, err := g.PredictHRG(HRGPredictionOptions{Samples: 1, Bins: 1, StartingModel: &wrong}); err == nil {
		t.Fatal("expected alignment error")
	}
	directed, err := NewGraphFromEdges(2, []Edge{{0, 1}}, true)
	if err != nil {
		t.Fatal(err)
	}
	defer directed.Close()
	if _, err := directed.ConsensusHRG(HRGAnalysisOptions{Samples: 1}); err == nil {
		t.Fatal("expected directed error")
	}
	looped, err := NewGraphFromEdges(2, []Edge{{0, 0}}, false)
	if err != nil {
		t.Fatal(err)
	}
	defer looped.Close()
	if _, err := looped.PredictHRG(HRGPredictionOptions{Samples: 1, Bins: 1}); err == nil {
		t.Fatal("expected simplicity error")
	}
}

func TestValidateHRGConsensus(t *testing.T) {
	valid := HRGConsensus{Parents: []int{2, 2, -1}, Weights: []float64{1}}
	if err := validateHRGConsensus(valid, 2, 1); err != nil {
		t.Fatal(err)
	}
	for name, result := range map[string]HRGConsensus{
		"nil": {}, "length": {Parents: []int{-1}, Weights: []float64{0}}, "parent": {Parents: []int{0}, Weights: []float64{}}, "roots": {Parents: []int{-1, -1}, Weights: []float64{0}}, "weight": {Parents: []int{1, -1}, Weights: []float64{2}}, "cycle": {Parents: []int{1, 0}, Weights: []float64{0}},
	} {
		t.Run(name, func(t *testing.T) {
			if err := validateHRGConsensus(result, 1, 1); err == nil {
				t.Fatal("expected error")
			}
		})
	}
}

func TestHRGAnalysisBoundaryGraphs(t *testing.T) {
	seed := uint64(7)
	empty, err := NewGraph()
	if err != nil {
		t.Fatal(err)
	}
	defer empty.Close()
	if _, err := empty.ConsensusHRG(HRGAnalysisOptions{Samples: 1, Seed: &seed}); err == nil {
		t.Fatal("expected empty consensus error")
	}
	if _, err := empty.PredictHRG(HRGPredictionOptions{Samples: 1, Bins: 1, Seed: &seed}); err == nil {
		t.Fatal("expected empty prediction error")
	}
	singleton, err := NewGraphFromEdges(1, nil, false)
	if err != nil {
		t.Fatal(err)
	}
	defer singleton.Close()
	if _, err := singleton.ConsensusHRG(HRGAnalysisOptions{Samples: 10, Seed: &seed}); err == nil {
		t.Fatal("expected singleton consensus error")
	}
	if _, err := singleton.PredictHRG(HRGPredictionOptions{Samples: 10, Bins: 2, Seed: &seed}); err == nil {
		t.Fatal("expected singleton prediction error")
	}
	edgeless, err := NewGraphFromEdges(4, nil, false)
	if err != nil {
		t.Fatal(err)
	}
	defer edgeless.Close()
	if _, err := edgeless.ConsensusHRG(HRGAnalysisOptions{Samples: 10, Seed: &seed}); err != nil {
		t.Fatalf("edgeless consensus: %v", err)
	}
	if _, err := edgeless.PredictHRG(HRGPredictionOptions{Samples: 10, Bins: 2, Seed: &seed}); err != nil {
		t.Fatalf("edgeless prediction: %v", err)
	}

	complete, err := NewFull(4, false, false)
	if err != nil {
		t.Fatal(err)
	}
	defer complete.Close()
	prediction, err := complete.PredictHRG(HRGPredictionOptions{Samples: 10, Bins: 2, Seed: &seed})
	if err != nil {
		t.Fatal(err)
	}
	if prediction.Edges == nil || prediction.Probabilities == nil || len(prediction.Edges) != 0 || len(prediction.Probabilities) != 0 {
		t.Fatalf("complete prediction = %#v", prediction)
	}

	disconnected, err := NewGraphFromEdges(4, []Edge{{0, 1}, {2, 3}}, false)
	if err != nil {
		t.Fatal(err)
	}
	defer disconnected.Close()
	if _, err := disconnected.ConsensusHRG(HRGAnalysisOptions{Samples: 10, Seed: &seed}); err != nil {
		t.Fatalf("disconnected consensus: %v", err)
	}
	if _, err := disconnected.PredictHRG(HRGPredictionOptions{Samples: 10, Bins: 2, Seed: &seed}); err != nil {
		t.Fatalf("disconnected prediction: %v", err)
	}

	parallel, err := NewGraphFromEdges(2, []Edge{{0, 1}, {0, 1}}, false)
	if err != nil {
		t.Fatal(err)
	}
	defer parallel.Close()
	if _, err := parallel.ConsensusHRG(HRGAnalysisOptions{Samples: 1}); err == nil {
		t.Fatal("expected parallel-edge error")
	}
}

func TestHRGAnalysisFailuresCloseResources(t *testing.T) {
	g := newHRGFitGraph(t)
	defer g.Close()
	forced := errors.New("forced failure")
	for _, prediction := range []bool{false, true} {
		name := "consensus"
		if prediction {
			name = "prediction"
		}
		t.Run(name, func(t *testing.T) {
			base := defaultHRGAnalysisAdapters()
			for failure, mutate := range map[string]func(*hrgAnalysisAdapters){
				"model-init": func(a *hrgAnalysisAdapters) { a.fresh = func(int) (*cHRG, error) { return nil, forced } },
				"int-init":   func(a *hrgAnalysisAdapters) { a.newInt = func() (*intVector, error) { return nil, forced } },
				"real-init":  func(a *hrgAnalysisAdapters) { a.newReal = func() (*realVector, error) { return nil, forced } },
				"upstream": func(a *hrgAnalysisAdapters) {
					if prediction {
						a.predict = func(*Graph, *intVector, *realVector, *cHRG, bool, int, int) error { return forced }
					} else {
						a.consensus = func(*Graph, *intVector, *realVector, *cHRG, bool, int) error { return forced }
					}
				},
				"conversion":    func(a *hrgAnalysisAdapters) { a.readInt = func(*intVector) ([]int, error) { return nil, forced } },
				"model-extract": func(a *hrgAnalysisAdapters) { a.extract = func(*cHRG) (HRGModel, error) { return HRGModel{}, forced } },
			} {
				t.Run(failure, func(t *testing.T) {
					a := base
					modelCloses, intCloses, realCloses := 0, 0, 0
					a.closeModel = func(h *cHRG) { modelCloses++; h.close() }
					a.closeInt = func(v *intVector) { intCloses++; v.close() }
					a.closeReal = func(v *realVector) { realCloses++; v.close() }
					mutate(&a)
					var err error
					if prediction {
						_, err = g.predictHRG(HRGPredictionOptions{Samples: 1, Bins: 1}, &a)
					} else {
						_, err = g.consensusHRG(HRGAnalysisOptions{Samples: 1}, &a)
					}
					if !errors.Is(err, forced) {
						t.Fatalf("error = %v", err)
					}
					wantModel, wantInt, wantReal := 1, 1, 1
					if failure == "model-init" {
						wantModel, wantInt, wantReal = 0, 0, 0
					}
					if failure == "int-init" {
						wantInt, wantReal = 0, 0
					}
					if failure == "real-init" {
						wantReal = 0
					}
					if modelCloses != wantModel || intCloses != wantInt || realCloses != wantReal {
						t.Fatalf("closes = model %d, int %d, real %d; want %d, %d, %d", modelCloses, intCloses, realCloses, wantModel, wantInt, wantReal)
					}
				})
			}
		})
	}
}

func TestHRGAnalysisConcurrentCallsAndClose(t *testing.T) {
	g := newHRGFitGraph(t)
	seed := uint64(42)
	var wg sync.WaitGroup
	for i := range 6 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var err error
			if i%2 == 0 {
				_, err = g.ConsensusHRG(HRGAnalysisOptions{Samples: 20, Seed: &seed})
			} else {
				_, err = g.PredictHRG(HRGPredictionOptions{Samples: 20, Bins: 4, Seed: &seed})
			}
			if err != nil && !errors.Is(err, ErrClosed) {
				t.Errorf("analysis close race: %v", err)
			}
		}()
	}
	wg.Add(1)
	go func() { defer wg.Done(); _ = g.Close() }()
	wg.Wait()
	if _, err := g.ConsensusHRG(HRGAnalysisOptions{Samples: 1}); !errors.Is(err, ErrClosed) {
		t.Fatalf("closed consensus error = %v", err)
	}
	if _, err := g.PredictHRG(HRGPredictionOptions{Samples: 1, Bins: 1}); !errors.Is(err, ErrClosed) {
		t.Fatalf("closed prediction error = %v", err)
	}
}

func TestHRGPredictionRejectsMalformedUpstreamOutput(t *testing.T) {
	g := newHRGFitGraph(t)
	defer g.Close()
	for name, output := range map[string]struct {
		endpoints []int
		prob      []float64
	}{
		"odd endpoints":          {[]int{0}, []float64{}},
		"length mismatch":        {[]int{0, 2}, []float64{}},
		"invalid endpoint":       {[]int{0, 6}, []float64{0.5}},
		"loop":                   {[]int{1, 1}, []float64{0.5}},
		"non-finite probability": {[]int{0, 2}, []float64{math.NaN()}},
		"probability range":      {[]int{0, 2}, []float64{1.1}},
	} {
		t.Run(name, func(t *testing.T) {
			a := defaultHRGAnalysisAdapters()
			a.readInt = func(*intVector) ([]int, error) { return output.endpoints, nil }
			a.readReal = func(*realVector) ([]float64, error) { return output.prob, nil }
			a.extract = func(*cHRG) (HRGModel, error) { return testHRGModel(t), nil }
			if _, err := g.predictHRG(HRGPredictionOptions{Samples: 1, Bins: 1}, &a); err == nil {
				t.Fatal("expected malformed-output error")
			}
		})
	}
}
