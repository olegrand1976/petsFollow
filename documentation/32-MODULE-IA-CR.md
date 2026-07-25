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

## Cycle

1. Activation → statut `trial`, `trial_ends_at = +90j`
2. Tracking `practice.ai_cr_usage_events`
3. ROI affiché dès **J60** (temps gagné estimé sur CR **finalisés** ayant eu transcribe/improve)
4. J75–J90 : conversion `monthly_39` / `annual_390` ou expiration
5. Feedback NPS + job friction → alerte commercial assigné

## Surfaces

| Rôle | URL |
|------|-----|
| Admin | `/admin/ai-modules` |
| Commercial | `/commercial/ai-modules` + guide `/commercial/ai-cr-playbook` |
| VetPro | widget dashboard |

## Ops

- Migration `000064_ai_cr_module`
- Job : `POST /api/v1/internal/ai-module-friction/run` + header `X-Ai-Module-Friction-Secret` (`AI_MODULE_FRICTION_SECRET`)
