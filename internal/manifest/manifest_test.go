package manifest

import (
	"strings"
	"testing"
	"time"
)

// TestSeed_EntryCount — 11 canonical entries.
func TestSeed_EntryCount(t *testing.T) {
	if got := len(Seed()); got != 11 {
		t.Fatalf("Seed entry count: got %d, want 11", got)
	}
}

// TestSeed_AllEntriesPinned — schema version pinned.
func TestSeed_AllEntriesPinned(t *testing.T) {
	for _, e := range Seed() {
		if e.SchemaVersion != SchemaVersion {
			t.Errorf("entry %q has wrong SchemaVersion: got %d, want %d", e.Key, e.SchemaVersion, SchemaVersion)
		}
	}
}

// TestSeed_KAT1HexAnchored — canonical hex pinned at R151 anchor.
func TestSeed_KAT1HexAnchored(t *testing.T) {
	for _, e := range Seed() {
		if e.Key == "cohort.r151.kat1_canonical_hex" {
			if !strings.Contains(e.Description, "239a7d0d3f1bbe3a98aede01e2ad818c2db60b7177c02e2f015035b2b5b7dbca") {
				t.Fatalf("KAT-1 anchor entry missing hex literal: %q", e.Description)
			}
			return
		}
	}
	t.Fatal("manifest missing cohort.r151.kat1_canonical_hex anchor entry")
}

// TestSeed_MethodologyPinsAreUnknownDate — Phase-2 sentinel.
func TestSeed_MethodologyPinsAreUnknownDate(t *testing.T) {
	for _, e := range Seed() {
		if !strings.HasPrefix(e.Key, "methodology_corpus.") {
			continue
		}
		if !e.FreshAt.Equal(FreshAtUnknown) {
			t.Errorf("methodology pin %q should use FreshAtUnknown sentinel; got %v", e.Key, e.FreshAt)
		}
		if e.Confidence != ConfidenceLow {
			t.Errorf("methodology pin %q should have ConfidenceLow; got %d", e.Key, e.Confidence)
		}
	}
}

// TestSeed_JurisdictionDistribution — UK + EU + None all represented.
func TestSeed_JurisdictionDistribution(t *testing.T) {
	seen := map[Jurisdiction]int{}
	for _, e := range Seed() {
		seen[e.Jurisdiction]++
	}
	if seen[JurisdictionUK] < 3 {
		t.Errorf("UK jurisdiction count: got %d, want >= 3", seen[JurisdictionUK])
	}
	if seen[JurisdictionEU] < 2 {
		t.Errorf("EU jurisdiction count: got %d, want >= 2", seen[JurisdictionEU])
	}
}

// TestIsStale_UnknownDate — sentinel returns true.
func TestIsStale_UnknownDate(t *testing.T) {
	e := Entry{FreshAt: FreshAtUnknown}
	if !e.IsStale(time.Now(), 24*time.Hour) {
		t.Fatal("FreshAtUnknown entry should be stale")
	}
}

// TestIsStale_BeyondMaxAge — exceeded threshold returns true.
func TestIsStale_BeyondMaxAge(t *testing.T) {
	e := Entry{FreshAt: time.Now().Add(-48 * time.Hour)}
	if !e.IsStale(time.Now(), 24*time.Hour) {
		t.Fatal("48h-old entry should be stale with 24h maxAge")
	}
}

// TestIsStale_WithinMaxAge — fresh returns false.
func TestIsStale_WithinMaxAge(t *testing.T) {
	e := Entry{FreshAt: time.Now().Add(-1 * time.Hour)}
	if e.IsStale(time.Now(), 24*time.Hour) {
		t.Fatal("1h-old entry should NOT be stale with 24h maxAge")
	}
}

// TestManifest_SortedKeys — keys in alphabetical order.
func TestManifest_SortedKeys(t *testing.T) {
	keys := Seed().SortedKeys()
	for i := 1; i < len(keys); i++ {
		if keys[i-1] > keys[i] {
			t.Fatalf("SortedKeys not sorted: %q > %q", keys[i-1], keys[i])
		}
	}
}

// TestManifest_ByJurisdiction — UK filter returns only UK.
func TestManifest_ByJurisdiction(t *testing.T) {
	uk := Seed().ByJurisdiction(JurisdictionUK)
	if len(uk) == 0 {
		t.Fatal("no UK entries returned")
	}
	for _, e := range uk {
		if e.Jurisdiction != JurisdictionUK {
			t.Errorf("ByJurisdiction(UK) returned non-UK entry: %q", e.Key)
		}
	}
}
