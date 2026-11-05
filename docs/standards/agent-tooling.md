# Agent tooling (optional, personal)

Tools that help an AI agent or a developer work faster. **None of them are needed to build, test or deploy**
this repository. `make check`, `make doctor` and `scripts/bootstrap.sh` never depend on them (P-0011).

Why they are not in the bootstrap: the repository stays tool-neutral (canonical role docs in
[`docs/agents/`](../agents/README.md), thin adapters per tool). Each tool runs code on your machine, so it is
a supply-chain change that needs its own review ([security trigger](security.md#security-review-triggers)).

## Tools in use

| Tool | What it is for | Used by | Installed as (this machine) | Version |
|---|---|---|---|---|
| codebase-memory-mcp | code graph: who calls what, architecture overview | all roles, before reading files | binary `~/.local/bin/codebase-memory-mcp` | 0.11.0 |
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
- **Project tools are different.** Playwright as the e2e test library, Go, Node, OpenTofu, buf,
  golangci-lint and OpenSpec are pinned in `.tool-versions` / `package.json` and installed by
  `scripts/bootstrap.sh` (D-0019). Agents wait for owner approval before running it after a failed doctor check.

## Adding a tool here
1. Check its source and license.
2. Pin the exact version.
3. Add a row to the table above and say which role uses it.
4. Owner approval if it runs with network access or changes how agents read or write files.
