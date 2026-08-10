# Module CR IA (VetPro)

Dictée / transcription / amélioration de comptes-rendus (Gemini) — **inclus nativement dans l’offre Pro** (plus d’add-on tarifé).

## Prix

| Offre | Tarif |
|-------|-------|
| CR IA | **Inclus** dans le SaaS Pro (834,71 € HTVA / an ou 2 253,72 € / 3 ans + setup 320 €) |

Pas de SKU 39 € / 390 €. Kill-switch admin (`disabled`) uniquement en ops.

## Accès

- Gate sur `POST .../report/transcribe` et `.../improve` : autorisé pour tout cabinet Pro **sauf** statut `disabled`.
- Absent de ligne `practice.ai_cr_modules` → **autorisé** (inclus par défaut).
- CR manuel (lecture / édition / finalize) reste toujours possible.
- `care_pro` : autorisé si le `practice_id` de la visite n’est pas `disabled`.

## Suivi adoption + ROI (sans conversion payante)

1. Activation suivi (admin / commercial) → row tracking statut **`active`** + email brandé **J0** (`j0_activation`) — optionnel pour le ROI dashboard (pas un essai payant).
2. Tracking `practice.ai_cr_usage_events`
3. Job quotidien (09:00 Brussels) envoie le drip **adoption** idempotent (`practice.ai_cr_email_sends`) aux modules `trial` **ou** `active` :

| Step | Condition | Contenu |
|------|-----------|---------|
| `j0_activation` | activation suivi | CTA premiers CR |
| `j3_nudge` / `j7_nudge` | 0 usage | Relance premier CR |
| `j14_nps` / `j45_nps` | calendaire | Demande NPS (+ digest à J45) |
| `j15_digest` / `j30_digest` | calendaire | Compteurs usage |
| `j60_roi` | ≥ J60 | Preuve temps / € estimé |

Les steps `j75_convert` / `j85_urgency` / `j90_last` (conversion payante) sont **retirés**.

4. ROI affiché dès **J60** (temps gagné estimé sur CR **finalisés** ayant eu transcribe/improve)
5. Feedback NPS in-app (tags friction) + alertes friction → commercial assigné

## Surfaces

| Rôle | URL |
|------|-----|
| Admin | `/admin/ai-modules` |
| Commercial | `/commercial/ai-modules` + guide `/commercial/ai-cr-playbook` |
| VetPro | widget dashboard (ROI, NPS, tags) |

## Ops

- Migrations `000064_ai_cr_module`, `000065_ai_cr_adhesion`
- Job unique : `POST /api/v1/internal/ai-module-friction/run` + header `X-Ai-Module-Friction-Secret` (`AI_MODULE_FRICTION_SECRET`)
  - expire trials legacy + **drip adoption** + alertes friction
- Secret Manager : `petsfollow-ai-module-friction-secret` (monté Cloud Run via `deploy-run-args.sh`)
- Scheduler : `./infra/gcp/setup-ai-module-friction-scheduler.sh` (`0 9 * * *` Europe/Brussels)

## Hors scope

- Checkout Stripe cabinet
- SPIFF commercial auto sur « convert IA »
- Playbook public anonymisé
- CR IA **avancé** multi-agents + RAG — voir [`44-AI-CR-ADVANCED.md`](44-AI-CR-ADVANCED.md)
