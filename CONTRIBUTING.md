# Contributing

This repository is built by humans and AI agents following the same rules. Start with
[`AGENTS.md`](AGENTS.md): precedence, memory protocol and roles.

## Prerequisites

Supported OSes: **macOS and Linux** (P-0009). On a new machine you only need `git` and `curl`:

```sh
git clone <repo> && cd <repo>
./scripts/bootstrap.sh      # safe to re-run; prints exactly what is still missing
make check                  # or `make-proj check` if bootstrap installed the no-sudo copy
```

What [`scripts/bootstrap.sh`](scripts/bootstrap.sh) does (D-0019):

| Step | How |
|---|---|
| `make` | uses the OS make. If missing: `xcode-select --install` (macOS); `sudo apt-get/dnf install -y make` (Linux, interactive terminal only); without sudo, a user-space copy named **`make-proj`** (never `make`, so it can't be mistaken for the OS one) |
| mise | pinned version, SHA-256 verified, installed into `~/.local/bin`, no sudo |
| Go, Node, OpenTofu, buf, golangci-lint, OpenSpec | `mise install`: the exact versions in [`.tool-versions`](.tool-versions) |
| Docker + Compose v2 | **checked only**, never installed (needs admin rights): Docker Desktop on macOS, Docker Engine on Linux |
| verify | `make doctor` |

For a human-maintained developer machine, whenever `make doctor` fails (e.g. after someone bumps
`.tool-versions`), run `./scripts/bootstrap.sh` again. Agents must follow the approval gate in
[`AGENTS.md`](AGENTS.md#1-start-of-every-session): explain the proposed bootstrap actions and wait for owner
approval before running it.
The `Makefile` uses mise's pinned tools even if mise is not activated in your shell. To also get them in
your terminal, add `eval "$(mise activate zsh)"` (or `bash`) to your shell rc.

Once the OS `make` is installed, you can delete the fallback copy: `rm ~/.local/bin/make-proj`.

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
