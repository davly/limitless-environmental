# Security posture — limitless-environmental

## Boundary

Pure-Go stdlib CLI. Zero runtime dependencies. Zero network calls
unless the operator explicitly wires Phase-2 Mirror-Mark + corpus
verification. No DB reads. No filesystem writes (CLI emits to stdout
only).

## Attack surface (Phase-1 scaffold)

- CLI argv parsing (stdlib `flag` package, fixed grammar).
- A closed-set vocabulary of permit outcomes (5) and BNG outcomes (4).
- No interpolation of user-provided strings into shell commands.
- No SQL.
- No HTTP server.

## R166 LIABILITY-FOOTER-CONST honest-default

This scaffold ships `ReviewedByCounsel = false`. All legal-disclosure
narratives are PLACEHOLDER and have NOT been reviewed by qualified
environmental-law counsel. Operator MUST commission counsel review
and flip the flag on a R145.B sibling branch before any live
permit-decision deployment.

## R153 REGULATED-DECISION-ESCAPE

The package emits `ENV_REGULATED_DECISION_ESCAPE` for every permit /
BNG outcome other than the happy-path. Consumers MUST route every
escape result to a qualified environmental officer at the relevant
regulator (EA / SEPA / NRW / NIEA) or a chartered environmental
consultant.

## Methodology version drift (R150)

The three regulator-published methodology corpora pinned in
`internal/manifest/` are PLACEHOLDER pins at Phase-1. Cold-verify
against the regulator-published canonical artefact (DEFRA + Natural
England Biodiversity Metric v4.0 / EU Commission IED BAT
Conclusions Implementing Decisions / EA Environmental Permitting
guidance) before any production decision.

## R151 KAT-1 substrate-parity firewall

`internal/mirrormark/` pins the cohort KAT-1 / KAT-6 / KAT-7 mark
literals byte-identical to foundation/pkg/mirrormark. Test
`TestVerify_KAT1Mark` fails CI if the Mirror-Mark v1 algorithm
drifts from the cohort.

## Reporting vulnerabilities

Open a private security advisory at
https://github.com/davly/limitless-environmental/security/advisories/new.
