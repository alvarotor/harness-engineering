# Design

## Context

See proposal.md for motivation. Decisions are in `docs/decisions/registry.yaml` and ADR-0001…0017.
No services exist yet. Installed tools: Go 1.26, Node 24, OpenSpec 1.14.1, OpenTofu 1.13.1,
buf 1.74.0, golangci-lint 2.14.0, Playwright 1.64.1.

## Goals / Non-Goals

**Goals:** durable memory, an explicit precedence model, a small set of roles, machine-checked consistency, and clear fix messages.

**Non-Goals:**
- Semantic conflict detection (it needs judgment; agents report conflicts and the reviewer checks).
- Compose, import-graph and coverage-of-services rules. Those land with the code they check.

## Decisions

1. **Canonical docs + thin tool adapters.** Roles live in `docs/agents/*.md` (tool-neutral). The adapters in
   `.opencode/agents/*.md` only add permissions and point to the canonical file.
   - Alternative: OpenCode-only agent files. Rejected because it locks the harness to one tool.
2. **9 roles** (owner decision P-0008, revising an earlier 6-role proposal):
   - An *orchestrator* makes the order of work explicit and enforced, instead of each agent following it from memory. It has no edit
     rights, so it cannot quietly do the work it should delegate.
   - The *security reviewer* is separate from the reviewer, so security gets an independent, focused review.
     It runs only on triggers, to keep the cost proportional.
   - The *documentation* role verifies, per change, that each role documented its own work (diff vs. docs), and fixes descriptive gaps.
     Authors are always responsible for their docs (owner decision). Archiving belongs to the planner, which owns OpenSpec changes.
   - *Migration/refactoring* is not a role: the architect plans it and the implementer executes it.
3. **The harness is a Go CLI (`tools/harness`).** It's testable, matches the default backend language, and needs no runtime beyond Go.
   Every rule has an ID, a message and a fix. A test ensures each rule ID is documented in
   `docs/standards/validation.md`, so docs and code cannot drift.
   - Alternative: shell scripts with yq. Rejected as hard to test and adding an extra dependency.
4. **`enforced_by` lists only paths that exist.** Future enforcement goes in `planned_enforced_by` (not
   validated), so the registry never claims enforcement that isn't there.
5. **Catalog status lifecycle** `planned → active → deprecated`. Only `active` entries must exist on disk with
   their required files, which lets the catalog describe the approved topology before code exists.
6. **Templates are deferred** to the change that builds the first service of each kind. A template is extracted from
   working code and verified by regenerating it. This avoids placeholder templates that cannot run.
7. **Tool versions** are pinned in `.tool-versions` (asdf/mise compatible). `make doctor` compares them with installed versions.

## Risks / Trade-offs

- [The harness checks structure, not meaning] → reviewer role checklist + precedence rule in `AGENTS.md`.
- [Registry and docs can still drift in prose] → the link, index and catalog-in-ARCHITECTURE checks catch the common cases.
- [Agents may ignore `AGENTS.md`] → `make check` runs in CI (Phase 4) and blocks merges.
