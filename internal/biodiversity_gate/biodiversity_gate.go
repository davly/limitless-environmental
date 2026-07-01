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
	"math"
	"math/big"
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

// ErrNonFiniteUnits — no biodiversity-units field may be NaN or ±Inf.
// Non-finite values defeat every threshold comparison (NaN comparisons
// are always false), so without this guard malformed input would fall
// through to the only non-escaping outcome (BNG_MEETS_THRESHOLD) and
// silently bypass R153 planning-authority / Natural England review.
var ErrNonFiniteUnits = errors.New("biodiversity_gate: biodiversity-units fields must be finite (no NaN/Inf)")

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
	// Fail closed on non-finite input before any numeric comparison: NaN
	// and ±Inf defeat the `< 0`/`== 0`/threshold guards below and would
	// otherwise fall through to BNG_MEETS_THRESHOLD (the only outcome that
	// does NOT route to R153 regulatory review).
	for _, x := range [...]float64{ctx.PreDevelopmentUnits, ctx.PostDevelopmentUnits, ctx.StatutoryCreditsPurchased} {
		if math.IsNaN(x) || math.IsInf(x, 0) {
			return "", 0, ErrNonFiniteUnits
		}
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

// ClassifyExact is the EXACT-arithmetic counterpart of Classify for the +10% BNG gate.
// It decides the statutory verdict over exact rationals so the +10% boundary can never
// flip on IEEE-754 rounding: pre=3.0, post=3.3 is EXACTLY +10% (MEETS), but the float
// path computes (3.3-3.0)/3.0*100 = 9.9999999999999929 < 10 and wrongly returns
// BNG_BELOW_THRESHOLD (Environment Act 2021 Sch.14, wave-2 ENV-1/ENV-4). Callers should
// parse the DEFRA Metric v4.0 decimal inputs directly to *big.Rat (new(big.Rat).SetString
// gives an EXACT value for any finite decimal, e.g. "3.3" -> 33/10) rather than through
// float64, which closes the ingestion seam. The returned gain is the exact gain %.
func ClassifyExact(siteRef string, pre, post, credits *big.Rat) (BNGOutcome, *big.Rat, error) {
	if siteRef == "" {
		return "", nil, ErrEmptySiteReference
	}
	if pre == nil || post == nil || credits == nil {
		return "", nil, ErrNonFiniteUnits
	}
	if pre.Sign() < 0 || post.Sign() < 0 || credits.Sign() < 0 {
		return "", nil, ErrNegativeUnits
	}
	if pre.Sign() == 0 {
		return "", nil, ErrZeroPreUnits
	}

	gainOnPost := percentGainRat(post, pre)

	// Net loss: post alone below pre AND no credits to compensate.
	if post.Cmp(pre) < 0 && credits.Sign() == 0 {
		return BNGNetLoss, gainOnPost, nil
	}

	total := new(big.Rat).Add(post, credits)
	gainTotal := percentGainRat(total, pre)
	ten := big.NewRat(int64(StatutoryMinimumGainPercent), 1)

	if gainTotal.Cmp(ten) < 0 {
		return BNGBelowThreshold, gainTotal, nil
	}
	// Meets threshold only via credits — flag for transparency.
	if gainOnPost.Cmp(ten) < 0 && credits.Sign() > 0 {
		return BNGCreditsRequired, gainTotal, nil
	}
	return BNGMeetsThreshold, gainTotal, nil
}

// percentGainRat returns the exact (x - pre) / pre * 100. pre must be non-zero.
func percentGainRat(x, pre *big.Rat) *big.Rat {
	g := new(big.Rat).Sub(x, pre)
	g.Quo(g, pre)
	return g.Mul(g, big.NewRat(100, 1))
}
