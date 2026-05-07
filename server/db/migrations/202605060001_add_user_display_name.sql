-- +goose Up
ALTER TABLE users
    ADD COLUMN display_name text,
    ADD CONSTRAINT users_display_name_not_empty CHECK (
        display_name IS NULL OR (
            length(btrim(display_name)) > 0 AND
            length(btrim(display_name)) <= 100
        )
    );

-- +goose Down
ALTER TABLE users
    DROP CONSTRAINT users_display_name_not_empty,
    DROP COLUMN display_name;
