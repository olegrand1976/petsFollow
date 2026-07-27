# UC-X-04 — Lien cabinet (invite / demande de rattachement)

| | |
|--|--|
| **ID** | `UC-X-04` |
| **Durée** | ~12 min |
| **Priorité** | Important |
| **Surfaces** | Web VetPro + Flutter Client |

## Objectif

Rattacher un propriétaire à un cabinet via invitation ou demande de lien, puis accepter côté véto.

## Acteurs

| Rôle | Compte | MDP |
|------|--------|-----|
| Véto | `vet.demo@petsfollow.test` (ou `vet.parc`) | `VetDemo123!` |
| Client | **Email jetable** (inscription app ou invite) — **pas** `client.vide` | selon création |

## Prérequis

- Staging Web + app.
- **Ne pas** utiliser `client.vide` (réservé à `UC-CL-02`).
- Créer un client de test avec un email unique, **ou** partir d’un compte non encore lié au cabinet cible.
- Lien / QR d’invitation selon l’écran disponible.

## Étapes

1. **Web** — depuis une fiche client ou la liste Clients : envoyer une **invitation app** / lien cabinet (vers l’email jetable).
2. **App** — côté client : suivre le flux d’invitation **ou** initier une **demande de rattachement** vers le cabinet.
3. **Web** — ouvrir les **invitations** / demandes en attente → **accepter**.
4. Vérifier que le client apparaît bien lié et que la messagerie / animaux sont cohérents.

## Résultat attendu

- Flux invite ou demande aboutit à un rattachement.
- Refus possible et statut mis à jour (si testé).

## Checklist

| | Résultat |
|--|----------|
| Invite / demande | OK / KO / N/A |
| Acceptation véto | OK / KO / N/A |
| Client lié visible | OK / KO / N/A |

## Zone retour

- Bug :
- Friction UX :
- Idée :

→ [`TEMPLATE-RETOUR.md`](../TEMPLATE-RETOUR.md)

---

*Réf. QA : H4 · invitations Pro C2.8–C2.9 · lien client F5*
