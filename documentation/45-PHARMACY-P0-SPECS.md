# 45 — Specs brouillon P0-4…P0-7 (pharmacie)

**Statut** : brouillons produit/juridique — **pas** encore figés.  
**Parent** : [37-ROADMAP-STOCK-FACTURATION.md](37-ROADMAP-STOCK-FACTURATION.md) §0.  
**Date** : 2026-08-11.

Ces cadres permettent de démarrer les discussions ; la Phase code (4.B / 4.E / 5.A–C) reste bloquée jusqu’à validation écrite (critère de sortie P0).

---

## P0-4 — Registre stupéfiants BE (→ Phase 4.B)

### Objectif
Tenir un registre conforme des substances contrôlées (stupéfiants / psychotropes vétérinaires) au niveau cabinet.

### Questions ouvertes (juridique)
1. Liste de référence officielle BE (codes CNK / ATC / annexe) à synchroniser ?
2. Durée de conservation minimale (souvent 5 ans — confirmer) vs rétention `pharmacy.*` déjà 5 ans (4.F) ?
3. Qui signe / clôture le registre (vétérinaire responsable vs assistante) ?
4. Export exigé : PDF horodaté, CSV, ou les deux ? Fréquence (à la demande / annuel) ?

### Modèle de données (proposition)
| Entité | Champs clés |
|--------|-------------|
| `ref_medications` | flag `is_controlled` (+ classe si besoin) |
| `controlled_movements` | lot, qty signée, motif (`receipt`/`daf`/`waste`/`adjust`), acteur, animal?, DAF?, timestamp |
| Export | snapshot période + hash + auteur |

### UI cible (après spec)
- Filtre `/stock` + journal dédié `/stock/controlled` (tag `dev`)
- Blocage finalize DAF si substance contrôlée sans motif / n° ordonnance interne

### Critère de sortie P0-4
Spec figée (ce doc ou annexe juridique) signée produit + juridique → ouvre 4.B.

---

## P0-5 — Grossiste pilote EDI (→ Phase 5.A)

### Objectif
Choisir **1** grossiste pilote et obtenir docs API (catalogue, commande, BL).

### Candidats
| Grossiste | Contact / docs | Décision |
|-----------|----------------|----------|
| Covetrus | ⬜ | |
| Alcyon | ⬜ | |
| Crocodil | ⬜ | |

### Checklist contact
1. Existence d’une API / EDI (REST, AS2, SFTP, portail) ?
2. Auth (clé, OAuth, VPN) ?
3. Formats catalogue + tarifs + n° dépôt ?
4. Flux commande → confirmation → BL électronique ?
5. Coût / contrat pilote staging ?

### Critère de sortie P0-5
Un pilote nommé + lien docs + accès sandbox → ouvre 5.A.

---

## P0-6 — Bigame / Vetcompendium build vs licence (→ Phase 5.B/C)

### Contexte
Admin Compendium PDF (`/admin/compendium-imports`, D11b) = **outil interne** d’enrichissement dictionnaire — **≠** licence notices commerciales Vetcompendium / Bigame.

### Options
| Option | Pros | Cons |
|--------|------|------|
| **A — Build** (étendre extract PDF + matching AFMPS) | Contrôle, pas de licence tierce | Couverture / màj manuelle, risque juridique notices |
| **B — Licence** Vetcompendium / Bigame | Contenu officiel, màj fournisseur | Coût, dépendance, intégration |
| **C — Hybride** | AFMPS = CNK source ; notices via licence | Complexité produit |

### Décision à écrire
- Choix A/B/C + budget + calendrier
- Périmètre espèces (compagnon vs production / Bigame)
- Qui porte la responsabilité éditoriale des posologies

### Critère de sortie P0-6
Note décisionnelle signée produit → ouvre 5.B et/ou 5.C.

---

## P0-7 — Certification DAF vs disclaimer (→ Phase 4.E)

### État actuel
PDF DAF généré avec **disclaimer « non certifié »** (module tag `dev`).

### Questions juridiques
1. Existe-t-il un chemin de certification AFMPS / organisme pour un DAF logiciel ?
2. Mentions obligatoires manquantes vs export papier carbone ?
3. Peut-on rester en disclaimer indéfiniment en GA pharmacie (hors facture) ?
4. Responsabilité cabinet vs éditeur petsFollow ?

### Livrable attendu
Note 1–2 pages : trajectoire (certifier / rester disclaimer / hors scope) + impacts UI PDF (4.E).

### Critère de sortie P0-7
Note validée juridique + produit → ouvre 4.E.

---

## Lien roadmap

Après chaque P0 clos : mettre à jour le tableau §0 de [37](37-ROADMAP-STOCK-FACTURATION.md) et le runbook [38](38-RUNBOOK-PHARMACIE-CABINET.md) si parcours cabinet impacté.
