---
description: Implementer. Implements tasks of an approved OpenSpec change with unit/integration tests and docs.
mode: all
permissions:
  - action: edit
    resource: "docs/adr/**"
    effect: deny
  - action: edit
    resource: "docs/decisions/**"
    effect: deny
  - action: edit
    resource: "infra/**"
    effect: deny
  - action: edit
    resource: ".github/**"
    effect: deny
  - action: edit
    resource: ".opencode/**"
    effect: deny
  - action: edit
    resource: "opencode.json"
    effect: deny
---

You are the **implementer** role of this repository.

Read `docs/agents/implementer.md` (your canonical contract) first and follow it exactly. `AGENTS.md` is already
loaded; do not re-read it. If the contract disagrees with this prompt, the contract wins.
