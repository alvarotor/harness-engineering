# AGENTS.md (__NAME__)

Scoped rules for this service. They add to the root [AGENTS.md](../../AGENTS.md) and
[services/AGENTS.md](../../services/AGENTS.md), which always apply.

## Layout
- `cmd/service`: wiring only (config, logger, routes, lifecycle). Exempt from unit tests and coverage.
- `internal/platform`: runtime contract (config, logs, request IDs, probes, shutdown). Tested, ≥ 80 % coverage.
- `internal/adapters` → `internal/app` → `internal/domain`: add these when the first feature needs them. Do not create
  empty layers.
  - `domain` and `app` never import adapters, `net/http`, `database/sql` or generated transport code.
  - Adapters map domain errors to HTTP/gRPC codes and never return raw database errors to clients.

## Rules
- Add a route through the layers, not in `cmd/service`. Keep `context.Context` as the first parameter of I/O functions.
- Logs go through the request logger (`platform.LoggerFrom(ctx)`). Never log secrets, tokens or personal data.
- Config changes: add the variable to `platform.LoadConfig` (validated, `_FILE` support is automatic) and to the README.
- Do not change the ops port, the probe paths, the 20 s shutdown limit or the non-root user (D-0003, D-0009).
- A database (D-0008) adds a `migrate` subcommand and `migrations/` (versioned SQL, expand/contract), run by the image,
  never at startup. It uses the `<svc>_app` role at runtime and `<svc>_owner` for migrations.
- Tests: `make -C services/__NAME__ test` must pass. Non-trivial new logic needs its own test.
- Compose: `compose.yaml` keeps `read_only`, `cap_drop: [ALL]`, `no-new-privileges` and `traefik.enable: "false"` unless the
  security review makes the service public.
