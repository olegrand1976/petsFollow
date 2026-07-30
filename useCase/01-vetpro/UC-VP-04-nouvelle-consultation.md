# UC-VP-04 — Nouvelle consultation (CR → DAF / facture)

| | |
|--|--|
| **ID** | `UC-VP-04` |
| **Durée** | ~12 min |
| **Priorité** | Démo |
| **Surface** | Web VetPro |

## Objectif

Depuis la liste clients, démarrer une consultation rapide : animal → compte-rendu → enchaînement DAF et/ou facture Billit.

## Acteurs

| Rôle | Compte | MDP |
|------|--------|-----|
| Véto | `vet.demo@petsfollow.test` | `VetDemo123!` |

## Prérequis

- Staging / local Web Pro.
- Flags `PHARMACY_ENABLED` / Billit selon l’environnement (CTA masqués si off).
- Client seed avec au moins un animal.

## Étapes

1. Se connecter en véto → **Clients**.
2. Sur une ligne client, cliquer **Nouvelle Consultation**.
3. Choisir l’animal (pré-sélectionné depuis **Animaux** ou la **fiche animal**, ou s’il n’y en a qu’un) → **Démarrer**.
4. Écran CR split :
   - **Gauche — Notes / dictée** : écrire ou dicter (consentement audio) / importer un fichier.
   - **Droite — Compte-rendu** : **Améliorer (IA)** (source = notes si présentes, sinon le CR) ou éditer manuellement ; aperçu markdown après IA.
   - **Annuler les modifications** (footer panel) restaure le dernier enregistrement **sans** fermer la consultation ; **Annuler** (footer modal) quitte la consult (confirm leave).
   - Historique des versions (replié) : transcription d’origine → proposition IA → dernière version enregistrée ; restore vers la bonne pane.
   - **Enregistrer** (ou Finaliser).
5. (Optionnel) Panneau **Traitements (DAF)** : protocole 1 clic ou search CNK → AMM auto → enregistrer brouillon ; preview FEFO ; finaliser **dans la modal** (confirm ProModal) ou ouvrir le wizard pour édition avancée.
6. Choisir :
   - **Créer un DAF & Facturer** → wizard si besoin, sinon finalize inline puis **Facturer** (`mode=fromDaf`, lignes + montant estimé mock).
   - **ou Facturer directement** → page facturation avec contrepartie préremplie.
   - **ou Terminer** → visite `done` + finalisation auto des brouillons CR non vides (visible côté app client).
7. (Optionnel) Même CTA depuis la fiche client **ou la fiche animal**. Voir aussi [`UC-VP-05`](UC-VP-05-pharmacie-stock-daf.md) pour le parcours stock/DAF détaillé.

## Résultat attendu

- Visite créée en `confirmed` sans friction agenda (session walk-in : **n’occupe pas** un créneau client ; hors congés cabinet ; `source=care_pro` en terrain).
- Fermeture sans enregistrement CR → **confirm** Enregistrer / Annuler la consultation / Rester ; Annuler → visite annulée (pas d’orphelin). Fermeture pendant un enregistrement CR → attend la fin du PUT ; si le CR est déjà persisté (409 serveur), la visite est conservée.
- Historique cabinet : `/consultations` (date décroissante, filtres, audio draft si disponible + durée d’enregistrement).
- Veille / switch profil : autosave CR + reprise de la consultation pour le **même** utilisateur.
- CR accessible immédiatement.
- Deep-links DAF / Billit cohérents avec le contexte consultation (montant facture saisi manuellement ; `visitId` persisté sur le document). CTA libellé **DAF** (≠ module prescriptions).
- Traitements CNK en consult → brouillon DAF `visit_id` ; finalize = FEFO + déduction stock (jamais silencieux à la clôture CR) ; leave prompt si traitements dirty.
- E2E `@p0` : `03b-consultation.spec.ts` (Terminer · close sans save · CTA DAF/facture · traitements→preview FEFO→finalize).
- E2E `@p1` : `03e-visit-report-versions.spec.ts` (split panes · discard · restore · escape).

## Checklist

| | Résultat |
|--|----------|
| CTA liste clients | OK / KO / N/A |
| Modal + CR | OK / KO / N/A |
| CTA DAF / facture | OK / KO / N/A |

## Zone retour

- Bug :
- Friction UX :
- Idée :

→ Remonter via [`TEMPLATE-RETOUR.md`](../TEMPLATE-RETOUR.md).

---

*Réf. QA : C2.13, C2.19–C2.22, C7.4 · companion [`UC-VP-05`](UC-VP-05-pharmacie-stock-daf.md)*
