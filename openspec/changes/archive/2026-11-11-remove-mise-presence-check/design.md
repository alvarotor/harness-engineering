# Design

## Context

ADR-0020 uses `mise` to install pinned host tools and permits bootstrap to install missing tools.
The owner wants developer setup to detect command presence, explain missing prerequisites, and wait
for approval before installation. Termux support from P-0012 remains in place.

## Decisions

1. Replace exact-version checks with command-presence checks for the documented developer tools.
   Keep application dependency lockfiles and container image pins unchanged.
2. Remove `.tool-versions`, the `mise` downloader and checksums, `mise install`, and Makefile shim
   injection. `mise` may exist on a developer's machine for other projects but is not required here.
3. Keep bootstrap as a read-only diagnostic. It reports each missing command and a suitable
   installation instruction for macOS, Linux, or Termux; it does not run package managers, download
   binaries, or execute installation commands.
4. Agent instructions require an explicit approval for each proposed tool installation. Approval to
   run checks or the diagnostic bootstrap is not approval to install. Commands requiring `sudo` are
   given to the owner to run, consistent with P-0007.
5. Preserve Termux detection and package guidance for Make. Do not restore `make-proj`.
6. Set `GOTOOLCHAIN=local` for Go commands run by repository Make targets. Keep `go.mod`'s Go
   compatibility floor, but prevent the Go command from silently downloading a toolchain when the
   installed compiler is missing or too old.

## Decision records

During apply, write ADR-0021 with `Supersedes: 0020`; mark ADR-0020 and D-0020 superseded; add D-0021
and update enforced paths. Mark P-0011 superseded by P-0013, preserving its optional personal AI
tooling guidance while removing host-tool pins. Add a changelog entry. Do not edit accepted ADR-0020
text.

## Dependencies and sequencing

The `termux-bootstrap` change is active and its developer-bootstrap/agent-role specs are not yet
accepted. Finish and archive it first. Apply this change only after its specs exist, then modify those
accepted specs through this change.

## Risks

- System package managers can provide different tool versions. Presence checks confirm a command is
  available but cannot guarantee version-specific behavior. CI remains the compatibility gate.
- Installation guidance must name the correct package or developer-tools action for each supported
  OS without executing it.
- A missing `make` prevents `make doctor`; the standalone diagnostic script must still identify that
  prerequisite and explain how to obtain it.
- A present Go command may still be too old for a module's `go.mod` floor. Repository checks should
  fail with that requirement and leave installation to the owner.
