# AGENTS.md — petsFollow

## Projet

Monorepo **petsFollow** : suivi cardiaque vétérinaire dual-face.

| Face | Stack | Port dev |
|------|-------|----------|
| **Pro** (véto/admin) | Nuxt 3 (`nuxtjs/`) | **3002** |
| **pets** (clients) | Flutter (`flutter/`) | — |
| **API** | Go (`go/`) | **8291** |

## Démarrage local

```bash
# Terminal 1 — infra + API
make up-infra && make migrate && make seed && make api-dev

# Terminal 2 — Nuxt Pro
make nuxtjs-dev   # http://localhost:3002
```

Après modification des tokens brand : `make brand-sync`.

## Comptes démo

Mot de passe commun véto : `VetDemo123!` · client : `ClientDemo123!` · admin : `AdminDemo123!`

| Rôle | Email | Cabinet |
|------|-------|---------|
| Véto (démo) | `vet.demo@petsfollow.test` | Cabinet VetPlus Demo |
| Véto | `vet.parc@petsfollow.test` | Clinique du Parc |
| Véto | `vet.lyon@petsfollow.test` | Centre Cardio Animaux Lyon |
| Admin | `admin.demo@petsfollow.test` | — (global) |
| Client (Flutter) | `client.demo@petsfollow.test` | VetPlus — Rex, Bella |
| Client | `client.marie@petsfollow.test` | Parc — Mimi, Chouchou |
| Client | `client.paul@petsfollow.test` | Parc — Max |
| Client | `client.julie@petsfollow.test` | Lyon — Oscar |
| Client | `client.thomas@petsfollow.test` | Lyon — Luna, Nico (pending) |

Relancer les données : `make seed`

## UI Pro (Nuxt)

- Design system : composants `Pro*` dans `nuxtjs/components/pro/`
- Logo : `components/PetsFollowLogo.vue` (variants default/compact/hero)
- Shell : `ProSidebar` + `ProTopbar` (notifs véto uniquement)
- Listes : `ProListToolbar` + bascule table/kanban (`useListView`)
- CSS : `nuxtjs/assets/css/pro-*.css` + tokens `--pf-vet-*`
- Règle Cursor : `.cursor/rules/petsfollow-pro-ui.mdc`
- Charte : `documentation/13-CHARTE-GRAPHIQUE.md`

**Ne pas** mélanger le thème dark Flutter dans Pro.

## API

- Base : `http://localhost:8291/api/v1`
- Réponses enveloppées `{ data: ... }` — BFF Nuxt proxy tel quel
- Côté pages : `const items = res.data ?? res`

## Tests

```bash
cd nuxtjs && npm run build
cd nuxtjs && npm run test:e2e   # Playwright
make smoke
```

## Structure clé

- `go/internal/handlers/` — routes API
- `nuxtjs/pages/` — pages Pro (véto + admin)
- `nuxtjs/server/api/` — BFF Nuxt
- `brand/tokens/design-tokens.json` — source tokens
