# Security standard

Baseline: **OWASP ASVS Level 1** (D-0012). This file maps the applicable areas to how they are met.
**Status** is honest: `planned` means not enforced yet.

| ASVS area | How we meet it | Enforced by | Status |
|---|---|---|---|
| V2/V3 Authentication & sessions | Better Auth magic link; single-use, short-lived (≤ 15 min) links; session cookie `HttpOnly; Secure; SameSite=Lax`, host-only; rate-limited endpoints | `services/auth` config + tests | planned |
| V4 Access control | JWT `role` claim (`customer`/`support`/`admin`) checked in each service's application layer; resource ownership checks; deny by default | service unit tests + e2e | planned |
| V5 Validation & encoding | Proto/schema validation at the edge; parameterized SQL only (sqlc / query builder); output encoding by Preact | lint + code review | planned |
| V6 Cryptography | No custom crypto. EdDSA JWT via Better Auth; TLS terminated at Traefik (ACME) | ADR-0005, ADR-0009 | planned |
| V7 Logging | No secrets, tokens, cookies, emails or magic-link URLs in logs; JSON with correlation IDs | code review + log tests | planned |
| V8 Data protection | Personal-data inventory per deployable in the catalog; minimum PII in JWT (`sub`, `role`, `email`) | `docs/catalog/services.yaml` | active |
| V9 Communications | HTTPS only on public hosts, HSTS; internal traffic on private Docker networks | Traefik config | planned |
| V10 Malicious code / dependencies | Lockfiles; Dependabot; `govulncheck`; `pnpm audit`; Trivy (fs + image + IaC); Syft SBOM | CI | planned |
| V12/V13 API | CORS allowlist per env (frontend origins only); no wildcard origins with credentials | service config tests | planned |
| V14 Configuration | Secrets only from GitHub Environments / env files with mode 0600; gitleaks on every PR; non-root, `cap_drop: ALL`, `no-new-privileges`, read-only rootfs | CI + `tools/harness` compose rules | planned |

## Known accepted risks (owner decisions)
| Risk | Decision | Mitigation / trigger |
|---|---|---|
| Staff accounts protected only by email (magic link) | D-0005 | Add TOTP/passkeys for staff before staff use prod with real data |
| No database backups | D-0008 | Mandatory before real user data in prod |
| Redis pub/sub loses events | D-0015 | Only loss-tolerable events; upgrade path to Streams/RabbitMQ |
| Traefik has Docker API access (through a proxy) | D-0009 | Read-only socket proxy, isolated network, `POST=0` |
| GDPR applicability undetermined | D-0012 | Personal-data inventory kept current; decide once the launch country is known (Q4) |

## Security review triggers
The orchestrator calls the security-reviewer, and the owner must approve, when a change touches any of these:

| Trigger | Examples |
|---|---|
| Identity and access | `services/auth/**`; JWT/JWKS verification; role or permission checks; session/cookie/token settings |
| Exposure | Traefik labels or config, published ports, networks, the socket proxy, firewall rules in `infra/tofu`, CORS allowlists |
| Secrets and config | env handling of secrets, `infra/env/**`, GitHub Environments usage, `.github/workflows/**` permissions |
| Supply chain | new or upgraded dependencies (`go.mod`, `package.json`, lockfiles), base images, GitHub Actions versions |
| Data | new personal-data fields, contracts or events carrying PII, migrations on tables with PII, logging changes |
| Containers | `Dockerfile` users, capabilities, mounts |
| This standard | any edit to this file or to accepted risks |
