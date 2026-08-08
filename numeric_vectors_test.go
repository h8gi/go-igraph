package igraph

import (
	"reflect"
	"testing"
)

func TestIntVectorRoundTrip(t *testing.T) {
	for _, values := range [][]int{nil, {}, {0, 1, -2, int(^uint(0) >> 1)}} {
		vector, err := newIntVector(values)
		if err != nil {
			t.Fatalf("newIntVector(%v) error = %v", values, err)
		}
		got, err := vector.slice()
		if err != nil {
			t.Fatalf("integer slice error = %v", err)
		}
		vector.close()

		want := append([]int{}, values...)
		if !reflect.DeepEqual(got, want) {
			t.Errorf("integer round trip = %#v, want %#v", got, want)
		}
		if got == nil {
			t.Error("integer round trip returned a nil slice")
		}
	}
}

func TestRealVectorRoundTrip(t *testing.T) {
	for _, values := range [][]float64{nil, {}, {0, 1.5, -2.25}} {
		vector, err := newRealVector(values)
		if err != nil {
			t.Fatalf("newRealVector(%v) error = %v", values, err)
		}
		got, err := vector.slice()
		if err != nil {
			t.Fatalf("real slice error = %v", err)
		}
		vector.close()

		want := append([]float64{}, values...)
		if !reflect.DeepEqual(got, want) {
			t.Errorf("real round trip = %#v, want %#v", got, want)
		}
		if got == nil {
			t.Error("real round trip returned a nil slice")
		}
	}
}

func TestNumericVectorResultsOwnTheirStorage(t *testing.T) {
	integers, err := newIntVector([]int{1, 2})
	if err != nil {
		t.Fatal(err)
	}
	intResult, err := integers.slice()
	if err != nil {
		t.Fatal(err)
	}
	integers.close()
	intResult[0] = 9
	if !reflect.DeepEqual(intResult, []int{9, 2}) {
		t.Fatalf("integer result after close = %v", intResult)
	}

	reals, err := newRealVector([]float64{1, 2})
	if err != nil {
		t.Fatal(err)
	}
	realResult, err := reals.slice()
	if err != nil {
		t.Fatal(err)
	}
	reals.close()
	realResult[0] = 9
	if !reflect.DeepEqual(realResult, []float64{9, 2}) {
		t.Fatalf("real result after close = %v", realResult)
	}
}
