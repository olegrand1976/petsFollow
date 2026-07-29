# UC-AD-01 — Ops admin & support

| | |
|--|--|
| **ID** | `UC-AD-01` |
| **Durée** | ~12 min |
| **Priorité** | Secondaire |
| **Surface** | Web Admin |

## Objectif

Contrôle rapide ops : métriques, utilisateurs, commercials, inbox support.

## Acteurs

| Rôle | Compte | MDP |
|------|--------|-----|
| Admin | `admin.demo@petsfollow.test` | `AdminDemo123!` |

## Prérequis

- Staging Web.
- Usage **interne** (pas nécessairement en démo client final).
- Compte seed multi-profils (`admin` + `client` + `vet` VetPlus) — bascule profil possible dans la topbar.

## Étapes

1. Se connecter avec `admin.demo`.
2. Vérifier le **tableau de bord admin** (métriques).
3. Ouvrir **Utilisateurs** : filtres / rôles visibles.
4. Ouvrir **Commerciaux** : liste + assignations lisibles.
5. Ouvrir **Support** : liste des tickets ; ouvrir un ticket si présent.
6. (Optionnel) Paiements / commissions — lecture seule.
7. (Optionnel) Menu profil → basculer vers **vet** (VetPlus) puis revenir **admin**.

## Résultat attendu

- Accès admin sans erreur.
- Pages utilisateurs / commerciaux / support utilisables.
- Un compte véto ne doit **pas** accéder à l’espace admin (test optionnel : se déconnecter → `vet.demo` → tenter d’ouvrir l’URL admin).

## Checklist

| | Résultat |
|--|----------|
| Dashboard admin | OK / KO / N/A |
| Users | OK / KO / N/A |
| Commercials | OK / KO / N/A |
| Support | OK / KO / N/A |

## Zone retour

- Bug :
- Friction UX :
- Idée :

→ [`TEMPLATE-RETOUR.md`](../TEMPLATE-RETOUR.md)

---

*Réf. QA : D1–D5, D14*
