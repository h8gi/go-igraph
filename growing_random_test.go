package igraph_test

import (
	"fmt"
	"math"
	"reflect"
	"testing"

	"github.com/h8gi/go-igraph"
)

func TestGrowingRandomGame(t *testing.T) {
	seed := uint64(337)
	options := igraph.GrowingRandomOptions{Seed: &seed, Directed: true, Citation: true}
	first, err := igraph.GrowingRandomGame(10, 2, options)
	if err != nil {
		t.Fatal(err)
	}
	defer first.Close()
	second, err := igraph.GrowingRandomGame(10, 2, options)
	if err != nil {
		t.Fatal(err)
	}
	defer second.Close()
	if !reflect.DeepEqual(mustEdges(t, first), mustEdges(t, second)) {
		t.Fatal("same seed differed")
	}
	vertices, edges := mustCounts(t, first)
	if vertices != 10 || edges != 18 {
		t.Fatalf("counts=%d/%d", vertices, edges)
	}
	if directed, _ := first.IsDirected(); !directed {
		t.Fatal("expected directed graph")
	}
	empty, err := igraph.GrowingRandomGame(0, 0, igraph.GrowingRandomOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer empty.Close()
	if v, e := mustCounts(t, empty); v != 0 || e != 0 {
		t.Fatalf("empty counts=%d/%d", v, e)
	}
	if _, err := igraph.GrowingRandomGame(-1, 0, igraph.GrowingRandomOptions{}); err == nil {
		t.Fatal("accepted negative vertices")
	}
	if _, err := igraph.GrowingRandomGame(1, -1, igraph.GrowingRandomOptions{}); err == nil {
		t.Fatal("accepted negative attachment")
	}
}

func TestForestFireGame(t *testing.T) {
	seed := uint64(337)
	options := igraph.ForestFireOptions{Seed: &seed, BackwardFactor: 0.5, Ambassadors: 1, Directed: true}
	first, err := igraph.ForestFireGame(12, 0.25, options)
	if err != nil {
		t.Fatal(err)
	}
	defer first.Close()
	second, err := igraph.ForestFireGame(12, 0.25, options)
	if err != nil {
		t.Fatal(err)
	}
	defer second.Close()
	if !reflect.DeepEqual(mustEdges(t, first), mustEdges(t, second)) {
		t.Fatal("same seed differed")
	}
	if v, _ := first.VertexCount(); v != 12 {
		t.Fatalf("vertices=%d", v)
	}
	isolated, err := igraph.ForestFireGame(3, 0, igraph.ForestFireOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer isolated.Close()
	if _, e := mustCounts(t, isolated); e != 0 {
		t.Fatalf("edges=%d", e)
	}
	for i, call := range []func() error{func() error { _, e := igraph.ForestFireGame(1, 1, igraph.ForestFireOptions{}); return e }, func() error {
		_, e := igraph.ForestFireGame(1, 0.6, igraph.ForestFireOptions{BackwardFactor: 2})
		return e
	}, func() error {
		_, e := igraph.ForestFireGame(1, 0, igraph.ForestFireOptions{BackwardFactor: math.Inf(1)})
		return e
	}, func() error { _, e := igraph.ForestFireGame(1, 0, igraph.ForestFireOptions{Ambassadors: -1}); return e }} {
		if call() == nil {
			t.Errorf("validation %d accepted", i)
		}
	}
}

func TestAgingGrowthGames(t *testing.T) {
	seed := uint64(337)
	schedule := igraph.AttachmentSchedule{OutSequence: []int{0, 1, 2, 1, 2, 1}}
	barabasi := igraph.BarabasiAgingOptions{Seed: &seed, Schedule: schedule, AttachmentExponent: 1, AgingExponent: -1, AgingBins: 3, ZeroDegreeAppeal: 1, ZeroAgeAppeal: 1, DegreeCoefficient: 1, AgeCoefficient: 1, Directed: true}
	g, err := igraph.BarabasiAgingGame(6, barabasi)
	if err != nil {
		t.Fatal(err)
	}
	defer g.Close()
	if v, e := mustCounts(t, g); v != 6 || e != 7 {
		t.Fatalf("Barabasi counts=%d/%d", v, e)
	}
	recent := igraph.RecentDegreeOptions{Seed: &seed, Schedule: igraph.AttachmentSchedule{EdgesPerStep: 1}, Exponent: 1, Window: 3, ZeroAppeal: 1, Directed: true}
	r1, err := igraph.RecentDegreeGame(8, recent)
	if err != nil {
		t.Fatal(err)
	}
	defer r1.Close()
	r2, err := igraph.RecentDegreeGame(8, recent)
	if err != nil {
		t.Fatal(err)
	}
	defer r2.Close()
	if !reflect.DeepEqual(mustEdges(t, r1), mustEdges(t, r2)) {
		t.Fatal("recent-degree same seed differed")
	}
	sequenceRecent, err := igraph.RecentDegreeGame(6, igraph.RecentDegreeOptions{Seed: &seed, Schedule: schedule, Exponent: 1, Window: 2, ZeroAppeal: 1})
	if err != nil {
		t.Fatal(err)
	}
	defer sequenceRecent.Close()
	if _, edges := mustCounts(t, sequenceRecent); edges != 7 {
		t.Fatalf("sequence recent edges=%d", edges)
	}
	combined := igraph.RecentDegreeAgingOptions{Seed: &seed, Schedule: igraph.AttachmentSchedule{EdgesPerStep: 1}, AttachmentExponent: 1, AgingExponent: -1, AgingBins: 4, Window: 3, ZeroAppeal: 1}
	ra, err := igraph.RecentDegreeAgingGame(8, combined)
	if err != nil {
		t.Fatal(err)
	}
	defer ra.Close()
	if v, e := mustCounts(t, ra); v != 8 || e != 7 {
		t.Fatalf("combined counts=%d/%d", v, e)
	}
	zero, err := igraph.RecentDegreeGame(1, igraph.RecentDegreeOptions{Exponent: 1, Window: 0, ZeroAppeal: 1})
	if err != nil {
		t.Fatal(err)
	}
	defer zero.Close()
	if v, e := mustCounts(t, zero); v != 1 || e != 0 {
		t.Fatalf("singleton counts=%d/%d", v, e)
	}

	badSchedule := igraph.AttachmentSchedule{OutSequence: []int{0, -1}}
	for i, call := range []func() error{
		func() error { _, e := igraph.BarabasiAgingGame(2, igraph.BarabasiAgingOptions{AgingBins: 0}); return e },
		func() error {
			_, e := igraph.BarabasiAgingGame(2, igraph.BarabasiAgingOptions{AgingBins: 1, AttachmentExponent: math.NaN()})
			return e
		},
		func() error {
			_, e := igraph.BarabasiAgingGame(2, igraph.BarabasiAgingOptions{AgingBins: 1, AgingExponent: math.Inf(1)})
			return e
		},
		func() error {
			_, e := igraph.BarabasiAgingGame(2, igraph.BarabasiAgingOptions{AgingBins: 1, ZeroDegreeAppeal: -1})
			return e
		},
		func() error {
			_, e := igraph.BarabasiAgingGame(2, igraph.BarabasiAgingOptions{AgingBins: 1, ZeroAgeAppeal: math.NaN()})
			return e
		},
		func() error {
			_, e := igraph.BarabasiAgingGame(2, igraph.BarabasiAgingOptions{AgingBins: 1, DegreeCoefficient: -1})
			return e
		},
		func() error {
			_, e := igraph.BarabasiAgingGame(2, igraph.BarabasiAgingOptions{AgingBins: 1, AgeCoefficient: -1})
			return e
		},
		func() error {
			_, e := igraph.RecentDegreeGame(2, igraph.RecentDegreeOptions{Schedule: badSchedule, Window: 1})
			return e
		},
		func() error {
			_, e := igraph.RecentDegreeGame(2, igraph.RecentDegreeOptions{Schedule: igraph.AttachmentSchedule{OutSequence: []int{0}}, Window: 1})
			return e
		},
		func() error { _, e := igraph.RecentDegreeGame(2, igraph.RecentDegreeOptions{Window: -1}); return e },
		func() error {
			_, e := igraph.RecentDegreeGame(2, igraph.RecentDegreeOptions{Window: 1, Exponent: math.NaN()})
			return e
		},
		func() error {
			_, e := igraph.RecentDegreeGame(2, igraph.RecentDegreeOptions{Window: 1, ZeroAppeal: -1})
			return e
		},
		func() error {
			_, e := igraph.RecentDegreeAgingGame(2, igraph.RecentDegreeAgingOptions{AgingBins: 1, Window: -1})
			return e
		},
		func() error {
			_, e := igraph.RecentDegreeAgingGame(2, igraph.RecentDegreeAgingOptions{AgingBins: 0})
			return e
		},
		func() error {
			_, e := igraph.RecentDegreeAgingGame(2, igraph.RecentDegreeAgingOptions{AgingBins: 1, AttachmentExponent: math.NaN()})
			return e
		},
		func() error {
			_, e := igraph.RecentDegreeAgingGame(2, igraph.RecentDegreeAgingOptions{AgingBins: 1, AgingExponent: math.Inf(1)})
			return e
		},
		func() error {
			_, e := igraph.RecentDegreeAgingGame(2, igraph.RecentDegreeAgingOptions{AgingBins: 1, ZeroAppeal: -1})
			return e
		},
	} {
		if call() == nil {
			t.Errorf("validation %d accepted", i)
		}
	}
}

func ExampleRecentDegreeGame() {
	seed := uint64(7)
	graph, err := igraph.RecentDegreeGame(10, igraph.RecentDegreeOptions{Seed: &seed, Schedule: igraph.AttachmentSchedule{EdgesPerStep: 1}, Exponent: 1, Window: 3, ZeroAppeal: 1, Directed: true})
	if err != nil {
		panic(err)
	}
	defer graph.Close()
	vertices, _ := graph.VertexCount()
	fmt.Println(vertices)
	// Output:
	// 10
}
