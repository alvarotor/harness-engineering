# Proposal

## Why

Rules for services exist only as prose, so every new service rebuilds about 20 rules by hand (runtime contract,
Dockerfile, Compose labels, tests, health). The repository is a reusable base (P-0016) and no product is chosen,
so the base must give a generic, tested starting point and enforce the container rules with the validation tool,
not with reviewers' memory.

## Lane and security triggers

- **Lane: L** (new `templates/` tree, new `infra/compose/` files and harness rules; `infra/` and `Dockerfile`
  changes are lane-L paths).
- Security triggers (`docs/standards/security.md`): **exposure** (the edge proxy and the socket proxy),
  **containers** (Dockerfile users, capabilities, mounts), **supply chain** (base images pinned by digest,
  third-party images), **secrets and config** (Compose secrets, the committed dev-only exception). The owner
  approves the change and the security-reviewer reviews it before the reviewer step.

## What Changes

- **Service template:** `templates/go-service` is a stdlib-only Go service that implements the runtime contract:
  env config, JSON `slog` logs, app port 8080, ops port 8081 with `/livez` and `/readyz`, SIGTERM drain of at most
  20 s, a `healthcheck` subcommand, explicit HTTP server timeouts (asserted by `TestNewServer_SetsAllTimeouts`), a
  two-stage Dockerfile whose final stage runs as a non-root user with a `HEALTHCHECK`, base images pinned by digest,
  a Makefile (`lint`, `test` with coverage ≥ 80 %, `build`, `image`), a `compose.yaml` fragment, README and AGENTS.md.
- **Generator:** `scripts/new-service.sh` copies the template to `services/<name>` (or `apps/<name>`) and runs
  through `make new-service name=<x>`. It refuses a name outside the kebab-case rule, a name with a path separator
  or `..`, or a folder that exists, and changes nothing in that case. Its refusal tests run in `make harness-test`.
- **Local infrastructure:** `infra/compose/compose.yaml` (base), `compose.local.yaml` (local overlay) and
  `compose.deploy.yaml` (deploy overlay, owned by ci-cd-template). The base runs:
  - the edge proxy (Traefik) as non-root with `sysctls: net.ipv4.ip_unprivileged_port_start=0` and `cap_drop: ALL`
    (ports 80/443 without a granted capability);
  - a socket proxy (`wollomatic/socket-proxy`) that is the only container mounting the Docker socket, read-only,
    with a GET path allow-list that denies container archive, export, logs and attach;
  - Postgres, the only service that uses a Compose `secrets:` entry;
  - Mailpit.
  `make up`, `make down`, `make ps`, `make logs`, `make compose-config`.
- **Secrets as files:** secret values reach containers as files, never as `environment:` values. Deploy writes
  every `DEPLOY_ENV_FILE` entry to a `0600` file under `~/har/secrets` on the host (ci-cd-template).
- **Owner-approved exception:** the committed dev-only local secrets `infra/env/local/compose.env` and
  `infra/env/local/secrets/postgres_password` hold local values only. Every other secret stays untracked.
- **Harness rules:** R-COMPOSE-001..005 (only the edge proxy publishes ports; only the socket proxy mounts the
  Docker socket, read-only; no `container_name`, `links` or 8081 routing; third-party images pinned by digest;
  deployable services drop `ALL` capabilities, set `no-new-privileges` and `read_only`). R-IMG-001 (every `FROM`
  pinned by digest; the final stage has a non-root `USER` and an active `HEALTHCHECK`). R-CAT-004 also requires
  `compose.yaml` for active deployables. `make template-test` joins `make check`.
- **Docs:** `CONTRIBUTING.md` lists the daily commands (`make new-service`, `make up`, `make check`).
- **Tooling:** `tofu` is removed from the required developer tools; IaC is deferred (Q9).

## Capabilities

### New Capabilities

- `service-template`: the generated service meets the runtime contract, has explicit HTTP timeouts, reads secrets as
  files, and generation is safe.

### Modified Capabilities

- `repo-validation`: container, Compose and Dockerfile rules are validated.

## Impact

- Decisions: D-0003 (container and ops port), D-0004 (Go default), D-0008 (own DB on shared Postgres), D-0009
  (no secrets in git, the edge proxy the only exposed container, socket proxy read-only), D-0011 (JSON logs),
  D-0017 (tests and coverage), D-0021 (developer tools), P-0002 (no new infrastructure before its trigger).
- Files: `templates/go-service/**`, `scripts/new-service.sh`, `Makefile`, `infra/compose/**`,
  `infra/env/local/**`, `infra/AGENTS.md`, `infra/README.md`, `tools/harness/internal/{check,bootstrap,generator}`,
  `docs/standards/{validation,engineering,agent-tooling,security}.md`, `docs/agents/{implementer,devops}.md`,
  `CONTRIBUTING.md`.
- No product service or product feature is generated. No new decision: the rules implement D-0003, D-0009 and D-0017.
