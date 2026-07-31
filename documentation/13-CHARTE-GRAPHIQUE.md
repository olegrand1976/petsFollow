# Charte graphique — petsFollow Pro (Web)

**Scope** : application web en ligne **petsFollow Pro** (`nuxtjs/`) — cabinets vétérinaires, commerciaux, admin.  
**Hors scope** : app mobile Flutter **pets** (dark) et **Pro Light** — mêmes couleurs de marque, autre shell.

Source de vérité tokens : [`brand/tokens/design-tokens.json`](../brand/tokens/design-tokens.json) → `make brand-sync` → [`nuxtjs/assets/css/tokens.css`](../nuxtjs/assets/css/tokens.css).  
Règle agent : [`.cursor/rules/petsfollow-pro-ui.mdc`](../.cursor/rules/petsfollow-pro-ui.mdc).

---

## 1. Positionnement visuel

| Axe | Direction |
|-----|-----------|
| Promesse | Santé animale professionnelle, confiance cabinet, clarté clinique |
| Ton | Clair, calme, moderne — **pas** gadget wellness, **pas** dark mode « tech » par défaut |
| Mode par défaut | **Light** (navy + teal sur fond clair) |
| Mode optionnel | Dark (toggle topbar) — mêmes accents teal / coral |
| À éviter | Violet/indigo générique IA, crème + serif terracotta, glow néon, emojis UI, Google Fonts CDN |

**Dual-face marque**

| Face | Surface | Mode | Primary |
|------|---------|------|---------|
| **Pro (Web)** | `nuxtjs/` | Light (+ dark opt.) | Navy `#1B3A4B` |
| **pets (mobile)** | `flutter/` | Dark | Sage `#52B788` |

Ne **pas** importer la palette / le look dark Flutter dans le shell Pro.

---

## 2. Logo & emblème

- Mark circulaire (chien + mains + pulse + wordmark) — `brand/logo/petsfollow-mark.png`
- Servi Nuxt : `/brand/logo-mark.png` (favicon `/brand/favicon-192.png`)
- Flutter : `assets/brand/petsfollow-mark.png` (+ icônes launcher Android/iOS)
- PDF consultation (partage CR) : `go/internal/platform/consultationpdf/assets/emblem.png` (copie du mark via `make brand-sync`)
- L’ancien emblème SVG patte (`brand/emblem/petsfollow-emblem.svg`) est conservé en archive, plus servi en UI ni dans les PDF.
- Composant : `PetsFollowLogo`

| Variant | Usage |
|---------|--------|
| `default` | Pages marketing / contenu |
| `compact` | Topbar authentifiée |
| `hero` (+ `animated`) | Panneau brand login |

Le logo est le **seul** SVG brand autorisé en UI ; le reste des icônes = Material Symbols.

---

## 3. Palette

### 3.1 Couleurs de marque (brand)

| Nom | Hex | Rôle |
|-----|-----|------|
| **Navy** | `#1B3A4B` | Autorité, titres, sidebar, texte principal |
| **Teal** | `#2A9D8F` | Accent, CTA, liens, focus, navigation active |
| **Gold** | `#E9C46A` | Mise en avant douce / warning light |
| **Sage** | `#52B788` | Succès, pont vers la face pets |
| **Coral** | `#F4A261` | Accent secondaire / alerte douce |

### 3.2 Tokens Pro Web (`--pf-vet-*`)

| Token CSS | Valeur | Usage |
|-----------|--------|-------|
| `--pf-vet-primary` | `#1B3A4B` | Titres, sidebar, texte fort |
| `--pf-vet-accent` | `#2A9D8F` | Boutons primary, liens, états actifs |
| `--pf-vet-alert` | `#E76F51` | Erreurs, danger, BPM hors seuil |
| `--pf-vet-text` | `#1B3A4B` | Corps |
| `--pf-vet-text-muted` | `#6B7280` | Labels, aides, métadonnées |
| `--pf-vet-bg` | `#F7F9FB` | Fond page |
| `--pf-vet-surface` | `#FFFFFF` | Cartes, inputs, topbar |
| `--pf-vet-border` | `#E2E6ED` | Séparateurs, contours |

Variables marque exposées aussi : `--pf-brand-navy`, `--pf-brand-teal`, `--pf-brand-gold`, `--pf-brand-sage`, `--pf-brand-coral`.

### 3.3 Gradients

| Token | Valeur | Usage |
|-------|--------|-------|
| `--pf-vet-gradient-login` | `135deg, #1B3A4B → #2A9D8F` | Panneau brand login |
| `--pf-vet-gradient-main` | `180deg, #F7F9FB → #FFFFFF 48%` | Zone contenu authentifiée |

### 3.4 Sémantique badges / statuts

| Variant | Couleurs (light) | Exemples |
|---------|------------------|----------|
| `success` | fond sage clair / texte sage | actif, payé, validé |
| `warning` | fond gold clair / texte gold | pending, en attente |
| `danger` | fond coral/alert clair / texte alert | past_due, alerte BPM |
| `neutral` | fond gris / texte muted | rôle, N/A |

### 3.5 Dark mode (optionnel)

Activé via `html.dark` (`pro-theme-dark.css`). Accents **teal** et **alert** conservés ; surfaces passent en navy profond (`#0f1923` / `#1a2733`).  
Default produit et maquettes = **light**.

---

## 4. Typographie

Polices **auto-hébergées** (`nuxtjs/assets/css/fonts.css` + `public/fonts/`) — **interdit** : CDN Google Fonts (RGPD / CSP).

| Rôle | Fonte | Token |
|------|-------|-------|
| UI / titres / corps | **DM Sans** (400–700) | `--pf-font` |
| Données (BPM, KPI, code) | **IBM Plex Mono** (400, 600) | `--pf-font-mono` |
| Icônes | **Material Symbols Outlined** | via `ProIcon` |

### Échelle

| Élément | Taille | Graisse |
|---------|--------|---------|
| h1 page | `1.75rem` | 700 |
| h2 section | `1.25rem` | 700 |
| h3 / titre carte | `1.05rem` | 600 |
| Corps | `1rem` | 400 |
| Bouton / label fort | `0.9375rem` | 600 |
| Muted / meta | `0.8125–0.9375rem` | 400 |
| Badge | `0.75rem` | 600 |

Interlignage corps : `1.5` · Titres : `1.2`  
Couleur titres : `--pf-vet-primary`

---

## 5. Forme, ombre, focus

| Token | Valeur | Usage |
|-------|--------|-------|
| `--pf-vet-radius` | `8px` | Boutons, inputs, badges |
| `--pf-vet-radius-lg` | `12px` | Cartes, boutons topbar |
| `--pf-vet-radius-xl` | `16px` | Surfaces larges |
| `--pf-vet-shadow-sm` | `0 1px 3px rgba(27,58,75,.08)` | Cartes au repos |
| `--pf-vet-shadow-md` | `0 4px 24px rgba(27,58,75,.08)` | Élévation |
| `--pf-vet-shadow-hover` | `0 8px 28px rgba(27,58,75,.12)` | Carte interactive |
| `--pf-vet-focus-ring` | `0 0 0 3px rgba(42,157,143,.35)` | `:focus-visible` |

**Touch targets** : minimum **44px** (boutons, inputs, icônes topbar).

---

## 6. Layout & shell

### Dimensions

| Token | Valeur |
|-------|--------|
| `--pf-vet-sidebar-width` | `260px` |
| `--pf-vet-content-max` | `1200px` |
| Padding main | `2rem` desktop · `1.25rem` mobile |
| Gap KPI | `1rem` · `minmax(180px, 1fr)` |
| Gap 2 colonnes | `minmax(280px, 1fr)` |

### Shell authentifié

```
ProApp
├── ProTopbar (full width)
│     logo compact · nom cabinet · thème · notifs · profil · logout
└── ProAppShell
      ├── ProSidebar (nav + icônes uniquement)
      └── ProAppBody → main (gradient fond)
```

- Topbar sticky, fond semi-transparent + `backdrop-filter`
- **Ne pas** remettre profil / logout dans la sidebar
- Lien nav actif : accent teal + fond léger

### Login (split-screen)

```
┌──────────────────────┬─────────────────┐
│  Brand (gradient)    │  Formulaire     │
│  emblème hero        │  logo + champs  │
│  accroche Pro        │  CTA teal       │
└──────────────────────┴─────────────────┘
```

Classes : `.pro-login-page`, `.pro-login-brand`, `.pro-login-form-panel`  
&lt; 960px : brand en bandeau haut, formulaire pleine largeur.

### Landing / pages publiques

Même palette navy/teal ; composition marketing (hero brand-first). Ne pas mélanger le shell authentifié (sidebar) sur les pages full-page publiques.

---

## 7. Composants design system

Dossier : `nuxtjs/components/pro/` · CSS : `pro-components.css`, `pro-forms.css`, `pro-layout.css`, `pro-base.css`.

| Composant | Rôle visuel |
|-----------|-------------|
| `ProTopbar` | Header sticky |
| `ProSidebar` | Nav latérale 260px |
| `ProIcon` | Material Symbols Outlined |
| `ProPageHeader` | Titre + sous-titre + actions |
| `ProCard` | Surface blanche bordée ; `--interactive` = lift hover |
| `ProButton` | `primary` (teal) / `secondary` (contour) / `ghost` (texte teal) |
| `ProInput` | Champ labelisé, radius 8px |
| `ProTable` | Table + empty state |
| `ProListToolbar` | Filtres + bascule vue |
| `ProViewToggle` | Liste / Kanban |
| `ProKanban` / `ProKanbanColumn` | Board colonnes |
| `ProKpi` | Tuile métrique (mono pour chiffres) |
| `ProBadge` | Statuts sémantiques |
| `ProEmptyState` | État vide illustré (icône Material) |
| `ProCommissionSheet` | Fiche commission |
| `PetsFollowLogo` | Marque |

### Boutons — règles

- **Primary** : fond teal, texte blanc — action principale (1 par zone)
- **Secondary** : contour border, texte navy — action secondaire
- **Ghost** : texte teal, sans fond — actions tertiaires / liens d’action
- Disabled : `opacity: 0.6`

### Cartes

- Fond surface, border 1px, radius 12px, shadow-sm
- Interactive : hover `translateY(-1px)` + shadow-hover
- Titre carte : 1.05rem / 600, séparateur bas optionnel

### Listes

1. `ProListToolbar` + slot `#filters`
2. Bascule table/kanban via `useListView(storageKey)`
3. Kanban clients : 0 / 1 / 2+ animaux
4. Kanban admin users : par rôle

---

## 8. Icôographie

- **Uniquement** Material Symbols Outlined via `ProIcon` (`name` = ligature, ex. `dashboard`, `notifications`, `light_mode`)
- Settings : `FILL` 0, `wght` 400, `opsz` 24
- **Interdit** : emojis UI, SVG maison hors logo, Font Awesome, etc.
- Exceptions : emblème brand, charts (`ProBpmChart`)

---

## 9. Accessibilité & motion

- Focus visible : anneau teal (`--pf-vet-focus-ring`) — ne pas supprimer
- Contraste texte navy / muted sur fond clair (mode light)
- Transitions courtes : `0.1s–0.15s` (hover bouton / carte) — pas d’animation décorative bruyante
- Login hero : animation logo autorisée (`animated`)
- Cibles tactiles ≥ 44px

---

## 10. Do / Don’t

| Do | Don’t |
|----|-------|
| Utiliser `--pf-vet-*` / `--pf-brand-*` | Hardcoder des hex hors tokens |
| CTA en teal accent | CTA violet ou navy plein comme primary bouton |
| Light par défaut pour maquettes | Importer le dark Flutter |
| Material Symbols via `ProIcon` | Emojis / icônes random |
| Fonts auto-hébergées | `fonts.googleapis.com` |
| Une action primary par zone | Empiler plusieurs boutons teal |
| Cartes pour interactions / contenu structuré | Cartes partout sur un hero marketing |

---

## 11. Fichiers de référence

| Fichier | Contenu |
|---------|---------|
| `brand/tokens/design-tokens.json` | Source tokens |
| `nuxtjs/assets/css/tokens.css` | CSS généré |
| `nuxtjs/assets/css/fonts.css` | @font-face |
| `nuxtjs/assets/css/pro-*.css` | Layout, composants, forms, dark |
| `brand/logo/petsfollow-mark.png` | Logo mark (source) |
| `go/.../consultationpdf/assets/emblem.png` | Logo mark embarqué PDF CR |
| `documentation/13-CHARTE-GRAPHIQUE.md` | Ce document |

Sync après modification tokens : `make brand-sync`.

---

*Charte Pro Web — alignée tokens juillet 2026.*
