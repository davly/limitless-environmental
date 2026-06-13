// Package firewall implements the R145.C FIREWALL-TEST-DISCIPLINE pin
// for limitless-environmental — structural firewall against
// internal/ + cmd/ drift.
//
// The firewall test fails CI if the on-disk package list drifts from
// the scaffold-pinned canonical list. This is the R145.B sibling-not-
// stacked discipline made mechanical: a new internal/<pkg>/ must be
// added to ExpectedPackages on its own branch, with a paired
// regression test, or the firewall fails.
package firewall

import (
	"os"
	"path/filepath"
	"sort"
)

// ExpectedPackages returns the canonical list of internal/ packages
// limitless-environmental ships as of 2026-05-28 I52 scaffold +
// calibration-uncertainty x-poll (2026-06-13).
//
// 8 packages: 5 cohort (firewall + honest + legal + manifest +
// mirrormark) + 3 domain (permit_gate + biodiversity_gate +
// calibration). calibration added: Brier/Murphy/ROC probabilistic
// forecast verification for the BNG gain estimate.
func ExpectedPackages() []string {
	return []string{
		"biodiversity_gate",
		"calibration",
		"firewall",
		"honest",
		"legal",
		"manifest",
		"mirrormark",
		"permit_gate",
	}
}

// ExpectedBinaries returns the canonical list of cmd/ binaries.
func ExpectedBinaries() []string {
	return []string{
		"environmental",
	}
}

// ScanInternal returns the subdirectories under repoRoot/internal/.
func ScanInternal(repoRoot string) ([]string, error) {
	return scanGoSubtree(filepath.Join(repoRoot, "internal"))
}

// ScanCmd returns the subdirectories under repoRoot/cmd/.
func ScanCmd(repoRoot string) ([]string, error) {
	cmdDir := filepath.Join(repoRoot, "cmd")
	entries, err := os.ReadDir(cmdDir)
	if err != nil {
		return nil, err
	}
	var out []string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		mainGo := filepath.Join(cmdDir, e.Name(), "main.go")
		if _, err := os.Stat(mainGo); err == nil {
			out = append(out, e.Name())
		}
	}
	sort.Strings(out)
	return out, nil
}

func scanGoSubtree(root string) ([]string, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, err
	}
	var out []string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name := e.Name()
		subPath := filepath.Join(root, name)
		hasGo, err := dirHasGoFile(subPath)
		if err != nil {
			return nil, err
		}
		if hasGo {
			out = append(out, name)
		}
	}
	sort.Strings(out)
	return out, nil
}

func dirHasGoFile(dir string) (bool, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false, err
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if filepath.Ext(e.Name()) == ".go" {
			return true, nil
		}
	}
	return false, nil
}
