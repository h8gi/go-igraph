package igraph_test

import (
	"fmt"
	"log"

	igraph "github.com/h8gi/go-igraph"
)

func ExampleGraph_CommunityMultilevel() {
	// Create two connected triangles (2 communities)
	graph, err := igraph.NewGraphFromEdges(6, []igraph.Edge{
		{From: 0, To: 1}, {From: 0, To: 2}, {From: 1, To: 2},
		{From: 3, To: 4}, {From: 3, To: 5}, {From: 4, To: 5},
		{From: 2, To: 3},
	}, false)
	if err != nil {
		log.Fatal(err)
	}
	defer graph.Close()

	seed := uint64(42)
	partition, err := graph.CommunityMultilevel(igraph.MultilevelOptions{
		Seed: &seed,
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(partition.CommunityCount)
	// Output: 2
}

func ExampleGraph_CommunityWalktrap() {
	graph, err := igraph.NewGraphFromEdges(6, []igraph.Edge{
		{From: 0, To: 1}, {From: 0, To: 2}, {From: 1, To: 2},
		{From: 3, To: 4}, {From: 3, To: 5}, {From: 4, To: 5},
		{From: 2, To: 3},
	}, false)
	if err != nil {
		log.Fatal(err)
	}
	defer graph.Close()

	dendrogram, err := graph.CommunityWalktrap(nil, 4)
	if err != nil {
		log.Fatal(err)
	}

	partition, err := dendrogram.OptimalMembership()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(partition.CommunityCount)
	// Output: 2
}
