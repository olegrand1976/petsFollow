# Use cases commerciaux — petsFollow

Scénarios de test **manuels** destinés aux commerciaux (et aux testeurs métier). Langage non-technique. Objectif : valider l’app sur **staging**, remonter bugs / frictions UX / idées d’évolution.

Réf. QA technique (équipe) : [`documentation/15-PLAN-TESTS.md`](../documentation/15-PLAN-TESTS.md).

---

## Glossaire

| Nom | Qui | Surface |
|-----|-----|---------|
| **VetPro** | Cabinet / vétérinaire | App **Web** cabinet |
| **VetLight / Pro Light** | Véto terrain ou care pro (maréchal, etc.) | App **mobile** Flutter (shell pro) |
| **Client** | Propriétaire d’animal | App **mobile** Flutter (shell particulier) |
| **Commercial** | Commercial terrain | App **Web** espace commercial |
| **Manager** | Responsable commercial | App **Web** pilotage équipe |
| **Admin** | Ops plateforme (interne) | App **Web** admin |

URL Web staging : [https://petsfollow.ll-it-sc.be](https://petsfollow.ll-it-sc.be)

---

## Environnement

### Staging (cible commerciaux)

| Surface | Accès |
|---------|--------|
| Web (tous rôles Pro) | [https://petsfollow.ll-it-sc.be](https://petsfollow.ll-it-sc.be) |
| App mobile | APK **staging** (Firebase App Distribution) |

Comptes démo : voir [`COMPTES.md`](COMPTES.md).

**Staging partagé** : plusieurs testeurs peuvent se marcher dessus (messages, Care, partage). Préférer des libellés de test uniques (`UC-X-01 test 14:32`). Les scénarios marqués **Destructif** consomment un compte seed — noter N/A ou demander un reset seed si déjà fait.

### Local (équipe tech uniquement)

```bash
make up-infra && make migrate && make seed
make api-dev          # :8291
make nuxtjs-dev       # :3002
```

---

## Comment utiliser

1. Choisir un fichier `UC-*.md` (ou la session démo ci-dessous).
2. Se connecter avec le(s) compte(s) indiqué(s).
3. Suivre les étapes ; cocher **OK / KO / N/A**.
4. Remonter un problème via [`TEMPLATE-RETOUR.md`](TEMPLATE-RETOUR.md) (bug, friction, idée).

**Solo vs croisé** : si vous avez déjà passé un parcours croisé (`UC-X-*`), vous pouvez **sauter** le solo équivalent — `UC-VP-03` ⊂ `UC-X-01` · `UC-CL-03` ⊂ `UC-X-02` · `UC-CO-02` ⊂ `UC-X-07`.

---

## Session démo recommandée (30–45 min)

Ordre pour un commercial sur staging :

1. [`UC-VP-01`](01-vetpro/UC-VP-01-cockpit-cabinet.md) — cockpit cabinet  
2. [`UC-VP-04`](01-vetpro/UC-VP-04-nouvelle-consultation.md) — consultation rapide (CR → DAF / facture)
3. [`UC-VP-05`](01-vetpro/UC-VP-05-pharmacie-stock-daf.md) — pharmacie stock → DAF → VAMReg (tag `dev`)
3. [`UC-X-01`](10-interactions/UC-X-01-messagerie-vet-client.md) — messagerie véto ↔ client *(couvre aussi VP-03)*  
4. [`UC-X-02`](10-interactions/UC-X-02-fc-client-vers-pro.md) — relevé cardiaque *(couvre aussi CL-03)*  
5. [`UC-X-05`](10-interactions/UC-X-05-partage-care-pro.md) — partage terrain  
6. [`UC-PL-01`](03-pro-light/UC-PL-01-agenda-terrain.md) — Pro Light  
7. [`UC-CL-02`](02-client/UC-CL-02-activation-paiement.md) — activation payante *(si temps — **Destructif**)*  
8. [`UC-CO-02`](04-commercial/UC-CO-02-encode-activation.md) — encode + commission *(si temps — **Destructif** ; ou `UC-X-07` pour le funnel complet)*

---

## Index des use cases

### VetPro — [`01-vetpro/`](01-vetpro/)

| ID | Fichier | Priorité |
|----|---------|----------|
| UC-VP-01 | [Cockpit cabinet](01-vetpro/UC-VP-01-cockpit-cabinet.md) | Démo |
| UC-VP-02 | [Onboarding](01-vetpro/UC-VP-02-onboarding.md) | Important — **Destructif** |
| UC-VP-03 | [Agenda & messagerie solo](01-vetpro/UC-VP-03-agenda-messagerie.md) | Démo — skip si X-01 |
| UC-VP-04 | [Nouvelle consultation](01-vetpro/UC-VP-04-nouvelle-consultation.md) | Démo |
| UC-VP-05 | [Pharmacie stock → DAF → VAMReg](01-vetpro/UC-VP-05-pharmacie-stock-daf.md) | Important — tag `dev` |

### Client — [`02-client/`](02-client/)

| ID | Fichier | Priorité |
|----|---------|----------|
| UC-CL-01 | [Parcours accueil](02-client/UC-CL-01-parcours-accueil.md) | Démo |
| UC-CL-02 | [Activation + paiement](02-client/UC-CL-02-activation-paiement.md) | Démo — **Destructif** |
| UC-CL-03 | [Relevé cardiaque solo](02-client/UC-CL-03-releve-cardiaque.md) | Démo — skip si X-02 |

### Pro Light — [`03-pro-light/`](03-pro-light/)

| ID | Fichier | Priorité |
|----|---------|----------|
| UC-PL-01 | [Agenda terrain](03-pro-light/UC-PL-01-agenda-terrain.md) | Démo |
| UC-PL-02 | [Switch profil pro ↔ client](03-pro-light/UC-PL-02-switch-profil.md) | Important |

### Commercial — [`04-commercial/`](04-commercial/)

| ID | Fichier | Priorité |
|----|---------|----------|
| UC-CO-01 | [Prospects CRM](04-commercial/UC-CO-01-prospects-crm.md) | Démo |
| UC-CO-02 | [Encode + activation](04-commercial/UC-CO-02-encode-activation.md) | Démo — **Destructif** — skip si X-07 |

### Manager — [`05-commercial-manager/`](05-commercial-manager/)

| ID | Fichier | Priorité |
|----|---------|----------|
| UC-CM-01 | [Dashboard équipe](05-commercial-manager/UC-CM-01-dashboard-equipe.md) | Important |

### Admin — [`06-admin/`](06-admin/)

| ID | Fichier | Priorité |
|----|---------|----------|
| UC-AD-01 | [Ops & support](06-admin/UC-AD-01-ops-support.md) | Secondaire |

### Équipe cabinet — [`07-equipe-cabinet/`](07-equipe-cabinet/)

| ID | Fichier | Priorité |
|----|---------|----------|
| UC-EQ-01 | [Rôles équipe](07-equipe-cabinet/UC-EQ-01-roles-team.md) | Important |
| UC-EQ-02 | [Switch poste partagé](07-equipe-cabinet/UC-EQ-02-switch-poste.md) | Important |

### Interactions — [`10-interactions/`](10-interactions/)

| ID | Fichier | Priorité |
|----|---------|----------|
| UC-X-01 | [Messagerie véto ↔ client](10-interactions/UC-X-01-messagerie-vet-client.md) | Démo |
| UC-X-02 | [FC client → VetPro](10-interactions/UC-X-02-fc-client-vers-pro.md) | Démo |
| UC-X-03 | [RDV bout-en-bout](10-interactions/UC-X-03-rdv-bout-en-bout.md) | Démo |
| UC-X-04 | [Lien cabinet](10-interactions/UC-X-04-lien-cabinet.md) | Important |
| UC-X-05 | [Partage care pro](10-interactions/UC-X-05-partage-care-pro.md) | Démo |
| UC-X-06 | [Continuité Care](10-interactions/UC-X-06-care-continuity.md) | Important |
| UC-X-07 | [Funnel commercial complet](10-interactions/UC-X-07-funnel-commercial-complet.md) | Démo — **Destructif** |
| UC-X-08 | [Envoi dossier animal → pro (lien 24 h)](10-interactions/UC-X-08-envoi-dossier-pro.md) | Démo |
| UC-X-09 | [Consultation client → partage PDF (lien 24 h)](10-interactions/UC-X-09-consultation-partage-pdf.md) | Démo |

---

## Maintenance (équipe / agents)

Quand une feature change un parcours utilisateur, un écran, un compte seed ou l’offre visible en démo :

1. Mettre à jour le(s) `UC-*.md` concerné(s) (étapes + résultat attendu).
2. Mettre à jour [`COMPTES.md`](COMPTES.md) si un compte seed change.
3. Ajuster la session démo / index ci-dessus si l’ordre ou les UC prioritaires changent.
4. Si le parcours est aussi P0/P1 QA → aligner `documentation/15-PLAN-TESTS.md`.

Règle Cursor : [`.cursor/rules/usecase-sync.mdc`](../.cursor/rules/usecase-sync.mdc).

Après toute modif UC : `make usecases-sync` (alimente la page Pro staging [`/usecases`](../nuxtjs/pages/usecases/index.vue)).

Convention : IDs `UC-XX-NN` **immuables** ; nouveau parcours → nouvel ID dans le bon dossier.
