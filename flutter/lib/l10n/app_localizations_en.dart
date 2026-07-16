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
  String get changePhoto => 'Change photo';

  @override
  String get addPhoto => 'Add a photo';

  @override
  String get photoUpdated => 'Photo updated';

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

  @override
  String get navHome => 'Home';

  @override
  String get navPets => 'Pets';

  @override
  String get navCare => 'Care';

  @override
  String get navMessages => 'Messages';

  @override
  String get navProfile => 'Profile';

  @override
  String get speciesDog => 'Dog';

  @override
  String get speciesCat => 'Cat';

  @override
  String get speciesHorse => 'Horse';

  @override
  String get speciesOther => 'Other';

  @override
  String get careComingSoon => 'Care reminders coming soon';

  @override
  String get emptyPetsTitle => 'No pets yet';

  @override
  String get emptyPetsBody =>
      'Add your first pet to start heart rate monitoring with your veterinarian.';

  @override
  String get discoveryTitle => 'Discover petsFollow';

  @override
  String get discoveryMission => 'Your 7-day journey';

  @override
  String get discoveryDay0Title => 'Day 0 — Welcome';

  @override
  String get discoveryDay0Body =>
      'Create your pet\'s profile and learn how to measure heart rate.';

  @override
  String get discoveryDay2Title => 'Day 2 — First reading';

  @override
  String get discoveryDay2Body =>
      'Take your first heart rate reading and get comfortable with the technique.';

  @override
  String get discoveryDay4Title => 'Day 4 — Routine';

  @override
  String get discoveryDay4Body =>
      'Build a daily measurement habit with personalized reminders.';

  @override
  String get discoveryDay6Title => 'Day 6 — Share with vet';

  @override
  String get discoveryDay6Body =>
      'Your readings are shared with your vet for optimal follow-up.';

  @override
  String get myVets => 'My veterinarians';

  @override
  String get addVetByEmail => 'Add a vet by email';

  @override
  String get vetEmailHint => 'email@practice.vet';

  @override
  String get noVets => 'No linked veterinarian';

  @override
  String get primaryVet => 'Primary veterinarian';

  @override
  String get setPrimaryVet => 'Set as primary vet';

  @override
  String get careTitle => 'Care';

  @override
  String get careDone => 'Done';

  @override
  String get carePostpone => 'Postpone';

  @override
  String get careOverdue => 'Overdue';

  @override
  String get visitHistory => 'Visit history';

  @override
  String get requestVisit => 'Request a visit';

  @override
  String get upcomingVisit => 'Upcoming visit';

  @override
  String get notificationHrTitle => 'petsFollow';

  @override
  String get notificationHrBody => 'Time for a heart rate reading for your pet';

  @override
  String get reviewAskTitle => 'Enjoying petsFollow?';

  @override
  String get reviewAskYes => 'Yes, rate the app';

  @override
  String get reviewAskNo => 'Later';

  @override
  String get carePlusUpsell =>
      'Care+ — personalized medication and care reminders';

  @override
  String get familyPackHint => 'Family pack — multiple pets, one subscription';

  @override
  String get discoveryMarkDone => 'Mission complete';

  @override
  String get notificationPreferences => 'Notification preferences';

  @override
  String get notificationPrefsHint =>
      'Choose which notification types you want to receive.';

  @override
  String get notificationPrefsSaved => 'Preferences saved';

  @override
  String get notificationPrefHr => 'Heart rate readings';

  @override
  String get notificationPrefCare => 'Care reminders';

  @override
  String get notificationPrefVisits => 'Visits';

  @override
  String get notificationPrefMessages => 'Messages';

  @override
  String get notificationPrefDiscovery => 'Discovery journey';

  @override
  String get notificationPrefBilling => 'Billing';

  @override
  String carePostponeDays(int days) {
    return 'Postpone by $days days';
  }

  @override
  String get noCareReminders => 'No pending care reminders';

  @override
  String get noThreads => 'No conversations';

  @override
  String get vetInviteSent => 'Invitation sent to veterinarian';

  @override
  String get visitRequested => 'Visit request sent';

  @override
  String get primaryVetSet => 'Primary veterinarian updated';

  @override
  String get visitStatusRequested => 'Requested';

  @override
  String get visitStatusConfirmed => 'Confirmed';

  @override
  String get visitStatusDone => 'Completed';

  @override
  String get visitStatusCancelled => 'Cancelled';

  @override
  String get horseHealthTitle => 'Horse health';

  @override
  String get horseContactsTitle => 'Contacts (farrier, dentist…)';

  @override
  String get horseCompetitionsTitle => 'Competitions';

  @override
  String get horseContactsSoon =>
      'Manage your professional contacts — coming soon.';

  @override
  String get horseCompetitionsSoon =>
      'Competition calendar and results — coming soon.';

  @override
  String get horsePackUpsell =>
      'Horse pack — farrier, fecal egg count and competition tracking';

  @override
  String get careTypeFarrier => 'Farrier';

  @override
  String get careTypeFecalEgg => 'Fecal egg count';

  @override
  String get careTypeVaccination => 'Vaccination';

  @override
  String get careTypeDeworming => 'Deworming';

  @override
  String get careTypeVetCheck => 'Vet check-up';

  @override
  String get careTypeDental => 'Dental care';

  @override
  String get careTypeCustom => 'Custom reminder';

  @override
  String get homeAddFirstVetTitle => 'Add your veterinarian';

  @override
  String get homeAddFirstVetBody =>
      'Link the practice that follows your pet to share readings and chat.';

  @override
  String get homeAddFirstVetCta => 'Add a veterinarian';

  @override
  String get photoFrameHint => 'Center the muzzle — pet profile preview';

  @override
  String get takePhoto => 'Take a photo';

  @override
  String get chooseFromGallery => 'Choose from gallery';

  @override
  String get attachMedia => 'Attach a photo or video';

  @override
  String get attachPhoto => 'Photo';

  @override
  String get attachVideo => 'Video';

  @override
  String get openMedia => 'Open';

  @override
  String get mediaVideoLabel => 'Video';
}
