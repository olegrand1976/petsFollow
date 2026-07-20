# Modèle de données — petsFollow

Source de vérité : migrations `go/internal/platform/db/migrations/` (000001 → 000023+).

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
| Commissions | `commission_tiers`, `commission_ledger`, `commercial_commission_ledger`, payout runs/lines, `commercial_bonus_awards` |
| Commercial | `sales.prospects` (+ assignation commercial ↔ véto) |
| FC | `heartrate.sessions` |
| Msg | `messaging.threads`, `messages`, `vet_availability` |

## Entitlements

- **Pet** : plan `annual` / `triennial` / `quinquennial` + `billing_mode` (`one_time` / `subscription`) + statut (`pending` → `active` / `past_due` / …) + `stripe_subscription_id` si sub.
- **Addon** : `family` / `kennel` / `care_plus` / `horse` — **abonnement Stripe annuel récurrent** (`subscription` `year`×1), `valid_until` ~365 j renouvelé via `invoice.paid`, `stripe_subscription_id`, statut `pending` / `active` / `past_due` / `cancelled` / `expired`. Scope owner. Family ≥2 (pas de plafond) ; Kennel ≥6 ; exclusifs (upgrade Kennel annule Family) ; pets.`litter_tag` (élevage).

Migrations billing utiles : `000019` commissions · `000020` `commercial_bonus_awards` · `000022` kennel / litter_tag / ledger addon · `000023` addon `stripe_subscription_id` + `past_due`.

## Notes

- TVA BE 21 % : calcul HTVA côté store (`vat.go`) pour commissions.
- Taux / facteurs : `commission_rates.go` + tiers seed migration `000019`.
- Accrual commission à l’activation checkout ; `invoice.paid` prolonge l’entitlement **sans** re-commission.
- SPIFF commercial : `commercial_bonuses.go` (`SyncCommercialBonusAwards`) ; palier véto 31 = affichage seul.
- Détail Stripe → [07-STRIPE-BILLING.md](07-STRIPE-BILLING.md).
