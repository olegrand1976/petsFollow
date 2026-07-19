# Politique tarifaire petsFollow (offre économique complète)

Positionnement : **logiciel de suivi prescrit par le véto**, sans hardware.  
Monétisation **B2B2C** : Pro gratuit pour le cabinet, client paie.

Objectif : grille **attractive pour véto et commercial**, **concurrentielle** pour le client (~2–3 €/mois), et **marge plateforme** après TVA + Stripe + commissions.

**Statut** : politique **en vigueur** (alignée code / seed) — BM Apporteur / Prescripteur + steer triennial.

---

## 1. Grille (offre complète)

Prix **TTC** client / Stripe. Les commissions partenaires se calculent sur le **HTVA** (TVA BE 21 %).

| Offre | Code | Prix TTC | ≈ / mois | Rôle économique |
|-------|------|----------|----------|-----------------|
| Annual | `annual` | **35 € / an** | 2,9 € | Ancre d’entrée |
| Triennial **recommandé** | `triennial` | **95 € / 3 ans** | 2,6 € | Cœur d’offre & pitch |
| Quinquennial | `quinquennial` | **145 € / 5 ans** | 2,4 € | Engagement long (**paiement unique** — pas de sub Stripe) |
| Family | `family` | **39 € / an** | — | Addon foyer (≥2) ; remise plans **−10 %** ; **pas de plafond** pets |
| Kennel | `kennel` | **119 € / an** | — | Addon élevage (≥6) ; remise plans **−15 %** ; encodage rapide |
| Care+ | `care_plus` | **19 € / an** | — | Upsell soins |
| Horse | `horse` | **39 € / an** | — | Pack équine |

**Exclusivité foyer** : Family et Kennel ne se cumulent pas. Achat Kennel avec Family **active** = upgrade (Family annulé à l’activation Kennel). Family **pending** bloque Kennel (évite double charge).

Modes Stripe : annual / triennial → `one_time` **ou** `subscription` (`year`×1 / `year`×3). Quinquennial → **`one_time` uniquement** (intervalle récurrent Stripe max **3 ans** ; entitlement app = 1825 j).

### Remises engagement (vs 3× / 5× annual)

| Plan | Plein tarif | Prix | Remise |
|------|-------------|------|--------|
| Triennial | 105 € | 95 € | **−10 %** |
| Quinquennial | 175 € | 145 € | **−17 %** |

Règle : remise max multi-ans **≤ 20 %**.

---

## 2. Commissions (partenaires)

Assiette = **HTVA**. Ledger `base_amount_cents` = HT.  
Aucune pénalité véto si commercial assigné.

### Véto — progressif × facteur plan

| Clients payants | Taux de base |
|-----------------|--------------|
| 1–10 | **7 %** |
| 11–30 | **9 %** |
| 31–60 | **11 %** |
| 61+ | **12 %** |

Facteur plan : annual / quinquennial **×0,67** · triennial **×1,00** (plafond effectif 8 / 12 / 8 %).  

| Addon | Taux HT véto | € indicatif |
|-------|--------------|-------------|
| Family 39 € | **5 %** | ~1,6 € |
| Kennel 119 € | **5 %** | ~4,9 € |
| Care+ / Horse | **0 %** | — |

Assiette commission = **HT du montant payé** (après remise foyer/élevage si applicable).

### Commercial — fixe par plan

| Offre | Taux HT | € indicatif |
|-------|---------|-------------|
| Annual | **8 %** | ~2,3 € |
| Triennial | **12 %** | ~**9,4 €** |
| Quinquennial | **8 %** | ~9,6 € |
| Family / Kennel | **10 %** | ~3,2 / **~9,8 €** |
| Care+ / Horse | **10 %** | ~1,6 / 3,2 € |

Fiches pitch : [18 véto](./18-FICHE-COMMISSION-VETO.md) · [19 commercial](./19-FICHE-COMMISSION-COMMERCIAL.md) · [20 admin](./20-FICHE-COMMISSION-ADMIN.md).

### SPIFF (V1 manuelle)

| Bonus | Montant | Condition |
|-------|---------|-----------|
| Ramp cabinet | 25 € | 5 pets / 60 j (commercial) |
| Mix triennial | 50 € / mois | ≥ 55 % activations triennial |
| Palier véto 31 | 50 € | 1er passage 31 clients |

---

## 3. Économie unitaire (pire cas plafond)

Hypothèses : Stripe **1,5 % + 0,25 €** (TTC) ; TVA 21 % sortie ; partners sur HTVA.

| Offre | Net / an | % TTC |
|-------|----------|-------|
| Annual 35 € | ~23,5 € | ~67 % |
| Triennial 95 € | ~**19,3 €** | ~61 % |
| Quinquennial 145 € | ~20,6 € | ~71 % |

### Gardes-fous

| Règle | Seuil |
|-------|--------|
| Marge nette après TVA + Stripe + commissions | **≥ 55 %** du TTC |
| Net annualisé / animal (cœur) | **≥ 17 € / an** |
| Remise max multi-ans | **≤ 20 %** |
| Prix d’entrée | **≤ ~3 € / mois** |
| Take rate max cœur | **≤ 24 %** HT |

**Risque accepté** : triennial avec remise Kennel **−15 %** → net ~**16,4 € / an** sous plafonds commissions (sous le garde-fou 17 €) — documenté et accepté.

---

## 4. Application

| Couche | Source |
|--------|--------|
| Montants TTC | `go/internal/billing/domain.go` (3500 / 9500 / 14500) |
| HTVA | `go/internal/store/vat.go` |
| Taux / facteurs | `go/internal/store/commission_rates.go` |
| Tiers seed | migration `000019` + `DefaultVetCommissionTiers` |
| Fiches UI | `ProCommissionSheet` (vet / commercial / admin) |
| Stripe | Prices `STRIPE_PRICE_*` : plans 35 / 95 / 145 · addons Family 39 / Kennel 119 / Care+ 19 / Horse 39 |
