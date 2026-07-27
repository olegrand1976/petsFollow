# Positionnement marketing — petsFollow

## Promesse

**petsFollow — la continuité de soins prescrite, matérialisée en passeport digital de l’animal : Web cabinet · mobile terrain (Pro Light / care pro) · app propriétaire.**

### Glossaire (noms figés)

| Nom | Surface | Public |
|-----|---------|--------|
| **Pro** (alias deck : VetPro) | App **Web** | Cabinet / vétérinaire |
| **Pro Light** (alias deck : VetLight) | App **mobile** | Terrain / care pro |
| **petsFollow** app client (alias : Client) | App **mobile** | Particulier / propriétaire |

Sous-ligne : Pro pilote · Pro Light documente · le propriétaire suit et paie — un même passeport partagé entre vétérinaires, pros de soins et foyer.

**Vocabulaire** : **passeport digital** = identité produit (dossier vivant multi-acteurs). « Carnet de santé » = alias UI côté propriétaire, pas le positionnement canon.

**Pro** (app **Web** + apps mobiles) facturé **hors ligne** (69 € HT/mois + setup) ; **Pro Light** (app **mobile** ProLight) gratuit ; le propriétaire paie un abonnement par animal (~2–3,5 €/mois) via l’app **mobile** pets.

Le relevé cardiaque (15 / 30 / 60 s, dans l’app) est une **fonctionnalité différenciante**, pas l’identité produit.

## Axes du passeport

| Lien | Vendable aujourd’hui | Hors pitch |
|------|----------------------|------------|
| **Pro ↔ propriétaire** | Messagerie, timeline, relevés FC, rappels Care/Horse, push | — |
| **Care pro ↔ propriétaire** | Notes, docs, CR visite, agenda, ACL `write_notes` | Messagerie care_pro |
| **Pro ↔ care pro / care pro ↔ care pro** | Partage de dossier (`pet_access` / `client_access`), CR, notes | Chat inter-pros |

Le passeport = **quoi** circule entre acteurs. La continuité prescrite = **pourquoi / comment** (prescription véto, 3 surfaces, modèle éco).

## Modèle

**B2B2C + SaaS cabinet** : commercial apporte le cabinet → véto prescrit → client paie (Stripe) ; abonnement Pro **facturé en externe** (pas de Stripe cabinet).

| Acteur | Bénéfice |
|--------|----------|
| Véto (Pro) | Continuité entre consultations + passeport partagé + commissions pouvant compenser le SaaS ; CR vocaux + amélioration IA (Web) |
| Pro Light (ProLight) | App mobile terrain gratuite (véto light / care pro), notes / CR / docs + partage ACL ; avec ou sans compte Web Pro |
| Client | App mobile simple : face propriétaire du passeport — suivi prescrit, messages (↔ véto), rappels Care/Horse, relevé FC |
| Commercial | Commission sur chaque nouvelle activation + SPIFF mix triennial |

## Offre cœur (TTC)

| Plan | Prix | Message |
|------|------|---------|
| Monthly | 3,50 € / mois | Flex (abo only) |
| Annual | 35 € / an | Entrée |
| **Triennial** | **95 € / 3 ans** (~2,6 €/mois) | **Recommandé** |

Care / Horse / foyer / encodage élevage : **inclus** dès entitlement animal actif (plus d’addons à vendre). Quinquennial hors vente.

## Différenciation

- Trois surfaces logicielles : **Pro (Web) · Pro Light (mobile) · app client (mobile)**
- Prescription véto (pas un gadget grand public isolé)
- **Passeport multi-acteurs** : vet / care pro / foyer sur le même animal
- Continuité multi-profil : Pro · Pro Light · Client · Care pro — 6 langues
- Complémentaire du PMS cabinet (ne le remplace pas)
- Alignement économique véto + commercial (pas de pénalité co-selling)

## Matériel commercial

- Fiche produit : [22-FICHE-PRODUIT-COMMERCIAL.md](22-FICHE-PRODUIT-COMMERCIAL.md)
- Playbook : [21-GTM-COMMERCIAL.md](21-GTM-COMMERCIAL.md)
- Grille & commissions : [17](17-POLITIQUE-TARIFAIRE.md), [18](18-FICHE-COMMISSION-VETO.md), [19](19-FICHE-COMMISSION-COMMERCIAL.md)
- Page Pro pitch : `/commercial/pitch` (`ProCommissionSheet`)

## Interdits pitch

Ne pas promettre un appareil à vendre, WebSocket temps réel, ni % sur TTC — voir interdits dans [22](22-FICHE-PRODUIT-COMMERCIAL.md).  
Ne pas cantonner le pitch à « une app cardiaque » — le FC est un module parmi d’autres.  
Ne pas centrer le pitch sur « sans boîtier » — parler des **apps** Web + mobile.  
**Ne pas promettre une messagerie entre care pro** — parler de **partage de dossier / CR / notes**.  
Push FCM livré (messages véto → client, confirmation RDV) — détail [08](08-MESSAGERIE-NOTIFICATIONS.md).
