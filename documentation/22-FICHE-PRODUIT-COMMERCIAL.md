# Fiche produit — petsFollow (commerciaux)

**Scope** : présenter le produit, l’offre et le pitch. Rémunération → [19](19-FICHE-COMMISSION-COMMERCIAL.md). Playbook → [21](21-GTM-COMMERCIAL.md).

## En une phrase

**Continuité de soins prescrite**, matérialisée en **passeport digital de l’animal** — **Web** cabinet (**Pro**), **mobile** terrain (**ProLight** / care pro), **app** propriétaire (**pets**).  
**Pro** (app Web + apps mobiles) facturé **hors ligne** ; **Pro Light** (app mobile **ProLight**) gratuit ; le client paie le suivi animal (~2–3,5 €/mois).

Le relevé cardiaque (tap, 15 / 30 / 60 s dans l’app) est une **feature** parmi d’autres (messagerie, Care/Horse, CR IA, agenda terrain, partage ACL). « Carnet de santé » = alias UI propriétaire ; identité = **passeport digital**.
## Pour qui / modèle

| Qui | Rôle | Surface | Tarif |
|-----|------|---------|-------|
| Cabinet véto | Prescripteur (B2B) | **Pro** (Web SaaS) | **834,71 € HTVA / an** (ou **2 253,72 € / 3 ans**, −10 %) + setup 320 € HTVA — **facturation externe** |
| Cabinet véto | Add-on CR IA | Dictée / improve Gemini | **39 € HT/mois** ou **390 € HT/an** — essai **90 j**, ROI dès J60 — [32](32-MODULE-IA-CR.md) |
| Pro terrain | App mobile ProLight | **Pro Light** (Flutter, avec ou sans compte Web Pro) | **Gratuit** (CR IA si cabinet activé) |
| Propriétaire | Payeur (B2B2C) | App mobile pets | 3,50 / 35 / 95 € TTC (Stripe) |
| Commercial | Apporteur | Pro web (espace commercial) | — |

Monétisation : **SaaS cabinet hors ligne** + **activations clients payantes** (commissions partenaires pouvant compenser le SaaS).

## Ce que ça fait (vendable aujourd’hui)

- **Passeport digital** de l’animal partagé entre cabinet, care pro et propriétaire
- Création / suivi des animaux
- Messagerie client ↔ véto (+ mode indisponible) et client ↔ care_pro (app Pro Light, fil distinct du cabinet)
- Timeline historique (messages, relevés validés, événements)
- Partage de dossier (`pet_access` / `client_access`) : collègue / care pro / pro externe — notes, CR, docs
- Relevé cardiaque **15 / 30 / 60 s** (tap dans l’app) — feature différenciante
- Rappels Care, pack Horse, foyer / encodage élevage — **inclus** dès entitlement animal actif
- **Rapports vocaux de consultation** et **amélioration IA** des CR (Pro Light + Pro ; édition structurée Web Pro) — **module add-on** 39 € HT/mois ou 390 € HT/an après essai 90 j ([32](32-MODULE-IA-CR.md))
- Agenda terrain (Pro Light) + calendrier cabinet (Pro)
- Langues **FR / NL / EN / ES / ET / IT**
- Push FCM : message véto → client, confirmation RDV (détail [08](08-MESSAGERIE-NOTIFICATIONS.md))

Ne pas promettre : appareil à vendre, WebSocket temps réel (refresh via ouverture app / push), pitch centré « sans boîtier », **messagerie entre care pro** (parler partage dossier / CR / notes).
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
3. **Activer** pets payants — commission à chaque activation ; SPIFF mix si ≥ 55 % triennial / mois

Détail 30 jours + SPIFF → [21](21-GTM-COMMERCIAL.md).

## Scripts

| Audience | Script |
|----------|--------|
| **Véto (30 s)** | « petsFollow : continuité de soins prescrite — un passeport digital de l’animal partagé entre votre cabinet, les pros de soins terrain et le propriétaire. Web Pro, Pro Light gratuit, app client. Pro 834,71 € HTVA / an (hors ligne ; −10 % en triennal), autofinançable via commissions. Clients ≤ 3,5 €/mois — steer 95 € / 3 ans. » |
| **Client (via véto)** | « Le passeport digital de votre animal — messages avec le cabinet, rappels, relevés — dans une app mobile, à partir de ~2,6 €/mois sur 3 ans. » |
| **Vous (interne)** | « Ouvrez le cabinet, activez des pets payants → commission sur chaque nouvelle activation. Steer triennial. Pitch = passeport multi-acteurs · Web + 2 mobiles. » |

### Objections

| Objection | Réponse |
|-----------|---------|
| « Encore un abonnement » | Prescrit par le véto, **≤ 3,5 €/mois** (triennial ~2,6 €), tout dans l’app. |
| « Je perds s’il y a un commercial » | **Non** — même plafond véto avec ou sans commercial. |
| « C’est juste une app cardio ? » | **Non** — passeport digital multi-acteurs : messagerie, Care/Horse, foyer, CR terrain + IA, partage care pro ; le FC est une feature parmi d’autres. |
| « Et Family / Care+ ? » | **Inclus** dès qu’un animal est payant — plus d’addons à acheter. |
| « On a déjà un PMS » | Complementary — petsFollow ajoute le passeport Web + mobile (cabinet / care pro / foyer), pas un 2ᵉ PMS. |
| « Les care pro peuvent chatter entre eux ? » | **Non** — partage sécurisé de dossier, notes et CR ; pas de messagerie inter-pros. |

## Démo terrain

| Rôle | Email | Mot de passe |
|------|-------|--------------|
| Véto | `vet.demo@petsfollow.test` | `VetDemo123!` |
| Client | `client.demo@petsfollow.test` | `ClientDemo123!` |
| Commercial | `commercial.demo@petsfollow.test` | `CommercialDemo123!` |

Local : Pro **http://localhost:3002** · API **http://localhost:8291**  
(`make up-infra && make migrate && make seed && make api-dev` + `make nuxtjs-dev`)

## Plaquette cabinet (leave-behind véto / ASV)

Support produit imprimable en PDF depuis l’espace commercial (ou responsable) : **`/commercial/brochure`**.

| Quand | Usage |
|-------|--------|
| Fin de visite cabinet | Imprimer / Enregistrer en PDF → laisser au véto (et copie standard) |
| Relance e-mail | Bouton « Envoyer par e-mail » — **joindre le PDF** généré côté navigateur (le lien auth n’est pas ouvert aux vétos) |

Focus : « pas un logiciel de plus » · continuité prescrite · passeport partagé 3 acteurs · features GA · offre compacte · bandeau ASV (mémo dédié à part). Nav Offre + CTA depuis `/commercial/pitch`.

## Mémo ASV (leave-behind 10 min)

Fiche adressée à l’**ASV / accueil**, imprimable en PDF depuis l’espace commercial (ou responsable) : **`/commercial/asv-memo`**.

| Quand | Usage |
|-------|--------|
| ASV présente en démo | Laisser lire / annoter pendant que le véto voit le cockpit clinique |
| Suite de visite cabinet | Imprimer / Enregistrer en PDF → laisser au standard |
| Relance e-mail | Bouton « Envoyer par e-mail » (joindre le PDF généré côté navigateur) |

Focus : soulagement du standard (messagerie, RDV, préconsult, Care, switch poste) — **pas** le tarif SaaS cabinet au centre. Nav Offre (commercial + manager) + CTA depuis `/commercial/pitch`.

## Interdits (à bien comprendre)

Un cabinet **commence forcément à 0 animal payant** — c’est normal. Ce qui est interdit, c’est de **confondre inscription et revenu**.

| Interdit | Pourquoi |
|----------|----------|
| **Compter une commission (ou un « deal gagné ») dès l’inscription du véto** | L’ouverture du cabinet = étape 1. Vous êtes payé quand un **animal passe payant**. Tant qu’il n’y a pas d’activation → **0 €**. SPIFF mix = ≥ 55 % activations triennial / mois. |
| **Promettre un % calculé sur le prix TTC** | Le client paie en TTC (ex. 95 €). Votre commission = **% du HTVA** uniquement (hors TVA 21 %). Dire « 12 % de 95 € » est faux. |
| **Dire au véto qu’il gagne moins parce qu’un commercial l’a apporté** | Les grilles sont **indépendantes**. Même plafond (~9,4 € sur le triennial). |
| **Cantonner le pitch au « suivi cardiaque »** | Identité = continuité prescrite + passeport digital (Web + mobiles). Le FC est vendable en démo, pas comme plafond d’offre. |
| **Centrer le pitch sur « sans boîtier »** | Parler des **apps** : Web Pro · ProLight · Client. Ne pas vendre / promettre un appareil. |
| **Promettre une messagerie entre care pro** | Vendable = partage de dossier / CR / notes (ACL). Chat inter-pros = **hors scope**. |
| **Promettre un appareil, du chat WebSocket, ou vendre des addons** | Vendable = Web Pro + apps mobiles (passeport, messagerie véto↔client, timeline, FC, Care/Horse/foyer inclus, CR IA, push FCM, partage ACL). WebSocket, addons payants = **pas à pitcher**. |

## Liens

| Doc | Contenu |
|-----|---------|
| [14](14-POSITIONNEMENT-MARKETING.md) | Positionnement |
| [17](17-POLITIQUE-TARIFAIRE.md) | Grille prix + économie |
| [18](18-FICHE-COMMISSION-VETO.md) | Commission véto (co-selling) |
| [19](19-FICHE-COMMISSION-COMMERCIAL.md) | Votre rémunération |
| [21](21-GTM-COMMERCIAL.md) | Playbook 30 j + SPIFF |
