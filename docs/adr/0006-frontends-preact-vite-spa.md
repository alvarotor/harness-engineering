# 0006 Frontends: Preact + Vite SPA served as static containers

- Status: accepted
- Date: 2026-10-15
- Source: user

## Decision
- `apps/web` and `apps/admin`: Preact + Vite + TypeScript (strict). pnpm workspaces; shared code
  in `packages/ts/*` (UI kit, generated API client, configs).
- API client: typed fetch generated from the proto-derived OpenAPI (ADR-0007). Auth via the
  Better Auth vanilla client.
- Runtime image: `nginxinc/nginx-unprivileged` (pinned digest), serving the static build.
  `/config.js` is rendered from env at container start, so one image works in every environment.
  `/livez` and `/readyz` are served on 8081.
- No SSR. Search-engine indexing of the customer app is not a requirement (B2C app, not a content site).

## Revisit when
SEO or first-paint performance becomes a product requirement. Then consider an SSR or prerender step.
