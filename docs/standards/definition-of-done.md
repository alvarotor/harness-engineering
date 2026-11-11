# Definition of Done

A change is done only when **every applicable item** holds and was **verified by running it** (P-0004).

## Always
- [ ] An OpenSpec change exists, was approved, and all its tasks are checked (lanes M and L; lane S needs none, P-0015).
- [ ] `make check` passes locally.
- [ ] No conflict with registry decisions, or the conflict was resolved through a decision change.
- [ ] New or changed preferences and decisions are persisted (registry, ADR, standards, AGENTS.md, templates, harness).
- [ ] Docs required by the [documentation map](../agents/README.md#documentation-map) are updated by you, in the same change, and listed in your handoff.
- [ ] No secrets, credentials or real `.env` values are committed.

## Code
- [ ] Unit tests for all new or changed code; coverage ≥ 80 % for the affected deployable or library.
- [ ] Integration tests for new or changed adapters.
- [ ] Lint and format are clean.
- [ ] Errors are handled; no PII or tokens in logs.

## Contracts and data
- [ ] `buf lint` and `buf breaking` pass. Breaking changes use a new version package.
- [ ] Migrations are expand/contract, tested on an empty DB and on the previous version.
- [ ] Catalog `depends_on`, `contracts` and `personal_data` are updated.

## User-facing
- [ ] Playwright e2e for the flow, or the reason it isn't possible is recorded in `tasks.md`.
- [ ] Accessible: labelled controls, keyboard navigation.

## Deployable / infra
- [ ] Image builds, runs non-root, and its health check passes; Trivy shows no fixable HIGH/CRITICAL.
- [ ] Compose config validates (`make compose-config`); IaC `validate`/`plan` is clean once IaC exists (Q9).
- [ ] Owner approval for gated items (security-sensitive, prod, new infra, destructive migration).

## Review
- [ ] Security-reviewer report, if a [security trigger](security.md#security-review-triggers) matched. Findings resolved or accepted by the owner.
- [ ] Reviewer report (code and design), with findings resolved or accepted by the owner.
- [ ] Documentation check: every changed area has its required docs (lane L: documentation report; lane M: part of the reviewer report).
- [ ] Lane M or L: the OpenSpec change is archived by the planner after review passes.
