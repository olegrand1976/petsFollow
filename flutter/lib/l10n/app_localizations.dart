import 'dart:async';

import 'package:flutter/foundation.dart';
import 'package:flutter/widgets.dart';
import 'package:flutter_localizations/flutter_localizations.dart';
import 'package:intl/intl.dart' as intl;

import 'app_localizations_en.dart';
import 'app_localizations_es.dart';
import 'app_localizations_et.dart';
import 'app_localizations_fr.dart';
import 'app_localizations_it.dart';
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
    Locale('it'),
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

  /// No description provided for @emailNotVerified.
  ///
  /// In fr, this message translates to:
  /// **'Confirmez d\'abord votre email (lien reçu à l\'inscription), puis reconnectez-vous.'**
  String get emailNotVerified;

  /// No description provided for @resendConfirmation.
  ///
  /// In fr, this message translates to:
  /// **'Renvoyer l\'email de confirmation'**
  String get resendConfirmation;

  /// No description provided for @resendConfirmationSent.
  ///
  /// In fr, this message translates to:
  /// **'Si le compte existe et n\'est pas encore confirmé, un nouvel email a été envoyé.'**
  String get resendConfirmationSent;

  /// No description provided for @resendConfirmationFailed.
  ///
  /// In fr, this message translates to:
  /// **'Envoi impossible. Réessayez dans un instant.'**
  String get resendConfirmationFailed;

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

  /// No description provided for @loginWithApple.
  ///
  /// In fr, this message translates to:
  /// **'Continuer avec Apple'**
  String get loginWithApple;

  /// No description provided for @appleComingSoon.
  ///
  /// In fr, this message translates to:
  /// **'La connexion avec Apple arrive bientôt.'**
  String get appleComingSoon;

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
  /// **'S\'inscrire'**
  String get registerCta;

  /// No description provided for @registerTitle.
  ///
  /// In fr, this message translates to:
  /// **'S\'inscrire'**
  String get registerTitle;

  /// No description provided for @registerSubtitle.
  ///
  /// In fr, this message translates to:
  /// **'Créez votre compte pour suivre la santé de votre animal. Un email de validation vous sera envoyé.'**
  String get registerSubtitle;

  /// No description provided for @registerSubmit.
  ///
  /// In fr, this message translates to:
  /// **'S\'inscrire'**
  String get registerSubmit;

  /// No description provided for @registerSuccess.
  ///
  /// In fr, this message translates to:
  /// **'Compte créé. Ouvrez le lien dans l\'email de validation, puis revenez vous connecter dans l\'app.'**
  String get registerSuccess;

  /// No description provided for @registerInviteNotApplied.
  ///
  /// In fr, this message translates to:
  /// **'Compte créé, mais le code d\'invitation n\'a pas pu être appliqué. Vous pourrez le ressaisir après connexion.'**
  String get registerInviteNotApplied;

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

  /// No description provided for @confirmEmailTitle.
  ///
  /// In fr, this message translates to:
  /// **'Confirmation de l\'email'**
  String get confirmEmailTitle;

  /// No description provided for @confirmEmailLoading.
  ///
  /// In fr, this message translates to:
  /// **'Confirmation en cours…'**
  String get confirmEmailLoading;

  /// No description provided for @confirmEmailDoneTitle.
  ///
  /// In fr, this message translates to:
  /// **'Email confirmé'**
  String get confirmEmailDoneTitle;

  /// No description provided for @confirmEmailDoneSubtitle.
  ///
  /// In fr, this message translates to:
  /// **'Votre compte est activé. Vous pouvez vous connecter.'**
  String get confirmEmailDoneSubtitle;

  /// No description provided for @confirmEmailFailedTitle.
  ///
  /// In fr, this message translates to:
  /// **'Confirmation impossible'**
  String get confirmEmailFailedTitle;

  /// No description provided for @confirmEmailFailed.
  ///
  /// In fr, this message translates to:
  /// **'Impossible de confirmer cet email.'**
  String get confirmEmailFailed;

  /// No description provided for @confirmEmailInvalidLink.
  ///
  /// In fr, this message translates to:
  /// **'Lien de confirmation invalide ou déjà utilisé.'**
  String get confirmEmailInvalidLink;

  /// No description provided for @confirmEmailBackToLogin.
  ///
  /// In fr, this message translates to:
  /// **'Retour à la connexion'**
  String get confirmEmailBackToLogin;

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

  /// No description provided for @proLightTourToday.
  ///
  /// In fr, this message translates to:
  /// **'Aujourd\'hui'**
  String get proLightTourToday;

  /// No description provided for @proLightTourWeek.
  ///
  /// In fr, this message translates to:
  /// **'7 jours'**
  String get proLightTourWeek;

  /// No description provided for @proLightTourAll.
  ///
  /// In fr, this message translates to:
  /// **'Tout'**
  String get proLightTourAll;

  /// No description provided for @proLightNoTourToday.
  ///
  /// In fr, this message translates to:
  /// **'Aucun rendez-vous aujourd\'hui'**
  String get proLightNoTourToday;

  /// No description provided for @proLightNoTourWeek.
  ///
  /// In fr, this message translates to:
  /// **'Aucun rendez-vous sur 7 jours'**
  String get proLightNoTourWeek;

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

  /// No description provided for @proLightReportHistoryTitle.
  ///
  /// In fr, this message translates to:
  /// **'Historique'**
  String get proLightReportHistoryTitle;

  /// No description provided for @proLightReportHistoryTranscript.
  ///
  /// In fr, this message translates to:
  /// **'Original (transcription)'**
  String get proLightReportHistoryTranscript;

  /// No description provided for @proLightReportHistoryImproved.
  ///
  /// In fr, this message translates to:
  /// **'Version IA'**
  String get proLightReportHistoryImproved;

  /// No description provided for @proLightReportHistorySaved.
  ///
  /// In fr, this message translates to:
  /// **'Version enregistrée'**
  String get proLightReportHistorySaved;

  /// No description provided for @proLightReportHistoryEmpty.
  ///
  /// In fr, this message translates to:
  /// **'Aucune version disponible'**
  String get proLightReportHistoryEmpty;

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
  /// **'L\'enregistrement sert uniquement à générer le compte rendu. Confirmez l\'accord oral du client. L\'audio est supprimé à la finalisation du CR.'**
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

  /// No description provided for @proLightSpecialtyGroomer.
  ///
  /// In fr, this message translates to:
  /// **'Toiletteur'**
  String get proLightSpecialtyGroomer;

  /// No description provided for @proLightSpecialtyBreeder.
  ///
  /// In fr, this message translates to:
  /// **'Éleveur'**
  String get proLightSpecialtyBreeder;

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
  /// **'Ce compte Google est déjà un profil Pro — utilisez l\'application web.'**
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

  /// No description provided for @exportMyData.
  ///
  /// In fr, this message translates to:
  /// **'Exporter mes données'**
  String get exportMyData;

  /// No description provided for @registerConsentPrefix.
  ///
  /// In fr, this message translates to:
  /// **'J\'accepte les '**
  String get registerConsentPrefix;

  /// No description provided for @registerConsentMiddle.
  ///
  /// In fr, this message translates to:
  /// **' et la '**
  String get registerConsentMiddle;

  /// No description provided for @registerConsentRequired.
  ///
  /// In fr, this message translates to:
  /// **'Vous devez accepter les conditions et la politique de confidentialité.'**
  String get registerConsentRequired;

  /// No description provided for @registerInviteCode.
  ///
  /// In fr, this message translates to:
  /// **'Code d\'invitation (optionnel)'**
  String get registerInviteCode;

  /// No description provided for @registerInviteCodeHint.
  ///
  /// In fr, this message translates to:
  /// **'Saisissez le code du QR / lien commercial'**
  String get registerInviteCodeHint;

  /// No description provided for @nearbyCommercialTitle.
  ///
  /// In fr, this message translates to:
  /// **'Commercial près de chez vous'**
  String get nearbyCommercialTitle;

  /// No description provided for @nearbyCommercialHint.
  ///
  /// In fr, this message translates to:
  /// **'Sans code d\'invitation, choisissez un commercial proche (optionnel).'**
  String get nearbyCommercialHint;

  /// No description provided for @nearbyCommercialUseLocation.
  ///
  /// In fr, this message translates to:
  /// **'Utiliser ma position'**
  String get nearbyCommercialUseLocation;

  /// No description provided for @nearbyCommercialPostalCode.
  ///
  /// In fr, this message translates to:
  /// **'Code postal'**
  String get nearbyCommercialPostalCode;

  /// No description provided for @nearbyCommercialSearch.
  ///
  /// In fr, this message translates to:
  /// **'Rechercher'**
  String get nearbyCommercialSearch;

  /// No description provided for @nearbyCommercialEmpty.
  ///
  /// In fr, this message translates to:
  /// **'Aucun commercial trouvé à proximité.'**
  String get nearbyCommercialEmpty;

  /// No description provided for @nearbyCommercialSkip.
  ///
  /// In fr, this message translates to:
  /// **'Ne pas rattacher'**
  String get nearbyCommercialSkip;

  /// No description provided for @nearbyCommercialGeoDenied.
  ///
  /// In fr, this message translates to:
  /// **'Position refusée — saisissez un code postal.'**
  String get nearbyCommercialGeoDenied;

  /// No description provided for @nearbyCommercialDistance.
  ///
  /// In fr, this message translates to:
  /// **'{km} km'**
  String nearbyCommercialDistance(String km);

  /// No description provided for @pushPermissionTitle.
  ///
  /// In fr, this message translates to:
  /// **'Notifications'**
  String get pushPermissionTitle;

  /// No description provided for @pushPermissionBody.
  ///
  /// In fr, this message translates to:
  /// **'petsFollow souhaite vous envoyer des notifications : messages de votre vétérinaire, confirmations de rendez-vous et rappels de soins. Vous pouvez les désactiver à tout moment dans les réglages de l\'app ou du téléphone.'**
  String get pushPermissionBody;

  /// No description provided for @pushPermissionContinue.
  ///
  /// In fr, this message translates to:
  /// **'Continuer'**
  String get pushPermissionContinue;

  /// No description provided for @exportDataSaved.
  ///
  /// In fr, this message translates to:
  /// **'Export enregistré : {path}'**
  String exportDataSaved(String path);

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

  /// No description provided for @heartRateShort.
  ///
  /// In fr, this message translates to:
  /// **'Cœur'**
  String get heartRateShort;

  /// No description provided for @weightShort.
  ///
  /// In fr, this message translates to:
  /// **'Poids'**
  String get weightShort;

  /// No description provided for @recordWeightTitle.
  ///
  /// In fr, this message translates to:
  /// **'Enregistrer le poids'**
  String get recordWeightTitle;

  /// No description provided for @weightKgLabel.
  ///
  /// In fr, this message translates to:
  /// **'Poids (kg)'**
  String get weightKgLabel;

  /// No description provided for @weightCommentLabel.
  ///
  /// In fr, this message translates to:
  /// **'Commentaire (optionnel)'**
  String get weightCommentLabel;

  /// No description provided for @weightCommentHint.
  ///
  /// In fr, this message translates to:
  /// **'Ex. après balade, à jeun…'**
  String get weightCommentHint;

  /// No description provided for @weightSave.
  ///
  /// In fr, this message translates to:
  /// **'Enregistrer'**
  String get weightSave;

  /// No description provided for @weightSentToVet.
  ///
  /// In fr, this message translates to:
  /// **'Poids enregistré'**
  String get weightSentToVet;

  /// No description provided for @weightInvalid.
  ///
  /// In fr, this message translates to:
  /// **'Indiquez un poids valide (0,01–999,99 kg)'**
  String get weightInvalid;

  /// No description provided for @weightLastLabel.
  ///
  /// In fr, this message translates to:
  /// **'Dernier poids : {kg} kg'**
  String weightLastLabel(String kg);

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
  /// **'Conditions générales d\'utilisation — petsFollow\n\nL\'application petsFollow permet aux propriétaires d\'animaux le suivi prescrit (messagerie, rappels Care/Horse, relevés cardiaques), de consulter l\'historique et de communiquer avec leur vétérinaire.\n\nLes services sont fournis dans le cadre de l\'abonnement choisi (paiement via Stripe). L\'utilisateur s\'engage à utiliser l\'application conformément à sa destination.\n\nVersion complète : https://petsfollow.ll-it-sc.be/legal/terms\n\nDate d\'actualisation : juillet 2026'**
  String get legalTermsBody;

  /// No description provided for @legalPrivacyBody.
  ///
  /// In fr, this message translates to:
  /// **'Politique de confidentialité — petsFollow\n\nDonnées collectées : identité (prénom, email), données animal (nom, espèce, race, photos), relevés de fréquence cardiaque (données de santé animale), messages et médias échangés avec le cabinet, comptes rendus de visite (texte et enregistrements audio), coordonnées GPS des visites à domicile (professionnels de soin), jetons de notification (FCM), données de paiement traitées par Stripe.\n\nFinalités : gestion du compte, continuité de soins (dont relevés cardiaques), messagerie vétérinaire, comptes rendus de visite, notifications, facturation.\n\nTraitement IA : Google Gemini est utilisé pour améliorer les comptes rendus de visite (audio traité en temps réel, non conservé par Google).\n\nSous-traitants / partenaires : Google (Sign-In, Firebase Cloud Messaging, Gemini), Stripe (paiements), hébergement cloud (GCP).\n\nConservation : jusqu\'à suppression du compte ; comptes inactifs purgés après 3 ans ; audio des comptes rendus conservé le temps du dossier.\n\nDroits RGPD (accès, rectification, suppression, portabilité) : Profil → Exporter mes données / Supprimer le compte, ou contact support@ll-it-sc.be.\n\nVersion complète : https://petsfollow.ll-it-sc.be/legal/privacy\n\nDate d\'actualisation : juillet 2026'**
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

  /// No description provided for @languageIt.
  ///
  /// In fr, this message translates to:
  /// **'Italiano'**
  String get languageIt;

  /// No description provided for @appearance.
  ///
  /// In fr, this message translates to:
  /// **'Apparence'**
  String get appearance;

  /// No description provided for @themeLight.
  ///
  /// In fr, this message translates to:
  /// **'Clair'**
  String get themeLight;

  /// No description provided for @themeDark.
  ///
  /// In fr, this message translates to:
  /// **'Sombre'**
  String get themeDark;

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

  /// No description provided for @editPet.
  ///
  /// In fr, this message translates to:
  /// **'Modifier l\'animal'**
  String get editPet;

  /// No description provided for @petName.
  ///
  /// In fr, this message translates to:
  /// **'Nom'**
  String get petName;

  /// No description provided for @petNameRequired.
  ///
  /// In fr, this message translates to:
  /// **'Indiquez le nom de l’animal'**
  String get petNameRequired;

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

  /// No description provided for @petMicrochipOptional.
  ///
  /// In fr, this message translates to:
  /// **'N° de puce (optionnel)'**
  String get petMicrochipOptional;

  /// No description provided for @petHealthBookNumberOptional.
  ///
  /// In fr, this message translates to:
  /// **'N° de carnet (optionnel)'**
  String get petHealthBookNumberOptional;

  /// No description provided for @petHealthBookAddPages.
  ///
  /// In fr, this message translates to:
  /// **'Ajouter des photos du carnet'**
  String get petHealthBookAddPages;

  /// No description provided for @petHealthBookReplacePages.
  ///
  /// In fr, this message translates to:
  /// **'Remplacer le PDF (photos)'**
  String get petHealthBookReplacePages;

  /// No description provided for @petHealthBookPagesCount.
  ///
  /// In fr, this message translates to:
  /// **'{count} page(s) sélectionnée(s)'**
  String petHealthBookPagesCount(int count);

  /// No description provided for @petHealthBookPdfAttached.
  ///
  /// In fr, this message translates to:
  /// **'PDF du carnet joint'**
  String get petHealthBookPdfAttached;

  /// No description provided for @petHealthBookRemovePdf.
  ///
  /// In fr, this message translates to:
  /// **'Supprimer'**
  String get petHealthBookRemovePdf;

  /// No description provided for @petHealthBookOpenPdf.
  ///
  /// In fr, this message translates to:
  /// **'Ouvrir le carnet (PDF)'**
  String get petHealthBookOpenPdf;

  /// No description provided for @petMicrochipLabel.
  ///
  /// In fr, this message translates to:
  /// **'Puce : {number}'**
  String petMicrochipLabel(String number);

  /// No description provided for @petHealthBookNumberLabel.
  ///
  /// In fr, this message translates to:
  /// **'Carnet : {number}'**
  String petHealthBookNumberLabel(String number);

  /// No description provided for @errorHealthBookUploadFailed.
  ///
  /// In fr, this message translates to:
  /// **'Animal enregistré, mais le carnet n\'a pas pu être envoyé'**
  String get errorHealthBookUploadFailed;

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
  /// **'Enregistrer et payer'**
  String get continueToPayment;

  /// No description provided for @petFormSave.
  ///
  /// In fr, this message translates to:
  /// **'Enregistrer'**
  String get petFormSave;

  /// No description provided for @petSavedPendingPayment.
  ///
  /// In fr, this message translates to:
  /// **'Animal enregistré — activez-le pour accéder aux fonctionnalités'**
  String get petSavedPendingPayment;

  /// No description provided for @paymentFeaturesLocked.
  ///
  /// In fr, this message translates to:
  /// **'Paiement requis pour utiliser les fonctionnalités de cet animal'**
  String get paymentFeaturesLocked;

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
  /// **'Abonnement requis pour utiliser cette fonctionnalité'**
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

  /// No description provided for @heartRateNotSupported.
  ///
  /// In fr, this message translates to:
  /// **'Le relevé cardiaque n’est pas disponible pour cette espèce'**
  String get heartRateNotSupported;

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
  /// **'Alerte : hausse significative vs le relevé précédent'**
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
  /// **'Ajoutez votre premier animal pour commencer le suivi prescrit avec votre vétérinaire.'**
  String get emptyPetsBody;

  /// No description provided for @discoveryTitle.
  ///
  /// In fr, this message translates to:
  /// **'Découvrir petsFollow'**
  String get discoveryTitle;

  /// No description provided for @discoveryMission.
  ///
  /// In fr, this message translates to:
  /// **'Votre parcours petsFollow'**
  String get discoveryMission;

  /// No description provided for @discoveryDay0Title.
  ///
  /// In fr, this message translates to:
  /// **'Étape 1 — Bienvenue'**
  String get discoveryDay0Title;

  /// No description provided for @discoveryDay0Body.
  ///
  /// In fr, this message translates to:
  /// **'Créez le profil de votre animal et découvrez l\'app — messagerie, rappels et relevés (dont la fréquence cardiaque).'**
  String get discoveryDay0Body;

  /// No description provided for @discoveryDay2Title.
  ///
  /// In fr, this message translates to:
  /// **'Étape 2 — Première mesure'**
  String get discoveryDay2Title;

  /// No description provided for @discoveryDay2Body.
  ///
  /// In fr, this message translates to:
  /// **'Effectuez votre premier relevé cardiaque et familiarisez-vous avec la technique.'**
  String get discoveryDay2Body;

  /// No description provided for @discoveryDay4Title.
  ///
  /// In fr, this message translates to:
  /// **'Étape 3 — Routine'**
  String get discoveryDay4Title;

  /// No description provided for @discoveryDay4Body.
  ///
  /// In fr, this message translates to:
  /// **'Installez une routine de mesure quotidienne avec les rappels personnalisés.'**
  String get discoveryDay4Body;

  /// No description provided for @discoveryDay6Title.
  ///
  /// In fr, this message translates to:
  /// **'Étape 4 — Partage véto'**
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

  /// No description provided for @vetLinkRequired.
  ///
  /// In fr, this message translates to:
  /// **'Liez un vétérinaire pour activer le suivi avec votre cabinet'**
  String get vetLinkRequired;

  /// No description provided for @linkVetAfterSaveTitle.
  ///
  /// In fr, this message translates to:
  /// **'Lier un vétérinaire'**
  String get linkVetAfterSaveTitle;

  /// No description provided for @linkVetAfterSaveBody.
  ///
  /// In fr, this message translates to:
  /// **'Liez un cabinet pour activer messagerie, visites et rappels de soins.'**
  String get linkVetAfterSaveBody;

  /// No description provided for @linkVetHomeTitle.
  ///
  /// In fr, this message translates to:
  /// **'Lier un vétérinaire ?'**
  String get linkVetHomeTitle;

  /// No description provided for @linkVetHomeBody.
  ///
  /// In fr, this message translates to:
  /// **'Votre animal est enregistré. Souhaitez-vous lier un vétérinaire ? C’est optionnel — vous pourrez le faire plus tard.'**
  String get linkVetHomeBody;

  /// No description provided for @linkVetLater.
  ///
  /// In fr, this message translates to:
  /// **'Plus tard'**
  String get linkVetLater;

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
  /// **'La réservation en ligne n\'est pas disponible pour ce cabinet. Appelez le cabinet pour prendre rendez-vous.'**
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

  /// No description provided for @calendarSelectVet.
  ///
  /// In fr, this message translates to:
  /// **'Choisissez un vétérinaire :'**
  String get calendarSelectVet;

  /// No description provided for @calendarCallPractice.
  ///
  /// In fr, this message translates to:
  /// **'Appeler le cabinet'**
  String get calendarCallPractice;

  /// No description provided for @calendarNoPhone.
  ///
  /// In fr, this message translates to:
  /// **'Aucun numéro de téléphone n\'est renseigné pour ce cabinet. Contactez-le par un autre moyen.'**
  String get calendarNoPhone;

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

  /// No description provided for @petBirthDate.
  ///
  /// In fr, this message translates to:
  /// **'Date de naissance'**
  String get petBirthDate;

  /// No description provided for @petBirthDateInvalid.
  ///
  /// In fr, this message translates to:
  /// **'Date de naissance invalide (AAAA-MM-JJ)'**
  String get petBirthDateInvalid;

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
  /// **'E{day}'**
  String discoveryDayBadge(int day);

  /// No description provided for @timelineTypeHeartrate.
  ///
  /// In fr, this message translates to:
  /// **'Fréquence cardiaque'**
  String get timelineTypeHeartrate;

  /// No description provided for @timelineTypeWeight.
  ///
  /// In fr, this message translates to:
  /// **'Poids'**
  String get timelineTypeWeight;

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

  /// No description provided for @messageNoMessagesYet.
  ///
  /// In fr, this message translates to:
  /// **'Pas encore de messages'**
  String get messageNoMessagesYet;

  /// No description provided for @messageNewConversation.
  ///
  /// In fr, this message translates to:
  /// **'Nouvelle conversation'**
  String get messageNewConversation;

  /// No description provided for @messageComposeTitle.
  ///
  /// In fr, this message translates to:
  /// **'Nouvelle conversation'**
  String get messageComposeTitle;

  /// No description provided for @messageChoosePro.
  ///
  /// In fr, this message translates to:
  /// **'Professionnel de soins'**
  String get messageChoosePro;

  /// No description provided for @messageChoosePet.
  ///
  /// In fr, this message translates to:
  /// **'Animal concerné'**
  String get messageChoosePet;

  /// No description provided for @messageStartConversation.
  ///
  /// In fr, this message translates to:
  /// **'Démarrer'**
  String get messageStartConversation;

  /// No description provided for @messageLockedTitle.
  ///
  /// In fr, this message translates to:
  /// **'Messagerie indisponible'**
  String get messageLockedTitle;

  /// No description provided for @messageLockedBody.
  ///
  /// In fr, this message translates to:
  /// **'Liez un vétérinaire pour discuter avec un professionnel de soins.'**
  String get messageLockedBody;

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
  /// **'Recherchez par nom, email ou cabinet. S’il est déjà sur petsFollow, une demande de liaison est envoyée au cabinet.'**
  String get addVetSearchHint;

  /// No description provided for @addVetSearchLabel.
  ///
  /// In fr, this message translates to:
  /// **'Rechercher un vétérinaire'**
  String get addVetSearchLabel;

  /// No description provided for @addVetSearchFieldHint.
  ///
  /// In fr, this message translates to:
  /// **'Nom, email ou cabinet'**
  String get addVetSearchFieldHint;

  /// No description provided for @addVetNotListed.
  ///
  /// In fr, this message translates to:
  /// **'Mon véto n’est pas listé'**
  String get addVetNotListed;

  /// No description provided for @addVetSuggestTitle.
  ///
  /// In fr, this message translates to:
  /// **'Nouveau vétérinaire'**
  String get addVetSuggestTitle;

  /// No description provided for @addVetSuggestBody.
  ///
  /// In fr, this message translates to:
  /// **'Indiquez l’email et le téléphone du cabinet. Nous le contacterons pour qu’il rejoigne petsFollow.'**
  String get addVetSuggestBody;

  /// No description provided for @addVetSuggestEmail.
  ///
  /// In fr, this message translates to:
  /// **'Email du cabinet'**
  String get addVetSuggestEmail;

  /// No description provided for @addVetSuggestPhone.
  ///
  /// In fr, this message translates to:
  /// **'Téléphone'**
  String get addVetSuggestPhone;

  /// No description provided for @addVetSuggestNameOptional.
  ///
  /// In fr, this message translates to:
  /// **'Nom du véto (optionnel)'**
  String get addVetSuggestNameOptional;

  /// No description provided for @addVetSuggestCta.
  ///
  /// In fr, this message translates to:
  /// **'Envoyer la suggestion'**
  String get addVetSuggestCta;

  /// No description provided for @vetSuggestSent.
  ///
  /// In fr, this message translates to:
  /// **'Merci — nous contactons le cabinet. Vous serez notifié quand il rejoindra petsFollow.'**
  String get vetSuggestSent;

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

  /// No description provided for @takeVideo.
  ///
  /// In fr, this message translates to:
  /// **'Filmer une vidéo'**
  String get takeVideo;

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

  /// No description provided for @compressingMedia.
  ///
  /// In fr, this message translates to:
  /// **'Compression de la vidéo…'**
  String get compressingMedia;

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

  /// No description provided for @appInviteTitle.
  ///
  /// In fr, this message translates to:
  /// **'QR invitation app'**
  String get appInviteTitle;

  /// No description provided for @appInviteHint.
  ///
  /// In fr, this message translates to:
  /// **'Affichez ce QR ou partagez le lien. Un nouveau client qui s’inscrit via ce lien est rattaché automatiquement.'**
  String get appInviteHint;

  /// No description provided for @appInviteHintClient.
  ///
  /// In fr, this message translates to:
  /// **'Partagez ce QR avec un proche. Il sera lié à vous (parrainage) et pourra rejoindre votre cabinet s’il n’en a pas encore.'**
  String get appInviteHintClient;

  /// No description provided for @appInviteHintCommercial.
  ///
  /// In fr, this message translates to:
  /// **'Deux liens : clients (app) et cabinets (inscription Pro avec votre code parrain).'**
  String get appInviteHintCommercial;

  /// No description provided for @appInviteHintShort.
  ///
  /// In fr, this message translates to:
  /// **'Lien de téléchargement et rattachement'**
  String get appInviteHintShort;

  /// No description provided for @appInviteCodeLabel.
  ///
  /// In fr, this message translates to:
  /// **'Code :'**
  String get appInviteCodeLabel;

  /// No description provided for @appInviteCopy.
  ///
  /// In fr, this message translates to:
  /// **'Copier le lien'**
  String get appInviteCopy;

  /// No description provided for @appInviteCopied.
  ///
  /// In fr, this message translates to:
  /// **'Lien copié'**
  String get appInviteCopied;

  /// No description provided for @appInviteLoadError.
  ///
  /// In fr, this message translates to:
  /// **'Impossible de charger le QR'**
  String get appInviteLoadError;

  /// No description provided for @appInviteRetry.
  ///
  /// In fr, this message translates to:
  /// **'Réessayer'**
  String get appInviteRetry;

  /// No description provided for @proLightVetTitle.
  ///
  /// In fr, this message translates to:
  /// **'Terrain véto'**
  String get proLightVetTitle;

  /// No description provided for @commercialFieldTitle.
  ///
  /// In fr, this message translates to:
  /// **'Commercial'**
  String get commercialFieldTitle;

  /// No description provided for @commercialFieldSubtitle.
  ///
  /// In fr, this message translates to:
  /// **'QR invitation clients et accès au site Pro.'**
  String get commercialFieldSubtitle;

  /// No description provided for @commercialManagerFieldSubtitle.
  ///
  /// In fr, this message translates to:
  /// **'Résultats d\'équipe, QR invitation et accès au site Pro.'**
  String get commercialManagerFieldSubtitle;

  /// No description provided for @commercialOpenProWeb.
  ///
  /// In fr, this message translates to:
  /// **'Ouvrir le site Pro'**
  String get commercialOpenProWeb;

  /// No description provided for @managerTeamCta.
  ///
  /// In fr, this message translates to:
  /// **'Résultats de mon équipe'**
  String get managerTeamCta;

  /// No description provided for @managerTeamTitle.
  ///
  /// In fr, this message translates to:
  /// **'Équipe'**
  String get managerTeamTitle;

  /// No description provided for @managerTeamSection.
  ///
  /// In fr, this message translates to:
  /// **'Résultats équipe'**
  String get managerTeamSection;

  /// No description provided for @managerSelfSection.
  ///
  /// In fr, this message translates to:
  /// **'Mes résultats'**
  String get managerSelfSection;

  /// No description provided for @managerMembersSection.
  ///
  /// In fr, this message translates to:
  /// **'Commerciaux'**
  String get managerMembersSection;

  /// No description provided for @managerTeamEmpty.
  ///
  /// In fr, this message translates to:
  /// **'Aucun commercial rattaché.'**
  String get managerTeamEmpty;

  /// No description provided for @managerKpiProspects.
  ///
  /// In fr, this message translates to:
  /// **'Prospects'**
  String get managerKpiProspects;

  /// No description provided for @managerKpiConverted.
  ///
  /// In fr, this message translates to:
  /// **'Convertis'**
  String get managerKpiConverted;

  /// No description provided for @managerKpiConversion.
  ///
  /// In fr, this message translates to:
  /// **'Taux de conversion'**
  String get managerKpiConversion;

  /// No description provided for @managerKpiAppointments.
  ///
  /// In fr, this message translates to:
  /// **'RDV à venir'**
  String get managerKpiAppointments;

  /// No description provided for @managerKpiStale.
  ///
  /// In fr, this message translates to:
  /// **'Stale pipeline'**
  String get managerKpiStale;

  /// No description provided for @managerKpiMonthEarned.
  ///
  /// In fr, this message translates to:
  /// **'Commissions (mois)'**
  String get managerKpiMonthEarned;

  /// No description provided for @managerKpiLifetimeEarned.
  ///
  /// In fr, this message translates to:
  /// **'Commissions (total)'**
  String get managerKpiLifetimeEarned;

  /// No description provided for @managerKpiVets.
  ///
  /// In fr, this message translates to:
  /// **'Véto assignés'**
  String get managerKpiVets;

  /// No description provided for @featureModules.
  ///
  /// In fr, this message translates to:
  /// **'Options'**
  String get featureModules;

  /// No description provided for @featureModulesCatalog.
  ///
  /// In fr, this message translates to:
  /// **'Découvrir les options'**
  String get featureModulesCatalog;

  /// No description provided for @featureModulesSubtitle.
  ///
  /// In fr, this message translates to:
  /// **'Activez Care+, Horse, Kennel ou Foyer selon vos besoins.'**
  String get featureModulesSubtitle;

  /// No description provided for @moduleCarePlus.
  ///
  /// In fr, this message translates to:
  /// **'Care+'**
  String get moduleCarePlus;

  /// No description provided for @moduleCarePlusDesc.
  ///
  /// In fr, this message translates to:
  /// **'Rappels de soins enrichis pour tous vos animaux.'**
  String get moduleCarePlusDesc;

  /// No description provided for @moduleHorse.
  ///
  /// In fr, this message translates to:
  /// **'Horse'**
  String get moduleHorse;

  /// No description provided for @moduleHorseDesc.
  ///
  /// In fr, this message translates to:
  /// **'Contacts pros et compétitions pour les chevaux.'**
  String get moduleHorseDesc;

  /// No description provided for @moduleKennel.
  ///
  /// In fr, this message translates to:
  /// **'Kennel'**
  String get moduleKennel;

  /// No description provided for @moduleKennelDesc.
  ///
  /// In fr, this message translates to:
  /// **'Encodage rapide d’élevage / portée.'**
  String get moduleKennelDesc;

  /// No description provided for @moduleFamily.
  ///
  /// In fr, this message translates to:
  /// **'Foyer'**
  String get moduleFamily;

  /// No description provided for @moduleFamilyDesc.
  ///
  /// In fr, this message translates to:
  /// **'Vue foyer de vos animaux.'**
  String get moduleFamilyDesc;

  /// No description provided for @moduleActivate.
  ///
  /// In fr, this message translates to:
  /// **'Activer'**
  String get moduleActivate;

  /// No description provided for @moduleActive.
  ///
  /// In fr, this message translates to:
  /// **'Actif'**
  String get moduleActive;

  /// No description provided for @switchProfile.
  ///
  /// In fr, this message translates to:
  /// **'Changer de profil'**
  String get switchProfile;

  /// No description provided for @profilePersonal.
  ///
  /// In fr, this message translates to:
  /// **'Personnel'**
  String get profilePersonal;

  /// No description provided for @profilePro.
  ///
  /// In fr, this message translates to:
  /// **'Professionnel'**
  String get profilePro;

  /// No description provided for @profileSwitched.
  ///
  /// In fr, this message translates to:
  /// **'Profil basculé'**
  String get profileSwitched;

  /// No description provided for @preconsultTitle.
  ///
  /// In fr, this message translates to:
  /// **'Pré-consultation'**
  String get preconsultTitle;

  /// No description provided for @preconsultTitlePet.
  ///
  /// In fr, this message translates to:
  /// **'Pré-consultation — {petName}'**
  String preconsultTitlePet(String petName);

  /// No description provided for @preconsultIntro.
  ///
  /// In fr, this message translates to:
  /// **'Indiquez l\'état de votre animal avant la visite. Votre vétérinaire pourra s\'y préparer.'**
  String get preconsultIntro;

  /// No description provided for @preconsultComplaint.
  ///
  /// In fr, this message translates to:
  /// **'Motif / plainte principale'**
  String get preconsultComplaint;

  /// No description provided for @preconsultComplaintRequired.
  ///
  /// In fr, this message translates to:
  /// **'Indiquez le motif de la visite'**
  String get preconsultComplaintRequired;

  /// No description provided for @preconsultDuration.
  ///
  /// In fr, this message translates to:
  /// **'Depuis quand ?'**
  String get preconsultDuration;

  /// No description provided for @preconsultDurationToday.
  ///
  /// In fr, this message translates to:
  /// **'Aujourd\'hui'**
  String get preconsultDurationToday;

  /// No description provided for @preconsultDurationFewDays.
  ///
  /// In fr, this message translates to:
  /// **'Quelques jours'**
  String get preconsultDurationFewDays;

  /// No description provided for @preconsultDurationWeek.
  ///
  /// In fr, this message translates to:
  /// **'Environ une semaine'**
  String get preconsultDurationWeek;

  /// No description provided for @preconsultDurationWeeks.
  ///
  /// In fr, this message translates to:
  /// **'Plusieurs semaines'**
  String get preconsultDurationWeeks;

  /// No description provided for @preconsultDurationMonths.
  ///
  /// In fr, this message translates to:
  /// **'Plusieurs mois'**
  String get preconsultDurationMonths;

  /// No description provided for @preconsultBehavior.
  ///
  /// In fr, this message translates to:
  /// **'Comportement'**
  String get preconsultBehavior;

  /// No description provided for @preconsultBehaviorNormal.
  ///
  /// In fr, this message translates to:
  /// **'Normal'**
  String get preconsultBehaviorNormal;

  /// No description provided for @preconsultBehaviorLethargic.
  ///
  /// In fr, this message translates to:
  /// **'Apathique'**
  String get preconsultBehaviorLethargic;

  /// No description provided for @preconsultBehaviorRestless.
  ///
  /// In fr, this message translates to:
  /// **'Agité'**
  String get preconsultBehaviorRestless;

  /// No description provided for @preconsultBehaviorAggressive.
  ///
  /// In fr, this message translates to:
  /// **'Agressif'**
  String get preconsultBehaviorAggressive;

  /// No description provided for @preconsultBehaviorAnxious.
  ///
  /// In fr, this message translates to:
  /// **'Anxieux'**
  String get preconsultBehaviorAnxious;

  /// No description provided for @preconsultBehaviorOther.
  ///
  /// In fr, this message translates to:
  /// **'Autre'**
  String get preconsultBehaviorOther;

  /// No description provided for @preconsultAppetite.
  ///
  /// In fr, this message translates to:
  /// **'Appétit'**
  String get preconsultAppetite;

  /// No description provided for @preconsultThirst.
  ///
  /// In fr, this message translates to:
  /// **'Soif'**
  String get preconsultThirst;

  /// No description provided for @preconsultElimination.
  ///
  /// In fr, this message translates to:
  /// **'Selles / urines'**
  String get preconsultElimination;

  /// No description provided for @preconsultScaleNormal.
  ///
  /// In fr, this message translates to:
  /// **'Normal'**
  String get preconsultScaleNormal;

  /// No description provided for @preconsultScaleDecreased.
  ///
  /// In fr, this message translates to:
  /// **'Diminué'**
  String get preconsultScaleDecreased;

  /// No description provided for @preconsultScaleIncreased.
  ///
  /// In fr, this message translates to:
  /// **'Augmenté'**
  String get preconsultScaleIncreased;

  /// No description provided for @preconsultUrgency.
  ///
  /// In fr, this message translates to:
  /// **'Urgence perçue'**
  String get preconsultUrgency;

  /// No description provided for @preconsultUrgencyLow.
  ///
  /// In fr, this message translates to:
  /// **'Faible'**
  String get preconsultUrgencyLow;

  /// No description provided for @preconsultUrgencyMedium.
  ///
  /// In fr, this message translates to:
  /// **'Moyenne'**
  String get preconsultUrgencyMedium;

  /// No description provided for @preconsultUrgencyHigh.
  ///
  /// In fr, this message translates to:
  /// **'Élevée'**
  String get preconsultUrgencyHigh;

  /// No description provided for @preconsultComment.
  ///
  /// In fr, this message translates to:
  /// **'Commentaire (optionnel)'**
  String get preconsultComment;

  /// No description provided for @preconsultUnknown.
  ///
  /// In fr, this message translates to:
  /// **'Je ne sais pas'**
  String get preconsultUnknown;

  /// No description provided for @preconsultSubmit.
  ///
  /// In fr, this message translates to:
  /// **'Envoyer'**
  String get preconsultSubmit;

  /// No description provided for @preconsultSubmitted.
  ///
  /// In fr, this message translates to:
  /// **'Pré-consultation envoyée'**
  String get preconsultSubmitted;

  /// No description provided for @preconsultAlreadySubmitted.
  ///
  /// In fr, this message translates to:
  /// **'Vous avez déjà envoyé cette pré-consultation.'**
  String get preconsultAlreadySubmitted;

  /// No description provided for @preconsultFillCta.
  ///
  /// In fr, this message translates to:
  /// **'Remplir la pré-consultation'**
  String get preconsultFillCta;

  /// No description provided for @proLightAudioConsentClientCheck.
  ///
  /// In fr, this message translates to:
  /// **'J\'ai obtenu l\'accord oral du client pour enregistrer'**
  String get proLightAudioConsentClientCheck;

  /// No description provided for @proLightReportAiProposalBanner.
  ///
  /// In fr, this message translates to:
  /// **'Proposition IA — validation obligatoire avant finalisation (diagnostic / médication inclus).'**
  String get proLightReportAiProposalBanner;

  /// No description provided for @proLightAiModuleRequired.
  ///
  /// In fr, this message translates to:
  /// **'Module CR IA non activé ou essai expiré — contactez votre commercial petsFollow.'**
  String get proLightAiModuleRequired;

  /// No description provided for @proLightAiModuleTrialBanner.
  ///
  /// In fr, this message translates to:
  /// **'Essai CR IA — {days} j restants. Dictez puis améliorez vos comptes rendus.'**
  String proLightAiModuleTrialBanner(int days);

  /// No description provided for @proLightAiModuleInactiveBanner.
  ///
  /// In fr, this message translates to:
  /// **'CR IA non activé pour ce cabinet. La dictée IA sera indisponible.'**
  String get proLightAiModuleInactiveBanner;

  /// No description provided for @proLightAiModuleVisitScopedBanner.
  ///
  /// In fr, this message translates to:
  /// **'CR IA disponible si le cabinet de la visite a le module activé.'**
  String get proLightAiModuleVisitScopedBanner;

  /// No description provided for @supportTitle.
  ///
  /// In fr, this message translates to:
  /// **'Signaler un problème'**
  String get supportTitle;

  /// No description provided for @supportHint.
  ///
  /// In fr, this message translates to:
  /// **'Décrivez le bug. Les diagnostics techniques des 15 dernières minutes sont joints automatiquement.'**
  String get supportHint;

  /// No description provided for @supportSubject.
  ///
  /// In fr, this message translates to:
  /// **'Sujet'**
  String get supportSubject;

  /// No description provided for @supportMessage.
  ///
  /// In fr, this message translates to:
  /// **'Description'**
  String get supportMessage;

  /// No description provided for @supportDiagnosticsAttached.
  ///
  /// In fr, this message translates to:
  /// **'Diagnostics joints automatiquement (erreurs, requêtes, configuration).'**
  String get supportDiagnosticsAttached;

  /// No description provided for @supportSubmit.
  ///
  /// In fr, this message translates to:
  /// **'Envoyer'**
  String get supportSubmit;

  /// No description provided for @supportSending.
  ///
  /// In fr, this message translates to:
  /// **'Envoi…'**
  String get supportSending;

  /// No description provided for @supportSuccess.
  ///
  /// In fr, this message translates to:
  /// **'Message envoyé. Merci !'**
  String get supportSuccess;

  /// No description provided for @supportErrorRateLimit.
  ///
  /// In fr, this message translates to:
  /// **'Trop de tickets récemment. Réessayez dans une heure.'**
  String get supportErrorRateLimit;

  /// No description provided for @supportErrorTooLarge.
  ///
  /// In fr, this message translates to:
  /// **'Diagnostics trop volumineux. Relancez l\'app et réessayez.'**
  String get supportErrorTooLarge;

  /// No description provided for @supportMenu.
  ///
  /// In fr, this message translates to:
  /// **'Support'**
  String get supportMenu;

  /// No description provided for @appInviteHintSales.
  ///
  /// In fr, this message translates to:
  /// **'Partagez votre code parrain avec un cabinet (inscription Pro) ou un client (invitation app).'**
  String get appInviteHintSales;

  /// No description provided for @appInviteCopyVet.
  ///
  /// In fr, this message translates to:
  /// **'Copier le lien inscription cabinet'**
  String get appInviteCopyVet;

  /// No description provided for @appInviteCopyClient.
  ///
  /// In fr, this message translates to:
  /// **'Copier le lien invitation client'**
  String get appInviteCopyClient;

  /// No description provided for @sendDossierToPro.
  ///
  /// In fr, this message translates to:
  /// **'Envoyer vers un pro'**
  String get sendDossierToPro;

  /// No description provided for @sendDossierEmailLabel.
  ///
  /// In fr, this message translates to:
  /// **'E-mail du professionnel'**
  String get sendDossierEmailLabel;

  /// No description provided for @sendDossierEmailHint.
  ///
  /// In fr, this message translates to:
  /// **'vet@cabinet.be'**
  String get sendDossierEmailHint;

  /// No description provided for @sendDossierConfirm.
  ///
  /// In fr, this message translates to:
  /// **'Envoyer'**
  String get sendDossierConfirm;

  /// No description provided for @sendDossierSuccess.
  ///
  /// In fr, this message translates to:
  /// **'Dossier envoyé — lien valable 24 h.'**
  String get sendDossierSuccess;

  /// No description provided for @sendDossierInvalidEmail.
  ///
  /// In fr, this message translates to:
  /// **'Adresse e-mail invalide.'**
  String get sendDossierInvalidEmail;

  /// No description provided for @sendDossierPhiWarning.
  ///
  /// In fr, this message translates to:
  /// **'Ce dossier contient des données de santé : comptes rendus de visite, carnet de santé et documents. Le lien reste valable 24 h et n\'importe qui le possédant pourra les consulter.'**
  String get sendDossierPhiWarning;

  /// No description provided for @sendDossierPhiConsent.
  ///
  /// In fr, this message translates to:
  /// **'J\'accepte de partager ces données de santé avec ce professionnel.'**
  String get sendDossierPhiConsent;
}

class _AppLocalizationsDelegate
    extends LocalizationsDelegate<AppLocalizations> {
  const _AppLocalizationsDelegate();

  @override
  Future<AppLocalizations> load(Locale locale) {
    return SynchronousFuture<AppLocalizations>(lookupAppLocalizations(locale));
  }

  @override
  bool isSupported(Locale locale) => <String>[
        'en',
        'es',
        'et',
        'fr',
        'it',
        'nl'
      ].contains(locale.languageCode);

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
    case 'it':
      return AppLocalizationsIt();
    case 'nl':
      return AppLocalizationsNl();
  }

  throw FlutterError(
      'AppLocalizations.delegate failed to load unsupported locale "$locale". This is likely '
      'an issue with the localizations generation tool. Please file an issue '
      'on GitHub with a reproducible sample app and the gen-l10n configuration '
      'that was used.');
}
