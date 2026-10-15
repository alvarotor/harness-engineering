# 0003 Containers everywhere; Compose now, Kubernetes-ready

- Status: accepted
- Date: 2026-10-15
- Source: user

## Decision
- Every deployable is an OCI image built from its own multi-stage `Dockerfile`, with base images
  pinned by digest, running as non-root, and following the runtime contract in ARCHITECTURE.md §9.
- Docker Compose runs local, dev and prod. Files: `infra/compose/compose.yaml` (base) plus
  `compose.local.yaml` (build, ports, Mailpit) and `compose.deploy.yaml` (image digests, restart policy, limits).
- Images are built once in CI and promoted to later environments by digest.
- Kubernetes-readiness rules are enforced now:
  - env-only config
  - stateless containers
  - `/livez` and `/readyz` on port 8081
  - graceful SIGTERM shutdown
  - migrations as a separate command
  - no `container_name`, `links`, or host-path volumes for app state

## Consequences
- Single VM per environment: no high availability, and a few seconds of downtime per deploy (assumption A2).

## Revisit when
Any Kubernetes trigger in ARCHITECTURE.md §13 is met.
