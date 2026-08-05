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
3. Choisir l’animal (pré-sélectionné depuis **Animaux** ou la **fiche animal**, ou s’il n’y en a qu’un) → **Démarrer** → ouverture de la fiche `/consultations/{id}` (même écran que l’historique / le CR).
4. Écran CR split (workspace unique) :
   - **Gauche — Notes / dictée** : écrire ou dicter (consentement audio) / importer un fichier.
   - **Droite — Compte-rendu** : **Améliorer (IA)** (source = notes si présentes, sinon le CR) ou éditer manuellement ; aperçu markdown après IA.
   - **Annuler les modifications** (footer panel) restaure le dernier enregistrement **sans** fermer la consultation ; **Annuler** (footer workspace) quitte la consult (confirm leave).
   - Historique des versions (replié) : transcription d’origine → proposition IA → dernière version enregistrée ; restore vers la bonne pane.
   - **Enregistrer** (ou Finaliser).
5. Hub post-CR : CTA **DAF** (wizard `/daf/nouveau` prérempli `visitId`) / **Facturer** / **Consignes** (si flags) — pas de panneau traitements in-workspace ; parcours DAF détaillé → [`UC-VP-05`](UC-VP-05-pharmacie-stock-daf.md).
6. Choisir :
   - **Créer un DAF & Facturer** → wizard si besoin, sinon finalize inline puis **Facturer** (`mode=fromDaf`, lignes + montant estimé mock).
   - **ou Facturer directement** → page facturation avec contrepartie préremplie.
   - **ou Terminer** → visite `done` + finalisation auto des brouillons CR non vides (visible côté app client) → retour liste `/consultations`.
7. (Optionnel) Même CTA depuis la fiche client **ou la fiche animal**. Voir aussi [`UC-VP-05`](UC-VP-05-pharmacie-stock-daf.md) pour le parcours stock/DAF détaillé.
8. (Optionnel) Depuis **Agenda** → détail d’un RDV : le détail n’affiche **plus** le CR — bouton **Nouvelle consultation** (RDV confirmé à venir, ou walk-in à reprendre) **et** **Voir la consultation** (RDV passé / `done`) ouvrent le **même** écran `/consultations/{id}` (stade édition ou lecture selon droits / statut ; fermer sans enregistrer **n’annule pas** un RDV agenda).

## Résultat attendu

- Visite créée en `confirmed` sans friction agenda (session walk-in : **n’occupe pas** un créneau client ; hors congés cabinet ; `source=care_pro` en terrain).
- Fermeture sans enregistrement CR → **confirm** Enregistrer / Annuler la consultation / Rester ; Annuler → visite annulée (pas d’orphelin). Fermeture pendant un enregistrement CR → attend la fin du PUT ; si le CR est déjà persisté (409 serveur), la visite est conservée.
- Historique cabinet : `/consultations` (walk-ins + RDV avec CR, date décroissante, filtres, audio draft si disponible + durée d’enregistrement).
- Veille / switch profil : autosave CR + reprise de la consultation pour le **même** utilisateur.
- CR accessible immédiatement.
- Deep-links DAF / Billit cohérents avec le contexte consultation (montant facture saisi manuellement ; `visitId` persisté sur le document). CTA libellé **DAF** (≠ module prescriptions).
- Hub CTA **DAF** → wizard avec `visitId` ; finalize FEFO + déduction stock dans le wizard (jamais silencieux à la clôture CR).
- E2E `@p0` : `03b-consultation.spec.ts` (Terminer · close sans save · leave-save · dirty post-save · CTA DAF/facture · hub sans traitements in-workspace · détail RDV agenda → visite conservée).
- E2E `@p1` : `03e-visit-report-versions.spec.ts` (split panes · discard · restore · escape).

## Checklist

| | Résultat |
|--|----------|
| CTA liste clients | OK / KO / N/A |
| Setup modal → workspace page | OK / KO / N/A |
| CTA DAF / facture | OK / KO / N/A |

## Zone retour

- Bug :
- Friction UX :
- Idée :

→ Remonter via [`TEMPLATE-RETOUR.md`](../TEMPLATE-RETOUR.md).

---

*Réf. QA : C2.13, C2.19–C2.22, C7.4 · companion [`UC-VP-05`](UC-VP-05-pharmacie-stock-daf.md)*
