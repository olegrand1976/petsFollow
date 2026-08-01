# Modèle de données — petsFollow

Source de vérité : migrations `go/internal/platform/db/migrations/` (000001 → 000091+).

## Schémas

| Schéma | Rôle |
|--------|------|
| `identity` | Users, tokens email/reset, OAuth/2FA, locale, payout commercial (`payout_iban`…), zone de base commercial (`base_lat`/`base_lng`/`base_city`/`base_postal_code` — `000058`) |
| `practice` | Cabinets (profil société/banque `000026`), clients liés, invitations, link-requests, `vet_schedule` / vacations (`000024`), import jobs (`000028`) |
| `pets` | Animaux, dossier events, relevés de poids (`weight_readings` — `000060`), tension (`blood_pressure_readings` — `000127`) |
| `heartrate` | Sessions relevé cardiaque |
| `labs` | Panels de prise de sang + résultats analytes (`000128`) — [41](41-TENSION-LABOS.md) |
| `messaging` | Threads, messages (+ media), dispo véto |
| `notifications` | Préférences, log, device tokens |
| `billing` | Entitlements pets (+ addons legacy), Stripe, commissions |
| `sales` | Prospects commerciaux |
| `care` | Rappels, contacts/compétitions horse |
| `visits` | Visites (+ reschedule pending, GPS, `visit_reports`, `preconsult_intakes`) |
| `discovery` | Onboarding client in-app + parcours email (`email_journey`, `email_sends`) |
| `pharmacy` | **Livré (flag)** — CNK, stocks FEFO, DAF + traçabilité `daf_id`/`daf_item_id` sur sorties ([27](27-PHARMACIE-BELGIQUE.md), [28](28-PLAN-STOCK-PEREMPTION.md), migrations `000081`–`000090`) |
| `prescriptions` | **Tag `dev` (flag)** — brouillons de prescriptions + preview PDF ([35](35-PRESCRIPTIONS.md), migration `000091`) ; `date_issued` / signature / PDF persisté = phase 2 |
| `research` | **Tag `dev` (flag)** — événements anonymisés + agrégats hebdo épidémio ([42](42-RESEARCH.md), migrations `000130`/`000131`) ; opt-in `practice.practices.research_opt_in_at` |

## Tables clés

| Domaine | Tables |
|---------|--------|
| Auth | `identity.users` (+ `professional_specialty` pour `care_pro` ; rôle `research` — `000130`), `email_verification_tokens`, `password_reset_tokens`, `identity.profiles` |
| Cabinet | `practice.practices` (+ `research_opt_in_at` / `research_opt_in_by` — `000131`), `practice_clients`, `client_access`, `client_vet_link_requests`, `vet_schedule`, `vet_vacations`, `team_members` (rôles + JSON `permissions` — caps `shares.read`/`shares.manage`, `pharmacy.read`/`pharmacy.write`, etc. via `DefaultTeamPermissions`) |
| Research (dev) | `research.anon_events`, `research.weekly_aggregates`, `research.etl_watermarks` — pas de PII ; `practice_id_hash` interne seulement |
| ACL pets | `pets.pet_access` (partage dossier) — lecture staff = `shares.read`, mutation = `shares.manage` |
| Import | `practice.client_import_jobs`, `client_import_rows` (+ grants `000029`) |
| Animal | `pets.pets`, `pets.dossier_events`, `pets.weight_readings`, `pets.blood_pressure_readings` |
| Billing | `pet_entitlements`, `addon_entitlements`, `stripe_customers`, `stripe_events` |
| Commissions | `commission_tiers`, `commission_ledger`, `commercial_commission_ledger`, payout runs/lines, `commercial_bonus_awards` |
| Commercial | `sales.prospects` (claim `commercial_user_id` nullable = pool libre ; inactivité 30 j) ; assignation commercial ↔ véto ; `manager_user_id` ; RDV / contact — `000031` ; `practice.commercial_referrals` (QR client / nearby) ; `practice.client_referrals` (QR parrainage client→client, `000074`) ; `practice.app_invite_codes` (Code Parrain cabinet+client+véto+care_pro) ; zone base commercial `000058` ; pool admin vétos `assigned_commercial_id IS NULL` |
| FC | `heartrate.sessions` |
| Poids | `pets.weight_readings` (historique) ; `pets.pets.weight_kg` = dernier `POST /weights` (peut diverger si PATCH fiche animal sans lecture) |
| Tension | `pets.blood_pressure_readings` (SYS/DIA, méthode, site) — [41](41-TENSION-LABOS.md) |
| Labos | `labs.panels`, `labs.panel_results` (catalogue analytes V1) — [41](41-TENSION-LABOS.md) |
| Msg | `messaging.threads`, `messages`, `vet_availability` |
| Discovery | `discovery.progress`, `discovery.email_journey`, `discovery.email_sends` |
| Pharmacie | `pharmacy.ref_medications` ; `practice_settings` ; `medication_deposits` ; `medication_batches` (`lot_number` non vide, `expires_on`) ; `stock_movements` (`daf_id` + `daf_item_id` obligatoires si `reason` ∈ {daf, daf_cancel} — `000087`/`000088`) ; `daf_documents` / `daf_items` / `daf_sequences` — API gated `pharmacy.read` / `pharmacy.write` (indépendant de `pets.write_clinical` une fois la clé pharma présente dans l’override ; sinon miroir legacy clinique) |
| Prescriptions (dev) | `prescriptions.prescriptions` — `status` draft\|signed\|sent\|archived ; `medications` JSONB ; `paper_format` A4\|A5 ; `date_issued` NULL en V1 (posé à la signature phase 2) ; `signature_id` / `pdf_*` réservés phase 2 (`000091`) |

## Entitlements

- **Pet** : `plan_code` vendable `monthly` / `annual` / `triennial` (+ `quinquennial` **legacy hors vente**) + `billing_mode` (`one_time` / `subscription` ; monthly = `subscription` only) + statut (`pending` → `active` / `past_due` / …) + `stripe_subscription_id` si sub. Entitlement animal actif ouvre Care / Horse / foyer / kennel ; premium FC + messagerie restent conditionnés au paiement.
- **Addon** (**legacy / plus vendus**) : `family` / `kennel` / `care_plus` / `horse` — table + API conservées pour entitlements existants. Historiquement paiement Stripe unique (`payment`) à vie (`valid_until` NULL) ; `stripe_subscription_id` + handlers `invoice.paid` / `subscription.*` pour lignes sub legacy. Statut `pending` / `active` / `past_due` / `cancelled` / `expired`. Scope owner. Family ≥2 ; Kennel ≥6 ; exclusifs (upgrade Kennel annule Family) ; pets.`litter_tag`.

Migrations utiles : `000019` commissions · `000020` `commercial_bonus_awards` · `000022` kennel / litter_tag / ledger addon · `000023` addon `stripe_subscription_id` + `past_due` · `000024` calendrier véto · `000026` profil payout véto · `000028`/`000029` import clients · `000031` commercial_manager + CRM tracking / directory.

## Notes

- TVA BE 21 % : calcul HTVA côté store (`vat.go`) pour commissions.
- Taux / facteurs : `commission_rates.go` + tiers seed migration `000019`.
- Accrual commission à l’activation checkout ; `invoice.paid` prolonge l’entitlement **sans** re-commission.
- SPIFF commercial (mix triennial only) : `commercial_bonuses.go` (`SyncCommercialBonusAwards`) ; palier véto 31 = affichage seul.
- Détail Stripe → [07-STRIPE-BILLING.md](07-STRIPE-BILLING.md).
