import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:petsfollow_mobile/core/api/api_client.dart';
import 'package:petsfollow_mobile/core/config/app_env.dart';
import 'package:petsfollow_mobile/features/client_ai/presentation/consultation_explain_screen.dart';
import 'package:petsfollow_mobile/features/pets/presentation/consultation_view_screen.dart';
import 'package:petsfollow_mobile/l10n/app_localizations_fr.dart';

import '../../helpers/pump_app.dart';

void main() {
  const visitId = 'visit-explain-1';
  late Interceptor mockInterceptor;

  setUp(() {
    AppEnv.debugClientAiOverride = true;
    ApiClient.instance.dio.interceptors.clear();
    ApiClient.instance.userId = 'user-1';
    mockInterceptor = InterceptorsWrapper(
      onRequest: (options, handler) {
        final path =
            options.uri.path.isNotEmpty ? options.uri.path : options.path;
        if (options.method == 'GET' &&
            path.contains('/client-consultation/explain')) {
          handler.resolve(
            Response(
              requestOptions: options,
              statusCode: 200,
              data: {
                'data': {
                  'visitId': visitId,
                  'cached': false,
                  'disclaimer': 'Pas un avis médical.',
                  'cards': [
                    {
                      'title': 'Souffle',
                      'body': 'Un souffle grade 2 est souvent léger.',
                      'kind': 'term',
                    },
                  ],
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
                  'petName': 'Bella',
                  'practiceName': 'VetPlus',
                  'status': 'done',
                  'reports': [
                    {
                      'id': 'r1',
                      'authorName': 'Dr Demo',
                      'bodyText': 'Souffle systolique grade 2/6.',
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
  });

  tearDown(() {
    AppEnv.debugClientAiOverride = null;
    ApiClient.instance.dio.interceptors.remove(mockInterceptor);
    ApiClient.instance.userId = null;
  });

  testWidgets('consultation view shows explain CTA and opens cards', (tester) async {
    await pumpApp(
      tester,
      home: const ConsultationViewScreen(visitId: visitId, petName: 'Bella'),
    );
    await tester.pumpAndSettle();

    expect(find.byKey(const Key('consultation_explain_cta_$visitId')), findsOneWidget);
    await tester.tap(find.byKey(const Key('consultation_explain_cta_$visitId')));
    await tester.pumpAndSettle();

    expect(find.text(AppLocalizationsFr().clientAiExplainTitle), findsOneWidget);
    expect(find.byKey(const Key('client_ai_explain_card_term')), findsOneWidget);
    expect(find.byKey(const Key('client_ai_explain_disclaimer')), findsOneWidget);
    expect(find.textContaining('souffle grade 2'), findsOneWidget);
  });

  testWidgets('explain screen loads cards directly', (tester) async {
    await pumpApp(
      tester,
      home: const ConsultationExplainScreen(visitId: visitId, petName: 'Bella'),
    );
    await tester.pumpAndSettle();
    expect(find.byKey(const Key('client_ai_explain_dev_badge')), findsOneWidget);
    expect(find.text('Souffle'), findsOneWidget);
  });
}
