-- Pitch training: scripts, agent prompt versions, simulations, commercial feedback, analyzer runs.

CREATE TABLE IF NOT EXISTS sales.pitch_scripts (
    id UUID PRIMARY KEY,
    slug TEXT NOT NULL,
    title TEXT NOT NULL,
    audience TEXT NOT NULL DEFAULT 'vet'
        CHECK (audience IN ('vet', 'client')),
    owner_user_id UUID REFERENCES identity.users(id) ON DELETE CASCADE,
    parent_script_id UUID REFERENCES sales.pitch_scripts(id) ON DELETE SET NULL,
    steps_json JSONB NOT NULL DEFAULT '[]'::jsonb,
    example_dialogue_json JSONB NOT NULL DEFAULT '[]'::jsonb,
    coach_hints TEXT NOT NULL DEFAULT '',
    locale TEXT NOT NULL DEFAULT 'fr',
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_pitch_scripts_slug_admin
    ON sales.pitch_scripts(slug)
    WHERE owner_user_id IS NULL;

CREATE INDEX IF NOT EXISTS idx_pitch_scripts_owner
    ON sales.pitch_scripts(owner_user_id)
    WHERE owner_user_id IS NOT NULL;

CREATE TABLE IF NOT EXISTS sales.agent_prompt_versions (
    id UUID PRIMARY KEY,
    agent_kind TEXT NOT NULL CHECK (agent_kind IN ('vet_live', 'coach')),
    version INT NOT NULL,
    content_json JSONB NOT NULL,
    changelog TEXT NOT NULL DEFAULT '',
    source TEXT NOT NULL DEFAULT 'seed'
        CHECK (source IN ('seed', 'admin', 'analyzer')),
    created_by UUID REFERENCES identity.users(id) ON DELETE SET NULL,
    analyzer_run_id UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (agent_kind, version)
);

CREATE TABLE IF NOT EXISTS sales.agent_prompt_current (
    agent_kind TEXT PRIMARY KEY CHECK (agent_kind IN ('vet_live', 'coach')),
    version_id UUID NOT NULL REFERENCES sales.agent_prompt_versions(id)
);

CREATE TABLE IF NOT EXISTS sales.pitch_analyzer_runs (
    id UUID PRIMARY KEY,
    started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    finished_at TIMESTAMPTZ,
    feedback_count INT NOT NULL DEFAULT 0,
    status TEXT NOT NULL DEFAULT 'noop'
        CHECK (status IN ('noop', 'applied', 'needs_review', 'failed')),
    input_summary_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    output_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    vet_version_id UUID REFERENCES sales.agent_prompt_versions(id),
    coach_version_id UUID REFERENCES sales.agent_prompt_versions(id)
);

ALTER TABLE sales.agent_prompt_versions
    DROP CONSTRAINT IF EXISTS agent_prompt_versions_analyzer_run_fk;
ALTER TABLE sales.agent_prompt_versions
    ADD CONSTRAINT agent_prompt_versions_analyzer_run_fk
    FOREIGN KEY (analyzer_run_id) REFERENCES sales.pitch_analyzer_runs(id) ON DELETE SET NULL;

CREATE TABLE IF NOT EXISTS sales.pitch_simulations (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES identity.users(id) ON DELETE CASCADE,
    script_id UUID NOT NULL REFERENCES sales.pitch_scripts(id),
    interest_level TEXT NOT NULL DEFAULT 'neutre'
        CHECK (interest_level IN ('hostile', 'sceptique', 'neutre', 'interesse', 'chaud')),
    voice_name TEXT NOT NULL DEFAULT 'Charon',
    vet_prompt_version_id UUID REFERENCES sales.agent_prompt_versions(id),
    coach_prompt_version_id UUID REFERENCES sales.agent_prompt_versions(id),
    outcome TEXT NOT NULL DEFAULT 'manual'
        CHECK (outcome IN ('appointment', 'hangup', 'timeout', 'manual', 'in_progress')),
    appointment_slot TEXT NOT NULL DEFAULT '',
    duration_sec INT NOT NULL DEFAULT 0,
    ended_at TIMESTAMPTZ,
    transcript_json JSONB NOT NULL DEFAULT '[]'::jsonb,
    coach_feedback_json JSONB,
    ai_score NUMERIC(4,1),
    user_score NUMERIC(4,1),
    audio_object_key TEXT NOT NULL DEFAULT '',
    is_top5 BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_pitch_sims_user_created
    ON sales.pitch_simulations(user_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_pitch_sims_user_top5
    ON sales.pitch_simulations(user_id, is_top5)
    WHERE is_top5 = true;

CREATE TABLE IF NOT EXISTS sales.pitch_sim_feedback (
    id UUID PRIMARY KEY,
    simulation_id UUID NOT NULL UNIQUE REFERENCES sales.pitch_simulations(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES identity.users(id) ON DELETE CASCADE,
    vet_realism SMALLINT NOT NULL CHECK (vet_realism BETWEEN 1 AND 5),
    coach_usefulness SMALLINT NOT NULL CHECK (coach_usefulness BETWEEN 1 AND 5),
    difficulty_felt TEXT NOT NULL DEFAULT 'ok'
        CHECK (difficulty_felt IN ('too_easy', 'ok', 'too_hard')),
    comment TEXT NOT NULL DEFAULT '',
    flags JSONB NOT NULL DEFAULT '[]'::jsonb,
    analyzer_processed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_pitch_feedback_unprocessed
    ON sales.pitch_sim_feedback(analyzer_processed_at)
    WHERE analyzer_processed_at IS NULL;

-- Seed: default pitch script + agent prompts (ids fixed for idempotent re-runs via ON CONFLICT).
INSERT INTO sales.pitch_scripts (
    id, slug, title, audience, owner_user_id, parent_script_id,
    steps_json, example_dialogue_json, coach_hints, locale, is_active
) VALUES (
    'a0000000-0000-4000-8000-000000000001',
    'vente-petsfollow-veto',
    'Vente petsFollow — cabinet véto',
    'vet',
    NULL,
    NULL,
    '[
      {"id":"open","title":"Ouverture 30 s","talkingPoints":["Se présenter","petsFollow: suivi cardiaque prescrit, sans boîtier","Cabinet gratuit, client < 3 €/mois"],"exampleLine":"Bonjour docteur, je suis [Prénom] de petsFollow…"},
      {"id":"value","title":"Valeur cabinet","talkingPoints":["Suivi prescrit sans hardware","Messagerie + timeline","Activation clients payants"],"exampleLine":"Vos clients paient, vous gagnez sur chaque activation."},
      {"id":"money","title":"Commission véto","talkingPoints":["Même plafond avec ou sans commercial","Steer triennial ~9,4 €","Bonus ramp 5 pets / 60 j"],"exampleLine":"Sur le triennial, jusqu’à environ 9,4 € pour vous."},
      {"id":"objections","title":"Objections","talkingPoints":["Pas un boîtier","Pas de % sur le TTC","Inscription ≠ revenu"],"exampleLine":"Non, pas d’appareil — relevé au doigt dans l’app."},
      {"id":"cta","title":"CTA RDV","talkingPoints":["Proposer une démo 20 min","Proposer un créneau concret"],"exampleLine":"On peut caler 20 minutes cette semaine ?"}
    ]'::jsonb,
    '[
      {"role":"commercial","text":"Bonjour docteur, je suis Léa de petsFollow. Je vous appelle pour un suivi cardiaque prescrit, sans boîtier."},
      {"role":"vet","text":"Allo… encore un outil ? On est déjà saturés."},
      {"role":"commercial","text":"Je comprends. petsFollow est gratuit pour le cabinet : ce sont vos clients qui paient moins de 3 € par mois."},
      {"role":"vet","text":"Et moi, j’y gagne quoi ?"},
      {"role":"commercial","text":"Une commission sur chaque activation — jusqu’à environ 9,4 € sur le plan triennal, même plafond avec ou sans commercial."},
      {"role":"vet","text":"Il faut un appareil ?"},
      {"role":"commercial","text":"Non, relevé au doigt dans l’app, durée définie par le cabinet. Je vous propose une démo de 20 minutes mardi 10 h ?"},
      {"role":"vet","text":"Mardi 10 h, d’accord. Envoyez-moi le lien."}
    ]'::jsonb,
    'Vérifier: pas de promesse boîtier; pas de % sur TTC; inscription ≠ revenu; CTA RDV clair.',
    'fr',
    true
) ON CONFLICT (id) DO NOTHING;

INSERT INTO sales.agent_prompt_versions (id, agent_kind, version, content_json, changelog, source)
VALUES (
    'a0000000-0000-4000-8000-000000000010',
    'vet_live',
    1,
    '{
      "basePersona": "Tu es un vétérinaire belgo-français d’un cabinet de ville. Tu réponds au téléphone en français uniquement. Tu commences toujours par « Allo ». Tu restes dans le rôle: tu ne parles jamais comme un assistant IA. Tu connais vaguement les logiciels vétos mais pas petsFollow. Objectifs d’appel possibles: accepter un RDV démo OU raccrocher si pas intéressé.",
      "productFacts": "petsFollow = suivi cardiaque animal prescrit, sans boîtier, relevé 15/30/60s dans l’app. Pro gratuit pour le cabinet; client paie (~2–3€/mois). Pas de hardware. Pas de chat WebSocket temps réel.",
      "difficulty": {
        "hostile": "Tu es impatient, coupes la parole, objections sèches. Tu refuses presque toujours. Utilise hang_up_not_interested rapidement si le pitch est faible.",
        "sceptique": "Tu écoutes peu, beaucoup d’objections (boîtier, abo, commission). RDV seulement si pitch excellent et CTA clair.",
        "neutre": "Tu es poli mais non convaincu. Tu poses 2–3 questions. RDV possible si valeur + CTA clairs.",
        "interesse": "Tu es curieux, objections légères. Tu acceptes un RDV si le commercial mène bien.",
        "chaud": "Tu es déjà positif, peu de friction. Tu acceptes presque toujours un RDV démo."
      },
      "tools": "Quand tu acceptes: appelle book_appointment avec un créneau fictif. Quand tu refuses: appelle hang_up_not_interested avec une raison courte."
    }'::jsonb,
    'Seed initial véto Live',
    'seed'
) ON CONFLICT (id) DO NOTHING;

INSERT INTO sales.agent_prompt_versions (id, agent_kind, version, content_json, changelog, source)
VALUES (
    'a0000000-0000-4000-8000-000000000011',
    'coach',
    1,
    '{
      "system": "Tu es un coach commercial senior spécialisé vente B2B logiciels vétérinaires (petsFollow). Tu analyses un transcript d’appel d’entraînement. Tu es direct, bienveillant, actionnable. Réponds uniquement en JSON valide.",
      "rubric": ["opener","listening","objections","offerClarity","cta"],
      "rules": "Contextualise selon interest_level (difficulté). Signale les interdits: promesse boîtier, % sur TTC, confondre inscription et revenu. score sur 10."
    }'::jsonb,
    'Seed initial coach',
    'seed'
) ON CONFLICT (id) DO NOTHING;

INSERT INTO sales.agent_prompt_current (agent_kind, version_id) VALUES
    ('vet_live', 'a0000000-0000-4000-8000-000000000010'),
    ('coach', 'a0000000-0000-4000-8000-000000000011')
ON CONFLICT (agent_kind) DO NOTHING;

DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'petsfollow_app') THEN
    GRANT SELECT, INSERT, UPDATE, DELETE ON
      sales.pitch_scripts,
      sales.agent_prompt_versions,
      sales.agent_prompt_current,
      sales.pitch_analyzer_runs,
      sales.pitch_simulations,
      sales.pitch_sim_feedback
      TO petsfollow_app;
  END IF;
END $$;
