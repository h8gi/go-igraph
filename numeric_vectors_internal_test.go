package igraph

import (
	"testing"
)

func TestNewRealVectorSizeNegativeError(t *testing.T) {
	if _, err := newRealVectorSize(-1); err == nil {
		t.Error("expected error for negative size in newRealVectorSize")
	}
}
