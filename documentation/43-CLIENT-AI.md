# 43 — Client AI (vulgarisation CR + triage 24/7)

Module **tag `dev`** : assistants IA pour l’app Flutter **client** (propriétaire). Distinct du module CR IA Pro ([`32-MODULE-IA-CR.md`](32-MODULE-IA-CR.md)).

## Surfaces

| Feature | Entrée Flutter | API Go |
|---------|----------------|--------|
| Vulgarisation CR | `ConsultationViewScreen` → « Comprendre mon compte-rendu » | `GET /api/v1/visits/{id}/client-consultation/explain` |
| Triage 24/7 | Home + Settings | `POST/GET /api/v1/client-ai/triage/sessions…` |

## Flags

| Couche | Variable |
|--------|----------|
| API | `CLIENT_AI_ENABLED` (défaut off ; `make api-dev` → true) |
| Flutter | `--dart-define=CLIENT_AI_ENABLED` (`make flutter-dev` → true) |

Endpoints → **404** `client_ai_disabled` si flag off.

## Modèle & coûts

- `GEMINI_LITE_MODEL` via `go/internal/platform/gemini`
- Cache DB pour les explications CR (`visits.visit_report_explanations`)
- Rate-limit triage : 20 messages / user / heure ; max 20 messages / session (insert atomique user+assistant)
- Rate-limit explain (génération Gemini) : 10 / user / heure (les hits cache ne comptent pas)
- Refresh explain : 1× / jour / visite (`?refresh=1`)

## Sécurité produit

- L’IA **ne modifie jamais** le diagnostic / le texte du CR (explication additive).
- Triage = pré-évaluation, **pas** un diagnostic ; pas de posologie / médicament conseillé.
- Escalade Rouge : téléphone du **cabinet rattaché** + messagerie + prise de RDV — **pas** d’annuaire de garde.
- Disclaimer UI systématique.

## RGPD

- Export : `visitReportExplanations`, `clientAiTriageSessions`, `clientAiUsageEvents`
- Purge : cascade pets → explanations ; `DELETE` sessions triage par `user_id` dans `purgeClientOwnedDataExec`
- Sous-traitant Gemini (privacy Flutter + [`36-RGPD.md`](36-RGPD.md))

## Hors scope MVP

Annuaire de garde, surface Nuxt, useCase commercial, facturation client de cette IA.
