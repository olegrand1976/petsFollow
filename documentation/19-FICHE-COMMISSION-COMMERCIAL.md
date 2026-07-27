# Fiche commissionnement — Commercial

**Scope** : votre rémunération **+** grille véto (pour le pitch co-selling).

## Assiette
- Prix client / Stripe = **TTC**
- Commission = **% du HTVA**
- Assiette = **HT du montant payé**

**Déclenchement** : **une fois** à chaque **nouvelle** activation payante (animal) du cabinet assigné. Pas de re-commission au renouvellement Stripe.

**Inscription sans QR** : le client peut choisir un commercial « près de chez moi » ou un `inviteCode`.  
- **Véto** : seul le **Code Parrain** (`inviteCode`) pose `assigned_commercial_id` à l’inscription. Sans code → pool admin.  
- Client → `commercial_referrals` ; si le véto lié n’a **pas** de commercial assigné, ce referral sert de fallback commission. Un commercial déjà posé sur le véto **gagne toujours**.  
- Chaîne **Comm → Véto → Client** : après rattachement cabinet, `ResolveVetCommercial` paie le commercial du véto ; la row `commercial_referrals` (QR / nearby) reste first-wins et redevient le fallback si le véto est unassign.  
- **QR client parrain** : le filleul peut hériter le commercial de référence du parrain (seed `commercial_referrals`) — toujours en **fallback** seulement ; si le filleul joint un cabinet dont le véto a déjà un commercial assigné, **ce commercial est payé** (`ResolveVetCommercial`). Pas de commission au client promoteur.

## Votre grille
| Offre | Taux HT | € indicatif |
|-------|---------|-------------|
| Monthly 3,50 € | **8 %** | ~0,23 € |
| Annual 35 € | **8 %** | ~2,3 € |
| **Triennial 95 €** | **12 %** | **~9,4 €** |

Steer : triennial = **meilleur taux** et **meilleur €**. Pas de commission addon (plus vendus).

## Bonus SPIFF
| Bonus | Montant | Condition |
|-------|---------|-----------|
| Mix mois | 50 € | ≥ 55 % activations triennial dans le mois |

Détection **automatique** (`SyncCommercialBonusAwards`) ; payout via admin **mark-paid**.

## Grille véto (pour votre pitch)
Progressif 7 → 9 → 11 → 12 % × facteur plan (plafond effectif 8 / 8 / 12 %).  
Sur le triennial au plafond : **~9,4 €** aussi pour le véto.  
**Le véto n’est pas pénalisé** si vous êtes assigné.

## Ne pas compter
- Inscription véto seule (cabinet à 0 payant = normal) → **0 €** de commission
- Revenu = animal qui passe **payant** ; SPIFF mix = ≥ 55 % activations triennial / mois
- Renouvellement d’un abo déjà commissionné → **0 €** supplémentaire
