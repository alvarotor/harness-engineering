# 0013 GitHub Actions with reusable, path-filtered workflows

- Status: accepted
- Date: 2026-10-15
- Source: user (GitHub Actions), default (structure)

## Decision
- `pr.yml`: detects changed deployables (their path plus the `packages/` and `contracts/` paths they
  depend on), then calls reusable workflows: `_go-service.yml`, `_node-service.yml`, `_ts-app.yml`,
  `_contracts.yml`, `_image.yml`, `_tofu.yml`, `_e2e.yml`, `_harness.yml`.
- Stages per deployable:
  1. format/lint
  2. unit tests + coverage gate (≥ 80 %)
  3. integration tests
  4. migration validation
  5. image build → Trivy → SBOM
- Repo-wide jobs:
  - gitleaks
  - `buf breaking`
  - `openspec validate`
  - `tools/harness check`
  - `docker compose config`
  - e2e on the Compose stack
  - `tofu fmt -check` / `validate` / `plan` for changed envs
- `main.yml`: push changed images to GHCR, then deploy to `dev`. `promote.yml` (manual): deploy a recorded manifest to `prod`.
  `tofu apply` is manual, per environment, with required reviewers.
- Actions are pinned by commit SHA, with least-privilege `permissions:`. Every CI step calls a `make` target, so local runs match CI.
