# Fiche commissionnement — Vétérinaire

**Scope** : rémunération du **prescripteur** uniquement (pas les taux commercial).

## Assiette
- Prix client / Stripe = **TTC**
- Commission = **% du HTVA** (TVA BE 21 %)
- Assiette = **HT du montant payé** (après remise Family −10 % / Kennel −15 % si applicable)

**Déclenchement** : **une fois** à l’activation payante (checkout animal ou addon Family/Kennel). Pas de commission au renouvellement Stripe (plans). Addons = paiement unique à vie.

## Grille progressive (base)
| Clients payants | Taux |
|-----------------|------|
| 1–10 | 7 % |
| 11–30 | 9 % |
| 31–60 | 11 % |
| 61+ | 12 % |

× facteur plan : **triennial ×1** · annual / quinquennial **×0,67**  
→ plafonds effectifs **12 % / 8 % / 8 %**

## € indicatifs (plafond)
| Plan TTC | € max approx. |
|----------|---------------|
| Annual 35 € | ~2,3 € |
| **Triennial 95 €** | **~9,4 €** |
| Quinquennial 145 € | ~9,6 € |

## Bonus
- **50 €** one-shot au 1er passage de **31** clients payants  
  Affichage / progression dans Pro — **payout hors système** (pas d’award ledger).

## Addons (assiette = HT payé)
| Addon | Taux HT | € indicatif |
|-------|---------|-------------|
| Family 39 € | **5 %** | ~1,6 € |
| Kennel 119 € | **5 %** | ~4,9 € |
| Care+ / Horse | **0 %** | — |

Accrual : `AccrueVetForAddon` → ledger `source_kind=addon`.

## Message
Poussez le **triennial** — meilleur taux dès le 1er client. Même plafond avec ou sans commercial assigné.

## Versements
- Les paiements des commissions sont effectués **en début de chaque mois**.
- Profil payout (Paramètres → Paiements & société) requis pour passer une ligne en **prêt à payer** :
  - raison sociale, n° TVA, n° entreprise (BCE), forme juridique
  - adresse de facturation (cabinet ou distincte)
  - IBAN + titulaire (BIC optionnel)
- Statuts ligne visibles par le véto : `accruing` → `missing_info` / `ready_to_pay` → `paid`.
