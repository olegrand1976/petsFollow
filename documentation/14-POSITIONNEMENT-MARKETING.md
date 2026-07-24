# Positionnement marketing — petsFollow

## Promesse

**Suivi cardiaque prescrit par le vétérinaire, sans boîtier.**  
**Pro** (Web + app clients) facturé **hors ligne** (69 € HT/mois + setup) ; **Pro Light** (app mobile **ProLight**) gratuit ; le propriétaire paie un abonnement par animal (~2–3,5 €/mois).

## Modèle

**B2B2C + SaaS cabinet** : commercial apporte le cabinet → véto prescrit → client paie (Stripe) ; abonnement Pro **facturé en externe** (pas de Stripe cabinet).

| Acteur | Bénéfice |
|--------|----------|
| Véto (Pro) | Suivi entre consultations + commissions pouvant compenser le SaaS ; CR vocaux + amélioration IA (Web) |
| Pro Light (ProLight) | App mobile terrain gratuite, avec ou sans compte Web Pro ; CR vocaux + amélioration IA |
| Client | App simple, relevé au doigt, lien direct avec son véto |
| Commercial | Commission sur chaque nouvelle activation + SPIFF ramp / mix |

## Offre cœur (TTC)

| Plan | Prix | Message |
|------|------|---------|
| Monthly | 3,50 € / mois | Flex (abo only) |
| Annual | 35 € / an | Entrée |
| **Triennial** | **95 € / 3 ans** (~2,6 €/mois) | **Recommandé** |

Care / Horse / foyer / encodage élevage : **inclus** dès entitlement animal actif (plus d’addons à vendre). Quinquennial hors vente.

## Différenciation

- Pas de hardware
- Prescription véto (pas un gadget grand public isolé)
- Dual-face Pro + pets, 5 langues
- Alignement économique véto + commercial (pas de pénalité co-selling)

## Matériel commercial

- Fiche produit : [22-FICHE-PRODUIT-COMMERCIAL.md](22-FICHE-PRODUIT-COMMERCIAL.md)
- Playbook : [21-GTM-COMMERCIAL.md](21-GTM-COMMERCIAL.md)
- Grille & commissions : [17](17-POLITIQUE-TARIFAIRE.md), [18](18-FICHE-COMMISSION-VETO.md), [19](19-FICHE-COMMISSION-COMMERCIAL.md)
- Page Pro pitch : `/commercial/pitch` (`ProCommissionSheet`)

## Interdits pitch

Ne pas promettre hardware, WebSocket temps réel, ni % sur TTC — voir interdits dans [22](22-FICHE-PRODUIT-COMMERCIAL.md).  
Push FCM livré (messages véto → client, confirmation RDV) — détail [08](08-MESSAGERIE-NOTIFICATIONS.md).
