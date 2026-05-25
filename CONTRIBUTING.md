# Contributing to vstar

Thanks for considering a contribution. This document covers the
shared rules across `spec/`, the Go reference implementation at the
repo root, and the future `ts/` tree. Per-language details live
alongside the code (see [`README.md`](README.md)).

## What you're contributing to

- [`spec/`](spec/) — V\* specification text, CC-BY-4.0.
- Go reference implementation at the repo root, Apache-2.0
  (see [ADR-0002](docs/adrs/0002-go-library-license.md)).
- [`ts/`](ts/) — TypeScript reference implementation (planned;
  not present at v0.1.0).
- [`testdata/`](testdata/) — shared conformance fixtures
  consumed by every implementation.
- [`docs/`](docs/) — ADRs and design plans.

The repository structure and rationale are in
[ADR-0001](docs/adrs/0001-monorepo-structure.md). The v0.1.0 design
brief was retired post-release; the surviving decision record lives
in [`docs/adrs/`](docs/adrs/).

## Licensing

- Code (Go reference at repo root, `ts/`) ships under **Apache-2.0**. Every `.go` file
  begins with:

  ```go
  // SPDX-License-Identifier: Apache-2.0
  ```

- Spec text (`spec/`) ships under **CC-BY-4.0** — leave the
  existing notice intact.
- **Inbound = outbound.** By opening a PR you agree your
  contribution is licensed under the same terms as the file you're
  touching. **No DCO sign-off, no CLA** — at least until V\*
  enters a foundation and the rules change. (Apache-2.0 §5
  governs.)

## Commit conventions

We use [Conventional Commits](https://www.conventionalcommits.org/).
Allowed types:

| Type | Use for |
|------|---------|
| `feat` | new user-visible functionality |
| `fix` | bug fix |
| `docs` | docs-only change (README, ADR, this file) |
| `style` | formatting (no code change) |
| `refactor` | restructure without behaviour change |
| `perf` | performance improvement |
| `test` | adding or revising tests |
| `build` | toolchain, deps, build scripts |
| `ci` | GitHub Actions, CI infrastructure |
| `chore` | release plumbing, housekeeping |

Optional scope is the area, e.g. `feat(rfc5545): ...`,
`fix(canonical): ...`, `ci: ...`. Use the imperative mood
("add", not "added"). Keep the subject under 72 chars.

Commit often, then squash before merge if the history is noisy.

## TDD posture (RED / GREEN / REFACTOR)

Every behavioural change lands as a sequence of TDD-phase commits:

1. **RED** — write the failing test first. Commit message ends in
   `— RED`. Tree fails CI.
2. **GREEN** — minimum implementation to pass. Commit message ends
   in `— GREEN`. Tree passes CI.
3. **REFACTOR** — tidy the code without expanding scope. Commit
   message ends in `— REFACTOR`. Tree still passes.

Pure config or docs changes (no test-able behaviour) skip the
RED commit and go straight to GREEN. The smoke test in the
scaffolding track demonstrates the full RED → GREEN flow.

## Architecture decisions

Significant design choices land as ADRs in
[`docs/adrs/`](docs/adrs/) using the standard four-section
template (Status / Context / Decision / Consequences).

Open an ADR PR before — or alongside — a code PR that depends on
the decision. Existing ADRs:

- [ADR-0001](docs/adrs/0001-monorepo-structure.md) — monorepo
  layout
- [ADR-0002](docs/adrs/0002-go-library-license.md) — Go library
  license (Apache-2.0)
- [ADR-0003](docs/adrs/0003-no-kit-dependency.md) — `hop.top/vstar`
  takes no `hop.top/kit` dependency

## No `hop.top/kit` dependency

`hop.top/vstar` is a library. It depends only on the Go standard
library and selected `golang.org/x/...` packages. **No
`hop.top/kit` import.** PRs that introduce a non-stdlib
non-`golang.org/x` dependency require an explicit reviewer sign-off
and an updated ADR.

See [ADR-0003](docs/adrs/0003-no-kit-dependency.md) for the
rationale.

## Shared `testdata/` corpus

Conformance fixtures live at the repository root in
[`testdata/`](testdata/), not under any single implementation.
This keeps Go, TypeScript, and any cross-validation harness
running against byte-identical inputs.

When you change a fixture, the spec text and at least one
implementation must move with it in the **same PR**. The
expected pattern:

1. Add or modify a fixture under `testdata/`.
2. Update the relevant section of `spec/` to describe the new
   behaviour.
3. Update each implementation (Go reference, `ts/` once it exists)
   to handle it; add tests that consume the new fixture.

Splitting these across PRs leaves implementations divergent
from the spec — refused at review.

## Per-language conventions

- **Go**: see [`README.md`](README.md) for build/test/lint
  invocations, Go toolchain pin (mise), formatter (gofumpt), and
  linter set (golangci-lint).
- **TypeScript**: TBD when `ts/` lands.

## Filing an issue

GitHub issues are the right place for spec ambiguities, bugs,
and feature requests. For private security reports, see
`SECURITY.md` (TBD; until then, email the owners listed in
ADR-0001).

## Reviewers

- **Sami** — engineering lead; reviews all releases and ADRs.
- **Reza** — primary contributor on the Go reference; reviews Go PRs.
- **Theo** — future owner of `ts/`; reviews when that tree lands.
