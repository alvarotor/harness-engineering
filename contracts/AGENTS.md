# contracts/AGENTS.md

API and event contracts. Adds to the root [`AGENTS.md`](../AGENTS.md).

- **Source of truth** (D-0007): protobuf only.
  - APIs: `proto/<domain>/v<N>/`
  - events: `proto/<domain>/events/v<N>/`
  - Managed with `buf` (`lint`, `breaking --against main`, `generate`).
- **REST edge:** every public RPC has a `google.api.http` annotation. `openapi/` is generated, committed, and never hand-edited.
- **Breaking changes are not allowed in place.** Add `v<N+1>`, keep serving `v<N>` until consumers have migrated, then remove it in a
  separate change.
- **Events** (D-0015):
  - Name: `<domain>.<event>.v<N>`, past tense (e.g. `orders.order_placed.v1`).
  - Metadata: `event_id`, `occurred_at`, `request_id`, `traceparent`.
  - The change that adds an event must state why losing it is acceptable (Redis pub/sub is at-most-once).
- **Ownership:** only the owning domain service changes its package. Consumers propose changes through OpenSpec.
- **No PII in event payloads** beyond IDs unless the change justifies it (personal-data inventory in the catalog).
