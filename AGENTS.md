# AGENTS.md

Rules for every AI agent and human in this repository. Scoped `AGENTS.md` files add rules for their
folder and never relax these. The repository is the memory; never rely on chat history for anything durable.
**Stage:** way of working + first basic architecture; no product features yet. This repo is a reusable product base.
Each product is its own repository created from this template; this repo never holds a product
([`docs/standards/new-product.md`](docs/standards/new-product.md)). The product is not chosen yet.
**Open questions with a `needed_by` trigger (Q1–Q8) are asked, not assumed:** before work that reaches a trigger,
stop and ask the owner.

## 1. Session start
1. Read [`docs/decisions/registry.yaml`](docs/decisions/registry.yaml) (current decisions and open questions).
   [`history.yaml`](docs/decisions/history.yaml) is read only when changing a decision.
2. Read your role in [`docs/agents/`](docs/agents/README.md) and the scoped `AGENTS.md` of folders you touch.
3. Read [`ARCHITECTURE.md`](ARCHITECTURE.md), ADRs and `docs/standards/` **only for the sections your task touches**.
4. `openspec list` for active changes, if the task is lane M or L.

## 2. Precedence (highest first)
Owner's current request > accepted OpenSpec specs and the approved change > accepted ADRs/registry >
[`ARCHITECTURE.md`](ARCHITECTURE.md) and [`docs/standards/`](docs/standards/) > scoped `AGENTS.md` > this file.
If the owner's request contradicts a lower layer, report the conflict (§4) and continue only after the owner
confirms (then it is a decision change, §5).

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
- Conflicts with: <decision/preference/spec id, or file and section> (<file>): "<recorded text>"
- Impact: <what would change or break>
- Options: (a) keep the recorded rule, (b) change it through a new decision (ADR <NNNN>), (c) scoped exception
- Recommendation: <one option + one-line reason>
```

## 5. Changing a rule
Only through an owner-approved OpenSpec change. Changing an accepted decision: new ADR with `Supersedes: NNNN`, old
ADR and registry entry `superseded` (+ `superseded_by`, moved to `history.yaml`), everything in the old `enforced_by`
updated, affected services migrated. Never edit the decision text of an accepted ADR. Changing a starting rule in
`ARCHITECTURE.md` or `docs/standards/` records the new choice as a decision. Details: [`CONTRIBUTING.md`](CONTRIBUTING.md).

## 6. Lanes
The orchestrator classifies every request. When in doubt, take the higher lane.

| Lane | When | Flow |
|---|---|---|
| S | fix, refactor, docs or tests; **no observable behavior change** | implementer → `make check` |
| M | feature inside an existing service; no new decision | light OpenSpec change (proposal, delta spec, tasks) → **owner approves** → implementer (code + tests + docs) → reviewer (also checks docs) → archive |
| L | new service/deployable, decision, `infra/` or `.github/`, contract break, destructive migration, any [security trigger](docs/standards/security.md#security-review-triggers) | full flow in [`docs/agents/README.md`](docs/agents/README.md#flow-run-by-the-orchestrator) |

Archiving needs no owner approval. Never claim a check passed unless you ran it and saw it pass.

## 7. Roles and approvals
Roles: [`docs/agents/`](docs/agents/README.md). The orchestrator is the default entry point and never edits.
Every role documents its own work in the same step and lists the docs in its handoff.
Never silently make expensive, security-sensitive or hard-to-reverse decisions: ask the owner.
Owner approval is always required for (the only list; other files link here): starting a lane M or L change
(proposal approval), accepting or superseding decisions, installing developer tools, new infrastructure or external
services, security-sensitive changes, production deploys, infrastructure-as-code `apply`, destructive migrations,
lowering a quality gate.

## 8. Hard rules (the source wins)
| Rule | Source |
|---|---|
| Backends in Go (hexagonal-lite) unless a registry exception exists | [engineering.md](docs/standards/engineering.md#go-default-backend) |
| One service per business domain; own DB on the shared Postgres server | [ARCHITECTURE §2, §4](ARCHITECTURE.md#2-service-boundaries) |
| Sync between services = gRPC; contracts in `contracts/proto`, versioned, no breaking changes in place | [`contracts/AGENTS.md`](contracts/AGENTS.md) |
| Deployables never import each other; `packages/` is build-time only | [ARCHITECTURE §2](ARCHITECTURE.md#2-service-boundaries) |
| Every deployable: container, non-root, env config, ops port 8081 `/livez` `/readyz` | [ARCHITECTURE §9](ARCHITECTURE.md#9-service-runtime-contract-every-deployable) |
| No secrets in git; Traefik is the only exposed container | [ARCHITECTURE §7](ARCHITECTURE.md#7-security-model-baseline-owasp-asvs-l1) |
| Unit tests for all code, coverage ≥ 80 %, e2e for user flows when possible | [testing.md](docs/standards/testing.md) |
| Every command runs through `make`; CI calls `make` targets; prod only from a tag `vX.Y.Z` with owner approval | [ARCHITECTURE §6](ARCHITECTURE.md#6-delivery-and-environments) |
| No new infrastructure (broker, gateway, cache, Kubernetes…) before its trigger: start simple and reversible | [ARCHITECTURE §13](ARCHITECTURE.md#13-deferred-components-and-trigger-conditions) |

## 9. Commands
`make help` · `make check` (run before every handoff) · `make harness-test` · `make spec-validate` · `make doctor`.
Run `make doctor` only when the owner explicitly asks to check developer-tool availability. It is diagnostic-only.
For missing tools, give the install guidance and wait for explicit owner approval before any installer; if it needs
`sudo`, give the owner the command. Never install a user-space copy named `make`.
