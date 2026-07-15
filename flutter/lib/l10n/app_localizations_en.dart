// ignore: unused_import
import 'package:intl/intl.dart' as intl;
import 'app_localizations.dart';

// ignore_for_file: type=lint

/// The translations for English (`en`).
class AppLocalizationsEn extends AppLocalizations {
  AppLocalizationsEn([String locale = 'en']) : super(locale);

  @override
  String get appTitle => 'petsFollow';

  @override
  String get appTagline => 'Health monitoring for your pet';

  @override
  String get email => 'Email';

  @override
  String get password => 'Password';

  @override
  String get login => 'Sign in';

  @override
  String get loginFailed => 'Sign in failed';

  @override
  String get myPets => 'My pets';

  @override
  String get language => 'Language';

  @override
  String get languageFr => 'Français';

  @override
  String get languageNl => 'Nederlands';

  @override
  String get languageEn => 'English';

  @override
  String get paymentResume => 'Resume payment';

  @override
  String get manageSubscription => 'Manage subscription';

  @override
  String get heartRate => 'Heart rate reading';

  @override
  String get history => 'History';

  @override
  String get vetMessaging => 'Vet messaging';

  @override
  String get badgeAutoRenew => 'Auto-renewal';

  @override
  String get badgeActive => 'Active';

  @override
  String get badgePendingPayment => 'Payment pending';

  @override
  String badgeExpiresOn(String date) {
    return 'expires $date';
  }

  @override
  String get newPet => 'New pet';

  @override
  String get petName => 'Name';

  @override
  String get species => 'Species';

  @override
  String get breed => 'Breed';

  @override
  String get choosePlan => 'Choose your plan';

  @override
  String get recommended => 'Recommended';

  @override
  String get autoRenewTitle => 'Auto-renew';

  @override
  String get autoRenewSubtitle => 'Charged at each renewal';

  @override
  String get continueToPayment => 'Continue to payment';

  @override
  String get paymentConfirmed => 'Payment confirmed — pet active';

  @override
  String get paymentPending => 'Payment pending — you can resume later';

  @override
  String errorGeneric(String message) {
    return 'Error: $message';
  }

  @override
  String planAnnualSub(String price) {
    return '$price, auto-renewed';
  }

  @override
  String get planTriennialSub => '€60 every 3 years, auto-renewed';

  @override
  String get planQuinquennialSub => '€75 every 5 years, auto-renewed';

  @override
  String planOneTime(String price) {
    return '$price, one-time payment';
  }

  @override
  String get heartRateInstructions => 'Tap on each beat for 60 seconds.';

  @override
  String get start => 'Start';

  @override
  String secondsLeft(int seconds) {
    return '$seconds s';
  }

  @override
  String beatsCount(int count) {
    return '$count beats';
  }

  @override
  String get tapHere => 'Tap here on each beat';

  @override
  String bpmLabel(String bpm) {
    return 'BPM: $bpm';
  }

  @override
  String beatsLabel(int count) {
    return 'Beats: $count';
  }

  @override
  String get thresholdAlert => 'Threshold alert';

  @override
  String get validateAndSend => 'Validate and send to vet';

  @override
  String get restart => 'Start over';

  @override
  String get sentToVet => 'Reading sent to vet';
}
