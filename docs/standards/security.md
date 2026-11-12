# Security standard

Baseline: **OWASP ASVS Level 1**. This file maps the applicable areas to how they are met.
**Status** is honest: `planned` means not enforced yet. Rows marked *Q7* depend on the identity choice.

| ASVS area | How we meet it | Enforced by | Status |
|---|---|---|---|
| V2/V3 Authentication & sessions | Decided with identity (Q7). Requirements: no custom crypto; short-lived tokens; session cookies `HttpOnly; Secure; SameSite=Lax`; rate-limited login endpoints | identity component config + tests | planned (Q7) |
| V4 Access control | Deny by default; coarse role from the token checked in each service's application layer; resource ownership checks | service unit tests + e2e | planned |
| V5 Validation & encoding | Proto/schema validation at the edge; parameterized SQL only; output encoding by the frontend framework | lint + code review | planned |
| V6 Cryptography | No custom crypto; TLS terminated at Traefik (ACME) | Traefik config | planned |
| V7 Logging | No secrets, tokens, cookies or personal data in logs; JSON with correlation IDs | code review + log tests | planned |
| V8 Data protection | Personal-data inventory per deployable in the catalog; minimum PII in tokens | `docs/catalog/services.yaml` | active |
| V9 Communications | HTTPS only on public hosts, HSTS; internal traffic on private Docker networks | Traefik config | planned |
| V10 Malicious code / dependencies | Lockfiles; gitleaks, `govulncheck` and Trivy in CI; images and actions pinned by digest/SHA | CI + `tools/harness` (R-COMPOSE-004, R-IMG-001, R-CI-001) | active |
| V12/V13 API | CORS allowlist per env (frontend origins only); no wildcard origins with credentials | service config tests | planned |
| V14 Configuration | Secrets only from GitHub Environments, as files with mode 0600; SSH host keys pinned; non-root, `cap_drop: ALL`, `no-new-privileges`, read-only rootfs | CI + `tools/harness` (R-COMPOSE-*, R-IMG-001, R-CI-*) | active |

## Known accepted risks (template defaults)
Accepted for the template. A product's owner reviews and confirms them when starting the product
([`new-product.md`](new-product.md) checklist); each later change to this table needs owner approval.

| Risk | Mitigation / trigger |
|---|---|
| No database backups | Mandatory before real user data in prod (decided with Q6) |
| Traefik has Docker API access (through a proxy) | Proxy allow-lists only the `GET` paths Traefik needs (ping, version, events, container list/inspect, networks); `archive`, `export`, `logs`, `attach` and all writes are denied; isolated network, non-root. Container inspect still shows config and env names, so secrets are passed as files (`<NAME>_FILE`), never in env values or labels |
| Deploy user is in the `docker` group (root-equivalent on the host) | One dedicated host per environment; SSH only with the deploy key and a pinned host key. Revisit with hosting (Q6) |
| CI installs OpenSpec with `npm install -g` without a lockfile | Exact version pinned; CI-only. Revisit when the repo has a Node lockfile |
| Committed dev-only Postgres password in `infra/env/local/` | Local only; deploy never reads `infra/env/local` |
| Unconfigured prod still shows the approval prompt before the deploy skips | Approve only when a real host is configured |
| No HSTS or security headers yet | Add a Traefik headers middleware before the first public route |
| Fixable HIGH/CRITICAL CVEs in third-party infra images (e.g. Postgres OS packages and bundled `gosu`) | Internal networks only, never published; a non-blocking CI job reports them; bump pinned digests when upstream rebuilds |
| Postgres starts as root for its init, then drops to uid 70 | Official image behaviour; data network only |
| Traefik may bind any port (`ip_unprivileged_port_start=0` in its own namespace) | Traefik is the only service with published ports; runs as non-root |
| No image signing or provenance | Images deployed by digest from GHCR; add signing when a second publisher or third-party consumers appear |
| Post-deploy smoke test accepts any status below 500 | No public route exists yet; require an expected status on a health path with the first public route |
| Traefik does not reject encoded dots (`%2e`) in paths (no option exists) | Go `net/http` routing cleans paths; backends must normalise and reject `..`; add a middleware with the first public route if needed |
| GDPR applicability undetermined | Personal-data inventory kept current; decide once the launch country is known (Q3) |

## Security review triggers
The orchestrator calls the security-reviewer, and the owner must approve, when a change touches any of these:

| Trigger | Examples |
|---|---|
| Identity and access | the identity component; token verification; role or permission checks; session/cookie/token settings |
| Exposure | Traefik labels or config, published ports, networks, the socket proxy, firewall rules, CORS allowlists |
| Secrets and config | env handling of secrets, `infra/env/**`, GitHub Environments usage, `.github/workflows/**` permissions, deploy scripts |
| Supply chain | new or upgraded dependencies (`go.mod`, `package.json`, lockfiles), base images, GitHub Actions versions |
| Data | new personal-data fields, contracts or events carrying PII, migrations on tables with PII, logging changes |
| Containers | `Dockerfile` users, capabilities, mounts |
| This standard | any edit to this file or to accepted risks |
