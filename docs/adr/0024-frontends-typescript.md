# 0024 Frontends in TypeScript; framework and rendering mode deferred

- Status: accepted
- Date: 2026-11-11
- Source: user
- Supersedes: 0006

## Context

ADR-0006 chose Preact + Vite SPAs. A plain SPA suits apps behind a login (e.g. a social network) but is
weak for search-engine traffic (e.g. a shop). The product is not chosen yet (P-0016).

## Decision

- Frontends (`apps/*`) are TypeScript, containerized, non-root, with runtime config from env (one image for
  every environment) and the standard runtime contract (8080 app, 8081 ops).
- **Deferred (Q11):** framework and rendering mode (SPA, server-rendered or static). Decide when the product
  is known; ADR-0006 holds the earlier SPA analysis.

## Consequences

The first frontend change must resolve Q11, then add a frontend template.

## Revisit when

The product is chosen (Q6) or a frontend needs a non-TypeScript stack.
