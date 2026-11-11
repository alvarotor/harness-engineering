# 0020 Developer bootstrap on macOS, Linux and Termux

- Status: superseded
- Date: 2026-11-09
- Source: user
- Supersedes: 0019

## Context

P-0009 limits supported developer hosts to macOS and Linux and requires a private
`make-proj` fallback when GNU Make cannot be installed. The owner wants to use the
repository from Termux on an Android phone and has rejected that fallback. ADR-0019
also requires explicit owner approval before an agent runs bootstrap; that gate remains
necessary because bootstrap can install tools.

## Decision

- Support macOS, Linux and Termux for GNU Make and repository commands whose host
  prerequisites are available.
- Detect Termux before generic Linux. When GNU Make is missing, use Termux's `pkg`
  package manager to install it. If that cannot be run, print the exact `pkg install
  make` command. On macOS and Linux, use their native developer-tools or package-manager
  paths and print the exact command when automatic installation is unavailable.
- Never extract or install a private Make binary as `make-proj`. Repository commands
  run through `make` (D-0002).
- Keep `scripts/bootstrap.sh` as the setup and repair command. It may install
  checksum-verified mise and the pinned tools in `.tool-versions`, and checks Docker
  Compose without installing it. Termux support does not imply that Docker Compose or
  every pinned tool works on Android; `make doctor` reports the available prerequisites.
- When an agent's initial `make doctor` check fails or Make is missing, the agent
  explains that it proposes to run `./scripts/bootstrap.sh`, summarizes the actions
  indicated by the check and script, and waits for the owner's explicit approval before
  running it. A direct owner request to run the script approves that run. Approval to
  run the script does not approve separate actions that require owner action, such as
  using `sudo` or starting Docker.

## Alternatives

- Keep the `make-proj` fallback and macOS/Linux-only support. This would not support the
  owner's requested Termux environment and retains a private Make installation path.
- Add Termux as a supported environment but leave Android Make installation to the
  owner. The package-manager path is straightforward and keeps setup consistent with
  the other supported hosts.

## Consequences

- Termux can obtain Make through its package manager without a private fallback binary.
- Docker Compose and some pinned tools may be unavailable on Android; this is reported
  by prerequisite checks and does not block Make targets that do not need them.
- Agents continue to pause for explicit owner approval before running bootstrap after a
  failed initial tool check.

## Revisit trigger

Revisit Android support if the pinned toolchain or Docker prerequisites become available
or are required for the intended Termux workflows.
