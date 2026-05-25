# Set up the dev environment

Clone the repo, install the pinned Go toolchain, and run the full
CI gate locally.

## Use this when

- You are about to file your first PR against `hop-top/vstar`.
- You want to confirm your environment matches the one CI uses
  before debugging a failing job.
- You hit `command not found` on `make ci` and need the toolchain.

## Result

After completing this guide, you will:

- Have the pinned Go version active.
- Run `make ci` and see green (vet + fmt-check + lint + cover +
  fixtures-verify).
- Run `make fixtures-verify` standalone and see no drift against
  the committed corpus.

## Before you begin

You need:

- Git.
- [mise](https://mise.jdx.dev/) installed and on `$PATH` —
  `vstar/go` pins its Go toolchain via `mise.toml`.
- GNU `make` (the bundled BSD `make` on macOS works too).
- `golangci-lint` and `gofumpt` (mise resolves them on demand if
  they are listed in `mise.toml`; otherwise install via
  `go install`).

## Quick version

```sh
git clone git@github.com:hop-top/vstar.git
cd vstar/go
mise install
make ci
```

`make ci` runs vet, fmt-check, lint, cover, and fixtures-verify
end-to-end. Green means you're ready to contribute.

## Steps

### 1. Clone the repository

```sh
git clone git@github.com:hop-top/vstar.git
cd vstar
```

The repo is a multi-tree monorepo:

- `spec/` — V* specification text (CC-BY-4.0).
- Repo root — Go reference implementation (Apache-2.0).
- `ts/` — planned TypeScript reference implementation (not present
  in v0.1).
- `testdata/` — shared conformance fixtures consumed by every
  implementation.
- `docs/` — ADRs, release artifacts, contributor guides.

The split rationale lives in [ADR-0001](../adrs/0001-monorepo-structure.md).

### 2. Install the pinned Go toolchain

```sh
mise install
```

`mise` reads `mise.toml` and installs the exact Go version
`vstar/go` builds against. Confirm:

```sh
mise current go
```

Expected: a single line printing the pinned version (e.g.
`go 1.22.x`).

### 3. Build and test

```sh
make build
make test
```

Both should complete cleanly. `make test` runs with the race
detector (`go test -race ./...`); slow on first run, fast on
subsequent runs once the test cache populates.

### 4. Run the full CI gate

```sh
make ci
```

`ci` aggregates:

- `vet` — `go vet ./...`.
- `fmt-check` — `gofumpt` formatting check (no rewrite).
- `lint` — `golangci-lint run ./...`.
- `cover` — race-mode tests with coverage profile.
- `fixtures-verify` — regenerates `testdata/` canonical/hash
  siblings + fuzz seeds, then asserts no drift.

A green `make ci` is the gate every PR must pass before review.

### 5. Verify the conformance corpus

```sh
make fixtures-verify
```

This regenerates every `.canonical` and `.hash` sibling under
`testdata/` and runs `git diff --exit-code testdata/`. If your
working tree is dirty after the regeneration, the target fails —
either commit the regenerated files (because you intentionally
changed implementation behavior) or revert the implementation
change (because the drift is unintentional).

The same target also keeps the codec fuzz-seed corpora
(`codec/rfc5545/testdata/fuzz/` etc.) in sync with the canonical
source under `testdata/fuzz-seed/`. Never hand-edit the go-fuzz
corpus.

## Troubleshooting

| Symptom | Likely cause | Fix |
|---|---|---|
| `mise: command not found` | mise not installed. | Install per [mise.jdx.dev](https://mise.jdx.dev/). On macOS: `brew install mise`. |
| `mise install` succeeds but `go version` shows the wrong version | `mise activate` not in your shell init. | Add `eval "$(mise activate zsh)"` (or `bash`) to your shell rc, then re-source. |
| `make: command not found` | macOS minimal install. | Install Xcode Command Line Tools: `xcode-select --install`. |
| `golangci-lint: command not found` on `make lint` | Linter not installed. | `go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest` or pin via `mise.toml`. |
| `make fixtures-verify` shows drift on a fresh checkout | Implementation behavior diverged on `main`. | Open an issue — fixtures should be clean on `main` per CI. |
| `make ci` passes locally but fails on CI | Different Go version, OS, or linter rule version. | Compare versions: `go version` vs the GitHub Actions log. mise pin should match the workflow's. |

## How it works

`vstar/go` is a Go module with no `hop.top/kit` dependency (see
[ADR-0003](../adrs/0003-no-kit-dependency.md)). The toolchain is
mise-managed so contributors and CI run identical Go versions
without per-machine drift.

`fixtures-verify` is the cross-cutting integrity check: it ensures
the Go reference produces byte-identical canonical bytes and hashes
for every fixture, every commit. Future TypeScript and AGR-Racket
V* implementations will run an equivalent check against the same
corpus — see [How to build a sister implementation](../user/how-to-implement-vstar.md).

## Next steps

- [Add a feature track end-to-end](contributing-flow.md) — TDD
  posture, Conventional Commits, ADR filing.
- [CONTRIBUTING.md](../../CONTRIBUTING.md) — repo-wide rules
  (licensing, commit style, reviewers).
- [ADR index](../adrs/README.md) — read recent ADRs to understand
  the decision style before filing your own.
