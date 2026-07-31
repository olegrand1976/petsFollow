# UC-X-08 — Client envoie le dossier animal à un pro (lien 24 h)

| | |
|--|--|
| **ID** | `UC-X-08` |
| **Durée** | ~10 min |
| **Priorité** | Démo |
| **Surface** | Flutter Client → e-mail → page Web publique /dossier/{token} |

## Objectif

Montrer qu’un propriétaire peut envoyer le **dossier complet** d’un animal (PDF + pièces + carnet + timeline/visites) à un professionnel **hors plateforme**, via un lien magique valable **24 h**, avec message marketing petsFollow et téléphone du commercial.

## Acteurs

| Rôle | Compte | MDP |
|------|--------|-----|
| Client | `client.demo@petsfollow.test` | `ClientDemo123!` |
| Destinataire | boîte MailHog / e-mail test | — |
| Commercial (contexte) | `commercial.demo@petsfollow.test` (Camille, rattachée via vet.demo) | `CommercialDemo123!` |

## Prérequis

- Seed + app client + API + SMTP (MailHog local ou staging).
- Animal actif chez `client.demo` (ex. Bella / Rex).
- Commercial avec téléphone profil (seed Camille) — sinon gate Pro à la connexion.

## Étapes

1. **App client** — login `client.demo` → ouvrir la fiche d’un animal actif.
2. Appuyer sur **Envoyer vers un pro**.
3. Lire l’avertissement « données de santé », saisir l’e-mail du professionnel destinataire, **cocher le consentement** (le bouton d’envoi reste grisé sans la case) → confirmer.
4. Ouvrir l’e-mail reçu : vérifier le ton marketing, le CTA inscription, le **téléphone commercial**, le lien 24 h.
5. Ouvrir le lien → page `/dossier/...` → **Télécharger le dossier**.
6. Vérifier le ZIP : `dossier.pdf` (identité, mesures, visites avec **pro consulté**, liens petsFollow, téléphone commercial) + documents / carnet si présents.
7. (Optionnel) Attendre / forcer l’expiration → page « lien expiré ».

## Résultat attendu

- Snackbar succès côté client.
- E-mail branding petsFollow + CTA pro + contact commercial.
- Téléchargement ZIP avant expiry ; refus après 24 h.
- Messagerie **non** incluse dans le pack.

## Checklist

| | Résultat |
|--|----------|
| CTA fiche animal | OK / KO / N/A |
| Avertissement PHI + consentement obligatoire | OK / KO / N/A |
| E-mail reçu | OK / KO / N/A |
| Téléphone commercial visible (mail + PDF) | OK / KO / N/A |
| ZIP + dossier.pdf | OK / KO / N/A |
| Visites avec pro consulté | OK / KO / N/A |
| Expiry 24 h | OK / KO / N/A |

## Zone retour

- Bug :
- Friction UX :
- Idée :

→ [`TEMPLATE-RETOUR.md`](../TEMPLATE-RETOUR.md)

---

*Réf. QA : envoi dossier animal · distinct de UC-X-05 (ACL care_pro)*
