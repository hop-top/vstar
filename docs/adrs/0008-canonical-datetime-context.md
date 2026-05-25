# ADR-0008: Canonical datetime context — two-tier API

## Status

Accepted (Sami, 2026-05-04)

## Date

2026-05-04

## Context

`spec/03-canonicalization.md` rule 5 requires that "Datetime forms.
UTC (`Z`-suffixed) preferred for V\*; if local time is needed, an
explicit `TZID=` parameter referencing a VTIMEZONE in the same
VCALENDAR." The implication for the canonical form is that any
datetime carrying a TZID parameter must be resolved against the
referenced VTIMEZONE and re-emitted in UTC form #2 — otherwise two
calendars carrying logically equivalent times in different
representations canonicalize to different bytes, breaking the
foundational invariant that hash equality follows logical equality.

Wave 3a shipped `Canonical(Component) []byte` but did not
implement rule 5. The reason recorded at review time: a Component,
in isolation, has no access to the VTIMEZONE registry — VTIMEZONE
components live as siblings inside the parent VCALENDAR, not
inside the Component being canonicalized. Resolving TZID requires
threading Calendar context through.

Three options surfaced when planning the fixup:

1. **Resolve unconditionally.** Always look up TZID, fail loudly
   when no Calendar context is available. Forces every caller to
   wrap a single Component in a synthesised Calendar, even when
   the component has no TZID datetimes (the common case).
2. **Resolve silently when possible.** Add a Calendar registry
   parameter; when missing, skip resolution. Caller cannot tell
   whether their bytes are the canonical form or a verbatim
   degradation.
3. **Two-tier API.** Two named entry points: a verbatim form for
   Components that don't carry TZIDs, and a context-taking form
   that resolves them. The function name itself signals which
   guarantee the caller gets.

## Decision

Adopt option 3: a two-tier API in `canonical`.

```go
// Verbatim datetime emit; for Components without TZID datetimes.
func Component(c vstar.Component) []byte

// Resolves TZID-tagged datetimes against cal's VTIMEZONE registry.
func ComponentInContext(c vstar.Component, cal vstar.Calendar) []byte

// Routes through ComponentInContext using c itself as the registry.
func Calendar(c vstar.Calendar) []byte
```

Implementation is single-source: `Component(c)` is a thin wrapper
around `ComponentInContext(c, vstar.Calendar{})`. The empty
Calendar carries no VTIMEZONE entries, so TZID resolution always
fails and the component falls back to verbatim emit — preserving
Wave 3a's behaviour exactly for callers who don't pass a Calendar.

`Card(c)` is unchanged; VCARD components have no datetime / TZID
concerns under V\* v0.1.

### Resolution semantics

For each property whose name is one of `DTSTAMP`, `DTSTART`,
`DTEND`, `DUE`, `COMPLETED`, `RECURRENCE-ID`, `CREATED`, or
`LAST-MODIFIED` (the v0.1 datetime allow-list):

- **No TZID parameter.** Pass through verbatim. UTC form #2
  (`Z`-suffixed) values are already canonical; bare local-time
  values without a TZID are non-conforming producer output and
  there is nothing canonical can do to recover them.
- **TZID parameter present, value is form #2 (`Z` suffix).**
  Treat as already canonical and pass through verbatim. (V\* form
  #2 + TZID is contradictory wire output; canonical preserves the
  producer bytes rather than re-shaping.)
- **TZID parameter present, value is form #1 (local), VTIMEZONE
  resolves under ADR-0007's RRULE subset.** Resolve via
  `vstar.ParseTimeWithTZID(value, tzid, cal)`, re-emit value as
  `vstar.FormatTime(t.UTC())`, drop the TZID parameter from the
  canonical Params slice.
- **TZID parameter present, VTIMEZONE missing OR outside the v0.1
  RRULE subset.** Resolution returns `(zero, false)`. Pass the
  value AND TZID parameter through verbatim. Canonical bytes are
  NOT deterministic across calendars carrying different VTIMEZONE
  definitions in this branch — the calendar producer is expected
  to ship VTIMEZONE coverage that satisfies ADR-0007. We do NOT
  annotate the wire bytes with a marker; that would pollute the
  canonical form with implementation state.

### Nested sub-components

`STANDARD` and `DAYLIGHT` children inside a VTIMEZONE carry a
wall-clock `DTSTART` that defines the transition rule itself —
those values are deliberately NOT TZID-tagged and pass through
verbatim. This falls out naturally: `prepareComponent` recursively
prepares sub-components with the same Calendar context, but the
TZID-resolution branch only fires when the property carries a
TZID parameter, which the rule-defining DTSTART does not.

## Rationale

- **Two named entry points beat a single overloaded one.** The
  function signature itself documents whether the caller is
  promising "no TZIDs" (Component) or "use this registry"
  (ComponentInContext). Reviewers and producers cannot
  accidentally call the wrong tier — the wrong call doesn't
  exist.
- **Wave 3a callers stay green.** `Component(c)` keeps its prior
  behaviour: wrapper around an empty Calendar, no resolution.
  Existing tests (TestComponent\_DeterministicAcross100Runs,
  TestGoldenFiles, etc.) keep passing without modification.
- **Calendar threads context automatically.** Users who hold a
  full VCALENDAR call `canonical.Calendar(c)` and never need to
  reason about the two-tier split. The split only matters at the
  Component layer, where TZID resolution is genuinely
  context-dependent.
- **Failure mode is documented, not hidden.** When TZID resolution
  fails, the verbatim fallback is in the doc comment, in this
  ADR, and reproducible — producers seeing non-deterministic
  hashes have a clear lookup path to "your VTIMEZONE coverage is
  incomplete or outside ADR-0007's subset".

## Consequences

- Public API expands by one symbol (`ComponentInContext`); old
  symbols keep their signatures and semantics. No breaking change
  for Wave 3a callers.
- `prepareComponent` and `prepareProperty` internal helpers gain a
  `cal vstar.Calendar` parameter; both forward it on recursion.
- A `datetimeProperties` allow-list mirrors the existing
  `textProperties` allow-list. New datetime property names that
  surface from fixtures get added there explicitly.
- ADR-0007's VTIMEZONE RRULE subset is a hard dependency on this
  rule: producers shipping VTIMEZONE definitions outside the
  subset (multiple STANDARD/DAYLIGHT entries, non-yearly RRULE,
  RDATE, UNTIL/COUNT) will hit the verbatim-fallback branch.
- Producers SHOULD prefer UTC form #2 wire output where possible
  (V\*'s baseline guidance in spec/03 rule 5); ADR-0008 catches
  the cases where that guidance is impractical (recurring local
  events with DST).
- Spec/03 rule 5 lock at v0.1 is implemented, not deferred.

## Spec linkage

- `spec/03-canonicalization.md` rule 5 ("Datetime forms. UTC
  (`Z`-suffixed) preferred for V\*; if local time is needed, an
  explicit `TZID=` parameter referencing a VTIMEZONE in the same
  VCALENDAR.") — this ADR resolves the canonical-form treatment
  of the TZID branch.

## Implementation linkage

- Track: `vstar-go-canonical`
- Files: `canonical/canonical.go` (`ComponentInContext`,
  `Component`, `Calendar`; internal `prepareComponent`,
  `prepareProperty`, `datetimeProperties`).
- Tests: `canonical/canonical_test.go`
  - `TestComponent_VerbatimDatetimeWithoutContext`
  - `TestComponentInContext_ResolvesTZIDToUTC`
  - `TestComponentInContext_MissingVTIMEZONE`
  - `TestCalendar_ResolvesTZIDForChildren`
  - `TestCanonical_AlreadyUTCDatetime_NoChange`

## Related

- ADR-0004 (canonical component order — companion v0.1 lock)
- ADR-0005 (canonical text normalization — companion v0.1 lock)
- ADR-0006 (canonical ATTACH handling — companion v0.1 lock)
- ADR-0007 (VTIMEZONE RRULE subset — direct dependency for the
  resolution path)
