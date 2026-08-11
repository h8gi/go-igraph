package igraph

import (
	"testing"
)

func TestNewIntVectorInitializationFailure(t *testing.T) {
	vector, err := newIntVectorWithInitializer([]int{1}, func(*intVector, int) int {
		return 1
	})
	if err == nil || vector != nil {
		t.Fatalf("newIntVectorWithInitializer(failure) = %v, %v", vector, err)
	}
}

func TestNewRealVectorSizeNegativeError(t *testing.T) {
	if _, err := newRealVectorSize(-1); err == nil {
		t.Error("expected error for negative size in newRealVectorSize")
	}
}
