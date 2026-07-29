# UC-X-09 — Client visualise / partage une consultation (PDF 24 h)

| | |
|--|--|
| **ID** | `UC-X-09` |
| **Durée** | ~10 min |
| **Priorité** | Démo |
| **Surface** | Flutter Client → e-mail → page Web publique `/consultation/{token}` |

## Objectif

Montrer qu’un propriétaire peut **ouvrir le compte-rendu finalisé** d’une visite dans l’app, puis l’envoyer à un vétérinaire **hors plateforme** via un lien magique **24 h** pour télécharger un **PDF** brandé petsFollow (promo + contact commercial).

## Acteurs

| Rôle | Compte | MDP |
|------|--------|-----|
| Client | `client.demo@petsfollow.test` | `ClientDemo123!` |
| Véto (auteur CR) | `vet.demo@petsfollow.test` | `VetDemo123!` |
| Destinataire | boîte MailHog / e-mail test | — |
| Commercial (contexte) | `commercial.demo@petsfollow.test` | `CommercialDemo123!` |

## Prérequis

- Seed + app client + API + SMTP (MailHog local ou staging).
- Au moins une visite avec **CR finalisé** pour un animal actif de `client.demo` (sinon créer une consultation Pro / Pro Light puis **Enregistrer** + **Terminer** — le mark-done finalise le brouillon non vide ; ou bouton Finaliser explicite).
- Distinct de UC-X-08 (dossier complet ZIP).

## Étapes

1. **App client** — login `client.demo` → fiche animal → **Historique des visites**.
2. Dans **Consultations**, ouvrir une visite avec compte-rendu.
3. Vérifier l’affichage : méta visite + section(s) CR (tous les auteurs finalisés).
4. Appuyer sur **Envoyer à un vétérinaire** → avertissement PHI + e-mail + **consentement** obligatoire → confirmer.
5. Ouvrir l’e-mail : marketing petsFollow, CTA inscription, téléphone commercial, lien 24 h.
6. Ouvrir `/consultation/...` → **Télécharger le PDF** → vérifier en-tête petsFollow + corps CR + bandeau commercial.
7. (Optionnel) Forcer l’expiration → page « lien expiré ».

## Résultat attendu

- Snackbar succès côté client.
- Brouillons **non** visibles / non partageables.
- PDF téléchargeable avant expiry ; 410 après 24 h.

## Checklist

| | Résultat |
|--|----------|
| Liste consultations cliquable | OK / KO / N/A |
| Lecture CR multi-auteurs | OK / KO / N/A |
| Consentement PHI | OK / KO / N/A |
| E-mail + tél. commercial | OK / KO / N/A |
| PDF brandé | OK / KO / N/A |
| Expiry 24 h | OK / KO / N/A |

## Zone retour

- Bug :
- Friction UX :
- Idée :

→ [`TEMPLATE-RETOUR.md`](../TEMPLATE-RETOUR.md)

---

*Réf. QA : H14 · distinct de UC-X-08 (dossier) et UC-X-05 (ACL care_pro)*
