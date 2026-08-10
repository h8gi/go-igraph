package igraph

import (
	"errors"
	"testing"
)

func TestDeleteVerticesWithHookFailures(t *testing.T) {
	for _, stage := range []deletionStage{
		deletionBeforeClone,
		deletionBeforeEdgeSnapshot,
		deletionBeforeSelectorInit,
		deletionBeforeFirstMappingInit,
		deletionBeforeSecondMappingInit,
		deletionAtMutation,
		deletionAfterMutation,
	} {
		g := testGraphFromEdges(t, 3, []Edge{{0, 1}, {1, 2}}, false)
		selector, err := VertexIDs(0)
		if err != nil {
			g.Close()
			t.Fatal(err)
		}

		targetStage := stage
		_, err = g.deleteVertices(selector, func(s deletionStage) error {
			if s == targetStage {
				return errors.New("simulated failure")
			}
			return nil
		})
		g.Close()
		if err == nil {
			t.Errorf("expected error for stage %d failure", targetStage)
		}
	}
}
