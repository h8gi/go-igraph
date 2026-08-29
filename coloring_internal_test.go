package igraph

import (
	"errors"
	"testing"
)

func TestColoringFailureCleanup(t *testing.T) {
	forced := errors.New("forced")
	g, err := NewGraphFromEdges(2, []Edge{{0, 1}}, false)
	if err != nil {
		t.Fatal(err)
	}
	defer g.Close()
	base := defaultColoringAdapters()

	failedInit := base
	failedInit.newInt = func([]int) (*intVector, error) { return nil, forced }
	if _, err := g.greedyVertexColoring(0, &failedInit); !errors.Is(err, forced) {
		t.Errorf("greedy init = %v", err)
	}
	if _, err := g.validateColoring([]int{0, 1}, false, &failedInit); !errors.Is(err, forced) {
		t.Errorf("validate init = %v", err)
	}

	closed := 0
	upstream := base
	upstream.closeInt = func(*intVector) { closed++ }
	upstream.greedy = func(*Graph, *intVector, ColoringHeuristic) int { return 4 }
	if _, err := g.greedyVertexColoring(0, &upstream); err == nil || closed != 1 {
		t.Errorf("greedy upstream = %v, closed %d", err, closed)
	}

	closed = 0
	convert := base
	convert.closeInt = func(*intVector) { closed++ }
	convert.intSlice = func(*intVector) ([]int, error) { return nil, forced }
	if _, err := g.greedyVertexColoring(0, &convert); !errors.Is(err, forced) || closed != 1 {
		t.Errorf("greedy conversion = %v, closed %d", err, closed)
	}

	closed = 0
	validation := base
	validation.closeInt = func(*intVector) { closed++ }
	validation.validate = func(*Graph, *intVector, bool) (bool, int) { return false, 4 }
	if _, err := g.validateColoring([]int{0, 1}, false, &validation); err == nil || closed != 1 {
		t.Errorf("validate upstream = %v, closed %d", err, closed)
	}

	boolInit := base
	boolInit.newBool = func([]bool) (*boolVector, error) { return nil, forced }
	if _, err := g.isBipartiteColoring(BipartitePartition{false, true}, &boolInit); !errors.Is(err, forced) {
		t.Errorf("bipartite init = %v", err)
	}
	boolClosed := 0
	boolUpstream := base
	boolUpstream.closeBool = func(*boolVector) { boolClosed++ }
	boolUpstream.bipartite = func(*Graph, *boolVector) (bool, DirectionMode, int) { return false, DirectionIn, 4 }
	if _, err := g.isBipartiteColoring(BipartitePartition{false, true}, &boolUpstream); err == nil || boolClosed != 1 {
		t.Errorf("bipartite upstream = %v, closed %d", err, boolClosed)
	}
}

func TestColoringValidationHelpers(t *testing.T) {
	if err := validateColors([]int{0, 2}, 2, "vertex"); err != nil {
		t.Fatal(err)
	}
	for _, colors := range [][]int{{0}, {0, -1}} {
		if err := validateColors(colors, 2, "vertex"); err == nil {
			t.Errorf("validateColors(%v) = nil", colors)
		}
	}
	if result := (BipartiteColoringResult{Valid: false, Direction: DirectionAll}); result.Direction != DirectionAll {
		t.Fatal(result)
	}
}
