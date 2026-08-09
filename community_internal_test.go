package igraph

import (
	"testing"
)

func TestInternalCommunityHelpers(t *testing.T) {
	t.Run("withRNG with seed and nil seed", func(t *testing.T) {
		seedVal := uint64(12345)
		err := withRNG(&seedVal, func() error {
			return nil
		})
		if err != nil {
			t.Fatalf("unexpected error with seed: %v", err)
		}

		err = withRNG(nil, func() error {
			return nil
		})
		if err != nil {
			t.Fatalf("unexpected error with nil seed: %v", err)
		}
	})

	t.Run("cMatrixIntToMerges nil and empty", func(t *testing.T) {
		res := cMatrixIntToMerges(nil)
		if len(res) != 0 {
			t.Errorf("expected empty merges for nil matrix")
		}

		m, cleanup, err := mergesToCMatrixInt([][2]int{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		defer cleanup()

		resEmpty := cMatrixIntToMerges(&m)
		if len(resEmpty) != 0 {
			t.Errorf("expected empty merges for empty matrix")
		}

		merges := [][2]int{{0, 1}, {2, 3}}
		m2, cleanup2, err := mergesToCMatrixInt(merges)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		defer cleanup2()

		res2 := cMatrixIntToMerges(&m2)
		if len(res2) != 2 || res2[0] != merges[0] || res2[1] != merges[1] {
			t.Errorf("expected %v, got %v", merges, res2)
		}
	})
}
