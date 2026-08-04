# 44 — SMS transactionnel (Telnyx)

**Statut : module tag `dev`, dry-run par défaut.** Aucun SMS ne quitte le système tant que `SMS_DRY_RUN=false` n'est pas posé volontairement.

## Pourquoi

Les notifications client passaient par email + push FCM uniquement (cf. [08-MESSAGERIE-NOTIFICATIONS.md](08-MESSAGERIE-NOTIFICATIONS.md)). Deux manques :

- le **rappel J-1 de RDV n'existait pas** (backlog de [31-PRECONSULTATION.md](31-PRECONSULTATION.md)) ;
- le push ne touche que les clients ayant installé l'app.

Le SMS couvre les deux, pour trois messages transactionnels liés au soin.

> **Positionnement** : le pitch commercial vend « fini les SMS personnels dispersés ». Ce module ne le contredit pas — il remplace le SMS *personnel du véto, non tracé* par un SMS *cabinet, journalisé au dossier, désactivable*. Distinction à tenir dans le discours commercial.

## Périmètre

| Kind | Déclencheur | Code |
|------|-------------|------|
| `visit_confirmed` | RDV confirmé (tous chemins) + reschedule accepté | `onVisitConfirmed` / `onVisitRescheduleAccepted` (`handlers/preconsult.go`) |
| `visit_reminder` | Cron J-1 | `internalRunVisitReminders` (`handlers/visit_reminders.go`) |
| `visit_reschedule` | Nouveau créneau proposé ou imposé | `propose_reschedule` / `reschedule_direct` (`handlers/client_enrichment.go`) |

Hors périmètre : envoi manuel depuis la fiche client, SMS marketing, SMS vers les pros.

## Architecture

```
handlers/sms.go            couche dispatch (gates, E.164, i18n, journal)
handlers/visit_reminders.go job cron J-1 (claim + push + SMS)
handlers/sms_webhook.go     webhooks DLR + SMS entrants
notifications/sms/          Sender (interface) · TelnyxClient · NopSender
                            NormalizeE164 · VerifyWebhookSignature · ParseWebhook
store/sms_log.go            journal, claim idempotent, opt-out, purge
```

`sms.Sender` suit le même contrat que `fcm.Pusher` : interface + implémentation no-op + constructeur à dégradation gracieuse (`NewFromConfig`) — un module mal configuré ne bloque jamais le boot de l'API. `TelnyxClient` calque `pharmacy/vamreg_client.go` (dry-run, timeout 30 s, `io.LimitReader`, erreurs `telnyx_http_<code>`). Aucun SDK : `net/http` seul.

### Numéros (E.164)

`identity.users.contact_phone` est du texte libre (validation trim + 40 runes). `NormalizeE164` fait une conversion **déterministe et volontairement conservatrice** : séparateurs retirés, `00` → `+`, `0` national → indicatif de `SMS_DEFAULT_REGION` (BE/FR/NL/LU), puis validation `^\+[1-9][0-9]{6,14}$`. Tout ce qui ne matche pas est **rejeté et journalisé** (`skipped/invalid_phone`) plutôt que deviné. Si le taux de skip devient élevé, passer à `nyaruka/phonenumbers`.

## Consentement (opt-out)

Double gate avant tout envoi : topic `visits` **et** canal `sms` sur `notifications.client_preferences` (`DEFAULT TRUE`).

- Client : `GET|PATCH /me/notification-preferences` (champ `sms`), écran Flutter *Préférences de notification* (`Key('settings_sms_pref_toggle')`).
- **STOP entrant** → `sms = false` automatiquement (voir webhooks). C'est l'opt-out légal ; il ne dépend d'aucune action du cabinet.
- `START` réactive le canal.

## Webhooks Telnyx

Deux URLs à coller dans le **Messaging Profile** (`Webhook API Version` = **v2**) :

| Champ portail | URL |
|---|---|
| Webhook URL | `https://<host-api>/api/v1/notifications/webhooks/telnyx` |
| Webhook Failover URL | `https://<host-api>/api/v1/notifications/webhooks/telnyx/failover` |

Les deux chemins exécutent le même traitement ; le chemin failover **marque l'origine** (`sms_inbound.via_failover`, log `via FAILOVER url`). Sans ce marquage, une défaillance de la primaire serait invisible.

> **Limite assumée** : les deux URLs pointent le même service Cloud Run. Si ce service est indisponible, les deux échouent — le failover ne protège donc que d'un incident propre à une route, pas d'une panne d'instance. Pour une vraie résilience il faudrait un domaine de panne distinct (sink Cloud Function → Pub/Sub pour rejeu). À arbitrer si la perte d'un STOP pendant une indisponibilité devient inacceptable.

Sécurité :

- signature **Ed25519** vérifiée (`telnyx-signature-ed25519`, `telnyx-timestamp`, payload signé = `<timestamp>|<body>`), clé publique via `TELNYX_PUBLIC_KEY` (portail Telnyx » Auth) ;
- tolérance d'horloge **5 min** (anti-rejeu) ;
- rate limiter 120/min ; corps borné à 1 Mo ; routes hors `AuthMiddleware` (comme Stripe/Billit) ;
- réponse 2xx immédiate — un non-2xx déclencherait les retries puis la bascule failover.

Traitements : `message.sent` / `message.finalized` → DLR rapproché par `provider_message_id` (`delivery_status`, `delivered_at`, `delivery_error`) ; `message.received` → `notifications.sms_inbound` + application de STOP/START. Relivraison idempotente (index unique sur `provider_message_id`).

Un numéro entrant est rapproché d'un compte par les **9 derniers chiffres** de `contact_phone`. En cas de zéro ou plusieurs correspondances, le message est conservé **sans rattachement** : on ne coupe jamais le canal d'un client sur une correspondance douteuse.

## Rappel J-1

`POST /api/v1/internal/visit-reminders/run`, header `X-Visit-Reminders-Secret` (comparaison temps constant, `secretHeaderOK`). Cron : `make gcp-visit-reminders-scheduler` → quotidien **17:00 Europe/Brussels**.

Fenêtre `[now+1h, now+VISIT_REMINDER_LOOKAHEAD_HOURS[` (défaut 30 h) : à 17:00 elle couvre exactement les RDV du lendemain. Exclusions : statut ≠ `confirmed`, `consultation_session` (walk-in).

**Idempotence** : `ClaimVisitReminderSms` insère en `status='sending'` avec `ON CONFLICT DO NOTHING` sur l'index unique partiel `sms_log_reminder_once (visit_id, scheduled_for) WHERE kind='visit_reminder'`. La clé est **(visite, créneau)** — un RDV déplacé redevient donc éligible à un rappel, ce qui est voulu. Garantie : **au plus un** rappel par créneau, même en concurrence ou sur run rejoué. Corollaire assumé : un téléphone corrigé pendant la fenêtre ne sera pas rattrapé.

Le job envoie aussi un **push** (gratuit, infra existante) avec sa propre gate `prefs.Visits`.

## Journal et RGPD

`notifications.sms_log` : `status` = `sending|sent|dry_run|skipped|error`, `error` porte aussi la raison de skip (`pref_opt_out`, `no_phone`, `invalid_phone`). **Le corps du message n'est jamais stocké** (kind + locale déterminent le gabarit) ; seul `to_phone` est une donnée personnelle. `notifications.sms_inbound` conserve le corps **tronqué à 160 caractères**, nécessaire au support (« j'ai répondu STOP »).

| Obligation | Couverture |
|---|---|
| Export | `store/user_export.go` — clés `smsNotifications` et `smsInbound` |
| Effacement | `ON DELETE CASCADE` / `SET NULL` vers `identity.users` (la purge client supprime le compte) |
| Rétention | `handlers/retention.go` — `smsLogRetention` = **12 mois** (minimisation : plus strict que les 3 ans du compte) |

## Configuration

| Variable | Défaut | Rôle |
|---|---|---|
| `SMS_ENABLED` | `false` | Monte le module (routes webhook incluses) |
| `SMS_DRY_RUN` | **`true`** | Logue au lieu d'appeler Telnyx |
| `TELNYX_API_KEY` | — | Envoi (Secret Manager `petsfollow-telnyx-api-key`) |
| `TELNYX_MESSAGING_PROFILE_ID` | — | Profil d'envoi |
| `TELNYX_FROM` | — | Expéditeur optionnel (long code / alphanumérique) |
| `TELNYX_PUBLIC_KEY` | — | Vérification des webhooks (`petsfollow-telnyx-public-key`) |
| `SMS_DEFAULT_REGION` | `BE` | Indicatif des numéros nationaux |
| `VISIT_REMINDERS_SECRET` | — | Header du cron J-1 (`petsfollow-visit-reminders-secret`) |
| `VISIT_REMINDER_LOOKAHEAD_HOURS` | `30` | Fin de fenêtre du rappel |

`ValidateSMS()` (boot) refuse le mode live hors env dev sans `TELNYX_API_KEY`, sans profil/expéditeur, sans `VISIT_REMINDERS_SECRET` et **sans `TELNYX_PUBLIC_KEY`** — sinon un STOP entrant serait rejeté, donc l'opt-out non honoré. Jamais de retombée silencieuse en dry-run. Staging/prod forcent `SMS_DRY_RUN=true` dans `infra/gcp/lib/deploy-run-args.sh` : le passage en live est un geste explicite.

## Rédaction des gabarits

Clés `sms.visit_confirmed|visit_reminder|visit_reschedule` dans les **8 locales** (`platform/i18n/locales/*.json`), variables `{petName}` et `{when}` (Europe/Brussels, `JJ/MM/AAAA HH:MM`).

`catalog_sms_test.go` verrouille deux régimes de longueur, car le coût est facturé **par segment** :

- **locales latines** : GSM-7, **160** caractères. Les caractères hors GSM-7 (`ê î ô û ë ï`, `á í ó ú`, `õ`…) sont **interdits** car ils basculeraient le message en UCS-2, donc à 70 caractères et au moins deux segments. `é è à ù ç ñ ä ö ü` sont GSM-7 et autorisés.
- **locales cyrilliques (uk, ru)** : le cyrillique n'existe pas en GSM-7, ces messages sont nécessairement UCS-2 → **70** caractères. Leurs gabarits sont donc plus courts (pas d'incitation « ouvrez l'app ») : à 70 caractères, l'information du créneau primes.

Ordre de grandeur du coût : ~0,045–0,075 € par segment sortant BE/FR. `sms_log` fournit la volumétrie avant tout passage en live.

## Tests

| Fichier | Couvre |
|---|---|
| `notifications/sms/telnyx_test.go` | httptest : succès, `telnyx_http_422`, dry-run sans réseau, clé requise |
| `notifications/sms/normalize_test.go` | E.164 (formats BE/FR, `00`, rejets) |
| `notifications/sms/webhook_test.go` | Ed25519 (corps altéré, horodatage périmé, mauvaise clé), parsing v2, STOP/START, statuts d'échec |
| `platform/i18n/catalog_sms_test.go` | parité 8 locales, interpolation, longueur GSM-7 / UCS-2 |
| `handlers/sms_integration_test.go` | confirmation → `sent` + E.164 ; opt-out → `pref_opt_out` ; téléphone invalide ; reschedule ; module off = zéro ligne |
| `handlers/visit_reminders_integration_test.go` | 401 sans secret ; envoi puis **2e run sans doublon** ; nouveau créneau = nouveau rappel ; opt-out sans claim ; walk-in et annulé exclus |
| `handlers/sms_webhook_integration_test.go` | signature rejetée ; DLR `delivered` / `delivery_failed` ; STOP → `sms=false` + relivraison idempotente ; START ; failover marqué |
| `flutter/test/features/settings/notification_prefs_sms_test.dart` | toggle → `PATCH sms:false` sans emporter les autres canaux ; relecture d'un canal coupé |

## Vérification locale

```bash
make up-infra && make migrate && make seed && make api-dev   # SMS_ENABLED=true, SMS_DRY_RUN=true
# Confirmer un RDV depuis le Web Pro, puis :
psql -c "SELECT kind, status, to_phone, delivery_status FROM notifications.sms_log ORDER BY created_at DESC LIMIT 5;"

# Rappel J-1 (le 2e appel ne doit rien renvoyer de plus)
curl -X POST localhost:8291/api/v1/internal/visit-reminders/run \
  -H "X-Visit-Reminders-Secret: dev-visit-reminders"
```

Les webhooks exigent une signature Ed25519 valide : les tester en local passe par les tests d'intégration (`go test ./internal/handlers/ -run TestTelnyxWebhook`), pas par un `curl` nu.

## Avant de passer en live

1. Acheter un numéro (long code) et créer le Messaging Profile ; renseigner les deux URLs de webhook ci-dessus + API version v2.
2. Créer les secrets : `petsfollow-telnyx-api-key`, `petsfollow-telnyx-public-key`, `petsfollow-visit-reminders-secret`, puis redéployer (montés automatiquement par `pf_api_secrets`).
3. `make gcp-visit-reminders-scheduler`.
4. Vérifier la volumétrie et les taux de skip dans `sms_log` en dry-run.
5. Seulement ensuite : `SMS_DRY_RUN=false`.

Choix de l'expéditeur : préférer un **long code** au sender alphanumérique — l'alphanumérique est de plus en plus restreint en France pour le transactionnel et n'accepte pas de réponse, donc pas de STOP. Ces messages étant transactionnels avec opt-out in-app, l'absence de STOP serait défendable, mais le long code évite le débat.

**Non traité en v1** : quiet hours (les événements suivent l'activité du cabinet, le rappel est fixé à 17:00), sink de failover en domaine de panne distinct.
