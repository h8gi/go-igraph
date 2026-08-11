package igraph_test

import (
	"fmt"
	"log"

	igraph "github.com/h8gi/go-igraph"
)

func ExampleGraph_Isomorphic() {
	left, err := igraph.NewGraphFromEdges(3, []igraph.Edge{{From: 0, To: 1}, {From: 1, To: 2}}, false)
	if err != nil {
		log.Fatal(err)
	}
	defer left.Close()
	right, err := igraph.NewGraphFromEdges(3, []igraph.Edge{{From: 2, To: 1}, {From: 1, To: 0}}, false)
	if err != nil {
		log.Fatal(err)
	}
	defer right.Close()

	match, err := left.Isomorphic(right)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("isomorphic:", match)
	// Output:
	// isomorphic: true
}

func ExampleGraph_ContainsSubgraphIsomorphicToVF2() {
	target, err := igraph.NewGraphFromEdges(4, []igraph.Edge{{From: 0, To: 1}, {From: 1, To: 2}, {From: 2, To: 3}}, false)
	if err != nil {
		log.Fatal(err)
	}
	defer target.Close()
	pattern, err := igraph.NewGraphFromEdges(2, []igraph.Edge{{From: 0, To: 1}}, false)
	if err != nil {
		log.Fatal(err)
	}
	defer pattern.Close()

	match, err := target.ContainsSubgraphIsomorphicToVF2(pattern, igraph.VF2SubgraphOptions{
		TargetVertexColors:  []int{1, 2, 3, 4},
		PatternVertexColors: []int{2, 3},
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("found:", match.Found)
	fmt.Println("pattern mapping length:", len(match.PatternToTarget))
	fmt.Println("unmatched target sentinel:", match.TargetToPattern[0] == igraph.RemovedID)
	// Output:
	// found: true
	// pattern mapping length: 2
	// unmatched target sentinel: true
}

func ExampleGraph_EnumerateIsomorphismsVF2() {
	left, err := igraph.NewRing(4, false, false)
	if err != nil {
		log.Fatal(err)
	}
	defer left.Close()
	right, err := igraph.NewRing(4, false, false)
	if err != nil {
		log.Fatal(err)
	}
	defer right.Close()

	result, err := left.EnumerateIsomorphismsVF2(right, igraph.VF2IsomorphismEnumerationOptions{MaxMappings: 2})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("collected:", len(result.Mappings))
	fmt.Println("more mappings exist:", result.Truncated)
	// Output:
	// collected: 2
	// more mappings exist: true
}

func ExampleGraph_CanonicalGraph() {
	graph, err := igraph.NewRing(4, false, false)
	if err != nil {
		log.Fatal(err)
	}
	defer graph.Close()
	canonical, err := graph.CanonicalGraph(nil)
	if err != nil {
		log.Fatal(err)
	}
	defer canonical.Graph.Close()
	size, err := graph.AutomorphismGroupSize(nil)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("permutation length:", len(canonical.SourceToCanonical))
	fmt.Println("automorphisms:", size)
	// Output:
	// permutation length: 4
	// automorphisms: 8
}
