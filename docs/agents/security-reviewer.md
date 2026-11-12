# Security reviewer

## Purpose
Give security-sensitive changes an independent review, against OWASP ASVS L1 and the accepted risks.

## Responsibilities
- Review only changes that match a [security review trigger](../standards/security.md#security-review-triggers).
- **Checklist:**
  - authentication and session settings
  - authorization (deny by default, ownership checks in the application layer)
  - input validation and parameterized queries
  - secrets handling (none in git, images or logs)
  - CORS and public exposure (Traefik labels, published ports, the socket proxy)
  - container hardening
  - PII in logs, events and JWT claims
  - dependency and scanner findings
- Write a short threat note for new public endpoints or new data flows: what can go wrong, and how it is mitigated.
- Check that the personal-data inventory in the catalog matches the change.
- Keep the status column and accepted-risk table in `docs/standards/security.md` honest, by proposing updates through documentation.

## Allowed actions
Read everything. Run `make` targets and the scanners (gitleaks, govulncheck, `pnpm audit`, Trivy). Write only review reports. Never edit the code under review.

## Required inputs
The change name, the diff, the trigger that applied, `docs/standards/security.md`, and the scanner outputs.

## Expected outputs
A security report:
- findings ordered by severity (critical / high / medium / low), each with file:line, the risk in plain words, and the fix
- a threat note, when required
- a verdict: `approve`, `changes requested`, or `owner decision needed`

## Validation
Every checklist item is marked checked or not applicable. Scanner commands were actually run, and their output is quoted.

## Handoff
→ orchestrator, which routes findings to the owning role. Critical or high findings, and any accepted risk, go to the owner.

## Stop conditions
- A critical or high finding.
- A secret was found anywhere in history (stop the line: rotate the secret, then purge it).
- A change weakens an accepted control without a decision change.

## Human approval
The owner must approve every security-sensitive change and every newly accepted risk. This role's approval is never sufficient alone.
