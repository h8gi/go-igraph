package igraph_test

import (
	"fmt"
	"math"
	"reflect"
	"sync"
	"testing"

	igraph "github.com/h8gi/go-igraph"
)

func TestMilestone22ReferenceWorkflows(t *testing.T) {
	seed := uint64(342)

	t.Run("expected degree to summaries", func(t *testing.T) {
		graph, err := igraph.ChungLuGame([]float64{1, 2, 2, 1}, nil, igraph.ChungLuOptions{Seed: &seed, Variant: igraph.ChungLuMaximumEntropy})
		if err != nil {
			t.Fatal(err)
		}
		defer graph.Close()
		degrees, err := graph.Degree(igraph.AllVertices(), igraph.DegreeOptions{})
		if err != nil || len(degrees) != 4 {
			t.Fatalf("degrees=%v, %v", degrees, err)
		}
		mean, err := graph.MeanDegree(false)
		if err != nil || math.IsNaN(mean) || mean < 0 {
			t.Fatalf("mean degree=%g, %v", mean, err)
		}
	})

	t.Run("preference to categorical mixing", func(t *testing.T) {
		preference := mustMilestone22Matrix(t, [][]float64{{0.8, 0.1}, {0.1, 0.7}})
		result, err := igraph.PreferenceGame(8, preference, igraph.PreferenceOptions{Seed: &seed, TypeCounts: []int{4, 4}})
		if err != nil {
			t.Fatal(err)
		}
		defer result.Graph.Close()
		mixing, err := result.Graph.CategoricalJointDistribution(igraph.IntegerCategories(result.Types), igraph.CategoryJointDistributionOptions{})
		if err != nil {
			t.Fatal(err)
		}
		if rows, columns := mixing.Matrix.Dims(); rows != 2 || columns != 2 {
			t.Fatalf("mixing dimensions=%dx%d", rows, columns)
		}
	})

	t.Run("latent samples to dot product graph", func(t *testing.T) {
		positions, err := igraph.SampleDirichlet(6, []float64{1, 1, 1}, igraph.LatentSampleOptions{Seed: &seed})
		if err != nil {
			t.Fatal(err)
		}
		graph, err := igraph.DotProductGame(positions, igraph.LatentGraphOptions{Seed: &seed})
		if err != nil {
			t.Fatal(err)
		}
		defer graph.Close()
		if vertices, _ := graph.VertexCount(); vertices != 6 {
			t.Fatalf("vertices=%d", vertices)
		}
	})

	t.Run("geometric coordinates to spatial analysis", func(t *testing.T) {
		result, err := igraph.GeometricRandomGame(10, 0.45, igraph.GeometricGraphOptions{Seed: &seed})
		if err != nil {
			t.Fatal(err)
		}
		defer result.Graph.Close()
		lengths, err := result.Graph.SpatialEdgeLengths(result.Coordinates, igraph.SpatialEuclidean)
		if err != nil {
			t.Fatal(err)
		}
		for _, length := range lengths {
			if length >= 0.45 {
				t.Fatalf("geometric edge length=%g", length)
			}
		}
	})

	t.Run("correlated pair to graph comparison", func(t *testing.T) {
		pair, err := igraph.CorrelatedPairGame(8, 1, 0.35, false, igraph.CorrelatedGraphOptions{Seed: &seed})
		if err != nil {
			t.Fatal(err)
		}
		defer pair.First.Close()
		defer pair.Second.Close()
		same, err := pair.First.Isomorphic(pair.Second)
		if err != nil || !same {
			t.Fatalf("isomorphic=%v, %v", same, err)
		}
	})
}

func TestMilestone22PackageSeedAndConcurrencyContract(t *testing.T) {
	seed := uint64(342)
	first, err := igraph.StaticPowerLawGame(20, 30, 2.5, igraph.StaticPowerLawOptions{Seed: &seed})
	if err != nil {
		t.Fatal(err)
	}
	defer first.Close()
	second, err := igraph.StaticPowerLawGame(20, 30, 2.5, igraph.StaticPowerLawOptions{Seed: &seed})
	if err != nil {
		t.Fatal(err)
	}
	defer second.Close()
	if !reflect.DeepEqual(mustEdges(t, first), mustEdges(t, second)) {
		t.Fatal("same package seed differed")
	}

	source, err := igraph.ErdosRenyiGNM(12, 20, false, false, igraph.ErdosRenyiOptions{Seed: &seed})
	if err != nil {
		t.Fatal(err)
	}
	defer source.Close()
	var wait sync.WaitGroup
	errors := make(chan error, 24)
	for index := range 8 {
		wait.Add(1)
		go func(seed uint64) {
			defer wait.Done()
			generated, err := igraph.SampleSphereVolume(4, 3, 1, igraph.LatentSampleOptions{Seed: &seed})
			if err != nil {
				errors <- err
			} else if rows, columns := generated.Dims(); rows != 4 || columns != 3 {
				errors <- fmt.Errorf("latent sample dimensions=%dx%d", rows, columns)
			}
			correlated, err := source.CorrelatedGame(0.5, 0.3, igraph.CorrelatedGraphOptions{Seed: &seed})
			if err != nil {
				errors <- err
			} else {
				_ = correlated.Close()
			}
		}(uint64(index + 1))
	}
	wait.Wait()
	close(errors)
	for err := range errors {
		t.Error(err)
	}
}

func mustMilestone22Matrix(t *testing.T, rows [][]float64) igraph.Matrix {
	t.Helper()
	matrix, err := igraph.NewMatrixFromRows(rows)
	if err != nil {
		t.Fatal(err)
	}
	return matrix
}
