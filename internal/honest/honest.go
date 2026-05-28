// Package honest implements R143 LOUD-ONCE-WARNING-FLAG discipline for
// limitless-environmental, with R157 substrate-native idiom (Go's
// `sync.Once`). limitless-environmental is the UK Environment Agency +
// Environment Act 2021 + EU IED compliance flagship — every advisory
// is liability-bearing or methodology-pin drift.
//
// limitless-environmental's 5 honest-defaults surfaces:
//
//  1. ENVIRONMENTAL_EA_PERMIT_VARIATION_PROCEDURE_REQUIRED — EPR 2016
//     Schedule 5 (Environmental Permitting (England and Wales)
//     Regulations 2016, SI 2016/1154) requires formal variation
//     procedure before any substantive change to a permitted activity.
//     Variation-by-omission is non-compliant.
//  2. ENVIRONMENTAL_EU_IED_BAT_CONCLUSIONS_PIN_REQUIRED — EU IED
//     (Directive 2010/75/EU) Article 14(3) requires operating permits
//     to be set with reference to the published BAT conclusions for the
//     relevant industrial sector. BAT conclusions are amended; permit
//     conditions MUST be reviewed against current BAT within 4 years
//     of publication of new BAT conclusions per Article 21(3).
//  3. ENVIRONMENTAL_ENV_ACT_2021_BNG_10_PERCENT_REQUIRED — Environment
//     Act 2021 Schedule 14 (Biodiversity Net Gain) requires a minimum
//     +10% biodiversity-units gain for development-consent-order
//     applications under TCPA 1990 from 2024-02-12 (small sites) /
//     2024-04 (major sites). A BNG calculation below +10% is
//     non-compliant.
//  4. ENVIRONMENTAL_METHODOLOGY_VERSION_PIN_REQUIRED — R150 PARALLEL-
//     MAP jurisdiction-version drift. The local copy of the
//     regulator-published Biodiversity Metric 4.0 / BAT conclusions /
//     EPR methodology corpus MUST be cold-verified against the
//     regulator-published canonical artefact before any permit
//     decision or BNG sign-off.
//  5. ENVIRONMENTAL_REVIEWED_BY_COUNSEL_FALSE — R166 LIABILITY-FOOTER-
//     CONST honest-default. Placeholder legal-disclosure narratives in
//     this scaffold have NOT been reviewed by qualified counsel.
package honest

import (
	"fmt"
	"io"
	"sync"
)

const LoudOncePrefix = "[LOUD-ONCE-WARNING]"

// Severity follows R143.A SEVERITY-LADDER-CONVENTION.
type Severity string

const (
	SeverityInfo  Severity = "INFO"
	SeverityWarn  Severity = "WARN"
	SeverityError Severity = "ERROR"
)

// Advisory is one R143 advisory entry — code + severity + message +
// doc-link pointing at the canonical regulator / R-rule citation.
type Advisory struct {
	Code     string
	Severity Severity
	Message  string
	DocLink  string
}

var canonicalAdvisories = []Advisory{
	{
		Code:     "ENVIRONMENTAL_EA_PERMIT_VARIATION_PROCEDURE_REQUIRED",
		Severity: SeverityError,
		Message:  "EPR 2016 (Environmental Permitting (England and Wales) Regulations 2016, SI 2016/1154) Regulation 20 + Schedule 5 require formal variation procedure (with public-consultation triggers per Schedule 5 Part 1) before any substantive change to a permitted activity (waste / installation / radioactive substances / water-discharge / groundwater-discharge / flood-risk). A material change applied without variation procedure is non-compliant and exposes the operator to enforcement notices under EPR 2016 Regulation 36 + Schedule 17 (Magistrates'/Crown unlimited fine, individual prosecution).",
		DocLink:  "SECURITY.md",
	},
	{
		Code:     "ENVIRONMENTAL_EU_IED_BAT_CONCLUSIONS_PIN_REQUIRED",
		Severity: SeverityError,
		Message:  "EU IED (Directive 2010/75/EU Industrial Emissions Directive) Article 14(3) requires operating permits be set with reference to the published BAT (Best Available Techniques) conclusions for the relevant industrial sector (Annex I activity). Per Article 21(3), permit conditions MUST be reconsidered and (if necessary) updated within 4 years of publication of new BAT conclusions. The local BAT conclusions methodology corpus pin MUST be cold-verified against the regulator-published canonical artefact (EU Commission Implementing Decision) before any permit-condition decision.",
		DocLink:  "SECURITY.md",
	},
	{
		Code:     "ENVIRONMENTAL_ENV_ACT_2021_BNG_10_PERCENT_REQUIRED",
		Severity: SeverityError,
		Message:  "Environment Act 2021 Schedule 14 (Biodiversity Net Gain) introduces a mandatory minimum +10% biodiversity-units gain for development-consent-order applications under TCPA 1990 (Town and Country Planning Act 1990) — small sites from 2024-02-12, major sites from 2024-04-02. The +10% threshold is measured via the statutory Biodiversity Metric (currently v4.0 published by Natural England / DEFRA). A BNG calculation that returns < +10% biodiversity-units gain is non-compliant; the consent application MUST either redesign the development OR purchase statutory biodiversity credits before the planning authority will determine.",
		DocLink:  "SECURITY.md",
	},
	{
		Code:     "ENVIRONMENTAL_METHODOLOGY_VERSION_PIN_REQUIRED",
		Severity: SeverityWarn,
		Message:  "R150 PARALLEL-MAP jurisdiction-version drift. Three regulator-published methodology corpora are pinned at scaffold-time (DEFRA + Natural England Biodiversity Metric v4.0 / EU Commission BAT conclusions implementing decisions / EA Environmental Permitting guidance). Regulators publish amendments. The local corpus SHA pin MUST be cold-verified against the regulator-published canonical artefact before any permit decision or BNG sign-off. A stale local pin is a methodology-drift compliance violation.",
		DocLink:  "CONTEXT.md",
	},
	{
		Code:     "ENVIRONMENTAL_REVIEWED_BY_COUNSEL_FALSE",
		Severity: SeverityWarn,
		Message:  "R166 LIABILITY-FOOTER-CONST honest-default. Phase-1 scaffold ships ReviewedByCounsel = false. Placeholder legal-disclosure narrative templates (EPR Schedule 5 variation procedure citations, IED BAT-conclusions cross-walks, Environment Act 2021 BNG calculation narratives) have NOT been reviewed by qualified environmental-law counsel. Operator MUST commission counsel review + flip ReviewedByCounsel to true on its own R145.B sibling branch before any live permit-decision deployment.",
		DocLink:  "SECURITY.md",
	},
}

// registry is package-global — single LoudOnce gate per advisory code.
var (
	registryMu sync.RWMutex
	registry   = map[string]*sync.Once{}
)

// LoudOnce emits the advisory exactly once per package process-lifetime.
// Goroutine-safe.
func LoudOnce(adv Advisory, w io.Writer) {
	registryMu.RLock()
	once, ok := registry[adv.Code]
	registryMu.RUnlock()
	if !ok {
		registryMu.Lock()
		once, ok = registry[adv.Code]
		if !ok {
			once = &sync.Once{}
			registry[adv.Code] = once
		}
		registryMu.Unlock()
	}
	once.Do(func() {
		_, _ = fmt.Fprintf(w, "%s %s %s: %s (see %s)\n",
			LoudOncePrefix, adv.Severity, adv.Code, adv.Message, adv.DocLink)
	})
}

// Reset clears the once-gate registry. Test-only.
func Reset() {
	registryMu.Lock()
	registry = map[string]*sync.Once{}
	registryMu.Unlock()
}

// CanonicalAdvisories returns a defensive copy of the 5 canonical advisories.
func CanonicalAdvisories() []Advisory {
	out := make([]Advisory, len(canonicalAdvisories))
	copy(out, canonicalAdvisories)
	return out
}

// FindAdvisory looks up a canonical advisory by Code.
func FindAdvisory(code string) (Advisory, bool) {
	for _, a := range canonicalAdvisories {
		if a.Code == code {
			return a, true
		}
	}
	return Advisory{}, false
}
