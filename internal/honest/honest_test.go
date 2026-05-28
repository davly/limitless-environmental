package honest

import (
	"bytes"
	"strings"
	"sync"
	"testing"
)

// TestCanonicalAdvisories_Count — exactly 5 canonical advisories.
func TestCanonicalAdvisories_Count(t *testing.T) {
	if got := len(CanonicalAdvisories()); got != 5 {
		t.Fatalf("CanonicalAdvisories count: got %d, want 5", got)
	}
}

// TestCanonicalAdvisories_Codes — all 5 codes present + correct prefix.
func TestCanonicalAdvisories_Codes(t *testing.T) {
	want := map[string]bool{
		"ENVIRONMENTAL_EA_PERMIT_VARIATION_PROCEDURE_REQUIRED": false,
		"ENVIRONMENTAL_EU_IED_BAT_CONCLUSIONS_PIN_REQUIRED":    false,
		"ENVIRONMENTAL_ENV_ACT_2021_BNG_10_PERCENT_REQUIRED":   false,
		"ENVIRONMENTAL_METHODOLOGY_VERSION_PIN_REQUIRED":       false,
		"ENVIRONMENTAL_REVIEWED_BY_COUNSEL_FALSE":              false,
	}
	for _, a := range CanonicalAdvisories() {
		if _, ok := want[a.Code]; !ok {
			t.Errorf("unexpected advisory code: %s", a.Code)
			continue
		}
		want[a.Code] = true
		if !strings.HasPrefix(a.Code, "ENVIRONMENTAL_") {
			t.Errorf("advisory %s missing ENVIRONMENTAL_ prefix", a.Code)
		}
	}
	for code, seen := range want {
		if !seen {
			t.Errorf("missing canonical advisory: %s", code)
		}
	}
}

// TestLoudOnce_FiresExactlyOnce — sync.Once R143 discipline.
func TestLoudOnce_FiresExactlyOnce(t *testing.T) {
	Reset()
	adv, _ := FindAdvisory("ENVIRONMENTAL_EA_PERMIT_VARIATION_PROCEDURE_REQUIRED")
	var buf bytes.Buffer
	LoudOnce(adv, &buf)
	first := buf.String()
	LoudOnce(adv, &buf)
	if buf.String() != first {
		t.Fatalf("LoudOnce fired twice: %q", buf.String())
	}
}

// TestLoudOnce_FormatIncludesLoudOncePrefix — wire format check.
func TestLoudOnce_FormatIncludesLoudOncePrefix(t *testing.T) {
	Reset()
	adv, _ := FindAdvisory("ENVIRONMENTAL_ENV_ACT_2021_BNG_10_PERCENT_REQUIRED")
	var buf bytes.Buffer
	LoudOnce(adv, &buf)
	if !strings.HasPrefix(buf.String(), LoudOncePrefix) {
		t.Fatalf("LoudOnce missing prefix: %q", buf.String())
	}
}

// TestLoudOnce_GoroutineSafe — concurrent callers don't double-fire.
func TestLoudOnce_GoroutineSafe(t *testing.T) {
	Reset()
	adv, _ := FindAdvisory("ENVIRONMENTAL_METHODOLOGY_VERSION_PIN_REQUIRED")
	var mu sync.Mutex
	var buf bytes.Buffer
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var local bytes.Buffer
			LoudOnce(adv, &local)
			mu.Lock()
			buf.Write(local.Bytes())
			mu.Unlock()
		}()
	}
	wg.Wait()
	count := strings.Count(buf.String(), adv.Code)
	if count != 1 {
		t.Fatalf("LoudOnce fired %d times under concurrency; want 1", count)
	}
}

// TestSeverityLadder — 3 advisories Error, 2 Warn.
func TestSeverityLadder(t *testing.T) {
	errors, warns := 0, 0
	for _, a := range CanonicalAdvisories() {
		switch a.Severity {
		case SeverityError:
			errors++
		case SeverityWarn:
			warns++
		}
	}
	if errors != 3 || warns != 2 {
		t.Fatalf("severity ladder drift: errors=%d, warns=%d; want 3,2", errors, warns)
	}
}

// TestFindAdvisory_NotFound — sentinel return.
func TestFindAdvisory_NotFound(t *testing.T) {
	if _, ok := FindAdvisory("DOES_NOT_EXIST"); ok {
		t.Fatal("FindAdvisory returned true for unknown code")
	}
}
