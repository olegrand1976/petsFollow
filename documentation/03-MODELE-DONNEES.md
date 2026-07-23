# Modèle de données — petsFollow

Source de vérité : migrations `go/internal/platform/db/migrations/` (000001 → 000031+).

## Schémas

| Schéma | Rôle |
|--------|------|
| `identity` | Users, tokens email/reset, OAuth/2FA, locale, payout commercial (`payout_iban`…) |
| `practice` | Cabinets (profil société/banque `000026`), clients liés, invitations, link-requests, `vet_schedule` / vacations (`000024`), import jobs (`000028`) |
| `pets` | Animaux, dossier events |
| `heartrate` | Sessions relevé cardiaque |
| `messaging` | Threads, messages (+ media), dispo véto |
| `notifications` | Préférences, log, device tokens |
| `billing` | Entitlements pets/addons, Stripe, commissions |
| `sales` | Prospects commerciaux |
| `care` | Rappels, contacts/compétitions horse |
| `visits` | Visites (+ reschedule pending) |
| `discovery` | Onboarding client in-app + parcours email (`email_journey`, `email_sends`) |
| `pharmacy` | **Spec** — référentiel CNK, stocks FEFO, DAF, jobs audit ([27](27-PHARMACIE-BELGIQUE.md), migrations `000039+` à venir) |

## Tables clés

| Domaine | Tables |
|---------|--------|
| Auth | `identity.users`, `email_verification_tokens`, `password_reset_tokens` |
| Cabinet | `practice.practices`, `practice_clients`, `client_vet_link_requests`, `vet_schedule`, `vet_vacations` |
| Import | `practice.client_import_jobs`, `client_import_rows` (+ grants `000029`) |
| Animal | `pets.pets`, `pets.dossier_events` |
| Billing | `pet_entitlements`, `addon_entitlements`, `stripe_customers`, `stripe_events` |
| Commissions | `commission_tiers`, `commission_ledger`, `commercial_commission_ledger`, payout runs/lines, `commercial_bonus_awards` |
| Commercial | `sales.prospects` (+ assignation commercial ↔ véto, `manager_user_id`, source `directory`, RDV / contact timestamps — `000031`) |
| FC | `heartrate.sessions` |
| Msg | `messaging.threads`, `messages`, `vet_availability` |
| Discovery | `discovery.progress`, `discovery.email_journey`, `discovery.email_sends` |

## Entitlements

- **Pet** : plan `annual` / `triennial` / `quinquennial` + `billing_mode` (`one_time` / `subscription`) + statut (`pending` → `active` / `past_due` / …) + `stripe_subscription_id` si sub.
- **Addon** : `family` / `kennel` / `care_plus` / `horse` — **paiement Stripe unique** (`payment`) **à vie** (`valid_until` NULL). Colonne `stripe_subscription_id` conservée pour les abonnements legacy ; handlers `invoice.paid` / `subscription.*` restent pour ces lignes. Statut `pending` / `active` / `past_due` / `cancelled` / `expired`. Scope owner. Family ≥2 (pas de plafond) ; Kennel ≥6 ; exclusifs (upgrade Kennel annule Family) ; pets.`litter_tag` (élevage).

Migrations utiles : `000019` commissions · `000020` `commercial_bonus_awards` · `000022` kennel / litter_tag / ledger addon · `000023` addon `stripe_subscription_id` + `past_due` · `000024` calendrier véto · `000026` profil payout véto · `000028`/`000029` import clients · `000031` commercial_manager + CRM tracking / directory.

## Notes

- TVA BE 21 % : calcul HTVA côté store (`vat.go`) pour commissions.
- Taux / facteurs : `commission_rates.go` + tiers seed migration `000019`.
- Accrual commission à l’activation checkout ; `invoice.paid` prolonge l’entitlement **sans** re-commission.
- SPIFF commercial : `commercial_bonuses.go` (`SyncCommercialBonusAwards`) ; palier véto 31 = affichage seul.
- Détail Stripe → [07-STRIPE-BILLING.md](07-STRIPE-BILLING.md).
