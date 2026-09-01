package igraph

import (
	"math"
	"testing"
)

func TestFitPowerLawDiscreteContinuousAndOwnership(t *testing.T) {
	discreteData := []float64{1, 1, 2, 2, 3, 4, 6, 10, 20, 40}
	discrete, err := FitPowerLaw(discreteData, PowerLawFitOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if discrete.Continuous || discrete.Alpha <= 1 || discrete.XMin < 0 || discrete.KSStatistic < 0 {
		t.Fatalf("discrete fit = %+v", discrete)
	}
	discreteData[0] = 999
	if discrete.data[0] != 1 {
		t.Fatalf("fit retained caller storage: %v", discrete.data)
	}

	xmin := 1.0
	continuous, err := FitPowerLaw([]float64{1, 1.5, 2, 2.5, 4, 8, 16}, PowerLawFitOptions{XMin: &xmin, ForceContinuous: true})
	if err != nil {
		t.Fatal(err)
	}
	if !continuous.Continuous || continuous.XMin != xmin {
		t.Fatalf("continuous fit = %+v", continuous)
	}
}

func TestPowerLawPValueSeedReplay(t *testing.T) {
	fit, err := FitPowerLaw([]float64{1, 1, 2, 2, 3, 4, 6, 10, 20, 40}, PowerLawFitOptions{})
	if err != nil {
		t.Fatal(err)
	}
	seed := uint64(8128)
	options := PowerLawPValueOptions{Precision: 0.25, Seed: &seed}
	first, err := fit.PValue(options)
	if err != nil {
		t.Fatal(err)
	}
	second, err := fit.PValue(options)
	if err != nil {
		t.Fatal(err)
	}
	if first != second || first < 0 || first > 1 {
		t.Fatalf("p-values = %v, %v", first, second)
	}
}

func TestFitPowerLawValidation(t *testing.T) {
	xminNegative := -1.0
	xminNaN := math.NaN()
	for name, call := range map[string]func() error{
		"nil":             func() error { _, err := FitPowerLaw(nil, PowerLawFitOptions{}); return err },
		"singleton":       func() error { _, err := FitPowerLaw([]float64{1}, PowerLawFitOptions{}); return err },
		"nan sample":      func() error { _, err := FitPowerLaw([]float64{1, math.NaN()}, PowerLawFitOptions{}); return err },
		"infinite sample": func() error { _, err := FitPowerLaw([]float64{1, math.Inf(1)}, PowerLawFitOptions{}); return err },
		"negative XMin": func() error {
			_, err := FitPowerLaw([]float64{1, 2}, PowerLawFitOptions{XMin: &xminNegative})
			return err
		},
		"NaN XMin": func() error { _, err := FitPowerLaw([]float64{1, 2}, PowerLawFitOptions{XMin: &xminNaN}); return err },
	} {
		t.Run(name, func(t *testing.T) {
			if err := call(); err == nil {
				t.Fatal("invalid input accepted")
			}
		})
	}
}

func TestPowerLawPValueValidation(t *testing.T) {
	fit, err := FitPowerLaw([]float64{1, 1, 2, 3, 5, 8, 13}, PowerLawFitOptions{})
	if err != nil {
		t.Fatal(err)
	}
	for _, precision := range []float64{0, -1, 0.51, math.NaN(), math.Inf(1)} {
		if _, err := fit.PValue(PowerLawPValueOptions{Precision: precision}); err == nil {
			t.Fatalf("precision %v accepted", precision)
		}
	}
	if _, err := (PowerLawFit{}).PValue(PowerLawPValueOptions{Precision: 0.25}); err == nil {
		t.Fatal("zero model accepted")
	}
	invalid := fit
	invalid.Alpha = 1
	if _, err := invalid.PValue(PowerLawPValueOptions{Precision: 0.25}); err == nil {
		t.Fatal("invalid model accepted")
	}
}

func TestPowerLawPValueConcurrent(t *testing.T) {
	fit, err := FitPowerLaw([]float64{1, 1, 2, 2, 3, 4, 6, 10, 20, 40}, PowerLawFitOptions{})
	if err != nil {
		t.Fatal(err)
	}
	errors := make(chan error, 4)
	for index := range 4 {
		go func() {
			seed := uint64(index + 1)
			_, err := fit.PValue(PowerLawPValueOptions{Precision: 0.5, Seed: &seed})
			errors <- err
		}()
	}
	for range 4 {
		if err := <-errors; err != nil {
			t.Fatal(err)
		}
	}
}
