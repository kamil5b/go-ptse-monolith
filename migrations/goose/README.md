# Postgres migrations

This project contains SQL migrations for Postgres under `migrations/postgres` (plain .sql files) and a set of goose-style migrations under `migrations/goose`.

Recommended: use `goose` (Pressly) to manage migrations. Steps:

1. Install `goose`:

```bash
go install github.com/pressly/goose/v3/cmd/goose@latest
```

2. Run migrations (Makefile helper):

```bash
DATABASE_URL="postgres://postgres:pass@localhost:5432/app?sslmode=disable" make migrate-up
```

3. Rollback one migration:

```bash
DATABASE_URL="..." make migrate-down
```
