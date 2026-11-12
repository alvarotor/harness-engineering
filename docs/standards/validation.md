# Validation rules

Run all of them with `make check`. Each violation prints `RULE  path: problem` followed by `fix: …`.
Rule IDs are stable. `tools/harness` fails (R-DOC-002) if a rule exists in code but not here, or the reverse.

| ID | Checks | Typical fix |
|---|---|---|
| R-STRUCT-001 | Required repository files exist (root docs, registry, catalog, standards, agent index, OpenSpec config, `scripts/bootstrap.sh`); the bootstrap script is executable | Restore the file / `chmod +x scripts/bootstrap.sh`. Never delete harness files without a decision change |
| R-REG-001 | Registry parses; IDs are well-formed (`D-NNNN`, `P-NNNN`, `A<n>`, `Q<n>`) and unique; `status`/`source` values are valid | Fix the YAML entry |
| R-REG-002 | Each decision's ADR file exists and its `- Status:` line equals the registry status | Align ADR and registry status |
| R-REG-003 | Every ADR file is referenced by a registry decision and listed in `docs/adr/README.md` with the same status | Add the registry entry or the index row |
| R-REG-004 | Every decision and preference `enforced_by` path exists | Create the enforcing file or move the path to `planned_enforced_by` |
| R-REG-005 | `superseded` decisions/preferences name an existing replacement of the same type; other entries don't have `superseded_by` | Follow the decision-change process in CONTRIBUTING.md |
| R-REG-006 | `docs/decisions/history.yaml` (optional) holds only superseded/deprecated decisions and preferences, resolved questions and the changelog; it is merged with the registry for R-REG-001..005 | Move current entries back to `registry.yaml` |
| R-CAT-001 | Catalog schema: unique kebab-case names, `kind` app/service with a matching path, valid `status`, `owner`, `domain` (services), `personal_data` present | Fix the catalog entry |
| R-CAT-002 | Language rule: apps use `typescript`, services use `go`, unless a registry decision with that `scope` sets another `default_language` or lists the deployable in its `exceptions` with the same language | Use the default language, or add a justified exception through an approved change |
| R-CAT-003 | Every folder in `apps/` and `services/` has a catalog entry; `active` entries exist on disk; `planned` entries don't | Add or update the catalog entry (`planned` → `active` when code lands) |
| R-CAT-004 | Active deployables have `Dockerfile`, `AGENTS.md`, `README.md`, `compose.yaml` and `go.mod` (Go) or `package.json` (TypeScript) | Add the missing file (templates generate them) |
| R-CAT-005 | `depends_on` names existing catalog entries | Fix the dependency name or add the entry |
| R-COMPOSE-001 | In `infra/compose/*.yaml`, `services/*/compose.yaml`, `apps/*/compose.yaml` and `templates/*/compose.yaml`, only the `traefik` service publishes host `ports`; every Compose file parses as YAML | Remove `ports` and route through Traefik; fix the YAML |
| R-COMPOSE-002 | Only the `socket-proxy` service mounts the container runtime socket (`/var/run/docker.sock` or `/run/docker.sock`), and only read-only | Route the request through socket-proxy; mount with `:ro` |
| R-COMPOSE-003 | No service sets `container_name` or `links`, and no label routes port 8081 (`loadbalancer.server.port: 8081`) | Remove the key or label; 8081 serves probes only |
| R-COMPOSE-004 | Every literal `image:` in those Compose files is pinned by `@sha256:`; references containing `${…}` are exempt because the variable carries the pin | Pin the image by digest (`docker buildx imagetools inspect <image>`) |
| R-COMPOSE-005 | Every service in a deployable fragment (`services/*`, `apps/*`, `templates/*`) drops `ALL` capabilities, sets `no-new-privileges:true` and has `read_only: true`. `infra/compose` services are exempt: they are hardened case by case | Add `cap_drop: [ALL]`, `security_opt: ["no-new-privileges:true"]` and `read_only: true` |
| R-IMG-001 | Dockerfiles of active deployables and of `templates/*/`: every `FROM` is pinned by `@sha256:` (except `scratch` and earlier stage names); the **final** stage has a non-root `USER` and an active `HEALTHCHECK` (earlier stages are discarded, so they do not count; `HEALTHCHECK NONE` does not count) | Pin the base by digest; set `USER 65532:65532` and `HEALTHCHECK` in the final stage |
| R-CI-001 | Every `uses:` in `.github/workflows/*.yml` is a full 40-character commit SHA, except local `./` actions and `docker://` references | Pin to the commit SHA and keep the tag in a comment |
| R-CI-002 | Every workflow has a top-level `permissions:` key, and no `run:` script interpolates `${{ }}` (pass values through `env:`) | Add `permissions: {contents: read}` and widen per job only where needed; move the expression into `env:` |
| R-AGENT-001 | Each role in `docs/agents/` has all required sections | Add the missing section |
| R-AGENT-002 | Each role has an OpenCode adapter in `.opencode/agents/` that references its canonical file | Add or fix the adapter |
| R-SCOPE-001 | Scoped `AGENTS.md` exist for `apps/`, `services/`, `contracts/`, `infra/` | Restore the file |
| R-DOC-001 | Relative Markdown links resolve (archived changes, generated tool files and dependencies are skipped) | Fix the link or restore the target |
| R-DOC-002 | Rules in code and in this table match | Document the rule here, or remove the stale row |
| R-DOC-003 | Every catalog deployable path is mentioned in `ARCHITECTURE.md` | Update ARCHITECTURE.md §2 |

## Other checks run by `make check`
- `make harness-test`: harness unit tests (coverage ≥ 80 %) and the `scripts/bootstrap.sh` tests (POSIX `sh` with fake tools: OS detection, missing-tool guidance, no installer calls, Termux Docker Compose reporting).
- `make harness-lint`: golangci-lint on the harness.
- `make template-test`: generates `demo` from `templates/go-service` into a temporary folder and runs its `lint` and `test` targets, so the template stays buildable.
- `go -C tools/harness test` also runs `scripts/new-service.sh` on a copy of the template (name rules, refusal without changes, generated service tests).
- `go -C tools/harness test` also runs `scripts/rename.sh` on a copy of the repository (every slug form replaced, file modes kept, the renamed harness builds, refusal without changes).
- `go -C tools/harness test` also runs `infra/deploy/deploy.sh` against a fake `ssh`: no `ssh` call without a pinned host key, secrets staged as 0600 files, `StrictHostKeyChecking=yes` on every call (skipped when `sh`, `tar` or `ssh-keygen` is missing).
- `make spec-validate`: `openspec validate --all --strict`.

## Planned (arrive with the code they check)
Go layer-import rule, cross-deployable import rule, test-file presence per module and per-deployable coverage.
Template conformance is checked by `make template-test` and R-IMG-001.
