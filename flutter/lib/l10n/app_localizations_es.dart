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
  String get emailNotVerified =>
      'Confirme primero su email (enlace enviado al registrarse) y vuelva a iniciar sesión.';

  @override
  String get resendConfirmation => 'Reenviar el correo de confirmación';

  @override
  String get resendConfirmationSent =>
      'Si la cuenta existe y aún no está confirmada, se ha enviado un nuevo correo.';

  @override
  String get resendConfirmationFailed =>
      'No se pudo enviar. Inténtelo de nuevo en un momento.';

  @override
  String get loginOr => 'o';

  @override
  String get loginWithGoogle => 'Continuar con Google';

  @override
  String get loginWithApple => 'Continuar con Apple';

  @override
  String get appleComingSoon => 'El inicio de sesión con Apple llegará pronto.';

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
  String get registerCta => 'Registrarse';

  @override
  String get registerTitle => 'Registrarse';

  @override
  String get registerSubtitle =>
      'Crea una cuenta para seguir la salud de tu animal. Recibirás un email de validación.';

  @override
  String get registerSubmit => 'Registrarse';

  @override
  String get registerSuccess =>
      'Cuenta creada. Abre el enlace del email de validación y vuelve a iniciar sesión en la app.';

  @override
  String get registerInviteNotApplied =>
      'Cuenta creada, pero no se pudo aplicar el código de invitación. Puede volver a introducirlo tras iniciar sesión.';

  @override
  String get registerFailed => 'No se pudo registrar';

  @override
  String get registerEmailExists => 'Este email ya está en uso';

  @override
  String get registerBackToLogin => 'Volver al inicio de sesión';

  @override
  String get confirmEmailTitle => 'Confirmar email';

  @override
  String get confirmEmailLoading => 'Confirmando…';

  @override
  String get confirmEmailDoneTitle => 'Email confirmado';

  @override
  String get confirmEmailDoneSubtitle =>
      'Tu cuenta está activa. Puedes iniciar sesión.';

  @override
  String get confirmEmailFailedTitle => 'Confirmación imposible';

  @override
  String get confirmEmailFailed => 'No se pudo confirmar este email.';

  @override
  String get confirmEmailInvalidLink =>
      'Enlace de confirmación inválido o ya usado.';

  @override
  String get confirmEmailBackToLogin => 'Volver al inicio de sesión';

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
  String get proLightTourToday => 'Hoy';

  @override
  String get proLightTourWeek => '7 días';

  @override
  String get proLightTourAll => 'Todo';

  @override
  String get proLightNoTourToday => 'No hay citas hoy';

  @override
  String get proLightNoTourWeek => 'No hay citas en los próximos 7 días';

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
  String get proLightNewConsultation => 'Nueva consulta';

  @override
  String get proLightConsultationNextTitle => 'Consulta guardada — ¿siguiente?';

  @override
  String get proLightConsultationCtaDaf => 'Crear receta y facturar';

  @override
  String get proLightConsultationCtaInvoice => 'Facturar directamente';

  @override
  String get proLightConsultationCtaDone => 'Terminar';

  @override
  String get proLightReportFinal => 'Finalizado';

  @override
  String get proLightReportHistoryTitle => 'Historial';

  @override
  String get proLightReportHistoryTranscript => 'Original (transcripción)';

  @override
  String get proLightReportHistoryImproved => 'Versión IA';

  @override
  String get proLightReportHistorySaved => 'Versión guardada';

  @override
  String get proLightReportHistoryEmpty => 'Ninguna versión disponible';

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
  String get petAccessSharedRead => 'Compartido · lectura';

  @override
  String get petAccessSharedNotes => 'Compartido · notas';

  @override
  String get petAccessSharedFull => 'Compartido · completo';

  @override
  String get proLightUseGps => 'GPS';

  @override
  String get proLightTranscribeAudio => 'Archivo de audio';

  @override
  String get proLightDictationStart => 'Dictar';

  @override
  String get proLightDictationStop => 'Detener';

  @override
  String get proLightRecordingInProgress => 'Grabación en curso';

  @override
  String get proLightAudioConsentTitle => 'Consentimiento de audio';

  @override
  String get proLightAudioConsentBody =>
      'La grabación solo sirve para redactar el informe. Confirme el consentimiento oral del cliente. El audio se elimina al finalizarlo.';

  @override
  String get proLightAudioConsentAccept => 'Acepto';

  @override
  String get proLightSpecialtyFarrier => 'Herrador';

  @override
  String get proLightSpecialtyPhysio => 'Fisio / osteo';

  @override
  String get proLightSpecialtyBehaviorist => 'Etólogo';

  @override
  String get proLightSpecialtyGroomer => 'Peluquero canino';

  @override
  String get proLightSpecialtyBreeder => 'Criador';

  @override
  String get proLightSpecialtyVetLight => 'Vet light';

  @override
  String get proLightReportHintFarrier =>
      'CR herraje: cascos, herraje, observaciones…';

  @override
  String get proLightEmptyFarrier => 'Ningún caballo / visita compartida';

  @override
  String get proLightMicDenied =>
      'Micrófono denegado — active el acceso en ajustes';

  @override
  String get proLightGpsDenied => 'Ubicación no disponible';

  @override
  String get googleNotConfigured =>
      'Inicio de sesión con Google no configurado';

  @override
  String get googleLoginFailed => 'No se pudo iniciar sesión con Google';

  @override
  String get googleWrongAudience =>
      'Esta cuenta de Google ya es un perfil Pro — use la aplicación web.';

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
  String get acceptTermsTitle => 'Condiciones de uso';

  @override
  String get acceptTermsSubtitle =>
      'Su cuenta fue creada por su clínica. Acepte las condiciones y la política de privacidad para continuar.';

  @override
  String get acceptTermsSubmit => 'Aceptar y continuar';

  @override
  String get acceptTermsFailed =>
      'No se pudo guardar el consentimiento. Inténtelo de nuevo.';

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
  String get exportMyData => 'Exportar mis datos';

  @override
  String get registerConsentPrefix => 'Acepto las ';

  @override
  String get registerConsentMiddle => ' y la ';

  @override
  String get registerConsentRequired =>
      'Debe aceptar las condiciones y la política de privacidad.';

  @override
  String get registerInviteCode => 'Código de invitación (opcional)';

  @override
  String get registerInviteCodeHint =>
      'Introduzca el código del QR / enlace comercial';

  @override
  String get nearbyCommercialTitle => 'Comercial cerca de usted';

  @override
  String get nearbyCommercialHint =>
      'Sin código de invitación, elija un comercial cercano (opcional).';

  @override
  String get nearbyCommercialUseLocation => 'Usar mi ubicación';

  @override
  String get nearbyCommercialPostalCode => 'Código postal';

  @override
  String get nearbyCommercialSearch => 'Buscar';

  @override
  String get nearbyCommercialEmpty => 'No se encontró ningún comercial cerca.';

  @override
  String get nearbyCommercialSkip => 'Omitir';

  @override
  String get nearbyCommercialGeoDenied =>
      'Ubicación denegada — introduzca un código postal.';

  @override
  String nearbyCommercialDistance(String km) {
    return '$km km';
  }

  @override
  String get pushPermissionTitle => 'Notificaciones';

  @override
  String get pushPermissionBody =>
      'petsFollow desea enviarle notificaciones: mensajes de su veterinario, confirmaciones de citas y recordatorios de cuidados. Puede desactivarlas en cualquier momento en los ajustes de la app o del teléfono.';

  @override
  String get pushPermissionContinue => 'Continuar';

  @override
  String exportDataSaved(String path) {
    return 'Exportación guardada: $path';
  }

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
  String get heartRateShort => 'Corazón';

  @override
  String get weightShort => 'Peso';

  @override
  String get recordWeightTitle => 'Registrar peso';

  @override
  String get weightKgLabel => 'Peso (kg)';

  @override
  String get weightCommentLabel => 'Comentario (opcional)';

  @override
  String get weightCommentHint => 'Ej. tras paseo, en ayunas…';

  @override
  String get weightSave => 'Guardar';

  @override
  String get weightSentToVet => 'Peso registrado';

  @override
  String get weightInvalid => 'Indique un peso válido (0,01–999,99 kg)';

  @override
  String weightLastLabel(String kg) {
    return 'Último peso: $kg kg';
  }

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
      'Condiciones de uso — petsFollow\n\nLa app petsFollow permite a los propietarios el seguimiento prescrito (mensajería, recordatorios Care/Horse, lecturas cardíacas), consultar el historial y comunicarse con su veterinario.\n\nLos servicios se prestan según la suscripción seleccionada (pagos vía Stripe).\n\nVersión completa: https://petsfollow.ll-it-sc.be/legal/terms\n\nÚltima actualización: julio de 2026';

  @override
  String get legalPrivacyBody =>
      'Política de privacidad — petsFollow\n\nDatos recogidos: identidad (nombre, correo), datos de la mascota (nombre, especie, raza, fotos), lecturas de frecuencia cardíaca (datos de salud animal), mensajes y medios con la clínica, informes de visita (texto y grabaciones de audio), coordenadas GPS de las visitas a domicilio (profesionales de cuidado), tokens de notificación (FCM), datos de pago tratados por Stripe.\n\nFinalidades: gestión de la cuenta, continuidad de cuidados (incluidas lecturas cardíacas), mensajería veterinaria, informes de visita, notificaciones, facturación.\n\nTratamiento IA: Google Gemini se utiliza para mejorar los informes de visita (audio procesado en tiempo real, no conservado por Google).\n\nEncargados / socios: Google (Sign-In, Firebase Cloud Messaging, Gemini), Stripe (pagos), hosting cloud (GCP).\n\nConservación: hasta la eliminación de la cuenta; cuentas inactivas purgadas tras 3 años; audio de los informes conservado mientras exista el expediente.\n\nDerechos RGPD (acceso, rectificación, supresión, portabilidad): Perfil → Exportar mis datos / Eliminar cuenta, o support@petsfollow.app.\n\nVersión completa: https://petsfollow.ll-it-sc.be/legal/privacy\n\nÚltima actualización: julio de 2026';

  @override
  String get legalNoticeBody =>
      'Aviso legal — petsFollow\n\nEditor: LL-IT-SC / petsFollow\nContacto: support@petsfollow.app\n\nAlojamiento: Google Cloud Platform (conforme al RGPD).\n\nVersión completa: https://petsfollow.ll-it-sc.be/legal/mentions\n\nÚltima actualización: julio de 2026';

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
  String get languageIt => 'Italiano';

  @override
  String get appearance => 'Apariencia';

  @override
  String get themeLight => 'Claro';

  @override
  String get themeDark => 'Oscuro';

  @override
  String get planMonthlyLabel => '3,50 € / mes';

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
  String get editPet => 'Editar mascota';

  @override
  String get petName => 'Nombre';

  @override
  String get petNameRequired => 'Indica el nombre del animal';

  @override
  String get species => 'Especie';

  @override
  String get breed => 'Raza';

  @override
  String get petMicrochipOptional => 'N.º de chip (opcional)';

  @override
  String get petHealthBookNumberOptional => 'N.º de cartilla (opcional)';

  @override
  String get petHealthBookAddPages => 'Añadir fotos de la cartilla';

  @override
  String get petHealthBookReplacePages => 'Sustituir PDF (fotos)';

  @override
  String petHealthBookPagesCount(int count) {
    return '$count página(s) seleccionada(s)';
  }

  @override
  String get petHealthBookPdfAttached => 'PDF de la cartilla adjunto';

  @override
  String get petHealthBookRemovePdf => 'Eliminar';

  @override
  String get petHealthBookOpenPdf => 'Abrir cartilla (PDF)';

  @override
  String petMicrochipLabel(String number) {
    return 'Chip: $number';
  }

  @override
  String petHealthBookNumberLabel(String number) {
    return 'Cartilla: $number';
  }

  @override
  String get errorHealthBookUploadFailed =>
      'Mascota guardada, pero no se pudo enviar la cartilla';

  @override
  String get choosePlan => 'Elija su plan';

  @override
  String get recommended => 'Recomendado';

  @override
  String get autoRenewTitle => 'Renovación automática';

  @override
  String get autoRenewSubtitle => 'Se cobra en cada renovación';

  @override
  String get continueToPayment => 'Guardar y pagar';

  @override
  String get petFormSave => 'Guardar';

  @override
  String get petSavedPendingPayment =>
      'Mascota guardada — actívela para acceder a las funciones';

  @override
  String get paymentFeaturesLocked =>
      'Pago requerido para usar las funciones de esta mascota';

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
      'Se requiere una suscripción para usar esta función';

  @override
  String get errorPhotoUploadFailed =>
      'Mascota creada, pero no se ha podido subir la foto';

  @override
  String get errorCouldNotOpenLink => 'No se ha podido abrir el enlace';

  @override
  String get planMonthlySub => '3,50 € / mes, renovación automática';

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
  String get heartRateNotSupported =>
      'La medición cardíaca no está disponible para esta especie';

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
  String get thresholdAlert =>
      'Alerta: aumento significativo respecto a la lectura anterior';

  @override
  String get validateAndSend => 'Validar y enviar al veterinario';

  @override
  String get heartRateCommentLabel => 'Comentario (opcional)';

  @override
  String get heartRateCommentHint =>
      'Ej. agitado, en reposo, tras el esfuerzo…';

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
      'Añada su primera mascota para empezar el seguimiento prescrito con su veterinario.';

  @override
  String get discoveryTitle => 'Descubra petsFollow';

  @override
  String get discoveryMission => 'Su recorrido petsFollow';

  @override
  String get discoveryDay0Title => 'Paso 1 — Bienvenida';

  @override
  String get discoveryDay0Body =>
      'Cree el perfil de su mascota y descubra la app — mensajería, recordatorios y lecturas (incluida la frecuencia cardíaca).';

  @override
  String get discoveryDay2Title => 'Paso 2 — Primera lectura';

  @override
  String get discoveryDay2Body =>
      'Tome su primera lectura de frecuencia cardíaca y familiarícese con la técnica.';

  @override
  String get discoveryDay4Title => 'Paso 3 — Rutina';

  @override
  String get discoveryDay4Body =>
      'Cree el hábito de medir a diario con recordatorios personalizados.';

  @override
  String get discoveryDay6Title => 'Paso 4 — Compartir con el veterinario';

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
  String get vetLinkRequired =>
      'Vincula un veterinario para activar el seguimiento con tu clínica';

  @override
  String get linkVetAfterSaveTitle => 'Vincular un veterinario';

  @override
  String get linkVetAfterSaveBody =>
      'Vincula una clínica para activar mensajería, visitas y recordatorios de cuidados.';

  @override
  String get linkVetHomeTitle => '¿Vincular un veterinario?';

  @override
  String get linkVetHomeBody =>
      'Tu mascota está guardada. ¿Quieres vincular un veterinario? Es opcional — puedes hacerlo más tarde.';

  @override
  String get linkVetLater => 'Más tarde';

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
      'La reserva en línea no está disponible para esta clínica. Llame a la clínica para concertar una cita.';

  @override
  String get calendarBookingDisabledReschedule =>
      'La reserva en línea no está disponible. Proponga una fecha manualmente.';

  @override
  String get calendarNoSlots =>
      'No hay huecos disponibles en los próximos 14 días.';

  @override
  String get calendarPickSlot => 'Elija un hueco:';

  @override
  String get calendarSelectVet => 'Elija un veterinario:';

  @override
  String get calendarCallPractice => 'Llamar a la clínica';

  @override
  String get calendarNoPhone =>
      'No hay número de teléfono para esta clínica. Contáctela de otra forma.';

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
  String get petBirthDate => 'Fecha de nacimiento';

  @override
  String get petBirthDateInvalid =>
      'Fecha de nacimiento no válida (AAAA-MM-DD)';

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
    return 'P$day';
  }

  @override
  String get timelineTypeHeartrate => 'Frecuencia cardíaca';

  @override
  String get timelineTypeWeight => 'Peso';

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
  String get messageNoMessagesYet => 'Aún no hay mensajes';

  @override
  String get messageNewConversation => 'Nueva conversación';

  @override
  String get messageComposeTitle => 'Nueva conversación';

  @override
  String get messageChoosePro => 'Profesional de cuidados';

  @override
  String get messageChooseClient => 'Cliente';

  @override
  String get messageChoosePet => 'Animal concernido';

  @override
  String get messageChoosePetOptional => 'Animal (opcional)';

  @override
  String get messageGeneralThread => 'Conversación general';

  @override
  String get messageStartConversation => 'Empezar';

  @override
  String get messageLockedTitle => 'Mensajería no disponible';

  @override
  String get messageLockedBody =>
      'Vincule un veterinario para hablar con un profesional de cuidados.';

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
      'Busque por nombre, correo o clínica. Si ya está en petsFollow, se envía una solicitud de vinculación a la clínica.';

  @override
  String get addVetSearchLabel => 'Buscar un veterinario';

  @override
  String get addVetSearchFieldHint => 'Nombre, correo o clínica';

  @override
  String get addVetNotListed => 'Mi veterinario no aparece';

  @override
  String get addVetSuggestTitle => 'Nuevo veterinario';

  @override
  String get addVetSuggestBody =>
      'Indique el correo y el teléfono de la clínica. Les contactaremos para que se unan a petsFollow.';

  @override
  String get addVetSuggestEmail => 'Correo de la clínica';

  @override
  String get addVetSuggestPhone => 'Teléfono';

  @override
  String get addVetSuggestNameOptional => 'Nombre del veterinario (opcional)';

  @override
  String get addVetSuggestCta => 'Enviar sugerencia';

  @override
  String get vetSuggestSent =>
      'Gracias — contactaremos a la clínica. Le avisaremos cuando se una a petsFollow.';

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
  String get takeVideo => 'Grabar un vídeo';

  @override
  String get chooseFromGallery => 'Elegir de la galería';

  @override
  String get attachMedia => 'Adjuntar una foto o un vídeo';

  @override
  String get attachPhoto => 'Foto';

  @override
  String get attachVideo => 'Vídeo';

  @override
  String get compressingMedia => 'Comprimiendo el vídeo…';

  @override
  String get openMedia => 'Abrir';

  @override
  String get mediaVideoLabel => 'Vídeo';

  @override
  String get appInviteTitle => 'QR de invitación app';

  @override
  String get appInviteHint =>
      'Muestre este QR o comparta el enlace. Un nuevo cliente que se registre con este enlace se vincula automáticamente.';

  @override
  String get appInviteHintClient =>
      'Comparte este QR con un amigo. Quedará vinculado a ti (padrinazgo) y podrá unirse a tu clínica si aún no tiene una.';

  @override
  String get appInviteHintCommercial =>
      'Dos enlaces: clientes (app) y clínicas (registro Pro con su código de patrocinio).';

  @override
  String get appInviteHintShort => 'Enlace de descarga y vinculación';

  @override
  String get appInviteCodeLabel => 'Código:';

  @override
  String get appInviteCopy => 'Copiar enlace';

  @override
  String get appInviteCopied => 'Enlace copiado';

  @override
  String get appInviteLoadError => 'No se pudo cargar el QR';

  @override
  String get appInviteRetry => 'Reintentar';

  @override
  String get proLightVetTitle => 'Campo véto';

  @override
  String get commercialFieldTitle => 'Comercial';

  @override
  String get commercialFieldSubtitle =>
      'QR de invitación a clientes y acceso al sitio Pro.';

  @override
  String get commercialManagerFieldSubtitle =>
      'Resultados del equipo, QR de invitación y acceso al sitio Pro.';

  @override
  String get commercialOpenProWeb => 'Abrir sitio Pro';

  @override
  String get managerTeamCta => 'Resultados de mi equipo';

  @override
  String get managerTeamTitle => 'Equipo';

  @override
  String get managerTeamSection => 'Resultados del equipo';

  @override
  String get managerSelfSection => 'Mis resultados';

  @override
  String get managerMembersSection => 'Comerciales';

  @override
  String get managerTeamEmpty => 'Ningún comercial asignado.';

  @override
  String get managerKpiProspects => 'Prospectos';

  @override
  String get managerKpiConverted => 'Convertidos';

  @override
  String get managerKpiConversion => 'Tasa de conversión';

  @override
  String get managerKpiAppointments => 'Citas próximas';

  @override
  String get managerKpiStale => 'Pipeline estancado';

  @override
  String get managerKpiMonthEarned => 'Comisiones (mes)';

  @override
  String get managerKpiLifetimeEarned => 'Comisiones (total)';

  @override
  String get managerKpiVets => 'Veterinarios asignados';

  @override
  String get featureModules => 'Opciones';

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
  String get moduleActivate => 'Activar';

  @override
  String get moduleActive => 'Activo';

  @override
  String get switchProfile => 'Cambiar de perfil';

  @override
  String get profilePersonal => 'Personal';

  @override
  String get profilePro => 'Profesional';

  @override
  String get profileSwitched => 'Profile switched';

  @override
  String get preconsultTitle => 'Preconsulta';

  @override
  String preconsultTitlePet(String petName) {
    return 'Preconsulta — $petName';
  }

  @override
  String get preconsultIntro =>
      'Indique el estado de su animal antes de la visita para que su veterinario pueda prepararse.';

  @override
  String get preconsultComplaint => 'Motivo / queja principal';

  @override
  String get preconsultComplaintRequired => 'Indique el motivo de la visita';

  @override
  String get preconsultDuration => '¿Desde cuándo?';

  @override
  String get preconsultDurationToday => 'Hoy';

  @override
  String get preconsultDurationFewDays => 'Unos días';

  @override
  String get preconsultDurationWeek => 'Aprox. una semana';

  @override
  String get preconsultDurationWeeks => 'Varias semanas';

  @override
  String get preconsultDurationMonths => 'Varios meses';

  @override
  String get preconsultBehavior => 'Comportamiento';

  @override
  String get preconsultBehaviorNormal => 'Normal';

  @override
  String get preconsultBehaviorLethargic => 'Apatía';

  @override
  String get preconsultBehaviorRestless => 'Inquieto';

  @override
  String get preconsultBehaviorAggressive => 'Agresivo';

  @override
  String get preconsultBehaviorAnxious => 'Ansioso';

  @override
  String get preconsultBehaviorOther => 'Otro';

  @override
  String get preconsultAppetite => 'Apetito';

  @override
  String get preconsultThirst => 'Sed';

  @override
  String get preconsultElimination => 'Heces / orina';

  @override
  String get preconsultScaleNormal => 'Normal';

  @override
  String get preconsultScaleDecreased => 'Disminuido';

  @override
  String get preconsultScaleIncreased => 'Aumentado';

  @override
  String get preconsultUrgency => 'Urgencia percibida';

  @override
  String get preconsultUrgencyLow => 'Baja';

  @override
  String get preconsultUrgencyMedium => 'Media';

  @override
  String get preconsultUrgencyHigh => 'Alta';

  @override
  String get preconsultComment => 'Comentario (opcional)';

  @override
  String get preconsultUnknown => 'No lo sé';

  @override
  String get preconsultSubmit => 'Enviar';

  @override
  String get preconsultSubmitted => 'Preconsulta enviada';

  @override
  String get preconsultAlreadySubmitted => 'Ya envió esta preconsulta.';

  @override
  String get preconsultFillCta => 'Completar preconsulta';

  @override
  String get proLightAudioConsentClientCheck =>
      'He obtenido el consentimiento oral del cliente para grabar';

  @override
  String get proLightReportAiProposalBanner =>
      'Propuesta de IA — validación veterinaria obligatoria antes de finalizar (incluye diagnóstico / medicación sugeridos).';

  @override
  String get proLightAiModuleRequired =>
      'Módulo CR IA no activado o prueba caducada — contacte a su comercial petsFollow.';

  @override
  String proLightAiModuleTrialBanner(int days) {
    return 'Prueba CR IA — $days días restantes. Dicte y mejore sus informes.';
  }

  @override
  String get proLightAiModuleInactiveBanner =>
      'CR IA no activado para este centro. La dictación IA no estará disponible.';

  @override
  String get proLightAiModuleVisitScopedBanner =>
      'CR IA disponible si el centro de la visita tiene el módulo activado.';

  @override
  String get supportTitle => 'Reportar un problema';

  @override
  String get supportHint =>
      'Describa el error. Los diagnósticos técnicos de los últimos 15 minutos se adjuntan automáticamente.';

  @override
  String get supportSubject => 'Asunto';

  @override
  String get supportMessage => 'Descripción';

  @override
  String get supportDiagnosticsAttached =>
      'Diagnósticos adjuntos automáticamente (errores, peticiones, configuración).';

  @override
  String get supportSubmit => 'Enviar';

  @override
  String get supportSending => 'Enviando…';

  @override
  String get supportSuccess => 'Mensaje enviado. ¡Gracias!';

  @override
  String get supportErrorRateLimit =>
      'Demasiados tickets recientemente. Inténtelo en una hora.';

  @override
  String get supportErrorTooLarge =>
      'Diagnósticos demasiado grandes. Reinicie la app e inténtelo de nuevo.';

  @override
  String get supportMenu => 'Soporte';

  @override
  String get appInviteHintSales =>
      'Comparte tu código con una clínica o un cliente.';

  @override
  String get appInviteCopyVet => 'Copiar enlace de alta clínica';

  @override
  String get appInviteCopyClient => 'Copiar enlace de invitación cliente';

  @override
  String get sendDossierToPro => 'Enviar a un profesional';

  @override
  String get sendDossierEmailLabel => 'Email del profesional';

  @override
  String get sendDossierEmailHint => 'vet@clinica.es';

  @override
  String get sendDossierConfirm => 'Enviar';

  @override
  String get sendDossierSuccess => 'Historial enviado — enlace válido 24 h.';

  @override
  String get sendDossierInvalidEmail => 'Email no válido.';

  @override
  String get sendDossierPhiWarning =>
      'Este historial contiene datos de salud: informes de visita, cartilla sanitaria y documentos. El enlace es válido 24 h y cualquiera que lo tenga podrá consultarlos.';

  @override
  String get sendDossierPhiConsent =>
      'Acepto compartir estos datos de salud con este profesional.';

  @override
  String get consultationsHistory => 'Consultas';

  @override
  String get consultationTitle => 'Consulta';

  @override
  String consultationTitleWithPet(String petName) {
    return 'Consulta — $petName';
  }

  @override
  String get consultationVisitMeta => 'Visita';

  @override
  String consultationReportBy(String author) {
    return 'Informe de $author';
  }

  @override
  String get consultationReportFallback => 'Informe';

  @override
  String get consultationReportEmpty => '(vacío)';

  @override
  String get consultationReportUnavailable =>
      'No hay informe disponible para esta visita.';

  @override
  String get sendConsultationToVet => 'Enviar a un veterinario';

  @override
  String get sendConsultationEmailLabel => 'Email del veterinario';

  @override
  String get sendConsultationEmailHint => 'vet@clinica.es';

  @override
  String get sendConsultationConfirm => 'Enviar';

  @override
  String get sendConsultationSuccess =>
      'Consulta enviada — enlace válido 24 h.';

  @override
  String get sendConsultationInvalidEmail => 'Email no válido.';

  @override
  String get sendConsultationPhiWarning =>
      'Este informe contiene datos de salud. El enlace es válido 24 h y cualquiera que lo tenga podrá descargar el PDF.';

  @override
  String get sendConsultationPhiConsent =>
      'Acepto compartir este informe con este veterinario.';
}
