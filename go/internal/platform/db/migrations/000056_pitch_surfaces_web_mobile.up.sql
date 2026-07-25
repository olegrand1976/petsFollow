-- Pitch copy: drop "sans boîtier" framing; emphasize Web Pro + mobile ProLight + mobile Client.

UPDATE sales.pitch_scripts SET
    steps_json = '[
      {"id":"open","title":"Ouverture 30 s","talkingPoints":["Se présenter","petsFollow: continuité de soins prescrite — Web cabinet + mobile ProLight + app client","Messagerie, Care, relevés FC inclus — client ≤ 3,5 €/mois"],"exampleLine":"Bonjour docteur, je suis [Prénom] de petsFollow…"},
      {"id":"value","title":"Valeur cabinet","talkingPoints":["Web Pro pour le cabinet","Mobile ProLight pour le terrain","App propriétaire pour le foyer"],"exampleLine":"Vos clients paient, vous gagnez sur chaque activation."},
      {"id":"money","title":"Commission véto","talkingPoints":["Même plafond avec ou sans commercial","Steer triennial ~9,4 €","Bonus ramp 5 pets / 60 j"],"exampleLine":"Sur le triennial, jusqu’à environ 9,4 € pour vous."},
      {"id":"objections","title":"Objections","talkingPoints":["Pas juste une app cardio","Pas un 2ᵉ PMS","Pas de % sur le TTC","Inscription ≠ revenu"],"exampleLine":"Ce n’est pas un PMS — c’est la continuité Web + mobile autour de vos patients."},
      {"id":"cta","title":"CTA RDV","talkingPoints":["Proposer une démo 20 min","Proposer un créneau concret"],"exampleLine":"On peut caler 20 minutes cette semaine ?"}
    ]'::jsonb,
    example_dialogue_json = '[
      {"role":"commercial","text":"Bonjour docteur, je suis Léa de petsFollow. Continuité de soins prescrite — Web pour le cabinet, mobile ProLight pour le terrain, app pour vos clients."},
      {"role":"vet","text":"Allo… encore un outil ? On est déjà saturés."},
      {"role":"commercial","text":"Je comprends. petsFollow complète votre PMS : messagerie, Care, relevés — ce sont vos clients qui paient ≤ 3,5 € par mois."},
      {"role":"vet","text":"Et moi, j’y gagne quoi ?"},
      {"role":"commercial","text":"Une commission sur chaque activation — jusqu’à environ 9,4 € sur le plan triennal, même plafond avec ou sans commercial."},
      {"role":"vet","text":"C’est juste du cardio ?"},
      {"role":"commercial","text":"Non — continuité Web + mobile ; le relevé cardiaque est une feature. Je vous propose une démo de 20 minutes mardi 10 h ?"},
      {"role":"vet","text":"Mardi 10 h, d’accord. Envoyez-moi le lien."}
    ]'::jsonb,
    coach_hints = 'Vérifier: identité = continuité Web+mobile (pas « app cardio », pas pitch « sans boîtier »); pas de % sur TTC; inscription ≠ revenu; CTA RDV clair.',
    updated_at = NOW()
WHERE id = 'a0000000-0000-4000-8000-000000000001';

UPDATE sales.agent_prompt_versions SET
    content_json = jsonb_set(
        content_json,
        '{productFacts}',
        '"petsFollow = continuité de soins animale prescrite via Web Pro (cabinet), mobile ProLight (terrain) et mobile Client (particulier). Features: messagerie, Care/Horse, relevés FC 15/30/60s, CR IA terrain. Pro SaaS cabinet; client paie (~2–3,5€/mois). Pas de chat WebSocket temps réel. Ne pas cantonner le pitch au cardio ni le centrer sur « sans boîtier »."'::jsonb
    ),
    changelog = 'Surfaces Web+mobile ; retrait framing sans boîtier'
WHERE id = 'a0000000-0000-4000-8000-000000000010'
  AND agent_kind = 'vet_live';
