# 0017 Testing strategy: unit tests for all code, ≥ 80 % coverage, e2e when possible

- Status: accepted
- Date: 2026-10-15
- Source: user (coverage, unit, e2e), default (tools)

## Decision
- **Unit tests for all hand-written code.**
  - Go: every package with non-generated `.go` files has `_test.go` files.
  - TS: every source module has a sibling `*.test.ts(x)`.
  - Exempt: type-only files, generated code, entrypoints (`main.go`, `main.tsx`, `index.ts` that only wires).
  - The harness enforces this.
- **Coverage ≥ 80 %** (lines, and branches where the tool reports them) per deployable and per
  `packages/*` library, with generated code excluded. CI fails below the threshold. The threshold
  may be raised per service but never lowered without a superseding ADR.
- Tools:
  - Go: `go test -race -coverprofile`
  - TS: Vitest with v8 coverage and `@testing-library/preact`
  - Integration: Testcontainers (Postgres, Redis)
- **Contract tests:** `buf breaking` for APIs and events. Generated clients must compile in CI.
- **Migration tests:** apply to an empty database and to a database at the previous release.
- **E2E:** Playwright in `e2e/` against the full Compose stack. The magic-link login reads emails from the
  Mailpit API. Every user-facing flow gets an e2e test when possible. When not possible, the reason
  is recorded in the OpenSpec change's `tasks.md`.
- Tests are deterministic: no real network, no sleeps for synchronization, no shared mutable fixtures.
