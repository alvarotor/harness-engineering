# Engineering standards

Layer 4 in the precedence model. Decisions referenced by ID win over this file.

## Repository conventions
- Deployables live in `apps/<name>` (frontends) and `services/<name>` (backends). Names are kebab-case, matching the
  catalog entry and the image `har-<name>`.
- Build-time libraries live in `packages/go/<name>` and `packages/ts/<name>`. A deployable never imports another deployable (D-0002).
- Every command developers or CI run is a `make` target. Per-deployable Makefiles expose the same target names
  (`lint`, `test`, `build`, `image`).
- Host developer tools are checked for presence without version pins or a version manager (D-0021). Application dependencies are pinned by lockfile (`go.sum`, `pnpm-lock.yaml`). Images are pinned by digest.

## Go (default backend, D-0004)
- Layout and dependency rule: [ADR-0004](../adr/0004-go-hexagonal-lite-backends.md).
- Formatting: `gofmt`/`goimports`. Lint: `golangci-lint` with the repo config. Lint warnings fail CI.
- Errors:
  - wrap with `%w` and context
  - domain errors are typed values in `internal/domain`
  - adapters map them to gRPC/HTTP codes
  - never return raw database errors to clients
- `context.Context` is the first parameter of every I/O function. No package-level mutable state.
- Constructors take explicit dependencies (no DI framework, no `init()` side effects).
- Logging via `log/slog` (JSON handler) only. No `fmt.Print*` in service code.

## TypeScript (apps, `services/auth`)
- `strict: true`, ESM only, Node 24 LTS for servers. Lint: ESLint (typescript-eslint, recommended-type-checked). Format: Prettier.
- Package manager: pnpm workspaces (via corepack). No `any` without a comment explaining why.
- Server code follows ports and adapters: `src/domain`, `src/app`, `src/adapters`, `src/main.ts` (wiring only).
- Logging via `pino` (JSON).

## Contracts (D-0007, D-0015)
See [`contracts/AGENTS.md`](../../contracts/AGENTS.md). Generated code is never edited by hand and is excluded from coverage.

## Data (D-0008)
- Migrations: forward-only versioned SQL, named `NNNN_description.sql`, reviewed like code.
- Expand/contract: add the new schema → deploy code that writes both → backfill → switch reads → remove the old schema in a later release.
- Runtime uses the `<svc>_app` role (DML only). Migrations use `<svc>_owner`.

## Runtime contract
Every deployable implements this ([ARCHITECTURE.md §9](../../ARCHITECTURE.md#9-service-runtime-contract-every-deployable)):

| Item | Rule |
|---|---|
| Ports | app `8080`, ops `8081` (`GET /livez`, `GET /readyz`) |
| Config | env vars, validated at startup, fail fast with the variable name; `<NAME>_FILE` for file secrets |
| Shutdown | SIGTERM → readiness false → drain → exit ≤ 20 s |
| Logs | JSON stdout with `ts level msg service version env request_id trace_id`; never secrets/tokens/emails |
| Correlation | read or generate `X-Request-ID`; propagate it and `traceparent` on HTTP, gRPC and event metadata |
| Telemetry | OpenTelemetry SDK; export only when `OTEL_EXPORTER_OTLP_ENDPOINT` is set |
| Image | multi-stage, base pinned by digest, UID 65532, `HEALTHCHECK` on 8081, no shell in Go images |
| State | stateless; data only in Postgres/Redis |

## Naming
- Env vars: `UPPER_SNAKE`, prefixed by concern (`DB_`, `AUTH_`, `OTEL_`). Required variables are documented in the service README.
- Proto packages: `har.<domain>.v<N>`. Events: `<domain>.<event>.v<N>`.
- Databases and roles: `<svc>`, `<svc>_owner`, `<svc>_app`.
