# limitless-environmental — Workshop Context

**Status**: I52 marathon 2026-05-28 Phase-1 scaffold shipped. NEW flagship.

UK Environment Agency + Environment Act 2021 + EU IED compliance
forge. Composes with aegis (UK GDPR) + energyaudit (UK BEIS) + green-
anchor (EU CSRD + UK SECR + SEC climate) for the **compound
UK+EU climate-environmental cohort**.

## Health (I52 scaffold 2026-05-28)

| Field | Value |
|---|---|
| **Language** | Go 1.22 |
| **Substrate** | Pure-Go stdlib CLI (zero runtime deps, zero env reads in default path, zero network, zero DB) |
| **Branch** | main |
| **Remote** | https://github.com/davly/limitless-environmental.git |
| **Source files (Go, non-test)** | 7 (5 cohort + 2 domain) |
| **Test files (Go)** | 7 (one per source package) |
| **Test funcs** | 47 |
| **Test status** | GREEN (all 47 PASS on local Go 1.22 toolchain) |
| **R-patterns covered** | 7 named (R115 + R143 + R145.C + R150 + R151 + R153 + R166) |
| **Tool versions** | golang 1.22 |

## Cohort packages (5)

| Pkg | File(s) | Responsibility |
|---|---|---|
| mirrormark | `internal/mirrormark/mirrormark.go` | L43 Mirror-Mark v1 byte-identical to foundation/pkg/mirrormark (R151 KAT-1 pinned, KAT-6 / KAT-7 cohort literals) |
| honest | `internal/honest/honest.go` | R143 LOUD-ONCE-WARNING-FLAG; 5 ENVIRONMENTAL_* advisories |
| legal | `internal/legal/legal.go` | R166 5-axis LIABILITY_FOOTER_TEMPLATE + R153 ENV_REGULATED_DECISION_ESCAPE (5 canonical decision classes) |
| manifest | `internal/manifest/manifest.go` | R150 11-entry envelope with R150 Class-3 jurisdiction-version anchor |
| firewall | `internal/firewall/firewall.go` | R145.C structural firewall against internal/ + cmd/ drift |

## Domain packages (2)

| Pkg | File(s) | Responsibility |
|---|---|---|
| permit_gate | `internal/permit_gate/permit_gate.go` | EPR 2016 + EU IED permit-state classifier; R115 single-enum 5-outcome (EA_PERMIT_FRESH / VARIATION_PENDING / BAT_CONCLUSIONS_DRIFT / SCHEDULE_5_REVIEW_DUE / NONCOMPLIANCE_NOTICE_OPEN) |
| biodiversity_gate | `internal/biodiversity_gate/biodiversity_gate.go` | Environment Act 2021 Schedule 14 BNG +10% gate; R115 single-enum 4-outcome (BNG_MEETS_THRESHOLD / BNG_BELOW_THRESHOLD / BNG_CREDITS_REQUIRED / BNG_NET_LOSS) |

## CLI binary

| Cmd | File | Responsibility |
|---|---|---|
| environmental | `cmd/environmental/main.go` | 6 subcommands: advisories list / manifest list / escape list / permit classify / bng classify / version |

## Methodology corpora pinned (R150 Class 3)

| Pin | Jurisdiction | Version | Status |
|---|---|---|---|
| methodology_corpus.uk.defra_biodiversity_metric_v4 | UK | v4.0 | Phase-2 cold-verify pending |
| methodology_corpus.eu.ied_bat_conclusions | EU | 2024-rolling | Phase-2 cold-verify pending |
| methodology_corpus.uk.ea_permitting_guidance | UK | 2024-current | Phase-2 cold-verify pending |

## Honest-defaults (R143 LOUD-ONCE)

5 ENVIRONMENTAL_* advisories — 3 Error + 2 Warn:

- ENVIRONMENTAL_EA_PERMIT_VARIATION_PROCEDURE_REQUIRED (Error)
- ENVIRONMENTAL_EU_IED_BAT_CONCLUSIONS_PIN_REQUIRED (Error)
- ENVIRONMENTAL_ENV_ACT_2021_BNG_10_PERCENT_REQUIRED (Error)
- ENVIRONMENTAL_METHODOLOGY_VERSION_PIN_REQUIRED (Warn)
- ENVIRONMENTAL_REVIEWED_BY_COUNSEL_FALSE (Warn)

## Regulated-decision escape (R153)

5 canonical decision classes route to qualified environmental officer:

- permit_grant_or_refusal
- permit_variation
- enforcement_notice_issuance
- statutory_bng_sign_off
- ied_bat_condition_setting

Sentinel: `ENV_REGULATED_DECISION_ESCAPE`.

## Composes with

- **aegis** (M59) — UK GDPR + EU AI Act compound regulatory cohort.
- **energyaudit** (I51) — UK BEIS ESOS energy-savings opportunity.
- **green-anchor** (M12) — EU CSRD + UK SECR + SEC climate disclosure.

Together: compound UK+EU climate-environmental cohort.

## Phase-2 obligations

- Cold-verify the 3 methodology-corpus SHAs against regulator-published artefacts.
- Counsel review the LIABILITY_FOOTER_TEMPLATE + flip ReviewedByCounsel to true on a R145.B sibling branch.
- Wire-in Mirror-Mark with `ENVIRONMENTAL_MIRRORMARK_ENABLED=true` env-gate (Phase-2 task).
- Add SEPA + NRW + NIEA jurisdiction-specific permit-state nuances (currently EA + EU-wide only).
