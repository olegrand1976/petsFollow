# Organisation MLM — préparation profils petsFollow

> Statut : **préparation** (flag `MLM_ORG_ENABLED`, défaut `false`).  
> Les commissions restent **flat** (grille commercial + SPIFF). Aucun override upline / residual / spillover.

Voir aussi : [06-FLUX-UTILISATEURS.md](06-FLUX-UTILISATEURS.md) · [19-FICHE-COMMISSION-COMMERCIAL.md](19-FICHE-COMMISSION-COMMERCIAL.md).

## Matrice profils

| Profil petsFollow | Rôle org MLM cible | Aujourd’hui | Réservé (flag / stub) |
|-------------------|--------------------|-------------|------------------------|
| `commercial` | Distributeur / leaf | CRM, encode, commissions flat, pitch, réseau (lecture) | Downline N niveaux |
| `commercial_manager` | Chef de branche / upline L1 | KPI équipe, suivi, reassign, quotas, leaderboard | Scope branche multi-niveaux |
| `admin` | Ops plateforme | CRUD users, branches, paie ledger, SPIFF | Sync `external_mlm_id`, audit généalogie |
| *(pas de 4ᵉ rôle)* | Ranks MLM | `sales_rank` int (affichage) | Grades dynamiques sans nouveau rôle Nuxt |

## Modèle données (migration `000068`)

- `sales.branches` — `name`, `code`, `external_mlm_id` (mapping org externe)
- `identity.users.branch_id` — rattachement commercial / manager
- `identity.users.sponsor_user_id` — upline MLM ; **sync** avec `manager_user_id` tant que profondeur 1
- `identity.users.sales_rank` — défaut `0`
- `sales.commercial_quotas` — objectifs mensuels coaching

## API

| Endpoint | Rôle |
|----------|------|
| `GET /commercial/network` | Branche + sponsor + downline (`maxDepth=1` si flag off) |
| `GET /admin/sales-branches` | Liste branches + `pendingAuto` (éligibles auto-création) |
| `POST /admin/sales-branches` | Création manuelle (secours) |
| `POST /admin/sales-branches/auto-run` | Lance immédiatement l’auto-création (+ emails) |
| `POST /internal/sales-branches-auto/run` | Job bi-quotidien 10h/18h (`X-Sales-Branches-Auto-Secret`) |
| `PATCH /admin/commercials/{id}/branch` | Assign branche |
| `GET /commercial-manager/leaderboard` | Classement € équipe |
| `GET/PUT /commercial-manager/quotas` | Objectifs |
| `PATCH /commercial-manager/prospects/{id}/reassign` | Réassignation |

### Auto-création de branche

Éligible : `role=commercial`, `branch_id` vide, sponsor **absent** ou **non** `commercial` (peer).

Dérivation depuis `full_name` (`Prénom … Nom`) :
- **name** : `Dupont D`
- **code** : `DUPONTD` (sans accents, collision → `DUPONTD2`, …)

Le commercial reçoit un email de félicitations. Réponse job/admin : `{ created, skipped, items, skippedItems }`. Scheduler GCP : `make gcp-sales-branches-scheduler` (`0 10,18 * * *` Europe/Brussels). Create+assign sont atomiques (transaction).

Helper store : `ListDownline(userID, maxDepth)` — prêt pour N quand `MLM_ORG_ENABLED=true`.

## UI

- Commercial / manager : section nav **Réseau** → `/commercial/network`
- Admin : `/admin/sales-branches` (liste + pending + run now + création manuelle secours) ; colonnes branche / manager sur commerciaux
- Seed démo : branches **Bruxelles** (`BRU`) et **Nord** (`NORD`)

## Hors scope (chantier MLM ultérieur)

Override %, residual, spillover, self-recruit, payout multi-bénéficiaires, sync org externe automatisée.
