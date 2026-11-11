# Testing standard

Decision: [ADR-0017](../adr/0017-testing-strategy.md). Summary of what every change must satisfy:

| Layer | Required | Tooling | Gate |
|---|---|---|---|
| Unit | **all hand-written code**: every Go package has `_test.go`; every TS module has a sibling `*.test.ts(x)` | `go test -race`; the frontend stack's runner (Q11) | coverage **≥ 80 %** per deployable and per `packages/*` library |
| Integration | adapters that touch Postgres, other data stores or HTTP | Testcontainers | must pass on PRs touching the deployable |
| Contract | every proto change | `buf breaking --against main`; generated clients compile | blocks merge |
| Migration | every migration | apply to an empty DB and to a DB at the previous release | blocks merge |
| E2E | every user-facing flow, **when possible** | Playwright in `/e2e` on the Compose stack; emails read from Mailpit | blocks merge for covered flows; when not possible, the reason goes in the change's `tasks.md` |

## Rules
- **Exempt from the unit-test-presence rule:** generated code, type-only files, pure wiring entrypoints (`main.go`, `main.ts(x)`).
  Exempt files are also excluded from coverage. Nothing else is excluded.
- **Coverage thresholds** may be raised per deployable but never lowered without a superseding ADR.
- **Deterministic tests:** no real external network, no sleeps for synchronization, inject clocks and ID generators, no order dependence.
- **Test names describe behavior:** `TestPlaceOrder_RejectsEmptyCart`, `it("rejects an empty cart")`.
- **Fixtures** live next to the tests (`testdata/` in Go). No shared mutable fixtures.
- **Bug fixes** start with a failing test that reproduces the bug.

## Current coverage-gated code
| Module | Command |
|---|---|
| `tools/harness` | `make harness-test` |
