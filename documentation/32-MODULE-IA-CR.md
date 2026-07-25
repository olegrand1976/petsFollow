# Module CR IA (VetPro)

Add-on cabinet : dictée / transcription / amélioration de comptes-rendus (Gemini).

## Prix (facture externe)

| Offre | Tarif HT |
|-------|----------|
| Mensuel | **39 € / mois / cabinet** |
| Annuel | **390 € / an** (≈ 32,50 €/mois) |

Essai **90 jours** à l’activation (admin ou commercial). **Pas d’intro tarifaire** : la conversion repose sur la preuve ROI.

## Accès

- Gate sur `POST .../report/transcribe` et `.../improve` uniquement.
- CR manuel (lecture / édition / finalize) reste possible sans module.
- `care_pro` : autorisé si le `practice_id` de la visite a un module `trial` valide ou `active`.

## Cycle + machine d’adhésion

1. Activation → statut `trial`, `trial_ends_at = +90j` + email brandé **J0** (`j0_activation`)
2. Tracking `practice.ai_cr_usage_events`
3. Job quotidien (09:00 Brussels) envoie le drip idempotent (`practice.ai_cr_email_sends`) :

| Step | Condition | Contenu |
|------|-----------|---------|
| `j0_activation` | activation | Essai 90 j, CTA 3 CR |
| `j3_nudge` / `j7_nudge` | 0 usage | Relance premier CR |
| `j14_nps` / `j45_nps` | calendaire | Demande NPS (+ digest à J45) |
| `j15_digest` / `j30_digest` | calendaire | Compteurs usage |
| `j60_roi` | ≥ J60 | Preuve temps / € estimé |
| `j75_convert` / `j85_urgency` / `j90_last` | trial | Conversion **39 € / 390 €** |

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
  - expire trials + **drip adhésion** + alertes friction
- Secret Manager : `petsfollow-ai-module-friction-secret` (monté Cloud Run via `deploy-run-args.sh`)
- Scheduler : `./infra/gcp/setup-ai-module-friction-scheduler.sh` (`0 9 * * *` Europe/Brussels)

## Hors scope V1

- Checkout Stripe cabinet
- SPIFF commercial auto
- Playbook public anonymisé
