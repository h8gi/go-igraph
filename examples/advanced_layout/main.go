// Command advanced_layout composes the Milestone 28 advanced layout APIs.
package main

import (
	"fmt"
	"log"

	igraph "github.com/h8gi/go-igraph"
)

func main() {
	seed := uint64(28)
	graph, err := igraph.NewRing(8, false, false)
	if err != nil {
		log.Fatal(err)
	}
	defer graph.Close()

	circle, err := graph.LayoutCircle(nil)
	if err != nil {
		log.Fatal(err)
	}
	gem, err := graph.LayoutGEM(igraph.GEMOptions{
		Seed: &seed, MaxIter: 100, InitialCoordinates: &circle,
	})
	if err != nil {
		log.Fatal(err)
	}
	aligned, err := graph.AlignLayout(gem)
	if err != nil {
		log.Fatal(err)
	}
	rows, columns := aligned.Dims()
	fmt.Printf("aligned GEM: %dx%d\n", rows, columns)

	drl, err := graph.LayoutDrL(igraph.DrLOptions{Seed: &seed})
	if err != nil {
		log.Fatal(err)
	}
	drl3D, err := graph.LayoutDrL3D(igraph.DrLOptions{Seed: &seed})
	if err != nil {
		log.Fatal(err)
	}
	_, drlColumns := drl.Dims()
	_, drl3DColumns := drl3D.Dims()
	fmt.Printf("DrL dimensions: %dD and %dD\n", drlColumns, drl3DColumns)

	tree, err := igraph.NewGraphFromEdges(5, []igraph.Edge{
		{From: 0, To: 1}, {From: 1, To: 2}, {From: 2, To: 3}, {From: 3, To: 4},
	}, false)
	if err != nil {
		log.Fatal(err)
	}
	defer tree.Close()
	roots, err := tree.TreeLayoutRoots(igraph.DegAll, igraph.TreeRootByEccentricity)
	if err != nil {
		log.Fatal(err)
	}
	treeCoordinates, err := tree.LayoutReingoldTilford(igraph.DegAll, roots)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("automatic tree roots: %v\n", roots)

	merged, err := igraph.MergeLayoutsDLA(
		[]*igraph.Graph{graph, tree},
		[]igraph.Matrix{aligned, treeCoordinates},
		igraph.DLAMergeOptions{Seed: &seed},
	)
	if err != nil {
		log.Fatal(err)
	}
	rows, columns = merged.Dims()
	fmt.Printf("DLA merge: %dx%d in graph/vertex row order\n", rows, columns)
}
