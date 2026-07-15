# Vision produit — MVP petsFollow

## Personas

- **Dr Martin** — véto libéral, dashboard Pro web
- **Sophie** — propriétaire chien senior, app mobile pets

## Périmètre MVP (inclus)

1. Création animal (client)
2. Suivi clients + animaux (véto)
3. Messagerie interne + mode indisponible véto
4. Relevé cardiaque 60s — Valider envoi véto / Recommencer
5. Timeline historique (messages, relevés validés, événements)

## Hors MVP

Stripe, FCM push, documents GCS, rappels récurrents, WebSocket.

## Comptes seed

Voir [AGENTS.md](../AGENTS.md) pour la liste complète (5 vétos, 7 clients, admin).

Mots de passe : `VetDemo123!` · `ClientDemo123!` · `AdminDemo123!`

Parcours spéciaux : `vet.onboarding@` (profil incomplet), `vet.unverified@` + token `demo-confirm-email`, `client.vide@` (sans animal).
