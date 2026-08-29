package igraph_test

import (
	"fmt"
	"math"
	"reflect"
	"testing"

	"github.com/h8gi/go-igraph"
)

func TestTraitGrowthGames(t *testing.T) {
	seed := uint64(338)
	pref := matrix(t, [][]float64{{0.8, 0.1}, {0.1, 0.7}})
	options := igraph.TraitGrowthOptions{Seed: &seed, TypeWeights: []float64{1, 2}}
	first, err := igraph.CallawayTraitsGame(12, 3, pref, options)
	if err != nil {
		t.Fatal(err)
	}
	defer first.Graph.Close()
	second, err := igraph.CallawayTraitsGame(12, 3, pref, options)
	if err != nil {
		t.Fatal(err)
	}
	defer second.Graph.Close()
	if !reflect.DeepEqual(first.Types, second.Types) || !reflect.DeepEqual(mustEdges(t, first.Graph), mustEdges(t, second.Graph)) {
		t.Fatal("same seed differed")
	}
	if len(first.Types) != 12 {
		t.Fatalf("types=%v", first.Types)
	}
	copyTypes := append([]int(nil), first.Types...)
	first.Graph.Close()
	if !reflect.DeepEqual(first.Types, copyTypes) {
		t.Fatal("types depended on graph")
	}
	established, err := igraph.EstablishmentGame(10, 2, pref, igraph.TraitGrowthOptions{Seed: &seed})
	if err != nil {
		t.Fatal(err)
	}
	defer established.Graph.Close()
	if len(established.Types) != 10 {
		t.Fatalf("establishment types=%v", established.Types)
	}
	if d, _ := established.Graph.IsDirected(); d {
		t.Fatal("expected undirected")
	}
	directed, err := igraph.CallawayTraitsGame(4, 1, matrix(t, [][]float64{{0, 1}, {0, 0}}), igraph.TraitGrowthOptions{Seed: &seed, Directed: true})
	if err != nil {
		t.Fatal(err)
	}
	defer directed.Graph.Close()
	empty, err := igraph.CallawayTraitsGame(0, 0, matrix(t, [][]float64{{1}}), igraph.TraitGrowthOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer empty.Graph.Close()
	if empty.Types == nil || len(empty.Types) != 0 {
		t.Fatalf("empty types=%v", empty.Types)
	}
	for i, call := range []func() error{
		func() error { _, e := igraph.CallawayTraitsGame(-1, 0, pref, igraph.TraitGrowthOptions{}); return e },
		func() error { _, e := igraph.EstablishmentGame(1, -1, pref, igraph.TraitGrowthOptions{}); return e },
		func() error {
			_, e := igraph.CallawayTraitsGame(1, 0, matrix(t, [][]float64{{0, 1}, {0, 0}}), igraph.TraitGrowthOptions{})
			return e
		},
		func() error {
			_, e := igraph.CallawayTraitsGame(1, 0, pref, igraph.TraitGrowthOptions{TypeWeights: []float64{1}})
			return e
		},
		func() error {
			_, e := igraph.CallawayTraitsGame(1, 0, pref, igraph.TraitGrowthOptions{TypeWeights: []float64{0, 0}})
			return e
		},
		func() error {
			_, e := igraph.CallawayTraitsGame(1, 0, pref, igraph.TraitGrowthOptions{TypeWeights: []float64{math.NaN(), 1}})
			return e
		},
		func() error {
			_, e := igraph.CallawayTraitsGame(1, 0, matrix(t, [][]float64{}), igraph.TraitGrowthOptions{})
			return e
		},
		func() error {
			_, e := igraph.CallawayTraitsGame(1, 0, matrix(t, [][]float64{{2}}), igraph.TraitGrowthOptions{})
			return e
		},
		func() error {
			_, e := igraph.CallawayTraitsGame(1, 0, pref, igraph.TraitGrowthOptions{TypeWeights: []float64{-1, 2}})
			return e
		},
	} {
		if call() == nil {
			t.Errorf("validation %d accepted", i)
		}
	}
}

func TestCitationGames(t *testing.T) {
	seed := uint64(338)
	options := igraph.CitationOptions{Seed: &seed, Directed: true}
	first, err := igraph.LastCitationGame(9, 2, []float64{3, 2, 1}, options)
	if err != nil {
		t.Fatal(err)
	}
	defer first.Close()
	second, err := igraph.LastCitationGame(9, 2, []float64{3, 2, 1}, options)
	if err != nil {
		t.Fatal(err)
	}
	defer second.Close()
	if !reflect.DeepEqual(mustEdges(t, first), mustEdges(t, second)) {
		t.Fatal("last-citation same seed differed")
	}
	if v, e := mustCounts(t, first); v != 9 || e != 16 {
		t.Fatalf("last-citation counts=%d/%d", v, e)
	}
	types := []int{0, 1, 0, 1, 1, 0}
	cited, err := igraph.CitedTypeGame(types, []float64{1, 3}, 2, options)
	if err != nil {
		t.Fatal(err)
	}
	defer cited.Close()
	if v, e := mustCounts(t, cited); v != 6 || e != 10 {
		t.Fatalf("cited counts=%d/%d", v, e)
	}
	if d, _ := cited.IsDirected(); !d {
		t.Fatal("citation graph not directed")
	}
	citing, err := igraph.CitingCitedTypeGame(types, matrix(t, [][]float64{{4, 1}, {1, 4}}), 1, options)
	if err != nil {
		t.Fatal(err)
	}
	defer citing.Close()
	if v, e := mustCounts(t, citing); v != 6 || e != 5 {
		t.Fatalf("citing/cited counts=%d/%d", v, e)
	}
	empty, err := igraph.CitedTypeGame(nil, nil, 0, igraph.CitationOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer empty.Close()
	if v, e := mustCounts(t, empty); v != 0 || e != 0 {
		t.Fatalf("empty counts=%d/%d", v, e)
	}
	emptyMatrix, _ := igraph.NewMatrix(0, 0)
	emptyCiting, err := igraph.CitingCitedTypeGame(nil, emptyMatrix, 0, igraph.CitationOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer emptyCiting.Close()
	for i, call := range []func() error{
		func() error { _, e := igraph.LastCitationGame(1, 0, []float64{1}, options); return e },
		func() error { _, e := igraph.LastCitationGame(1, 0, []float64{1, 0}, options); return e },
		func() error { _, e := igraph.LastCitationGame(1, 0, []float64{-1, 1}, options); return e },
		func() error { _, e := igraph.LastCitationGame(-1, 0, []float64{1, 1}, options); return e },
		func() error { _, e := igraph.LastCitationGame(1, -1, []float64{1, 1}, options); return e },
		func() error { _, e := igraph.LastCitationGame(1, 0, []float64{math.NaN(), 1}, options); return e },
		func() error { _, e := igraph.CitedTypeGame([]int{-1}, []float64{1}, 0, options); return e },
		func() error { _, e := igraph.CitedTypeGame([]int{0, 1}, []float64{1}, 1, options); return e },
		func() error { _, e := igraph.CitedTypeGame([]int{0, 1}, []float64{0, 1}, 1, options); return e },
		func() error { _, e := igraph.CitedTypeGame([]int{0}, []float64{math.Inf(1)}, 0, options); return e },
		func() error {
			_, e := igraph.CitingCitedTypeGame([]int{0, 1}, matrix(t, [][]float64{{1}}), 1, options)
			return e
		},
		func() error {
			_, e := igraph.CitingCitedTypeGame([]int{0}, matrix(t, [][]float64{{-1}}), 0, options)
			return e
		},
		func() error {
			_, e := igraph.CitingCitedTypeGame([]int{0}, matrix(t, [][]float64{{math.NaN()}}), 0, options)
			return e
		},
	} {
		if call() == nil {
			t.Errorf("validation %d accepted", i)
		}
	}
}

func ExampleCitedTypeGame() {
	graph, err := igraph.CitedTypeGame([]int{0, 1, 0, 1}, []float64{1, 2}, 1, igraph.CitationOptions{Directed: true})
	if err != nil {
		panic(err)
	}
	defer graph.Close()
	vertices, _ := graph.VertexCount()
	fmt.Println(vertices)
	// Output:
	// 4
}
