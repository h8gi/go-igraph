package igraph

// #cgo pkg-config: igraph
// #include <igraph.h>
// #include "statistical_utilities_cgo.h"
import "C"

import (
	"fmt"
	"math"
)

// IntegerSampleOptions controls sampling without replacement. Seed optionally
// resets the package C/igraph RNG; equal non-nil seeds replay the sample.
type IntegerSampleOptions struct {
	Seed *uint64
}

// RunningMean returns the mean of every consecutive window values in data.
// data is borrowed only for the call and is never modified. Window must be in
// [1, len(data)], so an empty input has no valid window. The non-nil returned
// slice has length len(data)-window+1, is Go-owned, and does not alias data.
// Every input and output value must be finite.
//
//igraph:bind igraph_running_mean
func RunningMean(data []float64, window int) ([]float64, error) {
	return runningMean(data, window, nil)
}

// SampleIntegers returns count distinct integers sampled without replacement
// from the inclusive interval [low, high]. The non-nil Go-owned result is in
// increasing order. A zero count returns an empty slice without consuming RNG
// state. Interval cardinality, integer conversion, and allocation bounds are
// validated before C execution.
//
//igraph:bind igraph_random_sample
func SampleIntegers(low, high, count int, options IntegerSampleOptions) ([]int, error) {
	return sampleIntegers(low, high, count, options, nil)
}

type statisticalUtilityAdapters struct {
	newReal     func([]float64) (*realVector, error)
	newInt      func([]int) (*intVector, error)
	runningMean func(*realVector, int, *realVector) int
	sample      func(int, int, int, *uint64, *intVector) int
	readReal    func(*realVector) ([]float64, error)
	readInt     func(*intVector) ([]int, error)
}

func defaultStatisticalUtilityAdapters() statisticalUtilityAdapters {
	return statisticalUtilityAdapters{
		newReal: newRealVector,
		newInt:  newIntVector,
		runningMean: func(data *realVector, window int, result *realVector) int {
			return int(C.go_igraph_running_mean(&data.value, C.igraph_int_t(window), &result.value))
		},
		sample: func(low, high, count int, seed *uint64, result *intVector) int {
			var cSeed C.uint64_t
			if seed != nil {
				cSeed = C.uint64_t(*seed)
			}
			return int(C.go_igraph_random_sample(
				C.igraph_int_t(low), C.igraph_int_t(high), C.igraph_int_t(count),
				C.igraph_bool_t(booltoint(seed != nil)), cSeed, &result.value,
			))
		},
		readReal: (*realVector).slice,
		readInt:  (*intVector).slice,
	}
}

func runningMean(data []float64, window int, adapters *statisticalUtilityAdapters) ([]float64, error) {
	if window < 1 || window > len(data) {
		return nil, fmt.Errorf("igraph: running-mean window must be in [1, %d]: %d", len(data), window)
	}
	if _, err := intToIgraphInt(len(data), "running-mean data length"); err != nil {
		return nil, err
	}
	if _, err := intToIgraphInt(window, "running-mean window"); err != nil {
		return nil, err
	}
	for index, value := range data {
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return nil, fmt.Errorf("igraph: running-mean value at index %d must be finite: %v", index, value)
		}
	}
	resolved := defaultStatisticalUtilityAdapters()
	if adapters != nil {
		resolved = *adapters
	}
	cData, err := resolved.newReal(data)
	if err != nil {
		return nil, err
	}
	defer cData.close()
	result, err := resolved.newReal(nil)
	if err != nil {
		return nil, err
	}
	defer result.close()
	if code := resolved.runningMean(cData, window, result); code != int(C.IGRAPH_SUCCESS) {
		return nil, igraphError("calculate running mean", code)
	}
	means, err := resolved.readReal(result)
	if err != nil {
		return nil, err
	}
	want := len(data) - window + 1
	if len(means) != want {
		return nil, fmt.Errorf("igraph: running-mean result has length %d, want %d", len(means), want)
	}
	for index, value := range means {
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return nil, fmt.Errorf("igraph: running-mean result at index %d is not finite: %v", index, value)
		}
	}
	return means, nil
}

func sampleIntegers(low, high, count int, options IntegerSampleOptions, adapters *statisticalUtilityAdapters) ([]int, error) {
	if low > high {
		return nil, fmt.Errorf("igraph: sample lower bound %d exceeds upper bound %d", low, high)
	}
	if count < 0 {
		return nil, fmt.Errorf("igraph: sample count must be non-negative: %d", count)
	}
	maxInt := int(^uint(0) >> 1)
	if low <= 0 && high > maxInt+low-1 {
		return nil, fmt.Errorf("igraph: sample interval cardinality overflows int: [%d, %d]", low, high)
	}
	population := high - low + 1
	if count > population {
		return nil, fmt.Errorf("igraph: sample count %d exceeds interval [%d, %d]", count, low, high)
	}
	if _, err := intToIgraphInt(low, "sample lower bound"); err != nil {
		return nil, err
	}
	if _, err := intToIgraphInt(high, "sample upper bound"); err != nil {
		return nil, err
	}
	if _, err := intToIgraphInt(count, "sample count"); err != nil {
		return nil, err
	}
	if count == 0 {
		return []int{}, nil
	}
	resolved := defaultStatisticalUtilityAdapters()
	if adapters != nil {
		resolved = *adapters
	}
	result, err := resolved.newInt(nil)
	if err != nil {
		return nil, err
	}
	defer result.close()
	rngMutex.Lock()
	code := resolved.sample(low, high, count, options.Seed, result)
	rngMutex.Unlock()
	if code != int(C.IGRAPH_SUCCESS) {
		return nil, igraphError("sample integers", code)
	}
	values, err := resolved.readInt(result)
	if err != nil {
		return nil, err
	}
	if err := validateIntegerSample(values, low, high, count); err != nil {
		return nil, err
	}
	return values, nil
}

func validateIntegerSample(values []int, low, high, count int) error {
	if len(values) != count {
		return fmt.Errorf("igraph: integer sample has length %d, want %d", len(values), count)
	}
	for index, value := range values {
		if value < low || value > high {
			return fmt.Errorf("igraph: integer sample value %d at index %d is outside [%d, %d]", value, index, low, high)
		}
		if index > 0 && value <= values[index-1] {
			return fmt.Errorf("igraph: integer sample is not strictly increasing at index %d: %d then %d", index, values[index-1], value)
		}
	}
	return nil
}
