# ADR-0005: Canonical text normalization — NFC required

## Status

Accepted (Sami, 2026-05-04)

## Date

2026-05-04

## Context

`spec/03-canonicalization.md` open question #3 (locale normalization
for text-bearing properties): two visually-identical strings can
have different byte sequences when one uses precomposed Unicode
characters and the other uses combining sequences. Without a
normalization rule, two implementations canonicalizing the same
logical content can produce different bytes — defeating the goal of
spec/03.

Two candidates surfaced:

1. **Pass-through.** Treat property values as opaque bytes; never
   normalize. Simple, lossless, but breaks set-equality across
   producers using different platforms (macOS HFS+ historically
   used NFD; most other systems use NFC).
2. **NFC required.** Normalize text-bearing property values to
   Unicode Normalization Form C (Canonical Decomposition followed
   by Canonical Composition) before encoding. The dominant form
   on the wire (W3C Working Group on Internationalization
   recommends NFC for the web; HTML5, XML, JSON, and PDF
   canonicalize toward NFC).

## Decision

**NFC required** for all text-bearing property values during
canonicalization. Specifically:

- Every property `Value` is normalized to NFC via
  `golang.org/x/text/unicode/norm.NFC` before encoding.
- Every parameter `Value` is normalized to NFC similarly.
- Property `Name` and parameter `Name` are not normalized — they
  are uppercased per RFC 5545 §3.1 / RFC 6350 §3.3 (already
  ASCII-only by spec).
- Group prefixes on vCard properties (`group.NAME`) preserve
  their case for round-trip fidelity (see codec/rfc6350) but
  their bytes are NFC-normalized as a unit.

## Rationale

Three reasons.

1. **Cross-implementation determinism.** Two implementations
   canonicalizing the same logical content MUST produce the same
   bytes. Without NFC, `é` (U+00E9) and `é` (U+0065 U+0301)
   canonicalize differently. AGR-Racket cross-validation against
   Go output requires a single normalization form; NFC is the
   universally-accepted choice.
2. **W3C/IETF alignment.** RFC 5198 ("Unicode Format for Network
   Interchange") recommends NFC for protocol text. The W3C
   character model (§4) requires "early uniform normalization" to
   NFC. JSON (RFC 8259) recommends NFC. Aligning V\* with the
   default reduces friction for downstream consumers.
3. **Stability.** NFC is closed under concatenation and idempotent
   (`NFC(NFC(x)) == NFC(x)`). The canonical form is stable
   under repeated canonicalization — a fixpoint property the
   spec's round-trip semantics depend on.

Against pass-through: technically simpler but punts the problem
to every consumer pair. The first time AGR-Racket and vstar-Go
disagree on `café`, the spec's reference-implementation cross-
validation fails, and the fix is "specify NFC". Specify it now.

## Consequences

### Dependency

This ADR introduces a new external dependency:
`golang.org/x/text/unicode/norm`. Per ADR-0003 ("No
`hop.top/kit` dependency"), `golang.org/x/...` is the documented
exception class — "stdlib-extension where the standard library is
missing functionality". Unicode normalization is not in the Go
standard library; `golang.org/x/text` is the canonical Go-team-
maintained normalization package. Adding it is consistent with
ADR-0003.

`go.mod` grows one `require` line:

```
require golang.org/x/text vX.Y.Z
```

`go.sum` grows accordingly.

### Behaviour

- Producer-supplied non-NFC text round-trips through `Canonical()`
  as NFC. The original byte form is lost. This is documented as
  intentional (the canonical form is the equivalence-class
  representative).
- An emitter that wants pass-through bytes uses
  `codec/rfc5545.Encode` directly; `Canonical()` is the
  hash-stable form.
- Codec parsers do NOT normalize on parse — they remain lossless
  with respect to the input bytes. Normalization is a
  canonicalization-time concern only.

### Performance

NFC normalization adds a per-property pass. For a typical
calendar with ~50 properties, the overhead is < 1ms on a 2024
laptop (measured in `canonical_test.go` benchmarks at task
T-0033). Acceptable for the intended use cases (interactive
agentic systems, not high-throughput streaming).

## Spec linkage

- `spec/03-canonicalization.md` open question #3 ("Locale-specific
  normalization for text-bearing properties (SUMMARY,
  DESCRIPTION) — NFC required?") — this ADR resolves the question.

## Implementation linkage

- Track: `vstar-go-canonical`
- Task: T-0033 (`Canonical(Component)` — applies NFC per rule 4
  of the canonicalization steps)
- File: `canonical.go` (`Canonical(Component) []byte`)
- Dependency: `golang.org/x/text/unicode/norm.NFC`

## Related

- ADR-0001 (monorepo structure)
- ADR-0003 (no kit dependency — defines stdlib-extension exception)
- ADR-0004 (canonical component order — companion v0.1 lock)
- ADR-0006 (canonical ATTACH handling — companion v0.1 lock)
- ADR-0007 (VTIMEZONE RRULE subset — companion carryover)
- RFC 5198 (Unicode Format for Network Interchange)
