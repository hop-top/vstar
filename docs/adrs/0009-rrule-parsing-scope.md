# ADR-0009: Generic RRULE parsing scope — v0.2

## Status

Accepted (Sami, 2026-05-05)

> **Editorial note**: this ADR is filed, scoped, and accepted by
> the same person (Sami in the v0.1.0 release authority role).
> A v0.2 implementation track (`vstar-go-rrule`) executes against
> these decisions; if implementation surfaces an unworkable
> constraint, the ADR is amended at v0.2-cycle, not blocked.

## Amendments

### 2026-05-05 — promote BYSETPOS, BYWEEKNO, BYYEARDAY to v0.2 scope

Sami review pre-PR-#25-merge: defer-list reduces v0.2's value
proposition. The "agentic-system iCal library" framing breaks
if adopters need arran4 for any single feature. The three
BY-* clauses are cheap insurance (~330 LOC combined) against
that gap.

**Promoted to v0.2 scope:**

- BYSETPOS — positional filter on expanded BY-set. Highest
  consumer value: enables "last business day of month",
  "second Tuesday", and similar common business-calendar
  patterns.
- BYWEEKNO — ISO 8601 week-number expansion (FREQ=YEARLY only),
  WKST-aware (default MO).
- BYYEARDAY — day-of-year expansion (FREQ=YEARLY only).

**Still deferred to v0.3+:**

- FREQ=SECONDLY, FREQ=MINUTELY — interaction with
  NextOccurrence's maxIterations cap requires the O(1)-jump
  structural fix. Not a v0.2 effort.
- RSCALE (RFC 7529) — non-Gregorian calendars. Out of V*
  roadmap indefinitely.

v0.2 surface is now: "all RFC 5545 §3.3.10 BY-clauses
supported; FREQ-extreme and non-Gregorian deferred."

## Date

2026-05-05

## Context

Surfaced post-v0.1.0 release via [tlc T-0125](#) (a request from
the tlc adopter blocked on T-1230 "migrate vtodo from
arran4/golang-ical to vstar").

vstar v0.1 treats `RRULE` as an opaque property string at the
codec/Component level. Producers can round-trip it; consumers
cannot inspect it without re-implementing RFC 5545 §3.3.10
parsing.

[ADR-0007](0007-vtimezone-rrule-subset.md) carved out a narrow
RRULE subset *for VTIMEZONE DST resolution only* — `FREQ=YEARLY`
+ `BYMONTH` + `BYDAY=<n><WEEKDAY>`, with everything else
rejected. That subset is deliberately not a generic RRULE parser
and explicitly inadequate for VTODO/VEVENT recurrence (which
typically use `FREQ=DAILY|WEEKLY|MONTHLY` + `INTERVAL` +
`UNTIL`/`COUNT`).

The gap means tlc and any other adopter parsing VTODO/VEVENT
recurrence must keep a second iCal library
(`github.com/arran4/golang-ical`) purely for `ParseRecurrenceRule`.
The "agentic system state as iCalendar" framing of vstar is
undermined when adopters need a second iCal library to use vstar.

## Decision

Ship a **generic RRULE parser + structured Rule type +
NextOccurrence evaluator** as a v0.2 feature. Scope is bounded
explicitly below; everything outside the bounded scope is deferred
to v0.3 or rejected.

### Public API surface

New package `rrule/`:

```go
// Rule is the structured form of an RFC 5545 §3.3.10 RRULE value.
// Fields hold the subset accepted by v0.2; unsupported parts of
// the input cause Parse to return ErrUnsupportedRRule.
type Rule struct {
    Freq       Freq        // FREQ — required
    Interval   int         // INTERVAL (default 1; >= 1)
    Until      time.Time   // UNTIL (mutually exclusive with Count)
    Count      int         // COUNT (mutually exclusive with Until)
    ByDay      []ByDay     // BYDAY (weekday + optional ordinal)
    ByMonth    []int       // BYMONTH (1-12)
    ByMonthDay []int       // BYMONTHDAY (-31..-1, 1..31; 0 invalid)
    ByHour     []int       // BYHOUR (0-23)
    ByMinute   []int       // BYMINUTE (0-59)
    BySecond   []int       // BYSECOND (0-60; 60 = leap)
    WeekStart  Weekday     // WKST (default Monday per RFC)
}

type Freq int
const (
    FreqHourly Freq = iota + 1
    FreqDaily
    FreqWeekly
    FreqMonthly
    FreqYearly
    // Note: FreqSecondly + FreqMinutely deliberately omitted from
    // v0.2 — see "Rejected" below.
)

type Weekday int  // SU=0..SA=6 per RFC 5545

type ByDay struct {
    Ordinal int      // -53..-1, 1..53; 0 = "all weekdays of this kind"
    Weekday Weekday
}

// ParseRRule decomposes s (the property value, no "RRULE:" prefix)
// into a Rule. Returns:
//   - ErrMalformed: syntactic garbage, unknown rule-part name,
//     non-integer where integer expected, etc.
//   - ErrUnsupportedRRule: syntactically valid but uses a feature
//     v0.2 explicitly defers (e.g. FREQ=SECONDLY, BYSETPOS,
//     BYWEEKNO, BYYEARDAY, RSCALE).
func ParseRRule(s string) (Rule, error)

// ValidateRRule returns nil if s parses cleanly; otherwise the
// same error ParseRRule would return. Cheaper for boundary checks
// where the caller doesn't need the structured form.
func ValidateRRule(s string) error

// NextOccurrence returns the next time t' > after for which the
// rule fires, computing relative to dtstart (the DTSTART of the
// component carrying the RRULE).
//
// Returns (zero, false, nil) when the rule has terminated
// (UNTIL passed or COUNT exhausted relative to dtstart).
// Returns (zero, false, ErrUnsupportedRRule) when the rule is
// parseable but contains a feature NextOccurrence cannot evaluate
// (see "Evaluator-deferred features" below).
func NextOccurrence(rule Rule, dtstart, after time.Time) (time.Time, bool, error)
```

### Sentinels (new in v0.2)

- `ErrUnsupportedRRule = errors.New("rrule: feature outside v0.2 scope")`
  — wrapped via `%w`; consumers match with `errors.Is`.

### Accepted rule-parts (v0.2 parser scope)

- `FREQ` — required. Values: `HOURLY`, `DAILY`, `WEEKLY`, `MONTHLY`,
  `YEARLY`. (See "Rejected" for SECONDLY/MINUTELY.)
- `INTERVAL` — positive integer; default 1 if omitted.
- `UNTIL` — RFC 5545 form #2 (UTC, `Z`-suffixed) only. Form #1
  (local) and form #3 (with TZID) are rejected — V*'s strict-UTC
  posture from [ADR-0007](0007-vtimezone-rrule-subset.md) carries
  forward.
- `COUNT` — positive integer. Mutually exclusive with `UNTIL` per
  RFC 5545 §3.3.10.
- `BYDAY` — list of `[<ordinal>]<weekday>`; ordinal in -53..-1,
  1..53, or omitted.
- `BYMONTH` — list of 1..12.
- `BYMONTHDAY` — list of -31..-1, 1..31. (0 explicitly rejected
  per RFC.)
- `BYHOUR` — list of 0..23.
- `BYMINUTE` — list of 0..59.
- `BYSECOND` — list of 0..60. (60 retained for leap seconds per
  RFC.)
- `WKST` — single weekday; default `MO` per RFC.

Order of rule-parts is irrelevant on parse (RFC 5545 §3.3.10 makes
no order requirement); `Rule` field order is for ergonomic display.

### Evaluator scope (`NextOccurrence`)

`NextOccurrence` evaluates the parsed rule against `dtstart` and
returns the next-fire time. Supported expansion behavior:

- `FREQ=HOURLY|DAILY|WEEKLY|MONTHLY|YEARLY` with any combination
  of `INTERVAL`, `UNTIL`, `COUNT`, `BYDAY`, `BYMONTH`,
  `BYMONTHDAY`, `BYHOUR`, `BYMINUTE`, `BYSECOND`.
- WKST-aware week boundaries (default MO).
- Leap-day handling: `BYMONTHDAY=29` for Feb on non-leap years
  silently skips that occurrence; `BYMONTHDAY=-1` always picks
  the last day correctly.
- DST boundaries: evaluator operates on `time.Time` with the zone
  the caller provides via `dtstart`; "wall-clock 02:30 doesn't
  exist on spring-forward day" is handled by `time.Time`'s own
  normalization (the caller sees the next valid instant).

### Rejected (v0.2 parser returns `ErrMalformed` or `ErrUnsupportedRRule`)

Syntactic errors → `ErrMalformed`:

- Unknown rule-part name (e.g. `FOO=BAR`).
- Missing `FREQ`.
- `BYMONTHDAY=0`, `BYDAY=0SU` (ordinal 0).
- `INTERVAL=0` or negative.
- Both `UNTIL` and `COUNT` present.
- `UNTIL` in form #1 or form #3 (V* requires UTC for RRULE
  bounds).

Out-of-scope-for-v0.2 → `ErrUnsupportedRRule`:

- `FREQ=SECONDLY`, `FREQ=MINUTELY` — extreme expansion, no
  realistic agentic use case in v0.2.
- ~~`BYSETPOS` — combinatorial complexity (filter applied to
  expanded set); deferred to v0.3 if a real consumer surfaces.~~
  *(Promoted to v0.2 scope per the 2026-05-05 amendment above.)*
- ~~`BYWEEKNO` — ISO week-number expansion; deferred.~~
  *(Promoted to v0.2 scope per the 2026-05-05 amendment above.)*
- ~~`BYYEARDAY` — day-of-year expansion; deferred.~~
  *(Promoted to v0.2 scope per the 2026-05-05 amendment above.)*
- `RSCALE` (RFC 7529 — non-Gregorian calendar systems);
  deferred indefinitely (not on V*'s roadmap).

### Evaluator-deferred features (NextOccurrence returns `ErrUnsupportedRRule`)

The parser accepts these but the evaluator does not handle them in
v0.2:

- None at v0.2 ship. Parser scope == evaluator scope. If a
  v0.2.x patch surfaces an evaluator gap, it lands as
  `ErrUnsupportedRRule` from `NextOccurrence` while `ParseRRule`
  continues to accept (data preservation > evaluation).

## Rationale

Three reasons the scope lands here.

1. **Closes the tlc adoption gap.** tlc's `ParseRecurrenceRule`
   need is exactly the parser surface above; tlc can drop its
   `arran4/golang-ical` dep on v0.2. The "agentic system state as
   iCalendar" framing holds when adopters don't need a second iCal
   library to consume vstar.

2. **NextOccurrence is in-scope, not deferred.** Sami's call
   2026-05-05 (overrode the originally-recommended deferral). A
   parser that returns Rule data without an evaluator forces every
   consumer to reimplement the same expansion logic — same failure
   pattern that motivated this ADR. NextOccurrence in v0.2
   prevents the "vstar 90% / arran4 10%" outcome.

3. **Bounded scope avoids the BYSETPOS rabbit hole.**
   `BYSETPOS`/`BYWEEKNO`/`BYYEARDAY`/`RSCALE` deferred via
   `ErrUnsupportedRRule`. Real-world VTODO/VEVENT producers
   overwhelmingly use the simple FREQ+INTERVAL+UNTIL/COUNT+BYDAY
   subset; the deferred features are ~5% of the spec and ~50% of
   the implementation complexity. Strict failure mode means
   consumers see torn data immediately and can fall back rather
   than silently miscompute.

## Consequences

- New `rrule/` subpackage. Independent of `vstar` root and
  other subpackages except for `vstar.ParseTime` (UTC parse for
  UNTIL).
- New sentinel `rrule.ErrUnsupportedRRule`. Re-exported from the
  root `vstar` package? **No** — keeps the rrule import boundary
  explicit. Consumers that need it import `rrule`.
- Conformance corpus extended with `testdata/rrule/` fixtures:
  one happy-path per FREQ value (5), one per BYxxx clause (6),
  one per Sentinel-rejection class (~6 negative cases). Each gets
  the `.notes.md` quartet companion (no `.canonical` / `.hash` —
  RRULE is parsed, not canonicalized at this layer).
- Validate package gains `VS050` (RRULE syntactically valid but
  uses unsupported feature) — Warning severity, not Error,
  because the property still round-trips correctly.
- ADR-0007's VTIMEZONE-DST RRULE subset is **not** subsumed by
  this work. The two parsers serve different purposes:
  - `time_tzid.parseYearlyRRULE` — narrow, internal to VTIMEZONE
    resolution, returns transition rules.
  - `rrule.ParseRRule` — generic, public, returns a `Rule` struct.
  v0.2 may consolidate if implementation surfaces obvious
  duplication; default is to keep them separate.
- Release: v0.2.0 (semver minor — new public API surface, no
  breaking changes to v0.1 surface). release-please will catch the
  `feat:` commits and propose v0.2.0 automatically.
- crm and AGR are downstream consumers of this gap; both can drop
  bespoke recurrence handling on v0.2.

## Spec linkage

- RFC 5545 §3.3.10 (RECUR value type) — the canonical spec for
  RRULE. v0.2 implements a documented subset; rejected features
  are noted with section references in code comments.
- RFC 7529 (Non-Gregorian Recurrence Rules) — RSCALE deferred
  indefinitely; out of V* scope.
- spec/03 canonicalization — RRULE properties pass through
  unchanged in canonical form (the property value is preserved
  byte-for-byte; canonical does not normalize rule-part order).
  This decision lands as an addendum in spec/03 §"Property values
  preserved verbatim" when v0.2 ships.

## Implementation linkage

- Track: `vstar-go-rrule` (v0.2 epic, file in tlc post-ADR
  acceptance).
- Files: `rrule/rrule.go`, `rrule/parse.go`,
  `rrule/validate.go`, `rrule/next.go`, plus `*_test.go`
  and `testdata/rrule/`.
- Validate integration: `validate/rrule.go` adds VS050
  warning rule that calls `rrule.ValidateRRule` on every RRULE
  property in a component.

## Related

- ADR-0001 (monorepo structure)
- ADR-0007 (VTIMEZONE RRULE subset — narrow internal precedent)
- tlc T-0125 (the consumer request that surfaced this gap)
- tlc T-1230 (the tlc-side migration this unblocks)
