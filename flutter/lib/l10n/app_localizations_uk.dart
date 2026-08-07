// ignore: unused_import
import 'package:intl/intl.dart' as intl;
import 'app_localizations.dart';

// ignore_for_file: type=lint

/// The translations for Ukrainian (`uk`).
class AppLocalizationsUk extends AppLocalizations {
  AppLocalizationsUk([String locale = 'uk']) : super(locale);

  @override
  String get appTitle => 'petsFollow';

  @override
  String get appTagline => 'Спостереження за здоров\'ям вашої тварини';

  @override
  String get email => 'Email';

  @override
  String get password => 'Пароль';

  @override
  String get login => 'Увійти';

  @override
  String get loginFailed => 'Не вдалося увійти';

  @override
  String get emailNotVerified =>
      'Спочатку підтвердьте свою електронну адресу (посилання надіслано при реєстрації), потім увійдіть знову.';

  @override
  String get resendConfirmation => 'Надіслати лист для підтвердження ще раз';

  @override
  String get resendConfirmationSent =>
      'Якщо акаунт існує і ще не підтверджений, надіслано новий лист.';

  @override
  String get resendConfirmationFailed =>
      'Не вдалося надіслати. Спробуйте за мить.';

  @override
  String get loginOr => 'або';

  @override
  String get loginWithGoogle => 'Продовжити з Google';

  @override
  String get loginWithApple => 'Продовжити з Apple';

  @override
  String get appleComingSoon => 'Вхід через Apple буде доступний невдовзі.';

  @override
  String get twoFaTitle => 'Перевірка 2FA';

  @override
  String get twoFaSubtitle =>
      'Введіть 6-цифровий код із вашого застосунку автентифікації.';

  @override
  String get twoFaCode => 'Код автентифікатора';

  @override
  String get twoFaSubmit => 'Підтвердити';

  @override
  String get twoFaBack => 'Повернутися до входу';

  @override
  String get twoFaInvalid => 'Код 2FA некоректний або застарілий';

  @override
  String get forgotPassword => 'Забули пароль?';

  @override
  String get forgotPasswordTitle => 'Забули пароль';

  @override
  String get forgotPasswordSubtitle =>
      'Укажіть email вашого акаунта. Якщо акаунт існує, буде надіслано посилання для скидання.';

  @override
  String get forgotPasswordSubmit => 'Надіслати посилання';

  @override
  String get forgotPasswordBack => 'Повернутися до входу';

  @override
  String get forgotPasswordFailed => 'Не вдалося надіслати';

  @override
  String get forgotPasswordSentTitle => 'Лист надіслано';

  @override
  String forgotPasswordSent(String email) {
    return 'Якщо акаунт для $email існує, посилання надіслано. Відкрийте його в браузері, щоб обрати новий пароль.';
  }

  @override
  String get emailRequired => 'Введіть коректну електронну адресу';

  @override
  String get resetPasswordTitle => 'Новий пароль';

  @override
  String get resetPasswordSubtitle => 'Мінімум 8 символів.';

  @override
  String get resetPasswordToken => 'Токен для скидання';

  @override
  String get resetPasswordSubmit => 'Зберегти';

  @override
  String get resetPasswordBackToLogin => 'Перейти до входу';

  @override
  String get resetPasswordInvalidLink => 'Некоректне посилання для скидання';

  @override
  String get resetPasswordFailed => 'Не вдалося скинути пароль';

  @override
  String get resetPasswordDoneTitle => 'Пароль оновлено';

  @override
  String get resetPasswordDoneSubtitle => 'Тепер ви можете увійти.';

  @override
  String get fullName => 'Повне ім\'я';

  @override
  String get registerCta => 'Зареєструватися';

  @override
  String get registerTitle => 'Реєстрація';

  @override
  String get registerSubtitle =>
      'Створіть акаунт, щоб спостерігати за здоров\'ям вашої тварини. Вам буде надіслано лист для підтвердження.';

  @override
  String get registerSubmit => 'Зареєструватися';

  @override
  String get registerSuccess =>
      'Акаунт створено. Відкрийте посилання в листі для підтвердження, потім повернітеся й увійдіть у застосунку.';

  @override
  String get registerInviteNotApplied =>
      'Акаунт створено, але код запрошення не вдалося застосувати. Ви зможете ввести його знову після входу.';

  @override
  String get registerFailed => 'Не вдалося зареєструватися';

  @override
  String get registerEmailExists => 'Ця електронна адреса вже використовується';

  @override
  String get registerBackToLogin => 'Повернутися до входу';

  @override
  String get confirmEmailTitle => 'Підтвердження електронної адреси';

  @override
  String get confirmEmailLoading => 'Підтвердження виконується…';

  @override
  String get confirmEmailDoneTitle => 'Електронну адресу підтверджено';

  @override
  String get confirmEmailDoneSubtitle =>
      'Ваш акаунт активовано. Ви можете увійти.';

  @override
  String get confirmEmailFailedTitle => 'Не вдалося підтвердити';

  @override
  String get confirmEmailFailed =>
      'Не вдалося підтвердити цю електронну адресу.';

  @override
  String get confirmEmailInvalidLink =>
      'Посилання для підтвердження некоректне або вже використане.';

  @override
  String get confirmEmailBackToLogin => 'Повернутися до входу';

  @override
  String get vetUseProWeb =>
      'Повний акаунт ветеринара використовується на вебсайті Pro.';

  @override
  String get unsupportedRoleApp =>
      'Цей акаунт не можна використовувати в застосунку pets. Скористайтеся вебсайтом Pro.';

  @override
  String get proLightTitle => 'Pro для виїзду';

  @override
  String get proLightAgenda => 'Розклад';

  @override
  String get proLightClients => 'Клієнти';

  @override
  String get proLightPets => 'Тварини';

  @override
  String get proLightLoadError => 'Не вдалося завантажити';

  @override
  String get proLightNoVisits => 'Немає візитів';

  @override
  String get proLightTourToday => 'Сьогодні';

  @override
  String get proLightTourWeek => '7 днів';

  @override
  String get proLightTourAll => 'Усе';

  @override
  String get proLightNoTourToday => 'Сьогодні немає візитів';

  @override
  String get proLightNoTourWeek => 'Немає візитів за 7 днів';

  @override
  String get proLightNoClients => 'Немає спільних клієнтів';

  @override
  String get proLightNoPets => 'Немає спільних тварин';

  @override
  String get proLightAddress => 'Адреса';

  @override
  String get proLightOpenMaps => 'Maps';

  @override
  String get proLightReportTitle => 'Звіт';

  @override
  String get proLightReportHint => 'Нотатки про візит…';

  @override
  String get proLightImproveAi => 'Покращити (ШІ)';

  @override
  String get proLightFinalizeReport => 'Завершити';

  @override
  String get proLightNewConsultation => 'Нова консультація';

  @override
  String get proLightConsultationNextTitle =>
      'Консультацію збережено — що далі?';

  @override
  String get proLightConsultationCtaDaf => 'Створити DAF і виставити рахунок';

  @override
  String get proLightConsultationCtaInvoice => 'Виставити рахунок напряму';

  @override
  String get proLightConsultationCtaDone => 'Завершити';

  @override
  String get proLightReportFinal => 'Завершено';

  @override
  String get proLightReportHistoryTitle => 'Історія';

  @override
  String get proLightReportHistoryTranscript => 'Оригінал (транскрипція)';

  @override
  String get proLightReportHistoryImproved => 'Версія ШІ';

  @override
  String get proLightReportHistorySaved => 'Збережена версія';

  @override
  String get proLightReportHistoryEmpty => 'Немає доступних версій';

  @override
  String get proLightSettings => 'Налаштування';

  @override
  String get proLightSpecialty => 'Спеціалізація';

  @override
  String get proLightDocuments => 'Документи';

  @override
  String get proLightNoDocuments => 'Немає документів';

  @override
  String get proLightTimeline => 'Хронологія';

  @override
  String get proLightNoTimeline => 'Немає подій';

  @override
  String get proLightReminders => 'Нагадування';

  @override
  String get proLightNoReminders => 'Немає нагадувань';

  @override
  String get proLightLitterTag => 'Тег / послід';

  @override
  String get proLightActionFailed => 'Дія неможлива';

  @override
  String get proLightReadOnly => 'Доступ лише для читання';

  @override
  String get petAccessSharedRead => 'Спільний · читання';

  @override
  String get petAccessSharedNotes => 'Спільний · нотатки';

  @override
  String get petAccessSharedFull => 'Спільний · повний';

  @override
  String get proLightUseGps => 'GPS';

  @override
  String get proLightTranscribeAudio => 'Аудіофайл';

  @override
  String get proLightDictationStart => 'Диктувати';

  @override
  String get proLightDictationStop => 'Зупинити';

  @override
  String get proLightRecordingInProgress => 'Запис триває';

  @override
  String get proLightAudioConsentTitle => 'Згода на аудіозапис';

  @override
  String get proLightAudioConsentBody =>
      'Запис використовується лише для створення звіту. Підтвердьте усну згоду клієнта. Аудіо видаляється після завершення звіту.';

  @override
  String get proLightAudioConsentAccept => 'Я погоджуюся';

  @override
  String get proLightSpecialtyFarrier => 'Коваль';

  @override
  String get proLightSpecialtyPhysio => 'Фізіо / остеопат';

  @override
  String get proLightSpecialtyBehaviorist => 'Зоопсихолог';

  @override
  String get proLightSpecialtyGroomer => 'Грумер';

  @override
  String get proLightSpecialtyBreeder => 'Розплідник';

  @override
  String get proLightSpecialtyVetLight => 'Ветеринар light';

  @override
  String get proLightReportHintFarrier =>
      'Звіт про підковування: копита, підкова, спостереження…';

  @override
  String get proLightEmptyFarrier => 'Немає спільних коней / втручань';

  @override
  String get proLightMicDenied =>
      'Доступ до мікрофона відхилено — дозвольте його в налаштуваннях';

  @override
  String get proLightGpsDenied => 'Місцезнаходження недоступне';

  @override
  String get googleNotConfigured => 'Вхід через Google не налаштовано';

  @override
  String get googleLoginFailed => 'Не вдалося увійти через Google';

  @override
  String get googleWrongAudience =>
      'Цей акаунт Google уже є профілем Pro — скористайтеся вебзастосунком.';

  @override
  String get myPets => 'Мої тварини';

  @override
  String get myData => 'Мої дані';

  @override
  String get settings => 'Налаштування';

  @override
  String get logout => 'Завершити сеанс';

  @override
  String get save => 'Зберегти';

  @override
  String get cancel => 'Скасувати';

  @override
  String get firstName => 'Ваше ім\'я';

  @override
  String get currentPassword => 'Поточний пароль';

  @override
  String get newPassword => 'Новий пароль';

  @override
  String get confirmNewPassword => 'Підтвердьте пароль';

  @override
  String get changePassword => 'Змінити пароль';

  @override
  String get forceChangePasswordTitle => 'Змінити пароль';

  @override
  String get forceChangePasswordSubtitle =>
      'Цей акаунт створено з тимчасовим паролем. Оберіть власний, щоб продовжити.';

  @override
  String get forceChangePasswordSubmit => 'Зберегти та продовжити';

  @override
  String get acceptTermsTitle => 'Умови користування';

  @override
  String get acceptTermsSubtitle =>
      'Ваш акаунт створено вашою клінікою. Прийміть умови та політику конфіденційності, щоб продовжити.';

  @override
  String get acceptTermsSubmit => 'Прийняти та продовжити';

  @override
  String get acceptTermsFailed =>
      'Не вдалося зберегти згоду. Спробуйте ще раз.';

  @override
  String get passwordTooShort => 'Мінімум 8 символів';

  @override
  String get passwordMismatch => 'Паролі не збігаються';

  @override
  String get passwordChangeFailed => 'Не вдалося змінити пароль';

  @override
  String get deleteAccount => 'Видалити акаунт';

  @override
  String get deleteAccountConfirm =>
      'Цю дію неможливо скасувати. Усі ваші тварини та дані будуть видалені.';

  @override
  String get exportMyData => 'Експортувати мої дані';

  @override
  String get registerConsentPrefix => 'Я приймаю ';

  @override
  String get registerConsentMiddle => ' і ';

  @override
  String get registerConsentRequired =>
      'Ви повинні прийняти умови та політику конфіденційності.';

  @override
  String get registerInviteCode => 'Код запрошення (необов\'язково)';

  @override
  String get registerInviteCodeHint =>
      'Введіть код із QR / комерційного посилання';

  @override
  String get nearbyCommercialTitle => 'Менеджер з продажу поблизу вас';

  @override
  String get nearbyCommercialHint =>
      'Без коду запрошення виберіть менеджера з продажу поблизу (необов\'язково).';

  @override
  String get nearbyCommercialUseLocation => 'Використати моє місцезнаходження';

  @override
  String get nearbyCommercialPostalCode => 'Поштовий індекс';

  @override
  String get nearbyCommercialSearch => 'Знайти';

  @override
  String get nearbyCommercialEmpty =>
      'Поблизу не знайдено менеджерів з продажу.';

  @override
  String get nearbyCommercialSkip => 'Не прив\'язувати';

  @override
  String get nearbyCommercialGeoDenied =>
      'Доступ до місцезнаходження відхилено — введіть поштовий індекс.';

  @override
  String nearbyCommercialDistance(String km) {
    return '$km км';
  }

  @override
  String get pushPermissionTitle => 'Повідомлення';

  @override
  String get pushPermissionBody =>
      'petsFollow хоче надсилати вам повідомлення: листи від вашого ветеринара, підтвердження візитів і нагадування про догляд. Ви можете вимкнути їх будь-коли в налаштуваннях застосунку або телефона.';

  @override
  String get pushPermissionContinue => 'Продовжити';

  @override
  String exportDataSaved(String path) {
    return 'Експорт збережено: $path';
  }

  @override
  String get profileSaved => 'Профіль збережено';

  @override
  String get changePhoto => 'Змінити фото';

  @override
  String get addPhoto => 'Додати фото';

  @override
  String get photoUpdated => 'Фото оновлено';

  @override
  String get passwordChanged => 'Пароль змінено';

  @override
  String greeting(String name) {
    return 'Вітаємо, $name,';
  }

  @override
  String get latestValues => 'Останні значення';

  @override
  String get startMeasurement => 'РОЗПОЧАТИ ВИМІРЮВАННЯ';

  @override
  String get heartRateShort => 'Дихання';

  @override
  String get weightShort => 'Вага';

  @override
  String get recordWeightTitle => 'Записати вагу';

  @override
  String get weightKgLabel => 'Вага (кг)';

  @override
  String get weightCommentLabel => 'Коментар (необов\'язково)';

  @override
  String get weightCommentHint => 'Напр. після прогулянки, натщесерце…';

  @override
  String get weightSave => 'Зберегти';

  @override
  String get weightSentToVet => 'Вагу збережено';

  @override
  String get weightInvalid => 'Укажіть коректну вагу (0,01–999,99 кг)';

  @override
  String weightLastLabel(String kg) {
    return 'Остання вага: $kg кг';
  }

  @override
  String get choosePetForMeasurement => 'Виберіть тварину';

  @override
  String get chooseDuration => 'Тривалість вимірювання';

  @override
  String durationSeconds(int seconds) {
    return '$seconds с';
  }

  @override
  String get howToMeasure => 'Як вимірювати?';

  @override
  String get howToMeasureIntro =>
      'Вимірюйте частоту дихання вашої тварини у стані спокою.';

  @override
  String get howToMeasureStep1 =>
      '1. Заспокойте тварину, вона має лежати або сидіти.';

  @override
  String get howToMeasureStep2 =>
      '2. Покладіть руку на грудну клітку й натискайте з кожним вдихом протягом зазначеного часу.';

  @override
  String get howToMeasureStep3 =>
      '3. Підтвердьте вимірювання, щоб надіслати його вашому ветеринарові.';

  @override
  String get howToMeasureWhyTitle => 'Чому це важливо?';

  @override
  String get howToMeasureWhyBody =>
      'Регулярне спостереження за частотою дихання дає змогу виявити зміни та скоригувати лікування разом із вашим ветеринаром.';

  @override
  String get reminders => 'Нагадування';

  @override
  String get remindersHint =>
      'Отримуйте щоденне нагадування зробити вимірювання дихання.';

  @override
  String get remindersEnabled => 'Увімкнути нагадування';

  @override
  String get remindersTime => 'Час нагадування';

  @override
  String get remindersSaved => 'Нагадування збережено';

  @override
  String get legalTermsTitle => 'Загальні умови користування';

  @override
  String get legalPrivacyTitle => 'Політика конфіденційності';

  @override
  String get legalNoticeTitle => 'Юридична інформація';

  @override
  String get legalOpenOnline => 'Переглянути версію онлайн';

  @override
  String get legalTermsBody =>
      'Загальні умови користування — petsFollow\n\nЗастосунок petsFollow дає власникам тварин змогу вести призначене спостереження (повідомлення, нагадування Care/Horse, вимірювання дихання), переглядати історію та спілкуватися зі своїм ветеринаром.\n\nПослуги надаються в межах обраної підписки (оплата через Stripe). Користувач зобов\'язується використовувати застосунок за призначенням.\n\nПовна версія: https://petsfollow.ll-it-sc.be/legal/terms\n\nДата оновлення: липень 2026';

  @override
  String get legalPrivacyBody =>
      'Політика конфіденційності — petsFollow\n\nЗбирані дані: ідентифікаційні дані (ім\'я, email), дані тварини (ім\'я, вид, порода, фотографії), результати вимірювання частоти дихання (дані про здоров\'я тварини), повідомлення та медіафайли, якими ви обмінюєтеся з клінікою, звіти про візити (текст і аудіозаписи), GPS-координати домашніх візитів (фахівці з догляду), токени повідомлень (FCM), платіжні дані, які обробляє Stripe.\n\nЦілі: керування акаунтом, безперервність допомоги (зокрема вимірювання дихання), спілкування з ветеринаром, звіти про візити, повідомлення, виставлення рахунків.\n\nОбробка з ШІ: Google Gemini використовується для покращення звітів про візити (аудіо обробляється в реальному часі й не зберігається Google) і, у клієнтському застосунку (експериментальний модуль), для спрощеного викладу завершених звітів та розмовного визначення терміновості.\n\nСубпідрядники / партнери: Google (Sign-In, Firebase Cloud Messaging, Gemini), Stripe (платежі), хмарний хостинг (GCP).\n\nЗберігання: до видалення акаунта; неактивні акаунти видаляються через 3 роки; аудіо звітів зберігається протягом існування картки.\n\nПрава за GDPR (доступ, виправлення, видалення, перенесення): Профіль → Експортувати мої дані / Видалити акаунт, або зверніться на barbara@petsfollow.app.\n\nПовна версія: https://petsfollow.ll-it-sc.be/legal/privacy\n\nДата оновлення: липень 2026';

  @override
  String get legalNoticeBody =>
      'Юридична інформація — petsFollow\n\nВидавець: LL-IT-SC / petsFollow\nКонтакт: barbara@petsfollow.app · 0478.02.33.77\n\nХостинг: Google Cloud Platform (відповідність GDPR).\n\nВідповідальний за публікацію: petsFollow.\n\nПовна версія: https://petsfollow.ll-it-sc.be/legal/mentions\n\nДата оновлення: липень 2026';

  @override
  String get language => 'Мова';

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
  String get appearance => 'Вигляд';

  @override
  String get themeLight => 'Світла';

  @override
  String get themeDark => 'Темна';

  @override
  String get planMonthlyLabel => '3,50 € / місяць';

  @override
  String get planAnnualLabel => '35 € / рік';

  @override
  String get planTriennialLabel => '95 € / 3 роки';

  @override
  String get planQuinquennialLabel => '145 € / 5 років';

  @override
  String get pushNewMessage => 'Нове повідомлення';

  @override
  String get pushVisitConfirmed => 'Візит підтверджено';

  @override
  String get pushVisitProposed => 'Пропозиція візиту';

  @override
  String get pushVisitReschedule => 'Перенесення візиту';

  @override
  String get notifChannelMessages => 'Повідомлення';

  @override
  String get notifChannelVisits => 'Візити';

  @override
  String get notifChannelCare => 'Догляд';

  @override
  String get paymentResume => 'Продовжити оплату';

  @override
  String get manageSubscription => 'Керувати підпискою';

  @override
  String get heartRate => 'Вимірювання дихання';

  @override
  String get history => 'Історія';

  @override
  String get vetMessaging => 'Повідомлення ветеринару';

  @override
  String get badgeAutoRenew => 'Автопоновлення';

  @override
  String get badgeActive => 'Активно';

  @override
  String get badgePendingPayment => 'Очікує оплати';

  @override
  String badgeExpiresOn(String date) {
    return 'спливає $date';
  }

  @override
  String get newPet => 'Нова тварина';

  @override
  String get editPet => 'Змінити тварину';

  @override
  String get petName => 'Ім\'я';

  @override
  String get petNameRequired => 'Укажіть ім\'я тварини';

  @override
  String get species => 'Вид';

  @override
  String get breed => 'Порода';

  @override
  String get petMicrochipOptional => 'Номер мікрочипа (необов\'язково)';

  @override
  String get petHealthBookNumberOptional => 'Номер книжки (необов\'язково)';

  @override
  String get petDomicileLocation => 'Місце утримання / конюшня';

  @override
  String get petDomicileHint => 'Напр. Конюшня «Верби» — Брюссель';

  @override
  String get petFoodChainStatus => 'Статус щодо харчового ланцюга';

  @override
  String get petFoodChainCompanion => 'Домашня тварина (поза ланцюгом)';

  @override
  String get petFoodChainFoodProducing =>
      'Продуктивна тварина / харчовий ланцюг';

  @override
  String get petFoodChainExcluded => 'Виключена з харчового ланцюга';

  @override
  String get petHealthBookAddPages => 'Додати фотографії книжки';

  @override
  String get petHealthBookReplacePages => 'Замінити PDF (фотографії)';

  @override
  String petHealthBookPagesCount(int count) {
    return 'Вибрано сторінок: $count';
  }

  @override
  String get petHealthBookPdfAttached => 'PDF книжки прикріплено';

  @override
  String get petHealthBookRemovePdf => 'Видалити';

  @override
  String get petHealthBookOpenPdf => 'Відкрити книжку (PDF)';

  @override
  String petMicrochipLabel(String number) {
    return 'Мікрочип: $number';
  }

  @override
  String petHealthBookNumberLabel(String number) {
    return 'Книжка: $number';
  }

  @override
  String get errorHealthBookUploadFailed =>
      'Тварину збережено, але книжку не вдалося надіслати';

  @override
  String get choosePlan => 'Виберіть свій тариф';

  @override
  String get recommended => 'Рекомендовано';

  @override
  String get autoRenewTitle => 'Поновлювати автоматично';

  @override
  String get autoRenewSubtitle => 'Списання на кожну дату продовження';

  @override
  String get continueToPayment => 'Зберегти та оплатити';

  @override
  String get petFormSave => 'Зберегти';

  @override
  String get petSavedPendingPayment =>
      'Тварину збережено — активуйте її, щоб отримати доступ до функцій';

  @override
  String get paymentFeaturesLocked =>
      'Для використання функцій цієї тварини потрібна оплата';

  @override
  String get paymentConfirmed => 'Оплату підтверджено — тварина активна';

  @override
  String get paymentPending =>
      'Платіж в очікуванні — ви зможете продовжити пізніше';

  @override
  String errorGeneric(String message) {
    return 'Помилка: $message';
  }

  @override
  String get errorNetwork =>
      'Не вдалося підключитися. Перевірте мережу та спробуйте ще раз.';

  @override
  String get retryAction => 'Спробувати ще раз';

  @override
  String get errorMediaTooLarge => 'Файл завеликий (макс. 25 МБ)';

  @override
  String get errorInvalidMediaType =>
      'Формат не підтримується (JPEG, PNG, WebP, MP4, MOV, WebM)';

  @override
  String get errorPaymentRequired =>
      'Для використання цієї функції потрібна підписка';

  @override
  String get errorPhotoUploadFailed =>
      'Тварину створено, але фото не вдалося надіслати';

  @override
  String get errorCouldNotOpenLink => 'Не вдалося відкрити посилання';

  @override
  String get planMonthlySub => '3,50 € / місяць, поновлюється автоматично';

  @override
  String planAnnualSub(String price) {
    return '$price, поновлюється автоматично';
  }

  @override
  String get planTriennialSub => '95 € кожні 3 роки, поновлюється автоматично';

  @override
  String get planQuinquennialSub => '145 € за 5 років, одноразовий платіж';

  @override
  String planOneTime(String price) {
    return '$price, одноразовий платіж';
  }

  @override
  String get heartRateInstructions =>
      'Натискайте з кожним вдихом протягом часу, зазначеного вашим ветеринаром.';

  @override
  String heartRateInstructionsDuration(int seconds) {
    return 'Натискайте з кожним вдихом протягом $seconds секунд.';
  }

  @override
  String get heartRateNoDurationConfigured =>
      'Для цієї клініки не налаштовано тривалість вимірювання. Зверніться до свого ветеринара.';

  @override
  String get heartRateNotSupported =>
      'Вимірювання дихання недоступне для цього виду';

  @override
  String get start => 'Розпочати';

  @override
  String secondsLeft(int seconds) {
    return '$seconds с';
  }

  @override
  String beatsCount(int count) {
    return '$count вдихів';
  }

  @override
  String get tapHere => 'Натискайте тут з кожним вдихом';

  @override
  String bpmLabel(String bpm) {
    return 'Частота: $bpm';
  }

  @override
  String beatsLabel(int count) {
    return 'Вдихів: $count';
  }

  @override
  String get thresholdAlert =>
      'Сповіщення: значне зростання порівняно з попереднім вимірюванням';

  @override
  String get validateAndSend => 'Підтвердити й надіслати ветеринарові';

  @override
  String get heartRateCommentLabel => 'Коментар (необов\'язково)';

  @override
  String get heartRateCommentHint =>
      'Напр. неспокійна, у спокої, після навантаження…';

  @override
  String get restart => 'Почати знову';

  @override
  String get sentToVet => 'Вимірювання надіслано ветеринарові';

  @override
  String get navHome => 'Головна';

  @override
  String get navPets => 'Тварини';

  @override
  String get navCare => 'Догляд';

  @override
  String get navMessages => 'Повідомлення';

  @override
  String get navProfile => 'Профіль';

  @override
  String get speciesDog => 'Пес';

  @override
  String get speciesCat => 'Кіт';

  @override
  String get speciesHorse => 'Кінь';

  @override
  String get speciesDonkey => 'Осел';

  @override
  String get speciesCattle => 'Велика рогата худоба';

  @override
  String get speciesSheep => 'Вівця';

  @override
  String get speciesGoat => 'Коза';

  @override
  String get speciesPig => 'Свиня';

  @override
  String get speciesPoultry => 'Свійська птиця';

  @override
  String get speciesRabbit => 'Кролик';

  @override
  String get speciesAlpaca => 'Альпака';

  @override
  String get speciesLlama => 'Лама';

  @override
  String get speciesOther => 'Інше';

  @override
  String get careComingSoon =>
      'Нагадування про догляд будуть доступні невдовзі';

  @override
  String get emptyPetsTitle => 'Немає тварин';

  @override
  String get emptyPetsBody =>
      'Додайте свою першу тварину, щоб розпочати призначене спостереження разом із вашим ветеринаром.';

  @override
  String get discoveryTitle => 'Знайомство з petsFollow';

  @override
  String get homeRecentActivity => 'Остання активність';

  @override
  String get homePetTipsTitle => 'Поради для ваших тварин';

  @override
  String get homePetTipsDisclaimer =>
      'Загальні поради — ваш ветеринар залишається орієнтиром.';

  @override
  String get discoveryMission => 'Ваш шлях у petsFollow';

  @override
  String get discoveryDay0Title => 'Крок 1 — Вітаємо';

  @override
  String get discoveryDay0Body =>
      'Створіть профіль своєї тварини й ознайомтеся із застосунком — повідомлення, нагадування та вимірювання (зокрема частоти дихання).';

  @override
  String get discoveryDay2Title => 'Крок 2 — Перше вимірювання';

  @override
  String get discoveryDay2Body =>
      'Зробіть перше вимірювання дихання й освойте техніку.';

  @override
  String get discoveryDay4Title => 'Крок 3 — Звичка';

  @override
  String get discoveryDay4Body =>
      'Створіть звичку щоденних вимірювань за допомогою власних нагадувань.';

  @override
  String get discoveryDay6Title => 'Крок 4 — Обмін із ветеринаром';

  @override
  String get discoveryDay6Body =>
      'Ваші вимірювання доступні вашому ветеринарові для найкращого спостереження.';

  @override
  String get myVets => 'Мої ветеринари';

  @override
  String get addVetByEmail => 'Додати ветеринара за email';

  @override
  String get vetEmailHint => 'email@klinika.vet';

  @override
  String get noVets => 'Немає прив\'язаних ветеринарів';

  @override
  String get vetLinkRequired =>
      'Прив\'яжіть ветеринара, щоб активувати спостереження з вашою клінікою';

  @override
  String get linkVetAfterSaveTitle => 'Прив\'язати ветеринара';

  @override
  String get linkVetAfterSaveBody =>
      'Прив\'яжіть клініку, щоб активувати повідомлення, візити та нагадування про догляд.';

  @override
  String get linkVetHomeTitle => 'Прив\'язати ветеринара?';

  @override
  String get linkVetHomeBody =>
      'Вашу тварину зареєстровано. Бажаєте прив\'язати ветеринара? Це необов\'язково — ви зможете зробити це пізніше.';

  @override
  String get linkVetLater => 'Пізніше';

  @override
  String get primaryVet => 'Основний ветеринар';

  @override
  String get setPrimaryVet => 'Призначити основним ветеринаром';

  @override
  String get careTitle => 'Догляд';

  @override
  String get careDone => 'Виконано';

  @override
  String get carePostpone => 'Відкласти';

  @override
  String get careOverdue => 'Прострочено';

  @override
  String get visitHistory => 'Історія візитів';

  @override
  String get requestVisit => 'Запросити візит';

  @override
  String get calendarBookingDisabled =>
      'Онлайн-бронювання недоступне для цієї клініки. Зателефонуйте в клініку, щоб записатися.';

  @override
  String get calendarBookingDisabledReschedule =>
      'Онлайн-бронювання недоступне. Запропонуйте дату вручну.';

  @override
  String get calendarNoSlots => 'Немає вільних місць на наступні 14 днів.';

  @override
  String get calendarPickSlot => 'Виберіть час:';

  @override
  String get calendarSelectVet => 'Виберіть ветеринара:';

  @override
  String get calendarCallPractice => 'Зателефонувати в клініку';

  @override
  String get calendarNoPhone =>
      'Для цієї клініки не вказано номера телефону. Зв\'яжіться з нею іншим способом.';

  @override
  String get visitConfirm => 'Підтвердити';

  @override
  String get visitProposeReschedule => 'Запропонувати інший час';

  @override
  String get visitRescheduleProposed => 'Пропозицію перенесення надіслано';

  @override
  String get paymentSuccessSnack => 'Платіж отримано — оновлення…';

  @override
  String get paymentCancelSnack => 'Платіж скасовано';

  @override
  String get visitRejectReschedule => 'Відмовити в перенесенні';

  @override
  String get visitAcceptReschedule => 'Прийняти новий час';

  @override
  String get upcomingVisit => 'Майбутній візит';

  @override
  String get notificationHrTitle => 'petsFollow';

  @override
  String get notificationHrBody =>
      'Час зробити вимірювання дихання для вашої тварини';

  @override
  String get reviewAskTitle => 'Вам подобається petsFollow?';

  @override
  String get reviewAskYes => 'Так, оцінити застосунок';

  @override
  String get reviewAskNo => 'Пізніше';

  @override
  String get careTypeMedication => 'Медикамент';

  @override
  String get horseAddContact => 'Додати контакт';

  @override
  String get horseAddCompetition => 'Додати змагання';

  @override
  String get horseContactName => 'Ім\'я';

  @override
  String get horseContactRole => 'Роль';

  @override
  String get horseCompetitionTitle => 'Подія';

  @override
  String get horseCompetitionDate => 'Дата (РРРР-ММ-ДД)';

  @override
  String familyHouseholdTitle(int count) {
    return 'Домогосподарство Family — $count тварин';
  }

  @override
  String kennelHouseholdTitle(int count) {
    return 'Домогосподарство Kennel — $count тварин';
  }

  @override
  String get familyHouseholdNext => 'Наступні нагадування домогосподарства';

  @override
  String get familyPetLimit =>
      'Пакет для домогосподарства вже активний або в процесі придбання';

  @override
  String get familyRequiresTwoPets =>
      'Пакет Family потребує щонайменше 2 тварин';

  @override
  String get kennelPackHint =>
      'Пакет Kennel — ≥6 тварин, −15 % на наступні підписки';

  @override
  String get kennelRequiresSixPets =>
      'Пакет Kennel потребує щонайменше 6 тварин';

  @override
  String get kennelQuickEncodeTitle => 'Внесення посліду (розплідник)';

  @override
  String get kennelRequired => 'Для пакетного внесення потрібен пакет Kennel';

  @override
  String get litterTag => 'Тег посліду';

  @override
  String get petBirthDate => 'Дата народження';

  @override
  String get petBirthDateInvalid => 'Некоректна дата народження (РРРР-ММ-ДД)';

  @override
  String get discoveryMarkDone => 'Завдання виконано';

  @override
  String get discoveryDismiss => 'Приховати шлях';

  @override
  String get appBuildInfo => 'Версія застосунку';

  @override
  String get notificationPreferences => 'Налаштування повідомлень';

  @override
  String get notificationPrefsHint =>
      'Виберіть типи повідомлень, які бажаєте отримувати.';

  @override
  String get notificationPrefsSaved => 'Налаштування збережено';

  @override
  String get notificationPrefHr => 'Вимірювання дихання';

  @override
  String get notificationPrefCare => 'Нагадування про догляд';

  @override
  String get notificationPrefVisits => 'Візити';

  @override
  String get notificationPrefMessages => 'Повідомлення';

  @override
  String get notificationPrefDiscovery => 'Ознайомлювальний шлях';

  @override
  String get notificationPrefBilling => 'Виставлення рахунків';

  @override
  String get notificationPrefSms => 'SMS';

  @override
  String get notificationPrefSmsHint =>
      'Підтвердження та нагадування про візити через SMS. Відповідь STOP також вимикає цей канал.';

  @override
  String carePostponeDays(int days) {
    return 'Відкласти на $days днів';
  }

  @override
  String get noCareReminders => 'Немає активних нагадувань про догляд';

  @override
  String get careAddReminder => 'Додати нагадування';

  @override
  String get careSelectPet => 'Тварина';

  @override
  String careDueInDays(int days) {
    return 'Термін через $days днів';
  }

  @override
  String get careReferenceModeDone => 'Уже виконано';

  @override
  String get careReferenceModeFirst => 'Уперше';

  @override
  String get careLastDateLabel => 'Остання дата';

  @override
  String get careLastDateDone => 'Дата останнього догляду';

  @override
  String get careLastDateFirst => 'Початкова дата циклу';

  @override
  String get careRecurrenceLabel => 'Періодичність';

  @override
  String get careRecurrenceNone => 'Немає (один термін)';

  @override
  String careRecurrenceDays(int days) {
    return 'Кожні $days днів';
  }

  @override
  String get careDueDateLabel => 'Термін';

  @override
  String get careDueDateComputed => 'Розрахований термін';

  @override
  String get careTooltipDoneWithRecurrence =>
      'Догляд уже виконано: термін = дата останнього догляду + періодичність.';

  @override
  String get careTooltipFirstWithRecurrence =>
      'Перше планування: укажіть початкову дату циклу. Термін = ця дата + періодичність.';

  @override
  String get careTooltipNoRecurrence =>
      'Без періодичності: введена дата і є єдиним терміном.';

  @override
  String get careTooltipDueExplained =>
      'Термін = остання дата + періодичність (якщо задана).';

  @override
  String get carePickDate => 'Вибрати дату';

  @override
  String discoveryDayBadge(int day) {
    return 'К$day';
  }

  @override
  String get timelineTypeHeartrate => 'Частота дихання';

  @override
  String get timelineTypeWeight => 'Вага';

  @override
  String get timelineTypeMessage => 'Повідомлення';

  @override
  String get timelineTypeCare => 'Догляд';

  @override
  String get timelineTypeVisit => 'Візит';

  @override
  String get timelineTypeEvent => 'Подія';

  @override
  String get visitCancelAction => 'Скасувати запит';

  @override
  String get upcomingVisits => 'Наступні візити';

  @override
  String get timelineEmpty => 'Поки що немає подій';

  @override
  String get noThreads => 'Немає розмов';

  @override
  String get messageNoMessagesYet => 'Повідомлень ще немає';

  @override
  String get messageNewConversation => 'Нова розмова';

  @override
  String get messageComposeTitle => 'Нова розмова';

  @override
  String get messageChoosePro => 'Фахівець із догляду';

  @override
  String get messageChooseClient => 'Клієнт';

  @override
  String get messageChoosePet => 'Тварина, про яку йдеться';

  @override
  String get messageChoosePetOptional => 'Тварина (необов\'язково)';

  @override
  String get messageGeneralThread => 'Загальна розмова';

  @override
  String get messageStartConversation => 'Розпочати';

  @override
  String get messageLockedTitle => 'Повідомлення недоступні';

  @override
  String get messageLockedBody =>
      'Прив\'яжіть ветеринара, щоб спілкуватися з фахівцем із догляду.';

  @override
  String get vetInviteSent =>
      'Запрошення надіслано — клініка повинна прийняти запит';

  @override
  String vetInviteSentNamed(String practice) {
    return 'Запит надіслано до $practice — клініка повинна його прийняти';
  }

  @override
  String get vetNotFound =>
      'Ветеринара з такою електронною адресою не знайдено';

  @override
  String get addVetSearchHint =>
      'Шукайте за іменем, email або клінікою. Якщо ветеринар уже в petsFollow, запит на прив\'язку буде надіслано клініці.';

  @override
  String get addVetSearchLabel => 'Знайти ветеринара';

  @override
  String get addVetSearchFieldHint => 'Ім\'я, email або клініка';

  @override
  String get addVetNotListed => 'Мого ветеринара немає в списку';

  @override
  String get addVetSuggestTitle => 'Новий ветеринар';

  @override
  String get addVetSuggestBody =>
      'Укажіть email і телефон клініки. Ми звернемося до неї, щоб вона приєдналася до petsFollow.';

  @override
  String get addVetSuggestEmail => 'Email клініки';

  @override
  String get addVetSuggestPhone => 'Телефон';

  @override
  String get addVetSuggestNameOptional => 'Ім\'я ветеринара (необов\'язково)';

  @override
  String get addVetSuggestCta => 'Надіслати пропозицію';

  @override
  String get vetSuggestSent =>
      'Дякуємо — ми звернемося до клініки. Ви отримаєте повідомлення, коли вона приєднається до petsFollow.';

  @override
  String get visitRequested => 'Запит на візит надіслано';

  @override
  String get primaryVetSet => 'Основного ветеринара оновлено';

  @override
  String get visitStatusRequested => 'Запитано';

  @override
  String get visitStatusConfirmed => 'Підтверджено';

  @override
  String get visitStatusDone => 'Завершено';

  @override
  String get visitStatusCancelled => 'Скасовано';

  @override
  String get visitStatusReschedulePending => 'Перенесення в очікуванні';

  @override
  String get horseHealthTitle => 'Здоров\'я коней';

  @override
  String get horseContactsTitle => 'Контакти (коваль, стоматолог…)';

  @override
  String get horseCompetitionsTitle => 'Змагання';

  @override
  String get horseContactsSoon =>
      'Активуйте пакет Horse, щоб керувати професійними контактами.';

  @override
  String get horseCompetitionsSoon =>
      'Активуйте пакет Horse для календаря змагань.';

  @override
  String get horsePackUpsell =>
      'Пакет Horse — коваль, копроскопія, контакти та змагання';

  @override
  String get careTypeFarrier => 'Коваль';

  @override
  String get careTypeFecalEgg => 'Копроскопія';

  @override
  String get careTypeVaccination => 'Вакцинація';

  @override
  String get careTypeDeworming => 'Дегельмінтизація';

  @override
  String get careTypeVetCheck => 'Ветеринарний огляд';

  @override
  String get careTypeDental => 'Стоматологічний догляд';

  @override
  String get careTypeCustom => 'Власне нагадування';

  @override
  String get homeAddFirstVetTitle => 'Додайте свого ветеринара';

  @override
  String get homeAddFirstVetBody =>
      'Прив\'яжіть клініку, яка спостерігає за вашою твариною, щоб обмінюватися вимірюваннями та спілкуватися.';

  @override
  String get homeAddFirstVetCta => 'Додати ветеринара';

  @override
  String get photoFrameHint =>
      'Розташуйте морду в центрі — попередній перегляд картки тварини';

  @override
  String get takePhoto => 'Зробити фото';

  @override
  String get takeVideo => 'Зняти відео';

  @override
  String get chooseFromGallery => 'Вибрати з галереї';

  @override
  String get attachMedia => 'Прикріпити фото або відео';

  @override
  String get attachPhoto => 'Фото';

  @override
  String get attachVideo => 'Відео';

  @override
  String get compressingMedia => 'Стиснення відео…';

  @override
  String get openMedia => 'Відкрити';

  @override
  String get mediaVideoLabel => 'Відео';

  @override
  String get appInviteTitle => 'QR-запрошення до застосунку';

  @override
  String get appInviteHint =>
      'Покажіть цей QR або поділіться посиланням. Новий клієнт, який зареєструється за цим посиланням, прив\'язується автоматично.';

  @override
  String get appInviteHintClient =>
      'Поділіться цим QR із близькою людиною. Вона буде прив\'язана до вас (запрошення) і зможе приєднатися до вашої клініки, якщо ще не має своєї.';

  @override
  String get appInviteHintCommercial =>
      'Два посилання: клієнти (застосунок) і клініки (реєстрація Pro з вашим кодом запрошення).';

  @override
  String get appInviteHintShort => 'Посилання для завантаження та прив\'язки';

  @override
  String get appInviteCodeLabel => 'Код:';

  @override
  String get appInviteCopy => 'Копіювати посилання';

  @override
  String get appInviteCopied => 'Посилання скопійовано';

  @override
  String get appInviteLoadError => 'Не вдалося завантажити QR';

  @override
  String get appInviteRetry => 'Спробувати ще раз';

  @override
  String get proLightVetTitle => 'Ветеринар на виїзді';

  @override
  String get commercialFieldTitle => 'Менеджер з продажу';

  @override
  String get commercialFieldSubtitle =>
      'QR-запрошення для клієнтів і доступ до сайту Pro.';

  @override
  String get commercialManagerFieldSubtitle =>
      'Результати команди, QR-запрошення та доступ до сайту Pro.';

  @override
  String get commercialOpenProWeb => 'Відкрити сайт Pro';

  @override
  String get managerTeamCta => 'Результати моєї команди';

  @override
  String get managerTeamTitle => 'Команда';

  @override
  String get managerTeamSection => 'Результати команди';

  @override
  String get managerSelfSection => 'Мої результати';

  @override
  String get managerMembersSection => 'Менеджери з продажу';

  @override
  String get managerTeamEmpty => 'Немає прикріплених менеджерів з продажу.';

  @override
  String get managerKpiProspects => 'Потенційні клієнти';

  @override
  String get managerKpiConverted => 'Конвертовані';

  @override
  String get managerKpiConversion => 'Коефіцієнт конверсії';

  @override
  String get managerKpiAppointments => 'Майбутні зустрічі';

  @override
  String get managerKpiStale => 'Прострочений пайплайн';

  @override
  String get managerKpiMonthEarned => 'Винагороди (місяць)';

  @override
  String get managerKpiLifetimeEarned => 'Винагороди (усього)';

  @override
  String get managerKpiVets => 'Призначені ветеринари';

  @override
  String get featureModules => 'Опції';

  @override
  String get featureModulesCatalog => 'Ознайомитися з опціями';

  @override
  String get featureModulesSubtitle =>
      'Активуйте Care+, Horse, Kennel або «Домогосподарство» відповідно до ваших потреб.';

  @override
  String get moduleCarePlus => 'Care+';

  @override
  String get moduleCarePlusDesc =>
      'Розширені нагадування про догляд для всіх ваших тварин.';

  @override
  String get moduleHorse => 'Horse';

  @override
  String get moduleHorseDesc => 'Професійні контакти та змагання для коней.';

  @override
  String get moduleKennel => 'Kennel';

  @override
  String get moduleKennelDesc => 'Швидке внесення розплідника / посліду.';

  @override
  String get moduleFamily => 'Домогосподарство';

  @override
  String get moduleFamilyDesc => 'Огляд домогосподарства ваших тварин.';

  @override
  String get moduleActivate => 'Активувати';

  @override
  String get moduleActive => 'Активно';

  @override
  String get switchProfile => 'Змінити профіль';

  @override
  String get profilePersonal => 'Особистий';

  @override
  String get profilePro => 'Професійний';

  @override
  String get profileSwitched => 'Профіль змінено';

  @override
  String get preconsultTitle => 'Передконсультація';

  @override
  String preconsultTitlePet(String petName) {
    return 'Передконсультація — $petName';
  }

  @override
  String get preconsultIntro =>
      'Опишіть стан вашої тварини перед візитом. Ваш ветеринар зможе підготуватися.';

  @override
  String get preconsultComplaint => 'Причина / основна скарга';

  @override
  String get preconsultComplaintRequired => 'Укажіть причину візиту';

  @override
  String get preconsultDuration => 'Як довго?';

  @override
  String get preconsultDurationToday => 'Сьогодні';

  @override
  String get preconsultDurationFewDays => 'Кілька днів';

  @override
  String get preconsultDurationWeek => 'Близько тижня';

  @override
  String get preconsultDurationWeeks => 'Кілька тижнів';

  @override
  String get preconsultDurationMonths => 'Кілька місяців';

  @override
  String get preconsultBehavior => 'Поведінка';

  @override
  String get preconsultBehaviorNormal => 'Звичайна';

  @override
  String get preconsultBehaviorLethargic => 'Апатична';

  @override
  String get preconsultBehaviorRestless => 'Неспокійна';

  @override
  String get preconsultBehaviorAggressive => 'Агресивна';

  @override
  String get preconsultBehaviorAnxious => 'Тривожна';

  @override
  String get preconsultBehaviorOther => 'Інша';

  @override
  String get preconsultAppetite => 'Апетит';

  @override
  String get preconsultThirst => 'Спрага';

  @override
  String get preconsultElimination => 'Випорожнення / сечовиділення';

  @override
  String get preconsultScaleNormal => 'У нормі';

  @override
  String get preconsultScaleDecreased => 'Знижений';

  @override
  String get preconsultScaleIncreased => 'Підвищений';

  @override
  String get preconsultUrgency => 'Відчутна терміновість';

  @override
  String get preconsultUrgencyLow => 'Низька';

  @override
  String get preconsultUrgencyMedium => 'Середня';

  @override
  String get preconsultUrgencyHigh => 'Висока';

  @override
  String get preconsultComment => 'Коментар (необов\'язково)';

  @override
  String get preconsultUnknown => 'Не знаю';

  @override
  String get preconsultSubmit => 'Надіслати';

  @override
  String get preconsultSubmitted => 'Передконсультацію надіслано';

  @override
  String get preconsultAlreadySubmitted =>
      'Ви вже надіслали цю передконсультацію.';

  @override
  String get preconsultFillCta => 'Заповнити передконсультацію';

  @override
  String get proLightAudioConsentClientCheck =>
      'Я отримав усну згоду клієнта на запис';

  @override
  String get proLightReportAiProposalBanner =>
      'Пропозиція ШІ — обов\'язкове підтвердження перед завершенням (включно з діагнозом / лікуванням).';

  @override
  String get proLightAiModuleRequired =>
      'Функція ШІ-звітів вимкнена для цієї клініки — зверніться до підтримки petsFollow.';

  @override
  String proLightAiModuleTrialBanner(int days) {
    return 'Пробний період ШІ-звітів — залишилося $days дн. Диктуйте, потім покращуйте свої звіти.';
  }

  @override
  String get proLightAiModuleInactiveBanner =>
      'ШІ-звіти не активовано для цієї клініки. Диктування з ШІ буде недоступне.';

  @override
  String get proLightAiModuleVisitScopedBanner =>
      'ШІ-звіти доступні, якщо в клініці, де відбувається візит, увімкнено модуль.';

  @override
  String get supportTitle => 'Повідомити про проблему';

  @override
  String get supportHint =>
      'Опишіть помилку. Технічну діагностику за останні 15 хвилин буде додано автоматично.';

  @override
  String get supportSubject => 'Тема';

  @override
  String get supportMessage => 'Опис';

  @override
  String get supportDiagnosticsAttached =>
      'Діагностику додано автоматично (помилки, запити, конфігурація).';

  @override
  String get supportSubmit => 'Надіслати';

  @override
  String get supportSending => 'Надсилання…';

  @override
  String get supportSuccess => 'Повідомлення надіслано. Дякуємо!';

  @override
  String get supportErrorTooLarge =>
      'Діагностика завелика. Перезапустіть застосунок і спробуйте ще раз.';

  @override
  String get supportMenu => 'Підтримка';

  @override
  String get appInviteHintSales =>
      'Поділіться своїм кодом запрошення з клінікою (реєстрація Pro) або з клієнтом (запрошення до застосунку).';

  @override
  String get appInviteCopyVet => 'Копіювати посилання для реєстрації клініки';

  @override
  String get appInviteCopyClient =>
      'Копіювати посилання-запрошення для клієнта';

  @override
  String get sendDossierToPro => 'Надіслати фахівцю';

  @override
  String get sendDossierEmailLabel => 'Email фахівця';

  @override
  String get sendDossierEmailHint => 'vet@klinika.be';

  @override
  String get sendDossierConfirm => 'Надіслати';

  @override
  String get sendDossierSuccess =>
      'Картку надіслано — посилання дійсне 24 год.';

  @override
  String get sendDossierInvalidEmail => 'Некоректна електронна адреса.';

  @override
  String get sendDossierPhiWarning =>
      'Ця картка містить дані про здоров\'я: звіти про візити, медичну книжку та документи. Посилання дійсне 24 год, і будь-хто, хто його має, зможе їх переглянути.';

  @override
  String get sendDossierPhiConsent =>
      'Я погоджуюся поділитися цими даними про здоров\'я з цим фахівцем.';

  @override
  String get consultationsHistory => 'Консультації';

  @override
  String get consultationTitle => 'Консультація';

  @override
  String consultationTitleWithPet(String petName) {
    return 'Консультація — $petName';
  }

  @override
  String get consultationVisitMeta => 'Візит';

  @override
  String consultationReportBy(String author) {
    return 'Звіт від $author';
  }

  @override
  String get consultationReportFallback => 'Звіт';

  @override
  String get consultationReportEmpty => '(порожньо)';

  @override
  String get consultationAvailableCta => 'Доступно';

  @override
  String get consultationPendingCta => 'Чернетка';

  @override
  String get sendConsultationToVet => 'Надіслати ветеринарові';

  @override
  String get sendConsultationEmailLabel => 'Email ветеринара';

  @override
  String get sendConsultationEmailHint => 'vet@klinika.be';

  @override
  String get sendConsultationConfirm => 'Надіслати';

  @override
  String get sendConsultationSuccess =>
      'Консультацію надіслано — посилання дійсне 24 год.';

  @override
  String get sendConsultationInvalidEmail => 'Некоректна електронна адреса.';

  @override
  String get sendConsultationPhiWarning =>
      'Цей звіт містить дані про здоров\'я. Посилання дійсне 24 год, і будь-хто, хто його має, зможе завантажити PDF.';

  @override
  String get sendConsultationPhiConsent =>
      'Я погоджуюся поділитися цим звітом із цим ветеринаром.';

  @override
  String get clientAiDevBadge => 'dev';

  @override
  String get clientAiSectionTitle => 'Допомога ШІ';

  @override
  String get clientAiExplainCta => 'Зрозуміти мій звіт';

  @override
  String get clientAiExplainTitle => 'Ваш звіт простими словами';

  @override
  String get clientAiExplainDisclaimer =>
      'Це не медичний висновок. Завжди дотримуйтеся вказівок вашого ветеринара.';

  @override
  String get clientAiExplainLoading => 'Підготовка пояснення…';

  @override
  String get clientAiExplainListTitle => 'Зрозуміти звіт';

  @override
  String get clientAiExplainListSubtitle =>
      'Спрощене пояснення завершеного звіту';

  @override
  String get clientAiExplainListEmpty =>
      'Поки що немає завершених звітів для пояснення.';

  @override
  String get clientAiTriageTitle => 'Екстрена допомога 24/7';

  @override
  String get clientAiTriageSubtitle =>
      'Опишіть ситуацію — ми оцінимо ступінь терміновості.';

  @override
  String get clientAiTriageHint => 'Напр.: мій пес з\'їв шоколад…';

  @override
  String get clientAiTriageSend => 'Надіслати';

  @override
  String get clientAiTriageLevelGreen => 'Порада';

  @override
  String get clientAiTriageLevelOrange => 'Записатися на візит';

  @override
  String get clientAiTriageLevelRed => 'Екстрений випадок';

  @override
  String get clientAiTriageWatchSigns => 'Ознаки, за якими слід спостерігати';

  @override
  String get clientAiTriageCallPractice => 'Зателефонувати в клініку';

  @override
  String get clientAiTriageBookVisit => 'Записатися на візит';

  @override
  String get clientAiTriageMessageVet => 'Звернутися до мого ветеринара';

  @override
  String get clientAiTriageSelectPet => 'Для якої тварини?';

  @override
  String get clientAiTriageNoPet => 'Продовжити без тварини';

  @override
  String get clientAiTriageStart => 'Розпочати';

  @override
  String get clientAiTriageEmergencyFallback =>
      'Якщо ви не можете зв\'язатися зі своєю клінікою, негайно зверніться до місцевої ветеринарної служби невідкладної допомоги.';

  @override
  String get clientAiTriageOpenMessages => 'Відкрити повідомлення';

  @override
  String get bloodPressureShort => 'Тиск';

  @override
  String get recordBloodPressureTitle => 'Записати тиск';

  @override
  String get bpSystolicLabel => 'Систолічний (мм рт. ст.)';

  @override
  String get bpDiastolicLabel => 'Діастолічний (мм рт. ст.)';

  @override
  String get bpMethodLabel => 'Метод';

  @override
  String get bpMethodDoppler => 'Доплер';

  @override
  String get bpMethodOscillometric => 'Осцилометричний';

  @override
  String get bpMethodUnknown => 'Не вказано';

  @override
  String get bloodPressureInvalid => 'Укажіть коректний тиск (СИС ≥ ДІА)';

  @override
  String get bloodPressureSaved => 'Тиск збережено';

  @override
  String get labsTitle => 'Аналізи';

  @override
  String get labsEmpty => 'Немає доступних аналізів';

  @override
  String get labsResults => 'Результати';

  @override
  String labsAbnormalCount(int count) {
    return '$count поза нормою';
  }

  @override
  String get labsOpen => 'Переглянути аналізи';

  @override
  String get labsFlagLow => 'Низький';

  @override
  String get labsFlagHigh => 'Високий';

  @override
  String get labsFlagNormal => 'У нормі';

  @override
  String get labsOpenDocument => 'Відкрити документ';

  @override
  String get bpMethodInvasive => 'Інвазивний';

  @override
  String get bpSiteLabel => 'Місце (необов\'язково)';

  @override
  String get bpCommentLabel => 'Коментар (необов\'язково)';

  @override
  String get bloodPressureSave => 'Зберегти';

  @override
  String get calendarSelectSite => 'Оберіть місце';

  @override
  String get proformaValidateCta => 'Підтвердити проформу';
}
