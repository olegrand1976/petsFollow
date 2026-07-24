import 'dart:async';

import 'package:flutter/foundation.dart';
import 'package:flutter/widgets.dart';
import 'package:flutter_localizations/flutter_localizations.dart';
import 'package:intl/intl.dart' as intl;

import 'app_localizations_en.dart';
import 'app_localizations_es.dart';
import 'app_localizations_et.dart';
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
    Locale('et'),
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

  /// No description provided for @twoFaTitle.
  ///
  /// In fr, this message translates to:
  /// **'Vérification 2FA'**
  String get twoFaTitle;

  /// No description provided for @twoFaSubtitle.
  ///
  /// In fr, this message translates to:
  /// **'Saisissez le code à 6 chiffres de votre application d\'authentification.'**
  String get twoFaSubtitle;

  /// No description provided for @twoFaCode.
  ///
  /// In fr, this message translates to:
  /// **'Code authenticator'**
  String get twoFaCode;

  /// No description provided for @twoFaSubmit.
  ///
  /// In fr, this message translates to:
  /// **'Valider'**
  String get twoFaSubmit;

  /// No description provided for @twoFaBack.
  ///
  /// In fr, this message translates to:
  /// **'Retour à la connexion'**
  String get twoFaBack;

  /// No description provided for @twoFaInvalid.
  ///
  /// In fr, this message translates to:
  /// **'Code 2FA invalide ou expiré'**
  String get twoFaInvalid;

  /// No description provided for @forgotPassword.
  ///
  /// In fr, this message translates to:
  /// **'Mot de passe oublié ?'**
  String get forgotPassword;

  /// No description provided for @forgotPasswordTitle.
  ///
  /// In fr, this message translates to:
  /// **'Mot de passe oublié'**
  String get forgotPasswordTitle;

  /// No description provided for @forgotPasswordSubtitle.
  ///
  /// In fr, this message translates to:
  /// **'Indiquez l\'email de votre compte. Si un compte existe, un lien de réinitialisation sera envoyé.'**
  String get forgotPasswordSubtitle;

  /// No description provided for @forgotPasswordSubmit.
  ///
  /// In fr, this message translates to:
  /// **'Envoyer le lien'**
  String get forgotPasswordSubmit;

  /// No description provided for @forgotPasswordBack.
  ///
  /// In fr, this message translates to:
  /// **'Retour à la connexion'**
  String get forgotPasswordBack;

  /// No description provided for @forgotPasswordFailed.
  ///
  /// In fr, this message translates to:
  /// **'Envoi impossible'**
  String get forgotPasswordFailed;

  /// No description provided for @forgotPasswordSentTitle.
  ///
  /// In fr, this message translates to:
  /// **'Email envoyé'**
  String get forgotPasswordSentTitle;

  /// No description provided for @forgotPasswordSent.
  ///
  /// In fr, this message translates to:
  /// **'Si un compte existe pour {email}, un lien a été envoyé. Ouvrez-le dans votre navigateur pour choisir un nouveau mot de passe.'**
  String forgotPasswordSent(String email);

  /// No description provided for @emailRequired.
  ///
  /// In fr, this message translates to:
  /// **'Saisissez une adresse email valide'**
  String get emailRequired;

  /// No description provided for @resetPasswordTitle.
  ///
  /// In fr, this message translates to:
  /// **'Nouveau mot de passe'**
  String get resetPasswordTitle;

  /// No description provided for @resetPasswordSubtitle.
  ///
  /// In fr, this message translates to:
  /// **'Minimum 8 caractères.'**
  String get resetPasswordSubtitle;

  /// No description provided for @resetPasswordToken.
  ///
  /// In fr, this message translates to:
  /// **'Jeton de réinitialisation'**
  String get resetPasswordToken;

  /// No description provided for @resetPasswordSubmit.
  ///
  /// In fr, this message translates to:
  /// **'Enregistrer'**
  String get resetPasswordSubmit;

  /// No description provided for @resetPasswordBackToLogin.
  ///
  /// In fr, this message translates to:
  /// **'Aller à la connexion'**
  String get resetPasswordBackToLogin;

  /// No description provided for @resetPasswordInvalidLink.
  ///
  /// In fr, this message translates to:
  /// **'Lien de réinitialisation invalide'**
  String get resetPasswordInvalidLink;

  /// No description provided for @resetPasswordFailed.
  ///
  /// In fr, this message translates to:
  /// **'Réinitialisation impossible'**
  String get resetPasswordFailed;

  /// No description provided for @resetPasswordDoneTitle.
  ///
  /// In fr, this message translates to:
  /// **'Mot de passe mis à jour'**
  String get resetPasswordDoneTitle;

  /// No description provided for @resetPasswordDoneSubtitle.
  ///
  /// In fr, this message translates to:
  /// **'Vous pouvez maintenant vous connecter.'**
  String get resetPasswordDoneSubtitle;

  /// No description provided for @fullName.
  ///
  /// In fr, this message translates to:
  /// **'Nom complet'**
  String get fullName;

  /// No description provided for @registerCta.
  ///
  /// In fr, this message translates to:
  /// **'Créer un compte'**
  String get registerCta;

  /// No description provided for @registerTitle.
  ///
  /// In fr, this message translates to:
  /// **'Créer un compte'**
  String get registerTitle;

  /// No description provided for @registerSubtitle.
  ///
  /// In fr, this message translates to:
  /// **'Suivez la santé de votre animal. Confirmez ensuite votre email.'**
  String get registerSubtitle;

  /// No description provided for @registerSubmit.
  ///
  /// In fr, this message translates to:
  /// **'S\'inscrire'**
  String get registerSubmit;

  /// No description provided for @registerSuccess.
  ///
  /// In fr, this message translates to:
  /// **'Compte créé. Vérifiez votre email pour confirmer, puis connectez-vous.'**
  String get registerSuccess;

  /// No description provided for @registerFailed.
  ///
  /// In fr, this message translates to:
  /// **'Inscription impossible'**
  String get registerFailed;

  /// No description provided for @registerEmailExists.
  ///
  /// In fr, this message translates to:
  /// **'Cet email est déjà utilisé'**
  String get registerEmailExists;

  /// No description provided for @registerBackToLogin.
  ///
  /// In fr, this message translates to:
  /// **'Retour à la connexion'**
  String get registerBackToLogin;

  /// No description provided for @vetUseProWeb.
  ///
  /// In fr, this message translates to:
  /// **'Le compte vétérinaire complet s\'utilise sur le site Pro web.'**
  String get vetUseProWeb;

  /// No description provided for @unsupportedRoleApp.
  ///
  /// In fr, this message translates to:
  /// **'Ce compte n\'est pas utilisable dans l\'app pets. Utilisez le site Pro web.'**
  String get unsupportedRoleApp;

  /// No description provided for @proLightTitle.
  ///
  /// In fr, this message translates to:
  /// **'Pro terrain'**
  String get proLightTitle;

  /// No description provided for @proLightAgenda.
  ///
  /// In fr, this message translates to:
  /// **'Agenda'**
  String get proLightAgenda;

  /// No description provided for @proLightClients.
  ///
  /// In fr, this message translates to:
  /// **'Clients'**
  String get proLightClients;

  /// No description provided for @proLightPets.
  ///
  /// In fr, this message translates to:
  /// **'Animaux'**
  String get proLightPets;

  /// No description provided for @proLightLoadError.
  ///
  /// In fr, this message translates to:
  /// **'Chargement impossible'**
  String get proLightLoadError;

  /// No description provided for @proLightNoVisits.
  ///
  /// In fr, this message translates to:
  /// **'Aucun rendez-vous'**
  String get proLightNoVisits;

  /// No description provided for @proLightNoClients.
  ///
  /// In fr, this message translates to:
  /// **'Aucun client partagé'**
  String get proLightNoClients;

  /// No description provided for @proLightNoPets.
  ///
  /// In fr, this message translates to:
  /// **'Aucun animal partagé'**
  String get proLightNoPets;

  /// No description provided for @proLightAddress.
  ///
  /// In fr, this message translates to:
  /// **'Adresse'**
  String get proLightAddress;

  /// No description provided for @proLightOpenMaps.
  ///
  /// In fr, this message translates to:
  /// **'Maps'**
  String get proLightOpenMaps;

  /// No description provided for @proLightReportTitle.
  ///
  /// In fr, this message translates to:
  /// **'Compte rendu'**
  String get proLightReportTitle;

  /// No description provided for @proLightReportHint.
  ///
  /// In fr, this message translates to:
  /// **'Notes de visite…'**
  String get proLightReportHint;

  /// No description provided for @proLightImproveAi.
  ///
  /// In fr, this message translates to:
  /// **'Améliorer (IA)'**
  String get proLightImproveAi;

  /// No description provided for @proLightFinalizeReport.
  ///
  /// In fr, this message translates to:
  /// **'Finaliser'**
  String get proLightFinalizeReport;

  /// No description provided for @proLightReportFinal.
  ///
  /// In fr, this message translates to:
  /// **'Finalisé'**
  String get proLightReportFinal;

  /// No description provided for @proLightSettings.
  ///
  /// In fr, this message translates to:
  /// **'Réglages'**
  String get proLightSettings;

  /// No description provided for @proLightSpecialty.
  ///
  /// In fr, this message translates to:
  /// **'Spécialité'**
  String get proLightSpecialty;

  /// No description provided for @proLightDocuments.
  ///
  /// In fr, this message translates to:
  /// **'Documents'**
  String get proLightDocuments;

  /// No description provided for @proLightNoDocuments.
  ///
  /// In fr, this message translates to:
  /// **'Aucun document'**
  String get proLightNoDocuments;

  /// No description provided for @proLightTimeline.
  ///
  /// In fr, this message translates to:
  /// **'Timeline'**
  String get proLightTimeline;

  /// No description provided for @proLightNoTimeline.
  ///
  /// In fr, this message translates to:
  /// **'Aucun événement'**
  String get proLightNoTimeline;

  /// No description provided for @proLightReminders.
  ///
  /// In fr, this message translates to:
  /// **'Rappels'**
  String get proLightReminders;

  /// No description provided for @proLightNoReminders.
  ///
  /// In fr, this message translates to:
  /// **'Aucun rappel'**
  String get proLightNoReminders;

  /// No description provided for @proLightLitterTag.
  ///
  /// In fr, this message translates to:
  /// **'Tag / portée'**
  String get proLightLitterTag;

  /// No description provided for @proLightActionFailed.
  ///
  /// In fr, this message translates to:
  /// **'Action impossible'**
  String get proLightActionFailed;

  /// No description provided for @proLightReadOnly.
  ///
  /// In fr, this message translates to:
  /// **'Accès lecture seule'**
  String get proLightReadOnly;

  /// No description provided for @petAccessSharedRead.
  ///
  /// In fr, this message translates to:
  /// **'Partagé · lecture'**
  String get petAccessSharedRead;

  /// No description provided for @petAccessSharedNotes.
  ///
  /// In fr, this message translates to:
  /// **'Partagé · notes'**
  String get petAccessSharedNotes;

  /// No description provided for @petAccessSharedFull.
  ///
  /// In fr, this message translates to:
  /// **'Partagé · complet'**
  String get petAccessSharedFull;

  /// No description provided for @proLightUseGps.
  ///
  /// In fr, this message translates to:
  /// **'GPS'**
  String get proLightUseGps;

  /// No description provided for @proLightTranscribeAudio.
  ///
  /// In fr, this message translates to:
  /// **'Fichier audio'**
  String get proLightTranscribeAudio;

  /// No description provided for @proLightDictationStart.
  ///
  /// In fr, this message translates to:
  /// **'Dicter'**
  String get proLightDictationStart;

  /// No description provided for @proLightDictationStop.
  ///
  /// In fr, this message translates to:
  /// **'Arrêter & transcrire'**
  String get proLightDictationStop;

  /// No description provided for @proLightAudioConsentTitle.
  ///
  /// In fr, this message translates to:
  /// **'Consentement audio'**
  String get proLightAudioConsentTitle;

  /// No description provided for @proLightAudioConsentBody.
  ///
  /// In fr, this message translates to:
  /// **'L\'enregistrement sert uniquement à générer le compte rendu. Il est supprimé à la finalisation du CR.'**
  String get proLightAudioConsentBody;

  /// No description provided for @proLightAudioConsentAccept.
  ///
  /// In fr, this message translates to:
  /// **'J\'accepte'**
  String get proLightAudioConsentAccept;

  /// No description provided for @proLightSpecialtyFarrier.
  ///
  /// In fr, this message translates to:
  /// **'Maréchal-ferrant'**
  String get proLightSpecialtyFarrier;

  /// No description provided for @proLightSpecialtyPhysio.
  ///
  /// In fr, this message translates to:
  /// **'Physio / ostéo'**
  String get proLightSpecialtyPhysio;

  /// No description provided for @proLightSpecialtyBehaviorist.
  ///
  /// In fr, this message translates to:
  /// **'Comportementaliste'**
  String get proLightSpecialtyBehaviorist;

  /// No description provided for @proLightSpecialtyVetLight.
  ///
  /// In fr, this message translates to:
  /// **'Véto light'**
  String get proLightSpecialtyVetLight;

  /// No description provided for @proLightReportHintFarrier.
  ///
  /// In fr, this message translates to:
  /// **'CR ferrage : pieds, fer, observations…'**
  String get proLightReportHintFarrier;

  /// No description provided for @proLightEmptyFarrier.
  ///
  /// In fr, this message translates to:
  /// **'Aucun cheval / intervention partagée'**
  String get proLightEmptyFarrier;

  /// No description provided for @proLightMicDenied.
  ///
  /// In fr, this message translates to:
  /// **'Microphone refusé — autorisez l\'accès dans les réglages'**
  String get proLightMicDenied;

  /// No description provided for @proLightGpsDenied.
  ///
  /// In fr, this message translates to:
  /// **'Position indisponible'**
  String get proLightGpsDenied;

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

  /// No description provided for @choosePetForMeasurement.
  ///
  /// In fr, this message translates to:
  /// **'Choisir un animal'**
  String get choosePetForMeasurement;

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

  /// No description provided for @legalOpenOnline.
  ///
  /// In fr, this message translates to:
  /// **'Voir la version en ligne'**
  String get legalOpenOnline;

  /// No description provided for @legalTermsBody.
  ///
  /// In fr, this message translates to:
  /// **'Conditions générales d\'utilisation — petsFollow\n\nL\'application petsFollow permet aux propriétaires d\'animaux de mesurer la fréquence cardiaque, de consulter l\'historique et de communiquer avec leur vétérinaire.\n\nLes services sont fournis dans le cadre de l\'abonnement choisi (paiement via Stripe). L\'utilisateur s\'engage à utiliser l\'application conformément à sa destination.\n\nVersion complète : https://petsfollow.ll-it-sc.be/legal/terms\n\nDate d\'actualisation : juillet 2026'**
  String get legalTermsBody;

  /// No description provided for @legalPrivacyBody.
  ///
  /// In fr, this message translates to:
  /// **'Politique de confidentialité — petsFollow\n\nDonnées collectées : identité (prénom, email), données animal (nom, espèce, race, photos), relevés de fréquence cardiaque (données de santé animale), messages et médias échangés avec le cabinet, jetons de notification (FCM), données de paiement traitées par Stripe.\n\nFinalités : gestion du compte, suivi cardiaque, messagerie vétérinaire, notifications, facturation.\n\nSous-traitants / partenaires : Google (Sign-In, Firebase Cloud Messaging), Stripe (paiements), hébergement cloud (GCP).\n\nConservation : jusqu\'à suppression du compte ou 3 ans d\'inactivité.\n\nDroits RGPD (accès, rectification, suppression) : Profil → Supprimer le compte, ou contact support@ll-it-sc.be.\n\nVersion complète : https://petsfollow.ll-it-sc.be/legal/privacy\n\nDate d\'actualisation : juillet 2026'**
  String get legalPrivacyBody;

  /// No description provided for @legalNoticeBody.
  ///
  /// In fr, this message translates to:
  /// **'Mentions légales — petsFollow\n\nÉditeur : LL-IT-SC / petsFollow\nContact : support@ll-it-sc.be\n\nHébergement : Google Cloud Platform (conformité RGPD).\n\nDirecteur de publication : petsFollow.\n\nVersion complète : https://petsfollow.ll-it-sc.be/legal/mentions\n\nDate d\'actualisation : juillet 2026'**
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

  /// No description provided for @languageEt.
  ///
  /// In fr, this message translates to:
  /// **'Eesti'**
  String get languageEt;

  /// No description provided for @planMonthlyLabel.
  ///
  /// In fr, this message translates to:
  /// **'3,50 € / mois'**
  String get planMonthlyLabel;

  /// No description provided for @planAnnualLabel.
  ///
  /// In fr, this message translates to:
  /// **'35 € / an'**
  String get planAnnualLabel;

  /// No description provided for @planTriennialLabel.
  ///
  /// In fr, this message translates to:
  /// **'95 € / 3 ans'**
  String get planTriennialLabel;

  /// No description provided for @planQuinquennialLabel.
  ///
  /// In fr, this message translates to:
  /// **'145 € / 5 ans'**
  String get planQuinquennialLabel;

  /// No description provided for @pushNewMessage.
  ///
  /// In fr, this message translates to:
  /// **'Nouveau message'**
  String get pushNewMessage;

  /// No description provided for @pushVisitConfirmed.
  ///
  /// In fr, this message translates to:
  /// **'Rendez-vous confirmé'**
  String get pushVisitConfirmed;

  /// No description provided for @pushVisitProposed.
  ///
  /// In fr, this message translates to:
  /// **'Proposition de rendez-vous'**
  String get pushVisitProposed;

  /// No description provided for @pushVisitReschedule.
  ///
  /// In fr, this message translates to:
  /// **'Déplacement de rendez-vous'**
  String get pushVisitReschedule;

  /// No description provided for @notifChannelMessages.
  ///
  /// In fr, this message translates to:
  /// **'Messages'**
  String get notifChannelMessages;

  /// No description provided for @notifChannelVisits.
  ///
  /// In fr, this message translates to:
  /// **'Visites'**
  String get notifChannelVisits;

  /// No description provided for @notifChannelCare.
  ///
  /// In fr, this message translates to:
  /// **'Soins'**
  String get notifChannelCare;

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

  /// No description provided for @errorNetwork.
  ///
  /// In fr, this message translates to:
  /// **'Connexion impossible. Vérifiez votre réseau et réessayez.'**
  String get errorNetwork;

  /// No description provided for @retryAction.
  ///
  /// In fr, this message translates to:
  /// **'Réessayer'**
  String get retryAction;

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

  /// No description provided for @planMonthlySub.
  ///
  /// In fr, this message translates to:
  /// **'3,50 € / mois, renouvelé automatiquement'**
  String get planMonthlySub;

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

  /// No description provided for @heartRateCommentLabel.
  ///
  /// In fr, this message translates to:
  /// **'Commentaire (optionnel)'**
  String get heartRateCommentLabel;

  /// No description provided for @heartRateCommentHint.
  ///
  /// In fr, this message translates to:
  /// **'Ex. agité, au repos, après effort…'**
  String get heartRateCommentHint;

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

  /// No description provided for @careReferenceModeDone.
  ///
  /// In fr, this message translates to:
  /// **'Déjà effectué'**
  String get careReferenceModeDone;

  /// No description provided for @careReferenceModeFirst.
  ///
  /// In fr, this message translates to:
  /// **'Première fois'**
  String get careReferenceModeFirst;

  /// No description provided for @careLastDateLabel.
  ///
  /// In fr, this message translates to:
  /// **'Dernière date'**
  String get careLastDateLabel;

  /// No description provided for @careLastDateDone.
  ///
  /// In fr, this message translates to:
  /// **'Date du dernier soin'**
  String get careLastDateDone;

  /// No description provided for @careLastDateFirst.
  ///
  /// In fr, this message translates to:
  /// **'Date de départ du cycle'**
  String get careLastDateFirst;

  /// No description provided for @careRecurrenceLabel.
  ///
  /// In fr, this message translates to:
  /// **'Récurrence'**
  String get careRecurrenceLabel;

  /// No description provided for @careRecurrenceNone.
  ///
  /// In fr, this message translates to:
  /// **'Aucune (échéance unique)'**
  String get careRecurrenceNone;

  /// No description provided for @careRecurrenceDays.
  ///
  /// In fr, this message translates to:
  /// **'Tous les {days} jours'**
  String careRecurrenceDays(int days);

  /// No description provided for @careDueDateLabel.
  ///
  /// In fr, this message translates to:
  /// **'Échéance'**
  String get careDueDateLabel;

  /// No description provided for @careDueDateComputed.
  ///
  /// In fr, this message translates to:
  /// **'Échéance calculée'**
  String get careDueDateComputed;

  /// No description provided for @careTooltipDoneWithRecurrence.
  ///
  /// In fr, this message translates to:
  /// **'Soin déjà fait : l’échéance = date du dernier soin + récurrence.'**
  String get careTooltipDoneWithRecurrence;

  /// No description provided for @careTooltipFirstWithRecurrence.
  ///
  /// In fr, this message translates to:
  /// **'Première planification : indiquez la date de départ du cycle. L’échéance = cette date + récurrence.'**
  String get careTooltipFirstWithRecurrence;

  /// No description provided for @careTooltipNoRecurrence.
  ///
  /// In fr, this message translates to:
  /// **'Sans récurrence : la date saisie est l’échéance unique.'**
  String get careTooltipNoRecurrence;

  /// No description provided for @careTooltipDueExplained.
  ///
  /// In fr, this message translates to:
  /// **'Échéance = dernière date + récurrence (si définie).'**
  String get careTooltipDueExplained;

  /// No description provided for @carePickDate.
  ///
  /// In fr, this message translates to:
  /// **'Choisir une date'**
  String get carePickDate;

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
      <String>['en', 'es', 'et', 'fr', 'nl'].contains(locale.languageCode);

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
    case 'et':
      return AppLocalizationsEt();
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
