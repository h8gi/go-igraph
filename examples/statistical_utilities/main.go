package main

import (
	"fmt"
	"log"

	igraph "github.com/h8gi/go-igraph"
)

func main() {
	data := []float64{1, 1, 2, 2, 3, 4, 6, 10, 20, 40}
	means, err := igraph.RunningMean(data, 2)
	if err != nil {
		log.Fatal(err)
	}
	seed := uint64(29)
	indices, err := igraph.SampleIntegers(0, len(means)-1, 4, igraph.IntegerSampleOptions{Seed: &seed})
	if err != nil {
		log.Fatal(err)
	}
	fit, err := igraph.FitPowerLaw(data, igraph.PowerLawFitOptions{})
	if err != nil {
		log.Fatal(err)
	}
	pValue, err := fit.PValue(igraph.PowerLawPValueOptions{Precision: 0.5, Seed: &seed})
	if err != nil {
		log.Fatal(err)
	}

	costs, err := igraph.NewMatrixFromRows([][]float64{{4, 1, 3}, {2, 0, 5}, {3, 2, 2}})
	if err != nil {
		log.Fatal(err)
	}
	assignment, err := igraph.SolveLinearAssignment(costs)
	if err != nil {
		log.Fatal(err)
	}
	total := 0.0
	for row, column := range assignment {
		cost, err := costs.At(row, column)
		if err != nil {
			log.Fatal(err)
		}
		total += cost
	}

	fmt.Printf("running means: %d; sampled indices: %v\n", len(means), indices)
	fmt.Printf("power-law alpha > 1: %t; p-value in range: %t\n", fit.Alpha > 1, pValue >= 0 && pValue <= 1)
	fmt.Printf("assignment: %v; total cost: %.0f\n", assignment, total)
}
