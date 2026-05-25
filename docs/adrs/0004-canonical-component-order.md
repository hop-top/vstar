# ADR-0004: Canonical component order — UID-lexicographic

## Status

Accepted (Sami, 2026-05-04)

## Date

2026-05-04

## Context

`spec/03-canonicalization.md` rule 6 requires a "stable, deterministic"
component order within a VCALENDAR but leaves the exact rule TBD
("likely UID-sorted"). Two independent implementations producing
byte-identical canonical forms cannot exist until this rule is
locked.

Three candidates surfaced during design:

1. **Append-order.** Preserve writer intent. Breaks set-equality
   semantics: two calendars with the same logical content but
   different write histories canonicalize differently.
2. **UID-lexicographic.** Sort components by their UID property
   value, byte-wise (UTF-8 lexicographic). Deterministic, well-
   defined for any input that has UIDs (RFC 5545 §3.8.4.7 makes UID
   mandatory for the major components: VEVENT, VTODO, VJOURNAL,
   VFREEBUSY).
3. **Topological by RELATED-TO.** Resolve `RELATED-TO` references
   to derive a partial order. Semantic but ambiguous on cycles
   (RFC 5545 does not forbid them) and gives no total order on
   independent components.

## Decision

**UID-lexicographic** sort: top-level components within a VCALENDAR
are sorted by `UID` property value using a byte-wise (UTF-8
lexicographic) comparison. Sort is stable: components with equal
UIDs (a producer bug — UIDs are meant to be unique per RFC 5545
§3.8.4.7) preserve their relative input order.

### Edge cases

- **Components without UID.** VTIMEZONE has TZID, not UID. When a
  VTIMEZONE appears at the top level of a VCALENDAR, treat its
  TZID value as the sort key (and document this in the rule). When
  a top-level component has neither UID nor TZID (a producer bug),
  sort it to the END of the calendar in stable input order.
- **Nested sub-components** (e.g. STANDARD/DAYLIGHT inside
  VTIMEZONE; VALARM inside VEVENT). These have neither UID nor a
  natural sort key — STANDARD/DAYLIGHT carry only DTSTART
  (deliberately a local-time partial value); VALARM carries
  ACTION + TRIGGER but not UID. Nested sub-components preserve
  their **input append order** during canonicalization. Producers
  controlling nested order is the documented contract.

## Rationale

UID-lexicographic wins for three reasons:

1. **Total, deterministic, stable** for the components RFC 5545
   makes UID-mandatory on. No second-pass disambiguation
   required.
2. **No semantic dependency.** Topological-by-RELATED-TO is
   structurally elegant but breaks down on cycles, on
   cross-VCALENDAR references (an open question deferred to
   v0.2 — see ADR README), and gives no order to siblings without
   `RELATED-TO`. UID-lexicographic side-steps all of this.
3. **Round-trip stability.** A round-trip of
   `Parse → Canonical → Parse → Canonical` MUST be a fixpoint at
   the byte level. UID-lexicographic guarantees this; append-order
   does not (it depends on input write history).

Against append-order: real benefit is "preserve writer intent"
which is undefined for a canonical form by construction. The
canonical form is the equivalence-class representative; different
writer intents that produce the same logical content collapse to
the same bytes by design.

Against topological: useful for human reading, not for canonical
hashing. A future v0.2 *display* order is a separate concern and
does not need to be the canonical order.

## Consequences

- Implemented in `Canonical(Calendar) []byte` (track:
  vstar-go-canonical, task T-0034). Top-level component sort is
  the only Calendar-level concern; per-component canonicalization
  is the task T-0033 responsibility.
- VTIMEZONE-with-TZID-as-sort-key is a documented branch in the
  Calendar canonicalizer; same source of truth lives in
  `canonical.go` so the rule is single-implementation.
- Nested sub-components preserve append order — this is a real
  semantic loss for callers who depend on it not being a loss.
  Documented in the spec/03 cross-link.
- Producer bug surfaces (duplicate UIDs, missing UID + missing
  TZID) do not crash canonicalization: the sort is stable, the
  fallback is "to end in input order".
- Spec/03 rule 6 is locked at v0.1.

## Spec linkage

- `spec/03-canonicalization.md` rule 6 ("Component order within a
  VCALENDAR. Stable, deterministic — exact ordering rule TBD
  (likely UID-sorted).") — this ADR resolves the TBD.

## Implementation linkage

- Track: `vstar-go-canonical`
- Task: T-0034 (`Canonical(Calendar)` — top-level calendar
  canonicalization)
- File: `canonical.go` (`Canonical(Calendar) []byte`)

## Related

- ADR-0001 (monorepo structure — context)
- ADR-0005 (canonical text normalization — companion v0.1 lock)
- ADR-0006 (canonical ATTACH handling — companion v0.1 lock)
- ADR-0007 (VTIMEZONE RRULE subset — companion carryover)
