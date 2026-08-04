# Messagerie & notifications — petsFollow

## Messagerie interne

- Threads **cabinet × client × animal** (`messaging.threads.pet_id`) ; un client peut avoir plusieurs threads (un par couple practice/pet).
- Threads **care_pro × client × animal** : `practice_id` **NULL**, unicité `(vet_user_id, client, pet)` — conversation personnelle distincte du fil cabinet (ACL `pet_access` / `client_access` requise).
- Création côté client : `POST /api/v1/messaging/threads` `{ practiceId, petId }` → `GetOrCreateThreadForPet` (ensure). Si l’animal est déjà rattaché à un autre cabinet → `403` `wrong_practice`.
- Création côté care_pro : `POST /api/v1/messaging/threads` `{ clientUserId, petId? }` → `GetOrCreateCareProThread` (403 sans grant).
- Sans véto lié : UI Flutter client verrouillée (pas d’envoi vers cabinet) ; composition après liaison : choix care pro / cabinet + animal filtré par practice.
- Messages texte + **media** (`POST …/messages/media`).
- Marquage lu thread / read-all.
- **Mode indisponible** véto cabinet : `PUT/GET /vet/availability` — le client voit l’indisponibilité (pas applicable aux threads care_pro).

Surfaces : Pro `/messages` · Flutter client (onglet Messages) · Flutter Pro Light **identique** pour staff cabinet et `care_pro` (onglet Messages : non-lus, composer client, envoi).

## Notifications email (livré)

Emails transactionnels via notifier Go (confirm email, reset MDP, etc.) selon locale user.

**Parcours découverte / fidélisation client** (drip 12 mois) : scheduler in-process + tables `discovery.email_*` — détail [23-PARCOURS-EMAIL-CLIENT.md](23-PARCOURS-EMAIL-CLIENT.md). Respecte `client_preferences.discovery` / `.billing`. Désabonnement : `GET/POST /api/v1/public/journey/unsubscribe?token=…`.

**Digest produit quotidien** (interne) : synthèse fonctionnelle des évolutions du jour → emails aux rôles `admin` / `commercial` / `commercial_manager` à 18:00 Europe/Brussels. Détail [25-PRODUCT-DIGEST.md](25-PRODUCT-DIGEST.md).

Préférences :

- Véto : `GET/PUT /vet/notification-preferences` (`emailOnMessage`, `emailOnHeartrate`, `emailOnVisitRequest`)
- Client : `GET/PATCH /me/notification-preferences` (`hr`, `care`, `visits`, `messages`, `discovery`, `billing`, `sms`)

Quand le **client** écrit un message et que le véto a `email_on_message`, un email est envoyé au véto.

Quand le **client** crée une demande de RDV (ou propose un déplacement au véto) et que `email_on_visit_request` est actif : e-mail avec CTA `/calendar?visit={id}`.

Quand un RDV passe à **`confirmed`** (confirm, `ConfirmDirect`, accept reschedule) : création d’un intake pré-consult `pending`, push `visit_confirmed` (+ `preconsult=1`), et **e-mail client** (si pref `visits` et intake nouvellement créé) avec CTA invite app `?preconsult={visitId}` — détail [31-PRECONSULTATION.md](31-PRECONSULTATION.md).

## Push FCM (livré)

Device tokens : `PUT /me/device-tokens` (enregistrés par l’app Flutter au login).

Envoi serveur (API Go, package `internal/notifications/fcm`) via Firebase Admin + ADC :

| Événement | Pref client | Payload `data.type` |
|-----------|-------------|---------------------|
| Pro (cabinet ou care_pro) envoie un message (texte/média) | `messages` | `message` (+ `threadId`) |
| Client répond → push au `vet_user_id` (véto ou care_pro) | (device tokens pro) | `message` (+ `threadId`) |
| Véto confirme un RDV | `visits` | `visit_confirmed` (+ `visitId`, `petId`) |
| Véto propose un RDV | `visits` | `visit_proposed` |
| Véto propose un déplacement | `visits` | `visit_reschedule` |
| Rappel J-1 avant RDV (cron) | `visits` | `visit_reminder` |

- Locale des titres/corps : `users.preferred_locale` (clés `push.*` dans `go/internal/platform/i18n/locales/`).
- Sans credentials ADC / si `FCM_ENABLED=false` : no-op (handlers restent 200).
- Tokens invalides (unregistered) : supprimés de `notifications.device_tokens`.

Flutter : handlers `onMessage` / `onMessageOpenedApp` + notif locale au premier plan ; tap → onglet Messages ou timeline animal.

Prérequis ops : projet Firebase `premedica-prod-2025`, ADC (`GOOGLE_APPLICATION_CREDENTIALS` en local, ou SA Cloud Run avec droits FCM).

## SMS transactionnel Telnyx (tag `dev`, dry-run)

Troisième canal, **client uniquement**, pour trois messages liés au RDV : confirmation, **rappel J-1** (nouveau job cron 17:00 Europe/Brussels) et reprogrammation.

- Double gate : topic `visits` **et** canal `sms` (`client_preferences.sms`, `DEFAULT TRUE` = opt-out).
- **STOP entrant** → `sms = false` automatiquement (webhook Telnyx, signature Ed25519).
- Journal `notifications.sms_log` (statuts + raisons de skip + DLR) ; corps jamais stocké.
- `SMS_DRY_RUN=true` par défaut, forcé en staging/prod : aucun envoi facturé sans geste explicite.

Détail complet (webhooks primaire/failover, idempotence du rappel, E.164, coûts, RGPD) : **[44-SMS-TELNYX.md](44-SMS-TELNYX.md)**.

## Hors scope actuel

WebSocket temps réel — refresh à l’ouverture de l’onglet Messages / à la réception push.

SMS : envoi manuel depuis la fiche client, SMS marketing, SMS vers les pros, quiet hours, et sink de failover webhook en domaine de panne distinct (cf. 44).
