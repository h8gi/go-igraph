package igraph_test

import (
	"fmt"
	"math"
	"reflect"
	"testing"

	"github.com/h8gi/go-igraph"
)

func matrix(t *testing.T, rows [][]float64) igraph.Matrix {
	t.Helper()
	m, err := igraph.NewMatrixFromRows(rows)
	if err != nil {
		t.Fatal(err)
	}
	return m
}

func TestHierarchicalSBMGames(t *testing.T) {
	seed := uint64(336)
	pref := matrix(t, [][]float64{{0.8, 0.1}, {0.1, 0.7}})
	first, err := igraph.HierarchicalSBMGame(12, 6, []float64{0.5, 0.5}, pref, 0.02, igraph.HierarchicalBlockOptions{Seed: &seed})
	if err != nil {
		t.Fatal(err)
	}
	defer first.Graph.Close()
	second, err := igraph.HierarchicalSBMGame(12, 6, []float64{0.5, 0.5}, pref, 0.02, igraph.HierarchicalBlockOptions{Seed: &seed})
	if err != nil {
		t.Fatal(err)
	}
	defer second.Graph.Close()
	if !reflect.DeepEqual(mustEdges(t, first.Graph), mustEdges(t, second.Graph)) {
		t.Fatal("same seed differed")
	}
	if !reflect.DeepEqual(first.Blocks, []int{0, 0, 0, 0, 0, 0, 1, 1, 1, 1, 1, 1}) {
		t.Fatalf("blocks=%v", first.Blocks)
	}
	if !reflect.DeepEqual(first.Clusters, []int{0, 0, 0, 1, 1, 1, 2, 2, 2, 3, 3, 3}) {
		t.Fatalf("clusters=%v", first.Clusters)
	}
	if directed, _ := first.Graph.IsDirected(); directed {
		t.Fatal("HSBM graph is directed")
	}

	listed, err := igraph.HierarchicalSBMListGame([]igraph.HierarchicalBlockSpec{
		{Size: 4, Proportions: []float64{1}, Preference: matrix(t, [][]float64{{0.4}})},
		{Size: 6, Proportions: []float64{0, 0.5, 0.5}, Preference: matrix(t, [][]float64{{0, 0, 0}, {0, 0.8, 0.2}, {0, 0.2, 0.8}})},
	}, 0.1, igraph.HierarchicalBlockOptions{Seed: &seed})
	if err != nil {
		t.Fatal(err)
	}
	defer listed.Graph.Close()
	if len(listed.Blocks) != 10 || len(listed.Clusters) != 10 {
		t.Fatalf("membership lengths=%d/%d", len(listed.Blocks), len(listed.Clusters))
	}
	if got, _ := listed.Graph.VertexCount(); got != 10 {
		t.Fatalf("vertices=%d", got)
	}

	bad := []func() error{
		func() error {
			_, e := igraph.HierarchicalSBMGame(5, 2, []float64{1}, matrix(t, [][]float64{{1}}), 0, igraph.HierarchicalBlockOptions{})
			return e
		},
		func() error {
			_, e := igraph.HierarchicalSBMGame(4, 2, []float64{0.3, 0.7}, pref, 0, igraph.HierarchicalBlockOptions{})
			return e
		},
		func() error {
			_, e := igraph.HierarchicalSBMListGame(nil, 0, igraph.HierarchicalBlockOptions{})
			return e
		},
		func() error {
			_, e := igraph.HierarchicalSBMGame(4, 2, []float64{1}, matrix(t, [][]float64{{1}}), -1, igraph.HierarchicalBlockOptions{})
			return e
		},
		func() error {
			_, e := igraph.HierarchicalSBMGame(4, 2, nil, matrix(t, [][]float64{}), 0, igraph.HierarchicalBlockOptions{})
			return e
		},
		func() error {
			_, e := igraph.HierarchicalSBMGame(4, 2, []float64{0.5, 0.5}, matrix(t, [][]float64{{1, 0}, {1, 1}}), 0, igraph.HierarchicalBlockOptions{})
			return e
		},
		func() error {
			_, e := igraph.HierarchicalSBMListGame([]igraph.HierarchicalBlockSpec{{Size: 0, Proportions: []float64{1}, Preference: matrix(t, [][]float64{{1}})}}, 0, igraph.HierarchicalBlockOptions{})
			return e
		},
		func() error {
			_, e := igraph.HierarchicalSBMListGame([]igraph.HierarchicalBlockSpec{{Size: 1, Proportions: []float64{1}, Preference: matrix(t, [][]float64{{1}})}}, 2, igraph.HierarchicalBlockOptions{})
			return e
		},
	}
	for i, fn := range bad {
		if fn() == nil {
			t.Errorf("validation %d accepted", i)
		}
	}
}

func TestPreferenceGames(t *testing.T) {
	seed := uint64(336)
	pref := matrix(t, [][]float64{{0.9, 0.1}, {0.1, 0.8}})
	r, err := igraph.PreferenceGame(8, pref, igraph.PreferenceOptions{Seed: &seed, TypeCounts: []int{3, 5}})
	if err != nil {
		t.Fatal(err)
	}
	defer r.Graph.Close()
	counts := []int{0, 0}
	for _, v := range r.Types {
		counts[v]++
	}
	if !reflect.DeepEqual(counts, []int{3, 5}) {
		t.Fatalf("counts=%v", counts)
	}
	copyTypes := append([]int(nil), r.Types...)
	r.Graph.Close()
	if !reflect.DeepEqual(r.Types, copyTypes) {
		t.Fatal("types depended on graph")
	}

	directed, err := igraph.PreferenceGame(10, matrix(t, [][]float64{{0.2, 0.8}, {0.1, 0.3}}), igraph.PreferenceOptions{Seed: &seed, Directed: true, Loops: true, TypeWeights: []float64{1, 2}})
	if err != nil {
		t.Fatal(err)
	}
	defer directed.Graph.Close()
	if d, _ := directed.Graph.IsDirected(); !d {
		t.Fatal("expected directed")
	}
	uniform, err := igraph.PreferenceGame(0, pref, igraph.PreferenceOptions{Seed: &seed})
	if err != nil {
		t.Fatal(err)
	}
	defer uniform.Graph.Close()
	if uniform.Types == nil || len(uniform.Types) != 0 {
		t.Fatalf("empty types=%v", uniform.Types)
	}
	dist := matrix(t, [][]float64{{1, 0}, {0, 1}})
	asym, err := igraph.AsymmetricPreferenceGame(12, matrix(t, [][]float64{{0.1, 0.9}, {0.2, 0.3}}), igraph.AsymmetricPreferenceOptions{Seed: &seed, TypeDistribution: &dist})
	if err != nil {
		t.Fatal(err)
	}
	defer asym.Graph.Close()
	if len(asym.OutTypes) != 12 || len(asym.InTypes) != 12 {
		t.Fatalf("type lengths=%d/%d", len(asym.OutTypes), len(asym.InTypes))
	}
	if d, _ := asym.Graph.IsDirected(); !d {
		t.Fatal("asymmetric graph must be directed")
	}
	uniformAsym, err := igraph.AsymmetricPreferenceGame(4, matrix(t, [][]float64{{0.2, 0.3, 0.4}}), igraph.AsymmetricPreferenceOptions{Seed: &seed, Loops: true})
	if err != nil {
		t.Fatal(err)
	}
	defer uniformAsym.Graph.Close()
	if len(uniformAsym.OutTypes) != 4 || len(uniformAsym.InTypes) != 4 {
		t.Fatal("uniform asymmetric types misaligned")
	}

	if _, e := igraph.PreferenceGame(3, pref, igraph.PreferenceOptions{TypeCounts: []int{1, 1}}); e == nil {
		t.Fatal("accepted invalid fixed counts")
	}
	if _, e := igraph.PreferenceGame(3, pref, igraph.PreferenceOptions{TypeWeights: []float64{0, 0}}); e == nil {
		t.Fatal("accepted zero weights")
	}
	if _, e := igraph.PreferenceGame(2, pref, igraph.PreferenceOptions{TypeCounts: []int{-1, 3}}); e == nil {
		t.Fatal("accepted negative type count")
	}
	if _, e := igraph.PreferenceGame(2, pref, igraph.PreferenceOptions{TypeWeights: []float64{-1, 2}}); e == nil {
		t.Fatal("accepted negative type weight")
	}
	if _, e := igraph.PreferenceGame(2, matrix(t, [][]float64{{1, 0}}), igraph.PreferenceOptions{}); e == nil {
		t.Fatal("accepted nonsquare matrix")
	}
	badDist := matrix(t, [][]float64{{1}})
	if _, e := igraph.AsymmetricPreferenceGame(3, pref, igraph.AsymmetricPreferenceOptions{TypeDistribution: &badDist}); e == nil {
		t.Fatal("accepted mismatched distribution")
	}
	badProbability := matrix(t, [][]float64{{1.1, 0}, {0, 1}})
	if _, e := igraph.PreferenceGame(2, badProbability, igraph.PreferenceOptions{}); e == nil {
		t.Fatal("accepted invalid probability")
	}
	nonsymmetric := matrix(t, [][]float64{{0, 1}, {0, 0}})
	if _, e := igraph.PreferenceGame(2, nonsymmetric, igraph.PreferenceOptions{}); e == nil {
		t.Fatal("accepted asymmetric undirected preference")
	}
	if _, e := igraph.PreferenceGame(2, pref, igraph.PreferenceOptions{TypeCounts: []int{1, 1}, TypeWeights: []float64{1, 1}}); e == nil {
		t.Fatal("accepted counts and weights")
	}
}

func TestSimpleInterconnectedIslandsGame(t *testing.T) {
	seed := uint64(336)
	r, err := igraph.SimpleInterconnectedIslandsGame(3, 4, 0.5, 2, igraph.IslandOptions{Seed: &seed})
	if err != nil {
		t.Fatal(err)
	}
	defer r.Graph.Close()
	if !reflect.DeepEqual(r.Islands, []int{0, 0, 0, 0, 1, 1, 1, 1, 2, 2, 2, 2}) {
		t.Fatalf("islands=%v", r.Islands)
	}
	if d, _ := r.Graph.IsDirected(); d {
		t.Fatal("islands graph is directed")
	}
	mustBeSimple(t, r.Graph)
	empty, err := igraph.SimpleInterconnectedIslandsGame(0, 0, 0, 0, igraph.IslandOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer empty.Graph.Close()
	if len(empty.Islands) != 0 {
		t.Fatal("empty memberships not empty")
	}
	if _, e := igraph.SimpleInterconnectedIslandsGame(2, 2, 0, 5, igraph.IslandOptions{}); e == nil {
		t.Fatal("accepted too many inter-island edges")
	}
	if _, e := igraph.SimpleInterconnectedIslandsGame(-1, 2, 0, 0, igraph.IslandOptions{}); e == nil {
		t.Fatal("accepted negative island count")
	}
	if _, e := igraph.SimpleInterconnectedIslandsGame(1, -1, 0, 0, igraph.IslandOptions{}); e == nil {
		t.Fatal("accepted negative island size")
	}
	if _, e := igraph.SimpleInterconnectedIslandsGame(1, 1, math.NaN(), 0, igraph.IslandOptions{}); e == nil {
		t.Fatal("accepted NaN probability")
	}
}

func ExamplePreferenceGame() {
	pref, _ := igraph.NewMatrixFromRows([][]float64{{0.8, 0.1}, {0.1, 0.7}})
	result, err := igraph.PreferenceGame(6, pref, igraph.PreferenceOptions{TypeCounts: []int{3, 3}})
	if err != nil {
		panic(err)
	}
	defer result.Graph.Close()
	fmt.Println(len(result.Types))
	// Output:
	// 6
}
