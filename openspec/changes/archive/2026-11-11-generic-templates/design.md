# Design

## Approach

The repository ships one generic service template instead of rules repeated in prose. The template is the reference
implementation of the runtime contract (D-0003, D-0011, D-0017). The harness enforces the container rules on every
Dockerfile and Compose file, so a generated service and a hand-written one face the same checks.

## Service template (`templates/go-service`)

- Standard library only (`net/http`, `log/slog`, `os/signal`). No database driver until a product needs one. No new
  dependency (lean rule).
- Config from environment variables with defaults for local use; invalid values fail at start.
- Two HTTP servers: app on 8080, ops on 8081. `/livez` returns 200 while the process runs; `/readyz` returns 200
  only while accepting work.
- HTTP timeouts on both servers: `ReadHeaderTimeout`, `ReadTimeout`, `WriteTimeout`, `IdleTimeout` as named
  constants in `internal/platform/server.go`. `TestNewServer_SetsAllTimeouts` asserts them. No timing test of slow clients: it is flaky
  and the assertion covers the configuration.
- Shutdown: on SIGTERM, readiness turns 503 first, in-flight requests drain, and the process exits within 20 s
  (`shortcut: fixed drain window, upgrade when requests need longer`). The `healthcheck` subcommand calls `/livez`
  for the container `HEALTHCHECK`.
- Dockerfile: two stages. The build stage compiles; the final distroless stage sets `USER 65532:65532`, the
  `HEALTHCHECK` and the entrypoint. Both `FROM` images are pinned by digest (R-IMG-001).
- Makefile exposes `lint`, `test` (coverage ≥ 80 %), `build` and `image`. `make -C <path> lint test image` is what
  `pr.yml` and `_publish.yml` run (ADR-0023).
- Secrets: the service reads secret values from files whose paths come from environment variables (`*_FILE`). The
  template has no secret yet; its tests cover the file reader.

## Generator (`scripts/new-service.sh`)

- Name must match `^[a-z][a-z0-9]*(-[a-z0-9]+)*$`. That rejects uppercase, underscores, leading digits, leading,
  trailing or doubled hyphens, and path separators. The target `services/<name>` (or `apps/<name>`) must not exist.
  On refusal the script exits non-zero before writing anything.
- Refusal and happy-path tests are Go tests in `tools/harness/internal/generator`, run by `make harness-test`.
- Copies the template and replaces the module and binary name. It does not edit the catalog: the change that approves
  the service adds the `planned` entry (catalog README).

## Harness rules (`tools/harness/internal/check`)

- Implemented in `tools/harness` with `gopkg.in/yaml.v3` already in `go.mod`; no new dependency.
- R-COMPOSE-001 to R-COMPOSE-005 parse `infra/compose/*.yaml`, each deployable's `compose.yaml` fragment and the
  template fragment (see `docs/standards/validation.md`). R-COMPOSE-005 applies to deployable fragments only:
  `infra/compose` services are hardened case by case, as described above.
- R-IMG-001 reads each `Dockerfile` under `templates/`, `services/` and `apps/`, and checks `FROM` digests in every
  stage and `USER` and `HEALTHCHECK` in the final stage only. A user set in the build stage alone fails.
- R-CAT-004 gains `compose.yaml` in its required files.

## Local infrastructure (`infra/compose`)

- `compose.yaml` (base):
  - **Edge proxy (Traefik):** the only service with published ports (80/443). Non-root through
    `sysctls: net.ipv4.ip_unprivileged_port_start=0` (lets a non-root process bind 80/443) and `cap_drop: ALL`. No
    capability is added.
  - **Socket proxy (`wollomatic/socket-proxy`):** the only service that mounts the Docker socket, read-only. It runs
    a GET path allow-list for the calls the edge proxy needs. It denies container archive, export, logs and attach.
    The probe matrix (27 cases) is the evidence.
  - **Postgres:** one database per service, created by an init script (D-0008). Its password comes from a Compose
    `secrets:` entry mounted as a file (`POSTGRES_PASSWORD_FILE`). It is the only `secrets:` entry in the base file.
  - **Mailpit** for local email.
- `compose.local.yaml` adds the local domain and dev defaults.
- Secrets: outside the documented exception, secrets come from untracked files, and only `.example` files are
  committed (D-0009).
- **Owner-approved exception:** `infra/env/local/compose.env` (loaded through `env_file:`) and
  `infra/env/local/secrets/postgres_password` are committed with dev-only values. They hold no production value and no
  value reused outside local development. Any other file under `infra/env/` or `secrets/` stays untracked. The
  exception is listed in `infra/README.md`.
- Makefile wraps `docker compose` (`up`, `down`, `ps`, `logs`, `compose-config`). The OpenTelemetry local profile
  (D-0011) is left out until the first service needs it (P-0002).

## Developer tools

`tofu` is removed from the required tools in `bootstrap` and in the docs; IaC is deferred (Q9). Presence checks
(D-0021) stay for the remaining tools.

## Alternatives considered

- **Cookiecutter or another external generator:** rejected; a shell script and a template folder cover the need without
  a new tool (D-0021).
- **`NET_BIND_SERVICE` for the edge proxy:** rejected in favour of the sysctl, which adds no capability at all.
- **Check Compose with `docker compose config` only:** rejected for the port, socket and capability rules, because they
  must work without a running Docker daemon and without third-party images being pulled.
