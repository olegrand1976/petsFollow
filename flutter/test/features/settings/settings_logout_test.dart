import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:petsfollow_mobile/features/settings/presentation/settings_menu_screen.dart';
import 'package:shared_preferences/shared_preferences.dart';

import '../../helpers/pump_app.dart';

void main() {
  setUp(() {
    SharedPreferences.setMockInitialValues({});
  });

  testWidgets('settings_logout key is present in menu', (tester) async {
    await pumpApp(
      tester,
      home: SettingsMenuScreen(onLogout: () {}),
    );
    await tester.pumpAndSettle();

    final logout = find.byKey(const Key('settings_logout'));
    await tester.scrollUntilVisible(
      logout,
      300,
      scrollable: find.byType(Scrollable).first,
    );
    await tester.pumpAndSettle();
    expect(logout, findsOneWidget);
  });
}
