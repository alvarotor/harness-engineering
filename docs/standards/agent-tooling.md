# Agent tooling (optional, personal)

Tools that help an AI agent or a developer work faster. **None of them are needed to build, test or deploy**
this repository. `make check`, `make doctor` and `scripts/bootstrap.sh` never depend on them.

Why they are not in the bootstrap: the repository stays tool-neutral (canonical role docs in
[`docs/agents/`](../agents/README.md), thin adapters per tool). Each tool runs code on the developer's machine, so it is
a supply-chain change that needs its own review ([security trigger](security.md#security-review-triggers)).

## Where tools are configured
Each developer configures optional tools in their **personal** agent configuration (for OpenCode:
`~/.config/opencode/`). The repository ships only the role adapters in `.opencode/agents/`, so no personal tool
is shared with the team.

## Rules
- **Pin before sharing.** Any `@latest` or unpinned `npx` entry needs an exact version before it goes into a shared
  config. Personal configs may stay unpinned.
- **Never required.** A role may use a tool when available, but its handoff must not depend on it. Checks
  must pass without it.
- **Tool output is evidence only** if the command and its result are included in the handoff, as with any check.
- **Project tools are different.** Go, Node, buf, golangci-lint and OpenSpec are checked for presence, not pinned to
  host-tool versions. `scripts/bootstrap.sh` reports missing tools and installation guidance but never installs
  them. Android/Termux compatibility is not guaranteed for every tool or Docker Compose. Agents offer
  installation and wait for explicit owner approval before running any installer; diagnostic bootstrap does not
  grant that approval.

## Models per role
Models are personal configuration, never committed: set `agents.<role>.model` (and optionally `steps`) in the
personal OpenCode configuration; it merges with the repo adapters in `.opencode/agents/`. Check the result with
`opencode debug agents`. Pick the model of the orchestrator session in the model selector; the configuration
only sets the model for subagents.

## Adding a tool to the team's setup
Only tools the repository itself depends on belong in the repository. For a personal tool:
1. Check its source and license.
2. Pin the exact version in the personal configuration.
3. Note which role uses it, in the personal configuration or the handoff.
4. Owner approval if it runs with network access or changes how agents read or write files.
