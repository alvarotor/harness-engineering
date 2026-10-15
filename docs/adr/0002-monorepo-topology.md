# 0002 Monorepo topology

- Status: accepted
- Date: 2026-10-15
- Source: default

## Context
There are 4 deployables, a single team (assumed), shared contracts, and a need for atomic cross-service changes.

## Decision
- One repository. Top-level folders: `apps/` (frontends), `services/` (backends), `packages/`
  (build-time libraries only), `contracts/`, `infra/`, `e2e/`, `openspec/`, `docs/`, `templates/`,
  `tools/`, `.github/`.
- There is no separate `deploy/` folder. All environment and deployment material lives in `infra/`,
  including OpenTofu (`infra/tofu/`). Future Kubernetes manifests go in `infra/k8s/`.
- Go uses one module per deployable plus `packages/go/*`, tied together with `go.work` for local development.
  TS uses pnpm workspaces.
- A deployable may import `packages/*` and generated contract code. It may **never** import
  another deployable.
- `Makefile` is the single command entry point, for both humans and CI.

## Consequences
- Path-filtered CI is required so unrelated services are not rebuilt.
- Splitting into multiple repos later stays possible, because deployables do not import each other.

## Revisit when
Separate teams need independent access control or release processes for parts of the repo.
