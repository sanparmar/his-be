---
applyTo: "**/migrations/**"
---

# Database Migration Conventions

## File Naming
```
<sequence>_<description>.up.sql
<sequence>_<description>.down.sql
```
Example: `000001_create_patients_table.up.sql`

Sequence numbers are 6 digits, zero-padded.

## Up Migration Rules
- Always create tables with `IF NOT EXISTS`
- Always add `tenant_id uuid NOT NULL` as the second column (after `id`)
- Always add audit columns at the end:

```sql
created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
created_by  UUID        NOT NULL,
updated_by  UUID        NOT NULL,
deleted_at  TIMESTAMPTZ
```

- Add a composite unique index on `(tenant_id, <natural_key>)` where relevant
- Add a partial index for soft deletes: `WHERE deleted_at IS NULL`
- Never reference another service's schema/database

## Down Migration Rules
- `DROP TABLE IF EXISTS <table_name>;` — always use `IF EXISTS`
- Drop indexes before tables
- Mirror every `up` action exactly in reverse

## Standard Table Template
```sql
CREATE TABLE IF NOT EXISTS <table_name> (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   UUID        NOT NULL,
    -- domain columns --
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by  UUID        NOT NULL,
    updated_by  UUID        NOT NULL,
    deleted_at  TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_<table_name>_tenant ON <table_name>(tenant_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_<table_name>_created ON <table_name>(created_at DESC);
```

## Idempotency
All DDL must be safe to re-run: use `IF NOT EXISTS`, `IF EXISTS`, `CREATE INDEX CONCURRENTLY IF NOT EXISTS`.

## PHI Columns
Mark PHI columns with a SQL comment:
```sql
name   BYTEA NOT NULL, -- PHI: encrypted at rest using AES-256-GCM
```
Never store raw PHI as `TEXT` or `VARCHAR`.

## No Cross-Schema References
Never use `FOREIGN KEY` constraints pointing to another service's tables.
Model relationships via UUIDs only.
