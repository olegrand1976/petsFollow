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
                    'reportStatus': 'final',
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
    expect(find.byKey(const Key('visit_consultation_cta_$visitId')), findsOneWidget);
    expect(find.byKey(const Key('timeline_section_consultations')), findsOneWidget);
    final consultationsTile = tester.widget<ExpansionTile>(
      find.byKey(const Key('timeline_section_consultations')),
    );
    expect(consultationsTile.initiallyExpanded, isTrue);
    final historyTile = tester.widget<ExpansionTile>(
      find.byKey(const Key('timeline_section_history')),
    );
    expect(historyTile.initiallyExpanded, isTrue);
    await tester.tap(find.byKey(const Key('visit_consultation_cta_$visitId')));
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

  testWidgets('shows consultation CTA on visit card in Historique', (tester) async {
    ApiClient.instance.dio.interceptors.remove(mockInterceptor);
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
                    'reportStatus': 'final',
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
              data: {
                'data': [
                  {
                    'id': visitId,
                    'type': 'visit',
                    'title': 'Visite',
                    'body': '',
                    'createdAt': '2026-01-10T10:00:00Z',
                    'meta': {
                      'visitId': visitId,
                      'hasReport': true,
                      'reportStatus': 'final',
                    },
                  },
                  {
                    'id': 'hr-1',
                    'type': 'heartrate',
                    'title': 'Relevé',
                    'body': 'BPM: 72',
                    'createdAt': '2026-01-09T10:00:00Z',
                    'meta': {},
                  },
                ],
              },
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
        handler.reject(
          DioException(
            requestOptions: options,
            error: 'unmocked ${options.method} $path',
          ),
        );
      },
    );
    ApiClient.instance.dio.interceptors.add(mockInterceptor);

    await pumpApp(
      tester,
      home: const PetTimelineScreen(petId: petId, petName: 'Bella', canWriteNotes: false),
    );
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 200));

    expect(find.byKey(const Key('timeline_section_consultations')), findsOneWidget);
    expect(find.byKey(const Key('visit_consultation_tile_$visitId')), findsOneWidget);
    expect(
      tester.widget<ExpansionTile>(find.byKey(const Key('timeline_section_consultations'))).initiallyExpanded,
      isTrue,
    );
    expect(
      tester.widget<ExpansionTile>(find.byKey(const Key('timeline_section_history'))).initiallyExpanded,
      isTrue,
    );
    // Visit card kept in Historique with the same CTA.
    expect(find.byKey(const Key('timeline_visit_report_$visitId')), findsOneWidget);
    expect(find.byKey(const Key('visit_consultation_cta_$visitId')), findsNWidgets(2));
    expect(find.byKey(const Key('timeline_item_hr-1')), findsOneWidget);
  });

  testWidgets('historique expands by default when no consultations', (tester) async {
    ApiClient.instance.dio.interceptors.remove(mockInterceptor);
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
            Response(requestOptions: options, statusCode: 200, data: {'data': <dynamic>[]}),
          );
          return;
        }
        if (options.method == 'GET' && path.contains('/pets/$petId/timeline')) {
          handler.resolve(
            Response(
              requestOptions: options,
              statusCode: 200,
              data: {
                'data': [
                  {
                    'id': 'hr-1',
                    'type': 'heartrate',
                    'title': 'Relevé',
                    'body': 'BPM: 72',
                    'createdAt': '2026-01-09T10:00:00Z',
                    'meta': {},
                  },
                ],
              },
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
        handler.reject(
          DioException(
            requestOptions: options,
            error: 'unmocked ${options.method} $path',
          ),
        );
      },
    );
    ApiClient.instance.dio.interceptors.add(mockInterceptor);

    await pumpApp(
      tester,
      home: const PetTimelineScreen(petId: petId, petName: 'Bella', canWriteNotes: false),
    );
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 200));

    expect(find.byKey(const Key('timeline_section_consultations')), findsNothing);
    expect(
      tester.widget<ExpansionTile>(find.byKey(const Key('timeline_section_history'))).initiallyExpanded,
      isTrue,
    );
    expect(find.byKey(const Key('timeline_item_hr-1')), findsOneWidget);
  });

  testWidgets('historique visit opens CR when ListVisits lacks hasFinalReport', (tester) async {
    ApiClient.instance.dio.interceptors.remove(mockInterceptor);
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
                    'hasFinalReport': false,
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
              data: {
                'data': [
                  {
                    'id': visitId,
                    'type': 'visit',
                    'title': 'Visite',
                    'body': '',
                    'createdAt': '2026-01-10T10:00:00Z',
                    'meta': {
                      'visitId': visitId,
                      'hasReport': true,
                      'reportStatus': 'final',
                    },
                  },
                ],
              },
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
        handler.reject(
          DioException(
            requestOptions: options,
            error: 'unmocked ${options.method} $path',
          ),
        );
      },
    );
    ApiClient.instance.dio.interceptors.add(mockInterceptor);

    await pumpApp(
      tester,
      home: const PetTimelineScreen(petId: petId, petName: 'Bella', canWriteNotes: false),
    );
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 200));

    expect(find.byKey(const Key('visit_consultation_tile_$visitId')), findsNothing);
    expect(find.byKey(const Key('timeline_section_history')), findsOneWidget);
    expect(find.byKey(const Key('timeline_visit_report_$visitId')), findsOneWidget);
    expect(find.byKey(const Key('visit_consultation_cta_$visitId')), findsOneWidget);
    await tester.tap(find.byKey(const Key('visit_consultation_cta_$visitId')));
    await tester.pumpAndSettle();
    expect(find.textContaining('Examen clinique OK'), findsOneWidget);
  });

  testWidgets('draft consultation shows disabled pending CTA', (tester) async {
    ApiClient.instance.dio.interceptors.remove(mockInterceptor);
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
                    'hasFinalReport': false,
                    'reportStatus': 'draft',
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
              data: {
                'data': [
                  {
                    'id': visitId,
                    'type': 'visit',
                    'title': 'Visite',
                    'body': '',
                    'createdAt': '2026-01-10T10:00:00Z',
                    'meta': {
                      'visitId': visitId,
                      'hasReport': false,
                      'reportStatus': 'draft',
                    },
                  },
                ],
              },
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
          handler.reject(
            DioException(
              requestOptions: options,
              error: 'should not open draft consultation',
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

    await pumpApp(
      tester,
      home: const PetTimelineScreen(petId: petId, petName: 'Bella', canWriteNotes: false),
    );
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 200));

    expect(find.byKey(const Key('visit_consultation_tile_$visitId')), findsOneWidget);
    // Consultations tile + Historique visit card.
    expect(find.byKey(const Key('visit_consultation_pending_$visitId')), findsNWidgets(2));
    expect(find.byKey(const Key('visit_consultation_cta_$visitId')), findsNothing);
    expect(find.byKey(const Key('timeline_visit_pending_$visitId')), findsOneWidget);
    expect(find.byKey(const Key('timeline_visit_report_$visitId')), findsNothing);

    final pending = tester.widgetList<TextButton>(
      find.byKey(const Key('visit_consultation_pending_$visitId')),
    );
    expect(pending, isNotEmpty);
    for (final btn in pending) {
      expect(btn.onPressed, isNull);
    }
    await tester.tap(find.byKey(const Key('visit_consultation_pending_$visitId')).first);
    await tester.pumpAndSettle();
    expect(find.textContaining('Examen clinique'), findsNothing);
  });

  testWidgets('historique visit without CR is not tappable', (tester) async {
    ApiClient.instance.dio.interceptors.remove(mockInterceptor);
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
                    'hasFinalReport': false,
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
              data: {
                'data': [
                  {
                    'id': visitId,
                    'type': 'visit',
                    'title': 'Visite',
                    'body': 'visite',
                    'createdAt': '2026-01-10T10:00:00Z',
                    'meta': {
                      'visitId': visitId,
                      'hasReport': false,
                    },
                  },
                ],
              },
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
        handler.reject(
          DioException(
            requestOptions: options,
            error: 'unmocked ${options.method} $path',
          ),
        );
      },
    );
    ApiClient.instance.dio.interceptors.add(mockInterceptor);

    await pumpApp(
      tester,
      home: const PetTimelineScreen(petId: petId, petName: 'Bella', canWriteNotes: false),
    );
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 200));

    expect(find.byKey(const Key('timeline_item_$visitId')), findsOneWidget);
    expect(find.byKey(const Key('visit_consultation_cta_$visitId')), findsNothing);
    expect(find.byKey(const Key('visit_consultation_pending_$visitId')), findsNothing);
    await tester.tap(find.byKey(const Key('timeline_item_$visitId')));
    await tester.pump();
    expect(find.textContaining('Aucun compte-rendu disponible'), findsNothing);
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
