import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:petsfollow_mobile/features/support/presentation/support_report_screen.dart';
import 'package:shared_preferences/shared_preferences.dart';

import '../../helpers/mock_api.dart';
import '../../helpers/pump_app.dart';

void main() {
  late MockApi mock;

  setUp(() {
    SharedPreferences.setMockInitialValues({});
    mock = MockApi()..install();
  });

  tearDown(() {
    mock.uninstall();
  });

  testWidgets('support form submits ticket', (tester) async {
    var posted = false;
    mock.on('POST', r'/api/v1/support/tickets', (options) {
      posted = true;
      final body = options.data as Map;
      expect(body['source'], 'flutter_client');
      expect(body['subject'], 'Bug UI');
      expect(body['message'], 'Le bouton ne marche pas');
      expect(body['diagnostics'], isA<Map>());
      return mock.ok(options, {
        'id': 't1',
        'status': 'open',
        'subject': body['subject'],
      }, status: 201);
    });

    await pumpApp(
      tester,
      home: const SupportReportScreen(source: 'flutter_client'),
    );
    await tester.pumpAndSettle();

    await tester.enterText(find.byKey(const Key('support_subject')), 'Bug UI');
    await tester.enterText(
      find.byKey(const Key('support_message')),
      'Le bouton ne marche pas',
    );
    await tester.tap(find.byKey(const Key('support_submit')));
    await tester.pump(); // start submit
    await tester.pump(const Duration(milliseconds: 900));
    await tester.pumpAndSettle();

    expect(posted, isTrue);
  });
}
