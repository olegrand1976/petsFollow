# Simulation prospection admin — 10 ans

> **Statut : backlog / non livré.** Aucune page `/admin/simulation` en code à ce jour. Document = plan d’implémentation uniquement.

Plan d’implémentation (hors correctifs flux API). Décisions figées faute de réponse utilisateur :

| Choix | Décision |
|-------|----------|
| **1A** | Page dédiée `/admin/simulation`, calcul **100 % navigateur** (pas d’API / pas de persistance) |
| **2B** | Revenus abonnements **3 plans** (3,50 €/mois · 35 €/an · 95 €/3 ans) + option affichage commission commercial (grille plan). *Ancien mix quinquennial + attach rates addons = obsolète (hors vente).* |

## Objectif

Permettre à un admin de projeter sur **10 ans** :
- revenus annuels (nouveaux + renouvellements)
- **cumul** des renouvellements
- multi-animaux par client
- mix de plans et taux de renouvellement

## UI

- Nav admin : entrée « Simulation » → [`nuxtjs/layouts/admin.vue`](../nuxtjs/layouts/admin.vue)
- Page [`nuxtjs/pages/admin/simulation.vue`](../nuxtjs/pages/admin/simulation.vue) : `ProPageHeader` + formulaire hypothèses + `ProKpi` + tableau année 1…10
- i18n FR / EN / NL sous `admin.simulation.*`
- Export CSV optionnel (client-side)

## Hypothèses (inputs)

| Champ | Défaut suggéré |
|-------|----------------|
| Cabinets acquis / an | 10 |
| Croissance cabinets (% / an) | 0 |
| Clients payants / cabinet (an 1) | 40 |
| Animaux moyens / client | 1.5 |
| Mix monthly / annual / triennial (%) | 15 / 25 / 60 |
| Taux de renouvellement en fin de période (%) | 80 |
| Afficher commission commercial (grille plan) | off |

Tarifs figés alignés billing Go : monthly 350 ct/mois · annual 3500 · triennial 9500 ct. Quinquennial + addons Family / Kennel / Care+ / Horse = **hors vente** — ne plus modéliser d’attach rate addon.

## Moteur

- Fichier pur TS : [`nuxtjs/utils/prospectionSimulation.ts`](../nuxtjs/utils/prospectionSimulation.ts) (+ test Vitest)
- Modèle cohortes : chaque année N, nouveaux pets × mix plans → cash N ; renouvellements aux échéances 1 mois / 1 an / 3 ans selon plan, avec attrition
- Sorties par année : `newRevenue`, `renewalRevenue`, `total`, `cumulativeRenewals`, `cumulativeTotal` (plus de `addonRevenue`)
- KPI : CA 10 ans, part renouvellements, animaux actifs fin an 10

## Hors scope

- Persistance scénarios (1B)
- Appels API / seed
- Flutter

## Tests

- Vitest unitaire sur le moteur (cas mono-animal annual 100 % renouvellement ; multi-animaux ; mix plans)
- Playwright smoke : admin ouvre `/admin/simulation`, KPIs visibles
