import 'dart:async';

import 'package:flutter/foundation.dart';
import 'package:flutter/widgets.dart';
import 'package:flutter_localizations/flutter_localizations.dart';
import 'package:intl/intl.dart' as intl;

import 'app_localizations_en.dart';
import 'app_localizations_es.dart';
import 'app_localizations_fr.dart';
import 'app_localizations_nl.dart';

// ignore_for_file: type=lint

/// Callers can lookup localized strings with an instance of AppLocalizations
/// returned by `AppLocalizations.of(context)`.
///
/// Applications need to include `AppLocalizations.delegate()` in their app's
/// `localizationDelegates` list, and the locales they support in the app's
/// `supportedLocales` list. For example:
///
/// ```dart
/// import 'l10n/app_localizations.dart';
///
/// return MaterialApp(
///   localizationsDelegates: AppLocalizations.localizationsDelegates,
///   supportedLocales: AppLocalizations.supportedLocales,
///   home: MyApplicationHome(),
/// );
/// ```
///
/// ## Update pubspec.yaml
///
/// Please make sure to update your pubspec.yaml to include the following
/// packages:
///
/// ```yaml
/// dependencies:
///   # Internationalization support.
///   flutter_localizations:
///     sdk: flutter
///   intl: any # Use the pinned version from flutter_localizations
///
///   # Rest of dependencies
/// ```
///
/// ## iOS Applications
///
/// iOS applications define key application metadata, including supported
/// locales, in an Info.plist file that is built into the application bundle.
/// To configure the locales supported by your app, you’ll need to edit this
/// file.
///
/// First, open your project’s ios/Runner.xcworkspace Xcode workspace file.
/// Then, in the Project Navigator, open the Info.plist file under the Runner
/// project’s Runner folder.
///
/// Next, select the Information Property List item, select Add Item from the
/// Editor menu, then select Localizations from the pop-up menu.
///
/// Select and expand the newly-created Localizations item then, for each
/// locale your application supports, add a new item and select the locale
/// you wish to add from the pop-up menu in the Value field. This list should
/// be consistent with the languages listed in the AppLocalizations.supportedLocales
/// property.
abstract class AppLocalizations {
  AppLocalizations(String locale)
      : localeName = intl.Intl.canonicalizedLocale(locale.toString());

  final String localeName;

  static AppLocalizations? of(BuildContext context) {
    return Localizations.of<AppLocalizations>(context, AppLocalizations);
  }

  static const LocalizationsDelegate<AppLocalizations> delegate =
      _AppLocalizationsDelegate();

  /// A list of this localizations delegate along with the default localizations
  /// delegates.
  ///
  /// Returns a list of localizations delegates containing this delegate along with
  /// GlobalMaterialLocalizations.delegate, GlobalCupertinoLocalizations.delegate,
  /// and GlobalWidgetsLocalizations.delegate.
  ///
  /// Additional delegates can be added by appending to this list in
  /// MaterialApp. This list does not have to be used at all if a custom list
  /// of delegates is preferred or required.
  static const List<LocalizationsDelegate<dynamic>> localizationsDelegates =
      <LocalizationsDelegate<dynamic>>[
    delegate,
    GlobalMaterialLocalizations.delegate,
    GlobalCupertinoLocalizations.delegate,
    GlobalWidgetsLocalizations.delegate,
  ];

  /// A list of this localizations delegate's supported locales.
  static const List<Locale> supportedLocales = <Locale>[
    Locale('en'),
    Locale('es'),
    Locale('fr'),
    Locale('nl')
  ];

  /// No description provided for @appTitle.
  ///
  /// In fr, this message translates to:
  /// **'petsFollow'**
  String get appTitle;

  /// No description provided for @appTagline.
  ///
  /// In fr, this message translates to:
  /// **'Suivi santé de votre animal'**
  String get appTagline;

  /// No description provided for @email.
  ///
  /// In fr, this message translates to:
  /// **'Email'**
  String get email;

  /// No description provided for @password.
  ///
  /// In fr, this message translates to:
  /// **'Mot de passe'**
  String get password;

  /// No description provided for @login.
  ///
  /// In fr, this message translates to:
  /// **'Se connecter'**
  String get login;

  /// No description provided for @loginFailed.
  ///
  /// In fr, this message translates to:
  /// **'Connexion impossible'**
  String get loginFailed;

  /// No description provided for @loginOr.
  ///
  /// In fr, this message translates to:
  /// **'ou'**
  String get loginOr;

  /// No description provided for @loginWithGoogle.
  ///
  /// In fr, this message translates to:
  /// **'Continuer avec Google'**
  String get loginWithGoogle;

  /// No description provided for @googleNotConfigured.
  ///
  /// In fr, this message translates to:
  /// **'Connexion Google non configurée'**
  String get googleNotConfigured;

  /// No description provided for @googleLoginFailed.
  ///
  /// In fr, this message translates to:
  /// **'Connexion Google impossible'**
  String get googleLoginFailed;

  /// No description provided for @googleClientNotFound.
  ///
  /// In fr, this message translates to:
  /// **'Aucun compte client pour cet email. Demandez une invitation à votre vétérinaire'**
  String get googleClientNotFound;

  /// No description provided for @googleWrongAudience.
  ///
  /// In fr, this message translates to:
  /// **'Ce compte Google n\'est pas un profil client'**
  String get googleWrongAudience;

  /// No description provided for @myPets.
  ///
  /// In fr, this message translates to:
  /// **'Mes animaux'**
  String get myPets;

  /// No description provided for @myData.
  ///
  /// In fr, this message translates to:
  /// **'Mes données'**
  String get myData;

  /// No description provided for @settings.
  ///
  /// In fr, this message translates to:
  /// **'Paramètres'**
  String get settings;

  /// No description provided for @logout.
  ///
  /// In fr, this message translates to:
  /// **'Fermer la session'**
  String get logout;

  /// No description provided for @save.
  ///
  /// In fr, this message translates to:
  /// **'Sauvegarder'**
  String get save;

  /// No description provided for @cancel.
  ///
  /// In fr, this message translates to:
  /// **'Annuler'**
  String get cancel;

  /// No description provided for @firstName.
  ///
  /// In fr, this message translates to:
  /// **'Votre prénom'**
  String get firstName;

  /// No description provided for @currentPassword.
  ///
  /// In fr, this message translates to:
  /// **'Mot de passe actuel'**
  String get currentPassword;

  /// No description provided for @newPassword.
  ///
  /// In fr, this message translates to:
  /// **'Nouveau mot de passe'**
  String get newPassword;

  /// No description provided for @confirmNewPassword.
  ///
  /// In fr, this message translates to:
  /// **'Confirmer le mot de passe'**
  String get confirmNewPassword;

  /// No description provided for @changePassword.
  ///
  /// In fr, this message translates to:
  /// **'Changer le mot de passe'**
  String get changePassword;

  /// No description provided for @forceChangePasswordTitle.
  ///
  /// In fr, this message translates to:
  /// **'Changer le mot de passe'**
  String get forceChangePasswordTitle;

  /// No description provided for @forceChangePasswordSubtitle.
  ///
  /// In fr, this message translates to:
  /// **'Ce compte a été créé avec un mot de passe temporaire. Choisissez le vôtre pour continuer.'**
  String get forceChangePasswordSubtitle;

  /// No description provided for @forceChangePasswordSubmit.
  ///
  /// In fr, this message translates to:
  /// **'Enregistrer et continuer'**
  String get forceChangePasswordSubmit;

  /// No description provided for @passwordTooShort.
  ///
  /// In fr, this message translates to:
  /// **'Minimum 8 caractères'**
  String get passwordTooShort;

  /// No description provided for @passwordMismatch.
  ///
  /// In fr, this message translates to:
  /// **'Les mots de passe ne correspondent pas'**
  String get passwordMismatch;

  /// No description provided for @passwordChangeFailed.
  ///
  /// In fr, this message translates to:
  /// **'Impossible de modifier le mot de passe'**
  String get passwordChangeFailed;

  /// No description provided for @deleteAccount.
  ///
  /// In fr, this message translates to:
  /// **'Supprimer le compte'**
  String get deleteAccount;

  /// No description provided for @deleteAccountConfirm.
  ///
  /// In fr, this message translates to:
  /// **'Cette action est irréversible. Tous vos animaux et données seront supprimés.'**
  String get deleteAccountConfirm;

  /// No description provided for @profileSaved.
  ///
  /// In fr, this message translates to:
  /// **'Profil enregistré'**
  String get profileSaved;

  /// No description provided for @changePhoto.
  ///
  /// In fr, this message translates to:
  /// **'Changer la photo'**
  String get changePhoto;

  /// No description provided for @addPhoto.
  ///
  /// In fr, this message translates to:
  /// **'Ajouter une photo'**
  String get addPhoto;

  /// No description provided for @photoUpdated.
  ///
  /// In fr, this message translates to:
  /// **'Photo mise à jour'**
  String get photoUpdated;

  /// No description provided for @passwordChanged.
  ///
  /// In fr, this message translates to:
  /// **'Mot de passe modifié'**
  String get passwordChanged;

  /// No description provided for @greeting.
  ///
  /// In fr, this message translates to:
  /// **'Bonjour {name},'**
  String greeting(String name);

  /// No description provided for @latestValues.
  ///
  /// In fr, this message translates to:
  /// **'Dernières valeurs'**
  String get latestValues;

  /// No description provided for @startMeasurement.
  ///
  /// In fr, this message translates to:
  /// **'DÉMARRER LA MESURE'**
  String get startMeasurement;

  /// No description provided for @chooseDuration.
  ///
  /// In fr, this message translates to:
  /// **'Durée de la mesure'**
  String get chooseDuration;

  /// No description provided for @durationSeconds.
  ///
  /// In fr, this message translates to:
  /// **'{seconds} s'**
  String durationSeconds(int seconds);

  /// No description provided for @howToMeasure.
  ///
  /// In fr, this message translates to:
  /// **'Comment mesurer ?'**
  String get howToMeasure;

  /// No description provided for @howToMeasureIntro.
  ///
  /// In fr, this message translates to:
  /// **'Mesurer la fréquence cardiaque de votre animal au repos.'**
  String get howToMeasureIntro;

  /// No description provided for @howToMeasureStep1.
  ///
  /// In fr, this message translates to:
  /// **'1. Placez votre animal au calme, allongé ou assis.'**
  String get howToMeasureStep1;

  /// No description provided for @howToMeasureStep2.
  ///
  /// In fr, this message translates to:
  /// **'2. Placez votre main sur le thorax et tapez à chaque battement pendant la durée indiquée.'**
  String get howToMeasureStep2;

  /// No description provided for @howToMeasureStep3.
  ///
  /// In fr, this message translates to:
  /// **'3. Validez le relevé pour l\'envoyer à votre vétérinaire.'**
  String get howToMeasureStep3;

  /// No description provided for @howToMeasureWhyTitle.
  ///
  /// In fr, this message translates to:
  /// **'Pourquoi mesurer ?'**
  String get howToMeasureWhyTitle;

  /// No description provided for @howToMeasureWhyBody.
  ///
  /// In fr, this message translates to:
  /// **'Le suivi régulier de la fréquence cardiaque permet de détecter des variations et d\'adapter le traitement avec votre vétérinaire.'**
  String get howToMeasureWhyBody;

  /// No description provided for @reminders.
  ///
  /// In fr, this message translates to:
  /// **'Rappels'**
  String get reminders;

  /// No description provided for @remindersHint.
  ///
  /// In fr, this message translates to:
  /// **'Recevez un rappel quotidien pour effectuer un relevé cardiaque.'**
  String get remindersHint;

  /// No description provided for @remindersEnabled.
  ///
  /// In fr, this message translates to:
  /// **'Activer les rappels'**
  String get remindersEnabled;

  /// No description provided for @remindersTime.
  ///
  /// In fr, this message translates to:
  /// **'Heure du rappel'**
  String get remindersTime;

  /// No description provided for @remindersSaved.
  ///
  /// In fr, this message translates to:
  /// **'Rappels enregistrés'**
  String get remindersSaved;

  /// No description provided for @legalTermsTitle.
  ///
  /// In fr, this message translates to:
  /// **'Conditions générales d\'utilisation'**
  String get legalTermsTitle;

  /// No description provided for @legalPrivacyTitle.
  ///
  /// In fr, this message translates to:
  /// **'Politique de confidentialité'**
  String get legalPrivacyTitle;

  /// No description provided for @legalNoticeTitle.
  ///
  /// In fr, this message translates to:
  /// **'Mentions légales'**
  String get legalNoticeTitle;

  /// No description provided for @legalTermsBody.
  ///
  /// In fr, this message translates to:
  /// **'Conditions générales d\'utilisation — petsFollow\n\nL\'application petsFollow permet aux propriétaires d\'animaux de mesurer la fréquence cardiaque, de consulter l\'historique et de communiquer avec leur vétérinaire.\n\nLes services sont fournis dans le cadre de l\'abonnement choisi. L\'utilisateur s\'engage à utiliser l\'application conformément à sa destination.\n\nDate d\'actualisation : juillet 2026'**
  String get legalTermsBody;

  /// No description provided for @legalPrivacyBody.
  ///
  /// In fr, this message translates to:
  /// **'Politique de confidentialité — petsFollow\n\nDonnées collectées : prénom, email, données animal (nom, espèce, race), relevés cardiaques, messages au vétérinaire.\n\nFinalités : gestion du compte, suivi santé, communication avec le cabinet vétérinaire.\n\nConservation : jusqu\'à suppression du compte ou 3 ans d\'inactivité.\n\nVous pouvez exercer vos droits (accès, rectification, suppression) via les paramètres de l\'application.\n\nDate d\'actualisation : juillet 2026'**
  String get legalPrivacyBody;

  /// No description provided for @legalNoticeBody.
  ///
  /// In fr, this message translates to:
  /// **'Mentions légales — petsFollow\n\nÉditeur : petsFollow\nContact : support@petsfollow.test\n\nHébergement : infrastructure cloud conforme RGPD.\n\nDirecteur de publication : petsFollow.\n\nDate d\'actualisation : juillet 2026'**
  String get legalNoticeBody;

  /// No description provided for @language.
  ///
  /// In fr, this message translates to:
  /// **'Langue'**
  String get language;

  /// No description provided for @languageFr.
  ///
  /// In fr, this message translates to:
  /// **'Français'**
  String get languageFr;

  /// No description provided for @languageNl.
  ///
  /// In fr, this message translates to:
  /// **'Nederlands'**
  String get languageNl;

  /// No description provided for @languageEn.
  ///
  /// In fr, this message translates to:
  /// **'English'**
  String get languageEn;

  /// No description provided for @languageEs.
  ///
  /// In fr, this message translates to:
  /// **'Español'**
  String get languageEs;

  /// No description provided for @paymentResume.
  ///
  /// In fr, this message translates to:
  /// **'Reprendre le paiement'**
  String get paymentResume;

  /// No description provided for @manageSubscription.
  ///
  /// In fr, this message translates to:
  /// **'Gérer mon abonnement'**
  String get manageSubscription;

  /// No description provided for @heartRate.
  ///
  /// In fr, this message translates to:
  /// **'Relevé cardiaque'**
  String get heartRate;

  /// No description provided for @history.
  ///
  /// In fr, this message translates to:
  /// **'Historique'**
  String get history;

  /// No description provided for @vetMessaging.
  ///
  /// In fr, this message translates to:
  /// **'Messagerie véto'**
  String get vetMessaging;

  /// No description provided for @badgeAutoRenew.
  ///
  /// In fr, this message translates to:
  /// **'Renouvellement auto'**
  String get badgeAutoRenew;

  /// No description provided for @badgeActive.
  ///
  /// In fr, this message translates to:
  /// **'Actif'**
  String get badgeActive;

  /// No description provided for @badgePendingPayment.
  ///
  /// In fr, this message translates to:
  /// **'En attente de paiement'**
  String get badgePendingPayment;

  /// No description provided for @badgeExpiresOn.
  ///
  /// In fr, this message translates to:
  /// **'expire {date}'**
  String badgeExpiresOn(String date);

  /// No description provided for @newPet.
  ///
  /// In fr, this message translates to:
  /// **'Nouvel animal'**
  String get newPet;

  /// No description provided for @petName.
  ///
  /// In fr, this message translates to:
  /// **'Nom'**
  String get petName;

  /// No description provided for @species.
  ///
  /// In fr, this message translates to:
  /// **'Espèce'**
  String get species;

  /// No description provided for @breed.
  ///
  /// In fr, this message translates to:
  /// **'Race'**
  String get breed;

  /// No description provided for @choosePlan.
  ///
  /// In fr, this message translates to:
  /// **'Choisissez votre formule'**
  String get choosePlan;

  /// No description provided for @recommended.
  ///
  /// In fr, this message translates to:
  /// **'Recommandé'**
  String get recommended;

  /// No description provided for @autoRenewTitle.
  ///
  /// In fr, this message translates to:
  /// **'Renouveler automatiquement'**
  String get autoRenewTitle;

  /// No description provided for @autoRenewSubtitle.
  ///
  /// In fr, this message translates to:
  /// **'Prélèvement à chaque échéance'**
  String get autoRenewSubtitle;

  /// No description provided for @continueToPayment.
  ///
  /// In fr, this message translates to:
  /// **'Continuer vers le paiement'**
  String get continueToPayment;

  /// No description provided for @paymentConfirmed.
  ///
  /// In fr, this message translates to:
  /// **'Paiement confirmé — animal actif'**
  String get paymentConfirmed;

  /// No description provided for @paymentPending.
  ///
  /// In fr, this message translates to:
  /// **'Paiement en attente — vous pourrez reprendre plus tard'**
  String get paymentPending;

  /// No description provided for @errorGeneric.
  ///
  /// In fr, this message translates to:
  /// **'Erreur: {message}'**
  String errorGeneric(String message);

  /// No description provided for @errorMediaTooLarge.
  ///
  /// In fr, this message translates to:
  /// **'Fichier trop volumineux (25 Mo max)'**
  String get errorMediaTooLarge;

  /// No description provided for @errorInvalidMediaType.
  ///
  /// In fr, this message translates to:
  /// **'Format non supporté (JPEG, PNG, WebP, MP4, MOV, WebM)'**
  String get errorInvalidMediaType;

  /// No description provided for @errorPaymentRequired.
  ///
  /// In fr, this message translates to:
  /// **'Abonnement requis pour envoyer des médias'**
  String get errorPaymentRequired;

  /// No description provided for @errorPhotoUploadFailed.
  ///
  /// In fr, this message translates to:
  /// **'Animal créé, mais la photo n\'a pas pu être envoyée'**
  String get errorPhotoUploadFailed;

  /// No description provided for @errorCouldNotOpenLink.
  ///
  /// In fr, this message translates to:
  /// **'Impossible d\'ouvrir le lien'**
  String get errorCouldNotOpenLink;

  /// No description provided for @planAnnualSub.
  ///
  /// In fr, this message translates to:
  /// **'{price}, renouvelé automatiquement'**
  String planAnnualSub(String price);

  /// No description provided for @planTriennialSub.
  ///
  /// In fr, this message translates to:
  /// **'95 € tous les 3 ans, renouvelé automatiquement'**
  String get planTriennialSub;

  /// No description provided for @planQuinquennialSub.
  ///
  /// In fr, this message translates to:
  /// **'145 € pour 5 ans, paiement unique'**
  String get planQuinquennialSub;

  /// No description provided for @planOneTime.
  ///
  /// In fr, this message translates to:
  /// **'{price}, paiement unique'**
  String planOneTime(String price);

  /// No description provided for @heartRateInstructions.
  ///
  /// In fr, this message translates to:
  /// **'Tapotez à chaque battement pendant la durée indiquée par votre vétérinaire.'**
  String get heartRateInstructions;

  /// No description provided for @heartRateInstructionsDuration.
  ///
  /// In fr, this message translates to:
  /// **'Tapotez à chaque battement pendant {seconds} secondes.'**
  String heartRateInstructionsDuration(int seconds);

  /// No description provided for @heartRateNoDurationConfigured.
  ///
  /// In fr, this message translates to:
  /// **'Aucune durée de mesure n’est configurée pour ce cabinet. Contactez votre vétérinaire.'**
  String get heartRateNoDurationConfigured;

  /// No description provided for @start.
  ///
  /// In fr, this message translates to:
  /// **'Démarrer'**
  String get start;

  /// No description provided for @secondsLeft.
  ///
  /// In fr, this message translates to:
  /// **'{seconds} s'**
  String secondsLeft(int seconds);

  /// No description provided for @beatsCount.
  ///
  /// In fr, this message translates to:
  /// **'{count} battements'**
  String beatsCount(int count);

  /// No description provided for @tapHere.
  ///
  /// In fr, this message translates to:
  /// **'Tapez ici à chaque battement'**
  String get tapHere;

  /// No description provided for @bpmLabel.
  ///
  /// In fr, this message translates to:
  /// **'BPM: {bpm}'**
  String bpmLabel(String bpm);

  /// No description provided for @beatsLabel.
  ///
  /// In fr, this message translates to:
  /// **'Battements: {count}'**
  String beatsLabel(int count);

  /// No description provided for @thresholdAlert.
  ///
  /// In fr, this message translates to:
  /// **'Alerte seuil'**
  String get thresholdAlert;

  /// No description provided for @validateAndSend.
  ///
  /// In fr, this message translates to:
  /// **'Valider et envoyer au véto'**
  String get validateAndSend;

  /// No description provided for @restart.
  ///
  /// In fr, this message translates to:
  /// **'Recommencer'**
  String get restart;

  /// No description provided for @sentToVet.
  ///
  /// In fr, this message translates to:
  /// **'Relevé envoyé au véto'**
  String get sentToVet;

  /// No description provided for @navHome.
  ///
  /// In fr, this message translates to:
  /// **'Accueil'**
  String get navHome;

  /// No description provided for @navPets.
  ///
  /// In fr, this message translates to:
  /// **'Animaux'**
  String get navPets;

  /// No description provided for @navCare.
  ///
  /// In fr, this message translates to:
  /// **'Soins'**
  String get navCare;

  /// No description provided for @navMessages.
  ///
  /// In fr, this message translates to:
  /// **'Messages'**
  String get navMessages;

  /// No description provided for @navProfile.
  ///
  /// In fr, this message translates to:
  /// **'Profil'**
  String get navProfile;

  /// No description provided for @speciesDog.
  ///
  /// In fr, this message translates to:
  /// **'Chien'**
  String get speciesDog;

  /// No description provided for @speciesCat.
  ///
  /// In fr, this message translates to:
  /// **'Chat'**
  String get speciesCat;

  /// No description provided for @speciesHorse.
  ///
  /// In fr, this message translates to:
  /// **'Cheval'**
  String get speciesHorse;

  /// No description provided for @speciesOther.
  ///
  /// In fr, this message translates to:
  /// **'Autre'**
  String get speciesOther;

  /// No description provided for @careComingSoon.
  ///
  /// In fr, this message translates to:
  /// **'Les rappels de soins arrivent bientôt'**
  String get careComingSoon;

  /// No description provided for @emptyPetsTitle.
  ///
  /// In fr, this message translates to:
  /// **'Aucun animal'**
  String get emptyPetsTitle;

  /// No description provided for @emptyPetsBody.
  ///
  /// In fr, this message translates to:
  /// **'Ajoutez votre premier animal pour commencer le suivi cardiaque avec votre vétérinaire.'**
  String get emptyPetsBody;

  /// No description provided for @discoveryTitle.
  ///
  /// In fr, this message translates to:
  /// **'Découvrir petsFollow'**
  String get discoveryTitle;

  /// No description provided for @discoveryMission.
  ///
  /// In fr, this message translates to:
  /// **'Votre parcours en 7 jours'**
  String get discoveryMission;

  /// No description provided for @discoveryDay0Title.
  ///
  /// In fr, this message translates to:
  /// **'Jour 0 — Bienvenue'**
  String get discoveryDay0Title;

  /// No description provided for @discoveryDay0Body.
  ///
  /// In fr, this message translates to:
  /// **'Créez le profil de votre animal et découvrez comment mesurer sa fréquence cardiaque.'**
  String get discoveryDay0Body;

  /// No description provided for @discoveryDay2Title.
  ///
  /// In fr, this message translates to:
  /// **'Jour 2 — Première mesure'**
  String get discoveryDay2Title;

  /// No description provided for @discoveryDay2Body.
  ///
  /// In fr, this message translates to:
  /// **'Effectuez votre premier relevé cardiaque et familiarisez-vous avec la technique.'**
  String get discoveryDay2Body;

  /// No description provided for @discoveryDay4Title.
  ///
  /// In fr, this message translates to:
  /// **'Jour 4 — Routine'**
  String get discoveryDay4Title;

  /// No description provided for @discoveryDay4Body.
  ///
  /// In fr, this message translates to:
  /// **'Installez une routine de mesure quotidienne avec les rappels personnalisés.'**
  String get discoveryDay4Body;

  /// No description provided for @discoveryDay6Title.
  ///
  /// In fr, this message translates to:
  /// **'Jour 6 — Partage véto'**
  String get discoveryDay6Title;

  /// No description provided for @discoveryDay6Body.
  ///
  /// In fr, this message translates to:
  /// **'Vos relevés sont partagés avec votre vétérinaire pour un suivi optimal.'**
  String get discoveryDay6Body;

  /// No description provided for @myVets.
  ///
  /// In fr, this message translates to:
  /// **'Mes vétérinaires'**
  String get myVets;

  /// No description provided for @addVetByEmail.
  ///
  /// In fr, this message translates to:
  /// **'Ajouter un véto par email'**
  String get addVetByEmail;

  /// No description provided for @vetEmailHint.
  ///
  /// In fr, this message translates to:
  /// **'email@cabinet.vet'**
  String get vetEmailHint;

  /// No description provided for @noVets.
  ///
  /// In fr, this message translates to:
  /// **'Aucun vétérinaire lié'**
  String get noVets;

  /// No description provided for @primaryVet.
  ///
  /// In fr, this message translates to:
  /// **'Vétérinaire principal'**
  String get primaryVet;

  /// No description provided for @setPrimaryVet.
  ///
  /// In fr, this message translates to:
  /// **'Définir comme véto principal'**
  String get setPrimaryVet;

  /// No description provided for @careTitle.
  ///
  /// In fr, this message translates to:
  /// **'Soins'**
  String get careTitle;

  /// No description provided for @careDone.
  ///
  /// In fr, this message translates to:
  /// **'Fait'**
  String get careDone;

  /// No description provided for @carePostpone.
  ///
  /// In fr, this message translates to:
  /// **'Reporter'**
  String get carePostpone;

  /// No description provided for @careOverdue.
  ///
  /// In fr, this message translates to:
  /// **'En retard'**
  String get careOverdue;

  /// No description provided for @visitHistory.
  ///
  /// In fr, this message translates to:
  /// **'Historique des visites'**
  String get visitHistory;

  /// No description provided for @requestVisit.
  ///
  /// In fr, this message translates to:
  /// **'Demander une visite'**
  String get requestVisit;

  /// No description provided for @calendarBookingDisabled.
  ///
  /// In fr, this message translates to:
  /// **'La réservation en ligne n\'est pas disponible pour ce cabinet. Vous pouvez envoyer une demande sans créneau.'**
  String get calendarBookingDisabled;

  /// No description provided for @calendarBookingDisabledReschedule.
  ///
  /// In fr, this message translates to:
  /// **'La réservation en ligne n\'est pas disponible. Proposez une date manuellement.'**
  String get calendarBookingDisabledReschedule;

  /// No description provided for @calendarNoSlots.
  ///
  /// In fr, this message translates to:
  /// **'Aucun créneau disponible sur les 14 prochains jours.'**
  String get calendarNoSlots;

  /// No description provided for @calendarPickSlot.
  ///
  /// In fr, this message translates to:
  /// **'Choisissez un créneau :'**
  String get calendarPickSlot;

  /// No description provided for @visitConfirm.
  ///
  /// In fr, this message translates to:
  /// **'Confirmer'**
  String get visitConfirm;

  /// No description provided for @visitProposeReschedule.
  ///
  /// In fr, this message translates to:
  /// **'Proposer un autre créneau'**
  String get visitProposeReschedule;

  /// No description provided for @visitRescheduleProposed.
  ///
  /// In fr, this message translates to:
  /// **'Proposition de déplacement envoyée'**
  String get visitRescheduleProposed;

  /// No description provided for @paymentSuccessSnack.
  ///
  /// In fr, this message translates to:
  /// **'Paiement reçu — actualisation…'**
  String get paymentSuccessSnack;

  /// No description provided for @paymentCancelSnack.
  ///
  /// In fr, this message translates to:
  /// **'Paiement annulé'**
  String get paymentCancelSnack;

  /// No description provided for @visitRejectReschedule.
  ///
  /// In fr, this message translates to:
  /// **'Refuser le déplacement'**
  String get visitRejectReschedule;

  /// No description provided for @visitAcceptReschedule.
  ///
  /// In fr, this message translates to:
  /// **'Accepter le nouveau créneau'**
  String get visitAcceptReschedule;

  /// No description provided for @upcomingVisit.
  ///
  /// In fr, this message translates to:
  /// **'Visite à venir'**
  String get upcomingVisit;

  /// No description provided for @notificationHrTitle.
  ///
  /// In fr, this message translates to:
  /// **'petsFollow'**
  String get notificationHrTitle;

  /// No description provided for @notificationHrBody.
  ///
  /// In fr, this message translates to:
  /// **'Il est temps de prendre un relevé cardiaque pour votre animal'**
  String get notificationHrBody;

  /// No description provided for @reviewAskTitle.
  ///
  /// In fr, this message translates to:
  /// **'Vous aimez petsFollow ?'**
  String get reviewAskTitle;

  /// No description provided for @reviewAskYes.
  ///
  /// In fr, this message translates to:
  /// **'Oui, noter l\'app'**
  String get reviewAskYes;

  /// No description provided for @reviewAskNo.
  ///
  /// In fr, this message translates to:
  /// **'Plus tard'**
  String get reviewAskNo;

  /// No description provided for @carePlusUpsell.
  ///
  /// In fr, this message translates to:
  /// **'Care+ — médicaments et rappels personnalisés'**
  String get carePlusUpsell;

  /// No description provided for @carePlusRequired.
  ///
  /// In fr, this message translates to:
  /// **'Care+ est requis pour les médicaments et rappels personnalisés.'**
  String get carePlusRequired;

  /// No description provided for @horsePackRequired.
  ///
  /// In fr, this message translates to:
  /// **'Le Pack Cheval est requis pour les rappels maréchal, contacts et compétitions.'**
  String get horsePackRequired;

  /// No description provided for @activateAddon.
  ///
  /// In fr, this message translates to:
  /// **'Activer'**
  String get activateAddon;

  /// No description provided for @careTypeMedication.
  ///
  /// In fr, this message translates to:
  /// **'Médicament'**
  String get careTypeMedication;

  /// No description provided for @horseAddContact.
  ///
  /// In fr, this message translates to:
  /// **'Ajouter un contact'**
  String get horseAddContact;

  /// No description provided for @horseAddCompetition.
  ///
  /// In fr, this message translates to:
  /// **'Ajouter une compétition'**
  String get horseAddCompetition;

  /// No description provided for @horseContactName.
  ///
  /// In fr, this message translates to:
  /// **'Nom'**
  String get horseContactName;

  /// No description provided for @horseContactRole.
  ///
  /// In fr, this message translates to:
  /// **'Rôle'**
  String get horseContactRole;

  /// No description provided for @horseCompetitionTitle.
  ///
  /// In fr, this message translates to:
  /// **'Événement'**
  String get horseCompetitionTitle;

  /// No description provided for @horseCompetitionDate.
  ///
  /// In fr, this message translates to:
  /// **'Date (AAAA-MM-JJ)'**
  String get horseCompetitionDate;

  /// No description provided for @familyPackHint.
  ///
  /// In fr, this message translates to:
  /// **'Pack Famille — vue foyer des rappels, −10 % dès le 2ᵉ abonnement animal payant'**
  String get familyPackHint;

  /// No description provided for @familyHouseholdTitle.
  ///
  /// In fr, this message translates to:
  /// **'Foyer Famille — {count} animaux'**
  String familyHouseholdTitle(int count);

  /// No description provided for @kennelHouseholdTitle.
  ///
  /// In fr, this message translates to:
  /// **'Foyer Élevage — {count} animaux'**
  String kennelHouseholdTitle(int count);

  /// No description provided for @familyHouseholdNext.
  ///
  /// In fr, this message translates to:
  /// **'Prochains rappels du foyer'**
  String get familyHouseholdNext;

  /// No description provided for @familyPetLimit.
  ///
  /// In fr, this message translates to:
  /// **'Un pack foyer est déjà actif ou en cours d\'achat'**
  String get familyPetLimit;

  /// No description provided for @familyRequiresTwoPets.
  ///
  /// In fr, this message translates to:
  /// **'Le pack Famille nécessite au moins 2 animaux'**
  String get familyRequiresTwoPets;

  /// No description provided for @kennelPackHint.
  ///
  /// In fr, this message translates to:
  /// **'Pack Élevage — ≥6 animaux, −15 % sur les abos suivants'**
  String get kennelPackHint;

  /// No description provided for @kennelRequiresSixPets.
  ///
  /// In fr, this message translates to:
  /// **'Le pack Élevage nécessite au moins 6 animaux'**
  String get kennelRequiresSixPets;

  /// No description provided for @kennelQuickEncodeTitle.
  ///
  /// In fr, this message translates to:
  /// **'Encodage portée (élevage)'**
  String get kennelQuickEncodeTitle;

  /// No description provided for @kennelRequired.
  ///
  /// In fr, this message translates to:
  /// **'Le pack Élevage est requis pour l\'encodage par lot'**
  String get kennelRequired;

  /// No description provided for @litterTag.
  ///
  /// In fr, this message translates to:
  /// **'Tag portée'**
  String get litterTag;

  /// No description provided for @discoveryMarkDone.
  ///
  /// In fr, this message translates to:
  /// **'Mission accomplie'**
  String get discoveryMarkDone;

  /// No description provided for @notificationPreferences.
  ///
  /// In fr, this message translates to:
  /// **'Préférences de notifications'**
  String get notificationPreferences;

  /// No description provided for @notificationPrefsHint.
  ///
  /// In fr, this message translates to:
  /// **'Choisissez les types de notifications que vous souhaitez recevoir.'**
  String get notificationPrefsHint;

  /// No description provided for @notificationPrefsSaved.
  ///
  /// In fr, this message translates to:
  /// **'Préférences enregistrées'**
  String get notificationPrefsSaved;

  /// No description provided for @notificationPrefHr.
  ///
  /// In fr, this message translates to:
  /// **'Relevés cardiaques'**
  String get notificationPrefHr;

  /// No description provided for @notificationPrefCare.
  ///
  /// In fr, this message translates to:
  /// **'Rappels de soins'**
  String get notificationPrefCare;

  /// No description provided for @notificationPrefVisits.
  ///
  /// In fr, this message translates to:
  /// **'Visites'**
  String get notificationPrefVisits;

  /// No description provided for @notificationPrefMessages.
  ///
  /// In fr, this message translates to:
  /// **'Messages'**
  String get notificationPrefMessages;

  /// No description provided for @notificationPrefDiscovery.
  ///
  /// In fr, this message translates to:
  /// **'Parcours découverte'**
  String get notificationPrefDiscovery;

  /// No description provided for @notificationPrefBilling.
  ///
  /// In fr, this message translates to:
  /// **'Facturation'**
  String get notificationPrefBilling;

  /// No description provided for @carePostponeDays.
  ///
  /// In fr, this message translates to:
  /// **'Reporter de {days} jours'**
  String carePostponeDays(int days);

  /// No description provided for @noCareReminders.
  ///
  /// In fr, this message translates to:
  /// **'Aucun rappel de soin en cours'**
  String get noCareReminders;

  /// No description provided for @careAddReminder.
  ///
  /// In fr, this message translates to:
  /// **'Ajouter un rappel'**
  String get careAddReminder;

  /// No description provided for @careSelectPet.
  ///
  /// In fr, this message translates to:
  /// **'Animal'**
  String get careSelectPet;

  /// No description provided for @careDueInDays.
  ///
  /// In fr, this message translates to:
  /// **'Échéance dans {days} jours'**
  String careDueInDays(int days);

  /// No description provided for @discoveryDayBadge.
  ///
  /// In fr, this message translates to:
  /// **'J{day}'**
  String discoveryDayBadge(int day);

  /// No description provided for @timelineTypeHeartrate.
  ///
  /// In fr, this message translates to:
  /// **'Fréquence cardiaque'**
  String get timelineTypeHeartrate;

  /// No description provided for @timelineTypeMessage.
  ///
  /// In fr, this message translates to:
  /// **'Message'**
  String get timelineTypeMessage;

  /// No description provided for @timelineTypeCare.
  ///
  /// In fr, this message translates to:
  /// **'Soin'**
  String get timelineTypeCare;

  /// No description provided for @timelineTypeVisit.
  ///
  /// In fr, this message translates to:
  /// **'Visite'**
  String get timelineTypeVisit;

  /// No description provided for @timelineTypeEvent.
  ///
  /// In fr, this message translates to:
  /// **'Événement'**
  String get timelineTypeEvent;

  /// No description provided for @visitCancelAction.
  ///
  /// In fr, this message translates to:
  /// **'Annuler la demande'**
  String get visitCancelAction;

  /// No description provided for @upcomingVisits.
  ///
  /// In fr, this message translates to:
  /// **'Prochaines visites'**
  String get upcomingVisits;

  /// No description provided for @timelineEmpty.
  ///
  /// In fr, this message translates to:
  /// **'Aucun événement pour le moment'**
  String get timelineEmpty;

  /// No description provided for @noThreads.
  ///
  /// In fr, this message translates to:
  /// **'Aucune conversation'**
  String get noThreads;

  /// No description provided for @vetInviteSent.
  ///
  /// In fr, this message translates to:
  /// **'Invitation envoyée — le cabinet doit accepter la demande'**
  String get vetInviteSent;

  /// No description provided for @vetInviteSentNamed.
  ///
  /// In fr, this message translates to:
  /// **'Demande envoyée à {practice} — le cabinet doit l’accepter'**
  String vetInviteSentNamed(String practice);

  /// No description provided for @vetNotFound.
  ///
  /// In fr, this message translates to:
  /// **'Aucun vétérinaire trouvé avec cet email'**
  String get vetNotFound;

  /// No description provided for @addVetSearchHint.
  ///
  /// In fr, this message translates to:
  /// **'Nous recherchons ce compte vétérinaire dans petsFollow. S’il existe, une demande de liaison est envoyée au cabinet.'**
  String get addVetSearchHint;

  /// No description provided for @visitRequested.
  ///
  /// In fr, this message translates to:
  /// **'Demande de visite envoyée'**
  String get visitRequested;

  /// No description provided for @primaryVetSet.
  ///
  /// In fr, this message translates to:
  /// **'Vétérinaire principal mis à jour'**
  String get primaryVetSet;

  /// No description provided for @visitStatusRequested.
  ///
  /// In fr, this message translates to:
  /// **'Demandée'**
  String get visitStatusRequested;

  /// No description provided for @visitStatusConfirmed.
  ///
  /// In fr, this message translates to:
  /// **'Confirmée'**
  String get visitStatusConfirmed;

  /// No description provided for @visitStatusDone.
  ///
  /// In fr, this message translates to:
  /// **'Terminée'**
  String get visitStatusDone;

  /// No description provided for @visitStatusCancelled.
  ///
  /// In fr, this message translates to:
  /// **'Annulée'**
  String get visitStatusCancelled;

  /// No description provided for @visitStatusReschedulePending.
  ///
  /// In fr, this message translates to:
  /// **'Déplacement en attente'**
  String get visitStatusReschedulePending;

  /// No description provided for @horseHealthTitle.
  ///
  /// In fr, this message translates to:
  /// **'Santé équine'**
  String get horseHealthTitle;

  /// No description provided for @horseContactsTitle.
  ///
  /// In fr, this message translates to:
  /// **'Contacts (maréchal, dentiste…)'**
  String get horseContactsTitle;

  /// No description provided for @horseCompetitionsTitle.
  ///
  /// In fr, this message translates to:
  /// **'Compétitions'**
  String get horseCompetitionsTitle;

  /// No description provided for @horseContactsSoon.
  ///
  /// In fr, this message translates to:
  /// **'Activez le Pack Cheval pour gérer vos contacts professionnels.'**
  String get horseContactsSoon;

  /// No description provided for @horseCompetitionsSoon.
  ///
  /// In fr, this message translates to:
  /// **'Activez le Pack Cheval pour le calendrier de compétitions.'**
  String get horseCompetitionsSoon;

  /// No description provided for @horsePackUpsell.
  ///
  /// In fr, this message translates to:
  /// **'Pack Cheval — maréchal, coproscopie, contacts et compétitions'**
  String get horsePackUpsell;

  /// No description provided for @careTypeFarrier.
  ///
  /// In fr, this message translates to:
  /// **'Maréchal-ferrant'**
  String get careTypeFarrier;

  /// No description provided for @careTypeFecalEgg.
  ///
  /// In fr, this message translates to:
  /// **'Coproscopie'**
  String get careTypeFecalEgg;

  /// No description provided for @careTypeVaccination.
  ///
  /// In fr, this message translates to:
  /// **'Vaccination'**
  String get careTypeVaccination;

  /// No description provided for @careTypeDeworming.
  ///
  /// In fr, this message translates to:
  /// **'Vermifuge'**
  String get careTypeDeworming;

  /// No description provided for @careTypeVetCheck.
  ///
  /// In fr, this message translates to:
  /// **'Contrôle vétérinaire'**
  String get careTypeVetCheck;

  /// No description provided for @careTypeDental.
  ///
  /// In fr, this message translates to:
  /// **'Soins dentaires'**
  String get careTypeDental;

  /// No description provided for @careTypeCustom.
  ///
  /// In fr, this message translates to:
  /// **'Rappel personnalisé'**
  String get careTypeCustom;

  /// No description provided for @homeAddFirstVetTitle.
  ///
  /// In fr, this message translates to:
  /// **'Ajoutez votre vétérinaire'**
  String get homeAddFirstVetTitle;

  /// No description provided for @homeAddFirstVetBody.
  ///
  /// In fr, this message translates to:
  /// **'Liez le cabinet qui suit votre animal pour partager les relevés et échanger.'**
  String get homeAddFirstVetBody;

  /// No description provided for @homeAddFirstVetCta.
  ///
  /// In fr, this message translates to:
  /// **'Ajouter un vétérinaire'**
  String get homeAddFirstVetCta;

  /// No description provided for @photoFrameHint.
  ///
  /// In fr, this message translates to:
  /// **'Cadrez le museau au centre — aperçu fiche animal'**
  String get photoFrameHint;

  /// No description provided for @takePhoto.
  ///
  /// In fr, this message translates to:
  /// **'Prendre une photo'**
  String get takePhoto;

  /// No description provided for @chooseFromGallery.
  ///
  /// In fr, this message translates to:
  /// **'Choisir dans la galerie'**
  String get chooseFromGallery;

  /// No description provided for @attachMedia.
  ///
  /// In fr, this message translates to:
  /// **'Joindre une photo ou une vidéo'**
  String get attachMedia;

  /// No description provided for @attachPhoto.
  ///
  /// In fr, this message translates to:
  /// **'Photo'**
  String get attachPhoto;

  /// No description provided for @attachVideo.
  ///
  /// In fr, this message translates to:
  /// **'Vidéo'**
  String get attachVideo;

  /// No description provided for @openMedia.
  ///
  /// In fr, this message translates to:
  /// **'Ouvrir'**
  String get openMedia;

  /// No description provided for @mediaVideoLabel.
  ///
  /// In fr, this message translates to:
  /// **'Vidéo'**
  String get mediaVideoLabel;
}

class _AppLocalizationsDelegate
    extends LocalizationsDelegate<AppLocalizations> {
  const _AppLocalizationsDelegate();

  @override
  Future<AppLocalizations> load(Locale locale) {
    return SynchronousFuture<AppLocalizations>(lookupAppLocalizations(locale));
  }

  @override
  bool isSupported(Locale locale) =>
      <String>['en', 'es', 'fr', 'nl'].contains(locale.languageCode);

  @override
  bool shouldReload(_AppLocalizationsDelegate old) => false;
}

AppLocalizations lookupAppLocalizations(Locale locale) {
  // Lookup logic when only language code is specified.
  switch (locale.languageCode) {
    case 'en':
      return AppLocalizationsEn();
    case 'es':
      return AppLocalizationsEs();
    case 'fr':
      return AppLocalizationsFr();
    case 'nl':
      return AppLocalizationsNl();
  }

  throw FlutterError(
      'AppLocalizations.delegate failed to load unsupported locale "$locale". This is likely '
      'an issue with the localizations generation tool. Please file an issue '
      'on GitHub with a reproducible sample app and the gen-l10n configuration '
      'that was used.');
}
