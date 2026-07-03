-- +goose Up
CREATE OR REPLACE VIEW public_status_page_assignment_scope AS
SELECT public_status_page_elements.public_page_id,
       public_status_page_elements.id AS element_id,
       probe_check_assignments.id AS assignment_id,
       probe_check_assignments.project_id,
       probe_check_assignments.probe_id,
       probe_check_assignments.check_id
FROM public_status_page_elements
JOIN probe_check_assignments
  ON public_status_page_elements.assignment_selection_mode = 'all_check'
 AND probe_check_assignments.project_id = public_status_page_elements.project_id
 AND probe_check_assignments.check_id = public_status_page_elements.check_id
 AND probe_check_assignments.deleted_at IS NULL
JOIN checks
  ON checks.project_id = probe_check_assignments.project_id
 AND checks.id = probe_check_assignments.check_id
 AND checks.deleted_at IS NULL
 AND checks.check_type IN ('ping', 'tcp')
WHERE public_status_page_elements.kind = 'assignment_group'
UNION ALL
SELECT public_status_page_elements.public_page_id,
       public_status_page_elements.id AS element_id,
       probe_check_assignments.id AS assignment_id,
       probe_check_assignments.project_id,
       probe_check_assignments.probe_id,
       probe_check_assignments.check_id
FROM public_status_page_elements
JOIN public_status_page_element_assignments
  ON public_status_page_element_assignments.public_page_id = public_status_page_elements.public_page_id
 AND public_status_page_element_assignments.element_id = public_status_page_elements.id
JOIN probe_check_assignments
  ON probe_check_assignments.id = public_status_page_element_assignments.assignment_id
 AND probe_check_assignments.project_id = public_status_page_elements.project_id
 AND probe_check_assignments.deleted_at IS NULL
JOIN checks
  ON checks.project_id = probe_check_assignments.project_id
 AND checks.id = probe_check_assignments.check_id
 AND checks.deleted_at IS NULL
 AND checks.check_type IN ('ping', 'tcp')
WHERE public_status_page_elements.kind = 'assignment_group'
  AND public_status_page_elements.assignment_selection_mode = 'selected_assignments';

-- +goose Down
DROP VIEW IF EXISTS public_status_page_assignment_scope;
