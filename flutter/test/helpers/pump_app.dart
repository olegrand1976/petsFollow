import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:petsfollow_mobile/l10n/app_localizations.dart';

/// Minimal MaterialApp with FR l10n for widget tests.
Future<void> pumpApp(
  WidgetTester tester, {
  required Widget home,
  Locale locale = const Locale('fr'),
}) async {
  await tester.pumpWidget(
    MaterialApp(
      locale: locale,
      localizationsDelegates: AppLocalizations.localizationsDelegates,
      supportedLocales: AppLocalizations.supportedLocales,
      home: home,
    ),
  );
  await tester.pump();
}
