package igraph_test

import (
	"fmt"

	igraph "github.com/h8gi/go-igraph"
)

func Example_hierarchicalRandomGraphWorkflow() {
	graph, err := igraph.NewGraphFromEdges(4, []igraph.Edge{{From: 0, To: 1}, {From: 1, To: 2}, {From: 2, To: 3}, {From: 3, To: 0}}, false)
	if err != nil {
		panic(err)
	}
	defer graph.Close()
	seed := uint64(2026)
	model, err := graph.FitHRG(igraph.HRGFitOptions{Steps: 50, Seed: &seed})
	if err != nil {
		panic(err)
	}
	consensus, err := graph.ConsensusHRG(igraph.HRGAnalysisOptions{Samples: 25, Seed: &seed, StartingModel: &model})
	if err != nil {
		panic(err)
	}
	prediction, err := graph.PredictHRG(igraph.HRGPredictionOptions{Samples: 25, Bins: 5, Seed: &seed, StartingModel: &model})
	if err != nil {
		panic(err)
	}
	samples, err := model.Sample(2, &seed)
	if err != nil {
		panic(err)
	}
	defer samples[0].Close()
	defer samples[1].Close()
	vertices, _ := samples[0].VertexCount()
	fmt.Println("model leaves:", model.LeafCount())
	fmt.Println("consensus aligned:", len(consensus.Parents) == model.LeafCount()+len(consensus.Weights))
	fmt.Println("prediction aligned:", len(prediction.Edges) == len(prediction.Probabilities))
	fmt.Printf("samples: %d x %d vertices\n", len(samples), vertices)
	// Output:
	// model leaves: 4
	// consensus aligned: true
	// prediction aligned: true
	// samples: 2 x 4 vertices
}
