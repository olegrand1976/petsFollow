# 42 — petsFollow Research (observatoire épidémio anonymisé)

Module **tag `dev`** : profil `research` + couche de données anonymisées (opt-in cabinet) + Observatoire Pro (agrégats régionaux). Objectif produit : offrir aux vétérinaires / experts une **donnée de masse** pour statistiques et détection précoce de signaux par région et espèce — sans PII.

## Décisions V1

| Choix | Valeur |
|-------|--------|
| Rôle | `research` (multi-profil, associable à `vet` / `admin`) |
| Contribution | **Opt-in cabinet** (`practice.practices.research_opt_in_at`) |
| Accès V1 | Observatoire agrégé ; k-anonymité ≥ 5 heatmap/alerts |
| Accès V2 | Data room micro-événements (k ≥ 5 **cabinets distincts**) + membership d’un groupe avec `dataroom_enabled` (toggle **admin**) |
| Flags | `RESEARCH_ENABLED` / `NUXT_PUBLIC_RESEARCH_ENABLED` |
| Surface | Nuxt Pro `/research` — pas de shell Flutter Research |

## Architecture

```
Sources cliniques (opt-in practice)
  → ETL interne (anonymisation)
  → research.anon_events + research.weekly_aggregates
  → API /research/* (rôle research)
  → UI Observatoire Nuxt
```

- Géolocalisation : **code postal + ville + pays du cabinet** (pas d’adresse client).
- Identifiants : `practice_id_hash` salé (purge opt-out) — **jamais** exposé à l’API Research.
- Textes libres (CR, `chiefComplaint`) : **exclus** V1.

## Rôle & multi-profil

- Migration CHECK : `identity.users` + `identity.profiles` incluent `research`.
- Kernel : `RoleResearch` — **hors** `IsPracticeStaff` / `IsSalesForce` / `IsOpsRole` ; **hors** `IsProRole` Go (pas d’auto-profil client).
- Nuxt : `isProRole('research')` + `homePathForRole` → `/research`.
- Attach admin : `POST /admin/users/{id}/profiles` avec `role=research`.
- Seed : `research.demo@petsfollow.test` (+ profil `research` sur `vet.demo` / `admin.demo`).

## Schéma `research`

| Table | Grain |
|-------|--------|
| `anon_events` | événement anonymisé (semaine ISO, CP, pays, espèce, age_band, signal_type, payload JSONB, source_hash, practice_id_hash) |
| `weekly_aggregates` | `week × postal_code × country × species × signal_type` + `event_count` |
| `etl_watermarks` | curseurs incrémentaux par source |
| `groups` / `group_members` | groupes collaboratifs V2 (owner/member) + `dataroom_enabled` (admin) |

### Signaux V1

| `signal_type` | Source | Payload |
|---------------|--------|---------|
| `preconsult_syndrome` | préconsult | urgency, appetite, thirst, behavior, duration (**pas** chiefComplaint) |
| `visit_volume` | visits | — |
| `lab_flag` | labs (flag ≠ normal) | analyte, flag |
| `care_preventive` | care completed | type (vaccination / deworming / fecal_egg) |
| `antibiotic_daf` | DAF `has_antibiotic` | — |
| `hr_alert` | heartrate `is_alert` | — |

## Flags

| Env | Défaut |
|-----|--------|
| `RESEARCH_ENABLED` | off · `make api-dev` → `true` |
| `NUXT_PUBLIC_RESEARCH_ENABLED` | idem (`make nuxtjs-dev`) |
| `RESEARCH_ETL_SECRET` | protège `POST /internal/research-etl/run` |
| `RESEARCH_ANON_SALT` | sel HMAC `practice_id_hash` (**obligatoire** hors local/test ; jamais dérivé du secret ETL ; `make api-dev` pose un sel local) |

404 `research_disabled` si flag off. Boot **fail-fast** si `RESEARCH_ENABLED` hors env seedable sans salt + secret ETL.

### Ops GCP

| Élément | Notes |
|---------|--------|
| Staging flags | `RESEARCH_ENABLED=true` + bake `NUXT_PUBLIC_RESEARCH_ENABLED=true` (`deploy-run-args` / Cloud Build) |
| Secrets SM | `petsfollow-research-etl-secret` · `petsfollow-research-anon-salt` |
| Scheduler | `make gcp-research-etl-scheduler` — `POST /internal/research-etl/run` toutes les 6 h (`Europe/Brussels`) |

## Endpoints API

| Méthode | Path | Accès |
|---------|------|-------|
| `GET` | `/api/v1/research/overview` | `research` |
| `GET` | `/api/v1/research/heatmap` | `research` |
| `GET` | `/api/v1/research/timeseries` | `research` |
| `GET` | `/api/v1/research/alerts` | `research` |
| `GET` | `/api/v1/admin/research/opt-ins` | `admin` / `dev` |
| `GET/POST` | `/api/v1/research/groups` | `research` |
| `GET/PATCH/DELETE` | `/api/v1/research/groups/{id}` | `research` (membre / owner) |
| `GET/POST` | `/api/v1/research/groups/{id}/members` | `research` (POST anti-énumération : toujours `{ok:true}`) |
| `DELETE` | `/api/v1/research/groups/{id}/members/{userId}` | `research` |
| `GET` | `/api/v1/admin/research/groups` | `admin` / `dev` |
| `PATCH` | `/api/v1/admin/research/groups/{id}/dataroom` | `admin` / `dev` — `{enabled}` |
| `GET` | `/api/v1/research/dataroom/events` | `research` + membre d’un groupe `dataroom_enabled` |
| `GET` | `/api/v1/vet/practice/research-opt-in` | staff `practice.settings` |
| `POST` | `/api/v1/vet/practice/research-opt-in` | staff `practice.settings` |
| `DELETE` | `/api/v1/vet/practice/research-opt-in` | staff `practice.settings` |
| `POST` | `/api/v1/internal/research-etl/run` | secret `X-Research-Etl-Secret` |

## RGPD

- Finalité **distincte** de la continuité de soins (observatoire / santé publique vétérinaire). Base légale à valider juridiquement (intérêt légitime + contrat opt-in cabinet recommandé).
- Minimisation : pas de nom, email, microchip, rue, UUID pet/user dans les réponses API.
- Data room API : **pas de `payload`** warehouse (reste en base pour ETL) — champs exposés = semaine / CP / pays / espèce / age_band / signal.
- Opt-out : stop ingest + purge `anon_events` par `practice_id_hash` ; rebuild agrégats.
- `GET /me/export` : **ne pas** inclure les agrégats research (pas de donnée personnelle).
- Voir aussi [36-RGPD.md](36-RGPD.md).

## Limites V1 (honnêteté produit)

- Pas de diagnostic / pathogène / maladie à déclaration.
- Pas de température / symptômes respiratoires structurés.
- Labs V1 sans sérologie/PCR.
- Biais géo « cabinet » (pas domicile animal).
- Densité limitée aux cabinets opt-in petsFollow.

## V2 (livré — tag `dev`)

- `research.groups` + `research.group_members` (owner/member) + `dataroom_enabled` (défaut off).
- Data room : `GET /research/dataroom/events` — sans `practice_id_hash` / `source_hash` / `city` / `payload` ; grain éligible si **COUNT(DISTINCT practice_id_hash) ≥ 5** ; membership d’un groupe **dataroom_enabled** (pas de solo-groupe).
- **Scope réseau** : le groupe est une **porte d’accès** (privilege), pas un filtre de données — un membre autorisé voit tous les micro-événements k-anon du réseau opt-in, pas seulement ceux « de son groupe ».
- Seed : groupe « Réseau démo BE » + flag Data room on + 5 hashes cabinet synthétiques (TRUNCATE au re-seed).

### Encore hors scope

- Signaux amont (fièvre, respiratoire, diagnostic codé).
- Flutter lecture seule optionnelle.
- Use cases commerciaux (`useCase/`) **uniquement à la GA** (règle modules tag `dev`).

## Surfaces UI

| Route | Rôle |
|-------|------|
| `/research` · `/heatmap` · `/timeseries` · `/alerts` · `/settings` | `research` |
| `/research/groups` · `/research/dataroom` | `research` (V2) |
| `/admin/research` | `admin` / `dev` — liste cabinets opt-in |
| Settings cabinet (toggle opt-in) | staff `practice.settings` |

Semaines ISO : borne lundi **Europe/Brussels**. Agrégats ETL : refresh **incrémental** des semaines touchées ; rebuild global à l’opt-out.

## Démo locale

```bash
make up-infra && make migrate && make seed   # opt-in VetPlus + ETL seed
make api-dev          # RESEARCH_ENABLED=true
make nuxtjs-dev       # NUXT_PUBLIC_RESEARCH_ENABLED=true
# Login research.demo@petsfollow.test / ResearchDemo123!
# ou switch profil research depuis vet.demo / admin.demo
curl -X POST http://localhost:8291/api/v1/internal/research-etl/run \
  -H "X-Research-Etl-Secret: $RESEARCH_ETL_SECRET"
```

## Tests

- Go : `TestResearch*` + `TestResearchGroupsAndDataRoom` (groups, gate Data room, k-anon sans PII).
- Playwright : `@p0` `22-research` — overview + timeseries + heatmap + groups + dataroom + admin opt-ins + toggle Data room.
- Plan : [15-PLAN-TESTS.md](15-PLAN-TESTS.md).
