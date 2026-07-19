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
  String get loginOr => 'of';

  @override
  String get loginWithGoogle => 'Doorgaan met Google';

  @override
  String get googleNotConfigured => 'Google-aanmelding is niet geconfigureerd';

  @override
  String get googleLoginFailed => 'Google-aanmelding mislukt';

  @override
  String get googleClientNotFound =>
      'Geen klantaccount voor dit e-mailadres. Vraag uw dierenarts om een uitnodiging';

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
  String get languageEs => 'Español';

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
      'Familiepakket — gezinsweergave, −10% op volgende abonnementen';

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
}
