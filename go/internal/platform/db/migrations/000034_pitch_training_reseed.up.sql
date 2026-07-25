-- Re-seed pitch training defaults if missing (staging / DBs migrated before seed rows).

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
) ON CONFLICT (id) DO UPDATE SET
    is_active = true,
    title = EXCLUDED.title,
    steps_json = EXCLUDED.steps_json,
    example_dialogue_json = EXCLUDED.example_dialogue_json,
    coach_hints = EXCLUDED.coach_hints,
    updated_at = NOW();

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
