// ignore: unused_import
import 'package:intl/intl.dart' as intl;
import 'app_localizations.dart';

// ignore_for_file: type=lint

/// The translations for French (`fr`).
class AppLocalizationsFr extends AppLocalizations {
  AppLocalizationsFr([String locale = 'fr']) : super(locale);

  @override
  String get appTitle => 'petsFollow';

  @override
  String get appTagline => 'Suivi santé de votre animal';

  @override
  String get email => 'Email';

  @override
  String get password => 'Mot de passe';

  @override
  String get login => 'Se connecter';

  @override
  String get loginFailed => 'Connexion impossible';

  @override
  String get myPets => 'Mes animaux';

  @override
  String get language => 'Langue';

  @override
  String get languageFr => 'Français';

  @override
  String get languageNl => 'Nederlands';

  @override
  String get languageEn => 'English';

  @override
  String get paymentResume => 'Reprendre le paiement';

  @override
  String get manageSubscription => 'Gérer mon abonnement';

  @override
  String get heartRate => 'Relevé cardiaque';

  @override
  String get history => 'Historique';

  @override
  String get vetMessaging => 'Messagerie véto';

  @override
  String get badgeAutoRenew => 'Renouvellement auto';

  @override
  String get badgeActive => 'Actif';

  @override
  String get badgePendingPayment => 'En attente de paiement';

  @override
  String badgeExpiresOn(String date) {
    return 'expire $date';
  }

  @override
  String get newPet => 'Nouvel animal';

  @override
  String get petName => 'Nom';

  @override
  String get species => 'Espèce';

  @override
  String get breed => 'Race';

  @override
  String get choosePlan => 'Choisissez votre formule';

  @override
  String get recommended => 'Recommandé';

  @override
  String get autoRenewTitle => 'Renouveler automatiquement';

  @override
  String get autoRenewSubtitle => 'Prélèvement à chaque échéance';

  @override
  String get continueToPayment => 'Continuer vers le paiement';

  @override
  String get paymentConfirmed => 'Paiement confirmé — animal actif';

  @override
  String get paymentPending =>
      'Paiement en attente — vous pourrez reprendre plus tard';

  @override
  String errorGeneric(String message) {
    return 'Erreur: $message';
  }

  @override
  String planAnnualSub(String price) {
    return '$price, renouvelé automatiquement';
  }

  @override
  String get planTriennialSub =>
      '60 € tous les 3 ans, renouvelé automatiquement';

  @override
  String get planQuinquennialSub =>
      '75 € tous les 5 ans, renouvelé automatiquement';

  @override
  String planOneTime(String price) {
    return '$price, paiement unique';
  }

  @override
  String get heartRateInstructions =>
      'Tapotez à chaque battement pendant 60 secondes.';

  @override
  String get start => 'Démarrer';

  @override
  String secondsLeft(int seconds) {
    return '$seconds s';
  }

  @override
  String beatsCount(int count) {
    return '$count battements';
  }

  @override
  String get tapHere => 'Tapez ici à chaque battement';

  @override
  String bpmLabel(String bpm) {
    return 'BPM: $bpm';
  }

  @override
  String beatsLabel(int count) {
    return 'Battements: $count';
  }

  @override
  String get thresholdAlert => 'Alerte seuil';

  @override
  String get validateAndSend => 'Valider et envoyer au véto';

  @override
  String get restart => 'Recommencer';

  @override
  String get sentToVet => 'Relevé envoyé au véto';
}
