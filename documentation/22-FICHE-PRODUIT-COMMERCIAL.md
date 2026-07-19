# Fiche produit — petsFollow (commerciaux)

**Scope** : présenter le produit, l’offre et le pitch. Rémunération → [19](19-FICHE-COMMISSION-COMMERCIAL.md). Playbook → [21](21-GTM-COMMERCIAL.md).

## En une phrase

Logiciel de **suivi cardiaque prescrit par le véto**, **sans boîtier**. Pro gratuit pour le cabinet ; le client paie (~2–3 €/mois).

## Pour qui / modèle

| Qui | Rôle | Surface |
|-----|------|---------|
| Cabinet véto | Prescripteur (B2B) | Pro web |
| Propriétaire | Payeur (B2B2C) | App mobile pets |
| Commercial | Apporteur | Pro web (espace commercial) |

Monétisation : **cabinet gratuit** → activation clients payants.

## Ce que ça fait (vendable aujourd’hui)

- Création / suivi des animaux
- Relevé cardiaque **15 / 30 / 60 s** (tap, sans hardware)
- Messagerie client ↔ véto (+ mode indisponible)
- Timeline historique (messages, relevés validés)
- Langues **FR / NL / EN / ES**
- Push FCM : message véto → client, confirmation RDV (détail [08](08-MESSAGERIE-NOTIFICATIONS.md))

Ne pas promettre : hardware, WebSocket temps réel (refresh via ouverture app / push).

## Offre à pitcher

Prix **TTC** client. Steer = **triennial**.

| Plan | Prix TTC | ≈ / mois | Message |
|------|----------|----------|---------|
| Annual | 35 € / an | 2,9 € | Entrée |
| **Triennial** | **95 € / 3 ans** | **2,6 €** | **Recommandé** |
| Quinquennial | 145 € / 5 ans | 2,4 € | Engagement long |

| Addon (abo annuel récurrent) | Prix TTC | Pitch |
|-------|----------|-------|
| Family | **39 € / an** | Dès 2 animaux ; vue foyer ; **−10 %** sur abos suivants ; pas de plafond |
| Kennel | **119 € / an** | Dès 6 animaux ; encodage rapide ; **−15 %** ; **exclusif** Family (upgrade) |
| Care+ | **19 € / an** | Médicaments / rappels perso |
| Horse | **39 € / an** | Pack équine (si ≥1 cheval) |

Détail économique → [17](17-POLITIQUE-TARIFAIRE.md).

Commission indicative triennial (plafond) : **~9,4 €** pour vous **et** pour le véto — grille complète [19](19-FICHE-COMMISSION-COMMERCIAL.md).

## Parcours de vente (3 étapes)

1. **Ouvrir** le cabinet (inscription / assignation)
2. **Onboard** profil cabinet complet (Pro)
3. **Activer** pets payants — objectif **5 pets / 60 j** (bonus ramp 25 €)

Détail 30 jours + SPIFF → [21](21-GTM-COMMERCIAL.md).

## Scripts

| Audience | Script |
|----------|--------|
| **Véto (30 s)** | « petsFollow : suivi cardiaque prescrit, sans boîtier. Vos clients paient moins de 3 €/mois ; vous gagnez sur chaque activation — jusqu’à ~9,4 € sur le triennial. Même plafond si un commercial vous a apporté. » |
| **Client (via véto)** | « Suivi prescrit par votre véto, sans boîtier — moins de 3 €/mois. » |
| **Vous (interne)** | « Ouvrez le cabinet, activez 5 pets en 60 j → bonus 25 € + commissions récurrentes. Steer triennial. » |

### Objections

| Objection | Réponse |
|-----------|---------|
| « Encore un abonnement » | Sans boîtier, prescrit par le véto, **&lt; 3 €/mois** (triennial ~2,6 €). |
| « Je perds s’il y a un commercial » | **Non** — même plafond véto avec ou sans commercial. |
| « Il faut un appareil ? » | **Non** — relevé au doigt dans l’app (durée définie par le cabinet). |

## Démo terrain

| Rôle | Email | Mot de passe |
|------|-------|--------------|
| Véto | `vet.demo@petsfollow.test` | `VetDemo123!` |
| Client | `client.demo@petsfollow.test` | `ClientDemo123!` |
| Commercial | `commercial.demo@petsfollow.test` | `CommercialDemo123!` |

Local : Pro **http://localhost:3002** · API **http://localhost:8291**  
(`make up-infra && make migrate && make seed && make api-dev` + `make nuxtjs-dev`)

## Interdits (à bien comprendre)

Un cabinet **commence forcément à 0 animal payant** — c’est normal. Ce qui est interdit, c’est de **confondre inscription et revenu**.

| Interdit | Pourquoi |
|----------|----------|
| **Compter une commission (ou un « deal gagné ») dès l’inscription du véto** | L’ouverture du cabinet = étape 1. Vous êtes payé quand un **animal passe payant** (abonnement client). Tant qu’il n’y a pas d’activation → **0 €**. Le bonus ramp (25 €) exige **5 pets payants / 60 j**, pas juste un compte créé. |
| **Promettre un % calculé sur le prix TTC** | Le client paie en TTC (ex. 95 €). Votre commission = **% du HTVA** uniquement (hors TVA 21 %). Dire « 12 % de 95 € » est faux. |
| **Dire au véto qu’il gagne moins parce qu’un commercial l’a apporté** | Les grilles sont **indépendantes**. Le véto n’est **pas pénalisé** si vous êtes assigné ; même plafond (~9,4 € sur le triennial). |
| **Promettre un boîtier ou du chat temps réel type WebSocket** | Vendable = app + Pro (relevé, messagerie, timeline, addons, push FCM messages/RDV). Hardware et WebSocket = **pas livré**. |

## Liens

| Doc | Contenu |
|-----|---------|
| [14](14-POSITIONNEMENT-MARKETING.md) | Positionnement |
| [17](17-POLITIQUE-TARIFAIRE.md) | Grille prix + économie |
| [18](18-FICHE-COMMISSION-VETO.md) | Commission véto (co-selling) |
| [19](19-FICHE-COMMISSION-COMMERCIAL.md) | Votre rémunération |
| [21](21-GTM-COMMERCIAL.md) | Playbook 30 j + SPIFF |
