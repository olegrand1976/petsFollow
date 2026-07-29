# 39 — VAMReg AFMPS (M2M readonly)

**Objectif** : client Go de communication avec l’API **lecture seule** VAMREG (FAMHP) pour les listes de référence — ICD `v20260701`.

| Méta | Valeur |
|------|--------|
| PDF source | [assets/VAMREG_ICD_readonly_v20260701_ENG.pdf](assets/VAMREG_ICD_readonly_v20260701_ENG.pdf) |
| Rôle FAMHP | `VAM_REG_ROLE_SOFTWARE_HOUSE` (pas d’UI VAMREG, pas d’écriture déclaration) |
| Code | `go/internal/pharmacy/vamreg_afmps.go` · types `vamreg_afmps_types.go` |
| Déclarations DAF | Toujours `VamregClient` / `VamregDeclarer` (dry-run défaut) — **hors** cet ICD |

## Contrat HTTP

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

```go
c := pharmacy.NewVamregAFMPSClient(cfg.VamregBaseURL, cfg.VamregAPIKey)
// BaseURL vide → https://app.fagg-afmps.be/vamreg/api
products, err := c.ListMedicinalProducts(ctx)
timed, err := c.ListMedicinalProductsWithTimestamp(ctx)
species, err := c.ListTargetSpecies(ctx)
```

Env existants : `VAMREG_BASE_URL`, `VAMREG_API_KEY` (Secret Manager `petsfollow-vamreg-api-key`).

## Hors scope de cet ICD

- POST déclarations stock / registre in/out (chemin déclaration live = **P0-1** + doc write séparée)
- Import CNK national catalogue petsFollow (**P0-3**) — ce client peut alimenter une sync ultérieure

## Tests

```bash
cd go && go test ./internal/pharmacy/ -run VamregAFMPS -count=1
```
