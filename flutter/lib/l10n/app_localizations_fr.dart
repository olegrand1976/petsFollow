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
  String get loginOr => 'ou';

  @override
  String get loginWithGoogle => 'Continuer avec Google';

  @override
  String get twoFaTitle => 'Vérification 2FA';

  @override
  String get twoFaSubtitle =>
      'Saisissez le code à 6 chiffres de votre application d\'authentification.';

  @override
  String get twoFaCode => 'Code authenticator';

  @override
  String get twoFaSubmit => 'Valider';

  @override
  String get twoFaBack => 'Retour à la connexion';

  @override
  String get twoFaInvalid => 'Code 2FA invalide ou expiré';

  @override
  String get forgotPassword => 'Mot de passe oublié ?';

  @override
  String get forgotPasswordTitle => 'Mot de passe oublié';

  @override
  String get forgotPasswordSubtitle =>
      'Indiquez l\'email de votre compte. Si un compte existe, un lien de réinitialisation sera envoyé.';

  @override
  String get forgotPasswordSubmit => 'Envoyer le lien';

  @override
  String get forgotPasswordBack => 'Retour à la connexion';

  @override
  String get forgotPasswordFailed => 'Envoi impossible';

  @override
  String get forgotPasswordSentTitle => 'Email envoyé';

  @override
  String forgotPasswordSent(String email) {
    return 'Si un compte existe pour $email, un lien a été envoyé. Ouvrez-le dans votre navigateur pour choisir un nouveau mot de passe.';
  }

  @override
  String get emailRequired => 'Saisissez une adresse email valide';

  @override
  String get resetPasswordTitle => 'Nouveau mot de passe';

  @override
  String get resetPasswordSubtitle => 'Minimum 8 caractères.';

  @override
  String get resetPasswordToken => 'Jeton de réinitialisation';

  @override
  String get resetPasswordSubmit => 'Enregistrer';

  @override
  String get resetPasswordBackToLogin => 'Aller à la connexion';

  @override
  String get resetPasswordInvalidLink => 'Lien de réinitialisation invalide';

  @override
  String get resetPasswordFailed => 'Réinitialisation impossible';

  @override
  String get resetPasswordDoneTitle => 'Mot de passe mis à jour';

  @override
  String get resetPasswordDoneSubtitle =>
      'Vous pouvez maintenant vous connecter.';

  @override
  String get fullName => 'Nom complet';

  @override
  String get registerCta => 'Créer un compte';

  @override
  String get registerTitle => 'Créer un compte';

  @override
  String get registerSubtitle =>
      'Suivez la santé de votre animal. Confirmez ensuite votre email.';

  @override
  String get registerSubmit => 'S\'inscrire';

  @override
  String get registerSuccess =>
      'Compte créé. Vérifiez votre email pour confirmer, puis connectez-vous.';

  @override
  String get registerFailed => 'Inscription impossible';

  @override
  String get registerEmailExists => 'Cet email est déjà utilisé';

  @override
  String get registerBackToLogin => 'Retour à la connexion';

  @override
  String get vetUseProWeb =>
      'Le compte vétérinaire complet s\'utilise sur le site Pro web.';

  @override
  String get unsupportedRoleApp =>
      'Ce compte n\'est pas utilisable dans l\'app pets. Utilisez le site Pro web.';

  @override
  String get proLightTitle => 'Pro terrain';

  @override
  String get proLightAgenda => 'Agenda';

  @override
  String get proLightClients => 'Clients';

  @override
  String get proLightPets => 'Animaux';

  @override
  String get proLightLoadError => 'Chargement impossible';

  @override
  String get proLightNoVisits => 'Aucun rendez-vous';

  @override
  String get proLightNoClients => 'Aucun client partagé';

  @override
  String get proLightNoPets => 'Aucun animal partagé';

  @override
  String get proLightAddress => 'Adresse';

  @override
  String get proLightOpenMaps => 'Maps';

  @override
  String get proLightReportTitle => 'Compte rendu';

  @override
  String get proLightReportHint => 'Notes de visite…';

  @override
  String get proLightImproveAi => 'Améliorer (IA)';

  @override
  String get proLightFinalizeReport => 'Finaliser';

  @override
  String get proLightReportFinal => 'Finalisé';

  @override
  String get proLightSettings => 'Réglages';

  @override
  String get proLightSpecialty => 'Spécialité';

  @override
  String get proLightDocuments => 'Documents';

  @override
  String get proLightNoDocuments => 'Aucun document';

  @override
  String get proLightTimeline => 'Timeline';

  @override
  String get proLightNoTimeline => 'Aucun événement';

  @override
  String get proLightReminders => 'Rappels';

  @override
  String get proLightNoReminders => 'Aucun rappel';

  @override
  String get proLightLitterTag => 'Tag / portée';

  @override
  String get proLightActionFailed => 'Action impossible';

  @override
  String get proLightReadOnly => 'Accès lecture seule';

  @override
  String get proLightUseGps => 'GPS';

  @override
  String get proLightTranscribeAudio => 'Audio → texte';

  @override
  String get proLightGpsDenied => 'Position indisponible';

  @override
  String get googleNotConfigured => 'Connexion Google non configurée';

  @override
  String get googleLoginFailed => 'Connexion Google impossible';

  @override
  String get googleClientNotFound =>
      'Aucun compte client pour cet email. Demandez une invitation à votre vétérinaire';

  @override
  String get googleWrongAudience =>
      'Ce compte Google n\'est pas un profil client';

  @override
  String get myPets => 'Mes animaux';

  @override
  String get myData => 'Mes données';

  @override
  String get settings => 'Paramètres';

  @override
  String get logout => 'Fermer la session';

  @override
  String get save => 'Sauvegarder';

  @override
  String get cancel => 'Annuler';

  @override
  String get firstName => 'Votre prénom';

  @override
  String get currentPassword => 'Mot de passe actuel';

  @override
  String get newPassword => 'Nouveau mot de passe';

  @override
  String get confirmNewPassword => 'Confirmer le mot de passe';

  @override
  String get changePassword => 'Changer le mot de passe';

  @override
  String get forceChangePasswordTitle => 'Changer le mot de passe';

  @override
  String get forceChangePasswordSubtitle =>
      'Ce compte a été créé avec un mot de passe temporaire. Choisissez le vôtre pour continuer.';

  @override
  String get forceChangePasswordSubmit => 'Enregistrer et continuer';

  @override
  String get passwordTooShort => 'Minimum 8 caractères';

  @override
  String get passwordMismatch => 'Les mots de passe ne correspondent pas';

  @override
  String get passwordChangeFailed => 'Impossible de modifier le mot de passe';

  @override
  String get deleteAccount => 'Supprimer le compte';

  @override
  String get deleteAccountConfirm =>
      'Cette action est irréversible. Tous vos animaux et données seront supprimés.';

  @override
  String get profileSaved => 'Profil enregistré';

  @override
  String get changePhoto => 'Changer la photo';

  @override
  String get addPhoto => 'Ajouter une photo';

  @override
  String get photoUpdated => 'Photo mise à jour';

  @override
  String get passwordChanged => 'Mot de passe modifié';

  @override
  String greeting(String name) {
    return 'Bonjour $name,';
  }

  @override
  String get latestValues => 'Dernières valeurs';

  @override
  String get startMeasurement => 'DÉMARRER LA MESURE';

  @override
  String get choosePetForMeasurement => 'Choisir un animal';

  @override
  String get chooseDuration => 'Durée de la mesure';

  @override
  String durationSeconds(int seconds) {
    return '$seconds s';
  }

  @override
  String get howToMeasure => 'Comment mesurer ?';

  @override
  String get howToMeasureIntro =>
      'Mesurer la fréquence cardiaque de votre animal au repos.';

  @override
  String get howToMeasureStep1 =>
      '1. Placez votre animal au calme, allongé ou assis.';

  @override
  String get howToMeasureStep2 =>
      '2. Placez votre main sur le thorax et tapez à chaque battement pendant la durée indiquée.';

  @override
  String get howToMeasureStep3 =>
      '3. Validez le relevé pour l\'envoyer à votre vétérinaire.';

  @override
  String get howToMeasureWhyTitle => 'Pourquoi mesurer ?';

  @override
  String get howToMeasureWhyBody =>
      'Le suivi régulier de la fréquence cardiaque permet de détecter des variations et d\'adapter le traitement avec votre vétérinaire.';

  @override
  String get reminders => 'Rappels';

  @override
  String get remindersHint =>
      'Recevez un rappel quotidien pour effectuer un relevé cardiaque.';

  @override
  String get remindersEnabled => 'Activer les rappels';

  @override
  String get remindersTime => 'Heure du rappel';

  @override
  String get remindersSaved => 'Rappels enregistrés';

  @override
  String get legalTermsTitle => 'Conditions générales d\'utilisation';

  @override
  String get legalPrivacyTitle => 'Politique de confidentialité';

  @override
  String get legalNoticeTitle => 'Mentions légales';

  @override
  String get legalOpenOnline => 'Voir la version en ligne';

  @override
  String get legalTermsBody =>
      'Conditions générales d\'utilisation — petsFollow\n\nL\'application petsFollow permet aux propriétaires d\'animaux de mesurer la fréquence cardiaque, de consulter l\'historique et de communiquer avec leur vétérinaire.\n\nLes services sont fournis dans le cadre de l\'abonnement choisi (paiement via Stripe). L\'utilisateur s\'engage à utiliser l\'application conformément à sa destination.\n\nVersion complète : https://petsfollow.ll-it-sc.be/legal/terms\n\nDate d\'actualisation : juillet 2026';

  @override
  String get legalPrivacyBody =>
      'Politique de confidentialité — petsFollow\n\nDonnées collectées : identité (prénom, email), données animal (nom, espèce, race, photos), relevés de fréquence cardiaque (données de santé animale), messages et médias échangés avec le cabinet, jetons de notification (FCM), données de paiement traitées par Stripe.\n\nFinalités : gestion du compte, suivi cardiaque, messagerie vétérinaire, notifications, facturation.\n\nSous-traitants / partenaires : Google (Sign-In, Firebase Cloud Messaging), Stripe (paiements), hébergement cloud (GCP).\n\nConservation : jusqu\'à suppression du compte ou 3 ans d\'inactivité.\n\nDroits RGPD (accès, rectification, suppression) : Profil → Supprimer le compte, ou contact support@ll-it-sc.be.\n\nVersion complète : https://petsfollow.ll-it-sc.be/legal/privacy\n\nDate d\'actualisation : juillet 2026';

  @override
  String get legalNoticeBody =>
      'Mentions légales — petsFollow\n\nÉditeur : LL-IT-SC / petsFollow\nContact : support@ll-it-sc.be\n\nHébergement : Google Cloud Platform (conformité RGPD).\n\nDirecteur de publication : petsFollow.\n\nVersion complète : https://petsfollow.ll-it-sc.be/legal/mentions\n\nDate d\'actualisation : juillet 2026';

  @override
  String get language => 'Langue';

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
  String get planAnnualLabel => '35 € / an';

  @override
  String get planTriennialLabel => '95 € / 3 ans';

  @override
  String get planQuinquennialLabel => '145 € / 5 ans';

  @override
  String get pushNewMessage => 'Nouveau message';

  @override
  String get pushVisitConfirmed => 'Rendez-vous confirmé';

  @override
  String get pushVisitProposed => 'Proposition de rendez-vous';

  @override
  String get pushVisitReschedule => 'Déplacement de rendez-vous';

  @override
  String get notifChannelMessages => 'Messages';

  @override
  String get notifChannelVisits => 'Visites';

  @override
  String get notifChannelCare => 'Soins';

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
  String get errorNetwork =>
      'Connexion impossible. Vérifiez votre réseau et réessayez.';

  @override
  String get retryAction => 'Réessayer';

  @override
  String get errorMediaTooLarge => 'Fichier trop volumineux (25 Mo max)';

  @override
  String get errorInvalidMediaType =>
      'Format non supporté (JPEG, PNG, WebP, MP4, MOV, WebM)';

  @override
  String get errorPaymentRequired =>
      'Abonnement requis pour envoyer des médias';

  @override
  String get errorPhotoUploadFailed =>
      'Animal créé, mais la photo n\'a pas pu être envoyée';

  @override
  String get errorCouldNotOpenLink => 'Impossible d\'ouvrir le lien';

  @override
  String planAnnualSub(String price) {
    return '$price, renouvelé automatiquement';
  }

  @override
  String get planTriennialSub =>
      '95 € tous les 3 ans, renouvelé automatiquement';

  @override
  String get planQuinquennialSub => '145 € pour 5 ans, paiement unique';

  @override
  String planOneTime(String price) {
    return '$price, paiement unique';
  }

  @override
  String get heartRateInstructions =>
      'Tapotez à chaque battement pendant la durée indiquée par votre vétérinaire.';

  @override
  String heartRateInstructionsDuration(int seconds) {
    return 'Tapotez à chaque battement pendant $seconds secondes.';
  }

  @override
  String get heartRateNoDurationConfigured =>
      'Aucune durée de mesure n’est configurée pour ce cabinet. Contactez votre vétérinaire.';

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

  @override
  String get navHome => 'Accueil';

  @override
  String get navPets => 'Animaux';

  @override
  String get navCare => 'Soins';

  @override
  String get navMessages => 'Messages';

  @override
  String get navProfile => 'Profil';

  @override
  String get speciesDog => 'Chien';

  @override
  String get speciesCat => 'Chat';

  @override
  String get speciesHorse => 'Cheval';

  @override
  String get speciesOther => 'Autre';

  @override
  String get careComingSoon => 'Les rappels de soins arrivent bientôt';

  @override
  String get emptyPetsTitle => 'Aucun animal';

  @override
  String get emptyPetsBody =>
      'Ajoutez votre premier animal pour commencer le suivi cardiaque avec votre vétérinaire.';

  @override
  String get discoveryTitle => 'Découvrir petsFollow';

  @override
  String get discoveryMission => 'Votre parcours en 7 jours';

  @override
  String get discoveryDay0Title => 'Jour 0 — Bienvenue';

  @override
  String get discoveryDay0Body =>
      'Créez le profil de votre animal et découvrez comment mesurer sa fréquence cardiaque.';

  @override
  String get discoveryDay2Title => 'Jour 2 — Première mesure';

  @override
  String get discoveryDay2Body =>
      'Effectuez votre premier relevé cardiaque et familiarisez-vous avec la technique.';

  @override
  String get discoveryDay4Title => 'Jour 4 — Routine';

  @override
  String get discoveryDay4Body =>
      'Installez une routine de mesure quotidienne avec les rappels personnalisés.';

  @override
  String get discoveryDay6Title => 'Jour 6 — Partage véto';

  @override
  String get discoveryDay6Body =>
      'Vos relevés sont partagés avec votre vétérinaire pour un suivi optimal.';

  @override
  String get myVets => 'Mes vétérinaires';

  @override
  String get addVetByEmail => 'Ajouter un véto par email';

  @override
  String get vetEmailHint => 'email@cabinet.vet';

  @override
  String get noVets => 'Aucun vétérinaire lié';

  @override
  String get primaryVet => 'Vétérinaire principal';

  @override
  String get setPrimaryVet => 'Définir comme véto principal';

  @override
  String get careTitle => 'Soins';

  @override
  String get careDone => 'Fait';

  @override
  String get carePostpone => 'Reporter';

  @override
  String get careOverdue => 'En retard';

  @override
  String get visitHistory => 'Historique des visites';

  @override
  String get requestVisit => 'Demander une visite';

  @override
  String get calendarBookingDisabled =>
      'La réservation en ligne n\'est pas disponible pour ce cabinet. Vous pouvez envoyer une demande sans créneau.';

  @override
  String get calendarBookingDisabledReschedule =>
      'La réservation en ligne n\'est pas disponible. Proposez une date manuellement.';

  @override
  String get calendarNoSlots =>
      'Aucun créneau disponible sur les 14 prochains jours.';

  @override
  String get calendarPickSlot => 'Choisissez un créneau :';

  @override
  String get visitConfirm => 'Confirmer';

  @override
  String get visitProposeReschedule => 'Proposer un autre créneau';

  @override
  String get visitRescheduleProposed => 'Proposition de déplacement envoyée';

  @override
  String get paymentSuccessSnack => 'Paiement reçu — actualisation…';

  @override
  String get paymentCancelSnack => 'Paiement annulé';

  @override
  String get visitRejectReschedule => 'Refuser le déplacement';

  @override
  String get visitAcceptReschedule => 'Accepter le nouveau créneau';

  @override
  String get upcomingVisit => 'Visite à venir';

  @override
  String get notificationHrTitle => 'petsFollow';

  @override
  String get notificationHrBody =>
      'Il est temps de prendre un relevé cardiaque pour votre animal';

  @override
  String get reviewAskTitle => 'Vous aimez petsFollow ?';

  @override
  String get reviewAskYes => 'Oui, noter l\'app';

  @override
  String get reviewAskNo => 'Plus tard';

  @override
  String get carePlusUpsell => 'Care+ — médicaments et rappels personnalisés';

  @override
  String get carePlusRequired =>
      'Care+ est requis pour les médicaments et rappels personnalisés.';

  @override
  String get horsePackRequired =>
      'Le Pack Cheval est requis pour les rappels maréchal, contacts et compétitions.';

  @override
  String get activateAddon => 'Activer';

  @override
  String get careTypeMedication => 'Médicament';

  @override
  String get horseAddContact => 'Ajouter un contact';

  @override
  String get horseAddCompetition => 'Ajouter une compétition';

  @override
  String get horseContactName => 'Nom';

  @override
  String get horseContactRole => 'Rôle';

  @override
  String get horseCompetitionTitle => 'Événement';

  @override
  String get horseCompetitionDate => 'Date (AAAA-MM-JJ)';

  @override
  String get familyPackHint =>
      'Pack Famille — vue foyer des rappels, −10 % dès le 2ᵉ abonnement animal payant';

  @override
  String familyHouseholdTitle(int count) {
    return 'Foyer Famille — $count animaux';
  }

  @override
  String kennelHouseholdTitle(int count) {
    return 'Foyer Élevage — $count animaux';
  }

  @override
  String get familyHouseholdNext => 'Prochains rappels du foyer';

  @override
  String get familyPetLimit =>
      'Un pack foyer est déjà actif ou en cours d\'achat';

  @override
  String get familyRequiresTwoPets =>
      'Le pack Famille nécessite au moins 2 animaux';

  @override
  String get kennelPackHint =>
      'Pack Élevage — ≥6 animaux, −15 % sur les abos suivants';

  @override
  String get kennelRequiresSixPets =>
      'Le pack Élevage nécessite au moins 6 animaux';

  @override
  String get kennelQuickEncodeTitle => 'Encodage portée (élevage)';

  @override
  String get kennelRequired =>
      'Le pack Élevage est requis pour l\'encodage par lot';

  @override
  String get litterTag => 'Tag portée';

  @override
  String get discoveryMarkDone => 'Mission accomplie';

  @override
  String get notificationPreferences => 'Préférences de notifications';

  @override
  String get notificationPrefsHint =>
      'Choisissez les types de notifications que vous souhaitez recevoir.';

  @override
  String get notificationPrefsSaved => 'Préférences enregistrées';

  @override
  String get notificationPrefHr => 'Relevés cardiaques';

  @override
  String get notificationPrefCare => 'Rappels de soins';

  @override
  String get notificationPrefVisits => 'Visites';

  @override
  String get notificationPrefMessages => 'Messages';

  @override
  String get notificationPrefDiscovery => 'Parcours découverte';

  @override
  String get notificationPrefBilling => 'Facturation';

  @override
  String carePostponeDays(int days) {
    return 'Reporter de $days jours';
  }

  @override
  String get noCareReminders => 'Aucun rappel de soin en cours';

  @override
  String get careAddReminder => 'Ajouter un rappel';

  @override
  String get careSelectPet => 'Animal';

  @override
  String careDueInDays(int days) {
    return 'Échéance dans $days jours';
  }

  @override
  String get careReferenceModeDone => 'Déjà effectué';

  @override
  String get careReferenceModeFirst => 'Première fois';

  @override
  String get careLastDateLabel => 'Dernière date';

  @override
  String get careLastDateDone => 'Date du dernier soin';

  @override
  String get careLastDateFirst => 'Date de départ du cycle';

  @override
  String get careRecurrenceLabel => 'Récurrence';

  @override
  String get careRecurrenceNone => 'Aucune (échéance unique)';

  @override
  String careRecurrenceDays(int days) {
    return 'Tous les $days jours';
  }

  @override
  String get careDueDateLabel => 'Échéance';

  @override
  String get careDueDateComputed => 'Échéance calculée';

  @override
  String get careTooltipDoneWithRecurrence =>
      'Soin déjà fait : l’échéance = date du dernier soin + récurrence.';

  @override
  String get careTooltipFirstWithRecurrence =>
      'Première planification : indiquez la date de départ du cycle. L’échéance = cette date + récurrence.';

  @override
  String get careTooltipNoRecurrence =>
      'Sans récurrence : la date saisie est l’échéance unique.';

  @override
  String get careTooltipDueExplained =>
      'Échéance = dernière date + récurrence (si définie).';

  @override
  String get carePickDate => 'Choisir une date';

  @override
  String discoveryDayBadge(int day) {
    return 'J$day';
  }

  @override
  String get timelineTypeHeartrate => 'Fréquence cardiaque';

  @override
  String get timelineTypeMessage => 'Message';

  @override
  String get timelineTypeCare => 'Soin';

  @override
  String get timelineTypeVisit => 'Visite';

  @override
  String get timelineTypeEvent => 'Événement';

  @override
  String get visitCancelAction => 'Annuler la demande';

  @override
  String get upcomingVisits => 'Prochaines visites';

  @override
  String get timelineEmpty => 'Aucun événement pour le moment';

  @override
  String get noThreads => 'Aucune conversation';

  @override
  String get vetInviteSent =>
      'Invitation envoyée — le cabinet doit accepter la demande';

  @override
  String vetInviteSentNamed(String practice) {
    return 'Demande envoyée à $practice — le cabinet doit l’accepter';
  }

  @override
  String get vetNotFound => 'Aucun vétérinaire trouvé avec cet email';

  @override
  String get addVetSearchHint =>
      'Nous recherchons ce compte vétérinaire dans petsFollow. S’il existe, une demande de liaison est envoyée au cabinet.';

  @override
  String get visitRequested => 'Demande de visite envoyée';

  @override
  String get primaryVetSet => 'Vétérinaire principal mis à jour';

  @override
  String get visitStatusRequested => 'Demandée';

  @override
  String get visitStatusConfirmed => 'Confirmée';

  @override
  String get visitStatusDone => 'Terminée';

  @override
  String get visitStatusCancelled => 'Annulée';

  @override
  String get visitStatusReschedulePending => 'Déplacement en attente';

  @override
  String get horseHealthTitle => 'Santé équine';

  @override
  String get horseContactsTitle => 'Contacts (maréchal, dentiste…)';

  @override
  String get horseCompetitionsTitle => 'Compétitions';

  @override
  String get horseContactsSoon =>
      'Activez le Pack Cheval pour gérer vos contacts professionnels.';

  @override
  String get horseCompetitionsSoon =>
      'Activez le Pack Cheval pour le calendrier de compétitions.';

  @override
  String get horsePackUpsell =>
      'Pack Cheval — maréchal, coproscopie, contacts et compétitions';

  @override
  String get careTypeFarrier => 'Maréchal-ferrant';

  @override
  String get careTypeFecalEgg => 'Coproscopie';

  @override
  String get careTypeVaccination => 'Vaccination';

  @override
  String get careTypeDeworming => 'Vermifuge';

  @override
  String get careTypeVetCheck => 'Contrôle vétérinaire';

  @override
  String get careTypeDental => 'Soins dentaires';

  @override
  String get careTypeCustom => 'Rappel personnalisé';

  @override
  String get homeAddFirstVetTitle => 'Ajoutez votre vétérinaire';

  @override
  String get homeAddFirstVetBody =>
      'Liez le cabinet qui suit votre animal pour partager les relevés et échanger.';

  @override
  String get homeAddFirstVetCta => 'Ajouter un vétérinaire';

  @override
  String get photoFrameHint =>
      'Cadrez le museau au centre — aperçu fiche animal';

  @override
  String get takePhoto => 'Prendre une photo';

  @override
  String get chooseFromGallery => 'Choisir dans la galerie';

  @override
  String get attachMedia => 'Joindre une photo ou une vidéo';

  @override
  String get attachPhoto => 'Photo';

  @override
  String get attachVideo => 'Vidéo';

  @override
  String get openMedia => 'Ouvrir';

  @override
  String get mediaVideoLabel => 'Vidéo';
}
