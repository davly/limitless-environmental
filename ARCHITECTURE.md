# limitless-environmental — Architecture

Go 1.22 pure-stdlib CLI implementing UK EPR 2016 + Environment Act 2021
+ EU IED compliance gates.

## Language and build

| Field | Value |
|---|---|
| **Language** | Go 1.22 |
| **Build** | `go build ./...` |
| **Test** | `go test ./...` |
| **Runtime deps** | Zero (stdlib only) |
| **Test deps** | Zero (stdlib testing only) |

## Module layout

```
limitless-environmental/
├── cmd/
│   └── environmental/         CLI binary
│       └── main.go            6 subcommands
├── internal/
│   ├── mirrormark/            L43 Mirror-Mark v1 (R151 KAT-pinned)
│   ├── honest/                R143 LOUD-ONCE-WARNING (5 advisories)
│   ├── legal/                 R166 footer + R153 escape (5 classes)
│   ├── manifest/              R150 envelope (11 entries)
│   ├── firewall/              R145.C structural firewall
│   ├── permit_gate/           EPR 2016 + EU IED (5 outcomes)
│   └── biodiversity_gate/     Env Act 2021 Schedule 14 (4 outcomes)
├── go.mod
├── LICENSE
├── README.md
├── CONTEXT.md
├── SECURITY.md
└── ARCHITECTURE.md
```

## Cohort discipline

- **R145.B** behavior-changing work gets its own branch.
- **R145.C** structural firewall: internal/firewall pins on-disk packages.
- **R151** KAT-AS-COHORT-INVARIANT cross-substrate pin: KAT-1 / KAT-6 / KAT-7 literals byte-identical to foundation/pkg/mirrormark.
- **R153** REGULATED-ROLE-ESCAPE-INVARIANT: 5 regulated-decision classes route to HUMAN.
- **R166** LIABILITY-FOOTER-CONST 5-axis: identity + scope + disclaimer + authority + counsel-review.
- **R143** LOUD-ONCE-WARNING (sync.Once R157 substrate-native idiom).
- **R150** PARALLEL-MAP envelope with Class-3 jurisdiction-version anchor.
- **R115** single-enum rejection-outcome shape (permit_gate 5-enum + biodiversity_gate 4-enum).

## Composes with

- **aegis** (M59) Rust UK GDPR + EU AI Act → compound regulatory cohort
- **energyaudit** (I51) UK BEIS ESOS energy-savings opportunity
- **green-anchor** (M12) Go EU CSRD + UK SECR + SEC climate disclosure
