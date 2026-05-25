# ADR-0006: Canonical ATTACH handling — reference-only (URI)

## Status

Accepted (Sami, 2026-05-04)

## Date

2026-05-04

## Context

`spec/03-canonicalization.md` open question #2: how does V\* handle
the `ATTACH` property (RFC 5545 §3.8.1.1)? RFC 5545 permits two
forms:

1. **Inline (BINARY).** The attachment payload is base64-encoded
   into the property value with `VALUE=BINARY` and
   `ENCODING=BASE64` parameters. Self-contained but explodes
   document size.
2. **Reference (URI).** The property value is a URI; the
   attachment lives wherever the URI points. Compact but
   introduces dangling-link risk.

Two candidates for V\* v0.1:

1. **Both forms supported on emit.** Maximum RFC fidelity.
   Producers choose the form per attachment.
2. **Reference-only on emit.** Inline (BINARY) is read-only —
   parsed, round-tripped on `Encode`, but not written by
   `Canonical()` or any V\* helper. Matches `crm/internal/vcal`
   precedent (binary-passthrough on read; no inline writing).

## Decision

**Reference-only on emit** for v0.1.

- `Canonical(Component)` emits ATTACH properties with `VALUE=URI`
  semantics: the property `Value` is the URI; no `VALUE=BINARY`
  on output.
- ATTACH properties with `VALUE=BINARY` on input round-trip
  through the codec layer losslessly (the parser preserves them
  byte-for-byte; the encoder emits them verbatim).
- `Canonical(Component)` on a Component containing a
  `VALUE=BINARY` ATTACH **strips the `VALUE=BINARY` parameter and
  any `ENCODING=BASE64` parameter** and treats the value as a
  URI. If the value is not a syntactically-valid URI, the
  property is preserved as-is (the canonical form is best-effort
  for malformed inputs).
- A future v0.2 may add inline-emit if a real consumer requires
  it. v0.1 explicitly does not.

## Rationale

Three reasons.

1. **Hash stability.** Inline (BINARY) ATTACH defeats the value
   proposition of `X-VSTAR-HASH`: the hash includes the
   attachment payload, so a 5MB PDF in an ATTACH means a 5MB
   hash input. Reference-only keeps hashes small and meaningful.
2. **Consumer parity.** `crm/internal/vcal` already implements
   reference-only on emit. AGR L2 (the other major V\* consumer
   for v0.1) produces references, not inline payloads. Aligning
   V\* with consumer behaviour reduces friction.
3. **Foundation-track readiness.** Inline binary in a structured-
   data spec invites scope creep (encoding negotiation,
   compression, content-type signaling). Reference-only keeps
   V\* squarely in the "structured metadata about resources"
   lane and out of the "transport for binary blobs" lane.

Against both-forms-supported: the "maximum RFC fidelity"
benefit is theoretical; in practice no V\* consumer has surfaced
a need for inline-emit. Adding it later (v0.2) is cheap; emitting
it speculatively is expensive (hash bloat, transport bloat,
spec ambiguity around base64 padding canonical form).

## Consequences

- `Canonical(Component)` strips `VALUE=BINARY` and
  `ENCODING=BASE64` parameters from any ATTACH property before
  encoding. Documented in `canonical.go`.
- ATTACH-with-`VALUE=BINARY` on input round-trips losslessly
  through `codec/rfc5545.Parse → codec/rfc5545.Encode`. The
  codec layer is intentionally separate from canonicalization;
  pass-through fidelity is preserved at that layer.
- `Validate(Component)` (separate track: vstar-go-validate) MAY
  warn on `VALUE=BINARY` ATTACH properties as non-conformant
  for V\* emit. v0.1 emits a soft warning, not a hard error.
- v0.1 conformance fixtures contain only `VALUE=URI` ATTACH
  examples (or no ATTACH). v0.2 may add a `VALUE=BINARY` round-
  trip fixture once the inline-emit decision is reopened.

## Spec linkage

- `spec/03-canonicalization.md` open question #2 ("Handling of
  binary attachments (`ATTACH`) — inline vs reference") — this
  ADR resolves the question for v0.1.

## Implementation linkage

- Track: `vstar-go-canonical`
- Task: T-0033 (`Canonical(Component)` — applies the
  ATTACH-strip rule during the canonical-form construction)
- File: `canonical.go` (`Canonical(Component) []byte`)

## Related

- ADR-0001 (monorepo structure)
- ADR-0004 (canonical component order — companion v0.1 lock)
- ADR-0005 (canonical text normalization — companion v0.1 lock)
- ADR-0007 (VTIMEZONE RRULE subset — companion carryover)
- RFC 5545 §3.8.1.1 (ATTACH property)
