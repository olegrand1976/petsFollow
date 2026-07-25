import 'package:flutter_test/flutter_test.dart';
import 'package:petsfollow_mobile/core/api/api_client.dart';
import 'package:petsfollow_mobile/l10n/app_localizations_fr.dart';
import 'package:petsfollow_mobile/core/api/api_errors.dart';
import 'package:dio/dio.dart';

import '../../helpers/mock_api.dart';

void main() {
  late MockApi mock;

  setUp(() {
    mock = MockApi();
    mock.install();
  });

  tearDown(() => mock.uninstall());

  test('createWeightReading posts weightKg and comment', () async {
    Map<String, dynamic>? body;
    mock.on('POST', r'/api/v1/pets/pet-1/weights', (options) {
      body = Map<String, dynamic>.from(options.data as Map);
      return mock.ok(options, {
        'id': 'w1',
        'weightKg': 12.5,
        'recordedAt': DateTime.now().toUtc().toIso8601String(),
      }, status: 201);
    });

    final res = await ApiClient.instance.createWeightReading(
      'pet-1',
      weightKg: 12.5,
      comment: 'note',
    );
    expect(body?['weightKg'], 12.5);
    expect(body?['comment'], 'note');
    expect(res['id'], 'w1');
  });

  test('mapApiError maps invalid_weight', () {
    final l10n = AppLocalizationsFr();
    final e = DioException(
      requestOptions: RequestOptions(path: '/weights'),
      response: Response(
        requestOptions: RequestOptions(path: '/weights'),
        statusCode: 400,
        data: {
          'error': {'code': 'invalid_weight', 'msgKey': 'invalid_weight'},
        },
      ),
      type: DioExceptionType.badResponse,
    );
    expect(mapApiError(e, l10n), l10n.weightInvalid);
  });
}
