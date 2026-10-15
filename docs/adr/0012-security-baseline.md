# 0012 Security baseline OWASP ASVS L1

- Status: accepted
- Date: 2026-10-15
- Source: user

## Decision
- OWASP ASVS Level 1 is the minimum for all deployables. `docs/standards/security.md` maps each
  applicable control to how it is met (code, config, or a CI check).
- CI gates:
  - gitleaks
  - Go: `govulncheck`
  - TS: `pnpm audit`
  - Trivy (filesystem, IaC and image; fails on fixable HIGH/CRITICAL)
  - Syft SBOM per image
  - Dependabot
- Containers: non-root, `cap_drop: ALL`, `no-new-privileges`, read-only root filesystem where possible.
- **GDPR applicability is undetermined** (launch country unknown). Each service keeps a minimal
  personal-data inventory in the service catalog, so the decision is cheap once the market is known.

## Revisit when
The launch country is known, a payment or health data requirement appears, or a SOC 2 request arrives.
