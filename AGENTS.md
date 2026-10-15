# AGENTS.md

Instructions for every AI agent and human contributor in this repository. Scoped `AGENTS.md`
files in subfolders add rules for that folder. They never relax the rules here.

**Current stage:** way of working + first basic architecture (P-0005). Do not build product
features unless the owner asks.

## 1. Start of every session

0. **Tools check:** run `make doctor` (or `make-proj doctor`). If `make` is not found or `doctor` fails, run
   `./scripts/bootstrap.sh` and re-check (D-0018). If it reports steps that need sudo or Docker, stop and give the
   owner the exact commands it printed (P-0007).

Then read, in this order:
1. [`docs/decisions/registry.yaml`](docs/decisions/registry.yaml): decisions, preferences, assumptions and open questions.
2. [`ARCHITECTURE.md`](ARCHITECTURE.md), plus the ADRs in [`docs/adr/`](docs/adr/) that are relevant to the task.
3. Active OpenSpec changes: `openspec list`. Accepted specs: `openspec list --specs`.
4. The scoped `AGENTS.md` of every folder you will touch.
5. Your role definition in [`docs/agents/`](docs/agents/README.md).

The repository is the memory. Do not rely on conversation history for anything that must survive the session.

## 2. Instruction precedence (highest first)

1. The explicit current request from the owner.
2. Accepted OpenSpec specs (`openspec/specs/`) and the approved change being applied.
3. Accepted ADRs and registry decisions.
4. Engineering standards ([`docs/standards/`](docs/standards/)).
5. Scoped `AGENTS.md`.
6. This file and repository defaults.

A lower layer may add detail but must never contradict a higher one. If **the owner's request** contradicts
layers 2–6, it does not silently win. Report the conflict (§4) and proceed only after the owner confirms.
The confirmation then becomes a decision change (§5).

## 3. Memory protocol: when the owner states something durable

When the owner states a preference, constraint, convention, correction or decision, persist it **before** continuing:

1. **Record it.** Add or update the entry in `docs/decisions/registry.yaml`:
   - `source: user`
   - the rationale
   - an entry in `changelog`
   - architectural ones also get an ADR (`docs/adr/`) and a row in the ADR index
2. **Update standards** in `docs/standards/` and `ARCHITECTURE.md` if affected.
3. **Update** templates, generators and the service catalog (`docs/catalog/services.yaml`).
4. **Update validation** in `tools/harness` and `docs/standards/validation.md` where the rule is checkable.
5. **Update** affected scoped `AGENTS.md` and `docs/agents/*`.
6. **Update OpenSpec**: `openspec/config.yaml` rules and affected specs, through a change.
7. **Apply it** to existing affected services, in the same change or a tracked follow-up change.
8. **List the files** that now enforce it under `enforced_by` (or `planned_enforced_by`).

If it contradicts an accepted decision, it is a decision change (§5), not a new entry.
Run `make check` when done.

## 4. Conflict reporting

Report conflicts in this exact form and stop the conflicting work:

```text
CONFLICT
- Request: <what was asked>
- Conflicts with: <D-/P-/spec id> (<file>): "<recorded text>"
- Impact: <what would change or break>
- Options: (a) keep recorded decision, (b) supersede via new ADR <NNNN>, (c) scoped exception
- Recommendation: <one option + one-line reason>
```

## 5. Changing an accepted decision

Only through an OpenSpec change approved by the owner. The change contains:
- a new ADR with `Supersedes: NNNN`
- the old ADR set to `superseded`
- the registry entry updated: old `status: superseded` + `superseded_by`, plus a new entry
- updates to everything listed in the old entry's `enforced_by`, and migration of affected services

Never edit the decision text of an accepted ADR. Details: [`CONTRIBUTING.md`](CONTRIBUTING.md).

## 6. Workflow

`explore → propose (OpenSpec change) → owner approves → apply (tasks) → make check → review → archive`

- Changes to behavior, architecture, contracts, infrastructure or the harness go through OpenSpec
  (`/opsx-propose`, `/opsx-apply`, `/opsx-archive` in OpenCode).
- Typos and doc fixes that don't change meaning can skip OpenSpec, but still need `make check`.
- Never claim a check passed unless you ran it and saw it pass (P-0004).

## 7. Roles

Defined in [`docs/agents/`](docs/agents/README.md): **orchestrator** (the entry point; sequences the work and never edits),
**architect**, **planner**, **implementer**, **test-engineer**, **security-reviewer** (only when a
[security trigger](docs/standards/security.md#security-review-triggers) matches), **reviewer**, **documentation**
(verifies each role documented its work), **devops**. OpenCode adapters live in `.opencode/agents/`. Select `orchestrator` as the primary agent for normal work.

**Every role documents its own work** in the same step, following the
[documentation map](docs/agents/README.md#documentation-map), and lists the docs in its handoff (P-0010).

Owner approval is always required for:
- accepting or superseding decisions
- adding new infrastructure components or external services
- security-sensitive changes
- production deploys
- destructive migrations
- lowering a quality gate

## 8. Hard rules (summary; the source wins)

| Rule | Source |
|---|---|
| Backends in Go (hexagonal-lite) unless a registry exception exists | D-0004 |
| One service per business domain; own DB on the shared Postgres server | D-0014, D-0008 |
| Sync between services = gRPC; async = Redis pub/sub, only for events whose loss is tolerable | D-0007, D-0015 |
| Contracts in `contracts/proto`, versioned; no breaking changes in place | D-0007 |
| Deployables never import each other; `packages/` is build-time only | D-0002 |
| Every deployable: container, non-root, env config, ops port 8081 `/livez` `/readyz` | D-0003, D-0009 |
| No secrets in git; Traefik is the only exposed container | D-0009, D-0010, D-0012 |
| Unit tests for all code, coverage ≥ 80 %, e2e for user flows when possible | D-0017 |
| Infra in OpenTofu, provider-neutral modules, AWS default | D-0016 |
| Every command runs through `make` | D-0002, D-0013 |
| No new infrastructure (broker, gateway, cache, Kubernetes…) before its trigger | P-0002 |

## 9. Commands

```sh
make help            # list targets
make check           # all repository validations (run before every handoff)
make harness-test    # harness unit tests + coverage gate
make spec-validate   # openspec strict validation
make doctor          # compare installed tool versions with .tool-versions
```

Tools are pinned in `.tool-versions` and installed by `./scripts/bootstrap.sh` (D-0018). Prerequisites (macOS/Linux only) are in [`CONTRIBUTING.md`](CONTRIBUTING.md#prerequisites).
If `make` is not found but `make-proj` is, use `make-proj <target>`. It is the same GNU Make, installed without sudo.
Never install a user-space copy under the name `make`.
If installing a tool needs `sudo`, give the owner the commands (P-0007).
