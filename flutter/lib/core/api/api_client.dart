import 'package:dio/dio.dart';
import 'package:petsfollow_mobile/core/discovery/discovery_controller.dart';
import 'package:petsfollow_mobile/core/locale/locale_controller.dart';
import 'package:petsfollow_mobile/core/models/care_reminder.dart';
import 'package:petsfollow_mobile/core/models/discovery_progress.dart';
import 'package:petsfollow_mobile/core/models/message_thread.dart';
import 'package:petsfollow_mobile/core/models/notification_prefs.dart';
import 'package:petsfollow_mobile/core/models/vet_link.dart';
import 'package:petsfollow_mobile/core/models/visit.dart';
import 'package:petsfollow_mobile/core/notifications/notification_service.dart';
import 'package:shared_preferences/shared_preferences.dart';

class ApiClient {
  ApiClient._();
  static final instance = ApiClient._();

  static const _tokenKey = 'pf_token';

  String? token;
  final dio = Dio(BaseOptions(
    baseUrl: const String.fromEnvironment('API_BASE', defaultValue: 'http://10.0.2.2:8291'),
    headers: {'Content-Type': 'application/json'},
  ));

  void loadToken() {
    dio.interceptors.clear();
    dio.interceptors.add(InterceptorsWrapper(
      onRequest: (options, handler) {
        if (token != null) {
          options.headers['Authorization'] = 'Bearer $token';
        }
        options.headers['Accept-Language'] = LocaleController.instance.languageCode;
        if (options.data is FormData) {
          options.headers.remove(Headers.contentTypeHeader);
          options.contentType = null;
        }
        handler.next(options);
      },
    ));
  }

  Future<void> restoreSession() async {
    final sp = await SharedPreferences.getInstance();
    token = sp.getString(_tokenKey);
    loadToken();
  }

  Future<void> _persistToken(String? value) async {
    final sp = await SharedPreferences.getInstance();
    if (value == null) {
      await sp.remove(_tokenKey);
    } else {
      await sp.setString(_tokenKey, value);
    }
  }

  Future<void> logout() async {
    token = null;
    await _persistToken(null);
    loadToken();
    NotificationService.instance.resetSession();
    await DiscoveryController.instance.clearLocal();
  }

  Future<Map<String, dynamic>> login(String email, String password) async {
    final res = await dio.post('/api/v1/auth/login', data: {
      'email': email,
      'password': password,
    });
    final data = res.data['data'] as Map<String, dynamic>;
    token = data['accessToken'] as String?;
    await _persistToken(token);
    loadToken();
    await syncLocaleFromMe();
    try {
      final me = await getMe();
      DiscoveryController.instance.bindUser(me['userId'] as String? ?? me['id'] as String?);
    } catch (_) {}
    await NotificationService.instance.onLogin();
    return data;
  }

  Future<Map<String, dynamic>> getMe() async {
    final res = await dio.get('/api/v1/me');
    return res.data['data'] as Map<String, dynamic>;
  }

  Future<Map<String, dynamic>> updateMe(String fullName) async {
    final res = await dio.patch('/api/v1/me', data: {'fullName': fullName});
    return res.data['data'] as Map<String, dynamic>;
  }

  Future<Map<String, dynamic>> uploadAvatar(String filePath) async {
    final form = FormData.fromMap({
      'file': await MultipartFile.fromFile(filePath, filename: 'avatar.jpg'),
    });
    final res = await dio.post('/api/v1/me/avatar', data: form);
    return res.data['data'] as Map<String, dynamic>;
  }

  Future<Map<String, dynamic>> uploadPetPhoto(String petId, String filePath) async {
    final form = FormData.fromMap({
      'file': await MultipartFile.fromFile(filePath, filename: 'photo.jpg'),
    });
    final res = await dio.post('/api/v1/pets/$petId/photo', data: form);
    return res.data['data'] as Map<String, dynamic>;
  }

  Future<void> changePassword(String currentPassword, String newPassword) async {
    await dio.patch('/api/v1/me/password', data: {
      'currentPassword': currentPassword,
      'newPassword': newPassword,
    });
  }

  Future<void> deleteAccount() async {
    await dio.delete('/api/v1/me');
    await logout();
  }

  Future<void> updateLocale(String locale) async {
    await dio.patch('/api/v1/me/locale', data: {'locale': locale});
    await LocaleController.instance.setLocale(locale);
  }

  Future<void> syncLocaleFromMe() async {
    try {
      final me = await getMe();
      final locale = me['preferredLocale'] as String?;
      if (locale != null) {
        await LocaleController.instance.setLocale(locale);
      }
    } catch (_) {
      /* ignore if /me unavailable */
    }
  }

  Future<List<dynamic>> getPets() async {
    final res = await dio.get('/api/v1/pets');
    return res.data['data'] as List<dynamic>;
  }

  Future<Map<String, dynamic>> createPet(Map<String, dynamic> body) async {
    final res = await dio.post('/api/v1/pets', data: body);
    return res.data['data'] as Map<String, dynamic>;
  }

  Future<Map<String, dynamic>> getPet(String petId) async {
    final res = await dio.get('/api/v1/pets/$petId');
    return res.data['data'] as Map<String, dynamic>;
  }

  Future<String> resumeCheckout(String petId) async {
    final res = await dio.post('/api/v1/pets/$petId/billing/checkout');
    return res.data['data']['checkoutUrl'] as String;
  }

  Future<String> billingPortal(String petId) async {
    final res = await dio.post('/api/v1/pets/$petId/billing/portal', data: {});
    return res.data['data']['url'] as String;
  }

  Future<List<dynamic>> getBillingPlans() async {
    final res = await dio.get('/api/v1/billing/plans');
    return res.data['data']['plans'] as List<dynamic>;
  }

  Future<List<dynamic>> getBillingAddons() async {
    final res = await dio.get('/api/v1/billing/addons');
    return res.data['data']['addons'] as List<dynamic>;
  }

  Future<List<dynamic>> getMyAddons() async {
    final res = await dio.get('/api/v1/billing/my-addons');
    final data = res.data['data'];
    if (data is List) return data;
    if (data is Map && data['addons'] is List) return data['addons'] as List<dynamic>;
    return const [];
  }

  Future<String> startAddonCheckout({required String addonCode, String? petId}) async {
    final res = await dio.post('/api/v1/billing/addons/checkout', data: {
      'addonCode': addonCode,
      if (petId != null && petId.isNotEmpty) 'petId': petId,
    });
    return res.data['data']['checkoutUrl'] as String;
  }

  /// Family privilege: household digest (requires active Family addon).
  Future<Map<String, dynamic>> getHousehold() async {
    final res = await dio.get('/api/v1/me/household');
    return Map<String, dynamic>.from(res.data['data'] as Map);
  }

  Future<List<dynamic>> getHorseContacts(String petId) async {
    final res = await dio.get('/api/v1/pets/$petId/horse-contacts');
    return res.data['data'] as List<dynamic>;
  }

  Future<Map<String, dynamic>> createHorseContact(
    String petId, {
    required String fullName,
    String role = '',
    String phone = '',
    String email = '',
    String notes = '',
  }) async {
    final res = await dio.post('/api/v1/pets/$petId/horse-contacts', data: {
      'fullName': fullName,
      'role': role,
      'phone': phone,
      'email': email,
      'notes': notes,
    });
    return Map<String, dynamic>.from(res.data['data'] as Map);
  }

  Future<void> deleteHorseContact(String id) async {
    await dio.delete('/api/v1/horse-contacts/$id');
  }

  Future<List<dynamic>> getHorseCompetitions(String petId) async {
    final res = await dio.get('/api/v1/pets/$petId/horse-competitions');
    return res.data['data'] as List<dynamic>;
  }

  Future<Map<String, dynamic>> createHorseCompetition(
    String petId, {
    required String title,
    required String eventDate,
    String location = '',
    String discipline = '',
    String result = '',
    String notes = '',
  }) async {
    final res = await dio.post('/api/v1/pets/$petId/horse-competitions', data: {
      'title': title,
      'eventDate': eventDate,
      'location': location,
      'discipline': discipline,
      'result': result,
      'notes': notes,
    });
    return Map<String, dynamic>.from(res.data['data'] as Map);
  }

  Future<void> deleteHorseCompetition(String id) async {
    await dio.delete('/api/v1/horse-competitions/$id');
  }

  Future<Map<String, dynamic>> startHeartRate(String petId, {int? durationSec}) async {
    final res = await dio.post(
      '/api/v1/pets/$petId/heartrate/sessions',
      data: durationSec != null ? {'durationSec': durationSec} : null,
    );
    return res.data['data'] as Map<String, dynamic>;
  }

  Future<List<dynamic>> getHeartRateSessions(String petId) async {
    final res = await dio.get('/api/v1/pets/$petId/heartrate/sessions');
    return res.data['data'] as List<dynamic>;
  }

  Future<Map<String, dynamic>> completeHeartRate(String sessionId, int tapCount) async {
    final res = await dio.patch('/api/v1/heartrate/sessions/$sessionId', data: {'tapCount': tapCount});
    return res.data['data'] as Map<String, dynamic>;
  }

  Future<Map<String, dynamic>> validateHeartRate(String sessionId) async {
    final res = await dio.post('/api/v1/heartrate/sessions/$sessionId/validate');
    return res.data['data'] as Map<String, dynamic>;
  }

  Future<void> cancelHeartRate(String sessionId) async {
    await dio.post('/api/v1/heartrate/sessions/$sessionId/cancel');
  }

  Future<List<dynamic>> getTimeline(String petId) async {
    final res = await dio.get('/api/v1/pets/$petId/timeline');
    return res.data['data'] as List<dynamic>;
  }

  Future<List<dynamic>> getThreads() async {
    final res = await dio.get('/api/v1/messaging/threads');
    return res.data['data'] as List<dynamic>;
  }

  Future<List<dynamic>> getMessages(String threadId) async {
    final res = await dio.get('/api/v1/messaging/threads/$threadId/messages');
    return res.data['data'] as List<dynamic>;
  }

  Future<void> sendMessage(String threadId, String body) async {
    await dio.post('/api/v1/messaging/threads/$threadId/messages', data: {'body': body});
  }

  Future<void> sendMessageMedia(String threadId, String filePath, {String? body, String? filename}) async {
    final form = FormData.fromMap({
      if (body != null && body.trim().isNotEmpty) 'body': body.trim(),
      'file': await MultipartFile.fromFile(
        filePath,
        filename: filename ?? filePath.split('/').last,
      ),
    });
    await dio.post(
      '/api/v1/messaging/threads/$threadId/messages/media',
      data: form,
      options: Options(sendTimeout: const Duration(minutes: 2), receiveTimeout: const Duration(minutes: 2)),
    );
  }

  // --- Client enrichment ---

  Future<List<VetLink>> getMyVets({String? primaryPracticeId}) async {
    final res = await dio.get('/api/v1/me/vets');
    final data = res.data['data'] as List<dynamic>;
    return data
        .map((v) => VetLink.fromJson(
              Map<String, dynamic>.from(v as Map),
              primaryPracticeId: primaryPracticeId,
            ))
        .toList();
  }

  Future<void> inviteVet(String email) async {
    await dio.post('/api/v1/me/vets/invite', data: {'email': email});
  }

  Future<void> setPetPrimaryPractice(String petId, String practiceId) async {
    await dio.patch('/api/v1/pets/$petId/primary-practice', data: {'practiceId': practiceId});
  }

  Future<List<CareReminder>> getCareReminders(String petId) async {
    final res = await dio.get('/api/v1/pets/$petId/care-reminders');
    final data = res.data['data'] as List<dynamic>;
    return data.map((r) => CareReminder.fromJson(Map<String, dynamic>.from(r as Map))).toList();
  }

  Future<CareReminder> createCareReminder(
    String petId, {
    String? title,
    String? type,
    int? dueDays,
    String? notes,
    int? recurrenceDays,
  }) async {
    final res = await dio.post('/api/v1/pets/$petId/care-reminders', data: {
      if (title != null && title.trim().isNotEmpty) 'title': title.trim(),
      if (type != null && type.isNotEmpty) 'type': type,
      if (dueDays != null) 'dueDays': dueDays,
      if (notes != null && notes.trim().isNotEmpty) 'notes': notes.trim(),
      if (recurrenceDays != null) 'recurrenceDays': recurrenceDays,
    });
    return CareReminder.fromJson(res.data['data'] as Map<String, dynamic>);
  }

  Future<CareReminder> markCareReminderDone(String id) async {
    final res = await dio.post('/api/v1/care-reminders/$id/done');
    return CareReminder.fromJson(res.data['data'] as Map<String, dynamic>);
  }

  Future<CareReminder> postponeCareReminder(String id, int days) async {
    final res = await dio.post('/api/v1/care-reminders/$id/postpone', data: {'days': days});
    return CareReminder.fromJson(res.data['data'] as Map<String, dynamic>);
  }

  Future<List<Visit>> getVisits(String petId) async {
    final res = await dio.get('/api/v1/pets/$petId/visits');
    final data = res.data['data'] as List<dynamic>;
    return data.map((v) => Visit.fromJson(Map<String, dynamic>.from(v as Map))).toList();
  }

  Future<Visit> createVisit(String petId, {String? notes}) async {
    final res = await dio.post('/api/v1/pets/$petId/visits', data: {
      if (notes != null) 'notes': notes,
    });
    return Visit.fromJson(res.data['data'] as Map<String, dynamic>);
  }

  Future<Visit> updateVisit(String id, String status) async {
    final res = await dio.patch('/api/v1/visits/$id', data: {'status': status});
    return Visit.fromJson(res.data['data'] as Map<String, dynamic>);
  }

  Future<DiscoveryProgress> getDiscovery() async {
    final res = await dio.get('/api/v1/me/discovery');
    return DiscoveryProgress.fromJson(res.data['data'] as Map<String, dynamic>);
  }

  Future<DiscoveryProgress> completeDiscoveryCard(String cardKey) async {
    final res = await dio.post('/api/v1/me/discovery/complete', data: {'cardKey': cardKey});
    return DiscoveryProgress.fromJson(res.data['data'] as Map<String, dynamic>);
  }

  Future<void> putDeviceToken(String token, String platform) async {
    await dio.put('/api/v1/me/device-tokens', data: {'token': token, 'platform': platform});
  }

  Future<NotificationPrefs> getNotificationPrefs() async {
    final res = await dio.get('/api/v1/me/notification-preferences');
    return NotificationPrefs.fromJson(res.data['data'] as Map<String, dynamic>);
  }

  Future<NotificationPrefs> updateNotificationPrefs(NotificationPrefs prefs) async {
    final res = await dio.patch('/api/v1/me/notification-preferences', data: prefs.toJson());
    return NotificationPrefs.fromJson(res.data['data'] as Map<String, dynamic>);
  }

  Future<List<MessageThread>> getMessageThreads() async {
    final res = await dio.get('/api/v1/messaging/threads');
    final data = res.data['data'] as List<dynamic>;
    return data.map((t) => MessageThread.fromJson(Map<String, dynamic>.from(t as Map))).toList();
  }

  Future<List<ChatMessage>> getChatMessages(String threadId) async {
    final res = await dio.get('/api/v1/messaging/threads/$threadId/messages');
    final data = res.data['data'] as List<dynamic>;
    return data.map((m) => ChatMessage.fromJson(Map<String, dynamic>.from(m as Map))).toList();
  }
}
