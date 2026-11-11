# 0016 Infrastructure as code with OpenTofu, provider-neutral, AWS default

- Status: deprecated
- Deprecated by: product choices deferred to Q9 (P-0016)
- Date: 2026-10-15
- Source: user

## Decision
- All cloud infrastructure is declared with **OpenTofu** (version pinned in `required_version`) under `infra/tofu/`.
  Local development needs no cloud resources (Compose only).
- Provider neutrality is achieved by **capability modules**:
  - `modules/<capability>/` defines the provider-agnostic interface (`variables.tf`, `outputs.tf`). Capabilities:
    - `network`
    - `host` (VM + cloud-init with Docker + persistent volume + static IP)
    - `dns`
  - `modules/<capability>/aws/` is the default implementation. Another provider adds a sibling folder; env compositions do not change shape.
  - `envs/{dev,prod}/` compose capabilities and select the provider.
- Workloads must not depend on provider-specific services (SQS, Cognito, RDS-only features…)
  without an ADR. Postgres and Redis run as containers on the host.
- State: S3 backend with native lockfile (AWS default), with **OpenTofu state encryption** enabled.
  A one-time `infra/tofu/bootstrap/` creates the state bucket.
- Firewall: 80/443 public. SSH only from an allowlisted CIDR, as the deploy user only.
- CI: `fmt -check`, `validate`, Trivy IaC scan and `plan` on PR. `apply` is manual per env with approval.

## Revisit when
Kubernetes is adopted (add a `cluster` capability; EKS default), or managed data services are approved.
