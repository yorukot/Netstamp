-- +goose NO TRANSACTION
-- +goose Up
-- +goose StatementBegin
DO $$
BEGIN
    CREATE TYPE public_status_assignment_selection_mode AS ENUM ('all_check', 'selected_assignments');
EXCEPTION
    WHEN duplicate_object THEN NULL;
END $$;
-- +goose StatementEnd

ALTER TYPE public_status_element_kind ADD VALUE IF NOT EXISTS 'assignment_group';

ALTER TABLE public_status_page_elements
    ADD COLUMN IF NOT EXISTS assignment_selection_mode public_status_assignment_selection_mode;

ALTER TABLE public_status_page_elements
    DROP CONSTRAINT IF EXISTS public_status_page_elements_folder_shape;

ALTER TABLE public_status_page_elements
    DROP CONSTRAINT IF EXISTS public_status_page_elements_shape;

UPDATE public_status_page_elements
SET kind = 'assignment_group',
    assignment_selection_mode = 'all_check'
WHERE kind = 'check';

ALTER TABLE public_status_page_elements
    ADD CONSTRAINT public_status_page_elements_shape CHECK (
        (kind = 'folder' AND check_id IS NULL AND assignment_selection_mode IS NULL)
        OR (kind = 'assignment_group' AND assignment_selection_mode = 'all_check' AND check_id IS NOT NULL)
        OR (kind = 'assignment_group' AND assignment_selection_mode = 'selected_assignments' AND check_id IS NULL)
    );

CREATE TABLE IF NOT EXISTS public_status_page_element_assignments (
    element_id uuid NOT NULL,
    public_page_id uuid NOT NULL,
    project_id uuid NOT NULL,
    assignment_id uuid NOT NULL REFERENCES probe_check_assignments(id) ON DELETE CASCADE,
    sort_order integer NOT NULL DEFAULT 0,
    created_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (element_id, assignment_id),
    CONSTRAINT public_status_page_element_assignments_sort_order_non_negative CHECK (sort_order >= 0),
    CONSTRAINT fk_public_status_page_element_assignments_element
        FOREIGN KEY (public_page_id, element_id) REFERENCES public_status_page_elements(public_page_id, id) ON DELETE CASCADE,
    CONSTRAINT fk_public_status_page_element_assignments_page_project
        FOREIGN KEY (public_page_id, project_id) REFERENCES public_status_pages(id, project_id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_public_status_page_element_assignments_page
    ON public_status_page_element_assignments (public_page_id, element_id, sort_order, assignment_id);
CREATE INDEX IF NOT EXISTS idx_public_status_page_element_assignments_assignment
    ON public_status_page_element_assignments (assignment_id);

-- +goose Down
DROP TABLE IF EXISTS public_status_page_element_assignments;

ALTER TABLE public_status_page_elements
    DROP CONSTRAINT IF EXISTS public_status_page_elements_shape;

UPDATE public_status_page_elements
SET kind = 'check',
    assignment_selection_mode = NULL
WHERE kind = 'assignment_group';

ALTER TABLE public_status_page_elements
    DROP COLUMN IF EXISTS assignment_selection_mode;

ALTER TABLE public_status_page_elements
    ADD CONSTRAINT public_status_page_elements_folder_shape CHECK (
        (kind = 'folder' AND check_id IS NULL)
        OR (kind = 'check' AND check_id IS NOT NULL)
    );

DROP TYPE IF EXISTS public_status_assignment_selection_mode;
