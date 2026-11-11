## 1. Decision and prerequisite gates

- [x] 1.1 Owner approved this OpenSpec proposal before apply.
- [x] 1.2 Remove the unsupported `rationale` field from D-0019 in `docs/decisions/registry.yaml`; its rationale remains recorded in ADR-0019. Accepted ADR text is unchanged.
- [x] 1.3 Archive `agent-bootstrap-approval` and validate its accepted `agent-roles` spec before applying this change's modification.
- [x] 1.4 Architect added ADR-0020, superseded D-0019 and P-0009, added D-0020 and P-0012, and updated the ADR index and registry changelog. Final validation remains in task 4.2.

## 2. Bootstrap, registry validation and tests

- [x] 2.1 Update `scripts/bootstrap.sh` to detect Termux before generic Linux, install missing Make through `pkg`, keep macOS/Linux native paths, remove every `make-proj` branch, and run the doctor via `make`. When mise is missing in Termux, do not download the generic Linux binary; report `pkg install mise` and stop before pinned tool installation.
- [x] 2.2 Update `tools/harness/internal/bootstrap` tests to remove fallback expectations and cover Termux detection, package setup, missing mise guidance and no generic Linux download.
- [x] 2.3 Run the focused bootstrap Go tests; passed.
- [x] 2.4 Extend harness registry preference validation to accept optional `status` and `superseded_by`, validate same-type successor references, and add unit tests for superseded preferences. Update `docs/standards/validation.md` with the rule.
- [x] 2.5 Run the harness Go tests; passed with 97.9% coverage.
- [x] 2.6 Update Makefile comments and bootstrap test coverage documentation.

## 3. Instructions and specifications

- [x] 3.1 Update `AGENTS.md`, `CONTRIBUTING.md`, `docs/agents/orchestrator.md`, and `docs/standards/agent-tooling.md` to use `make` directly, document Termux Make setup and the `pkg install mise` owner step, remove operational `make-proj` references, preserve the D-0020 approval gate, and state Android tool/Docker compatibility limits.
- [x] 3.2 After task 1.3, update the accepted `agent-roles` spec through this change to remove `make-proj doctor` while retaining the owner approval scenarios; then rerun strict validation.
- [x] 3.3 Search for remaining operational `make-proj` references. Only historical decision/OpenSpec text and the deliberate test that confirms the fallback is ignored remain.

## 4. Validation, review and archive

- [x] 4.1 `openspec validate termux-bootstrap --strict` exited 0 and reported the change valid.
- [x] 4.2 `make check` passed: 0 harness violations, 97.9% coverage, lint 0 issues, and all OpenSpec changes valid (with the expected INFO before the `developer-bootstrap` spec is archived).
- [x] 4.3 Reviewed the complete diff and audited all prior D-0019 enforcement paths. Reviewer and documentation handoffs completed; security review approved, with scanners unavailable.
- [x] 4.4 Owner requested archiving this change.
- [ ] 4.5 Archive after tasks 1.3, 2.3, 2.5, 3.2, 4.2, 4.3 and 4.4 are complete.

## Apply audit

Validation evidence:

- `openspec validate termux-bootstrap --strict`: passed.
- Focused bootstrap Go tests: passed.
- Harness Go tests: passed with 97.9% coverage.
- `make check`: passed with 0 harness violations, 97.9% coverage, golangci-lint 0 issues, and all OpenSpec changes valid. OpenSpec emitted the expected INFO that archiving the `developer-bootstrap` delta requires this change to be archived first.
- Reviewer and documentation handoffs completed; security review approved. Security scanners were unavailable.

The owner approved archive by requesting “archive commit and push.”
