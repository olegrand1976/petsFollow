# Charte graphique

## Dual-face unifiée

| Face | App | Mode | Primary |
|------|-----|------|---------|
| **petsFollow Pro** | nuxtjs/ | Light | `#1B3A4B` |
| **petsFollow pets** | flutter/ | Dark | `#52B788` |

## Logo

Emblème patte friendly + pulse ECG : `brand/emblem/petsfollow-emblem.svg`  
Servi côté Nuxt : `/brand/emblem.svg`

Variantes composant `PetsFollowLogo` : `default`, `compact` (topbar), `hero` (login animé).

## Shell authentifié (topbar)

```
┌──────────┬─────────────────────────────────────────────┐
│ Sidebar  │ Topbar : logo · notifications · profil      │
│ (nav)    ├─────────────────────────────────────────────┤
│          │ Contenu page                                │
└──────────┴─────────────────────────────────────────────┘
```

- `ProTopbar` : thème (`light_mode` / `dark_mode`), cloche (`notifications`), menu profil, déconnexion
- `ProSidebar` : navigation avec icônes Material Symbols via `ProIcon`
- Icônes UI : **Material Symbols Outlined** (Google Fonts) — composant `ProIcon`

## Composants Pro (`nuxtjs/components/pro/`)

| Composant | Rôle |
|-----------|------|
| `ProTopbar` | Header logo + notifs + profil |
| `ProIcon` | Icône Material Symbols Outlined |
| `ProPageHeader` | Titre + sous-titre + actions |
| `ProListToolbar` | Filtres + bascule vue |
| `ProViewToggle` | Liste / Kanban |
| `ProKanban` / `ProKanbanColumn` | Board colonnes |
| `ProCard` | Surface blanche |
| `ProButton` | CTA |
| `ProInput` | Champ accessible |
| `ProTable` | Table + empty state |
| `ProKpi` | KPI admin |
| `ProBadge` | Statuts colorés |
| `ProEmptyState` | État vide |
| `ProSidebar` | Nav latérale |

## Listes (table / kanban)

- **Clients** : filtres recherche, animaux, tri ; kanban Sans dossier / 1 animal / Multi
- **Admin inscriptions** : filtres rôle + paiement ; kanban par rôle
- Persistance vue : `localStorage` via `useListView`

## Tokens Pro (ajouts)

- `--pf-vet-gradient-main` — fond zone principale
- `--pf-vet-shadow-hover` — hover cartes

## Typographie

- UI : **DM Sans** (`--pf-font-sans`)
- Données BPM / KPI : **IBM Plex Mono** (`--pf-font-mono`)
- Icônes : **Material Symbols Outlined** (`ProIcon`)
- Titres : 600–700 ; corps : 400

### Échelle typographique Pro

| Élément | Taille |
|---------|--------|
| h1 page | 1.75rem |
| h2 card | 1.05rem |
| corps | 1rem |
| muted / label | 0.8125–0.9375rem |
| badge | 0.75rem |

## Tokens

Source : `brand/tokens/design-tokens.json`  
Sync : `make brand-sync` → `nuxtjs/assets/css/tokens.css`, `flutter/lib/core/theme/app_colors.dart`

### Tokens Pro (`pro` dans design-tokens.json)

- `--pf-vet-shadow-sm`, `--pf-vet-shadow-md`
- `--pf-vet-radius-lg` (12px), `--pf-vet-radius-xl` (16px)
- `--pf-vet-sidebar-width`, `--pf-vet-content-max`
- `--pf-vet-gradient-login` — navy → teal
- `--pf-vet-focus-ring` — teal 2px

## Layout login split-screen

```
┌─────────────────────┬──────────────────┐
│  Panneau brand      │  Formulaire      │
│  gradient navy/teal │  logo + titre    │
│  emblème SVG        │  email / mdp     │
│  accroche Pro       │  CTA teal        │
└─────────────────────┴──────────────────┘
```

Classes : `.pro-login-page`, `.pro-login-brand`, `.pro-login-form-panel`  
Mobile (&lt; 960px) : brand en bandeau top, formulaire pleine largeur.

## Grille spacing

- Padding main : `2rem` (desktop), `1.25rem` (mobile)
- Gap grilles KPI : `1rem` — `minmax(180px, 1fr)`
- Gap grilles 2 col : `minmax(280px, 1fr)`
- Touch targets : min 44px (boutons, inputs)

## Badges statuts

| Contexte | Variant | Exemple |
|----------|---------|---------|
| Paiement OK | `success` | actif, payé |
| En attente | `warning` | pending |
| Impayé / alerte BPM | `danger` | past_due, Alerte |
| Neutre | `neutral` | rôle, N/A |

## Palette Pro

| Token | Valeur | Usage |
|-------|--------|-------|
| Primary | `#1B3A4B` | Sidebar, titres |
| Accent | `#2A9D8F` | CTA, liens actifs |
| Alert | `#F4A261` | Alertes BPM |
| Surface | `#FFFFFF` | Cartes |
| BG | `#F8FAFB` | Fond page |

## Règle Cursor

Voir `.cursor/rules/petsfollow-pro-ui.mdc` pour les conventions agent.
