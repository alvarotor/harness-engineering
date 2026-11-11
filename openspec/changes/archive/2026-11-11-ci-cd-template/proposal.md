# Proposal

## Why

The repository has no CI. The owner wants branch-based delivery that deploys dev and prod and grows with the
product (ADR-0023). It must work before the GitHub repository and the servers exist, so the workflows are written
and validated now, and the deploy steps skip until the environments are configured.

## Lane and security triggers

- **Lane: L** (changes `.github/` and `infra/`; new deploy path; new CI secrets handling).
- Security triggers (`docs/standards/security.md`): **exposure** (SSH from CI to the hosts), **containers** (images
  built, published and run), **supply chain** (GitHub Actions, base and third-party images, Go modules), **secrets
  and config** (GitHub Environments, deploy secrets, workflow `permissions:`, the deploy script). The owner approves
  the change and the security-reviewer reviews it before the reviewer step.

## What Changes

- **Pull requests:** `.github/workflows/pr.yml` runs repository validation and `make compose-config`. For each
  changed deployable it runs lint, test and image through `_deployable.yml` (no publishing), plus an image scan.
  A secret scan of the git history and a Go vulnerability check (`make vuln`) run on every pull request. A report of
  third-party image vulnerabilities runs too; it is not gated (owner decision).
- **Merges to main:** `main.yml` builds every deployable image once through `_publish.yml`, scans it (fixable HIGH
  and CRITICAL block), and publishes it to GHCR tagged with the raw commit SHA. Only `_publish.yml` holds
  `packages: write`, and only `main.yml` calls it. Then `_deploy.yml` deploys the dev environment.
- **Releases:** `release.yml` runs on a tag `vX.Y.Z`. Its verify job checks, before any approval prompt, that the
  tagged commit is on `main`. The prod deploy then waits for owner approval through the GitHub Environment `prod`.
  After approval, the deploy resolves the digest of each published image from the commit SHA tag and fails closed if
  one is missing. Rollback is a re-run of the release for an earlier tag.
- **Reusable workflows:** `_deployable.yml` (input `path`: lint, test, image, image scan), `_publish.yml` (input
  `path` and `tag`: the same plus publish), `_deploy.yml` (input `ref`: commit SHA; environment `dev` or `prod`).
- **Skip rule:** inside the deploy job, the deploy steps are skipped, and the workflow succeeds, while the
  environment's `DEPLOY_HOST` variable is unset. Environment protection applies when the job starts, so an
  unconfigured prod still shows the approval prompt before the skip.
- **SSH host keys pinned (P-0018):** once a host is configured, the deploy reads `DEPLOY_KNOWN_HOSTS` and uses
  `StrictHostKeyChecking=yes`. Without it the deploy fails before sending anything.
- **Secrets as files:** each `DEPLOY_ENV_FILE` entry is written to its own `0600` file under `~/har/secrets` on the
  host. The release env file holds digests and non-secret values only. Secrets never go on the command line.
- **Pinned and least-privilege workflows:** actions pinned by commit SHA (R-CI-001); every workflow declares
  `permissions:` and no `run:` script interpolates an expression (R-CI-002).
- **Docs:** `infra/README.md` describes the compose files, variables and secrets, the GitHub settings checklist,
  the approval gate, the owner-approved local exception, and rollback.

## Capabilities

### New Capabilities

- `delivery`: pull requests are checked and scanned, merges publish once and deploy to dev, tags release to prod after
  approval, deploy host keys are pinned, deploy secrets are files, and workflows are pinned.

## Impact

- Decisions: D-0023 (delivery model, ADR-0023), P-0018 (SSH host keys pinned), D-0012 (security baseline, secrets
  handling in deploys), D-0009 (no secrets in git; the edge proxy is the only exposed container), P-0003 (ask the owner
  before hard-to-reverse or security-sensitive steps: prod deploys and secrets).
- Files: `.github/workflows/{pr,main,release,_deployable,_publish,_deploy}.yml`, `infra/compose/compose.deploy.yaml`,
  `infra/deploy/deploy.sh`, `tools/harness/internal/deploy`, `Makefile` (`deployables`, `deploy`, `secrets-scan`,
  `vuln`, `image-scan`, `infra-images`), `tools/harness/internal/check`, `docs/standards/validation.md`,
  `infra/README.md`, `infra/AGENTS.md`, `docs/agents/devops.md`.
- Owner approval required before any prod environment, secret or host is configured in GitHub; this change adds none.
- No application code and no product feature changes. No new infrastructure service (P-0002): the deploy target is
  the existing Compose stack on a host.
