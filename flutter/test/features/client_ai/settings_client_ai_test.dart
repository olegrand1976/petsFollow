import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:petsfollow_mobile/core/api/api_client.dart';
import 'package:petsfollow_mobile/core/config/app_env.dart';
import 'package:petsfollow_mobile/features/client_ai/presentation/explain_reports_list_screen.dart';
import 'package:petsfollow_mobile/features/settings/presentation/settings_menu_screen.dart';
import 'package:petsfollow_mobile/l10n/app_localizations_fr.dart';
import 'package:shared_preferences/shared_preferences.dart';

import '../../helpers/pump_app.dart';

void main() {
  setUp(() {
    SharedPreferences.setMockInitialValues({});
    AppEnv.debugClientAiOverride = true;
    ApiClient.instance.dio.interceptors.clear();
    ApiClient.instance.userId = 'user-1';
  });

  tearDown(() {
    AppEnv.debugClientAiOverride = null;
    ApiClient.instance.dio.interceptors.clear();
  });

  testWidgets('settings shows Assistance IA section and explain entry when flag on',
      (tester) async {
    await pumpApp(
      tester,
      home: SettingsMenuScreen(onLogout: () {}),
    );
    await tester.pumpAndSettle();

    final section = find.byKey(const Key('settings_client_ai_section'));
    await tester.scrollUntilVisible(
      section,
      300,
      scrollable: find.byType(Scrollable).first,
    );
    await tester.pumpAndSettle();
    expect(section, findsOneWidget);
    expect(find.byKey(const Key('settings_client_ai_explain')), findsOneWidget);
    expect(find.byKey(const Key('settings_client_ai_triage')), findsOneWidget);
    expect(find.text(AppLocalizationsFr().clientAiSectionTitle), findsOneWidget);
    expect(find.text(AppLocalizationsFr().clientAiExplainListTitle), findsOneWidget);
  });

  testWidgets('settings explain entry opens list of finalized reports', (tester) async {
    ApiClient.instance.dio.interceptors.add(
      InterceptorsWrapper(
        onRequest: (options, handler) {
          final path =
              options.uri.path.isNotEmpty ? options.uri.path : options.path;
          if (options.method == 'GET' && path.endsWith('/pets')) {
            handler.resolve(
              Response(
                requestOptions: options,
                statusCode: 200,
                data: {
                  'data': [
                    {'id': 'pet-1', 'name': 'Rex', 'species': 'dog'},
                  ],
                },
              ),
            );
            return;
          }
          if (options.method == 'GET' && path.contains('/pets/pet-1/visits')) {
            handler.resolve(
              Response(
                requestOptions: options,
                statusCode: 200,
                data: {
                  'data': [
                    {
                      'id': 'visit-1',
                      'petId': 'pet-1',
                      'status': 'done',
                      'hasFinalReport': true,
                      'reportStatus': 'final',
                      'scheduledAt': '2026-07-30T10:00:00Z',
                    },
                  ],
                },
              ),
            );
            return;
          }
          handler.reject(
            DioException(
              requestOptions: options,
              error: 'unexpected ${options.method} $path',
            ),
          );
        },
      ),
    );

    await pumpApp(
      tester,
      home: SettingsMenuScreen(onLogout: () {}),
    );
    await tester.pumpAndSettle();

    final explain = find.byKey(const Key('settings_client_ai_explain'));
    await tester.scrollUntilVisible(
      explain,
      300,
      scrollable: find.byType(Scrollable).first,
    );
    await tester.tap(explain);
    await tester.pumpAndSettle();

    expect(find.byType(ExplainReportsListScreen), findsOneWidget);
    expect(
      find.byKey(const Key('client_ai_explain_list_item_visit-1')),
      findsOneWidget,
    );
    expect(find.text('Rex'), findsOneWidget);
  });

  testWidgets('settings hides Client AI when flag off', (tester) async {
    AppEnv.debugClientAiOverride = false;
    await pumpApp(
      tester,
      home: SettingsMenuScreen(onLogout: () {}),
    );
    await tester.pumpAndSettle();
    expect(find.byKey(const Key('settings_client_ai_section')), findsNothing);
    expect(find.byKey(const Key('settings_client_ai_explain')), findsNothing);
    expect(find.byKey(const Key('settings_client_ai_triage')), findsNothing);
  });
}
