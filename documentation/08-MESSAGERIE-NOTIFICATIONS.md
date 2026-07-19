# Messagerie & notifications — petsFollow

## Messagerie interne

- Threads client ↔ véto (lié au cabinet / relation).
- Messages texte + **media** (`POST …/messages/media`).
- Marquage lu thread / read-all.
- **Mode indisponible** véto : `PUT/GET /vet/availability` — le client voit l’indisponibilité.

Surfaces : Pro `/messages` · Flutter messagerie animal/véto.

## Notifications email (livré)

Emails transactionnels via notifier Go (confirm email, reset MDP, etc.) selon locale user.

Préférences :

- Véto : `GET/PUT /vet/notification-preferences`
- Client : `GET/PATCH /me/notification-preferences`

## Push FCM

- Device tokens : `PUT /me/device-tokens` (schéma présent).
- **Envoi push = post-MVP** (pas de pipeline FCM opérationnel dans le périmètre actuel).

## Hors scope actuel

WebSocket temps réel — polling / refresh client pour l’instant.
