package igraph

import (
	"testing"
)

func TestNewOrderVertexSelector(t *testing.T) {
	t.Run("nil order returns all selector", func(t *testing.T) {
		sel, err := newOrderVertexSelector(nil, 5)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		defer sel.close()
		if sel.owned {
			t.Errorf("got owned %v, want false for all selector", sel.owned)
		}
	})

	t.Run("mismatched order length", func(t *testing.T) {
		if _, err := newOrderVertexSelector([]int{0, 1}, 5); err == nil {
			t.Error("expected error for mismatched order length")
		}
	})

	t.Run("out of bounds order index", func(t *testing.T) {
		if _, err := newOrderVertexSelector([]int{0, 1, 2, 3, 10}, 5); err == nil {
			t.Error("expected error for out of bounds vertex ID")
		}
		if _, err := newOrderVertexSelector([]int{0, 1, -1, 3, 4}, 5); err == nil {
			t.Error("expected error for negative vertex ID")
		}
	})

	t.Run("valid custom order", func(t *testing.T) {
		sel, err := newOrderVertexSelector([]int{4, 3, 2, 1, 0}, 5)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		defer sel.close()
		if !sel.owned {
			t.Errorf("got owned %v, want true for vector selector", sel.owned)
		}
	})
}
