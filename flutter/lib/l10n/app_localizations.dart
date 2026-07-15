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
