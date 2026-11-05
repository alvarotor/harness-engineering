# Architecture Decision Records

Format: short MADR. One decision per file: `NNNN-kebab-title.md`.
- Edit `proposed` ADRs freely.
- Never edit the decision text of an `accepted` ADR. To change it, write a new ADR with `Supersedes: NNNN`
  and set the old one to `superseded`.
- Every ADR has an entry in [`../decisions/registry.yaml`](../decisions/registry.yaml); the harness checks they match.

Status values: `proposed` → `accepted` → (`superseded` | `deprecated`).
`Source` is either `user` (stated or approved by the owner) or `default` (an agent proposal that was not contested).

| ADR | Title | Status | Source |
|---|---|---|---|
| [0001](0001-repository-as-project-memory.md) | Repository as project memory, precedence and change control | accepted | user |
| [0002](0002-monorepo-topology.md) | Monorepo topology | accepted | default |
| [0003](0003-containers-compose-k8s-ready.md) | Containers everywhere; Compose now, Kubernetes-ready | accepted | user |
| [0004](0004-go-hexagonal-lite-backends.md) | Go is the default backend language (override rule) | accepted | user |
| [0005](0005-identity-better-auth.md) | Identity via Better Auth (Node), magic link, JWT/JWKS | accepted | user |
| [0006](0006-frontends-preact-vite-spa.md) | Frontends: Preact + Vite SPA served as static containers | accepted | user |
| [0007](0007-contracts-proto-connect-rest.md) | Protobuf contracts, gRPC for sync calls, REST edge via transcoding | accepted | user |
| [0008](0008-postgres-db-per-service.md) | One PostgreSQL server, database per service, explicit migrations | accepted | user |
| [0009](0009-edge-and-network.md) | Traefik edge via Docker labels and socket proxy | accepted | user |
| [0010](0010-deployment-and-secrets.md) | VM deployment, digest promotion, GitHub Environments secrets | accepted | user |
| [0011](0011-observability.md) | Structured logs, correlation IDs, OpenTelemetry opt-in | accepted | user |
| [0012](0012-security-baseline.md) | Security baseline OWASP ASVS L1 | accepted | user |
| [0013](0013-ci-cd-github-actions.md) | GitHub Actions with reusable, path-filtered workflows | accepted | user+default |
| [0014](0014-service-per-domain.md) | One service per business domain | accepted | user |
| [0015](0015-async-redis-pubsub.md) | Asynchronous messaging via Redis pub/sub | accepted | user |
| [0016](0016-infrastructure-opentofu.md) | Infrastructure as code with OpenTofu, provider-neutral, AWS default | accepted | user |
| [0017](0017-testing-strategy.md) | Testing: unit tests for all code, ≥ 80 % coverage, e2e when possible | accepted | user+default |
| [0018](0018-dev-bootstrap-mise.md) | Developer machine bootstrap with mise | superseded | user |
| [0019](0019-agent-approval-before-bootstrap.md) | Owner approval before agent bootstrap | accepted | user |
