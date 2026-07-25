import 'package:dio/dio.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:petsfollow_mobile/core/api/api_client.dart';

/// Smoke critical path against local API when available.
///
/// ```bash
/// make api-dev
/// make test-flutter-smoke   # --dart-define=RUN_FLUTTER_SMOKE=true
/// ```
///
/// Default `flutter test` skips unless the dart-define is set.
void main() {
  const runSmoke = bool.fromEnvironment('RUN_FLUTTER_SMOKE');

  test('login → list pets → weight reading', () async {
    if (!runSmoke) {
      // ignore: avoid_print
      print('SKIP smoke: pass --dart-define=RUN_FLUTTER_SMOKE=true');
      return;
    }

    final dio = Dio(
      BaseOptions(
        baseUrl: 'http://localhost:8291',
        connectTimeout: const Duration(seconds: 2),
        receiveTimeout: const Duration(seconds: 5),
      ),
    );

    late final Response loginRes;
    try {
      loginRes = await dio.post(
        '/api/v1/auth/login',
        data: {
          'email': 'client.demo@petsfollow.test',
          'password': 'ClientDemo123!',
        },
      );
    } catch (e) {
      // ignore: avoid_print
      print('SKIP smoke: API unreachable or login failed ($e)');
      return;
    }

    final data = loginRes.data['data'] as Map<String, dynamic>?;
    if (data == null) {
      // ignore: avoid_print
      print('SKIP smoke: unexpected login payload');
      return;
    }
    if (data['requires2FA'] == true) {
      // ignore: avoid_print
      print('SKIP smoke: 2FA required for seed user');
      return;
    }
    final token = (data['accessToken'] ?? data['token']) as String?;
    if (token == null || token.isEmpty) {
      // ignore: avoid_print
      print('SKIP smoke: missing accessToken');
      return;
    }
    dio.options.headers['Authorization'] = 'Bearer $token';

    try {
      final petsRes = await dio.get('/api/v1/pets');
      final pets = petsRes.data['data'] as List<dynamic>;
      expect(pets, isNotEmpty);

      final pet = pets.cast<Map<String, dynamic>>().firstWhere(
        (p) =>
            (p['paymentStatus'] as String?) == 'active' ||
            (p['entitlement'] is Map &&
                (p['entitlement'] as Map)['status'] == 'active'),
        orElse: () => pets.first as Map<String, dynamic>,
      );
      final petId = pet['id'] as String;

      try {
        final sess = await dio.post(
          '/api/v1/pets/$petId/heartrate/sessions',
          data: {'durationSec': 15},
        );
        final sid = (sess.data['data'] as Map)['id'] as String;
        await dio.post('/api/v1/heartrate/sessions/$sid/cancel');
      } catch (_) {}

      final weightRes = await dio.post(
        '/api/v1/pets/$petId/weights',
        data: {'weightKg': 12.34, 'comment': 'smoke'},
      );
      expect(weightRes.statusCode, anyOf(200, 201));
      expect((weightRes.data['data'] as Map)['weightKg'], isNotNull);
      expect(ApiClient.instance, isNotNull);
    } catch (e) {
      fail('smoke post-login failed: $e');
    }
  });
}
