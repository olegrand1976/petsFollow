-- Reposition pitch training: continuity of care (cardiac = feature, not identity).

UPDATE sales.pitch_scripts SET
    steps_json = '[
      {"id":"open","title":"Ouverture 30 s","talkingPoints":["Se présenter","petsFollow: continuité de soins prescrite, sans boîtier","Messagerie, Care, relevés FC inclus — client ≤ 3,5 €/mois"],"exampleLine":"Bonjour docteur, je suis [Prénom] de petsFollow…"},
      {"id":"value","title":"Valeur cabinet","talkingPoints":["Continuité cabinet / terrain / foyer","Messagerie + timeline + Care/Horse","Activation clients payants"],"exampleLine":"Vos clients paient, vous gagnez sur chaque activation."},
      {"id":"money","title":"Commission véto","talkingPoints":["Même plafond avec ou sans commercial","Steer triennial ~9,4 €","Bonus ramp 5 pets / 60 j"],"exampleLine":"Sur le triennial, jusqu’à environ 9,4 € pour vous."},
      {"id":"objections","title":"Objections","talkingPoints":["Pas juste une app cardio","Pas un boîtier","Pas de % sur le TTC","Inscription ≠ revenu"],"exampleLine":"Non, pas d’appareil — et le FC n’est qu’une feature parmi d’autres."},
      {"id":"cta","title":"CTA RDV","talkingPoints":["Proposer une démo 20 min","Proposer un créneau concret"],"exampleLine":"On peut caler 20 minutes cette semaine ?"}
    ]'::jsonb,
    example_dialogue_json = '[
      {"role":"commercial","text":"Bonjour docteur, je suis Léa de petsFollow. Je vous appelle pour la continuité de soins prescrite, sans boîtier — cabinet, terrain et foyer."},
      {"role":"vet","text":"Allo… encore un outil ? On est déjà saturés."},
      {"role":"commercial","text":"Je comprends. petsFollow complète votre PMS : messagerie, Care, relevés cardiaques — ce sont vos clients qui paient ≤ 3,5 € par mois."},
      {"role":"vet","text":"Et moi, j’y gagne quoi ?"},
      {"role":"commercial","text":"Une commission sur chaque activation — jusqu’à environ 9,4 € sur le plan triennal, même plafond avec ou sans commercial."},
      {"role":"vet","text":"C’est juste du cardio ?"},
      {"role":"commercial","text":"Non — continuité complète ; le relevé cardiaque est une feature. Je vous propose une démo de 20 minutes mardi 10 h ?"},
      {"role":"vet","text":"Mardi 10 h, d’accord. Envoyez-moi le lien."}
    ]'::jsonb,
    coach_hints = 'Vérifier: identité = continuité prescrite (pas « app cardio only »); pas de promesse boîtier; pas de % sur TTC; inscription ≠ revenu; CTA RDV clair.',
    updated_at = NOW()
WHERE id = 'a0000000-0000-4000-8000-000000000001';

UPDATE sales.agent_prompt_versions SET
    content_json = jsonb_set(
        content_json,
        '{productFacts}',
        '"petsFollow = continuité de soins animale prescrite (cabinet / terrain / foyer), sans boîtier. Features: messagerie, Care/Horse, relevés FC 15/30/60s, CR IA terrain. Pro SaaS cabinet; client paie (~2–3,5€/mois). Pas de hardware. Pas de chat WebSocket temps réel. Ne pas cantonner le pitch au cardio."'::jsonb
    ),
    changelog = 'Repositionnement continuité de soins (FC = feature)'
WHERE id = 'a0000000-0000-4000-8000-000000000010'
  AND agent_kind = 'vet_live';
