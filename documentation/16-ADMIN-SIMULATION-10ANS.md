# Simulation prospection admin — 10 ans

Plan d’implémentation (hors correctifs flux API). Décisions figées faute de réponse utilisateur :

| Choix | Décision |
|-------|----------|
| **1A** | Page dédiée `/admin/simulation`, calcul **100 % navigateur** (pas d’API / pas de persistance) |
| **2B** | Revenus abonnements (29 / 79 / 115 €) **+** addons Family / Care+ / Horse + option affichage commission commercial 12 % |

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
| Mix annual / triennial / quinquennial (%) | 25 / 60 / 15 |
| Taux de renouvellement en fin de période (%) | 80 |
| Attach rate Family / Care+ / Horse (% clients ou pets) | 10 / 15 / 5 |
| Afficher commission commercial 12 % | off |

Tarifs figés alignés billing Go : annual 2900 ct, triennial 7900, quinquennial 11500 ; addons 5500 / 1900 / 3900.

## Moteur

- Fichier pur TS : [`nuxtjs/utils/prospectionSimulation.ts`](../nuxtjs/utils/prospectionSimulation.ts) (+ test Vitest)
- Modèle cohortes : chaque année N, nouveaux pets × mix plans → cash N ; renouvellements aux échéances 1/3/5 ans selon plan, avec attrition
- Sorties par année : `newRevenue`, `renewalRevenue`, `addonRevenue`, `total`, `cumulativeRenewals`, `cumulativeTotal`
- KPI : CA 10 ans, part renouvellements, animaux actifs fin an 10

## Hors scope

- Persistance scénarios (1B)
- Appels API / seed
- Flutter

## Tests

- Vitest unitaire sur le moteur (cas mono-animal annual 100 % renouvellement ; multi-animaux ; mix plans)
- Playwright smoke : admin ouvre `/admin/simulation`, KPIs visibles
