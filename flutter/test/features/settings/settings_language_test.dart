import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:petsfollow_mobile/core/api/api_client.dart';
import 'package:petsfollow_mobile/core/locale/locale_controller.dart';
import 'package:petsfollow_mobile/features/settings/presentation/settings_menu_screen.dart';
import 'package:petsfollow_mobile/l10n/app_localizations_fr.dart';
import 'package:shared_preferences/shared_preferences.dart';

import '../../helpers/mock_api.dart';
import '../../helpers/pump_app.dart';

void main() {
  late MockApi mock;
  String? patchedLocale;

  setUp(() async {
    SharedPreferences.setMockInitialValues({});
    await LocaleController.instance.setLocale('fr');
    patchedLocale = null;
    mock = MockApi()..install();
    ApiClient.instance.token = 'test-token';
    ApiClient.instance.userId = 'user-1';
    mock.on('PATCH', r'/api/v1/me/locale', (options) {
      patchedLocale = (options.data as Map)['locale']?.toString();
      return mock.ok(options, {'ok': true});
    });
  });

  tearDown(() {
    mock.uninstall();
    ApiClient.instance.token = null;
    ApiClient.instance.userId = null;
  });

  testWidgets('settings language sheet lists all 8 locales and can pick uk',
      (tester) async {
    final l10n = AppLocalizationsFr();
    await pumpApp(
      tester,
      home: SettingsMenuScreen(onLogout: () {}),
    );
    await tester.pumpAndSettle();

    final tile = find.byKey(const Key('settings_language'));
    expect(tile, findsOneWidget);
    await tester.tap(tile);
    await tester.pumpAndSettle();

    expect(find.byKey(const Key('language_picker_list')), findsOneWidget);
    for (final code in LocaleController.supportedCodes) {
      expect(
        find.byKey(Key('language_option_$code')),
        findsOneWidget,
        reason: 'option $code must be reachable (scrollable sheet)',
      );
    }
    expect(find.text(l10n.languageUk), findsOneWidget);
    expect(find.text(l10n.languageRu), findsOneWidget);

    final ukOption = find.byKey(const Key('language_option_uk'));
    await tester.scrollUntilVisible(
      ukOption,
      80,
      scrollable: find.descendant(
        of: find.byKey(const Key('language_picker_list')),
        matching: find.byType(Scrollable),
      ),
    );
    await tester.pumpAndSettle();
    await tester.tap(ukOption);
    await tester.pumpAndSettle();

    expect(patchedLocale, 'uk');
    expect(LocaleController.instance.languageCode, 'uk');
  });
}
