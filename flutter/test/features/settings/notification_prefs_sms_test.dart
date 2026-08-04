import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:petsfollow_mobile/core/api/api_client.dart';
import 'package:petsfollow_mobile/core/models/notification_prefs.dart';
import 'package:petsfollow_mobile/features/settings/presentation/notification_preferences_screen.dart';

import '../../helpers/mock_api.dart';
import '../../helpers/pump_app.dart';

const _prefsPath = '/api/v1/me/notification-preferences';

Map<String, dynamic> _prefsPayload({bool sms = true}) => {
      'userId': 'user-1',
      'hr': true,
      'care': true,
      'visits': true,
      'messages': true,
      'discovery': true,
      'billing': true,
      'sms': sms,
    };

void main() {
  late MockApi mock;

  setUp(() {
    mock = MockApi();
    mock.install();
    // loadPrefs court-circuite l'appel réseau quand le token est nul.
    ApiClient.instance.token = 'test-token';
  });

  tearDown(() {
    mock.uninstall();
    ApiClient.instance.token = null;
  });

  testWidgets('settings_sms_pref_toggle envoie sms:false au serveur',
      (tester) async {
    Map<String, dynamic>? sentBody;
    mock.json('GET', _prefsPath, data: _prefsPayload());
    mock.on('PATCH', _prefsPath, (options) {
      sentBody = Map<String, dynamic>.from(options.data as Map);
      return mock.ok(options, _prefsPayload(sms: false));
    });

    await pumpApp(tester, home: const NotificationPreferencesScreen());
    await tester.pumpAndSettle();

    final toggle = find.byKey(const Key('settings_sms_pref_toggle'));
    await tester.scrollUntilVisible(
      toggle,
      200,
      scrollable: find.byType(Scrollable).first,
    );
    await tester.pumpAndSettle();
    expect(toggle, findsOneWidget);

    // Opt-out par défaut : le canal arrive activé.
    expect(tester.widget<SwitchListTile>(toggle).value, isTrue);

    await tester.tap(find.descendant(of: toggle, matching: find.byType(Switch)));
    await tester.pumpAndSettle();
    expect(tester.widget<SwitchListTile>(toggle).value, isFalse);

    await tester.tap(find.byType(FilledButton));
    await tester.pumpAndSettle();

    expect(sentBody, isNotNull);
    expect(sentBody!['sms'], isFalse);
    // Les autres canaux ne doivent pas être emportés par la bascule SMS.
    expect(sentBody!['visits'], isTrue);
    expect(sentBody!['messages'], isTrue);
  });

  testWidgets('un canal SMS coupé côté serveur (STOP) revient désactivé',
      (tester) async {
    mock.json('GET', _prefsPath, data: _prefsPayload(sms: false));

    await pumpApp(tester, home: const NotificationPreferencesScreen());
    await tester.pumpAndSettle();

    final toggle = find.byKey(const Key('settings_sms_pref_toggle'));
    await tester.scrollUntilVisible(
      toggle,
      200,
      scrollable: find.byType(Scrollable).first,
    );
    await tester.pumpAndSettle();
    expect(tester.widget<SwitchListTile>(toggle).value, isFalse);
  });

  test('NotificationPrefs sérialise le canal sms', () {
    const prefs = NotificationPrefs(userId: 'u', sms: false);
    expect(prefs.toJson()['sms'], isFalse);
    // Défaut opt-out : une réponse serveur sans le champ garde le canal actif.
    expect(NotificationPrefs.fromJson({'userId': 'u'}).sms, isTrue);
    expect(NotificationPrefs.fromJson({'userId': 'u', 'sms': false}).sms, isFalse);
    expect(prefs.copyWith(sms: true).sms, isTrue);
  });
}
