# Proposal

## Why

The harness costs too many tokens and owner approvals per change (audit, 2026-11-11): every request runs
through up to 9 roles and 3 owner gates, and every session reads the full registry (including superseded
entries and the changelog), `ARCHITECTURE.md` and a 140-line `AGENTS.md` that every adapter re-reads.

## What Changes

- **Lanes** (new P-0015): S (no behavior change) → implementer + `make check`; M (feature in an existing
  service) → light change, one owner approval, implementer, reviewer (also checks docs); L (risky work) →
  the full flow. Archiving no longer needs owner approval.
- **Lean context:** `AGENTS.md` rewritten as a short index; session start reads only the current registry
  and the role; `ARCHITECTURE.md`, ADRs and standards only for the sections the task touches. Adapters stop
  re-reading `AGENTS.md`.
- **Registry split:** superseded entries, resolved questions and the changelog move to
  `docs/decisions/history.yaml`; the harness merges both and rejects current entries in history (R-REG-006).
- **Owner preferences recorded:** P-0015 lanes, P-0016 repository is a reusable product base (applied in a
  follow-up change), P-0017 per-role models are personal config (Haiku 5.5 high for all roles).

## Capabilities

### Modified Capabilities

- `agent-roles`: lanes, lane-M documentation check by the reviewer, archive without owner approval.
- `project-memory`: current registry and history file are separate.

## Impact

- Decisions/preferences: new P-0015, P-0016, P-0017; P-0010 (documentation check moves to the reviewer
  in lane M); ADR-0001 memory layout (registry now split; no decision text changes).
- Files: `AGENTS.md`, `CONTRIBUTING.md`, `openspec/config.yaml`, `docs/agents/*`, `.opencode/agents/*`,
  `docs/decisions/{registry,history}.yaml`, `docs/standards/{definition-of-done,validation,agent-tooling}.md`,
  `tools/harness/internal/check`.
- No application services, infrastructure or product features change.
