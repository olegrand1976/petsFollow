# Relevé cardiaque V1

## Flux client (Flutter)

1. **Prêt** — consignes + Démarrer
2. **En cours** — timer 60s, tap à chaque battement
3. **Résultat** — BPM + alerte seuil
4. **Valider et envoyer au véto** ou **Recommencer**

## Statuts API

`in_progress` → `pending_validation` → `validated` | `cancelled`

Seuls les relevés **validated** sont visibles du véto.

## Endpoints

- `POST /api/v1/pets/{id}/heartrate/sessions`
- `PATCH /api/v1/heartrate/sessions/{id}` (tapCount)
- `POST /api/v1/heartrate/sessions/{id}/validate`
- `POST /api/v1/heartrate/sessions/{id}/cancel`

## BPM

`BPM = round(tap_count / 60 × 60)`

Seuils défaut : 60–140 BPM (`HEARTRATE_MIN_BPM`, `HEARTRATE_MAX_BPM`).
