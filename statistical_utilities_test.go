package igraph

import (
	"math"
	"reflect"
	"testing"
)

func TestRunningMeanKnownAnswersAndOwnership(t *testing.T) {
	data := []float64{1, 2, 3, 4, 5}
	got, err := RunningMean(data, 3)
	if err != nil {
		t.Fatal(err)
	}
	if want := []float64{2, 3, 4}; !reflect.DeepEqual(got, want) {
		t.Fatalf("means = %v, want %v", got, want)
	}
	data[0] = 100
	got[0] = 200
	again, err := RunningMean([]float64{1, 2, 3, 4, 5}, 3)
	if err != nil || !reflect.DeepEqual(again, []float64{2, 3, 4}) {
		t.Fatalf("second means = %v, %v", again, err)
	}
	identity, _ := RunningMean([]float64{-2, 4}, 1)
	if !reflect.DeepEqual(identity, []float64{-2, 4}) {
		t.Fatalf("window one = %v", identity)
	}
	full, _ := RunningMean([]float64{1, 2, 6}, 3)
	if !reflect.DeepEqual(full, []float64{3}) {
		t.Fatalf("full window = %v", full)
	}
}

func TestRunningMeanValidation(t *testing.T) {
	for name, call := range map[string]func() error{
		"empty":           func() error { _, err := RunningMean(nil, 1); return err },
		"zero window":     func() error { _, err := RunningMean([]float64{1}, 0); return err },
		"large window":    func() error { _, err := RunningMean([]float64{1}, 2); return err },
		"NaN":             func() error { _, err := RunningMean([]float64{math.NaN()}, 1); return err },
		"infinity":        func() error { _, err := RunningMean([]float64{math.Inf(1)}, 1); return err },
		"output overflow": func() error { _, err := RunningMean([]float64{math.MaxFloat64, math.MaxFloat64}, 2); return err },
	} {
		t.Run(name, func(t *testing.T) {
			if err := call(); err == nil {
				t.Fatal("invalid input accepted")
			}
		})
	}
}

func TestSampleIntegersBoundariesAndSeedReplay(t *testing.T) {
	zero, err := SampleIntegers(7, 7, 0, IntegerSampleOptions{})
	if err != nil || zero == nil || len(zero) != 0 {
		t.Fatalf("zero sample = %#v, %v", zero, err)
	}
	singleton, err := SampleIntegers(7, 7, 1, IntegerSampleOptions{})
	if err != nil || !reflect.DeepEqual(singleton, []int{7}) {
		t.Fatalf("singleton = %v, %v", singleton, err)
	}
	full, err := SampleIntegers(-2, 2, 5, IntegerSampleOptions{})
	if err != nil || !reflect.DeepEqual(full, []int{-2, -1, 0, 1, 2}) {
		t.Fatalf("full sample = %v, %v", full, err)
	}
	seed := uint64(2029)
	options := IntegerSampleOptions{Seed: &seed}
	first, err := SampleIntegers(-100, 100, 10, options)
	if err != nil {
		t.Fatal(err)
	}
	second, err := SampleIntegers(-100, 100, 10, options)
	if err != nil || !reflect.DeepEqual(first, second) {
		t.Fatalf("seed replay = %v, %v, %v", first, second, err)
	}
}

func TestSampleIntegersValidation(t *testing.T) {
	maxInt := int(^uint(0) >> 1)
	minInt := -maxInt - 1
	for name, call := range map[string]func() error{
		"reversed":             func() error { _, err := SampleIntegers(2, 1, 0, IntegerSampleOptions{}); return err },
		"negative count":       func() error { _, err := SampleIntegers(0, 1, -1, IntegerSampleOptions{}); return err },
		"too many":             func() error { _, err := SampleIntegers(0, 1, 3, IntegerSampleOptions{}); return err },
		"cardinality overflow": func() error { _, err := SampleIntegers(minInt, maxInt, 1, IntegerSampleOptions{}); return err },
	} {
		t.Run(name, func(t *testing.T) {
			if err := call(); err == nil {
				t.Fatal("invalid sample accepted")
			}
		})
	}
}

func TestStatisticalUtilitiesConcurrent(t *testing.T) {
	errors := make(chan error, 8)
	for index := range 4 {
		go func() {
			_, err := RunningMean([]float64{1, 2, 3, 4}, 2)
			errors <- err
		}()
		go func() {
			seed := uint64(index)
			_, err := SampleIntegers(0, 100, 5, IntegerSampleOptions{Seed: &seed})
			errors <- err
		}()
	}
	for range 8 {
		if err := <-errors; err != nil {
			t.Fatal(err)
		}
	}
}
