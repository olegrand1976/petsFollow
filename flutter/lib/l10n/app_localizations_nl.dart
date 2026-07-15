// ignore: unused_import
import 'package:intl/intl.dart' as intl;
import 'app_localizations.dart';

// ignore_for_file: type=lint

/// The translations for Dutch Flemish (`nl`).
class AppLocalizationsNl extends AppLocalizations {
  AppLocalizationsNl([String locale = 'nl']) : super(locale);

  @override
  String get appTitle => 'petsFollow';

  @override
  String get appTagline => 'Gezondheidsmonitoring van uw huisdier';

  @override
  String get email => 'E-mail';

  @override
  String get password => 'Wachtwoord';

  @override
  String get login => 'Inloggen';

  @override
  String get loginFailed => 'Inloggen mislukt';

  @override
  String get myPets => 'Mijn huisdieren';

  @override
  String get myData => 'Mijn gegevens';

  @override
  String get settings => 'Instellingen';

  @override
  String get logout => 'Afmelden';

  @override
  String get save => 'Opslaan';

  @override
  String get cancel => 'Annuleren';

  @override
  String get firstName => 'Voornaam';

  @override
  String get currentPassword => 'Huidig wachtwoord';

  @override
  String get newPassword => 'Nieuw wachtwoord';

  @override
  String get changePassword => 'Wachtwoord wijzigen';

  @override
  String get deleteAccount => 'Account verwijderen';

  @override
  String get deleteAccountConfirm =>
      'Deze actie is onomkeerbaar. Al uw huisdieren en gegevens worden verwijderd.';

  @override
  String get profileSaved => 'Profiel opgeslagen';

  @override
  String get changePhoto => 'Foto wijzigen';

  @override
  String get addPhoto => 'Foto toevoegen';

  @override
  String get photoUpdated => 'Foto bijgewerkt';

  @override
  String get passwordChanged => 'Wachtwoord gewijzigd';

  @override
  String greeting(String name) {
    return 'Hallo $name,';
  }

  @override
  String get latestValues => 'Laatste waarden';

  @override
  String get startMeasurement => 'METING STARTEN';

  @override
  String get chooseDuration => 'Meetduur';

  @override
  String durationSeconds(int seconds) {
    return '$seconds s';
  }

  @override
  String get howToMeasure => 'Hoe meten?';

  @override
  String get howToMeasureIntro => 'Meet de hartslag van uw huisdier in rust.';

  @override
  String get howToMeasureStep1 =>
      '1. Houd uw huisdier rustig, liggend of zittend.';

  @override
  String get howToMeasureStep2 =>
      '2. Leg uw hand op de borst en tik bij elke slag gedurende de aangegeven tijd.';

  @override
  String get howToMeasureStep3 =>
      '3. Valideer de meting om deze naar uw dierenarts te sturen.';

  @override
  String get howToMeasureWhyTitle => 'Waarom meten?';

  @override
  String get howToMeasureWhyBody =>
      'Regelmatige hartslagmonitoring helpt veranderingen op te sporen en de behandeling met uw dierenarts aan te passen.';

  @override
  String get reminders => 'Herinneringen';

  @override
  String get remindersHint =>
      'Ontvang een dagelijkse herinnering voor een hartslagmeting.';

  @override
  String get remindersEnabled => 'Herinneringen inschakelen';

  @override
  String get remindersTime => 'Tijdstip herinnering';

  @override
  String get remindersSaved => 'Herinneringen opgeslagen';

  @override
  String get legalTermsTitle => 'Gebruiksvoorwaarden';

  @override
  String get legalPrivacyTitle => 'Privacybeleid';

  @override
  String get legalNoticeTitle => 'Juridische vermeldingen';

  @override
  String get legalTermsBody =>
      'Gebruiksvoorwaarden — petsFollow\n\nDe petsFollow-app laat eigenaars de hartslag meten, de geschiedenis bekijken en communiceren met hun dierenarts.\n\nDiensten worden geleverd in het kader van het gekozen abonnement.\n\nLaatst bijgewerkt: juli 2026';

  @override
  String get legalPrivacyBody =>
      'Privacybeleid — petsFollow\n\nVerzamelde gegevens: voornaam, e-mail, huisdiergegevens, hartslagmetingen, berichten aan de dierenarts.\n\nDoeleinden: accountbeheer, gezondheidsmonitoring, communicatie met de praktijk.\n\nBewaring: tot verwijdering van het account of 3 jaar inactiviteit.\n\nLaatst bijgewerkt: juli 2026';

  @override
  String get legalNoticeBody =>
      'Juridische vermeldingen — petsFollow\n\nUitgever: petsFollow\nContact: support@petsfollow.test\n\nLaatst bijgewerkt: juli 2026';

  @override
  String get language => 'Taal';

  @override
  String get languageFr => 'Français';

  @override
  String get languageNl => 'Nederlands';

  @override
  String get languageEn => 'English';

  @override
  String get paymentResume => 'Betaling hervatten';

  @override
  String get manageSubscription => 'Abonnement beheren';

  @override
  String get heartRate => 'Hartslagmeting';

  @override
  String get history => 'Geschiedenis';

  @override
  String get vetMessaging => 'Berichten dierenarts';

  @override
  String get badgeAutoRenew => 'Automatische verlenging';

  @override
  String get badgeActive => 'Actief';

  @override
  String get badgePendingPayment => 'Betaling in behandeling';

  @override
  String badgeExpiresOn(String date) {
    return 'verloopt $date';
  }

  @override
  String get newPet => 'Nieuw huisdier';

  @override
  String get petName => 'Naam';

  @override
  String get species => 'Soort';

  @override
  String get breed => 'Ras';

  @override
  String get choosePlan => 'Kies uw formule';

  @override
  String get recommended => 'Aanbevolen';

  @override
  String get autoRenewTitle => 'Automatisch verlengen';

  @override
  String get autoRenewSubtitle => 'Incasso bij elke vervaldatum';

  @override
  String get continueToPayment => 'Doorgaan naar betaling';

  @override
  String get paymentConfirmed => 'Betaling bevestigd — huisdier actief';

  @override
  String get paymentPending =>
      'Betaling in behandeling — u kunt later verdergaan';

  @override
  String errorGeneric(String message) {
    return 'Fout: $message';
  }

  @override
  String planAnnualSub(String price) {
    return '$price, automatisch verlengd';
  }

  @override
  String get planTriennialSub => '60 € elke 3 jaar, automatisch verlengd';

  @override
  String get planQuinquennialSub => '75 € elke 5 jaar, automatisch verlengd';

  @override
  String planOneTime(String price) {
    return '$price, eenmalige betaling';
  }

  @override
  String get heartRateInstructions =>
      'Tik bij elke hartslag gedurende 60 seconden.';

  @override
  String heartRateInstructionsDuration(int seconds) {
    return 'Tik bij elke hartslag gedurende $seconds seconden.';
  }

  @override
  String get start => 'Starten';

  @override
  String secondsLeft(int seconds) {
    return '$seconds s';
  }

  @override
  String beatsCount(int count) {
    return '$count slagen';
  }

  @override
  String get tapHere => 'Tik hier bij elke slag';

  @override
  String bpmLabel(String bpm) {
    return 'BPM: $bpm';
  }

  @override
  String beatsLabel(int count) {
    return 'Slagen: $count';
  }

  @override
  String get thresholdAlert => 'Drempelwaarschuwing';

  @override
  String get validateAndSend => 'Valideren en naar dierenarts sturen';

  @override
  String get restart => 'Opnieuw beginnen';

  @override
  String get sentToVet => 'Meting naar dierenarts gestuurd';
}
