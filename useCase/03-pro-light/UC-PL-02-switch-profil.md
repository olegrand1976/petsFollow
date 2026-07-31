# UC-PL-02 — Switch profil pro ↔ client

| | |
|--|--|
| **ID** | `UC-PL-02` |
| **Durée** | ~8 min |
| **Priorité** | Important |
| **Surface** | Flutter Pro Light / Client |

## Objectif

Vérifier qu’un care pro dual (ex. farrier) peut basculer entre son profil professionnel et son profil client personnel.

## Acteurs

| Rôle | Compte | MDP |
|------|--------|-----|
| Farrier dual | `farrier.demo@petsfollow.test` | `CareProDemo123!` |

## Prérequis

- Compte seed avec double profil (pro + client).
- App staging.

## Étapes

1. Se connecter en `farrier.demo` → arriver en **Pro Light**.
2. Trouver l’action de **changement de profil** (réglages ou sélecteur).
3. Basculer vers le profil **client**.
4. Vérifier le application propriétaire (onglets client).
5. Revenir au profil **pro**.

## Résultat attendu

- Passage pro → client et retour sans se déconnecter complètement.
- Interfaces distinctes selon le profil actif.

## Checklist

| | Résultat |
|--|----------|
| Pro → Client | OK / KO / N/A |
| Client → Pro | OK / KO / N/A |

## Zone retour

- Bug :
- Friction UX :
- Idée :

→ [`TEMPLATE-RETOUR.md`](../TEMPLATE-RETOUR.md)

---

*Réf. QA : K1–K2, K5*
