# Prescriptions vétérinaires (module tag `dev`)

> **Statut** : V1 brouillon + preview PDF — **tag `dev`** (badge `nav.tagDev`), **pas** GA, **pas** signature électronique, **pas** partage client.
> Flag : `PRESCRIPTIONS_ENABLED` / `NUXT_PUBLIC_PRESCRIPTIONS_ENABLED` (on via `make api-dev`).

### Surfaces tag `dev` (obligatoires tant que non GA)

| Surface | Attendu |
|---------|---------|
| Nav Pro `/prescriptions` | `tag: t('nav.tagDev')` + testid `nav-prescriptions-dev-tag` |
| Pages liste / nouveau / détail | `ProBadge` `nav.tagDev` (`prescriptions-*-dev-badge`) |
| API | `PRESCRIPTIONS_ENABLED=false` → 404 `prescriptions_disabled` |
| UI flag | `NUXT_PUBLIC_PRESCRIPTIONS_ENABLED` déclaré (miroir) ; **nav toujours visible** — gate API only, comme pharmacie |
| Commercial | pas d’entrée `useCase/` tant que tag `dev` |

## Périmètre V1

| Inclus | Exclus (phase 2) |
|--------|------------------|
| CRUD brouillons (`status=draft`) | Signature / pad / eIDAS |
| Preview PDF A4/A5 à la volée | PDF immuable + `pdf_sha256` stocké |
| `country_code` depuis le cabinet | `date_issued` renseigné (NULL tant que draft) |
| UI Pro `/prescriptions` + badge `nav.tagDev` | Templates légaux certifiés BE/FR/IT/ES |
| Export RGPD `owner_id` | Attache `pets.documents` |
| Soft-link CNK / `ref_medication_id` (UI picker) | Push messagerie |
| Pont optionnel `POST …/daf/from-prescription` → draft DAF | Finalize silencieux / VAMReg depuis prescription |

## Différence avec le DAF pharmacie

| | **Prescription** | **DAF** ([27](27-PHARMACIE-BELGIQUE.md)) |
|--|----------------|------------------------------------------|
| But | Document propriétaire / officine | Administration / fourniture cabinet BE |
| Schéma | `prescriptions.prescriptions` | `pharmacy.daf_*` |
| Stock | Non | FEFO + mouvements |
| Flag | `PRESCRIPTIONS_ENABLED` | `PHARMACY_ENABLED` |

Soft-link optionnel dans le JSONB médicaments : `cnk`, `ref_medication_id` (pas de FK).

## API

Base : `/api/v1/vet/prescriptions` (BFF `/api/vet/prescriptions`).

| Méthode | Route | Rôle |
|---------|-------|------|
| `GET` | `/` | Liste (`status`, `petId`) |
| `POST` | `/` | Crée un draft |
| `GET/PATCH/DELETE` | `/{id}` | Détail / maj / delete draft |
| `GET` | `/{id}/pdf` | Stream PDF (généré, non persisté) |

Pont pharmacie (si `PHARMACY_ENABLED`) : `POST /api/v1/vet/pharmacy/daf/from-prescription` `{ prescriptionId }` → draft DAF (lignes avec `ref_medication_id` uniquement).

Consultation → DAF : `GET/PUT /api/v1/vet/pharmacy/daf/for-visit?visitId=` (un draft max par visite).

404 `prescriptions_disabled` si flag off. Transitions `signed` / `sent` / `archived` refusées en V1.

## Schéma

Migration `000091_prescriptions` — table `prescriptions.prescriptions` :
`practice_id`, `veterinary_id`, `pet_id` (CASCADE), `owner_id`, `country_code`,
`date_issued` (**NULL en V1** — renseigné à la signature / finalize phase 2), `valid_until`,
`signature_id` (réservé), `pdf_object_key` / `pdf_sha256` (réservés),
`status`, `paper_format`, `medications` JSONB, `notes`.

ACL V1 : création avec `CanAccessPet(write_notes)` ; lecture/écriture ultérieure = staff du `practice_id` (dossier cabinet, comme DAF). Pas de contrainte « seul l’auteur » sur PATCH/DELETE.

Préfixe média PHI réservé : `prescriptions/` (`IsSensitiveObjectKey`).

## Tests

```bash
cd go && go test ./internal/handlers/ -run 'TestPrescriptions|TestPharmacyDAFUpsertForVisit|TestPharmacyDAFFromPrescription' -count=1
```

Checklist : [15-PLAN-TESTS.md](15-PLAN-TESTS.md) (section Prescriptions).

## Phase 2 (non implémentée)

1. Finalize + attestation (hash) ou pad
2. Upload PDF `prescriptions/{practice}/{id}.pdf` + immutabilité
3. Copie / lien vers `pets.documents`
4. Message texte optionnel (pas de PDF chat)
5. Templates pays renforcés
