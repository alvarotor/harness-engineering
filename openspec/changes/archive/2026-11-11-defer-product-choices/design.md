# Design

## Approach

The base stays generic (P-0016). Each product choice that is not made yet becomes an open question with a
`needed_by` trigger in `docs/decisions/registry.yaml`. Before work reaches a trigger, the orchestrator runs
`docs/standards/new-product.md` with the owner; the answer becomes a decision and the question moves to
`docs/decisions/history.yaml` (R-REG-006 keeps the two files consistent).

Decided choices that no longer fit the generic base are handled by status, not deletion:

| Entry | Status | Replacement | Reason |
|---|---|---|---|
| D-0006 frontends (Preact SPA) | superseded | D-0024 / ADR-0024 (TypeScript, framework deferred Q11) | framework is a product choice |
| D-0007 proto + Connect/REST | superseded | D-0022 / ADR-0022 (proto + gRPC, browser API deferred Q8) | browser API is a product choice |
| D-0010, D-0013 deploy and CI | superseded | D-0023 / ADR-0023 | tied to OpenTofu VM and CI for services that do not exist |
| D-0005 identity (Better Auth) | deprecated | none (Q10) | identity is a product choice |
| D-0015 Redis pub/sub | deprecated | none (Q7) | messaging is a product choice |
| D-0016 OpenTofu/AWS | deprecated | none (Q9) | hosting is a product choice |
| P-0001 B2C roles | deprecated | none (Q6) | product users are unknown |

Superseded and deprecated entries keep their ADR files with the status line updated (R-REG-002); the
decision text is not edited (AGENTS.md §5).

## Alternatives considered

- **Keep the entries accepted and mark them `planned`:** rejected, because the harness would keep enforcing
  rules for services that do not exist.
- **Delete the entries:** rejected, because the history of the choices is needed when a trigger is reached
  (ADR-0001).

## Risks

- A trigger that is not noticed. Mitigation: `needed_by` is read by the orchestrator at each proposal (see
  `openspec/config.yaml`).
