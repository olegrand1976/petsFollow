# UC-VP-06 — Lieux (sites) du cabinet

| | |
|--|--|
| **ID** | `UC-VP-06` |
| **Durée** | ~5 min |
| **Priorité** | Démo |
| **Surface** | Web VetPro |

## Objectif

Montrer qu’un cabinet multi-sites peut lister, créer et gérer ses **lieux** (antenne + salles) depuis une entrée dédiée **Sites**, et basculer l’agenda via le switcher topbar.

## Acteurs

| Rôle | Compte | MDP |
|------|--------|-----|
| Véto | `vet.demo@petsfollow.test` | `VetDemo123!` |

## Prérequis

- Staging Web avec soft-GA sites (`NUXT_PUBLIC_SITES_UI_ENABLED`).
- Seed VetPlus : primary + Antenne Liège (ou créer un 2ᵉ site pendant la démo).

## Étapes

1. Se connecter en `vet.demo`.
2. Ouvrir **Sites** dans la nav (section Cabinet), ou **Paramètres → Calendrier → Gérer les sites**.
3. Vérifier la liste des lieux (au moins le site principal) ; créer une antenne si besoin (nom + ville optionnelle).
4. Sur un site actif : ajouter une **salle**, la renommer, la désactiver puis la réactiver.
5. Revenir au **Calendrier** : le switcher topbar apparaît dès qu’il y a ≥2 sites ; filtrer un site puis « Tous les sites ».

## Résultat attendu

- Page `/sites` accessible et opérationnelle.
- Création / rename / désactivation de sites et salles sans erreur.
- Switcher topbar visible en multi-sites ; agenda filtré correctement.

## Notes démo

- Soft-GA : si le flag est off, l’entrée **Sites** disparaît (filtre forcé sur le site par défaut).
- Les créneaux client et vacances restent dans **Paramètres → Calendrier** (par site via le switcher).
