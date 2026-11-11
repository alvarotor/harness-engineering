# contracts/AGENTS.md

API and event contracts. Adds to the root [`AGENTS.md`](../AGENTS.md).

- **Source of truth** (D-0022): protobuf only.
  - APIs: `proto/<domain>/v<N>/`
  - events: `proto/<domain>/events/v<N>/`
  - Managed with `buf` (`lint`, `breaking --against main`, `generate`). Proto packages: `har.<domain>.v<N>`.
- **Browser-facing API style** (REST transcoding, Connect JSON, other) is not decided: the first public API needs Q8.
- **Breaking changes are not allowed in place.** Add `v<N+1>`, keep serving `v<N>` until consumers have migrated, then remove it in a
  separate change.
- **Events:** the first event needs Q7 (messaging) answered. Names `<domain>.<event>.v<N>`, past tense; metadata
  `event_id`, `occurred_at`, `request_id`, `traceparent`.
- **Ownership:** only the owning domain service changes its package. Consumers propose changes through OpenSpec.
- **No PII in event payloads** beyond IDs unless the change justifies it (personal-data inventory in the catalog).
