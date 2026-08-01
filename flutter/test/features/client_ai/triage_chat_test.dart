import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:petsfollow_mobile/core/api/api_client.dart';
import 'package:petsfollow_mobile/core/config/app_env.dart';
import 'package:petsfollow_mobile/core/models/pet.dart';
import 'package:petsfollow_mobile/features/client_ai/presentation/triage_chat_screen.dart';
import 'package:petsfollow_mobile/l10n/app_localizations_fr.dart';

import '../../helpers/pump_app.dart';

void main() {
  const sessionId = 'sess-1';
  const petId = 'pet-1';
  late Interceptor mockInterceptor;
  String? postedBody;

  setUp(() {
    AppEnv.debugClientAiOverride = true;
    postedBody = null;
    ApiClient.instance.dio.interceptors.clear();
    ApiClient.instance.userId = 'user-1';
    mockInterceptor = InterceptorsWrapper(
      onRequest: (options, handler) {
        final path =
            options.uri.path.isNotEmpty ? options.uri.path : options.path;
        if (options.method == 'POST' &&
            path.endsWith('/client-ai/triage/sessions')) {
          handler.resolve(
            Response(
              requestOptions: options,
              statusCode: 201,
              data: {
                'data': {
                  'session': {'id': sessionId, 'petId': petId},
                  'petName': 'Bella',
                  'escalation': {
                    'petId': petId,
                    'practiceId': 'prac-1',
                    'practicePhone': '0123456789',
                    'canMessage': true,
                    'canBookVisit': true,
                  },
                  'messages': [],
                },
              },
            ),
          );
          return;
        }
        if (options.method == 'POST' && path.contains('/messages')) {
          postedBody = (options.data as Map?)?['body'] as String?;
          handler.resolve(
            Response(
              requestOptions: options,
              statusCode: 200,
              data: {
                'data': {
                  'userMessage': {
                    'id': 'm1',
                    'role': 'user',
                    'body': postedBody,
                  },
                  'assistantMessage': {
                    'id': 'm2',
                    'role': 'assistant',
                    'body': 'Danger potentiel — contactez le cabinet.',
                    'level': 'red',
                  },
                  'level': 'red',
                  'watchSigns': <String>[],
                  'recommendedAction': 'Appeler',
                  'escalation': {
                    'petId': petId,
                    'practiceId': 'prac-1',
                    'practicePhone': '0123456789',
                    'canMessage': true,
                    'canBookVisit': true,
                  },
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

  testWidgets('triage send shows red CTAs', (tester) async {
    await pumpApp(
      tester,
      home: TriageChatScreen(
        initialPetId: petId,
        initialPetName: 'Bella',
        petsOverride: const [
          Pet(
            id: petId,
            name: 'Bella',
            species: 'dog',
            breed: '',
            ownerUserId: 'user-1',
          ),
        ],
      ),
    );
    await tester.pumpAndSettle();

    expect(find.byKey(const Key('client_ai_triage_composer')), findsOneWidget);
    await tester.enterText(
      find.byKey(const Key('client_ai_triage_composer')),
      'Mon chien a mangé du chocolat',
    );
    await tester.tap(find.byKey(const Key('client_ai_triage_send')));
    await tester.pumpAndSettle();

    expect(postedBody, 'Mon chien a mangé du chocolat');
    expect(find.byKey(const Key('client_ai_triage_level_red')), findsOneWidget);
    expect(find.byKey(const Key('client_ai_triage_call')), findsOneWidget);
    expect(find.byKey(const Key('client_ai_triage_book')), findsOneWidget);
    expect(find.byKey(const Key('client_ai_triage_message')), findsOneWidget);
    expect(find.text(AppLocalizationsFr().clientAiTriageLevelRed), findsWidgets);
  });
}
