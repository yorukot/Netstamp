-- +goose Up
ALTER TABLE public_page_folder_checks
    DROP CONSTRAINT pk_public_page_folder_checks;

ALTER TABLE public_page_folder_checks
    ADD CONSTRAINT pk_public_page_folder_checks PRIMARY KEY (public_page_id, folder_id, check_id);

-- +goose Down
ALTER TABLE public_page_folder_checks
    DROP CONSTRAINT pk_public_page_folder_checks;

WITH ranked AS (
    SELECT ctid,
           row_number() OVER (
               PARTITION BY public_page_id, check_id
               ORDER BY sort_order ASC, created_at ASC, folder_id ASC
           ) AS row_number
    FROM public_page_folder_checks
)
DELETE FROM public_page_folder_checks
WHERE ctid IN (
    SELECT ctid
    FROM ranked
    WHERE row_number > 1
);

ALTER TABLE public_page_folder_checks
    ADD CONSTRAINT pk_public_page_folder_checks PRIMARY KEY (public_page_id, check_id);
