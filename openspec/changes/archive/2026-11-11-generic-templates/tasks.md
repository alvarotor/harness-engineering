## 1. Service template, generator and harness rules (implementer)

- [x] 1.1 Create `templates/go-service` (stdlib-only; env config, JSON logs, app 8080, ops 8081 `/livez` `/readyz`, SIGTERM drain ≤ 20 s, `healthcheck` subcommand, two-stage Dockerfile with a non-root final stage and digest-pinned `FROM`, Makefile `lint`/`test`/`build`/`image`, `compose.yaml` fragment, README, AGENTS.md). Verify: template coverage ≥ 80 %.
  Evidence: `make check` green; `make template-test` 100 %; demo image healthy as uid 65532, clean SIGTERM.
- [x] 1.2 Add `scripts/new-service.sh` and `make new-service name=<x>`; refuse names outside `^[a-z][a-z0-9]*(-[a-z0-9]+)*$`, names with path separators or `..`, and existing folders, changing nothing. Verify: refusal and happy-path tests in `make harness-test`.
  Evidence: `make check` green (generator tests in `tools/harness/internal/generator` run in `make harness-test`; `make template-test` generates and tests the demo).
- [x] 1.3 Add HTTP server timeouts (read-header, read, write, idle) to both servers as named constants. Verify: `TestNewServer_SetsAllTimeouts` asserts them (the slow-client timing test is dropped).
  Evidence: `TestNewServer_SetsAllTimeouts` passes in `make template-test`; `make check` green.
- [x] 1.4 Add harness rules R-COMPOSE-001..005 and R-IMG-001 (final-stage `USER` and `HEALTHCHECK`, digest `FROM` in every stage) with unit tests (coverage ≥ 80 %); extend R-CAT-004 to require `compose.yaml`. Verify: `make harness-test`.
  Evidence: `make check` green: 26 rules checked, 0 violations, harness coverage 97.0 %.
- [x] 1.5 Document the rules in `docs/standards/validation.md` (rows for R-COMPOSE-001..005, R-IMG-001, R-CAT-004); update `docs/standards/engineering.md` and `docs/standards/agent-tooling.md` (remove `tofu` from required developer tools); add the daily commands to `CONTRIBUTING.md`; update `docs/agents/implementer.md` for the generator. Verify: R-DOC-001 and R-DOC-002 pass.
  Evidence: `make check` green; R-DOC-001 and R-DOC-002 pass.
- [x] 1.6 Remove `tofu` from the required developer tools in `tools/harness/internal/bootstrap` and its tests. Verify: `make harness-test`.
  Evidence: `make check` green; bootstrap tests pass in `make harness-test`.

## 2. Local infrastructure and runtime hardening (devops)

- [x] 2.1 Add `infra/compose/compose.yaml` (edge proxy as the only service with published ports; socket proxy; Postgres with an init script; Mailpit) and `compose.local.yaml`. Verify: `make compose-config` passes.
  Evidence: `make compose-config` valid; `make up` brings every service healthy.
- [x] 2.2 Run the edge proxy as non-root with `sysctls: net.ipv4.ip_unprivileged_port_start=0` and `cap_drop: ALL`; run the socket proxy (`wollomatic/socket-proxy`) as non-root with a GET allow-list that denies container archive, export, logs and attach; no privileged mode. Verify: socket-proxy probe matrix.
  Evidence: `make up`: all services healthy as non-root; socket-proxy probe matrix 27 cases pass.
- [x] 2.3 Deliver secrets as files: Postgres reads its password from a Compose `secrets:` entry (`POSTGRES_PASSWORD_FILE`); no secret goes into `environment:`; deploy writes every secret to a `0600` file (tested in the deploy tests). Verify: `make up` healthy with the secret mounted; deploy tests pass.
  Evidence: `make up` healthy with the Postgres secret file mounted; deploy tests pass in `make harness-test`.
- [x] 2.4 Keep `infra/env/local/compose.env` and `infra/env/local/secrets/postgres_password` committed as the owner-approved dev-only exception. Document in `infra/README.md` what they may contain (dev-only values, no production value, no value reused outside local development) and that every other file under `infra/env/` and `secrets/` stays untracked. Verify: R-DOC-001 passes.
  Evidence: `make check` green (R-DOC-001); owner-approved exception recorded in `infra/README.md`.
- [x] 2.5 Add Makefile targets `up`, `down`, `ps`, `logs`, `compose-config`. Verify: `make compose-config`.
  Evidence: `make compose-config` valid; `make up` healthy.
- [x] 2.6 Update `infra/AGENTS.md` and add the local workflow to `infra/README.md` (the GitHub settings checklist is ci-cd-template task 1.6). Verify: R-DOC-001 passes.
  Evidence: `make check` green (R-DOC-001).

## 3. Check

- [x] 3.1 Run `make check` (includes `make template-test` and `make harness-test`) and record the result.
  Evidence: `make check` green: 26 rules, 0 violations; harness coverage 97.0 %; template-test 100 %; deploy tests pass; 7 specs valid. `make compose-config` valid.
