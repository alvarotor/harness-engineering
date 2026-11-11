# Contributing

This repository is built by humans and AI agents following the same rules. Start with
[`AGENTS.md`](AGENTS.md): precedence, memory protocol and roles.

## Prerequisites

Supported developer environments: **macOS, Linux, and Termux** (P-0012). The diagnostic needs `git` and a POSIX shell; Make and the tools below are needed for repository commands:

```sh
git clone <repo> && cd <repo>
./scripts/bootstrap.sh      # prints missing tools and setup guidance; use this when make itself is missing
make doctor                 # same diagnostic, through make
```

[`scripts/bootstrap.sh`](scripts/bootstrap.sh) is diagnostic-only: it does not install packages, download
tools, or invoke package managers. Install tools yourself, then use `make doctor` to check their
presence and `make check` to run repository validation.

| Step | How |
|---|---|
| Make | macOS: `xcode-select --install`; Linux: use the distribution package manager (for example, `sudo apt-get install -y make` or `sudo dnf install -y make`); Termux: `pkg install make`. No private Make copy is installed. |
| Go | Install through your OS package manager; Termux: `pkg install golang`. Make targets set `GOTOOLCHAIN=local`, so Go never downloads another toolchain implicitly. |
| Node.js | Install through your OS package manager; Termux: `pkg install nodejs-lts`. |
| buf, golangci-lint, OpenSpec | Install using the tools' documented setup for your OS. These tools are not version-pinned by the repository. |
| Docker + Compose v2 | Docker Desktop is for macOS; Docker Engine is for Linux. Availability in Termux is not guaranteed. |
| verify | `make doctor` checks that `make`, `go`, `node`, `buf`, `golangci-lint`, `openspec`, and `docker compose` are present, with setup guidance for each missing one. It does not compare versions. |

Agents must follow the approval gate in [`AGENTS.md`](AGENTS.md#9-commands): offer to install missing
user-space tools and wait for owner approval before running any installer. Running the diagnostic does
not approve an install.
Termux support covers bootstrap and Make targets whose host dependencies are available. The full
`make check` and Compose workflow require the listed tools and Docker Compose; their availability on
Android is not guaranteed.

## Daily commands

| Task | Command |
|---|---|
| Run every repository validation (before every handoff) | `make check` |
| Start the local stack (waits until healthy; port 80 must be free) | `make up` |
| Stop it (volumes are kept) | `make down` |
| Status and logs | `make ps`; `make logs` (all services) or `make logs SERVICE=<name>` |
| Validate the Compose merges | `make compose-config` |
| New backend service (lane L: needs Q6 answered and a `planned` catalog entry first) | `make new-service name=<kebab-case>` |
| Check the service template still builds and tests | `make template-test` |
| Release (`vX.Y.Z` on a `main` commit; owner approves the production gate) | `git tag vX.Y.Z && git push origin vX.Y.Z` |

GitHub Actions setup (secrets, environments, branch protection) and the deploy and rollback steps are in
[`infra/README.md`](infra/README.md). Run `make check` before every handoff.

## Change workflow

1. **Explore** (optional): clarify the idea (`/opsx-explore`).
2. **Propose:** create an OpenSpec change (`/opsx-propose <name>` or `openspec new change <name>`) with
   `proposal.md`, delta `specs/`, `design.md` (when needed) and `tasks.md`.
   Rules for these artifacts are in [`openspec/config.yaml`](openspec/config.yaml).
3. **Approve:** the owner approves the proposal (lanes M and L).
4. **Apply:** implement the tasks in order (`/opsx-apply`). Each task group lands its own tests and docs.
5. **Check:** `make check` must pass locally.
6. **Review:** the reviewer role (code and design; in lane M also docs). Lane L adds the test-engineer, the
   security-reviewer when a security trigger matches, the documentation role, and the owner for gated items.
7. **Archive:** `openspec archive <name>` merges the delta specs into `openspec/specs/`. No owner approval needed.

Every request runs in a lane (P-0015, [`AGENTS.md` §6](AGENTS.md#6-lanes-p-0015)). Lane S (no observable
behavior change) skips OpenSpec but still runs `make check`.

## Changing an established decision

Accepted decisions are stable on purpose. To change one:
1. Open an OpenSpec change named `supersede-d-NNNN-<topic>`. Its proposal includes a **Decision change** section:
   the old decision, the new one, why, and what is affected.
2. Write a new ADR with `Supersedes: NNNN`. Set the old ADR's status to `superseded`; don't change its decision text.
3. In `docs/decisions/`:
   - set the old entry to `status: superseded` with `superseded_by: D-XXXX` and move it to `history.yaml`
   - add the new entry to `registry.yaml`
   - add a `changelog` line in `history.yaml`
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
