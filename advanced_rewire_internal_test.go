package igraph

import (
	"errors"
	"reflect"
	"testing"
)

func TestRewireDirectedEdgesFailuresAreAtomic(t *testing.T) {
	failure := errors.New("injected")
	for _, stage := range []directedRewireStage{directedRewireBeforeClone, directedRewireAtMutation, directedRewireAfterMutation} {
		graph, err := NewGraphFromEdges(3, []Edge{{From: 0, To: 1}, {From: 1, To: 2}}, true)
		if err != nil {
			t.Fatal(err)
		}
		before, err := edgeSlice(&graph.graph)
		if err != nil {
			t.Fatal(err)
		}
		err = graph.rewireDirectedEdges(1, DirectionOut, true, RewireOptions{}, func(current directedRewireStage) error {
			if current == stage {
				return failure
			}
			return nil
		})
		if !errors.Is(err, failure) {
			t.Fatalf("stage %d error=%v", stage, err)
		}
		after, sliceErr := edgeSlice(&graph.graph)
		if sliceErr != nil {
			t.Fatal(sliceErr)
		}
		if !reflect.DeepEqual(before, after) {
			t.Fatalf("stage %d changed graph: %v -> %v", stage, before, after)
		}
		_ = graph.Close()
	}
}
