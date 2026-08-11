# 37 — Roadmap stock opérationnel + réglementaire + facturation

**Objectif** : rendre opérationnelle et conforme la boucle cabinet (stock → DAF → VAMReg → facture), face au deal-breaker « pharmacie + facturation ».

| Méta | Valeur |
|------|--------|
| Statut | **Cadré** · S4 ✅ · Phase 2.A–2.E ✅ · 2.F catalogue ✅ (staging+prod) · 4.C/4.D/4.F ✅ · S6 ✅ · S5 / Phase 3 gelés (P0-2) |
| Socle Phase 1 | [27-PHARMACIE-BELGIQUE.md](27-PHARMACIE-BELGIQUE.md) · [28-PLAN-STOCK-PEREMPTION.md](28-PLAN-STOCK-PEREMPTION.md) |
| Facturation | [33-BILLIT-INTEGRATION.md](33-BILLIT-INTEGRATION.md) · [34-BILLIT-RESELLER-TECH.md](34-BILLIT-RESELLER-TECH.md) |
| Dernière revue | 2026-08-11 (P0-1/P0-2 mails envoyés · 45 décisions produit finalisées) |

**Hors scope** (retour véto clinique / offline) : normes vitals, templates dentaires UGent, canvas protocoles, mode offline Flutter — backlog séparé.

---

## 0. Phase 0 — Cadrage partenaires & données (suivi)

Sans ces prérequis, les phases techniques restent un « presque conforme ».

| ID | Livrable | Owner | Statut | Critère de sortie |
|----|----------|-------|--------|-------------------|
| P0-1 | Accès / contrat API **VAMReg** (dry-run + calendrier go-live) | Ops / juridique | 🟡 Readonly OK · **mail write envoyé (2026-08-11)** — attente credentials | Credentials SM + env test + ICD write |
| P0-2 | **Accès reseller Billit** | Ops | 🟡 **Mail reseller envoyé (2026-08-11)** — attente accès ; S5/Phase 3 gelés | Credentials reseller → débloque S5 + Phase 3 |
| P0-3 | Source officielle catalogue **AFMPS / CNK** (licence, cadence maj) | Produit / Ops | 🟢 Licence + cadence + pipeline cron ✅ · **staging + prod** (~2738 CNK) — [38 § Prod](38-RUNBOOK-PHARMACIE-CABINET.md#prod-premier-catalogue) | Catalogue national en base staging/prod |
| P0-4 | Inventaire obligations **stupéfiants BE** + modèle registre | Produit / juridique | 🟡 Décision produit figée ([45](45-PHARMACY-P0-SPECS.md)) — ☐ juridique | Spec + OK juridique → Phase 4.B |
| P0-5 | Contact / docs API **grossistes** (1 pilote) | Produit | 🟡 Process figé ([45](45-PHARMACY-P0-SPECS.md)) — ☐ choix pilote commercial | Pilote + docs → Phase 5.A |
| P0-6 | **Bigame / Vetcompendium** — build vs licence | Produit | 🟡 Décision produit : build Compendium + licence différée ([45](45-PHARMACY-P0-SPECS.md)) — ☐ budget | OK produit → 5.B/C reportés |
| P0-7 | Trajectoire **certification DAF** vs disclaimer | Juridique | 🟡 Décision produit : rester disclaimer ([45](45-PHARMACY-P0-SPECS.md)) — ☐ juridique | OK juridique → polish 4.E |

### Suivi P0 partenaires (hors code) — 2026-08-11

Actions ops / juridique à pousser **en parallèle** du code (pas de chantier technique tant que le critère de sortie n’est pas atteint) :

| ID | Prochaine action concrète | Bloque |
|----|---------------------------|--------|
| **P0-1** | **Mail envoyé (2026-08-11)** — attente credentials write ; ticket [`527c7803…`](https://petsfollow.ll-it-sc.be/admin/support/527c7803-b27e-43d4-bb46-33aa175b4c8f) · garder `VAMREG_DRY_RUN=true` ([39](39-VAMREG-AFMPS-READONLY.md)) | 4.A production |
| **P0-2** | **Mail envoyé (2026-08-11)** — attente reseller ; ticket [`572d88c4…`](https://petsfollow.ll-it-sc.be/admin/support/572d88c4-9489-4668-bc82-b07dee145be8) · S5/Phase 3 gelés | 3.C–3.E, GA facture |
| **P0-3** | ✅ Staging + prod — [38 § Prod](38-RUNBOOK-PHARMACIE-CABINET.md#prod-premier-catalogue) | — |
| **P0-4** | Décisions produit [45 § P0-4](45-PHARMACY-P0-SPECS.md) — ☐ validation juridique | 4.B |
| **P0-5** | Process [45 § P0-5](45-PHARMACY-P0-SPECS.md) — ☐ retenir 1 pilote | 5.A |
| **P0-6** | Décision [45 § P0-6](45-PHARMACY-P0-SPECS.md) (build + licence différée) — ☐ OK budget | 5.B/C reportés |
| **P0-7** | Décision [45 § P0-7](45-PHARMACY-P0-SPECS.md) (disclaimer) — ☐ validation juridique | 4.E |

**Note** : l’admin Compendium PDF (`/admin/compendium-imports`, D11b) est **livré** sous flag `dev` (revue dual-list + pagination) — distinct de P0-6 / Phase 5.C (notices commerciales Vetcompendium).

### Archive mails partenaires (envoyés 2026-08-11)

Mails **envoyés** — corps conservé pour relance / audit. Tickets : P0-1 [`527c7803…`](https://petsfollow.ll-it-sc.be/admin/support/527c7803-b27e-43d4-bb46-33aa175b4c8f) · P0-2 [`572d88c4…`](https://petsfollow.ll-it-sc.be/admin/support/572d88c4-9489-4668-bc82-b07dee145be8).

<details>
<summary>Corps P0-1 — VAMReg write</summary>

```
Objet : petsFollow — accès API VAMReg déclaration (write) + env test

Bonjour,

Nous développons petsFollow (software house vétérinaire, Belgique).
Le client API VAMReg en lecture seule (listes espèces / indications / substances)
est déjà opérationnel côté produit (réf. interne doc 39-VAMREG-AFMPS-READONLY).

Pour passer en déclaration live (ICD write), merci de nous communiquer :

1. Credentials / certificat ICD **write** (environnement de test d’abord)
2. URL(s) de l’environnement de test et procédure d’accès
3. Calendrier / critères de go-live production
4. Contacts techniques pour le dépôt des secrets (nous stockons côté Cloud Run
   Secret Manager, variables VAMREG_*)

Tant que ces éléments ne sont pas reçus, nous restons en VAMREG_DRY_RUN=true
(aucune déclaration réelle).

Merci d’avance,
[Signature ops / juridique petsFollow]
```

</details>

<details>
<summary>Corps P0-2 — Billit reseller</summary>

```
Objet : petsFollow — accès reseller API Billit

Bonjour,

petsFollow intègre déjà Billit côté code (mapping DAF → lignes de facture / BIL-9,
préremplissage GET /practices/me/invoicing/prefill). Ce flux est **gelé** en attendant
les accès reseller (réf. interne docs 33 / 34-BILLIT-RESELLER-TECH).

Merci de nous fournir :

1. Compte **reseller** (sandbox puis production)
2. Credentials API (clés / tokens) et documentation d’auth
3. Procédure Peppol / envoi cabinet → clients si applicable au modèle reseller
4. Contact technique pour le branchement Secret Manager Cloud Run

Dès réception, nous branchons les secrets et dégelons EnqueueInvoicesConnect —
aucun pilote DAF→facture live sans ces accès.

Merci,
[Signature ops petsFollow]
```

</details>

### Prérequis facturation — Billit reseller (figé)

**On attend encore les accès reseller Billit.** Tant qu’ils ne sont pas reçus :

- **Gelé** : Phase 3, Peppol live cabinet, GA facturation.
- **Livré sans reseller** : BIL-9 — lignes de facture préremplies (acte du type de RDV + DAF finalisé de la visite) via `GET /practices/me/invoicing/prefill`, en mock comme en live.
- **Poursuivable** : S4 VAMReg, S6 staging pharmacie, Phase 2 stock ops, Phase 4 réglementaire (hors lien facture), Phase 5 grossistes/Bigame.
- Socle Billit déjà dans le repo : **ne pas ré-implémenter** ; ne pas pousser DAF→facture en pilote réel sans reseller.

---

## 1. Architecture cible

```mermaid
flowchart LR
  Catalog[Catalogue CNK prix] --> Stock[Stock lots FEFO]
  Wholesaler[Grossistes EDI] --> Stock
  Stock --> DAF[DAF finalize]
  DAF --> VAMReg[VAMReg]
  DAF --> Billit[Facture Billit Peppol]
  Rules[Regles chaines alimentaires] --> DAF
  Narcotics[Registre stupefiants] --> Stock
```

---

## 2. Phase 1 — Fermer le socle (S4–S6)

Voir détail d’exécution dans [28-PLAN-STOCK-PEREMPTION.md](28-PLAN-STOCK-PEREMPTION.md).

| Sprint | Contenu | Statut |
|--------|---------|--------|
| **S4** VAMReg | `job_audit`, Asynq opt-in, dry-run sync, retry UI | ✅ Dry-run |
| **S5** DAF→Billit | BIL-9 lignes préremplies (acte + DAF) | ✅ Livré · worker `invoices.connect` toujours ⏸ (P0-2) |
| **S6** Ops staging | `PHARMACY_ENABLED`, `VAMREG_DRY_RUN`, workers env, `pg_trgm`, smoke, UC | ✅ Env/secrets · `pg_trgm` · smoke MVP · `smoke-pharmacy-s6-staging` · UC-VP-05 |

**Done when (chemin stock)** : pilote staging — receipt → DAF → VAMReg dry-run OK.  
**Done when (chemin facture)** : + draft Billit lié DAF — **après** P0-2.

---

## 3. Phase 2 — Stock cabinet opérationnel

Objectif : tourner sans Excel (hors EDI).

| ID | Chantier | Contenu | Done when |
|----|----------|---------|-----------|
| 2.A | Catalogue prix practice | Prix achat / vente HT + TVA par CNK | ✅ API + **UI** `/stock` (carte prix) |
| 2.B | Seuils & alertes réassort | Seuil min + `GET …/reorder-alerts` | ✅ API + **UI** édition seuil + alertes |
| 2.C | Commandes internes | Brouillon + e-mail fournisseur (CSV) | ✅ API `orders` + UI `/stock` + `SendVetAlertWithCSV` |
| 2.D | Réception enrichie | BL manuel (n°, fournisseur, lignes → lots) + notif | ✅ API `delivery-notes` + UI BL optionnel |
| 2.E | Inventaire annuel | Session, écarts → adjust, export | ✅ API `inventory/sessions` + UI `/stock` + CSV |
| 2.F | Import CNK national | Pipeline AFMPS admin/CLI ✅ · job interne mensuel gate 1 (`/internal/afmps-import/run`) ✅ · catalogue national en base | ✅ **Staging + prod** catalogue commité (~2738 CNK, 2026-08-11) — [38 § Prod](38-RUNBOOK-PHARMACIE-CABINET.md#prod-premier-catalogue) |
| — | UI journal / settings | Mouvements, waste reasons, settings expiry, adjust manuel | ✅ composants `pharmacy/Stock*` |

**Dépendances** : P0-3 pour 2.F ; 2.A utile à Phase 3.

---

## 4. Phase 3 — Facturation associée (après reseller)

**Prérequis** : P0-2 (accès reseller Billit branchés SM + staging).

**Gate code** : `pharmacy.EnqueueInvoicesConnect` → `ErrInvoicesConnectResellerPending` jusqu’au déblocage (ne pas brancher sur finalize).

### Blocage actuel (2026-08-11)

| Chantier | Statut | Débloque quand |
|----------|--------|----------------|
| **S5** worker `invoices.connect` + Phase **3.C–3.E** | ⏸ gelé | Credentials **P0-2** branchés SM |
| **4.A** VAMReg live | ⏸ dry-run only | Credentials **P0-1** write + `VAMREG_DRY_RUN=false` |

Aucun code Phase 3 / 4.A à démarrer tant que les tickets P0-1 / P0-2 n’ont pas de réponse partenaire.

| ID | Chantier | Contenu | Statut |
|----|----------|---------|--------|
| 3.A | Mapping DAF → lignes | Lignes proposées par l’API (qty, libellé, prix catalogue, TVA) ; revue humaine avant envoi | ✅ Livré (`/invoicing/prefill`) |
| 3.B | Parcours consultation | CTA post-CR : acte tarifé + lignes DAF préremplies | ✅ Livré |
| 3.C | Avoirs / retours | Credit note + cohérence stock | ⏸ |
| 3.D | KYC & quotas | Connexion `active`, plafonds, runbook rejet | ⏸ |
| 3.E | Reporting | CA médicaments, marge, export comptable | ⏸ |

Lien ticket : **BIL-9** — [33 § Phase 5](33-BILLIT-INTEGRATION.md) — **livré** ; le reste de la Phase 3 (avoirs/stock, KYC, reporting) attend les accès reseller.

---

## 5. Phase 4 — Conformité réglementaire cœur (BE / UE)

| ID | Chantier | Contenu | Écart actuel |
|----|----------|---------|--------------|
| 4.A | VAMReg production | Dry-run → API réelle ; audit ; alertes | Gate + worker dry-run · attend P0-1 |
| 4.B | Registre stupéfiants | Flag substances ; mouvements ; export ; UI | Absent — attend P0-4 |
| 4.C | Temps d’attente structurés | Viande / lait / œufs sur référentiel + DAF | ✅ Colonnes ref + snapshot DAF + PDF + `PATCH …/withdrawal` |
| 4.D | Chaîne alimentaire | Statut animal ; DAF obligatoire ; éviction (ex. phénylbutazone) | ✅ `food_chain_status` + gate finalize + ban flag |
| 4.E | Mentions DAF renforcées | Alignement AFMPS + revue juridique PDF | Disclaimer non certifié — attend P0-7 |
| 4.F | Traçabilité 5 ans | Rétention / export registres ; pas de purge destructive mouvements | ✅ Job rétention users **n’efface pas** `pharmacy.*` · `REVOKE UPDATE/DELETE` mouvements (000172) · `GET …/movements/retention-stats` · inventaire CSV · runbook 38 |

---

## 6. Phase 5 — Intégrations externes

| ID | Chantier | Contenu | Prérequis |
|----|----------|---------|-----------|
| 5.A | Grossistes EDI | Catalogue, n° dépôt, tarifs, commande B2B, BL électronique | P0-5 |
| 5.B | Bigame | Animaux de production | P0-6 |
| 5.C | Vetcompendium | Notices / posologies / temps d’attente | P0-6 |

---

## 7. Phase 6 — GA pharmacie (+ facturation si P0-2)

| Critère | Statut / action |
|---------|-----------------|
| Flags staging + prod documentés | 🟡 Staging pharmacy + VAMReg dry-run/workers dans [10](10-GCP-DEPLOIEMENT.md) ; prod opt-in |
| Retrait `nav.tagDev` | ⬜ Décision produit explicite |
| UC commerciaux + `make usecases-sync` | ✅ [UC-VP-05](../useCase/01-vetpro/UC-VP-05-pharmacie-stock-daf.md) |
| Matrice P0 [15-PLAN-TESTS.md](15-PLAN-TESTS.md) | ✅ C7.1–C7.15 |
| Runbook cabinet | ✅ [38-RUNBOOK-PHARMACIE-CABINET.md](38-RUNBOOK-PHARMACIE-CABINET.md) |
| Fiche produit / concurrence | ⬜ Claim aligné sur livré réel |

Sans reseller : GA possible **pharmacie seule** (tag `dev` facturation conservé).

---

## 8. Ordre & dépendances

```mermaid
flowchart TB
  P0[Phase0 Cadre partenaires]
  BillitWait[Acces reseller Billit]
  P1a[S4 VAMReg + S6 staging]
  P1b[S5 DAF vers Billit]
  P2[Phase2 Stock ops]
  P3[Phase3 Facturation DAF]
  P4[Phase4 Reglementaire]
  P5[Phase5 EDI Bigame Vetcomp]
  P6[Phase6 GA]
  P0 --> P1a
  P1a --> P2
  P1a --> P4
  BillitWait --> P1b
  P1b --> P3
  P2 --> P3
  P4 --> P5
  P4 --> P6
  P3 --> P6
  P5 --> P6
```

**Priorité actuelle** : P0-3 ✅ · **P0-1/P0-2** mails envoyés (attente credentials) · **P0-4…7** décisions produit dans [45](45-PHARMACY-P0-SPECS.md) (☐ juridique/commercial) · Phase 3 / 4.A **gelées**.  
**Facturation** : dès credentials P0-2 → S5 → Phase 3.  
**Prochain code** : uniquement après réponse P0-1 ou P0-2 (ou coche juridique 45 pour 4.B/4.E).

---

## 9. Ordre de grandeur (indicatif, 1–2 devs)

| Phase | Effort | Nature |
|-------|--------|--------|
| 0 | 2–4 sem | Juridique / partenaires |
| 1 | 3–5 sem | S4–S6 (S5 gelé) |
| 2 | 4–6 sem | Stock ops + prix + inventaire |
| 3 | 3–5 sem | DAF↔Billit (après reseller) |
| 4 | 6–10 sem | Stupéfiants + chaîne alimentaire + VAMReg live |
| 5 | 8–16 sem | EDI + Bigame + Vetcompendium |
| 6 | 1–2 sem | GA / docs / UC / flags |

---

## 10. Liens

| Doc | Rôle |
|-----|------|
| [27](27-PHARMACIE-BELGIQUE.md) | Spec technique Phase 1 |
| [28](28-PLAN-STOCK-PEREMPTION.md) | Suivi sprints S0–S6 |
| [33](33-BILLIT-INTEGRATION.md) | Billit + BIL-9 |
| [15](15-PLAN-TESTS.md) | P0 / QA |
| [45](45-PHARMACY-P0-SPECS.md) | Brouillons P0-4…7 |
| [04-MODULES-METIER.md](04-MODULES-METIER.md) | Modules |
