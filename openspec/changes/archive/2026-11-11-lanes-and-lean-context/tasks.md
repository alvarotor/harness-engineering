## 1. Memory

- [x] 1.1 Record P-0015, P-0016, P-0017 and a changelog line. Verify with `make harness`.
- [x] 1.2 Split `docs/decisions/registry.yaml` into current entries and `history.yaml`; harness merges both and adds R-REG-006 with unit tests; document it in `docs/standards/validation.md`. Verify with `make harness-test`.

## 2. Lanes

- [x] 2.1 Lanes in `AGENTS.md`, `docs/agents/README.md`, `orchestrator.md`, `reviewer.md`, `planner.md`, `definition-of-done.md`, `CONTRIBUTING.md`, `openspec/config.yaml`.
- [x] 2.2 Archive needs no owner approval in every doc that listed it.

## 3. Lean context

- [x] 3.1 Rewrite `AGENTS.md` as a short index; session start reads only the current registry, the role and scoped files.
- [x] 3.2 Adapters no longer re-read `AGENTS.md`; the orchestrator passes paths, not file contents.
- [x] 3.3 Document per-role models as personal config in `docs/standards/agent-tooling.md`.

## 4. Check

- [x] 4.1 Run `make check` and record the result.
  Result: make check passed (18 rules, coverage 97.5%, 5 items valid).
