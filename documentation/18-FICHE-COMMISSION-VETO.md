# Fiche commissionnement — Vétérinaire

**Scope** : rémunération du **prescripteur** uniquement (pas les taux commercial).

## Assiette
- Prix client / Stripe = **TTC**
- Commission = **% du HTVA** (TVA BE 21 %)
- Assiette = **HT du montant payé**

**Déclenchement** : **une fois** à l’activation payante (checkout animal). Pas de commission au renouvellement Stripe.

## Grille progressive (base)
| Clients payants | Taux |
|-----------------|------|
| 1–10 | 7 % |
| 11–30 | 9 % |
| 31–60 | 11 % |
| 61+ | 12 % |

× facteur plan : **triennial ×1** · monthly / annual **×0,67**  
→ plafonds effectifs **12 % / 8 % / 8 %**

## € indicatifs (plafond)
| Plan TTC | € max approx. |
|----------|---------------|
| Monthly 3,50 € | ~0,23 € |
| Annual 35 € | ~2,3 € |
| **Triennial 95 €** | **~9,4 €** |

## Bonus
- **50 €** one-shot au 1er passage de **31** clients payants  
  Affichage / progression dans Pro — **payout hors système** (pas d’award ledger).

## Message
Poussez le **triennial** — meilleur taux dès le 1er client. Même plafond avec ou sans commercial assigné.  
Pas de commission addon (addons plus vendus ; features incluses avec l’entitlement animal).

## Versements
- Les paiements des commissions sont effectués **en début de chaque mois**.
- Profil payout (Paramètres → Paiements & société) requis pour passer une ligne en **prêt à payer** :
  - raison sociale, n° TVA, n° entreprise (BCE), forme juridique
  - adresse de facturation (cabinet ou distincte)
  - IBAN + titulaire (BIC optionnel)
- Statuts ligne visibles par le véto : `accruing` → `missing_info` / `ready_to_pay` → `paid`.
