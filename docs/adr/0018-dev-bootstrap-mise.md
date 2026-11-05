# 0018 Developer machine bootstrap with mise

- Status: superseded
- Date: 2026-10-15
- Source: user
- Superseded by: 0019

## Context
A fresh macOS/Linux machine has none of the pinned tools, and `make` may be missing. Every project
command runs through `make` (D-0002), so the first step cannot be a make target.

## Decision
- `scripts/bootstrap.sh` (POSIX `sh`) is the single first step on a new machine, and the fix whenever `make doctor` fails. It:
  1. ensures `make`: the OS make, sudo install in an interactive terminal, or a no-sudo user-space copy named `make-proj` (P-0009);
  2. installs **mise** (pinned version, SHA-256 verified) into `~/.local/bin`, without sudo;
  3. runs `mise install` for the exact versions in `.tool-versions`;
  4. checks Docker + Compose (never installs it: it needs admin rights);
  5. runs `make doctor`, and prints every remaining problem with its fix. Exit 1 if anything is missing.
- It is safe to re-run: installed things are only checked.
- The `Makefile` puts mise's shims first on `PATH`, so make targets use the pinned versions even if mise is not activated in the shell.
- `make doctor` points to the script on failure. Agents run the script at session start when `make` is missing or `doctor`
  fails, and hand sudo steps to the owner (P-0007).
- The harness requires the script and its executable bit (R-STRUCT-001). The script's tests run in `make harness-test`.

## Alternatives
- asdf: same `.tool-versions` file, but slower and needs one plugin per tool.
- Manual installs plus `make doctor`: works, but each machine drifts and onboarding is slow.
- Committing binaries (e.g. make): platform-specific and bloats git.

## Consequences
- One more tool (mise) in the chain, itself pinned and checksum-verified.
- Bumping a tool means editing `.tool-versions`, then `./scripts/bootstrap.sh`. Bumping mise means updating
  `MISE_VERSION` and its 4 checksums in the script.
