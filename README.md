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

Mot de passe véto : `VetDemo123!` · client : `ClientDemo123!` · admin : `AdminDemo123!`

| Rôle | Email | Notes |
|------|-------|-------|
| Véto | vet.demo@petsfollow.test | VetPlus — profil complet |
| Véto | vet.parc@petsfollow.test | Clinique du Parc |
| Véto | vet.lyon@petsfollow.test | Lyon — indisponible |
| Véto | vet.onboarding@petsfollow.test | Profil cabinet à compléter |
| Véto | vet.unverified@petsfollow.test | Email non confirmé |
| Admin | admin.demo@petsfollow.test | — |
| Client | client.demo@petsfollow.test | Rex + Bella |
| Client | client.vide@petsfollow.test | Sans animal (kanban) |
| Client | client.marie@petsfollow.test | Mimi + Chouchou |
| Client | client.paul@petsfollow.test | Max |
| Client | client.julie@petsfollow.test | Oscar |
| Client | client.thomas@petsfollow.test | Luna + Nico (pending) |

Confirm email : `http://localhost:3002/confirm-email?token=demo-confirm-email`

Relancer les données : `make seed`

## Staging GCP

- Web: https://petsfollow.ll-it-sc.be
- API: https://api.petsfollow.ll-it-sc.be

## Documentation

Voir [documentation/README.md](documentation/README.md)

## MVP

- Création animal (client)
- Suivi clients + animaux (véto)
- Messagerie interne + mode indisponible
- Relevé cardiaque (durée configurée par le véto : 15/30/60 s)
- Historique timeline
