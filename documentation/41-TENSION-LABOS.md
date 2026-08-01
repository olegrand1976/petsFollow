# Tension artérielle & prises de sang

## Principe

Deux couches distinctes dans le dossier animal :

| Donnée | Qui écrit | Modèle | Timeline |
|--------|-----------|--------|----------|
| **Tension** | Client owner (premium) + practice staff (`pets.write_clinical`) + admin | `pets.blood_pressure_readings` | `blood_pressure` |
| **Prise de sang** | Practice staff (`pets.write_clinical`) + admin — **pas** `care_pro` | `labs.panels` + `labs.panel_results` | `lab_panel` |

Pas de workflow de validation type FC. Alerte seuil tension = hors V1.

## Tension

Migration `000127`. Champs : `systolic_mmhg`, `diastolic_mmhg`, `mean_mmhg` optionnel, `method` (`doppler` \| `oscillometric` \| `invasive` \| `unknown`), `site`, `comment`, `recorded_at`.

### API

- `POST /api/v1/pets/{id}/blood-pressure` — client owner (premium) ou staff `pets.write_clinical` (pas care_pro)
- `GET /api/v1/pets/{id}/blood-pressure`

### UI

- **Pro** : KPI dernière tension, chart SYS/DIA (onglet vitals), tableau + formulaire (SYS/DIA, méthode, site, commentaire)
- **Flutter** : sheet saisie (mêmes champs) + chart dual + entrée timeline

## Prises de sang (panels)

Migration `000128` (schéma `labs`). Panel = métadonnées + lignes analytes du **catalogue V1** + `document_id` optionnel (`pets.documents`).

Catalogue : `crea`, `urea`, `bun`, `alat`/`alt`, `asat`/`ast`, `alp`, `ggt`, `hct`, `wbc`, `plt`, `glucose`, `tp`/`protein`.

Flags : `low` / `normal` / `high` / `unknown` (calculés vs `ref_low` / `ref_high`).

### API

- `POST/GET /api/v1/pets/{id}/lab-panels`
- `GET/PATCH/DELETE /api/v1/pets/{id}/lab-panels/{panelId}` — PATCH exige `results` non vide (`{}` → 400 `results_required`)
- `GET /api/v1/pets/{id}/lab-analytes/{analyteCode}/trend`

### UI

- **Pro** : section Labos sous vitals (création / édition résultats avec `valueNum` **ou** `valueText`, tendance analyte, suppression, lien document)
- **Flutter** : lecture seule (`LabPanelsScreen`) depuis la fiche ou un événement timeline `lab_panel` (`initialPanelId`) ; ouverture PDF si `documentId`

## RGPD

Export `GET /me/export` : agrégats `bloodPressureReadings`, `labPanels`, `labPanelResults`. Purge client via `DELETE pets` (CASCADE).

## Seed démo

Sur **Rex** (`client.demo`) : 3 tensions + 2 panels BioVet (créat. hors norme). Relancer `make seed`.

## Tests

- Go : `TestBloodPressureCreateListClientAndVet`, `TestLabPanelCRUDAndTimeline`
- Flutter : `pet_quick_actions_test` (sheet tension) · `lab_panels_screen_test`
- E2E : `09-pet-detail.spec.ts` — `pet detail — tension + panel labo @p1` (dont édition résultats)
- Smoke API : `make smoke` — POST/GET tension + panel (créat. + `valueText`) + trend
- Relancer : `make test-go` · `cd flutter && flutter test test/features/pets/pet_quick_actions_test.dart test/features/pets/lab_panels_screen_test.dart` · tag Playwright `@p1` · `make smoke`
