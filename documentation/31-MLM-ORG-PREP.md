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
| `GET/POST /admin/sales-branches` | CRUD branches |
| `PATCH /admin/commercials/{id}/branch` | Assign branche |
| `GET /commercial-manager/leaderboard` | Classement € équipe |
| `GET/PUT /commercial-manager/quotas` | Objectifs |
| `PATCH /commercial-manager/prospects/{id}/reassign` | Réassignation |

Helper store : `ListDownline(userID, maxDepth)` — prêt pour N quand `MLM_ORG_ENABLED=true`.

## UI

- Commercial / manager : section nav **Réseau** → `/commercial/network`
- Admin : `/admin/sales-branches` + colonnes branche / manager sur commerciaux
- Seed démo : branches **Bruxelles** (`BRU`) et **Nord** (`NORD`)

## Hors scope (chantier MLM ultérieur)

Override %, residual, spillover, self-recruit, payout multi-bénéficiaires, sync org externe automatisée.
