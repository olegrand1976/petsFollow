import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:petsfollow_mobile/core/auth/pending_login_hint.dart';
import 'package:petsfollow_mobile/core/deeplink/payment_deeplink.dart';
import 'package:petsfollow_mobile/features/auth/presentation/confirm_email_screen.dart';
import 'package:petsfollow_mobile/features/auth/presentation/login_screen.dart';
import 'package:petsfollow_mobile/l10n/app_localizations.dart';
import 'package:petsfollow_mobile/l10n/app_localizations_fr.dart';

void main() {
  tearDown(PendingLoginHint.take);

  test('PendingLoginHint take clears values', () {
    PendingLoginHint.set(email: 'a@test.com', infoMessage: 'ok');
    expect(PendingLoginHint.hasPending, isTrue);
    final first = PendingLoginHint.take();
    expect(first.email, 'a@test.com');
    expect(first.infoMessage, 'ok');
    expect(PendingLoginHint.hasPending, isFalse);
    final second = PendingLoginHint.take();
    expect(second.email, isNull);
    expect(second.infoMessage, isNull);
  });

  test('parseConfirmEmailDeepLink token and statusOk', () {
    final withToken = parseConfirmEmailDeepLink(
      Uri.parse('petsfollow://confirm-email?token=abc123'),
    );
    expect(withToken.kind, ConfirmEmailDeepLinkKind.token);
    expect(withToken.token, 'abc123');

    final statusOk = parseConfirmEmailDeepLink(
      Uri.parse('petsfollow://confirm-email?status=ok&email=a%40b.test'),
    );
    expect(statusOk.kind, ConfirmEmailDeepLinkKind.statusOk);
    expect(statusOk.email, 'a@b.test');

    final ignored = parseConfirmEmailDeepLink(
      Uri.parse('petsfollow://confirm-email'),
    );
    expect(ignored.kind, ConfirmEmailDeepLinkKind.none);

    final other = parseConfirmEmailDeepLink(
      Uri.parse('petsfollow://reset-password?token=x'),
    );
    expect(other.kind, ConfirmEmailDeepLinkKind.none);
  });

  test('onLoginHint setter flushes pending hint (cold start)', () {
    PendingLoginHint.set(email: 'cold@test.com', infoMessage: 'hi');
    var flushed = 0;
    AppDeepLink.instance.onLoginHint = () => flushed++;
    expect(flushed, 1);
    AppDeepLink.instance.onLoginHint = null;
  });

  testWidgets('ConfirmEmailScreen shows loading then invalid for empty token', (tester) async {
    final l10n = AppLocalizationsFr();
    await tester.pumpWidget(
      MaterialApp(
        locale: const Locale('fr'),
        localizationsDelegates: AppLocalizations.localizationsDelegates,
        supportedLocales: AppLocalizations.supportedLocales,
        home: const ConfirmEmailScreen(token: ''),
      ),
    );
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 50));
    expect(find.text(l10n.confirmEmailInvalidLink), findsOneWidget);
    expect(find.text(l10n.confirmEmailBackToLogin), findsOneWidget);
  });

  testWidgets('LoginScreen maps emailNotVerified key exists', (tester) async {
    final l10n = AppLocalizationsFr();
    expect(l10n.emailNotVerified, contains('email'));
    expect(l10n.registerCta, "S'inscrire");

    PendingLoginHint.set(email: 'new@petsfollow.test', infoMessage: l10n.confirmEmailDoneSubtitle);
    await tester.pumpWidget(
      MaterialApp(
        locale: const Locale('fr'),
        localizationsDelegates: AppLocalizations.localizationsDelegates,
        supportedLocales: AppLocalizations.supportedLocales,
        home: LoginScreen(onLoggedIn: () {}),
      ),
    );
    await tester.pumpAndSettle();

    expect(find.text(l10n.confirmEmailDoneSubtitle), findsOneWidget);
    expect(find.text('new@petsfollow.test'), findsOneWidget);
  });
}
