package igraph

import (
	"errors"
	"testing"
)

func TestNewVectorFromSlice(t *testing.T) {
	values := []float64{1, 2.5, -3, 4}
	v, err := NewVectorFromSlice(values)
	if err != nil {
		t.Fatalf("NewVectorFromSlice() error = %v", err)
	}
	t.Cleanup(func() { _ = v.Close() })
	values[0] = 99
	want := []float64{1, 2.5, -3, 4}
	for i, expected := range want {
		got, err := v.Get(i)
		if err != nil {
			t.Fatalf("Get(%d) error = %v", i, err)
		}
		if got != expected {
			t.Errorf("Get(%d) = %v, want %v", i, got, expected)
		}
	}
}

func TestNewVectorFromNilSlice(t *testing.T) {
	v, err := NewVectorFromSlice(nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = v.Close() })
	if _, err := v.Get(0); err == nil {
		t.Error("Get(0) on zero-length vector error = nil")
	}
}

func TestVectorSetAndGet(t *testing.T) {
	v, err := NewVector(2)
	if err != nil {
		t.Fatalf("NewVector() error = %v", err)
	}
	t.Cleanup(func() { _ = v.Close() })
	if err := v.Set(1, 42.5); err != nil {
		t.Fatalf("Set() error = %v", err)
	}
	got, err := v.Get(1)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if got != 42.5 {
		t.Errorf("Get(1) = %v, want 42.5", got)
	}
}

func TestVectorRejectsOutOfRangeIndexes(t *testing.T) {
	for _, size := range []int{0, 2} {
		v, err := NewVector(size)
		if err != nil {
			t.Fatalf("NewVector(%d) error = %v", size, err)
		}
		t.Cleanup(func() { _ = v.Close() })
		for _, index := range []int{-1, size} {
			if err := v.Set(index, 1); err == nil {
				t.Errorf("size %d: Set(%d) error = nil", size, index)
			}
			if _, err := v.Get(index); err == nil {
				t.Errorf("size %d: Get(%d) error = nil", size, index)
			}
		}
	}
}

func TestNewVectorRejectsNegativeSize(t *testing.T) {
	if _, err := NewVector(-1); err == nil {
		t.Error("NewVector(-1) error = nil")
	}
}

func TestVectorClose(t *testing.T) {
	v, err := NewVector(1)
	if err != nil {
		t.Fatalf("NewVector() error = %v", err)
	}
	if err := v.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	if err := v.Close(); err != nil {
		t.Fatalf("second Close() error = %v", err)
	}
	if err := v.Set(0, 1); !errors.Is(err, ErrClosed) {
		t.Errorf("Set() after Close error = %v, want %v", err, ErrClosed)
	}
	if _, err := v.Get(0); !errors.Is(err, ErrClosed) {
		t.Errorf("Get() after Close error = %v, want %v", err, ErrClosed)
	}
}

func TestNilVector(t *testing.T) {
	var v *Vector
	if err := v.Close(); err != nil {
		t.Fatalf("nil Close() error = %v", err)
	}
	if err := v.Set(0, 1); !errors.Is(err, ErrClosed) {
		t.Errorf("nil Set() error = %v, want %v", err, ErrClosed)
	}
	if _, err := v.Get(0); !errors.Is(err, ErrClosed) {
		t.Errorf("nil Get() error = %v, want %v", err, ErrClosed)
	}
}

func TestVectorFinalizerFallback(t *testing.T) {
	v, err := NewVector(1)
	if err != nil {
		t.Fatalf("NewVector() error = %v", err)
	}
	v.finalize()
	if _, err := v.Get(0); !errors.Is(err, ErrClosed) {
		t.Errorf("Get() after finalize error = %v, want %v", err, ErrClosed)
	}
}
