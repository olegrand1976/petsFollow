# 45 — Specs P0-4…P0-7 (pharmacie) — décisions produit

**Statut** : **finalisé produit (2026-08-11)** — validation juridique / commerciale encore ouverte (cases ☐).  
**Parent** : [37-ROADMAP-STOCK-FACTURATION.md](37-ROADMAP-STOCK-FACTURATION.md) §0.  
**Effet** : ces décisions **cadrent** les Phases 4.B / 4.E / 5.A–C ; le **code** ne démarre qu’après coche juridique/commercial ci-dessous (ou amendement écrit).

| P0 | Décision produit | Validation restante |
|----|------------------|---------------------|
| P0-4 | Modèle registre + rétention 5 ans + exports | ☐ Juridique BE |
| P0-5 | Process pilote EDI (1 grossiste) | ☐ Commercial (choix nom) |
| P0-6 | Hybride : AFMPS + Compendium build ; **5.B/C reportés** | ☐ Produit / budget (valide le report) |
| P0-7 | Rester en disclaimer non certifié jusqu’à trajectoire certif claire | ☐ Juridique |

---

## P0-4 — Registre stupéfiants BE (→ Phase 4.B)

### Décision produit
1. **Périmètre** : substances contrôlées (stupéfiants / psychotropes) utilisées en cabinet — flag sur le référentiel + journal d’écritures immuable.
2. **Liste** : démarrer par flag manuel admin (`is_controlled` sur `ref_medications`) + import/annotation ; pas de scraper légal inventé. La liste officielle BE exacte = **validation juridique** (annexe / ATC / CNK).
3. **Rétention** : **5 ans** minimum, alignée sur 4.F (`pharmacy.*` déjà hors purge destructive mouvements).
4. **Responsable** : le **vétérinaire** (profil `vet`) clôture / atteste ; assistante peut saisir sous supervision (ACL à préciser en 4.B).
5. **Export** : **PDF horodaté + CSV** à la demande (période) ; pas d’obligation de cron annuel en V1.

### Modèle de données (cible 4.B)
| Entité | Champs clés |
|--------|-------------|
| `ref_medications` | `is_controlled` (+ `controlled_class` optionnel) |
| `pharmacy.controlled_movements` (**table dédiée**, immuable) | lot, qty signée, motif (`receipt`/`daf`/`waste`/`adjust`), acteur, animal?, DAF?, timestamp, attestation |
| Export | snapshot période + hash + auteur |

**Non-choix** : pas de simple vue filtrée sur `stock_movements` — le journal stock actuel n’a pas les champs registre (attestation, motif légal) et n’est pas un substitut de registre stupéfiants. Les mouvements stock peuvent **alimenter** le registre (hook receipt/DAF/waste) mais la source de vérité exportable = `controlled_movements`.

### UI cible (4.B)
- Filtre `/stock` + journal `/stock/controlled` (tag `dev`)
- Finalize DAF : si ligne contrôlée → motif / n° interne requis

### Critère de sortie P0-4
- [x] Décision produit écrite (ce §)
- [ ] Validation juridique BE (liste + signature + export) → ouvre **4.B**

---

## P0-5 — Grossiste pilote EDI (→ Phase 5.A)

### Décision produit
1. **Un seul pilote** avant tout code EDI.
2. **Candidats** (ordre de prise de contact commercial) : Covetrus → Alcyon → Crocodil.
3. **Critères de choix** : API/docs sandbox, auth claire, catalogue + commande + BL, coût pilote acceptable.
4. **Hors V1 pharmacie** tant qu’aucun pilote n’a répondu favorablement — le réassort e-mail CSV actuel reste le filet.

### Checklist contact (commercial)
1. API / EDI (REST, AS2, SFTP, portail) ?
2. Auth (clé, OAuth, VPN) ?
3. Formats catalogue + tarifs + n° dépôt ?
4. Flux commande → confirmation → BL électronique ?
5. Coût / contrat pilote staging ?

| Grossiste | Contacté | Docs reçues | Retenu |
|-----------|----------|-------------|--------|
| Covetrus | ☐ | ☐ | ☐ |
| Alcyon | ☐ | ☐ | ☐ |
| Crocodil | ☐ | ☐ | ☐ |

### Critère de sortie P0-5
- [x] Process + ordre de contact figés (ce §)
- [ ] Pilote nommé + docs sandbox → ouvre **5.A**

---

## P0-6 — Bigame / Vetcompendium build vs licence (→ Phase 5.B/C)

### Contexte
Admin Compendium PDF (`/admin/compendium-imports`, D11b) = **outil interne** d’enrichissement dictionnaire — **≠** licence notices commerciales.

### Décision produit (hybride / différé)
| Couche | Choix |
|--------|--------|
| CNK / catalogue national | **AFMPS** (P0-3 ✅) |
| Enrichissement fiches (labo, substance, force, matching) | **Build** via Compendium admin (D11b) — poursuivre sous tag `dev` |
| Notices / posologies commerciales Vetcompendium | **Différé** — pas de licence V1 ; pas de promesse UI « notice officielle » |
| Bigame (animaux de production) | **Hors V1** compagnon — réévaluer après GA pharmacie cœur |

Responsabilité éditoriale des textes extraits PDF : **cabinet / ops** (revue humaine gates) — l’IA n’est qu’assistant d’extraction.

### Critère de sortie P0-6
- [x] Décision produit écrite (ce §) : **report explicite** de 5.B/C (licence / Bigame) ; D11b = chemin build V1
- [ ] OK budget / roadmap commerciale = **validation du report** (pas d’ouverture de 5.B/C) — coche = « on assume le différé »

---

## P0-7 — Certification DAF vs disclaimer (→ Phase 4.E)

### Décision produit
1. **Rester en disclaimer « non certifié »** pour le PDF DAF tant qu’aucune trajectoire de certification AFMPS / organisme n’est ouverte et budgétée.
2. **GA pharmacie** (Phase 6) possible **avec** disclaimer visible — ne pas présenter le PDF comme substitut légal d’un document certifié.
3. Mentions PDF : améliorer clarté du disclaimer + champs déjà livrés (retrait, chaîne alimentaire) ; **pas** de chantier « certif » sans note juridique.
4. Responsabilité : le **cabinet** reste émetteur du DAF ; petsFollow = outil logiciel.

### Critère de sortie P0-7
- [x] Décision produit écrite (ce §)
- [ ] Validation juridique (disclaimer suffisant pour GA) → ouvre polish **4.E** (mentions) sans certif

---

## Après validation

1. Cocher les ☐ ci-dessus (ou amender).
2. Mettre à jour [37](37-ROADMAP-STOCK-FACTURATION.md) §0 (P0-4…7 → 🟢 / 🟡).
3. Ouvrir le chantier code correspondant **uniquement** alors :
   - P0-4 validé → **4.B**
   - P0-5 pilote nommé → **5.A**
   - P0-7 validé → polish **4.E** (mentions, sans certif)
   - P0-6 coché → **rien à coder** pour 5.B/C (report confirmé) ; D11b continue sous `dev`
