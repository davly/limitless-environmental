package firewall

import (
	"path/filepath"
	"reflect"
	"runtime"
	"testing"
)

// TestExpectedPackages_Count — exactly 7 canonical packages.
func TestExpectedPackages_Count(t *testing.T) {
	if got := len(ExpectedPackages()); got != 7 {
		t.Fatalf("ExpectedPackages count: got %d, want 7", got)
	}
}

// TestExpectedPackages_Sorted — alphabetical order pinned.
func TestExpectedPackages_Sorted(t *testing.T) {
	pkgs := ExpectedPackages()
	for i := 1; i < len(pkgs); i++ {
		if pkgs[i-1] >= pkgs[i] {
			t.Fatalf("ExpectedPackages not strictly sorted at %d: %q vs %q", i, pkgs[i-1], pkgs[i])
		}
	}
}

// TestExpectedBinaries_Single — only the environmental CLI.
func TestExpectedBinaries_Single(t *testing.T) {
	bins := ExpectedBinaries()
	if !reflect.DeepEqual(bins, []string{"environmental"}) {
		t.Fatalf("ExpectedBinaries drift: %v", bins)
	}
}

// TestFirewall_OnDiskMatchesExpected — structural pin: on-disk == expected.
func TestFirewall_OnDiskMatchesExpected(t *testing.T) {
	repoRoot := findRepoRoot(t)
	got, err := ScanInternal(repoRoot)
	if err != nil {
		t.Fatalf("ScanInternal: %v", err)
	}
	if !reflect.DeepEqual(got, ExpectedPackages()) {
		t.Fatalf("internal/ on-disk drift:\n  got:  %v\n  want: %v", got, ExpectedPackages())
	}
}

// TestFirewall_CmdMatchesExpected — cmd/ on-disk == expected.
func TestFirewall_CmdMatchesExpected(t *testing.T) {
	repoRoot := findRepoRoot(t)
	got, err := ScanCmd(repoRoot)
	if err != nil {
		t.Fatalf("ScanCmd: %v", err)
	}
	if !reflect.DeepEqual(got, ExpectedBinaries()) {
		t.Fatalf("cmd/ on-disk drift:\n  got:  %v\n  want: %v", got, ExpectedBinaries())
	}
}

// findRepoRoot returns the repo root from this test file's path.
func findRepoRoot(t *testing.T) string {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	// .../limitless-environmental/internal/firewall/firewall_test.go
	return filepath.Dir(filepath.Dir(filepath.Dir(thisFile)))
}
