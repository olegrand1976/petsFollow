# Modèle de données — petsFollow

Source de vérité : migrations `go/internal/platform/db/migrations/` (000001 → 000019+).

## Schémas

| Schéma | Rôle |
|--------|------|
| `identity` | Users, tokens email/reset, OAuth/2FA, locale |
| `practice` | Cabinets, clients liés, invitations, link-requests |
| `pets` | Animaux, dossier events |
| `heartrate` | Sessions relevé cardiaque |
| `messaging` | Threads, messages (+ media), dispo véto |
| `notifications` | Préférences, log, device tokens |
| `billing` | Entitlements pets/addons, Stripe, commissions |
| `sales` | Prospects commerciaux |
| `care` | Rappels, contacts/compétitions horse |
| `visits` | Visites |
| `discovery` | Onboarding client |

## Tables clés

| Domaine | Tables |
|---------|--------|
| Auth | `identity.users`, `email_verification_tokens`, `password_reset_tokens` |
| Cabinet | `practice.practices`, `practice_clients`, `client_vet_link_requests` |
| Animal | `pets.pets`, `pets.dossier_events` |
| Billing | `pet_entitlements`, `addon_entitlements`, `stripe_customers`, `stripe_events` |
| Commissions | `commission_tiers`, `commission_ledger`, `commercial_commission_ledger`, payout runs/lines |
| Commercial | `sales.prospects` (+ assignation commercial ↔ véto) |
| FC | `heartrate.sessions` |
| Msg | `messaging.threads`, `messages`, `vet_availability` |

## Entitlements

- **Pet** : plan `annual` / `triennial` / `quinquennial` + statut (`pending` → `active` …).
- **Addon** : `family` / `kennel` / `care_plus` / `horse` (durée 365 j), scope owner. Family ≥2 (pas de plafond) ; Kennel ≥6 ; pets.`litter_tag` (élevage).

## Notes

- TVA BE 21 % : calcul HTVA côté store (`vat.go`) pour commissions.
- Taux / facteurs : `commission_rates.go` + tiers seed migration `000019`.
