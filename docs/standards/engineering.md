# Engineering standards

Layer 4 in the precedence model. Decisions referenced by ID win over this file.

## Repository conventions
- Deployables live in `apps/<name>` (frontends) and `services/<name>` (backends). Names are kebab-case, matching the
  catalog entry and the image `harness-engineering-<name>`.
- Build-time libraries live in `packages/go/<name>` and `packages/ts/<name>`. A deployable never imports another deployable.
- Every command developers or CI run is a `make` target. Per-deployable Makefiles expose the same target names
  (`lint`, `test`, `build`, `image`).
- Host developer tools are checked for presence without version pins or a version manager. Application dependencies are pinned by lockfile (`go.sum`, `pnpm-lock.yaml`). Images are pinned by digest.
- Make targets set `GOTOOLCHAIN=local`: Go never downloads another toolchain implicitly.

## Go (default backend)
- Backend services are written in Go. Another language only when its core dependency requires it: justify it in
  the OpenSpec change that creates the service, record it as a registry decision with `scope: services/*`,
  `default_language: go` and an `exceptions` entry (path, language, reason), and set the catalog `language`.
  The service still follows the same runtime contract, layering and test policy.
- New services start from `templates/go-service` (`make new-service name=<x>`). One Go module per deployable plus
  `packages/go/*`, tied together with `go.work` for local development.
- Layout (hexagonal-lite). The template creates `cmd/service` and `internal/platform`; add the other layers with the
  first feature that needs them, never as empty folders:

  ```text
  services/<name>/
    cmd/service/main.go       # wiring only: config, adapters; `migrate` subcommand with the first database
    internal/platform/        # runtime contract: config, logs, request IDs, probes, shutdown
    internal/domain/          # entities, value objects, domain errors. stdlib only
    internal/app/             # use cases + port interfaces (repositories, publishers, clients)
    internal/adapters/        # grpc/ (Connect handlers), postgres/ (sqlc queries) implementing app ports
    migrations/               # goose SQL migrations, embedded
  ```
- Dependency rule: `domain` ← `app` ← `adapters` ← `cmd`. `domain` and `app` never import adapters, `net/http`,
  `database/sql` or generated transport code. Enforcement is planned (ARCHITECTURE.md §13).
- "Lite": no interface unless there are two implementations or a test double needs one. No DI framework.
- Default libraries, chosen at first use: `connectrpc.com/connect`, `sqlc`, `pgx/v5`, `goose`, `log/slog`,
  OpenTelemetry SDK. A product may replace one by recording a decision.
- Formatting: `gofmt`/`goimports`. Lint: `golangci-lint` with the repo config. Lint warnings fail CI.
- Errors:
  - wrap with `%w` and context
  - domain errors are typed values in `internal/domain`
  - adapters map them to gRPC/HTTP codes
  - never return raw database errors to clients
- `context.Context` is the first parameter of every I/O function. No package-level mutable state.
- Constructors take explicit dependencies (no DI framework, no `init()` side effects).
- Logging via `log/slog` (JSON handler) only. No `fmt.Print*` in service code.

## TypeScript (apps)
- `strict: true`, ESM only, Node LTS for servers. Lint: ESLint (typescript-eslint, recommended-type-checked). Format: Prettier.
- Framework and rendering mode are decided per product (Q8); this section grows with that decision.
- Package manager: pnpm workspaces (via corepack). No `any` without a comment explaining why.
- Server code follows ports and adapters: `src/domain`, `src/app`, `src/adapters`, `src/main.ts` (wiring only).
- Logging via `pino` (JSON).

## Contracts
See [`contracts/AGENTS.md`](../../contracts/AGENTS.md). Generated code is never edited by hand and is excluded from coverage.

## Data
- Migrations: forward-only versioned SQL, named `NNNN_description.sql`, reviewed like code.
- Expand/contract: add the new schema → deploy code that writes both → backfill → switch reads → remove the old schema in a later release.
- Runtime uses the `<svc>_app` role (DML only). Migrations use `<svc>_owner`.

## Runtime contract
Every deployable implements [ARCHITECTURE.md §9](../../ARCHITECTURE.md#9-service-runtime-contract-every-deployable)
(ports, probes, config, shutdown, image, resources) and the logging and correlation rules of
[§8](../../ARCHITECTURE.md#8-observability-model). Those sections are the only copy of the contract.

## Naming
- Env vars: `UPPER_SNAKE`, prefixed by concern (`DB_`, `AUTH_`, `OTEL_`). Required variables are documented in the service README.
- Proto packages: `harness_engineering.<domain>.v<N>`. Events: `<domain>.<event>.v<N>`.
- Databases and roles: `<svc>`, `<svc>_owner`, `<svc>_app`.
