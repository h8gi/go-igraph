package igraph

import (
	"errors"
	"testing"
)

func TestDerivedGraphUninitializedListError(t *testing.T) {
	var list graphList
	if _, err := list.takeGraphs(); err == nil {
		t.Error("expected error for uninitialized graphList.takeGraphs")
	}
	if err := list.appendCopy(nil); err == nil {
		t.Error("expected error for uninitialized graphList.appendCopy")
	}
}

func TestGraphListTakeGraphsHooks(t *testing.T) {
	g, err := NewGraphFromEdges(2, []Edge{{0, 1}}, false)
	if err != nil {
		t.Fatal(err)
	}
	defer g.Close()

	list, err := newGraphList()
	if err != nil {
		t.Fatal(err)
	}
	if err := list.appendCopy(&g.graph); err != nil {
		list.close()
		t.Fatal(err)
	}

	_, err = list.takeGraphsWithHooks(graphListExtractionHooks{
		beforeRemove: func(index int) error {
			return errors.New("simulated beforeRemove error")
		},
	})
	if err == nil {
		t.Error("expected error for beforeRemove hook failure")
	}

	list2, err := newGraphList()
	if err != nil {
		t.Fatal(err)
	}
	if err := list2.appendCopy(&g.graph); err != nil {
		list2.close()
		t.Fatal(err)
	}

	_, err = list2.takeGraphsWithHooks(graphListExtractionHooks{
		beforeAdopt: func(index int) error {
			return errors.New("simulated beforeAdopt error")
		},
	})
	if err == nil {
		t.Error("expected error for beforeAdopt hook failure")
	}
}
