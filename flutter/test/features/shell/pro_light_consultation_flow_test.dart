import 'package:dio/dio.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:petsfollow_mobile/core/api/api_client.dart';
import 'package:petsfollow_mobile/features/shell/consultation_errors.dart';

import '../../helpers/mock_api.dart';

void main() {
  group('isConsultationHasReportError', () {
    test('matches msgKey consultation_has_report', () {
      final e = DioException(
        requestOptions: RequestOptions(path: '/api/v1/visits/x'),
        response: Response(
          requestOptions: RequestOptions(path: '/api/v1/visits/x'),
          statusCode: 409,
          data: {
            'error': {
              'code': 'conflict',
              'msgKey': 'consultation_has_report',
            },
          },
        ),
        type: DioExceptionType.badResponse,
      );
      expect(isConsultationHasReportError(e), isTrue);
    });

    test('ignores other 409 conflicts', () {
      final e = DioException(
        requestOptions: RequestOptions(path: '/api/v1/visits/x'),
        response: Response(
          requestOptions: RequestOptions(path: '/api/v1/visits/x'),
          statusCode: 409,
          data: {
            'error': {'code': 'conflict', 'msgKey': 'slot_taken'},
          },
        ),
        type: DioExceptionType.badResponse,
      );
      expect(isConsultationHasReportError(e), isFalse);
    });
  });

  group('createVisit consultation flags', () {
    late MockApi mock;

    setUp(() {
      mock = MockApi();
      mock.install();
    });

    tearDown(() => mock.uninstall());

    test('sends confirmDirect + silentConfirm + consultationSession', () async {
      Map<String, dynamic>? body;
      mock.on('POST', RegExp(r'/api/v1/pets/.+/visits'), (options) {
        body = Map<String, dynamic>.from(options.data as Map);
        return mock.ok(
          options,
          {
            'id': 'visit-walkin',
            'petId': 'pet-1',
            'practiceId': 'prac-1',
            'status': 'confirmed',
            'source': 'care_pro',
            'consultationSession': true,
          },
          status: 201,
        );
      });

      final visit = await ApiClient.instance.createVisit(
        'pet-1',
        notes: 'terrain',
        scheduledAt: DateTime.utc(2026, 7, 28, 12),
        confirmDirect: true,
        silentConfirm: true,
        consultationSession: true,
        durationMinutes: 30,
      );

      expect(visit.id, 'visit-walkin');
      expect(visit.consultationSession, isTrue);
      expect(body?['confirmDirect'], isTrue);
      expect(body?['silentConfirm'], isTrue);
      expect(body?['consultationSession'], isTrue);
      expect(body?['durationMinutes'], 30);
    });

    test('discard keeps visit on consultation_has_report 409', () async {
      mock.on('PATCH', RegExp(r'/api/v1/visits/.+'), (options) {
        return Response(
          requestOptions: options,
          statusCode: 409,
          data: {
            'error': {
              'code': 'conflict',
              'msgKey': 'consultation_has_report',
            },
          },
        );
      });

      var kept = false;
      try {
        await ApiClient.instance.updateVisit('visit-1', 'cancelled');
      } on DioException catch (e) {
        kept = isConsultationHasReportError(e);
      }
      expect(kept, isTrue);
    });
  });
}
