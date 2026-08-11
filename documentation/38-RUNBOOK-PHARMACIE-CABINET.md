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

1. **Réception** — `/stock` : lot + DLC + qty ; dépôt (multi-sites) ; optionnel n° BL / fournisseur.
2. **Dépôts** — carte `/stock` : créer un site/armoire (code + défaut).
3. **Prix & seuils** — carte `/stock` : prix achat/vente/TVA + seuil réassort par CNK.
4. **Réglages péremption** — auto-quarantaine, digest hebdo, notifs.
5. **Réassort** — seuils → alertes → commande e-mail CSV fournisseur.
6. **Inventaire** — démarrer session → compter toutes les lignes → clôturer (écarts → adjust). Export CSV registre. Adjust manuel aussi disponible sur un lot actif.
7. **Journal** — carte mouvements (receipt / DAF / adjust / waste / quarantaine).
8. **Fiche médicament** — `/medicaments` : search CNK → détail, temps d’attente cabinet, prix, lots actifs.
9. **DAF** — `/daf/nouveau` : lignes + AMM ; antibiotique → espèce / indication / durée / **posologie** ; finaliser → PDF + VAMReg dry-run.
10. **Chaîne alimentaire / domicile** — `PATCH …/pets/{id}/food-chain` (perm `pets.write_clinical`, **hors** gate pharmacy) accepte `{ foodChainStatus?, domicileLocation? }` en **un seul UPDATE**. Animal de rente : temps d’attente V/L/O requis sinon finalize bloqué. Espèces UI domicile/statut : cheval, âne, bovin, ovin, caprin, porcin, volaille, lapin, alpaga, lama — fiche Pro, wizard DAF et PDF. Création client : rente (bovin/ovin/caprin/porcin/volaille/alpaga/lama) → défaut `food_producing`.
11. **Temps d’attente** — `PATCH …/medications/{id}/withdrawal` (viande / lait / œufs jours) ; snapshot sur lignes DAF + PDF ; aussi éditable sur fiche `/medicaments`.

## Quarantaine / waste

- Lots périmés : job expiry (Scheduler) → quarantaine auto.
- Sortie définitive : action **waste** avec motif (`expired` / `supplier_return` / `destruction`) — pas un simple adjust à zéro.

## Hors périmètre pilote

- Import CNK national AFMPS pack officiel (`2.F` / P0-3) — la **procédure 3 gates** est opérationnelle (voir ci-dessous) ; le fichier national attend la licence
- VAMReg production (P0-1)
- Stupéfiants registre (P0-4)
- EDI grossistes / Bigame / Vetcompendium (Phase 5)
- Facture Peppol auto depuis DAF (S5 / Phase 3)

## Import AFMPS CSV — triple contrôle (ops)

Aucun upsert silencieux dans `pharmacy.ref_medications`. Surfaces : admin Pro `/admin/afmps-imports` (flag `PHARMACY_ENABLED`) · CLI `import-cnk` · e2e `23-afmps-admin` · détail technique [27 § AFMPS](27-PHARMACIE-BELGIQUE.md).

| Gate | Action | Statut |
|------|--------|--------|
| **1 — Validation** | Upload CSV → parse + KPIs (skip CNK vides, dédup, erreurs dures ≤ 5 %, collisions) | `validated` ou `blocked` |
| **2 — Revue** | UI : filtres collision + exclude lignes + checkbox accusé → **Marquer revu** (CLI : `--mark-reviewed`) | `reviewed` |
| **3 — Commit** | Phrase `IMPORT AFMPS` ou `IMPORT_AFMPS` → upsert transactionnel + fusion `afmps_meta` | `completed` |

### UI admin (staging / local)

1. Admin → **Import AFMPS** → *Importer un CSV* → choisir le fichier → *Valider (gate 1)*.
2. Sur la fiche job : vérifier KPIs (ready / insert / update / deactivate preview) et le preview lignes.
3. Cocher l’accusé de revue → *Marquer revu*.
4. Saisir `IMPORT_AFMPS` → *Commit*.  
   **Ne cochez `deactivate-missing` qu’après lecture du compteur** : soft-disable tous les CNK actifs absents du fichier (destructif sur un catalogue partiel / mini-CSV).

Fixture e2e / smoke local : [`nuxtjs/tests/e2e/fixtures/afmps-mini.csv`](../nuxtjs/tests/e2e/fixtures/afmps-mini.csv) (2 lignes vétérinaires).

### CLI

```bash
# Gate 1 — validation + staging (aucune écriture dictionnaire)
go run ./cmd/petsfollow-api import-cnk \
  --file=/chemin/export-afmps-conditionnement.csv \
  --validate

# Gate 2
go run ./cmd/petsfollow-api import-cnk --job=<UUID> --mark-reviewed

# Gate 3 — commit (phrase obligatoire)
go run ./cmd/petsfollow-api import-cnk \
  --job=<UUID> --commit --confirm=IMPORT_AFMPS

# Opt-in seulement si le compteur deactivatePreview a été lu et accepté :
#   --deactivate-missing
```

`--dry-run` = alias de `--validate`. Pas d’upsert direct `--file` sans job.
## Smoke local

```bash
make up-infra && make migrate && make seed && make api-dev
# autre terminal
make nuxtjs-dev
# tests
cd go && go test ./internal/handlers/ -run 'TestPharmacy|TestAFMPS' -count=1
```

E2E pharmacie : `17-pharmacy-stock-daf.spec.ts` · AFMPS admin : `23-afmps-admin.spec.ts` (`@pharmacy`).

## Rétention 5 ans

Les tables `pharmacy.*` (mouvements, DAF, audits) **ne sont pas** purgées par le job RGPD 3 ans d’inactivité utilisateurs. Conservation registres typique 5 ans — voir commentaire `internalRunRetentionPurge`.
