# RFC 6350 conformance fixtures

Round-trip golden VCARDs used by the Go codec round-trip tests
and (later) the cross-language conformance suite.

Files use **LF** line terminators on disk for diff-friendliness;
the parser is liberal on input (CRLF or LF) and the encoder always
emits CRLF on output. Round-trip semantic equality is asserted in
`codec/rfc6350/`.

## Catalogue

| File | Purpose |
|------|---------|
| `minimal.vcf` | Single FN-only VCARD |
| `kind_org.vcf` | VCARD with KIND=org |
| `kind_group.vcf` | VCARD with KIND=group + 3 MEMBER properties |
| `escaping.vcf` | VCARD exercising RFC 6350 §3.4 TEXT escaping |
| `grouped.vcf` | VCARD with property grouping |
| `with_extensions.vcf` | One X-* per spec/04 scope (vstar, system, exp) |

## Hash goldens

Each fixture has a sibling `<fixture>.hash` file containing the
content hash of its canonical form, in the format

```
sha256:<64 lowercase hex chars>
```

followed by a single LF terminator (file is exactly 72 bytes).

The hash is computed by `hashing.Card` over the canonical byte
form (see `spec/03-canonicalization.md` and `canonical/`).
The `X-VSTAR-HASH` property — when present in the input — is
stripped before hashing per spec/03 §7.

Future TypeScript and AGR-Racket V\* implementations MUST produce
byte-identical `.hash` content for these fixtures. CI enforces
this via `TestHashGoldens` in `hashing/golden_test.go`.
