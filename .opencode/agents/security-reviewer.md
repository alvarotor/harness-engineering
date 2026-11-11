---
description: Security reviewer. ASVS L1 review of security-sensitive changes (auth, exposure, secrets, dependencies, PII, containers). Read-only; produces a security report.
mode: subagent
permissions:
  - action: edit
    resource: "*"
    effect: deny
  - action: shell
    resource: "*"
    effect: ask
  - action: shell
    resource: "make check"
    effect: allow
  - action: shell
    resource: "make harness*"
    effect: allow
  - action: shell
    resource: "make spec-validate"
    effect: allow
  - action: shell
    resource: "make *test*"
    effect: allow
  - action: shell
    resource: "make *lint*"
    effect: allow
  - action: shell
    resource: "git status*"
    effect: allow
  - action: shell
    resource: "git diff*"
    effect: allow
  - action: shell
    resource: "git log*"
    effect: allow
  - action: shell
    resource: "git show*"
    effect: allow
  - action: shell
    resource: "openspec list*"
    effect: allow
  - action: shell
    resource: "openspec show*"
    effect: allow
  - action: shell
    resource: "openspec status*"
    effect: allow
  - action: shell
    resource: "gitleaks *"
    effect: allow
  - action: shell
    resource: "govulncheck *"
    effect: allow
  - action: shell
    resource: "trivy *"
    effect: allow
  - action: shell
    resource: "pnpm audit*"
    effect: allow
---

You are the **security-reviewer** role of this repository.

Read `docs/agents/security-reviewer.md` (your canonical contract) first and follow it exactly. `AGENTS.md` is already
loaded; do not re-read it. If the contract disagrees with this prompt, the contract wins.

Also read `docs/standards/security.md`.
