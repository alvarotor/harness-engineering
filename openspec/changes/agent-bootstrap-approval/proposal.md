# Proposal

## Why

The session-start instructions currently tell agents to run `scripts/bootstrap.sh` automatically after `make doctor` fails. That script can install pinned developer tools, and its output may identify steps that need owner action. The owner wants to review the proposed bootstrap run before the agent executes it.

## What Changes

- Change the agent session-start flow to run `make doctor`, explain the bootstrap actions indicated by the failure and the script, then stop for explicit owner approval before running `./scripts/bootstrap.sh`.
- Treat a direct owner request to run `./scripts/bootstrap.sh` as approval for that run. Approval does not authorize separate actions such as starting Docker or using `sudo`.
- Keep the bootstrap script and pinned-tool setup behavior unchanged.
- Supersede ADR-0018 with ADR-0019, which preserves the bootstrap setup decision while changing the agent execution gate. ADR-0019 is accepted after owner approval.

## Capabilities

### Modified Capabilities
- `agent-roles`: define the approval gate for an agent-proposed bootstrap run.

## Impact

- Decisions: D-0018 is superseded by D-0019; D-0019 adds the agent approval gate. P-0007 remains unchanged.
- Enforcement files to audit/update during apply: `AGENTS.md`, `CONTRIBUTING.md`, `docs/standards/agent-tooling.md`, `.opencode/agents/*`, `docs/agents/*`, `Makefile`, `scripts/bootstrap.sh`, and `tools/harness` references to D-0018 or automatic agent execution. Preserve technical bootstrap behavior; update only references that encode the old agent flow.
- ADR and registry records: `docs/adr/0018-dev-bootstrap-mise.md`, `docs/adr/0019-agent-approval-before-bootstrap.md`, `docs/adr/README.md`, and `docs/decisions/registry.yaml`.
- No product services, runtime behavior, infrastructure, or catalog entries are affected.

## Decision change

ADR-0019 supersedes ADR-0018. The owner approved this OpenSpec change; the decision records and agent instructions have been updated accordingly.
