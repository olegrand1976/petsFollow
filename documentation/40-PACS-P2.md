# 40-PACS — Plan P2 (vers GA clinique)

> **Statut** : **P2.0** hotfix EOF déployé (`b32145b`) · **P2.1 démarré** (CSP wasm, metadata API, Cornerstone opt-in, flag `NUXT_PUBLIC_PACS_VIEWER_ENGINE`).  
> Prérequis : P0 + P1 livrés · arbitrage **A Cornerstone3D** acté.  
> Doc module : [`40-PACS.md`](40-PACS.md) · règle tag `dev` : [`.cursor/rules/modules-tag-dev.mdc`](../.cursor/rules/modules-tag-dev.mdc).

## Objectif

Passer le module Imagerie d’un **outil démo tag `dev`** (previews PNG Orthanc + canvas) à un **viewer clinique crédible** prêt pour une **décision GA** explicite — sans C-STORE, sans partage client, sans archivage multi-pays.

## Déjà livré (hors P2)

| Lot | Contenu |
|-----|---------|
| P0 | `UNIQUE(orthanc_study_id)`, ParentSeries authz, launch pad, isolation tests, deploy staging |
| P1 | Multi-série, multi-frame, download `.dcm`, admin wake, purge Orthanc testée |
| Review | Frame EOF = 404 only, reset `waking`, magic DICM download |

Checklist GA G5 / G6 / G7 : **faits** (rester documentés dans `40-PACS.md`).

## Hors scope P2 (inchangé)

- C-STORE / DIMSE modalités
- Partage d’images vers le client Flutter
- PgBouncer dédié Orthanc
- Conformité archivage légal multi-pays
- Use case commercial **avant** décision G1 / retrait `nav.tagDev`

---

## Arbitrage viewer (tranché)

| Option | Approche | Décision |
|--------|----------|----------|
| **A — Cornerstone3D** | Lazy client-only + WASM ; pixels via WADO-RS Orthanc (proxy Go) ou decode `.dcm` | **Retenu** (2026-07-31) |
| B — Canvas + metadata | PNG Orthanc + `PixelSpacing` | Rejeté (plafond clinique) |

**Conséquences** : P2.1 = spike CSP + Cornerstone3D ; G3/G4 sur moteur Cornerstone ; fallback canvas conservé derrière flag `pacsViewerEngine` le temps de la migration.

Tant que G1 n’est pas signé, **ne pas** retirer `nav.tagDev` et **ne pas** inventer de UC commercial.

---

## Phases

```mermaid
flowchart TD
  P20[P2.0 hygiene staging]
  P21[P2.1 viewer clinique G3]
  P22[P2.2 mesures G4]
  P23[P2.3 filet tests]
  P24[P2.4 produit go-live]
  P20 --> P21
  P21 --> P22
  P22 --> P23
  P23 --> P24
```

### P2.0 — Hygiene & alignement staging

- [x] Déployer `7e148d4` puis hotfix `b32145b` (EOF 400|404 → 404 client).
- [x] Smoke MVP ; PACS wake → ready → preview/file.
- [ ] Re-smoke frame hors plage → **404** après deploy `b32145b`.
- Docs P2 + checklist G5/G6 : faits.

### P2.1 — Viewer clinique (G3) — Cornerstone3D *(en cours)*

1. [x] Spike CSP : `wasm-unsafe-eval` + `worker-src 'self' blob:` (pas `*`).
2. [x] Packages `@cornerstonejs/*` + Vite exclude loader ; flag `NUXT_PUBLIC_PACS_VIEWER_ENGINE` (défaut **canvas**).
3. [x] Metadata proxy Go `GET …/instances/{id}/metadata` (PixelSpacing, W/L tags) — BFF Nuxt.
4. [x] `CornerstoneDicomViewer.vue` opt-in (wadouri → BFF `/file`) ; canvas reste défaut.
5. [ ] Tools Cornerstone (pan/zoom/W/L HU/Length) + e2e smoke engine=cornerstone.
6. Authz `TestPacs*` inchangée.

**Done P2.1** : W/L HU démontrable via Cornerstone en staging avec flag on ; e2e non soft-skip.

### P2.2 — Mesures calibrées (G4)

- Lire `PixelSpacing` (ou `ImagerPixelSpacing`) via Orthanc tags / metadata.
- Mesure longueur en **mm** (fallback px si spacing absent + hint UI).
- i18n unités ; test unitaire helper calibration ; e2e mesure non-régression.

**Done** : mesure affiche mm sur fixture avec spacing ; px documenté si absent.

### P2.3 — Filet anti-régression GA

| Surface | Ajout |
|---------|--------|
| Go | Metadata / WADO proxy si nouveau ; garder isolation tenant |
| Vitest | Helper spacing + (si A) smoke mount CSP-safe |
| Playwright `@p0` | Download, frame next (404 EOF), admin wake ; si A : ouverture Cornerstone |
| Doc | `15-PLAN-TESTS.md` section PACS |

**Done** : `make test-go` (`TestPacs*`) + `make test-e2e-p0` verts avec flag on.

### P2.4 — Produit & go-live (G1, G2, G8)

1. **G1** — Note produit écrite (GA vs rester `dev`) dans ce fichier ou ADR court.
2. **G2** — Doc opt-in prod `PACS_ENABLED` / `NUXT_PUBLIC_PACS_ENABLED` (défaut **off** prod) + runbook smoke.
3. Si GA décidé :
   - Retirer `nav.tagDev` / badges Imagerie + admin (règle modules-tag-dev).
   - **G8** — UC commercial `useCase/` + `make usecases-sync` + session démo README.
4. Sinon : rester `dev`, pas de UC inventé.

**Done** : décision tracée ; si GA → badge retiré + UC sync + flags prod documentés.

---

## Critères GA restants (après P2)

| # | Critère | Entrée P2 |
|---|---------|-----------|
| G1 | Décision produit | P2.4 |
| G2 | Flag prod opt-in | P2.4 |
| G3 | W/L clinique / Cornerstone | P2.1 |
| G4 | Mesures calibrées | P2.2 |
| G5–G7 | Multi-série/frame, download, erreurs | **Déjà OK** |
| G8 | Use case commercial | P2.4 (si GA) |
| G9 | Index doc | Maintenir `40` + ce plan |
| G10 | Hors V1 | Toujours hors |

---

## Ordre d’exécution proposé

1. ~~Valider Option A vs B~~ → **A acté**.
2. Exécuter **P2.0** (deploy `7e148d4` + smoke).
3. **P2.1 → P2.2 → P2.3** sur branche `feat/pacs-ga` (ou équivalent) depuis `staging`.
4. Review + staging deploy + smoke.
5. **P2.4** uniquement après OK produit.

## Estimation indicative

| Phase | Ordre de grandeur |
|-------|-------------------|
| P2.0 | 0,5 j |
| P2.1 Cornerstone | 3–6 j |
| P2.2 | 1–2 j |
| P2.3 | 1–2 j |
| P2.4 | 0,5–1 j |

---

## Validation

- [x] Arbitrage **A (Cornerstone3D)** acté (2026-07-31)
- [ ] Démarrer P2.0 (deploy review + smoke)
- [ ] P2.1+ uniquement après P2.0 vert
