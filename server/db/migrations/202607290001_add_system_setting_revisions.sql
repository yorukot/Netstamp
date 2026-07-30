-- +goose Up
CREATE TABLE system_setting_revisions (
    resource text PRIMARY KEY,
    revision bigint NOT NULL DEFAULT 1,
    CONSTRAINT system_setting_revisions_resource_not_empty CHECK (length(btrim(resource)) > 0),
    CONSTRAINT system_setting_revisions_revision_positive CHECK (revision > 0)
);

INSERT INTO system_setting_revisions (resource)
VALUES ('access'),
       ('smtp'),
       ('auth.oidc'),
       ('auth.google'),
       ('auth.github');

DELETE FROM system_settings
WHERE secret = true
  AND octet_length(encrypted_value) = 16
  AND key IN (
      'smtp.password',
      'auth.provider.oidc.client_secret',
      'auth.provider.google.client_secret',
      'auth.provider.github.client_secret'
  );

-- +goose Down
DROP TABLE IF EXISTS system_setting_revisions;
