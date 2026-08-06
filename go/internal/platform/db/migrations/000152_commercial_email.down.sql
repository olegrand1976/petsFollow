DROP TABLE IF EXISTS sales.email_clicks;
DROP TABLE IF EXISTS sales.email_sends;
DROP TABLE IF EXISTS sales.email_templates;

ALTER TABLE sales.prospects DROP COLUMN IF EXISTS email_opt_out;
