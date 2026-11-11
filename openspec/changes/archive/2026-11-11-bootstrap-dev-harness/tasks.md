# Tasks

## 1. Accept decisions

- [x] 1.1 Set ADR-0001…0017, the ADR index and registry decisions to `accepted`; move not-yet-existing `enforced_by` paths to `planned_enforced_by`. Verify: `make check` rule R-REG-* passes.

## 2. Project memory and instructions

- [x] 2.1 Write root `AGENTS.md` (precedence, memory protocol, conflict reporting, workflow, roles, commands). Verify: links resolve (R-DOC-001).
- [x] 2.2 Write scoped `AGENTS.md` for `apps/`, `services/`, `contracts/`, `infra/`. Verify: R-SCOPE-001.
- [x] 2.3 Write `CONTRIBUTING.md` with the change workflow and the decision-change process. Verify: R-STRUCT-001.

## 3. Standards and catalog

- [x] 3.1 Write `docs/standards/{engineering,security,testing,definition-of-done,validation}.md`. Verify: R-STRUCT-001, R-DOC-002.
- [x] 3.2 Write `docs/catalog/services.yaml` (4 planned deployables) and its README. Verify: R-CAT-*.

## 4. Agent roles

- [x] 4.1 Write `docs/agents/README.md` and 9 role definitions with all required sections. Verify: R-AGENT-001.
- [x] 4.2 Write OpenCode adapters in `.opencode/agents/`. Verify: R-AGENT-002.

## 5. Validation tooling

- [x] 5.1 Implement `tools/harness` (rules, reporter, CLI) with a unit test per rule. Verify: `make harness-test` with coverage ≥ 80 %.
- [x] 5.2 Add `Makefile` (`check`, `harness-test`, `spec-validate`, `doctor`, `help`), `.tool-versions`, `.gitignore`, `.editorconfig`. Verify: `make doctor` and `make check` exit 0.
- [x] 5.3 Lint the harness with golangci-lint. Verify: `make harness-lint` exits 0.

## 6. Integration check

- [x] 6.1 Run `make check` (harness + tests + lint + openspec strict validation) and record the result in this file.

## Workflow follow-up

- Archive this change with `openspec archive bootstrap-dev-harness` after 6.1 passes.

## Verification record

2026-10-15, `make check` exit 0:
- harness: 17 rules checked, 0 violations (re-run after adding orchestrator, security-reviewer, documentation: 9 roles, still 0 violations)
- harness-test: coverage 97.8 % (min 80 %)
- harness-lint: golangci-lint 0 issues
- spec-validate: 1 passed, 0 failed
- Negative check: a mutated copy of the repo (ADR status drift, orphan service folder, undeclared language, broken link) produced 5 violations with fixes, exit 1.

## 7. Machine bootstrap (added after owner request, D-0018)

- [x] 7.1 Write `scripts/bootstrap.sh` (POSIX sh: make / make-proj, pinned + checksum-verified mise, `mise install`, Docker check, `make doctor`). Verify: 10 tests in `tools/harness/internal/bootstrap` pass under dash; a mutation disabling the checksum check makes a test fail.
- [x] 7.2 `Makefile`: mise shims on PATH; `doctor` checks Docker and points to the bootstrap on failure. Verify: `make-proj doctor`.
- [x] 7.3 Harness R-STRUCT-001 requires the script and its executable bit. Verify: unit test "bootstrap not executable".
- [x] 7.4 Record ADR-0018 + D-0018; update `AGENTS.md` §1 step 0, `CONTRIBUTING.md` prerequisites, `validation.md`. Verify: `make check`.
- [x] 7.5 Real run on this machine. Verify: first run installed mise 2026.10.11 + 6 tools, exit 0; re-run in ~4 s with nothing installed, exit 0.
