# UC-VP-05 — Pharmacie stock → DAF (consultation)

| | |
|--|--|
| **ID** | `UC-VP-05` |
| **Durée** | ~15 min |
| **Priorité** | Démo |
| **Surface** | Web VetPro |

## Objectif

Montrer le parcours **one-sitting** : CR → traitements (protocole 1 clic / CNK) → preview FEFO → finalize DAF → facture mock — sans quitter la modal consultation, avec stock démo prêt.

## Acteurs

| Rôle | Compte | MDP |
|------|--------|-----|
| Véto | `vet.demo@petsfollow.test` | `VetDemo123!` |

## Prérequis

- `PHARMACY_ENABLED` / `NUXT_PUBLIC_PHARMACY_ENABLED` on.
- Seed : CNK démo + lots `SEED-*` VetPlus + protocoles cliniques.
- Module tag **dev** (nav pharmacie / DAF).

## Étapes

1. Clients → **Nouvelle Consultation** → animal → **Démarrer** → enregistrer un CR.
2. Panneau **Traitements (DAF)** :
   - (Optionnel) cliquer un **protocole** (ex. « Antibiothérapie courte ») → lignes + AMM préremplies ; espèce VAMReg préremplie si antibiotique.
   - ou search CNK `Amoxicilline` / `2712345`.
3. **Enregistrer le brouillon** → **Preview FEFO** (lots seed visibles).
4. Si rupture : mini-réception lot inline → re-preview (sans aller sur `/stock`).
5. **Finaliser le DAF** → confirm ProModal → stock déduit.
6. CTA **Facturer** (si Billit UI on) → `/invoicing?dafId=&mode=fromDaf` avec lignes + montant HT estimé si prix stock.
7. Fiche animal → bloc **Dispenses DAF** (lecture).
8. `/daf?status=draft` : badge âge si brouillon > 1 h ; `/consultations` : badge « Brouillon DAF >1 h ».

## Résultat attendu

- Finalize **dans** la modal (pas besoin du wizard sauf édition avancée).
- Timeline client : événement non-PHI « Traitement cabinet ».
- Pas de finalize silencieux à la clôture CR ; dirty traitements → leave prompt.

## Checklist

| | Résultat |
|--|----------|
| Protocole / CNK + FEFO | OK / KO / N/A |
| Finalize + stock | OK / KO / N/A |
| Dispenses fiche animal | OK / KO / N/A |

## Zone retour

- Bug :
- Friction UX :
- Idée :

→ Remonter via [`TEMPLATE-RETOUR.md`](../TEMPLATE-RETOUR.md).

---

*Réf. QA : C7.1–C7.4, C2.16 · doc [37](../../documentation/37-ROADMAP-STOCK-FACTURATION.md)*
