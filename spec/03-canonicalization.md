# V* — Canonicalization

> Status: **v0.1** (locked 2026-05-04). Byte-for-byte rules
> finalized via ADRs 0004/0005/0006/0007/0008 (all Accepted) and
> implemented in `hop-top/vstar/go` v0.1.0. The TS reference
> implementation will cross-validate against the same ADRs;
> divergence is a bug in whichever implementation drifts.
>
> **v0.1 ADR locks** (filed during the Go reference-implementation
> work — see `docs/adrs/`):
>
> - [ADR-0004](../docs/adrs/0004-canonical-component-order.md):
>   Component order — UID-lexicographic (resolves rule 6).
> - [ADR-0005](../docs/adrs/0005-canonical-text-normalization.md):
>   Text normalization — NFC required (resolves open question on
>   SUMMARY/DESCRIPTION normalization).
> - [ADR-0006](../docs/adrs/0006-canonical-attach-handling.md):
>   ATTACH handling — reference-only on emit (resolves open
>   question on inline vs reference).
> - [ADR-0007](../docs/adrs/0007-vtimezone-rrule-subset.md):
>   VTIMEZONE RRULE subset for TZID resolution (informs rule 5).
> - [ADR-0008](../docs/adrs/0008-canonical-datetime-context.md):
>   Canonical datetime context — two-tier API (resolves rule 5).

## Goal

Two V\* documents containing the same logical content MUST produce
identical canonical byte sequences and identical `X-VSTAR-HASH`
values.

## Areas requiring canonical form

1. **Line endings.** CRLF, per RFC 5545.
2. **Property order within a component.** Alphabetical by property
   name; parameters in the same alphabetical order.
3. **Folding.** RFC 5545 line folding (75-octet boundary) applied
   AFTER property assembly, not before.
4. **Property parameter values.** Quoted forms canonical (per RFC).
5. **Datetime forms.** UTC (`Z`-suffixed) preferred for V\*; if local
   time is needed, an explicit `TZID=` parameter referencing a
   VTIMEZONE in the same VCALENDAR.
6. **Component order within a VCALENDAR.** Stable, deterministic —
   exact ordering rule TBD (likely UID-sorted).
7. **`X-VSTAR-HASH` exclusion.** When computing the hash for a
   component, the `X-VSTAR-HASH` property itself is excluded.
8. **`RRULE` property values preserved verbatim.** `RRULE` property
   values are preserved byte-for-byte in canonical form; the
   canonical layer does not normalize rule-part order, drop
   defaulted rule-parts, or re-format integer lists. Two RRULEs
   that decompose to the same `rrule.Rule` but differ in wire
   shape (e.g. `FREQ=DAILY;INTERVAL=1` vs `FREQ=DAILY` —
   semantically identical because INTERVAL defaults to 1) hash
   differently. Producers needing semantic equivalence MUST
   normalize the wire form before emit; canonicalization is a
   byte-level discipline. See
   [ADR-0009](../docs/adrs/0009-rrule-parsing-scope.md) for the
   rrule scope and the rationale for keeping rule-part order out
   of canonical's responsibility.

## Hashing

`X-VSTAR-HASH` SHOULD be `sha256:<hex>` of the component's
canonical form (with the hash property removed). The `sha256:`
prefix allows future algorithm migration.

## Open questions

Resolved at v0.1 by ADRs filed in the canonical-track work
(see header — ADRs 0004 through 0007):

- ~~Locale-specific normalization for text-bearing properties
  (SUMMARY, DESCRIPTION) — NFC required?~~ → ADR-0005 (NFC required).
- ~~Handling of binary attachments (`ATTACH`) — inline vs
  reference~~ → ADR-0006 (reference-only on emit).

**Deferred to v0.2:**

- vCard profile selection (vCard 4.0 baseline vs allowing 3.0?).
- Cross-VCALENDAR references (URI scheme for `RELATED-TO` across
  ledger files).

Two reference implementations (Go in this repo at v0.1.0, then
TypeScript via a forthcoming `vstar-ts` library) validate the
locked rules. AGR's Racket emitter cross-validates against the
Go output (non-blocking for v0.1.0).
