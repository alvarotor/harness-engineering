# Validation rules

Run all of them with `make check`. Each violation prints `RULE  path: problem` followed by `fix: …`.
Rule IDs are stable. `tools/harness` fails (R-DOC-002) if a rule exists in code but not here, or the reverse.

| ID | Checks | Typical fix |
|---|---|---|
| R-STRUCT-001 | Required repository files exist (root docs, registry, catalog, standards, agent index, OpenSpec config, `scripts/bootstrap.sh`); the bootstrap script is executable | Restore the file / `chmod +x scripts/bootstrap.sh`. Never delete harness files without a decision change |
| R-REG-001 | Registry parses; IDs are well-formed (`D-NNNN`, `P-NNNN`, `A<n>`, `Q<n>`) and unique; `status`/`source` values are valid | Fix the YAML entry |
| R-REG-002 | Each decision's ADR file exists and its `- Status:` line equals the registry status | Align ADR and registry status |
| R-REG-003 | Every ADR file is referenced by a registry decision and listed in `docs/adr/README.md` with the same status | Add the registry entry or the index row |
| R-REG-004 | Every `enforced_by` path exists | Create the enforcing file or move the path to `planned_enforced_by` |
| R-REG-005 | `superseded` decisions/preferences name an existing replacement of the same type; other entries don't have `superseded_by` | Follow the decision-change process in CONTRIBUTING.md |
| R-CAT-001 | Catalog schema: unique kebab-case names, `kind` app/service with a matching path, valid `status`, `owner`, `domain` (services), `personal_data` present | Fix the catalog entry |
| R-CAT-002 | Language rule: apps use D-0006's language; services use D-0004's default unless listed in its `exceptions` with the same language | Use the default language, or add a justified exception through an approved change |
| R-CAT-003 | Every folder in `apps/` and `services/` has a catalog entry; `active` entries exist on disk; `planned` entries don't | Add or update the catalog entry (`planned` → `active` when code lands) |
| R-CAT-004 | Active deployables have `Dockerfile`, `AGENTS.md`, `README.md` and `go.mod` (Go) or `package.json` (TypeScript) | Add the missing file (templates generate them) |
| R-CAT-005 | `depends_on` names existing catalog entries | Fix the dependency name or add the entry |
| R-AGENT-001 | Each role in `docs/agents/` has all required sections | Add the missing section |
| R-AGENT-002 | Each role has an OpenCode adapter in `.opencode/agents/` that references its canonical file | Add or fix the adapter |
| R-SCOPE-001 | Scoped `AGENTS.md` exist for `apps/`, `services/`, `contracts/`, `infra/` | Restore the file |
| R-DOC-001 | Relative Markdown links resolve (archived changes, generated tool files and dependencies are skipped) | Fix the link or restore the target |
| R-DOC-002 | Rules in code and in this table match | Document the rule here, or remove the stale row |
| R-DOC-003 | Every catalog deployable path is mentioned in `ARCHITECTURE.md` | Update ARCHITECTURE.md §2 |

## Other checks run by `make check`
- `make harness-test`: harness unit tests (coverage ≥ 80 %) and the `scripts/bootstrap.sh` tests (POSIX `sh` with fake tools: OS detection, missing-tool guidance, no installer calls, Termux Docker Compose reporting).
- `make harness-lint`: golangci-lint on the harness.
- `make spec-validate`: `openspec validate --all --strict`.

## Planned (arrive with the code they check)
Compose rules (published ports, socket mount, port 8081 not routed, no `container_name`), Go layer-import
rule, cross-deployable import rule, test-file presence per module, per-deployable coverage, and
template conformance.
