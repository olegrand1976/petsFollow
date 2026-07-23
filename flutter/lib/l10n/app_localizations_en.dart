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
  String get loginOr => 'or';

  @override
  String get loginWithGoogle => 'Continue with Google';

  @override
  String get googleNotConfigured => 'Google sign-in is not configured';

  @override
  String get googleLoginFailed => 'Google sign-in failed';

  @override
  String get googleClientNotFound =>
      'No client account for this email. Ask your vet for an invite';

  @override
  String get googleWrongAudience =>
      'This Google account is not a client profile';

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
  String get confirmNewPassword => 'Confirm password';

  @override
  String get changePassword => 'Change password';

  @override
  String get forceChangePasswordTitle => 'Change your password';

  @override
  String get forceChangePasswordSubtitle =>
      'This account was created with a temporary password. Choose your own to continue.';

  @override
  String get forceChangePasswordSubmit => 'Save and continue';

  @override
  String get passwordTooShort => 'At least 8 characters';

  @override
  String get passwordMismatch => 'Passwords do not match';

  @override
  String get passwordChangeFailed => 'Could not change password';

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
  String get choosePetForMeasurement => 'Choose a pet';

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
  String get legalOpenOnline => 'View online version';

  @override
  String get legalTermsBody =>
      'Terms of use — petsFollow\n\nThe petsFollow app lets pet owners measure heart rate, view history and communicate with their veterinarian.\n\nServices are provided under the selected subscription (payments via Stripe). Users must use the app as intended.\n\nFull version: https://petsfollow.ll-it-sc.be/legal/terms\n\nLast updated: July 2026';

  @override
  String get legalPrivacyBody =>
      'Privacy policy — petsFollow\n\nData collected: identity (first name, email), pet data (name, species, breed, photos), heart rate readings (animal health data), messages and media with the practice, notification tokens (FCM), payment data processed by Stripe.\n\nPurposes: account management, cardiac monitoring, vet messaging, notifications, billing.\n\nProcessors / partners: Google (Sign-In, Firebase Cloud Messaging), Stripe (payments), cloud hosting (GCP).\n\nRetention: until account deletion or 3 years of inactivity.\n\nGDPR rights (access, rectification, deletion): Profile → Delete account, or support@ll-it-sc.be.\n\nFull version: https://petsfollow.ll-it-sc.be/legal/privacy\n\nLast updated: July 2026';

  @override
  String get legalNoticeBody =>
      'Legal notice — petsFollow\n\nPublisher: LL-IT-SC / petsFollow\nContact: support@ll-it-sc.be\n\nHosting: Google Cloud Platform (GDPR-compliant).\n\nPublication director: petsFollow.\n\nFull version: https://petsfollow.ll-it-sc.be/legal/mentions\n\nLast updated: July 2026';

  @override
  String get language => 'Language';

  @override
  String get languageFr => 'Français';

  @override
  String get languageNl => 'Nederlands';

  @override
  String get languageEn => 'English';

  @override
  String get languageEs => 'Español';

  @override
  String get planAnnualLabel => '€35 / year';

  @override
  String get planTriennialLabel => '€95 / 3 years';

  @override
  String get planQuinquennialLabel => '€145 / 5 years';

  @override
  String get pushNewMessage => 'New message';

  @override
  String get pushVisitConfirmed => 'Appointment confirmed';

  @override
  String get pushVisitProposed => 'Appointment proposal';

  @override
  String get pushVisitReschedule => 'Appointment reschedule';

  @override
  String get notifChannelMessages => 'Messages';

  @override
  String get notifChannelVisits => 'Visits';

  @override
  String get notifChannelCare => 'Care';

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
  String get errorNetwork =>
      'Unable to connect. Check your network and try again.';

  @override
  String get retryAction => 'Retry';

  @override
  String get errorMediaTooLarge => 'File too large (25 MB max)';

  @override
  String get errorInvalidMediaType =>
      'Unsupported format (JPEG, PNG, WebP, MP4, MOV, WebM)';

  @override
  String get errorPaymentRequired => 'Subscription required to send media';

  @override
  String get errorPhotoUploadFailed =>
      'Pet created, but the photo could not be uploaded';

  @override
  String get errorCouldNotOpenLink => 'Could not open the link';

  @override
  String planAnnualSub(String price) {
    return '$price, auto-renewed';
  }

  @override
  String get planTriennialSub => '€95 every 3 years, auto-renewed';

  @override
  String get planQuinquennialSub => '€145 for 5 years, one-time payment';

  @override
  String planOneTime(String price) {
    return '$price, one-time payment';
  }

  @override
  String get heartRateInstructions =>
      'Tap on each beat for the duration set by your veterinarian.';

  @override
  String heartRateInstructionsDuration(int seconds) {
    return 'Tap on each beat for $seconds seconds.';
  }

  @override
  String get heartRateNoDurationConfigured =>
      'No measurement duration is configured for this practice. Contact your veterinarian.';

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
  String get calendarBookingDisabled =>
      'Online booking is not available for this practice. You can still send a request without a time slot.';

  @override
  String get calendarBookingDisabledReschedule =>
      'Online booking is not available. Propose a date manually.';

  @override
  String get calendarNoSlots => 'No slots available in the next 14 days.';

  @override
  String get calendarPickSlot => 'Pick a slot:';

  @override
  String get visitConfirm => 'Confirm';

  @override
  String get visitProposeReschedule => 'Propose another time';

  @override
  String get visitRescheduleProposed => 'Reschedule proposal sent';

  @override
  String get paymentSuccessSnack => 'Payment received — refreshing…';

  @override
  String get paymentCancelSnack => 'Payment cancelled';

  @override
  String get visitRejectReschedule => 'Decline reschedule';

  @override
  String get visitAcceptReschedule => 'Accept new time';

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
  String get carePlusUpsell => 'Care+ — medications and personalized reminders';

  @override
  String get carePlusRequired =>
      'Care+ is required for medications and custom reminders.';

  @override
  String get horsePackRequired =>
      'The Horse pack is required for farrier reminders, contacts and competitions.';

  @override
  String get activateAddon => 'Activate';

  @override
  String get careTypeMedication => 'Medication';

  @override
  String get horseAddContact => 'Add a contact';

  @override
  String get horseAddCompetition => 'Add a competition';

  @override
  String get horseContactName => 'Name';

  @override
  String get horseContactRole => 'Role';

  @override
  String get horseCompetitionTitle => 'Event';

  @override
  String get horseCompetitionDate => 'Date (YYYY-MM-DD)';

  @override
  String get familyPackHint =>
      'Family pack — household care view, −10% from the 2nd paying pet plan';

  @override
  String familyHouseholdTitle(int count) {
    return 'Family household — $count pets';
  }

  @override
  String kennelHouseholdTitle(int count) {
    return 'Kennel household — $count pets';
  }

  @override
  String get familyHouseholdNext => 'Upcoming household reminders';

  @override
  String get familyPetLimit =>
      'A household pack is already active or being purchased';

  @override
  String get familyRequiresTwoPets => 'Family pack requires at least 2 pets';

  @override
  String get kennelPackHint => 'Kennel pack — 6+ pets, −15% on next pet plans';

  @override
  String get kennelRequiresSixPets => 'Kennel pack requires at least 6 pets';

  @override
  String get kennelQuickEncodeTitle => 'Litter quick encode';

  @override
  String get kennelRequired => 'Kennel pack is required for batch encoding';

  @override
  String get litterTag => 'Litter tag';

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
  String get careAddReminder => 'Add a reminder';

  @override
  String get careSelectPet => 'Pet';

  @override
  String careDueInDays(int days) {
    return 'Due in $days days';
  }

  @override
  String get careReferenceModeDone => 'Already done';

  @override
  String get careReferenceModeFirst => 'First time';

  @override
  String get careLastDateLabel => 'Reference date';

  @override
  String get careLastDateDone => 'Date of last care';

  @override
  String get careLastDateFirst => 'Cycle start date';

  @override
  String get careRecurrenceLabel => 'Recurrence';

  @override
  String get careRecurrenceNone => 'None (one-off due date)';

  @override
  String careRecurrenceDays(int days) {
    return 'Every $days days';
  }

  @override
  String get careDueDateLabel => 'Due date';

  @override
  String get careDueDateComputed => 'Computed due date';

  @override
  String get careTooltipDoneWithRecurrence =>
      'Already done: due date = last care date + recurrence.';

  @override
  String get careTooltipFirstWithRecurrence =>
      'First schedule: enter the cycle start date. Due date = that date + recurrence.';

  @override
  String get careTooltipNoRecurrence =>
      'No recurrence: the date you enter is the single due date.';

  @override
  String get careTooltipDueExplained =>
      'Due date = reference date + recurrence (when set).';

  @override
  String get carePickDate => 'Pick a date';

  @override
  String discoveryDayBadge(int day) {
    return 'D$day';
  }

  @override
  String get timelineTypeHeartrate => 'Heart rate';

  @override
  String get timelineTypeMessage => 'Message';

  @override
  String get timelineTypeCare => 'Care';

  @override
  String get timelineTypeVisit => 'Visit';

  @override
  String get timelineTypeEvent => 'Event';

  @override
  String get visitCancelAction => 'Cancel request';

  @override
  String get upcomingVisits => 'Upcoming visits';

  @override
  String get timelineEmpty => 'No events yet';

  @override
  String get noThreads => 'No conversations';

  @override
  String get vetInviteSent =>
      'Invitation sent — the practice must accept the request';

  @override
  String vetInviteSentNamed(String practice) {
    return 'Request sent to $practice — the practice must accept it';
  }

  @override
  String get vetNotFound => 'No veterinarian found with this email';

  @override
  String get addVetSearchHint =>
      'We look up this veterinarian account in petsFollow. If it exists, a link request is sent to the practice.';

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
  String get visitStatusReschedulePending => 'Reschedule pending';

  @override
  String get horseHealthTitle => 'Horse health';

  @override
  String get horseContactsTitle => 'Contacts (farrier, dentist…)';

  @override
  String get horseCompetitionsTitle => 'Competitions';

  @override
  String get horseContactsSoon =>
      'Activate the Horse pack to manage professional contacts.';

  @override
  String get horseCompetitionsSoon =>
      'Activate the Horse pack for the competition calendar.';

  @override
  String get horsePackUpsell =>
      'Horse pack — farrier, fecal egg count, contacts and competitions';

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
