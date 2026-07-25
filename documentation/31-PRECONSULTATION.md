# Pré-consultation — petsFollow

Formulaire client d’état de l’animal **après confirmation** d’un RDV (`visits.visits` → `confirmed`).

## Flux

```text
ConfirmVisit / ConfirmDirect / AcceptReschedule (→ confirmed)
  → EnsurePreconsultPending (idempotent)
  → Push FCM visit_confirmed (+ deep link pré-consult)
  → Email client (si intake nouvellement créé + pref visits)
       CTA = /invite/{code}?preconsult={visitId}
       Deep link app = petsfollow://preconsult?visitId=
  → Client remplit l’app (PUT)
  → Véto lit dans l’agenda Pro (GET)
```

## Schéma JSON `answers`

| Champ | Type | Valeurs |
|-------|------|---------|
| `chiefComplaint` | string | motif / plainte (requis à la soumission) |
| `duration` | enum | `today` · `few_days` · `week` · `weeks` · `months` · `unknown` |
| `behavior` | enum | `normal` · `lethargic` · `restless` · `aggressive` · `anxious` · `other` · `unknown` |
| `appetite` | enum | `normal` · `decreased` · `increased` · `unknown` |
| `thirst` | enum | idem appetite |
| `elimination` | enum | `normal` · `decreased` · `increased` · `unknown` |
| `urgency` | enum | `low` · `medium` · `high` |
| `comment` | string | libre (optionnel) |

Statuts intake : `pending` → `submitted` (ou `skipped` réservé).

## API

| Méthode | Route | Qui |
|---------|-------|-----|
| `GET` | `/visits/{visitID}/preconsult` | owner client · staff `calendar.manage` · care_pro lecture pet |
| `PUT` | `/visits/{visitID}/preconsult` | owner client, visite `confirmed`, pas encore `done` |

## Prefs / RGPD

- Respecte `client_preferences.visits` (même flag que le push RDV).
- Export : clé `preconsultIntakes` dans `GET /me/export`.
- Purge : cascade `pets` → `visits` → `preconsult_intakes`.
- Deep link hors session : `visitId` persisté (`PreconsultVisitStore`) jusqu’à ouverture post-login.
- CTA e-mail : invite cabinet (`/invite/{code}?preconsult=`) via référence véto ou 1er véto du cabinet ; fallback `petsfollow://preconsult?visitId=`.

## Hors scope MVP

Guest sans compte, SMS, rappel J-1, templates par espèce, scoring IA.
