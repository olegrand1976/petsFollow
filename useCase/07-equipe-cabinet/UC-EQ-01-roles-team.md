# UC-EQ-01 — Rôles équipe cabinet

| | |
|--|--|
| **ID** | `UC-EQ-01` |
| **Durée** | ~12 min |
| **Priorité** | Important |
| **Surface** | Web VetPro |

## Objectif

Vérifier la page équipe et que collègue / assistante / secrétaire se connectent avec des droits adaptés.

## Acteurs

| Rôle | Compte | MDP |
|------|--------|-----|
| Véto référence | `vet.demo@petsfollow.test` | `VetDemo123!` |
| Collègue | `vet.colleague@petsfollow.test` | `VetDemo123!` |
| Assistante | `vet.assist@petsfollow.test` | `VetDemo123!` |
| Secrétaire | `secretary.demo@petsfollow.test` | `VetDemo123!` |

## Prérequis

- Staging Web.
- Seed équipe VetPlus.

## Étapes

1. Se connecter en `vet.demo` → ouvrir la page **Équipe**.
2. Vérifier la liste des membres (collègue, assistante, secrétaire).
3. Vérifier que les **droits** affichent des libellés clairs (ex. « Voir les clients ») avec info-bulle au survol — pas de codes techniques.
4. Se déconnecter → se connecter en `vet.colleague` : accès cabinet cohérent.
5. Se connecter en `vet.assist` : noter ce qui est accessible / masqué (droits réduits) — **Histo. consultations** visible.
6. Se connecter en `secretary.demo` : focus secrétariat (**Agenda** + clients) ; **pas** d’entrée **Histo. consultations** dans le menu (droit `consultations.history.read` OFF par défaut ; réactivable via Équipe).

## Résultat attendu

- Page Équipe lisible côté véto référence (droits compréhensibles ; tip Agenda ≠ historique).
- Chaque rôle se connecte ; l’espace **Admin** reste inaccessible.
- Différences de droits **perceptibles** entre assistante / secrétaire / collègue (dont historique consultations).

## Checklist

| | Résultat |
|--|----------|
| Page équipe (vet.demo) | OK / KO / N/A |
| Login collègue | OK / KO / N/A |
| Login assistante | OK / KO / N/A |
| Login secrétaire | OK / KO / N/A |

## Zone retour

- Bug :
- Friction UX :
- Idée :

→ [`TEMPLATE-RETOUR.md`](../TEMPLATE-RETOUR.md)

---

*Réf. QA : L1–L2*
