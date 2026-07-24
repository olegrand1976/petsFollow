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
  String get emailNotVerified =>
      'Bevestig eerst uw e-mail (link bij registratie), log daarna opnieuw in.';

  @override
  String get loginOr => 'of';

  @override
  String get loginWithGoogle => 'Doorgaan met Google';

  @override
  String get twoFaTitle => '2FA-verificatie';

  @override
  String get twoFaSubtitle =>
      'Voer de 6-cijferige code van uw authenticator-app in.';

  @override
  String get twoFaCode => 'Authenticatorcode';

  @override
  String get twoFaSubmit => 'Bevestigen';

  @override
  String get twoFaBack => 'Terug naar inloggen';

  @override
  String get twoFaInvalid => 'Ongeldige of verlopen 2FA-code';

  @override
  String get forgotPassword => 'Wachtwoord vergeten?';

  @override
  String get forgotPasswordTitle => 'Wachtwoord vergeten';

  @override
  String get forgotPasswordSubtitle =>
      'Geef het e-mailadres van uw account op. Als er een account bestaat, wordt een resetlink verzonden.';

  @override
  String get forgotPasswordSubmit => 'Link versturen';

  @override
  String get forgotPasswordBack => 'Terug naar inloggen';

  @override
  String get forgotPasswordFailed => 'Verzenden mislukt';

  @override
  String get forgotPasswordSentTitle => 'E-mail verzonden';

  @override
  String forgotPasswordSent(String email) {
    return 'Als er een account bestaat voor $email, is er een link verzonden. Open deze in uw browser om een nieuw wachtwoord te kiezen.';
  }

  @override
  String get emailRequired => 'Voer een geldig e-mailadres in';

  @override
  String get resetPasswordTitle => 'Nieuw wachtwoord';

  @override
  String get resetPasswordSubtitle => 'Minimaal 8 tekens.';

  @override
  String get resetPasswordToken => 'Resettoken';

  @override
  String get resetPasswordSubmit => 'Opslaan';

  @override
  String get resetPasswordBackToLogin => 'Naar inloggen';

  @override
  String get resetPasswordInvalidLink => 'Ongeldige resetlink';

  @override
  String get resetPasswordFailed => 'Wachtwoord resetten mislukt';

  @override
  String get resetPasswordDoneTitle => 'Wachtwoord bijgewerkt';

  @override
  String get resetPasswordDoneSubtitle => 'U kunt nu inloggen.';

  @override
  String get fullName => 'Volledige naam';

  @override
  String get registerCta => 'Registreren';

  @override
  String get registerTitle => 'Registreren';

  @override
  String get registerSubtitle =>
      'Maak een account om de gezondheid van uw dier te volgen. U ontvangt een bevestigingsmail.';

  @override
  String get registerSubmit => 'Registreren';

  @override
  String get registerSuccess =>
      'Account aangemaakt. Open de link in de bevestigingsmail en kom terug om in te loggen.';

  @override
  String get registerFailed => 'Registratie mislukt';

  @override
  String get registerEmailExists => 'Dit e-mailadres is al in gebruik';

  @override
  String get registerBackToLogin => 'Terug naar inloggen';

  @override
  String get confirmEmailTitle => 'E-mail bevestigen';

  @override
  String get confirmEmailLoading => 'Bevestigen…';

  @override
  String get confirmEmailDoneTitle => 'E-mail bevestigd';

  @override
  String get confirmEmailDoneSubtitle =>
      'Uw account is actief. U kunt inloggen.';

  @override
  String get confirmEmailFailedTitle => 'Bevestiging mislukt';

  @override
  String get confirmEmailFailed => 'Deze e-mail kon niet worden bevestigd.';

  @override
  String get confirmEmailInvalidLink =>
      'Ongeldige of al gebruikte bevestigingslink.';

  @override
  String get confirmEmailBackToLogin => 'Terug naar inloggen';

  @override
  String get vetUseProWeb =>
      'Volledige dierenartsaccounts gebruiken de Pro-website.';

  @override
  String get unsupportedRoleApp =>
      'Dit account kan niet in de pets-app. Gebruik de Pro-website.';

  @override
  String get proLightTitle => 'Pro terrein';

  @override
  String get proLightAgenda => 'Agenda';

  @override
  String get proLightClients => 'Klanten';

  @override
  String get proLightPets => 'Dieren';

  @override
  String get proLightLoadError => 'Laden mislukt';

  @override
  String get proLightNoVisits => 'Geen afspraken';

  @override
  String get proLightTourToday => 'Vandaag';

  @override
  String get proLightTourWeek => '7 dagen';

  @override
  String get proLightTourAll => 'Alles';

  @override
  String get proLightNoTourToday => 'Geen afspraken vandaag';

  @override
  String get proLightNoTourWeek => 'Geen afspraken in de komende 7 dagen';

  @override
  String get proLightNoClients => 'Geen gedeelde klanten';

  @override
  String get proLightNoPets => 'Geen gedeelde dieren';

  @override
  String get proLightAddress => 'Adres';

  @override
  String get proLightOpenMaps => 'Maps';

  @override
  String get proLightReportTitle => 'Verslag';

  @override
  String get proLightReportHint => 'Bezoeknotities…';

  @override
  String get proLightImproveAi => 'Verbeteren (IA)';

  @override
  String get proLightFinalizeReport => 'Afronden';

  @override
  String get proLightReportFinal => 'Afgerond';

  @override
  String get proLightReportHistoryTitle => 'Geschiedenis';

  @override
  String get proLightReportHistoryTranscript => 'Origineel (transcriptie)';

  @override
  String get proLightReportHistoryImproved => 'IA-versie';

  @override
  String get proLightReportHistorySaved => 'Opgeslagen versie';

  @override
  String get proLightReportHistoryEmpty => 'Geen versie beschikbaar';

  @override
  String get proLightSettings => 'Instellingen';

  @override
  String get proLightSpecialty => 'Specialiteit';

  @override
  String get proLightDocuments => 'Documenten';

  @override
  String get proLightNoDocuments => 'Geen documenten';

  @override
  String get proLightTimeline => 'Tijdlijn';

  @override
  String get proLightNoTimeline => 'Geen gebeurtenissen';

  @override
  String get proLightReminders => 'Herinneringen';

  @override
  String get proLightNoReminders => 'Geen herinneringen';

  @override
  String get proLightLitterTag => 'Nest / tag';

  @override
  String get proLightActionFailed => 'Actie mislukt';

  @override
  String get proLightReadOnly => 'Alleen-lezen toegang';

  @override
  String get petAccessSharedRead => 'Gedeeld · lezen';

  @override
  String get petAccessSharedNotes => 'Gedeeld · notities';

  @override
  String get petAccessSharedFull => 'Gedeeld · volledig';

  @override
  String get proLightUseGps => 'GPS';

  @override
  String get proLightTranscribeAudio => 'Audiobestand';

  @override
  String get proLightDictationStart => 'Dicteren';

  @override
  String get proLightDictationStop => 'Stop & transcriberen';

  @override
  String get proLightAudioConsentTitle => 'Audiotoestemming';

  @override
  String get proLightAudioConsentBody =>
      'De opname dient alleen om het verslag te maken. Ze wordt verwijderd bij finalisatie.';

  @override
  String get proLightAudioConsentAccept => 'Ik ga akkoord';

  @override
  String get proLightSpecialtyFarrier => 'Hoefsmid';

  @override
  String get proLightSpecialtyPhysio => 'Fysio / osteo';

  @override
  String get proLightSpecialtyBehaviorist => 'Gedragstherapeut';

  @override
  String get proLightSpecialtyVetLight => 'Dierenarts light';

  @override
  String get proLightReportHintFarrier =>
      'Beslagverslag: hoeven, ijzer, observaties…';

  @override
  String get proLightEmptyFarrier => 'Geen gedeeld paard / bezoek';

  @override
  String get proLightMicDenied =>
      'Microfoon geweigerd — geef toegang in de instellingen';

  @override
  String get proLightGpsDenied => 'Locatie niet beschikbaar';

  @override
  String get googleNotConfigured => 'Google-aanmelding is niet geconfigureerd';

  @override
  String get googleLoginFailed => 'Google-aanmelding mislukt';

  @override
  String get googleWrongAudience => 'Dit Google-account is geen klantprofiel';

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
  String get confirmNewPassword => 'Bevestig wachtwoord';

  @override
  String get changePassword => 'Wachtwoord wijzigen';

  @override
  String get forceChangePasswordTitle => 'Wachtwoord wijzigen';

  @override
  String get forceChangePasswordSubtitle =>
      'Dit account is aangemaakt met een tijdelijk wachtwoord. Kies uw eigen wachtwoord om verder te gaan.';

  @override
  String get forceChangePasswordSubmit => 'Opslaan en doorgaan';

  @override
  String get passwordTooShort => 'Minimaal 8 tekens';

  @override
  String get passwordMismatch => 'Wachtwoorden komen niet overeen';

  @override
  String get passwordChangeFailed => 'Wachtwoord wijzigen mislukt';

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
  String get choosePetForMeasurement => 'Kies een huisdier';

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
  String get legalOpenOnline => 'Onlineversie bekijken';

  @override
  String get legalTermsBody =>
      'Gebruiksvoorwaarden — petsFollow\n\nDe petsFollow-app laat eigenaars de hartslag meten, de geschiedenis bekijken en communiceren met hun dierenarts.\n\nDiensten worden geleverd in het kader van het gekozen abonnement (betalingen via Stripe).\n\nVolledige versie: https://petsfollow.ll-it-sc.be/legal/terms\n\nLaatst bijgewerkt: juli 2026';

  @override
  String get legalPrivacyBody =>
      'Privacybeleid — petsFollow\n\nVerzamelde gegevens: identiteit (voornaam, e-mail), huisdiergegevens (naam, soort, ras, foto\'s), hartslagmetingen (diergezondheidsgegevens), berichten en media met de praktijk, notificatietokens (FCM), betalingsgegevens via Stripe.\n\nDoeleinden: accountbeheer, hartmonitoring, dierenartsberichten, notificaties, facturatie.\n\nVerwerkers / partners: Google (Sign-In, Firebase Cloud Messaging), Stripe (betalingen), cloudhosting (GCP).\n\nBewaring: tot verwijdering van het account of 3 jaar inactiviteit.\n\nAVG-rechten: Profiel → Account verwijderen, of support@ll-it-sc.be.\n\nVolledige versie: https://petsfollow.ll-it-sc.be/legal/privacy\n\nLaatst bijgewerkt: juli 2026';

  @override
  String get legalNoticeBody =>
      'Juridische vermeldingen — petsFollow\n\nUitgever: LL-IT-SC / petsFollow\nContact: support@ll-it-sc.be\n\nHosting: Google Cloud Platform (AVG-conform).\n\nVolledige versie: https://petsfollow.ll-it-sc.be/legal/mentions\n\nLaatst bijgewerkt: juli 2026';

  @override
  String get language => 'Taal';

  @override
  String get languageFr => 'Français';

  @override
  String get languageNl => 'Nederlands';

  @override
  String get languageEn => 'English';

  @override
  String get languageEs => 'Español';

  @override
  String get languageEt => 'Eesti';

  @override
  String get planMonthlyLabel => '3,50 € / maand';

  @override
  String get planAnnualLabel => '35 € / jaar';

  @override
  String get planTriennialLabel => '95 € / 3 jaar';

  @override
  String get planQuinquennialLabel => '145 € / 5 jaar';

  @override
  String get pushNewMessage => 'Nieuw bericht';

  @override
  String get pushVisitConfirmed => 'Afspraak bevestigd';

  @override
  String get pushVisitProposed => 'Afspraakvoorstel';

  @override
  String get pushVisitReschedule => 'Afspraak verplaatst';

  @override
  String get notifChannelMessages => 'Berichten';

  @override
  String get notifChannelVisits => 'Bezoeken';

  @override
  String get notifChannelCare => 'Zorgen';

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
  String get errorNetwork =>
      'Geen verbinding. Controleer uw netwerk en probeer opnieuw.';

  @override
  String get retryAction => 'Opnieuw proberen';

  @override
  String get errorMediaTooLarge => 'Bestand te groot (max. 25 MB)';

  @override
  String get errorInvalidMediaType =>
      'Niet-ondersteund formaat (JPEG, PNG, WebP, MP4, MOV, WebM)';

  @override
  String get errorPaymentRequired => 'Abonnement vereist om media te versturen';

  @override
  String get errorPhotoUploadFailed =>
      'Huisdier aangemaakt, maar de foto kon niet worden geüpload';

  @override
  String get errorCouldNotOpenLink => 'Link kon niet worden geopend';

  @override
  String get planMonthlySub => '3,50 € / maand, automatisch verlengd';

  @override
  String planAnnualSub(String price) {
    return '$price, automatisch verlengd';
  }

  @override
  String get planTriennialSub => '95 € elke 3 jaar, automatisch verlengd';

  @override
  String get planQuinquennialSub => '145 € voor 5 jaar, eenmalige betaling';

  @override
  String planOneTime(String price) {
    return '$price, eenmalige betaling';
  }

  @override
  String get heartRateInstructions =>
      'Tik bij elke hartslag gedurende de tijd die uw dierenarts heeft ingesteld.';

  @override
  String heartRateInstructionsDuration(int seconds) {
    return 'Tik bij elke hartslag gedurende $seconds seconden.';
  }

  @override
  String get heartRateNoDurationConfigured =>
      'Er is geen meetduur geconfigureerd voor deze praktijk. Neem contact op met uw dierenarts.';

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
  String get heartRateCommentLabel => 'Opmerking (optioneel)';

  @override
  String get heartRateCommentHint => 'Bv. onrustig, in rust, na inspanning…';

  @override
  String get restart => 'Opnieuw beginnen';

  @override
  String get sentToVet => 'Meting naar dierenarts gestuurd';

  @override
  String get navHome => 'Home';

  @override
  String get navPets => 'Dieren';

  @override
  String get navCare => 'Zorg';

  @override
  String get navMessages => 'Berichten';

  @override
  String get navProfile => 'Profiel';

  @override
  String get speciesDog => 'Hond';

  @override
  String get speciesCat => 'Kat';

  @override
  String get speciesHorse => 'Paard';

  @override
  String get speciesOther => 'Anders';

  @override
  String get careComingSoon => 'Zorgherinneringen komen binnenkort';

  @override
  String get emptyPetsTitle => 'Geen huisdieren';

  @override
  String get emptyPetsBody =>
      'Voeg uw eerste huisdier toe om te beginnen met hartslagmonitoring bij uw dierenarts.';

  @override
  String get discoveryTitle => 'Ontdek petsFollow';

  @override
  String get discoveryMission => 'Uw 7-daagse traject';

  @override
  String get discoveryDay0Title => 'Dag 0 — Welkom';

  @override
  String get discoveryDay0Body =>
      'Maak het profiel van uw huisdier aan en leer hoe u de hartslag meet.';

  @override
  String get discoveryDay2Title => 'Dag 2 — Eerste meting';

  @override
  String get discoveryDay2Body =>
      'Doe uw eerste hartslagmeting en oefen de techniek.';

  @override
  String get discoveryDay4Title => 'Dag 4 — Routine';

  @override
  String get discoveryDay4Body =>
      'Bouw een dagelijkse meetroutine op met gepersonaliseerde herinneringen.';

  @override
  String get discoveryDay6Title => 'Dag 6 — Delen met dierenarts';

  @override
  String get discoveryDay6Body =>
      'Uw metingen worden gedeeld met uw dierenarts voor optimale opvolging.';

  @override
  String get myVets => 'Mijn dierenartsen';

  @override
  String get addVetByEmail => 'Dierenarts toevoegen via e-mail';

  @override
  String get vetEmailHint => 'email@praktijk.vet';

  @override
  String get noVets => 'Geen gekoppelde dierenarts';

  @override
  String get primaryVet => 'Hoofddierenarts';

  @override
  String get setPrimaryVet => 'Instellen als hoofddierenarts';

  @override
  String get careTitle => 'Zorg';

  @override
  String get careDone => 'Gedaan';

  @override
  String get carePostpone => 'Uitstellen';

  @override
  String get careOverdue => 'Te laat';

  @override
  String get visitHistory => 'Bezoekgeschiedenis';

  @override
  String get requestVisit => 'Bezoek aanvragen';

  @override
  String get calendarBookingDisabled =>
      'Online reserveren is niet beschikbaar voor deze praktijk. Bel de praktijk om een afspraak te maken.';

  @override
  String get calendarBookingDisabledReschedule =>
      'Online reserveren is niet beschikbaar. Stel handmatig een datum voor.';

  @override
  String get calendarNoSlots =>
      'Geen slots beschikbaar in de komende 14 dagen.';

  @override
  String get calendarPickSlot => 'Kies een slot:';

  @override
  String get calendarSelectVet => 'Kies een dierenarts:';

  @override
  String get calendarCallPractice => 'Bel de praktijk';

  @override
  String get calendarNoPhone =>
      'Er is geen telefoonnummer voor deze praktijk. Neem op een andere manier contact op.';

  @override
  String get visitConfirm => 'Bevestigen';

  @override
  String get visitProposeReschedule => 'Ander moment voorstellen';

  @override
  String get visitRescheduleProposed => 'Verplaatsingsvoorstel verzonden';

  @override
  String get paymentSuccessSnack => 'Betaling ontvangen — vernieuwen…';

  @override
  String get paymentCancelSnack => 'Betaling geannuleerd';

  @override
  String get visitRejectReschedule => 'Verplaatsing weigeren';

  @override
  String get visitAcceptReschedule => 'Nieuw moment aanvaarden';

  @override
  String get upcomingVisit => 'Komend bezoek';

  @override
  String get notificationHrTitle => 'petsFollow';

  @override
  String get notificationHrBody =>
      'Tijd voor een hartslagmeting van uw huisdier';

  @override
  String get reviewAskTitle => 'Bevalt petsFollow?';

  @override
  String get reviewAskYes => 'Ja, app beoordelen';

  @override
  String get reviewAskNo => 'Later';

  @override
  String get carePlusUpsell =>
      'Care+ — medicatie en gepersonaliseerde herinneringen';

  @override
  String get carePlusRequired =>
      'Care+ is vereist voor medicatie en aangepaste herinneringen.';

  @override
  String get horsePackRequired =>
      'Het paardenpakket is vereist voor hoefsmid, contacten en wedstrijden.';

  @override
  String get activateAddon => 'Activeren';

  @override
  String get careTypeMedication => 'Medicatie';

  @override
  String get horseAddContact => 'Contact toevoegen';

  @override
  String get horseAddCompetition => 'Wedstrijd toevoegen';

  @override
  String get horseContactName => 'Naam';

  @override
  String get horseContactRole => 'Rol';

  @override
  String get horseCompetitionTitle => 'Evenement';

  @override
  String get horseCompetitionDate => 'Datum (JJJJ-MM-DD)';

  @override
  String get familyPackHint =>
      'Familiepakket — gezinsweergave, −10% vanaf het 2e betalende dierabonnement';

  @override
  String familyHouseholdTitle(int count) {
    return 'Familiehuishouden — $count dieren';
  }

  @override
  String kennelHouseholdTitle(int count) {
    return 'Kennelhuishouden — $count dieren';
  }

  @override
  String get familyHouseholdNext => 'Komende gezinsherinneringen';

  @override
  String get familyPetLimit =>
      'Er is al een huishoudpakket actief of in aankoop';

  @override
  String get familyRequiresTwoPets => 'Familiepakket vereist minstens 2 dieren';

  @override
  String get kennelPackHint =>
      'Kennelpakket — 6+ dieren, −15% op volgende abonnementen';

  @override
  String get kennelRequiresSixPets => 'Kennelpakket vereist minstens 6 dieren';

  @override
  String get kennelQuickEncodeTitle => 'Nest snel encoderen';

  @override
  String get kennelRequired => 'Kennelpakket is vereist voor batch-encoding';

  @override
  String get litterTag => 'Nest-tag';

  @override
  String get discoveryMarkDone => 'Missie voltooid';

  @override
  String get notificationPreferences => 'Meldingsvoorkeuren';

  @override
  String get notificationPrefsHint =>
      'Kies welke meldingstypes u wilt ontvangen.';

  @override
  String get notificationPrefsSaved => 'Voorkeuren opgeslagen';

  @override
  String get notificationPrefHr => 'Hartslagmetingen';

  @override
  String get notificationPrefCare => 'Zorgherinneringen';

  @override
  String get notificationPrefVisits => 'Bezoeken';

  @override
  String get notificationPrefMessages => 'Berichten';

  @override
  String get notificationPrefDiscovery => 'Ontdekkingsreis';

  @override
  String get notificationPrefBilling => 'Facturering';

  @override
  String carePostponeDays(int days) {
    return 'Uitstellen met $days dagen';
  }

  @override
  String get noCareReminders => 'Geen openstaande zorgherinneringen';

  @override
  String get careAddReminder => 'Herinnering toevoegen';

  @override
  String get careSelectPet => 'Huisdier';

  @override
  String careDueInDays(int days) {
    return 'Vervalt over $days dagen';
  }

  @override
  String get careReferenceModeDone => 'Al uitgevoerd';

  @override
  String get careReferenceModeFirst => 'Eerste keer';

  @override
  String get careLastDateLabel => 'Referentiedatum';

  @override
  String get careLastDateDone => 'Datum van laatste zorg';

  @override
  String get careLastDateFirst => 'Startdatum van de cyclus';

  @override
  String get careRecurrenceLabel => 'Herhaling';

  @override
  String get careRecurrenceNone => 'Geen (eenmalige deadline)';

  @override
  String careRecurrenceDays(int days) {
    return 'Elke $days dagen';
  }

  @override
  String get careDueDateLabel => 'Deadline';

  @override
  String get careDueDateComputed => 'Berekende deadline';

  @override
  String get careTooltipDoneWithRecurrence =>
      'Al uitgevoerd: deadline = datum van laatste zorg + herhaling.';

  @override
  String get careTooltipFirstWithRecurrence =>
      'Eerste planning: geef de startdatum van de cyclus. Deadline = die datum + herhaling.';

  @override
  String get careTooltipNoRecurrence =>
      'Zonder herhaling: de ingevoerde datum is de enige deadline.';

  @override
  String get careTooltipDueExplained =>
      'Deadline = referentiedatum + herhaling (indien ingesteld).';

  @override
  String get carePickDate => 'Kies een datum';

  @override
  String discoveryDayBadge(int day) {
    return 'D$day';
  }

  @override
  String get timelineTypeHeartrate => 'Hartslag';

  @override
  String get timelineTypeMessage => 'Bericht';

  @override
  String get timelineTypeCare => 'Zorg';

  @override
  String get timelineTypeVisit => 'Bezoek';

  @override
  String get timelineTypeEvent => 'Gebeurtenis';

  @override
  String get visitCancelAction => 'Aanvraag annuleren';

  @override
  String get upcomingVisits => 'Komende bezoeken';

  @override
  String get timelineEmpty => 'Nog geen gebeurtenissen';

  @override
  String get noThreads => 'Geen gesprekken';

  @override
  String get vetInviteSent =>
      'Uitnodiging verzonden — de praktijk moet de aanvraag aanvaarden';

  @override
  String vetInviteSentNamed(String practice) {
    return 'Aanvraag verzonden naar $practice — de praktijk moet ze aanvaarden';
  }

  @override
  String get vetNotFound => 'Geen dierenarts gevonden met dit e-mailadres';

  @override
  String get addVetSearchHint =>
      'We zoeken dit dierenartsaccount in petsFollow. Als het bestaat, wordt een koppelingsaanvraag naar de praktijk gestuurd.';

  @override
  String get visitRequested => 'Bezoekaanvraag verzonden';

  @override
  String get primaryVetSet => 'Hoofddierenarts bijgewerkt';

  @override
  String get visitStatusRequested => 'Aangevraagd';

  @override
  String get visitStatusConfirmed => 'Bevestigd';

  @override
  String get visitStatusDone => 'Afgerond';

  @override
  String get visitStatusCancelled => 'Geannuleerd';

  @override
  String get visitStatusReschedulePending => 'Verplaatsing in afwachting';

  @override
  String get horseHealthTitle => 'Paardengezondheid';

  @override
  String get horseContactsTitle => 'Contacten (hoefsmid, tandarts…)';

  @override
  String get horseCompetitionsTitle => 'Wedstrijden';

  @override
  String get horseContactsSoon =>
      'Activeer het paardenpakket om contacten te beheren.';

  @override
  String get horseCompetitionsSoon =>
      'Activeer het paardenpakket voor de wedstrijdkalender.';

  @override
  String get horsePackUpsell =>
      'Paardenpakket — hoefsmid, mestonderzoek, contacten en wedstrijden';

  @override
  String get careTypeFarrier => 'Hoefsmid';

  @override
  String get careTypeFecalEgg => 'Mestonderzoek';

  @override
  String get careTypeVaccination => 'Vaccinatie';

  @override
  String get careTypeDeworming => 'Ontworming';

  @override
  String get careTypeVetCheck => 'Dierenartscontrole';

  @override
  String get careTypeDental => 'Gebitsverzorging';

  @override
  String get careTypeCustom => 'Aangepaste herinnering';

  @override
  String get homeAddFirstVetTitle => 'Voeg uw dierenarts toe';

  @override
  String get homeAddFirstVetBody =>
      'Koppel de praktijk die uw dier volgt om metingen te delen en te chatten.';

  @override
  String get homeAddFirstVetCta => 'Dierenarts toevoegen';

  @override
  String get photoFrameHint => 'Centreer de snuit — voorvertoning dierenfiche';

  @override
  String get takePhoto => 'Foto maken';

  @override
  String get chooseFromGallery => 'Kiezen uit galerij';

  @override
  String get attachMedia => 'Foto of video toevoegen';

  @override
  String get attachPhoto => 'Foto';

  @override
  String get attachVideo => 'Video';

  @override
  String get openMedia => 'Openen';

  @override
  String get mediaVideoLabel => 'Video';

  @override
  String get appInviteTitle => 'App-uitnodiging QR';

  @override
  String get appInviteHint =>
      'Toon deze QR of deel de link. Een nieuwe cliënt die via deze link registreert, wordt automatisch gekoppeld.';

  @override
  String get appInviteHintShort => 'Download- en koppelingslink';

  @override
  String get appInviteCodeLabel => 'Code:';

  @override
  String get appInviteCopy => 'Link kopiëren';

  @override
  String get appInviteCopied => 'Link gekopieerd';

  @override
  String get appInviteLoadError => 'QR laden mislukt';

  @override
  String get appInviteRetry => 'Opnieuw';

  @override
  String get proLightVetTitle => 'Veld véto';

  @override
  String get commercialFieldTitle => 'Commercial';

  @override
  String get commercialFieldSubtitle =>
      'QR-uitnodiging clients en toegang tot de Pro-site.';

  @override
  String get commercialOpenProWeb => 'Pro-site openen';
}
