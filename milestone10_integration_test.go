package igraph_test

import (
	"math/big"
	"sync"
	"testing"

	igraph "github.com/h8gi/go-igraph"
)

func TestMilestone10IntegrationPipeline(t *testing.T) {
	source, err := igraph.NewRing(4, false, false)
	if err != nil {
		t.Fatal(err)
	}
	canonical, err := source.CanonicalGraph([]int{1, 1, 1, 1})
	if err != nil {
		t.Fatal(err)
	}
	defer canonical.Graph.Close()

	match, err := source.IsomorphicVF2(canonical.Graph, igraph.VF2IsomorphismOptions{
		SourceVertexColors: []int{1, 1, 1, 1},
		TargetVertexColors: []int{1, 1, 1, 1},
		SourceEdgeColors:   []int{2, 2, 2, 2},
		TargetEdgeColors:   []int{2, 2, 2, 2},
	})
	if err != nil || !match.Found {
		t.Fatalf("color-aware canonical match = %+v, %v", match, err)
	}

	pattern, err := igraph.NewGraphFromEdges(2, []igraph.Edge{{From: 0, To: 1}}, false)
	if err != nil {
		t.Fatal(err)
	}
	enumeration, err := canonical.Graph.EnumerateSubgraphIsomorphismsVF2(pattern, igraph.VF2SubgraphEnumerationOptions{
		Colors: igraph.VF2SubgraphOptions{
			TargetVertexColors:  []int{1, 1, 1, 1},
			PatternVertexColors: []int{1, 1},
		},
		MaxMappings: 3,
	})
	if err != nil || len(enumeration.Mappings) != 3 || !enumeration.Truncated {
		t.Fatalf("bounded canonical subgraph enumeration = %+v, %v", enumeration, err)
	}

	size, err := canonical.Graph.AutomorphismGroupSize([]int{1, 1, 1, 1})
	if err != nil || size.Cmp(big.NewInt(8)) != 0 {
		t.Fatalf("canonical cycle automorphism size = %v, %v", size, err)
	}
	generators, err := canonical.Graph.AutomorphismGenerators(nil)
	if err != nil || len(generators) == 0 {
		t.Fatalf("canonical cycle generators = %v, %v", generators, err)
	}

	if err := source.Close(); err != nil {
		t.Fatal(err)
	}
	if err := pattern.Close(); err != nil {
		t.Fatal(err)
	}
	if len(canonical.SourceToCanonical) != 4 || len(match.SourceToTarget) != 4 || len(enumeration.Mappings[0]) != 2 || len(generators[0]) != 4 {
		t.Fatal("Go-owned Milestone 10 result did not survive source closure")
	}
	if vertices, err := canonical.Graph.VertexCount(); err != nil || vertices != 4 {
		t.Fatalf("canonical graph after source closure = %d, %v", vertices, err)
	}
}

func TestMilestone10RepeatedAndReversedConcurrency(t *testing.T) {
	left, err := igraph.NewRing(6, false, false)
	if err != nil {
		t.Fatal(err)
	}
	defer left.Close()
	right, err := igraph.NewRing(6, false, false)
	if err != nil {
		t.Fatal(err)
	}
	defer right.Close()

	var wg sync.WaitGroup
	errorsCh := make(chan error, 80)
	for round := 0; round < 20; round++ {
		wg.Add(4)
		go func() { defer wg.Done(); _, err := left.Isomorphic(left); errorsCh <- err }()
		go func() {
			defer wg.Done()
			_, err := right.IsomorphicVF2(right, igraph.VF2IsomorphismOptions{})
			errorsCh <- err
		}()
		go func() { defer wg.Done(); _, err := left.Isomorphic(right); errorsCh <- err }()
		go func() {
			defer wg.Done()
			_, err := right.EnumerateSubgraphIsomorphismsLAD(left, igraph.LADEnumerationOptions{MaxMappings: 1})
			errorsCh <- err
		}()
	}
	wg.Wait()
	close(errorsCh)
	for err := range errorsCh {
		if err != nil {
			t.Fatalf("Milestone 10 concurrent call: %v", err)
		}
	}
}
