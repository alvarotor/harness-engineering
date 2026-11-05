# Tasks

## 1. Decision records

- [x] 1.1 Accept ADR-0019 and set ADR-0018 to `superseded`, preserving ADR-0018's technical setup decisions in the replacement where still applicable. Verify the ADR index and registry agree.
- [x] 1.2 Update `docs/decisions/registry.yaml`: mark D-0018 `superseded` with `superseded_by: D-0019`, add D-0019 with `source: user`, rationale, `enforced_by` / `planned_enforced_by`, and a changelog entry. Verify all named enforcement paths exist or are explicitly planned.

## 2. Agent startup flow

- [x] 2.1 Update `AGENTS.md` so a failed initial doctor check results in an explanation of the proposed bootstrap actions and a stop for explicit owner approval; preserve the instruction to hand separate Docker or sudo actions to the owner.
- [x] 2.2 Audit and update `CONTRIBUTING.md`, `docs/standards/agent-tooling.md`, `docs/agents/*`, and `.opencode/agents/*` wherever they instruct or imply that agents automatically execute bootstrap. Keep human setup instructions clear.
- [x] 2.3 Audit `Makefile`, `scripts/bootstrap.sh`, and `tools/harness` against D-0018's prior `enforced_by` list. Change them only if they encode the agent's automatic execution; otherwise document that their technical behavior remains unchanged. Verify the audit covers every prior enforcement path.

## 3. Specification and integration

- [ ] 3.1 Archive the updated `agent-roles` spec through this OpenSpec change after separate owner approval. Verify the scenarios cover failure, approval, and separate owner actions.
- [x] 3.2 Run `openspec validate agent-bootstrap-approval --strict`; result: passed. `make check` was not run.
- [x] 3.3 Review the final diff and hand it to the owner for gated approval before archiving the change.

## Approval gate

The owner approved applying the proposal. Archive remains pending separate owner approval under repository policy. `make check` remains unverified.

## Audit record

- `AGENTS.md`, `CONTRIBUTING.md`, `docs/agents/orchestrator.md`, and `docs/standards/agent-tooling.md` updated to require owner approval before agent bootstrap.
- `.opencode/agents/*` contains no automatic bootstrap instruction; its orchestrator adapter points agents to the canonical repository instructions.
- `Makefile`, `scripts/bootstrap.sh`, and `tools/harness` contain no agent auto-execution behavior and remain unchanged.
- `openspec validate agent-bootstrap-approval --strict` passed. `make check` was not run.
