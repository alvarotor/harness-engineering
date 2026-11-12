# Architecture

> Status: **starting architecture** of the template. This repository is a reusable base for new products; each
> product is its own repository created from it. It fixes the generic, technical architecture; product choices are
> open questions (Q1–Q8 in [`docs/decisions/registry.yaml`](docs/decisions/registry.yaml)) decided when their
> trigger is reached ([`docs/standards/new-product.md`](docs/standards/new-product.md)). A product changes a rule
> here by recording a decision (registry + ADR); if this file then disagrees with an accepted ADR, the ADR wins and
> this file must be fixed.

## 1. System context

```mermaid
flowchart LR
  user([Browser / client]) -->|HTTPS| edge[Traefik<br/>TLS, routing, headers]
  edge --> apps[apps/*<br/>TypeScript frontends]
  edge --> svcs[services/*<br/>Go services]
  svcs -->|gRPC| svcs
  svcs --> pg[(Postgres server<br/>one database per service)]
```

Traefik is the only exposed container. Which apps and services exist depends on the product (Q1); the
[service catalog](docs/catalog/services.yaml) is the source of truth.

## 2. Service boundaries

- **One service per business domain** (a bounded context such as catalog, orders, billing). A new domain means a
  new service with its own database and proto package, generated from `templates/go-service`. A new entity of an
  existing domain stays in that service. Not one service per entity (chatty calls, no transactions inside a domain).
  Start with a few coarse services: splitting later is cheap, merging is not.
- **Go is the default backend language.** Another language only when its core dependency requires it, recorded as
  a registry exception ([engineering.md](docs/standards/engineering.md#go-default-backend)).
- **Frontends are TypeScript** containers; framework and rendering mode are decided per product (Q8).
- **Monorepo:** one repository per product. Deployables never import each other; `packages/` holds build-time
  libraries only. Cross-domain workflows cannot share a database transaction; they go through gRPC or events.

## 3. Communication model

| Kind | Mechanism | Status |
|---|---|---|
| Service → service, synchronous | gRPC (connect-go) on the private network, contracts in `contracts/proto/<domain>/v<N>/` | starting rule |
| Browser / third party → backend | REST transcoding, Connect JSON client or other | **deferred, Q5** (first public API) |
| Service → service, asynchronous | none yet | **deferred, Q4** (first event) |
| Identity propagation | a token from the identity component, verified locally by each service | **deferred, Q7** (first signed-in feature) |

Breaking contract changes need a new version package (`v<N+1>`). Messaging, when added, sits behind
`Publisher`/`Subscriber` ports so the transport can change inside one adapter.

## 4. Data ownership

- **One Postgres server, one database per service** in every environment, major version pinned by image digest.
  Each database has `<svc>_owner` (DDL, migrations only) and `<svc>_app` (DML, runtime) roles; `CONNECT` is granted
  only to that service's roles.
- No service reads another service's database. No cross-database foreign keys; reference other data by opaque IDs.
- Migrations: versioned SQL in the owning service, applied by the explicit `migrate` command of **the same
  image**, before rollout, never at app startup. Changes are expand/contract across releases.
- Each service gets its own connection URL, so a database can move to another server through config alone.
- **Backups:** decided with hosting (Q6). Mandatory before real user data reaches prod.

## 5. Repository topology

```text
.
├── AGENTS.md  ARCHITECTURE.md  CONTRIBUTING.md  Makefile
├── apps/                     # frontends (deployable), once Q8 is decided
├── services/                 # backends (deployable), one per domain, from templates/go-service
├── packages/                 # build-time libraries only; never deployable
├── contracts/proto/          # source of truth for service APIs and events, versioned
├── infra/
│   ├── compose/              # base + local + deploy overlays (same images everywhere)
│   ├── traefik/              # static config
│   ├── postgres/             # init: databases + roles per service
│   ├── env/{local,dev,prod}/ # NON-secret config per environment
│   ├── deploy/               # deploy script (SSH + Compose)
│   └── k8s/                  # planned: only after a decision approves Kubernetes
├── templates/                # generators: go-service (frontend/auth templates once decided)
├── e2e/                      # planned: end-to-end suite, with the first user flow
├── openspec/                 # specs, changes, project context
├── docs/{adr,decisions,standards,catalog,agents}/
├── tools/harness/            # Go CLI: repository validation rules
└── .github/workflows/        # pr.yml, main.yml, release.yml + reusable workflows
```

## 6. Delivery and environments

| Env | Runs on | Images | Trigger | Secrets |
|---|---|---|---|---|
| local | developer machine, `make up` | built locally | — | `infra/env/local/*.env` (dev-only values) |
| dev | a Linux host with Docker Compose, over SSH | GHCR, by digest | merge to `main` | GitHub Environment `dev` → `0600` secret files |
| prod | a Linux host with Docker Compose, over SSH | **same digests** as dev | tag `vX.Y.Z` + owner approval | GitHub Environment `prod` → `0600` secret files |

```text
change/<name> ──PR──► pr.yml      make check + checks of changed deployables
merge to main ──────► main.yml    build every image once (tag = git SHA) → GHCR → deploy dev
tag vX.Y.Z ─────────► release.yml same digests → owner approval → deploy prod   (rollback: re-run previous tag)
```

- Branches: short-lived `change/<name>` branches into `main`, one OpenSpec change per branch (lane S may use
  `fix/<name>`). `main` is always deployable. GitHub Actions are pinned by commit SHA with least-privilege `permissions:`.
- Deploy: upload Compose files, a non-secret `.env` (digests, domain) and `0600` secret files → `docker compose pull`
  → `up -d --wait` → smoke test. A `migrate` step before `up` is **planned**: it is added to `infra/deploy/deploy.sh`
  with the first service that has a database.
- Deploys pin SSH host keys: the target's `known_hosts` lines come from the GitHub Environment variable
  `DEPLOY_KNOWN_HOSTS`, with `StrictHostKeyChecking=yes`; a deploy without them fails instead of trusting a first key.
- Deploy jobs skip themselves until the environment's host variables exist, so CI works before any server does.
- How hosts are created (cloud provider, infrastructure as code), DNS and TLS domain are **deferred (Q6, Q2)**.

## 7. Security model (baseline: OWASP ASVS L1)

- **Authentication:** deferred (Q7). Whatever is chosen: no custom crypto, short-lived tokens, verified locally.
- **Authorization:** deny by default; coarse roles from the token, resource ownership checks in each
  service's application layer.
- **CORS:** backends allow only the product's frontend origins, configured per environment.
- **Network:**
  - Only Traefik publishes ports (80/443).
  - Traefik reaches Docker through a **read-only socket proxy** on an isolated network; the socket is never mounted into Traefik.
  - `exposedByDefault=false`. Data stores and ops ports are never routed.
  - Hosts come from env: `${DOMAIN}` and its subdomains (decided per product). Locally: `*.localhost`.
  - The Traefik dashboard and API are disabled on dev and prod hosts.
- **Containers:** non-root, `cap_drop: ALL`, `no-new-privileges`, read-only root filesystem where possible,
  base images pinned by digest.
- **Secrets:** never in git. Runtime injection from GitHub Environments.
- **Compliance:** GDPR applicability decided with the launch country (Q3); the catalog keeps a personal-data inventory per deployable.

## 8. Observability model

- **Logs:** JSON to stdout. Fields: `ts level msg service version env request_id`, plus `trace_id` once tracing
  exists (planned). Never log secrets, tokens, cookies or personal data.
- **Correlation:** read or generate `X-Request-ID` (in the template). Propagating it and `traceparent` on outgoing
  HTTP, gRPC and event metadata is **planned**: it arrives with the first service-to-service call.
- **Traces and metrics:** **planned**, not in the template. Rule when added: OpenTelemetry SDK, exporting only when
  `OTEL_EXPORTER_OTLP_ENDPOINT` is set. A local telemetry stack is not part of Compose yet (§13).

## 9. Service runtime contract (every deployable)

| Item | Standard |
|---|---|
| App port | `8080` |
| Ops port | `8081`, never routed by Traefik: `GET /livez` (process alive), `GET /readyz` (dependencies OK) |
| Routing | Compose labels on the service. `traefik.enable=true` only for public services. Labels never reference 8081 |
| Config | env vars only. Validated at startup; fail fast with a clear message |
| Secrets | env vars injected at runtime; `<NAME>_FILE` variant supported |
| Shutdown | SIGTERM: stop readiness → drain → exit within 20 s |
| State | stateless; volumes only for data stores |
| Image | multi-stage, pinned base digest, non-root UID 65532, `HEALTHCHECK` on the ops port, no shell in Go images |
| Resources | CPU and memory limits in the Compose fragment (`deploy.resources`); they become Kubernetes requests/limits |
| Frontend config | runtime config rendered from env at container start, so one image works in every environment |

`templates/go-service` implements this contract; `tools/harness` checks the parts that can be checked.

## 10. Testing model

| Layer | Tooling | Gate |
|---|---|---|
| Unit | `go test -race` (TS: the frontend stack's runner) | **≥ 80 % coverage** per deployable; generated code excluded |
| Integration | Testcontainers (Postgres, other stores) | every PR touching the service |
| Contract | `buf breaking`; generated clients compile | blocks merge |
| Migration | apply to an empty DB and to a DB at the previous version | blocks merge |
| E2E | Playwright against the Compose stack (emails read from Mailpit) | every user-facing flow when possible |

## 11. Key trade-offs

| Decision | Gain | Cost |
|---|---|---|
| Compose on one host per env | simple, cheap | single point of failure; seconds of downtime per deploy |
| One Postgres server, DB per service | cheap; clear ownership | shared failure domain and resource contention |
| Service per domain | clear ownership; independent deploys | more services to operate; no cross-domain transactions |
| Traefik via labels + socket proxy | routing next to the service; maps to Kubernetes IngressRoute | Docker API access (mitigated by a read-only proxy) |
| Product choices deferred | no rework when the product is chosen | the first product sessions must answer Q1–Q8 |
| Build every image on each merge | simple, always a full release set | slower CI as deployables grow |

## 12. Kubernetes migration path

Already in place: same images, env-only config, probes on `:8081`, stateless services, graceful SIGTERM,
resource limits in each deployable's Compose fragment, DNS service names only. Planned: migrations as a separate
`migrate` command (with the first database).

| Compose today | Kubernetes later |
|---|---|
| service + `deploy.resources` | Deployment + requests/limits |
| healthcheck on `:8081` | liveness/readiness/startup probes |
| `docker compose run <svc> migrate` | Job / Helm pre-upgrade hook |
| Traefik + Docker labels | Traefik IngressRoute (or Gateway API) + cert-manager |
| secret files from GitHub Environments | ConfigMap + Secret via External Secrets |
| networks `edge` / `data` / `socket` | NetworkPolicies |
| Postgres container | managed service (URL change only) |

## 13. Deferred components and trigger conditions

| Component | Introduce when |
|---|---|
| Kubernetes | need > 1 node for availability, zero-downtime deploys, autoscaling, or > ~8 deployables |
| API gateway | third-party API consumers (keys/quotas), or shared policies across > 3 public backends |
| Messaging (Q4) | the first event between services; durable if losing it is unacceptable |
| Service mesh / mTLS | compliance requires encryption between services, or > 10 services with traffic policies |
| Cache | measured latency or DB load problem |
| Secrets manager | > 1 host per env, mandatory rotation, or audit requirement |
| Infrastructure as code (Q6) | the first deploy to a real dev host |
| Staging environment | the owner asks for it, or prod releases need a production-like rehearsal |
| Postgres backups | **before any real user data in prod** |
| Creating a new service's database on an existing server (the init scripts run only on an empty volume) | the first service with a database |
| `migrate` subcommand in the template and the `migrate` step in `infra/deploy/deploy.sh` | the first service with a database |
| OpenTelemetry SDK, `trace_id` in logs, `X-Request-ID`/`traceparent` propagation on outgoing calls | the first service-to-service call |
| Contract tooling: `buf.yaml`, code generation, a `buf breaking` check | the first gRPC call between services |
| Shared Go code for all services (today each service copies `internal/platform`) | the second service |
| Import rules enforced with golangci-lint `depguard` (layers inside a service, no imports between deployables) | the first service with business code |
| Local telemetry stack (for example `grafana/otel-lgtm` as a Compose profile) | traces are needed to debug |
| Dependabot and an SBOM per image | the first deployable |
| `pnpm audit` in CI | the first TypeScript deployable |
