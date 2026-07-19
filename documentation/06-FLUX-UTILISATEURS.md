# Flux utilisateurs — petsFollow

## Rôles

| Rôle | Surface | Mission |
|------|---------|---------|
| `vet` | Pro | Prescrire / suivre / messagerie |
| `client` | Flutter pets | Animaux, FC, paiement, messages |
| `commercial` | Pro | Apporter cabinets, prospects, activations |
| `admin` | Pro | Ops plateforme, commissions, commercials |

## Parcours véto

```mermaid
flowchart TD
  Reg[Register + confirm email] --> Onb[Onboarding profil cabinet]
  Onb --> Dash[Dashboard / clients]
  Dash --> Msg[Messagerie]
  Dash --> FC[Relevés FC validés]
  Dash --> Comm[Commissions]
  Dash --> Req[Link-requests]
```

## Parcours client

```mermaid
flowchart TD
  Login[Login / Google] --> Pets[Créer animal + plan]
  Pets --> Pay[Stripe Checkout]
  Pay --> Active[Entitlement active]
  Active --> HR[Relevé cardiaque]
  Active --> Thread[Messagerie véto]
  Active --> Addons[Addons Family Care+ Horse]
```

## Parcours commercial

```mermaid
flowchart TD
  LoginC[Login commercial] --> Over[Overview]
  Over --> Pros[Prospects CRM]
  Over --> Encode[Encoder / suivre vétos]
  Encode --> Act[Activer pets payants]
  Act --> Earn[Commissions ledger]
```

## Parcours admin

Login → métriques → users / payments → commercials (créer, assigner) → clôture périodes commissions véto & commercial.

## Démo

Comptes seed : [AGENTS.md](../AGENTS.md) · fiche produit commercial : [22](22-FICHE-PRODUIT-COMMERCIAL.md).
