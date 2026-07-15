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
  String get myData => 'My data';

  @override
  String get settings => 'Settings';

  @override
  String get logout => 'Sign out';

  @override
  String get save => 'Save';

  @override
  String get cancel => 'Cancel';

  @override
  String get firstName => 'First name';

  @override
  String get currentPassword => 'Current password';

  @override
  String get newPassword => 'New password';

  @override
  String get changePassword => 'Change password';

  @override
  String get deleteAccount => 'Delete account';

  @override
  String get deleteAccountConfirm =>
      'This action cannot be undone. All your pets and data will be deleted.';

  @override
  String get profileSaved => 'Profile saved';

  @override
  String get passwordChanged => 'Password changed';

  @override
  String greeting(String name) {
    return 'Hello $name,';
  }

  @override
  String get latestValues => 'Latest values';

  @override
  String get startMeasurement => 'START MEASUREMENT';

  @override
  String get chooseDuration => 'Measurement duration';

  @override
  String durationSeconds(int seconds) {
    return '$seconds s';
  }

  @override
  String get howToMeasure => 'How to measure?';

  @override
  String get howToMeasureIntro => 'Measure your pet\'s resting heart rate.';

  @override
  String get howToMeasureStep1 =>
      '1. Keep your pet calm, lying down or sitting.';

  @override
  String get howToMeasureStep2 =>
      '2. Place your hand on the chest and tap on each beat for the indicated duration.';

  @override
  String get howToMeasureStep3 =>
      '3. Validate the reading to send it to your veterinarian.';

  @override
  String get howToMeasureWhyTitle => 'Why measure?';

  @override
  String get howToMeasureWhyBody =>
      'Regular heart rate monitoring helps detect changes and adjust treatment with your vet.';

  @override
  String get reminders => 'Reminders';

  @override
  String get remindersHint =>
      'Receive a daily reminder to take a heart rate reading.';

  @override
  String get remindersEnabled => 'Enable reminders';

  @override
  String get remindersTime => 'Reminder time';

  @override
  String get remindersSaved => 'Reminders saved';

  @override
  String get legalTermsTitle => 'Terms of use';

  @override
  String get legalPrivacyTitle => 'Privacy policy';

  @override
  String get legalNoticeTitle => 'Legal notice';

  @override
  String get legalTermsBody =>
      'Terms of use — petsFollow\n\nThe petsFollow app lets pet owners measure heart rate, view history and communicate with their veterinarian.\n\nServices are provided under the selected subscription. Users must use the app as intended.\n\nLast updated: July 2026';

  @override
  String get legalPrivacyBody =>
      'Privacy policy — petsFollow\n\nData collected: first name, email, pet data (name, species, breed), heart rate readings, messages to the vet.\n\nPurposes: account management, health monitoring, communication with the veterinary practice.\n\nRetention: until account deletion or 3 years of inactivity.\n\nYou may exercise your rights (access, rectification, deletion) via app settings.\n\nLast updated: July 2026';

  @override
  String get legalNoticeBody =>
      'Legal notice — petsFollow\n\nPublisher: petsFollow\nContact: support@petsfollow.test\n\nHosting: GDPR-compliant cloud infrastructure.\n\nPublication director: petsFollow.\n\nLast updated: July 2026';

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
  String heartRateInstructionsDuration(int seconds) {
    return 'Tap on each beat for $seconds seconds.';
  }

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
