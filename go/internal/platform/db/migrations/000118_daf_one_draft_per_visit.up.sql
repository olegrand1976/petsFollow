-- Deduplicate draft DAFs sharing the same visit (keep newest), then enforce uniqueness.
WITH ranked AS (
    SELECT id,
           ROW_NUMBER() OVER (PARTITION BY visit_id ORDER BY created_at DESC, id DESC) AS rn
    FROM pharmacy.daf_documents
    WHERE visit_id IS NOT NULL AND status = 'draft'
)
DELETE FROM pharmacy.daf_items
WHERE daf_id IN (SELECT id FROM ranked WHERE rn > 1);

WITH ranked AS (
    SELECT id,
           ROW_NUMBER() OVER (PARTITION BY visit_id ORDER BY created_at DESC, id DESC) AS rn
    FROM pharmacy.daf_documents
    WHERE visit_id IS NOT NULL AND status = 'draft'
)
DELETE FROM pharmacy.daf_documents
WHERE id IN (SELECT id FROM ranked WHERE rn > 1);

-- At most one draft DAF per visit (consultation → DAF automation).
CREATE UNIQUE INDEX IF NOT EXISTS idx_daf_documents_visit_draft_unique
    ON pharmacy.daf_documents (visit_id)
    WHERE visit_id IS NOT NULL AND status = 'draft';
