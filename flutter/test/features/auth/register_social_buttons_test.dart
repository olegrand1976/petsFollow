import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:petsfollow_mobile/core/api/api_errors.dart';
import 'package:petsfollow_mobile/core/auth/google_login_flow.dart';
import 'package:petsfollow_mobile/features/auth/presentation/register_screen.dart';
import 'package:petsfollow_mobile/l10n/app_localizations.dart';
import 'package:petsfollow_mobile/l10n/app_localizations_fr.dart';

DioException _dioError(Object? data) {
  final options = RequestOptions(path: '/api/v1/auth/google');
  return DioException(
    requestOptions: options,
    response: Response(requestOptions: options, statusCode: 400, data: data),
  );
}

Widget _wrapRegister() {
  return const MaterialApp(
    locale: Locale('fr'),
    localizationsDelegates: AppLocalizations.localizationsDelegates,
    supportedLocales: AppLocalizations.supportedLocales,
    home: RegisterScreen(),
  );
}

void main() {
  group('apiErrorCode', () {
    test('lit error.code dans la réponse enveloppée', () {
      final e = _dioError({
        'error': {'code': 'google_client_only', 'message': 'nope'},
      });
      expect(apiErrorCode(e), 'google_client_only');
    });

    test('retombe sur msgKey puis message', () {
      expect(
        apiErrorCode(_dioError({'error': {'msgKey': 'email_not_verified'}})),
        'email_not_verified',
      );
      expect(
        apiErrorCode(_dioError({'error': {'message': 'boom'}})),
        'boom',
      );
    });

    test('lit code à plat', () {
      expect(apiErrorCode(_dioError({'code': 'conflict'})), 'conflict');
    });

    test('null pour données inattendues ou erreur non Dio', () {
      expect(apiErrorCode(_dioError('plain text')), isNull);
      expect(apiErrorCode(_dioError(null)), isNull);
      expect(apiErrorCode(StateError('x')), isNull);
    });
  });

  group('GoogleLoginFlow.errorMessage', () {
    final l10n = AppLocalizationsFr();

    test('mappe les codes API vers les messages localisés', () {
      String messageFor(String code) => GoogleLoginFlow.errorMessage(
            l10n,
            _dioError({'error': {'code': code}}),
          );

      expect(messageFor('not_configured'), l10n.googleNotConfigured);
      expect(messageFor('google_client_only'), l10n.googleWrongAudience);
      expect(messageFor('google_pro_only'), l10n.googleWrongAudience);
      expect(messageFor('email_not_verified'), l10n.emailNotVerified);
      expect(messageFor('anything_else'), l10n.googleLoginFailed);
    });

    test('erreur non Dio → message générique', () {
      expect(
        GoogleLoginFlow.errorMessage(l10n, StateError('x')),
        l10n.googleLoginFailed,
      );
    });
  });

  group('RegisterScreen boutons sociaux', () {
    final l10n = AppLocalizationsFr();

    testWidgets('iOS : bouton Apple visible, tap affiche appleComingSoon',
        (tester) async {
      // L'invariant du binding exige un reset avant la fin du corps du test.
      debugDefaultTargetPlatformOverride = TargetPlatform.iOS;
      try {
        await tester.pumpWidget(_wrapRegister());
        await tester.pumpAndSettle();

        final appleButton = find.text(l10n.loginWithApple);
        expect(appleButton, findsOneWidget);
        expect(find.text(l10n.loginOr), findsOneWidget);
        // GOOGLE_SERVER_CLIENT_ID absent en test → bouton Google masqué.
        expect(find.text(l10n.loginWithGoogle), findsNothing);

        await tester.ensureVisible(appleButton);
        await tester.tap(appleButton);
        await tester.pump();
        expect(find.text(l10n.appleComingSoon), findsOneWidget);
      } finally {
        debugDefaultTargetPlatformOverride = null;
      }
    });

    testWidgets('Android : ni bouton Apple ni séparateur (Google non configuré)',
        (tester) async {
      debugDefaultTargetPlatformOverride = TargetPlatform.android;
      try {
        await tester.pumpWidget(_wrapRegister());
        await tester.pumpAndSettle();

        expect(find.text(l10n.loginWithApple), findsNothing);
        expect(find.text(l10n.loginOr), findsNothing);
        expect(find.text(l10n.loginWithGoogle), findsNothing);
        expect(
          find.widgetWithText(FilledButton, l10n.registerSubmit),
          findsOneWidget,
        );
      } finally {
        debugDefaultTargetPlatformOverride = null;
      }
    });
  });
}
