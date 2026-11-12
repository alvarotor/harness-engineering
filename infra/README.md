# Infrastructure

Compose stack, Traefik and Postgres configuration, the deploy script and the delivery workflows
([`ARCHITECTURE.md` §6–§7](../ARCHITECTURE.md#6-delivery-and-environments)). Rules for agents: [`infra/AGENTS.md`](AGENTS.md).

## Run locally

Needs Docker with the Compose plugin, and port 80 free.

```sh
make up              # build deployables, start the stack, wait until healthy
make ps              # status and health
make logs SERVICE=traefik
make down            # stop (volumes are kept)
```

Then open `http://mail.localhost/` (Mailpit UI; services send mail to `mailpit:1025`).

Local values are dev-only: [`env/local/compose.env`](env/local/compose.env) and the dev Postgres password in
[`env/local/secrets/`](env/local/secrets/). Deploys never read `env/local`. `make compose-config` validates both
merges (local and deploy); CI runs it on every pull request.

`DOCKER_GID` (default: the group of `/var/run/docker.sock`, set by `make`) is the group the non-root socket proxy
joins. If your socket group differs (for example on Docker Desktop), override it: `make up DOCKER_GID=<gid>`.

## Compose layout

`make` merges these files, always with `--project-directory .` so every path is relative to the repository root:

| File | Used by | Contents |
|---|---|---|
| `compose/compose.yaml` | local and deploy | Base stack: `traefik`, `socket-proxy`, `postgres`; networks; volume `pgdata`; secret `postgres_password` |
| `compose/compose.local.yaml` | `make up` | `mailpit` (dev mail catcher, UI at `mail.${DOMAIN}`) |
| `compose/compose.deploy.yaml` | deploy | Port 443, TLS with Let's Encrypt, HTTP→HTTPS redirect, one-shot `letsencrypt-init`, volume `letsencrypt` |
| `services/*/compose.yaml`, `apps/*/compose.yaml` | both | One fragment per deployable (added by the generator) |

Rules that the harness checks and the design depends on:

- Only `traefik` publishes ports: `80` locally, `80` and `443` on a deploy.
- Only `socket-proxy` mounts `/var/run/docker.sock`, read-only. Traefik reaches the Docker API only through it, and the
  proxy allows only the paths in the [Docker API](#docker-api-socket-proxy) section.
- Networks: `edge` (Traefik and deployables), `data` (internal: Postgres and backends), `socket` (internal: Traefik and the proxy).
- Every third-party image is pinned by digest. Deployable images come from the release env file, not from tags.
- No `container_name`, no `links`, no routing labels for port 8081 (the ops port).

Users (checked with `docker inspect --format '{{.Config.User}}'`):

| Service | User | Why |
|---|---|---|
| traefik | `65532:65532` | Non-root. Binds 80 and 443 with `sysctls: net.ipv4.ip_unprivileged_port_start=0` (per container net namespace) |
| socket-proxy | `65534:65534` | Non-root, read-only root filesystem, no tmpfs. `group_add: ${DOCKER_GID}` gives read access to the socket |
| postgres | starts as root, runs as uid 70 | The official entrypoint needs root for the data directory, then drops privileges |
| mailpit (local) | `65532:65532` | Non-root, read-only root filesystem |

Pinned images (bump with `docker buildx imagetools inspect <image>` and take the index digest):

| Service | Image |
|---|---|
| traefik | `traefik:v3.7.18@sha256:9c437ac4f09ce6f4980a8544902d3475fdf70a1a7c578729581d2535dfa7d635` |
| socket-proxy | `wollomatic/socket-proxy:1.13.2@sha256:ff95d6ff6c27853005ef85eb8f4ab5923dc3d2f1e055f98fb6aeeaceec3fe848` |
| postgres | `postgres:17.11-alpine@sha256:3540c0362c809df084ffac0ae69fe794515f9de92692f02178bcdce0ef0302ac` |
| mailpit (local) | `axllent/mailpit:v1.33.1@sha256:aef2aacc11f075ec0fb6874275ee4a421e688e11a96ebb296134d1735c7761ce` |

### Traefik configuration

Static settings are in [`traefik/traefik.yml`](traefik/traefik.yml), mounted as `/etc/traefik/static.yml`. Two rules
keep it correct:

- **Keep the deploy overlay in sync.** Traefik ignores flags and environment variables once it loads a config file,
  and a file cannot read `${ACME_EMAIL}`. So `compose.deploy.yaml` drops `--configFile` and repeats the settings as
  flags. Change both when you change one.
- **Never mount a file at `/etc/traefik/traefik.yml`.** Traefik loads that path automatically and then ignores flags.

Deny by default: `global.aliasHeadersStrategy: reject` rejects a request with a header that
aliases another (`X.Auth.User`, `X_Auth_User`) with 400, and `encodedCharacters` rejects encoded `/`, `\`, NUL, `;`,
`%`, `?` and `#` in the path, also with 400. A product that needs one of these needs an owner decision and a
change to both files. Access-log fields are not restricted yet: do this before the first public route.

Postgres init scripts (one database and two roles per service) are in [`postgres/init/`](postgres/init/README.md).

### Docker API (socket proxy)

Traefik's docker provider needs the Docker API. The stock Tecnativa proxy opened the whole `/containers` prefix, so
`GET /containers/<id>/archive` returned files, including the Postgres secret file.
The socket proxy now is `wollomatic/socket-proxy` with an anchored regex allow-list ([`compose.yaml`](compose/compose.yaml)):

| Method | Allowed paths (optional `/v<api>` prefix) | Denied |
|---|---|---|
| GET | `_ping`, `version`, `events`, `containers/json`, `containers/{id}/json`, `networks[/{id}]` | everything else (403 for an unknown path, 405 for a method) |
| HEAD | `_ping` (Traefik ≥ 3.6.1 uses HEAD `/_ping`) | everything else |

It listens on `socket:2375` only, `-allowfrom=traefik`, and has a health check (`/healthcheck`, `-allowhealthcheck`).
It has no shell and no tmpfs; its root filesystem is read-only.

What it does **not** stop: `containers/{id}/json` returns the container's configuration (environment names and values,
mounts, labels). Secrets are therefore files, not environment values. Configuration that the operator puts in a label or
environment variable is visible to a compromised Traefik. Do not put secrets in either.

Test it (the probe matrix ran against the stack from inside `traefik`, the allowed client). Expect 200 for the allowed
paths and 403 or 405 for the rest; the archive response body must be empty:

```sh
docker exec harness-engineering-traefik-1 sh -c "timeout 5 wget -S -O /dev/null http://socket-proxy:2375/containers/<pg-id>/archive?path=/run/secrets/postgres_password"   # 403
docker exec harness-engineering-traefik-1 sh -c "timeout 5 wget -S -O /dev/null http://socket-proxy:2375/containers/json"                                              # 200
```

Traefik's own discovery is the functional test: `http://mail.localhost/` returns 200, so the routes were discovered.

### Encoded characters and dots

Traefik's `encodedCharacters` options cover `/`, `\`, NUL, `;`, `%`, `?` and `#` (denied above). They do not cover the
encoded dot. `%2e` is not denied: a request for `/%2e%2e/x` reaches the backend (verified).
Traefik's `sanitizePath` removes raw `/../` only. No static option denies encoded dots, so this is a known limit: every
backend must normalise and reject `..` after decoding. Add a middleware (owner decision) if a product needs more.

## Secrets

Runtime secrets are **files, never environment variables**: `docker inspect` and the Docker API
then show a path, not a value. The Compose files read them from `HARNESS_ENGINEERING_SECRETS_DIR` (set in the env file):

- Local: `HARNESS_ENGINEERING_SECRETS_DIR=./infra/env/local/secrets` (dev-only values, committed on purpose: dev-only exception, see
  docs/standards/security.md accepted risks).
- Host: `HARNESS_ENGINEERING_SECRETS_DIR=./secrets` (`~/harness-engineering/secrets`, written by the deploy with mode `0600`; folder `0700`).

Each secret is one file named after its variable in lower case: `POSTGRES_PASSWORD` becomes `secrets/postgres_password`.
Services read them with the `<NAME>_FILE` convention (`templates/go-service` supports it). Non-secret values and image
digests stay in the env file. Committing a real secret is a stop-the-line incident: rotate it, then purge it.

## Deploy

Environments: `dev` (automatic on merge to `main`) and `prod` (a `vX.Y.Z` tag on a `main` commit, after owner
approval). Each is one Linux host running Docker Compose, reached over SSH. Hosts are not provisioned by this
repository yet (Q6).

What `make deploy` does ([`deploy/deploy.sh`](deploy/deploy.sh)):

1. Checks the environment name and that `KNOWN_HOSTS` has an entry for `HOST`. Without it the deploy stops.
2. Turns the secrets file into `0600` files in a temporary folder. Values never reach a log; errors name a line number.
3. Archives `infra/compose`, `infra/traefik`, `infra/postgres`, `infra/env/<env>` and every deployable fragment, plus the
   env file (as `.env`, with `HARNESS_ENGINEERING_SECRETS_DIR=./secrets`) and the secret files. A failed archive stops the deploy.
4. Sends the archive over SSH with `StrictHostKeyChecking=yes`, `UserKnownHostsFile` set to the pinned file and
   `GlobalKnownHostsFile=/dev/null`. On the host, `~/harness-engineering` is `755`, `.env` and `secrets/` keep their modes.
5. `docker compose ... pull --ignore-buildable`, then `up -d --no-build --remove-orphans --wait`. Migrations go between
   these two steps once a service has a database (commented in the script).
6. Fails unless every service is running and healthy, and the one-shot `letsencrypt-init` exited with 0. If `DOMAIN`
   is set, requires `https://$DOMAIN/` to answer below 500.

The images are the ones `main.yml` built for the commit (tag = commit SHA, pushed to `ghcr.io/<owner>/harness-engineering-<name>`).
The release env file holds `IMAGE_<NAME>=ghcr.io/...@sha256:...` (`<NAME>` in upper case with `_`, for example
`IMAGE_ORDERS_API`), `ENV`, `DOMAIN` and `ACME_EMAIL`. If an image for the commit is missing, or has no digest,
the deploy fails: `main` has not built that commit.

Run a deploy by hand only from a checkout of the commit, with files you created yourself:
`make deploy ENV=dev HOST=<host> SSH_USER=<user> ENV_FILE=<file> SECRETS_FILE=<file> KNOWN_HOSTS=<file> [SSH_KEY=<file>]`.

### Recording the host key (owner)

The deploy never trusts a key on first use. Record the key once per host, after checking its fingerprint out of band:

1. Get the fingerprint from the host itself, through a channel you trust (the provider's console, or a shell you
   already have): `ssh-keygen -lf /etc/ssh/ssh_host_ed25519_key.pub`.
2. From your workstation, fetch the key the name resolves to: `ssh-keyscan -t ed25519 <DEPLOY_HOST> > host_key.txt`.
3. Compare: `ssh-keygen -lf host_key.txt` must print the same fingerprint as step 1. If it does not, stop: someone else
   may be answering for that name. Do not record the key.
4. Store the **whole** line from `host_key.txt` (it starts with the name you used in step 2, for example
   `203.0.113.10 ssh-ed25519 AAAA...`) as the GitHub Environment variable `DEPLOY_KNOWN_HOSTS`, for `dev` and for
   `prod` separately. Use the same name as `DEPLOY_HOST`. Delete `host_key.txt` afterwards.

Re-record only after a host rebuild, and only with the same out-of-band check. A changed key fails the deploy on
purpose.

### GitHub setup checklist (owner, once, before any secret is stored)

Nothing here is configured yet. Until `DEPLOY_HOST` exists, the deploy job skips itself and succeeds. The rules
below are settings only GitHub can enforce; the workflows cannot check them.

1. **Environment `dev`:** deployment branches = **`main` only**. No required reviewers.
2. **Environment `prod`:** **required reviewers = the owner**, deployment branches/tags = **tags `v*.*.*` only**.
   An unconfigured `prod` still shows the approval prompt before the step-level skip, so the owner approves
   a release even when no host is set yet.
3. Variables per environment: `DEPLOY_HOST` (empty = deploy skipped), `DEPLOY_USER`, `DEPLOY_KNOWN_HOSTS` (see the
   host-key procedure above), `DOMAIN`, `ACME_EMAIL`.
4. Secrets per environment: `DEPLOY_SSH_KEY` (private key of `DEPLOY_USER`, no passphrase) and `DEPLOY_ENV_FILE`
   (`NAME=value` lines, one per line; becomes the secret files). Never store them as repository secrets.
5. **No repository-level secrets.** Only environment secrets exist, so a PR job cannot read them.
6. **Fork pull requests get no secrets:** keep "Send secrets to workflows from fork pull requests" off.
7. **Branch protection on `main`:** pull request required, status check **`check`** (from the `pr` workflow)
   required. Also require `security` once the Go vulnerability finding above is fixed.
8. **Tag ruleset on `v*`:** only the owner or repository admins can create tags. The approver checks the tag commit is
   on `main` before approving `prod` (the workflow's ancestry check is a fast fail, not the gate).
9. **Default workflow permissions:** Settings → Actions → General → "Read repository contents" (read-only), and
   "Allow GitHub Actions to create and approve pull requests" off.
10. Host prerequisites: Docker Engine and Compose v2; `DEPLOY_USER` in the `docker` group (see the residual risk below);
    ports 80 and 443 open (the host firewall allows 22, 80 and 443 only: Docker-published ports bypass UFW rules,
    so only `traefik` publishes); a DNS record for `DOMAIN` pointing at the host.

### Release

```sh
git switch main && git pull
git tag vX.Y.Z && git push origin vX.Y.Z
```

`release.yml` fails unless the tag is `vX.Y.Z` and its commit is on `main`. It then waits for approval in the
`prod` environment and deploys the digests built for that commit.

### Rollback

Re-run the release of the previous tag: GitHub → Actions → `release` → the run for the previous `vX.Y.Z` → **Re-run
all jobs**, then approve `prod`. Its images stay in GHCR under their commit SHA, so no rebuild happens. Rollback
changes only images: migrations must be backward compatible (expand/contract).

## CI workflows

| File | Trigger | Permissions (ceiling) | What it does |
|---|---|---|---|
| `pr.yml` | pull request to `main` | `contents: read` | `check` (`make check`, `make compose-config`), `security` (`make secrets-scan`, `make vuln`), `changes` → `deployable` per changed deployable |
| `_deployable.yml` | called by `pr.yml` | `contents: read` | `make -C <path> lint test image`, then `make image-scan` |
| `main.yml` | push to `main` | caller grants per job | build and push every deployable (`_publish.yml`), then deploy `dev` |
| `_publish.yml` | called by `main.yml` | `contents: read`, `packages: write` | lint, test, build, scan and push one image under the commit SHA |
| `release.yml` | tag `v*.*.*` | `contents: read` | tag check, then deploy `prod` (environment approval) |
| `_deploy.yml` | called by `main.yml` and `release.yml` | `contents: read`, `packages: read` | skip or deploy one environment (see above) |

Why `_publish.yml` exists: a called job cannot hold more permissions than its caller. A single reusable workflow with
`packages: write` would force every pull request run to hold it. So the push path has its own file, and pull requests
never get `packages: write`. Top-level `permissions: {}` in the reusable files grants nothing; every job sets its own.

Every action is pinned by commit SHA (R-CI-001). No step interpolates `${{ }}` inside `run:`; values reach the
shell through `env:` (R-CI-002).

### Security scans

Run in CI, and locally with the same targets. Images are pinned by digest.

- `make secrets-scan`: gitleaks `v8.30.1` over the full git history with [`gitleaks.toml`](gitleaks.toml). The only
  allowlisted paths are exactly `infra/env/local/compose.env` and `infra/env/local/secrets/postgres_password` (dev-only
  exception, see docs/standards/security.md accepted risks).
- `make vuln`: govulncheck `v1.9.0` over `tools/harness` and every Go deployable. Fails on a reachable vulnerability.
  It needs a local Go of at least 1.26.1: the harness and the service template pin `toolchain go1.26.1`, and
  `go 1.26.0` is the minimum language version. Install 1.26.1 or newer before running it.
- `make image-scan IMAGE=<image>`: Trivy `0.76.2`, `--severity HIGH,CRITICAL --ignore-unfixed`, exit 1 on a finding.
  CI runs it on every deployable image.
- Infra images (socket proxy, Postgres, Traefik, Mailpit) are scanned in CI by the `infra-images` job in `pr.yml`.
  Every image listed by `make infra-images` is scanned. The job is report-only (`continue-on-error`): a finding is
  reported, never blocks a merge. The owner accepted this.

Current results (2026-11-11, local runs): `make vuln` fails on GO-2026-4602 (standard library `os`, fixed in Go 1.26.1)
while the local toolchain is 1.26.0; upgrading the toolchain clears it. `make image-scan` on the pinned infra images:
`wollomatic/socket-proxy:1.13.2` is clean (replaced `tecnativa/docker-socket-proxy:v0.5.0`, 9 fixable HIGH);
`postgres:17.11-alpine` has 44 fixable findings: 18 HIGH in the OS and 25 HIGH plus 1 CRITICAL in the bundled
`gosu` binary (the current tag digest has the same findings; owner accepted). Traefik `v3.7.18` and Mailpit `v1.33.1` are clean.

## Known limits

- Traefik access logs include the default fields. Restrict them before a public route exists.
- HSTS and security headers are not configured: no public route exists yet.
- Smoke test accepts any status below 500. Tighten with a health path when public routes exist.
- Encoded dots (`%2e`) are not denied by Traefik (see Encoded characters and dots).
- `main.yml` builds every deployable on each merge. Its `shortcut:` comment and ARCHITECTURE.md §11 describe when to change it.

### Residual risk (for `docs/standards/security.md`)

- **Docker API read access (reduced by the allow-list):** the proxy allows only the paths in the Docker API section.
  A compromised Traefik can still read container configuration (`containers/{id}/json`) and events. Secrets are files,
  and no secret goes into an environment value or label. Logs, archive and export are denied.
- **Root on the deploy host (accepted):** `DEPLOY_USER` is in the `docker` group, which is root on the host. A
  leaked `DEPLOY_SSH_KEY` gives host root. Mitigated by a dedicated host, a key that exists only in `prod`/`dev`
  environments, and reviewers on `prod`.
- **Unpinned transitive npm install (accepted by owner):** `openspec@1.14.1` is installed without a lockfile.
- **Infra image CVEs (owner accepted):** reported in CI by the `infra-images` job, not gated (see Security scans above).
- **Traefik encoded dots (owner decision):** `%2e` is not denied (see Encoded characters and dots).
- **No image signing or provenance:** a compromised GHCR push is trusted by digest.
- **`:ro` on the socket is not a control:** only `POST=0` and the proxy's path rules restrict the Docker API.
