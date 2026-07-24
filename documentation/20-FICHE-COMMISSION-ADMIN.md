# Fiche commissionnement — Admin (vue d’ensemble)

**Scope** : modèle complet Apporteur / Prescripteur.

## Prix TTC
**3,50 € / mois** · **35 € / an** · **95 € / 3 ans** (monthly / annual / triennial).  
Hors vente : quinquennial · addons Family / Kennel / Care+ / Horse.

## Assiette
Commissions sur **HTVA** (`DefaultVATRateBps = 2100`). Stripe Prices = TTC.  
Assiette = **HT du montant payé**.

**Déclenchement** : accrual **une fois** à l’activation payante (checkout animal). Pas de re-commission au renouvellement Stripe.

## Véto
- Tiers : 7 / 9 / 11 / 12 % (1–10 / 11–30 / 31–60 / 61+) — éditables admin (`PUT /admin/commissions/tiers`)
- Facteur plan : monthly & annual ×0,67 · triennial ×1
- Pas de commission addon (plus vendus)
- SPIFF : 50 € @ 31 clients — **affichage / progression seule** (pas d’award DB ; payout hors système)
- Payouts mensuels (versement **début de mois**) :
  - run : `open` → `closed` → `partially_paid` → `paid`
  - lignes : `accruing` (preview) → `missing_info` / `ready_to_pay` (à la clôture) → `paid`
  - profil société/banque sur `practice.practices` (migration `000026`)
  - admin : mark-paid **ligne** ou **bulk des lignes prêtes** (`/admin/commissions/…`)

## Commercial
- Plans : **8 / 8 / 12 %** (monthly / annual / triennial) — **constantes code** (pas éditables ; `PUT /admin/commissions/settings` rejette)
- SPIFF ramp 25 € · mix 50 €/mois : **détection auto** + mark-paid admin (`/admin/commercial-bonuses`)
- Pas de commission sur inscription véto seule (ni sans commercial assigné)
- Payouts : miroir `/admin/commercial-commissions/…` + profil IBAN commercial

## Gardes-fous
| Règle | Seuil |
|-------|--------|
| Marge nette | ≥ 55 % TTC |
| Net/an cœur | ≥ 17 € |
| Take rate max cœur | ≤ 24 % HT |
| Remise multi-ans | ≤ 20 % |

## Code
- `go/internal/store/commission_rates.go` — taux plans
- `go/internal/store/commissions.go` — accrual animal (legacy addon helpers éventuels)
- `go/internal/store/commercial_bonuses.go` — `SyncCommercialBonusAwards`, mark-paid
- `go/internal/store/vat.go`
- migrations : `000019_commission_plan_rates` · `000020` `commercial_bonus_awards` · vet payout profile `000026`
- UI : `ProCommissionSheet` audience `admin` · pages `/admin/commissions`, `/admin/commercial-commissions`, `/admin/commercial-bonuses`
