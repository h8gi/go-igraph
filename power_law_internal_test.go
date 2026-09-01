package igraph

import (
	"errors"
	"math"
	"testing"
)

func TestFitPowerLawFailureAdapters(t *testing.T) {
	data := []float64{1, 2, 3, 4}
	forced := errors.New("forced power-law failure")

	initialization := defaultPowerLawAdapters()
	initialization.newReal = func([]float64) (*realVector, error) { return nil, forced }
	if _, err := fitPowerLaw(data, PowerLawFitOptions{}, &initialization); !errors.Is(err, forced) {
		t.Fatalf("initialization error = %v", err)
	}

	upstream := defaultPowerLawAdapters()
	upstream.fit = func(*realVector, float64, bool) (powerLawRawResult, int) { return powerLawRawResult{}, 4 }
	if _, err := fitPowerLaw(data, PowerLawFitOptions{}, &upstream); err == nil {
		t.Fatal("upstream error not propagated")
	}

	malformed := defaultPowerLawAdapters()
	malformed.fit = func(*realVector, float64, bool) (powerLawRawResult, int) {
		return powerLawRawResult{alpha: math.NaN(), xmin: 1, logLikelihood: -1, ksStatistic: 0.1}, 0
	}
	if _, err := fitPowerLaw(data, PowerLawFitOptions{}, &malformed); err == nil {
		t.Fatal("malformed fit accepted")
	}
}

func TestPowerLawPValueFailureAdapters(t *testing.T) {
	fit, err := FitPowerLaw([]float64{1, 2, 3, 4}, PowerLawFitOptions{})
	if err != nil {
		t.Fatal(err)
	}
	forced := errors.New("forced p-value failure")

	initialization := defaultPowerLawAdapters()
	initialization.newReal = func([]float64) (*realVector, error) { return nil, forced }
	if _, err := fit.powerLawPValue(PowerLawPValueOptions{Precision: 0.5}, &initialization); !errors.Is(err, forced) {
		t.Fatalf("initialization error = %v", err)
	}

	upstream := defaultPowerLawAdapters()
	upstream.pValue = func(*realVector, powerLawRawResult, float64, *uint64) (float64, int) { return 0, 4 }
	if _, err := fit.powerLawPValue(PowerLawPValueOptions{Precision: 0.5}, &upstream); err == nil {
		t.Fatal("upstream error not propagated")
	}

	malformed := defaultPowerLawAdapters()
	malformed.pValue = func(*realVector, powerLawRawResult, float64, *uint64) (float64, int) { return 2, 0 }
	if _, err := fit.powerLawPValue(PowerLawPValueOptions{Precision: 0.5}, &malformed); err == nil {
		t.Fatal("malformed p-value accepted")
	}
}
