# services/AGENTS.md

Backend services. Adds to the root [`AGENTS.md`](../AGENTS.md).

- **One service per business domain** (D-0014). Before adding an entity, check
  [`docs/catalog/services.yaml`](../docs/catalog/services.yaml). If the entity belongs to an existing domain, add it
  there. Otherwise propose a new service through OpenSpec, stating the domain, the owned data, and why it is not part of an existing domain.
- **Language:** Go by default (D-0004). Another language requires a registry exception with a reason.
  Current exception: `auth` (Node, Better Auth).
- **Layout (Go):** `cmd/<name>` → `internal/adapters` → `internal/app` → `internal/domain`.
  `domain` and `app` never import adapters, `net/http`, `database/sql`, Redis or generated transport code.
  Non-Go services follow the same ports-and-adapters split.
- **Data** (D-0008): own database only, on the shared server. Separate `<svc>_owner` and `<svc>_app` roles. Versioned SQL
  migrations in `migrations/`, run by the image's `migrate` command, never at startup. Changes are expand/contract.
- **Communication** (D-0007, D-0015): sync = gRPC through `contracts/proto`; async = Redis pub/sub
  behind `Publisher`/`Subscriber` ports, loss-tolerable events only. Never call another service's database.
- **Runtime contract:** [`docs/standards/engineering.md`](../docs/standards/engineering.md#runtime-contract).
- **Required files per service:** `Dockerfile`, `AGENTS.md`, `README.md`, plus `go.mod` (Go) or `package.json` (Node). Checked by R-CAT-004.
- **Tests:** unit tests for all code, coverage ≥ 80 %, integration tests with Testcontainers (D-0017).
- **Templates** (`make new-service`) arrive with the first service of each kind. Until then, follow this file and the standards.
