import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:petsfollow_mobile/core/api/api_client.dart';
import 'package:petsfollow_mobile/core/auth/google_auth.dart';
import 'package:petsfollow_mobile/features/auth/presentation/login_screen.dart';
import 'package:petsfollow_mobile/l10n/app_localizations_fr.dart';
import 'package:shared_preferences/shared_preferences.dart';

import '../../helpers/mock_api.dart';
import '../../helpers/pump_app.dart';

void main() {
  late MockApi mock;
  final l10n = AppLocalizationsFr();

  setUp(() async {
    SharedPreferences.setMockInitialValues({});
    mock = MockApi()..install();
    ApiClient.instance.token = null;
    ApiClient.instance.userId = null;
    GoogleAuth.debugForceConfigured = null;
  });

  tearDown(() {
    GoogleAuth.debugForceConfigured = null;
    mock.uninstall();
    ApiClient.instance.token = null;
    ApiClient.instance.userId = null;
  });

  testWidgets('login hides Google button when not configured', (tester) async {
    GoogleAuth.debugForceConfigured = false;
    await pumpApp(tester, home: LoginScreen(onLoggedIn: () {}));
    await tester.pumpAndSettle();
    expect(find.byKey(const Key('login_google')), findsNothing);
  });

  testWidgets('login shows Google button when configured', (tester) async {
    GoogleAuth.debugForceConfigured = true;
    await pumpApp(tester, home: LoginScreen(onLoggedIn: () {}));
    await tester.pumpAndSettle();
    expect(find.byKey(const Key('login_google')), findsOneWidget);
    expect(find.text(l10n.loginWithGoogle), findsOneWidget);
  });

  test('loginWithGoogle posts audience=client and consent', () async {
    Map<String, dynamic>? body;
    mock.on('POST', '/api/v1/auth/google', (options) {
      body = Map<String, dynamic>.from(options.data as Map);
      // MFA challenge skips _completeLogin (secure storage / getMe side effects).
      return mock.ok(options, {
        'requires2FA': true,
        'mfaToken': 'mfa-tok',
        'expiresIn': 300,
      });
    });

    final data =
        await ApiClient.instance.loginWithGoogle('fake-id-token', consent: true);

    expect(body, isNotNull);
    expect(body!['idToken'], 'fake-id-token');
    expect(body!['audience'], 'client');
    expect(body!['consent'], true);
    expect(data['requires2FA'], true);
  });
}
