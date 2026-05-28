## limitless-environmental

UK Environment Agency + Environment Act 2021 + EU IED compliance forge.

Phase-1 scaffold (I52 marathon 2026-05-28). Pure-Go stdlib, zero
runtime dependencies.

### What it does

Five domain surfaces:

1. R143 honest-defaults — 5 ENVIRONMENTAL_* advisories cover EPR 2016
   Schedule 5 variation procedure, EU IED BAT-conclusions pinning,
   Environment Act 2021 BNG +10% minimum, methodology version pinning,
   counsel-review status.

2. R150 manifest — 11 entries pinning DEFRA Biodiversity Metric v4.0,
   EU IED BAT conclusions, EA permitting guidance, plus regulator
   citations and L43 Mirror-Mark + R151 KAT-1 cohort anchors.

3. R153 regulated-decision escape — 5 canonical decision classes
   (permit grant/refusal, permit variation, enforcement notice,
   statutory BNG sign-off, IED BAT condition setting) all route to
   the qualified environmental officer.

4. permit_gate — EPR 2016 + EU IED permit-state classifier emitting
   one of 5 outcomes (EA_PERMIT_FRESH / VARIATION_PENDING /
   BAT_CONCLUSIONS_DRIFT / SCHEDULE_5_REVIEW_DUE /
   NONCOMPLIANCE_NOTICE_OPEN).

5. biodiversity_gate — Environment Act 2021 Schedule 14 BNG +10%
   gate emitting one of 4 outcomes (BNG_MEETS_THRESHOLD /
   BNG_BELOW_THRESHOLD / BNG_CREDITS_REQUIRED / BNG_NET_LOSS).

### CLI

```
environmental advisories list
environmental manifest list
environmental escape list
environmental permit classify --ref EPR/AB1234XY --issued 2024-01-01
environmental bng classify --site 23/04567/FUL --pre 100 --post 112
```

### Cohort posture

- R145.B BEHAVIOR-CHANGING-WORK-GETS-ITS-OWN-BRANCH compliant: each
  domain surface ships on its own additive branch.
- R145.C FIREWALL-TEST-DISCIPLINE: internal/firewall checks the
  on-disk package list matches the canonical 7-pack at every test run.
- R151 KAT-AS-COHORT-INVARIANT-CROSS-SUBSTRATE-PIN: KAT-1 / KAT-6 /
  KAT-7 cohort literals re-derived in internal/mirrormark with hex
  embedded-digest check connecting the mark literal back to OpenSSL.
- R153 REGULATED-ROLE-ESCAPE-INVARIANT: every permit / BNG outcome
  except the happy-path returns IsRegulatoryEscape() == true.
- R166 LIABILITY-FOOTER-CONST: 5-axis const surfaced in internal/legal.

### Composes with

- aegis (M59) — UK GDPR + EU AI Act compound regulatory cohort.
- energyaudit (I51) — UK BEIS energy-savings opportunity compliance.
- green-anchor (M12) — EU CSRD + UK SECR + SEC climate disclosure.

### Status

| Field | Value |
|---|---|
| Substrate | Go 1.22 pure-stdlib |
| Runtime deps | Zero |
| Tests | 47 |
| Packages | 7 (5 cohort + 2 domain) |
| Counsel-reviewed | FALSE (Phase-1 scaffold) |
