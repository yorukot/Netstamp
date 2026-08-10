# Database Migrations

Netstamp v0.1.0 starts with the single fresh-install Goose baseline `00001_v0_1_0.sql`. It creates the complete PostgreSQL and TimescaleDB schema required by the release.

The baseline replaces all pre-release migration history. Existing development databases and PostgreSQL volumes that ran the old migrations cannot be upgraded in place: back them up if needed, then recreate them before starting v0.1.0. Do not point the squashed migration directory at an old `goose_db_version` history.

The v0.1.0 baseline intentionally has no destructive rollback. Running `migrate down` against version 1 returns an irreversible-baseline error and leaves the schema and Goose version intact. Restore a backup or recreate the database when a reset is required.

After v0.1.0, migration history is append-only. Never edit, delete, rename, or renumber the baseline or another released migration. Add schema changes as new timestamped Goose migrations with a version greater than the retired pre-release maximum `202607300001`, for example:

```text
202608100001_create_example_table.sql
```

Keep schema changes forward-compatible when they are deployed with application changes.
