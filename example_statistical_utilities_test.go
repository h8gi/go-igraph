package igraph_test

import (
	"fmt"

	igraph "github.com/h8gi/go-igraph"
)

func ExampleRunningMean() {
	means, _ := igraph.RunningMean([]float64{1, 3, 5, 7}, 2)
	seed := uint64(29)
	sample, _ := igraph.SampleIntegers(0, 20, 5, igraph.IntegerSampleOptions{Seed: &seed})
	fit, _ := igraph.FitPowerLaw(
		[]float64{1, 1, 2, 2, 3, 4, 6, 10, 20, 40},
		igraph.PowerLawFitOptions{},
	)
	pValue, _ := fit.PValue(igraph.PowerLawPValueOptions{Precision: 0.5, Seed: &seed})

	fmt.Println(means)
	fmt.Println(len(sample), sample[0] >= 0, sample[len(sample)-1] <= 20)
	fmt.Println(fit.Alpha > 1, pValue >= 0 && pValue <= 1)
	// Output:
	// [2 4 6]
	// 5 true true
	// true true
}
