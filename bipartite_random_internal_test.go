package igraph

import (
	"errors"
	"testing"
)

func TestRandomBipartiteFailureAdapters(t *testing.T) {
	failure := errors.New("injected failure")
	base := func() *randomBipartiteAdapters {
		return &randomBipartiteAdapters{
			newBool:     newBoolVector,
			convertBool: (*boolVector).slice,
			closeBool:   (*boolVector).close,
			call: func(types *boolVector, _ DirectionMode) bipartiteGraphCallResult {
				return defaultBipartiteAdapters().full(1, 1, false, DirectionOut, types)
			},
		}
	}

	initFailure := base()
	initFailure.newBool = func([]bool) (*boolVector, error) { return nil, failure }
	if _, err := randomBipartiteGraphWithAdapters(1, 1, BipartiteRandomOptions{}, "test", nil, initFailure); !errors.Is(err, failure) {
		t.Fatalf("initialization error = %v", err)
	}

	upstreamFailure := base()
	upstreamFailure.call = func(*boolVector, DirectionMode) bipartiteGraphCallResult {
		return bipartiteGraphCallResult{code: 4}
	}
	if _, err := randomBipartiteGraphWithAdapters(1, 1, BipartiteRandomOptions{}, "test", nil, upstreamFailure); err == nil {
		t.Fatal("expected upstream error")
	}

	conversionFailure := base()
	conversionFailure.convertBool = func(*boolVector) ([]bool, error) { return nil, failure }
	if _, err := randomBipartiteGraphWithAdapters(1, 1, BipartiteRandomOptions{}, "test", nil, conversionFailure); !errors.Is(err, failure) {
		t.Fatalf("conversion error = %v", err)
	}
}

func TestBipartiteSimpleEdgeLimitOverflow(t *testing.T) {
	maximum := int(^uint(0) >> 1)
	if _, err := bipartiteSimpleEdgeLimit(maximum, 2, BipartiteRandomOptions{}); err == nil {
		t.Fatal("expected multiplication overflow")
	}
	if err := validateBipartiteModeSizes(maximum, 1); err == nil {
		t.Fatal("expected addition overflow")
	}
}
