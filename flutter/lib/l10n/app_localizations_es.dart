// ignore: unused_import
import 'package:intl/intl.dart' as intl;
import 'app_localizations.dart';

// ignore_for_file: type=lint

/// The translations for Spanish Castilian (`es`).
class AppLocalizationsEs extends AppLocalizations {
  AppLocalizationsEs([String locale = 'es']) : super(locale);

  @override
  String get appTitle => 'petsFollow';

  @override
  String get appTagline => 'Seguimiento de la salud de su mascota';

  @override
  String get email => 'Correo';

  @override
  String get password => 'Contraseña';

  @override
  String get login => 'Iniciar sesión';

  @override
  String get loginFailed => 'Error al iniciar sesión';

  @override
  String get loginOr => 'o';

  @override
  String get loginWithGoogle => 'Continuar con Google';

  @override
  String get twoFaTitle => 'Verificación 2FA';

  @override
  String get twoFaSubtitle =>
      'Introduzca el código de 6 dígitos de su aplicación de autenticación.';

  @override
  String get twoFaCode => 'Código autenticador';

  @override
  String get twoFaSubmit => 'Validar';

  @override
  String get twoFaBack => 'Volver al inicio de sesión';

  @override
  String get twoFaInvalid => 'Código 2FA inválido o caducado';

  @override
  String get forgotPassword => '¿Olvidó su contraseña?';

  @override
  String get forgotPasswordTitle => 'Contraseña olvidada';

  @override
  String get forgotPasswordSubtitle =>
      'Indique el correo de su cuenta. Si existe una cuenta, se enviará un enlace de restablecimiento.';

  @override
  String get forgotPasswordSubmit => 'Enviar enlace';

  @override
  String get forgotPasswordBack => 'Volver al inicio de sesión';

  @override
  String get forgotPasswordFailed => 'No se pudo enviar';

  @override
  String get forgotPasswordSentTitle => 'Correo enviado';

  @override
  String forgotPasswordSent(String email) {
    return 'Si existe una cuenta para $email, se ha enviado un enlace. Ábralo en el navegador para elegir una nueva contraseña.';
  }

  @override
  String get emailRequired => 'Introduzca un correo electrónico válido';

  @override
  String get resetPasswordTitle => 'Nueva contraseña';

  @override
  String get resetPasswordSubtitle => 'Mínimo 8 caracteres.';

  @override
  String get resetPasswordToken => 'Token de restablecimiento';

  @override
  String get resetPasswordSubmit => 'Guardar';

  @override
  String get resetPasswordBackToLogin => 'Ir al inicio de sesión';

  @override
  String get resetPasswordInvalidLink => 'Enlace de restablecimiento inválido';

  @override
  String get resetPasswordFailed => 'No se pudo restablecer la contraseña';

  @override
  String get resetPasswordDoneTitle => 'Contraseña actualizada';

  @override
  String get resetPasswordDoneSubtitle => 'Ya puede iniciar sesión.';

  @override
  String get fullName => 'Nombre completo';

  @override
  String get registerCta => 'Crear una cuenta';

  @override
  String get registerTitle => 'Crear una cuenta';

  @override
  String get registerSubtitle =>
      'Sigue la salud de tu animal. Confirma tu email a continuación.';

  @override
  String get registerSubmit => 'Registrarse';

  @override
  String get registerSuccess =>
      'Cuenta creada. Confirma tu email y luego inicia sesión.';

  @override
  String get registerFailed => 'No se pudo registrar';

  @override
  String get registerEmailExists => 'Este email ya está en uso';

  @override
  String get registerBackToLogin => 'Volver al inicio de sesión';

  @override
  String get vetUseProWeb =>
      'Las cuentas veterinarias completas usan el sitio Pro web.';

  @override
  String get unsupportedRoleApp =>
      'Esta cuenta no se puede usar en la app pets. Use el sitio Pro web.';

  @override
  String get proLightTitle => 'Pro terreno';

  @override
  String get proLightAgenda => 'Agenda';

  @override
  String get proLightClients => 'Clientes';

  @override
  String get proLightPets => 'Animales';

  @override
  String get proLightLoadError => 'No se pudo cargar';

  @override
  String get proLightNoVisits => 'Sin citas';

  @override
  String get proLightNoClients => 'Sin clientes compartidos';

  @override
  String get proLightNoPets => 'Sin animales compartidos';

  @override
  String get proLightAddress => 'Dirección';

  @override
  String get proLightOpenMaps => 'Maps';

  @override
  String get proLightReportTitle => 'Informe de visita';

  @override
  String get proLightReportHint => 'Notas de visita…';

  @override
  String get proLightImproveAi => 'Mejorar (IA)';

  @override
  String get proLightFinalizeReport => 'Finalizar';

  @override
  String get proLightReportFinal => 'Finalizado';

  @override
  String get proLightSettings => 'Ajustes';

  @override
  String get proLightSpecialty => 'Especialidad';

  @override
  String get proLightDocuments => 'Documentos';

  @override
  String get proLightNoDocuments => 'Sin documentos';

  @override
  String get proLightTimeline => 'Cronología';

  @override
  String get proLightNoTimeline => 'Sin eventos';

  @override
  String get proLightReminders => 'Recordatorios';

  @override
  String get proLightNoReminders => 'Sin recordatorios';

  @override
  String get proLightLitterTag => 'Camada / etiqueta';

  @override
  String get proLightActionFailed => 'Acción imposible';

  @override
  String get proLightReadOnly => 'Acceso de solo lectura';

  @override
  String get proLightUseGps => 'GPS';

  @override
  String get proLightTranscribeAudio => 'Audio → texto';

  @override
  String get proLightGpsDenied => 'Ubicación no disponible';

  @override
  String get googleNotConfigured =>
      'Inicio de sesión con Google no configurado';

  @override
  String get googleLoginFailed => 'No se pudo iniciar sesión con Google';

  @override
  String get googleClientNotFound =>
      'No hay cuenta de cliente para este correo. Pida una invitación a su veterinario';

  @override
  String get googleWrongAudience =>
      'Esta cuenta de Google no es un perfil de cliente';

  @override
  String get myPets => 'Mis mascotas';

  @override
  String get myData => 'Mis datos';

  @override
  String get settings => 'Ajustes';

  @override
  String get logout => 'Cerrar sesión';

  @override
  String get save => 'Guardar';

  @override
  String get cancel => 'Cancelar';

  @override
  String get firstName => 'Nombre';

  @override
  String get currentPassword => 'Contraseña actual';

  @override
  String get newPassword => 'Nueva contraseña';

  @override
  String get confirmNewPassword => 'Confirmar contraseña';

  @override
  String get changePassword => 'Cambiar contraseña';

  @override
  String get forceChangePasswordTitle => 'Cambiar la contraseña';

  @override
  String get forceChangePasswordSubtitle =>
      'Esta cuenta se creó con una contraseña temporal. Elija la suya para continuar.';

  @override
  String get forceChangePasswordSubmit => 'Guardar y continuar';

  @override
  String get passwordTooShort => 'Mínimo 8 caracteres';

  @override
  String get passwordMismatch => 'Las contraseñas no coinciden';

  @override
  String get passwordChangeFailed => 'No se pudo cambiar la contraseña';

  @override
  String get deleteAccount => 'Eliminar cuenta';

  @override
  String get deleteAccountConfirm =>
      'Esta acción no se puede deshacer. Se eliminarán todas sus mascotas y datos.';

  @override
  String get profileSaved => 'Perfil guardado';

  @override
  String get changePhoto => 'Cambiar foto';

  @override
  String get addPhoto => 'Añadir una foto';

  @override
  String get photoUpdated => 'Foto actualizada';

  @override
  String get passwordChanged => 'Contraseña cambiada';

  @override
  String greeting(String name) {
    return 'Hola $name,';
  }

  @override
  String get latestValues => 'Últimos valores';

  @override
  String get startMeasurement => 'EMPEZAR MEDICIÓN';

  @override
  String get choosePetForMeasurement => 'Elegir un animal';

  @override
  String get chooseDuration => 'Duración de la medición';

  @override
  String durationSeconds(int seconds) {
    return '$seconds s';
  }

  @override
  String get howToMeasure => '¿Cómo medir?';

  @override
  String get howToMeasureIntro =>
      'Mida la frecuencia cardíaca en reposo de su mascota.';

  @override
  String get howToMeasureStep1 =>
      '1. Mantenga a su mascota tranquila, tumbada o sentada.';

  @override
  String get howToMeasureStep2 =>
      '2. Coloque la mano en el pecho y pulse en cada latido durante la duración indicada.';

  @override
  String get howToMeasureStep3 =>
      '3. Valide la lectura para enviarla a su veterinario.';

  @override
  String get howToMeasureWhyTitle => '¿Por qué medir?';

  @override
  String get howToMeasureWhyBody =>
      'El seguimiento regular de la frecuencia cardíaca ayuda a detectar cambios y ajustar el tratamiento con su veterinario.';

  @override
  String get reminders => 'Recordatorios';

  @override
  String get remindersHint =>
      'Reciba un recordatorio diario para tomar una lectura de frecuencia cardíaca.';

  @override
  String get remindersEnabled => 'Activar recordatorios';

  @override
  String get remindersTime => 'Hora del recordatorio';

  @override
  String get remindersSaved => 'Recordatorios guardados';

  @override
  String get legalTermsTitle => 'Condiciones de uso';

  @override
  String get legalPrivacyTitle => 'Política de privacidad';

  @override
  String get legalNoticeTitle => 'Aviso legal';

  @override
  String get legalOpenOnline => 'Ver la versión en línea';

  @override
  String get legalTermsBody =>
      'Condiciones de uso — petsFollow\n\nLa app petsFollow permite a los propietarios medir la frecuencia cardíaca, consultar el historial y comunicarse con su veterinario.\n\nLos servicios se prestan según la suscripción seleccionada (pagos vía Stripe).\n\nVersión completa: https://petsfollow.ll-it-sc.be/legal/terms\n\nÚltima actualización: julio de 2026';

  @override
  String get legalPrivacyBody =>
      'Política de privacidad — petsFollow\n\nDatos recogidos: identidad (nombre, correo), datos de la mascota (nombre, especie, raza, fotos), lecturas de frecuencia cardíaca (datos de salud animal), mensajes y medios con la clínica, tokens de notificación (FCM), datos de pago tratados por Stripe.\n\nFinalidades: gestión de la cuenta, seguimiento cardíaco, mensajería veterinaria, notificaciones, facturación.\n\nEncargados / socios: Google (Sign-In, Firebase Cloud Messaging), Stripe (pagos), hosting cloud (GCP).\n\nConservación: hasta la eliminación de la cuenta o 3 años de inactividad.\n\nDerechos RGPD: Perfil → Eliminar cuenta, o support@ll-it-sc.be.\n\nVersión completa: https://petsfollow.ll-it-sc.be/legal/privacy\n\nÚltima actualización: julio de 2026';

  @override
  String get legalNoticeBody =>
      'Aviso legal — petsFollow\n\nEditor: LL-IT-SC / petsFollow\nContacto: support@ll-it-sc.be\n\nAlojamiento: Google Cloud Platform (conforme al RGPD).\n\nVersión completa: https://petsfollow.ll-it-sc.be/legal/mentions\n\nÚltima actualización: julio de 2026';

  @override
  String get language => 'Idioma';

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
  String get planAnnualLabel => '35 € / año';

  @override
  String get planTriennialLabel => '95 € / 3 años';

  @override
  String get planQuinquennialLabel => '145 € / 5 años';

  @override
  String get pushNewMessage => 'Nuevo mensaje';

  @override
  String get pushVisitConfirmed => 'Cita confirmada';

  @override
  String get pushVisitProposed => 'Propuesta de cita';

  @override
  String get pushVisitReschedule => 'Cambio de cita';

  @override
  String get notifChannelMessages => 'Mensajes';

  @override
  String get notifChannelVisits => 'Visitas';

  @override
  String get notifChannelCare => 'Cuidados';

  @override
  String get paymentResume => 'Reanudar el pago';

  @override
  String get manageSubscription => 'Gestionar suscripción';

  @override
  String get heartRate => 'Lectura de frecuencia cardíaca';

  @override
  String get history => 'Historial';

  @override
  String get vetMessaging => 'Mensajería con el veterinario';

  @override
  String get badgeAutoRenew => 'Renovación automática';

  @override
  String get badgeActive => 'Activa';

  @override
  String get badgePendingPayment => 'Pago pendiente';

  @override
  String badgeExpiresOn(String date) {
    return 'caduca el $date';
  }

  @override
  String get newPet => 'Nueva mascota';

  @override
  String get petName => 'Nombre';

  @override
  String get species => 'Especie';

  @override
  String get breed => 'Raza';

  @override
  String get choosePlan => 'Elija su plan';

  @override
  String get recommended => 'Recomendado';

  @override
  String get autoRenewTitle => 'Renovación automática';

  @override
  String get autoRenewSubtitle => 'Se cobra en cada renovación';

  @override
  String get continueToPayment => 'Continuar al pago';

  @override
  String get paymentConfirmed => 'Pago confirmado — mascota activa';

  @override
  String get paymentPending => 'Pago pendiente — puede reanudarlo más tarde';

  @override
  String errorGeneric(String message) {
    return 'Error: $message';
  }

  @override
  String get errorNetwork =>
      'No se puede conectar. Compruebe su red e inténtelo de nuevo.';

  @override
  String get retryAction => 'Reintentar';

  @override
  String get errorMediaTooLarge => 'Archivo demasiado grande (máx. 25 MB)';

  @override
  String get errorInvalidMediaType =>
      'Formato no admitido (JPEG, PNG, WebP, MP4, MOV, WebM)';

  @override
  String get errorPaymentRequired =>
      'Se requiere una suscripción para enviar medios';

  @override
  String get errorPhotoUploadFailed =>
      'Mascota creada, pero no se ha podido subir la foto';

  @override
  String get errorCouldNotOpenLink => 'No se ha podido abrir el enlace';

  @override
  String planAnnualSub(String price) {
    return '$price, renovación automática';
  }

  @override
  String get planTriennialSub => '95 € cada 3 años, renovación automática';

  @override
  String get planQuinquennialSub => '145 € por 5 años, pago único';

  @override
  String planOneTime(String price) {
    return '$price, pago único';
  }

  @override
  String get heartRateInstructions =>
      'Pulse en cada latido durante la duración fijada por su veterinario.';

  @override
  String heartRateInstructionsDuration(int seconds) {
    return 'Pulse en cada latido durante $seconds segundos.';
  }

  @override
  String get heartRateNoDurationConfigured =>
      'No hay ninguna duración de medición configurada para esta clínica. Contacte con su veterinario.';

  @override
  String get start => 'Empezar';

  @override
  String secondsLeft(int seconds) {
    return '$seconds s';
  }

  @override
  String beatsCount(int count) {
    return '$count latidos';
  }

  @override
  String get tapHere => 'Pulse aquí en cada latido';

  @override
  String bpmLabel(String bpm) {
    return 'BPM: $bpm';
  }

  @override
  String beatsLabel(int count) {
    return 'Latidos: $count';
  }

  @override
  String get thresholdAlert => 'Alerta de umbral';

  @override
  String get validateAndSend => 'Validar y enviar al veterinario';

  @override
  String get restart => 'Empezar de nuevo';

  @override
  String get sentToVet => 'Lectura enviada al veterinario';

  @override
  String get navHome => 'Inicio';

  @override
  String get navPets => 'Mascotas';

  @override
  String get navCare => 'Cuidados';

  @override
  String get navMessages => 'Mensajes';

  @override
  String get navProfile => 'Perfil';

  @override
  String get speciesDog => 'Perro';

  @override
  String get speciesCat => 'Gato';

  @override
  String get speciesHorse => 'Caballo';

  @override
  String get speciesOther => 'Otro';

  @override
  String get careComingSoon => 'Recordatorios de cuidados próximamente';

  @override
  String get emptyPetsTitle => 'Aún no hay mascotas';

  @override
  String get emptyPetsBody =>
      'Añada su primera mascota para empezar el seguimiento de frecuencia cardíaca con su veterinario.';

  @override
  String get discoveryTitle => 'Descubra petsFollow';

  @override
  String get discoveryMission => 'Su recorrido de 7 días';

  @override
  String get discoveryDay0Title => 'Día 0 — Bienvenida';

  @override
  String get discoveryDay0Body =>
      'Cree el perfil de su mascota y aprenda a medir la frecuencia cardíaca.';

  @override
  String get discoveryDay2Title => 'Día 2 — Primera lectura';

  @override
  String get discoveryDay2Body =>
      'Tome su primera lectura de frecuencia cardíaca y familiarícese con la técnica.';

  @override
  String get discoveryDay4Title => 'Día 4 — Rutina';

  @override
  String get discoveryDay4Body =>
      'Cree el hábito de medir a diario con recordatorios personalizados.';

  @override
  String get discoveryDay6Title => 'Día 6 — Compartir con el veterinario';

  @override
  String get discoveryDay6Body =>
      'Sus lecturas se comparten con su veterinario para un seguimiento óptimo.';

  @override
  String get myVets => 'Mis veterinarios';

  @override
  String get addVetByEmail => 'Añadir un veterinario por correo';

  @override
  String get vetEmailHint => 'correo@clinica.vet';

  @override
  String get noVets => 'Ningún veterinario vinculado';

  @override
  String get primaryVet => 'Veterinario principal';

  @override
  String get setPrimaryVet => 'Establecer como veterinario principal';

  @override
  String get careTitle => 'Cuidados';

  @override
  String get careDone => 'Hecho';

  @override
  String get carePostpone => 'Aplazar';

  @override
  String get careOverdue => 'Atrasado';

  @override
  String get visitHistory => 'Historial de visitas';

  @override
  String get requestVisit => 'Solicitar una visita';

  @override
  String get calendarBookingDisabled =>
      'La reserva en línea no está disponible para esta clínica. Puede enviar una solicitud sin horario.';

  @override
  String get calendarBookingDisabledReschedule =>
      'La reserva en línea no está disponible. Proponga una fecha manualmente.';

  @override
  String get calendarNoSlots =>
      'No hay huecos disponibles en los próximos 14 días.';

  @override
  String get calendarPickSlot => 'Elija un hueco:';

  @override
  String get visitConfirm => 'Confirmar';

  @override
  String get visitProposeReschedule => 'Proponer otro horario';

  @override
  String get visitRescheduleProposed => 'Propuesta de cambio enviada';

  @override
  String get paymentSuccessSnack => 'Pago recibido — actualizando…';

  @override
  String get paymentCancelSnack => 'Pago cancelado';

  @override
  String get visitRejectReschedule => 'Rechazar cambio';

  @override
  String get visitAcceptReschedule => 'Aceptar nuevo horario';

  @override
  String get upcomingVisit => 'Próxima visita';

  @override
  String get notificationHrTitle => 'petsFollow';

  @override
  String get notificationHrBody =>
      'Es el momento de una lectura de frecuencia cardíaca para su mascota';

  @override
  String get reviewAskTitle => '¿Le gusta petsFollow?';

  @override
  String get reviewAskYes => 'Sí, valorar la app';

  @override
  String get reviewAskNo => 'Más tarde';

  @override
  String get carePlusUpsell =>
      'Care+ — medicación y recordatorios personalizados';

  @override
  String get carePlusRequired =>
      'Se requiere Care+ para medicación y recordatorios personalizados.';

  @override
  String get horsePackRequired =>
      'Se requiere el pack Caballo para recordatorios de herrador, contactos y competiciones.';

  @override
  String get activateAddon => 'Activar';

  @override
  String get careTypeMedication => 'Medicación';

  @override
  String get horseAddContact => 'Añadir un contacto';

  @override
  String get horseAddCompetition => 'Añadir una competición';

  @override
  String get horseContactName => 'Nombre';

  @override
  String get horseContactRole => 'Rol';

  @override
  String get horseCompetitionTitle => 'Evento';

  @override
  String get horseCompetitionDate => 'Fecha (AAAA-MM-DD)';

  @override
  String get familyPackHint =>
      'Pack Familia — vista del hogar, −10% desde el 2.º plan de mascota de pago';

  @override
  String familyHouseholdTitle(int count) {
    return 'Hogar Familia — $count mascotas';
  }

  @override
  String kennelHouseholdTitle(int count) {
    return 'Hogar Criadero — $count mascotas';
  }

  @override
  String get familyHouseholdNext => 'Próximos recordatorios del hogar';

  @override
  String get familyPetLimit => 'Ya hay un pack hogar activo o en compra';

  @override
  String get familyRequiresTwoPets =>
      'El pack Familia requiere al menos 2 mascotas';

  @override
  String get kennelPackHint =>
      'Pack Criadero — ≥6 mascotas, −15% en siguientes planes';

  @override
  String get kennelRequiresSixPets =>
      'El pack Criadero requiere al menos 6 mascotas';

  @override
  String get kennelQuickEncodeTitle => 'Codificación rápida de camada';

  @override
  String get kennelRequired =>
      'El pack Criadero es necesario para la codificación por lotes';

  @override
  String get litterTag => 'Etiqueta de camada';

  @override
  String get discoveryMarkDone => 'Misión completada';

  @override
  String get notificationPreferences => 'Preferencias de notificación';

  @override
  String get notificationPrefsHint =>
      'Elija qué tipos de notificación desea recibir.';

  @override
  String get notificationPrefsSaved => 'Preferencias guardadas';

  @override
  String get notificationPrefHr => 'Lecturas de frecuencia cardíaca';

  @override
  String get notificationPrefCare => 'Recordatorios de cuidados';

  @override
  String get notificationPrefVisits => 'Visitas';

  @override
  String get notificationPrefMessages => 'Mensajes';

  @override
  String get notificationPrefDiscovery => 'Recorrido de descubrimiento';

  @override
  String get notificationPrefBilling => 'Facturación';

  @override
  String carePostponeDays(int days) {
    return 'Aplazar $days días';
  }

  @override
  String get noCareReminders => 'Sin recordatorios de cuidados pendientes';

  @override
  String get careAddReminder => 'Añadir un recordatorio';

  @override
  String get careSelectPet => 'Mascota';

  @override
  String careDueInDays(int days) {
    return 'Vence en $days días';
  }

  @override
  String get careReferenceModeDone => 'Ya realizado';

  @override
  String get careReferenceModeFirst => 'Primera vez';

  @override
  String get careLastDateLabel => 'Fecha de referencia';

  @override
  String get careLastDateDone => 'Fecha del último cuidado';

  @override
  String get careLastDateFirst => 'Fecha de inicio del ciclo';

  @override
  String get careRecurrenceLabel => 'Recurrencia';

  @override
  String get careRecurrenceNone => 'Ninguna (vencimiento único)';

  @override
  String careRecurrenceDays(int days) {
    return 'Cada $days días';
  }

  @override
  String get careDueDateLabel => 'Vencimiento';

  @override
  String get careDueDateComputed => 'Vencimiento calculado';

  @override
  String get careTooltipDoneWithRecurrence =>
      'Ya realizado: el vencimiento = fecha del último cuidado + recurrencia.';

  @override
  String get careTooltipFirstWithRecurrence =>
      'Primera planificación: indique la fecha de inicio del ciclo. Vencimiento = esa fecha + recurrencia.';

  @override
  String get careTooltipNoRecurrence =>
      'Sin recurrencia: la fecha introducida es el único vencimiento.';

  @override
  String get careTooltipDueExplained =>
      'Vencimiento = fecha de referencia + recurrencia (si está definida).';

  @override
  String get carePickDate => 'Elegir una fecha';

  @override
  String discoveryDayBadge(int day) {
    return 'D$day';
  }

  @override
  String get timelineTypeHeartrate => 'Frecuencia cardíaca';

  @override
  String get timelineTypeMessage => 'Mensaje';

  @override
  String get timelineTypeCare => 'Cuidados';

  @override
  String get timelineTypeVisit => 'Visita';

  @override
  String get timelineTypeEvent => 'Evento';

  @override
  String get visitCancelAction => 'Cancelar solicitud';

  @override
  String get upcomingVisits => 'Próximas visitas';

  @override
  String get timelineEmpty => 'Aún no hay eventos';

  @override
  String get noThreads => 'Sin conversaciones';

  @override
  String get vetInviteSent =>
      'Invitación enviada — la clínica debe aceptar la solicitud';

  @override
  String vetInviteSentNamed(String practice) {
    return 'Solicitud enviada a $practice — la clínica debe aceptarla';
  }

  @override
  String get vetNotFound =>
      'No se ha encontrado ningún veterinario con este correo';

  @override
  String get addVetSearchHint =>
      'Buscamos esta cuenta de veterinario en petsFollow. Si existe, se envía una solicitud de vinculación a la clínica.';

  @override
  String get visitRequested => 'Solicitud de visita enviada';

  @override
  String get primaryVetSet => 'Veterinario principal actualizado';

  @override
  String get visitStatusRequested => 'Solicitada';

  @override
  String get visitStatusConfirmed => 'Confirmada';

  @override
  String get visitStatusDone => 'Completada';

  @override
  String get visitStatusCancelled => 'Cancelada';

  @override
  String get visitStatusReschedulePending => 'Cambio pendiente';

  @override
  String get horseHealthTitle => 'Salud del caballo';

  @override
  String get horseContactsTitle => 'Contactos (herrador, dentista…)';

  @override
  String get horseCompetitionsTitle => 'Competiciones';

  @override
  String get horseContactsSoon =>
      'Active el pack Caballo para gestionar contactos profesionales.';

  @override
  String get horseCompetitionsSoon =>
      'Active el pack Caballo para el calendario de competiciones.';

  @override
  String get horsePackUpsell =>
      'Pack Caballo — herrador, recuento de huevos fecales, contactos y competiciones';

  @override
  String get careTypeFarrier => 'Herrador';

  @override
  String get careTypeFecalEgg => 'Recuento de huevos fecales';

  @override
  String get careTypeVaccination => 'Vacunación';

  @override
  String get careTypeDeworming => 'Desparasitación';

  @override
  String get careTypeVetCheck => 'Revisión veterinaria';

  @override
  String get careTypeDental => 'Cuidado dental';

  @override
  String get careTypeCustom => 'Recordatorio personalizado';

  @override
  String get homeAddFirstVetTitle => 'Añada a su veterinario';

  @override
  String get homeAddFirstVetBody =>
      'Vincule la clínica que sigue a su mascota para compartir lecturas y chatear.';

  @override
  String get homeAddFirstVetCta => 'Añadir un veterinario';

  @override
  String get photoFrameHint =>
      'Centre el hocico — vista previa del perfil de la mascota';

  @override
  String get takePhoto => 'Hacer una foto';

  @override
  String get chooseFromGallery => 'Elegir de la galería';

  @override
  String get attachMedia => 'Adjuntar una foto o un vídeo';

  @override
  String get attachPhoto => 'Foto';

  @override
  String get attachVideo => 'Vídeo';

  @override
  String get openMedia => 'Abrir';

  @override
  String get mediaVideoLabel => 'Vídeo';
}
