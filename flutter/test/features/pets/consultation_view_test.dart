import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:petsfollow_mobile/core/api/api_client.dart';
import 'package:petsfollow_mobile/features/pets/presentation/consultation_view_screen.dart';
import 'package:petsfollow_mobile/features/pets/presentation/pet_timeline_screen.dart';
import 'package:petsfollow_mobile/l10n/app_localizations_fr.dart';

import '../../helpers/pump_app.dart';

void main() {
  const petId = 'pet-bella';
  const visitId = 'visit-1';
  late Interceptor mockInterceptor;
  String? postedEmail;

  setUp(() {
    postedEmail = null;
    ApiClient.instance.dio.interceptors.clear();
    ApiClient.instance.userId = 'user-1';
    ApiClient.instance.token = null;
    mockInterceptor = InterceptorsWrapper(
      onRequest: (options, handler) {
        final path = options.uri.path.isNotEmpty ? options.uri.path : options.path;
        if (options.method == 'GET' && path.contains('/heartrate/sessions')) {
          handler.resolve(
            Response(requestOptions: options, statusCode: 200, data: {'data': []}),
          );
          return;
        }
        if (options.method == 'GET' && path.contains('/pets/$petId/visits')) {
          handler.resolve(
            Response(
              requestOptions: options,
              statusCode: 200,
              data: {
                'data': [
                  {
                    'id': visitId,
                    'petId': petId,
                    'status': 'done',
                    'scheduledAt': '2026-01-10T10:00:00Z',
                    'hasFinalReport': true,
                    'consultationSession': true,
                  },
                ],
              },
            ),
          );
          return;
        }
        if (options.method == 'GET' && path.contains('/pets/$petId/timeline')) {
          handler.resolve(
            Response(
              requestOptions: options,
              statusCode: 200,
              data: {'data': <dynamic>[]},
            ),
          );
          return;
        }
        if (options.method == 'GET' &&
            (path.endsWith('/pets/$petId') || path.contains('/pets/$petId?'))) {
          handler.resolve(
            Response(
              requestOptions: options,
              statusCode: 200,
              data: {
                'data': {
                  'id': petId,
                  'name': 'Bella',
                  'ownerUserId': 'user-1',
                  'canWriteNotes': true,
                },
              },
            ),
          );
          return;
        }
        if (options.method == 'GET' && path.contains('/client-consultation')) {
          handler.resolve(
            Response(
              requestOptions: options,
              statusCode: 200,
              data: {
                'data': {
                  'visitId': visitId,
                  'petId': petId,
                  'petName': 'Bella',
                  'practiceName': 'VetPlus',
                  'status': 'done',
                  'scheduledAt': '2026-01-10T10:00:00Z',
                  'reports': [
                    {
                      'id': 'r1',
                      'authorName': 'Dr Demo',
                      'bodyText': 'Examen clinique OK.',
                      'finalizedAt': '2026-01-10T11:00:00Z',
                    },
                  ],
                },
              },
            ),
          );
          return;
        }
        if (options.method == 'POST' && path.contains('/consultation-shares')) {
          postedEmail = (options.data as Map?)?['email'] as String?;
          handler.resolve(
            Response(
              requestOptions: options,
              statusCode: 201,
              data: {
                'data': {'ok': true, 'expiresAt': '2026-07-30T12:00:00Z'},
              },
            ),
          );
          return;
        }
        handler.reject(
          DioException(
            requestOptions: options,
            error: 'unmocked ${options.method} $path',
          ),
        );
      },
    );
    ApiClient.instance.dio.interceptors.add(mockInterceptor);
  });

  tearDown(() {
    ApiClient.instance.dio.interceptors.remove(mockInterceptor);
    ApiClient.instance.userId = null;
  });

  testWidgets('timeline tap opens consultation then share', (tester) async {
    await pumpApp(
      tester,
      home: const PetTimelineScreen(petId: petId, petName: 'Bella', canWriteNotes: false),
    );
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 200));

    expect(find.byKey(const Key('visit_consultation_tile_$visitId')), findsOneWidget);
    await tester.tap(find.byKey(const Key('visit_consultation_tile_$visitId')));
    await tester.pumpAndSettle();

    expect(find.textContaining('Examen clinique OK'), findsOneWidget);
    await tester.tap(find.byKey(const Key('consultation_share_cta_$visitId')));
    await tester.pumpAndSettle();

    expect(find.byKey(const Key('consultation_share_dialog')), findsOneWidget);
    await tester.enterText(find.byKey(const Key('consultation_share_email')), 'vet.externe@petsfollow.test');
    await tester.tap(find.byKey(const Key('consultation_share_consent')));
    await tester.pump();
    await tester.tap(find.byKey(const Key('consultation_share_confirm')));
    await tester.pumpAndSettle();

    expect(postedEmail, 'vet.externe@petsfollow.test');
    expect(find.text(AppLocalizationsFr().sendConsultationSuccess), findsOneWidget);
  });

  testWidgets('consultation view loads reports', (tester) async {
    await pumpApp(
      tester,
      home: const ConsultationViewScreen(visitId: visitId, petName: 'Bella'),
    );
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 200));
    expect(find.textContaining('Dr Demo'), findsOneWidget);
    expect(find.byKey(const Key('consultation_share_cta_$visitId')), findsOneWidget);
  });
}
