# 39 — VAMReg AFMPS (M2M readonly)

**Objectif** : client Go de communication avec l’API **lecture seule** VAMREG (FAMHP) pour les listes de référence — ICD `v20260701`.

| Méta | Valeur |
|------|--------|
| PDF source | [assets/VAMREG_ICD_readonly_v20260701_ENG.pdf](assets/VAMREG_ICD_readonly_v20260701_ENG.pdf) |
| Rôle FAMHP | `VAM_REG_ROLE_SOFTWARE_HOUSE` (pas d’UI VAMREG, pas d’écriture déclaration) |
| Code | `go/internal/pharmacy/vamreg_afmps.go` · types `vamreg_afmps_types.go` |
| Déclarations DAF | `VamregClient` / `VamregDeclarer` — **hors** cet ICD · **dry-run partout** |

## Politique env (une clé software-house)

Une seule clé AFMPS (listes GET) **ne justifie pas** `VAMREG_DRY_RUN=false` : ce flag pilote les **déclarations** DAF (`Bearer` + chemin provisoire), pas le client listes (`FAMHP-SEC-KEY`).

| Flag / env | Rôle | Staging | Main / prod |
|------------|------|---------|-------------|
| `VAMREG_DRY_RUN` | Déclarations DAF | **true** (forcé au deploy) | **true** (forcé au deploy) |
| `VAMREG_AFMPS_API_KEY` | Listes GET | Secret SM monté | Même secret (ou clone) |
| `VAMREG_AFMPS_BASE_URL` | Host listes | `https://app.fagg-afmps.be/vamreg/api` | idem (seul host ICD) |
| `VAMREG_API_KEY` / `VAMREG_BASE_URL` | Write déclarant | Absents tant que dry-run | Absents tant que dry-run |

Passer `VAMREG_DRY_RUN=false` uniquement quand : ICD **write** FAMHP + credentials déclarant + client write réécrit. Jusque-là les listes peuvent être **live GET** sur le host prod AFMPS (pas d’env test public documenté).

Garde-fou boot : `ValidateVamreg` refuse un `VAMREG_BASE_URL` pointant vers `/vamreg/api` (host listes) en mode live declare.

## Contrat HTTP (listes)

| Élément | Valeur |
|---------|--------|
| Host prod | `https://app.fagg-afmps.be` |
| Base path | `/vamreg/api/` |
| Auth | Header `FAMHP-SEC-KEY` (clé générée sur le portail FAMHP Connect) |
| Accept | `application/json` |
| Succès / refus | `200` / `403` |

Clé prod (longueur) : `xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx_…=`

## Endpoints (GET)

Documentés ICD §2.1.1 :

- `/medicinal-product` · `?timestamp=true`
- `/foreign-medicinal-product` · `?timestamp=true`
- `/active-substance`
- `/pharmaceutical-form`
- `/unit`

Listes codées illustrées §2.1.2.4 (chemins déduits du même pattern — à confirmer live) :

- `/unit-out` · `/target-species` · `/indication` · `/provider-type` · `/product-type` · `/usage`

## Usage

Env séparés (ne pas mélanger avec la déclaration) :

| Env | Secret Manager | Rôle |
|-----|----------------|------|
| `VAMREG_AFMPS_API_KEY` | `petsfollow-vamreg-afmps-api-key` | Header `FAMHP-SEC-KEY` |
| `VAMREG_AFMPS_BASE_URL` | (env) défaut `https://app.fagg-afmps.be/vamreg/api` | Base path |
| `VAMREG_API_KEY` / `VAMREG_BASE_URL` | `petsfollow-vamreg-api-key` | Déclaration write (**P0-1**, hors ICD) — ne pas monter tant que dry-run |

```go
c := pharmacy.NewVamregAFMPSClient(cfg.VamregAfmpsBaseURL, cfg.VamregAfmpsAPIKey)
products, err := c.ListMedicinalProducts(ctx)
```

### Sync listes → Postgres (V3.1)

Job interne (dry-run par défaut) :

```bash
curl -X POST "$API/api/v1/internal/pharmacy/vamreg-ref-sync" \
  -H "X-Pharmacy-Vamreg-Ref-Sync-Secret: $PHARMACY_VAMREG_REF_SYNC_SECRET" \
  -H "Content-Type: application/json" \
  -d '{"dryRun":true}'
# apply : {"dryRun":false}
```

Persiste `target_species` / `indication` / `pharmaceutical_form` dans `pharmacy.vamreg_ref_codes`.  
Lecture cabinet : `GET /vet/pharmacy/vamreg-refs?kind=target_species` (perm `pharmacy.read`).

Provisionnement staging / attach Cloud Run :

```bash
./infra/gcp/setup-vamreg-afmps-secret.sh /chemin/vers/cle --attach-run
```

(`--attach-run` force aussi `VAMREG_DRY_RUN=true`.)

## Hors scope de cet ICD

- POST déclarations stock / registre in/out (chemin déclaration live = **P0-1** + doc write séparée)
- Import CNK national catalogue petsFollow (**P0-3**) — ce client peut alimenter une sync ultérieure
- Env AFMPS « test » séparé (URL non documentée / non publique)

## Tests

```bash
cd go && go test ./internal/pharmacy/ -run VamregAFMPS -count=1
cd go && go test ./internal/platform/config/ -run ValidateVamreg -count=1
```
