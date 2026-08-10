package igraph_test

import (
	"reflect"
	"sync"
	"testing"

	"github.com/h8gi/go-igraph"
)

func TestMilestone8IntegrationPipeline(t *testing.T) {
	seed := uint64(2026)

	// Step 1: Generate Barabási-Albert graph
	baGraph, err := igraph.BarabasiGame(30, 2, 1.0, 1.0, false, igraph.BarabasiOptions{Seed: &seed})
	if err != nil {
		t.Fatalf("BarabasiGame failed: %v", err)
	}
	defer baGraph.Close()

	vCount, eCount := mustCounts(t, baGraph)
	if vCount != 30 || eCount == 0 {
		t.Errorf("expected 30 vertices and non-zero edges, got %d and %d", vCount, eCount)
	}

	// Step 2: Edge rewiring in-place
	degBefore, err := baGraph.Degree(igraph.AllVertices(), igraph.DegreeOptions{Direction: igraph.DirectionAll})
	if err != nil {
		t.Fatalf("Degree failed: %v", err)
	}
	if err := baGraph.Rewire(50, igraph.RewireSimple, igraph.RewireOptions{Seed: &seed}); err != nil {
		t.Fatalf("Rewire failed: %v", err)
	}
	degAfter, err := baGraph.Degree(igraph.AllVertices(), igraph.DegreeOptions{Direction: igraph.DirectionAll})
	if err != nil {
		t.Fatalf("Degree after rewire failed: %v", err)
	}
	if !reflect.DeepEqual(degBefore, degAfter) {
		t.Errorf("expected degree sequence preservation after rewire")
	}

	// Step 3: Rewire edges to produce an independent copy
	rewiredCopy, err := baGraph.RewireEdges(0.1, false, false, igraph.RewireOptions{Seed: &seed})
	if err != nil {
		t.Fatalf("RewireEdges failed: %v", err)
	}
	defer rewiredCopy.Close()

	// Step 4: Random walk sampling
	vPath, ePath, err := rewiredCopy.RandomWalk(0, 10, igraph.DirectionOut, nil, igraph.RandomWalkOptions{Seed: &seed})
	if err != nil {
		t.Fatalf("RandomWalk failed: %v", err)
	}
	if len(vPath) == 0 {
		t.Errorf("expected non-empty vertex path")
	}
	_ = ePath

	// Step 5: Downstream analysis (ConnectedComponents and CommunityMultilevel)
	components, err := rewiredCopy.ConnectedComponents(igraph.ConnectednessWeak)
	if err != nil {
		t.Fatalf("ConnectedComponents failed: %v", err)
	}
	if len(components.Sizes) == 0 {
		t.Errorf("expected component sizes")
	}

	partition, err := rewiredCopy.CommunityMultilevel(igraph.MultilevelOptions{})
	if err != nil {
		t.Fatalf("CommunityMultilevel failed: %v", err)
	}
	if len(partition.Membership) != 30 {
		t.Errorf("expected 30 membership entries, got %d", len(partition.Membership))
	}
}

func TestMilestone8ConcurrentSeedIsolationRace(t *testing.T) {
	const goroutines = 10
	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			seed := uint64(100 + id)

			g, err := igraph.ErdosRenyiGNM(20, 40, false, false, igraph.ErdosRenyiOptions{Seed: &seed})
			if err != nil {
				t.Errorf("goroutine %d ErdosRenyiGNM failed: %v", id, err)
				return
			}
			defer g.Close()

			if err := g.Rewire(10, igraph.RewireSimple, igraph.RewireOptions{Seed: &seed}); err != nil {
				t.Errorf("goroutine %d Rewire failed: %v", id, err)
				return
			}

			vPath, _, err := g.RandomWalk(0, 5, igraph.DirectionOut, nil, igraph.RandomWalkOptions{Seed: &seed})
			if err != nil {
				t.Errorf("goroutine %d RandomWalk failed: %v", id, err)
				return
			}
			if len(vPath) == 0 {
				t.Errorf("goroutine %d empty random walk path", id)
			}
		}(i)
	}

	wg.Wait()
}
