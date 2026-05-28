package permit_gate

import (
	"testing"
	"time"
)

// TestAllOutcomes_Count — exactly 5 canonical outcomes.
func TestAllOutcomes_Count(t *testing.T) {
	if got := len(AllOutcomes()); got != 5 {
		t.Fatalf("AllOutcomes count: got %d, want 5", got)
	}
}

// TestIsRegulatoryEscape_OnlyFreshDoesNotEscape — R153 boundary.
func TestIsRegulatoryEscape_OnlyFreshDoesNotEscape(t *testing.T) {
	for _, o := range AllOutcomes() {
		escapes := o.IsRegulatoryEscape()
		if o == EAPermitFresh && escapes {
			t.Errorf("EA_PERMIT_FRESH should NOT escape")
		}
		if o != EAPermitFresh && !escapes {
			t.Errorf("%s should escape (R153)", o)
		}
	}
}

// TestClassify_FreshPermit — happy path: fresh permit, no issues.
func TestClassify_FreshPermit(t *testing.T) {
	now := time.Date(2026, 5, 28, 0, 0, 0, 0, time.UTC)
	ctx := PermitContext{
		PermitReference: "EPR/AB1234XY",
		IssuanceDate:    now.AddDate(-1, 0, 0),
	}
	out, err := Classify(ctx, now)
	if err != nil {
		t.Fatalf("Classify err: %v", err)
	}
	if out != EAPermitFresh {
		t.Fatalf("got %s, want EA_PERMIT_FRESH", out)
	}
}

// TestClassify_EnforcementOverridesAll — precedence 1.
func TestClassify_EnforcementOverridesAll(t *testing.T) {
	now := time.Date(2026, 5, 28, 0, 0, 0, 0, time.UTC)
	ctx := PermitContext{
		PermitReference:       "EPR/AB1234XY",
		IssuanceDate:          now.AddDate(-10, 0, 0),
		EnforcementNoticeOpen: true,
		VariationInFlight:     true, // would otherwise yield VARIATION_PENDING
	}
	out, _ := Classify(ctx, now)
	if out != NoncomplianceNoticeOpen {
		t.Fatalf("enforcement should override: got %s", out)
	}
}

// TestClassify_BATDrift — IED Article 21(3) trigger.
// Scenario: permit issued before BAT publication; BAT now > 4 yrs old;
// no post-BAT variation has reset the review clock.
func TestClassify_BATDrift(t *testing.T) {
	now := time.Date(2026, 5, 28, 0, 0, 0, 0, time.UTC)
	ctx := PermitContext{
		PermitReference:               "EPR/AB1234XY",
		IssuanceDate:                  now.AddDate(-7, 0, 0), // 2019
		BATConclusionsPublicationDate: now.AddDate(-5, 0, 0), // 2021, > 4yr
		Schedule5ReviewIntervalYears:  10, // suppress Schedule 5 trigger
	}
	out, _ := Classify(ctx, now)
	if out != BATConclusionsDrift {
		t.Fatalf("BAT-drift not detected: got %s", out)
	}
}

// TestClassify_Schedule5ReviewDue — EPR review-interval reached.
func TestClassify_Schedule5ReviewDue(t *testing.T) {
	now := time.Date(2026, 5, 28, 0, 0, 0, 0, time.UTC)
	ctx := PermitContext{
		PermitReference: "EPR/AB1234XY",
		IssuanceDate:    now.AddDate(-5, 0, 0), // > 4-year default
	}
	out, _ := Classify(ctx, now)
	if out != Schedule5ReviewDue {
		t.Fatalf("Schedule 5 review-due not detected: got %s", out)
	}
}

// TestClassify_VariationPending — Schedule 5 procedure in flight.
func TestClassify_VariationPending(t *testing.T) {
	now := time.Date(2026, 5, 28, 0, 0, 0, 0, time.UTC)
	ctx := PermitContext{
		PermitReference:   "EPR/AB1234XY",
		IssuanceDate:      now.AddDate(-1, 0, 0),
		VariationInFlight: true,
	}
	out, _ := Classify(ctx, now)
	if out != VariationPending {
		t.Fatalf("VariationPending not detected: got %s", out)
	}
}

// TestClassify_LastVariationResetsClock — variation resets review clock.
func TestClassify_LastVariationResetsClock(t *testing.T) {
	now := time.Date(2026, 5, 28, 0, 0, 0, 0, time.UTC)
	ctx := PermitContext{
		PermitReference:   "EPR/AB1234XY",
		IssuanceDate:      now.AddDate(-10, 0, 0),  // would trigger Schedule 5
		LastVariationDate: now.AddDate(-1, 0, 0),   // but variation reset
	}
	out, _ := Classify(ctx, now)
	if out != EAPermitFresh {
		t.Fatalf("LastVariation should reset clock: got %s", out)
	}
}

// TestClassify_EmptyPermitRef_Error — input validation.
func TestClassify_EmptyPermitRef_Error(t *testing.T) {
	_, err := Classify(PermitContext{}, time.Now())
	if err != ErrEmptyPermitReference {
		t.Fatalf("err: got %v, want ErrEmptyPermitReference", err)
	}
}

// TestClassify_ZeroIssuance_Error — IssuanceDate required.
func TestClassify_ZeroIssuance_Error(t *testing.T) {
	_, err := Classify(PermitContext{PermitReference: "X"}, time.Now())
	if err != ErrInvalidIssuanceDate {
		t.Fatalf("err: got %v, want ErrInvalidIssuanceDate", err)
	}
}

// TestClassify_CustomReviewInterval — operator can override.
func TestClassify_CustomReviewInterval(t *testing.T) {
	now := time.Date(2026, 5, 28, 0, 0, 0, 0, time.UTC)
	ctx := PermitContext{
		PermitReference:              "EPR/AB1234XY",
		IssuanceDate:                 now.AddDate(-3, 0, 0),
		Schedule5ReviewIntervalYears: 2, // tighter than default 4
	}
	out, _ := Classify(ctx, now)
	if out != Schedule5ReviewDue {
		t.Fatalf("custom 2-yr interval missed: got %s", out)
	}
}

// TestClassify_BATDriftNeedsLastReviewBeforeBAT — invariant test.
func TestClassify_BATDriftRequiresPreBATReview(t *testing.T) {
	now := time.Date(2026, 5, 28, 0, 0, 0, 0, time.UTC)
	// BAT published 5y ago, but operator already varied last year.
	ctx := PermitContext{
		PermitReference:               "EPR/AB1234XY",
		IssuanceDate:                  now.AddDate(-7, 0, 0),
		LastVariationDate:             now.AddDate(-1, 0, 0),
		BATConclusionsPublicationDate: now.AddDate(-5, 0, 0),
	}
	out, _ := Classify(ctx, now)
	if out != EAPermitFresh {
		t.Fatalf("BAT drift should be cleared by post-BAT variation: got %s", out)
	}
}

// TestPermitOutcome_ExactStringValues — wire format pin.
func TestPermitOutcome_ExactStringValues(t *testing.T) {
	if string(EAPermitFresh) != "EA_PERMIT_FRESH" ||
		string(VariationPending) != "VARIATION_PENDING" ||
		string(BATConclusionsDrift) != "BAT_CONCLUSIONS_DRIFT" ||
		string(Schedule5ReviewDue) != "SCHEDULE_5_REVIEW_DUE" ||
		string(NoncomplianceNoticeOpen) != "NONCOMPLIANCE_NOTICE_OPEN" {
		t.Fatal("PermitOutcome string-value drift")
	}
}
