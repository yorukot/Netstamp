-- +goose Up
CREATE TYPE public_status_theme AS ENUM ('light', 'dark', 'auto');
CREATE TYPE public_status_element_display_mode AS ENUM ('status', 'history', 'latency', 'map');

ALTER TABLE public_status_pages
    ADD COLUMN footer_text text,
    ADD COLUMN banner_image_url text,
    ADD COLUMN theme public_status_theme NOT NULL DEFAULT 'auto',
    ADD COLUMN show_targets boolean NOT NULL DEFAULT false,
    ADD COLUMN show_probe_names boolean NOT NULL DEFAULT false,
    ADD COLUMN show_probe_locations boolean NOT NULL DEFAULT false,
    ADD COLUMN show_incident_history boolean NOT NULL DEFAULT true,
    ADD COLUMN show_generated_at boolean NOT NULL DEFAULT true,
    ADD COLUMN custom_css text,
    ADD CONSTRAINT public_status_pages_footer_text_valid CHECK (footer_text IS NULL OR (length(btrim(footer_text)) > 0 AND length(footer_text) <= 2048)),
    ADD CONSTRAINT public_status_pages_banner_image_url_valid CHECK (banner_image_url IS NULL OR (length(btrim(banner_image_url)) > 0 AND length(banner_image_url) <= 2048)),
    ADD CONSTRAINT public_status_pages_custom_css_valid CHECK (custom_css IS NULL OR (length(btrim(custom_css)) > 0 AND length(custom_css) <= 65536));

ALTER TABLE public_status_page_elements
    ADD COLUMN display_mode public_status_element_display_mode NOT NULL DEFAULT 'status';

-- +goose Down
ALTER TABLE public_status_page_elements
    DROP COLUMN IF EXISTS display_mode;

ALTER TABLE public_status_pages
    DROP CONSTRAINT IF EXISTS public_status_pages_custom_css_valid,
    DROP CONSTRAINT IF EXISTS public_status_pages_banner_image_url_valid,
    DROP CONSTRAINT IF EXISTS public_status_pages_footer_text_valid,
    DROP COLUMN IF EXISTS custom_css,
    DROP COLUMN IF EXISTS show_generated_at,
    DROP COLUMN IF EXISTS show_incident_history,
    DROP COLUMN IF EXISTS show_probe_locations,
    DROP COLUMN IF EXISTS show_probe_names,
    DROP COLUMN IF EXISTS show_targets,
    DROP COLUMN IF EXISTS theme,
    DROP COLUMN IF EXISTS banner_image_url,
    DROP COLUMN IF EXISTS footer_text;

DROP TYPE IF EXISTS public_status_element_display_mode;
DROP TYPE IF EXISTS public_status_theme;
