# Add a feature track end-to-end

Land a new feature in `vstar/go` from RED test through GREEN
implementation, ADR (when architectural), and conformance fixture
updates.

## Use this when

- You are about to ship a non-trivial change to `vstar/go` —
  anything that adds a public symbol, changes wire-form behavior,
  or touches the canonical/hash discipline.
- Your change crosses the spec/implementation boundary (touches
  both `spec/` and the Go reference or `testdata/`).
- You need to file an ADR alongside a code PR.

## Result

After completing this guide, you will:

- Have a feature branch with TDD-phase commits (RED → GREEN →
  REFACTOR).
- Pass `make ci` and `make fixtures-verify` locally.
- Land the change with a coordinated PR covering spec, code, and
  fixtures together.

## Before you begin

You need:

- A working dev environment per [setup](setup.md).
- A clear statement of the user-visible behavior change. If you
  cannot describe it in one sentence, the scope is wrong.
- For architectural changes: a draft ADR. See [ADR index](../adrs/README.md)
  for the standard four-section template (Status / Context /
  Decision / Consequences).

## Quick version

```sh
git checkout -b feat/<area>-<short-summary>

# RED commit
go test ./<pkg> # fails
git add <pkg>/<feature>_test.go
git commit -m "test(<area>): <feature> — RED"

# GREEN commit
go test ./<pkg> # passes
git add <pkg>/<feature>.go
git commit -m "feat(<area>): <feature> — GREEN"

# REFACTOR commit (if cleanup needed)
git commit -am "refactor(<area>): tidy <feature> — REFACTOR"

# Coordinated update if fixtures or spec move
make fixtures-verify  # regenerates testdata/, fails on drift
git add testdata/ spec/
git commit -m "feat(<area>): regenerate corpus + spec amendment"

make ci
gh pr create --base main
```

## Steps

### 1. Choose a Conventional Commit type

`vstar` uses [Conventional Commits](https://www.conventionalcommits.org/).
Allowed types (per [CONTRIBUTING.md](../../CONTRIBUTING.md)):

| Type | Use for |
|---|---|
| `feat` | New user-visible functionality. |
| `fix` | Bug fix. |
| `docs` | Docs-only change. |
| `style` | Formatting only, no code change. |
| `refactor` | Restructure without behavior change. |
| `perf` | Performance improvement. |
| `test` | Adding or revising tests. |
| `build` | Toolchain, deps, build scripts. |
| `ci` | GitHub Actions, CI infrastructure. |
| `chore` | Release plumbing, housekeeping. |

Optional scope is the area: `feat(rfc5545): ...`, `fix(canonical): ...`,
`ci(go): ...`. Use the imperative mood ("add", not "added"). Keep
the subject under 72 chars.

### 2. Write the RED test first

Every behavioral change lands as a sequence of TDD-phase commits.
The first commit is the failing test:

```sh
git add <pkg>/<feature>_test.go
git commit -m "test(<area>): <feature> — RED"
```

The commit message ends in `— RED`. The tree fails CI on this
commit by design. CI does not gate intermediate commits — it gates
the merge.

If the change is pure config or docs (no testable behavior), skip
RED and go straight to GREEN.

### 3. Implement the minimum to pass — GREEN

```sh
git add <pkg>/<feature>.go
git commit -m "feat(<area>): <feature> — GREEN"
```

Resist scope creep. Only add code the RED test exercises. Anything
else lands as a separate sequence (or a follow-up REFACTOR commit).

Run the package's tests to confirm green:

```sh
go test ./<pkg>
```

### 4. REFACTOR (optional)

If the GREEN implementation is messy — repeated branches, awkward
naming, helpers that want extracting — clean up in a third commit:

```sh
git commit -am "refactor(<area>): tidy <feature> — REFACTOR"
```

Tree must still pass. Commit message ends in `— REFACTOR`.

### 5. File the ADR for architectural changes

Significant design choices land as ADRs in
[`docs/adrs/`](../adrs/) using the standard four-section template:

```
# ADR-NNNN: <decision title>

## Status

Proposed | Accepted (date)

## Context

<what forces this decision>

## Decision

<what we will do>

## Consequences

<positive + negative outcomes>
```

Open the ADR PR before — or alongside — the code PR that depends
on the decision. Read recent ADRs (0004–0009) for tone and depth
exemplars. Number sequentially from the highest filed ADR; do not
reuse numbers.

If your change is purely additive (new helper, new optional flag,
new fixture without semantic shift), no ADR is required.

### 6. Update the conformance corpus

If your change affects canonical bytes or hash output for any
existing fixture, regenerate the corpus:

```sh
make fixtures-verify
```

The target regenerates every `.canonical` and `.hash` and then
runs `git diff --exit-code testdata/`. If your working tree is
dirty, the target fails — commit the regenerated files in the
same PR:

```sh
git add testdata/ codec/rfc5545/testdata/fuzz/ codec/rfc6350/testdata/fuzz/
git commit -m "feat(<area>): regenerate corpus for <change>"
```

If the change is intentional, the regeneration commit is part of
the PR. If unintentional, revert your implementation change — you
broke a canonical-form invariant.

### 7. Coordinate spec + fixture + implementation

When the change touches the spec/implementation boundary, the PR
must move all three together:

1. Add or modify a fixture under `testdata/`.
2. Update the relevant section of `spec/` to describe the new
   behavior.
3. Update each implementation (Go reference, `ts/` once it exists) to
   handle it; add tests that consume the new fixture.

Splitting these across PRs leaves implementations divergent from
the spec — refused at review.

### 8. Run the full CI gate locally

```sh
make ci
```

Green locally is a prerequisite — CI runs the same target.

### 9. Open the PR

```sh
gh pr create --base main \
  --title "feat(<area>): <one-line summary>" \
  --body "$(cat <<'EOF'
## Summary

<2-3 bullets covering the user-visible change>

## Test plan

- [ ] make ci
- [ ] make fixtures-verify
- [ ] <any feature-specific exercise>

## ADR

Closes ADR-NNNN (or: no ADR — additive change).
EOF
)"
```

Reviewers per [CONTRIBUTING.md](../../CONTRIBUTING.md):

- **Sami** — engineering lead; reviews all releases and ADRs.
- **Reza** — primary contributor on the Go reference; reviews Go PRs.
- **Theo** — future owner of `ts/`; reviews when that tree lands.

## Troubleshooting

| Symptom | Likely cause | Fix |
|---|---|---|
| `make fixtures-verify` shows drift you didn't expect | Implementation behavior accidentally changed canonical bytes. | `git diff testdata/` to see the byte-level difference; revert the implementation change unless the drift is the point of the PR. |
| RED commit fails CI and you cannot push | CI is running per-commit, not per-merge. | Check `.github/workflows/go.yml` — vstar's CI gates the merge, not intermediate commits. If your fork enforces per-commit, squash before pushing. |
| Conventional Commit linter rejects your subject | Subject exceeds 72 chars or uses past tense. | Rewrite imperative + concise: "add X" not "added X with a long explanation". |
| Reviewer asks for an ADR on a change you thought was additive | The change touches the canonical/hash discipline or adds a public sentinel. | File the ADR — public sentinels and canonical-form rules are architectural. See [ADR-0009](../adrs/0009-rrule-parsing-scope.md) as an example of a scope-bounding ADR. |
| `gofumpt` fails fmt-check | Formatter version mismatch with CI. | Pin via `mise.toml` or run `gofumpt -w .` and commit. |

## How it works

The TDD posture (RED/GREEN/REFACTOR) is named in commit messages
so reviewers and `git log --oneline` readers can see the discipline
without opening diffs. Squash-on-merge collapses the trio to a
single commit on `main`, but the original sequence is preserved on
the feature branch.

`make fixtures-verify` is the cross-cutting integrity check.
Implementations may freely refactor as long as canonical bytes
stay byte-identical; any drift surfaces as a hard CI failure with
the diff visible. This is what keeps Go, future TypeScript, and
AGR-Racket cross-validators agreeing forever.

ADRs are the mechanism for moving past "is this a refactor or a
decision?" Anything that constrains a future contributor — wire
formats, error sentinels, package boundaries, scope cuts — goes in
an ADR. Read [ADR-0009](../adrs/0009-rrule-parsing-scope.md) for
an example of a scope-bounding decision with an explicit
"Rejected" + "Amendments" trail.

## Next steps

- [Set up the dev environment](setup.md) — toolchain prerequisites
  if you skipped past it.
- [CONTRIBUTING.md](../../CONTRIBUTING.md) — full repo-wide rules
  (licensing, reviewers, security disclosure).
- [ADR index](../adrs/README.md) — exemplars to model your own
  decision records on.
