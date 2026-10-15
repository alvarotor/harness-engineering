# 0007 Protobuf contracts, gRPC for sync calls, REST edge via transcoding

- Status: accepted
- Date: 2026-10-15
- Source: user

## Decision
- Single source of truth: `contracts/proto/<domain>/v<N>/*.proto` (APIs) and
  `contracts/proto/<domain>/events/v<N>/*.proto` (events), managed with `buf`
  (`buf lint`, `buf breaking --against main`, `buf generate`).
- **Synchronous service-to-service calls use gRPC.** Go services implement handlers with `connect-go`,
  which also serves gRPC-Web and Connect. Clients between services use the gRPC protocol.
- **Edge:** `vanguard-go` exposes REST/JSON from `google.api.http` annotations on the same handlers and port.
  Browsers and external clients use REST only.
- OpenAPI is generated from the protos into `contracts/openapi/`: committed, never hand-edited, and used
  for the TS client and docs. The owner confirmed the generator supports `google.api.http`
  annotations; the exact plugin version is pinned in `buf.gen.yaml`.
- Breaking changes need a new package version (`v2`). The old version keeps being served until consumers migrate.
- Asynchronous communication: ADR-0015.
