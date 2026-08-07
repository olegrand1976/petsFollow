// ignore: unused_import
import 'package:intl/intl.dart' as intl;
import 'app_localizations.dart';

// ignore_for_file: type=lint

/// The translations for Italian (`it`).
class AppLocalizationsIt extends AppLocalizations {
  AppLocalizationsIt([String locale = 'it']) : super(locale);

  @override
  String get appTitle => 'petsFollow';

  @override
  String get appTagline => 'Monitoraggio sanitario del tuo animale';

  @override
  String get email => 'E-mail';

  @override
  String get password => 'Password';

  @override
  String get login => 'Login';

  @override
  String get loginFailed => 'Impossibile connettersi';

  @override
  String get emailNotVerified =>
      'Per prima cosa conferma la tua email (link ricevuto al momento della registrazione), quindi effettua nuovamente l\'accesso.';

  @override
  String get resendConfirmation => 'Reinvia l\'e-mail di conferma';

  @override
  String get resendConfirmationSent =>
      'Se l\'account esiste e non è ancora confermato, è stata inviata una nuova e-mail.';

  @override
  String get resendConfirmationFailed =>
      'Invio non riuscito. Riprova tra poco.';

  @override
  String get loginOr => 'O';

  @override
  String get loginWithGoogle => 'Continua con Google';

  @override
  String get loginWithApple => 'Continua con Apple';

  @override
  String get appleComingSoon =>
      'La connessione con Apple sarà presto disponibile.';

  @override
  String get twoFaTitle => 'Verifica 2FA';

  @override
  String get twoFaSubtitle =>
      'Inserisci il codice di 6 cifre per la tua app di autenticazione.';

  @override
  String get twoFaCode => 'Codice dell\'autenticatore';

  @override
  String get twoFaSubmit => 'Per convalidare';

  @override
  String get twoFaBack => 'Torna al login';

  @override
  String get twoFaInvalid => 'Codice 2FA non valido o scaduto';

  @override
  String get forgotPassword => 'Password dimenticata?';

  @override
  String get forgotPasswordTitle => 'Password dimenticata';

  @override
  String get forgotPasswordSubtitle =>
      'Inserisci l\'e-mail del tuo account. Se esiste un account, verrà inviato un collegamento di reimpostazione.';

  @override
  String get forgotPasswordSubmit => 'Invia collegamento';

  @override
  String get forgotPasswordBack => 'Torna al login';

  @override
  String get forgotPasswordFailed => 'Impossibile inviare';

  @override
  String get forgotPasswordSentTitle => 'E-mail inviata';

  @override
  String forgotPasswordSent(String email) {
    return 'Se esiste un account per $email, è stato inviato un collegamento. Aprilo nel tuo browser per scegliere una nuova password.';
  }

  @override
  String get emailRequired => 'Inserisci un indirizzo email valido';

  @override
  String get resetPasswordTitle => 'Nuova password';

  @override
  String get resetPasswordSubtitle => 'Minimo 8 caratteri.';

  @override
  String get resetPasswordToken => 'Token di reset';

  @override
  String get resetPasswordSubmit => 'Salva';

  @override
  String get resetPasswordBackToLogin => 'Vai al login';

  @override
  String get resetPasswordInvalidLink =>
      'Collegamento di reimpostazione non valido';

  @override
  String get resetPasswordFailed => 'Impossibile reimpostare';

  @override
  String get resetPasswordDoneTitle => 'Password aggiornata';

  @override
  String get resetPasswordDoneSubtitle => 'Ora puoi accedere.';

  @override
  String get fullName => 'Nome e cognome';

  @override
  String get registerCta => 'Registro';

  @override
  String get registerTitle => 'Registro';

  @override
  String get registerSubtitle =>
      'Crea il tuo account per monitorare la salute del tuo animale domestico. Ti verrà inviata un\'e-mail di convalida.';

  @override
  String get registerSubmit => 'Registro';

  @override
  String get registerSuccess =>
      'Conto creato. Apri il collegamento nell\'e-mail di convalida, quindi accedi nuovamente all\'app.';

  @override
  String get registerInviteNotApplied =>
      'Account creato, ma il codice invito non è stato applicato. Potrai inserirlo di nuovo dopo l\'accesso.';

  @override
  String get registerFailed => 'Impossibile registrarsi';

  @override
  String get registerEmailExists => 'Questa email è già in uso';

  @override
  String get registerBackToLogin => 'Torna al login';

  @override
  String get confirmEmailTitle => 'Conferma e-mail';

  @override
  String get confirmEmailLoading => 'Conferma in corso…';

  @override
  String get confirmEmailDoneTitle => 'E-mail confermata';

  @override
  String get confirmEmailDoneSubtitle =>
      'Il tuo account è attivato. Puoi accedere.';

  @override
  String get confirmEmailFailedTitle => 'Impossibile confermare';

  @override
  String get confirmEmailFailed => 'Impossibile confermare questa email.';

  @override
  String get confirmEmailInvalidLink =>
      'Link di conferma non valido o già utilizzato.';

  @override
  String get confirmEmailBackToLogin => 'Torna al login';

  @override
  String get vetUseProWeb =>
      'L\'account veterinario completo viene utilizzato sul sito Web Pro.';

  @override
  String get unsupportedRoleApp =>
      'Questo account non può essere utilizzato nell\'app Animali domestici. Utilizzare il sito Web Pro.';

  @override
  String get proLightTitle => 'Campo professionistico';

  @override
  String get proLightAgenda => 'Diario';

  @override
  String get proLightClients => 'Clienti';

  @override
  String get proLightPets => 'Animali';

  @override
  String get proLightLoadError => 'Impossibile caricare';

  @override
  String get proLightNoVisits => 'Nessun appuntamento';

  @override
  String get proLightTourToday => 'Oggi';

  @override
  String get proLightTourWeek => '7 giorni';

  @override
  String get proLightTourAll => 'Tutto';

  @override
  String get proLightNoTourToday => 'Nessun appuntamento oggi';

  @override
  String get proLightNoTourWeek => 'Nessun appuntamento per più di 7 giorni';

  @override
  String get proLightNoClients => 'Nessun client condiviso';

  @override
  String get proLightNoPets => 'Nessun animale condiviso';

  @override
  String get proLightAddress => 'Indirizzo';

  @override
  String get proLightOpenMaps => 'Mappe';

  @override
  String get proLightReportTitle => 'Rapporto';

  @override
  String get proLightReportHint => 'Note di visita...';

  @override
  String get proLightImproveAi => 'Migliora (AI)';

  @override
  String get proLightFinalizeReport => 'Finalizzare';

  @override
  String get proLightNewConsultation => 'Nuova consultazione';

  @override
  String get proLightConsultationNextTitle => 'Consultazione salvata — dopo?';

  @override
  String get proLightConsultationCtaDaf => 'Crea DAF e fattura';

  @override
  String get proLightConsultationCtaInvoice => 'Fattura direttamente';

  @override
  String get proLightConsultationCtaDone => 'Termina';

  @override
  String get proLightReportFinal => 'Finalizzato';

  @override
  String get proLightReportHistoryTitle => 'Storico';

  @override
  String get proLightReportHistoryTranscript => 'Originale (trascrizione)';

  @override
  String get proLightReportHistoryImproved => 'Versione AI';

  @override
  String get proLightReportHistorySaved => 'Versione registrata';

  @override
  String get proLightReportHistoryEmpty => 'Nessuna versione disponibile';

  @override
  String get proLightSettings => 'Impostazioni';

  @override
  String get proLightSpecialty => 'Specialità';

  @override
  String get proLightDocuments => 'Documenti';

  @override
  String get proLightNoDocuments => 'Nessun documento';

  @override
  String get proLightTimeline => 'Cronologia';

  @override
  String get proLightNoTimeline => 'Nessun evento';

  @override
  String get proLightReminders => 'Promemoria';

  @override
  String get proLightNoReminders => 'Nessun promemoria';

  @override
  String get proLightLitterTag => 'Etichetta/ambito';

  @override
  String get proLightActionFailed => 'Azione impossibile';

  @override
  String get proLightReadOnly => 'Accesso in sola lettura';

  @override
  String get petAccessSharedRead => '· Lettura condivisa';

  @override
  String get petAccessSharedNotes => '· Note condivise';

  @override
  String get petAccessSharedFull => 'Condiviso · completo';

  @override
  String get proLightUseGps => 'GPS';

  @override
  String get proLightTranscribeAudio => 'File audio';

  @override
  String get proLightDictationStart => 'Dettare';

  @override
  String get proLightDictationStop => 'Ferma';

  @override
  String get proLightRecordingInProgress => 'Registrazione in corso';

  @override
  String get proLightAudioConsentTitle => 'Consenso audio';

  @override
  String get proLightAudioConsentBody =>
      'La registrazione viene utilizzata solo per generare il report. Confermare l\'accordo orale del cliente. L\'audio viene rimosso una volta finalizzata la CR.';

  @override
  String get proLightAudioConsentAccept => 'Accetto';

  @override
  String get proLightSpecialtyFarrier => 'Maniscalco';

  @override
  String get proLightSpecialtyPhysio => 'Fisio / osteo';

  @override
  String get proLightSpecialtyBehaviorist => 'Comportamentista';

  @override
  String get proLightSpecialtyGroomer => 'Toelettatore';

  @override
  String get proLightSpecialtyBreeder => 'Allevatore';

  @override
  String get proLightSpecialtyVetLight => 'Veterinario leggero';

  @override
  String get proLightReportHintFarrier =>
      'CR ferratura: piedi, ferro, osservazioni...';

  @override
  String get proLightEmptyFarrier => 'Nessun cavallo / intervento condiviso';

  @override
  String get proLightMicDenied =>
      'Microfono negato: consenti l\'accesso nelle impostazioni';

  @override
  String get proLightGpsDenied => 'Posizione non disponibile';

  @override
  String get googleNotConfigured => 'Connessione Google non configurata';

  @override
  String get googleLoginFailed => 'Impossibile connettersi a Google';

  @override
  String get googleWrongAudience =>
      'Questo account Google è già un profilo Pro: utilizza l\'app Web.';

  @override
  String get myPets => 'I miei animali';

  @override
  String get myData => 'I miei dati';

  @override
  String get settings => 'Parametri';

  @override
  String get logout => 'Esci';

  @override
  String get save => 'Per salvaguardare';

  @override
  String get cancel => 'Cancellare';

  @override
  String get firstName => 'Il tuo nome';

  @override
  String get currentPassword => 'Password attuale';

  @override
  String get newPassword => 'Nuova password';

  @override
  String get confirmNewPassword => 'Conferma password';

  @override
  String get changePassword => 'Cambiare la password';

  @override
  String get forceChangePasswordTitle => 'Cambiare la password';

  @override
  String get forceChangePasswordSubtitle =>
      'Questo account è stato creato con una password temporanea. Scegli il tuo per continuare.';

  @override
  String get forceChangePasswordSubmit => 'Salva e continua';

  @override
  String get acceptTermsTitle => 'Condizioni d\'uso';

  @override
  String get acceptTermsSubtitle =>
      'Il tuo account è stato creato dalla clinica. Accetta le condizioni e l\'informativa sulla privacy per continuare.';

  @override
  String get acceptTermsSubmit => 'Accetta e continua';

  @override
  String get acceptTermsFailed => 'Impossibile salvare il consenso. Riprova.';

  @override
  String get passwordTooShort => 'Minimo 8 caratteri';

  @override
  String get passwordMismatch => 'Le password non corrispondono';

  @override
  String get passwordChangeFailed => 'Impossibile modificare la password';

  @override
  String get deleteAccount => 'Elimina account';

  @override
  String get deleteAccountConfirm =>
      'Questa azione è irreversibile. Tutti i tuoi animali domestici e i tuoi dati verranno eliminati.';

  @override
  String get exportMyData => 'Esporta i miei dati';

  @override
  String get registerConsentPrefix => 'Accetto il';

  @override
  String get registerConsentMiddle => 'e il';

  @override
  String get registerConsentRequired =>
      'È necessario accettare i termini e l\'informativa sulla privacy.';

  @override
  String get registerInviteCode => 'Codice invito (opzionale)';

  @override
  String get registerInviteCodeHint =>
      'Inserisci il codice del QR / link commerciale';

  @override
  String get nearbyCommercialTitle => 'Commerciale vicino a te';

  @override
  String get nearbyCommercialHint =>
      'Senza un codice di invito, scegli un venditore nelle vicinanze (facoltativo).';

  @override
  String get nearbyCommercialUseLocation => 'Usa la mia posizione';

  @override
  String get nearbyCommercialPostalCode => 'Codice postale';

  @override
  String get nearbyCommercialSearch => 'Fare ricerca';

  @override
  String get nearbyCommercialEmpty =>
      'Nessun annuncio pubblicitario trovato nelle vicinanze.';

  @override
  String get nearbyCommercialSkip => 'Non riattaccare';

  @override
  String get nearbyCommercialGeoDenied =>
      'Posizione negata: inserisci un codice postale.';

  @override
  String nearbyCommercialDistance(String km) {
    return '$km km';
  }

  @override
  String get pushPermissionTitle => 'Notifiche';

  @override
  String get pushPermissionBody =>
      'petsFollow vuole inviarti notifiche: messaggi dal tuo veterinario, conferme di appuntamenti e promemoria di cure. Puoi disattivarli in qualsiasi momento nell\'app o nelle impostazioni del telefono.';

  @override
  String get pushPermissionContinue => 'Continuare';

  @override
  String exportDataSaved(String path) {
    return 'Esportazione salvata: $path';
  }

  @override
  String get profileSaved => 'Profilo salvato';

  @override
  String get changePhoto => 'Cambia foto';

  @override
  String get addPhoto => 'Aggiungi una foto';

  @override
  String get photoUpdated => 'Foto aggiornata';

  @override
  String get passwordChanged => 'La password è cambiata';

  @override
  String greeting(String name) {
    return 'Ciao $name,';
  }

  @override
  String get latestValues => 'Ultimi valori';

  @override
  String get startMeasurement => 'INIZIA LA MISURAZIONE';

  @override
  String get heartRateShort => 'Respiro';

  @override
  String get weightShort => 'Peso';

  @override
  String get recordWeightTitle => 'Registrare il peso';

  @override
  String get weightKgLabel => 'Peso (kg)';

  @override
  String get weightCommentLabel => 'Commento (facoltativo)';

  @override
  String get weightCommentHint => 'Ex. dopo una passeggiata, a stomaco vuoto…';

  @override
  String get weightSave => 'Salva';

  @override
  String get weightSentToVet => 'Peso registrato';

  @override
  String get weightInvalid => 'Inserisci un peso valido (0,01–999,99 kg)';

  @override
  String weightLastLabel(String kg) {
    return 'Ultimo peso: $kg kg';
  }

  @override
  String get choosePetForMeasurement => 'Scegli un animale';

  @override
  String get chooseDuration => 'Durata della misurazione';

  @override
  String durationSeconds(int seconds) {
    return '${seconds}s';
  }

  @override
  String get howToMeasure => 'Come misurare?';

  @override
  String get howToMeasureIntro =>
      'Misura la frequenza respiratoria del tuo animale domestico a riposo.';

  @override
  String get howToMeasureStep1 =>
      '1. Metti il ​​tuo animale domestico in un posto tranquillo, sdraiato o seduto.';

  @override
  String get howToMeasureStep2 =>
      '2. Metti la mano sul petto e picchietta ad ogni respiro per la durata indicata.';

  @override
  String get howToMeasureStep3 =>
      '3. Convalida il rapporto per inviarlo al tuo veterinario.';

  @override
  String get howToMeasureWhyTitle => 'Perché misurare?';

  @override
  String get howToMeasureWhyBody =>
      'Il monitoraggio regolare della frequenza respiratoria consente di rilevare eventuali variazioni e di adattare il trattamento con il veterinario.';

  @override
  String get reminders => 'Promemoria';

  @override
  String get remindersHint =>
      'Ricevi un promemoria quotidiano per una lettura respiratoria.';

  @override
  String get remindersEnabled => 'Abilita promemoria';

  @override
  String get remindersTime => 'Tempo di promemoria';

  @override
  String get remindersSaved => 'Promemoria salvati';

  @override
  String get legalTermsTitle => 'Condizioni generali d\'uso';

  @override
  String get legalPrivacyTitle => 'politica sulla riservatezza';

  @override
  String get legalNoticeTitle => 'Avvisi legali';

  @override
  String get legalOpenOnline => 'Visualizza la versione online';

  @override
  String get legalTermsBody =>
      'Condizioni generali d\'uso — petsFollow\n\nL\'app petsFollow consente ai proprietari di animali domestici di monitorare prescritti (messaggi, promemoria per cura/cavallo, letture respiratorie), visualizzare la cronologia e comunicare con il proprio veterinario.\n\nI servizi sono forniti come parte dell\'abbonamento scelto (pagamento tramite Stripe). L\'utente si impegna a utilizzare l\'applicazione in conformità con lo scopo previsto.\n\nVersione completa: https://petsfollow.ll-it-sc.be/legal/terms\n\nData aggiornata: luglio 2026';

  @override
  String get legalPrivacyBody =>
      'Informativa sulla privacy — petsFollow\n\nDati raccolti: identità (nome, email), dati degli animali (nome, specie, razza, foto), letture della frequenza respiratoria (dati sulla salute degli animali), messaggi e media scambiati con lo studio, resoconti delle visite (registrazioni di testo e audio), coordinate GPS delle visite a domicilio (professionisti dell\'assistenza), token di notifica (FCM), dati di pagamento elaborati da Stripe.\n\nFinalità: gestione account, continuità delle cure (comprese letture respiratorie), messaggistica veterinaria, resoconti visite, notifiche, fatturazione.\n\nElaborazione AI: Google Gemini viene utilizzato per migliorare i report sulle visite (audio elaborato in tempo reale, non archiviato da Google) e, nell\'app cliente (modulo sperimentale), per spiegare i referti finalizzati e offrire un triage conversazionale di urgenza.\n\nSubappaltatori/partner: Google (Sign-In, Firebase Cloud Messaging, Gemini), Stripe (pagamenti), cloud hosting (GCP).\n\nConservazione: fino alla cancellazione dell\'account; account inattivi eliminati dopo 3 anni; audio delle relazioni conservate per tutta la durata del fascicolo.\n\nDiritti GDPR (accesso, rettifica, cancellazione, portabilità): Profilo → Esporta i miei dati / Elimina account o contatta barbara@petsfollow.app.\n\nVersione completa: https://petsfollow.ll-it-sc.be/legal/privacy\n\nData aggiornata: luglio 2026';

  @override
  String get legalNoticeBody =>
      'Note legali — petsFollow\n\nEditore: LL-IT-SC / petsFollow\nContatto: barbara@petsfollow.app · 0478.02.33.77\n\nHosting: Google Cloud Platform (conformità GDPR).\n\nDirettore della pubblicazione: petsFollow.\n\nVersione completa: https://petsfollow.ll-it-sc.be/legal/mentions\n\nData aggiornata: luglio 2026';

  @override
  String get language => 'Lingua';

  @override
  String get languageFr => 'francese';

  @override
  String get languageNl => 'Paesi Bassi';

  @override
  String get languageEn => 'Inglese';

  @override
  String get languageEs => 'spagnolo';

  @override
  String get languageEt => 'Eesti';

  @override
  String get languageIt => 'Italiano';

  @override
  String get languageUk => 'Українська';

  @override
  String get languageRu => 'Русский';

  @override
  String get appearance => 'Aspetto';

  @override
  String get themeLight => 'Chiaro';

  @override
  String get themeDark => 'Scuro';

  @override
  String get planMonthlyLabel => '€ 3,50/mese';

  @override
  String get planAnnualLabel => '35€/anno';

  @override
  String get planTriennialLabel => '95 € / 3 anni';

  @override
  String get planQuinquennialLabel => '145 € / 5 anni';

  @override
  String get pushNewMessage => 'Nuovo messaggio';

  @override
  String get pushVisitConfirmed => 'Appuntamento confermato';

  @override
  String get pushVisitProposed => 'Suggerimento per l\'incontro';

  @override
  String get pushVisitReschedule => 'Viaggio su appuntamento';

  @override
  String get notifChannelMessages => 'Messaggi';

  @override
  String get notifChannelVisits => 'Visite';

  @override
  String get notifChannelCare => 'Cura';

  @override
  String get paymentResume => 'Riprendere il pagamento';

  @override
  String get manageSubscription => 'Gestisci il mio abbonamento';

  @override
  String get heartRate => 'Lettura respiratoria';

  @override
  String get history => 'Storico';

  @override
  String get vetMessaging => 'Messaggi veterinari';

  @override
  String get badgeAutoRenew => 'Rinnovo automatico';

  @override
  String get badgeActive => 'Attivo';

  @override
  String get badgePendingPayment => 'In attesa di pagamento';

  @override
  String badgeExpiresOn(String date) {
    return 'scade $date';
  }

  @override
  String get newPet => 'Nuovo animale';

  @override
  String get editPet => 'Modifica animale';

  @override
  String get petName => 'Nome';

  @override
  String get petNameRequired => 'Indica il nome dell’animale';

  @override
  String get species => 'Specie';

  @override
  String get breed => 'Razza';

  @override
  String get petMicrochipOptional => 'N. microchip (opzionale)';

  @override
  String get petHealthBookNumberOptional => 'N. libretto (opzionale)';

  @override
  String get petDomicileLocation => 'Domicilio / scuderia';

  @override
  String get petDomicileHint => 'Es. Scuderia dei Salici — Bruxelles';

  @override
  String get petFoodChainStatus => 'Stato catena alimentare';

  @override
  String get petFoodChainCompanion => 'Animale da compagnia (fuori catena)';

  @override
  String get petFoodChainFoodProducing =>
      'Animale da reddito / catena alimentare';

  @override
  String get petFoodChainExcluded => 'Escluso dalla catena alimentare';

  @override
  String get petHealthBookAddPages => 'Aggiungi foto del libretto';

  @override
  String get petHealthBookReplacePages => 'Sostituisci PDF (foto)';

  @override
  String petHealthBookPagesCount(int count) {
    return '$count pagina/e selezionata/e';
  }

  @override
  String get petHealthBookPdfAttached => 'PDF del libretto allegato';

  @override
  String get petHealthBookRemovePdf => 'Rimuovi';

  @override
  String get petHealthBookOpenPdf => 'Apri libretto (PDF)';

  @override
  String petMicrochipLabel(String number) {
    return 'Chip: $number';
  }

  @override
  String petHealthBookNumberLabel(String number) {
    return 'Libretto: $number';
  }

  @override
  String get errorHealthBookUploadFailed =>
      'Animale salvato, ma il libretto non è stato caricato';

  @override
  String get choosePlan => 'Scegli la tua formula';

  @override
  String get recommended => 'Raccomandato';

  @override
  String get autoRenewTitle => 'Rinnova automaticamente';

  @override
  String get autoRenewSubtitle => 'Addebito diretto ad ogni data di scadenza';

  @override
  String get continueToPayment => 'Salva e paga';

  @override
  String get petFormSave => 'Salva';

  @override
  String get petSavedPendingPayment =>
      'Animale salvato — attivalo per accedere alle funzionalità';

  @override
  String get paymentFeaturesLocked =>
      'Pagamento richiesto per usare le funzionalità di questo animale';

  @override
  String get paymentConfirmed => 'Pagamento confermato — animale attivo';

  @override
  String get paymentPending =>
      'Pagamento in sospeso: puoi riprendere più tardi';

  @override
  String errorGeneric(String message) {
    return 'Errore: $message';
  }

  @override
  String get errorNetwork =>
      'Impossibile connettersi. Controlla la tua rete e riprova.';

  @override
  String get retryAction => 'Riprova';

  @override
  String get errorMediaTooLarge => 'File troppo grande (25 MB massimo)';

  @override
  String get errorInvalidMediaType =>
      'Formato non supportato (JPEG, PNG, WebP, MP4, MOV, WebM)';

  @override
  String get errorPaymentRequired =>
      'Abbonamento richiesto per usare questa funzione';

  @override
  String get errorPhotoUploadFailed =>
      'Animale creato, ma non è stato possibile inviare la foto';

  @override
  String get errorCouldNotOpenLink => 'Impossibile aprire il collegamento';

  @override
  String get planMonthlySub => '€ 3,50/mese, rinnovo automatico';

  @override
  String planAnnualSub(String price) {
    return '$price, rinnovato automaticamente';
  }

  @override
  String get planTriennialSub => '95€ ogni 3 anni, rinnovata automaticamente';

  @override
  String get planQuinquennialSub => '€145 per 5 anni, unica soluzione';

  @override
  String planOneTime(String price) {
    return '$price, pagamento unico';
  }

  @override
  String get heartRateInstructions =>
      'Picchietta con ogni respiro per la durata indicata dal tuo veterinario.';

  @override
  String heartRateInstructionsDuration(int seconds) {
    return 'Tocca con ogni respiro per $seconds secondi.';
  }

  @override
  String get heartRateNoDurationConfigured =>
      'Per questo studio non è configurata alcuna durata di misurazione. Contatta il tuo veterinario.';

  @override
  String get heartRateNotSupported =>
      'La misurazione della frequenza respiratoria non è disponibile per questa specie';

  @override
  String get start => 'Per iniziare';

  @override
  String secondsLeft(int seconds) {
    return '${seconds}s';
  }

  @override
  String beatsCount(int count) {
    return '$count respiri';
  }

  @override
  String get tapHere => 'Tocca qui su ogni respiro';

  @override
  String bpmLabel(String bpm) {
    return 'B/M: $bpm';
  }

  @override
  String beatsLabel(int count) {
    return 'Respiri: $count';
  }

  @override
  String get thresholdAlert =>
      'Avviso: aumento significativo rispetto alla lettura precedente';

  @override
  String get validateAndSend => 'Convalidare e inviare al veterinario';

  @override
  String get heartRateCommentLabel => 'Commento (facoltativo)';

  @override
  String get heartRateCommentHint =>
      'Ex. agitato, a riposo, dopo l\'esercizio...';

  @override
  String get restart => 'Ricominciare';

  @override
  String get sentToVet => 'Dichiarazione inviata al veterinario';

  @override
  String get navHome => 'Benvenuto';

  @override
  String get navPets => 'Animali';

  @override
  String get navCare => 'Cura';

  @override
  String get navMessages => 'Messaggi';

  @override
  String get navProfile => 'Profilo';

  @override
  String get speciesDog => 'Cane';

  @override
  String get speciesCat => 'Gatto';

  @override
  String get speciesHorse => 'Cavallo';

  @override
  String get speciesDonkey => 'Asino';

  @override
  String get speciesCattle => 'Bovino';

  @override
  String get speciesSheep => 'Ovino';

  @override
  String get speciesGoat => 'Caprino';

  @override
  String get speciesPig => 'Suino';

  @override
  String get speciesPoultry => 'Pollame';

  @override
  String get speciesRabbit => 'Coniglio';

  @override
  String get speciesAlpaca => 'Alpaca';

  @override
  String get speciesLlama => 'Llama';

  @override
  String get speciesOther => 'Altro';

  @override
  String get careComingSoon => 'I promemoria per la cura arriveranno presto';

  @override
  String get emptyPetsTitle => 'Nessun animale';

  @override
  String get emptyPetsBody =>
      'Aggiungi il tuo primo animale per iniziare il follow-up prescritto dal tuo veterinario.';

  @override
  String get discoveryTitle => 'Scopri petsFollow';

  @override
  String get homeRecentActivity => 'Attività recente';

  @override
  String get discoveryMission => 'Il tuo percorso petsFollow';

  @override
  String get discoveryDay0Title => 'Passo 1 — Benvenuto';

  @override
  String get discoveryDay0Body =>
      'Crea il profilo del tuo animale domestico ed esplora l\'app: messaggi, promemoria e letture (inclusa la frequenza respiratoria).';

  @override
  String get discoveryDay2Title => 'Passo 2 — Prima misurazione';

  @override
  String get discoveryDay2Body =>
      'Effettua la prima lettura della frequenza respiratoria e acquisisci familiarità con la tecnica.';

  @override
  String get discoveryDay4Title => 'Passo 3 – Routine';

  @override
  String get discoveryDay4Body =>
      'Imposta una routine di misurazione quotidiana con promemoria personalizzati.';

  @override
  String get discoveryDay6Title => 'Passo 4 – Condivisione del veto';

  @override
  String get discoveryDay6Body =>
      'Le tue letture vengono condivise con il tuo veterinario per un monitoraggio ottimale.';

  @override
  String get myVets => 'I miei veterinari';

  @override
  String get addVetByEmail => 'Aggiungi un veterinario via e-mail';

  @override
  String get vetEmailHint => 'email@cabinet.vet';

  @override
  String get noVets => 'Nessun veterinario collegato';

  @override
  String get vetLinkRequired =>
      'Collega un veterinario per attivare il follow-up con il tuo studio';

  @override
  String get linkVetAfterSaveTitle => 'Collega un veterinario';

  @override
  String get linkVetAfterSaveBody =>
      'Collega uno studio per attivare messaggi, visite e promemoria di cura.';

  @override
  String get linkVetHomeTitle => 'Collegare un veterinario?';

  @override
  String get linkVetHomeBody =>
      'Il tuo animale è salvato. Vuoi collegare un veterinario? È facoltativo — potrai farlo più tardi.';

  @override
  String get linkVetLater => 'Più tardi';

  @override
  String get primaryVet => 'Veterinario senior';

  @override
  String get setPrimaryVet => 'Imposta come veterinario principale';

  @override
  String get careTitle => 'Cura';

  @override
  String get careDone => 'Fare';

  @override
  String get carePostpone => 'Rimandare';

  @override
  String get careOverdue => 'Tardi';

  @override
  String get visitHistory => 'Visita la storia';

  @override
  String get requestVisit => 'Richiedi una visita';

  @override
  String get calendarBookingDisabled =>
      'Per questa pratica non è disponibile la prenotazione online. Chiama in ufficio per fissare un appuntamento.';

  @override
  String get calendarBookingDisabledReschedule =>
      'La prenotazione online non è disponibile. Suggerisci una data manualmente.';

  @override
  String get calendarNoSlots =>
      'Nessuno spazio disponibile nei prossimi 14 giorni.';

  @override
  String get calendarPickSlot => 'Scegli una nicchia:';

  @override
  String get calendarSelectVet => 'Scegli un veterinario:';

  @override
  String get calendarCallPractice => 'Chiama lo studio';

  @override
  String get calendarNoPhone =>
      'Per questo ufficio non è fornito alcun numero di telefono. Contattalo in un altro modo.';

  @override
  String get visitConfirm => 'Confermare';

  @override
  String get visitProposeReschedule => 'Suggerisci un altro slot';

  @override
  String get visitRescheduleProposed => 'Proposta di viaggio inviata';

  @override
  String get paymentSuccessSnack => 'Pagamento ricevuto: aggiorna...';

  @override
  String get paymentCancelSnack => 'Pagamento annullato';

  @override
  String get visitRejectReschedule => 'Rifiuta il trasloco';

  @override
  String get visitAcceptReschedule => 'Accetta il nuovo slot';

  @override
  String get upcomingVisit => 'Prossima visita';

  @override
  String get notificationHrTitle => 'petsFollow';

  @override
  String get notificationHrBody =>
      'È ora di fare una lettura respiratoria per il tuo animale domestico';

  @override
  String get reviewAskTitle => 'Ti piacciono gli petsFollow?';

  @override
  String get reviewAskYes => 'Sì, valuta l\'app';

  @override
  String get reviewAskNo => 'Dopo';

  @override
  String get careTypeMedication => 'Medicinale';

  @override
  String get horseAddContact => 'Aggiungi un contatto';

  @override
  String get horseAddCompetition => 'Aggiungi una competizione';

  @override
  String get horseContactName => 'Nome';

  @override
  String get horseContactRole => 'Ruolo';

  @override
  String get horseCompetitionTitle => 'Evento';

  @override
  String get horseCompetitionDate => 'Data (AAAA-MM-GG)';

  @override
  String familyHouseholdTitle(int count) {
    return 'Casa Famiglia — $count animali';
  }

  @override
  String kennelHouseholdTitle(int count) {
    return 'Bestiame domestico — $count animali';
  }

  @override
  String get familyHouseholdNext => 'Prossimi promemoria per la casa';

  @override
  String get familyPetLimit =>
      'Un pacchetto famiglia è già attivo o in fase di acquisto';

  @override
  String get familyRequiresTwoPets =>
      'Il pacchetto Famiglia richiede almeno 2 animali';

  @override
  String get kennelPackHint =>
      'Pacchetto bestiame: ≥6 animali, −15% sugli abbonamenti successivi';

  @override
  String get kennelRequiresSixPets =>
      'Il pacchetto di allevamento richiede almeno 6 animali';

  @override
  String get kennelQuickEncodeTitle => 'Codifica della gamma (allevamento)';

  @override
  String get kennelRequired =>
      'Il pacchetto di allevamento è richiesto per la codifica batch';

  @override
  String get litterTag => 'Ambito dell\'etichetta';

  @override
  String get petBirthDate => 'Data di nascita';

  @override
  String get petBirthDateInvalid => 'Data di nascita non valida (AAAA-MM-GG)';

  @override
  String get discoveryMarkDone => 'Missione compiuta';

  @override
  String get discoveryDismiss => 'Nascondi il percorso';

  @override
  String get appBuildInfo => 'Versione dell\'app';

  @override
  String get notificationPreferences => 'Preferenze notifiche';

  @override
  String get notificationPrefsHint =>
      'Scegli i tipi di notifiche che desideri ricevere.';

  @override
  String get notificationPrefsSaved => 'Preferenze salvate';

  @override
  String get notificationPrefHr => 'Letture respiratorie';

  @override
  String get notificationPrefCare => 'Promemoria per la cura';

  @override
  String get notificationPrefVisits => 'Visite';

  @override
  String get notificationPrefMessages => 'Messaggi';

  @override
  String get notificationPrefDiscovery => 'Sentiero di scoperta';

  @override
  String get notificationPrefBilling => 'Fatturazione';

  @override
  String get notificationPrefSms => 'SMS';

  @override
  String get notificationPrefSmsHint =>
      'Conferme e promemoria degli appuntamenti via SMS. Rispondere STOP disattiva anche questo canale.';

  @override
  String carePostponeDays(int days) {
    return 'Rinviare $days giorni';
  }

  @override
  String get noCareReminders => 'Nessun promemoria di cura attuale';

  @override
  String get careAddReminder => 'Aggiungi un promemoria';

  @override
  String get careSelectPet => 'Animale';

  @override
  String careDueInDays(int days) {
    return 'Scadenza tra $days giorni';
  }

  @override
  String get careReferenceModeDone => 'Già fatto';

  @override
  String get careReferenceModeFirst => 'Prima volta';

  @override
  String get careLastDateLabel => 'Ultimo appuntamento';

  @override
  String get careLastDateDone => 'Data dell\'ultimo trattamento';

  @override
  String get careLastDateFirst => 'Data di inizio ciclo';

  @override
  String get careRecurrenceLabel => 'Ricorrenza';

  @override
  String get careRecurrenceNone => 'Nessuno (scadenza unica)';

  @override
  String careRecurrenceDays(int days) {
    return 'Ogni $days giorni';
  }

  @override
  String get careDueDateLabel => 'Scadenza';

  @override
  String get careDueDateComputed => 'Data di scadenza calcolata';

  @override
  String get careTooltipDoneWithRecurrence =>
      'Trattamento già effettuato: data di scadenza = data dell\'ultimo trattamento + recidiva.';

  @override
  String get careTooltipFirstWithRecurrence =>
      'Prima programmazione: indicare la data di inizio del ciclo. La scadenza = questa data + ricorrenza.';

  @override
  String get careTooltipNoRecurrence =>
      'Senza ricorrenza: la data inserita è l\'unica data di scadenza.';

  @override
  String get careTooltipDueExplained =>
      'Data di scadenza = ultima data + ricorrenza (se definita).';

  @override
  String get carePickDate => 'Scegli una data';

  @override
  String discoveryDayBadge(int day) {
    return 'P$day';
  }

  @override
  String get timelineTypeHeartrate => 'Frequenza respiratoria';

  @override
  String get timelineTypeWeight => 'Peso';

  @override
  String get timelineTypeMessage => 'Messaggio';

  @override
  String get timelineTypeCare => 'Cura';

  @override
  String get timelineTypeVisit => 'Visita';

  @override
  String get timelineTypeEvent => 'Evento';

  @override
  String get visitCancelAction => 'Annulla richiesta';

  @override
  String get upcomingVisits => 'Prossime visite';

  @override
  String get timelineEmpty => 'Nessun evento al momento';

  @override
  String get noThreads => 'Nessuna conversazione';

  @override
  String get messageNoMessagesYet => 'Ancora nessun messaggio';

  @override
  String get messageNewConversation => 'Nuova conversazione';

  @override
  String get messageComposeTitle => 'Nuova conversazione';

  @override
  String get messageChoosePro => 'Professionista della cura';

  @override
  String get messageChooseClient => 'Cliente';

  @override
  String get messageChoosePet => 'Animale interessato';

  @override
  String get messageChoosePetOptional => 'Animale (opzionale)';

  @override
  String get messageGeneralThread => 'Conversazione generale';

  @override
  String get messageStartConversation => 'Avvia';

  @override
  String get messageLockedTitle => 'Messaggistica non disponibile';

  @override
  String get messageLockedBody =>
      'Collega un veterinario per chattare con un professionista della cura.';

  @override
  String get vetInviteSent =>
      'Invito inviato: lo studio deve accettare la richiesta';

  @override
  String vetInviteSentNamed(String practice) {
    return 'Richiesta inviata a $practice: lo studio deve accettarla';
  }

  @override
  String get vetNotFound => 'Nessun veterinario trovato con questa email';

  @override
  String get addVetSearchHint =>
      'Cerca per nome, e-mail o studio. Se è già su petsFollow, viene inviata una richiesta di collegamento allo studio.';

  @override
  String get addVetSearchLabel => 'Cerca un veterinario';

  @override
  String get addVetSearchFieldHint => 'Nome, e-mail o studio';

  @override
  String get addVetNotListed => 'Il mio veterinario non è in elenco';

  @override
  String get addVetSuggestTitle => 'Nuovo veterinario';

  @override
  String get addVetSuggestBody =>
      'Indichi e-mail e telefono dello studio. Li contatteremo affinché possano unirsi a petsFollow.';

  @override
  String get addVetSuggestEmail => 'E-mail dello studio';

  @override
  String get addVetSuggestPhone => 'Telefono';

  @override
  String get addVetSuggestNameOptional => 'Nome del veterinario (facoltativo)';

  @override
  String get addVetSuggestCta => 'Invia suggerimento';

  @override
  String get vetSuggestSent =>
      'Grazie — contatteremo lo studio. Sarà avvisato quando si unirà a petsFollow.';

  @override
  String get visitRequested => 'Richiesta di visita inviata';

  @override
  String get primaryVetSet => 'Veterinario senior aggiornato';

  @override
  String get visitStatusRequested => 'Richiesto';

  @override
  String get visitStatusConfirmed => 'Confermato';

  @override
  String get visitStatusDone => 'Completato';

  @override
  String get visitStatusCancelled => 'Annullato';

  @override
  String get visitStatusReschedulePending => 'In attesa del viaggio';

  @override
  String get horseHealthTitle => 'Salute equina';

  @override
  String get horseContactsTitle => 'Contatti (maniscalco, dentista, ecc.)';

  @override
  String get horseCompetitionsTitle => 'Concorsi';

  @override
  String get horseContactsSoon =>
      'Attiva il Horse Pack per gestire i tuoi contatti professionali.';

  @override
  String get horseCompetitionsSoon =>
      'Attiva il pacchetto cavalli per il calendario delle gare.';

  @override
  String get horsePackUpsell =>
      'Horse Pack – maniscalco, coproscopia, contatti e gare';

  @override
  String get careTypeFarrier => 'Maniscalco';

  @override
  String get careTypeFecalEgg => 'Coproscopia';

  @override
  String get careTypeVaccination => 'Vaccinazione';

  @override
  String get careTypeDeworming => 'Vermifugo';

  @override
  String get careTypeVetCheck => 'Controllo veterinario';

  @override
  String get careTypeDental => 'Cure dentistiche';

  @override
  String get careTypeCustom => 'Promemoria personalizzato';

  @override
  String get homeAddFirstVetTitle => 'Aggiungi il tuo veterinario';

  @override
  String get homeAddFirstVetBody =>
      'Collega lo studio che segue il tuo animale per condividere letture e discutere.';

  @override
  String get homeAddFirstVetCta => 'Aggiungi un veterinario';

  @override
  String get photoFrameHint =>
      'Inquadra il muso al centro — anteprima scheda animale';

  @override
  String get takePhoto => 'Scatta una foto';

  @override
  String get takeVideo => 'Registra un video';

  @override
  String get chooseFromGallery => 'Scegli dalla galleria';

  @override
  String get attachMedia => 'Allega una foto o un video';

  @override
  String get attachPhoto => 'Foto';

  @override
  String get attachVideo => 'Video';

  @override
  String get compressingMedia => 'Compressione del video…';

  @override
  String get openMedia => 'Aprire';

  @override
  String get mediaVideoLabel => 'Video';

  @override
  String get appInviteTitle => 'Applicazione di invito QR';

  @override
  String get appInviteHint =>
      'Visualizza questo QR o condividi il collegamento. Un nuovo cliente che si registra tramite questo link viene automaticamente collegato.';

  @override
  String get appInviteHintClient =>
      'Condividi questo QR con un amico. Sarà collegato a te (referral) e potrà unirsi al tuo studio se non ne ha ancora uno.';

  @override
  String get appInviteHintCommercial =>
      'Due link: clienti (app) e cliniche (registrazione Pro con il tuo codice referral).';

  @override
  String get appInviteHintShort => 'Link per il download e collegamento';

  @override
  String get appInviteCodeLabel => 'Codice:';

  @override
  String get appInviteCopy => 'Copia collegamento';

  @override
  String get appInviteCopied => 'Collegamento copiato';

  @override
  String get appInviteLoadError => 'Impossibile caricare il QR';

  @override
  String get appInviteRetry => 'Riprova';

  @override
  String get proLightVetTitle => 'Campo veterinario';

  @override
  String get commercialFieldTitle => 'Commerciale';

  @override
  String get commercialFieldSubtitle =>
      'Invito cliente QR e accesso al sito Pro.';

  @override
  String get commercialManagerFieldSubtitle =>
      'Risultati del team, invito QR e accesso al sito Pro.';

  @override
  String get commercialOpenProWeb => 'Apri il sito Pro';

  @override
  String get managerTeamCta => 'Risultati del mio team';

  @override
  String get managerTeamTitle => 'Team';

  @override
  String get managerTeamSection => 'Risultati team';

  @override
  String get managerSelfSection => 'I miei risultati';

  @override
  String get managerMembersSection => 'Commerciali';

  @override
  String get managerTeamEmpty => 'Nessun commerciale assegnato.';

  @override
  String get managerKpiProspects => 'Prospect';

  @override
  String get managerKpiConverted => 'Convertiti';

  @override
  String get managerKpiConversion => 'Tasso di conversione';

  @override
  String get managerKpiAppointments => 'Appuntamenti in arrivo';

  @override
  String get managerKpiStale => 'Pipeline ferma';

  @override
  String get managerKpiMonthEarned => 'Commissioni (mese)';

  @override
  String get managerKpiLifetimeEarned => 'Commissioni (totale)';

  @override
  String get managerKpiVets => 'Veterinari assegnati';

  @override
  String get featureModules => 'Opzioni';

  @override
  String get featureModulesCatalog => 'Scopri le opzioni';

  @override
  String get featureModulesSubtitle =>
      'Attiva Care+, Horse, Kennel o Foyer a seconda delle tue esigenze.';

  @override
  String get moduleCarePlus => 'Care+';

  @override
  String get moduleCarePlusDesc =>
      'Promemoria di cura arricchiti per tutti i tuoi animali domestici.';

  @override
  String get moduleHorse => 'cavallo';

  @override
  String get moduleHorseDesc =>
      'Contatti professionali e concorsi per cavalli.';

  @override
  String get moduleKennel => 'Canile';

  @override
  String get moduleKennelDesc => 'Codifica rapida della riproduzione/figliata.';

  @override
  String get moduleFamily => 'Focolare';

  @override
  String get moduleFamilyDesc => 'Visualizzazione iniziale dei tuoi animali.';

  @override
  String get moduleActivate => 'Abilitare';

  @override
  String get moduleActive => 'Attivo';

  @override
  String get switchProfile => 'Cambia profilo';

  @override
  String get profilePersonal => 'Personale';

  @override
  String get profilePro => 'Professionale';

  @override
  String get profileSwitched => 'Profilo inclinato';

  @override
  String get preconsultTitle => 'Pre-consultazione';

  @override
  String preconsultTitlePet(String petName) {
    return 'Preconsultazione — $petName';
  }

  @override
  String get preconsultIntro =>
      'Indica lo stato del tuo animale domestico prima della visita. Il tuo veterinario può prepararsi per questo.';

  @override
  String get preconsultComplaint => 'Motivo/reclamo principale';

  @override
  String get preconsultComplaintRequired => 'Indicare il motivo della visita';

  @override
  String get preconsultDuration => 'Da quando?';

  @override
  String get preconsultDurationToday => 'Oggi';

  @override
  String get preconsultDurationFewDays => 'Pochi giorni';

  @override
  String get preconsultDurationWeek => 'Circa una settimana';

  @override
  String get preconsultDurationWeeks => 'Diverse settimane';

  @override
  String get preconsultDurationMonths => 'Diversi mesi';

  @override
  String get preconsultBehavior => 'Comportamento';

  @override
  String get preconsultBehaviorNormal => 'Normale';

  @override
  String get preconsultBehaviorLethargic => 'Apatico';

  @override
  String get preconsultBehaviorRestless => 'Irrequieto';

  @override
  String get preconsultBehaviorAggressive => 'Aggressivo';

  @override
  String get preconsultBehaviorAnxious => 'Ansioso';

  @override
  String get preconsultBehaviorOther => 'Altro';

  @override
  String get preconsultAppetite => 'Appetito';

  @override
  String get preconsultThirst => 'Assetato';

  @override
  String get preconsultElimination => 'Feci/urina';

  @override
  String get preconsultScaleNormal => 'Normale';

  @override
  String get preconsultScaleDecreased => 'Diminuito';

  @override
  String get preconsultScaleIncreased => 'Aumento';

  @override
  String get preconsultUrgency => 'Urgenza percepita';

  @override
  String get preconsultUrgencyLow => 'Debole';

  @override
  String get preconsultUrgencyMedium => 'Media';

  @override
  String get preconsultUrgencyHigh => 'Alto';

  @override
  String get preconsultComment => 'Commento (facoltativo)';

  @override
  String get preconsultUnknown => 'Non lo so';

  @override
  String get preconsultSubmit => 'Inviare';

  @override
  String get preconsultSubmitted => 'Preconsultazione inviata';

  @override
  String get preconsultAlreadySubmitted =>
      'Hai già inviato questa pre-consultazione.';

  @override
  String get preconsultFillCta => 'Completa la pre-consultazione';

  @override
  String get proLightAudioConsentClientCheck =>
      'Ho ottenuto l\'accordo orale da parte del cliente per la registrazione';

  @override
  String get proLightReportAiProposalBanner =>
      'Proposta di IA: convalida obbligatoria prima della finalizzazione (diagnosi/farmaci inclusi).';

  @override
  String get proLightAiModuleRequired =>
      'Funzione CR IA disattivata per questo studio — contatta il supporto petsFollow.';

  @override
  String proLightAiModuleTrialBanner(int days) {
    return 'Test CR IA — $days giorni rimanenti. Detta quindi migliora i tuoi rapporti.';
  }

  @override
  String get proLightAiModuleInactiveBanner =>
      'CR AI non attivata per questo studio. La dettatura AI non sarà disponibile.';

  @override
  String get proLightAiModuleVisitScopedBanner =>
      'CR IA disponibile se lo studio visite ha il modulo attivato.';

  @override
  String get supportTitle => 'Segnala un problema';

  @override
  String get supportHint =>
      'Descrivi il bug. I diagnostici tecnici degli ultimi 15 minuti sono allegati automaticamente.';

  @override
  String get supportSubject => 'Oggetto';

  @override
  String get supportMessage => 'Descrizione';

  @override
  String get supportDiagnosticsAttached =>
      'Diagnostici allegati automaticamente (errori, richieste, configurazione).';

  @override
  String get supportSubmit => 'Invia';

  @override
  String get supportSending => 'Invio…';

  @override
  String get supportSuccess => 'Messaggio inviato. Grazie!';

  @override
  String get supportErrorTooLarge =>
      'Diagnostici troppo grandi. Riavvia l\'app e riprova.';

  @override
  String get supportMenu => 'Supporto';

  @override
  String get appInviteHintSales =>
      'Condividi il codice sponsor con uno studio o un cliente.';

  @override
  String get appInviteCopyVet => 'Copia link iscrizione studio';

  @override
  String get appInviteCopyClient => 'Copia link invito cliente';

  @override
  String get sendDossierToPro => 'Invia a un professionista';

  @override
  String get sendDossierEmailLabel => 'Email del professionista';

  @override
  String get sendDossierEmailHint => 'vet@clinica.it';

  @override
  String get sendDossierConfirm => 'Invia';

  @override
  String get sendDossierSuccess => 'Cartella inviata — link valido 24 ore.';

  @override
  String get sendDossierInvalidEmail => 'Indirizzo email non valido.';

  @override
  String get sendDossierPhiWarning =>
      'Questa cartella contiene dati sanitari: referti delle visite, libretto sanitario e documenti. Il link resta valido 24 ore e chiunque lo possieda potrà consultarli.';

  @override
  String get sendDossierPhiConsent =>
      'Accetto di condividere questi dati sanitari con questo professionista.';

  @override
  String get consultationsHistory => 'Consultazioni';

  @override
  String get consultationTitle => 'Consultazione';

  @override
  String consultationTitleWithPet(String petName) {
    return 'Consultazione — $petName';
  }

  @override
  String get consultationVisitMeta => 'Visita';

  @override
  String consultationReportBy(String author) {
    return 'Referto di $author';
  }

  @override
  String get consultationReportFallback => 'Referto';

  @override
  String get consultationReportEmpty => '(vuoto)';

  @override
  String get consultationAvailableCta => 'Disponibile';

  @override
  String get consultationPendingCta => 'Bozza';

  @override
  String get sendConsultationToVet => 'Invia a un veterinario';

  @override
  String get sendConsultationEmailLabel => 'Email del veterinario';

  @override
  String get sendConsultationEmailHint => 'vet@clinica.it';

  @override
  String get sendConsultationConfirm => 'Invia';

  @override
  String get sendConsultationSuccess =>
      'Consultazione inviata — link valido 24 ore.';

  @override
  String get sendConsultationInvalidEmail => 'Indirizzo email non valido.';

  @override
  String get sendConsultationPhiWarning =>
      'Questo referto contiene dati sanitari. Il link resta valido 24 ore e chiunque lo possieda potrà scaricare il PDF.';

  @override
  String get sendConsultationPhiConsent =>
      'Accetto di condividere questo referto con questo veterinario.';

  @override
  String get clientAiDevBadge => 'dev';

  @override
  String get clientAiSectionTitle => 'Assistenza IA';

  @override
  String get clientAiExplainCta => 'Capire il mio referto';

  @override
  String get clientAiExplainTitle => 'Il tuo referto spiegato';

  @override
  String get clientAiExplainDisclaimer =>
      'Questo non è un parere medico. Segui sempre le istruzioni del veterinario.';

  @override
  String get clientAiExplainLoading => 'Preparazione della spiegazione…';

  @override
  String get clientAiExplainListTitle => 'Capire un referto';

  @override
  String get clientAiExplainListSubtitle =>
      'Spiegazione semplificata di un referto finalizzato';

  @override
  String get clientAiExplainListEmpty =>
      'Nessun referto finalizzato da spiegare per ora.';

  @override
  String get clientAiTriageTitle => 'Aiuto urgenza 24/7';

  @override
  String get clientAiTriageSubtitle =>
      'Descrivi la situazione — valutiamo il grado di urgenza.';

  @override
  String get clientAiTriageHint => 'Es.: il mio cane ha mangiato cioccolato…';

  @override
  String get clientAiTriageSend => 'Invia';

  @override
  String get clientAiTriageLevelGreen => 'Consiglio';

  @override
  String get clientAiTriageLevelOrange => 'Prenota visita';

  @override
  String get clientAiTriageLevelRed => 'Urgenza';

  @override
  String get clientAiTriageWatchSigns => 'Segni da osservare';

  @override
  String get clientAiTriageCallPractice => 'Chiama la clinica';

  @override
  String get clientAiTriageBookVisit => 'Prenota una visita';

  @override
  String get clientAiTriageMessageVet => 'Contatta il veterinario';

  @override
  String get clientAiTriageSelectPet => 'Per quale animale?';

  @override
  String get clientAiTriageNoPet => 'Continua senza animale';

  @override
  String get clientAiTriageStart => 'Inizia';

  @override
  String get clientAiTriageEmergencyFallback =>
      'Se non riesci a contattare la clinica, chiama subito un servizio di guardia veterinaria locale.';

  @override
  String get clientAiTriageOpenMessages => 'Apri messaggi';

  @override
  String get bloodPressureShort => 'PA';

  @override
  String get recordBloodPressureTitle => 'Registra pressione';

  @override
  String get bpSystolicLabel => 'Sistolica (mmHg)';

  @override
  String get bpDiastolicLabel => 'Diastolica (mmHg)';

  @override
  String get bpMethodLabel => 'Metodo';

  @override
  String get bpMethodDoppler => 'Doppler';

  @override
  String get bpMethodOscillometric => 'Oscillometrico';

  @override
  String get bpMethodUnknown => 'Non specificato';

  @override
  String get bloodPressureInvalid => 'Inserire una PA valida (SYS ≥ DIA)';

  @override
  String get bloodPressureSaved => 'Pressione salvata';

  @override
  String get labsTitle => 'Analisi';

  @override
  String get labsEmpty => 'Nessuna analisi';

  @override
  String get labsResults => 'Risultati';

  @override
  String labsAbnormalCount(int count) {
    return '$count fuori norma';
  }

  @override
  String get labsOpen => 'Vedi analisi';

  @override
  String get labsFlagLow => 'Basso';

  @override
  String get labsFlagHigh => 'Alto';

  @override
  String get labsFlagNormal => 'Normale';

  @override
  String get labsOpenDocument => 'Apri documento';

  @override
  String get bpMethodInvasive => 'Invasivo';

  @override
  String get bpSiteLabel => 'Sito (facoltativo)';

  @override
  String get bpCommentLabel => 'Commento (facoltativo)';

  @override
  String get bloodPressureSave => 'Salva';

  @override
  String get calendarSelectSite => 'Scegli una sede';

  @override
  String get proformaValidateCta => 'Valida preventivo';
}
