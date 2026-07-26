import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:petsfollow_mobile/features/auth/presentation/login_screen.dart';
import 'package:petsfollow_mobile/l10n/app_localizations_fr.dart';

import '../../helpers/fixtures.dart';
import '../../helpers/mock_api.dart';
import '../../helpers/pump_app.dart';

void main() {
  late MockApi mock;

  setUp(() {
    mock = MockApi();
    mock.install();
  });

  tearDown(() => mock.uninstall());

  testWidgets('login_submit shows error on bad credentials', (tester) async {
    mock.on('POST', r'/api/v1/auth/login', (options) {
      return mock.err(
        options,
        status: 401,
        code: 'invalid_credentials',
        message: 'bad',
      );
    });

    var loggedIn = false;
    await pumpApp(
      tester,
      home: LoginScreen(onLoggedIn: () => loggedIn = true),
    );
    await tester.pumpAndSettle();

    await tester.enterText(
      find.byKey(const Key('login_email')),
      'bad@test.com',
    );
    await tester.enterText(
      find.byKey(const Key('login_password')),
      'wrong',
    );
    await tester.tap(find.byKey(const Key('login_submit')));
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 50));

    expect(loggedIn, isFalse);
    expect(find.byKey(const Key('login_error')), findsOneWidget);
    expect(find.text(AppLocalizationsFr().loginFailed), findsOneWidget);
  });

  testWidgets('login_submit keys are present', (tester) async {
    mock.json('POST', r'/api/v1/auth/login', data: Fixtures.loginSuccess());
    await pumpApp(
      tester,
      home: LoginScreen(onLoggedIn: () {}),
    );
    await tester.pumpAndSettle();
    expect(find.byKey(const Key('login_email')), findsOneWidget);
    expect(find.byKey(const Key('login_password')), findsOneWidget);
    expect(find.byKey(const Key('login_submit')), findsOneWidget);
  });

  testWidgets('email_not_verified shows resend confirmation action', (tester) async {
    var resendCalls = 0;
    mock.on('POST', r'/api/v1/auth/login', (options) {
      return mock.err(
        options,
        status: 403,
        code: 'email_not_verified',
        message: 'unverified',
      );
    });
    mock.on('POST', r'/api/v1/auth/resend-confirmation', (options) {
      resendCalls++;
      return mock.ok(options, {'message': 'ok'});
    });

    await pumpApp(
      tester,
      home: LoginScreen(onLoggedIn: () {}),
    );
    await tester.pumpAndSettle();

    await tester.enterText(
      find.byKey(const Key('login_email')),
      'unverified@test.com',
    );
    await tester.enterText(
      find.byKey(const Key('login_password')),
      'ClientPass123!',
    );
    await tester.tap(find.byKey(const Key('login_submit')));
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 50));

    expect(find.text(AppLocalizationsFr().emailNotVerified), findsOneWidget);
    expect(find.byKey(const Key('login_resend_confirmation')), findsOneWidget);

    await tester.tap(find.byKey(const Key('login_resend_confirmation')));
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 50));

    expect(resendCalls, 1);
    expect(find.text(AppLocalizationsFr().resendConfirmationSent), findsOneWidget);
  });
}
