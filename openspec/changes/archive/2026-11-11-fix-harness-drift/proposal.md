# Proposal

## Why

An audit of the harness (owner request, 2026-11-11) found drift and gaps in the current files:
- Role write boundaries were enforced for the `edit` tool only. Read-only roles could still write
  through the shell, and the implementer and documentation roles could edit their own agent
  adapters, raising their own permissions.
- OpenCode starts sessions in the built-in `build` agent, so the role system is skipped by default.
- Docs contradict each other (reviewer scope, stale decision IDs, a broken approval-gate link).
- The required tool list exists twice (`Makefile` doctor and `scripts/bootstrap.sh`).

## What Changes

- Add a repository `opencode.json` that selects `orchestrator` as the default agent.
- Read-only roles (orchestrator, reviewer, security-reviewer) ask before any shell command that is
  not on a read-only allowlist.
- No role may edit `.opencode/**` or `opencode.json` without owner approval; only the architect may
  propose those edits (each one asks the owner).
- The documentation role can no longer edit `docs/decisions/registry.yaml`; it reports changelog
  lines to the architect instead.
- `make doctor` runs `scripts/bootstrap.sh`, so the tool list lives in one place.
- Fix stale docs: CONTRIBUTING reviewer scope and approval-gate link, reviewer adapter description,
  `agent-tooling.md` (P-0011 → P-0013), `documentation.md` (P-0008 → P-0010), and the
  `repo-validation` spec purpose left as TBD by the previous archive.

## Capabilities

### Modified Capabilities

- `agent-roles`: tool adapters enforce role write boundaries for every tool, and agents cannot
  change their own permissions.

## Impact

- Decisions/preferences: D-0001 (agent instructions), D-0021 / P-0013 (doctor uses the bootstrap
  diagnostic), P-0008, P-0010. No decision changes.
- Files: `opencode.json`, `.opencode/agents/*.md`, `Makefile`, `tools/harness/internal/bootstrap`,
  `CONTRIBUTING.md`, `docs/agents/documentation.md`, `docs/agents/architect.md`,
  `docs/standards/agent-tooling.md`, `openspec/specs/repo-validation/spec.md`.
- No application services, infrastructure or product features change.
