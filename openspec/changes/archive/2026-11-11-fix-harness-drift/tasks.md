## 1. Agent permissions

- [x] 1.1 Add `opencode.json` with `default_agent: orchestrator`. Verify with `opencode debug agents` / config output.
- [x] 1.2 Orchestrator, reviewer and security-reviewer: `shell` asks by default; read-only commands (`make check|harness*|spec-validate|*test*|*lint*`, `git status|diff|log|show`, `openspec list|show|status|validate`, scanners for security-reviewer) are allowed.
- [x] 1.3 Deny `edit` on `.opencode/**` and `opencode.json` for implementer and documentation; architect gets `ask` on them plus `allow` on `docs/agents/**`. Documentation loses edit on `docs/decisions/registry.yaml`. Update `docs/agents/architect.md` and `docs/agents/documentation.md` to match.
- [x] 1.4 Fix the reviewer adapter description (code and design only).

## 2. Single tool list

- [x] 2.1 `make doctor` runs `scripts/bootstrap.sh`; remove the duplicated tool loop. Update `tools/harness/internal/bootstrap` doctor tests and `CONTRIBUTING.md`. Verify with `make harness-test`.

## 3. Doc drift

- [x] 3.1 `CONTRIBUTING.md`: reviewer scope, approval-gate link (`#9-commands`).
- [x] 3.2 `docs/standards/agent-tooling.md`: P-0011 → P-0013; codebase graph is for structure queries, not before every read.
- [x] 3.3 `docs/agents/documentation.md`: P-0008 → P-0010.
- [x] 3.4 `openspec/specs/repo-validation/spec.md`: replace the TBD purpose.

## 4. Check

- [x] 4.1 Run `make check` and record the result.
  Result: make check passed (17 rules, coverage 97.9%, 5 specs valid).
