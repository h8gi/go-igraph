package igraph

import (
	"testing"
)

func TestSpectralCommunityInternalHelpers(t *testing.T) {
	t.Run("SpincommUpdateRule cValue validation", func(t *testing.T) {
		validRules := []SpincommUpdateRule{SpincommUpdateSimple, SpincommUpdateConfig}
		for _, r := range validRules {
			if _, err := r.cValue(); err != nil {
				t.Errorf("expected valid update rule %d, got err: %v", r, err)
			}
		}
		invalidRule := SpincommUpdateRule(99)
		if _, err := invalidRule.cValue(); err == nil {
			t.Errorf("expected error for invalid update rule %d", invalidRule)
		}
	})

	t.Run("SpinglassImplementation cValue validation", func(t *testing.T) {
		validImpls := []SpinglassImplementation{SpinglassImplementationOriginal, SpinglassImplementationNegative}
		for _, impl := range validImpls {
			if _, err := impl.cValue(); err != nil {
				t.Errorf("expected valid implementation %d, got err: %v", impl, err)
			}
		}
		invalidImpl := SpinglassImplementation(99)
		if _, err := invalidImpl.cValue(); err == nil {
			t.Errorf("expected error for invalid implementation %d", invalidImpl)
		}
	})
}
