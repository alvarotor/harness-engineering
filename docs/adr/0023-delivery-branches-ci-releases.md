# 0023 Delivery: trunk-based branches, CI, dev → prod by git tag

- Status: accepted
- Date: 2026-11-11
- Source: user
- Supersedes: 0010, 0013

## Context

ADR-0010 tied deployment to an OpenTofu-provisioned VM and ADR-0013 listed CI jobs for tools and services
that do not exist yet. Hosting and infrastructure-as-code are now deferred (Q9). The owner wants a generic
CI/CD template that works before the GitHub repository and servers exist and grows with the product.

## Decision

- **Branches:** short-lived `change/<name>` branches into `main`, one OpenSpec change per branch (lane S may
  use `fix/<name>`). `main` is always deployable.
- **Pull request (`pr.yml`):** `make check` plus `make -C <deployable> lint test image` for every deployable that
  changed (its folder, or `templates/`, `packages/`, `contracts/` for all). Every step calls a `make` target.
- **Merge to `main` (`main.yml`):** build every deployable image once, tag it with the git SHA, push to
  `ghcr.io/<github.repository_owner>/har-<name>`, then deploy to `dev` automatically.
- **Release (`release.yml`):** pushing a tag `vX.Y.Z` that points to a commit on `main` deploys the **same image
  digests** to `prod` after owner approval (GitHub Environment `prod` with required reviewers).
  Rollback = re-run the release of the previous tag.
- **Environments:** `dev` and `prod` only for now. Each is a Linux host running Docker Compose, reached
  over SSH. Host address and SSH key come from GitHub Environment variables and secrets; while they are
  unset, deploy jobs skip themselves so CI stays green.
- **Deploy steps:** upload the Compose files and an env file (mode `0600`) with secrets and image digests →
  `docker compose pull` → `up -d --wait` → smoke test.
- Actions are pinned by commit SHA (checked by the harness) with least-privilege `permissions:`.
- **Deferred (Q9):** how hosts are created (OpenTofu, cloud provider), DNS and TLS domain (Q3), staging.

## Consequences

- Prod always runs images already built and deployed to dev; nothing is rebuilt for prod.
- Building every image on each merge is simple but slows down with many deployables.

## Revisit when

A staging environment is needed, builds get slow (build changed images only), more than one host per
environment, or Kubernetes (ARCHITECTURE.md §13).
