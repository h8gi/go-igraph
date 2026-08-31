package main

import (
	"fmt"
	"log"

	igraph "github.com/h8gi/go-igraph"
)

func main() {
	graph, err := igraph.NewGraphFromEdges(6, []igraph.Edge{{From: 0, To: 1}, {From: 1, To: 2}, {From: 2, To: 3}, {From: 3, To: 4}, {From: 4, To: 5}, {From: 5, To: 0}, {From: 0, To: 3}}, false)
	if err != nil {
		log.Fatal(err)
	}
	defer graph.Close()
	seed := uint64(2026)
	model, err := graph.FitHRG(igraph.HRGFitOptions{Steps: 100, Seed: &seed})
	if err != nil {
		log.Fatal(err)
	}
	consensus, err := graph.ConsensusHRG(igraph.HRGAnalysisOptions{Samples: 50, Seed: &seed, StartingModel: &model})
	if err != nil {
		log.Fatal(err)
	}
	prediction, err := graph.PredictHRG(igraph.HRGPredictionOptions{Samples: 50, Bins: 5, Seed: &seed, StartingModel: &model})
	if err != nil {
		log.Fatal(err)
	}
	samples, err := model.Sample(3, &seed)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		for _, sample := range samples {
			_ = sample.Close()
		}
	}()
	fmt.Printf("fitted %d-leaf HRG\n", model.LeafCount())
	fmt.Printf("consensus: %d vertices, %d group weights\n", len(consensus.Parents), len(consensus.Weights))
	fmt.Printf("predicted %d missing edges with aligned probabilities: %t\n", len(prediction.Edges), len(prediction.Edges) == len(prediction.Probabilities))
	fmt.Printf("sampled %d independently owned graphs\n", len(samples))
}
