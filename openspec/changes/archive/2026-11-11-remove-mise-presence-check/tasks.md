## 1. Decision and prerequisite gates

- [x] 1.1 Owner approved the decision-change route and this OpenSpec proposal before apply.
- [x] 1.2 Complete and archive `termux-bootstrap`, including its prerequisite `agent-bootstrap-approval` archive; Go tests and `make check` passed, and the accepted specs exist before applying this change's MODIFIED deltas.
- [x] 1.3 Architect wrote ADR-0021 (`Supersedes: 0020`), marked ADR-0020/D-0020 superseded, added D-0021 and P-0013, superseded P-0011, and updated the ADR index, registry, changelog, and enforcement paths. Accepted ADR text remains unchanged.

## 2. Host tool checks and bootstrap

- [x] 2.1 Removed `.tool-versions` and its harness required-file rule; removed the Makefile's mise shim path setup. Preserved application dependency manifests and lockfiles.
- [x] 2.2 Changed `make doctor` to check command presence for Go, Node, OpenTofu, buf, golangci-lint, OpenSpec, and Docker Compose. Removed exact developer-tool version comparisons.
- [x] 2.3 Set `GOTOOLCHAIN=local` for Go commands run by repository Make targets; retained the `go.mod` compatibility floor. Focused tests cover no toolchain downloads.
- [x] 2.4 Made `scripts/bootstrap.sh` diagnostic-only: it reports Make and other missing prerequisites with platform-specific installation guidance and invokes no package managers, installers, or downloads. Preserved macOS/Linux/Termux detection and no `make-proj` behavior.
- [x] 2.5 Updated bootstrap and doctor unit tests for present/missing tools, presence-only behavior, no version manager, no package-manager/install invocation, Go no-download behavior, and OS-specific guidance. Focused tests passed.

## 3. Instructions, standards and specifications

- [x] 3.1 Updated `AGENTS.md`, `CONTRIBUTING.md`, `docs/agents/*`, and `docs/standards/agent-tooling.md` so agents report missing tools, offer installation, and wait for explicit approval before any install. Diagnostic bootstrap is documented as non-installing.
- [x] 3.2 Updated `docs/standards/engineering.md` and `docs/standards/validation.md` for unpinned host tools, presence-only checks, and the new required-file set. Updated harness rule comments.
- [x] 3.3 Updated the accepted `developer-bootstrap` and `agent-roles` specs after task 1.2; preserved Termux Make behavior and removed mise/version-pin requirements.
- [x] 3.4 Searched operational documentation and implementation for stale mise, `.tool-versions`, and version-pin instructions; historical ADRs and changelog records retain prior choices.

## 4. Validation, review and archive

- [x] 4.1 Focused bootstrap and harness unit tests passed using `GOTOOLCHAIN=local`.
- [x] 4.2 `openspec validate remove-mise-presence-check --strict` passed. `make check` passed with the already installed Go 1.26, OpenTofu, buf, and golangci-lint binaries explicitly added to PATH; it reported 0 harness violations, 97.9% coverage, 0 lint issues, and 5/5 strict OpenSpec validations. No tools were installed.
- [x] 4.3 Reviewed all files in D-0020's enforcement paths and the full diff. Reviewer verdict: approve; no correctness or design findings. Security review: not triggered (no identity/access, exposure, secrets, supply-chain dependency, data, container, or security-standard changes). Documentation review: decision records and index, contributor/agent guidance, standards, bootstrap and doctor behavior, and test coverage align with the change; the stale devops `.tool-versions` reference was removed. The implementer and architect documented their handoffs above.
- [x] 4.4 Owner explicitly approved archiving after review on 2026-11-11.
- [x] 4.5 Archive after tasks 1.2, 2.4, 3.3, 4.2, and 4.3 completed.
