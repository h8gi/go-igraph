package igraph

import (
	"errors"
	"reflect"
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
	if _, err := g.ConvertToUndirectedInPlace(UndirectedConversionMode(99), nil); err == nil {
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
		_, err = gSub.convertToUndirectedInPlace(UndirectedConversionCollapse, nil, func(s graphTransformationStage) error {
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

func TestContractionAndReversalHookFailuresAreAtomic(t *testing.T) {
	for _, stage := range []graphTransformationStage{
		graphTransformationAtClone,
		graphTransformationAtTransform,
		graphTransformationAfterTransform,
	} {
		contract := transformationGraph(t, 3, []Edge{{0, 1}, {1, 2}}, true)
		before, _ := contract.Edges()
		target := stage
		_, err := contract.contractVerticesInPlace([]int{0, 0, 1}, nil, func(current graphTransformationStage) error {
			if current == target {
				return errors.New("injected contraction failure")
			}
			return nil
		})
		if err == nil {
			t.Errorf("contraction stage %d error = nil", stage)
		}
		after, _ := contract.Edges()
		vertices, _ := contract.VertexCount()
		if vertices != 3 || !reflect.DeepEqual(after, before) {
			t.Errorf("contraction stage %d mutated graph: vertices=%d edges=%v", stage, vertices, after)
		}

		reverse := transformationGraph(t, 3, []Edge{{0, 1}, {1, 2}}, true)
		before, _ = reverse.Edges()
		_, err = reverse.reverseEdgesInPlace(AllEdges(), func(current graphTransformationStage) error {
			if current == target {
				return errors.New("injected reversal failure")
			}
			return nil
		})
		if err == nil {
			t.Errorf("reversal stage %d error = nil", stage)
		}
		after, _ = reverse.Edges()
		if !reflect.DeepEqual(after, before) {
			t.Errorf("reversal stage %d mutated graph: %v", stage, after)
		}
	}
}

func TestTransformationNormalizationHelpers(t *testing.T) {
	if got := uniqueSelectedEdgeIDs([]int{3, 1, 3, 2, 1}); !reflect.DeepEqual(got, []int{3, 1, 2}) {
		t.Errorf("unique edge IDs = %v", got)
	}
	if _, _, _, err := normalizeContractionMapping([]int{0}, 2); err == nil {
		t.Error("mapping length error = nil")
	}
	if _, err := contractionResult([]int{0}, 0, 0); err == nil {
		t.Error("invalid contraction vertex mapping error = nil")
	}
	if _, err := contractionResult(nil, 0, -1); err == nil {
		t.Error("invalid contraction edge count error = nil")
	}
	if _, err := identityGraphTransformationResult(-1, 0); err == nil {
		t.Error("invalid identity vertex count error = nil")
	}
	if _, err := identityGraphTransformationResult(0, -1); err == nil {
		t.Error("invalid identity edge count error = nil")
	}
}
