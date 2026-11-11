# Agent tooling (optional, personal)

Tools that help an AI agent or a developer work faster. **None of them are needed to build, test or deploy**
this repository. `make check`, `make doctor` and `scripts/bootstrap.sh` never depend on them (P-0013).

Why they are not in the bootstrap: the repository stays tool-neutral (canonical role docs in
[`docs/agents/`](../agents/README.md), thin adapters per tool). Each tool runs code on your machine, so it is
a supply-chain change that needs its own review ([security trigger](security.md#security-review-triggers)).

## Tools in use

| Tool | What it is for | Used by | Installed as (this machine) | Version |
|---|---|---|---|---|
| codebase-memory-mcp | code graph: who calls what, architecture overview | all roles, for structure questions (callers, call chains), not before every read | binary `~/.local/bin/codebase-memory-mcp` | 0.11.0 |
| lean-ctx | compressed reads, shell and session memory | all roles | `lean-ctx` binary (npm `lean-ctx-bin`) | 3.10.1 |
| mgrep | semantic search over files | implementer, reviewer | npm `@mixedbread/mgrep` (`mgrep mcp`) | 0.1.13 |
| chrome-devtools MCP | drive and inspect Chrome from an agent | test-engineer | `npx -y chrome-devtools-mcp` | **unpinned** |
| Playwright MCP | drive a browser from an agent | test-engineer | `npx @playwright/mcp@latest` | **unpinned** |
| context7 | current library documentation | implementer, architect | `npx -y @upstash/context7-mcp` | **unpinned** |
| caveman | terse reply style (skill, not a binary) | all roles, personal preference | skills under the OpenCode config | n/a |

Where they are configured: `~/.config/opencode/opencode.jsonc` (`mcp` section). The repository has no
`opencode.json`, so nothing here is shared with the team.

## Rules
- **Pin before sharing.** Any `@latest` or unpinned `npx` entry must get an exact version before it is
  written into a shared config. Current state: chrome-devtools, Playwright MCP and context7 are unpinned.
- **Never required.** A role may use a tool when available, but its handoff must not depend on it. Checks
  must pass without it.
- **Tool output is evidence only** if the command and its result are included in the handoff, as with any check.
- **Project tools are different.** Go, Node, buf, golangci-lint and OpenSpec are checked
  for presence, not pinned to host-tool versions. `scripts/bootstrap.sh` reports missing tools and
  installation guidance but never installs them (D-0021). Android/Termux compatibility is not
  guaranteed for every tool or Docker Compose. Agents offer installation and wait for explicit
  owner approval before running any installer; diagnostic bootstrap does not grant that approval.

## Models per role (P-0017)
Models are personal configuration, never committed: set `agents.<role>.model` (and optionally `steps`) in
`~/.config/opencode/opencode.jsonc`; it merges with the repo adapters in `.opencode/agents/`. Check the result
with `opencode debug agents`. Current owner setup: `github-copilot/claude-haiku-5.5#high` for every role.
No `steps` limit: on the last step OpenCode ends with an assistant prefill, which this provider rejects. Pick the model of the orchestrator session in the
model selector; the config only sets the model for subagents.

## Adding a tool here
1. Check its source and license.
2. Pin the exact version.
3. Add a row to the table above and say which role uses it.
4. Owner approval if it runs with network access or changes how agents read or write files.
