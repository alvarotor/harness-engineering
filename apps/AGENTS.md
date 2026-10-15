# apps/AGENTS.md

Frontends. Adds to the root [`AGENTS.md`](../AGENTS.md).

- **Stack** (D-0006): Preact + Vite + TypeScript (strict), SPA. No SSR without a superseding ADR.
- **No data ownership.** Apps call backends only through REST (generated client from `contracts/openapi`)
  and `services/auth` for login. Never embed secrets. Everything in the bundle is public.
- **Runtime config:** `/config.js` rendered from env at container start. Never bake environment URLs into the build.
- **Auth:** email magic link via the Better Auth client. JWT is kept in memory only (no localStorage) and sent as
  Bearer to APIs. Staff-only UI lives in `apps/admin`, never in `apps/web`.
- **Container:** static files served by non-root nginx on 8080; `/livez` and `/readyz` on 8081.
- **Shared code** goes in `packages/ts/*` (UI kit, API client). Apps never import each other.
- **Tests** (D-0017):
  - Vitest + Testing Library for every module, coverage ≥ 80 %.
  - Playwright e2e in `/e2e` for every user flow when possible.
- **Accessibility:** semantic HTML, labelled controls, keyboard navigable.
