# V* — Conformance

> Status: **draft**. Conformance criteria firm up alongside a
> conformance test suite (planned for v0.2).

## What "V\* conformant" means

An implementation is V\* conformant when it:

1. **Emits valid RFC 5545 / RFC 6350.** Output passes a generic
   iCalendar/vCard validator.

2. **Emits the required common properties** on every component:
   `UID`, `DTSTAMP`, `X-VSTAR-HASH` (per `02-component-mapping.md`).

3. **Honors canonicalization rules** (per `03-canonicalization.md`):
   identical logical content → identical bytes → identical hash.

4. **Uses the correct component types** for the agentic concepts it
   represents (per the mapping table).

5. **Respects the extension namespace** (per `04-extensions.md`):
   no top-level new components; all extensions in `X-*`.

6. **Is append-only**: original components are never mutated;
   state changes use the supersession pattern.

## Implementation classes

V\* implementations come in three flavors:

- **Emitter**: produces V\* documents (e.g. AGR's compiler).
- **Consumer**: reads V\* documents and projects state.
- **Round-trip**: both, with byte-identical re-emit guaranteed
  for documents it emitted.

The conformance test suite (v0.2) will exercise each class
separately.

## Reference implementations

| Implementation | Class | Status | Repo |
|---|---|---|---|
| AGR (Racket) | Emitter + Consumer | In development | `hop-top/agr` |
| `vstar-ts` (TypeScript) | TBD | Planned (v0.2) | TBD |

## Self-certification

Until a formal conformance suite exists, implementations
SHOULD publish a `VSTAR-CONFORMANCE.md` documenting:

- Spec revision targeted (commit hash of this repo)
- Implementation class (emitter / consumer / round-trip)
- Known deviations + rationale
- Test artifacts (golden V\* documents + their `X-VSTAR-HASH`
  values)

This is the v0.1 honor-system substitute for a real test suite.
