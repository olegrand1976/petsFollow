# Fiche commissionnement — Vétérinaire

**Scope** : rémunération du **prescripteur** uniquement (pas les taux commercial).

## Assiette
- Prix client / Stripe = **TTC**
- Commission = **% du HTVA** (TVA BE 21 %)
- Assiette = **HT du montant payé**

**Déclenchement** : **une fois** à l’activation payante (checkout animal). Pas de commission au renouvellement Stripe.

**Attribution** : le véto rétribué est celui lié via `practice.practice_clients` (dernier lien cabinet du owner). Les invitations app (QR / email `/invite/{code}`) auto-rattachent le nouveau client au véto émetteur à l’inscription ou au claim deep link.

**Care pro** : taux configurables dans Admin → Rétributions (`billing.profile_commission_rates`, clés `care_pro.*`). Défaut **0 %** — pas de ligne tant que le taux reste à 0. Dès qu'un taux > 0 est défini, l'activation d'un plan animal écrit une ligne `source_kind='care_pro'` dans `billing.commission_ledger` (base HT × taux, une ligne par care_pro lié au foyer, période ouverte) — les payouts suivent le circuit runs/lines existant. Claim QR care_pro → grant `client_access` (`write_notes`).

**Commercial** : QR invite enregistre `practice.commercial_referrals` ; la commission commercial reste basée sur le véto assigné (inchangé).

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
