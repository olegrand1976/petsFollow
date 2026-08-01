import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:petsfollow_mobile/core/api/api_client.dart';
import 'package:petsfollow_mobile/features/pets/presentation/lab_panels_screen.dart';
import 'package:petsfollow_mobile/l10n/app_localizations.dart';
import 'package:petsfollow_mobile/l10n/app_localizations_fr.dart';

void main() {
  const petId = 'pet-labs-1';
  const panelId = 'panel-1';
  late Interceptor mockInterceptor;
  var detailGets = 0;

  setUp(() {
    detailGets = 0;
    ApiClient.instance.dio.interceptors.clear();
    mockInterceptor = InterceptorsWrapper(
      onRequest: (options, handler) {
        final path = options.path;
        if (options.method == 'GET' &&
            path.endsWith('/pets/$petId/lab-panels')) {
          handler.resolve(
            Response(
              requestOptions: options,
              statusCode: 200,
              data: {
                'data': [
                  {
                    'id': panelId,
                    'labName': 'BioVet Test',
                    'collectedAt': '2026-07-01T10:00:00Z',
                    'abnormalCount': 1,
                  },
                ],
              },
            ),
          );
          return;
        }
        if (options.method == 'GET' &&
            path.contains('/pets/$petId/lab-panels/$panelId')) {
          detailGets++;
          handler.resolve(
            Response(
              requestOptions: options,
              statusCode: 200,
              data: {
                'data': {
                  'id': panelId,
                  'labName': 'BioVet Test',
                  'collectedAt': '2026-07-01T10:00:00Z',
                  'results': [
                    {
                      'analyteCode': 'crea',
                      'valueNum': 2.4,
                      'unit': 'mg/dL',
                      'flag': 'high',
                    },
                    {
                      'analyteCode': 'alat',
                      'valueText': 'N',
                      'unit': 'U/L',
                      'flag': 'normal',
                    },
                  ],
                },
              },
            ),
          );
          return;
        }
        if (options.method == 'GET' &&
            path.contains('/pets/$petId/documents')) {
          handler.resolve(
            Response(
              requestOptions: options,
              statusCode: 200,
              data: {'data': []},
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
  });

  testWidgets('lists panels and opens detail on tap', (tester) async {
    final l10n = AppLocalizationsFr();
    await tester.pumpWidget(
      MaterialApp(
        locale: const Locale('fr'),
        localizationsDelegates: AppLocalizations.localizationsDelegates,
        supportedLocales: AppLocalizations.supportedLocales,
        home: const LabPanelsScreen(petId: petId),
      ),
    );
    await tester.pumpAndSettle();

    expect(find.byKey(const Key('lab_panel_$panelId')), findsOneWidget);
    expect(find.text('BioVet Test'), findsOneWidget);

    await tester.tap(find.byKey(const Key('lab_panel_$panelId')));
    await tester.pumpAndSettle();

    expect(detailGets, 1);
    expect(find.text(l10n.labsResults), findsOneWidget);
    expect(find.text('CREA'), findsOneWidget);
    expect(find.textContaining('2.4'), findsOneWidget);
    expect(find.text('ALAT'), findsOneWidget);
    expect(find.textContaining('N'), findsOneWidget);
  });

  testWidgets('initialPanelId opens detail after list load', (tester) async {
    await tester.pumpWidget(
      MaterialApp(
        locale: const Locale('fr'),
        localizationsDelegates: AppLocalizations.localizationsDelegates,
        supportedLocales: AppLocalizations.supportedLocales,
        home: const LabPanelsScreen(
          petId: petId,
          initialPanelId: panelId,
        ),
      ),
    );
    await tester.pumpAndSettle();

    expect(detailGets, 1);
    expect(find.text('CREA'), findsOneWidget);
    expect(find.textContaining('2.4'), findsOneWidget);
  });
}
