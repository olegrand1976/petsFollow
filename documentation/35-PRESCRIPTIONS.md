# Consignes client (module tag `dev`)

> **Statut** : V1 brouillon + preview fiche PDF — **tag `dev`** (badge `nav.tagDev`), **pas** GA, **pas** partage client structuré.
> Flag : `PRESCRIPTIONS_ENABLED` / `NUXT_PUBLIC_PRESCRIPTIONS_ENABLED` (on via `make api-dev`).
> **Code technique** : routes `/prescriptions`, schéma `prescriptions.*`, package Go `prescription` (noms historiques inchangés).
> **Libellé produit UI** : **Consignes**.

### Frontière — ordonnance légale

L’**ordonnance légale** reste sur **papier carbone**, remplissage **manuel obligatoire**. petsFollow **ne gère pas** l’ordonnance (pas de signature eIDAS, pas de templates pays certifiés, pas de finalize légal). Ce module produit une **fiche consignes** (soins, médicaments à donner, posologie, conseils) destinée à la continuité / transmission au propriétaire.

### Surfaces tag `dev` (obligatoires tant que non GA)

| Surface | Attendu |
|---------|---------|
| Nav Pro `/prescriptions` | Label **Consignes** + `tag: t('nav.tagDev')` + testid `nav-prescriptions-dev-tag` |
| Pages liste / nouveau / détail | `ProBadge` `nav.tagDev` (`prescriptions-*-dev-badge`) |
| API | `PRESCRIPTIONS_ENABLED=false` → 404 `prescriptions_disabled` |
| UI flag | `NUXT_PUBLIC_PRESCRIPTIONS_ENABLED` (gate nav) |
| Commercial | pas d’entrée `useCase/` tant que tag `dev` |

## Périmètre V1

| Inclus | Exclus |
|--------|--------|
| CRUD brouillons (`status=draft`), **sans médication obligatoire** | Ordonnance légale / signature / eIDAS |
| Preview PDF A4/A5 (fiche consignes) | PDF immuable + `pdf_sha256` stocké |
| `care_advice` (conseils / soins client) | Partage client / messagerie |
| `visit_id` optionnel (lien consultation) | Finalize / sent / archived métier |
| Pré-remplissage IA depuis le CR (`suggest-from-visit`) | Inventaire structuré de gestes hors texte |
| Soft-link CNK / `ref_medication_id` (UI picker) | |
| Liste filtrable (statut, recherche, dates) + actions (ouvrir, PDF, fiche animal, supprimer brouillon) | |
| Pont optionnel `POST …/daf/from-prescription` → draft DAF | |
| Export RGPD `owner_id` | |

## Différence avec le DAF pharmacie

| | **Consignes** | **DAF** ([27](27-PHARMACIE-BELGIQUE.md)) |
|--|----------------|------------------------------------------|
| But | Fiche propriétaire (soins / médication / conseils) | Administration / fourniture cabinet BE |
| Schéma | `prescriptions.prescriptions` | `pharmacy.daf_*` |
| Stock | Non | FEFO + mouvements |
| Flag | `PRESCRIPTIONS_ENABLED` | `PHARMACY_ENABLED` |

Soft-link optionnel dans le JSONB médicaments : `cnk`, `ref_medication_id` (pas de FK).

## API

Base : `/api/v1/vet/prescriptions` (BFF `/api/vet/prescriptions`).

| Méthode | Route | Rôle |
|---------|-------|------|
| `GET` | `/` | Liste (`status`, `petId`, `q`, `from`, `to` RFC3339) |
| `POST` | `/` | Crée un draft (`careAdvice`, `visitId` optionnels ; `medications` peut être `[]`) |
| `POST` | `/suggest-from-visit` | Proposition IA depuis CR (`{ visitId }`) — non persistée |
| `GET/PATCH/DELETE` | `/{id}` | Détail / maj / delete draft |
| `GET` | `/{id}/pdf` | Stream PDF fiche consignes (généré, non persisté) |

Pont pharmacie (si `PHARMACY_ENABLED`) : `POST /api/v1/vet/pharmacy/daf/from-prescription` `{ prescriptionId }` → draft DAF (lignes avec `ref_medication_id` uniquement).

404 `prescriptions_disabled` si flag off. Transitions `signed` / `sent` / `archived` refusées en V1.

`suggest-from-visit` : ACL pet + visite cabinet ; rate-limit `vetSuggestRL` (clé `consignes-suggest:`) ; CR tronqué (12k runes) ; CR vide → 400 `no_visit_report` ; Gemini off → 503 `not_configured`. N’invente pas de posologie absente du CR.

## Schéma

Migrations `000091_prescriptions` + `000149_prescription_consignes` — table `prescriptions.prescriptions` :
`practice_id`, `veterinary_id`, `pet_id` (CASCADE), `owner_id`, `visit_id` (NULL, SET NULL), `country_code`,
`date_issued` / `signature_id` / `pdf_*` (colonnes réservées, **non productisées**),
`status`, `paper_format` (format d’impression fiche), `medications` JSONB, `notes`, `care_advice`.

ACL V1 : création avec `CanAccessPet(write_notes)` ; lecture/écriture ultérieure = staff du `practice_id`.

Préfixe média PHI réservé : `prescriptions/` (`IsSensitiveObjectKey`).

## Tests

```bash
cd go && go test ./internal/handlers/ -run 'TestPrescriptions' -count=1
cd go && go test ./internal/platform/gemini/ -run 'Consignes' -count=1
```

Checklist : [15-PLAN-TESTS.md](15-PLAN-TESTS.md) (section Consignes / Prescriptions).
