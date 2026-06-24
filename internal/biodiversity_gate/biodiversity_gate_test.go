package biodiversity_gate

import (
	"math"
	"testing"
)

// TestStatutoryMinimum_Is10 — statutory threshold pinned.
func TestStatutoryMinimum_Is10(t *testing.T) {
	if StatutoryMinimumGainPercent != 10.0 {
		t.Fatalf("statutory minimum drift: got %v, want 10.0", StatutoryMinimumGainPercent)
	}
}

// TestAllOutcomes_Count — exactly 4 canonical outcomes.
func TestAllOutcomes_Count(t *testing.T) {
	if got := len(AllOutcomes()); got != 4 {
		t.Fatalf("AllOutcomes count: got %d, want 4", got)
	}
}

// TestIsRegulatoryEscape_OnlyMeetsThresholdDoesNotEscape — R153 boundary.
func TestIsRegulatoryEscape_OnlyMeetsThresholdDoesNotEscape(t *testing.T) {
	for _, o := range AllOutcomes() {
		escapes := o.IsRegulatoryEscape()
		if o == BNGMeetsThreshold && escapes {
			t.Errorf("BNG_MEETS_THRESHOLD should NOT escape")
		}
		if o != BNGMeetsThreshold && !escapes {
			t.Errorf("%s should escape", o)
		}
	}
}

// TestClassify_MeetsThreshold — exactly +10%.
func TestClassify_MeetsThreshold(t *testing.T) {
	ctx := BNGContext{
		SiteReference:        "23/04567/FUL",
		PreDevelopmentUnits:  100.0,
		PostDevelopmentUnits: 110.0,
	}
	out, gain, err := Classify(ctx)
	if err != nil {
		t.Fatalf("Classify: %v", err)
	}
	if out != BNGMeetsThreshold {
		t.Fatalf("got %s, want BNG_MEETS_THRESHOLD", out)
	}
	if math.Abs(gain-10.0) > 0.0001 {
		t.Fatalf("gain%%: got %v, want 10.0", gain)
	}
}

// TestClassify_AboveThreshold — exceeds +10%.
func TestClassify_AboveThreshold(t *testing.T) {
	ctx := BNGContext{
		SiteReference:        "23/04567/FUL",
		PreDevelopmentUnits:  100.0,
		PostDevelopmentUnits: 125.5,
	}
	out, _, _ := Classify(ctx)
	if out != BNGMeetsThreshold {
		t.Fatalf("got %s, want BNG_MEETS_THRESHOLD", out)
	}
}

// TestClassify_BelowThreshold — +5%, no credits.
func TestClassify_BelowThreshold(t *testing.T) {
	ctx := BNGContext{
		SiteReference:        "23/04567/FUL",
		PreDevelopmentUnits:  100.0,
		PostDevelopmentUnits: 105.0,
	}
	out, _, _ := Classify(ctx)
	if out != BNGBelowThreshold {
		t.Fatalf("got %s, want BNG_BELOW_THRESHOLD", out)
	}
}

// TestClassify_CreditsBridgeGap — credits bring under-spec to over.
func TestClassify_CreditsBridgeGap(t *testing.T) {
	ctx := BNGContext{
		SiteReference:             "23/04567/FUL",
		PreDevelopmentUnits:       100.0,
		PostDevelopmentUnits:      105.0, // only +5%
		StatutoryCreditsPurchased: 10.0,  // +15% total with credits
	}
	out, _, _ := Classify(ctx)
	if out != BNGCreditsRequired {
		t.Fatalf("got %s, want BNG_CREDITS_REQUIRED", out)
	}
}

// TestClassify_NetLoss — post < pre, no credits.
func TestClassify_NetLoss(t *testing.T) {
	ctx := BNGContext{
		SiteReference:        "23/04567/FUL",
		PreDevelopmentUnits:  100.0,
		PostDevelopmentUnits: 90.0,
	}
	out, gain, _ := Classify(ctx)
	if out != BNGNetLoss {
		t.Fatalf("got %s, want BNG_NET_LOSS", out)
	}
	if gain >= 0 {
		t.Fatalf("net-loss gain%% should be negative: got %v", gain)
	}
}

// TestClassify_NetLossWithCredits_NotNetLoss — credits change outcome.
func TestClassify_NetLossWithCredits_NotNetLoss(t *testing.T) {
	ctx := BNGContext{
		SiteReference:             "23/04567/FUL",
		PreDevelopmentUnits:       100.0,
		PostDevelopmentUnits:      90.0,
		StatutoryCreditsPurchased: 30.0,
	}
	out, _, _ := Classify(ctx)
	if out == BNGNetLoss {
		t.Fatal("credits should prevent NET_LOSS classification")
	}
}

// TestClassify_EmptyReference_Error — input validation.
func TestClassify_EmptyReference_Error(t *testing.T) {
	_, _, err := Classify(BNGContext{})
	if err != ErrEmptySiteReference {
		t.Fatalf("err: got %v, want ErrEmptySiteReference", err)
	}
}

// TestClassify_ZeroPreUnits_Error — divide-by-zero guard.
func TestClassify_ZeroPreUnits_Error(t *testing.T) {
	_, _, err := Classify(BNGContext{SiteReference: "X"})
	if err != ErrZeroPreUnits {
		t.Fatalf("err: got %v, want ErrZeroPreUnits", err)
	}
}

// TestClassify_NegativeUnits_Error — defensive guard.
func TestClassify_NegativeUnits_Error(t *testing.T) {
	_, _, err := Classify(BNGContext{
		SiteReference:        "X",
		PreDevelopmentUnits:  100.0,
		PostDevelopmentUnits: -5.0,
	})
	if err != ErrNegativeUnits {
		t.Fatalf("err: got %v, want ErrNegativeUnits", err)
	}
}

// TestClassify_NonFiniteUnits_Error — NaN/+Inf/-Inf on any biodiversity-
// units field must fail closed with ErrNonFiniteUnits, never silently
// returning the only non-escaping outcome (BNG_MEETS_THRESHOLD).
//
// Discrimination: each case fails (Classify returns BNGMeetsThreshold with
// nil error) if the finite-check guard in Classify is reverted — NaN/+Inf
// comparisons against the existing `< 0`/`== 0`/threshold guards are all
// false, so control would fall through to BNGMeetsThreshold.
func TestClassify_NonFiniteUnits_Error(t *testing.T) {
	nan := math.NaN()
	posInf := math.Inf(1)
	negInf := math.Inf(-1)

	cases := []struct {
		name string
		ctx  BNGContext
	}{
		{"pre=NaN", BNGContext{SiteReference: "23/04567/FUL", PreDevelopmentUnits: nan, PostDevelopmentUnits: 100.0}},
		{"pre=+Inf", BNGContext{SiteReference: "23/04567/FUL", PreDevelopmentUnits: posInf, PostDevelopmentUnits: 100.0}},
		{"pre=-Inf", BNGContext{SiteReference: "23/04567/FUL", PreDevelopmentUnits: negInf, PostDevelopmentUnits: 100.0}},
		{"post=NaN", BNGContext{SiteReference: "23/04567/FUL", PreDevelopmentUnits: 100.0, PostDevelopmentUnits: nan}},
		{"post=+Inf", BNGContext{SiteReference: "23/04567/FUL", PreDevelopmentUnits: 100.0, PostDevelopmentUnits: posInf}},
		{"post=-Inf", BNGContext{SiteReference: "23/04567/FUL", PreDevelopmentUnits: 100.0, PostDevelopmentUnits: negInf}},
		{"credits=NaN", BNGContext{SiteReference: "23/04567/FUL", PreDevelopmentUnits: 100.0, PostDevelopmentUnits: 110.0, StatutoryCreditsPurchased: nan}},
		{"credits=+Inf", BNGContext{SiteReference: "23/04567/FUL", PreDevelopmentUnits: 100.0, PostDevelopmentUnits: 110.0, StatutoryCreditsPurchased: posInf}},
		{"credits=-Inf", BNGContext{SiteReference: "23/04567/FUL", PreDevelopmentUnits: 100.0, PostDevelopmentUnits: 110.0, StatutoryCreditsPurchased: negInf}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out, gain, err := Classify(tc.ctx)
			if err != ErrNonFiniteUnits {
				t.Fatalf("err: got %v (out=%s, gain=%v), want ErrNonFiniteUnits", err, out, gain)
			}
			if out == BNGMeetsThreshold {
				t.Fatalf("non-finite input must NOT classify as the non-escaping BNG_MEETS_THRESHOLD")
			}
		})
	}
}

// TestBNGOutcome_ExactStringValues — wire format pin.
func TestBNGOutcome_ExactStringValues(t *testing.T) {
	if string(BNGMeetsThreshold) != "BNG_MEETS_THRESHOLD" ||
		string(BNGBelowThreshold) != "BNG_BELOW_THRESHOLD" ||
		string(BNGCreditsRequired) != "BNG_CREDITS_REQUIRED" ||
		string(BNGNetLoss) != "BNG_NET_LOSS" {
		t.Fatal("BNGOutcome string-value drift")
	}
}
