// Package manifest implements the R150 cohort-canonical schematised-
// knowledge envelope for limitless-environmental, extended with the
// R150 Class-3 jurisdiction-version anchor (the moat surface for
// methodology-corpus pinning).
//
// limitless-environmental's manifest pins 3 regulator-published
// methodology corpora (DEFRA Biodiversity Metric v4.0 / EU IED BAT
// conclusions / EA EPR guidance) + 4 regulation-citation references +
// 2 cohort-canonical anchors (KAT-1 + L43 Mirror-Mark) + 1
// counsel-review honest-default + 1 R85 parity.
package manifest

import (
	"sort"
	"time"
)

const SchemaVersion = 1

// FreshAtUnknown is the sentinel value for FreshAt fields that predate
// the R150 envelope discipline.
var FreshAtUnknown = time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)

const (
	SourceDEFRABiodiversityMetric    = "DEFRA + Natural England Biodiversity Metric v4.0 (statutory metric under Environment Act 2021 Schedule 14)"
	SourceEUIEDBATConclusions        = "EU IED BAT Conclusions Implementing Decisions (Directive 2010/75/EU Article 13)"
	SourceEAPermittingGuidance       = "Environment Agency Environmental Permitting Guidance (EPR 2016 SI 2016/1154 + EA technical guidance)"
	SourceEnvAct2021Schedule14       = "Environment Act 2021 c. 30 Schedule 14 (Biodiversity Gain in Planning Permission)"
	SourceEPR2016                    = "Environmental Permitting (England and Wales) Regulations 2016, SI 2016/1154"
	SourceIEDDirective201075EU       = "Directive 2010/75/EU of the European Parliament and of the Council on industrial emissions (Industrial Emissions Directive)"
	SourceContextDoc                 = "limitless-environmental CONTEXT.md"
	SourceR85ParityMarker            = "limitless-environmental R85 CLEAN-PARITY between code + CONTEXT.md"
)

type Confidence int

const (
	ConfidenceHigh   Confidence = 3
	ConfidenceMedium Confidence = 2
	ConfidenceLow    Confidence = 1
)

type Jurisdiction string

const (
	JurisdictionEU   Jurisdiction = "EU"
	JurisdictionUK   Jurisdiction = "UK"
	JurisdictionNone Jurisdiction = ""
)

// Entry is one R150 manifest entry with R150 Class-3 (jurisdiction-
// version) anchor extension.
type Entry struct {
	Key           string
	Description   string
	FreshAt       time.Time
	Source        string
	SchemaVersion int
	Confidence    Confidence
	Jurisdiction  Jurisdiction
	Version       string
}

// IsStale returns true when the entry's FreshAt has aged beyond maxAge
// or is the sentinel FreshAtUnknown value.
func (e Entry) IsStale(now time.Time, maxAge time.Duration) bool {
	if e.FreshAt.Equal(FreshAtUnknown) {
		return true
	}
	return now.Sub(e.FreshAt) > maxAge
}

type Manifest []Entry

func (m Manifest) SortedKeys() []string {
	keys := make([]string, 0, len(m))
	for _, e := range m {
		keys = append(keys, e.Key)
	}
	sort.Strings(keys)
	return keys
}

func (m Manifest) StaleEntries(now time.Time, maxAge time.Duration) []Entry {
	var out []Entry
	for _, e := range m {
		if e.IsStale(now, maxAge) {
			out = append(out, e)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Key < out[j].Key })
	return out
}

func (m Manifest) ByJurisdiction(j Jurisdiction) []Entry {
	var out []Entry
	for _, e := range m {
		if e.Jurisdiction == j {
			out = append(out, e)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Key < out[j].Key })
	return out
}

func AllJurisdictions() []Jurisdiction {
	return []Jurisdiction{JurisdictionEU, JurisdictionUK, JurisdictionNone}
}

// Seed returns the canonical R150 manifest for limitless-environmental.
// 11 entries: 3 methodology-corpus pins + 4 regulator-citation refs +
// 2 cohort-canonical anchors + 1 counsel placeholder + 1 R85 parity.
func Seed() Manifest {
	scaffold := time.Date(2026, 5, 28, 0, 0, 0, 0, time.UTC)
	tbdPhase2 := FreshAtUnknown

	return Manifest{
		// === R150 Class 3 jurisdiction-version anchor: 3 corpus pins ===
		{
			Key:           "methodology_corpus.uk.defra_biodiversity_metric_v4",
			Description:   "DEFRA + Natural England Biodiversity Metric v4.0 statutory corpus pin (Environment Act 2021 Schedule 14 BNG calculation anchor). Local SHA pin awaiting Phase-2 cold-verify against regulator-published canonical artefact.",
			FreshAt:       tbdPhase2,
			Source:        SourceDEFRABiodiversityMetric,
			SchemaVersion: SchemaVersion,
			Confidence:    ConfidenceLow,
			Jurisdiction:  JurisdictionUK,
			Version:       "v4.0",
		},
		{
			Key:           "methodology_corpus.eu.ied_bat_conclusions",
			Description:   "EU IED BAT Conclusions Implementing Decisions corpus pin (Directive 2010/75/EU Article 13 anchor for permit-condition setting per Article 14(3)). Local SHA pin awaiting Phase-2 cold-verify.",
			FreshAt:       tbdPhase2,
			Source:        SourceEUIEDBATConclusions,
			SchemaVersion: SchemaVersion,
			Confidence:    ConfidenceLow,
			Jurisdiction:  JurisdictionEU,
			Version:       "2024-rolling",
		},
		{
			Key:           "methodology_corpus.uk.ea_permitting_guidance",
			Description:   "Environment Agency Environmental Permitting guidance corpus pin (EPR 2016 SI 2016/1154 + EA technical-guidance series anchor). Local SHA pin awaiting Phase-2 cold-verify.",
			FreshAt:       tbdPhase2,
			Source:        SourceEAPermittingGuidance,
			SchemaVersion: SchemaVersion,
			Confidence:    ConfidenceLow,
			Jurisdiction:  JurisdictionUK,
			Version:       "2024-current",
		},
		// === Regulator-citation references ===
		{
			Key:           "regulation.uk.env_act_2021_schedule_14",
			Description:   "UK Environment Act 2021 c. 30 Schedule 14 (Biodiversity Gain in Planning Permission). Mandatory +10% BNG for development-consent applications under TCPA 1990 from 2024-02-12 (small sites) / 2024-04-02 (major sites).",
			FreshAt:       scaffold,
			Source:        SourceEnvAct2021Schedule14,
			SchemaVersion: SchemaVersion,
			Confidence:    ConfidenceHigh,
			Jurisdiction:  JurisdictionUK,
			Version:       "2021-11-09",
		},
		{
			Key:           "regulation.uk.epr_2016",
			Description:   "Environmental Permitting (England and Wales) Regulations 2016, SI 2016/1154. Permit grant/variation/refusal procedure; enforcement-notice authority (Regulation 36 + Schedule 17). Operative authority: Environment Agency (EA) + Natural Resources Wales (NRW).",
			FreshAt:       scaffold,
			Source:        SourceEPR2016,
			SchemaVersion: SchemaVersion,
			Confidence:    ConfidenceHigh,
			Jurisdiction:  JurisdictionUK,
			Version:       "2016-12-30",
		},
		{
			Key:           "regulation.eu.ied_directive_2010_75",
			Description:   "Directive 2010/75/EU of the European Parliament and of the Council on industrial emissions (IED). Article 14(3) BAT-conclusions reference; Article 21(3) 4-year permit-review trigger.",
			FreshAt:       scaffold,
			Source:        SourceIEDDirective201075EU,
			SchemaVersion: SchemaVersion,
			Confidence:    ConfidenceHigh,
			Jurisdiction:  JurisdictionEU,
			Version:       "2010-11-24",
		},
		{
			Key:           "regulation.uk.epr_2016_schedule_5",
			Description:   "EPR 2016 SI 2016/1154 Schedule 5 — variation procedure (Part 1 application + Part 2 public consultation triggers + Part 3 determination). Substantive change to a permitted activity without Schedule 5 procedure is non-compliant.",
			FreshAt:       scaffold,
			Source:        SourceEPR2016,
			SchemaVersion: SchemaVersion,
			Confidence:    ConfidenceHigh,
			Jurisdiction:  JurisdictionUK,
			Version:       "2016-12-30",
		},
		// === Cohort-canonical anchors ===
		{
			Key:           "cohort.l43.mirrormark_v1",
			Description:   "L43 Mirror-Mark v1 receipt algorithm byte-identical to foundation/pkg/mirrormark. 62-char 'lore@v1:' prefix + 54-char base64url body (8-byte corpus prefix + 32-byte HMAC-SHA256 digest).",
			FreshAt:       scaffold,
			Source:        SourceContextDoc,
			SchemaVersion: SchemaVersion,
			Confidence:    ConfidenceHigh,
			Jurisdiction:  JurisdictionNone,
			Version:       "v1",
		},
		{
			Key:           "cohort.r151.kat1_canonical_hex",
			Description:   "R151 KAT-1 cross-substrate hex anchor: 239a7d0d3f1bbe3a98aede01e2ad818c2db60b7177c02e2f015035b2b5b7dbca. Re-derivable via 'openssl dgst -sha256 -mac hmac -macopt key:' against canonical 33-byte input.",
			FreshAt:       scaffold,
			Source:        SourceContextDoc,
			SchemaVersion: SchemaVersion,
			Confidence:    ConfidenceHigh,
			Jurisdiction:  JurisdictionNone,
			Version:       "v1",
		},
		// === Honest-defaults + parity ===
		{
			Key:           "placeholder.counsel_review_status",
			Description:   "R166 LIABILITY-FOOTER-CONST honest-default: ReviewedByCounsel = false. Phase-1 ships placeholder legal narrative templates; counsel review + flip to true on its own R145.B sibling branch.",
			FreshAt:       scaffold,
			Source:        SourceContextDoc,
			SchemaVersion: SchemaVersion,
			Confidence:    ConfidenceLow,
			Jurisdiction:  JurisdictionNone,
			Version:       "phase-1",
		},
		{
			Key:           "r85.parity.code_vs_context",
			Description:   "R85 CLEAN-PARITY anchor — CONTEXT.md status row vs runtime ground truth. I52 marathon 2026-05-28 scaffold.",
			FreshAt:       scaffold,
			Source:        SourceR85ParityMarker,
			SchemaVersion: SchemaVersion,
			Confidence:    ConfidenceHigh,
			Jurisdiction:  JurisdictionNone,
			Version:       "v1",
		},
	}
}
