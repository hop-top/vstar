# V* documentation

Navigation hub for the `docs/` tree. Use the section that matches what
you came here to do.

## For adopters

You are importing `hop.top/vstar` into your project, OR
building a sister implementation (TypeScript, Racket, Python, …) that
must agree byte-for-byte with the Go reference.

- [Quickstart](user/quickstart.md) — parse your first VCALENDAR in
  five minutes.
- [How-to: validate and hash a Calendar](user/how-to-validate-and-hash.md) —
  compute `X-VSTAR-HASH`, run `validate.Validate`, interpret diagnostic
  codes.
- [How-to: parse and evaluate RRULE](user/how-to-recurrence.md) —
  v0.2 recurrence parser + `NextOccurrence` evaluator with the
  common patterns (DAILY, BYDAY, BYSETPOS).
- [How-to: build a sister vstar implementation](user/how-to-implement-vstar.md) —
  cross-validate against the Go reference using the conformance
  corpus and `X-VSTAR-HASH` algorithm.

## Reference

Audience-agnostic — any reader can link in.

- [Diagnostic code catalog (`VS001`–`VS051`)](validate-codes.md) —
  every `validate.Diagnostic.Code` the Go reference emits.
- [Specification](../spec/) — V* normative text, CC-BY-4.0.
- [Conformance corpus](../testdata/) — fixtures shared by every
  implementation.

## For developers

You are contributing to vstar itself (Go reference, future TypeScript,
spec edits).

- [Setup the dev environment](dev/setup.md) — clone, mise install,
  `make ci`, `make fixtures-verify`.
- [Add a feature track end-to-end](dev/contributing-flow.md) —
  TDD posture, Conventional Commits, ADR filing, conformance
  fixtures.

## Architecture and decisions

ADRs (Architecture Decision Records) capture the reasoning behind
each binding choice. Read them when you need the *why* behind a
canonical-form rule, an extension namespace constraint, or a parser
scope boundary.

- [ADR index](adrs/README.md)
- [ADR-0004 — canonical component order](adrs/0004-canonical-component-order.md)
- [ADR-0005 — text NFC normalization](adrs/0005-canonical-text-normalization.md)
- [ADR-0006 — ATTACH handling](adrs/0006-canonical-attach-handling.md)
- [ADR-0007 — VTIMEZONE RRULE subset](adrs/0007-vtimezone-rrule-subset.md)
- [ADR-0008 — canonical datetime context](adrs/0008-canonical-datetime-context.md)
- [ADR-0009 — generic RRULE parsing scope (v0.2)](adrs/0009-rrule-parsing-scope.md)

## Release process

Releases are cut automatically by [release-please](https://github.com/googleapis/release-please);
see [`.github/release-please-config.json`](../.github/release-please-config.json) and
[`CHANGELOG.md`](../CHANGELOG.md) once the first release lands.
