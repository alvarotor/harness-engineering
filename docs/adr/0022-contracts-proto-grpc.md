# 0022 Protobuf contracts and gRPC between services; browser API style deferred

- Status: accepted
- Date: 2026-11-11
- Source: user
- Supersedes: 0007

## Context

ADR-0007 fixed both the service-to-service contract (protobuf + gRPC) and the browser edge (REST via
Vanguard transcoding, generated OpenAPI). The repository is now a reusable base for products that are not
chosen yet (P-0016). Service-to-service contracts are generic; how browsers call APIs depends on the product.

## Decision

- Contracts live in `contracts/proto/<domain>/v<N>/` (APIs) and `contracts/proto/<domain>/events/v<N>/`
  (events), managed with `buf` (`lint`, `breaking`, `generate`).
- Synchronous service-to-service calls use gRPC (Go: `connect-go` handlers, gRPC protocol between services).
- Breaking changes need a new package version (`v<N+1>`); the old version is served until consumers migrate.
- **Deferred (Q8):** how browsers and third parties call APIs (REST transcoding, Connect JSON from a generated
  client, or another style). Decide when the first public API is built; ADR-0007 holds the earlier analysis.

## Consequences

- No codegen for a REST edge until it is needed.
- The first public API change must resolve Q8 first.

## Revisit when

A product needs a browser-facing API (Q8), or a service cannot use gRPC.
