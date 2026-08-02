# Relevé respiratoire (FR)

> Nom technique historique inchangé : schéma/API `heartrate.*` (sessions, BPM). Vocabulaire produit = **fréquence respiratoire**.


## Principe

La durée du relevé (**15 / 30 / 60 s**) est **définie par le vétérinaire** du cabinet (paramètres Pro : onboarding + `/settings`). Le client ne peut démarrer une session qu’avec une durée autorisée pour le `practice_id` de l’animal.

## Flux client (Flutter)

1. **Prêt** — durée(s) proposées = `pet.heartrateDurationsSec` (cabinet) ; défaut UI = **plus longue** durée activée
2. **En cours** — timer = durée choisie, tap à chaque respiration
3. **Résultat** — BPM + alerte seuil + **commentaire optionnel** (max 500 car.)
4. **Valider et envoyer au véto** ou **Recommencer**

Le commentaire est stocké sur la session et visible côté Pro dans le tableau des relevés validés (et dans le corps de l’entrée timeline associée).

## Configuration Pro (véto)

- Cases à cocher 15 / 30 / 60 s dans le profil cabinet
- Au moins une durée requise
- Stockage : `practice.practices.heartrate_durations_sec`

## Statuts API

`in_progress` → `pending_validation` → `validated` | `cancelled`

Seuls les relevés **validated** sont visibles du véto.

## Endpoints

- `POST /api/v1/pets/{id}/heartrate/sessions` — body `{ "durationSec": N }` ; si omis → **max** des durées cabinet
- `PATCH /api/v1/heartrate/sessions/{id}` (tapCount)
- `POST /api/v1/heartrate/sessions/{id}/validate` — body optionnel `{ "comment": "…" }` (trim, max 500 runes)
- `POST /api/v1/heartrate/sessions/{id}/cancel`

## BPM

`BPM = (tap_count × 60) / duration_sec`

### Espèces

- **dog / cat / horse** : relevé FR autorisé.
- **other** : pas de contrôle FR (UI masquée ; `POST …/heartrate/sessions` → `403` `heartrate_not_supported`). Le poids reste autorisé.

### Alerte seuil (hausse vs précédent)

Alerte si **hausse** par rapport au **dernier relevé validé** du même animal :

`bpm_actuel − bpm_précédent ≥ delta_espèce`

- Premier relevé (pas d’historique validé) : **pas** d’alerte.
- Seuils en base : `heartrate.species_alert_deltas` (migration `000078`, seed **dog / cat / horse = 30**).
- À la validation d’un relevé `is_alert` : email véto `SendHeartrateThresholdAlert` (si pref `emailOnHeartrate`).

Les anciens seuils absolus 60–140 (`HEARTRATE_MIN/MAX_BPM`) ont été retirés ; seule la hausse vs précédent pilote `is_alert`.

## Accueil Flutter

Actions compactes **par animal** (cœur + poids) sur les cartes Home et la fiche animal — pas de gros CTA global. Bouton cœur masqué si `species == other`.

## Poids (lié fiche animal)

- Saisie client : dialog kg + commentaire optionnel → `POST /api/v1/pets/{id}/weights` (immédiat, visible sur timeline / fiche Pro — **pas** d’email/push dédié).
- CTA Flutter : « Enregistrer » / snackbar « Poids enregistré » (pas « envoyé au véto »).
- Historique : `pets.weight_readings` ; `pets.pets.weight_kg` synchronisé sur le dernier `POST /weights` (un PATCH fiche peut diverger).
- Pro : chart + tableau sur la fiche pet (`GET /api/v1/pets/{id}/weights`).
- Timeline : type `weight` (titre localisé côté client/Pro via `type`, pas le title SQL FR).
