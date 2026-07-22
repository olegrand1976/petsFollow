# Messagerie & notifications — petsFollow

## Messagerie interne

- Threads client ↔ véto (lié au cabinet / relation).
- Messages texte + **media** (`POST …/messages/media`).
- Marquage lu thread / read-all.
- **Mode indisponible** véto : `PUT/GET /vet/availability` — le client voit l’indisponibilité.

Surfaces : Pro `/messages` · Flutter messagerie animal/véto.

## Notifications email (livré)

Emails transactionnels via notifier Go (confirm email, reset MDP, etc.) selon locale user.

**Parcours découverte / fidélisation client** (drip 12 mois) : scheduler in-process + tables `discovery.email_*` — détail [23-PARCOURS-EMAIL-CLIENT.md](23-PARCOURS-EMAIL-CLIENT.md). Respecte `client_preferences.discovery` / `.billing`. Désabonnement : `GET/POST /api/v1/public/journey/unsubscribe?token=…`.

**Digest produit quotidien** (interne) : synthèse fonctionnelle des évolutions du jour → emails aux rôles `admin` / `commercial` / `commercial_manager` à 18:00 Europe/Brussels. Détail [25-PRODUCT-DIGEST.md](25-PRODUCT-DIGEST.md).

Préférences :

- Véto : `GET/PUT /vet/notification-preferences` (`emailOnMessage`, `emailOnHeartrate`, `emailOnVisitRequest`)
- Client : `GET/PATCH /me/notification-preferences` (`hr`, `care`, `visits`, `messages`, `discovery`, `billing`)

Quand le **client** écrit un message et que le véto a `email_on_message`, un email est envoyé au véto.

Quand le **client** crée une demande de RDV (ou propose un déplacement au véto) et que `email_on_visit_request` est actif : e-mail avec CTA `/calendar?visit={id}`.

## Push FCM (livré)

Device tokens : `PUT /me/device-tokens` (enregistrés par l’app Flutter au login).

Envoi serveur (API Go, package `internal/notifications/fcm`) via Firebase Admin + ADC :

| Événement | Pref client | Payload `data.type` |
|-----------|-------------|---------------------|
| Véto envoie un message (texte/média) | `messages` | `message` (+ `threadId`) |
| Véto confirme un RDV | `visits` | `visit_confirmed` (+ `visitId`, `petId`) |
| Véto propose un RDV | `visits` | `visit_proposed` |
| Véto propose un déplacement | `visits` | `visit_reschedule` |

- Locale des titres/corps : `users.preferred_locale` (clés `push.*` dans `go/internal/platform/i18n/locales/`).
- Sans credentials ADC / si `FCM_ENABLED=false` : no-op (handlers restent 200).
- Tokens invalides (unregistered) : supprimés de `notifications.device_tokens`.

Flutter : handlers `onMessage` / `onMessageOpenedApp` + notif locale au premier plan ; tap → onglet Messages ou timeline animal.

Prérequis ops : projet Firebase `premedica-prod-2025`, ADC (`GOOGLE_APPLICATION_CREDENTIALS` en local, ou SA Cloud Run avec droits FCM).

## Hors scope actuel

WebSocket temps réel — refresh à l’ouverture de l’onglet Messages / à la réception push.
