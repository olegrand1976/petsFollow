import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:petsfollow_mobile/core/theme/theme_controller.dart';
import 'package:petsfollow_mobile/features/settings/presentation/settings_menu_screen.dart';
import 'package:shared_preferences/shared_preferences.dart';

import '../../helpers/pump_app.dart';

void main() {
  setUp(() {
    SharedPreferences.setMockInitialValues({});
  });

  testWidgets('settings_appearance toggles ThemeController', (tester) async {
    await ThemeController.instance.load();
    expect(ThemeController.instance.isDark, isTrue);

    await pumpApp(
      tester,
      home: SettingsMenuScreen(onLogout: () {}),
    );
    await tester.pumpAndSettle();

    final tile = find.byKey(const Key('settings_appearance'));
    await tester.scrollUntilVisible(
      tile,
      200,
      scrollable: find.byType(Scrollable).first,
    );
    await tester.pumpAndSettle();
    expect(tile, findsOneWidget);

    await tester.tap(find.descendant(of: tile, matching: find.byType(Switch)));
    await tester.pumpAndSettle();

    expect(ThemeController.instance.isDark, isFalse);

    final prefs = await SharedPreferences.getInstance();
    expect(prefs.getString('pf_theme'), 'light');

    // Restore default for other tests.
    await ThemeController.instance.setDark(true);
  });
}
