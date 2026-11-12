# services/AGENTS.md

Backend services. Adds to the root [`AGENTS.md`](../AGENTS.md).

- **One service per business domain**. Before adding an entity, check
  [`docs/catalog/services.yaml`](../docs/catalog/services.yaml). If the entity belongs to an existing domain, add it
  there. Otherwise propose a new service through OpenSpec (lane L), stating the domain, the owned data, and why it
  is not part of an existing domain. The first service needs the product answer (Q1).
- **Create services with `make new-service name=<x>`** (from `templates/go-service`). Never hand-copy another service.
- **Language:** Go by default. Another language requires a registry exception with a reason.
- **Layout (Go):** `cmd/service` → `internal/adapters` → `internal/app` → `internal/domain`.
  `domain` and `app` never import adapters, `net/http`, `database/sql` or generated transport code.
- **Data**: own database only, on the shared server. Separate `<svc>_owner` and `<svc>_app` roles. Versioned SQL
  migrations in `migrations/`, run by the image's `migrate` command, never at startup. Changes are expand/contract.
- **Communication**: sync = gRPC through `contracts/proto`. Async messaging is not decided: the first event
  between services needs Q4 answered. Never call another service's database.
- **Runtime contract:** [`ARCHITECTURE.md` §9](../ARCHITECTURE.md#9-service-runtime-contract-every-deployable) and the logging rules in §8.
- **Required files per service:** `Dockerfile`, `AGENTS.md`, `README.md`, `compose.yaml`, plus `go.mod` (Go). Checked by the harness.
- **Tests:** unit tests for all code, coverage ≥ 80 % (`make -C services/<x> test`), integration tests with Testcontainers.
