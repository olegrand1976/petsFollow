import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:petsfollow_mobile/core/api/api_client.dart';
import 'package:petsfollow_mobile/features/shell/presentation/commercial_field_shell_screen.dart';
import 'package:petsfollow_mobile/l10n/app_localizations_fr.dart';

import '../../helpers/mock_api.dart';
import '../../helpers/pump_app.dart';

void main() {
  late MockApi mock;

  setUp(() {
    mock = MockApi();
    mock.json('GET', '/api/v1/me/app-invite', data: {
      'code': 'demo',
      'proSiteUrl': 'https://example.test',
    });
    mock.json('GET', '/api/v1/commercial-manager/overview', data: {
      'teamProspectsTotal': 8,
      'teamProspectsConverted': 2,
      'teamProspectsContacted': 4,
      'teamAppointmentsUpcoming': 1,
      'teamStaleInPipeline': 0,
      'teamMonthEarnedCents': 1200,
      'conversionRateBps': 5000,
      'directoryTotal': 0,
      'team': [
        {
          'userId': 'rep-1',
          'fullName': 'Camille Rep',
          'email': 'camille@test',
          'prospectsConverted': 2,
          'staleInPipeline': 0,
          'monthEarnedCents': 1200,
        },
      ],
      'self': {
        'assignedVets': 1,
        'prospectsTotal': 2,
        'prospectsConverted': 0,
        'monthEarnedCents': 0,
      },
    });
    mock.install();
  });

  tearDown(() {
    mock.uninstall();
    ApiClient.instance.userRole = null;
  });

  /// Flush Dio zero-duration timers scheduled by shell initState.
  Future<void> flushApi(WidgetTester tester) async {
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 50));
  }

  testWidgets('manager sees team CTA; commercial does not', (tester) async {
    ApiClient.instance.userRole = 'commercial';
    await pumpApp(
      tester,
      home: CommercialFieldShellScreen(onLogout: () {}),
    );
    await flushApi(tester);
    expect(find.byKey(const Key('commercial_manager_team')), findsNothing);

    ApiClient.instance.userRole = 'commercial_manager';
    await pumpApp(
      tester,
      home: CommercialFieldShellScreen(onLogout: () {}),
    );
    await flushApi(tester);
    expect(find.byKey(const Key('commercial_manager_team')), findsOneWidget);
    expect(find.text(AppLocalizationsFr().managerTeamCta), findsOneWidget);
  });

  testWidgets('manager team CTA opens overview with member row', (tester) async {
    ApiClient.instance.userRole = 'commercial_manager';
    await pumpApp(
      tester,
      home: CommercialFieldShellScreen(onLogout: () {}),
    );
    await flushApi(tester);

    await tester.tap(find.byKey(const Key('commercial_manager_team')));
    // Avoid pumpAndSettle: CircularProgressIndicator never settles while loading.
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 50));

    expect(find.text(AppLocalizationsFr().managerTeamTitle), findsOneWidget);

    final row = find.byKey(
      const Key('manager_team_row_rep-1'),
      skipOffstage: false,
    );
    await tester.ensureVisible(row);
    await tester.pump();
    expect(row, findsOneWidget);
    expect(find.text('Camille Rep'), findsOneWidget);
  });
}
