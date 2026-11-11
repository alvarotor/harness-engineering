## 1. Decision record and instructions

- [x] 1.1 Architect records the owner's durable preference in `docs/decisions/registry.yaml` with
  source, rationale, changelog entry, and updated enforcement paths. Verify registry validation
  through `make check`.
- [x] 1.2 Update `AGENTS.md`, `docs/agents/orchestrator.md`, and affected tooling guidance to say
  `make doctor` runs only on explicit owner request. Verify by searching for remaining automatic
  session-start instructions.

## 2. Registry validator

- [x] 2.1 Extend the harness registry preference schema to accept `rationale`, `enforced_by`, and
  `planned_enforced_by`, without changing existing decision or preference semantics.
- [x] 2.2 Extend R-REG-004 validation to check `enforced_by` paths for preferences as well as
  decisions. Add unit tests for valid preference metadata and missing preference enforcement paths.
- [x] 2.3 Confirm or clarify `docs/standards/validation.md` so R-REG-004 explicitly covers both
  decisions and preferences; document the validator change in the harness guidance if needed.

## 3. Specification and archive

- [x] 3.1 Run `make check` and record its actual result before archive.
- [x] 3.2 After owner approval, merge the `agent-roles` and `project-memory` deltas into the
  accepted specs through `openspec archive on-demand-doctor`; verify with
  `openspec validate --all --strict`.
- [x] 3.3 Run `make check` after archive and record its actual result before handoff.
