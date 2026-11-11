# Proposal

## Why

This repository is a reusable product base (P-0016): the product is not chosen yet. Several recorded
decisions fixed product choices too early (identity provider, frontend framework, messaging, OpenTofu/AWS,
delivery model), and the harness carried rules for services that do not exist. Agents rebuilt ~20 rules per
service from prose. The owner asked to defer product choices until their trigger and to keep the base generic.

## Lane and security triggers

- **Lane: L** (decisions and ADRs; `docs/standards/**`, `ARCHITECTURE.md`, `openspec/config.yaml` and
  `docs/catalog/**` change; the decision change is owner-approved).
- Security triggers (`docs/standards/security.md`): **this standard** (the file is rewritten; accepted risks change)
  and **identity and access** (the identity requirements are deferred to Q10 and D-0005 is deprecated). The owner
  approves the change and the security-reviewer reviews it before the reviewer step.

## What Changes

- **Deprecated (no replacement):** D-0005 (Better Auth, magic link), D-0015 (Redis pub/sub), D-0016
  (OpenTofu/AWS), P-0001 (B2C roles).
- **Superseded:** D-0006 by D-0024 (ADR-0024: TypeScript frontends, framework deferred); D-0007 by D-0022
  (ADR-0022: protobuf + gRPC, browser API style deferred); D-0010 and D-0013 by D-0023 (ADR-0023: delivery
  with `change/<name>` branches, `pr.yml`, `main.yml` builds once and deploys dev, tag `vX.Y.Z` deploys prod
  after owner approval).
- Superseded, deprecated and resolved entries move to `docs/decisions/history.yaml`. D-0004 loses its
  `services/auth` exception; assumptions A3–A5 are removed.
- **Open questions with triggers:** Q6 (product), Q3 (email and domain), Q4 (launch country), Q7 (messaging),
  Q8 (browser API), Q9 (hosting and IaC), Q10 (identity), Q11 (frontend framework). Each has a `needed_by`
  trigger. New `docs/standards/new-product.md` is the checklist the orchestrator runs with the owner before
  work reaches a trigger.
- **Slimmed docs:** `ARCHITECTURE.md`, `security.md`, `engineering.md`, `testing.md`, scoped `AGENTS.md`
  files and `openspec/config.yaml` no longer describe the deferred stack. `docs/catalog/services.yaml` is empty.

## Capabilities

### Modified Capabilities

- `project-memory`: deferred choices are recorded as open questions with a trigger and asked at that trigger.

## Decision change

This change supersedes or deprecates decisions the owner approved in conversation. Superseding ADRs
(owner-approved as part of this change):

- ADR-0022 (`docs/adr/0022-contracts-proto-grpc.md`) supersedes D-0007.
- ADR-0023 (`docs/adr/0023-delivery-branches-ci-releases.md`) supersedes D-0010 and D-0013.
- ADR-0024 (`docs/adr/0024-frontends-typescript.md`) supersedes D-0006.

Deprecated without replacement: D-0005, D-0015, D-0016, P-0001.

## Impact

- Decisions: D-0004, D-0005, D-0006, D-0007, D-0010, D-0013, D-0015, D-0016, D-0022, D-0023, D-0024,
  P-0001, P-0016.
- Files: `docs/decisions/{registry,history}.yaml`, `docs/adr/*`, `docs/standards/*` (new `new-product.md`),
  `ARCHITECTURE.md`, scoped `AGENTS.md`, `openspec/config.yaml`, `docs/catalog/services.yaml`,
  `docs/agents/*`, `.opencode/agents/*`, `tools/harness/internal/check`.
- No application code, infrastructure or product features change.
