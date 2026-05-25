# V* — Calendar/Card-shaped data convention for agentic systems

V\* is a convention for representing **agentic-system state as iCalendar
+ vCard components**. It maps the moving parts of agent-driven work
(worlds, missions, players, turns, observations, decisions, …) onto the
stable, well-tooled component types defined in RFC 5545 (iCalendar) and
RFC 6350 (vCard) — VCALENDAR, VTODO, VEVENT, VJOURNAL, VFREEBUSY,
VTIMEZONE, VALARM, VCARD.

The result: agentic work that interoperates with calendars, schedulers,
contact directories, and any other RFC 5545/6350-aware tool — with no
proprietary serialization, no bespoke data model, and a clear
extensibility namespace (`X-*`).

## Status

**Draft / RFC.** v0.1 work-in-progress. Specification extracted from
the AGR (Agent Game Runtime) project, where V\* originated as the
runtime's emission target.

## Documentation

Two audiences, two starting points:

- **Adopters** (importing `hop.top/vstar` or building a sister
  implementation) → [`docs/user/quickstart.md`](docs/user/quickstart.md)
  for the 5-minute path; the full
  [`docs/INDEX.md`](docs/INDEX.md) lists how-tos and reference.
- **Contributors** (changing vstar itself) →
  [`CONTRIBUTING.md`](CONTRIBUTING.md) for repo-wide rules;
  [`docs/dev/setup.md`](docs/dev/setup.md) and
  [`docs/dev/contributing-flow.md`](docs/dev/contributing-flow.md)
  for the dev loop.

Specification text lives under [`spec/`](spec/) (CC-BY-4.0).
Architecture decisions live under [`docs/adrs/`](docs/adrs/).

## Why this exists as its own repo

V\* started inside AGR (`agr/docs/specs/vstar-mapping.md`). It's
extracted here because:

- The convention is reusable beyond AGR — any agent-protocol or
  orchestration system can emit V\*
- Conformance must not be coupled to one runtime's release cycle
- Future implementations (TypeScript, Go, Python, …) need a stable
  reference point
- Foundation conversations are easier with an independent spec

## What's in this repo

```
vstar/
├── README.md             — this file
├── INDEX.md              — guide to spec sections
├── GLOSSARY.md           — V\* terminology
├── LICENSE.md            — CC-BY-4.0
└── spec/
    ├── 01-overview.md
    ├── 02-component-mapping.md
    ├── 03-canonicalization.md
    ├── 04-extensions.md
    └── 05-conformance.md
```

Examples directory will appear once the v0.1 surface stabilizes.

## Implementations

| Implementation | Status                          | Module / Path                          |
|----------------|---------------------------------|----------------------------------------|
| Go (reference) | v0.1.0-alpha prerelease channel (tag `vstar/v${version}`) | `hop.top/vstar` |
| TypeScript     | planned (post-Go v0.1.0)        | `vstar-ts` — separate repo, TBD        |
| Racket (AGR)   | downstream emitter, cross-validates against Go output (non-blocking) | in `hop-top/agr` |

The Go implementation is the v0.1 reference. Conformance is
defined by the corpus + canonical/hash siblings in
[`testdata/`](testdata/), regenerated and checked by `make
fixtures-verify` at the repo root.

## Relationship to AGR

AGR is the first V\*-emitting runtime. Its compiler converts
Racket-described worlds into V\* component sequences appended to a
ledger. AGR's `docs/specs/vstar-mapping.md` now points at this spec
as the canonical reference; AGR-internal extensions live in the
`X-AGR-*` namespace.

## Relationship to AGNTCY TS SDK

The AGNTCY TS SDK is a separate downstream consumer candidate — agents
participating in AGNTCY can emit V\* as their canonical action log.
Not yet wired; tracked separately.

## License

CC-BY-4.0. See `LICENSE.md`.
