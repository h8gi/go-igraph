package igraph

import "testing"

func TestNewVectorFromSlice(t *testing.T) {
	want := []float64{1, 2.5, -3, 4}
	v := NewVectorFromSlice(want)
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

func TestVectorSetAndGet(t *testing.T) {
	v := NewVector(2)
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
		v := NewVector(size)
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

func TestVectorViewCopiesInput(t *testing.T) {
	values := []float64{1, 2}
	v := VectorView(values)
	values[0] = 99

	got, err := v.Get(0)
	if err != nil {
		t.Fatalf("Get(0) error = %v", err)
	}
	if got != 1 {
		t.Errorf("Get(0) = %v after input mutation, want 1", got)
	}
}
