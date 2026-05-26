# V\*

Calendar/Card-shaped data convention for agentic systems.

[![Latest tag](https://img.shields.io/github/v/tag/hop-top/vstar?filter=vstar/*&label=release&color=00ADD8&sort=semver)](https://github.com/hop-top/vstar/releases)
[![CI](https://img.shields.io/github/actions/workflow/status/hop-top/vstar/go.yml?branch=main&label=ci)](https://github.com/hop-top/vstar/actions/workflows/go.yml?query=branch%3Amain)
[![Spec](https://img.shields.io/badge/spec-draft%20v0.1-blue)](spec/)
[![Code](https://img.shields.io/badge/code-Apache--2.0-green)](LICENSE)
[![Spec license](https://img.shields.io/badge/spec-CC--BY--4.0-lightgrey)](LICENSE.md)

V\* represents agentic-system state — worlds, missions, players, turns,
observations, decisions — as **iCalendar (RFC 5545) + vCard (RFC 6350)
components**. Agent work that already has time, identity, and sequence
semantics gets to ride existing calendar/scheduler/contact tooling
instead of a bespoke protocol.

## Use this when

- You emit agent state and want it to interop with calendars,
  schedulers, or contact directories with no custom serialization.
- You need a stable reference point so multiple runtimes (Go, TS,
  Python, Racket, …) can produce byte-identical output.
- You want a published spec, conformance fixtures, and a reference
  implementation — not a vendored file inside someone else's runtime.

**Skip this if:** you need a streaming wire protocol, or your data
has no time / identity / sequence semantics.

## Status

| Layer | State |
|-------|-------|
| Spec | Draft v0.1 — breaking changes allowed pre-1.0 |
| Go reference (`hop.top/vstar`) | Alpha — `vstar/v0.1.0-alpha.0` |
| TypeScript reference | Planned, separate repo |
| Racket emitter (in `hop-top/agr`) | Downstream, cross-validates against Go |

Spec is the source of truth; conformance lives in
[`testdata/`](testdata/).

## Quick start (Go)

```sh
go get hop.top/vstar
```

```go
import (
    "strings"

    "hop.top/vstar/codec/rfc5545"
)

cal, err := rfc5545.Parse(strings.NewReader(input))
```

Full walkthrough (parse + hash + walk components):
[`docs/user/quickstart.md`](docs/user/quickstart.md).

## Where to go next

| You are… | Start here |
|----------|------------|
| **Adopting `hop.top/vstar`** | [`docs/user/quickstart.md`](docs/user/quickstart.md) → [`docs/INDEX.md`](docs/INDEX.md) for how-tos and reference |
| **Building a sister implementation** (TS / Python / …) | [`spec/05-conformance.md`](spec/05-conformance.md) + [`testdata/`](testdata/) + [`docs/user/how-to-implement-vstar.md`](docs/user/how-to-implement-vstar.md) |
| **Changing vstar itself** | [`CONTRIBUTING.md`](CONTRIBUTING.md) → [`docs/dev/setup.md`](docs/dev/setup.md) → [`docs/dev/contributing-flow.md`](docs/dev/contributing-flow.md) |
| **Understanding a design choice** | [`docs/adrs/`](docs/adrs/) |

## Why this exists as its own repo

V\* started inside AGR (`agr/docs/specs/vstar-mapping.md`) and was
extracted because:

- The convention is reusable beyond AGR — any agent-protocol or
  orchestration system can emit V\*.
- Conformance must not be coupled to one runtime's release cycle.
- Future implementations (TS, Python, …) need a stable reference
  point.
- Foundation conversations are easier with an independent spec.

## Related projects

- **AGR** — origin runtime and first downstream consumer. Its
  compiler converts Racket-described worlds into V\* component
  sequences. AGR-internal extensions live in `X-AGR-*`.
- **AGNTCY TS SDK** — candidate consumer; agents can emit V\* as
  their canonical action log. Not yet wired.

## Contributing

See [`CONTRIBUTING.md`](CONTRIBUTING.md) for repo-wide rules
(licensing, commit conventions, TDD posture, shared `testdata/`
discipline) and [`docs/dev/`](docs/dev/) for the dev loop.

Security reports: until `SECURITY.md` lands, see
[`CONTRIBUTING.md`](CONTRIBUTING.md) for the private-report path.

## License

- **Code** — Apache-2.0, see [`LICENSE`](LICENSE)
- **Spec text** (`spec/`) — CC-BY-4.0, see [`LICENSE.md`](LICENSE.md)
