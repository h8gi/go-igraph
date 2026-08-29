package igraph_test

import (
	"errors"
	"fmt"
	"reflect"
	"sync"
	"testing"

	igraph "github.com/h8gi/go-igraph"
)

func assertSIRTrajectories(t *testing.T, trajectories []igraph.SIRTrajectory, runs, population int) {
	t.Helper()
	if trajectories == nil || len(trajectories) != runs {
		t.Fatalf("trajectory count = %#v, want %d", trajectories, runs)
	}
	for run, trajectory := range trajectories {
		n := len(trajectory.Times)
		if n == 0 || len(trajectory.Susceptible) != n || len(trajectory.Infected) != n || len(trajectory.Recovered) != n {
			t.Fatalf("run %d is not event-aligned: %#v", run, trajectory)
		}
		for i := range n {
			if i > 0 && trajectory.Times[i] < trajectory.Times[i-1] {
				t.Errorf("run %d times decrease", run)
			}
			s, infected, recovered := trajectory.Susceptible[i], trajectory.Infected[i], trajectory.Recovered[i]
			if s < 0 || infected < 0 || recovered < 0 || s+infected+recovered != population {
				t.Errorf("run %d event %d has invalid compartments", run, i)
			}
		}
		if trajectory.Infected[n-1] != 0 {
			t.Errorf("run %d terminal infected = %d", run, trajectory.Infected[n-1])
		}
	}
}

func TestSIRReproducibilityOwnershipAndSpecialGraphs(t *testing.T) {
	seed := uint64(25)
	tests := []struct {
		name     string
		vertices int
		edges    []igraph.Edge
		directed bool
	}{
		{"singleton", 1, nil, false},
		{"edgeless", 4, nil, false},
		{"disconnected", 5, []igraph.Edge{{From: 0, To: 1}, {From: 1, To: 2}}, false},
		{"directed", 4, []igraph.Edge{{From: 0, To: 1}, {From: 1, To: 2}, {From: 2, To: 3}}, true},
		{"connected", 5, []igraph.Edge{{From: 0, To: 1}, {From: 1, To: 2}, {From: 2, To: 3}, {From: 3, To: 4}, {From: 4, To: 0}}, false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g, err := igraph.NewGraphFromEdges(test.vertices, test.edges, test.directed)
			if err != nil {
				t.Fatal(err)
			}
			first, err := g.SIR(0.8, 1.2, 3, igraph.SIROptions{Seed: &seed})
			if err != nil {
				t.Fatal(err)
			}
			second, err := g.SIR(0.8, 1.2, 3, igraph.SIROptions{Seed: &seed})
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(first, second) {
				t.Fatal("equal seeds did not replay")
			}
			assertSIRTrajectories(t, first, 3, test.vertices)
			_ = g.Close()
			assertSIRTrajectories(t, first, 3, test.vertices)
			if _, err := g.SIR(1, 1, 1, igraph.SIROptions{}); !errors.Is(err, igraph.ErrClosed) {
				t.Fatalf("post-Close error = %v", err)
			}
		})
	}
	for _, edges := range [][]igraph.Edge{{{From: 0, To: 0}}, {{From: 0, To: 1}, {From: 0, To: 1}}} {
		g, err := igraph.NewGraphFromEdges(2, edges, false)
		if err != nil {
			t.Fatal(err)
		}
		_, err = g.SIR(1, 1, 1, igraph.SIROptions{})
		_ = g.Close()
		if err == nil {
			t.Errorf("SIR accepted loop/parallel-edge graph: %v", edges)
		}
	}
}

func TestSIRZeroRateEmptyAndValidation(t *testing.T) {
	g, err := igraph.NewGraphFromEdges(3, []igraph.Edge{{From: 0, To: 1}, {From: 1, To: 2}}, false)
	if err != nil {
		t.Fatal(err)
	}
	defer g.Close()
	trajectories, err := g.SIR(0, 1, 2, igraph.SIROptions{})
	if err != nil {
		t.Fatal(err)
	}
	assertSIRTrajectories(t, trajectories, 2, 3)
	for _, trajectory := range trajectories {
		if trajectory.Recovered[len(trajectory.Recovered)-1] != 1 {
			t.Fatalf("zero-rate recovered = %v", trajectory.Recovered)
		}
	}
	for _, input := range []struct {
		beta, gamma float64
		runs        int
	}{{-1, 1, 1}, {1, 0, 1}, {1, -1, 1}, {1, 1, 0}, {1, 1, -1}} {
		if _, err := g.SIR(input.beta, input.gamma, input.runs, igraph.SIROptions{}); err == nil {
			t.Errorf("accepted invalid input %#v", input)
		}
	}
	empty, err := igraph.NewGraph()
	if err != nil {
		t.Fatal(err)
	}
	defer empty.Close()
	if _, err := empty.SIR(1, 1, 1, igraph.SIROptions{}); err == nil {
		t.Fatal("empty graph accepted")
	}
}

func TestSIRConcurrentCalls(t *testing.T) {
	g, err := igraph.NewGraphFromEdges(4, []igraph.Edge{{From: 0, To: 1}, {From: 1, To: 2}, {From: 2, To: 3}}, false)
	if err != nil {
		t.Fatal(err)
	}
	defer g.Close()
	var wg sync.WaitGroup
	for i := range 8 {
		wg.Add(1)
		go func(seed uint64) {
			defer wg.Done()
			got, err := g.SIR(1, 1, 2, igraph.SIROptions{Seed: &seed})
			if err != nil {
				t.Error(err)
				return
			}
			assertSIRTrajectories(t, got, 2, 4)
		}(uint64(i))
	}
	wg.Wait()
}

func ExampleGraph_SIR() {
	g, _ := igraph.NewGraphFromEdges(4, []igraph.Edge{{From: 0, To: 1}, {From: 1, To: 2}, {From: 2, To: 3}}, false)
	defer g.Close()
	seed := uint64(25)
	runs, _ := g.SIR(0.7, 1.0, 2, igraph.SIROptions{Seed: &seed})
	valid := true
	for _, run := range runs {
		valid = valid && run.Infected[len(run.Infected)-1] == 0
	}
	fmt.Printf("runs=%d terminally-recovered=%v\n", len(runs), valid)
	// Output: runs=2 terminally-recovered=true
}
