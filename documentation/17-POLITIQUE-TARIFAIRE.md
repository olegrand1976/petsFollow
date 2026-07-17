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
| Triennial **recommandé** | `triennial` | **79 € / 3 ans** | 2,2 € | Cœur d’offre & pitch commercial |
| Quinquennial | `quinquennial` | **115 € / 5 ans** | 1,9 € | Engagement long sans dumping |
| Family | `family` | **55 € / an** | — | Addon foyer **2–3** animaux + digest consolidé |
| Care+ | `care_plus` | **19 € / an** | — | Upsell soins / médicaments (compte) |
| Horse | `horse` | **39 € / an** | — | Pack équine (compte, pets horse) |

Modes : `one_time` ou `subscription` (interval 1 / 3 / 5 ans).

### Remises engagement (vs 3× / 5× annual)

| Plan | Plein tarif | Prix | Remise |
|------|-------------|------|--------|
| Triennial | 87 € | 79 € | **−9 %** |
| Quinquennial | 145 € | 115 € | **−21 %** |

Règle : ne jamais descendre sous **≈ −25 %** sur le 5 ans.

### Family — règle produit

- Scope **owner** (compte) : pack foyer **2–3** animaux (achat et plafond).
- **Addon** (complément), pas un substitut des abonnements animal.
- Achat : foyer avec **2 ou 3** animaux déjà créés.
- Avec Family actif (ou pending checkout de moins de 24 h) : plafond **3 animaux** (création du 4ᵉ bloquée).
- **Sans Family** : la liste Care **par animal** (et la navigation multi-animaux) reste accessible — aucun blocage des rappels standards.
- **Avec Family** — privilège payant : **digest consolidé foyer** (`GET /me/household`) — prochains rappels de soins agrégés sur tous les animaux du compte.
- Copy UI : « Forfait foyer — jusqu’à 3 animaux » (pas « abo illimité »).

### Care+ — privilèges

Sans Care+ : rappels standards seedés (vaccin, vermifuge, contrôle, dentaire).

Avec Care+ (**livré**) : rappels **médicaments**, rappels **personnalisés**.

**Roadmap** (annoncé, non livré) : historique / export carnet, relances email J-3 / J0.

### Horse — privilèges

Sans pack (même si `species=horse`) : fiche + relève + rappels généraux uniquement.

Avec Horse : rappels maréchal / coproscopie, carnet maréchal, **contacts pro**, **compétitions**.  
Promo in-app **uniquement** si le foyer a ≥1 cheval.

---

## 2. Commissions (partenaires)

| Acteur | Règle | Commentaire |
|--------|--------|-------------|
| **Véto** | Progressif **5 % → 8 % → 10 % → 12 %** (volume clients) | Plafond **12 %** — éditable admin |
| **Commercial** | **12 % fixe** sur abonnements **et** addons | Éditable admin (`commission_settings`) |
| **Véto addons** | Aucune | Inchangé |

### Commission absolue (plafond véto 12 % + commercial 12 %)

| Offre | Prix | Comm. véto | Comm. commercial |
|-------|------|------------|------------------|
| Annual | 29 € | **3,48 €** | **3,48 €** |
| Triennial | 79 € | **9,48 €** | **9,48 €** |
| Quinquennial | 115 € | **13,80 €** | **13,80 €** |
| Family | 55 € | — | **6,60 €** |
| Care+ | 19 € | — | **2,28 €** |
| Horse | 39 € | — | **4,68 €** |

**Pitch commercial** : vendre le **triennial** (~**9,5 €** / animal @12 %) plutôt que l’annual (~3,5 €).  
**Pitch véto** : au plafond 12 %, même € que le commercial sur les abos.

---

## 3. Économie unitaire plateforme

Hypothèses Stripe EU approx. : **1,5 % + 0,25 €**.

### Au pire (véto 12 % + commercial 12 %)

| Offre | Brut | Stripe | Partenaires | **Net plateforme** | % net | Net / an |
|-------|------|--------|-------------|---------------------|-------|----------|
| Annual 29 € | 29 | 0,69 | 6,96 | **21,35 €** | 74 % | 21,35 € |
| Triennial 79 € | 79 | 1,44 | 18,96 | **58,60 €** | 74 % | **19,53 €** |
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
| Apps software-only | 50–150 €/an | Bas / milieu bas de fourchette |

Message pitch : *suivi prescrit par votre véto, sans boîtier*.

---

## 5. Mix commercial recommandé

| Plan | Mix cible |
|------|-----------|
| Annual | 25 % |
| Triennial | **55 %** |
| Quinquennial | 20 % |

Family dès le 2ᵉ animal ; Care+ en upsell post-activation.

---

## 6. Application

| Couche | Source |
|--------|--------|
| Montants | `go/internal/billing/domain.go` |
| Labels | i18n Go `billing.*` + Flutter ARB + Nuxt locales |
| Comm. commercial | `DefaultCommercialCommissionRateBps` = **1200** |
| Stripe | Prices `STRIPE_PRICE_*` à aligner (29 / 79 / 115 ; 55 / 19 / 39) |
