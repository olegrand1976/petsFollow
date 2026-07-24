// ignore: unused_import
import 'package:intl/intl.dart' as intl;
import 'app_localizations.dart';

// ignore_for_file: type=lint

/// The translations for Estonian (`et`).
class AppLocalizationsEt extends AppLocalizations {
  AppLocalizationsEt([String locale = 'et']) : super(locale);

  @override
  String get appTitle => 'petsFollow';

  @override
  String get appTagline => 'Teie lemmiklooma tervise jälgimine';

  @override
  String get email => 'E-post';

  @override
  String get password => 'Parool';

  @override
  String get login => 'Logi sisse';

  @override
  String get loginFailed => 'Sisselogimine ebaõnnestus';

  @override
  String get loginOr => 'või';

  @override
  String get loginWithGoogle => 'Jätka Google\'iga';

  @override
  String get twoFaTitle => '2FA kinnitus';

  @override
  String get twoFaSubtitle =>
      'Sisestage autentimisrakendusest 6-kohaline kood.';

  @override
  String get twoFaCode => 'Autentimiskood';

  @override
  String get twoFaSubmit => 'Kinnita';

  @override
  String get twoFaBack => 'Tagasi sisselogimise juurde';

  @override
  String get twoFaInvalid => 'Kehtetu või aegunud 2FA kood';

  @override
  String get forgotPassword => 'Unustasid parooli?';

  @override
  String get forgotPasswordTitle => 'Unustatud parool';

  @override
  String get forgotPasswordSubtitle =>
      'Sisestage oma konto e-post. Kui konto on olemas, saadetakse lähtestamislink.';

  @override
  String get forgotPasswordSubmit => 'Saada link';

  @override
  String get forgotPasswordBack => 'Tagasi sisselogimise juurde';

  @override
  String get forgotPasswordFailed => 'E-kirja saatmine ebaõnnestus';

  @override
  String get forgotPasswordSentTitle => 'E-kiri saadetud';

  @override
  String forgotPasswordSent(String email) {
    return 'Kui kontol $email on olemas, on link saadetud. Avage see brauseris, et valida uus parool.';
  }

  @override
  String get emailRequired => 'Sisestage kehtiv e-posti aadress';

  @override
  String get resetPasswordTitle => 'Uus parool';

  @override
  String get resetPasswordSubtitle => 'Vähemalt 8 tähemärki.';

  @override
  String get resetPasswordToken => 'Lähtestamise märgis';

  @override
  String get resetPasswordSubmit => 'Salvesta';

  @override
  String get resetPasswordBackToLogin => 'Mine sisselogimisele';

  @override
  String get resetPasswordInvalidLink => 'Kehtetu lähtestamislink';

  @override
  String get resetPasswordFailed => 'Parooli lähtestamine ebaõnnestus';

  @override
  String get resetPasswordDoneTitle => 'Parool uuendatud';

  @override
  String get resetPasswordDoneSubtitle => 'Saate nüüd sisse logida.';

  @override
  String get fullName => 'Täisnimi';

  @override
  String get registerCta => 'Loo konto';

  @override
  String get registerTitle => 'Loo konto';

  @override
  String get registerSubtitle =>
      'Jälgige oma lemmiklooma tervist. Järgmisena kinnitage e-post.';

  @override
  String get registerSubmit => 'Registreeru';

  @override
  String get registerSuccess =>
      'Konto loodud. Kontrollige e-posti kinnitamiseks, seejärel logige sisse.';

  @override
  String get registerFailed => 'Registreerimine ebaõnnestus';

  @override
  String get registerEmailExists => 'See e-post on juba kasutusel';

  @override
  String get registerBackToLogin => 'Tagasi sisselogimise juurde';

  @override
  String get vetUseProWeb =>
      'Täielikud loomaarsti kontod kasutavad Pro veebisaiti.';

  @override
  String get unsupportedRoleApp =>
      'Seda kontot ei saa pets rakenduses kasutada. Kasutage Pro veebisaiti.';

  @override
  String get proLightTitle => 'Välipro';

  @override
  String get proLightAgenda => 'Päevakava';

  @override
  String get proLightClients => 'Kliendid';

  @override
  String get proLightPets => 'Lemmikloomad';

  @override
  String get proLightLoadError => 'Laadimine ebaõnnestus';

  @override
  String get proLightNoVisits => 'Kohtumisi pole';

  @override
  String get proLightNoClients => 'Jagatud kliente pole';

  @override
  String get proLightNoPets => 'Jagatud lemmikloomi pole';

  @override
  String get proLightAddress => 'Aadress';

  @override
  String get proLightOpenMaps => 'Kaardid';

  @override
  String get proLightReportTitle => 'Visiidi aruanne';

  @override
  String get proLightReportHint => 'Visiidi märkmed…';

  @override
  String get proLightImproveAi => 'Paranda (AI)';

  @override
  String get proLightFinalizeReport => 'Lõpeta';

  @override
  String get proLightReportFinal => 'Lõpetatud';

  @override
  String get proLightSettings => 'Seaded';

  @override
  String get proLightSpecialty => 'Eriala';

  @override
  String get proLightDocuments => 'Dokumendid';

  @override
  String get proLightNoDocuments => 'Dokumente pole';

  @override
  String get proLightTimeline => 'Ajajoon';

  @override
  String get proLightNoTimeline => 'Sündmusi pole';

  @override
  String get proLightReminders => 'Meeldetuletused';

  @override
  String get proLightNoReminders => 'Meeldetuletusi pole';

  @override
  String get proLightLitterTag => 'Pesakond / silt';

  @override
  String get proLightActionFailed => 'Tegevus ebaõnnestus';

  @override
  String get proLightReadOnly => 'Ainult lugemisõigus';

  @override
  String get proLightUseGps => 'GPS';

  @override
  String get proLightTranscribeAudio => 'Audio → tekst';

  @override
  String get proLightGpsDenied => 'Asukoht pole saadaval';

  @override
  String get googleNotConfigured =>
      'Google\'iga sisselogimine pole seadistatud';

  @override
  String get googleLoginFailed => 'Google\'iga sisselogimine ebaõnnestus';

  @override
  String get googleClientNotFound =>
      'Selle e-posti jaoks pole kliendikontot. Paluge loomaarstilt kutset';

  @override
  String get googleWrongAudience => 'See Google\'i konto ei ole kliendiprofiil';

  @override
  String get myPets => 'Minu lemmikloomad';

  @override
  String get myData => 'Minu andmed';

  @override
  String get settings => 'Seaded';

  @override
  String get logout => 'Logi välja';

  @override
  String get save => 'Salvesta';

  @override
  String get cancel => 'Tühista';

  @override
  String get firstName => 'Eesnimi';

  @override
  String get currentPassword => 'Praegune parool';

  @override
  String get newPassword => 'Uus parool';

  @override
  String get confirmNewPassword => 'Kinnita parool';

  @override
  String get changePassword => 'Muuda parooli';

  @override
  String get forceChangePasswordTitle => 'Muutke oma parooli';

  @override
  String get forceChangePasswordSubtitle =>
      'See konto loodi ajutise parooliga. Jätkamiseks valige enda oma.';

  @override
  String get forceChangePasswordSubmit => 'Salvesta ja jätka';

  @override
  String get passwordTooShort => 'Vähemalt 8 tähemärki';

  @override
  String get passwordMismatch => 'Paroolid ei ühti';

  @override
  String get passwordChangeFailed => 'Parooli muutmine ebaõnnestus';

  @override
  String get deleteAccount => 'Kustuta konto';

  @override
  String get deleteAccountConfirm =>
      'Seda toimingut ei saa tagasi võtta. Kõik teie lemmikloomad ja andmed kustutatakse.';

  @override
  String get profileSaved => 'Profiil salvestatud';

  @override
  String get changePhoto => 'Muuda fotot';

  @override
  String get addPhoto => 'Lisa foto';

  @override
  String get photoUpdated => 'Foto uuendatud';

  @override
  String get passwordChanged => 'Parool muudetud';

  @override
  String greeting(String name) {
    return 'Tere $name,';
  }

  @override
  String get latestValues => 'Viimased väärtused';

  @override
  String get startMeasurement => 'ALUSTA MÕÕTMIST';

  @override
  String get choosePetForMeasurement => 'Valige lemmikloom';

  @override
  String get chooseDuration => 'Mõõtmise kestus';

  @override
  String durationSeconds(int seconds) {
    return '$seconds s';
  }

  @override
  String get howToMeasure => 'Kuidas mõõta?';

  @override
  String get howToMeasureIntro =>
      'Mõõtke oma lemmiklooma puhkeoleku südame löögisagedust.';

  @override
  String get howToMeasureStep1 =>
      '1. Hoidke lemmikloom rahulikuna, lamades või istudes.';

  @override
  String get howToMeasureStep2 =>
      '2. Asetage käsi rinnale ja puudutage iga löögi peale näidatud kestuse jooksul.';

  @override
  String get howToMeasureStep3 =>
      '3. Kinnitage näit, et saata see oma loomaarstile.';

  @override
  String get howToMeasureWhyTitle => 'Miks mõõta?';

  @override
  String get howToMeasureWhyBody =>
      'Regulaarne südame löögisageduse jälgimine aitab muutusi märgata ja ravi koos loomaarstiga kohandada.';

  @override
  String get reminders => 'Meeldetuletused';

  @override
  String get remindersHint =>
      'Saage igapäevane meeldetuletus südame löögisageduse mõõtmiseks.';

  @override
  String get remindersEnabled => 'Luba meeldetuletused';

  @override
  String get remindersTime => 'Meeldetuletuse aeg';

  @override
  String get remindersSaved => 'Meeldetuletused salvestatud';

  @override
  String get legalTermsTitle => 'Kasutustingimused';

  @override
  String get legalPrivacyTitle => 'Privaatsuspoliitika';

  @override
  String get legalNoticeTitle => 'Juriidiline teave';

  @override
  String get legalOpenOnline => 'Vaata veebiversiooni';

  @override
  String get legalTermsBody =>
      'Kasutustingimused — petsFollow\n\nPetsFollowi rakendus võimaldab lemmikloomaomanikel mõõta südame löögisagedust, vaadata ajalugu ja suhelda oma loomaarstiga.\n\nTeenuseid osutatakse valitud tellimuse alusel (maksed Stripe\'i kaudu). Kasutajad peavad rakendust kasutama ettenähtud otstarbel.\n\nTäielik versioon: https://petsfollow.ll-it-sc.be/legal/terms\n\nViimati uuendatud: juuli 2026';

  @override
  String get legalPrivacyBody =>
      'Privaatsuspoliitika — petsFollow\n\nKogutavad andmed: isikuandmed (eesnimi, e-post), lemmiklooma andmed (nimi, liik, tõug, fotod), südame löögisageduse näidud (looma terviseandmed), sõnumid ja meedia praktikaga, teavituste märgid (FCM), Stripe\'i töödeldud makseandmed.\n\nEesmärgid: konto haldamine, südame jälgimine, loomaarsti sõnumid, teavitused, arveldus.\n\nTöötlejad / partnerid: Google (Sign-In, Firebase Cloud Messaging), Stripe (maksed), pilvemajutus (GCP).\n\nSäilitamine: kuni konto kustutamiseni või 3 aastat tegevusetust.\n\nGDPR õigused (juurdepääs, parandamine, kustutamine): Profiil → Kustuta konto või support@ll-it-sc.be.\n\nTäielik versioon: https://petsfollow.ll-it-sc.be/legal/privacy\n\nViimati uuendatud: juuli 2026';

  @override
  String get legalNoticeBody =>
      'Juriidiline teave — petsFollow\n\nVäljaandja: LL-IT-SC / petsFollow\nKontakt: support@ll-it-sc.be\n\nMajutus: Google Cloud Platform (GDPR-ga kooskõlas).\n\nVäljaande juht: petsFollow.\n\nTäielik versioon: https://petsfollow.ll-it-sc.be/legal/mentions\n\nViimati uuendatud: juuli 2026';

  @override
  String get language => 'Keel';

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
  String get planAnnualLabel => '35 € / aasta';

  @override
  String get planTriennialLabel => '95 € / 3 aastat';

  @override
  String get planQuinquennialLabel => '145 € / 5 aastat';

  @override
  String get pushNewMessage => 'Uus sõnum';

  @override
  String get pushVisitConfirmed => 'Kohtumine kinnitatud';

  @override
  String get pushVisitProposed => 'Kohtumise ettepanek';

  @override
  String get pushVisitReschedule => 'Kohtumise ümberplaneerimine';

  @override
  String get notifChannelMessages => 'Sõnumid';

  @override
  String get notifChannelVisits => 'Visiidid';

  @override
  String get notifChannelCare => 'Hooldus';

  @override
  String get paymentResume => 'Jätka makset';

  @override
  String get manageSubscription => 'Halda tellimust';

  @override
  String get heartRate => 'Südame löögisageduse näit';

  @override
  String get history => 'Ajalugu';

  @override
  String get vetMessaging => 'Loomaarsti sõnumid';

  @override
  String get badgeAutoRenew => 'Automaatne uuendamine';

  @override
  String get badgeActive => 'Aktiivne';

  @override
  String get badgePendingPayment => 'Makse ootel';

  @override
  String badgeExpiresOn(String date) {
    return 'aegub $date';
  }

  @override
  String get newPet => 'Uus lemmikloom';

  @override
  String get petName => 'Nimi';

  @override
  String get species => 'Liik';

  @override
  String get breed => 'Tõug';

  @override
  String get choosePlan => 'Valige oma plaan';

  @override
  String get recommended => 'Soovitatud';

  @override
  String get autoRenewTitle => 'Automaatne uuendamine';

  @override
  String get autoRenewSubtitle => 'Võetakse iga uuendamise ajal';

  @override
  String get continueToPayment => 'Jätka maksele';

  @override
  String get paymentConfirmed => 'Makse kinnitatud — lemmikloom aktiivne';

  @override
  String get paymentPending => 'Makse ootel — saate hiljem jätkata';

  @override
  String errorGeneric(String message) {
    return 'Viga: $message';
  }

  @override
  String get errorNetwork =>
      'Ühendus ebaõnnestus. Kontrollige võrku ja proovige uuesti.';

  @override
  String get retryAction => 'Proovi uuesti';

  @override
  String get errorMediaTooLarge => 'Fail on liiga suur (maks. 25 MB)';

  @override
  String get errorInvalidMediaType =>
      'Toetamata vorming (JPEG, PNG, WebP, MP4, MOV, WebM)';

  @override
  String get errorPaymentRequired => 'Meedia saatmiseks on vaja tellimust';

  @override
  String get errorPhotoUploadFailed =>
      'Lemmikloom loodud, kuid fotot ei õnnestunud üles laadida';

  @override
  String get errorCouldNotOpenLink => 'Linki ei õnnestunud avada';

  @override
  String planAnnualSub(String price) {
    return '$price, automaatselt uuendatud';
  }

  @override
  String get planTriennialSub =>
      '95 € iga 3 aasta tagant, automaatselt uuendatud';

  @override
  String get planQuinquennialSub => '145 € 5 aastaks, ühekordne makse';

  @override
  String planOneTime(String price) {
    return '$price, ühekordne makse';
  }

  @override
  String get heartRateInstructions =>
      'Puudutage iga löögi peale teie loomaarsti määratud kestuse jooksul.';

  @override
  String heartRateInstructionsDuration(int seconds) {
    return 'Puudutage iga löögi peale $seconds sekundi jooksul.';
  }

  @override
  String get heartRateNoDurationConfigured =>
      'Selle praktika jaoks pole mõõtmise kestust seadistatud. Võtke ühendust oma loomaarstiga.';

  @override
  String get start => 'Alusta';

  @override
  String secondsLeft(int seconds) {
    return '$seconds s';
  }

  @override
  String beatsCount(int count) {
    return '$count lööki';
  }

  @override
  String get tapHere => 'Puudutage siin iga löögi peale';

  @override
  String bpmLabel(String bpm) {
    return 'BPM: $bpm';
  }

  @override
  String beatsLabel(int count) {
    return 'Lööke: $count';
  }

  @override
  String get thresholdAlert => 'Lävehoiatus';

  @override
  String get validateAndSend => 'Kinnita ja saada loomaarstile';

  @override
  String get restart => 'Alusta uuesti';

  @override
  String get sentToVet => 'Näit saadetud loomaarstile';

  @override
  String get navHome => 'Avaleht';

  @override
  String get navPets => 'Lemmikloomad';

  @override
  String get navCare => 'Hooldus';

  @override
  String get navMessages => 'Sõnumid';

  @override
  String get navProfile => 'Profiil';

  @override
  String get speciesDog => 'Koer';

  @override
  String get speciesCat => 'Kass';

  @override
  String get speciesHorse => 'Horse';

  @override
  String get speciesOther => 'Muu';

  @override
  String get careComingSoon => 'Hoolduse meeldetuletused tulevad peagi';

  @override
  String get emptyPetsTitle => 'Lemmikloomi pole veel';

  @override
  String get emptyPetsBody =>
      'Lisage oma esimene lemmikloom, et alustada südame löögisageduse jälgimist koos loomaarstiga.';

  @override
  String get discoveryTitle => 'Avastage petsFollow';

  @override
  String get discoveryMission => 'Teie 7-päevane teekond';

  @override
  String get discoveryDay0Title => 'Päev 0 — Tere tulemast';

  @override
  String get discoveryDay0Body =>
      'Looge oma lemmiklooma profiil ja õppige südame löögisagedust mõõtma.';

  @override
  String get discoveryDay2Title => 'Päev 2 — Esimene näit';

  @override
  String get discoveryDay2Body =>
      'Tehke esimene südame löögisageduse mõõtmine ja harjuge tehnikaga.';

  @override
  String get discoveryDay4Title => 'Päev 4 — Rutiin';

  @override
  String get discoveryDay4Body =>
      'Looge igapäevane mõõtmisharjumus isiklike meeldetuletustega.';

  @override
  String get discoveryDay6Title => 'Päev 6 — Jagamine loomaarstiga';

  @override
  String get discoveryDay6Body =>
      'Teie näidud jagatakse loomaarstiga optimaalseks jälgimiseks.';

  @override
  String get myVets => 'Minu loomaarstid';

  @override
  String get addVetByEmail => 'Lisa loomaarst e-posti järgi';

  @override
  String get vetEmailHint => 'email@practice.vet';

  @override
  String get noVets => 'Seotud loomaarsti pole';

  @override
  String get primaryVet => 'Peamine loomaarst';

  @override
  String get setPrimaryVet => 'Määra peamiseks loomaarstiks';

  @override
  String get careTitle => 'Hooldus';

  @override
  String get careDone => 'Tehtud';

  @override
  String get carePostpone => 'Lükka edasi';

  @override
  String get careOverdue => 'Tähtaeg ületatud';

  @override
  String get visitHistory => 'Visiitide ajalugu';

  @override
  String get requestVisit => 'Taotle visiiti';

  @override
  String get calendarBookingDisabled =>
      'Veebibroneering pole selle praktika jaoks saadaval. Saate siiski saata taotluse ilma ajaslottita.';

  @override
  String get calendarBookingDisabledReschedule =>
      'Veebibroneering pole saadaval. Paku kuupäev käsitsi.';

  @override
  String get calendarNoSlots => 'Järgmise 14 päeva jooksul vabu aegu pole.';

  @override
  String get calendarPickSlot => 'Valige aeg:';

  @override
  String get visitConfirm => 'Kinnita';

  @override
  String get visitProposeReschedule => 'Paku teist aega';

  @override
  String get visitRescheduleProposed => 'Ümberplaneerimise ettepanek saadetud';

  @override
  String get paymentSuccessSnack => 'Makse vastu võetud — värskendamine…';

  @override
  String get paymentCancelSnack => 'Makse tühistatud';

  @override
  String get visitRejectReschedule => 'Keeldu ümberplaneerimisest';

  @override
  String get visitAcceptReschedule => 'Nõustu uue ajaga';

  @override
  String get upcomingVisit => 'Eelseisev visiit';

  @override
  String get notificationHrTitle => 'petsFollow';

  @override
  String get notificationHrBody =>
      'Aeg teie lemmiklooma südame löögisagedust mõõta';

  @override
  String get reviewAskTitle => 'Meeldib petsFollow?';

  @override
  String get reviewAskYes => 'Jah, hinda rakendust';

  @override
  String get reviewAskNo => 'Hiljem';

  @override
  String get carePlusUpsell =>
      'Care+ — ravimid ja isikupärastatud meeldetuletused';

  @override
  String get carePlusRequired =>
      'Care+ on vajalik ravimite ja kohandatud meeldetuletuste jaoks.';

  @override
  String get horsePackRequired =>
      'Horse pakett on vajalik sepa meeldetuletuste, kontaktide ja võistluste jaoks.';

  @override
  String get activateAddon => 'Aktiveeri';

  @override
  String get careTypeMedication => 'Ravim';

  @override
  String get horseAddContact => 'Lisa kontakt';

  @override
  String get horseAddCompetition => 'Lisa võistlus';

  @override
  String get horseContactName => 'Nimi';

  @override
  String get horseContactRole => 'Roll';

  @override
  String get horseCompetitionTitle => 'Sündmus';

  @override
  String get horseCompetitionDate => 'Kuupäev (AAAA-KK-PP)';

  @override
  String get familyPackHint =>
      'Family pakett — majapidamise hoolduse vaade, −10% alates 2. tasulisest lemmiklooma plaanist';

  @override
  String familyHouseholdTitle(int count) {
    return 'Family majapidamine — $count lemmiklooma';
  }

  @override
  String kennelHouseholdTitle(int count) {
    return 'Kennel majapidamine — $count lemmiklooma';
  }

  @override
  String get familyHouseholdNext => 'Eelseisvad majapidamise meeldetuletused';

  @override
  String get familyPetLimit =>
      'Majapidamise pakett on juba aktiivne või ostmisel';

  @override
  String get familyRequiresTwoPets =>
      'Family pakett nõuab vähemalt 2 lemmiklooma';

  @override
  String get kennelPackHint =>
      'Kennel pakett — 6+ lemmiklooma, −15% järgmistele lemmiklooma plaanidele';

  @override
  String get kennelRequiresSixPets =>
      'Kennel pakett nõuab vähemalt 6 lemmiklooma';

  @override
  String get kennelQuickEncodeTitle => 'Pesakonna kiirkodeerimine';

  @override
  String get kennelRequired => 'Partii kodeerimiseks on vaja Kennel paketti';

  @override
  String get litterTag => 'Pesakonna silt';

  @override
  String get discoveryMarkDone => 'Missioon täidetud';

  @override
  String get notificationPreferences => 'Teavituste eelistused';

  @override
  String get notificationPrefsHint =>
      'Valige, milliseid teavitusi soovite vastu võtta.';

  @override
  String get notificationPrefsSaved => 'Eelistused salvestatud';

  @override
  String get notificationPrefHr => 'Südame löögisageduse näidud';

  @override
  String get notificationPrefCare => 'Hoolduse meeldetuletused';

  @override
  String get notificationPrefVisits => 'Visiidid';

  @override
  String get notificationPrefMessages => 'Sõnumid';

  @override
  String get notificationPrefDiscovery => 'Avastamise teekond';

  @override
  String get notificationPrefBilling => 'Arveldus';

  @override
  String carePostponeDays(int days) {
    return 'Lükka edasi $days päeva';
  }

  @override
  String get noCareReminders => 'Ootel hoolduse meeldetuletusi pole';

  @override
  String get careAddReminder => 'Lisa meeldetuletus';

  @override
  String get careSelectPet => 'Lemmikloom';

  @override
  String careDueInDays(int days) {
    return 'Tähtaeg $days päeva pärast';
  }

  @override
  String get careReferenceModeDone => 'Juba tehtud';

  @override
  String get careReferenceModeFirst => 'Esimene kord';

  @override
  String get careLastDateLabel => 'Viitekuupäev';

  @override
  String get careLastDateDone => 'Viimase hoolduse kuupäev';

  @override
  String get careLastDateFirst => 'Tsükli alguskuupäev';

  @override
  String get careRecurrenceLabel => 'Korduvus';

  @override
  String get careRecurrenceNone => 'Puudub (ühekordne tähtaeg)';

  @override
  String careRecurrenceDays(int days) {
    return 'Iga $days päeva tagant';
  }

  @override
  String get careDueDateLabel => 'Tähtaeg';

  @override
  String get careDueDateComputed => 'Arvutatud tähtaeg';

  @override
  String get careTooltipDoneWithRecurrence =>
      'Juba tehtud: tähtaeg = viimase hoolduse kuupäev + korduvus.';

  @override
  String get careTooltipFirstWithRecurrence =>
      'Esimene ajakava: sisestage tsükli alguskuupäev. Tähtaeg = see kuupäev + korduvus.';

  @override
  String get careTooltipNoRecurrence =>
      'Ilma korduvuseta: sisestatud kuupäev on ainus tähtaeg.';

  @override
  String get careTooltipDueExplained =>
      'Tähtaeg = viitekuupäev + korduvus (kui määratud).';

  @override
  String get carePickDate => 'Valige kuupäev';

  @override
  String discoveryDayBadge(int day) {
    return 'P$day';
  }

  @override
  String get timelineTypeHeartrate => 'Südame löögisagedus';

  @override
  String get timelineTypeMessage => 'Sõnum';

  @override
  String get timelineTypeCare => 'Hooldus';

  @override
  String get timelineTypeVisit => 'Visiit';

  @override
  String get timelineTypeEvent => 'Sündmus';

  @override
  String get visitCancelAction => 'Tühista taotlus';

  @override
  String get upcomingVisits => 'Eelseisvad visiidid';

  @override
  String get timelineEmpty => 'Sündmusi pole veel';

  @override
  String get noThreads => 'Vestlusi pole';

  @override
  String get vetInviteSent =>
      'Kutse saadetud — praktika peab taotluse vastu võtma';

  @override
  String vetInviteSentNamed(String practice) {
    return 'Taotlus saadetud praktikale $practice — praktika peab selle vastu võtma';
  }

  @override
  String get vetNotFound => 'Selle e-postiga loomaarsti ei leitud';

  @override
  String get addVetSearchHint =>
      'Otsime seda loomaarsti kontot petsFollowist. Kui see on olemas, saadetakse praktikale ühenduse taotlus.';

  @override
  String get visitRequested => 'Visiidi taotlus saadetud';

  @override
  String get primaryVetSet => 'Peamine loomaarst uuendatud';

  @override
  String get visitStatusRequested => 'Taotletud';

  @override
  String get visitStatusConfirmed => 'Kinnitatud';

  @override
  String get visitStatusDone => 'Lõpetatud';

  @override
  String get visitStatusCancelled => 'Tühistatud';

  @override
  String get visitStatusReschedulePending => 'Ümberplaneerimine ootel';

  @override
  String get horseHealthTitle => 'Horse tervis';

  @override
  String get horseContactsTitle => 'Kontaktid (sepp, hambaarst…)';

  @override
  String get horseCompetitionsTitle => 'Võistlused';

  @override
  String get horseContactsSoon =>
      'Aktiveerige Horse pakett professionaalsete kontaktide haldamiseks.';

  @override
  String get horseCompetitionsSoon =>
      'Aktiveerige Horse pakett võistluskalendri jaoks.';

  @override
  String get horsePackUpsell =>
      'Horse pakett — sepp, väljaheidete munade loendus, kontaktid ja võistlused';

  @override
  String get careTypeFarrier => 'Sepp';

  @override
  String get careTypeFecalEgg => 'Väljaheidete munade loendus';

  @override
  String get careTypeVaccination => 'Vaktsineerimine';

  @override
  String get careTypeDeworming => 'Ussirohi';

  @override
  String get careTypeVetCheck => 'Loomaarsti kontroll';

  @override
  String get careTypeDental => 'Hambaravi';

  @override
  String get careTypeCustom => 'Kohandatud meeldetuletus';

  @override
  String get homeAddFirstVetTitle => 'Lisage oma loomaarst';

  @override
  String get homeAddFirstVetBody =>
      'Siduge praktika, mis teie lemmiklooma jälgib, et jagada näite ja vestelda.';

  @override
  String get homeAddFirstVetCta => 'Lisa loomaarst';

  @override
  String get photoFrameHint => 'Keskele koon — lemmiklooma profiili eelvaade';

  @override
  String get takePhoto => 'Tee foto';

  @override
  String get chooseFromGallery => 'Vali galeriist';

  @override
  String get attachMedia => 'Lisa foto või video';

  @override
  String get attachPhoto => 'Foto';

  @override
  String get attachVideo => 'Video';

  @override
  String get openMedia => 'Ava';

  @override
  String get mediaVideoLabel => 'Video';
}
