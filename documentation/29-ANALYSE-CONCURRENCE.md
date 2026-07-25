# 29 — Analyse concurrence (FR / BE / ES)

## Objectif

Fournir une base **analyste marché** pour le pitch commercial et le positionnement petsFollow face aux acteurs PMS, apps propriétaires, wearables et téléconsult — sans dénigrement, avec des points de comparaison actionnables (`wins` / `watchouts` / `talkTrack`).

## Source de données

Fichier machine-readable :

[`nuxtjs/data/competition/markets.json`](../nuxtjs/data/competition/markets.json)

- `lastReviewedAt` : **2026-07-25**
- Marchés : `fr`, `be`, `es` (5 produits chacun)
- Bannières petsFollow + synthèse par marché (FR/EN)
- Catégories : `pms` | `telemed` | `wearables` | `owner_app` | `insurance_tech` | `other`

## Méthodologie

1. **Périmètre** : logiciels et services visibles par un cabinet ou un propriétaire en France, Belgique et Espagne (présence locale ou pan-européenne réelle).
2. **Sources** : sites publics, grilles affichées, observations terrain commercial, documentation interne tarifaire petsFollow ([17-POLITIQUE-TARIFAIRE](17-POLITIQUE-TARIFAIRE.md), [22-FICHE-PRODUIT-COMMERCIAL](22-FICHE-PRODUIT-COMMERCIAL.md)).
3. **Prix** : si non publics → `rangeHint` / notes = **« sur devis »**. Pas d’invention de montants cabinet.
4. **Ton** : factuel, complémentaire vs substitutif quand le concurrent est un PMS ; distinguer clairement **GPS grand public** vs **suivi cardiaque prescrit**.
5. **Angle petsFollow (récurrent)** : pas de boîtier · prescription véto · B2B2C · Pro **69 € HT/mois** (+ setup) · client **3,50 / 35 / 95 €** · Care/Horse inclus · 5 langues (FR/NL/EN/ES/ET).

## Caveats

- Les parts de marché et années de fondation sont **indicatives** (ordre de grandeur analyste), pas un audit financier.
- Les offres évoluent vite (modules, bundles réseaux) : retraiter `lastReviewedAt` à chaque revue.
- Un même vendor peut apparaître sur plusieurs marchés (ex. Tractive) avec des notes de présence différentes.
- Ce document **ne remplace pas** une due diligence juridique / réglementaire (téléconsult, dispositifs médicaux).

## Revue

| Champ | Valeur |
|-------|--------|
| Dernière revue | 2026-07-25 |
| Prochaine revue suggérée | ≤ 6 mois ou avant campagne commerciale majeure |
