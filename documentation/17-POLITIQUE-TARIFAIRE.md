# Politique tarifaire petsFollow (offre économique complète)

Positionnement : **logiciel de suivi prescrit par le véto**, sans hardware.  
Monétisation **B2B2C** : Pro gratuit pour le cabinet, client paie.

Objectif : grille **attractive pour véto et commercial**, **concurrentielle** pour le client (~2–3 €/mois), et **marge plateforme** après Stripe + commissions.

**Statut** : politique **en vigueur** (alignée code / seed).

---

## 1. Grille (offre complète)

| Offre | Code | Prix | ≈ / mois | Rôle économique |
|-------|------|------|----------|-----------------|
| Annual | `annual` | **29 € / an** | 2,4 € | Ancre d’entrée |
| Triennial **recommandé** | `triennial` | **75 € / 3 ans** | 2,1 € | Cœur d’offre & pitch commercial |
| Quinquennial | `quinquennial` | **115 € / 5 ans** | 1,9 € | Engagement long sans dumping |
| Family | `family` | **55 € / an** | — | Addon foyer (compte, ≤3 animaux) |
| Care+ | `care_plus` | **19 € / an** | — | Upsell soins / médicaments (compte) |
| Horse | `horse` | **39 € / an** | — | Pack équine (compte, pets horse) |

Modes : `one_time` ou `subscription` (interval 1 / 3 / 5 ans).

### Remises engagement (vs 3× / 5× annual)

| Plan | Plein tarif | Prix | Remise |
|------|-------------|------|--------|
| Triennial | 87 € | 75 € | **−14 %** |
| Quinquennial | 145 € | 115 € | **−21 %** |

Règle : ne jamais descendre sous **≈ −25 %** sur le 5 ans.

### Family — règle produit

- Scope **owner** (compte) : pack foyer jusqu’à **3 animaux**.
- **Addon** (complément), pas un substitut des abonnements animal.
- Copy UI : « Forfait foyer — jusqu’à 3 animaux », jamais « abo illimité ».

### Care+ — privilèges

Sans Care+ : rappels standards seedés (vaccin, vermifuge, contrôle, dentaire).

Avec Care+ : rappels **médicaments**, rappels **personnalisés**, historique / export carnet, relances email J-3 / J0.

### Horse — privilèges

Sans pack (même si `species=horse`) : fiche + relève + rappels généraux uniquement.

Avec Horse : rappels maréchal / coproscopie, carnet maréchal, **contacts pro**, **compétitions**.  
Promo in-app **uniquement** si le foyer a ≥1 cheval.

---

## 2. Commissions (partenaires)

| Acteur | Règle | Commentaire |
|--------|--------|-------------|
| **Véto** | Progressif **5 % → 8 % → 10 % → 12 %** (volume clients) | Plafond **12 %** — éditable admin |
| **Commercial** | **12 % fixe** sur abonnements **et** addons | Plus de miroir — éditable admin |
| **Véto addons** | Aucune | Inchangé |

### Commission absolue (plafond véto 12 % + commercial 12 %)

| Offre | Prix | Comm. véto | Comm. commercial |
|-------|------|------------|------------------|
| Annual | 29 € | **3,48 €** | **3,48 €** |
| Triennial | 75 € | **9,00 €** | **9,00 €** |
| Quinquennial | 115 € | **13,80 €** | **13,80 €** |
| Family | 55 € | — | **6,60 €** |
| Care+ | 19 € | — | **2,28 €** |
| Horse | 39 € | — | **4,68 €** |

**Pitch commercial** : vendre le **triennial** (~**9 €** / animal @12 %) plutôt que l’annual (~3,5 €).  
**Pitch véto** : au plafond 12 %, même € que le commercial sur les abos.

---

## 3. Économie unitaire plateforme

Hypothèses Stripe EU approx. : **1,5 % + 0,25 €**.

### Au pire (véto 12 % + commercial 12 %)

| Offre | Brut | Stripe | Partenaires | **Net plateforme** | % net | Net / an |
|-------|------|--------|-------------|---------------------|-------|----------|
| Annual 29 € | 29 | 0,69 | 6,96 | **21,35 €** | 74 % | 21,35 € |
| Triennial 75 € | 75 | 1,38 | 18,00 | **55,62 €** | 74 % | **18,54 €** |
| Quinquennial 115 € | 115 | 1,98 | 27,60 | **85,42 €** | 74 % | **17,08 €** |

### Addons (commercial 12 % seulement)

| Addon | Brut | Stripe | Comm. | Net | % net |
|-------|------|--------|-------|-----|-------|
| Family 55 € | 55 | 1,08 | 6,60 | **47,3 €** | 86 % |
| Care+ 19 € | 19 | 0,54 | 2,28 | **16,2 €** | 85 % |
| Horse 39 € | 39 | 0,84 | 4,68 | **33,5 €** | 86 % |

### Gardes-fous

| Règle | Seuil |
|-------|--------|
| Marge nette GMV après Stripe + commissions | **≥ 55 %** |
| Net annualisé / animal (mix multi-ans) | **≥ 14 € / an** |
| Remise max quinquennial vs 5× annual | **≤ 25 %** |
| Prix d’entrée client | **≤ ~2,5 € / mois** |

---

## 4. Concurrentiel (client)

| Repère marché | Ordre de grandeur | petsFollow |
|---------------|-------------------|------------|
| Collier médical | 250–430 €/an + hardware | ~10× moins — normal (pas de capteur) |
| GPS health light | 90–120 €/an + device | ~3–4× moins |
| Apps software-only | 50–180 €/an | **Bas de fourchette** |

Message : *suivi prescrit par votre véto, sans boîtier*.

---

## 5. Mix commercial recommandé

| Plan | Mix cible | Pourquoi |
|------|-----------|----------|
| Annual | 25 % | Conversion / essai |
| Triennial | **60 %** | Ticket / marge / commission |
| Quinquennial | 15 % | Fidélisation long terme |

Family proposé dès le 2ᵉ animal (complément) ; Care+ en upsell post-activation ; Horse seulement si cheval.

---

## 6. Paie commerciale (ops)

- Périodes `open` → `closed` → `paid` (tables `commercial_payout_*`)
- Admin : tableaux **prévisionnel / à payer / payé**, récap mensuel global et par personne + IBAN
- Commercial : suivi commissions + encodage IBAN / BIC / titulaire dans son profil

---

## 7. Application

| Couche | Action |
|--------|--------|
| Doc | ce fichier (source de vérité **politique**) |
| Code | `go/internal/billing/domain.go` + i18n + Flutter |
| Stripe | Prices `STRIPE_PRICE_*` + addons |
| Comm. | `billing.commission_settings` + paliers véto (admin) |
| Simu 10 ans | `documentation/16-…` + moteur Nuxt |
