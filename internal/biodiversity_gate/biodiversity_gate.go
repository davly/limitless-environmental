// Package biodiversity_gate implements the Environment Act 2021
// Schedule 14 Biodiversity Net Gain (BNG) +10% minimum-gain gate for
// limitless-environmental.
//
// The statutory minimum under Environment Act 2021 Schedule 14 is
// +10% biodiversity-units gain over the pre-development baseline,
// measured via the DEFRA + Natural England Biodiversity Metric v4.0.
//
// This package consumes a BNGContext (pre + post biodiversity units,
// statutory-credits-purchased units) and emits a BNGOutcome reflecting
// whether the +10% threshold is met.
//
// IMPORTANT: this is INFORMATIONAL only — the statutory BNG sign-off
// is a regulated decision (R153) routed to the local planning
// authority + Natural England. The package always pairs with the
// R153 escape-invariant in internal/legal/.
package biodiversity_gate

import (
	"errors"
)

// StatutoryMinimumGainPercent is the Environment Act 2021 Schedule 14
// statutory minimum (+10%).
const StatutoryMinimumGainPercent = 10.0

// BNGOutcome is the R115 single-enum gate result.
type BNGOutcome string

const (
	// BNGMeetsThreshold — calculated gain >= +10%.
	BNGMeetsThreshold BNGOutcome = "BNG_MEETS_THRESHOLD"
	// BNGBelowThreshold — gain < +10%, redesign or credits required.
	BNGBelowThreshold BNGOutcome = "BNG_BELOW_THRESHOLD"
	// BNGCreditsRequired — base development below threshold but
	// statutory credits make up the difference (still requires
	// planning-authority sign-off).
	BNGCreditsRequired BNGOutcome = "BNG_CREDITS_REQUIRED"
	// BNGNetLoss — post < pre; the development causes net biodiversity
	// loss before any credits.
	BNGNetLoss BNGOutcome = "BNG_NET_LOSS"
)

// AllOutcomes returns the canonical 4-outcome list.
func AllOutcomes() []BNGOutcome {
	return []BNGOutcome{
		BNGMeetsThreshold,
		BNGBelowThreshold,
		BNGCreditsRequired,
		BNGNetLoss,
	}
}

// IsRegulatoryEscape reports whether the outcome requires planning-
// authority routing per R153. Everything except BNG_MEETS_THRESHOLD
// escapes (and even MEETS still requires statutory sign-off, but does
// not need a redesign-conversation).
func (o BNGOutcome) IsRegulatoryEscape() bool {
	return o != BNGMeetsThreshold
}

// BNGContext is the canonical input.
type BNGContext struct {
	// SiteReference — local-planning-authority reference (e.g. "23/04567/FUL").
	SiteReference string
	// PreDevelopmentUnits — Biodiversity Metric v4.0 pre-development
	// baseline biodiversity-units total (sum across habitat /
	// hedgerow / watercourse modules).
	PreDevelopmentUnits float64
	// PostDevelopmentUnits — Biodiversity Metric v4.0 post-development
	// total (on-site retained + created + enhanced).
	PostDevelopmentUnits float64
	// StatutoryCreditsPurchased — statutory biodiversity-credits units
	// purchased under DEFRA's statutory credits scheme (Environment
	// Act 2021 Schedule 14 paragraph 4).
	StatutoryCreditsPurchased float64
}

// ErrEmptySiteReference — SiteReference required.
var ErrEmptySiteReference = errors.New("biodiversity_gate: SiteReference is required")

// ErrZeroPreUnits — PreDevelopmentUnits must be > 0.
var ErrZeroPreUnits = errors.New("biodiversity_gate: PreDevelopmentUnits must be > 0")

// ErrNegativeUnits — no field may be negative.
var ErrNegativeUnits = errors.New("biodiversity_gate: biodiversity-units fields must be >= 0")

// Classify returns the canonical BNGOutcome.
//
//   - post < pre AND no credits: BNG_NET_LOSS
//   - (post + credits) >= pre * 1.10: meets threshold; if credits>0
//     and post alone < threshold, returns BNG_CREDITS_REQUIRED
//   - otherwise: BNG_BELOW_THRESHOLD
func Classify(ctx BNGContext) (BNGOutcome, float64, error) {
	if ctx.SiteReference == "" {
		return "", 0, ErrEmptySiteReference
	}
	if ctx.PreDevelopmentUnits < 0 || ctx.PostDevelopmentUnits < 0 || ctx.StatutoryCreditsPurchased < 0 {
		return "", 0, ErrNegativeUnits
	}
	if ctx.PreDevelopmentUnits == 0 {
		return "", 0, ErrZeroPreUnits
	}

	// Calculate gain % on post alone (no credits) and on (post + credits).
	gainOnPost := percentGain(ctx.PostDevelopmentUnits, ctx.PreDevelopmentUnits)

	// Net loss: post alone is below pre AND no credits to compensate.
	if ctx.PostDevelopmentUnits < ctx.PreDevelopmentUnits && ctx.StatutoryCreditsPurchased == 0 {
		return BNGNetLoss, gainOnPost, nil
	}

	gainTotal := percentGain(ctx.PostDevelopmentUnits+ctx.StatutoryCreditsPurchased, ctx.PreDevelopmentUnits)

	if gainTotal < StatutoryMinimumGainPercent {
		return BNGBelowThreshold, gainTotal, nil
	}

	// Meets threshold via credits — flag for transparency.
	if gainOnPost < StatutoryMinimumGainPercent && ctx.StatutoryCreditsPurchased > 0 {
		return BNGCreditsRequired, gainTotal, nil
	}

	return BNGMeetsThreshold, gainTotal, nil
}

// percentGain returns (post - pre) / pre * 100. pre must be > 0.
func percentGain(post, pre float64) float64 {
	return (post - pre) / pre * 100.0
}
