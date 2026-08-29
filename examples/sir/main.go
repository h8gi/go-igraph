package main

import (
	"fmt"
	"log"

	igraph "github.com/h8gi/go-igraph"
)

func main() {
	graph, err := igraph.NewGraphFromEdges(6, []igraph.Edge{{From: 0, To: 1}, {From: 1, To: 2}, {From: 2, To: 3}, {From: 3, To: 4}, {From: 4, To: 5}, {From: 5, To: 0}}, false)
	if err != nil {
		log.Fatal(err)
	}
	defer graph.Close()
	seed := uint64(25)
	runs, err := graph.SIR(0.8, 1.0, 5, igraph.SIROptions{Seed: &seed})
	if err != nil {
		log.Fatal(err)
	}
	for i, run := range runs {
		last := len(run.Times) - 1
		fmt.Printf("run %d: events=%d duration=%.3f recovered=%d\n", i+1, len(run.Times), run.Times[last], run.Recovered[last])
	}
}
