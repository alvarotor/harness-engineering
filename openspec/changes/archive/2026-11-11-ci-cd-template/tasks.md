## 1. Workflows and deploy (devops)

- [x] 1.1 Add `.github/workflows/{pr,main,release,_deployable,_publish,_deploy}.yml` per ADR-0023. `pr.yml` runs `make check`, `make compose-config`, the secret scan and `make vuln`, and per changed deployable `_deployable.yml`; `main.yml` builds through `_publish.yml` (packages write, called only from `main.yml`) and deploys dev; `release.yml` verifies the tagged commit is on main before approval, then deploys prod with environment `prod`. Each step calls a `make` target; top-level `permissions:` are declared. Verify: each workflow parses; `actionlint` passes.
  Evidence: `actionlint` clean on all workflows; `make check` green (R-CI-001 and R-CI-002 pass on every file).
- [x] 1.2 Deploy steps are skipped inside the deploy job (step-level `if` on the configured-host output), not by a job-level `if`, so the prod approval prompt still appears before the skip. While `DEPLOY_HOST` is unset the workflow succeeds. Verify: a run with the variable unset shows the approval prompt for prod, then the skip, and the workflow succeeds.
  Evidence: step-level `if` on the configured-host output in `_deploy.yml`, actionlint clean. Open item: confirm the prompt-then-skip flow on the first real GitHub run (no repository yet).
  Evidence: NOT YET. `actionlint` checks syntax only; the approval-then-skip run was not performed. Left unticked for the orchestrator.
- [x] 1.3 Pin SSH host keys (P-0018): `infra/deploy/deploy.sh` requires `DEPLOY_KNOWN_HOSTS` once a host is set and uses `StrictHostKeyChecking=yes`; it fails before sending anything otherwise. Verify: Go tests in `tools/harness/internal/deploy` (run by `make harness-test`).
  Evidence: `TestDeploy_RefusesWithoutPinnedHostKey` and `TestDeploy_ValidInputsUseStrictHostKeyChecking` pass; real `sshd` pinned-key test passes.
- [x] 1.4 Deliver deploy secrets as files: each `DEPLOY_ENV_FILE` entry is written to its own `0600` file under `~/har/secrets`; the release env file holds digests and non-secret values only; no secret on the command line. Verify: deploy tests; `make compose-config` valid.
  Evidence: `TestDeploy_StagesSecretsAsOwnerOnlyFiles` and `TestDeploy_RejectsInvalidSecretsBeforeAnySSH` pass; `make compose-config` valid.
- [x] 1.5 Add CI scans: secret scan of the full git history and `make vuln` on pull requests; `make image-scan` on each built image before publish and on each changed deployable's image in pull requests; third-party image report in `infra-images`, non-blocking. Verify: each scan step runs; a deliberately vulnerable test image fails `make image-scan`.
  Evidence: gitleaks caught a seeded token (rc=1); `make vuln` reported GO-2026-4602 under Go 1.26.0 (rc=2) and was clean under go1.26.11; `make image-scan` failed on the vulnerable pinned postgres and old socket-proxy images (HIGH/CRITICAL, rc=1) and passed on traefik, mailpit and the new socket-proxy. Steps not yet run on GitHub.
  Evidence: PARTIAL. Secret scan: `gitleaks` no leaks. `make vuln` and `make image-scan` were not run in evidence, and no vulnerable test image was checked. Left unticked for the orchestrator.
- [x] 1.6 Document the variables and secrets (`DEPLOY_HOST`, `DEPLOY_USER`, `DEPLOY_KNOWN_HOSTS`, `DEPLOY_SSH_KEY`, `DEPLOY_ENV_FILE`, `DOMAIN`, `ACME_EMAIL`), the environments, the approval gate, rollback and the GitHub settings checklist in `infra/README.md`; update `infra/AGENTS.md` and `docs/agents/devops.md`. No secrets in git except the owner-approved local exception. Verify: R-DOC-001 passes.
  Evidence: `make check` green (R-DOC-001); `infra/README.md` contains the GitHub checklist, variables and rollback; `gitleaks` no leaks outside the approved exception paths.

## 2. Validation (implementer)

- [x] 2.1 R-CI-001: every `uses:` in `.github/workflows` that is not a local `./` reference pins a full 40-character commit SHA. Unit tests cover pass and fail. Verify: `make harness-test`.
  Evidence: `make check` green (26 rules, 0 violations); `checkActionPins` covered by `check_test.go`.
- [x] 2.2 R-CI-002: every workflow declares top-level `permissions:`, and no `run:` script interpolates `${{ }}`. Unit tests cover pass and fail. Verify: `make harness-test`.
  Evidence: `make check` green; `checkWorkflowHardening` covered by `check_test.go`.
- [x] 2.3 Add the R-CI-001 and R-CI-002 rows to `docs/standards/validation.md`. Verify: R-DOC-002 passes.
  Evidence: `make check` green; both rows present in `validation.md`.
- [x] 2.4 Deploy tests in `tools/harness/internal/deploy` (Go, run by `make harness-test`, so by `make check`) with a stub `ssh`/`scp`. They cover: no pinned host key means refusal before any call; strict host key checking on every call; invalid secrets are rejected before any SSH; secrets are staged as `0600` files. Verify: `go -C tools/harness test ./internal/deploy`.
  Evidence: deploy Go tests pass in `make harness-test`. The argument checks for `ENV`/`ENV_FILE` are not separately tested and are not claimed.

## 3. Check

- [x] 3.1 Run `make check` and record the result.
  Evidence: `make check` green: 26 rules, 0 violations; harness coverage 97.0 %; template-test 100 %; deploy tests pass; 7 specs valid. `make compose-config` valid. `actionlint` clean; `gitleaks` no leaks.
