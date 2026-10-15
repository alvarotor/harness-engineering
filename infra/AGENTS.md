# infra/AGENTS.md

Environment and deployment infrastructure. Adds to the root [`AGENTS.md`](../AGENTS.md).
The devops role owns this folder ([`docs/agents/devops.md`](../docs/agents/devops.md)).

Planned layout:
- `compose/`
- `traefik/`
- `postgres/`
- `env/{local,dev,prod}/`
- `tofu/`
- `k8s/` (reserved; empty until an ADR approves Kubernetes)

Rules:
- **Same images everywhere** (D-0003). Compose base plus `local`/`deploy` overlays. Deploy by digest, never by
  mutable tag. No `container_name`, no `links`, no host-path volumes for app state.
- **Edge** (D-0009):
  - Traefik is the only service with published ports.
  - Routes come from labels with `exposedByDefault=false`.
  - Only the socket proxy mounts `/var/run/docker.sock` (read-only, `POST=0`).
  - Labels never route port 8081.
- **Networks:**
  - `edge`: Traefik and the deployables.
  - `data`: Postgres, Redis, backends and migrate jobs.
  - `socket`: Traefik and the socket proxy.
  - Frontends are never on `data`.
- **Config:** `env/<env>/` holds **non-secret** values only. Secrets come from GitHub Environments at deploy time (D-0010).
  Committing a secret is a stop-the-line incident: rotate it, then purge it from history.
- **OpenTofu** (D-0016):
  - Capability modules (`modules/<capability>/`) define provider-neutral interfaces; the AWS implementation lives in `modules/<capability>/aws/`.
  - `envs/{dev,prod}` only compose capabilities.
  - State is encrypted.
  - `apply` is manual, with owner approval.
- **No new infrastructure components** (broker, gateway, cache, cluster, managed DB) without an accepted decision (P-0002).
- **Owner approval is required** for: prod deploys, `tofu apply`, firewall/exposure changes, data volume changes.
