package igraph

import (
	"errors"
	"testing"
)

func TestTransposeMatrixRejectsOverflow(t *testing.T) {
	maximum := int(^uint(0) >> 1)
	if _, err := transposeMatrix(Matrix{rows: maximum, columns: 2}); err == nil {
		t.Fatal("transposeMatrix accepted overflowing dimensions")
	}
}

func TestDotProductGamePropagatesTransposeFailure(t *testing.T) {
	if _, err := DotProductGame(Matrix{columns: -1}, LatentGraphOptions{}); err == nil {
		t.Fatal("DotProductGame accepted an invalid internal matrix shape")
	}
}

func TestSampledRowsFailurePaths(t *testing.T) {
	want := errors.New("injected")
	if _, err := sampledRowsWithAdapters(1, 2, LatentSampleOptions{}, "sample", latentSampleAdapters{newMatrix: func(Matrix) (*cMatrix, error) { return nil, want }}); !errors.Is(err, want) {
		t.Fatalf("initialization error=%v", err)
	}
	if _, err := sampledRowsWithAdapters(1, 2, LatentSampleOptions{}, "sample", latentSampleAdapters{newMatrix: newCMatrix, invoke: func(*cMatrix) int { return 4 }, convert: (*cMatrix).matrix}); err == nil {
		t.Fatal("accepted injected upstream error")
	}
	if _, err := sampledRowsWithAdapters(1, 2, LatentSampleOptions{}, "sample", latentSampleAdapters{newMatrix: newCMatrix, invoke: func(*cMatrix) int { return 0 }, convert: func(*cMatrix) (Matrix, error) { return Matrix{}, want }}); !errors.Is(err, want) {
		t.Fatalf("conversion error=%v", err)
	}
}

func TestSampledRowsRejectsOverflow(t *testing.T) {
	maximum := int(^uint(0) >> 1)
	_, err := sampledRows(maximum, 2, LatentSampleOptions{}, "test", nil)
	if err == nil {
		t.Fatal("sampledRows accepted overflowing dimensions")
	}
	if _, err := sampledRows(1, -1, LatentSampleOptions{}, "test", nil); err == nil {
		t.Fatal("sampledRows accepted a negative dimension")
	}
}
