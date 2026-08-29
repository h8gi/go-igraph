package igraph

import (
	"errors"
	"reflect"
	"testing"
)

func TestMycielskianInjectedFailuresClosePartialResult(t *testing.T) {
	source := testGraphFromEdges(t, 2, []Edge{{0, 1}}, false)
	defer source.Close()
	for _, stage := range []mycielskiFailureStage{mycielskiAfterConstruction, mycielskiAfterProvenance} {
		target := stage
		result, err := source.mycielskian(1, func(current mycielskiFailureStage) error {
			if current == target {
				return errors.New("injected Mycielski failure")
			}
			return nil
		})
		if err == nil || result.Graph != nil {
			t.Errorf("Mycielski stage %d = %#v, %v", stage, result, err)
		}
		assertGraphShape(t, source, 2, 1, false)
	}
}

func TestMycielskiProvenanceHelpers(t *testing.T) {
	if _, _, err := checkedMycielskiSize(-1, 0, 0); err == nil {
		t.Error("negative source count error = nil")
	}
	if _, _, err := checkedMycielskiSize(0, -1, 0); err == nil {
		t.Error("negative edge count error = nil")
	}
	if vertices, edges, err := checkedMycielskiSize(2, 1, 2); err != nil || vertices != 11 || edges != 20 {
		t.Errorf("checked size = %d/%d, %v", vertices, edges, err)
	}
	vertices, sources := mycielskiProvenance(0, 2)
	want := []MycielskiVertexProvenance{
		{Kind: MycielskiVertexApex, Generation: 1, SourceVertex: RemovedID, ParentVertex: RemovedID},
		{Kind: MycielskiVertexApex, Generation: 2, SourceVertex: RemovedID, ParentVertex: RemovedID},
	}
	if !reflect.DeepEqual(vertices, want) || sources == nil || len(sources) != 0 {
		t.Errorf("empty-source provenance = %#v / %#v", vertices, sources)
	}
	if err := runMycielskiFailureHook(nil, mycielskiAfterConstruction); err != nil {
		t.Errorf("nil hook: %v", err)
	}
}
