# Brief présentation & débrief commercial — petsFollow

> **Usages**  
> 1. **Présentation** — coller dans Gemini : *« Crée une présentation professionnelle à partir de ce brief… »*  
> 2. **Débrief commercial** — check-list post-RDV / coaching (section 12)  
> Sources : [22](22-FICHE-PRODUIT-COMMERCIAL.md), [14](14-POSITIONNEMENT-MARKETING.md), [21](21-GTM-COMMERCIAL.md), [28](28-MULTI-PROFILS-PRO.md), [17](17-POLITIQUE-TARIFAIRE.md).

---

## Consigne pour Gemini (pitch deck)

Tu es un designer de pitch deck B2B santé animale. À partir du brief ci-dessous :

1. Produis **12 à 18 slides** (titre + 3–6 bullets max, ou schéma simple).
2. Organisation obligatoire : **Vue d’ensemble → VetPro (Web) → VetLight (mobile) → Client (mobile) → Écosystème → Offre & modèle → Différenciation → Closing**.
3. Langue : **français**. Style : clair, confiant, concrêt (bénéfices avant features).
4. **Identité produit = continuité de soins prescrite** via **trois apps** : Web cabinet · mobile ProLight · mobile particulier. Le relevé cardiaque est **une feature parmi d’autres** — ne jamais présenter petsFollow comme « une app cardiaque ».
5. **Ne pas centrer le pitch sur « sans boîtier »** — parler des surfaces logicielles (Web + mobile). Ne promets **jamais** : chat WebSocket temps réel, addons payants Family/Care+/Horse, ni un appareil à vendre.
6. Steer commercial client : plan **triennial 95 € / 3 ans**.
7. Propose en fin de deck une **slide « Démo terrain »** (parcours 5 minutes).

---

## 1. En une phrase

**petsFollow** = **continuité de soins prescrite** — **Web** pour le cabinet (**Pro** / VetPro), **mobile** pour le terrain (**Pro Light** / VetLight) et le particulier (**app client**).

### Glossaire

| Nom produit | Alias deck | Surface |
|-------------|------------|---------|
| **Pro** | VetPro | App Web cabinet |
| **Pro Light** | VetLight | App mobile terrain |
| **petsFollow** (app client) | Client | App mobile particulier |

Sous-ligne : Pro pilote · Pro Light documente · le propriétaire suit et paie — **messagerie, Care/Horse, foyer, relevés cardiaques** inclus.

Trois faces complémentaires :

| Solution | Qui | Surface | Tarif |
|----------|-----|---------|-------|
| **VetPro** | Cabinet / vétérinaire | **App Web** SaaS (Nuxt Pro) | **69 € HT/mois** + setup **320 € HT** (facture hors ligne) |
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

**Flux de valeur** : Commercial ouvre le cabinet → VetPro (Web) onboard → VetLight (mobile terrain) → Client (mobile) active un animal payant → messagerie + Care/Horse/foyer + relevés cardiaques **inclus**.

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
- Recevoir les **relevés cardiaques** validés (feature)
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
| Cabinet | Onboarding profil, préférences email, mode messagerie indisponible, durées FC 15/30/60 s |
| Clients | Liste, invitations / link-requests, dossiers animaux, photos, envoi lien app |
| Messagerie | Threads client ↔ véto (+ médias), push côté client |
| Timeline | Messages, événements, relevés validés |
| Cardiaque | Relevés **validés** par le client → visibles Pro (**feature**) |
| Agenda | Calendrier RDV, plages, vacances, booking client optionnel |
| CR / IA | Édition structurée des rapports de visite ; historique transcription / IA / version finale |
| Équipe | Page équipe, partage animal / client (ACL) |
| Business | Commissions véto, overview dashboard, Care overdue |
| i18n | FR / NL / EN / ES / ET |

### Parcours type (VetPro)

1. Inscription + confirmation email  
2. Onboarding profil cabinet complet  
3. Création / rattachement clients & animaux  
4. Prescription du suivi (lien app mobile / activation)  
5. Messagerie + timeline au quotidien (relevés FC quand prescrits)  
6. Agenda & CR (Web + terrain via VetLight mobile)

### Tarif VetPro

- **69 € HT / mois**
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
- Saisir un **compte rendu** (dictée vocale → transcription → amélioration IA → finalisation)
- Noter GPS / adresse de visite, ouvrir Maps
- Marquer une visite **faite**

### Relation VetPro ↔ VetLight

| | VetPro | VetLight |
|---|--------|----------|
| Surface | **App Web** cabinet | **App mobile** terrain |
| Prix | SaaS payant | Gratuit |
| Force | Pilotage, messagerie, config, équipe | Agenda, CR vocaux, mobilité |
| Compte | Rôle `vet` | Même compte `vet` **ou** `care_pro` + specialty |
| Messagerie | Oui (cœur) | Hors scope care_pro ; véto utilise surtout le Web pour le chat |

Un véto peut avoir **VetPro + VetLight** (Web + mobile). Un care pro peut n’avoir que **VetLight**.

### Bénéfices (à pitcher)

- Zéro friction : app mobile gratuite, utile dès le premier RDV terrain
- CR vocaux + IA : moins de saisie au cabinet le soir
- Continuité dossier : ce qui est prescrit / partagé côté Web Pro est visible sur mobile
- GPS visite & tournées : agenda opérationnel

### Fonctionnalités clés

| Domaine | Ce que fait VetLight |
|---------|----------------------|
| Shell | Onglets Agenda · Clients · Settings (+ drill-down animal / docs / CR) |
| Agenda | Filtres Aujourd’hui / 7 j / Tout ; bouton Fait ; GPS / Maps |
| Dossiers | Clients & pets selon droits (read / write_notes / full) |
| CR | Micro → upload → transcription Gemini → « améliorer » (SOAP / specialty) → finaliser |
| Sécurité PHI | Audio CR non public ; stream auth ; purge à la finalisation |
| Langues | FR / NL / EN / ES / ET |

### Parcours type (VetLight)

1. Login (compte véto ou care_pro créé / rattaché)  
2. Agenda du jour → ouvrir visite  
3. Consulter fiche animal  
4. Dictée CR → améliorer IA → enregistrer / finaliser  
5. Marquer la visite faite  

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
- Relevé cardiaque au doigt (15/30/60 s) — **feature** dans l’app, durée prescrite par le cabinet

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
| Langues | FR / NL / EN / ES / ET |

### Parcours type (Client)

1. Register / Login / Google (+ consentement CGU)  
2. Créer animal + choisir plan (steer triennial)  
3. Paiement Stripe → entitlement actif  
4. Premier usage : message / Care / (si prescrit) relevé cardiaque  
5. Relation continue avec le cabinet  

### Offre client (TTC)

| Plan | Prix | Message slide |
|------|------|---------------|
| Mensuel | **3,50 € / mois** | Flexibilité (abonnement Stripe) |
| Annuel | **35 € / an** (~2,9 €/mois) | Entrée |
| **Triennal** | **95 € / 3 ans** (~2,6 €/mois) | **Recommandé** (−10 % vs 3× annuel) |

**Inclus** dès animal payant : messagerie + push, timeline, Care/Horse, foyer, relevés cardiaques — **pas d’addons à vendre**.

---

## 6. Qui fait quoi — matrice synthétique

| Besoin | VetPro (Web) | VetLight (mobile) | Client (mobile) |
|--------|:------------:|:-----------------:|:---------------:|
| Configurer le cabinet / durées FC | ✓ | | |
| Messagerie cabinet ↔ propriétaire | ✓ | | ✓ |
| Care / Horse / foyer | ✓ (vue) | | ✓ |
| Agenda terrain + GPS | ✓ (web) | ✓ | (booking optionnel) |
| CR vocal + IA | ✓ (édition) | ✓ (dictée) | |
| Suivre les relevés validés | ✓ | | (envoie) |
| Relevé cardiaque au doigt | | | ✓ |
| Payer le suivi animal | | | ✓ |
| Commissions / business cabinet | ✓ | | |
| Prix | Payant SaaS | Gratuit | Abo animal |

---

## 7. Messages clés par audience

### Au vétérinaire (VetPro + VetLight)

> « petsFollow : continuité de soins prescrite — Web pour votre cabinet, mobile ProLight pour le terrain, app pour vos clients. VetPro à 69 € HT/mois (facture hors ligne), autofinançable via commissions. ProLight gratuit. Vos clients paient ≤ 3,5 €/mois — steer 95 € / 3 ans. Messagerie, Care/Horse, relevés cardiaques inclus. »

### Au propriétaire (via le véto)

> « Le suivi que votre vétérinaire vous prescrit — messages, rappels, relevés — dans une app mobile, à partir de ~2,6 €/mois sur 3 ans. »

### Au commercial (interne)

> « Ouvrez le cabinet, activez des pets payants → commission sur chaque nouvelle activation. Steer triennial. Pitch = Web + 2 mobiles (ProLight + particulier), pas “app cardio”, pas “anti-boîtier”. »

### Différenciation (1 slide)

- Trois surfaces logicielles : **Web Pro · mobile ProLight · mobile Client**  
- Prescription vétérinaire (pas un gadget grand public)  
- Continuité multi-profil : VetPro · VetLight · Client · Care pro — 5 langues  
- Complémentaire du PMS (ne le remplace pas)  
- VetLight gratuit pour le terrain + CR IA  
- Alignement économique véto / commercial (pas de pénalité co-selling)  

### Objections fréquentes

| Objection | Réponse courte |
|-----------|----------------|
| « Encore un abonnement » | Prescrit par le véto, **≤ 3,5 €/mois** (triennial ~2,6 €), tout dans l’app |
| « C’est juste une app cardio ? » | **Non** — messagerie, Care/Horse, foyer, CR terrain + IA ; le FC est une feature |
| « Et Family / Care+ ? » | Inclus dès qu’un animal est payant |
| « Je perds s’il y a un commercial » | Non — même plafond commission véto |
| « On a déjà un logiciel / PMS » | Complementary — petsFollow ajoute la continuité Web + mobile propriétaire, pas un 2ᵉ PMS |
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
2. Promesse — continuité prescrite (Web + mobile + mobile)  
3. Le problème (fil perdu entre consultations)  
4. La réponse — 3 apps  
5. Schéma écosystème  
6. VetPro Web — pour qui / pourquoi  
7. VetPro — fonctionnalités  
8. VetLight mobile — pour qui / pourquoi  
9. VetLight — CR & terrain  
10. Client mobile — pour qui / pourquoi  
11. Client — suivi prescrit (messages, Care, **et** relevé FC)  
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
- **Cantonner petsFollow à « suivi cardiaque »** — identité = continuité prescrite  

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
| Promesse = **continuité** cabinet / terrain / foyer | | |
| Surfaces = **Web Pro + mobile ProLight + mobile Client** | | |
| FC présenté comme **feature**, pas comme produit | | |
| Pas de pitch centré « sans boîtier » | | |
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
