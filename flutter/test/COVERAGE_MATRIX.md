# Flutter action coverage matrix

Legend: `✓` covered · `·` not yet · `~` partial

| Action | Unit | Widget | Smoke | Notes |
|--------|------|--------|-------|-------|
| Login email/mdp + erreurs | · | ✓ | ✓ | `login_screen_test` + smoke |
| Forgot / reset / confirm | · | ~ | · | confirm partiel ; resend depuis login (`login_resend_confirmation`) |
| Register + consent | · | ~ | · | social buttons existants |
| Home HR / poids keys | · | ✓ | · | `pet_quick_actions_test` |
| Weight sheet validate/save | ✓ | ✓ | ✓ | min 0.01 + POST mock + smoke |
| Pet.weightKg parse | ✓ | · | · | `pet_weight_test` |
| HR start/taps | · | ✓ | ✓ | flow + smoke cancel |
| HR keys start/validate | · | ✓ | · | `heart_rate_validate_keys_test` |
| Book visit | · | ✓ | · | `book_visit_screen_test` |
| Manage subscription portal | · | ✓ | · | `pet_manage_subscription_test` |
| Send pet dossier to pro | · | ✓ | · | `pet_send_dossier_test` (consentement PHI requis ; fiche + dialogue sans débordement en 360 dp clavier ouvert) |
| View / share consultation PDF | · | ✓ | · | `consultation_view_test` (timeline tap → CR + share consent) |
| New pet → save without payment / pay CTA | · | ✓ | ✓ | `pet_form_screen_test` sticky save + skipCheckout + pop/snackbar ; smoke create→list |
| Edit pet → puce + n° carnet | · | ✓ | · | `pet_edit_screen_test` PUT microchip/healthBook |
| New pet → sans cabinet (post-save link vet) | · | ✓ | · | `PetCreateResult` + snack host ; dialog Home seulement si déjà des vétos (`hasLinkedVets`) ; sinon bandeau |
| New pet → createPet parse envelope | ✓ | · | · | `create_pet_api_test` _asMap + entitlement |
| Kennel quick encode → POST /pets/batch | · | ✓ | · | `kennel_quick_encode_test` submit + empty skip + post-save link dialog |
| Add / suggest vet (lookup) | · | ✓ | · | `my_vets_screen_test` invite + suggest |
| Préconsult models + submit key | ✓ | ✓ | · | models + key contract |
| Care create/done/postpone keys | ✓ | ✓ | · | `care_actions_test` |
| Messaging send keys | · | ✓ | · | `messaging_keys_test` |
| Messaging attach camera/gallery keys | · | ✓ | · | `messaging_keys_test` |
| Messaging attach sheet flow | · | ✓ | · | `messaging_attach_sheet_test` |
| Messaging staff mode (composer + ensure clientUserId) | · | ✓ | · | `staff_messaging_test` |
| Messaging video upload basename | ✓ | · | · | `message_media_upload_test` |
| Settings logout | · | ✓ | · | `settings_logout_test` |
| Settings appearance Light/Dark | · | ✓ | · | `settings_appearance_test` |
| Payment deeplink path | ✓ | · | · | `payment_deeplink_test` |
| Pro Light CR keys | · | ✓ | · | `pro_light_cr_keys_test` (dictation/save/finalize ; no improve IA) |
| Pro Light consultation keys | · | ✓ | · | `pro_light_consultation_keys_test` (nouvelle consultation + CTA post-CR) |
| Pro Light nav 5 tabs | · | ✓ | · | `pro_light_shell_test` (Messages inclus care_pro) |
| Pro Light audio consent checkbox | · | ✓ | · | `pro_light_audio_consent_test` |
| Commercial logout key | · | ✓ | · | messaging_keys_test |
| Commercial invite QR dual links | · | ✓ | · | `app_invite_qr_screen_test` cabinet + client |
| Client invite QR (parrainage) | · | ✓ | · | `app_invite_qr_screen_test` single copy, no cabinet |
| Commercial manager team results | ✓ | ✓ | · | overview parse + CTA gated + list |
| Smoke login→HR→poids | · | · | ✓ | `test/smoke/…` — `accessToken` ; `make test-flutter-smoke` |

## Run

```bash
make test-flutter
# Smoke (API :8291 + seed) — skip si API down :
make test-flutter-smoke
```

## Rule

Toute nouvelle action mutante Flutter = Key stable + ligne ici + test widget (ou smoke) dans la même PR.
