# Contributing

This repository is built by humans and AI agents following the same rules. Start with
[`AGENTS.md`](AGENTS.md): precedence, memory protocol and roles.

## Prerequisites

Supported developer environments: **macOS, Linux, and Termux** (P-0012). The diagnostic needs `git` and a POSIX shell; Make and the tools below are needed for repository commands:

```sh
git clone <repo> && cd <repo>
./scripts/bootstrap.sh      # prints missing tools and setup guidance; does not install anything
make doctor                 # checks required tool presence without running tool version commands
```

[`scripts/bootstrap.sh`](scripts/bootstrap.sh) is diagnostic-only: it does not install packages, download
tools, or invoke package managers. Install tools yourself, then use `make doctor` to check their
presence and `make check` to run repository validation.

| Step | How |
|---|---|
| Make | macOS: `xcode-select --install`; Linux: use the distribution package manager (for example, `sudo apt-get install -y make` or `sudo dnf install -y make`); Termux: `pkg install make`. No private Make copy is installed. |
| Go | Install through your OS package manager; Termux: `pkg install golang`. Make targets set `GOTOOLCHAIN=local`, so Go never downloads another toolchain implicitly. |
| Node.js | Install through your OS package manager; Termux: `pkg install nodejs-lts`. |
| OpenTofu, buf, golangci-lint, OpenSpec | Install using the tools' documented setup for your OS. These tools are not version-pinned by the repository. |
| Docker + Compose v2 | `make doctor` checks Docker Compose presence only. Docker Desktop is for macOS; Docker Engine is for Linux. Availability in Termux is not guaranteed. |
| verify | `make doctor` checks that `go`, `node`, `tofu`, `buf`, `golangci-lint`, `openspec`, and `docker compose` are present. It does not compare versions. |

For a human-maintained developer machine, whenever `make doctor` fails, run `./scripts/bootstrap.sh`
to see missing tools and setup guidance. Agents must follow the approval gate in
[`AGENTS.md`](AGENTS.md#1-start-of-every-session): offer to install missing user-space tools and wait for
owner approval before running any installer. Running the diagnostic script does not approve an install.
Termux support covers bootstrap and Make targets whose host dependencies are available. The full
`make check` and Compose workflow require the listed tools and Docker Compose; their availability on
Android is not guaranteed.

## Change workflow

1. **Explore** (optional): clarify the idea (`/opsx-explore`).
2. **Propose:** create an OpenSpec change (`/opsx-propose <name>` or `openspec new change <name>`) with
   `proposal.md`, delta `specs/`, `design.md` (when needed) and `tasks.md`.
   Rules for these artifacts are in [`openspec/config.yaml`](openspec/config.yaml).
3. **Approve:** the owner approves the proposal. Required for any new decision, new service, new
   infrastructure or security-sensitive change.
4. **Apply:** implement the tasks in order (`/opsx-apply`). Each task group lands its own tests and docs.
5. **Check:** `make check` must pass locally (CI runs the same targets).
6. **Review:** the reviewer role (code, security and docs pass) and a human for gated changes.
7. **Archive:** `openspec archive <name>` merges the delta specs into `openspec/specs/`.

Skip OpenSpec only for changes that don't alter meaning (typos, formatting, broken links), but still run `make check`.

## Changing an established decision

Accepted decisions are stable on purpose. To change one:
1. Open an OpenSpec change named `supersede-d-NNNN-<topic>`. Its proposal includes a **Decision change** section:
   the old decision, the new one, why, and what is affected.
2. Write a new ADR with `Supersedes: NNNN`. Set the old ADR's status to `superseded`; don't change its decision text.
3. In `docs/decisions/registry.yaml`:
   - set the old entry to `status: superseded` with `superseded_by: D-XXXX`
   - add the new entry
   - add a `changelog` line
4. Update everything in the old entry's `enforced_by` and `planned_enforced_by`: standards, templates,
   harness rules, agent instructions, catalog, and existing services.
5. `make check`, then owner approval, then merge.

**Scoped exceptions** (one service differs from a default) do not supersede a decision. Add them to the
decision's `exceptions` list with a reason, through a change the owner approves.

## Commits and branches

- [Conventional Commits](https://www.conventionalcommits.org/): `feat(core): …`, `fix(auth): …`, `docs: …`, `chore(harness): …`.
- One OpenSpec change per branch: `change/<change-name>`.
- Never commit secrets, `.env` files with real values, or generated artifacts other than `contracts/openapi/`.

## Definition of Done

See [`docs/standards/definition-of-done.md`](docs/standards/definition-of-done.md).
