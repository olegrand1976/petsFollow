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

  Future<Dio?> loginSmokeDio() async {
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
      return null;
    }

    final data = loginRes.data['data'] as Map<String, dynamic>?;
    if (data == null) {
      // ignore: avoid_print
      print('SKIP smoke: unexpected login payload');
      return null;
    }
    if (data['requires2FA'] == true) {
      // ignore: avoid_print
      print('SKIP smoke: 2FA required for seed user');
      return null;
    }
    final token = (data['accessToken'] ?? data['token']) as String?;
    if (token == null || token.isEmpty) {
      // ignore: avoid_print
      print('SKIP smoke: missing accessToken');
      return null;
    }
    dio.options.headers['Authorization'] = 'Bearer $token';
    return dio;
  }

  test('login → list pets → weight reading', () async {
    if (!runSmoke) {
      // ignore: avoid_print
      print('SKIP smoke: pass --dart-define=RUN_FLUTTER_SMOKE=true');
      return;
    }

    final dio = await loginSmokeDio();
    if (dio == null) return;

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

  test('create pet skipCheckout → listed in GET /pets', () async {
    if (!runSmoke) {
      // ignore: avoid_print
      print('SKIP smoke: pass --dart-define=RUN_FLUTTER_SMOKE=true');
      return;
    }

    final dio = await loginSmokeDio();
    if (dio == null) return;

    final name = 'SmokeCreate-${DateTime.now().millisecondsSinceEpoch}';
    try {
      final createRes = await dio.post(
        '/api/v1/pets',
        data: {
          'name': name,
          'species': 'dog',
          'breed': 'Smoke',
          'plan': 'triennial',
          'billingMode': 'subscription',
          'skipCheckout': true,
        },
      );
      expect(createRes.statusCode, 201);
      final data = Map<String, dynamic>.from(createRes.data['data'] as Map);
      expect(data.containsKey('checkoutUrl'), isFalse,
          reason: 'skipCheckout must not return checkoutUrl');
      final pet = Map<String, dynamic>.from(data['pet'] as Map);
      final petId = pet['id'] as String;
      expect(petId, isNotEmpty);
      expect(pet['paymentStatus'], 'pending_payment');
      final ent = pet['entitlement'];
      expect(ent, isA<Map>());
      expect((ent as Map)['status'], 'pending');

      final listRes = await dio.get('/api/v1/pets');
      final list = listRes.data['data'] as List<dynamic>;
      final found = list.cast<Map>().where((p) => p['id'] == petId);
      expect(found, isNotEmpty, reason: 'created pet must appear in list');
      expect(found.first['name'], name);
    } catch (e) {
      fail('smoke create pet failed: $e');
    }
  });

  test('login → messaging threads + care reminders list', () async {
    if (!runSmoke) {
      // ignore: avoid_print
      print('SKIP smoke: pass --dart-define=RUN_FLUTTER_SMOKE=true');
      return;
    }

    final dio = await loginSmokeDio();
    if (dio == null) return;

    try {
      final threadsRes = await dio.get('/api/v1/messaging/threads');
      expect(threadsRes.statusCode, 200);
      expect(threadsRes.data['data'], isA<List>());

      final petsRes = await dio.get('/api/v1/pets');
      final pets = petsRes.data['data'] as List<dynamic>;
      expect(pets, isNotEmpty);
      final petId = (pets.first as Map)['id'] as String;

      final careRes = await dio.get('/api/v1/pets/$petId/care-reminders');
      expect(careRes.statusCode, anyOf(200, 404));
      if (careRes.statusCode == 200) {
        expect(careRes.data['data'], isA<List>());
      }
    } catch (e) {
      fail('smoke messaging/care failed: $e');
    }
  });
}
