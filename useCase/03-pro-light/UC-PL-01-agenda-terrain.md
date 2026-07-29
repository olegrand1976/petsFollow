# UC-PL-01 — Agenda terrain Pro Light

| | |
|--|--|
| **ID** | `UC-PL-01` |
| **Durée** | ~10 min |
| **Priorité** | Démo |
| **Surface** | Flutter Pro Light |

## Objectif

Se connecter en care pro / VetLight, voir l’agenda du jour et marquer une visite comme faite.

## Acteurs

| Rôle | Compte | MDP |
|------|--------|-----|
| Farrier (ou vet_light) | `farrier.demo@petsfollow.test` **ou** `vetlight.demo@petsfollow.test` | `CareProDemo123!` |

## Prérequis

- App staging (même APK ; application terrain après login care_pro).
- Seed avec animal partagé (ex. Spirit pour farrier) — voir aussi `UC-X-05`.

## Étapes

1. Se connecter avec `farrier.demo` (ou `vetlight.demo`).
2. Vérifier le shell Pro Light : AppBar **logo petsFollow** seul ; tabs Agenda · Clients · Animaux · **Messages** · Réglages (identique staff / care_pro).
3. Ouvrir **Agenda** → vue **Aujourd’hui** (ou équivalent).
4. Ouvrir une visite / créneau si présent.
5. Marquer **Fait** (ou action équivalente).
6. (Optionnel) Compte rendu : dictée (bandeau micro + Arrêter) ou fichier audio ; **pas** de bouton Améliorer IA ; Enregistrer / Finaliser.
7. **Nouvelle Consultation** (VetLight / farrier avec `write_notes`) : fiche animal → CTA → visite confirmée immédiate (`source=care_pro`, hors overlap agenda) → CR → après enregistrement : CTA Web Pro (DAF / facture) ou **Terminer** (finalise le brouillon CR non vide pour le client).
8. Onglet **Messages** → composer vers le client partagé → envoyer (voir aussi `UC-X-01` étape care_pro).

## Résultat attendu

- Application terrain distincte de l’application propriétaire.
- Agenda utilisable ; action « Fait » prise en compte.
- Messages et CR utilisables sans confusion de boutons.
- Consultation terrain créable hors créneau préexistant (visite `confirmed`).

## Checklist

| | Résultat |
|--|----------|
| Login Pro Light | OK / KO / N/A |
| Agenda Aujourd’hui | OK / KO / N/A |
| Marquer Fait | OK / KO / N/A |
| Nouvelle Consultation + CR | OK / KO / N/A |

## Zone retour

- Bug :
- Friction UX :
- Idée :

→ [`TEMPLATE-RETOUR.md`](../TEMPLATE-RETOUR.md)

---

*Réf. QA : A9, G*
