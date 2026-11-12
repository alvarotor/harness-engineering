# Starting a product

This repository is a **template**. Each product is a **new repository** created from it. This repository
never holds a product's code, decisions or specs. Product choices are **open questions** in
[`docs/decisions/registry.yaml`](../decisions/registry.yaml) with a `needed_by` trigger.

Product repositories do not receive later harness changes automatically. Port them by hand when useful.

## Checklist: create a product repository (owner, manual)
1. **Create the repository from this template** (GitHub "Use this template"). It starts with fresh history.
2. **Rename the project slug**: `make rename slug=<product-slug>` (kebab-case). It replaces the kebab, snake and
   UPPER_SNAKE forms of the current slug (the Compose project `name:`) in every text file: Go module paths, image
   names, Compose project and network, `<SLUG>_SECRETS_DIR`, the deploy folder on the host, proto packages and docs.
   Run it once, before any product code, then review `git diff`. The `tools/harness` folder and `harness` command keep their names.
3. **Remove template-only wording**: the "reusable base" and "template" statements in `AGENTS.md`, `ARCHITECTURE.md`,
   `openspec/config.yaml` and this file are replaced with the product's own status. Keep the generic harness rules.
4. **Run `make check`** and record the result.
5. **Review the accepted risks** in [`security.md`](security.md#known-accepted-risks-template-defaults) with the owner:
   they are template defaults until the owner confirms them for this product.
6. **Answer Q1 with the owner** (product, users and roles, tenancy). Record the answer as a decision in the registry
   (and an ADR when architectural), then move the question to `history.yaml` as resolved. Answer Q2–Q8 only when a
   first slice needs them (their triggers are in the table below).
7. **Write the product's first OpenSpec proposal** (`openspec-propose`), lane L. This is the first change in the repo.
8. **Start with a few coarse services**. Each new service is lane L and needs its catalog entry (`planned`),
   its business domain and a `make new-service name=<kebab-case>`.

## Rule for agents
Before starting work that reaches an open question's `needed_by` trigger, the orchestrator **stops and asks the
owner**, offering 2–3 options with their trade-offs. The owner's answer is recorded by the architect as a decision (registry + ADR when architectural); the
question moves to `history.yaml` as a resolved question. Work continues only after that.

## First product session (Q1)
Triggered by the first product feature, service or frontend. Ask, in this order:
1. **Product:** what it does, who uses it, B2B or B2C, single or multi-tenant.
2. **Users and roles:** which roles exist and what each may do.
3. **First domains:** the business domains of the first services, with owned data.
4. **Slug and domain:** rename the slug (checklist step 2); public domain and email provider (Q2) if known.
5. **Which deferred choices the first slice needs now** (Q3–Q8); decide only those.

Then a lane-L change records the decisions, adds `planned` catalog entries, and adds the product-specific
templates (identity, frontend, messaging adapter) the decisions need.

## Deferred choices and their triggers
| ID | Choice | Ask before |
|---|---|---|
| Q1 | product, users, roles, tenancy | any product feature, service or frontend |
| Q2 | email provider, public domain | first dev deploy that sends email or needs TLS |
| Q3 | launch country (GDPR) | prod |
| Q4 | async messaging | first event published between services |
| Q5 | browser / third-party API style | first public API |
| Q6 | hosting, infrastructure as code, backups | first deploy to a real dev host |
| Q7 | identity and login | first feature with a signed-in user |
| Q8 | frontend framework and rendering mode (SEO needs) | first frontend |
