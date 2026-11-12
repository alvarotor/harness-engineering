# Architecture Decision Records

Format: short MADR. One decision per file: `NNNN-kebab-title.md`, numbered from `0001`. The ADR number equals the
decision number in [`../decisions/registry.yaml`](../decisions/registry.yaml).

The template ships with **no ADRs**: its starting rules live in [`ARCHITECTURE.md`](../../ARCHITECTURE.md),
[`docs/standards/`](../standards/) and the `AGENTS.md` files. A product's first architectural decision is ADR `0001`.

- Edit `proposed` ADRs freely.
- Never edit the decision text of an `accepted` ADR. To change it, write a new ADR with `Supersedes: NNNN`
  and set the old one to `superseded`.
- Every ADR has an entry in the registry; the harness checks they match.

Status values: `proposed` → `accepted` → (`superseded` | `deprecated`). `deprecated` = withdrawn without a replacement.
`Source` is either `user` (stated or approved by the owner) or `default` (an agent proposal that was not contested).

| ADR | Title | Status | Source |
|---|---|---|---|
