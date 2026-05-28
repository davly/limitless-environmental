// Package legal carries the R166 LIABILITY-FOOTER-CONST + R153 REGULATED-
// ROLE-ESCAPE-INVARIANT discipline for limitless-environmental.
//
// limitless-environmental is the UK Environment Agency + Environment
// Act 2021 + EU IED compliance flagship. The legal package surfaces:
//
//   - LIABILITY_FOOTER_TEMPLATE — R166 5-axis const stating the
//     guidance is INFORMATIONAL only, not a permit-decision and not
//     environmental-law advice.
//   - ENV_REGULATED_DECISION_ESCAPE — R153 sentinel: any decision
//     touching a permit-grant / permit-variation / enforcement-notice /
//     statutory-BNG-sign-off MUST route to a qualified environmental
//     officer (EA / SEPA / NRW / NIEA / consultant ecologist).
package legal

// LIABILITY_FOOTER_TEMPLATE is the R166 5-axis const:
//
//	1. Identity: limitless-environmental I52 marathon 2026-05-28 scaffold
//	2. Scope: UK EPR 2016 + Environment Act 2021 + EU IED 2010/75/EU
//	3. Disclaimer: INFORMATIONAL only, not environmental-law advice
//	4. Authority: regulator-published methodology corpora cited
//	5. Counsel-review: false (Phase-1 scaffold)
const LIABILITY_FOOTER_TEMPLATE = `--- limitless-environmental I52 scaffold 2026-05-28 ---
Scope: UK EPR 2016 (SI 2016/1154) + Environment Act 2021 + EU IED
       (Directive 2010/75/EU) + statutory Biodiversity Net Gain.
This output is INFORMATIONAL ONLY and is not a permit decision,
enforcement notice, or environmental-law advice. Permit and BNG
decisions MUST be taken by a qualified environmental officer at
the relevant regulator (EA / SEPA / NRW / NIEA) or a chartered
environmental consultant. Methodology corpora are cited at
scaffold-time; cold-verify against the regulator-published
canonical artefact before any production decision.
Reviewed by qualified counsel: FALSE (Phase-1 scaffold).
---`

// ENV_REGULATED_DECISION_ESCAPE is the R153 escape-invariant: when a
// scenario matches one of the canonical regulated-decision classes,
// the limitless-environmental verdict MUST route to HUMAN_ESCAPE.
const ENV_REGULATED_DECISION_ESCAPE = "ENV_REGULATED_DECISION_ESCAPE"

// RegulatedDecisionClass enumerates the 5 canonical decision classes
// that trigger ENV_REGULATED_DECISION_ESCAPE under R153.
type RegulatedDecisionClass string

const (
	// PermitGrantOrRefusal — EPR 2016 Regulation 13 + Schedule 5 Part 1:
	// the determination decision itself. NOT an automatable output.
	PermitGrantOrRefusal RegulatedDecisionClass = "permit_grant_or_refusal"

	// PermitVariation — EPR 2016 Regulation 20 + Schedule 5: substantive
	// change to a permitted activity. Requires regulator sign-off.
	PermitVariation RegulatedDecisionClass = "permit_variation"

	// EnforcementNoticeIssuance — EPR 2016 Regulation 36 + Schedule 17:
	// enforcement notice / suspension notice / revocation. Regulator-only.
	EnforcementNoticeIssuance RegulatedDecisionClass = "enforcement_notice_issuance"

	// StatutoryBNGSignOff — Environment Act 2021 Schedule 14 + TCPA 1990:
	// the +10% biodiversity-units gain certification for planning consent.
	// Requires planning authority + (where applicable) Natural England.
	StatutoryBNGSignOff RegulatedDecisionClass = "statutory_bng_sign_off"

	// IEDBATConditionSetting — EU IED Article 14(3): permit conditions
	// set with reference to BAT conclusions. Regulator-only.
	IEDBATConditionSetting RegulatedDecisionClass = "ied_bat_condition_setting"
)

// AllRegulatedDecisionClasses returns the canonical 5-class list.
func AllRegulatedDecisionClasses() []RegulatedDecisionClass {
	return []RegulatedDecisionClass{
		PermitGrantOrRefusal,
		PermitVariation,
		EnforcementNoticeIssuance,
		StatutoryBNGSignOff,
		IEDBATConditionSetting,
	}
}

// LiabilityFooter wraps a payload with the R166 footer.
func LiabilityFooter(payload string) string {
	return payload + "\n\n" + LIABILITY_FOOTER_TEMPLATE
}

// IsRegulatedDecision returns true when the given class is one of the
// canonical R153 escape-invariant classes.
func IsRegulatedDecision(class RegulatedDecisionClass) bool {
	for _, c := range AllRegulatedDecisionClasses() {
		if c == class {
			return true
		}
	}
	return false
}
