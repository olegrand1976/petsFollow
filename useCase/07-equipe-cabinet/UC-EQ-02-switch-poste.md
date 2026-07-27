# UC-EQ-02 — Switch poste partagé (veille)

| | |
|--|--|
| **ID** | `UC-EQ-02` |
| **Durée** | ~10 min |
| **Priorité** | Important |
| **Surface** | Web VetPro |

## Objectif

Sur un PC partagé au cabinet, basculer rapidement d’un compte équipe à un autre avec mot de passe, et vérifier que l’inactivité met le poste en veille.

## Acteurs

| Rôle | Compte | MDP |
|------|--------|-----|
| Véto référence | `vet.demo@petsfollow.test` | `VetDemo123!` |
| Secrétaire | `secretary.demo@petsfollow.test` | `VetDemo123!` |

## Prérequis

- Staging Web.
- Seed équipe VetPlus.

## Étapes

1. Se connecter en `vet.demo` → ouvrir **Clients** (ou une fiche animal).
2. Dans le **header**, repérer les avatars de l’équipe.
3. Cliquer l’avatar de la secrétaire → saisir son mot de passe → valider.
4. Vérifier que l’on arrive sur le dernier écran connu de ce compte (ou l’accueil si première fois).
5. Naviguer vers **Agenda** en secrétaire.
6. Rebasculer vers `vet.demo` (mot de passe) → vérifier le retour vers l’écran précédent du véto si possible.
7. Laisser le poste inactif **2 minutes** (ou demander un lock forcé à l’équipe tech en démo) → écran de veille.
8. Déverrouiller avec un compte équipe + mot de passe.

## Résultat attendu

- Switch toujours protégé par mot de passe (jamais de bascule silencieuse).
- Veille après inactivité : plus d’accès aux données sans re-saisie du mot de passe.
- Chaque utilisateur retrouve son dernier écran.

## Checklist

| | Résultat |
|--|----------|
| Avatars équipe dans le header | OK / KO / N/A |
| Switch avec mot de passe | OK / KO / N/A |
| Restauration dernier écran | OK / KO / N/A |
| Veille 2 min + déverrouillage | OK / KO / N/A |

## Zone retour

- Bug :
- Friction UX :
- Idée :

→ [`TEMPLATE-RETOUR.md`](../TEMPLATE-RETOUR.md)

---

*Réf. QA : L7*
