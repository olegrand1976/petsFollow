# UC-VP-12 — RDV rapide Nouveau client puis identification

| Champ | Valeur |
|-------|--------|
| ID | UC-VP-12 |
| Profil | VetPro |
| Priorité | P1 |
| Surfaces | Agenda Pro · Consultation |
| Tag | — |

## Objectif

Prendre un rendez-vous (ou démarrer une consultation) **sans connaître encore le client**, grâce au créneau système **Nouveau client / Nouvel animal**, puis **identifier** l’identité réelle avant le compte-rendu.

## Prérequis

- Compte véto seed (`vet.demo@petsfollow.test` / `VetDemo123!`)
- Cabinet avec le couple placeholder (créé automatiquement à la création du cabinet / au listage clients)

## Étapes

1. Ouvrir **Agenda** → **Nouveau RDV**.
2. Constater que **Nouveau client** (badge « À identifier ») et **Nouvel animal** sont pré-sélectionnés.
3. Saisir le **téléphone de rappel** (obligatoire) + créneau + type de visite → créer le RDV.
4. Ouvrir le détail du RDV → pastille « À identifier » + téléphone cliquable → **Nouvelle consultation**.
5. L’écran **Identifier le client** s’affiche (pas le CR) :
   - **Créer le client** : prénom, nom, téléphone (+ email optionnel) + animal, **ou**
   - **Client connu** : recherche → confirmation explicite → animal existant ou nouveau.
6. Valider → le workspace CR s’ouvre sur l’identité réelle.
7. Les placeholders **Nouveau client / Nouvel animal** restent disponibles pour le prochain cas (non modifiés).

## Résultat attendu

- RDV bookable en quelques secondes sans encoder une fausse identité.
- Impossible de finaliser un CR / DAF / facture tant que la visite pointe encore sur le placeholder.
- Après identification, la visite est rattachée au vrai animal ; le créneau système est réutilisable.

## Variante walk-in

Depuis **Clients** → ouvrir consultation sur **Nouveau client** → démarrer → même gate d’identification.
