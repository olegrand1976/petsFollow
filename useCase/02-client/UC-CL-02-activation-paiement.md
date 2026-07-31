# UC-CL-02 — Activation animal + paiement

| | |
|--|--|
| **ID** | `UC-CL-02` |
| **Durée** | ~12 min |
| **Priorité** | Démo |
| **Surface** | Flutter Client |
| **Destructif** | Oui — consomme `client.vide` (ne plus le réutiliser pour `UC-X-04`) |

## Objectif

Créer un animal depuis un compte vide et activer un plan payant (mensuel / annuel / triennal).

## Acteurs

| Rôle | Compte | MDP |
|------|--------|-----|
| Client vide | `client.vide@petsfollow.test` | `ClientDemo123!` |

## Prérequis

- App staging.
- Sur staging, paiement Stripe **test** ou parcours mode test selon config — noter si le paiement échoue.
- Si le compte a **déjà** un animal payant → **N/A** (ou reset seed).
- Réservé à ce scénario : **ne pas** utiliser `client.vide` pour le lien cabinet.

## Étapes

1. Se connecter avec `client.vide`.
2. Créer un **nouvel animal** (nom, espèce, infos demandées).
3. Choisir un plan : **3,50 € / mois**, **35 € / an** ou **95 € / 3 ans** (recommandé).
4. Aller jusqu’au paiement / activation.
5. Après succès : vérifier que l’animal est **actif** (accès Care, messages, relevés selon offre).

## Résultat attendu

- Plans affichés correctement (pas d’addons Family/Care+/Horse vendus).
- Après paiement : abonnement animal actif actif, fonctionnalités de suivi accessibles.

## Checklist

| | Résultat |
|--|----------|
| Création animal | OK / KO / N/A |
| Affichage plans | OK / KO / N/A |
| Activation réussie | OK / KO / N/A |

## Zone retour

- Bug :
- Friction UX :
- Idée :

→ [`TEMPLATE-RETOUR.md`](../TEMPLATE-RETOUR.md)

---

*Réf. QA : F1.4–F1.5, H6*
