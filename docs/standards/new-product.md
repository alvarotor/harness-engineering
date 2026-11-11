# Starting a product (P-0016)

This repository is a reusable base. Product choices are **open questions** in
[`docs/decisions/registry.yaml`](../decisions/registry.yaml) with a `needed_by` trigger.

## Rule for agents
Before starting work that reaches an open question's `needed_by` trigger, the orchestrator **stops and asks the
owner**, offering 2–3 options with their trade-offs (earlier analyses live in
[`history.yaml`](../decisions/history.yaml) and the deprecated or superseded ADRs). The owner's answer is
recorded by the architect as a decision (registry + ADR when architectural); the question moves to
`history.yaml` as a resolved question. Work continues only after that.

## First product session (Q6)
Triggered by the first product feature, service or frontend. Ask, in this order:
1. **Product:** what it does, who uses it, B2B or B2C, single or multi-tenant.
2. **Users and roles:** which roles exist and what each may do.
3. **First domains:** the business domains of the first services (D-0014), with owned data.
4. **Slug and domain:** keep `har` or rename (A1); public domain and email provider (Q3) if known.
5. **Which deferred choices the first slice needs now** (Q7–Q11); decide only those.

Then a lane-L change records the decisions, adds `planned` catalog entries, and adds the product-specific
templates (identity, frontend, messaging adapter) the decisions need.

## Deferred choices and their triggers
| ID | Choice | Ask before |
|---|---|---|
| Q6 | product, users, roles, tenancy | any product feature, service or frontend |
| Q3 | email provider, public domain | first dev deploy that sends email or needs TLS |
| Q4 | launch country (GDPR) | prod |
| Q7 | async messaging | first event published between services |
| Q8 | browser / third-party API style | first public API |
| Q9 | hosting, infrastructure as code, backups | first deploy to a real dev host |
| Q10 | identity and login | first feature with a signed-in user |
| Q11 | frontend framework and rendering mode (SEO needs) | first frontend |
