# Proposal

## Why

The project has 17 decisions but no way to make humans and AI agents follow them consistently across
sessions. Without a durable, validated harness, preferences get lost, decisions are silently
overwritten, and drift between docs and code goes unnoticed. This must exist before any service is
built, so that every service is born compliant.

## What Changes

- Accept ADR-0001…0017 and the decision registry (status `proposed` → `accepted`).
- Add root `AGENTS.md` (precedence model, memory protocol, conflict reporting, workflow) and scoped
  `AGENTS.md` for `apps/`, `services/`, `contracts/`, `infra/`.
- Define 9 agent roles (incl. orchestrator, security-reviewer, documentation) in `docs/agents/`, with OpenCode adapters in `.opencode/agents/`.
- Add engineering, security, testing, validation and Definition-of-Done standards in `docs/standards/`.
- Add a service catalog (`docs/catalog/services.yaml`) with ownership, dependencies and personal-data inventory.
- Add `CONTRIBUTING.md` with the change workflow and the controlled process for changing decisions.
- Add `tools/harness`, a Go CLI that validates repository structure, registry/ADR consistency,
  catalog rules, agent definitions and documentation links, with an explanation of how to fix each violation.
- Add `scripts/bootstrap.sh` (fresh-machine setup with mise, D-0018).
- Add `Makefile` entry points (`make check`, `make harness-test`, `make spec-validate`, `make doctor`)
  and pin tool versions in `.tool-versions`.

Out of scope (later changes): service templates (they arrive with the first service of each kind, so they are
proven by real code), Compose/Traefik, CI workflows, OpenTofu, and compose/import-rule checks.

## Capabilities

### New Capabilities
- `project-memory`: how decisions, preferences and conventions are recorded, ranked, changed, and kept consistent.
- `agent-roles`: the set of AI/human roles, their contracts, handoffs and approval gates.
- `repo-validation`: automated repository checks and how violations are reported.

### Modified Capabilities

## Impact

- Decisions: D-0001 (implemented), D-0002 (Makefile, structure), D-0004 (language rule check),
  D-0006 (apps language check), D-0014 (catalog), D-0017 (harness tests ≥ 80 %). All D-* move to `accepted`.
- New Go module `tools/harness` (dependency: a YAML parser).
- No runtime services are affected; none exist yet.
