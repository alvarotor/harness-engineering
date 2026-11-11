# 0021 Developer tool presence checks without automatic installation

- Status: accepted
- Date: 2026-11-09
- Source: user
- Supersedes: 0020

## Context

ADR-0020 supports Termux and preserves a bootstrap flow that can install mise and the
host-tool versions selected through `.tool-versions`. The owner wants the repository to
check which tools are available, explain what is missing, and offer installation without
installing anything automatically.

## Decision

- Keep macOS, Linux and Termux as supported developer environments, including the Termux
  Make guidance and the prohibition on a private `make-proj` fallback (P-0012).
- Check required developer executables for presence only. Do not pin host developer-tool
  versions in `.tool-versions`; application dependencies remain governed by their
  manifests and lockfiles, and container images remain pinned by digest.
- Remove mise from the developer workflow. Repository Make targets use the tools available
  on `PATH` and do not select tool shims.
- Keep `scripts/bootstrap.sh` as a diagnostic command. It reports missing commands and
  gives installation guidance for the detected environment. It never invokes package
  managers, downloads, or installers.
- Agents explain which tools are missing, provide the relevant installation guidance, and
  offer to install them. An agent may install a developer tool only after explicit owner
  approval. Running checks or the diagnostic bootstrap does not authorize installation.
  If installation requires privileges, the owner runs the exact command (P-0007).
- Set `GOTOOLCHAIN=local` for Go commands invoked by repository Make targets. An absent or
  too-old local Go toolchain must fail with its compatibility requirement instead of
  downloading another toolchain automatically.
- Keep optional personal AI tools separate from required developer tools. They remain
  optional, documented in `docs/standards/agent-tooling.md`, and pinned before being
  shared (P-0013).

## Alternatives

- Keep mise and exact host-tool versions. This provides consistent local tool versions but
  requires a version manager and downloads during setup.
- Let bootstrap install missing tools after the owner starts it. This still makes a
  diagnostic command perform installation and can change the machine unexpectedly.

## Consequences

- A developer must install required tools through the platform's package manager or other
  documented setup path. Bootstrap only reports the missing tools and guidance.
- Local tool versions can differ. Repository checks and CI verify compatibility, while
  `go.mod`, application lockfiles and image digests retain their existing constraints.
- Agents must wait for explicit owner approval before installing even user-space developer
  tools; privileged installs remain owner-run.

## Revisit trigger

Revisit the presence-only policy if CI reveals incompatibilities that cannot be diagnosed
without local version pins, or if the owner asks to restore managed tool versions.
