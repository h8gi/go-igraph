package igraph_test

import (
	"fmt"
	"log"
	"sort"

	igraph "github.com/h8gi/go-igraph"
)

func ExampleGraph_Cliques() {
	graph, err := igraph.NewGraphFromEdges(5, []igraph.Edge{
		{From: 0, To: 1}, {From: 0, To: 2}, {From: 1, To: 2}, {From: 3, To: 4},
	}, false)
	if err != nil {
		log.Fatal(err)
	}
	defer graph.Close()

	cliqueNumber, err := graph.CliqueNumber()
	if err != nil {
		log.Fatal(err)
	}
	maximumWeight, err := graph.MaximumWeightCliques([]int{1, 1, 1, 10, 10}, 1)
	if err != nil {
		log.Fatal(err)
	}
	independent, err := graph.LargestIndependentVertexSets(6)
	if err != nil {
		log.Fatal(err)
	}
	sort.Slice(independent.Sets, func(i, j int) bool {
		if independent.Sets[i][0] != independent.Sets[j][0] {
			return independent.Sets[i][0] < independent.Sets[j][0]
		}
		return independent.Sets[i][1] < independent.Sets[j][1]
	})

	fmt.Println("clique number:", cliqueNumber)
	fmt.Println("maximum weight:", maximumWeight.Sets)
	fmt.Println("largest independent sets:", independent.Sets)
	// Output:
	// clique number: 3
	// maximum weight: [[3 4]]
	// largest independent sets: [[0 3] [0 4] [1 3] [1 4] [2 3] [2 4]]
}
