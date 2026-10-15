# 0004 Go is the default backend language, with hexagonal-lite layout

- Status: accepted
- Date: 2026-10-15
- Source: user

## Decision
**Default:** backend services are written in Go (toolchain pinned in `go.mod`) and generated from
`templates/go-service`:

```text
services/<name>/
  cmd/<name>/main.go        # wiring only: config, adapters, `serve` and `migrate` subcommands
  internal/domain/          # entities, value objects, domain errors. stdlib only
  internal/app/             # use cases + port interfaces (repositories, publishers, clients)
  internal/adapters/
    grpc/                   # Connect handlers (gRPC) + Vanguard REST, auth interceptor
    postgres/               # sqlc-generated queries implementing app ports
    messaging/              # Redis pub/sub publisher/subscriber implementing app ports
  migrations/               # goose SQL migrations, embedded
```

Dependency rule: `domain` ← `app` ← `adapters` ← `cmd`. `domain` and `app` must not import adapters,
`net/http`, `database/sql`, Redis or generated transport code. The harness enforces this.

"Lite" means: no interface unless there are two implementations or a test double needs one. No DI framework.

Libraries:
- `connectrpc.com/connect`
- `connectrpc.com/vanguard`
- `sqlc`
- `pgx/v5`
- `goose`
- `go-redis/v9`
- `log/slog`
- OpenTelemetry SDK
- `lestrrat-go/jwx`

**Override rule:** a service may use another language only when its core dependency
requires it. The override must be:
1. justified in the OpenSpec change that creates the service
2. listed under `exceptions` of D-0004 in the registry
3. recorded in the service catalog

The service still follows the same runtime contract, the same layering (ports and adapters) and the same test policy.

## Current exceptions
- `services/auth`: Node 24 + TS, because Better Auth is a TypeScript-only library (ADR-0005).
  Generated from `templates/node-service`.
