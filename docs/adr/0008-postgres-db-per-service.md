# 0008 One PostgreSQL server, database per service, explicit migrations

- Status: accepted
- Date: 2026-10-15
- Source: user

## Decision
- **One Postgres server per environment** hosts one database per stateful service. Major version
  pinned by image digest in Phase 4.
- Roles per database: `<svc>_owner` (DDL, used only by `migrate`) and `<svc>_app` (DML only, runtime).
  `CONNECT` on a database is granted only to that service's roles.
- Each service receives its own connection URLs. Nothing assumes a shared host, so moving one database later is config only.
- No cross-database queries or foreign keys. Other services' entities are referenced by opaque IDs.
- Migrations are versioned SQL in `services/<svc>/migrations/`. They run with `<image> migrate` as a one-off job before rollout,
  are backward compatible with the running version (expand/contract), and are validated in CI against an
  empty database and against a database at the previous version.
- Dev/prod: the Postgres container runs on the environment VM with a persistent volume (provider-neutral).
- **Backups: deferred by owner decision.**

## Consequences
- The shared server is one failure domain, and its resources are shared by every service.
- Without backups, loss of the volume or VM means permanent data loss.

## Revisit when
- Before any real user data reaches prod: add backups (mandatory).
- Contention or availability problems: split onto separate servers or move to a managed service.
