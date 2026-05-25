# V\* validate — diagnostic code catalog

This file enumerates every `Diagnostic.Code` emitted by the
`hop.top/vstar/validate` package.

## Stability

Codes are part of the library's public surface. They are **stable
across Go library minor versions** per semver:

- A given Code never changes its meaning.
- A code is never recycled for an unrelated rule.
- New codes may be added in any release; consumers must tolerate
  unknown codes (and log them rather than crash).
- Renaming a code is a major-version event.

`Diagnostic.Message` is human-readable and may evolve across minor
versions; do not match it programmatically. Use Code instead.

## Severity

| Severity            | Meaning                                                     |
|---------------------|-------------------------------------------------------------|
| `SeverityError`     | MUST violation per spec/05 — document is not V\* conformant |
| `SeverityWarning`   | SHOULD violation or stylistic concern                       |

## Codes

### §1 — required common properties (spec/05 §1)

| Code   | Severity | Rule                                                |
|--------|----------|-----------------------------------------------------|
| VS001  | Error    | Required common property `UID` is missing.         |
| VS002  | Error    | Required common property `DTSTAMP` is missing.     |
| VS003  | Error    | Required common property `X-VSTAR-HASH` is missing.|

### §2 — X-VSTAR-HASH integrity (spec/05 §2)

| Code   | Severity | Rule                                                            |
|--------|----------|-----------------------------------------------------------------|
| VS010  | Error    | `X-VSTAR-HASH` is present but does not match the recomputed hash. |

VS010 fires only when the hash property is **present and wrong**.
When it is absent entirely, VS003 fires instead.

### §3 — extension namespace compliance (spec/05 §3, spec/04)

| Code   | Severity | Rule                                                                      |
|--------|----------|---------------------------------------------------------------------------|
| VS020  | Warning  | Property name is not on the RFC 5545/6350 allow-list and lacks the `X-` prefix. |

The standard property allow-list is sourced from RFC 5545 §3.7-§3.8
and RFC 6350 §6. See `validate/standard_properties.go`. The
sister `vstar-go-extensions` track will publish a
`ext.ScopeOf(name)` helper; once it lands, a follow-up PR will
collapse the inline allow-list onto that shared API. The
inline check is intentional during Wave 4 (sister-package
imports are forbidden across the parallel branches).

### §4 — supersession discipline (spec/05 §4)

| Code   | Severity | Rule                                                                                    |
|--------|----------|-----------------------------------------------------------------------------------------|
| VS030  | Error    | A supersession `VJOURNAL` (CATEGORIES contains `status-supersession`) is missing a required property: `RELATED-TO` or `X-VSTAR-EFFECTIVE-STATUS`. |
| VS031  | Error    | A supersession `VJOURNAL`'s `RELATED-TO` does not resolve to any component in the same Calendar (orphan supersession). |

VS030 fires from both `Validate` and `ValidateComponent` (it is a
component-local check). VS031 requires cross-component resolution,
so it fires from `Validate` only — `ValidateComponent` (single-
component entry) silently skips it because it has no ledger to
resolve against.

### §5 — type-specific required properties (spec/05 §5)

| Code   | Severity | Rule                                                                                |
|--------|----------|-------------------------------------------------------------------------------------|
| VS040  | Error    | `VTODO` requires `DUE`, OR `STATUS=COMPLETED` paired with `COMPLETED`.              |
| VS041  | Error    | `VEVENT` requires `DTSTART`.                                                        |
| VS042  | Error    | `VFREEBUSY` requires `DTSTART` AND `DTEND`.                                         |
| VS043  | Error    | `VCARD` (modeled as `Component{Type: "VCARD"}`) requires `VERSION` AND `UID`.       |

### §6 — RRULE conformance (ADR-0009)

| Code   | Severity | Rule                                                                                |
|--------|----------|-------------------------------------------------------------------------------------|
| VS050  | Warning  | `RRULE` value parses but uses a feature outside the v0.2 rrule scope (FREQ=SECONDLY/MINUTELY, RSCALE — see [ADR-0009](adrs/0009-rrule-parsing-scope.md) and its BYSETPOS/BYWEEKNO/BYYEARDAY amendment). |
| VS051  | Error    | `RRULE` value is malformed per RFC 5545 §3.3.10 (missing FREQ, INTERVAL≤0, both UNTIL+COUNT, BYMONTHDAY=0, UNTIL not in form #2, etc.). |

VS050 is a Warning because the property still round-trips through
the codec layer; only its recurrence semantics are inaccessible
to the v0.2 evaluator. Consumers using `rrule.NextOccurrence` MUST
check for VS050 before relying on the result.

VS051 is an Error because a malformed RRULE means no consumer
(vstar or otherwise) can evaluate the recurrence correctly.

## Path syntax

`Diagnostic.Path` is a dotted component/property locator.

Examples:

| Path                                 | Meaning                                  |
|--------------------------------------|------------------------------------------|
| `VCALENDAR`                          | Calendar-level diagnostic.               |
| `VCALENDAR.VTODO[uid=foo]`           | Component-level diagnostic on a VTODO.   |
| `VCALENDAR.VTODO[uid=foo].DTSTAMP`   | Property-level diagnostic on its DTSTAMP. |
| `VCALENDAR.VTODO[#3]`                | UID-less VTODO at positional index 3.    |
| `VTODO[uid=foo].DTSTAMP`             | `ValidateComponent` (no calendar prefix).|
