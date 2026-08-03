# Politique tarifaire petsFollow

Positionnement : **continuité de soins prescrite** + **passeport digital** de l’animal — Web Pro · mobile ProLight / care pro · mobile Client. Le relevé respiratoire est une feature.  
Monétisation **double** :

1. **SaaS cabinet Pro** — facturation **externe** (pas de Stripe)  
2. **B2B2C client** — abonnements animal (Stripe)

**Pro Light** (app mobile **ProLight**, profil pro terrain — avec ou sans compte Web Pro) : **gratuit** pour les professionnels.

Objectif : grille **attractive pour véto et commercial**, **concurrentielle** pour le client (~2–3,5 €/mois), et **marge plateforme** après TVA + Stripe + commissions.

**Statut** : politique **en vigueur** (alignée code / seed / pages offre) — BM Apporteur / Prescripteur + steer triennial + SaaS Pro hors ligne.

---

## 0. Offre professionnelle (cabinet)

| Offre | Surface | Tarif | Paiement |
|-------|---------|-------|----------|
| **Pro** | Web SaaS + app clients | **834,71 € HTVA / an** (+ setup **320 € HTVA** one-shot) · **2 253,72 € HTVA / 3 ans** (−10 %) | **Facturation externe** (commercial / compta) — **pas** de checkout Stripe cabinet |
| **Pro Light** | App mobile ProLight (Flutter) | **0 €** | — |

Engagements Pro (facturation externe) :

| Prestation | Tarif HTVA | Notes |
|------------|------------|-------|
| Mise en place & formation | **320 €** | One-shot |
| Annuel | **834,71 € / an** | ≈ 69,56 € / mois — **CR IA inclus** |
| Triennal (recommandé) | **2 253,72 € / 3 ans** | −10 % vs 3× annuel (économie **250,41 €** · ≈ 62,60 €/mois) — **CR IA inclus** |
| Migration données | Sur devis (≥ **350 €**) | Option |
| **CR IA** | **Inclus** dans Pro | Dictée / improve — plus d’add-on 39/390 € — voir `documentation/32-MODULE-IA-CR.md` |

Les commissions partenaires (activations clients) peuvent **compenser** la facture SaaS hors ligne (ordre de grandeur : ~8 activations triennales / mois ≈ couverture de l’équivalent mensuel ≈ 69,56 € HTVA).

---

## 1. Grille client (app)

Prix **TTC** client / Stripe. Les commissions partenaires se calculent sur le **HTVA** (TVA BE 21 %).

| Offre | Code | Prix TTC | ≈ / mois | Rôle | Modes Stripe |
|-------|------|----------|----------|------|--------------|
| Monthly | `monthly` | **3,50 € / mois** | 3,5 € | Flexibilité | **`subscription` only** (`month`×1) |
| Annual | `annual` | **35 € / an** | 2,9 € | Ancre d’entrée | `one_time` **ou** `subscription` (`year`×1) |
| Triennial **recommandé** | `triennial` | **95 € / 3 ans** | 2,6 € | Cœur d’offre & pitch | `one_time` **ou** `subscription` (`year`×3) |

**Hors vente** : quinquennial (`quinquennial`, legacy) · addons Family / Kennel / Care+ / Horse (plus vendus).

**Inclus dès entitlement animal actif** : foyer, encodage élevage, Care (rappels), Horse (maréchal / contacts / compétitions). Relevés HR premium + messagerie restent liés à l’entitlement payant.

Mensuel ≈ **42 € / an** (> annuel) — flexibilité, pas le meilleur prix.

### Remise engagement (vs 3× annual)

| Plan | Plein tarif | Prix | Remise |
|------|-------------|------|--------|
| Triennial | 105 € | 95 € | **−10 %** |

Règle : remise max multi-ans **≤ 20 %**.

---

## 2. Commissions (partenaires)

Assiette = **HTVA**. Ledger `base_amount_cents` = HT.  
Aucune pénalité véto si commercial assigné.

### Véto — progressif × facteur plan

| Clients payants | Taux de base |
|-----------------|--------------|
| 1–10 | **7,50 %** |
| 11–30 | **7,67 %** |
| 31–60 | **7,83 %** |
| 61+ | **8 %** |

Facteur plan : monthly / annual **×0,67** · triennial **×1,00**, puis borne effective **5–8 %** (plafond effectif ~5,4 / ~5,4 / 8 %).

### Commercial — fixe par plan

| Offre | Taux HT | € indicatif |
|-------|---------|-------------|
| Monthly 3,50 € | **8 %** | ~0,23 € |
| Annual 35 € | **8 %** | ~2,3 € |
| Triennial 95 € | **12 %** | ~**9,4 €** |

**Déclenchement** : commission accrétée **une fois** à l’activation payante (`checkout.session.completed` → animal). Idempotent par entitlement. Renouvellements Stripe (`invoice.paid`) prolongent l’abo **sans** nouvelle ligne ledger.

Fiches pitch : [18 véto](./18-FICHE-COMMISSION-VETO.md) · [19 commercial](./19-FICHE-COMMISSION-COMMERCIAL.md) · [20 admin](./20-FICHE-COMMISSION-ADMIN.md).

### SPIFF

| Bonus | Montant | Condition | Statut technique |
|-------|---------|-----------|------------------|
| Mix triennial | 50 € / mois | ≥ 55 % activations triennial | Détection auto (`SyncCommercialBonusAwards`) + mark-paid admin |
| Palier véto 31 | 50 € | 1er passage 31 clients payants | Affichage / progression seule — payout hors système |

---

## 3. Économie unitaire (pire cas plafond)

Hypothèses : Stripe **1,5 % + 0,25 €** (TTC) ; TVA 21 % sortie ; partners sur HTVA.

| Offre | Net / an (approx.) | % TTC |
|-------|--------------------|-------|
| Monthly 3,50 € ×12 | ~29 € | ~69 % |
| Annual 35 € | ~24,3 € | ~69 % |
| Triennial 95 € | ~**20,4 €** | ~64 % |

### Gardes-fous

| Règle | Seuil |
|-------|--------|
| Marge nette après TVA + Stripe + commissions | **≥ 55 %** du TTC |
| Net annualisé / animal (cœur) | **≥ 17 € / an** |
| Remise max multi-ans | **≤ 20 %** |
| Prix d’entrée | **≤ ~3,5 € / mois** |
| Take rate max cœur | **≤ 20 %** HT |

---

## 4. Application

| Couche | Source |
|--------|--------|
| Montants TTC | `go/internal/billing/domain.go` (350 / 3500 / 9500) |
| HTVA | `go/internal/store/vat.go` |
| Taux / facteurs | `go/internal/store/commission_rates.go` |
| Tiers seed | migration `000140` + `DefaultVetCommissionTiers` |
| SPIFF commercial | `go/internal/store/commercial_bonuses.go` · mig `000020` · UI `/admin/commercial-bonuses` |
| Fiches UI | `ProCommissionSheet` (vet / commercial / admin) |
| Stripe | Prices `STRIPE_PRICE_*` : monthly **sub** · annual / triennial one_time+sub |
