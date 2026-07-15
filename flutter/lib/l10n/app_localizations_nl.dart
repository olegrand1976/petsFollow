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
