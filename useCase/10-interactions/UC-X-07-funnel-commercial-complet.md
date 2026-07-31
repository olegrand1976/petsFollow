# UC-X-07 — Funnel commercial complet

| | |
|--|--|
| **ID** | `UC-X-07` |
| **Durée** | ~20 min |
| **Priorité** | Démo |
| **Surfaces** | Web Commercial (+ Client / VetPro si activation réelle) |
| **Destructif** | Oui — crée prospects / comptes / activations de test |

> Couvre aussi [`UC-CO-02`](../04-commercial/UC-CO-02-encode-activation.md) : pas besoin de refaire CO-02 après.

## Objectif

Enchaîner le parcours commercial de bout en bout : prospect → encode → client / animal payant → commission.

## Acteurs

| Rôle | Compte | MDP |
|------|--------|-----|
| Commercial | `commercial.demo@petsfollow.test` | `CommercialDemo123!` |

## Prérequis

- Staging Web.
- Emails **jetables** pour l’encode (ne pas écraser les comptes `*.petsfollow.test`).
- Complète / approfondit `UC-CO-01` + `UC-CO-02`.

## Étapes

1. Login `commercial.demo` → aperçu commercial.
2. Ouvrir / mettre à jour un **prospect** (contact → RDV → résultat) sur une fiche de test.
3. **Encoder** un véto de test **ou** un client lié à un cabinet déjà assigné (Camille → VetPlus).
4. Pousser l’activation d’un **animal payant** (paiement test / mode test).
5. Ouvrir **Commissions** : vérifier l’apparition d’une ligne / montant cohérent.
6. (Optionnel) Se connecter côté VetPro du cabinet pour montrer le client activé.

## Résultat attendu

- Chaîne commerciale compréhensible pour une démo terrain.
- Commission liée à l’activation.
- Messages d’erreur clairs si cabinet déjà assigné à un autre commercial.

## Checklist

| | Résultat |
|--|----------|
| Prospect CRM | OK / KO / N/A |
| Encode | OK / KO / N/A |
| Activation | OK / KO / N/A |
| Commission | OK / KO / N/A |

## Zone retour

- Bug :
- Friction UX :
- Idée :

→ [`TEMPLATE-RETOUR.md`](../TEMPLATE-RETOUR.md)

---

*Réf. QA : E1 · flux commercial [`06-FLUX-UTILISATEURS.md`](../../documentation/06-FLUX-UTILISATEURS.md)*
