# RFC 5545 conformance fixtures

Round-trip golden VCALENDARs used by the Go codec round-trip tests
and (later) the cross-language conformance suite per the v0.1
design doc § "fixtures grow throughout the work".

Files use **LF** line terminators on disk for diff-friendliness;
the parser is liberal on input (CRLF or LF) and the encoder always
emits CRLF on output. Round-trip semantic equality is asserted by
`TestRoundTripFixture_*` in `codec/rfc5545/`.

## Catalogue

| File | Purpose |
|------|---------|
| `empty.ics` | Calendar with PRODID + VERSION only |
| `one_vtodo.ics` | Single VTODO with UID/DTSTAMP/SUMMARY/PRIORITY |
| `nested_vtimezone.ics` | VTIMEZONE containing STANDARD + DAYLIGHT subs |
| `vevent_valarm.ics` | VEVENT containing one VALARM |
| `vjournal.ics` | Single VJOURNAL component |
| `vfreebusy.ics` | Single VFREEBUSY with two FREEBUSY periods |
| `world.ics` | VCALENDAR with three VTODOs of varying STATUS |

The `vstar-go-fixtures` track will finalise the directory layout
and register cross-language guarantees; this seed set exists so the
codec can prove round-trip fidelity today.

## Hash goldens

Each fixture has a sibling `<fixture>.hash` file containing the
content hash of its canonical form, in the format

```
sha256:<64 lowercase hex chars>
```

followed by a single LF terminator (file is exactly 72 bytes).

The hash is computed by `hashing.Calendar` (for `.ics`) or
`hashing.Card` (for `.vcf`) over the canonical byte form (see
`spec/03-canonicalization.md` and `canonical/`). The
`X-VSTAR-HASH` property — when present in the input — is stripped
before hashing per spec/03 §7.

Future TypeScript and AGR-Racket V\* implementations MUST produce
byte-identical `.hash` content for these fixtures. CI enforces
this via `TestHashGoldens` in `hashing/golden_test.go`. To
regenerate after a deliberate canonical change, run a small
generator that walks the fixtures, calls `hashing.Calendar` /
`hashing.Card`, and rewrites the `.hash` siblings.
