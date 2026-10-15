# 0014 One service per business domain

- Status: accepted
- Date: 2026-10-15
- Source: user

## Decision
- A **new business domain** (bounded context such as catalog, orders, billing) means a **new service**:
  - its own database on the shared Postgres server
  - its own proto package
  - generated from the default template (`make new-service`)
- A **new entity** that belongs to an existing domain (e.g. `OrderLine` in orders) stays in that domain's service.
- `services/core` holds only the first domain and is renamed once that domain is named (assumption A5).
- The OpenSpec change that introduces a domain must state: name, owned entities, owned data, the
  contracts it exposes or consumes, and why it is not part of an existing domain.

## Consequences
- More deployables to operate as the product grows (watch the Kubernetes trigger).
- Cross-domain workflows cannot use database transactions; they go through gRPC or events.
- Rejected: one service per entity (chatty calls, no transactional consistency inside a domain).
