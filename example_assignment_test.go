package igraph_test

import (
	"fmt"

	igraph "github.com/h8gi/go-igraph"
)

func ExampleSolveLinearAssignment() {
	costs, _ := igraph.NewMatrixFromRows([][]float64{
		{4, 1, 3},
		{2, 0, 5},
		{3, 2, 2},
	})
	assignment, _ := igraph.SolveLinearAssignment(costs)
	total := 0.0
	for agent, task := range assignment {
		cost, _ := costs.At(agent, task)
		total += cost
	}
	fmt.Println(assignment)
	fmt.Println(total)
	// Output:
	// [1 0 2]
	// 5
}
