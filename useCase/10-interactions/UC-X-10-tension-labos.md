# UC-X-10 — Tension & prise de sang (client ↔ VetPro)

| | |
|--|--|
| **ID** | `UC-X-10` |
| **Durée** | ~15 min |
| **Priorité** | Démo |
| **Surfaces** | Flutter Client → Web VetPro |

## Objectif

Enregistrer une tension côté client, puis côté cabinet saisir un panel de prise de sang et vérifier timeline + lecture client.

## Acteurs

| Rôle | Compte | MDP |
|------|--------|-----|
| Client | `client.demo@petsfollow.test` | `ClientDemo123!` |
| Véto | `vet.demo@petsfollow.test` | `VetDemo123!` |

## Prérequis

- Animal seed actif côté client.demo / VetPlus.

## Étapes

1. **App** — login `client.demo` → animal → **Tension** → 140/90 Doppler → enregistrer.
2. **Web** — login `vet.demo` → Clients → même animal → onglet **Vitals**.
3. Vérifier KPI / chart / tableau tension (valeur client).
4. Ajouter une tension cabinet (ex. 130/80 oscillométrique).
5. Créer un **panel labo** (labo « BioVet », créatinine hors norme + ALAT normale).
6. Ouvrir le détail panel ; vérifier badge hors norme + timeline.
7. **App** — rouvrir l’animal → **Voir les analyses** → détail du panel.

## Résultat attendu

- Tension client + Pro visibles côté Pro.
- Panel labo Pro-only en écriture ; client en lecture seule.
- Timeline : types tension et prise de sang.

## Checklist

| | Résultat |
|--|----------|
| Tension client → Pro | OK / KO / N/A |
| Panel labo Pro | OK / KO / N/A |
| Lecture client analyses | OK / KO / N/A |

## Zone retour

- Bug :
- Amélioration :
