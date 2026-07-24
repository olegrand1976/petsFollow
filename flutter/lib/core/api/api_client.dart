import 'dart:async';

import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';
import 'package:petsfollow_mobile/core/auth/google_auth.dart';
import 'package:petsfollow_mobile/core/discovery/discovery_controller.dart';
import 'package:petsfollow_mobile/core/invite/invite_code_store.dart';
import 'package:petsfollow_mobile/core/locale/locale_controller.dart';
import 'package:petsfollow_mobile/core/models/care_reminder.dart';
import 'package:petsfollow_mobile/core/models/discovery_progress.dart';
import 'package:petsfollow_mobile/core/models/message_thread.dart';
import 'package:petsfollow_mobile/core/models/notification_prefs.dart';
import 'package:petsfollow_mobile/core/models/practice_availability.dart';
import 'package:petsfollow_mobile/core/models/vet_link.dart';
import 'package:petsfollow_mobile/core/models/visit.dart';
import 'package:petsfollow_mobile/core/notifications/notification_service.dart';
import 'package:shared_preferences/shared_preferences.dart';

class ApiClient {
  ApiClient._() {
    if (kReleaseMode && !_apiBase.startsWith('https://')) {
      throw StateError(
        'API_BASE must be an https:// URL in release builds (got: "$_apiBase"). '
        'Pass --dart-define=API_BASE=https://…',
      );
    }
  }
  static final instance = ApiClient._();

  static const _tokenKey = 'pf_token';
  static const _apiBaseDefined = String.fromEnvironment('API_BASE');

  /// Platform-aware local default (Android emu vs iOS sim). Override with `--dart-define=API_BASE=…`.
  static String get _apiBase {
    if (_apiBaseDefined.isNotEmpty) return _apiBaseDefined;
    if (defaultTargetPlatform == TargetPlatform.android) {
      return 'http://10.0.2.2:8291';
    }
    return 'http://localhost:8291';
  }

  String? token;

  /// Fired after a 401 clears the local session (AuthGate rebuilds → login).
  void Function()? onSessionInvalidated;

  bool _handlingUnauthorized = false;
  int _authGeneration = 0;

  late final dio = Dio(BaseOptions(
    baseUrl: _apiBase,
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
      onError: (error, handler) {
        if (error.response?.statusCode == 401 &&
            token != null &&
            !_isPublicAuthPath(error.requestOptions.path)) {
          // Clear token sync so AuthGate sees token == null immediately.
          unawaited(_invalidateSessionFromUnauthorized());
        }
        handler.next(error);
      },
    ));
  }

  static bool _isPublicAuthPath(String path) {
    return path.contains('/api/v1/auth/');
  }

  /// Clears session after an authenticated 401. Safe to call multiple times.
  Future<void> _invalidateSessionFromUnauthorized() async {
    if (_handlingUnauthorized || token == null) return;
    _handlingUnauthorized = true;
    final generation = ++_authGeneration;
    token = null;
    userId = null;
    userRole = null;
    userSpecialty = null;
    onSessionInvalidated?.call();
    try {
      final sp = await SharedPreferences.getInstance();
      // Abort if a new login started after we cleared.
      if (_authGeneration != generation || token != null) return;
      await sp.remove(_tokenKey);
      if (_authGeneration != generation || token != null) {
        if (token != null) await _persistToken(token);
        return;
      }
      NotificationService.instance.resetSession();
      await DiscoveryController.instance.clearLocal();
      if (_authGeneration != generation || token != null) return;
      await GoogleAuth.signOut();
    } finally {
      _handlingUnauthorized = false;
    }
  }

  Future<void> restoreSession() async {
    final sp = await SharedPreferences.getInstance();
    token = sp.getString(_tokenKey);
    loadToken();
    if (token != null) {
      try {
        final me = await getMe();
        userId = me['userId'] as String? ?? me['id'] as String?;
        userRole = me['role'] as String?;
        userSpecialty = me['professionalSpecialty'] as String?;
      } catch (_) {
        await logout();
      }
    }
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
    userId = null;
    userRole = null;
    userSpecialty = null;
    await _persistToken(null);
    loadToken();
    NotificationService.instance.resetSession();
    await DiscoveryController.instance.clearLocal();
    await GoogleAuth.signOut();
  }

  Future<Map<String, dynamic>> login(String email, String password) async {
    final res = await dio.post('/api/v1/auth/login', data: {
      'email': email,
      'password': password,
    });
    final data = res.data['data'] as Map<String, dynamic>;
    if (_isMfaChallenge(data)) return data;
    return _completeLogin(data);
  }

  Future<void> registerClient({
    required String email,
    required String password,
    required String fullName,
    String? locale,
    String? inviteCode,
  }) async {
    final code = inviteCode ?? await InviteCodeStore.instance.peek();
    await dio.post(
      '/api/v1/auth/register-client',
      data: {
        'email': email,
        'password': password,
        'fullName': fullName,
        if (code != null && code.isNotEmpty) 'inviteCode': code,
      },
      options: Options(
        headers: {
          if (locale != null && locale.isNotEmpty) 'Accept-Language': locale,
        },
      ),
    );
    if (code != null && code.isNotEmpty) {
      await InviteCodeStore.instance.save(null);
    }
  }

  Future<List<dynamic>> listCareProVisits() async {
    final res = await dio.get('/api/v1/care-pro/visits');
    final data = res.data['data'];
    return data is List ? data : [];
  }

  Future<List<dynamic>> listCareProClients() async {
    final res = await dio.get('/api/v1/care-pro/clients');
    final data = res.data['data'];
    return data is List ? data : [];
  }

  Future<List<dynamic>> listCareProPets() async {
    final res = await dio.get('/api/v1/care-pro/pets');
    final data = res.data['data'];
    return data is List ? data : [];
  }

  Future<List<dynamic>> listVetVisits() async {
    final res = await dio.get('/api/v1/vet/visits');
    final data = res.data['data'];
    return data is List ? data : [];
  }

  Future<List<dynamic>> listVetClients() async {
    final res = await dio.get('/api/v1/clients');
    final data = res.data['data'];
    return data is List ? data : [];
  }

  Future<List<dynamic>> listVetPets() async {
    final res = await dio.get('/api/v1/vet/pets');
    final data = res.data['data'];
    return data is List ? data : [];
  }

  Future<Map<String, dynamic>> getMyAppInvite() async {
    final res = await dio.get('/api/v1/me/app-invite');
    return Map<String, dynamic>.from(res.data['data'] as Map);
  }

  Future<({List<dynamic> visits, List<dynamic> clients, List<dynamic> pets})>
      loadProTerrainLists() async {
    if (userRole == 'vet') {
      return (
        visits: await listVetVisits(),
        clients: await listVetClients(),
        pets: await listVetPets(),
      );
    }
    return (
      visits: await listCareProVisits(),
      clients: await listCareProClients(),
      pets: await listCareProPets(),
    );
  }

  Future<Map<String, dynamic>> getVisitReport(String visitId) async {
    final res = await dio.get('/api/v1/visits/$visitId/report');
    return Map<String, dynamic>.from(res.data['data'] as Map);
  }

  Future<Map<String, dynamic>> putVisitReport(String visitId, String bodyText) async {
    final res = await dio.put('/api/v1/visits/$visitId/report', data: {
      'bodyText': bodyText,
    });
    return Map<String, dynamic>.from(res.data['data'] as Map);
  }

  Future<Map<String, dynamic>> improveVisitReport(String visitId) async {
    final res = await dio.post('/api/v1/visits/$visitId/report/improve');
    return Map<String, dynamic>.from(res.data['data'] as Map);
  }

  Future<Map<String, dynamic>> finalizeVisitReport(String visitId) async {
    final res = await dio.post('/api/v1/visits/$visitId/report/finalize');
    return Map<String, dynamic>.from(res.data['data'] as Map);
  }

  Future<List<dynamic>> listPetDocuments(String petId) async {
    final res = await dio.get('/api/v1/pets/$petId/documents');
    final data = res.data['data'];
    return data is List ? data : [];
  }

  Future<Map<String, dynamic>> updateVisitLocation(
    String visitId,
    String addressText, {
    double? lat,
    double? lng,
    bool clearCoords = false,
  }) async {
    final res = await dio.patch('/api/v1/visits/$visitId/location', data: {
      'addressText': addressText,
      if (clearCoords) 'clearCoords': true,
      if (!clearCoords && lat != null) 'lat': lat,
      if (!clearCoords && lng != null) 'lng': lng,
    });
    return Map<String, dynamic>.from(res.data['data'] as Map);
  }

  Future<Map<String, dynamic>> transcribeVisitReport(
    String visitId,
    String filePath, {
    String? filename,
    String? hint,
  }) async {
    final form = FormData.fromMap({
      if (hint != null && hint.trim().isNotEmpty) 'hint': hint.trim(),
      'audio': await MultipartFile.fromFile(
        filePath,
        filename: filename ?? filePath.split('/').last,
      ),
    });
    final res = await dio.post(
      '/api/v1/visits/$visitId/report/transcribe',
      data: form,
    );
    return Map<String, dynamic>.from(res.data['data'] as Map);
  }

  String? userId;
  String? userRole;
  String? userSpecialty;

  /// Google Sign-In for pets clients. [idToken] must be issued for the same
  /// Web client ID as API `GOOGLE_OAUTH_CLIENT_ID`.
  Future<Map<String, dynamic>> loginWithGoogle(String idToken) async {
    final inviteCode = await InviteCodeStore.instance.peek();
    final res = await dio.post('/api/v1/auth/google', data: {
      'idToken': idToken,
      'audience': 'client',
      if (inviteCode != null && inviteCode.isNotEmpty) 'inviteCode': inviteCode,
    });
    final data = res.data['data'] as Map<String, dynamic>;
    if (_isMfaChallenge(data)) return data;
    return _completeLogin(data);
  }

  /// Completes MFA after [login] / [loginWithGoogle] returned `requires2FA`.
  Future<Map<String, dynamic>> verify2FA(String mfaToken, String code) async {
    final res = await dio.post('/api/v1/auth/2fa/verify', data: {
      'mfaToken': mfaToken,
      'code': code,
    });
    return _completeLogin(res.data['data'] as Map<String, dynamic>);
  }

  Future<void> forgotPassword(String email) async {
    await dio.post('/api/v1/auth/forgot-password', data: {'email': email});
  }

  Future<void> resetPassword(String token, String password) async {
    await dio.post('/api/v1/auth/reset-password', data: {
      'token': token,
      'password': password,
    });
  }

  static bool _isMfaChallenge(Map<String, dynamic> data) {
    return data['requires2FA'] == true &&
        (data['mfaToken'] as String?)?.isNotEmpty == true;
  }

  Future<Map<String, dynamic>> _completeLogin(Map<String, dynamic> data) async {
    token = data['accessToken'] as String?;
    if (token == null || token!.isEmpty) {
      throw DioException(
        requestOptions: RequestOptions(path: '/api/v1/auth/login'),
        message: 'mfa_required_or_missing_token',
      );
    }
    _authGeneration++;
    await _persistToken(token);
    loadToken();
    await syncLocaleFromMe();
    try {
      final me = await getMe();
      userId = me['userId'] as String? ?? me['id'] as String?;
      userRole = me['role'] as String?;
      userSpecialty = me['professionalSpecialty'] as String?;
      DiscoveryController.instance.bindUser(userId);
    } catch (_) {
      // Keep token but force safe ACL defaults (Pet.isOwner → false without userId).
      userId = null;
    }
    await NotificationService.instance.onLogin();
    await _claimPendingInvite();
    return data;
  }

  Future<void> _claimPendingInvite() async {
    final code = await InviteCodeStore.instance.peek();
    if (code == null || code.isEmpty) return;
    try {
      await claimVetInvite(code);
      await InviteCodeStore.instance.save(null);
    } catch (e) {
      debugPrint('claim invite failed: $e');
    }
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
    final body = <String, dynamic>{'newPassword': newPassword};
    if (currentPassword.isNotEmpty) {
      body['currentPassword'] = currentPassword;
    }
    await dio.patch('/api/v1/me/password', data: body);
  }

  Future<bool> mustChangePassword() async {
    try {
      final me = await getMe();
      return me['mustChangePassword'] == true;
    } catch (_) {
      return false;
    }
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

  static List<dynamic> _asList(dynamic raw) {
    if (raw is List) return raw;
    return const [];
  }

  static Map<String, dynamic> _asMap(dynamic raw) {
    if (raw is Map) return Map<String, dynamic>.from(raw);
    throw StateError('expected map in API data envelope');
  }

  Future<List<dynamic>> getPets() async {
    final res = await dio.get('/api/v1/pets');
    return _asList(res.data is Map ? res.data['data'] : null);
  }

  Future<Map<String, dynamic>> createPet(Map<String, dynamic> body) async {
    final res = await dio.post('/api/v1/pets', data: body);
    return res.data['data'] as Map<String, dynamic>;
  }

  /// Kennel privilege: create several pets in one call (`POST /pets/batch`).
  Future<Map<String, dynamic>> createPetsBatch(List<Map<String, dynamic>> pets) async {
    final res = await dio.post('/api/v1/pets/batch', data: {'pets': pets});
    return Map<String, dynamic>.from(res.data['data'] as Map);
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

  Future<String> startAddonCheckout({required String addonCode}) async {
    final res = await dio.post('/api/v1/billing/addons/checkout', data: {
      'addonCode': addonCode,
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

  Future<Map<String, dynamic>> validateHeartRate(
    String sessionId, {
    String? comment,
  }) async {
    final trimmed = comment?.trim();
    final res = await dio.post(
      '/api/v1/heartrate/sessions/$sessionId/validate',
      data: (trimmed != null && trimmed.isNotEmpty) ? {'comment': trimmed} : null,
    );
    return res.data['data'] as Map<String, dynamic>;
  }

  Future<void> cancelHeartRate(String sessionId) async {
    await dio.post('/api/v1/heartrate/sessions/$sessionId/cancel');
  }

  Future<List<dynamic>> getTimeline(String petId) async {
    final res = await dio.get('/api/v1/pets/$petId/timeline');
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
    final data = _asList(res.data is Map ? res.data['data'] : null);
    return data
        .whereType<Map>()
        .map((v) => VetLink.fromJson(
              Map<String, dynamic>.from(v),
              primaryPracticeId: primaryPracticeId,
            ))
        .toList();
  }

  Future<Map<String, dynamic>> inviteVet(String email) async {
    final res = await dio.post('/api/v1/me/vets/invite', data: {'email': email});
    final data = res.data['data'];
    if (data is Map) {
      return Map<String, dynamic>.from(data);
    }
    return {'found': false, 'status': 'not_found'};
  }

  Future<Map<String, dynamic>> claimVetInvite(String code) async {
    final res = await dio.post('/api/v1/me/vets/claim-invite', data: {
      'code': code.trim().toUpperCase(),
    });
    final data = res.data['data'];
    if (data is Map) {
      return Map<String, dynamic>.from(data);
    }
    return {'status': 'linked'};
  }

  Future<void> setPetPrimaryPractice(String petId, String practiceId) async {
    await dio.patch('/api/v1/pets/$petId/primary-practice', data: {'practiceId': practiceId});
  }

  Future<List<CareReminder>> getCareReminders(String petId) async {
    final res = await dio.get('/api/v1/pets/$petId/care-reminders');
    final data = _asList(res.data is Map ? res.data['data'] : null);
    return data
        .whereType<Map>()
        .map((r) => CareReminder.fromJson(Map<String, dynamic>.from(r)))
        .toList();
  }

  Future<CareReminder> createCareReminder(
    String petId, {
    String? title,
    String? type,
    int? dueDays,
    String? dueAt,
    String? notes,
    int? recurrenceDays,
  }) async {
    final res = await dio.post('/api/v1/pets/$petId/care-reminders', data: {
      if (title != null && title.trim().isNotEmpty) 'title': title.trim(),
      if (type != null && type.isNotEmpty) 'type': type,
      if (dueDays != null) 'dueDays': dueDays,
      if (dueAt != null && dueAt.isNotEmpty) 'dueAt': dueAt,
      if (notes != null && notes.trim().isNotEmpty) 'notes': notes.trim(),
      if (recurrenceDays != null) 'recurrenceDays': recurrenceDays,
    });
    return CareReminder.fromJson(_asMap(res.data is Map ? res.data['data'] : null));
  }

  Future<CareReminder> markCareReminderDone(String id) async {
    final res = await dio.post('/api/v1/care-reminders/$id/done');
    return CareReminder.fromJson(_asMap(res.data is Map ? res.data['data'] : null));
  }

  Future<CareReminder> postponeCareReminder(String id, int days) async {
    final res = await dio.post('/api/v1/care-reminders/$id/postpone', data: {'days': days});
    return CareReminder.fromJson(_asMap(res.data is Map ? res.data['data'] : null));
  }

  Future<List<Visit>> getVisits(String petId) async {
    final res = await dio.get('/api/v1/pets/$petId/visits');
    final data = _asList(res.data is Map ? res.data['data'] : null);
    return data
        .whereType<Map>()
        .map((v) => Visit.fromJson(Map<String, dynamic>.from(v)))
        .toList();
  }

  Future<Visit> createVisit(
    String petId, {
    String? notes,
    DateTime? scheduledAt,
  }) async {
    final res = await dio.post('/api/v1/pets/$petId/visits', data: {
      if (notes != null) 'notes': notes,
      if (scheduledAt != null) 'scheduledAt': scheduledAt.toUtc().toIso8601String(),
    });
    return Visit.fromJson(res.data['data'] as Map<String, dynamic>);
  }

  Future<Visit> updateVisit(String id, String status) async {
    final res = await dio.patch('/api/v1/visits/$id', data: {'status': status});
    return Visit.fromJson(res.data['data'] as Map<String, dynamic>);
  }

  Future<Visit> visitAction(
    String id, {
    required String action,
    DateTime? proposedScheduledAt,
  }) async {
    final res = await dio.patch('/api/v1/visits/$id', data: {
      'action': action,
      if (proposedScheduledAt != null)
        'proposedScheduledAt': proposedScheduledAt.toUtc().toIso8601String(),
    });
    return Visit.fromJson(res.data['data'] as Map<String, dynamic>);
  }

  Future<Map<String, dynamic>> getPracticeAvailability(
    String practiceId, {
    required DateTime from,
    required DateTime to,
  }) async {
    final res = await dio.get(
      '/api/v1/practices/$practiceId/availability',
      queryParameters: {
        'from': from.toUtc().toIso8601String(),
        'to': to.toUtc().toIso8601String(),
      },
    );
    return Map<String, dynamic>.from(res.data['data'] as Map);
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

  /// Typed messaging threads (single wrapper for `/messaging/threads`).
  Future<List<MessageThread>> getMessageThreads() async {
    final res = await dio.get('/api/v1/messaging/threads');
    final data = res.data['data'] as List<dynamic>;
    return data.map((t) => MessageThread.fromJson(Map<String, dynamic>.from(t as Map))).toList();
  }

  /// Alias of [getMessageThreads] — prefer this or [getMessageThreads], not a raw duplicate.
  Future<List<MessageThread>> getThreads() => getMessageThreads();

  Future<List<ChatMessage>> getChatMessages(String threadId) async {
    final res = await dio.get('/api/v1/messaging/threads/$threadId/messages');
    final data = res.data['data'] as List<dynamic>;
    return data.map((m) => ChatMessage.fromJson(Map<String, dynamic>.from(m as Map))).toList();
  }

  Future<void> markThreadRead(String threadId) async {
    await dio.post('/api/v1/messaging/threads/$threadId/read');
  }
}
