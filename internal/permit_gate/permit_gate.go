// Package permit_gate implements the EPR 2016 + EU IED permit-state
// classification gate for limitless-environmental.
//
// The package consumes a permit context (issuance date, last variation
// date, BAT-conclusions-publication date, compliance-notice status)
// and emits a single rejection-outcome enum (R115 single-enum
// rejection-outcome pattern). The 5 canonical outcomes:
//
//   - EA_PERMIT_FRESH: permit fresh, within Schedule 5 + IED windows.
//   - VARIATION_PENDING: Schedule 5 variation procedure in flight.
//   - BAT_CONCLUSIONS_DRIFT: published BAT conclusions are > 4 years
//     since the last permit-condition review (IED Article 21(3) trigger).
//   - SCHEDULE_5_REVIEW_DUE: EPR 2016 Schedule 5 review interval reached.
//   - NONCOMPLIANCE_NOTICE_OPEN: enforcement / suspension / revocation
//     notice open against the permit (EPR 2016 Reg 36 + Schedule 17).
//
// Mirror-Mark + counsel-escape composition: a permit_gate outcome
// other than EA_PERMIT_FRESH composes with the R153 escape-invariant in
// internal/legal/ — the caller MUST route to a qualified
// environmental officer (EA / SEPA / NRW / NIEA) regardless of the
// upstream Mirror-Mark verify result.
package permit_gate

import (
	"errors"
	"time"
)

// PermitOutcome is the canonical R115 single-enum rejection-outcome.
type PermitOutcome string

const (
	EAPermitFresh             PermitOutcome = "EA_PERMIT_FRESH"
	VariationPending          PermitOutcome = "VARIATION_PENDING"
	BATConclusionsDrift       PermitOutcome = "BAT_CONCLUSIONS_DRIFT"
	Schedule5ReviewDue        PermitOutcome = "SCHEDULE_5_REVIEW_DUE"
	NoncomplianceNoticeOpen   PermitOutcome = "NONCOMPLIANCE_NOTICE_OPEN"
)

// AllOutcomes returns the canonical 5-outcome list.
func AllOutcomes() []PermitOutcome {
	return []PermitOutcome{
		EAPermitFresh,
		VariationPending,
		BATConclusionsDrift,
		Schedule5ReviewDue,
		NoncomplianceNoticeOpen,
	}
}

// IsRegulatoryEscape reports whether an outcome requires routing to
// the qualified environmental officer per R153. Everything except
// EA_PERMIT_FRESH escapes.
func (o PermitOutcome) IsRegulatoryEscape() bool {
	return o != EAPermitFresh
}

// PermitContext is the canonical 7-field input to Classify.
type PermitContext struct {
	// PermitReference — the EA permit reference (e.g. "EPR/AB1234XY").
	PermitReference string
	// IssuanceDate — original permit grant date.
	IssuanceDate time.Time
	// LastVariationDate — last EPR 2016 Schedule 5 variation date
	// (zero value if no variation has occurred since issuance).
	LastVariationDate time.Time
	// BATConclusionsPublicationDate — date the relevant EU IED BAT
	// conclusions Implementing Decision was published.
	BATConclusionsPublicationDate time.Time
	// VariationInFlight — true when a Schedule 5 variation procedure
	// is currently open (regardless of stage).
	VariationInFlight bool
	// EnforcementNoticeOpen — true when an EPR 2016 Regulation 36 +
	// Schedule 17 enforcement / suspension / revocation notice is open.
	EnforcementNoticeOpen bool
	// Schedule5ReviewIntervalYears — operator-configured review
	// interval. EA practice is typically 4 years.
	Schedule5ReviewIntervalYears int
}

// ErrEmptyPermitReference — PermitReference required.
var ErrEmptyPermitReference = errors.New("permit_gate: PermitReference is required")

// ErrInvalidIssuanceDate — IssuanceDate must be non-zero.
var ErrInvalidIssuanceDate = errors.New("permit_gate: IssuanceDate is required (non-zero)")

// IEDArticle21ReviewYears is the canonical 4-year permit-review trigger
// under EU IED Article 21(3).
const IEDArticle21ReviewYears = 4

// DefaultSchedule5ReviewYears is the EA standard 4-year Schedule 5
// review interval. Operators can override per-permit.
const DefaultSchedule5ReviewYears = 4

// Classify returns the canonical PermitOutcome for the given context.
//
// Precedence (descending):
//   1. NONCOMPLIANCE_NOTICE_OPEN — most severe; gate output regardless.
//   2. BAT_CONCLUSIONS_DRIFT — IED Article 21(3) statutory trigger.
//   3. SCHEDULE_5_REVIEW_DUE — EA Schedule 5 review interval.
//   4. VARIATION_PENDING — Schedule 5 variation procedure in flight.
//   5. EA_PERMIT_FRESH — happy path.
func Classify(ctx PermitContext, now time.Time) (PermitOutcome, error) {
	if ctx.PermitReference == "" {
		return "", ErrEmptyPermitReference
	}
	if ctx.IssuanceDate.IsZero() {
		return "", ErrInvalidIssuanceDate
	}

	if ctx.EnforcementNoticeOpen {
		return NoncomplianceNoticeOpen, nil
	}

	// IED Article 21(3) — 4-year trigger from BAT conclusions publication
	// to permit-condition reconsideration.
	if !ctx.BATConclusionsPublicationDate.IsZero() {
		batAge := now.Sub(ctx.BATConclusionsPublicationDate)
		if batAge > time.Duration(IEDArticle21ReviewYears)*365*24*time.Hour {
			lastReview := ctx.LastVariationDate
			if lastReview.IsZero() {
				lastReview = ctx.IssuanceDate
			}
			if lastReview.Before(ctx.BATConclusionsPublicationDate) {
				return BATConclusionsDrift, nil
			}
		}
	}

	// EPR 2016 Schedule 5 review interval.
	reviewYears := ctx.Schedule5ReviewIntervalYears
	if reviewYears <= 0 {
		reviewYears = DefaultSchedule5ReviewYears
	}
	lastReview := ctx.LastVariationDate
	if lastReview.IsZero() {
		lastReview = ctx.IssuanceDate
	}
	reviewAge := now.Sub(lastReview)
	if reviewAge > time.Duration(reviewYears)*365*24*time.Hour {
		return Schedule5ReviewDue, nil
	}

	if ctx.VariationInFlight {
		return VariationPending, nil
	}

	return EAPermitFresh, nil
}
