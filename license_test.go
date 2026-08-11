package igraph_test

import (
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/google/licensecheck"
)

func TestProjectLicenseIsDetectable(t *testing.T) {
	text, err := os.ReadFile("LICENSE")
	if err != nil {
		t.Fatalf("read LICENSE: %v", err)
	}

	coverage := licensecheck.Scan(text)
	ids := make([]string, 0, len(coverage.Match))
	for _, match := range coverage.Match {
		ids = append(ids, match.ID)
	}
	if !slices.Contains(ids, "GPL-2.0") && !slices.Contains(ids, "GPL-2.0-or-later") {
		t.Fatalf("LICENSE was not detected as GPL-2.0: coverage %.1f%%, matches %v", coverage.Percent, ids)
	}
	if coverage.Percent < 99 {
		t.Fatalf("LICENSE detection covered only %.1f%% of the text", coverage.Percent)
	}

	notice, err := os.ReadFile("COPYRIGHT")
	if err != nil {
		t.Fatalf("read COPYRIGHT: %v", err)
	}
	if !strings.Contains(string(notice), "SPDX-License-Identifier: GPL-2.0-or-later") {
		t.Fatal("COPYRIGHT does not declare GPL-2.0-or-later")
	}
}
