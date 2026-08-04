-- Catalogue d'espèces piloté par l'admin + règles réglementaires par pays.
--
-- Deux tables :
--   pets.species               — « quels animaux existent » (universel)
--   pets.species_country_rules — « ce que le régulateur en dit ici » (par pays)
--
-- Le seed BE reproduit à l'identique kernel.IsFoodChainSpecies / kernel.DefaultFoodChainStatus
-- (go/pkg/kernel/kernel.go) → aucune régression sur le comportement chaîne alimentaire.
--
-- Libellés : seedés en fr/nl/en. Les 13 espèces historiques ont déjà des clés i18n compilées
-- côté Nuxt (common.species.*) et Flutter (.arb) dans les 8 locales ; les labels JSONB ne
-- servent que de repli pour les espèces ajoutées par l'admin sans clé compilée.

CREATE TABLE IF NOT EXISTS pets.species (
    code                TEXT PRIMARY KEY,
    labels              JSONB NOT NULL DEFAULT '{}'::jsonb,
    sort_order          INT NOT NULL DEFAULT 100,
    is_active           BOOLEAN NOT NULL DEFAULT true,
    supports_heart_rate BOOLEAN NOT NULL DEFAULT false,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- daf_required :
--   never      — espèce non productrice de denrées → DAF refusé
--   always     — espèce productrice de denrées par nature → DAF autorisé
--   per_animal — statut décidé par individu (équidés : passeport) → DAF autorisé
--                sauf si pets.pets.food_chain_status = 'excluded_from_food_chain'
CREATE TABLE IF NOT EXISTS pets.species_country_rules (
    species_code    TEXT NOT NULL REFERENCES pets.species(code) ON DELETE CASCADE,
    country_code    CHAR(2) NOT NULL,
    is_large_animal BOOLEAN NOT NULL DEFAULT false,
    daf_required    TEXT NOT NULL DEFAULT 'never'
        CHECK (daf_required IN ('never', 'always', 'per_animal')),
    is_food_chain   BOOLEAN NOT NULL DEFAULT false,
    default_food_chain_status TEXT NOT NULL DEFAULT 'companion'
        CHECK (default_food_chain_status IN ('companion', 'food_producing', 'excluded_from_food_chain')),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (species_code, country_code)
);

CREATE INDEX IF NOT EXISTS idx_species_rules_country
    ON pets.species_country_rules (country_code);

-- ---------------------------------------------------------------------------
-- Catalogue : 13 espèces historiques (actives) + extensions (inactives)
-- ---------------------------------------------------------------------------
-- `other` garde sort_order 999 : plusieurs tests verrouillent sa position finale.
INSERT INTO pets.species (code, labels, sort_order, is_active, supports_heart_rate) VALUES
    ('dog',     '{"fr":"Chien","nl":"Hond","en":"Dog"}',           10,  true,  true),
    ('cat',     '{"fr":"Chat","nl":"Kat","en":"Cat"}',             20,  true,  true),
    ('horse',   '{"fr":"Cheval","nl":"Paard","en":"Horse"}',       30,  true,  true),
    ('donkey',  '{"fr":"Âne","nl":"Ezel","en":"Donkey"}',          40,  true,  false),
    ('cattle',  '{"fr":"Bovin","nl":"Rund","en":"Cattle"}',        50,  true,  false),
    ('sheep',   '{"fr":"Ovin","nl":"Schaap","en":"Sheep"}',        60,  true,  false),
    ('goat',    '{"fr":"Caprin","nl":"Geit","en":"Goat"}',         70,  true,  false),
    ('pig',     '{"fr":"Porcin","nl":"Varken","en":"Pig"}',        80,  true,  false),
    ('poultry', '{"fr":"Volaille","nl":"Pluimvee","en":"Poultry"}', 90, true,  false),
    ('rabbit',  '{"fr":"Lapin","nl":"Konijn","en":"Rabbit"}',      100, true,  false),
    ('alpaca',  '{"fr":"Alpaga","nl":"Alpaca","en":"Alpaca"}',     110, true,  false),
    ('llama',   '{"fr":"Lama","nl":"Lama","en":"Llama"}',          120, true,  false),

    -- Gros animaux producteurs de denrées (inactifs : cf. correctif Flutter avant activation)
    ('buffalo', '{"fr":"Buffle","nl":"Buffel","en":"Buffalo"}',                        200, false, false),
    ('deer',    '{"fr":"Cervidé d''élevage","nl":"Gekweekt hert","en":"Farmed deer"}', 210, false, false),
    ('camel',   '{"fr":"Camélidé (chameau)","nl":"Kameelachtige","en":"Camel"}',       220, false, false),
    ('ostrich', '{"fr":"Ratite (autruche)","nl":"Loopvogel","en":"Ratite (ostrich)"}', 230, false, false),
    ('mule',    '{"fr":"Mulet","nl":"Muildier","en":"Mule"}',                          240, false, false),

    -- Petits animaux producteurs de denrées
    ('bee',         '{"fr":"Abeille","nl":"Bij","en":"Bee"}',                             300, false, false),
    ('farmed_fish', '{"fr":"Poisson d''aquaculture","nl":"Kweekvis","en":"Farmed fish"}', 310, false, false),
    ('pigeon',      '{"fr":"Pigeon","nl":"Duif","en":"Pigeon"}',                          320, false, false),

    -- NAC mammifères
    ('ferret',     '{"fr":"Furet","nl":"Fret","en":"Ferret"}',              400, false, false),
    ('guinea_pig', '{"fr":"Cochon d''Inde","nl":"Cavia","en":"Guinea pig"}', 410, false, false),
    ('hamster',    '{"fr":"Hamster","nl":"Hamster","en":"Hamster"}',         420, false, false),
    ('rat',        '{"fr":"Rat","nl":"Rat","en":"Rat"}',                     430, false, false),
    ('mouse',      '{"fr":"Souris","nl":"Muis","en":"Mouse"}',               440, false, false),
    ('gerbil',     '{"fr":"Gerbille","nl":"Gerbil","en":"Gerbil"}',          450, false, false),
    ('chinchilla', '{"fr":"Chinchilla","nl":"Chinchilla","en":"Chinchilla"}', 460, false, false),
    ('hedgehog',   '{"fr":"Hérisson","nl":"Egel","en":"Hedgehog"}',          470, false, false),

    -- Oiseaux non producteurs
    ('parrot',          '{"fr":"Perroquet","nl":"Papegaai","en":"Parrot"}',                500, false, false),
    ('ornamental_bird', '{"fr":"Oiseau d''ornement","nl":"Siervogel","en":"Ornamental bird"}', 510, false, false),
    ('raptor',          '{"fr":"Rapace","nl":"Roofvogel","en":"Raptor"}',                  520, false, false),

    -- Reptiles / amphibiens
    ('tortoise',  '{"fr":"Tortue","nl":"Schildpad","en":"Tortoise"}',   600, false, false),
    ('snake',     '{"fr":"Serpent","nl":"Slang","en":"Snake"}',         610, false, false),
    ('lizard',    '{"fr":"Lézard","nl":"Hagedis","en":"Lizard"}',       620, false, false),
    ('amphibian', '{"fr":"Amphibien","nl":"Amfibie","en":"Amphibian"}', 630, false, false),

    -- Aquatique d'ornement / divers
    ('ornamental_fish', '{"fr":"Poisson d''ornement","nl":"Siervis","en":"Ornamental fish"}', 700, false, false),
    ('wildlife',        '{"fr":"Faune sauvage","nl":"Wilde fauna","en":"Wildlife"}',          800, false, false),
    ('exotic_other',    '{"fr":"Autre exotique","nl":"Andere exoot","en":"Other exotic"}',    900, false, false),

    ('other',   '{"fr":"Autre","nl":"Andere","en":"Other"}',            999, true,  false)
ON CONFLICT (code) DO NOTHING;

-- ---------------------------------------------------------------------------
-- Règles Belgique — AR du 21 juillet 2016 (animaux producteurs de denrées)
-- ---------------------------------------------------------------------------
INSERT INTO pets.species_country_rules
    (species_code, country_code, is_large_animal, daf_required, is_food_chain, default_food_chain_status) VALUES
    -- Animaux de compagnie — jamais de DAF
    ('dog',   'BE', false, 'never', false, 'companion'),
    ('cat',   'BE', false, 'never', false, 'companion'),
    ('other', 'BE', false, 'never', false, 'companion'),

    -- Équidés — producteurs de denrées sauf exclusion par passeport
    ('horse',  'BE', true, 'per_animal', true, 'companion'),
    ('donkey', 'BE', true, 'per_animal', true, 'companion'),
    ('mule',   'BE', true, 'per_animal', true, 'companion'),

    -- Ruminants, porcins, camélidés, gibier d'élevage — producteurs de denrées
    ('cattle',  'BE', true, 'always', true, 'food_producing'),
    ('buffalo', 'BE', true, 'always', true, 'food_producing'),
    ('sheep',   'BE', true, 'always', true, 'food_producing'),
    ('goat',    'BE', true, 'always', true, 'food_producing'),
    ('pig',     'BE', true, 'always', true, 'food_producing'),
    ('alpaca',  'BE', true, 'always', true, 'food_producing'),
    ('llama',   'BE', true, 'always', true, 'food_producing'),
    ('camel',   'BE', true, 'always', true, 'food_producing'),
    ('deer',    'BE', true, 'always', true, 'food_producing'),
    ('ostrich', 'BE', true, 'always', true, 'food_producing'),

    -- Producteurs de denrées qui ne sont PAS des gros animaux
    ('poultry',     'BE', false, 'always',     true, 'food_producing'),
    ('rabbit',      'BE', false, 'always',     true, 'companion'),
    ('bee',         'BE', false, 'always',     true, 'food_producing'),  -- miel
    ('farmed_fish', 'BE', false, 'always',     true, 'food_producing'),
    ('pigeon',      'BE', false, 'per_animal', true, 'companion'),       -- colombophilie

    -- NAC, oiseaux d'ornement, reptiles, faune sauvage — jamais de DAF
    ('ferret',          'BE', false, 'never', false, 'companion'),
    ('guinea_pig',      'BE', false, 'never', false, 'companion'),
    ('hamster',         'BE', false, 'never', false, 'companion'),
    ('rat',             'BE', false, 'never', false, 'companion'),
    ('mouse',           'BE', false, 'never', false, 'companion'),
    ('gerbil',          'BE', false, 'never', false, 'companion'),
    ('chinchilla',      'BE', false, 'never', false, 'companion'),
    ('hedgehog',        'BE', false, 'never', false, 'companion'),
    ('parrot',          'BE', false, 'never', false, 'companion'),
    ('ornamental_bird', 'BE', false, 'never', false, 'companion'),
    ('raptor',          'BE', false, 'never', false, 'companion'),
    ('tortoise',        'BE', false, 'never', false, 'companion'),
    ('snake',           'BE', false, 'never', false, 'companion'),
    ('lizard',          'BE', false, 'never', false, 'companion'),
    ('amphibian',       'BE', false, 'never', false, 'companion'),
    ('ornamental_fish', 'BE', false, 'never', false, 'companion'),
    ('wildlife',        'BE', false, 'never', false, 'companion'),
    ('exotic_other',    'BE', false, 'never', false, 'companion')
ON CONFLICT (species_code, country_code) DO NOTHING;

-- ---------------------------------------------------------------------------
-- Débloquer le cardio : le CHECK figé sur dog/cat/horse empêchait toute extension.
-- ---------------------------------------------------------------------------
ALTER TABLE heartrate.species_alert_deltas
    DROP CONSTRAINT IF EXISTS species_alert_deltas_species_check;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_constraint WHERE conname = 'species_alert_deltas_species_fkey'
  ) THEN
    ALTER TABLE heartrate.species_alert_deltas
      ADD CONSTRAINT species_alert_deltas_species_fkey
      FOREIGN KEY (species) REFERENCES pets.species(code) ON DELETE CASCADE;
  END IF;
END $$;
