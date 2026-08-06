-- Commercial CRM outbound mail: shared templates, send history, open/click tracking.

ALTER TABLE sales.prospects
    ADD COLUMN IF NOT EXISTS email_opt_out BOOLEAN NOT NULL DEFAULT FALSE;

CREATE TABLE IF NOT EXISTS sales.email_templates (
    id UUID PRIMARY KEY,
    slug TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    category TEXT NOT NULL DEFAULT 'intro'
        CHECK (category IN ('intro', 'rdv', 'nurture', 'post', 'reactivation')),
    locale TEXT NOT NULL DEFAULT 'fr',
    subject TEXT NOT NULL,
    body_html TEXT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    updated_by UUID REFERENCES identity.users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_sales_email_templates_category
    ON sales.email_templates(category) WHERE is_active;

CREATE TABLE IF NOT EXISTS sales.email_sends (
    id UUID PRIMARY KEY,
    prospect_id UUID NOT NULL REFERENCES sales.prospects(id) ON DELETE CASCADE,
    template_id UUID REFERENCES sales.email_templates(id) ON DELETE SET NULL,
    commercial_user_id UUID NOT NULL REFERENCES identity.users(id) ON DELETE CASCADE,
    to_email TEXT NOT NULL,
    subject TEXT NOT NULL,
    body_html_rendered TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'sent'
        CHECK (status IN ('sent', 'failed')),
    error TEXT NOT NULL DEFAULT '',
    open_token TEXT NOT NULL UNIQUE,
    opened_at TIMESTAMPTZ,
    open_count INT NOT NULL DEFAULT 0,
    sent_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_sales_email_sends_prospect
    ON sales.email_sends(prospect_id, sent_at DESC NULLS LAST);
CREATE INDEX IF NOT EXISTS idx_sales_email_sends_commercial
    ON sales.email_sends(commercial_user_id, sent_at DESC NULLS LAST);

CREATE TABLE IF NOT EXISTS sales.email_clicks (
    id UUID PRIMARY KEY,
    send_id UUID NOT NULL REFERENCES sales.email_sends(id) ON DELETE CASCADE,
    click_token TEXT NOT NULL UNIQUE,
    target_url TEXT NOT NULL,
    clicked_at TIMESTAMPTZ,
    click_count INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_sales_email_clicks_send
    ON sales.email_clicks(send_id);

DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'petsfollow_app') THEN
    GRANT SELECT, INSERT, UPDATE, DELETE ON
        sales.email_templates, sales.email_sends, sales.email_clicks
        TO petsfollow_app;
  END IF;
END $$;
