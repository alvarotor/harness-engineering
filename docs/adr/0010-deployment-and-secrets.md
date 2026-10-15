# 0010 VM deployment, digest promotion, GitHub Environments secrets

- Status: accepted
- Date: 2026-10-15
- Source: user

## Decision
- One Linux VM per environment (`dev`, `prod`), provisioned with OpenTofu (ADR-0016) and running Docker Compose.
- Images: `ghcr.io/${{ github.repository_owner }}/har-<service>`, tagged with the git SHA and deployed
  **by digest**. No owner is hard-coded; the repo does not exist yet.
- `dev` deploys automatically on merge to `main`. `prod` deploys by manually promoting a release
  manifest already deployed to dev, gated by GitHub Environment required reviewers.
- Secrets live in GitHub Environment secrets. Deploy writes `/etc/har/<env>.env` (`0600 root`) on the VM.
  Non-secret config lives in `infra/env/<env>/`.
- Deploy: pull → migrate → `up -d --wait` → smoke test. Rollback = redeploy the previous manifest.
- No domain yet. TLS and DNS are wired through `DOMAIN` config when one exists.

## Revisit when
More than one VM per env, a rotation or audit requirement, or Kubernetes. Then move to SOPS, AWS Secrets Manager, or Vault.
