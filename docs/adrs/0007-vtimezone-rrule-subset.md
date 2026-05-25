# ADR-0007: VTIMEZONE RRULE subset for ParseTimeWithTZID — v0.1

## Status

Accepted

## Date

2026-05-04

## Context

`vstar-go-time` (Wave 2) shipped `ParseTimeWithTZID(s, tzid, cal)` —
a strict parser for RFC 5545 §3.3.5 form #1 wall-clock times
resolved against a VTIMEZONE in the same VCALENDAR. Implementation
landed at `time_tzid.go`; testdata lives at `testdata/time/`.

VTIMEZONE in RFC 5545 is open-ended: a producer may declare any
combination of STANDARD and DAYLIGHT children, each potentially
carrying RDATE entries, RRULE recurrence patterns of arbitrary
complexity, multiple historical entries (split-zone histories),
and a long-tail of edge cases. A complete VTIMEZONE-resolution
implementation is roughly 600+ LOC and effectively duplicates the
IANA tz database.

V\*'s use case is narrower: agentic systems exchanging calendar
data with explicit, simple zones (America/Toronto,
Europe/London, …). The RFC 5545 form #1 + TZID combo needs to
work for the dominant case; obscure historical zone surgery is a
separate concern.

## Decision

`ParseTimeWithTZID` accepts a documented **subset** of VTIMEZONE
shapes for v0.1:

### Accepted shapes

1. **Single STANDARD only** (no DAYLIGHT). Fixed offset zones
   (e.g. UTC, fixed-offset historical zones, GMT). The
   STANDARD child:
   - MUST carry `TZOFFSETTO` and `TZOFFSETFROM` (formats `±HHMM`
     or `±HHMMSS`).
   - MUST carry `DTSTART` (form #1).
   - MAY carry `TZNAME` (used to label the resulting `FixedZone`).
   - MAY OMIT `RRULE` (zone is fixed-offset for all time).
2. **Single STANDARD + single DAYLIGHT**. Both children:
   - MUST carry `TZOFFSETTO`, `TZOFFSETFROM`, `DTSTART`.
   - MUST carry `RRULE` with `FREQ=YEARLY`.
   - The RRULE accepts `BYMONTH=<1-12>` and
     `BYDAY=<n><WEEKDAY>` (ordinal `n` is a non-zero signed
     small integer; weekday is `SU|MO|TU|WE|TH|FR|SA`).
   - The RRULE accepts `INTERVAL=1` (silently no-op). Any other
     INTERVAL value is rejected.
3. **DAYLIGHT-only** is treated as a single fixed-offset rule
   using the DAYLIGHT child's `TZOFFSETTO`. Edge case for
   producers shipping only the active rule.

### Rejected shapes (return `(zero, false)`)

- Multiple STANDARD or multiple DAYLIGHT children (split-zone
  histories).
- RRULE without `FREQ=YEARLY` (e.g. `FREQ=MONTHLY`, `FREQ=WEEKLY`).
- RRULE containing `UNTIL`, `COUNT`, `BYWEEKNO`, `BYSETPOS`,
  `BYHOUR`, `BYMINUTE`, `BYSECOND`, `BYYEARDAY`, `BYMONTHDAY`,
  `WKST`, or any unknown key.
- BYDAY without an explicit ordinal (`SU`, `0SU`).
- RDATE-only zones (no RRULE, transitions enumerated as discrete
  dates).
- Missing `TZOFFSETTO` or `TZOFFSETFROM` in any child.
- Missing `DTSTART` in any child.

## Rationale

Three reasons.

1. **Coverage of the dominant case.** Single-rule
   STANDARD+DAYLIGHT zones with `FREQ=YEARLY;BYMONTH;BYDAY`
   covers all current IANA zones with active DST (US, EU, AU,
   most of South America). The subset accepted is exactly what
   producers writing modern timezones need.
2. **Implementation cost vs benefit.** Full VTIMEZONE
   resolution (multiple STANDARD entries, RDATE, BYWEEKNO,
   BYSETPOS) adds ~600 LOC and a second IANA-style transition
   database. v0.1 scope is "agentic systems exchange calendar
   data"; historical zone surgery is not on that path.
3. **Strict failure mode.** When a producer ships a VTIMEZONE
   outside the v0.1 subset, `ParseTimeWithTZID` returns
   `(zero, false)` rather than silently applying a partial
   rule. Consumers see torn data immediately and can fall back
   to UTC-only properties or upgrade to a full IANA-tz library.

## Consequences

- `time_tzid.go` documents the accepted/rejected shape list
  in its package-level comment (lines 16-30 currently — see
  `parseYearlyRRULE` and `loadTZRules`).
- `testdata/time/README.md` lists which fixtures exercise
  which subset boundary.
- `testdata/time/america_montreal.ics` is the canonical
  STANDARD+DAYLIGHT+YEARLY fixture; `testdata/time/utc_only.ics`
  is the STANDARD-only fixture.
- A future v0.2 may extend the subset (RDATE support, multiple
  STANDARD entries) or punt to an IANA tz library
  (`time.LoadLocation`) for any TZID resolution. The decision
  is deferred until a real consumer surfaces a need.
- The wave-2 time track shipped this subset as code; this ADR
  documents the decision after the fact ("Accepted, already
  implemented"). The ADR is filed in the canonical track only
  because canonical is the first track to introduce a `docs/adrs/`
  directory in the worktree.
- `Canonical(Component)` for any TZID-bearing property uses
  `ParseTimeWithTZID` to resolve to UTC for emission. When the
  TZID resolution fails (out-of-subset VTIMEZONE), the canonical
  form preserves the original local-time value with TZID
  parameter. This is best-effort: hash-stability is preserved
  but two implementations resolving the same VTIMEZONE
  differently could diverge. The v0.1 subset is small enough that
  this is unlikely in practice.

## Spec linkage

- `spec/03-canonicalization.md` rule 5 ("Datetime forms. UTC
  preferred for V\*; if local time is needed, an explicit TZID=
  parameter referencing a VTIMEZONE in the same VCALENDAR.") —
  this ADR documents the VTIMEZONE shape `Canonical()` resolves.

## Implementation linkage

- Track: `vstar-go-time` (already shipped, Wave 2)
- File: `time_tzid.go` — `ParseTimeWithTZID`, `loadTZRules`,
  `parseYearlyRRULE`, `parseBYDAY`.
- Testdata: `testdata/time/america_montreal.ics`,
  `testdata/time/utc_only.ics`, `testdata/time/README.md`.

## Related

- ADR-0001 (monorepo structure)
- ADR-0003 (no kit dependency — informs "no IANA tz database
  vendored" choice)
- ADR-0004 (canonical component order)
- ADR-0005 (canonical text normalization)
- ADR-0006 (canonical ATTACH handling)
- RFC 5545 §3.6.5 (VTIMEZONE), §3.8.5.3 (RRULE)
