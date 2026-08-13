# Lecture eID belge — préremplissage client (Pro)

## Objectif

Sur les cabinets **BE**, permettre de préremplir la modale **Nouveau client** et
l’onglet **Identité** depuis :

1. un export **eID Viewer** (`.eid` / `.xml`) — chemin par défaut ;
2. **Web eID** (lecteur USB + extension + PIN) — si le poste est équipé.

Hors scope v1 : IdP OIDC e-Contract, import PDF via IA, photo carte, date de
naissance persistée sur le compte client.

## Flags

| Variable | Rôle |
|----------|------|
| `EID_ENABLED` | Active les routes API `/vet/eid/*` |
| `NUXT_PUBLIC_EID_ENABLED` | Affiche le bloc UI `ProEidReader` |
| `EID_SITE_ORIGIN` | Origine liée au jeton Web eID (défaut = `PETSFOLLOW_PUBLIC_SITE_URL`) |
| `WEB_EID_DISABLE_OCSP` | Dev only — saute OCSP |

`make api-dev` / `make nuxtjs-dev` posent ces flags à `true` en local.
Staging Cloud Run : `EID_ENABLED` + `NUXT_PUBLIC_EID_ENABLED` on par défaut
(`infra/gcp/lib/deploy-run-args.sh`) ; prod opt-in. Module tag **dev**
(badge `nav.tagDev` sur `ProEidReader`).

L’API refuse aussi les cabinets dont `practice.country_code ≠ BE`
(`eid_not_available`).

Export RGPD (`GET /me/export`) : agrégat `eidReadings` (audit hashé du véto).

## Endpoints

| Méthode | Chemin | Description |
|---------|--------|-------------|
| `POST` | `/api/v1/vet/eid/import` | Multipart `file` → identité JSON |
| `GET` | `/api/v1/vet/eid/web-eid/challenge` | Nonce (Redis obligatoire hors local ; mémoire en local/dev/test, TTL 5 min) |
| `POST` | `/api/v1/vet/eid/web-eid/verify` | `{ "token": … }` → identité JSON (sans photo) |

Réponse import/verify : identité publique uniquement — **pas** de `photo_jpeg_base64`.

Staging/prod sans Redis → `503 eid_redis_required` (fail-closed multi-instances Cloud Run).

BFF Nuxt : `/api/vet/eid/import`, `/api/vet/eid/web-eid/challenge`, `/api/vet/eid/web-eid/verify`.

Permission : `clients.write`.

## Mapping formulaire

| eID | Champ petsFollow |
|-----|------------------|
| `firstname` / `lastname` | `firstName` / `lastName` |
| `niss` | `nationalRegistryNumber` |
| rue (+ n°) | `address` + `billingStreet` |
| CP / ville | `billingPostal` / `billingCity` |
| pays | `billingCountry=BE` |
| — | `billingCustomerKind=individual` si vide |

E-mail et téléphone restent manuels. Web eID ne fournit **pas** l’adresse.

## Audit

Table `practice.eid_readings` : outil, succès, champs lus, **hash** NISS
(`SHA-256(niss:JWT_SIGNING_KEY)`), pas de NISS brut.

Rétention : purge automatique **1 an** via `POST /internal/retention/run`
(`PurgeOldEidReadings`).

## Prérequis poste

- **Viewer** : eID Viewer / BEid → export `.eid` → upload.
- **Web eID** : lecteur + app native + extension navigateur + PIN.
  Lib JS : `@web-eid/web-eid-library` (GitHub `web-eid/web-eid.js`).

## Tests

```bash
cd go && go test ./internal/eid/ -count=1
cd go && go test ./internal/handlers/ -run 'TestEid' -count=1
# Playwright : scénario prefill mock dans 03-clients.spec.ts
```

Réf. checklist : `documentation/15-PLAN-TESTS.md` (C2.3 / C2.6).
