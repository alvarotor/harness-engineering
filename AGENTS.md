# AGENTS.md

Rules for every AI agent and human in this repository. Scoped `AGENTS.md` files add rules for their
folder and never relax these. The repository is the memory; never rely on chat history for anything durable.
**Stage:** way of working + first basic architecture (P-0005). This repo is a reusable product base (P-0016):
the product is not chosen yet. **Open questions with a `needed_by` trigger (Q6–Q11) are asked, not assumed:** before
work that reaches a trigger, stop and ask the owner ([`docs/standards/new-product.md`](docs/standards/new-product.md)).

## 1. Session start
1. Read [`docs/decisions/registry.yaml`](docs/decisions/registry.yaml) (current decisions only).
   [`history.yaml`](docs/decisions/history.yaml) is read only when changing a decision.
2. Read your role in [`docs/agents/`](docs/agents/README.md) and the scoped `AGENTS.md` of folders you touch.
3. Read [`ARCHITECTURE.md`](ARCHITECTURE.md), ADRs and `docs/standards/` **only for the sections your task touches**.
4. `openspec list` for active changes, if the task is lane M or L.

## 2. Precedence (highest first)
Owner's current request > accepted OpenSpec specs and the approved change > accepted ADRs/registry >
[`docs/standards/`](docs/standards/) > scoped `AGENTS.md` > this file. If the owner's request contradicts a
lower layer, report the conflict (§4) and continue only after the owner confirms (then it is a decision change, §5).

## 3. Memory protocol
When the owner states a durable preference, constraint, convention, correction or decision, persist it before continuing:
add or update the registry entry (`source: user`, rationale, `enforced_by`/`planned_enforced_by`) and a
changelog line in `history.yaml`; architectural ones also get an ADR and an index row. Then update what it
affects: standards, `ARCHITECTURE.md`, catalog, templates, harness rules (`tools/harness` +
`docs/standards/validation.md`), scoped `AGENTS.md`, `docs/agents/*`, OpenSpec specs, existing services.
Run `make check`.

## 4. Conflict reporting
Stop the conflicting work and report:
```text
CONFLICT
- Request: <what was asked>
- Conflicts with: <D-/P-/spec id> (<file>): "<recorded text>"
- Impact: <what would change or break>
- Options: (a) keep recorded decision, (b) supersede via new ADR <NNNN>, (c) scoped exception
- Recommendation: <one option + one-line reason>
```

## 5. Changing an accepted decision
Only through an owner-approved OpenSpec change: new ADR with `Supersedes: NNNN`, old ADR and registry entry
`superseded` (+ `superseded_by`, moved to `history.yaml`), everything in the old `enforced_by` updated, affected
services migrated. Never edit the decision text of an accepted ADR. Details: [`CONTRIBUTING.md`](CONTRIBUTING.md).

## 6. Lanes (P-0015)
The orchestrator classifies every request. When in doubt, take the higher lane.

| Lane | When | Flow |
|---|---|---|
| S | fix, refactor, docs or tests; **no observable behavior change** | implementer → `make check` |
| M | feature inside an existing service; no new decision | light OpenSpec change (proposal, delta spec, tasks) → **owner approves** → implementer (code + tests + docs) → reviewer (also checks docs) → archive |
| L | new service/deployable, decision, `infra/` or `.github/`, contract break, destructive migration, any [security trigger](docs/standards/security.md#security-review-triggers) | full flow in [`docs/agents/README.md`](docs/agents/README.md#flow-run-by-the-orchestrator) |

Never claim a check passed unless you ran it and saw it pass (P-0004).

## 7. Roles and approvals
Roles: [`docs/agents/`](docs/agents/README.md). The orchestrator is the default entry point and never edits.
Every role documents its own work and lists the docs in its handoff (P-0010).
Owner approval is always required for: accepting or superseding decisions, installing developer tools
(D-0021), new infrastructure or external services, security-sensitive changes, production deploys,
destructive migrations, lowering a quality gate.

## 8. Hard rules (the source wins)
| Rule | Source |
|---|---|
| Backends in Go (hexagonal-lite) unless a registry exception exists | D-0004 |
| One service per business domain; own DB on the shared Postgres server | D-0014, D-0008 |
| Sync between services = gRPC; contracts in `contracts/proto`, versioned, no breaking changes in place | D-0022 |
| Deployables never import each other; `packages/` is build-time only | D-0002 |
| Every deployable: container, non-root, env config, ops port 8081 `/livez` `/readyz` | D-0003, D-0009 |
| No secrets in git; Traefik is the only exposed container | D-0009, D-0012 |
| Unit tests for all code, coverage ≥ 80 %, e2e for user flows when possible | D-0017 |
| Every command runs through `make`; CI calls `make` targets; prod only from a tag `vX.Y.Z` with owner approval | D-0002, D-0023 |
| No new infrastructure (broker, gateway, cache, Kubernetes…) before its trigger | P-0002 |

## 9. Commands
`make help` · `make check` (run before every handoff) · `make harness-test` · `make spec-validate` · `make doctor`.
Run `make doctor` only when the owner asks (P-0014). It is diagnostic-only. For missing tools, give the install
guidance and wait for explicit owner approval before any installer; if it needs `sudo`, give the owner the
command (D-0021, P-0007). Never install a user-space copy named `make`.
