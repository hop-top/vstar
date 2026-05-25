# vstar — Architecture Decision Records

The canonical home of the V\* decision record. Each ADR captures one
locked decision; new decisions land here as a numbered file with
status `Proposed`, then flip to `Accepted` once Sami signs off.

## Release status

All v0.1.0 canonical ADRs (0004, 0005, 0006, 0007, 0008) and the
RRULE-scope ADR (0009) are **Accepted**. `spec/03-canonicalization.md`
is locked at v0.1 and cross-references the accepted ADRs. The
`v0.1.0` tag is cut and live at `hop.top/vstar@v0.1.0`.

## Index

| ADR  | Title                                           | Status   | Notes                                                    |
|------|-------------------------------------------------|----------|----------------------------------------------------------|
| 0001 | Monorepo structure                              | Accepted | (file pending — see CONTRIBUTING.md for rationale)       |
| 0002 | Go library license — Apache-2.0                 | Accepted | (file pending — referenced from `README.md`)          |
| 0003 | No `hop.top/kit` dependency                     | Accepted | (file pending — referenced from `README.md`)          |
| 0004 | Canonical component order — UID-lexicographic   | Accepted | locked 2026-05-04                                        |
| 0005 | Canonical text normalization — NFC required     | Accepted | locked 2026-05-04                                        |
| 0006 | Canonical ATTACH handling — reference-only      | Accepted | locked 2026-05-04                                        |
| 0007 | VTIMEZONE RRULE subset — v0.1                   | Accepted | locked during `vstar-go-time` track; documented after the fact |
| 0008 | Canonical datetime context — two-tier API       | Accepted | locked 2026-05-04                                        |
| 0009 | RRULE parsing scope                             | Accepted | locked 2026-05-05                                        |

## Spec/03 open questions deferred to v0.2

Per the canonical-track plan, two of the spec/03 open questions
are NOT addressed by v0.1 ADRs and remain open for v0.2:

- **vCard profile selection** (vCard 4.0 baseline vs allowing 3.0?)
  — fixtures track or release agent picks this up.
- **Cross-VCALENDAR `RELATED-TO` references** (URI scheme for
  cross-ledger refs) — fixtures track or release agent picks
  this up.
