# 0005 Identity via Better Auth (Node), magic link, JWT/JWKS

- Status: accepted
- Date: 2026-10-15
- Source: user

## Context
Writing auth ourselves is high-risk. The owner chose Better Auth, which is TypeScript-only. The
service therefore runs on Node under the override rule in ADR-0004.

## Decision
- `services/auth`: Node 24 LTS + Hono + Better Auth, layered as ports and adapters. It owns the `auth` database.
- Login: **email magic link only**, for all roles. No passwords, no passkeys for now.
- Plugins:
  - `magicLink`
  - `admin` (roles `customer`, `support`, `admin`; staff are invite-only and cannot self-register)
  - `jwt` (EdDSA, TTL ≤ 15 min, claims `sub role email iss aud`)
- Session cookie: host-only on `auth.DOMAIN`, `HttpOnly; Secure; SameSite=Lax`. Frontends call
  `GET /api/auth/token` and send the JWT as Bearer to Go services.
- Go services verify JWTs locally against `/api/auth/jwks` (cached; refreshed on unknown `kid`).
- Better Auth's built-in rate limiter is enabled with database storage.
- Schema: SQL generated from Better Auth's schema, committed and reviewed, then applied by the image's explicit `migrate` command.
- Email: SMTP adapter. Mailpit locally. The provider is not chosen yet (Q3).

## Consequences
- Account security equals email security, which is a material risk for staff accounts.
- Revocation (ban, role change) takes effect within the JWT TTL.

## Revisit when
- Before staff use prod with real data: add TOTP or passkeys for staff.
- Enterprise SSO is needed, or immediate revocation is required.
