## 1. Decisions and memory (architect)

- [x] 1.1 Deprecate D-0005, D-0015, D-0016 and P-0001; supersede D-0006 by D-0024, D-0007 by D-0022, D-0010 and D-0013 by D-0023; write ADR-0022, ADR-0023, ADR-0024 and their index rows.
- [x] 1.2 Move superseded, deprecated and resolved entries to `docs/decisions/history.yaml`; remove the D-0004 `services/auth` exception and assumptions A3–A5.
- [x] 1.3 Record Q6, Q3, Q4, Q7, Q8, Q9, Q10, Q11 as open questions with `needed_by` triggers.

## 2. Triggers and standards (architect)

- [x] 2.1 Write `docs/standards/new-product.md` (checklist run with the owner at Q6–Q11 triggers) and link it from `AGENTS.md` and the orchestrator role.
- [x] 2.2 Update `openspec/config.yaml` so the proposal rules require the owner's answer as a decision before a trigger.

## 3. Slimmed docs (architect)

- [x] 3.1 Slim `ARCHITECTURE.md`, `security.md`, `engineering.md`, `testing.md` and scoped `AGENTS.md` files to the decided stack.
- [x] 3.2 Empty `docs/catalog/services.yaml` and verify `docs/catalog/README.md` says no deployables yet.

## 4. Check (planner)

- [x] 4.1 Run `make check` and record the result (registry, catalog and docs rules must pass).
  Evidence: `make check` passed (registry, catalog and docs rules; 7 specs valid).
