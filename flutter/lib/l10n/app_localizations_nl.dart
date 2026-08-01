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
  String get resendConfirmation => 'Bevestigingsmail opnieuw versturen';

  @override
  String get resendConfirmationSent =>
      'Als het account bestaat en nog niet bevestigd is, is er een nieuwe e-mail verzonden.';

  @override
  String get resendConfirmationFailed =>
      'Verzenden mislukt. Probeer het zo opnieuw.';

  @override
  String get loginOr => 'of';

  @override
  String get loginWithGoogle => 'Doorgaan met Google';

  @override
  String get loginWithApple => 'Doorgaan met Apple';

  @override
  String get appleComingSoon => 'Inloggen met Apple komt binnenkort.';

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
  String get registerInviteNotApplied =>
      'Account aangemaakt, maar de uitnodigingscode kon niet worden toegepast. U kunt die na het inloggen opnieuw invoeren.';

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
  String get proLightNewConsultation => 'Nieuwe consultatie';

  @override
  String get proLightConsultationNextTitle =>
      'Consultatie opgeslagen — verder?';

  @override
  String get proLightConsultationCtaDaf => 'DAF aanmaken & factureren';

  @override
  String get proLightConsultationCtaInvoice => 'Direct factureren';

  @override
  String get proLightConsultationCtaDone => 'Afronden';

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
  String get proLightDictationStop => 'Stoppen';

  @override
  String get proLightRecordingInProgress => 'Opname bezig';

  @override
  String get proLightAudioConsentTitle => 'Audiotoestemming';

  @override
  String get proLightAudioConsentBody =>
      'De opname dient alleen om het verslag te maken. Bevestig de mondelinge toestemming van de klant. Audio wordt verwijderd bij finalisatie.';

  @override
  String get proLightAudioConsentAccept => 'Ik ga akkoord';

  @override
  String get proLightSpecialtyFarrier => 'Hoefsmid';

  @override
  String get proLightSpecialtyPhysio => 'Fysio / osteo';

  @override
  String get proLightSpecialtyBehaviorist => 'Gedragstherapeut';

  @override
  String get proLightSpecialtyGroomer => 'Trimmer';

  @override
  String get proLightSpecialtyBreeder => 'Fokker';

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
  String get googleWrongAudience =>
      'Dit Google-account is al een Pro-profiel — gebruik de webapp.';

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
  String get acceptTermsTitle => 'Gebruiksvoorwaarden';

  @override
  String get acceptTermsSubtitle =>
      'Uw account is aangemaakt door uw praktijk. Aanvaard de voorwaarden en het privacybeleid om verder te gaan.';

  @override
  String get acceptTermsSubmit => 'Aanvaarden en doorgaan';

  @override
  String get acceptTermsFailed =>
      'Toestemming kon niet worden opgeslagen. Probeer opnieuw.';

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
  String get exportMyData => 'Mijn gegevens exporteren';

  @override
  String get registerConsentPrefix => 'Ik aanvaard de ';

  @override
  String get registerConsentMiddle => ' en het ';

  @override
  String get registerConsentRequired =>
      'U moet de voorwaarden en het privacybeleid aanvaarden.';

  @override
  String get registerInviteCode => 'Uitnodigingscode (optioneel)';

  @override
  String get registerInviteCodeHint =>
      'Voer de code van de QR / commercieel link in';

  @override
  String get nearbyCommercialTitle => 'Vertegenwoordiger bij u in de buurt';

  @override
  String get nearbyCommercialHint =>
      'Zonder uitnodigingscode kunt u een nabije vertegenwoordiger kiezen (optioneel).';

  @override
  String get nearbyCommercialUseLocation => 'Mijn locatie gebruiken';

  @override
  String get nearbyCommercialPostalCode => 'Postcode';

  @override
  String get nearbyCommercialSearch => 'Zoeken';

  @override
  String get nearbyCommercialEmpty =>
      'Geen vertegenwoordiger in de buurt gevonden.';

  @override
  String get nearbyCommercialSkip => 'Overslaan';

  @override
  String get nearbyCommercialGeoDenied =>
      'Locatie geweigerd — voer een postcode in.';

  @override
  String nearbyCommercialDistance(String km) {
    return '$km km';
  }

  @override
  String get pushPermissionTitle => 'Notificaties';

  @override
  String get pushPermissionBody =>
      'petsFollow wil u notificaties sturen: berichten van uw dierenarts, afspraakbevestigingen en zorgherinneringen. U kunt ze op elk moment uitschakelen in de app- of telefooninstellingen.';

  @override
  String get pushPermissionContinue => 'Doorgaan';

  @override
  String exportDataSaved(String path) {
    return 'Export opgeslagen: $path';
  }

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
  String get heartRateShort => 'Hart';

  @override
  String get weightShort => 'Gewicht';

  @override
  String get recordWeightTitle => 'Gewicht registreren';

  @override
  String get weightKgLabel => 'Gewicht (kg)';

  @override
  String get weightCommentLabel => 'Opmerking (optioneel)';

  @override
  String get weightCommentHint => 'Bv. na wandeling, nuchter…';

  @override
  String get weightSave => 'Opslaan';

  @override
  String get weightSentToVet => 'Gewicht opgeslagen';

  @override
  String get weightInvalid => 'Geef een geldig gewicht op (0,01–999,99 kg)';

  @override
  String weightLastLabel(String kg) {
    return 'Laatste gewicht: $kg kg';
  }

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
      'Gebruiksvoorwaarden — petsFollow\n\nDe petsFollow-app biedt eigenaars voorgeschreven opvolging (berichten, Care/Horse-herinneringen, hartslagmetingen), geschiedenis en communicatie met hun dierenarts.\n\nDiensten worden geleverd in het kader van het gekozen abonnement (betalingen via Stripe).\n\nVolledige versie: https://petsfollow.ll-it-sc.be/legal/terms\n\nLaatst bijgewerkt: juli 2026';

  @override
  String get legalPrivacyBody =>
      'Privacybeleid — petsFollow\n\nVerzamelde gegevens: identiteit (voornaam, e-mail), huisdiergegevens (naam, soort, ras, foto\'s), hartslagmetingen (diergezondheidsgegevens), berichten en media met de praktijk, bezoekverslagen (tekst en audio-opnamen), GPS-coördinaten van huisbezoeken (zorgprofessionals), notificatietokens (FCM), betalingsgegevens via Stripe.\n\nDoeleinden: accountbeheer, zorgcontinuïteit (inclusief hartslagmetingen), dierenartsberichten, bezoekverslagen, notificaties, facturatie.\n\nAI-verwerking: Google Gemini wordt gebruikt om bezoekverslagen te verbeteren (audio in realtime verwerkt, niet bewaard door Google) en, in de cliënt-app (experimentele module), om afgeronde verslagen uit te leggen en conversatie-triage bij spoedgevallen te bieden.\n\nVerwerkers / partners: Google (Sign-In, Firebase Cloud Messaging, Gemini), Stripe (betalingen), cloudhosting (GCP).\n\nBewaring: tot verwijdering van het account; inactieve accounts na 3 jaar verwijderd; audio van bezoekverslagen bewaard zolang het dossier bestaat.\n\nAVG-rechten (inzage, rectificatie, wissing, overdraagbaarheid): Profiel → Mijn gegevens exporteren / Account verwijderen, of support@petsfollow.app.\n\nVolledige versie: https://petsfollow.ll-it-sc.be/legal/privacy\n\nLaatst bijgewerkt: juli 2026';

  @override
  String get legalNoticeBody =>
      'Juridische vermeldingen — petsFollow\n\nUitgever: LL-IT-SC / petsFollow\nContact: support@petsfollow.app\n\nHosting: Google Cloud Platform (AVG-conform).\n\nVolledige versie: https://petsfollow.ll-it-sc.be/legal/mentions\n\nLaatst bijgewerkt: juli 2026';

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
  String get languageIt => 'Italiano';

  @override
  String get appearance => 'Weergave';

  @override
  String get themeLight => 'Licht';

  @override
  String get themeDark => 'Donker';

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
  String get editPet => 'Huisdier bewerken';

  @override
  String get petName => 'Naam';

  @override
  String get petNameRequired => 'Geef de naam van het dier op';

  @override
  String get species => 'Soort';

  @override
  String get breed => 'Ras';

  @override
  String get petMicrochipOptional => 'Chipnummer (optioneel)';

  @override
  String get petHealthBookNumberOptional => 'Paspoortnummer (optioneel)';

  @override
  String get petDomicileLocation => 'Domicilie / stal';

  @override
  String get petDomicileHint => 'Bv. Stal De Wilgen — Brussel';

  @override
  String get petFoodChainStatus => 'Status voedselketen';

  @override
  String get petFoodChainCompanion => 'Gezelschapsdier (buiten voedselketen)';

  @override
  String get petFoodChainFoodProducing => 'Productiedier / voedselketen';

  @override
  String get petFoodChainExcluded => 'Uitgesloten van de voedselketen';

  @override
  String get petHealthBookAddPages => 'Foto\'s van het paspoort toevoegen';

  @override
  String get petHealthBookReplacePages => 'PDF vervangen (foto\'s)';

  @override
  String petHealthBookPagesCount(int count) {
    return '$count pagina(\'s) geselecteerd';
  }

  @override
  String get petHealthBookPdfAttached => 'PDF van het paspoort bijgevoegd';

  @override
  String get petHealthBookRemovePdf => 'Verwijderen';

  @override
  String get petHealthBookOpenPdf => 'Paspoort openen (PDF)';

  @override
  String petMicrochipLabel(String number) {
    return 'Chip: $number';
  }

  @override
  String petHealthBookNumberLabel(String number) {
    return 'Paspoort: $number';
  }

  @override
  String get errorHealthBookUploadFailed =>
      'Huisdier opgeslagen, maar het paspoort kon niet worden geüpload';

  @override
  String get choosePlan => 'Kies uw formule';

  @override
  String get recommended => 'Aanbevolen';

  @override
  String get autoRenewTitle => 'Automatisch verlengen';

  @override
  String get autoRenewSubtitle => 'Incasso bij elke vervaldatum';

  @override
  String get continueToPayment => 'Opslaan en betalen';

  @override
  String get petFormSave => 'Opslaan';

  @override
  String get petSavedPendingPayment =>
      'Huisdier opgeslagen — activeer het om functies te gebruiken';

  @override
  String get paymentFeaturesLocked =>
      'Betaling vereist om de functies van dit huisdier te gebruiken';

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
  String get errorPaymentRequired =>
      'Abonnement vereist om deze functie te gebruiken';

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
  String get heartRateNotSupported =>
      'Hartslagmeting is niet beschikbaar voor deze diersoort';

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
  String get thresholdAlert =>
      'Waarschuwing: significante stijging t.o.v. vorige meting';

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
  String get speciesDonkey => 'Ezel';

  @override
  String get speciesCattle => 'Rund';

  @override
  String get speciesSheep => 'Schaap';

  @override
  String get speciesGoat => 'Geit';

  @override
  String get speciesPig => 'Varken';

  @override
  String get speciesPoultry => 'Pluimvee';

  @override
  String get speciesRabbit => 'Konijn';

  @override
  String get speciesAlpaca => 'Alpaca';

  @override
  String get speciesLlama => 'Lama';

  @override
  String get speciesOther => 'Anders';

  @override
  String get careComingSoon => 'Zorgherinneringen komen binnenkort';

  @override
  String get emptyPetsTitle => 'Geen huisdieren';

  @override
  String get emptyPetsBody =>
      'Voeg uw eerste huisdier toe om te beginnen met voorgeschreven opvolging bij uw dierenarts.';

  @override
  String get discoveryTitle => 'Ontdek petsFollow';

  @override
  String get discoveryMission => 'Uw petsFollow-traject';

  @override
  String get discoveryDay0Title => 'Stap 1 — Welkom';

  @override
  String get discoveryDay0Body =>
      'Maak het profiel van uw huisdier aan en ontdek de app — berichten, herinneringen en metingen (inclusief hartslag).';

  @override
  String get discoveryDay2Title => 'Stap 2 — Eerste meting';

  @override
  String get discoveryDay2Body =>
      'Doe uw eerste hartslagmeting en oefen de techniek.';

  @override
  String get discoveryDay4Title => 'Stap 3 — Routine';

  @override
  String get discoveryDay4Body =>
      'Bouw een dagelijkse meetroutine op met gepersonaliseerde herinneringen.';

  @override
  String get discoveryDay6Title => 'Stap 4 — Delen met dierenarts';

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
  String get vetLinkRequired =>
      'Koppel een dierenarts om de opvolging met uw praktijk te activeren';

  @override
  String get linkVetAfterSaveTitle => 'Dierenarts koppelen';

  @override
  String get linkVetAfterSaveBody =>
      'Koppel een praktijk om berichten, bezoeken en zorgherinneringen te activeren.';

  @override
  String get linkVetHomeTitle => 'Dierenarts koppelen?';

  @override
  String get linkVetHomeBody =>
      'Uw dier is opgeslagen. Wilt u een dierenarts koppelen? Optioneel — u kunt dit later doen.';

  @override
  String get linkVetLater => 'Later';

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
  String get petBirthDate => 'Geboortedatum';

  @override
  String get petBirthDateInvalid => 'Ongeldige geboortedatum (JJJJ-MM-DD)';

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
    return 'S$day';
  }

  @override
  String get timelineTypeHeartrate => 'Hartslag';

  @override
  String get timelineTypeWeight => 'Gewicht';

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
  String get messageNoMessagesYet => 'Nog geen berichten';

  @override
  String get messageNewConversation => 'Nieuw gesprek';

  @override
  String get messageComposeTitle => 'Nieuw gesprek';

  @override
  String get messageChoosePro => 'Zorgprofessional';

  @override
  String get messageChooseClient => 'Cliënt';

  @override
  String get messageChoosePet => 'Betrokken dier';

  @override
  String get messageChoosePetOptional => 'Dier (optioneel)';

  @override
  String get messageGeneralThread => 'Algemeen gesprek';

  @override
  String get messageStartConversation => 'Starten';

  @override
  String get messageLockedTitle => 'Berichten niet beschikbaar';

  @override
  String get messageLockedBody =>
      'Koppel een dierenarts om met een zorgprofessional te chatten.';

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
      'Zoek op naam, e-mail of praktijk. Als ze al op petsFollow staan, wordt een koppelingsaanvraag naar de praktijk gestuurd.';

  @override
  String get addVetSearchLabel => 'Zoek een dierenarts';

  @override
  String get addVetSearchFieldHint => 'Naam, e-mail of praktijk';

  @override
  String get addVetNotListed => 'Mijn dierenarts staat niet in de lijst';

  @override
  String get addVetSuggestTitle => 'Nieuwe dierenarts';

  @override
  String get addVetSuggestBody =>
      'Geef het e-mailadres en telefoonnummer van de praktijk. Wij nemen contact op zodat ze petsFollow kunnen gebruiken.';

  @override
  String get addVetSuggestEmail => 'E-mail van de praktijk';

  @override
  String get addVetSuggestPhone => 'Telefoon';

  @override
  String get addVetSuggestNameOptional => 'Naam dierenarts (optioneel)';

  @override
  String get addVetSuggestCta => 'Suggestie versturen';

  @override
  String get vetSuggestSent =>
      'Bedankt — we contacteren de praktijk. U wordt verwittigd wanneer ze petsFollow gebruiken.';

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
  String get takeVideo => 'Video opnemen';

  @override
  String get chooseFromGallery => 'Kiezen uit galerij';

  @override
  String get attachMedia => 'Foto of video toevoegen';

  @override
  String get attachPhoto => 'Foto';

  @override
  String get attachVideo => 'Video';

  @override
  String get compressingMedia => 'Video comprimeren…';

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
  String get appInviteHintClient =>
      'Deel deze QR met een vriend. Die wordt aan jou gekoppeld (referral) en kan jouw praktijk volgen als die er nog geen heeft.';

  @override
  String get appInviteHintCommercial =>
      'Twee links: cliënten (app) en praktijken (Pro-registratie met uw doorverwijscode).';

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
  String get commercialManagerFieldSubtitle =>
      'Teamresultaten, QR-uitnodiging en toegang tot de Pro-site.';

  @override
  String get commercialOpenProWeb => 'Pro-site openen';

  @override
  String get managerTeamCta => 'Resultaten van mijn team';

  @override
  String get managerTeamTitle => 'Team';

  @override
  String get managerTeamSection => 'Teamresultaten';

  @override
  String get managerSelfSection => 'Mijn resultaten';

  @override
  String get managerMembersSection => 'Commercials';

  @override
  String get managerTeamEmpty => 'Geen commercials toegewezen.';

  @override
  String get managerKpiProspects => 'Prospects';

  @override
  String get managerKpiConverted => 'Geconverteerd';

  @override
  String get managerKpiConversion => 'Conversieratio';

  @override
  String get managerKpiAppointments => 'Komende afspraken';

  @override
  String get managerKpiStale => 'Stale pipeline';

  @override
  String get managerKpiMonthEarned => 'Commissies (maand)';

  @override
  String get managerKpiLifetimeEarned => 'Commissies (totaal)';

  @override
  String get managerKpiVets => 'Toegewezen dierenartsen';

  @override
  String get featureModules => 'Opties';

  @override
  String get featureModulesCatalog => 'Discover options';

  @override
  String get featureModulesSubtitle =>
      'Enable Care+, Horse, Kennel or Household as needed.';

  @override
  String get moduleCarePlus => 'Care+';

  @override
  String get moduleCarePlusDesc => 'Richer care reminders for all your pets.';

  @override
  String get moduleHorse => 'Horse';

  @override
  String get moduleHorseDesc => 'Pro contacts and competitions for horses.';

  @override
  String get moduleKennel => 'Kennel';

  @override
  String get moduleKennelDesc => 'Quick litter / kennel encoding.';

  @override
  String get moduleFamily => 'Household';

  @override
  String get moduleFamilyDesc => 'Household view of your pets.';

  @override
  String get moduleActivate => 'Enable';

  @override
  String get moduleActive => 'Active';

  @override
  String get switchProfile => 'Profiel wisselen';

  @override
  String get profilePersonal => 'Persoonlijk';

  @override
  String get profilePro => 'Professioneel';

  @override
  String get profileSwitched => 'Profile switched';

  @override
  String get preconsultTitle => 'Preconsultatie';

  @override
  String preconsultTitlePet(String petName) {
    return 'Preconsultatie — $petName';
  }

  @override
  String get preconsultIntro =>
      'Deel de toestand van uw dier vóór het bezoek zodat uw dierenarts zich kan voorbereiden.';

  @override
  String get preconsultComplaint => 'Hoofdreden / klacht';

  @override
  String get preconsultComplaintRequired => 'Geef de reden van het bezoek op';

  @override
  String get preconsultDuration => 'Sinds wanneer?';

  @override
  String get preconsultDurationToday => 'Vandaag';

  @override
  String get preconsultDurationFewDays => 'Enkele dagen';

  @override
  String get preconsultDurationWeek => 'Ongeveer een week';

  @override
  String get preconsultDurationWeeks => 'Meerdere weken';

  @override
  String get preconsultDurationMonths => 'Meerdere maanden';

  @override
  String get preconsultBehavior => 'Gedrag';

  @override
  String get preconsultBehaviorNormal => 'Normaal';

  @override
  String get preconsultBehaviorLethargic => 'Lusteloos';

  @override
  String get preconsultBehaviorRestless => 'Onrustig';

  @override
  String get preconsultBehaviorAggressive => 'Agressief';

  @override
  String get preconsultBehaviorAnxious => 'Angstig';

  @override
  String get preconsultBehaviorOther => 'Anders';

  @override
  String get preconsultAppetite => 'Eetlust';

  @override
  String get preconsultThirst => 'Dorst';

  @override
  String get preconsultElimination => 'Ontlasting / urine';

  @override
  String get preconsultScaleNormal => 'Normaal';

  @override
  String get preconsultScaleDecreased => 'Verminderd';

  @override
  String get preconsultScaleIncreased => 'Verhoogd';

  @override
  String get preconsultUrgency => 'Gevoelde urgentie';

  @override
  String get preconsultUrgencyLow => 'Laag';

  @override
  String get preconsultUrgencyMedium => 'Gemiddeld';

  @override
  String get preconsultUrgencyHigh => 'Hoog';

  @override
  String get preconsultComment => 'Opmerking (optioneel)';

  @override
  String get preconsultUnknown => 'Ik weet het niet';

  @override
  String get preconsultSubmit => 'Verzenden';

  @override
  String get preconsultSubmitted => 'Preconsultatie verzonden';

  @override
  String get preconsultAlreadySubmitted =>
      'U heeft deze preconsultatie al verzonden.';

  @override
  String get preconsultFillCta => 'Preconsultatie invullen';

  @override
  String get proLightAudioConsentClientCheck =>
      'Ik heb mondelinge toestemming van de klant om op te nemen';

  @override
  String get proLightReportAiProposalBanner =>
      'AI-voorstel — validatie door de dierenarts verplicht vóór finalisatie (incl. diagnose / medicatie).';

  @override
  String get proLightAiModuleRequired =>
      'AI-CR functie uitgeschakeld voor deze praktijk — neem contact op met petsFollow support.';

  @override
  String proLightAiModuleTrialBanner(int days) {
    return 'CR IA-proef — nog $days dagen. Dicteer en verbeter uw verslagen.';
  }

  @override
  String get proLightAiModuleInactiveBanner =>
      'CR IA niet geactiveerd voor dit kabinet. AI-dictatie is niet beschikbaar.';

  @override
  String get proLightAiModuleVisitScopedBanner =>
      'CR IA beschikbaar als het kabinet van het bezoek de module heeft geactiveerd.';

  @override
  String get supportTitle => 'Probleem melden';

  @override
  String get supportHint =>
      'Beschrijf de bug. Technische diagnostiek van de laatste 15 minuten wordt automatisch bijgevoegd.';

  @override
  String get supportSubject => 'Onderwerp';

  @override
  String get supportMessage => 'Beschrijving';

  @override
  String get supportDiagnosticsAttached =>
      'Diagnostiek automatisch bijgevoegd (fouten, requests, configuratie).';

  @override
  String get supportSubmit => 'Verzenden';

  @override
  String get supportSending => 'Verzenden…';

  @override
  String get supportSuccess => 'Bericht verzonden. Dank je!';

  @override
  String get supportErrorRateLimit =>
      'Te veel tickets recent. Probeer over een uur.';

  @override
  String get supportErrorTooLarge =>
      'Diagnostiek te groot. Herstart de app en probeer opnieuw.';

  @override
  String get supportMenu => 'Support';

  @override
  String get appInviteHintSales =>
      'Deel je sponsorcode met een praktijk of een cliënt.';

  @override
  String get appInviteCopyVet => 'Kopieer praktijk-aanmeldlink';

  @override
  String get appInviteCopyClient => 'Kopieer cliënt-uitnodigingslink';

  @override
  String get sendDossierToPro => 'Naar een pro sturen';

  @override
  String get sendDossierEmailLabel => 'E-mail van de professional';

  @override
  String get sendDossierEmailHint => 'vet@kliniek.be';

  @override
  String get sendDossierConfirm => 'Versturen';

  @override
  String get sendDossierSuccess => 'Dossier verstuurd — link 24 uur geldig.';

  @override
  String get sendDossierInvalidEmail => 'Ongeldig e-mailadres.';

  @override
  String get sendDossierPhiWarning =>
      'Dit dossier bevat gezondheidsgegevens: consultverslagen, gezondheidsboekje en documenten. De link blijft 24 uur geldig en iedereen die hem heeft, kan ze inkijken.';

  @override
  String get sendDossierPhiConsent =>
      'Ik ga ermee akkoord deze gezondheidsgegevens met deze professional te delen.';

  @override
  String get consultationsHistory => 'Consultaties';

  @override
  String get consultationTitle => 'Consultatie';

  @override
  String consultationTitleWithPet(String petName) {
    return 'Consultatie — $petName';
  }

  @override
  String get consultationVisitMeta => 'Bezoek';

  @override
  String consultationReportBy(String author) {
    return 'Verslag door $author';
  }

  @override
  String get consultationReportFallback => 'Verslag';

  @override
  String get consultationReportEmpty => '(leeg)';

  @override
  String get consultationAvailableCta => 'Beschikbaar';

  @override
  String get consultationPendingCta => 'Concept';

  @override
  String get sendConsultationToVet => 'Naar een dierenarts sturen';

  @override
  String get sendConsultationEmailLabel => 'E-mail van de dierenarts';

  @override
  String get sendConsultationEmailHint => 'vet@kliniek.be';

  @override
  String get sendConsultationConfirm => 'Versturen';

  @override
  String get sendConsultationSuccess =>
      'Consultatie verstuurd — link 24 uur geldig.';

  @override
  String get sendConsultationInvalidEmail => 'Ongeldig e-mailadres.';

  @override
  String get sendConsultationPhiWarning =>
      'Dit verslag bevat gezondheidsgegevens. De link blijft 24 uur geldig en iedereen die hem heeft, kan de PDF downloaden.';

  @override
  String get sendConsultationPhiConsent =>
      'Ik ga ermee akkoord dit verslag met deze dierenarts te delen.';

  @override
  String get clientAiDevBadge => 'dev';

  @override
  String get clientAiSectionTitle => 'AI-assistentie';

  @override
  String get clientAiExplainCta => 'Mijn verslag begrijpen';

  @override
  String get clientAiExplainTitle => 'Uw verslag uitgelegd';

  @override
  String get clientAiExplainDisclaimer =>
      'Dit is geen medisch advies. Volg altijd de instructies van uw dierenarts.';

  @override
  String get clientAiExplainLoading => 'Uitleg voorbereiden…';

  @override
  String get clientAiExplainListTitle => 'Een verslag begrijpen';

  @override
  String get clientAiExplainListSubtitle =>
      'Vereenvoudigde uitleg van een afgerond verslag';

  @override
  String get clientAiExplainListEmpty =>
      'Nog geen afgerond bezoekverslag om uit te leggen.';

  @override
  String get clientAiTriageTitle => 'Hulp bij spoedgevallen 24/7';

  @override
  String get clientAiTriageSubtitle =>
      'Beschrijf de situatie — wij beoordelen de urgentie.';

  @override
  String get clientAiTriageHint => 'Bv. mijn hond heeft chocolade gegeten…';

  @override
  String get clientAiTriageSend => 'Verzenden';

  @override
  String get clientAiTriageLevelGreen => 'Advies';

  @override
  String get clientAiTriageLevelOrange => 'Afspraak maken';

  @override
  String get clientAiTriageLevelRed => 'Spoed';

  @override
  String get clientAiTriageWatchSigns => 'Tekenen om te bewaken';

  @override
  String get clientAiTriageCallPractice => 'Praktijk bellen';

  @override
  String get clientAiTriageBookVisit => 'Afspraak maken';

  @override
  String get clientAiTriageMessageVet => 'Mijn dierenarts berichten';

  @override
  String get clientAiTriageSelectPet => 'Voor welk dier?';

  @override
  String get clientAiTriageNoPet => 'Doorgaan zonder dier';

  @override
  String get clientAiTriageStart => 'Start';

  @override
  String get clientAiTriageEmergencyFallback =>
      'Als u uw praktijk niet kunt bereiken, neem dan onmiddellijk contact op met een lokale spoeddierenarts.';

  @override
  String get clientAiTriageOpenMessages => 'Berichten openen';

  @override
  String get bloodPressureShort => 'RR';

  @override
  String get recordBloodPressureTitle => 'Bloeddruk registreren';

  @override
  String get bpSystolicLabel => 'Systolisch (mmHg)';

  @override
  String get bpDiastolicLabel => 'Diastolisch (mmHg)';

  @override
  String get bpMethodLabel => 'Methode';

  @override
  String get bpMethodDoppler => 'Doppler';

  @override
  String get bpMethodOscillometric => 'Oscillometrisch';

  @override
  String get bpMethodUnknown => 'Niet gespecificeerd';

  @override
  String get bloodPressureInvalid => 'Geldige RR invoeren (SYS ≥ DIA)';

  @override
  String get bloodPressureSaved => 'Bloeddruk opgeslagen';

  @override
  String get labsTitle => 'Labresultaten';

  @override
  String get labsEmpty => 'Geen labresultaten';

  @override
  String get labsResults => 'Resultaten';

  @override
  String labsAbnormalCount(int count) {
    return '$count buiten bereik';
  }

  @override
  String get labsOpen => 'Labresultaten bekijken';

  @override
  String get labsFlagLow => 'Laag';

  @override
  String get labsFlagHigh => 'Hoog';

  @override
  String get labsFlagNormal => 'Normaal';

  @override
  String get labsOpenDocument => 'Document openen';

  @override
  String get bpMethodInvasive => 'Invasief';

  @override
  String get bpSiteLabel => 'Meetplaats (optioneel)';

  @override
  String get bpCommentLabel => 'Opmerking (optioneel)';

  @override
  String get bloodPressureSave => 'Opslaan';
}
