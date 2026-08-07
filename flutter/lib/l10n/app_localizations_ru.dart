// ignore: unused_import
import 'package:intl/intl.dart' as intl;
import 'app_localizations.dart';

// ignore_for_file: type=lint

/// The translations for Russian (`ru`).
class AppLocalizationsRu extends AppLocalizations {
  AppLocalizationsRu([String locale = 'ru']) : super(locale);

  @override
  String get appTitle => 'petsFollow';

  @override
  String get appTagline => 'Наблюдение за здоровьем вашего животного';

  @override
  String get email => 'Email';

  @override
  String get password => 'Пароль';

  @override
  String get login => 'Войти';

  @override
  String get loginFailed => 'Не удалось войти';

  @override
  String get emailNotVerified =>
      'Сначала подтвердите свой адрес электронной почты (ссылка отправлена при регистрации), затем войдите снова.';

  @override
  String get resendConfirmation => 'Отправить письмо для подтверждения снова';

  @override
  String get resendConfirmationSent =>
      'Если аккаунт существует и ещё не подтверждён, отправлено новое письмо.';

  @override
  String get resendConfirmationFailed =>
      'Не удалось отправить. Попробуйте через мгновение.';

  @override
  String get loginOr => 'или';

  @override
  String get loginWithGoogle => 'Продолжить с Google';

  @override
  String get loginWithApple => 'Продолжить с Apple';

  @override
  String get appleComingSoon => 'Вход через Apple будет доступен скоро.';

  @override
  String get twoFaTitle => 'Проверка 2FA';

  @override
  String get twoFaSubtitle =>
      'Введите 6-значный код из вашего приложения аутентификации.';

  @override
  String get twoFaCode => 'Код аутентификатора';

  @override
  String get twoFaSubmit => 'Подтвердить';

  @override
  String get twoFaBack => 'Вернуться к входу';

  @override
  String get twoFaInvalid => 'Код 2FA некорректен или устарел';

  @override
  String get forgotPassword => 'Забыли пароль?';

  @override
  String get forgotPasswordTitle => 'Забыли пароль';

  @override
  String get forgotPasswordSubtitle =>
      'Укажите email вашего аккаунта. Если аккаунт существует, будет отправлена ссылка для сброса.';

  @override
  String get forgotPasswordSubmit => 'Отправить ссылку';

  @override
  String get forgotPasswordBack => 'Вернуться к входу';

  @override
  String get forgotPasswordFailed => 'Не удалось отправить';

  @override
  String get forgotPasswordSentTitle => 'Письмо отправлено';

  @override
  String forgotPasswordSent(String email) {
    return 'Если аккаунт для $email существует, ссылка отправлена. Откройте её в браузере, чтобы выбрать новый пароль.';
  }

  @override
  String get emailRequired => 'Введите корректный адрес электронной почты';

  @override
  String get resetPasswordTitle => 'Новый пароль';

  @override
  String get resetPasswordSubtitle => 'Минимум 8 символов.';

  @override
  String get resetPasswordToken => 'Токен для сброса';

  @override
  String get resetPasswordSubmit => 'Сохранить';

  @override
  String get resetPasswordBackToLogin => 'Перейти к входу';

  @override
  String get resetPasswordInvalidLink => 'Некорректная ссылка для сброса';

  @override
  String get resetPasswordFailed => 'Не удалось сбросить пароль';

  @override
  String get resetPasswordDoneTitle => 'Пароль обновлён';

  @override
  String get resetPasswordDoneSubtitle => 'Теперь вы можете войти.';

  @override
  String get fullName => 'Полное имя';

  @override
  String get registerCta => 'Зарегистрироваться';

  @override
  String get registerTitle => 'Регистрация';

  @override
  String get registerSubtitle =>
      'Создайте аккаунт, чтобы наблюдать за здоровьем вашего животного. Вам будет отправлено письмо для подтверждения.';

  @override
  String get registerSubmit => 'Зарегистрироваться';

  @override
  String get registerSuccess =>
      'Аккаунт создан. Откройте ссылку в письме для подтверждения, затем вернитесь и войдите в приложении.';

  @override
  String get registerInviteNotApplied =>
      'Аккаунт создан, но код приглашения не удалось применить. Вы сможете ввести его снова после входа.';

  @override
  String get registerFailed => 'Не удалось зарегистрироваться';

  @override
  String get registerEmailExists =>
      'Этот адрес электронной почты уже используется';

  @override
  String get registerBackToLogin => 'Вернуться к входу';

  @override
  String get confirmEmailTitle => 'Подтверждение адреса электронной почты';

  @override
  String get confirmEmailLoading => 'Подтверждение выполняется…';

  @override
  String get confirmEmailDoneTitle => 'Адрес электронной почты подтверждён';

  @override
  String get confirmEmailDoneSubtitle =>
      'Ваш аккаунт активирован. Вы можете войти.';

  @override
  String get confirmEmailFailedTitle => 'Не удалось подтвердить';

  @override
  String get confirmEmailFailed =>
      'Не удалось подтвердить этот адрес электронной почты.';

  @override
  String get confirmEmailInvalidLink =>
      'Ссылка для подтверждения некорректна или уже использована.';

  @override
  String get confirmEmailBackToLogin => 'Вернуться к входу';

  @override
  String get vetUseProWeb =>
      'Полный аккаунт ветеринара используется на веб-сайте Pro.';

  @override
  String get unsupportedRoleApp =>
      'Этот аккаунт нельзя использовать в приложении pets. Воспользуйтесь веб-сайтом Pro.';

  @override
  String get proLightTitle => 'Pro для выезда';

  @override
  String get proLightAgenda => 'Расписание';

  @override
  String get proLightClients => 'Клиенты';

  @override
  String get proLightPets => 'Животные';

  @override
  String get proLightLoadError => 'Не удалось загрузить';

  @override
  String get proLightNoVisits => 'Нет визитов';

  @override
  String get proLightTourToday => 'Сегодня';

  @override
  String get proLightTourWeek => '7 дней';

  @override
  String get proLightTourAll => 'Всё';

  @override
  String get proLightNoTourToday => 'Сегодня нет визитов';

  @override
  String get proLightNoTourWeek => 'Нет визитов за 7 дней';

  @override
  String get proLightNoClients => 'Нет общих клиентов';

  @override
  String get proLightNoPets => 'Нет общих животных';

  @override
  String get proLightAddress => 'Адрес';

  @override
  String get proLightOpenMaps => 'Maps';

  @override
  String get proLightReportTitle => 'Отчёт';

  @override
  String get proLightReportHint => 'Заметки о визите…';

  @override
  String get proLightImproveAi => 'Улучшить (ИИ)';

  @override
  String get proLightFinalizeReport => 'Завершить';

  @override
  String get proLightNewConsultation => 'Новая консультация';

  @override
  String get proLightConsultationNextTitle =>
      'Консультация сохранена — что дальше?';

  @override
  String get proLightConsultationCtaDaf => 'Создать DAF и выставить счёт';

  @override
  String get proLightConsultationCtaInvoice => 'Выставить счёт напрямую';

  @override
  String get proLightConsultationCtaDone => 'Завершить';

  @override
  String get proLightReportFinal => 'Завершено';

  @override
  String get proLightReportHistoryTitle => 'История';

  @override
  String get proLightReportHistoryTranscript => 'Оригинал (транскрипция)';

  @override
  String get proLightReportHistoryImproved => 'Версия ИИ';

  @override
  String get proLightReportHistorySaved => 'Сохранённая версия';

  @override
  String get proLightReportHistoryEmpty => 'Нет доступных версий';

  @override
  String get proLightSettings => 'Настройки';

  @override
  String get proLightSpecialty => 'Специализация';

  @override
  String get proLightDocuments => 'Документы';

  @override
  String get proLightNoDocuments => 'Нет документов';

  @override
  String get proLightTimeline => 'Хронология';

  @override
  String get proLightNoTimeline => 'Нет событий';

  @override
  String get proLightReminders => 'Напоминания';

  @override
  String get proLightNoReminders => 'Нет напоминаний';

  @override
  String get proLightLitterTag => 'Тег / помёт';

  @override
  String get proLightActionFailed => 'Действие невозможно';

  @override
  String get proLightReadOnly => 'Доступ только для чтения';

  @override
  String get petAccessSharedRead => 'Общий · чтение';

  @override
  String get petAccessSharedNotes => 'Общий · заметки';

  @override
  String get petAccessSharedFull => 'Общий · полный';

  @override
  String get proLightUseGps => 'GPS';

  @override
  String get proLightTranscribeAudio => 'Аудиофайл';

  @override
  String get proLightDictationStart => 'Диктовать';

  @override
  String get proLightDictationStop => 'Остановить';

  @override
  String get proLightRecordingInProgress => 'Запись идёт';

  @override
  String get proLightAudioConsentTitle => 'Согласие на аудиозапись';

  @override
  String get proLightAudioConsentBody =>
      'Запись используется только для создания отчёта. Подтвердите устное согласие клиента. Аудио удаляется после завершения отчёта.';

  @override
  String get proLightAudioConsentAccept => 'Я согласен';

  @override
  String get proLightSpecialtyFarrier => 'Кузнец';

  @override
  String get proLightSpecialtyPhysio => 'Физио / остеопат';

  @override
  String get proLightSpecialtyBehaviorist => 'Зоопсихолог';

  @override
  String get proLightSpecialtyGroomer => 'Грумер';

  @override
  String get proLightSpecialtyBreeder => 'Заводчик';

  @override
  String get proLightSpecialtyVetLight => 'Ветеринар light';

  @override
  String get proLightReportHintFarrier =>
      'Отчёт о подковке: копыта, подкова, наблюдения…';

  @override
  String get proLightEmptyFarrier => 'Нет общих лошадей / вмешательств';

  @override
  String get proLightMicDenied =>
      'Доступ к микрофону отклонён — разрешите его в настройках';

  @override
  String get proLightGpsDenied => 'Местоположение недоступно';

  @override
  String get googleNotConfigured => 'Вход через Google не настроен';

  @override
  String get googleLoginFailed => 'Не удалось войти через Google';

  @override
  String get googleWrongAudience =>
      'Этот аккаунт Google уже является профилем Pro — воспользуйтесь веб-приложением.';

  @override
  String get myPets => 'Мои животные';

  @override
  String get myData => 'Мои данные';

  @override
  String get settings => 'Настройки';

  @override
  String get logout => 'Завершить сеанс';

  @override
  String get save => 'Сохранить';

  @override
  String get cancel => 'Отмена';

  @override
  String get firstName => 'Ваше имя';

  @override
  String get currentPassword => 'Текущий пароль';

  @override
  String get newPassword => 'Новый пароль';

  @override
  String get confirmNewPassword => 'Подтвердите пароль';

  @override
  String get changePassword => 'Изменить пароль';

  @override
  String get forceChangePasswordTitle => 'Изменить пароль';

  @override
  String get forceChangePasswordSubtitle =>
      'Этот аккаунт создан с временным паролем. Выберите свой, чтобы продолжить.';

  @override
  String get forceChangePasswordSubmit => 'Сохранить и продолжить';

  @override
  String get acceptTermsTitle => 'Условия использования';

  @override
  String get acceptTermsSubtitle =>
      'Ваш аккаунт создан вашей клиникой. Примите условия и политику конфиденциальности, чтобы продолжить.';

  @override
  String get acceptTermsSubmit => 'Принять и продолжить';

  @override
  String get acceptTermsFailed =>
      'Не удалось сохранить согласие. Попробуйте снова.';

  @override
  String get passwordTooShort => 'Минимум 8 символов';

  @override
  String get passwordMismatch => 'Пароли не совпадают';

  @override
  String get passwordChangeFailed => 'Не удалось изменить пароль';

  @override
  String get deleteAccount => 'Удалить аккаунт';

  @override
  String get deleteAccountConfirm =>
      'Это действие необратимо. Все ваши животные и данные будут удалены.';

  @override
  String get exportMyData => 'Экспортировать мои данные';

  @override
  String get registerConsentPrefix => 'Я принимаю ';

  @override
  String get registerConsentMiddle => ' и ';

  @override
  String get registerConsentRequired =>
      'Вы должны принять условия и политику конфиденциальности.';

  @override
  String get registerInviteCode => 'Код приглашения (необязательно)';

  @override
  String get registerInviteCodeHint =>
      'Введите код из QR / коммерческой ссылки';

  @override
  String get nearbyCommercialTitle => 'Менеджер по продажам рядом с вами';

  @override
  String get nearbyCommercialHint =>
      'Без кода приглашения выберите менеджера по продажам рядом (необязательно).';

  @override
  String get nearbyCommercialUseLocation => 'Использовать моё местоположение';

  @override
  String get nearbyCommercialPostalCode => 'Почтовый индекс';

  @override
  String get nearbyCommercialSearch => 'Найти';

  @override
  String get nearbyCommercialEmpty =>
      'Рядом не найдено менеджеров по продажам.';

  @override
  String get nearbyCommercialSkip => 'Не привязывать';

  @override
  String get nearbyCommercialGeoDenied =>
      'Доступ к местоположению отклонён — введите почтовый индекс.';

  @override
  String nearbyCommercialDistance(String km) {
    return '$km км';
  }

  @override
  String get pushPermissionTitle => 'Уведомления';

  @override
  String get pushPermissionBody =>
      'petsFollow хочет отправлять вам уведомления: сообщения от вашего ветеринара, подтверждения визитов и напоминания об уходе. Вы можете отключить их в любое время в настройках приложения или телефона.';

  @override
  String get pushPermissionContinue => 'Продолжить';

  @override
  String exportDataSaved(String path) {
    return 'Экспорт сохранён: $path';
  }

  @override
  String get profileSaved => 'Профиль сохранён';

  @override
  String get changePhoto => 'Изменить фото';

  @override
  String get addPhoto => 'Добавить фото';

  @override
  String get photoUpdated => 'Фото обновлено';

  @override
  String get passwordChanged => 'Пароль изменён';

  @override
  String greeting(String name) {
    return 'Здравствуйте, $name,';
  }

  @override
  String get latestValues => 'Последние значения';

  @override
  String get startMeasurement => 'НАЧАТЬ ИЗМЕРЕНИЕ';

  @override
  String get heartRateShort => 'Дыхание';

  @override
  String get weightShort => 'Вес';

  @override
  String get recordWeightTitle => 'Записать вес';

  @override
  String get weightKgLabel => 'Вес (кг)';

  @override
  String get weightCommentLabel => 'Комментарий (необязательно)';

  @override
  String get weightCommentHint => 'Напр. после прогулки, натощак…';

  @override
  String get weightSave => 'Сохранить';

  @override
  String get weightSentToVet => 'Вес сохранён';

  @override
  String get weightInvalid => 'Укажите корректный вес (0,01–999,99 кг)';

  @override
  String weightLastLabel(String kg) {
    return 'Последний вес: $kg кг';
  }

  @override
  String get choosePetForMeasurement => 'Выберите животное';

  @override
  String get chooseDuration => 'Длительность измерения';

  @override
  String durationSeconds(int seconds) {
    return '$seconds с';
  }

  @override
  String get howToMeasure => 'Как измерять?';

  @override
  String get howToMeasureIntro =>
      'Измеряйте частоту дыхания вашего животного в состоянии покоя.';

  @override
  String get howToMeasureStep1 =>
      '1. Успокойте животное, оно должно лежать или сидеть.';

  @override
  String get howToMeasureStep2 =>
      '2. Положите руку на грудную клетку и нажимайте при каждом вдохе в течение указанного времени.';

  @override
  String get howToMeasureStep3 =>
      '3. Подтвердите измерение, чтобы отправить его вашему ветеринару.';

  @override
  String get howToMeasureWhyTitle => 'Почему это важно?';

  @override
  String get howToMeasureWhyBody =>
      'Регулярное наблюдение за частотой дыхания позволяет выявить изменения и скорректировать лечение вместе с вашим ветеринаром.';

  @override
  String get reminders => 'Напоминания';

  @override
  String get remindersHint =>
      'Получайте ежедневное напоминание сделать измерение дыхания.';

  @override
  String get remindersEnabled => 'Включить напоминания';

  @override
  String get remindersTime => 'Время напоминания';

  @override
  String get remindersSaved => 'Напоминания сохранены';

  @override
  String get legalTermsTitle => 'Общие условия использования';

  @override
  String get legalPrivacyTitle => 'Политика конфиденциальности';

  @override
  String get legalNoticeTitle => 'Юридическая информация';

  @override
  String get legalOpenOnline => 'Посмотреть версию онлайн';

  @override
  String get legalTermsBody =>
      'Общие условия использования — petsFollow\n\nПриложение petsFollow позволяет владельцам животных вести назначенное наблюдение (сообщения, напоминания Care/Horse, измерения дыхания), просматривать историю и общаться со своим ветеринаром.\n\nУслуги предоставляются в рамках выбранной подписки (оплата через Stripe). Пользователь обязуется использовать приложение по назначению.\n\nПолная версия: https://petsfollow.ll-it-sc.be/legal/terms\n\nДата обновления: июль 2026';

  @override
  String get legalPrivacyBody =>
      'Политика конфиденциальности — petsFollow\n\nСобираемые данные: идентификационные данные (имя, email), данные животного (имя, вид, порода, фотографии), результаты измерения частоты дыхания (данные о здоровье животного), сообщения и медиафайлы, которыми вы обмениваетесь с клиникой, отчёты о визитах (текст и аудиозаписи), GPS-координаты домашних визитов (специалисты по уходу), токены уведомлений (FCM), платёжные данные, которые обрабатывает Stripe.\n\nЦели: управление аккаунтом, непрерывность помощи (включая измерения дыхания), общение с ветеринаром, отчёты о визитах, уведомления, выставление счетов.\n\nОбработка с ИИ: Google Gemini используется для улучшения отчётов о визитах (аудио обрабатывается в реальном времени и не сохраняется Google) и, в клиентском приложении (экспериментальный модуль), для упрощённого изложения завершённых отчётов и разговорного определения срочности.\n\nСубподрядчики / партнёры: Google (Sign-In, Firebase Cloud Messaging, Gemini), Stripe (платежи), облачный хостинг (GCP).\n\nХранение: до удаления аккаунта; неактивные аккаунты удаляются через 3 года; аудио отчётов хранится в течение существования карты.\n\nПрава по GDPR (доступ, исправление, удаление, переносимость): Профиль → Экспортировать мои данные / Удалить аккаунт, или обратитесь на barbara@petsfollow.app.\n\nПолная версия: https://petsfollow.ll-it-sc.be/legal/privacy\n\nДата обновления: июль 2026';

  @override
  String get legalNoticeBody =>
      'Юридическая информация — petsFollow\n\nИздатель: LL-IT-SC / petsFollow\nКонтакт: barbara@petsfollow.app · 0478.02.33.77\n\nХостинг: Google Cloud Platform (соответствие GDPR).\n\nОтветственный за публикацию: petsFollow.\n\nПолная версия: https://petsfollow.ll-it-sc.be/legal/mentions\n\nДата обновления: июль 2026';

  @override
  String get language => 'Язык';

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
  String get languageUk => 'Українська';

  @override
  String get languageRu => 'Русский';

  @override
  String get appearance => 'Внешний вид';

  @override
  String get themeLight => 'Светлая';

  @override
  String get themeDark => 'Тёмная';

  @override
  String get planMonthlyLabel => '3,50 € / месяц';

  @override
  String get planAnnualLabel => '35 € / год';

  @override
  String get planTriennialLabel => '95 € / 3 года';

  @override
  String get planQuinquennialLabel => '145 € / 5 лет';

  @override
  String get pushNewMessage => 'Новое сообщение';

  @override
  String get pushVisitConfirmed => 'Визит подтверждён';

  @override
  String get pushVisitProposed => 'Предложение визита';

  @override
  String get pushVisitReschedule => 'Перенос визита';

  @override
  String get notifChannelMessages => 'Сообщения';

  @override
  String get notifChannelVisits => 'Визиты';

  @override
  String get notifChannelCare => 'Уход';

  @override
  String get paymentResume => 'Продолжить оплату';

  @override
  String get manageSubscription => 'Управлять подпиской';

  @override
  String get heartRate => 'Измерение дыхания';

  @override
  String get history => 'История';

  @override
  String get vetMessaging => 'Сообщения ветеринару';

  @override
  String get badgeAutoRenew => 'Автопродление';

  @override
  String get badgeActive => 'Активно';

  @override
  String get badgePendingPayment => 'Ожидает оплаты';

  @override
  String badgeExpiresOn(String date) {
    return 'истекает $date';
  }

  @override
  String get newPet => 'Новое животное';

  @override
  String get editPet => 'Изменить животное';

  @override
  String get petName => 'Имя';

  @override
  String get petNameRequired => 'Укажите имя животного';

  @override
  String get species => 'Вид';

  @override
  String get breed => 'Порода';

  @override
  String get petMicrochipOptional => 'Номер микрочипа (необязательно)';

  @override
  String get petHealthBookNumberOptional => 'Номер книжки (необязательно)';

  @override
  String get petDomicileLocation => 'Место содержания / конюшня';

  @override
  String get petDomicileHint => 'Напр. Конюшня «Вербы» — Брюссель';

  @override
  String get petFoodChainStatus => 'Статус в отношении пищевой цепи';

  @override
  String get petFoodChainCompanion => 'Домашнее животное (вне цепи)';

  @override
  String get petFoodChainFoodProducing =>
      'Продуктивное животное / пищевая цепь';

  @override
  String get petFoodChainExcluded => 'Исключено из пищевой цепи';

  @override
  String get petHealthBookAddPages => 'Добавить фотографии книжки';

  @override
  String get petHealthBookReplacePages => 'Заменить PDF (фотографии)';

  @override
  String petHealthBookPagesCount(int count) {
    return 'Выбрано страниц: $count';
  }

  @override
  String get petHealthBookPdfAttached => 'PDF книжки прикреплён';

  @override
  String get petHealthBookRemovePdf => 'Удалить';

  @override
  String get petHealthBookOpenPdf => 'Открыть книжку (PDF)';

  @override
  String petMicrochipLabel(String number) {
    return 'Микрочип: $number';
  }

  @override
  String petHealthBookNumberLabel(String number) {
    return 'Книжка: $number';
  }

  @override
  String get errorHealthBookUploadFailed =>
      'Животное сохранено, но книжку не удалось отправить';

  @override
  String get choosePlan => 'Выберите свой тариф';

  @override
  String get recommended => 'Рекомендуется';

  @override
  String get autoRenewTitle => 'Продлевать автоматически';

  @override
  String get autoRenewSubtitle => 'Списание на каждую дату продления';

  @override
  String get continueToPayment => 'Сохранить и оплатить';

  @override
  String get petFormSave => 'Сохранить';

  @override
  String get petSavedPendingPayment =>
      'Животное сохранено — активируйте его, чтобы получить доступ к функциям';

  @override
  String get paymentFeaturesLocked =>
      'Для использования функций этого животного требуется оплата';

  @override
  String get paymentConfirmed => 'Оплата подтверждена — животное активно';

  @override
  String get paymentPending =>
      'Платёж в ожидании — вы сможете продолжить позже';

  @override
  String errorGeneric(String message) {
    return 'Ошибка: $message';
  }

  @override
  String get errorNetwork =>
      'Не удалось подключиться. Проверьте сеть и попробуйте снова.';

  @override
  String get retryAction => 'Попробовать снова';

  @override
  String get errorMediaTooLarge => 'Файл слишком большой (макс. 25 МБ)';

  @override
  String get errorInvalidMediaType =>
      'Формат не поддерживается (JPEG, PNG, WebP, MP4, MOV, WebM)';

  @override
  String get errorPaymentRequired =>
      'Для использования этой функции требуется подписка';

  @override
  String get errorPhotoUploadFailed =>
      'Животное создано, но фото не удалось отправить';

  @override
  String get errorCouldNotOpenLink => 'Не удалось открыть ссылку';

  @override
  String get planMonthlySub => '3,50 € / месяц, продлевается автоматически';

  @override
  String planAnnualSub(String price) {
    return '$price, продлевается автоматически';
  }

  @override
  String get planTriennialSub =>
      '95 € каждые 3 года, продлевается автоматически';

  @override
  String get planQuinquennialSub => '145 € за 5 лет, единоразовый платёж';

  @override
  String planOneTime(String price) {
    return '$price, единоразовый платёж';
  }

  @override
  String get heartRateInstructions =>
      'Нажимайте при каждом вдохе в течение времени, указанного вашим ветеринаром.';

  @override
  String heartRateInstructionsDuration(int seconds) {
    return 'Нажимайте при каждом вдохе в течение $seconds секунд.';
  }

  @override
  String get heartRateNoDurationConfigured =>
      'Для этой клиники не настроена длительность измерения. Обратитесь к своему ветеринару.';

  @override
  String get heartRateNotSupported =>
      'Измерение дыхания недоступно для этого вида';

  @override
  String get start => 'Начать';

  @override
  String secondsLeft(int seconds) {
    return '$seconds с';
  }

  @override
  String beatsCount(int count) {
    return '$count вдохов';
  }

  @override
  String get tapHere => 'Нажимайте здесь при каждом вдохе';

  @override
  String bpmLabel(String bpm) {
    return 'Частота: $bpm';
  }

  @override
  String beatsLabel(int count) {
    return 'Вдохов: $count';
  }

  @override
  String get thresholdAlert =>
      'Оповещение: значительный рост по сравнению с предыдущим измерением';

  @override
  String get validateAndSend => 'Подтвердить и отправить ветеринару';

  @override
  String get heartRateCommentLabel => 'Комментарий (необязательно)';

  @override
  String get heartRateCommentHint =>
      'Напр. беспокойное, в покое, после нагрузки…';

  @override
  String get restart => 'Начать снова';

  @override
  String get sentToVet => 'Измерение отправлено ветеринару';

  @override
  String get navHome => 'Главная';

  @override
  String get navPets => 'Животные';

  @override
  String get navCare => 'Уход';

  @override
  String get navMessages => 'Сообщения';

  @override
  String get navProfile => 'Профиль';

  @override
  String get speciesDog => 'Собака';

  @override
  String get speciesCat => 'Кот';

  @override
  String get speciesHorse => 'Лошадь';

  @override
  String get speciesDonkey => 'Осёл';

  @override
  String get speciesCattle => 'Крупный рогатый скот';

  @override
  String get speciesSheep => 'Овца';

  @override
  String get speciesGoat => 'Коза';

  @override
  String get speciesPig => 'Свинья';

  @override
  String get speciesPoultry => 'Домашняя птица';

  @override
  String get speciesRabbit => 'Кролик';

  @override
  String get speciesAlpaca => 'Альпака';

  @override
  String get speciesLlama => 'Лама';

  @override
  String get speciesOther => 'Другое';

  @override
  String get careComingSoon => 'Напоминания об уходе будут доступны скоро';

  @override
  String get emptyPetsTitle => 'Нет животных';

  @override
  String get emptyPetsBody =>
      'Добавьте своё первое животное, чтобы начать назначенное наблюдение вместе с вашим ветеринаром.';

  @override
  String get discoveryTitle => 'Знакомство с petsFollow';

  @override
  String get homeRecentActivity => 'Недавняя активность';

  @override
  String get homePetTipsTitle => 'Советы для ваших питомцев';

  @override
  String get homePetTipsDisclaimer =>
      'Общие советы — ваш ветеринар остаётся ориентиром.';

  @override
  String get discoveryMission => 'Ваш путь в petsFollow';

  @override
  String get discoveryDay0Title => 'Шаг 1 — Добро пожаловать';

  @override
  String get discoveryDay0Body =>
      'Создайте профиль своего животного и познакомьтесь с приложением — сообщения, напоминания и измерения (включая частоту дыхания).';

  @override
  String get discoveryDay2Title => 'Шаг 2 — Первое измерение';

  @override
  String get discoveryDay2Body =>
      'Сделайте первое измерение дыхания и освойте технику.';

  @override
  String get discoveryDay4Title => 'Шаг 3 — Привычка';

  @override
  String get discoveryDay4Body =>
      'Создайте привычку ежедневных измерений с помощью персональных напоминаний.';

  @override
  String get discoveryDay6Title => 'Шаг 4 — Обмен с ветеринаром';

  @override
  String get discoveryDay6Body =>
      'Ваши измерения доступны вашему ветеринару для оптимального наблюдения.';

  @override
  String get myVets => 'Мои ветеринары';

  @override
  String get addVetByEmail => 'Добавить ветеринара по email';

  @override
  String get vetEmailHint => 'email@klinika.vet';

  @override
  String get noVets => 'Нет привязанных ветеринаров';

  @override
  String get vetLinkRequired =>
      'Привяжите ветеринара, чтобы активировать наблюдение с вашей клиникой';

  @override
  String get linkVetAfterSaveTitle => 'Привязать ветеринара';

  @override
  String get linkVetAfterSaveBody =>
      'Привяжите клинику, чтобы активировать сообщения, визиты и напоминания об уходе.';

  @override
  String get linkVetHomeTitle => 'Привязать ветеринара?';

  @override
  String get linkVetHomeBody =>
      'Ваше животное зарегистрировано. Хотите привязать ветеринара? Это необязательно — вы сможете сделать это позже.';

  @override
  String get linkVetLater => 'Позже';

  @override
  String get primaryVet => 'Основной ветеринар';

  @override
  String get setPrimaryVet => 'Назначить основным ветеринаром';

  @override
  String get careTitle => 'Уход';

  @override
  String get careDone => 'Выполнено';

  @override
  String get carePostpone => 'Отложить';

  @override
  String get careOverdue => 'Просрочено';

  @override
  String get visitHistory => 'История визитов';

  @override
  String get requestVisit => 'Запросить визит';

  @override
  String get calendarBookingDisabled =>
      'Онлайн-бронирование недоступно для этой клиники. Позвоните в клинику, чтобы записаться.';

  @override
  String get calendarBookingDisabledReschedule =>
      'Онлайн-бронирование недоступно. Предложите дату вручную.';

  @override
  String get calendarNoSlots => 'Нет свободных мест на следующие 14 дней.';

  @override
  String get calendarPickSlot => 'Выберите время:';

  @override
  String get calendarSelectVet => 'Выберите ветеринара:';

  @override
  String get calendarCallPractice => 'Позвонить в клинику';

  @override
  String get calendarNoPhone =>
      'Для этой клиники не указан номер телефона. Свяжитесь с ней другим способом.';

  @override
  String get visitConfirm => 'Подтвердить';

  @override
  String get visitProposeReschedule => 'Предложить другое время';

  @override
  String get visitRescheduleProposed => 'Предложение переноса отправлено';

  @override
  String get paymentSuccessSnack => 'Платёж получен — обновление…';

  @override
  String get paymentCancelSnack => 'Платёж отменён';

  @override
  String get visitRejectReschedule => 'Отказать в переносе';

  @override
  String get visitAcceptReschedule => 'Принять новое время';

  @override
  String get upcomingVisit => 'Предстоящий визит';

  @override
  String get notificationHrTitle => 'petsFollow';

  @override
  String get notificationHrBody =>
      'Время сделать измерение дыхания для вашего животного';

  @override
  String get reviewAskTitle => 'Вам нравится petsFollow?';

  @override
  String get reviewAskYes => 'Да, оценить приложение';

  @override
  String get reviewAskNo => 'Позже';

  @override
  String get careTypeMedication => 'Медикамент';

  @override
  String get horseAddContact => 'Добавить контакт';

  @override
  String get horseAddCompetition => 'Добавить соревнование';

  @override
  String get horseContactName => 'Имя';

  @override
  String get horseContactRole => 'Роль';

  @override
  String get horseCompetitionTitle => 'Событие';

  @override
  String get horseCompetitionDate => 'Дата (ГГГГ-ММ-ДД)';

  @override
  String familyHouseholdTitle(int count) {
    return 'Домохозяйство Family — $count животных';
  }

  @override
  String kennelHouseholdTitle(int count) {
    return 'Домохозяйство Kennel — $count животных';
  }

  @override
  String get familyHouseholdNext => 'Следующие напоминания домохозяйства';

  @override
  String get familyPetLimit =>
      'Пакет для домохозяйства уже активен или в процессе покупки';

  @override
  String get familyRequiresTwoPets =>
      'Пакет Family требует не менее 2 животных';

  @override
  String get kennelPackHint =>
      'Пакет Kennel — ≥6 животных, −15 % на последующие подписки';

  @override
  String get kennelRequiresSixPets =>
      'Пакет Kennel требует не менее 6 животных';

  @override
  String get kennelQuickEncodeTitle => 'Внесение помёта (питомник)';

  @override
  String get kennelRequired => 'Для пакетного внесения требуется пакет Kennel';

  @override
  String get litterTag => 'Тег помёта';

  @override
  String get petBirthDate => 'Дата рождения';

  @override
  String get petBirthDateInvalid => 'Некорректная дата рождения (ГГГГ-ММ-ДД)';

  @override
  String get discoveryMarkDone => 'Задание выполнено';

  @override
  String get discoveryDismiss => 'Скрыть путь';

  @override
  String get appBuildInfo => 'Версия приложения';

  @override
  String get notificationPreferences => 'Настройки уведомлений';

  @override
  String get notificationPrefsHint =>
      'Выберите типы уведомлений, которые хотите получать.';

  @override
  String get notificationPrefsSaved => 'Настройки сохранены';

  @override
  String get notificationPrefHr => 'Измерения дыхания';

  @override
  String get notificationPrefCare => 'Напоминания об уходе';

  @override
  String get notificationPrefVisits => 'Визиты';

  @override
  String get notificationPrefMessages => 'Сообщения';

  @override
  String get notificationPrefDiscovery => 'Ознакомительный путь';

  @override
  String get notificationPrefBilling => 'Выставление счетов';

  @override
  String get notificationPrefSms => 'SMS';

  @override
  String get notificationPrefSmsHint =>
      'Подтверждения и напоминания о визитах по SMS. Ответ STOP также отключает этот канал.';

  @override
  String carePostponeDays(int days) {
    return 'Отложить на $days дней';
  }

  @override
  String get noCareReminders => 'Нет активных напоминаний об уходе';

  @override
  String get careAddReminder => 'Добавить напоминание';

  @override
  String get careSelectPet => 'Животное';

  @override
  String careDueInDays(int days) {
    return 'Срок через $days дней';
  }

  @override
  String get careReferenceModeDone => 'Уже выполнено';

  @override
  String get careReferenceModeFirst => 'Впервые';

  @override
  String get careLastDateLabel => 'Последняя дата';

  @override
  String get careLastDateDone => 'Дата последнего ухода';

  @override
  String get careLastDateFirst => 'Начальная дата цикла';

  @override
  String get careRecurrenceLabel => 'Периодичность';

  @override
  String get careRecurrenceNone => 'Нет (один срок)';

  @override
  String careRecurrenceDays(int days) {
    return 'Каждые $days дней';
  }

  @override
  String get careDueDateLabel => 'Срок';

  @override
  String get careDueDateComputed => 'Рассчитанный срок';

  @override
  String get careTooltipDoneWithRecurrence =>
      'Уход уже выполнен: срок = дата последнего ухода + периодичность.';

  @override
  String get careTooltipFirstWithRecurrence =>
      'Первое планирование: укажите начальную дату цикла. Срок = эта дата + периодичность.';

  @override
  String get careTooltipNoRecurrence =>
      'Без периодичности: введённая дата и есть единственный срок.';

  @override
  String get careTooltipDueExplained =>
      'Срок = последняя дата + периодичность (если задана).';

  @override
  String get carePickDate => 'Выбрать дату';

  @override
  String discoveryDayBadge(int day) {
    return 'Ш$day';
  }

  @override
  String get timelineTypeHeartrate => 'Частота дыхания';

  @override
  String get timelineTypeWeight => 'Вес';

  @override
  String get timelineTypeMessage => 'Сообщение';

  @override
  String get timelineTypeCare => 'Уход';

  @override
  String get timelineTypeVisit => 'Визит';

  @override
  String get timelineTypeEvent => 'Событие';

  @override
  String get visitCancelAction => 'Отменить запрос';

  @override
  String get upcomingVisits => 'Следующие визиты';

  @override
  String get timelineEmpty => 'Пока нет событий';

  @override
  String get noThreads => 'Нет разговоров';

  @override
  String get messageNoMessagesYet => 'Сообщений ещё нет';

  @override
  String get messageNewConversation => 'Новый разговор';

  @override
  String get messageComposeTitle => 'Новый разговор';

  @override
  String get messageChoosePro => 'Специалист по уходу';

  @override
  String get messageChooseClient => 'Клиент';

  @override
  String get messageChoosePet => 'Животное, о котором речь';

  @override
  String get messageChoosePetOptional => 'Животное (необязательно)';

  @override
  String get messageGeneralThread => 'Общий разговор';

  @override
  String get messageStartConversation => 'Начать';

  @override
  String get messageLockedTitle => 'Сообщения недоступны';

  @override
  String get messageLockedBody =>
      'Привяжите ветеринара, чтобы общаться со специалистом по уходу.';

  @override
  String get vetInviteSent =>
      'Приглашение отправлено — клиника должна принять запрос';

  @override
  String vetInviteSentNamed(String practice) {
    return 'Запрос отправлен в $practice — клиника должна его принять';
  }

  @override
  String get vetNotFound =>
      'Ветеринар с таким адресом электронной почты не найден';

  @override
  String get addVetSearchHint =>
      'Искать по имени, email или клинике. Если ветеринар уже в petsFollow, запрос на привязку будет отправлен клинике.';

  @override
  String get addVetSearchLabel => 'Найти ветеринара';

  @override
  String get addVetSearchFieldHint => 'Имя, email или клиника';

  @override
  String get addVetNotListed => 'Моего ветеринара нет в списке';

  @override
  String get addVetSuggestTitle => 'Новый ветеринар';

  @override
  String get addVetSuggestBody =>
      'Укажите email и телефон клиники. Мы обратимся к ней, чтобы она присоединилась к petsFollow.';

  @override
  String get addVetSuggestEmail => 'Email клиники';

  @override
  String get addVetSuggestPhone => 'Телефон';

  @override
  String get addVetSuggestNameOptional => 'Имя ветеринара (необязательно)';

  @override
  String get addVetSuggestCta => 'Отправить предложение';

  @override
  String get vetSuggestSent =>
      'Спасибо — мы обратимся к клинике. Вы получите уведомление, когда она присоединится к petsFollow.';

  @override
  String get visitRequested => 'Запрос на визит отправлен';

  @override
  String get primaryVetSet => 'Основной ветеринар обновлён';

  @override
  String get visitStatusRequested => 'Запрошено';

  @override
  String get visitStatusConfirmed => 'Подтверждено';

  @override
  String get visitStatusDone => 'Завершено';

  @override
  String get visitStatusCancelled => 'Отменено';

  @override
  String get visitStatusReschedulePending => 'Перенос в ожидании';

  @override
  String get horseHealthTitle => 'Здоровье лошадей';

  @override
  String get horseContactsTitle => 'Контакты (кузнец, стоматолог…)';

  @override
  String get horseCompetitionsTitle => 'Соревнования';

  @override
  String get horseContactsSoon =>
      'Активируйте пакет Horse, чтобы управлять профессиональными контактами.';

  @override
  String get horseCompetitionsSoon =>
      'Активируйте пакет Horse для календаря соревнований.';

  @override
  String get horsePackUpsell =>
      'Пакет Horse — кузнец, копроскопия, контакты и соревнования';

  @override
  String get careTypeFarrier => 'Кузнец';

  @override
  String get careTypeFecalEgg => 'Копроскопия';

  @override
  String get careTypeVaccination => 'Вакцинация';

  @override
  String get careTypeDeworming => 'Дегельминтизация';

  @override
  String get careTypeVetCheck => 'Ветеринарный осмотр';

  @override
  String get careTypeDental => 'Стоматологический уход';

  @override
  String get careTypeCustom => 'Своё напоминание';

  @override
  String get homeAddFirstVetTitle => 'Добавьте своего ветеринара';

  @override
  String get homeAddFirstVetBody =>
      'Привяжите клинику, которая наблюдает за вашим животным, чтобы обмениваться измерениями и общаться.';

  @override
  String get homeAddFirstVetCta => 'Добавить ветеринара';

  @override
  String get photoFrameHint =>
      'Расположите морду в центре — предварительный просмотр карточки животного';

  @override
  String get takePhoto => 'Сделать фото';

  @override
  String get takeVideo => 'Снять видео';

  @override
  String get chooseFromGallery => 'Выбрать из галереи';

  @override
  String get attachMedia => 'Прикрепить фото или видео';

  @override
  String get attachPhoto => 'Фото';

  @override
  String get attachVideo => 'Видео';

  @override
  String get compressingMedia => 'Сжатие видео…';

  @override
  String get openMedia => 'Открыть';

  @override
  String get mediaVideoLabel => 'Видео';

  @override
  String get appInviteTitle => 'QR-приглашение в приложение';

  @override
  String get appInviteHint =>
      'Покажите этот QR или поделитесь ссылкой. Новый клиент, который зарегистрируется по этой ссылке, привязывается автоматически.';

  @override
  String get appInviteHintClient =>
      'Поделитесь этим QR с близким человеком. Он будет привязан к вам (приглашение) и сможет присоединиться к вашей клинике, если у него ещё нет своей.';

  @override
  String get appInviteHintCommercial =>
      'Две ссылки: клиенты (приложение) и клиники (регистрация Pro с вашим кодом приглашения).';

  @override
  String get appInviteHintShort => 'Ссылка для скачивания и привязки';

  @override
  String get appInviteCodeLabel => 'Код:';

  @override
  String get appInviteCopy => 'Копировать ссылку';

  @override
  String get appInviteCopied => 'Ссылка скопирована';

  @override
  String get appInviteLoadError => 'Не удалось загрузить QR';

  @override
  String get appInviteRetry => 'Попробовать снова';

  @override
  String get proLightVetTitle => 'Ветеринар на выезде';

  @override
  String get commercialFieldTitle => 'Менеджер по продажам';

  @override
  String get commercialFieldSubtitle =>
      'QR-приглашение для клиентов и доступ к сайту Pro.';

  @override
  String get commercialManagerFieldSubtitle =>
      'Результаты команды, QR-приглашение и доступ к сайту Pro.';

  @override
  String get commercialOpenProWeb => 'Открыть сайт Pro';

  @override
  String get managerTeamCta => 'Результаты моей команды';

  @override
  String get managerTeamTitle => 'Команда';

  @override
  String get managerTeamSection => 'Результаты команды';

  @override
  String get managerSelfSection => 'Мои результаты';

  @override
  String get managerMembersSection => 'Менеджеры по продажам';

  @override
  String get managerTeamEmpty => 'Нет прикреплённых менеджеров по продажам.';

  @override
  String get managerKpiProspects => 'Потенциальные клиенты';

  @override
  String get managerKpiConverted => 'Конвертированные';

  @override
  String get managerKpiConversion => 'Коэффициент конверсии';

  @override
  String get managerKpiAppointments => 'Предстоящие встречи';

  @override
  String get managerKpiStale => 'Просроченный пайплайн';

  @override
  String get managerKpiMonthEarned => 'Вознаграждения (месяц)';

  @override
  String get managerKpiLifetimeEarned => 'Вознаграждения (всего)';

  @override
  String get managerKpiVets => 'Назначенные ветеринары';

  @override
  String get featureModules => 'Опции';

  @override
  String get featureModulesCatalog => 'Познакомиться с опциями';

  @override
  String get featureModulesSubtitle =>
      'Активируйте Care+, Horse, Kennel или «Домохозяйство» в соответствии с вашими потребностями.';

  @override
  String get moduleCarePlus => 'Care+';

  @override
  String get moduleCarePlusDesc =>
      'Расширенные напоминания об уходе для всех ваших животных.';

  @override
  String get moduleHorse => 'Horse';

  @override
  String get moduleHorseDesc =>
      'Профессиональные контакты и соревнования для лошадей.';

  @override
  String get moduleKennel => 'Kennel';

  @override
  String get moduleKennelDesc => 'Быстрое внесение питомника / помёта.';

  @override
  String get moduleFamily => 'Домохозяйство';

  @override
  String get moduleFamilyDesc => 'Обзор домохозяйства ваших животных.';

  @override
  String get moduleActivate => 'Активировать';

  @override
  String get moduleActive => 'Активно';

  @override
  String get switchProfile => 'Изменить профиль';

  @override
  String get profilePersonal => 'Личный';

  @override
  String get profilePro => 'Профессиональный';

  @override
  String get profileSwitched => 'Профиль изменён';

  @override
  String get preconsultTitle => 'Предконсультация';

  @override
  String preconsultTitlePet(String petName) {
    return 'Предконсультация — $petName';
  }

  @override
  String get preconsultIntro =>
      'Опишите состояние вашего животного перед визитом. Ваш ветеринар сможет подготовиться.';

  @override
  String get preconsultComplaint => 'Причина / основная жалоба';

  @override
  String get preconsultComplaintRequired => 'Укажите причину визита';

  @override
  String get preconsultDuration => 'Как долго?';

  @override
  String get preconsultDurationToday => 'Сегодня';

  @override
  String get preconsultDurationFewDays => 'Несколько дней';

  @override
  String get preconsultDurationWeek => 'Около недели';

  @override
  String get preconsultDurationWeeks => 'Несколько недель';

  @override
  String get preconsultDurationMonths => 'Несколько месяцев';

  @override
  String get preconsultBehavior => 'Поведение';

  @override
  String get preconsultBehaviorNormal => 'Обычное';

  @override
  String get preconsultBehaviorLethargic => 'Апатичное';

  @override
  String get preconsultBehaviorRestless => 'Беспокойное';

  @override
  String get preconsultBehaviorAggressive => 'Агрессивное';

  @override
  String get preconsultBehaviorAnxious => 'Тревожное';

  @override
  String get preconsultBehaviorOther => 'Другое';

  @override
  String get preconsultAppetite => 'Аппетит';

  @override
  String get preconsultThirst => 'Жажда';

  @override
  String get preconsultElimination => 'Стул / мочеиспускание';

  @override
  String get preconsultScaleNormal => 'В норме';

  @override
  String get preconsultScaleDecreased => 'Снижен';

  @override
  String get preconsultScaleIncreased => 'Повышен';

  @override
  String get preconsultUrgency => 'Ощущаемая срочность';

  @override
  String get preconsultUrgencyLow => 'Низкая';

  @override
  String get preconsultUrgencyMedium => 'Средняя';

  @override
  String get preconsultUrgencyHigh => 'Высокая';

  @override
  String get preconsultComment => 'Комментарий (необязательно)';

  @override
  String get preconsultUnknown => 'Не знаю';

  @override
  String get preconsultSubmit => 'Отправить';

  @override
  String get preconsultSubmitted => 'Предконсультация отправлена';

  @override
  String get preconsultAlreadySubmitted =>
      'Вы уже отправили эту предконсультацию.';

  @override
  String get preconsultFillCta => 'Заполнить предконсультацию';

  @override
  String get proLightAudioConsentClientCheck =>
      'Я получил устное согласие клиента на запись';

  @override
  String get proLightReportAiProposalBanner =>
      'Предложение ИИ — обязательное подтверждение перед завершением (включая диагноз / лечение).';

  @override
  String get proLightAiModuleRequired =>
      'Функция ИИ-отчётов отключена для этой клиники — обратитесь в поддержку petsFollow.';

  @override
  String proLightAiModuleTrialBanner(int days) {
    return 'Пробный период ИИ-отчётов — осталось $days дн. Диктуйте, затем улучшайте свои отчёты.';
  }

  @override
  String get proLightAiModuleInactiveBanner =>
      'ИИ-отчёты не активированы для этой клиники. Диктовка с ИИ будет недоступна.';

  @override
  String get proLightAiModuleVisitScopedBanner =>
      'ИИ-отчёты доступны, если в клинике, где проходит визит, включён модуль.';

  @override
  String get supportTitle => 'Сообщить о проблеме';

  @override
  String get supportHint =>
      'Опишите ошибку. Техническая диагностика за последние 15 минут будет добавлена автоматически.';

  @override
  String get supportSubject => 'Тема';

  @override
  String get supportMessage => 'Описание';

  @override
  String get supportDiagnosticsAttached =>
      'Диагностика добавлена автоматически (ошибки, запросы, конфигурация).';

  @override
  String get supportSubmit => 'Отправить';

  @override
  String get supportSending => 'Отправка…';

  @override
  String get supportSuccess => 'Сообщение отправлено. Спасибо!';

  @override
  String get supportErrorTooLarge =>
      'Диагностика слишком большая. Перезапустите приложение и попробуйте снова.';

  @override
  String get supportMenu => 'Поддержка';

  @override
  String get appInviteHintSales =>
      'Поделитесь своим кодом приглашения с клиникой (регистрация Pro) или с клиентом (приглашение в приложение).';

  @override
  String get appInviteCopyVet => 'Копировать ссылку для регистрации клиники';

  @override
  String get appInviteCopyClient => 'Копировать ссылку-приглашение для клиента';

  @override
  String get sendDossierToPro => 'Отправить специалисту';

  @override
  String get sendDossierEmailLabel => 'Email специалиста';

  @override
  String get sendDossierEmailHint => 'vet@klinika.be';

  @override
  String get sendDossierConfirm => 'Отправить';

  @override
  String get sendDossierSuccess =>
      'Карта отправлена — ссылка действительна 24 ч.';

  @override
  String get sendDossierInvalidEmail => 'Некорректный адрес электронной почты.';

  @override
  String get sendDossierPhiWarning =>
      'Эта карта содержит данные о здоровье: отчёты о визитах, медицинскую книжку и документы. Ссылка действительна 24 ч, и любой, у кого она есть, сможет их посмотреть.';

  @override
  String get sendDossierPhiConsent =>
      'Я согласен поделиться этими данными о здоровье с этим специалистом.';

  @override
  String get consultationsHistory => 'Консультации';

  @override
  String get consultationTitle => 'Консультация';

  @override
  String consultationTitleWithPet(String petName) {
    return 'Консультация — $petName';
  }

  @override
  String get consultationVisitMeta => 'Визит';

  @override
  String consultationReportBy(String author) {
    return 'Отчёт от $author';
  }

  @override
  String get consultationReportFallback => 'Отчёт';

  @override
  String get consultationReportEmpty => '(пусто)';

  @override
  String get consultationAvailableCta => 'Доступно';

  @override
  String get consultationPendingCta => 'Черновик';

  @override
  String get sendConsultationToVet => 'Отправить ветеринару';

  @override
  String get sendConsultationEmailLabel => 'Email ветеринара';

  @override
  String get sendConsultationEmailHint => 'vet@klinika.be';

  @override
  String get sendConsultationConfirm => 'Отправить';

  @override
  String get sendConsultationSuccess =>
      'Консультация отправлена — ссылка действительна 24 ч.';

  @override
  String get sendConsultationInvalidEmail =>
      'Некорректный адрес электронной почты.';

  @override
  String get sendConsultationPhiWarning =>
      'Этот отчёт содержит данные о здоровье. Ссылка действительна 24 ч, и любой, у кого она есть, сможет скачать PDF.';

  @override
  String get sendConsultationPhiConsent =>
      'Я согласен поделиться этим отчётом с этим ветеринаром.';

  @override
  String get clientAiDevBadge => 'dev';

  @override
  String get clientAiSectionTitle => 'Помощь ИИ';

  @override
  String get clientAiExplainCta => 'Понять мой отчёт';

  @override
  String get clientAiExplainTitle => 'Ваш отчёт простыми словами';

  @override
  String get clientAiExplainDisclaimer =>
      'Это не медицинское заключение. Всегда следуйте указаниям вашего ветеринара.';

  @override
  String get clientAiExplainLoading => 'Подготовка объяснения…';

  @override
  String get clientAiExplainListTitle => 'Понять отчёт';

  @override
  String get clientAiExplainListSubtitle =>
      'Упрощённое объяснение завершённого отчёта';

  @override
  String get clientAiExplainListEmpty =>
      'Пока нет завершённых отчётов для объяснения.';

  @override
  String get clientAiTriageTitle => 'Экстренная помощь 24/7';

  @override
  String get clientAiTriageSubtitle =>
      'Опишите ситуацию — мы оценим степень срочности.';

  @override
  String get clientAiTriageHint => 'Напр.: моя собака съела шоколад…';

  @override
  String get clientAiTriageSend => 'Отправить';

  @override
  String get clientAiTriageLevelGreen => 'Совет';

  @override
  String get clientAiTriageLevelOrange => 'Записаться на визит';

  @override
  String get clientAiTriageLevelRed => 'Экстренный случай';

  @override
  String get clientAiTriageWatchSigns =>
      'Признаки, за которыми следует наблюдать';

  @override
  String get clientAiTriageCallPractice => 'Позвонить в клинику';

  @override
  String get clientAiTriageBookVisit => 'Записаться на визит';

  @override
  String get clientAiTriageMessageVet => 'Связаться с моим ветеринаром';

  @override
  String get clientAiTriageSelectPet => 'Для какого животного?';

  @override
  String get clientAiTriageNoPet => 'Продолжить без животного';

  @override
  String get clientAiTriageStart => 'Начать';

  @override
  String get clientAiTriageEmergencyFallback =>
      'Если вы не можете связаться со своей клиникой, немедленно обратитесь в местную ветеринарную службу неотложной помощи.';

  @override
  String get clientAiTriageOpenMessages => 'Открыть сообщения';

  @override
  String get bloodPressureShort => 'Давление';

  @override
  String get recordBloodPressureTitle => 'Записать давление';

  @override
  String get bpSystolicLabel => 'Систолическое (мм рт. ст.)';

  @override
  String get bpDiastolicLabel => 'Диастолическое (мм рт. ст.)';

  @override
  String get bpMethodLabel => 'Метод';

  @override
  String get bpMethodDoppler => 'Доплер';

  @override
  String get bpMethodOscillometric => 'Осциллометрический';

  @override
  String get bpMethodUnknown => 'Не указано';

  @override
  String get bloodPressureInvalid => 'Укажите корректное давление (СИС ≥ ДИА)';

  @override
  String get bloodPressureSaved => 'Давление сохранено';

  @override
  String get labsTitle => 'Анализы';

  @override
  String get labsEmpty => 'Нет доступных анализов';

  @override
  String get labsResults => 'Результаты';

  @override
  String labsAbnormalCount(int count) {
    return '$count вне нормы';
  }

  @override
  String get labsOpen => 'Посмотреть анализы';

  @override
  String get labsFlagLow => 'Низкий';

  @override
  String get labsFlagHigh => 'Высокий';

  @override
  String get labsFlagNormal => 'В норме';

  @override
  String get labsOpenDocument => 'Открыть документ';

  @override
  String get bpMethodInvasive => 'Инвазивный';

  @override
  String get bpSiteLabel => 'Место (необязательно)';

  @override
  String get bpCommentLabel => 'Комментарий (необязательно)';

  @override
  String get bloodPressureSave => 'Сохранить';

  @override
  String get calendarSelectSite => 'Выберите площадку';

  @override
  String get proformaValidateCta => 'Подтвердить проформу';
}
