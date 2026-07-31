import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:petsfollow_mobile/core/api/api_client.dart';
import 'package:petsfollow_mobile/features/heartrate/presentation/heart_rate_flow_screen.dart';
import 'package:petsfollow_mobile/l10n/app_localizations.dart';
import 'package:petsfollow_mobile/l10n/app_localizations_fr.dart';

void main() {
  late Interceptor mockInterceptor;
  int? startedDurationSec;

  setUp(() {
    startedDurationSec = null;
    ApiClient.instance.dio.interceptors.clear();
    mockInterceptor = InterceptorsWrapper(
      onRequest: (options, handler) {
        final path = options.path;
        if (options.method == 'POST' &&
            path.contains('/heartrate/sessions') &&
            !path.contains('/validate') &&
            !path.contains('/cancel') &&
            !path.contains('/seen')) {
          final body = options.data as Map<String, dynamic>?;
          final duration = body?['durationSec'] as int? ?? 30;
          startedDurationSec = duration;
          handler.resolve(Response(
            requestOptions: options,
            statusCode: 201,
            data: {
              'data': {
                'id': 'sess-1',
                'durationSec': duration,
                'status': 'in_progress',
              },
            },
          ));
          return;
        }
        if (options.method == 'POST' && path.contains('/cancel')) {
          handler.resolve(Response(
            requestOptions: options,
            statusCode: 200,
            data: {'data': {}},
          ));
          return;
        }
        if (options.method == 'PATCH' && path.contains('/heartrate/sessions/')) {
          final body = options.data as Map<String, dynamic>?;
          final taps = body?['tapCount'] as int? ?? 0;
          handler.resolve(Response(
            requestOptions: options,
            statusCode: 200,
            data: {
              'data': {
                'id': 'sess-1',
                'bpm': taps * 2,
                'tapCount': taps,
                'isAlert': false,
              },
            },
          ));
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
    ApiClient.instance.loadToken();
  });

  Future<void> pumpFlow(WidgetTester tester, {required List<int> durations}) async {
    await tester.pumpWidget(
      MaterialApp(
        locale: const Locale('fr'),
        localizationsDelegates: AppLocalizations.localizationsDelegates,
        supportedLocales: AppLocalizations.supportedLocales,
        home: HeartRateFlowScreen(
          petId: 'pet-1',
          durationsSec: durations,
          species: 'dog',
        ),
      ),
    );
    await tester.pumpAndSettle();
  }

  testWidgets('shows only practice-enabled duration chips', (tester) async {
    final l10n = AppLocalizationsFr();
    await pumpFlow(tester, durations: [15, 30]);

    expect(find.text(l10n.durationSeconds(15)), findsOneWidget);
    expect(find.text(l10n.durationSeconds(30)), findsOneWidget);
    expect(find.text(l10n.durationSeconds(60)), findsNothing);
  });

  testWidgets('start then taps increment beat count', (tester) async {
    final l10n = AppLocalizationsFr();
    await pumpFlow(tester, durations: [15, 30]);

    await tester.tap(find.text(l10n.start));
    // Do not pumpAndSettle — the 1s periodic timer would drain the whole session.
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 50));

    // Clinical default = longest enabled duration.
    expect(startedDurationSec, 30);

    expect(find.text(l10n.beatsCount(0)), findsOneWidget);
    expect(find.text(l10n.tapHere), findsOneWidget);

    final tapTarget = find.byKey(const Key('hr_tap_zone'));
    expect(tapTarget, findsOneWidget);

    await tester.tap(tapTarget);
    await tester.pump();
    expect(find.text(l10n.beatsCount(1)), findsOneWidget);

    // Debounce uses DateTime.now() (wall clock); runAsync escapes fake async.
    await tester.runAsync(() => Future<void>.delayed(const Duration(milliseconds: 160)));
    await tester.tap(tapTarget);
    await tester.pump();
    expect(find.text(l10n.beatsCount(2)), findsOneWidget);

    // Tear down while interceptor still registered (dispose cancels session via Dio).
    await tester.pumpWidget(const SizedBox.shrink());
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 1));
  });
}
