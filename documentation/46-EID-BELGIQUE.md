# Lecture eID belge — préremplissage client (Pro)

## Objectif

Sur les cabinets **BE**, permettre de préremplir la modale **Nouveau client** et
l’onglet **Identité** depuis :

1. **Web eID** (parcours officiel CSAM / BOSA : lecteur USB + app native + extension
   navigateur + PIN) — chemin privilégié ;
2. un export **eID Viewer** (`.eid` / `.xml`) — chemin alternatif.

Hors scope v1 : IdP OIDC e-Contract, import PDF via IA, photo carte, date de
naissance persistée sur le compte client.

## Trust model

| Chemin | Confiance |
|--------|-----------|
| **Web eID** | Jeton signé, chaîne CA BE embarquée, OCSP (sauf `WEB_EID_DISABLE_OCSP`) |
| **Viewer `.eid`** | Fichier fourni par un pro authentifié (`clients.write`) — **pas** de signature crypto vérifiée côté serveur (`signature_verified=false`). Un XML forgé peut préremplir ; le staff reste responsable de la saisie. |

## Flags

| Variable | Rôle |
|----------|------|
| `EID_ENABLED` | Active les routes API `/vet/eid/*` |
| `NUXT_PUBLIC_EID_ENABLED` | Affiche le bloc UI `ProEidReader` |
| `EID_SITE_ORIGIN` | Origine liée au jeton Web eID (défaut = `PETSFOLLOW_PUBLIC_SITE_URL`) |
| `WEB_EID_DISABLE_OCSP` | Dev only — saute OCSP |

`make api-dev` / `make nuxtjs-dev` posent ces flags à `true` en local.
Staging Cloud Run : `EID_ENABLED` + `NUXT_PUBLIC_EID_ENABLED` on par défaut
(`infra/gcp/lib/deploy-run-args.sh`) ; **prod opt-in** (défaut `false`). Module tag **dev**
(badge `nav.tagDev` sur `ProEidReader`).

### Production (checklist)

Les CA belges sont **embarqués dans le binaire API** (`//go:embed certs/*.crt`) —
pas d’upload GCS / Secret Manager. Au prochain `deploy-gcp-prod` qui inclut le
pack (~430 CA, commit trust store sur `staging` / `main`), elles sont déjà là.

Pour activer le module sur prod (toujours tag `dev` côté UI) :

1. Merge / promote le commit trust store jusqu’à la branche déployée en prod.
2. Deploy : `workflow_dispatch` [`.github/workflows/deploy-gcp-prod.yml`](../.github/workflows/deploy-gcp-prod.yml)
   ou `make gcp-deploy-prod`, avec opt-in explicite :
   - `EID_ENABLED=true`
   - `NUXT_PUBLIC_EID_ENABLED=true`
   - `EID_SITE_ORIGIN=https://petsfollow.app` (ou laisser le défaut =
     `PETSFOLLOW_PUBLIC_SITE_URL`)
3. Prérequis : Redis prod branché (`REDIS_ADDR`) — sinon `503 eid_redis_required`.
   OCSP **live** (ne pas poser `WEB_EID_DISABLE_OCSP` en prod — fail-fast au boot).
4. Post-deploy : `make smoke-eid` contre l’API prod **uniquement** si un compte
   véto BE de test existe (le smoke écrit un `eid_readings`) ; sinon QA terrain
   Web eID sur https://petsfollow.app + carte réelle.

Ne pas activer en prod tant que le parcours n’est pas validé en staging (PIN réel).


L’API refuse aussi les cabinets dont `practice.country_code ≠ BE`
(`eid_not_available`).

Export RGPD (`GET /me/export`) : agrégat `eidReadings` (audit hashé du véto).

## Endpoints

| Méthode | Chemin | Description |
|---------|--------|-------------|
| `POST` | `/api/v1/vet/eid/import` | Multipart `file` → identité JSON |
| `GET` | `/api/v1/vet/eid/web-eid/challenge` | Nonce (Redis obligatoire hors local ; mémoire en local/dev/test, TTL 5 min) |
| `POST` | `/api/v1/vet/eid/web-eid/verify` | `{ "token": … }` → identité JSON (préremplissage) |

Réponse import/verify : champs utiles au formulaire seulement (noms, NISS, adresse,
pays, outil) — **pas** de photo / date de naissance / genre / n° de carte.

Staging/prod sans Redis → `503 eid_redis_required` (fail-closed multi-instances Cloud Run).
`WEB_EID_DISABLE_OCSP` est refusé hors local/dev/test (fail-fast au boot). Une panne
OCSP/réseau au verify → `503 eid_token_infra` (challenge restauré) ; erreur Redis au
take nonce → `503 eid_nonce_store_failed` (≠ expired).

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

- **Web eID (prioritaire)** : lecteur USB + application native officielle + extension
  navigateur Web eID (écosystème CSAM / BOSA, install via https://eid.belgium.be/) + PIN.
  Lib JS : `@web-eid/web-eid-library` (pin commit SHA GitHub `web-eid/web-eid.js`).
  Communication via **native messaging** (hors CSP HTTP) — pas d’élargissement
  `connect-src` requis pour l’extension.
  En local, `localhost` et `127.0.0.1` sont des origines distinctes pour Web eID : le
  challenge reprend l’origine de l’onglet (via BFF `X-PF-Web-Eid-Origin`) tant que
  c’est un alias loopback de `EID_SITE_ORIGIN`. Les GET same-origin n’envoient
  souvent **pas** le header `Origin` : la BFF dérive alors `getRequestURL` (Host)
  puis `Referer` (`resolveWebEidOrigin` / `webEidUpstreamHeaders`). Le header n’est
  honoré côté Go que si `X-PF-Proxy-Secret` matche `BFF_PROXY_SECRET` (même gate
  que `X-PF-Client-IP`, forcé avec l’origine même sans IP client) — un appel API
  direct ne peut pas spoofe l’origine du challenge.
  **Limitation locale** : l’extension (surtout Firefox) n’injecte souvent **pas**
  de content script sur `http://localhost` / `127.0.0.1` alors que le même poste
  fonctionne en HTTPS staging — UI `loopbackUnavailable` + repli Viewer. Tester
  Web eID en local : même origine que `EID_SITE_ORIGIN`, extension autorisée pour
  le site, ou utiliser staging.
- **Viewer** : eID Viewer / BEid → export `.eid` → upload.
  Pas de détection d’extensions Chrome tierces (beID Connect, etc.) côté Pro.

## Tests

Flux complet (auto) :

```bash
cd go && go test ./internal/eid/ -count=1
cd go && go test ./internal/handlers/ -run 'TestEid' -count=1
cd nuxtjs && npm test -- tests/unit/eid-prefill.spec.ts tests/unit/use-eid-prefill.spec.ts
# Playwright @p0 — mock Viewer + Web eID (pas de PIN) :
cd nuxtjs && npx playwright test tests/e2e/specs/03-clients.spec.ts --grep 'eID'
make smoke-eid            # local :8291 (Viewer + challenge + verify bad token + gate BFF)
make smoke-eid-staging
```

Réf. checklist : `documentation/15-PLAN-TESTS.md` (C2.3 / C2.6 + section Z eID).

**QA terrain** : Web eID avec PIN réel nécessite lecteur + extension sur un poste BE
(non automatisable). Le smoke vérifie import Viewer + binding `EID_SITE_ORIGIN` +
contrat verify (422 / nonce one-shot). E2E utilise `__PF_WEB_EID_MOCK__` pour
simuler la lib sans extension.

**PIN** : le parcours Web eID **exige** le PIN d’authentification de la carte
(signature du challenge). Il n’existe pas d’API Web eID « lecture identité sans PIN ».
Sans PIN : exporter une fiche via eID Viewer (`.eid`) puis « Importer fiche eID Viewer ».

UI lecture carte : loader + message d’attente pendant `authenticate` (dialogue PIN) ;
timeout / annulation / extension absente / verify 422 → message d’erreur visible
(`eid-reader-error`). Trust store CA : `go/internal/eid/certs/` (~430 CA uniques,
PEM ; lots historiques expirés **conservés**) — Citizen (`citizen*`, `eidc*`) **et**
Foreigner (`foreigner*`, `eidf*`) + racines (`belgiumrca*`, dont **Root CA6**).
Re-sync : `make sync-eid-certs` (miroirs BOSA en HTTP — réseau de confiance).
Une carte dont le lot CA manque → `eid_cert_untrusted`. Les erreurs go-web-eid sont
classées via `exceptions.Error.Code` (`ClassifyWebEidVerifyError`).