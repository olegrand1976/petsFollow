# UC-VP-02 — Onboarding véto incomplet

| | |
|--|--|
| **ID** | `UC-VP-02` |
| **Durée** | ~8 min |
| **Priorité** | Important |
| **Surface** | Web VetPro |
| **Destructif** | Oui — une fois l’onboarding terminé, le compte seed n’est plus « incomplet » |

## Objectif

Vérifier qu’un vétérinaire au profil incomplet est forcé de compléter son profil avant d’accéder au cabinet.

## Acteurs

| Rôle | Compte | MDP |
|------|--------|-----|
| Véto onboarding | `vet.onboarding@petsfollow.test` | `VetDemo123!` |

## Prérequis

- Staging Web.
- Si le profil est **déjà complété** sur cet environnement → **N/A** (ou demander un reset seed à l’équipe).

## Étapes

1. Se connecter avec `vet.onboarding@petsfollow.test`.
2. Constater l’écran de **complétion de profil** (pas le tableau de bord cabinet).
3. Remplir les champs demandés (infos cabinet, au moins une durée de relevé respiratoire si proposé).
4. Valider / terminer.
5. Vérifier l’accès au **tableau de bord**.

## Résultat attendu

- Impossible d’accéder au tableau de bord tant que le profil est incomplet.
- Après complétion : accès normal au cabinet.

## Checklist

| | Résultat |
|--|----------|
| Redirect onboarding | OK / KO / N/A |
| Complétion → dashboard | OK / KO / N/A |

## Zone retour

- Bug :
- Friction UX :
- Idée :

→ [`TEMPLATE-RETOUR.md`](../TEMPLATE-RETOUR.md)

---

*Réf. QA : C1.1–C1.2*
