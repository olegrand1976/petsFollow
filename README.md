# petsFollow

Suivi vétérinaire pour animaux de compagnie — MVP.

## Stack

| Composant | Dossier | Rôle |
|-----------|---------|------|
| API Go | `go/` | Backend modulaire (identity, pets, heartrate, messaging, timeline) |
| Web véto | `nuxtjs/` | petsFollow Pro |
| Mobile client | `flutter/` | petsFollow pets |
| Brand | `brand/` | Tokens + logo unifié |

## Démarrage local

```bash
make env
make up
make smoke
```

- API: http://localhost:8291
- Web véto: http://localhost:3002
- MailHog: http://localhost:8026

### Comptes demo (seed)

| Rôle | Email | Mot de passe |
|------|-------|--------------|
| Véto | vet.demo@petsfollow.test | VetDemo123! |
| Client | client.demo@petsfollow.test | ClientDemo123! |

## Staging GCP

- Web: https://petsfollow.ll-it-sc.be
- API: https://api.petsfollow.ll-it-sc.be

## Documentation

Voir [documentation/README.md](documentation/README.md)

## MVP

- Création animal (client)
- Suivi clients + animaux (véto)
- Messagerie interne + mode indisponible
- Relevé cardiaque 60s (valider / recommencer)
- Historique timeline
