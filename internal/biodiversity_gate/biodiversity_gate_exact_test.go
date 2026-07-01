package biodiversity_gate

import (
	"math/big"
	"testing"
)

func rat(t *testing.T, s string) *big.Rat {
	t.Helper()
	r, ok := new(big.Rat).SetString(s)
	if !ok {
		t.Fatalf("bad rat %q", s)
	}
	return r
}

// TestClassifyExact_TenPercentBoundary pins the ENV-1/ENV-4 seam: pre=3.0, post=3.3 is
// EXACTLY +10% BNG (MEETS), but the float path computes 9.99999999999999 < 10 and wrongly
// returns BNG_BELOW_THRESHOLD. The exact path must return MEETS.
func TestClassifyExact_TenPercentBoundary(t *testing.T) {
	// The float path (still) flips here — documents WHY exact matters.
	if out, _, err := Classify(BNGContext{SiteReference: "23/1/FUL", PreDevelopmentUnits: 3.0, PostDevelopmentUnits: 3.3}); err != nil || out != BNGBelowThreshold {
		t.Fatalf("precondition: float Classify should (wrongly) return BELOW at the 3.0->3.3 boundary; got %s err=%v", out, err)
	}
	out, gain, err := ClassifyExact("23/1/FUL", rat(t, "3.0"), rat(t, "3.3"), rat(t, "0"))
	if err != nil {
		t.Fatalf("ClassifyExact: %v", err)
	}
	if out != BNGMeetsThreshold {
		t.Fatalf("exact +10%% boundary must MEET; got %s", out)
	}
	if gain.Cmp(big.NewRat(10, 1)) != 0 {
		t.Fatalf("exact gain must be exactly 10; got %s", gain.FloatString(6))
	}
}

func TestClassifyExact_Table(t *testing.T) {
	cases := []struct {
		pre, post, credits string
		want               BNGOutcome
		note               string
	}{
		{"3.0", "3.3", "0", BNGMeetsThreshold, "exactly +10%"},
		{"3.0", "3.29", "0", BNGBelowThreshold, "just under +10%"},
		{"3.0", "3.31", "0", BNGMeetsThreshold, "just over +10%"},
		{"3.0", "2.9", "0", BNGNetLoss, "post below pre, no credits"},
		{"3.0", "3.0", "0.3", BNGCreditsRequired, "meets only via credits"},
		{"10", "11", "0", BNGMeetsThreshold, "integer +10%"},
	}
	for _, c := range cases {
		out, _, err := ClassifyExact("s", rat(t, c.pre), rat(t, c.post), rat(t, c.credits))
		if err != nil || out != c.want {
			t.Errorf("ClassifyExact(pre=%s post=%s cr=%s)=%s err=%v; want %s (%s)", c.pre, c.post, c.credits, out, err, c.want, c.note)
		}
	}
}

func TestClassifyExact_Guards(t *testing.T) {
	if _, _, err := ClassifyExact("", rat(t, "3"), rat(t, "3.3"), rat(t, "0")); err != ErrEmptySiteReference {
		t.Fatalf("empty site must error")
	}
	if _, _, err := ClassifyExact("s", rat(t, "0"), rat(t, "1"), rat(t, "0")); err != ErrZeroPreUnits {
		t.Fatalf("zero pre must error")
	}
	if _, _, err := ClassifyExact("s", rat(t, "3"), rat(t, "-1"), rat(t, "0")); err != ErrNegativeUnits {
		t.Fatalf("negative must error")
	}
}
