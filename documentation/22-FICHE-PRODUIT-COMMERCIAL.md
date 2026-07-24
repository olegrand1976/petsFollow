# Fiche produit — petsFollow (commerciaux)

**Scope** : présenter le produit, l’offre et le pitch. Rémunération → [19](19-FICHE-COMMISSION-COMMERCIAL.md). Playbook → [21](21-GTM-COMMERCIAL.md).

## En une phrase

Logiciel de **suivi cardiaque prescrit par le véto**, **sans boîtier**.  
**Pro** (Web + app clients) facturé **hors ligne** ; **Pro Light** (app mobile **ProLight**) gratuit ; le client paie le suivi animal (~2–3,5 €/mois).

## Pour qui / modèle

| Qui | Rôle | Surface | Tarif |
|-----|------|---------|-------|
| Cabinet véto | Prescripteur (B2B) | **Pro** (Web SaaS) | **69 € HT/mois** + setup 320 € HT — **facturation externe** |
| Pro terrain | App mobile ProLight | **Pro Light** (Flutter, avec ou sans compte Web Pro) | **Gratuit** |
| Propriétaire | Payeur (B2B2C) | App mobile pets | 3,50 / 35 / 95 € TTC (Stripe) |
| Commercial | Apporteur | Pro web (espace commercial) | — |

Monétisation : **SaaS cabinet hors ligne** + **activations clients payantes** (commissions partenaires pouvant compenser le SaaS).

## Ce que ça fait (vendable aujourd’hui)

- Création / suivi des animaux
- Relevé cardiaque **15 / 30 / 60 s** (tap, sans hardware)
- Messagerie client ↔ véto (+ mode indisponible)
- Timeline historique (messages, relevés validés)
- Rappels Care, pack Horse, foyer / encodage élevage — **inclus** dès entitlement animal actif
- **Rapports vocaux de consultation** et **amélioration IA** des CR (Pro Light + Pro ; édition structurée Web Pro) — historiques transcription / IA / version enregistrée
- Langues **FR / NL / EN / ES / ET**
- Push FCM : message véto → client, confirmation RDV (détail [08](08-MESSAGERIE-NOTIFICATIONS.md))

Ne pas promettre : hardware, WebSocket temps réel (refresh via ouverture app / push).

## Offre à pitcher

Prix **TTC** client. Steer = **triennial**. Pas d’addons à vendre ; pas de plan 5 ans.

| Plan | Prix TTC | ≈ / mois | Message |
|------|----------|----------|---------|
| Monthly | 3,50 € / mois | 3,5 € | Flex (abo Stripe only) |
| Annual | 35 € / an | 2,9 € | Entrée |
| **Triennial** | **95 € / 3 ans** | **2,6 €** | **Recommandé** |

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
| **Véto (30 s)** | « petsFollow : suivi cardiaque prescrit, sans boîtier. Pro à 69 € HT/mois (facture hors ligne), autofinançable via commissions. Pro Light gratuit pour le terrain (ProLight). Vos clients paient ≤ 3,5 €/mois — steer triennial 95 € / 3 ans. » |
| **Client (via véto)** | « Suivi prescrit par votre véto, sans boîtier — à partir de ~2,6 €/mois sur 3 ans. » |
| **Vous (interne)** | « Ouvrez le cabinet, activez 5 pets en 60 j → bonus 25 € + commission sur chaque nouvelle activation. Steer triennial. » |

### Objections

| Objection | Réponse |
|-----------|---------|
| « Encore un abonnement » | Sans boîtier, prescrit par le véto, **≤ 3,5 €/mois** (triennial ~2,6 €). |
| « Je perds s’il y a un commercial » | **Non** — même plafond véto avec ou sans commercial. |
| « Il faut un appareil ? » | **Non** — relevé au doigt dans l’app (durée définie par le cabinet). |
| « Et Family / Care+ ? » | **Inclus** dès qu’un animal est payant — plus d’addons à acheter. |

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
| **Compter une commission (ou un « deal gagné ») dès l’inscription du véto** | L’ouverture du cabinet = étape 1. Vous êtes payé quand un **animal passe payant**. Tant qu’il n’y a pas d’activation → **0 €**. Le bonus ramp (25 €) exige **5 pets payants / 60 j**. |
| **Promettre un % calculé sur le prix TTC** | Le client paie en TTC (ex. 95 €). Votre commission = **% du HTVA** uniquement (hors TVA 21 %). Dire « 12 % de 95 € » est faux. |
| **Dire au véto qu’il gagne moins parce qu’un commercial l’a apporté** | Les grilles sont **indépendantes**. Même plafond (~9,4 € sur le triennial). |
| **Promettre un boîtier, du chat WebSocket, ou vendre des addons** | Vendable = app + Pro (relevé, messagerie, timeline, Care/Horse/foyer inclus, push FCM). Hardware, WebSocket, addons payants = **pas à pitcher**. |

## Liens

| Doc | Contenu |
|-----|---------|
| [14](14-POSITIONNEMENT-MARKETING.md) | Positionnement |
| [17](17-POLITIQUE-TARIFAIRE.md) | Grille prix + économie |
| [18](18-FICHE-COMMISSION-VETO.md) | Commission véto (co-selling) |
| [19](19-FICHE-COMMISSION-COMMERCIAL.md) | Votre rémunération |
| [21](21-GTM-COMMERCIAL.md) | Playbook 30 j + SPIFF |
