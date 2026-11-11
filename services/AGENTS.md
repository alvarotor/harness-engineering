# services/AGENTS.md

Backend services. Adds to the root [`AGENTS.md`](../AGENTS.md).

- **One service per business domain** (D-0014). Before adding an entity, check
  [`docs/catalog/services.yaml`](../docs/catalog/services.yaml). If the entity belongs to an existing domain, add it
  there. Otherwise propose a new service through OpenSpec (lane L), stating the domain, the owned data, and why it
  is not part of an existing domain. The first service needs the product answer (Q6).
- **Create services with `make new-service name=<x>`** (from `templates/go-service`). Never hand-copy another service.
- **Language:** Go by default (D-0004). Another language requires a registry exception with a reason.
- **Layout (Go):** `cmd/<name>` → `internal/adapters` → `internal/app` → `internal/domain`.
  `domain` and `app` never import adapters, `net/http`, `database/sql` or generated transport code.
- **Data** (D-0008): own database only, on the shared server. Separate `<svc>_owner` and `<svc>_app` roles. Versioned SQL
  migrations in `migrations/`, run by the image's `migrate` command, never at startup. Changes are expand/contract.
- **Communication** (D-0022): sync = gRPC through `contracts/proto`. Async messaging is not decided: the first event
  between services needs Q7 answered. Never call another service's database.
- **Runtime contract:** [`docs/standards/engineering.md`](../docs/standards/engineering.md#runtime-contract).
- **Required files per service:** `Dockerfile`, `AGENTS.md`, `README.md`, `compose.yaml`, plus `go.mod` (Go). Checked by the harness.
- **Tests:** unit tests for all code, coverage ≥ 80 % (`make -C services/<x> test`), integration tests with Testcontainers (D-0017).
