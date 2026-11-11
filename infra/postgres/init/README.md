# Postgres init scripts

Mounted read-only at `/docker-entrypoint-initdb.d` in the `postgres` service. The official image runs them
**once**, on the first start of an empty data volume (`pgdata`). Later changes need a manual migration, not an
edit here.

Rules (ADR-0008, D-0008):

- One database per service, named after the service: `<svc>`.
- Two roles per service: `<svc>_owner` (owns the database and runs migrations) and `<svc>_app` (runtime user, DML
  only). Passwords come from secret files (`/run/secrets/<name>`, see infra/README.md Secrets), never from these scripts.
- One file per service, named `NN-<svc>.sql` (two-digit prefix sets the order). Scripts run in name order.
- The official image runs only `*.sh`, `*.sql` and `*.sql.gz` files. Keep this README here; it is ignored.

The folder is empty until the first service has a database. Example for a service `orders`, to copy when needed:

```sql
-- 10-orders.sql
CREATE ROLE orders_owner LOGIN;
CREATE ROLE orders_app LOGIN;
CREATE DATABASE orders OWNER orders_owner;
REVOKE ALL ON DATABASE orders FROM PUBLIC;
GRANT CONNECT ON DATABASE orders TO orders_app;
```

Passwords are set after creation, for example with `ALTER ROLE ... PASSWORD` from a value read from a secret file. Never write a
password into a script. Scripts run against the local stack and against dev and prod, so keep them idempotent
where possible. Document any change in the owning service's `README.md`.
