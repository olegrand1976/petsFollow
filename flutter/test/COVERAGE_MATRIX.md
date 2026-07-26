# Flutter action coverage matrix

Legend: `✓` covered · `·` not yet · `~` partial

| Action | Unit | Widget | Smoke | Notes |
|--------|------|--------|-------|-------|
| Login email/mdp + erreurs | · | ✓ | ✓ | `login_screen_test` + smoke |
| Forgot / reset / confirm | · | ~ | · | confirm partiel existant |
| Register + consent | · | ~ | · | social buttons existants |
| Home HR / poids keys | · | ✓ | · | `pet_quick_actions_test` |
| Weight sheet validate/save | ✓ | ✓ | ✓ | min 0.01 + POST mock + smoke |
| Pet.weightKg parse | ✓ | · | · | `pet_weight_test` |
| HR start/taps | · | ✓ | ✓ | flow + smoke cancel |
| HR keys start/validate | · | ✓ | · | `heart_rate_validate_keys_test` |
| Book visit | · | ✓ | · | `book_visit_screen_test` |
| Manage subscription portal | · | ✓ | · | `pet_manage_subscription_test` |
| New pet → continue payment | · | ✓ | · | `pet_form_screen_test` sticky CTA |
| Préconsult models + submit key | ✓ | ✓ | · | models + key contract |
| Care create/done/postpone keys | ✓ | ✓ | · | `care_actions_test` |
| Messaging send keys | · | ✓ | · | `messaging_keys_test` |
| Messaging attach camera/gallery keys | · | ✓ | · | `messaging_keys_test` |
| Messaging attach sheet flow | · | ✓ | · | `messaging_attach_sheet_test` |
| Messaging video upload basename | ✓ | · | · | `message_media_upload_test` |
| Messaging attach camera/gallery keys | · | ✓ | · | `messaging_keys_test` |
| Settings logout | · | ✓ | · | `settings_logout_test` |
| Payment deeplink path | ✓ | · | · | `payment_deeplink_test` |
| Pro Light CR keys | · | ✓ | · | `pro_light_cr_keys_test` |
| Pro Light audio consent checkbox | · | ✓ | · | `pro_light_audio_consent_test` |
| Commercial logout key | · | ✓ | · | messaging_keys_test |
| Smoke login→HR→poids | · | · | ✓ | `test/smoke/…` — `accessToken` ; `make test-flutter-smoke` |

## Run

```bash
make test-flutter
# Smoke (API :8291 + seed) — skip si API down :
make test-flutter-smoke
```

## Rule

Toute nouvelle action mutante Flutter = Key stable + ligne ici + test widget (ou smoke) dans la même PR.
