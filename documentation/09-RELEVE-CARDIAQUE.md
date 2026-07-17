# Relevé cardiaque

## Principe

La durée du relevé (**15 / 30 / 60 s**) est **définie par le vétérinaire** du cabinet (paramètres Pro : onboarding + `/settings`). Le client ne peut démarrer une session qu’avec une durée autorisée pour le `practice_id` de l’animal.

## Flux client (Flutter)

1. **Prêt** — durée(s) proposées = `pet.heartrateDurationsSec` (cabinet) ; défaut UI = **plus longue** durée activée
2. **En cours** — timer = durée choisie, tap à chaque battement
3. **Résultat** — BPM + alerte seuil
4. **Valider et envoyer au véto** ou **Recommencer**

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
- `POST /api/v1/heartrate/sessions/{id}/validate`
- `POST /api/v1/heartrate/sessions/{id}/cancel`

## BPM

`BPM = (tap_count × 60) / duration_sec`

Seuils défaut alerte : 60–140 BPM (`HEARTRATE_MIN_BPM`, `HEARTRATE_MAX_BPM`).
