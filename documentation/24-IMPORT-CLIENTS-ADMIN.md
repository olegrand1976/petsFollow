# 24 — Import clients admin (CSV / Excel + Gemini)

## Objectif

Permettre à un **admin** d’importer la liste clients d’un vétérinaire (CSV / XLS / XLSX) afin d’éviter le ré-encodage manuel. Les **animaux** restent créés par le client dans l’app.

## Parcours

1. Admin choisit un véto + upload fichier → staging (`practice.client_import_jobs` / `client_import_rows`).
2. Mapping colonnes : suggestion **Gemini** (`GEMINI_API_KEY`) ou mapping manuel.
3. Prévisualisation + correction / exclusion de lignes.
4. **Validation manuelle** (commit) → création des comptes via `CreateClientForVet` avec `SkipJourney=true`.
5. Téléchargement one-shot du CSV des mots de passe temporaires (TTL 24 h).

## Règles produit

| Règle | Comportement |
|-------|----------------|
| Périmètre | Clients uniquement (`email`, `fullName`, `locale` optionnelle) |
| Doublons email | Ligne en erreur (`email_already_exists`) — pas de liaison auto |
| Emails | Aucun `send-app-link`, aucun enrôlement journey à l’import |
| MDP | Aléatoire 16 car. + `must_change_password=true` |
| Formats | CSV (`;` ou `,`) · XLSX · max 5 Mo / 2000 lignes |

## API

Préfixe : `/api/v1/admin/client-imports` (rôle admin).

| Méthode | Route |
|---------|-------|
| `POST` | `/` multipart `file` + `vetUserId` |
| `GET` | `/` liste |
| `GET` | `/{id}` détail + rows |
| `POST` | `/{id}/suggest-mapping` |
| `PUT` | `/{id}/mapping` |
| `PATCH` | `/{id}/rows/{rowId}` |
| `POST` | `/{id}/commit` |
| `GET` | `/{id}/credentials?token=` |

BFF Nuxt : `/api/admin/client-imports/*` · UI : `/admin/client-imports`.

## Config

```bash
GEMINI_API_KEY=...          # requis pour suggest-mapping
GEMINI_MODEL=gemini-3.5-flash
```

Sans clé Gemini, le mapping manuel reste possible.

### Staging GCP

- Secret Manager : `petsfollow-gemini-api-key` → Cloud Run `GEMINI_API_KEY` (via `pf_api_secrets`)
- Env non secrète : `GEMINI_MODEL` (via `pf_write_api_env_file`, défaut `gemini-3.5-flash`)

```bash
# Mettre à jour la clé depuis le .env local
source .env
echo -n "$GEMINI_API_KEY" | gcloud secrets versions add petsfollow-gemini-api-key \
  --data-file=- --project=premedica-prod-2025
```

## Code

- Store : `go/internal/store/client_import.go`
- Handlers : `go/internal/handlers/client_import.go`
- Parser : `go/internal/platform/spreadsheet`
- Gemini : `go/internal/platform/gemini`
- Migration : `000028_client_import`
