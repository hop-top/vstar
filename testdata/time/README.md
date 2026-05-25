# vstar/go time test fixtures

These fixtures back the `time` package's TZID-aware parser
(`ParseTimeWithTZID`). They are not yet wire-decoded — the rfc5545
codec lands in a sister track. Each fixture's `.ics` file is a
documentation reference for the intended VTIMEZONE shape; the test
suite reconstructs the equivalent `Calendar` value in code (see
`time_test.go::americaMontrealCalendar` etc.).

## Fixtures

- `america_montreal.ics` — STANDARD + DAYLIGHT with `FREQ=YEARLY`
  RRULEs (BYMONTH=3;BYDAY=2SU and BYMONTH=11;BYDAY=1SU). Exercises
  the standard US/Canada DST transitions.
- `utc_only.ics` — STANDARD-only VTIMEZONE for the trivial fixed-
  offset path.

## VTIMEZONE feature subset (v0.1)

The TZID parser implements a deliberately small slice of
RFC 5545 §3.6.5 — enough to handle the world's most common DST
zones expressed via `FREQ=YEARLY` RRULEs, plus single-offset zones.

Supported:

- One STANDARD child, no DAYLIGHT — fixed offset.
- One DAYLIGHT child, no STANDARD — fixed offset (rare, tolerated).
- One STANDARD + one DAYLIGHT — both must carry an
  `RRULE:FREQ=YEARLY` with `BYMONTH=<m>` and `BYDAY=<n><WD>` parts.
  `INTERVAL=1` accepted; any other INTERVAL rejected.

Rejected (returns `(zero, false)`):

- Multiple STANDARD or DAYLIGHT entries (historical zone with
  offset changes — e.g. America/Caracas pre-2007). Workaround:
  caller maintains its own zone registry.
- `RDATE`-only zones (no RRULE).
- `FREQ` other than `YEARLY`.
- `UNTIL`, `COUNT`, `BYWEEKNO`, `BYSETPOS`, `BYMONTHDAY`, `WKST`,
  or any RRULE part not enumerated above.
- `BYDAY` without an ordinal (`SU` instead of `2SU`) — RFC permits
  it but VTIMEZONE rules require an explicit nth-of-month.

When a fixture exercises a rejected case, the test should assert
`ok == false` rather than try to coerce the zone.

This subset is documented inline in `time_tzid.go` and will be
formalised into a `spec/decisions/` ADR alongside the v0.1 release
(touching `hops/main/` is out of scope for the time track).
