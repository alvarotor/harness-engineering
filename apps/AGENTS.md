# apps/AGENTS.md

Frontends. Adds to the root [`AGENTS.md`](../AGENTS.md).

- **Stack**: TypeScript (strict). Framework and rendering mode (SPA, server-rendered, static) are not
  decided: before creating the first app, resolve Q8 with the owner ([`new-product.md`](../docs/standards/new-product.md)).
- **No data ownership.** Apps call backends only through their public APIs (style decided by Q5). Never embed
  secrets; everything sent to the browser is public.
- **Runtime config:** rendered from env at container start. Never bake environment URLs into the build.
- **Auth:** decided by Q7. Never keep tokens in `localStorage`.
- **Container:** non-root, app on 8080, `/livez` and `/readyz` on 8081.
- **Shared code** goes in `packages/ts/*`. Apps never import each other.
- **Tests**: unit tests for every module, coverage ≥ 80 %; Playwright e2e in `/e2e` for every user flow when possible.
- **Accessibility:** semantic HTML, labelled controls, keyboard navigable.
