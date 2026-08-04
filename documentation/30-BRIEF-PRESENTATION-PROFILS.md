# Brief présentation & débrief commercial — petsFollow

> **Usages**  
> 1. **Présentation in-app** — Pro (admin / commercial / manager) : [`/presentation`](../nuxtjs/pages/presentation/index.vue) (pages successives, focus cabinet + IA) · diagrammes Mermaid : [`/flux-ia`](../nuxtjs/pages/flux-ia/index.vue)  
> 2. **Présentation Gemini** — coller dans Gemini : *« Crée une présentation professionnelle à partir de ce brief… »*  
> 3. **Débrief commercial** — check-list post-RDV / coaching (section 12)  
> Sources : [22](22-FICHE-PRODUIT-COMMERCIAL.md), [14](14-POSITIONNEMENT-MARKETING.md), [21](21-GTM-COMMERCIAL.md), [28](28-MULTI-PROFILS-PRO.md), [17](17-POLITIQUE-TARIFAIRE.md), [32](32-MODULE-IA-CR.md).

---

## Présentation in-app (shell Pro)

Parcours pages successives (`?step=`) pour **cabinet vétérinaire** :

1. Accueil / positionnement → constat → 3 surfaces → VetPro → continuité soins  
2. **Focus IA** : intégration CR IA (dictée → transcription → Améliorer → finalize)  
3. **Focus IA** : automatisations adoption (drip J0–J60, ROI, friction)  
4. Écosystème → offre → closing (CTA fonctionnel)

Diagrammes Mermaid associés : `/flux-ia?profile=vet` (consultation CR IA, adoption, continuité) — profils secondaires commercial / care_pro.

Ancienne route `/commercial/pitch-deck` → redirect vers `/presentation`.

---

## Consigne pour Gemini (pitch deck)

Tu es un designer de pitch deck B2B santé animale. À partir du brief ci-dessous :

1. Produis **12 à 18 slides** (titre + 3–6 bullets max, ou schéma simple).
2. Organisation obligatoire : **Vue d’ensemble → VetPro (Web) → Focus IA (CR inclus + automatisations) → VetLight (mobile) → Client (mobile) → Écosystème → Offre & modèle → Différenciation → Closing**.
3. Langue : **français**. Style : clair, confiant, concrêt (bénéfices avant features).
4. **Identité produit = continuité de soins prescrite + passeport digital de l’animal** via **trois apps** : Web cabinet · mobile ProLight / care pro · mobile particulier. Le relevé respiratoire est **une feature parmi d’autres** — ne jamais présenter petsFollow comme « une app cardiaque ».
5. **Ne pas centrer le pitch sur « sans boîtier »** — parler des surfaces logicielles (Web + mobile). Ne promets **jamais** : chat WebSocket temps réel, messagerie entre care pro, addons payants Family/Care+/Horse, ni un appareil à vendre. Pitch care pro = **partage de dossier / CR / notes**.
6. Steer commercial client : plan **triennial 95 € / 3 ans**.
7. Inclure **2 slides IA** : (a) comment le CR IA est intégré dans VetPro ; (b) automatisations d’adoption / ROI J60.
8. Propose en fin de deck une **slide « Démo terrain »** (parcours 5 minutes).
---

## 1. En une phrase

**petsFollow** = **continuité de soins prescrite**, matérialisée en **passeport digital** de l’animal — **Web** cabinet (**Pro** / VetPro), **mobile** terrain (**Pro Light** / VetLight / care pro), **app** propriétaire (**Client**).

### Glossaire

| Nom produit | Alias deck | Surface |
|-------------|------------|---------|
| **Pro** | VetPro | App Web cabinet |
| **Pro Light** | VetLight | App mobile terrain |
| **petsFollow** (app client) | Client | App mobile particulier |

Sous-ligne : Pro pilote · Pro Light documente · le propriétaire suit et paie — **un même passeport** partagé entre vétérinaires, pros de soins et foyer (messagerie véto↔client, Care/Horse, foyer, relevés, partage ACL).

Trois faces complémentaires :

| Solution | Qui | Surface | Tarif |
|----------|-----|---------|-------|
| **VetPro** | Cabinet / vétérinaire | **App Web** SaaS (Nuxt Pro) | **834,71 € HTVA/an** (ou **2 253,72 € / 3 ans**) + setup **320 € HTVA** (facture hors ligne) |
| **VetLight** | Véto terrain (et pros santé associés) | **App mobile** Flutter **Pro Light** | **Gratuit** |
| **Client** | Propriétaire d’animal | **App mobile** Flutter **pets** | **3,50 € / mois** · **35 € / an** · **95 € / 3 ans** (TTC, Stripe) |

Modèle : **B2B2C + SaaS cabinet** — le véto prescrit, le client paie le suivi animal ; le SaaS Pro est facturé au cabinet en externe.

---

## 2. Carte de l’écosystème

```text
                    ┌─────────────────────┐
                    │  VetPro (Web)       │
                    │  Cabinet · relation │
                    └──────────┬──────────┘
                               │ prescrit / suit
              ┌────────────────┼────────────────┐
              ▼                                 ▼
   ┌──────────────────┐              ┌──────────────────┐
   │ VetLight (mobile)│              │ Client (mobile)  │
   │ Terrain · CR IA  │◄── partage ──│ Particulier      │
   │ Agenda · fiches  │              │ Messages · Care  │
   └──────────────────┘              └──────────────────┘
```

**Flux de valeur** : Commercial ouvre le cabinet → VetPro (Web) onboard → VetLight (mobile terrain) → Client (mobile) active un animal payant → messagerie + Care/Horse/foyer + relevés respiratoires **inclus**.

---

## 3. Solution VetPro (app Web cabinet)

### Profil cible

- **Dr Martin** — vétérinaire libéral / cabinet
- Besoin : garder le fil avec les patients **entre** les consultations, avec un outil Web cabinet sérieux

### Objectif produit

Donner au cabinet un **cockpit Web Pro** pour :

- Gérer clients & animaux
- Échanger en **messagerie** avec le propriétaire
- Voir l’**historique** (timeline : messages, événements, relevés)
- Recevoir les **relevés respiratoires** validés (feature)
- Piloter agenda / RDV, équipe, paramètres cabinet
- Éditer les **comptes rendus** (y compris versions IA)
- Suivre les **commissions** liées aux activations clients

### Bénéfices (à pitcher)

- Continuité de soins entre deux visites
- Prescription claire (« votre animal est suivi dans petsFollow »)
- Une seule plateforme logicielle : Web cabinet + apps mobiles
- SaaS autofinançable via **commissions** sur activations (même plafond avec ou sans commercial)
- Équipe cabinet : collègues, partage de dossiers, calendrier

### Fonctionnalités clés

| Domaine | Ce que fait VetPro |
|---------|--------------------|
| Cabinet | Onboarding profil, préférences email, mode messagerie indisponible, durées FR 15/30/60 s |
| Clients | Liste, invitations / link-requests, dossiers animaux, photos, envoi lien app |
| Messagerie | Threads client ↔ véto (+ médias), push côté client |
| Timeline | Messages, événements, relevés validés |
| Cardiaque | Relevés **validés** par le client → visibles Pro (**feature**) |
| Agenda | Calendrier RDV, plages, vacances, booking client optionnel |
| CR / IA | Édition structurée des rapports de visite ; historique transcription / IA / version finale |
| Équipe | Page équipe, partage animal / client (ACL) |
| Business | Commissions véto, overview dashboard, Care overdue |
| i18n | FR / NL / EN / ES / ET / IT — 6 locales (Web Pro ; UK / RU réservés aux apps mobiles) |

### Parcours type (VetPro)

1. Inscription + confirmation email  
2. Onboarding profil cabinet complet  
3. Création / rattachement clients & animaux  
4. Prescription du suivi (lien app mobile / activation)  
5. Messagerie + timeline au quotidien (relevés FR quand prescrits)  
6. Agenda & CR (Web + terrain via VetLight mobile)

### Tarif VetPro

- **834,71 € HTVA / an** (ou **2 253,72 € HTVA / 3 ans**, −10 %)
- Setup **320 € HT**
- Facturation **externe** (pas Stripe cabinet)
- Objectif commercial terrain : **activer des pets payants** (commission à chaque activation ; SPIFF mix triennial)

---

## 4. Solution VetLight (app mobile terrain)

### Profil cible

- Même vétérinaire **en déplacement** (visite domicile, écurie, terrain)
- Aussi : autres **care pro** santé animale via le même shell Pro Light (maréchal-ferrant, physio, comportementaliste, toiletteur, éleveur) — specialty `vet_light` = positionnement « véto light »

### Objectif produit

Une **app mobile Pro Light gratuite** pour travailler **sur le terrain** :

- Voir l’**agenda** du jour / 7 jours / tout
- Accéder aux **fiches clients & animaux** partagés
- Saisir un **compte rendu** (dictée → transcription → finalisation ; Améliorer IA sur Web Pro)
- Échanger en **messagerie** avec le propriétaire (fil distinct du cabinet)
- Noter GPS / adresse de visite, ouvrir Maps
- Marquer une visite **faite**

### Relation VetPro ↔ VetLight

| | VetPro | VetLight |
|---|--------|----------|
| Surface | **App Web** cabinet | **App mobile** terrain |
| Prix | SaaS payant | Gratuit |
| Force | Pilotage, messagerie, config, équipe | Agenda, CR vocaux, mobilité, messagerie client |
| Compte | Rôle `vet` | Même compte `vet` **ou** `care_pro` + specialty |
| Messagerie | Oui (cœur Web) | Oui sur Pro Light (staff **et** care_pro ↔ client ; fil distinct du cabinet). Pas de chat **entre** care pro |

Un véto peut avoir **VetPro + VetLight** (Web + mobile). Un care pro peut n’avoir que **VetLight**.

### Bénéfices (à pitcher)

- Zéro friction : app mobile gratuite, utile dès le premier RDV terrain
- CR vocaux + IA : moins de saisie au cabinet le soir
- Continuité dossier : ce qui est prescrit / partagé côté Web Pro est visible sur mobile
- GPS visite & tournées : agenda opérationnel

### Fonctionnalités clés

| Domaine | Ce que fait VetLight |
|---------|----------------------|
| Shell | Onglets Agenda · Clients · Animaux · Messages · Settings |
| Agenda | Filtres Aujourd’hui / 7 j / Tout ; bouton Fait ; GPS / Maps |
| Dossiers | Clients & pets selon droits (read / write_notes / full) |
| Messagerie | Threads care_pro ↔ client (ACL) ; staff cabinet = fil practice |
| CR | Micro → transcription Gemini → éditer → enregistrer / finaliser (Améliorer IA = Web Pro) |
| Sécurité PHI | Audio CR non public ; stream auth ; purge à la finalisation |
| Langues | FR / NL / EN / ES / ET / IT / UK / RU |

### Parcours type (VetLight)

1. Login (compte véto ou care_pro créé / rattaché)  
2. Agenda du jour → ouvrir visite  
3. Consulter fiche animal  
4. Dictée CR → enregistrer / finaliser (Améliorer IA sur Web Pro)  
5. Messages → composer vers le client partagé  
6. Marquer la visite faite  

---

## 5. Solution Client (app mobile particulier)

### Profil cible

- **Sophie** — propriétaire (chien senior, chat suivi, cheval, multi-animaux foyer…)
- Entrée : prescription du véto **ou** self-inscription puis liaison cabinet

### Objectif produit

Une **app mobile** simple et rassurante pour le **suivi prescrit** :

- Créer / gérer ses animaux (foyer)
- Messager avec son cabinet (+ push)
- Rappels Care / Horse — **inclus** dès abonnement actif
- Timeline historique
- Relevé respiratoire au doigt (15/30/60 s) — **feature** dans l’app, durée prescrite par le cabinet

### Bénéfices (à pitcher)

- Prescrit par le véto → confiance médicale, pas un gadget wellness isolé
- Tout dans l’app mobile (lien direct avec le Web Pro du cabinet)
- Prix accessible : **≤ 3,5 €/mois** ; recommandé **~2,6 €/mois** sur 3 ans
- Une seule app pour le foyer (plusieurs animaux)

### Fonctionnalités clés

| Domaine | Ce que fait Client |
|---------|-------------------|
| Compte | Email/MDP, Google, confirm email, locale, export RGPD, suppression compte |
| Animaux | Création, photo, plans Stripe, foyer / litters |
| Messagerie | Thread avec le cabinet ; mode indisponible respecté |
| Care / Horse | Rappels, contacts pro, compétitions — **inclus** (plus d’addons) |
| Timeline | Messages + événements + relevés validés |
| Cardiaque | Sessions 15/30/60 s → BPM + commentaire → envoi véto (**feature**) |
| Engagement | Missions discovery in-app + emails éducatifs (opt-out possible) |
| Langues | FR / NL / EN / ES / ET / IT / UK / RU |

### Parcours type (Client)

1. Register / Login / Google (+ consentement CGU)  
2. Créer animal + choisir plan (steer triennial)  
3. Paiement Stripe → entitlement actif  
4. Premier usage : message / Care / (si prescrit) relevé respiratoire  
5. Relation continue avec le cabinet  

### Offre client (TTC)

| Plan | Prix | Message slide |
|------|------|---------------|
| Mensuel | **3,50 € / mois** | Flexibilité (abonnement Stripe) |
| Annuel | **35 € / an** (~2,9 €/mois) | Entrée |
| **Triennal** | **95 € / 3 ans** (~2,6 €/mois) | **Recommandé** (−10 % vs 3× annuel) |

**Inclus** dès animal payant : messagerie + push, timeline, Care/Horse, foyer, relevés respiratoires — **pas d’addons à vendre**.

---

## 6. Qui fait quoi — matrice synthétique

| Besoin | VetPro (Web) | VetLight (mobile) | Client (mobile) |
|--------|:------------:|:-----------------:|:---------------:|
| Configurer le cabinet / durées FR | ✓ | | |
| Messagerie cabinet ↔ propriétaire | ✓ | | ✓ |
| Care / Horse / foyer | ✓ (vue) | | ✓ |
| Agenda terrain + GPS | ✓ (web) | ✓ | (booking optionnel) |
| CR vocal + IA | ✓ (édition) | ✓ (dictée) | |
| Suivre les relevés validés | ✓ | | (envoie) |
| Relevé respiratoire au doigt | | | ✓ |
| Payer le suivi animal | | | ✓ |
| Commissions / business cabinet | ✓ | | |
| Prix | Payant SaaS | Gratuit | Abo animal |

---

## 7. Messages clés par audience

### Au vétérinaire (VetPro + VetLight)

> « petsFollow : continuité de soins prescrite — un passeport digital de l’animal partagé entre votre cabinet, les pros de soins terrain et le propriétaire. Web Pro, Pro Light gratuit, app client. VetPro à 834,71 € HTVA/an (facture hors ligne ; −10 % en triennal), autofinançable via commissions. Clients ≤ 3,5 €/mois — steer 95 € / 3 ans. »

### Au propriétaire (via le véto)

> « Le passeport digital de votre animal — messages avec le cabinet, rappels, relevés — dans une app mobile, à partir de ~2,6 €/mois sur 3 ans. »

### Au commercial (interne)

> « Ouvrez le cabinet, activez des pets payants → commission sur chaque nouvelle activation. Steer triennial. Pitch = passeport multi-acteurs · Web + 2 mobiles (ProLight + particulier), pas “app cardio”, pas “anti-boîtier”, pas messagerie **entre** care pro (oui care_pro ↔ client sur Pro Light). »

### Différenciation (1 slide)

- Trois surfaces logicielles : **Web Pro · mobile ProLight · mobile Client**  
- Prescription vétérinaire (pas un gadget grand public)  
- **Passeport multi-acteurs** : vet / care pro / foyer sur le même animal  
- Continuité multi-profil : VetPro · VetLight · Client · Care pro — 8 langues sur les apps mobiles (6 sur le Web Pro)  
- Complémentaire du PMS (ne le remplace pas)  
- VetLight gratuit pour le terrain + CR (IA sur Web) + partage ACL + messagerie client  
- Alignement économique véto / commercial (pas de pénalité co-selling)  

### Objections fréquentes

| Objection | Réponse courte |
|-----------|----------------|
| « Encore un abonnement » | Prescrit par le véto, **≤ 3,5 €/mois** (triennial ~2,6 €), tout dans l’app |
| « C’est juste une app cardio ? » | **Non** — passeport digital : messagerie, Care/Horse, foyer, CR + partage care pro ; la FR est une feature |
| « Et Family / Care+ ? » | Inclus dès qu’un animal est payant |
| « Je perds s’il y a un commercial » | Non — même plafond commission véto |
| « On a déjà un logiciel / PMS » | Complementary — passeport Web + mobile (cabinet / care pro / foyer), pas un 2ᵉ PMS |
| « Les care pro peuvent chatter entre eux ? » | **Non** — partage dossier / CR / notes ; pas de messagerie inter-pros |
| « Encore une app à installer ? » | Une pour le propriétaire ; le cabinet a le Web + ProLight terrain si besoin |

---

## 8. Scénario démo 5 minutes (slide dédiée)

1. **VetPro (Web)** — dashboard + dossier animal + messagerie  
2. **Client (mobile)** — message / Care + (option) relevé 30 s → Valider  
3. **VetPro (Web)** — timeline / relevé + réponse au propriétaire  
4. **VetLight (mobile)** — agenda du jour → dictée CR → améliorer IA  
5. **Closing** — prix client triennial + SaaS Web Pro + ProLight gratuit  

Comptes démo locaux (si besoin) : voir `AGENTS.md`  
(`vet.demo@` / `client.demo@` / `vetlight.demo@`).

---

## 9. Structure de slides recommandée

1. Titre — petsFollow  
2. Promesse — continuité prescrite + passeport digital (Web + mobile + mobile)  
3. Le problème (fil perdu entre consultations **et** entre soignants)  
4. La réponse — passeport multi-acteurs · 3 apps  
5. Schéma écosystème  
6. VetPro Web — pour qui / pourquoi  
7. VetPro — fonctionnalités  
8. VetLight mobile — pour qui / pourquoi  
9. VetLight — CR & terrain  
10. Client mobile — pour qui / pourquoi  
11. Client — suivi prescrit (messages, Care, **et** relevé FR)  
12. Matrice « qui fait quoi »  
13. Offre & prix (cabinet + client)  
14. Parcours de valeur (prescribe → active → suit)  
15. Différenciation  
16. Objections  
17. Démo 5 min  
18. Call to action / contact  

---

## 10. Interdits (ne pas mettre dans les slides ni le pitch)

- Vendre ou promettre un **appareil / wearable** (ce n’est pas l’offre)  
- Centrer le pitch sur « sans boîtier » (parler des **apps** Web + mobile)  
- Chat temps réel type WebSocket  
- Vente d’addons Family / Kennel / Care+ / Horse  
- Plan 5 ans (quinquennial) comme offre active  
- Confondre **inscription cabinet** et **revenu** (revenu = animal payant)  
- Calculer une commission sur le **TTC** (base = HTVA)  
- **Cantonner petsFollow à « suivi cardiaque »** — identité = continuité prescrite + passeport digital  
- **Promettre une messagerie entre care pro** — parler partage de dossier / CR / notes  

---

## 11. Identité visuelle (indications Gemini)

- Marque : **petsFollow** (hero-level sur la slide titre)  
- Univers : santé animale, confiance vétérinaire, moderne — **pas** dark mode gadget, **pas** violet générique IA  
- Couleurs : s’appuyer sur la charte Pro (teal / vétérinaire — voir `documentation/13-CHARTE-GRAPHIQUE.md` si images)  
- Visuels : **écrans Web + mobiles**, pas d’illustration de collier / boîtier  
- Une idée visuelle dominante par slide ; peu de cartes ; pas de clusters de pills  

---

## 12. Débrief commercial (post-RDV / coaching)

À remplir après chaque démo ou appel. Aligné [21](21-GTM-COMMERCIAL.md) / [22](22-FICHE-PRODUIT-COMMERCIAL.md).

### 12.1 Ce qui a été dit (auto-contrôle)

| Check | Oui / Non | Note |
|-------|-----------|------|
| Promesse = **continuité + passeport** cabinet / care pro / foyer | | |
| Surfaces = **Web Pro + mobile ProLight + mobile Client** | | |
| FR présenté comme **feature**, pas comme produit | | |
| Pas de pitch centré « sans boîtier » | | |
| Pas de promesse **messagerie entre care pro** (client ↔ care_pro OK) | | |
| Steer **triennial 95 €** | | |
| Care/Horse/foyer = **inclus** | | |
| Pas de promesse WebSocket / addons / appareil à vendre | | |
| Pas de % calculé sur TTC | | |
| Inscription ≠ revenu rappelé si besoin | | |

### 12.2 Résultat RDV

| Champ | Valeur |
|-------|--------|
| Cabinet / contact | |
| Date / format (visio / sur place / tel) | |
| Intérêt (chaud / tiède / froid) | |
| Objection principale | |
| Next step (démo 2, essai, non) | |
| Pets payants estimés J+60 | |
| CRM mis à jour (`/commercial/prospects`) | Oui / Non |

### 12.3 Points forts à rejouer

- _
- _

### 12.4 Points à corriger au prochain pitch

- _
- _

### 12.5 Angle gagnant observé (cocher)

- [ ] Continuité entre consultations  
- [ ] Trio Web + ProLight + app particulier  
- [ ] VetLight terrain + CR IA  
- [ ] Autofinancement via commissions  
- [ ] Prix client ≤ 3,5 € / steer triennial  
- [ ] Complément PMS (pas de remplacement)  
- [ ] Autre : ___

### 12.6 Phrase de closing utilisée

> _

---

*Brief présentation + débrief commercial — Web Pro · mobile ProLight · mobile Client (juillet 2026).*
