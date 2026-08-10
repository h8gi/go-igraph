package igraph

import (
	"testing"
)

func TestConstructorsInvalidModes(t *testing.T) {
	if _, err := StarMode(99).cValue(); err == nil {
		t.Error("expected error for invalid StarMode")
	}
	if _, err := TreeMode(99).cValue(); err == nil {
		t.Error("expected error for invalid TreeMode")
	}
}

func TestValidateConstructorSizeNegative(t *testing.T) {
	if err := validateConstructorSize("test size", -1); err == nil {
		t.Error("expected error for negative size in validateConstructorSize")
	}
}
