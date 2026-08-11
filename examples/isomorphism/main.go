// Command isomorphism demonstrates the Milestone 10 graph-isomorphism,
// bounded matching, canonical-labeling, and automorphism APIs.
package main

import (
	"fmt"
	"log"

	igraph "github.com/h8gi/go-igraph"
)

func main() {
	target, err := igraph.NewRing(4, false, false)
	if err != nil {
		log.Fatal(err)
	}
	defer target.Close()
	pattern, err := igraph.NewGraphFromEdges(2, []igraph.Edge{{From: 0, To: 1}}, false)
	if err != nil {
		log.Fatal(err)
	}
	defer pattern.Close()

	canonical, err := target.CanonicalGraph(nil)
	if err != nil {
		log.Fatal(err)
	}
	defer canonical.Graph.Close()
	fmt.Printf("Canonicalized %d vertices with source-to-canonical mapping %v.\n", len(canonical.SourceToCanonical), canonical.SourceToCanonical)

	match, err := target.ContainsSubgraphIsomorphicToVF2(pattern, igraph.VF2SubgraphOptions{
		TargetVertexColors:  []int{1, 1, 1, 1},
		PatternVertexColors: []int{1, 1},
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Color-aware VF2 found a pattern mapping: %t.\n", match.Found)

	mappings, err := target.EnumerateSubgraphIsomorphismsLAD(pattern, igraph.LADEnumerationOptions{MaxMappings: 3})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Bounded LAD collected %d mappings; truncated=%t.\n", len(mappings.Mappings), mappings.Truncated)

	size, err := target.AutomorphismGroupSize(nil)
	if err != nil {
		log.Fatal(err)
	}
	generators, err := target.AutomorphismGenerators(nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("The cycle has %s automorphisms represented by %d generators.\n", size.String(), len(generators))
}
