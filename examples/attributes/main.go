// Command attributes demonstrates typed attributes and an attributed GraphML round trip.
package main

import (
	"fmt"
	"log"
	"os"

	igraph "github.com/h8gi/go-igraph"
)

func main() {
	graph, err := igraph.NewGraphFromEdges(3, []igraph.Edge{{From: 0, To: 1}, {From: 1, To: 2}}, true)
	if err != nil {
		log.Fatal(err)
	}
	defer graph.Close()
	if err := graph.SetGraphStringAttribute("name", "delivery network"); err != nil {
		log.Fatal(err)
	}
	if err := graph.SetVertexStringAttributes("depot", []string{"north", "central", "south"}); err != nil {
		log.Fatal(err)
	}
	if err := graph.SetEdgeNumericAttributes("minutes", []float64{6, 9}); err != nil {
		log.Fatal(err)
	}

	file, err := os.CreateTemp("", "go-igraph-attributes-*.graphml")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	path := file.Name()
	defer os.Remove(path)
	if err := graph.WriteGraphML(file, false); err != nil {
		log.Fatal(err)
	}
	if err := graph.Close(); err != nil {
		log.Fatal(err)
	}
	if _, err := file.Seek(0, 0); err != nil {
		log.Fatal(err)
	}
	copy, err := igraph.ReadGraphML(file, 0)
	if err != nil {
		log.Fatal(err)
	}
	defer copy.Close()
	if err := file.Close(); err != nil {
		log.Fatal(err)
	}

	name, _ := copy.GraphStringAttribute("name")
	depots, _ := copy.VertexStringAttributes("depot")
	minutes, _ := copy.EdgeNumericAttributes("minutes")
	fmt.Printf("%s: %v, edge minutes %v\n", name, depots, minutes)
}
