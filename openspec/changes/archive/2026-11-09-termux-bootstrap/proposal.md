# Proposal

## Why

The developer bootstrap currently supports macOS and Linux and falls back to a private `make-proj`
binary when GNU Make is missing and sudo is unavailable. That fallback is no longer wanted, and a
Termux phone is currently rejected as an unsupported host even though Termux can install GNU Make
through its own package manager.

## What Changes

- Remove all `make-proj` detection, installation, invocation and documentation.
- Detect Termux before treating Android's Linux kernel as a generic Linux host. Install GNU Make
  through Termux's package manager when needed, and report the exact package command if it cannot be
  run.
- When mise is missing in Termux, do not download the generic Linux mise binary; report
  `pkg install mise` for the owner to run.
- Keep OS-native Make installation on macOS and Linux, reporting the appropriate install command
  when automatic installation is unavailable.
- Update bootstrap tests and repository instructions to cover Termux and the removal of the
  `make-proj` fallback.
- Extend preference registry validation to support and test `status` and `superseded_by` metadata
  for superseded preferences.
- Add Termux to the supported developer environments for bootstrap and Make targets whose host
  dependencies are available. This change does not claim that Docker Compose or every pinned tool
  works on Android; those requirements remain visible in `make doctor`.

## Capabilities

### New Capabilities
- `developer-bootstrap`: acquire and use the host's GNU Make on supported developer environments.
- `project-memory`: validate supersession metadata for preferences in the decision registry.

### Modified Capabilities
- `agent-roles`: use `make doctor` directly; the retired `make-proj doctor` path is removed.

## Impact

- Decisions: D-0019 and P-0009 change; new D-0020 and P-0012 capture the replacement policy. ADR-0019's agent approval gate remains in force under its superseding ADR.
- Bootstrap and validation: `scripts/bootstrap.sh`, `tools/harness/internal/bootstrap`, preference registry validation in `tools/harness`, `Makefile`, and `docs/standards/validation.md`.
- Developer documentation: `AGENTS.md`, `CONTRIBUTING.md`, `docs/agents/orchestrator.md`, and `docs/standards/agent-tooling.md`.
- Decision records: `docs/decisions/registry.yaml`, `docs/adr/README.md`, and new ADR-0020. ADR-0018 and ADR-0019 decision text will not be edited; ADR-0019 will be marked superseded.
- OpenSpec: the existing `agent-bootstrap-approval` change has a completed implementation but is still awaiting its separately required archive approval. Its `agent-roles` spec must be archived before this change can modify that accepted capability.
- No service, product feature, infrastructure, or catalog changes.

## Decision change

The owner selected the new-ADR path for changing P-0009. This change proposes ADR-0020, which supersedes ADR-0019 to replace its technical host and Make setup policy while retaining the explicit owner-approval gate for agents running bootstrap. The registry will mark D-0019 and P-0009 superseded and add D-0020 and P-0012. The owner must approve this concrete OpenSpec proposal before apply; the decision records will only be accepted during apply.

## Scope note

Termux Make/bootstrap support is the target. Full `make check` support on Android depends on the pinned tools and Docker Compose being available there and is not promised by this change.
