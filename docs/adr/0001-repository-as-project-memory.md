# 0001 Repository as project memory, precedence and change control

- Status: accepted
- Date: 2026-10-15
- Source: user

## Context
Decisions made in chat are lost between sessions and agents. Humans and AI agents need one
durable, version-controlled source of truth, and a clear rule for which instruction wins.

## Decision
1. Every accepted decision, preference, constraint or convention is persisted in
   `docs/decisions/registry.yaml`. Architectural ones also get an ADR. Each registry entry lists
   the files that enforce it (standards, templates, harness rules, agent instructions).
2. **Instruction precedence** (highest first):
   1. The explicit current request from the human owner.
   2. Accepted OpenSpec specs (`openspec/specs/`) and approved changes in progress.
   3. Accepted ADRs and registry decisions.
   4. Engineering standards (`docs/standards/`).
   5. Scoped `AGENTS.md` (service or folder).
   6. Root `AGENTS.md` and repository defaults.
3. If a request contradicts a higher-numbered layer, it applies **only after** the conflict is
   reported and the owner confirms. Agents never silently overwrite a recorded decision.
4. **Changing a decision:** open an OpenSpec change that includes a superseding ADR and the
   registry update, then apply it to templates, standards, validation, agent instructions and
   every affected service in the same change. The harness checks the registry against ADRs.

## Consequences
- Some ceremony for every durable decision. In return, any agent session can rebuild context from the repo alone.
- The registry is machine-readable, so validation rules can reference decision IDs.
