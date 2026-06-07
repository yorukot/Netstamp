-- +goose Up
DROP TABLE public_page_folder_checks;
DROP TABLE public_page_folders;
DROP TABLE public_pages;

-- +goose Down
CREATE TABLE public_pages (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id uuid NOT NULL REFERENCES projects(id),
    slug citext NOT NULL,
    title text NOT NULL,
    description text,
    enabled boolean NOT NULL DEFAULT true,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz,
    CONSTRAINT uq_public_pages_id_project UNIQUE (id, project_id),
    CONSTRAINT public_pages_slug_not_empty CHECK (length(btrim(slug::text)) > 0),
    CONSTRAINT public_pages_title_not_empty CHECK (length(btrim(title)) > 0),
    CONSTRAINT public_pages_description_not_empty CHECK (description IS NULL OR length(btrim(description)) > 0),
    CONSTRAINT public_pages_deleted_at_after_created_at CHECK (deleted_at IS NULL OR deleted_at >= created_at)
);

CREATE UNIQUE INDEX uq_public_pages_slug ON public_pages (slug);
CREATE INDEX idx_public_pages_project_id ON public_pages (project_id);

CREATE TRIGGER set_public_pages_updated_at
    BEFORE UPDATE ON public_pages
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

CREATE TABLE public_page_folders (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    public_page_id uuid NOT NULL REFERENCES public_pages(id) ON DELETE CASCADE,
    parent_id uuid,
    name text NOT NULL,
    description text,
    sort_order integer NOT NULL DEFAULT 0,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT uq_public_page_folders_page_id_id UNIQUE (public_page_id, id),
    CONSTRAINT public_page_folders_name_not_empty CHECK (length(btrim(name)) > 0),
    CONSTRAINT public_page_folders_description_not_empty CHECK (description IS NULL OR length(btrim(description)) > 0),
    CONSTRAINT public_page_folders_sort_order_non_negative CHECK (sort_order >= 0),
    CONSTRAINT public_page_folders_parent_not_self CHECK (parent_id IS NULL OR parent_id <> id),
    CONSTRAINT fk_public_page_folders_parent
        FOREIGN KEY (public_page_id, parent_id) REFERENCES public_page_folders(public_page_id, id) ON DELETE CASCADE
);

CREATE INDEX idx_public_page_folders_page_parent ON public_page_folders (public_page_id, parent_id);

CREATE TRIGGER set_public_page_folders_updated_at
    BEFORE UPDATE ON public_page_folders
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

CREATE TABLE public_page_folder_checks (
    public_page_id uuid NOT NULL,
    project_id uuid NOT NULL,
    folder_id uuid NOT NULL,
    check_id uuid NOT NULL,
    sort_order integer NOT NULL DEFAULT 0,
    created_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT public_page_folder_checks_sort_order_non_negative CHECK (sort_order >= 0),
    CONSTRAINT pk_public_page_folder_checks PRIMARY KEY (public_page_id, folder_id, check_id),
    CONSTRAINT fk_public_page_folder_checks_page_project
        FOREIGN KEY (public_page_id, project_id) REFERENCES public_pages(id, project_id) ON DELETE CASCADE,
    CONSTRAINT fk_public_page_folder_checks_folder
        FOREIGN KEY (public_page_id, folder_id) REFERENCES public_page_folders(public_page_id, id) ON DELETE CASCADE,
    CONSTRAINT fk_public_page_folder_checks_check
        FOREIGN KEY (project_id, check_id) REFERENCES checks(project_id, id) ON DELETE CASCADE
);

CREATE INDEX idx_public_page_folder_checks_folder ON public_page_folder_checks (public_page_id, folder_id);
CREATE INDEX idx_public_page_folder_checks_check ON public_page_folder_checks (check_id);
