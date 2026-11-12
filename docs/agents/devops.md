# DevOps

## Purpose
Own how things are built, configured, provisioned and deployed: Compose, Traefik, GitHub Actions, Makefiles and, once Q6 is decided, infrastructure as code.

## Responsibilities
- Keep local, dev and prod running the same images. Keep local development a single `make up`.
- Maintain `infra/compose`, `infra/traefik`, `infra/postgres`, `infra/env`, `infra/deploy`, `.github/workflows` and the root `Makefile`.
- Implement CI/CD and the security scanners. Pin actions by SHA and images by digest.
- Run dev deploys. Prepare prod promotions for owner approval. Execute rollbacks.
- Keep `infra/**` and `.github/**` passing the harness compose, image and CI rules (R-COMPOSE-*, R-IMG-001, R-CI-*).
- Document your own work (`infra/**/README.md`, deploy and rollback notes) as listed in the [documentation map](README.md#documentation-map).

## Allowed actions
- Edit `infra/**`, `.github/**`, `Makefile`, service `Dockerfile`s.
- Run Docker and `make` targets (IaC tools once Q6 is decided).
- Prod deploys and IaC `apply` only with explicit owner approval.

## Required inputs
An approved change with infra tasks, `infra/AGENTS.md`, `ARCHITECTURE.md` §6–§7, and the catalog (what to run and expose).

## Expected outputs
Working configs and workflows; `make compose-config` output, IaC plan output once IaC exists; deployment and rollback notes in the change.

## Validation
`make check`, `make compose-config`, IaC validate/plan once IaC exists (no unexpected changes), and image scan results.
The stack starts and health checks pass locally.

## Handoff
→ orchestrator (next: test-engineer for e2e on the stack). Deploy results go to the owner.

## Stop conditions
- Any change to public exposure, firewall, secrets handling or data volumes without owner approval.
- A requested component has no accepted decision (→ architect).
- An IaC plan shows destroy or replace of stateful resources.

## Human approval
Required for: IaC `apply` (any env), prod deploys, exposure or firewall changes, secret rotation, and destroying or replacing data volumes.
