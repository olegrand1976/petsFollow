# UC-CO-02 — Encode véto/client + activation

| | |
|--|--|
| **ID** | `UC-CO-02` |
| **Durée** | ~15 min |
| **Priorité** | Démo |
| **Surface** | Web Commercial |
| **Destructif** | Oui — crée des comptes / activations de test |

> **Skip** si [`UC-X-07`](../10-interactions/UC-X-07-funnel-commercial-complet.md) déjà passé (funnel complet).

## Objectif

Vérifier le parcours commercial : encoder un vétérinaire ou un client lié, puis constater le chemin vers activation payante / commission.

## Acteurs

| Rôle | Compte | MDP |
|------|--------|-----|
| Commercial | `commercial.demo@petsfollow.test` | `CommercialDemo123!` |

## Prérequis

- Staging Web.
- Utiliser un **email jetable** unique pour l’encode (ex. `demo.uc.co02+1430@example.com`) — ne pas toucher aux comptes `*.petsfollow.test`.
- Sur staging partagé : préférer encoder un client de test plutôt qu’un véto déjà assigné.

## Étapes

1. Se connecter en `commercial.demo`.
2. Aller dans **Vétérinaires** (encode) **ou** encode **client lié** au cabinet.
3. Créer un compte de test avec email unique + consentement si demandé.
4. Si encode véto : vérifier l’assignation au commercial.
5. Si encode client : rattacher / créer un animal et pousser jusqu’à **activation payante** (ou mode test selon env.).
6. Ouvrir **Commissions** et vérifier qu’une ligne apparaît (ou le chemin pour y arriver est clair).

## Résultat attendu

- Encode réussi (ou message clair si déjà assigné ailleurs — 409 métier compréhensible).
- Chemin activation → commission compréhensible pour une démo.

## Checklist

| | Résultat |
|--|----------|
| Encode | OK / KO / N/A |
| Activation / abonnement animal actif | OK / KO / N/A |
| Commission visible | OK / KO / N/A |

## Zone retour

- Bug :
- Friction UX :
- Idée :

→ [`TEMPLATE-RETOUR.md`](../TEMPLATE-RETOUR.md)

---

*Réf. QA : E1.3–E1.7*
