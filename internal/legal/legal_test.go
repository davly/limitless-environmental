package legal

import (
	"strings"
	"testing"
)

// TestLiabilityFooterTemplate_Contents — R166 5-axis pin.
func TestLiabilityFooterTemplate_Contents(t *testing.T) {
	required := []string{
		"limitless-environmental",                  // identity
		"EPR 2016",                                 // scope
		"INFORMATIONAL ONLY",                       // disclaimer
		"regulator",                                // authority
		"Reviewed by qualified counsel: FALSE",     // counsel-review honest-default
		"Environment Act 2021",                     // scope completeness
		"Directive 2010/75/EU",                     // EU IED scope
	}
	for _, s := range required {
		if !strings.Contains(LIABILITY_FOOTER_TEMPLATE, s) {
			t.Errorf("liability footer missing required token: %q", s)
		}
	}
}

// TestLiabilityFooter_WrapsPayload — append discipline.
func TestLiabilityFooter_WrapsPayload(t *testing.T) {
	got := LiabilityFooter("payload body")
	if !strings.HasPrefix(got, "payload body") {
		t.Fatal("LiabilityFooter did not preserve payload")
	}
	if !strings.Contains(got, LIABILITY_FOOTER_TEMPLATE) {
		t.Fatal("LiabilityFooter did not append template")
	}
}

// TestAllRegulatedDecisionClasses_Count — exactly 5 canonical classes.
func TestAllRegulatedDecisionClasses_Count(t *testing.T) {
	if got := len(AllRegulatedDecisionClasses()); got != 5 {
		t.Fatalf("regulated-decision-class count: got %d, want 5", got)
	}
}

// TestIsRegulatedDecision_PositiveCases — all 5 canonical match.
func TestIsRegulatedDecision_PositiveCases(t *testing.T) {
	for _, c := range AllRegulatedDecisionClasses() {
		if !IsRegulatedDecision(c) {
			t.Errorf("canonical class %q rejected by IsRegulatedDecision", c)
		}
	}
}

// TestIsRegulatedDecision_NegativeCase — invalid class.
func TestIsRegulatedDecision_NegativeCase(t *testing.T) {
	if IsRegulatedDecision(RegulatedDecisionClass("not_a_class")) {
		t.Fatal("IsRegulatedDecision accepted unknown class")
	}
}

// TestRegulatedDecisionEscape_SentinelValue — exact sentinel string.
func TestRegulatedDecisionEscape_SentinelValue(t *testing.T) {
	if ENV_REGULATED_DECISION_ESCAPE != "ENV_REGULATED_DECISION_ESCAPE" {
		t.Fatalf("escape sentinel drift: %q", ENV_REGULATED_DECISION_ESCAPE)
	}
}
