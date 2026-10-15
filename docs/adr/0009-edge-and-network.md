# 0009 Traefik edge via Docker labels and socket proxy, subdomains, ops port

- Status: accepted
- Date: 2026-10-15
- Source: user

## Decision
- **Traefik** (v3, pinned) is the only container publishing ports (80/443). It handles automatic TLS
  via ACME once a domain exists, plus shared middlewares (security headers, compression, optional IP allowlist for admin).
- Routing is declared with **Compose labels** on each service.
  - `providers.docker.exposedByDefault=false`.
  - Only public services set `traefik.enable=true`.
  - The load-balancer port is always `8080`, never `8081`.
- Traefik reaches the Docker API through a **socket proxy** (pinned image), which:
  - allows only read-only container/network listing (`CONTAINERS=1`, `NETWORKS=1`, `POST=0`, everything else denied)
  - is reachable only on an internal `socket` network shared with Traefik
  - is the only container that mounts the socket (read-only); Traefik never mounts it
- Hosts come from env: `${DOMAIN}`, `admin.${DOMAIN}`, `api.${DOMAIN}`, `auth.${DOMAIN}`. Locally: `*.localhost`.
- Networks:
  - `edge`: Traefik and the deployables.
  - `data`: Postgres, Redis, backends, migrate jobs.
  - `socket`: Traefik and the socket proxy.
  - Frontends are not on `data`.
- Every deployable serves `/livez` and `/readyz` on port 8081, which is never routed.
- Traefik's own dashboard and API are disabled in dev/prod and enabled locally only.

## Consequences
- Docker API access is a privilege, minimised by the proxy. The harness checks that only the socket proxy mounts the socket.
- Labels translate directly to Traefik IngressRoute on Kubernetes.
