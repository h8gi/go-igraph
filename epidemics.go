package igraph

/*
#include <igraph.h>
#include "epidemics_cgo.h"
*/
import "C"

import (
	"fmt"
	"math"
)

// SIROptions controls a batch of susceptible-infected-recovered simulations.
// Its zero value uses the package RNG's current state.
type SIROptions struct {
	// Seed optionally seeds the package-wide C/igraph random number generator.
	Seed *uint64
}

// SIRTrajectory is one Go-owned, event-aligned epidemic trajectory. Every
// slice has the same non-zero length. Times are finite and nondecreasing;
// compartment values are non-negative and sum to the graph's vertex count.
type SIRTrajectory struct {
	Times       []float64
	Susceptible []int
	Infected    []int
	Recovered   []int
}

// SIR simulates runs independent continuous-time susceptible-infected-
// recovered processes. Each run starts with one uniformly selected infected
// vertex and ends when no infected vertices remain. infectionRate is the rate
// per susceptible-infected edge and may be zero; recoveryRate is the rate per
// infected vertex and must be positive. Both rates must be finite.
//
// Pinned C/igraph treats directed edges as undirected for this model. This is
// the stable Go policy too; the upstream warning is intentionally suppressed.
// Loops and parallel edges are rejected by pinned upstream igraph. Empty graphs
// are rejected because no initial infected vertex can be selected.
//
// The complete call and extraction hold the graph read lock and package RNG
// lock. A non-nil seed makes equal calls replay exactly. Returned trajectories
// are fully Go-owned and remain valid after the graph is closed.
//
//igraph:bind igraph_sir
func (g *Graph) SIR(infectionRate, recoveryRate float64, runs int, options SIROptions) ([]SIRTrajectory, error) {
	return g.sir(infectionRate, recoveryRate, runs, options, nil)
}

type sirResults struct{ value C.igraph_vector_ptr_t }

type sirAdapters struct {
	newResults func() (*sirResults, error)
	close      func(*sirResults)
	run        func(*Graph, float64, float64, int, *sirResults) int
	extract    func(*sirResults) ([]SIRTrajectory, error)
}

func defaultSIRAdapters() sirAdapters {
	return sirAdapters{
		newResults: newSIRResults,
		close:      (*sirResults).close,
		run: func(g *Graph, beta, gamma float64, runs int, result *sirResults) int {
			return int(C.go_igraph_sir_run(&g.graph, C.igraph_real_t(beta), C.igraph_real_t(gamma), C.igraph_int_t(runs), &result.value))
		},
		extract: (*sirResults).trajectories,
	}
}

//igraph:internal igraph_sir_init
func newSIRResults() (*sirResults, error) {
	result := &sirResults{}
	if code := C.go_igraph_sir_results_init(&result.value); code != C.IGRAPH_SUCCESS {
		return nil, igraphError("initialize SIR result list", int(code))
	}
	return result, nil
}

//igraph:internal igraph_sir_destroy
func (result *sirResults) close() { C.go_igraph_sir_results_destroy(&result.value) }

func (result *sirResults) trajectories() ([]SIRTrajectory, error) {
	size, err := igraphIntToInt(C.go_igraph_sir_results_size(&result.value), "SIR result count")
	if err != nil {
		return nil, err
	}
	trajectories := make([]SIRTrajectory, size)
	for i := range trajectories {
		sir := C.go_igraph_sir_results_get(&result.value, C.igraph_int_t(i))
		if sir == nil {
			return nil, fmt.Errorf("igraph: SIR result %d is nil", i)
		}
		times, err := (&realVector{value: sir.times}).slice()
		if err != nil {
			return nil, fmt.Errorf("igraph: convert SIR result %d times: %w", i, err)
		}
		susceptible, err := intVectorSlice(&sir.no_s)
		if err != nil {
			return nil, fmt.Errorf("igraph: convert SIR result %d susceptible values: %w", i, err)
		}
		infected, err := intVectorSlice(&sir.no_i)
		if err != nil {
			return nil, fmt.Errorf("igraph: convert SIR result %d infected values: %w", i, err)
		}
		recovered, err := intVectorSlice(&sir.no_r)
		if err != nil {
			return nil, fmt.Errorf("igraph: convert SIR result %d recovered values: %w", i, err)
		}
		trajectories[i] = SIRTrajectory{Times: times, Susceptible: susceptible, Infected: infected, Recovered: recovered}
	}
	return trajectories, nil
}

func (g *Graph) sir(beta, gamma float64, runs int, options SIROptions, adapters *sirAdapters) ([]SIRTrajectory, error) {
	if math.IsNaN(beta) || math.IsInf(beta, 0) || beta < 0 {
		return nil, fmt.Errorf("igraph: infection rate must be finite and non-negative: %v", beta)
	}
	if math.IsNaN(gamma) || math.IsInf(gamma, 0) || gamma <= 0 {
		return nil, fmt.Errorf("igraph: recovery rate must be finite and positive: %v", gamma)
	}
	if runs <= 0 {
		return nil, fmt.Errorf("igraph: SIR run count must be positive: %d", runs)
	}
	if _, err := intToIgraphInt(runs, "SIR run count"); err != nil {
		return nil, err
	}
	if g == nil {
		return nil, ErrClosed
	}
	g.mu.RLock()
	defer g.mu.RUnlock()
	if g.closed {
		return nil, ErrClosed
	}
	op := defaultSIRAdapters()
	if adapters != nil {
		op = *adapters
	}
	result, err := op.newResults()
	if err != nil {
		return nil, err
	}
	defer op.close(result)
	err = withRNG(options.Seed, func() error {
		if code := op.run(g, beta, gamma, runs, result); code != int(C.IGRAPH_SUCCESS) {
			return igraphError("simulate SIR epidemics", code)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	trajectories, err := op.extract(result)
	if err != nil {
		return nil, err
	}
	if trajectories == nil {
		trajectories = []SIRTrajectory{}
	}
	population, err := igraphIntToInt(C.igraph_vcount(&g.graph), "SIR population")
	if err != nil {
		return nil, err
	}
	if len(trajectories) != runs {
		return nil, fmt.Errorf("igraph: SIR returned %d trajectories, want %d", len(trajectories), runs)
	}
	for i, trajectory := range trajectories {
		if err := validateSIRTrajectory(trajectory, population); err != nil {
			return nil, fmt.Errorf("igraph: invalid SIR trajectory %d: %w", i, err)
		}
	}
	return trajectories, nil
}

func validateSIRTrajectory(trajectory SIRTrajectory, population int) error {
	n := len(trajectory.Times)
	if n == 0 || len(trajectory.Susceptible) != n || len(trajectory.Infected) != n || len(trajectory.Recovered) != n {
		return fmt.Errorf("event slices are not non-empty and aligned")
	}
	for i := range n {
		if math.IsNaN(trajectory.Times[i]) || math.IsInf(trajectory.Times[i], 0) || i > 0 && trajectory.Times[i] < trajectory.Times[i-1] {
			return fmt.Errorf("event %d has invalid time %v", i, trajectory.Times[i])
		}
		s, infected, recovered := trajectory.Susceptible[i], trajectory.Infected[i], trajectory.Recovered[i]
		if s < 0 || infected < 0 || recovered < 0 || s+infected+recovered != population {
			return fmt.Errorf("event %d has invalid compartments (%d, %d, %d)", i, s, infected, recovered)
		}
	}
	if trajectory.Infected[n-1] != 0 {
		return fmt.Errorf("terminal infected count is %d", trajectory.Infected[n-1])
	}
	return nil
}
