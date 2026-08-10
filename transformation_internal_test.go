package igraph

import (
	"errors"
	"testing"
)

func TestToUndirectedWithHookFailures(t *testing.T) {
	g, err := NewGraphFromEdges(2, []Edge{{0, 1}}, true)
	if err != nil {
		t.Fatal(err)
	}
	defer g.Close()

	if _, err := g.ConvertToDirectedInPlace(DirectedConversionMode(99)); err == nil {
		t.Error("expected error for invalid DirectedConversionMode")
	}
	if _, err := g.ConvertToUndirectedInPlace(UndirectedConversionMode(99)); err == nil {
		t.Error("expected error for invalid UndirectedConversionMode")
	}

	for _, stage := range []graphTransformationStage{
		graphTransformationAtClone,
		graphTransformationAtTransform,
		graphTransformationAfterTransform,
	} {
		gSub, err := NewGraphFromEdges(2, []Edge{{0, 1}}, true)
		if err != nil {
			t.Fatal(err)
		}
		targetStage := stage
		_, err = gSub.convertToUndirectedInPlace(UndirectedConversionCollapse, func(s graphTransformationStage) error {
			if s == targetStage {
				return errors.New("simulated transformation failure")
			}
			return nil
		})
		gSub.Close()
		if err == nil {
			t.Errorf("expected error for transformation stage %d failure", targetStage)
		}
	}
}
