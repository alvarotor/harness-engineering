# 0019 Owner approval before agent bootstrap

- Status: superseded
- Date: 2026-11-05
- Source: user
- Supersedes: 0018

## Context
`make doctor` identifies missing or mismatched developer tools and points to
`scripts/bootstrap.sh`. The bootstrap can install tools and may report that
Docker must be started or that privileged steps are needed. Automatically
running it can make machine-level changes before the owner has reviewed what it
will do.

## Decision
- Keep `scripts/bootstrap.sh` as the documented setup and repair command for
  developer machines.
- Preserve ADR-0018's technical setup behavior: the script ensures `make` (using
  the OS package manager only in an interactive terminal, or `make-proj` without
  sudo), installs checksum-verified pinned mise, installs the versions in
  `.tool-versions`, checks Docker + Compose without installing them, and runs
  `make doctor` to report remaining issues.
- When an agent's initial `make doctor` (or `make-proj doctor`) check fails, the
  agent explains that it proposes to run `./scripts/bootstrap.sh`, summarizes
  the actions the script may take based on its output or the script's behavior,
  and stops before running it.
- The agent runs bootstrap only after the owner explicitly approves that run.
  Approval to run bootstrap does not authorize any separate action that the
  script reports as requiring owner action, such as starting Docker or using
  `sudo`.
- A direct owner request to run `./scripts/bootstrap.sh` is explicit approval
  for that run.
- This approval gate changes agent behavior only. It does not change the
  bootstrap script, its safety properties, or the requirement to use pinned
  tools.

## Alternatives
- Automatically run bootstrap after a failed doctor check, as specified by
  ADR-0018. This is convenient but can start installations before the owner
  knows what the script will do.
- Stop on a failed doctor check without proposing the repair command. This
  avoids unapproved machine changes but leaves the owner to discover the
  documented repair path.

## Consequences
- Agents may pause at session setup until the owner approves a bootstrap run.
- Owners see the proposed setup action before it is run.
- The bootstrap script remains the single documented setup and repair path.

## Revisit trigger
Revisit if bootstrap becomes read-only, or if the project adds a reliable
preview mode that shows all changes before execution.
