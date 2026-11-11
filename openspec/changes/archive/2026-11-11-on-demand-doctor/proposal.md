# Proposal

## Why

Agents currently run `make doctor` at the start of every session, even when the owner's request
does not involve checking developer tools. The owner wants this check run only when explicitly
requested.

## What Changes

- Remove the automatic session-start `make doctor` check from repository agent instructions.
- Require agents to run `make doctor` only when the owner explicitly asks them to check developer
  tool availability.
- Preserve the existing guidance for reporting missing tools and getting approval before an
  installation when a requested check reports missing tools.
- Extend strict registry validation so durable preferences can carry the rationale and enforcement
  paths required by the repository memory protocol, and validate those paths under R-REG-004.

## Capabilities

### Modified Capabilities

- `agent-roles`: run developer-tool checks only on explicit owner request.
- `project-memory`: validate preference rationale and enforcement metadata, including current
  enforcement paths.

## Impact

- Decisions/preferences: P-0013 (developer tooling setup and checks) and proposed P-0014 (run
  `make doctor` only when explicitly requested).
- Instructions: `AGENTS.md`, `docs/agents/orchestrator.md`, and related tool-check guidance.
- Harness: `tools/harness/internal/check/registry.go` and its unit tests must accept preference
  metadata and validate preference `enforced_by` paths.
- Validation docs: confirm R-REG-004 describes enforcement-path validation for decisions and
  preferences.
- Specifications: `openspec/specs/agent-roles/spec.md` and `openspec/specs/project-memory/spec.md`.
- No application services, infrastructure, dependencies, or product features change.

## Decision change

This request adds a durable agent-workflow preference. The architect should record it in the
decision registry, changelog, and affected enforcement paths during apply. Supporting the required
preference metadata and enforcing its paths is necessary for the registry validator to accept this
record consistently with the repository memory protocol. No accepted ADR is superseded.

## Owner gate

The owner asked for this behavior change. This proposal, including the registry-validator support
described above, must be reviewed and approved before apply, as required by the repository workflow.
