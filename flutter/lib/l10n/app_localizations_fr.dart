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
  String get changePassword => 'Changer le mot de passe';

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
  String get legalTermsBody =>
      'Conditions générales d\'utilisation — petsFollow\n\nL\'application petsFollow permet aux propriétaires d\'animaux de mesurer la fréquence cardiaque, de consulter l\'historique et de communiquer avec leur vétérinaire.\n\nLes services sont fournis dans le cadre de l\'abonnement choisi. L\'utilisateur s\'engage à utiliser l\'application conformément à sa destination.\n\nDate d\'actualisation : juillet 2026';

  @override
  String get legalPrivacyBody =>
      'Politique de confidentialité — petsFollow\n\nDonnées collectées : prénom, email, données animal (nom, espèce, race), relevés cardiaques, messages au vétérinaire.\n\nFinalités : gestion du compte, suivi santé, communication avec le cabinet vétérinaire.\n\nConservation : jusqu\'à suppression du compte ou 3 ans d\'inactivité.\n\nVous pouvez exercer vos droits (accès, rectification, suppression) via les paramètres de l\'application.\n\nDate d\'actualisation : juillet 2026';

  @override
  String get legalNoticeBody =>
      'Mentions légales — petsFollow\n\nÉditeur : petsFollow\nContact : support@petsfollow.test\n\nHébergement : infrastructure cloud conforme RGPD.\n\nDirecteur de publication : petsFollow.\n\nDate d\'actualisation : juillet 2026';

  @override
  String get language => 'Langue';

  @override
  String get languageFr => 'Français';

  @override
  String get languageNl => 'Nederlands';

  @override
  String get languageEn => 'English';

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
      '79 € tous les 3 ans, renouvelé automatiquement';

  @override
  String get planQuinquennialSub =>
      '115 € tous les 5 ans, renouvelé automatiquement';

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
      'Pack Famille — vue foyer des rappels, jusqu\'à 3 animaux';

  @override
  String familyHouseholdTitle(int count, int max) {
    return 'Foyer Famille — $count/$max animaux';
  }

  @override
  String get familyHouseholdNext => 'Prochains rappels du foyer';

  @override
  String get familyPetLimit => 'Pack Famille limité à 3 animaux';

  @override
  String get familyRequiresTwoPets =>
      'Le pack Famille nécessite au moins 2 animaux';

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
  String get vetInviteSent => 'Invitation envoyée au vétérinaire';

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
