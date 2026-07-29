import 'dart:async';
import 'dart:convert';

import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import 'package:petsfollow_mobile/core/api/api_errors.dart';
import 'package:petsfollow_mobile/core/auth/google_auth.dart';
import 'package:petsfollow_mobile/core/config/app_env.dart';
import 'package:petsfollow_mobile/core/discovery/discovery_controller.dart';
import 'package:petsfollow_mobile/core/invite/invite_code_store.dart';
import 'package:petsfollow_mobile/core/invite/preconsult_visit_store.dart';
import 'package:petsfollow_mobile/core/locale/locale_controller.dart';
import 'package:petsfollow_mobile/core/support/support_diagnostics_buffer.dart';
import 'package:petsfollow_mobile/core/models/care_reminder.dart';
import 'package:petsfollow_mobile/core/models/discovery_progress.dart';
import 'package:petsfollow_mobile/core/models/manager_overview.dart';
import 'package:petsfollow_mobile/core/models/message_thread.dart';
import 'package:petsfollow_mobile/core/models/notification_prefs.dart';
import 'package:petsfollow_mobile/core/models/practice_availability.dart';
import 'package:petsfollow_mobile/core/models/vet_link.dart';
import 'package:petsfollow_mobile/core/models/vet_lookup_hit.dart';
import 'package:petsfollow_mobile/core/models/visit.dart';
import 'package:petsfollow_mobile/core/notifications/notification_service.dart';
import 'package:petsfollow_mobile/core/notifications/push_navigation.dart';
import 'package:shared_preferences/shared_preferences.dart';

/// Payload JSON for HR validate when a non-blank comment is provided.
Map<String, String>? heartRateCommentPayload(String? comment) {
  final trimmed = comment?.trim();
  if (trimmed == null || trimmed.isEmpty) return null;
  return {'comment': trimmed};
}

/// Outcome of [ApiClient] refresh — never treat network blips as logout.
enum _TokenRefreshResult { success, invalid, transient }

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
  static const _refreshKey = 'pf_refresh';
  static const _sessionMetaKey = 'pf_session_meta';
  static const _apiBaseDefined = String.fromEnvironment('API_BASE');

  /// JWT en Keystore/Keychain — jamais en SharedPreferences (lisible sur device rooté/backup).
  static const _secureStorage = FlutterSecureStorage(
    aOptions: AndroidOptions(encryptedSharedPreferences: true),
  );

  /// Platform-aware local default (Android emu vs iOS sim). Override with `--dart-define=API_BASE=…`.
  static String get _apiBase {
    if (_apiBaseDefined.isNotEmpty) return _apiBaseDefined;
    if (defaultTargetPlatform == TargetPlatform.android) {
      return 'http://10.0.2.2:8291';
    }
    return 'http://localhost:8291';
  }

  String? token;
  String? refreshToken;

  /// Fired after a 401 clears the local session (AuthGate rebuilds → login).
  void Function()? onSessionInvalidated;

  /// Fired after a successful login / confirm-email session is persisted.
  void Function()? onSessionEstablished;

  bool _handlingUnauthorized = false;
  int _authGeneration = 0;
  Completer<_TokenRefreshResult>? _refreshInFlight;

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
      onError: (error, handler) async {
        final path = error.requestOptions.path;
        final alreadyRetried = error.requestOptions.extra['pf_auth_retried'] == true;
        if (error.response?.statusCode == 401 &&
            !_isPublicAuthPath(path) &&
            !alreadyRetried &&
            (token != null || refreshToken != null)) {
          final refreshResult = await _refreshTokens();
          switch (refreshResult) {
            case _TokenRefreshResult.success:
              final req = error.requestOptions;
              req.extra['pf_auth_retried'] = true;
              req.headers['Authorization'] = 'Bearer $token';
              try {
                final response = await dio.fetch(req);
                return handler.resolve(response);
              } on DioException catch (retryErr) {
                // Only wipe session if the server still rejects auth after refresh.
                if (retryErr.response?.statusCode == 401) {
                  unawaited(_invalidateSessionFromUnauthorized());
                }
                return handler.next(retryErr);
              } catch (_) {
                return handler.next(error);
              }
            case _TokenRefreshResult.invalid:
              unawaited(_invalidateSessionFromUnauthorized());
            case _TokenRefreshResult.transient:
              // Keep tokens; propagate the original 401 to the caller.
              break;
          }
        }
        handler.next(error);
      },
    ));
    dio.interceptors.add(SupportDiagnosticsBuffer.instance.dioInterceptor());
  }

  static bool _isPublicAuthPath(String path) {
    return path.contains('/api/v1/auth/');
  }

  /// Clears session after an authenticated 401. Safe to call multiple times.
  Future<void> _invalidateSessionFromUnauthorized() async {
    if (_handlingUnauthorized || (token == null && refreshToken == null)) return;
    _handlingUnauthorized = true;
    final generation = ++_authGeneration;
    token = null;
    refreshToken = null;
    userId = null;
    userRole = null;
    userSpecialty = null;
    onSessionInvalidated?.call();
    try {
      // Abort if a new login started after we cleared.
      if (_authGeneration != generation || token != null) return;
      await _persistTokens(null, null);
      await _persistSessionMeta();
      if (_authGeneration != generation || token != null) {
        if (token != null) {
          await _persistTokens(token, refreshToken);
          await _persistSessionMeta();
        }
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

  /// Restores session from Keystore. Access JWT (~15 min) is refreshed via
  /// [refreshToken] (API `JWT_REFRESH_TTL`, défaut 30 j). Network errors do
  /// not wipe the session. Forgot-password / reset stay independent of this.
  Future<void> restoreSession() async {
    token = await _readPersistedSecret(_tokenKey);
    refreshToken = await _readPersistedSecret(_refreshKey);
    // Legacy access JWT may still live in SharedPreferences.
    token ??= await _migrateLegacyAccessToken();
    await _restoreSessionMeta();
    loadToken();
    if (token == null && refreshToken == null) return;

    if (token == null) {
      switch (await _refreshTokens()) {
        case _TokenRefreshResult.success:
          break;
        case _TokenRefreshResult.invalid:
          await logout();
          return;
        case _TokenRefreshResult.transient:
          // Hors-ligne : garder refresh + meta rôle pour le shell.
          return;
      }
    }

    try {
      await _hydrateMe();
    } on DioException catch (e) {
      if (_isTransientNetworkError(e)) return;
      if (e.response?.statusCode == 401) {
        // Interceptor may already have refreshed or invalidated.
        if (token == null && refreshToken == null) return;
        switch (await _refreshTokens()) {
          case _TokenRefreshResult.success:
            try {
              await _hydrateMe();
            } on DioException catch (e2) {
              if (_isTransientNetworkError(e2)) return;
              if (e2.response?.statusCode == 401) await logout();
            } catch (_) {}
          case _TokenRefreshResult.invalid:
            await logout();
          case _TokenRefreshResult.transient:
            break;
        }
        return;
      }
      // Autre erreur HTTP : garder les tokens (API down, 5xx…).
    } catch (_) {
      // Ne pas détruire la session sur erreur inattendue au boot.
    }
  }

  Future<void> _hydrateMe() async {
    final me = await getMe();
    userId = me['userId'] as String? ?? me['id'] as String?;
    userRole = me['role'] as String?;
    userSpecialty = me['professionalSpecialty'] as String?;
    await _persistSessionMeta();
  }

  Future<void> _restoreSessionMeta() async {
    final raw = await _readPersistedSecret(_sessionMetaKey);
    if (raw == null || raw.isEmpty) return;
    try {
      final map = jsonDecode(raw) as Map<String, dynamic>;
      userId = map['userId'] as String?;
      userRole = map['role'] as String?;
      userSpecialty = map['specialty'] as String?;
    } catch (_) {}
  }

  Future<void> _persistSessionMeta() async {
    if (userId == null && userRole == null && userSpecialty == null) {
      try {
        await _secureStorage.delete(key: _sessionMetaKey);
      } catch (_) {}
      return;
    }
    try {
      await _secureStorage.write(
        key: _sessionMetaKey,
        value: jsonEncode({
          'userId': userId,
          'role': userRole,
          'specialty': userSpecialty,
        }),
      );
    } catch (_) {}
  }

  static bool _isTransientNetworkError(DioException e) {
    switch (e.type) {
      case DioExceptionType.connectionTimeout:
      case DioExceptionType.sendTimeout:
      case DioExceptionType.receiveTimeout:
      case DioExceptionType.connectionError:
      case DioExceptionType.transformTimeout:
        return true;
      case DioExceptionType.badResponse:
      case DioExceptionType.cancel:
      case DioExceptionType.badCertificate:
      case DioExceptionType.unknown:
        return e.response == null;
    }
  }

  static bool _isAuthRejectionStatus(int? status) {
    // Only a true unauthorized from the refresh endpoint should wipe the session.
    // 400/403 from proxies/WAF must not force logout.
    return status == 401;
  }

  /// Proactively exchange the refresh JWT so a post-payment API storm does not
  /// wipe the session on a stale 15-min access token.
  Future<bool> ensureFreshSession() async {
    if (refreshToken == null || refreshToken!.isEmpty) {
      refreshToken = await _readPersistedSecret(_refreshKey);
    }
    if (refreshToken == null || refreshToken!.isEmpty) return token != null;
    switch (await _refreshTokens()) {
      case _TokenRefreshResult.success:
        return true;
      case _TokenRefreshResult.transient:
        return token != null;
      case _TokenRefreshResult.invalid:
        unawaited(_invalidateSessionFromUnauthorized());
        return false;
    }
  }

  /// Exchange refresh JWT for a new access (+ rotated refresh). Single-flight.
  Future<_TokenRefreshResult> _refreshTokens() async {
    if (_refreshInFlight != null) return _refreshInFlight!.future;
    final stored = refreshToken ?? await _readPersistedSecret(_refreshKey);
    if (stored == null || stored.isEmpty) return _TokenRefreshResult.invalid;

    final completer = Completer<_TokenRefreshResult>();
    _refreshInFlight = completer;
    try {
      final res = await dio.post(
        '/api/v1/auth/refresh',
        data: {'refreshToken': stored},
        options: Options(extra: {'pf_auth_retried': true}),
      );
      final data = res.data['data'] as Map<String, dynamic>?;
      final access = data?['accessToken'] as String?;
      final nextRefresh = data?['refreshToken'] as String? ?? stored;
      if (access == null || access.isEmpty) {
        completer.complete(_TokenRefreshResult.invalid);
        return _TokenRefreshResult.invalid;
      }
      token = access;
      refreshToken = nextRefresh;
      await _persistTokens(token, refreshToken);
      // onRequest lit [token] à chaud — pas de loadToken() ici (évite de
      // recréer l'interceptor pendant un onError en cours).
      completer.complete(_TokenRefreshResult.success);
      return _TokenRefreshResult.success;
    } on DioException catch (e) {
      // 5xx / pas de status / réseau → transient ; seul 401 refresh = invalid.
      // 400/403 (WAF/proxy) restent transient pour ne pas forcer un logout.
      final resolved = e.response?.statusCode == null || _isTransientNetworkError(e)
          ? _TokenRefreshResult.transient
          : (_isAuthRejectionStatus(e.response?.statusCode)
              ? _TokenRefreshResult.invalid
              : _TokenRefreshResult.transient);
      completer.complete(resolved);
      return resolved;
    } catch (_) {
      completer.complete(_TokenRefreshResult.transient);
      return _TokenRefreshResult.transient;
    } finally {
      if (identical(_refreshInFlight, completer)) {
        _refreshInFlight = null;
      }
    }
  }

  Future<String?> _readPersistedSecret(String key) async {
    try {
      return await _secureStorage.read(key: key);
    } catch (_) {
      return null;
    }
  }

  Future<String?> _migrateLegacyAccessToken() async {
    final sp = await SharedPreferences.getInstance();
    final legacy = sp.getString(_tokenKey);
    if (legacy == null) return null;
    try {
      await _secureStorage.write(key: _tokenKey, value: legacy);
      await sp.remove(_tokenKey);
    } catch (_) {
      // Réessaiera au prochain démarrage.
    }
    return legacy;
  }

  Future<void> _persistTokens(String? access, String? refresh) async {
    try {
      if (access == null) {
        await _secureStorage.delete(key: _tokenKey);
      } else {
        await _secureStorage.write(key: _tokenKey, value: access);
      }
    } catch (_) {}
    try {
      if (refresh == null) {
        await _secureStorage.delete(key: _refreshKey);
      } else {
        await _secureStorage.write(key: _refreshKey, value: refresh);
      }
    } catch (_) {}
    // Purge systématique de l'ancien emplacement en clair.
    final sp = await SharedPreferences.getInstance();
    await sp.remove(_tokenKey);
  }

  Future<void> logout() async {
    token = null;
    refreshToken = null;
    userId = null;
    userRole = null;
    userSpecialty = null;
    await _persistTokens(null, null);
    await _persistSessionMeta();
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

  Future<Map<String, dynamic>> registerClient({
    required String email,
    required String password,
    required String fullName,
    String? locale,
    String? inviteCode,
    String? commercialUserId,
    bool consent = false,
  }) async {
    final fromStore = await InviteCodeStore.instance.peek();
    final code = (inviteCode != null && inviteCode.trim().isNotEmpty)
        ? inviteCode.trim().toUpperCase()
        : fromStore;
    if (code != null && code.isNotEmpty) {
      await InviteCodeStore.instance.save(code);
    }
    final res = await dio.post(
      '/api/v1/auth/register-client',
      data: {
        'email': email,
        'password': password,
        'fullName': fullName,
        'consent': consent,
        if (code != null && code.isNotEmpty) 'inviteCode': code,
        if ((code == null || code.isEmpty) &&
            commercialUserId != null &&
            commercialUserId.isNotEmpty)
          'commercialUserId': commercialUserId,
      },
      options: Options(
        headers: {
          if (locale != null && locale.isNotEmpty) 'Accept-Language': locale,
        },
      ),
    );
    final data = res.data is Map ? res.data['data'] : null;
    final inviteStatus =
        data is Map ? data['inviteStatus']?.toString() ?? '' : '';
    // Clear after success or known-invalid code; keep on soft failure for login retry.
    if (_inviteClaimSettled(inviteStatus)) {
      await InviteCodeStore.instance.save(null);
    }
    return data is Map
        ? Map<String, dynamic>.from(data)
        : <String, dynamic>{};
  }

  static bool _inviteClaimSucceeded(String status) {
    switch (status) {
      case 'referred':
      case 'linked':
      case 'granted':
      case 'already_linked':
        return true;
      default:
        return false;
    }
  }

  /// Clear pending invite after a definitive outcome (success or known-invalid).
  static bool _inviteClaimSettled(String status) {
    return _inviteClaimSucceeded(status) || status == 'ignored';
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

  /// Agenda terrain véto : plage calendrier (confirmed + scheduled), pas le pending-only de [listVetVisits].
  /// Cap API `/vet/calendar` : plage ≤ 62 jours (`parseFromTo`).
  Future<List<dynamic>> listVetTourVisits({int pastDays = 7, int futureDays = 52}) async {
    final now = DateTime.now();
    final from = DateTime(now.year, now.month, now.day).subtract(Duration(days: pastDays));
    final to = DateTime(now.year, now.month, now.day).add(Duration(days: futureDays + 1));
    // Date-only — évite le décalage minuit local→UTC qui élargit la plage.
    String ymd(DateTime d) =>
        '${d.year.toString().padLeft(4, '0')}-${d.month.toString().padLeft(2, '0')}-${d.day.toString().padLeft(2, '0')}';
    final res = await dio.get('/api/v1/vet/calendar', queryParameters: {
      'from': ymd(from),
      'to': ymd(to),
    });
    final data = Map<String, dynamic>.from(res.data['data'] as Map);
    final byId = <String, Map<String, dynamic>>{};
    for (final raw in [
      ...(data['visits'] as List? ?? const []),
      ...(data['pending'] as List? ?? const []),
    ]) {
      final m = Map<String, dynamic>.from(raw as Map);
      final id = m['id']?.toString() ?? '';
      if (id.isEmpty) continue;
      byId.putIfAbsent(id, () => m);
    }
    return byId.values.toList();
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

  Future<ManagerOverview> getCommercialManagerOverview() async {
    final res = await dio.get('/api/v1/commercial-manager/overview');
    return ManagerOverview.fromJson(
      Map<String, dynamic>.from(res.data['data'] as Map),
    );
  }

  Future<CommercialSelfStats> getCommercialManagerMemberOverview(
    String memberUserId,
  ) async {
    final res = await dio.get(
      '/api/v1/commercial-manager/team/$memberUserId/overview',
    );
    return CommercialSelfStats.fromJson(
      Map<String, dynamic>.from(res.data['data'] as Map),
    );
  }

  Future<({List<dynamic> visits, List<dynamic> clients, List<dynamic> pets})>
      loadProTerrainLists() async {
    if (userRole == 'vet' ||
        userRole == 'vet_assistant' ||
        userRole == 'secretary') {
      return (
        visits: await listVetTourVisits(),
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

  Future<Map<String, dynamic>> getAiModule() async {
    final res = await dio.get('/api/v1/me/ai-module');
    return Map<String, dynamic>.from(res.data['data'] as Map? ?? res.data as Map);
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
    bool clientAudioConsent = false,
    int? audioDurationSec,
  }) async {
    final form = FormData.fromMap({
      if (hint != null && hint.trim().isNotEmpty) 'hint': hint.trim(),
      'clientAudioConsent': clientAudioConsent ? 'true' : 'false',
      if (audioDurationSec != null && audioDurationSec > 0)
        'audioDurationSec': '$audioDurationSec',
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
  ///
  /// [consent] must be true to create a new client (create-if-absent).
  Future<Map<String, dynamic>> loginWithGoogle(
    String idToken, {
    bool consent = false,
    String? commercialUserId,
  }) async {
    final inviteCode = await InviteCodeStore.instance.peek();
    final res = await dio.post('/api/v1/auth/google', data: {
      'idToken': idToken,
      'audience': 'client',
      if (consent) 'consent': true,
      if (inviteCode != null && inviteCode.isNotEmpty) 'inviteCode': inviteCode,
      if ((inviteCode == null || inviteCode.isEmpty) &&
          commercialUserId != null &&
          commercialUserId.isNotEmpty)
        'commercialUserId': commercialUserId,
    });
    final data = res.data['data'] as Map<String, dynamic>;
    final inviteStatus = data['inviteStatus']?.toString() ?? '';
    if (_inviteClaimSettled(inviteStatus)) {
      await InviteCodeStore.instance.save(null);
    }
    if (_isMfaChallenge(data)) return data;
    return _completeLogin(data);
  }

  Future<List<Map<String, dynamic>>> listNearbyCommercials({
    double? lat,
    double? lng,
    String? postalCode,
    int limit = 5,
  }) async {
    final res = await dio.get('/api/v1/commercials/nearby', queryParameters: {
      if (lat != null) 'lat': lat,
      if (lng != null) 'lng': lng,
      if (postalCode != null && postalCode.isNotEmpty) 'postalCode': postalCode,
      'limit': limit,
    });
    final data = res.data['data'];
    if (data is! List) return [];
    return data
        .whereType<Map>()
        .map((e) => Map<String, dynamic>.from(e))
        .toList();
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

  /// Always 200 — does not reveal whether the email exists / is already verified.
  Future<void> resendConfirmation(String email) async {
    await dio.post('/api/v1/auth/resend-confirmation', data: {'email': email});
  }

  Future<void> resetPassword(String token, String password) async {
    await dio.post('/api/v1/auth/reset-password', data: {
      'token': token,
      'password': password,
    });
  }

  /// Confirms email via token and establishes a client session (auto-login).
  Future<Map<String, dynamic>> confirmEmail(String token) async {
    final res = await dio.post('/api/v1/auth/confirm-email', data: {'token': token});
    return _completeLogin(res.data['data'] as Map<String, dynamic>);
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
    refreshToken = data['refreshToken'] as String?;
    _authGeneration++;
    await _persistTokens(token, refreshToken);
    loadToken();
    await syncLocaleFromMe();
    try {
      final me = await getMe();
      userId = me['userId'] as String? ?? me['id'] as String?;
      userRole = me['role'] as String?;
      userSpecialty = me['professionalSpecialty'] as String?;
      DiscoveryController.instance.bindUser(userId);
      await _persistSessionMeta();
    } catch (_) {
      // Keep token but force safe ACL defaults (Pet.isOwner → false without userId).
      userId = null;
    }
    await NotificationService.instance.onLogin();
    await _claimPendingInvite();
    await _openPendingPreconsult();
    onSessionEstablished?.call();
    return data;
  }

  Future<void> _claimPendingInvite() async {
    final code = await InviteCodeStore.instance.peek();
    if (code == null || code.isEmpty) return;
    try {
      final result = await claimVetInvite(code);
      final status = result['status']?.toString() ?? '';
      if (_inviteClaimSucceeded(status) || status.isNotEmpty) {
        await InviteCodeStore.instance.save(null);
      }
    } on DioException catch (e) {
      final err = apiErrorCode(e);
      // Unknown / invalid code: stop retrying forever.
      if (err == 'invite_not_found' ||
          err == 'not_found' ||
          err == 'invalid_role' ||
          err == 'bad_request' ||
          e.response?.statusCode == 404 ||
          e.response?.statusCode == 400) {
        await InviteCodeStore.instance.save(null);
        return;
      }
      debugPrint('claim invite failed: $e');
    } catch (e) {
      debugPrint('claim invite failed: $e');
    }
  }

  Future<void> _openPendingPreconsult() async {
    final visitId = await PreconsultVisitStore.instance.peek();
    if (visitId == null || visitId.isEmpty) return;
    // Defer until navigator / shell callbacks are ready.
    Future<void>.delayed(const Duration(milliseconds: 500), () async {
      final still = await PreconsultVisitStore.instance.peek();
      if (still == null || still.isEmpty) return;
      PushNavigation.instance.openPreconsult(still);
    });
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

  /// Multi-image health book → server builds one compressed PDF.
  Future<Map<String, dynamic>> uploadPetHealthBook(
    String petId,
    List<String> filePaths,
  ) async {
    final files = <MultipartFile>[];
    for (var i = 0; i < filePaths.length; i++) {
      files.add(await MultipartFile.fromFile(
        filePaths[i],
        filename: 'page_$i.jpg',
      ));
    }
    final form = FormData.fromMap({'files': files});
    final res = await dio.post('/api/v1/pets/$petId/health-book', data: form);
    return res.data['data'] as Map<String, dynamic>;
  }

  Future<Map<String, dynamic>> deletePetHealthBook(String petId) async {
    final res = await dio.delete('/api/v1/pets/$petId/health-book');
    return res.data['data'] as Map<String, dynamic>;
  }

  /// Authenticated PDF bytes (PHI — never a public media URL).
  Future<List<int>> downloadPetHealthBook(String petId) async {
    final res = await dio.get<List<int>>(
      '/api/v1/pets/$petId/health-book',
      options: Options(responseType: ResponseType.bytes),
    );
    return res.data ?? const <int>[];
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

  /// Portabilité RGPD : export JSON brut des données du compte.
  Future<Map<String, dynamic>> exportMyData() async {
    final res = await dio.get('/api/v1/me/export');
    return res.data['data'] as Map<String, dynamic>;
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
    return _asMap(res.data is Map ? res.data['data'] : null);
  }

  /// Kennel privilege: create several pets in one call (`POST /pets/batch`).
  Future<void> updatePet(String petId, Map<String, dynamic> body) async {
    await dio.put('/api/v1/pets/$petId', data: body);
  }

  Future<Map<String, dynamic>> createPetsBatch(List<Map<String, dynamic>> pets) async {
    final res = await dio.post('/api/v1/pets/batch', data: {'pets': pets});
    return _asMap(res.data is Map ? res.data['data'] : null);
  }

  Future<Map<String, dynamic>> getPet(String petId) async {
    final res = await dio.get('/api/v1/pets/$petId');
    return res.data['data'] as Map<String, dynamic>;
  }

  /// Send a 24h download link for the full pet dossier to a care professional.
  Future<Map<String, dynamic>> sendPetDossierShare(String petId, String email) async {
    final res = await dio.post(
      '/api/v1/pets/$petId/dossier-shares',
      data: {'email': email.trim()},
    );
    return _asMap(res.data is Map ? res.data['data'] : null);
  }

  /// Finalized consultation report(s) for the pet owner.
  Future<Map<String, dynamic>> getClientConsultation(String visitId) async {
    final res = await dio.get('/api/v1/visits/$visitId/client-consultation');
    return _asMap(res.data is Map ? res.data['data'] : null);
  }

  /// Send a 24h PDF download link for a finalized consultation to a vet.
  Future<Map<String, dynamic>> sendConsultationShare(String visitId, String email) async {
    final res = await dio.post(
      '/api/v1/visits/$visitId/consultation-shares',
      data: {'email': email.trim()},
    );
    return _asMap(res.data is Map ? res.data['data'] : null);
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

  /// Household digest (all clients — pack derived from pet count).
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
    final res = await dio.post(
      '/api/v1/heartrate/sessions/$sessionId/validate',
      data: heartRateCommentPayload(comment),
    );
    return res.data['data'] as Map<String, dynamic>;
  }

  Future<void> cancelHeartRate(String sessionId) async {
    await dio.post('/api/v1/heartrate/sessions/$sessionId/cancel');
  }

  Future<Map<String, dynamic>> createWeightReading(
    String petId, {
    required double weightKg,
    String? comment,
  }) async {
    final data = <String, dynamic>{'weightKg': weightKg};
    final trimmed = comment?.trim();
    if (trimmed != null && trimmed.isNotEmpty) {
      data['comment'] = trimmed;
    }
    final res = await dio.post('/api/v1/pets/$petId/weights', data: data);
    return Map<String, dynamic>.from(res.data['data'] as Map);
  }

  Future<List<dynamic>> getWeightReadings(String petId) async {
    final res = await dio.get('/api/v1/pets/$petId/weights');
    return res.data['data'] as List<dynamic>;
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

  Future<List<VetLookupHit>> lookupVets(String query) async {
    final q = query.trim();
    if (q.length < 3) return [];
    final res = await dio.get('/api/v1/me/vets/lookup', queryParameters: {'q': q});
    final data = _asList(res.data is Map ? res.data['data'] : null);
    return data
        .whereType<Map>()
        .map((v) => VetLookupHit.fromJson(Map<String, dynamic>.from(v)))
        .toList();
  }

  Future<Map<String, dynamic>> inviteVet({String? email, String? vetUserId}) async {
    final body = <String, dynamic>{};
    if (vetUserId != null && vetUserId.trim().isNotEmpty) {
      body['vetUserId'] = vetUserId.trim();
    } else if (email != null && email.trim().isNotEmpty) {
      body['email'] = email.trim();
    }
    final res = await dio.post('/api/v1/me/vets/invite', data: body);
    final data = res.data['data'];
    if (data is Map) {
      return Map<String, dynamic>.from(data);
    }
    return {'found': false, 'status': 'not_found'};
  }

  Future<Map<String, dynamic>> suggestVet({
    required String email,
    required String phone,
    String? fullName,
    String? practiceName,
  }) async {
    final res = await dio.post('/api/v1/me/vets/suggest', data: {
      'email': email.trim(),
      'phone': phone.trim(),
      if (fullName != null && fullName.trim().isNotEmpty) 'fullName': fullName.trim(),
      if (practiceName != null && practiceName.trim().isNotEmpty) 'practiceName': practiceName.trim(),
    });
    final data = res.data['data'];
    if (data is Map) {
      return Map<String, dynamic>.from(data);
    }
    return {'status': 'suggested', 'found': false};
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

  Future<Map<String, dynamic>> getPreconsult(String visitId) async {
    final res = await dio.get('/api/v1/visits/$visitId/preconsult');
    return _asMap(res.data is Map ? res.data['data'] : null);
  }

  Future<Map<String, dynamic>> submitPreconsult(
    String visitId,
    Map<String, dynamic> answers,
  ) async {
    final res = await dio.put('/api/v1/visits/$visitId/preconsult', data: {
      'answers': answers,
    });
    return _asMap(res.data is Map ? res.data['data'] : null);
  }

  Future<Visit> createVisit(
    String petId, {
    String? notes,
    DateTime? scheduledAt,
    bool confirmDirect = false,
    bool silentConfirm = false,
    bool consultationSession = false,
    int? durationMinutes,
  }) async {
    final res = await dio.post('/api/v1/pets/$petId/visits', data: {
      if (notes != null) 'notes': notes,
      if (scheduledAt != null) 'scheduledAt': scheduledAt.toUtc().toIso8601String(),
      if (confirmDirect) 'confirmDirect': true,
      if (silentConfirm) 'silentConfirm': true,
      if (consultationSession) 'consultationSession': true,
      if (durationMinutes != null) 'durationMinutes': durationMinutes,
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

  Future<PracticeAvailability> getPracticeAvailability(
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
    return PracticeAvailability.fromJson(
      Map<String, dynamic>.from(res.data['data'] as Map),
    );
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

  Future<Map<String, dynamic>> getFeatureModules() async {
    final res = await dio.get('/api/v1/me/feature-modules');
    return Map<String, dynamic>.from(res.data['data'] as Map);
  }

  Future<Map<String, dynamic>> updateFeatureModules(Map<String, dynamic> body) async {
    final res = await dio.patch('/api/v1/me/feature-modules', data: body);
    return Map<String, dynamic>.from(res.data['data'] as Map);
  }

  Future<List<Map<String, dynamic>>> getProfiles() async {
    final res = await dio.get('/api/v1/me/profiles');
    final data = res.data['data'] as List<dynamic>;
    return data.map((e) => Map<String, dynamic>.from(e as Map)).toList();
  }

  Future<Map<String, dynamic>> switchProfile(String profileId) async {
    final res = await dio.post('/api/v1/me/profiles/switch', data: {'profileId': profileId});
    final data = Map<String, dynamic>.from(res.data['data'] as Map);
    return _completeLogin(data);
  }

  /// Typed messaging threads (single wrapper for `/messaging/threads`).
  Future<List<MessageThread>> getMessageThreads() async {
    final res = await dio.get('/api/v1/messaging/threads');
    final raw = res.data['data'];
    final data = raw is List ? raw : const <dynamic>[];
    return data
        .whereType<Map>()
        .map((t) => MessageThread.fromJson(Map<String, dynamic>.from(t)))
        .toList();
  }

  /// Alias of [getMessageThreads] — prefer this or [getMessageThreads], not a raw duplicate.
  Future<List<MessageThread>> getThreads() => getMessageThreads();

  /// Client: [practiceId] + [petId]. Staff: [clientUserId] (+ optional [petId]).
  Future<MessageThread> ensureMessageThread({
    String? practiceId,
    String? petId,
    String? clientUserId,
  }) async {
    final data = <String, dynamic>{};
    final cid = clientUserId?.trim() ?? '';
    if (cid.isNotEmpty) {
      data['clientUserId'] = cid;
      final pid = petId?.trim() ?? '';
      if (pid.isNotEmpty) data['petId'] = pid;
    } else {
      data['practiceId'] = practiceId?.trim() ?? '';
      data['petId'] = petId?.trim() ?? '';
    }
    final res = await dio.post('/api/v1/messaging/threads', data: data);
    return MessageThread.fromJson(Map<String, dynamic>.from(res.data['data'] as Map));
  }

  Future<List<ChatMessage>> getChatMessages(String threadId) async {
    final res = await dio.get('/api/v1/messaging/threads/$threadId/messages');
    final data = res.data['data'] as List<dynamic>;
    return data.map((m) => ChatMessage.fromJson(Map<String, dynamic>.from(m as Map))).toList();
  }

  Future<void> markThreadRead(String threadId) async {
    await dio.post('/api/v1/messaging/threads/$threadId/read');
  }

  Future<Map<String, dynamic>> createSupportTicket({
    required String source,
    required String subject,
    required String message,
    required Map<String, dynamic> diagnostics,
    String? route,
    String? appVersion,
  }) async {
    final res = await dio.post('/api/v1/support/tickets', data: {
      'source': source,
      'subject': subject,
      'message': message,
      'diagnostics': diagnostics,
      'locale': LocaleController.instance.languageCode,
      'route': route ?? '',
      'appVersion': appVersion ?? 'flutter/${AppEnv.value}',
      'userAgent': 'petsfollow-flutter/${AppEnv.value}',
    });
    final data = res.data['data'];
    if (data is Map<String, dynamic>) return data;
    if (data is Map) return Map<String, dynamic>.from(data);
    return {};
  }
}
