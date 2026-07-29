import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:petsfollow_mobile/features/auth/presentation/accept_terms_screen.dart';
import 'package:petsfollow_mobile/l10n/app_localizations_fr.dart';

import '../../helpers/mock_api.dart';
import '../../helpers/pump_app.dart';

void main() {
  late MockApi mock;

  setUp(() {
    mock = MockApi();
    mock.install();
  });

  tearDown(() => mock.uninstall());

  testWidgets('accept_terms_submit requires consent checkbox', (tester) async {
    var accepted = false;
    await pumpApp(
      tester,
      home: AcceptTermsScreen(onAccepted: () => accepted = true),
    );
    await tester.pumpAndSettle();

    await tester.tap(find.byKey(const Key('accept_terms_submit')));
    await tester.pump();

    expect(accepted, isFalse);
    expect(find.text(AppLocalizationsFr().registerConsentRequired), findsOneWidget);
  });

  testWidgets('accept_terms_submit posts consent and calls onAccepted', (tester) async {
    var posts = 0;
    mock.on('POST', r'/api/v1/me/accept-terms', (options) {
      posts++;
      final body = options.data as Map?;
      expect(body?['consent'], isTrue);
      return mock.ok(options, {
        'termsAcceptedAt': '2026-07-29T10:00:00Z',
      });
    });

    var accepted = false;
    await pumpApp(
      tester,
      home: AcceptTermsScreen(onAccepted: () => accepted = true),
    );
    await tester.pumpAndSettle();

    await tester.tap(find.byKey(const Key('accept_terms_consent')));
    await tester.pump();
    await tester.tap(find.byKey(const Key('accept_terms_submit')));
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 50));

    expect(posts, 1);
    expect(accepted, isTrue);
  });
}
