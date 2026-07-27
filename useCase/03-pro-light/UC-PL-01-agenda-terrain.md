# UC-PL-01 — Agenda terrain Pro Light

| | |
|--|--|
| **ID** | `UC-PL-01` |
| **Durée** | ~10 min |
| **Priorité** | Démo |
| **Surface** | Flutter Pro Light |

## Objectif

Se connecter en care pro / VetLight, voir l’agenda du jour et marquer une visite comme faite.

## Acteurs

| Rôle | Compte | MDP |
|------|--------|-----|
| Farrier (ou vet_light) | `farrier.demo@petsfollow.test` **ou** `vetlight.demo@petsfollow.test` | `CareProDemo123!` |

## Prérequis

- App staging (même APK ; application terrain après login care_pro).
- Seed avec animal partagé (ex. Spirit pour farrier) — voir aussi `UC-X-05`.

## Étapes

1. Se connecter avec `farrier.demo` (ou `vetlight.demo`).
2. Vérifier le shell **Pro Light** (Agenda / Clients / Animaux / Réglages — pas les 5 onglets client).
3. Ouvrir **Agenda** → vue **Aujourd’hui** (ou équivalent).
4. Ouvrir une visite / créneau si présent.
5. Marquer **Fait** (ou action équivalente).
6. (Optionnel) Consulter une fiche animal partagée / notes.

## Résultat attendu

- application terrain distinct du application propriétaire.
- Agenda utilisable ; action « Fait » prise en compte.

## Checklist

| | Résultat |
|--|----------|
| Login Pro Light | OK / KO / N/A |
| Agenda Aujourd’hui | OK / KO / N/A |
| Marquer Fait | OK / KO / N/A |

## Zone retour

- Bug :
- Friction UX :
- Idée :

→ [`TEMPLATE-RETOUR.md`](../TEMPLATE-RETOUR.md)

---

*Réf. QA : A9, G*
