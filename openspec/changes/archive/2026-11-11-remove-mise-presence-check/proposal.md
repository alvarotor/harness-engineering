# Proposal

## Why

The repository currently pins developer-tool versions in `.tool-versions` and uses `mise` to
install and select them. The owner wants to remove that version-management layer. Setup should say
which commands are missing and offer installation, while never installing anything automatically.

## What Changes

- Remove `mise` from bootstrap, Makefile shims, contributor instructions, and required-tool checks.
- Remove `.tool-versions` and stop enforcing exact versions for host developer tools. `make doctor`
  checks whether each required command exists on `PATH`.
- Keep the Go language/tool compatibility floor in `go.mod`, and set `GOTOOLCHAIN=local` for Go
  commands run through repository targets so an older or missing local toolchain is reported instead
  of downloaded automatically.
- Make `scripts/bootstrap.sh` diagnostic-only. It reports missing commands and OS-specific install
  guidance; it never invokes package managers, downloads, or installers.
- Update agent instructions so agents list missing prerequisites and offer to install them, then wait
  for explicit owner approval before running any install command. A request to run checks does not
  approve an installation.
- Preserve macOS, Linux, and Termux support, including Termux's `pkg install make` guidance and the
  removal of the private `make-proj` fallback.
- Keep application dependency manifests/lockfiles and container image digest rules in force; this
  change removes pins only for developer tools installed on the host. There is no root `package.json`
  in this repository.

## Capabilities

### Modified Capabilities
- `developer-bootstrap`: check prerequisite presence and report install guidance without installing.
- `agent-roles`: offer installation and wait for approval before any installation.

## Impact

- Decisions/preferences: D-0020 is superseded by D-0021. P-0011 is superseded by P-0013, which
  preserves the optional personal AI tooling guidance but removes host-tool version pinning. P-0007
  is clarified: local installation remains possible only after explicit owner approval; privileged
  commands remain owner-run. P-0012's supported OS scope remains unchanged.
- Host setup and checks: `.tool-versions`, `Makefile`, `scripts/bootstrap.sh`, `tools/harness`,
  and bootstrap/registry tests.
- Documentation: `AGENTS.md`, `CONTRIBUTING.md`, `docs/standards/engineering.md`,
  `docs/standards/validation.md`, `docs/standards/agent-tooling.md`, and `docs/agents/*` as needed.
- Decision records: new ADR-0021 supersedes ADR-0020; update its status, the ADR index, registry, and
  changelog. Do not edit accepted ADR-0020 text.
- No application services, runtime dependencies, catalog entries, infrastructure, or product
  features change.

## Decision change

The owner requests that developer tools be detected by presence only, with no host-tool version pins
or `mise`. This contradicts accepted D-0020 and P-0011. This change proposes ADR-0021, superseding
ADR-0020, and P-0013, superseding P-0011. The owner must approve this concrete OpenSpec change
before apply; the new ADR and registry status changes will be accepted during apply.

## Prerequisite

The active `termux-bootstrap` change must first pass its required tests and `make check`, archive the
`agent-bootstrap-approval` prerequisite, update its accepted specs, and be archived. This change
modifies the resulting `developer-bootstrap` and `agent-roles` capabilities and must not be applied
before that archive is complete.

## Scope note

Presence-only checks apply to developer executables such as Go, Node, OpenTofu, buf, golangci-lint,
OpenSpec, and Docker Compose. Project dependency versions in manifests/lockfiles and image digests
remain governed by their existing policies.
