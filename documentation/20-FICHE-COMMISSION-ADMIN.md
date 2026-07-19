# Fiche commissionnement — Admin (vue d’ensemble)

**Scope** : modèle complet Apporteur / Prescripteur.

## Prix TTC
35 / 95 / 145 € (annual / triennial / quinquennial) · addons Family 39 · Kennel 119 · Care+ 19 · Horse 39 €

## Assiette
Commissions sur **HTVA** (`DefaultVATRateBps = 2100`). Stripe Prices = TTC.  
Assiette = **HT du montant payé** (après remise foyer/élevage si applicable).

## Véto
- Tiers : 7 / 9 / 11 / 12 % (1–10 / 11–30 / 31–60 / 61+)
- Facteur plan : annual & quin ×0,67 · triennial ×1
- Addons : Family / Kennel **5 %** · Care+ / Horse **0 %**
- SPIFF : 50 € @ 31 clients (manuel V1)

## Commercial
- Plans : 8 / 12 / 8 % · addons **10 %** (Family / Kennel / Care+ / Horse)
- SPIFF : ramp 25 € · mix 50 €/mois (manuel V1)
- Pas de commission sur inscription véto seule

## Gardes-fous
| Règle | Seuil |
|-------|--------|
| Marge nette | ≥ 55 % TTC |
| Net/an cœur | ≥ 17 € |
| Take rate max cœur | ≤ 24 % HT |
| Remise multi-ans | ≤ 20 % |

## Code
- `go/internal/store/commission_rates.go` — taux plans / addons
- `go/internal/store/commissions.go` — `AccrueCommercialForAddon`, `AccrueVetForAddon`
- `go/internal/store/vat.go`
- migration `000019_commission_plan_rates` · ledger addon `000022_kennel_litter_tag` · addon sub `000023_addon_subscription`
- UI : `ProCommissionSheet` audience `admin`
