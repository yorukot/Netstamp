-- +goose Up
CREATE TYPE public_status_chart_mode AS ENUM ('inherit', 'off', 'compact');
CREATE TYPE public_status_chart_range AS ENUM ('24h', '7d', '30d');
CREATE TYPE public_status_element_kind AS ENUM ('folder', 'check');

CREATE TABLE public_status_pages (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id uuid NOT NULL REFERENCES projects(id),
    slug citext NOT NULL,
    title text NOT NULL,
    description text,
    enabled boolean NOT NULL DEFAULT true,
    default_chart_mode public_status_chart_mode NOT NULL DEFAULT 'off',
    default_chart_range public_status_chart_range NOT NULL DEFAULT '24h',
    created_by_user_id uuid NOT NULL REFERENCES users(id),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz,
    CONSTRAINT uq_public_status_pages_id_project UNIQUE (id, project_id),
    CONSTRAINT public_status_pages_slug_not_empty CHECK (length(btrim(slug::text)) > 0),
    CONSTRAINT public_status_pages_title_not_empty CHECK (length(btrim(title)) > 0),
    CONSTRAINT public_status_pages_description_not_empty CHECK (description IS NULL OR length(btrim(description)) > 0),
    CONSTRAINT public_status_pages_deleted_at_after_created_at CHECK (deleted_at IS NULL OR deleted_at >= created_at)
);

CREATE UNIQUE INDEX uq_public_status_pages_slug ON public_status_pages (slug);
CREATE INDEX idx_public_status_pages_project_active
    ON public_status_pages (project_id)
    WHERE deleted_at IS NULL;

CREATE TRIGGER set_public_status_pages_updated_at
    BEFORE UPDATE ON public_status_pages
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

CREATE TABLE public_status_page_elements (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    public_page_id uuid NOT NULL,
    project_id uuid NOT NULL,
    parent_element_id uuid,
    kind public_status_element_kind NOT NULL,
    check_id uuid,
    title text,
    description text,
    sort_order integer NOT NULL DEFAULT 0,
    chart_mode public_status_chart_mode NOT NULL DEFAULT 'inherit',
    chart_range public_status_chart_range,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT uq_public_status_page_elements_page_id_id UNIQUE (public_page_id, id),
    CONSTRAINT public_status_page_elements_title_not_empty CHECK (title IS NULL OR length(btrim(title)) > 0),
    CONSTRAINT public_status_page_elements_description_not_empty CHECK (description IS NULL OR length(btrim(description)) > 0),
    CONSTRAINT public_status_page_elements_sort_order_non_negative CHECK (sort_order >= 0),
    CONSTRAINT public_status_page_elements_folder_shape CHECK (
        (kind = 'folder' AND check_id IS NULL)
        OR (kind = 'check' AND check_id IS NOT NULL)
    ),
    CONSTRAINT public_status_page_elements_parent_not_self CHECK (parent_element_id IS NULL OR parent_element_id <> id),
    CONSTRAINT fk_public_status_page_elements_page_project
        FOREIGN KEY (public_page_id, project_id) REFERENCES public_status_pages(id, project_id) ON DELETE CASCADE,
    CONSTRAINT fk_public_status_page_elements_parent
        FOREIGN KEY (public_page_id, parent_element_id) REFERENCES public_status_page_elements(public_page_id, id) ON DELETE CASCADE,
    CONSTRAINT fk_public_status_page_elements_check
        FOREIGN KEY (project_id, check_id) REFERENCES checks(project_id, id) ON DELETE CASCADE
);

CREATE INDEX idx_public_status_page_elements_parent_order
    ON public_status_page_elements (public_page_id, parent_element_id, sort_order, created_at, id);
CREATE INDEX idx_public_status_page_elements_project_check
    ON public_status_page_elements (project_id, check_id)
    WHERE check_id IS NOT NULL;

CREATE TRIGGER set_public_status_page_elements_updated_at
    BEFORE UPDATE ON public_status_page_elements
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

-- +goose Down
DROP TABLE IF EXISTS public_status_page_elements;
DROP TABLE IF EXISTS public_status_pages;

DROP TYPE IF EXISTS public_status_element_kind;
DROP TYPE IF EXISTS public_status_chart_range;
DROP TYPE IF EXISTS public_status_chart_mode;
