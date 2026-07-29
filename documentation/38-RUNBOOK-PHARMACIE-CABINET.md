# Runbook cabinet — pharmacie (stock / DAF / inventaire)

Guide opérationnel pour un cabinet pilote staging (module tag `dev`).

## Prérequis

| Élément | Valeur |
|---------|--------|
| Flag API | `PHARMACY_ENABLED=true` (staging Cloud Run par défaut) |
| Flag Nuxt | `NUXT_PUBLIC_PHARMACY_ENABLED=true` |
| Extension PG | `pg_trgm` (requis autocomplete CNK) |
| VAMReg | `VAMREG_DRY_RUN=true` (défaut staging) — pas d’envoi live |
| Workers Asynq | `PHARMACY_WORKERS_ENABLED=false` par défaut ; activer seulement si Redis + file OK |
| Facturation DAF→Billit | **gelée** jusqu’accès reseller (`P0-2`) |

Secrets optionnels : `PHARMACY_EXPIRY_SECRET` (job péremption), `VAMREG_API_KEY` (Phase 4.A live).

## Parcours quotidien

1. **Réception** — `/stock` : lot + DLC + qty ; optionnel n° BL / fournisseur.
2. **Réassort** — seuils → alertes → commande e-mail CSV fournisseur.
3. **Inventaire** — démarrer session → compter toutes les lignes → clôturer (écarts → adjust). Export CSV registre.
4. **DAF** — `/daf/nouveau` : lignes + AMM ; antibiotique → espèce / indication / durée ; finaliser → PDF + VAMReg dry-run.
5. **Chaîne alimentaire / domicile** — `PATCH …/pets/{id}/food-chain` (perm `pets.write_clinical`, **hors** gate pharmacy) accepte `{ foodChainStatus?, domicileLocation? }` en **un seul UPDATE**. Animal de rente : temps d’attente V/L/O requis sinon finalize bloqué. Espèces UI domicile/statut : cheval, âne, bovin, ovin, caprin, porcin, volaille, lapin, alpaga, lama — fiche Pro, wizard DAF et PDF. Création client : rente (bovin/ovin/caprin/porcin/volaille/alpaga/lama) → défaut `food_producing`.
6. **Temps d’attente** — `PATCH …/medications/{id}/withdrawal` (viande / lait / œufs jours) ; snapshot sur lignes DAF + PDF.

## Quarantaine / waste

- Lots périmés : job expiry (Scheduler) → quarantaine auto.
- Sortie définitive : action **waste** uniquement (pas un simple adjust à zéro).

## Hors périmètre pilote

- Import CNK national AFMPS (`2.F` / P0-3)
- VAMReg production (P0-1)
- Stupéfiants registre (P0-4)
- EDI grossistes / Bigame / Vetcompendium (Phase 5)
- Facture Peppol auto depuis DAF (S5 / Phase 3)

## Smoke local

```bash
make up-infra && make migrate && make seed && make api-dev
# autre terminal
make nuxtjs-dev
# tests
cd go && go test ./internal/handlers/ -run 'TestPharmacy' -count=1
```

E2E : `nuxtjs/tests/e2e/specs/17-pharmacy-stock-daf.spec.ts` (`@pharmacy` — P0 trace DAF + P1 inventaire / réassort / VAMReg dry-run).

## Rétention 5 ans

Les tables `pharmacy.*` (mouvements, DAF, audits) **ne sont pas** purgées par le job RGPD 3 ans d’inactivité utilisateurs. Conservation registres typique 5 ans — voir commentaire `internalRunRetentionPurge`.
