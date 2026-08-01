-- +goose Up
LOCK TABLE system_setting_revisions IN ACCESS EXCLUSIVE MODE;

-- Preserve only a rollback epoch, not the per-resource concurrency state.
-- A rollback must not make an ETag issued before this migration valid again.
CREATE SEQUENCE system_setting_revisions_rollback_epoch AS bigint;

SELECT setval(
    'system_setting_revisions_rollback_epoch',
    CASE
        WHEN count(*) FILTER (
            WHERE resource IN ('access', 'smtp', 'auth.oidc', 'auth.google', 'auth.github')
        ) = 5 THEN max(revision) + 1
        ELSE 0
    END,
    false
)
FROM system_setting_revisions;

DROP TABLE system_setting_revisions;

-- +goose Down
CREATE TABLE system_setting_revisions (
    resource text PRIMARY KEY,
    revision bigint NOT NULL DEFAULT 1,
    CONSTRAINT system_setting_revisions_resource_not_empty CHECK (length(btrim(resource)) > 0),
    CONSTRAINT system_setting_revisions_revision_positive CHECK (revision > 0)
);

WITH rollback_revision AS MATERIALIZED (
    SELECT nextval('system_setting_revisions_rollback_epoch') AS revision
)
INSERT INTO system_setting_revisions (resource, revision)
SELECT resources.resource, rollback_revision.revision
FROM (
    VALUES ('access'),
           ('smtp'),
           ('auth.oidc'),
           ('auth.google'),
           ('auth.github')
) AS resources(resource)
CROSS JOIN rollback_revision;

DROP SEQUENCE system_setting_revisions_rollback_epoch;
