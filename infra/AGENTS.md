# infra/AGENTS.md

Environment and deployment infrastructure. Adds to the root [`AGENTS.md`](../AGENTS.md).
The devops role owns this folder ([`docs/agents/devops.md`](../docs/agents/devops.md)).

Layout:
- `compose/`: `compose.yaml` (base), `compose.local.yaml`, `compose.deploy.yaml`. Each deployable adds its own
  `compose.yaml` fragment in its folder; `make up` merges them all.
- `traefik/`: static config.
- `postgres/`: init scripts (databases + roles per service).
- `env/{local,dev,prod}/`: non-secret config (local may hold dev-only values).
- `deploy/`: deploy script (SSH + Compose, D-0023).
- Planned: `k8s/` only after an ADR approves Kubernetes; infrastructure as code with Q9.

Rules:
- **Same images everywhere** (D-0003). Deploy by digest, never by mutable tag. No `container_name`, no `links`,
  no host-path volumes for app state.
- **Edge** (D-0009):
  - Traefik is the only service with published ports.
  - Routes come from labels with `exposedByDefault=false`.
  - Only the socket proxy mounts `/var/run/docker.sock` (read-only, `POST=0`).
  - Labels never route port 8081.
- **Networks:**
  - `edge`: Traefik and the deployables.
  - `data`: data stores, backends and migrate jobs.
  - `socket`: Traefik and the socket proxy.
  - Frontends are never on `data`.
- **Config:** `env/<env>/` holds **non-secret** values only. Secrets come from GitHub Environments at deploy time (D-0023).
  Committing a secret is a stop-the-line incident: rotate it, then purge it from history.
- **Delivery** (D-0023): merge to `main` deploys dev; tag `vX.Y.Z` deploys the same digests to prod after owner approval.
- **No new infrastructure components** (broker, gateway, cache, cluster, managed DB, IaC) without an accepted decision (P-0002).
- **Owner approval is required** for: prod deploys, exposure or firewall changes, data volume changes.
