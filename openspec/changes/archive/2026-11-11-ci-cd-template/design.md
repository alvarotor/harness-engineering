# Design

## Approach

Delivery follows ADR-0023 and P-0018; this design maps them onto files. Every workflow step calls a `make` target,
so the same commands run locally and in CI. Setup steps only install pinned tools.

## Workflows

| File | Trigger | Jobs |
|---|---|---|
| `pr.yml` | pull request | `check` (`make check`, `make compose-config`); `security` (`make secrets-scan` with full history, `make vuln`); `infra-images` (`make image-scan` for each third-party image, `continue-on-error`, report only); `changes` (`make deployables BASE=<base>`, all of them when `templates/`, `packages/` or `contracts/` change); `deployable` (`_deployable.yml` per changed path). |
| `_deployable.yml` | `workflow_call`, input `path` | `make -C <path> lint test image`, then `make image-scan` on the local image. Permissions `contents: read`. No publish. |
| `main.yml` | push to `main` | `list` (`make deployables ALL=1`); `build` (`_publish.yml` per deployable, `packages: write` granted here only); `deploy-dev` (`_deploy.yml`, `environment: dev`, `ref: <commit-sha>`). |
| `_publish.yml` | `workflow_call`, inputs `path`, `tag` | `make -C <path> lint test image` with the GHCR image name, `make image-scan` before any push, login, push tagged with the raw commit SHA. |
| `release.yml` | tag `v*.*.*` | `verify`: fetch with depth 0, fail unless the tagged commit is an ancestor of `origin/main`, before any approval. `deploy-prod`: `_deploy.yml`, `environment: prod`, `ref: ${{ github.sha }}`, after `verify`. |
| `_deploy.yml` | `workflow_call`, inputs `environment`, `ref` | Environment on the job, so the approval gate applies at job start. Steps are skipped while `DEPLOY_HOST` is unset. Otherwise: fail closed without `DEPLOY_KNOWN_HOSTS`; resolve each digest from `:<ref>` with `docker buildx imagetools` and fail when one is missing; write the release env file (digests, `ENV`, `DOMAIN`, `ACME_EMAIL`); write each `DEPLOY_ENV_FILE` entry to a `0600` file; run `infra/deploy/deploy.sh` over SSH; remove the runner's copies with an `always()` step. |

Images are promoted by digest: prod receives the digests published when `main.yml` built the commit, never a rebuild
(D-0003). Images carry the raw git SHA; `<ref>` in `_deploy.yml` is that SHA.

## Why the skip is inside the job

GitHub applies environment protection when a job that names the environment starts. A job-level `if` on the host
variable would skip the job before the approval prompt, so an unconfigured prod would never ask the owner. A step-level
skip keeps the prompt: an unconfigured prod shows the prompt, then skips the deploy steps, and the workflow succeeds.

## Why the digest check is after approval

The verify job only checks that the commit is on main. The digest lookup is part of the deploy, and the owner's
approval must be the first gate for everything the deploy does (the deploy job holds `packages: read` and the SSH
key). So the lookup runs inside the approved job. It fails closed on a missing image or digest, before the release
env file is written or the host is contacted.

## Permissions

- Top-level `permissions:` in every workflow (R-CI-002); jobs widen only where needed.
- `packages: write` only in `main.yml`'s build caller and therefore only in `_publish.yml`. `packages: read` for the
  deploy. `contents: read` everywhere else.
- Scans run with `contents: read`. Third-party image findings are reported and not gated (owner decision).

## Environments and secrets

- GitHub Environments `dev` and `prod`. `DEPLOY_HOST`, `DEPLOY_USER` and `DEPLOY_KNOWN_HOSTS` are variables;
  `DEPLOY_SSH_KEY`, `DEPLOY_ENV_FILE` (one `NAME=value` per line) are secrets.
- `prod` requires reviewers (owner). This change creates no environment, host or secret; the owner configures them when
  the hosts exist (Q9). `infra/README.md` holds the checklist.
- Each `DEPLOY_ENV_FILE` entry goes to its own `0600` file under `~/har/secrets` on the host. The Compose base mounts
  only the Postgres password from a `secrets:` entry; the other secret files are available to services that add their
  own `secrets:` entries.
- Secrets never go on the command line. The release env file holds digests and non-secret values only.

## SSH host keys (P-0018)

- `deploy.sh` writes `DEPLOY_KNOWN_HOSTS` to a file and runs `ssh` and `scp` with `-o StrictHostKeyChecking=yes`
  and `-o UserKnownHostsFile=<file>`.
- When the host is set and `DEPLOY_KNOWN_HOSTS` is empty, the script exits non-zero before any connection or copy.
- Tests: Go tests in `tools/harness/internal/deploy` run by `make harness-test` (with a stub `ssh`/`scp`), and an
  integration check against a real `sshd` with a pinned key.

## Security

- Deploy credentials and CI access to hosts make this a security trigger (`docs/standards/security.md`: exposure,
  secrets and config, supply chain, containers). The security-reviewer reviews this change before the reviewer step.
- Actions pinned by commit SHA (R-CI-001); every workflow declares `permissions:`, and no `run:` script interpolates an
  expression (R-CI-002).
- Scans: secret scan of the full history and Go vulnerability check on every pull request; image scan before any
  publish and on each changed deployable in pull requests. Third-party image report is non-blocking.
- `actionlint` and `gitleaks` ran clean on these workflows (evidence in tasks).

## Alternatives considered

- **One workflow with every job:** rejected; reusable workflows keep deployable and deploy logic in one place as the
  number of deployables grows.
- **Job-level skip on the host variable:** rejected; it hides the approval prompt for prod (see above).
- **Digest lookup in the verify job:** rejected; it would run registry calls before the owner's approval.
- **Rebuild images for prod:** rejected; promotion by digest is the point of the release (ADR-0023).
- **Deploy through a platform (Kubernetes, managed PaaS):** deferred until the trigger in ADR-0023 (Q9, P-0002).
