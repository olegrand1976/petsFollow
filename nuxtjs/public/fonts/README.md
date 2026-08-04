# Polices auto-hébergées

Aucune requête vers `fonts.googleapis.com` (RGPD) — voir `assets/css/fonts.css`
et `buildCsp()` (`font-src 'self'`).

| Fichier(s) | Famille CSS | Plage | Licence |
|---|---|---|---|
| `dm-sans-latin*.woff2` | `DM Sans` | latin, latin-ext | SIL OFL 1.1 |
| `ibm-plex-mono-*-latin*.woff2` | `IBM Plex Mono` | latin, latin-ext | SIL OFL 1.1 |
| `noto-sans-{400,700}-cyrillic.woff2` | `DM Sans` | cyrillic | SIL OFL 1.1 |
| `noto-sans-mono-{400,600}-cyrillic.woff2` | `IBM Plex Mono` | cyrillic | SIL OFL 1.1 |
| `material-symbols-outlined.woff2` | `Material Symbols Outlined` | icônes | Apache 2.0 |

## Pourquoi Noto sous le nom « DM Sans » / « IBM Plex Mono »

DM Sans et IBM Plex Mono ne publient pas de sous-set cyrillique. Les faces Noto
sont déclarées **sous les mêmes noms de famille**, restreintes à la plage
cyrillique par `unicode-range` : le navigateur ne les télécharge que si la page
contient du cyrillique, et le latin continue d'utiliser DM Sans. Aucun token de
marque (`--pf-font`) n'est modifié.

Noto Sans Mono a la même avance monospace qu'IBM Plex Mono (0,600 em) —
l'alignement des colonnes `.text-mono` est donc préservé.

## Régénérer les sous-sets cyrilliques

Nécessite `fonttools[woff]` (+ `brotli`) et les TTF Noto (`fonts-noto-core`) :

```bash
CYR="U+0301,U+0400-045F,U+0490-0491,U+04B0-04B1,U+2116"

pyftsubset /usr/share/fonts/truetype/noto/NotoSans-Regular.ttf \
  --unicodes="$CYR" --flavor=woff2 --layout-features='*' \
  --output-file=public/fonts/noto-sans-400-cyrillic.woff2
# idem NotoSans-Bold.ttf → noto-sans-700-cyrillic.woff2
# idem NotoSansMono-{Regular,Bold}.ttf → noto-sans-mono-{400,600}-cyrillic.woff2
```

Le sous-set est volontairement limité à `cyrillic` (~11 Ko/poids). `cyrillic-ext`
pèse 31 Ko/poids pour des glyphes qu'aucune des deux langues cyrilliques
supportées (uk, ru) n'utilise — à ajouter seulement si une langue le nécessite
(kazakh, ouzbek…).

## Preload

Seuls `dm-sans-latin.woff2` et la police d'icônes sont préchargés
(`nuxt.config.ts`). Les sous-sets cyrilliques ne le sont **pas** : cela
pénaliserait les locales latines, qui ne les chargent jamais.
