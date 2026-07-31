-- Unique Billit order id (webhook apply must target exactly one document).
-- Dedupe leftovers from mock ord_N collisions before creating the index.
WITH ranked AS (
    SELECT id,
           ROW_NUMBER() OVER (
               PARTITION BY billit_order_id
               ORDER BY updated_at DESC NULLS LAST, created_at DESC
           ) AS rn
    FROM invoicing.documents
    WHERE billit_order_id IS NOT NULL AND billit_order_id <> ''
)
UPDATE invoicing.documents d
SET billit_order_id = NULL,
    updated_at = now()
FROM ranked r
WHERE d.id = r.id AND r.rn > 1;

CREATE UNIQUE INDEX IF NOT EXISTS idx_invoicing_documents_billit_order_id
    ON invoicing.documents (billit_order_id)
    WHERE billit_order_id IS NOT NULL AND billit_order_id <> '';
