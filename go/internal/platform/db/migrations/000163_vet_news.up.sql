-- Veille scientifique / sanitaire vétérinaire (multi-sources).
CREATE TABLE IF NOT EXISTS ops.vet_news_articles (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  source_id TEXT NOT NULL,
  source_name TEXT NOT NULL,
  source_url TEXT NOT NULL,
  title TEXT NOT NULL,
  summary TEXT NOT NULL DEFAULT '',
  image_url TEXT,
  published_at TIMESTAMPTZ,
  category TEXT NOT NULL DEFAULT 'general'
    CHECK (category IN ('epidemio', 'regulatory', 'clinical', 'science', 'practice', 'general')),
  importance TEXT NOT NULL DEFAULT 'low'
    CHECK (importance IN ('critical', 'high', 'medium', 'low')),
  tags TEXT[] NOT NULL DEFAULT '{}',
  content_hash TEXT NOT NULL DEFAULT '',
  fetched_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS vet_news_articles_source_url_uidx
  ON ops.vet_news_articles (source_url);

CREATE INDEX IF NOT EXISTS vet_news_articles_published_idx
  ON ops.vet_news_articles (published_at DESC NULLS LAST);

CREATE INDEX IF NOT EXISTS vet_news_articles_importance_idx
  ON ops.vet_news_articles (importance);

CREATE INDEX IF NOT EXISTS vet_news_articles_source_id_idx
  ON ops.vet_news_articles (source_id);
