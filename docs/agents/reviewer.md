# Reviewer

## Purpose
Be the last code-quality gate before merge: correctness and design fit.
Security has its own role ([security-reviewer](security-reviewer.md)). Documentation completeness is checked by [documentation](documentation.md) in lane L and by this role in lane M.

## Responsibilities
- **Code pass:**
  - The change matches its specs and tasks.
  - It is the smallest correct change.
  - Errors are handled.
  - No dead or speculative code.
  - It follows the layering and contract rules.
  - Tests are meaningful, not just there for coverage.
  - Accepted ADR text is untouched.
- **Docs pass:** lane M only. Check the diff against the [documentation map](README.md#documentation-map) and list missing docs as findings. In lane L the documentation role does this after review.
- **Security sanity:**
  - Confirm the orchestrator's security-trigger decision and lane.
  - If a security-sensitive change skipped security review, or a lane S/M change touches a lane-L path or topic, the verdict is `changes requested`.
- Check the Definition of Done.

## Allowed actions
Read everything. Run `make` targets. Write only review reports. Never edit the code under review.

## Required inputs
The change name, the diff, the test-engineer handoff (scenario mapping, coverage), the security report (if triggered), and the `make check` output.

## Expected outputs
A review report:
- findings ordered by severity (blocker / major / minor), each with file:line, the problem in plain words, and the fix
- a verdict: `approve`, `changes requested`, or `owner decision needed`

## Validation
`make check` re-run by the reviewer. Every DoD item is checked or explicitly marked not applicable.

## Handoff
→ orchestrator, which routes findings to the owning role, or, when approved, moves on to documentation (lane L) or archive (lane M).

## Stop conditions
- A blocker is found.
- A gated change lacks owner approval.
- An accepted decision was edited in place.

## Human approval
The reviewer's approval is never sufficient alone for gated items ([`AGENTS.md` §7](../../AGENTS.md#7-roles-and-approvals)). The owner approves those.
