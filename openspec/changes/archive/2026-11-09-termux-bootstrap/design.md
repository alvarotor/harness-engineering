# Design

## Context

P-0009 limits supported developer hosts to macOS and Linux and requires the no-sudo `make-proj`
fallback. ADR-0019 preserves that bootstrap behavior while requiring owner approval before an agent
runs bootstrap. The requested Termux support changes the former policy; it does not remove the
agent approval gate.

## Decisions

1. Detect Termux before generic Linux because Android reports a Linux kernel. Use Termux's own
   environment markers and verify that its `pkg` command is available before using the Termux path.
2. Install GNU Make with `pkg install make` in Termux. If `pkg` cannot be used, report that exact
   command and stop the bootstrap; do not extract or copy a private GNU Make binary.
3. If mise is missing in Termux, do not download the generic Linux release. Report `pkg install mise`
   for the owner to run, then stop before attempting pinned tool installation.
4. On macOS and Linux, retain the existing native installation paths. If Make cannot be installed
   automatically, print the host's exact package/developer-tools command.
5. Remove `make-proj` as an accepted executable name from bootstrap, instructions and agent startup
   flows. Make remains the repository command entry point (D-0002).
6. Add Termux support for bootstrap and targets when their prerequisites exist. Do not infer that
   Docker Compose or every pinned tool is supported on Android. The phone's `make doctor` output
   remains the source of truth for those prerequisites.

7. Extend the registry validator so preferences may declare a lifecycle `status` and, when
   `superseded`, a `superseded_by` preference ID. Validate the target exists and is a preference.
   Existing preferences without lifecycle metadata remain current. Add unit coverage for a valid
   superseded preference and invalid/missing targets, and document the rule in validation standards.

## Decision records

Add proposed ADR-0020 with `Supersedes: 0019`; retain ADR-0019's approval gate in the new decision.
During apply, mark D-0019 superseded by D-0020, add D-0020, mark P-0009 superseded by P-0012, and
add P-0012 for macOS, Linux and Termux with package-manager-provided GNU Make and no private
fallback. Do not edit accepted ADR text.

## Risks

- Android/Bionic compatibility of the pinned tool versions and Docker Compose has not been
  established. The change makes this limit explicit and scopes support to Make/bootstrap and
  commands whose dependencies are present. In Termux, missing mise is an owner setup step using
  `pkg install mise`; the bootstrap must not download the generic Linux binary.
- Termux environment markers and `pkg` behavior must be represented in isolated shell tests so
  generic Linux detection does not accidentally take precedence.

## Owner gate

The existing `agent-bootstrap-approval` change must be archived with its own owner approval before
this change archives an update to the same `agent-roles` capability. This change itself also needs
owner approval before apply.
