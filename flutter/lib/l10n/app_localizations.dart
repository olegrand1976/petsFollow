import 'dart:async';

import 'package:flutter/foundation.dart';
import 'package:flutter/widgets.dart';
import 'package:flutter_localizations/flutter_localizations.dart';
import 'package:intl/intl.dart' as intl;

import 'app_localizations_en.dart';
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

  /// No description provided for @changePassword.
  ///
  /// In fr, this message translates to:
  /// **'Changer le mot de passe'**
  String get changePassword;

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

  /// No description provided for @planAnnualSub.
  ///
  /// In fr, this message translates to:
  /// **'{price}, renouvelé automatiquement'**
  String planAnnualSub(String price);

  /// No description provided for @planTriennialSub.
  ///
  /// In fr, this message translates to:
  /// **'60 € tous les 3 ans, renouvelé automatiquement'**
  String get planTriennialSub;

  /// No description provided for @planQuinquennialSub.
  ///
  /// In fr, this message translates to:
  /// **'75 € tous les 5 ans, renouvelé automatiquement'**
  String get planQuinquennialSub;

  /// No description provided for @planOneTime.
  ///
  /// In fr, this message translates to:
  /// **'{price}, paiement unique'**
  String planOneTime(String price);

  /// No description provided for @heartRateInstructions.
  ///
  /// In fr, this message translates to:
  /// **'Tapotez à chaque battement pendant 60 secondes.'**
  String get heartRateInstructions;

  /// No description provided for @heartRateInstructionsDuration.
  ///
  /// In fr, this message translates to:
  /// **'Tapotez à chaque battement pendant {seconds} secondes.'**
  String heartRateInstructionsDuration(int seconds);

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
      <String>['en', 'fr', 'nl'].contains(locale.languageCode);

  @override
  bool shouldReload(_AppLocalizationsDelegate old) => false;
}

AppLocalizations lookupAppLocalizations(Locale locale) {
  // Lookup logic when only language code is specified.
  switch (locale.languageCode) {
    case 'en':
      return AppLocalizationsEn();
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
