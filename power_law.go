package igraph

// #cgo pkg-config: igraph
// #cgo linux LDFLAGS: -ldl
// #include <igraph.h>
// #include "power_law_cgo.h"
import "C"

import (
	"fmt"
	"math"
)

// PowerLawFitOptions controls power-law fitting. A nil XMin asks igraph to
// select the lower cutoff automatically. A non-nil XMin is borrowed and must
// be finite and non-negative; zero includes every sample. ForceContinuous
// selects a continuous model even when all samples are integral. Otherwise
// igraph chooses a discrete model for integral data and a continuous model for
// data containing a non-integral sample.
type PowerLawFitOptions struct {
	XMin            *float64
	ForceContinuous bool
}

// PowerLawPValueOptions controls Monte Carlo goodness-of-fit estimation.
// Precision must be finite and in (0, 0.5]; approximately
// 0.25/Precision^2 resamples are performed. Seed optionally resets the package
// C/igraph RNG, making equal non-nil seeds replay the calculation.
type PowerLawPValueOptions struct {
	Precision float64
	Seed      *uint64
}

// PowerLawFit is a Go-owned fitted power-law model. Alpha is the fitted
// exponent, XMin is the fitted lower cutoff, LogLikelihood is the fitted
// log-likelihood, and KSStatistic is the Kolmogorov-Smirnov distance (smaller
// is a closer fit). Continuous reports the model family selected by igraph.
//
// The value contains a private Go-owned copy of the original samples so that
// PValue can reconstruct the upstream model without retaining C memory or the
// caller's input slice.
type PowerLawFit struct {
	Continuous    bool
	Alpha         float64
	XMin          float64
	LogLikelihood float64
	KSStatistic   float64

	data []float64
}

// FitPowerLaw fits a power-law distribution to borrowed sample values. The
// input is copied and never retained or modified. At least two finite samples
// are required. The returned model and its retained sample copy are Go-owned.
//
//igraph:bind igraph_power_law_fit
func FitPowerLaw(data []float64, options PowerLawFitOptions) (PowerLawFit, error) {
	return fitPowerLaw(data, options, nil)
}

// PValue estimates the probability that data drawn from the fitted model
// would have a larger Kolmogorov-Smirnov statistic. The model is borrowed and
// reconstructed in temporary C storage for the synchronous call. RNG access is
// serialized package-wide; the model remains entirely Go-owned.
//
//igraph:bind igraph_plfit_result_calculate_p_value
func (fit PowerLawFit) PValue(options PowerLawPValueOptions) (float64, error) {
	return fit.powerLawPValue(options, nil)
}

type powerLawRawResult struct {
	continuous    bool
	alpha         float64
	xmin          float64
	logLikelihood float64
	ksStatistic   float64
}

type powerLawAdapters struct {
	newReal func([]float64) (*realVector, error)
	fit     func(*realVector, float64, bool) (powerLawRawResult, int)
	pValue  func(*realVector, powerLawRawResult, float64, *uint64) (float64, int)
}

func defaultPowerLawAdapters() powerLawAdapters {
	return powerLawAdapters{
		newReal: newRealVector,
		fit: func(data *realVector, xmin float64, forceContinuous bool) (powerLawRawResult, int) {
			var continuous C.igraph_bool_t
			var alpha, fittedXMin, likelihood, statistic C.igraph_real_t
			code := C.go_igraph_power_law_fit(
				&data.value, C.igraph_real_t(xmin), C.igraph_bool_t(booltoint(forceContinuous)),
				&continuous, &alpha, &fittedXMin, &likelihood, &statistic,
			)
			return powerLawRawResult{
				continuous:    bool(continuous),
				alpha:         float64(alpha),
				xmin:          float64(fittedXMin),
				logLikelihood: float64(likelihood),
				ksStatistic:   float64(statistic),
			}, int(code)
		},
		pValue: func(data *realVector, model powerLawRawResult, precision float64, seed *uint64) (float64, int) {
			var result C.igraph_real_t
			var cSeed C.uint64_t
			if seed != nil {
				cSeed = C.uint64_t(*seed)
			}
			code := C.go_igraph_power_law_p_value(
				&data.value, C.igraph_bool_t(booltoint(model.continuous)),
				C.igraph_real_t(model.alpha), C.igraph_real_t(model.xmin),
				C.igraph_real_t(model.logLikelihood), C.igraph_real_t(model.ksStatistic),
				C.igraph_real_t(precision), C.igraph_bool_t(booltoint(seed != nil)), cSeed, &result,
			)
			return float64(result), int(code)
		},
	}
}

func fitPowerLaw(data []float64, options PowerLawFitOptions, adapters *powerLawAdapters) (PowerLawFit, error) {
	if len(data) < 2 {
		return PowerLawFit{}, fmt.Errorf("igraph: power-law fitting requires at least two samples: %d", len(data))
	}
	if _, err := intToIgraphInt(len(data), "power-law sample count"); err != nil {
		return PowerLawFit{}, err
	}
	for index, value := range data {
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return PowerLawFit{}, fmt.Errorf("igraph: power-law sample at index %d must be finite: %v", index, value)
		}
	}
	xmin := -1.0
	if options.XMin != nil {
		xmin = *options.XMin
		if math.IsNaN(xmin) || math.IsInf(xmin, 0) || xmin < 0 {
			return PowerLawFit{}, fmt.Errorf("igraph: power-law XMin must be finite and non-negative: %v", xmin)
		}
	}
	resolved := defaultPowerLawAdapters()
	if adapters != nil {
		resolved = *adapters
	}
	cData, err := resolved.newReal(data)
	if err != nil {
		return PowerLawFit{}, err
	}
	defer cData.close()
	raw, code := resolved.fit(cData, xmin, options.ForceContinuous)
	if code != int(C.IGRAPH_SUCCESS) {
		return PowerLawFit{}, igraphError("fit power-law model", code)
	}
	if err := validatePowerLawRawResult(raw); err != nil {
		return PowerLawFit{}, err
	}
	return PowerLawFit{
		Continuous:    raw.continuous,
		Alpha:         raw.alpha,
		XMin:          raw.xmin,
		LogLikelihood: raw.logLikelihood,
		KSStatistic:   raw.ksStatistic,
		data:          append([]float64{}, data...),
	}, nil
}

func (fit PowerLawFit) powerLawPValue(options PowerLawPValueOptions, adapters *powerLawAdapters) (float64, error) {
	if len(fit.data) < 2 {
		return 0, fmt.Errorf("igraph: power-law model has no fitted sample data")
	}
	if math.IsNaN(options.Precision) || math.IsInf(options.Precision, 0) || options.Precision <= 0 || options.Precision > 0.5 {
		return 0, fmt.Errorf("igraph: power-law p-value precision must be finite and in (0, 0.5]: %v", options.Precision)
	}
	raw := powerLawRawResult{
		continuous:    fit.Continuous,
		alpha:         fit.Alpha,
		xmin:          fit.XMin,
		logLikelihood: fit.LogLikelihood,
		ksStatistic:   fit.KSStatistic,
	}
	if err := validatePowerLawRawResult(raw); err != nil {
		return 0, fmt.Errorf("igraph: invalid power-law model: %w", err)
	}
	resolved := defaultPowerLawAdapters()
	if adapters != nil {
		resolved = *adapters
	}
	cData, err := resolved.newReal(fit.data)
	if err != nil {
		return 0, err
	}
	defer cData.close()
	var result float64
	// Seeding and sampling must occur in one cgo call because igraph RNG state
	// is thread-local. The package lock prevents other stochastic operations
	// from interleaving that call.
	rngMutex.Lock()
	result, code := resolved.pValue(cData, raw, options.Precision, options.Seed)
	rngMutex.Unlock()
	if code != int(C.IGRAPH_SUCCESS) {
		return 0, igraphError("calculate power-law p-value", code)
	}
	if math.IsNaN(result) || math.IsInf(result, 0) || result < 0 || result > 1 {
		return 0, fmt.Errorf("igraph: power-law p-value result is outside [0, 1]: %v", result)
	}
	return result, nil
}

func validatePowerLawRawResult(result powerLawRawResult) error {
	if math.IsNaN(result.alpha) || math.IsInf(result.alpha, 0) || result.alpha <= 1 {
		return fmt.Errorf("power-law exponent must be finite and greater than one: %v", result.alpha)
	}
	if math.IsNaN(result.xmin) || math.IsInf(result.xmin, 0) || result.xmin < 0 {
		return fmt.Errorf("power-law XMin must be finite and non-negative: %v", result.xmin)
	}
	if math.IsNaN(result.logLikelihood) || math.IsInf(result.logLikelihood, 0) {
		return fmt.Errorf("power-law log-likelihood must be finite: %v", result.logLikelihood)
	}
	if math.IsNaN(result.ksStatistic) || math.IsInf(result.ksStatistic, 0) || result.ksStatistic < 0 {
		return fmt.Errorf("power-law KS statistic must be finite and non-negative: %v", result.ksStatistic)
	}
	return nil
}
