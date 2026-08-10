package igraph

import (
	"testing"
)

func TestVectorInternalMethodsAndValidations(t *testing.T) {
	if _, err := NewVector(-1); err == nil {
		t.Error("expected error for negative size in NewVector")
	}

	var nilVec *Vector
	if err := nilVec.Set(0, 1.0); err != ErrClosed {
		t.Errorf("nil Vector.Set error = %v, want ErrClosed", err)
	}
	if _, err := nilVec.Get(0); err != ErrClosed {
		t.Errorf("nil Vector.Get error = %v, want ErrClosed", err)
	}
	if err := nilVec.Close(); err != nil {
		t.Errorf("nil Vector.Close error = %v, want nil", err)
	}

	vec, err := NewVector(2)
	if err != nil {
		t.Fatal(err)
	}
	defer vec.Close()

	if err := vec.Set(-1, 1.0); err == nil {
		t.Error("expected error for out of range index in Vector.Set")
	}
	if err := vec.Set(5, 1.0); err == nil {
		t.Error("expected error for out of range index in Vector.Set")
	}

	if _, err := vec.Get(-1); err == nil {
		t.Error("expected error for out of range index in Vector.Get")
	}
	if _, err := vec.Get(5); err == nil {
		t.Error("expected error for out of range index in Vector.Get")
	}

	// Double Close
	if err := vec.Close(); err != nil {
		t.Errorf("first Close error = %v", err)
	}
	if err := vec.Close(); err != nil {
		t.Errorf("second Close error = %v", err)
	}

	// Finalizer call
	vec.finalize()
}
