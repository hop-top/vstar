# V\* conformance corpus

This directory is the cross-language conformance corpus for the V\*
specification. Every implementation (Go now; future TypeScript;
future AGR-Racket cross-validation) consumes these fixtures
identically — same input, same canonical bytes, same hash. CI
asserts the property; drift fails the PR.

## Directory layout

| Subdir            | Format          | Purpose                                              |
|-------------------|-----------------|------------------------------------------------------|
| `rfc5545/`        | `.ics`          | VCALENDAR happy-path round-trip goldens.             |
| `rfc6350/`        | `.vcf`          | VCARD happy-path round-trip goldens.                 |
| `supersession/`   | `.ics`          | Supersession edge cases (linear, multi-step, cross-component, corrupt). |
| `malformed/`      | `.ics` / `.vcf` | Inputs that MUST fail to parse with a documented sentinel. |
| `fuzz-seed/`      | `.bytes`        | Minimal-but-diverse byte sequences seeding the codec fuzz targets. |

The `rfc5545/` and `rfc6350/` directories were established by
earlier waves (codec/canonical/hashing tracks) and are kept as the
canonical layout for happy-path corpora. New categories live in
their own subdirs to keep the surface easy to navigate.

## Fixture quartet convention

Every fixture is a quartet of sibling files sharing a stem `<name>`:

| File                | Required            | Content                                                      |
|---------------------|---------------------|--------------------------------------------------------------|
| `<name>.ics` / `.vcf` | yes               | The fixture input. LF on disk for diff-friendliness.         |
| `<name>.canonical`  | parseable inputs    | Canonical bytes from `canonical.Calendar` or `canonical.Card`. |
| `<name>.hash`       | parseable inputs    | `sha256:<64 hex>` + LF (exactly 72 bytes) from `hashing.Calendar` / `hashing.Card`. |
| `<name>.notes.md`   | optional            | Human prose describing the fixture's shape and purpose.      |
| `<name>.error`      | malformed inputs    | Name of the expected `vstar.Err*` sentinel (one per line).   |

Fixtures under `malformed/` carry **no** `.canonical` or `.hash`
because they fail to parse. They carry `.error` instead.

Files use **LF** line terminators on disk for diff-friendliness;
the parser is liberal on input (CRLF or LF) and the encoder always
emits CRLF on output. Round-trip semantic equality is asserted by
codec tests; byte-identity of `.canonical` and `.hash` is asserted
by `make fixtures-verify` (see CI parity below).

## Implementation-class promise

For every fixture in `rfc5545/`, `rfc6350/`, and `supersession/`
that has a `.canonical` and `.hash` sibling:

- An **emitter** (build a Calendar/Card via the API and encode it)
  MUST produce bytes that, when canonicalized, match `.canonical`
  exactly and hash to `.hash` exactly.
- A **consumer** (parse the input, canonicalize, hash) MUST produce
  identical `.canonical` bytes and `.hash` content.
- A **round-trip** (parse, encode, parse again) MUST yield a
  semantically equal Calendar/Card.

This is enforced for the Go reference implementation by
`make fixtures-verify` and by `TestHashGoldens` in
`hashing/golden_test.go`. Future TypeScript and AGR-Racket V\*
implementations MUST produce byte-identical `.canonical` and
`.hash` content for these fixtures.

## Migration policy

Changing a fixture is a **coordinated PR** across:

1. Spec text (the `spec/` change that motivates the new behavior).
2. Implementation (the codec / canonical / hashing change).
3. Fixture quartet (regenerated `.canonical` and `.hash`).

Run `make fixtures-verify` at the repo root. The target regenerates every
`.canonical` and `.hash` and then runs `git diff --exit-code
testdata/`. If the working tree is dirty, the target fails — commit
the regenerated files in the same PR as the implementation change.

Renaming a fixture is also a coordinated PR; the stem change ripples
to every sibling.

## CI parity

`.github/workflows/go.yml` runs `make fixtures-verify` after the
test step. A green run proves:

- The Go implementation reproduces every `.canonical` byte-for-byte.
- The Go implementation reproduces every `.hash` byte-for-byte.

Drift in either direction fails the PR. The expected reaction is
"investigate root cause" — either the implementation drifted
(revert the change) or the spec/encoding intentionally changed
(regenerate, commit, document).

## Per-directory READMEs

- [`rfc5545/README.md`](rfc5545/README.md) — VCALENDAR happy paths.
- [`rfc6350/README.md`](rfc6350/README.md) — VCARD happy paths.
- [`supersession/README.md`](supersession/README.md) — append-only ledger edge cases.
- [`malformed/README.md`](malformed/README.md) — sentinel-bearing parse failures.
- [`fuzz-seed/README.md`](fuzz-seed/README.md) — fuzz-target byte seeds.

## Fuzz seed convention

`fuzz-seed/rfc5545/seed_*.bytes` and `fuzz-seed/rfc6350/seed_*.bytes`
are the **canonical source** for the codec fuzz seed corpora. Each
file is the raw byte input; no go-fuzz wrapper. The matching
go-fuzz corpus directories live at:

- `codec/rfc5545/testdata/fuzz/FuzzParse_RFC5545/`
- `codec/rfc6350/testdata/fuzz/FuzzParse_RFC6350/`

These are kept in sync via `make fixtures-verify`, which also
copies any new `.bytes` file into the matching go-fuzz format
wrapper. Update the canonical source under `fuzz-seed/` and rerun
the target; never hand-edit the go-fuzz corpus.
