# Pré-consultation — petsFollow (VetPro / lien public)

Formulaire client d’état de l’animal **sur opt-in VetPro** après confirmation d’un RDV (`visits.visits.request_preconsult`).

**VetPro** = surface Web Pro + permission `calendar.manage` (pas care_pro / VetLight seul).

## Flux

```text
Créer / confirmer RDV (checkbox « Demander une pré-consultation »)
  → visits.request_preconsult = true|false
  → Push FCM visit_confirmed (toujours)

  Si request_preconsult:
    → EnsurePreconsultPending
    → IssuePreconsultToken (opaque, TTL 30 j)
    → Email SendVisitPreconsult
         CTA = {ProPublicSiteURL}/preconsult/{token}
    → Client remplit GET/POST /api/v1/public/preconsult/{token}
    → Page remerciement (bénéfices + QR stores admin + lien invite)

  Sinon:
    → Email SendVisitConfirmedAffiliate
         CTA = /invite/{code} (affiliation cabinet / commission commercial)
```

L’auto-création d’intake + mail à **chaque** confirm est **désactivée**. Le formulaire Flutter auth reste secondaire ; la source de vérité est le **formulaire web public**.

## Schéma JSON `answers`

| Champ | Type | Valeurs |
|-------|------|---------|
| `chiefComplaint` | string | motif / plainte (requis ; plain text, pas de HTML) |
| `duration` | enum | `today` · `few_days` · `week` · `weeks` · `months` · `unknown` |
| `behavior` | enum | `normal` · `lethargic` · `restless` · `aggressive` · `anxious` · `other` · `unknown` |
| `appetite` | enum | `normal` · `decreased` · `increased` · `unknown` |
| `thirst` | enum | idem appetite |
| `elimination` | enum | idem appetite |
| `urgency` | enum | `low` · `medium` · `high` |
| `comment` | string | libre (optionnel ; plain text) |

Statuts intake : `pending` → `submitted` (ou `skipped` réservé).

## API

| Méthode | Route | Qui |
|---------|-------|-----|
| `POST/PATCH` | visite + `requestPreconsult` | staff `calendar.manage` |
| `GET` | `/visits/{visitID}/preconsult` | owner · staff · care_pro lecture |
| `PUT` | `/visits/{visitID}/preconsult` | owner client (legacy app) |
| `GET/POST` | `/public/preconsult/{token}` | public, rate-limité |
| `GET` | `/public/brand-assets` | QR stores publics |
| `GET/POST/PATCH` | `/admin/brand-assets` | admin (QR Android/iOS + URLs store) |

## Sécurité

- Token opaque (pas de PII dans l’URL), expiration, rate-limit authRL.
- Texte libre : trim, max length, **rejet** si balises/`<>` (`html_not_allowed`).
- Affichage Pro : plain text (pas de HTML brut).
- Pref client `visits` respectée pour les e-mails.

## QR stores & pack profil

- Assets globaux `platform.brand_assets` (`qr_android`, `qr_ios`) — admin `/admin/brand-assets`.
- Consommés sur page thanks pré-consult + payload invite (`/me/app-invite`).
- « Recalculer le pack » : recharge QR invite dynamique + QRs stores ; si asset absent → masqué / fallback lien download.

## Prefs / RGPD

- Export : `preconsultIntakes` dans `GET /me/export`.
- Purge : cascade visits → intakes / tokens.

## Hors scope

Deferred deep link Play Store natif, SMS, rappel J-1, templates par espèce, abonnement facturé « VetPro 69 € ».
