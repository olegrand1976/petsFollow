import 'package:flutter_test/flutter_test.dart';
import 'package:petsfollow_mobile/core/api/api_client.dart';
import 'package:shared_preferences/shared_preferences.dart';

import '../../helpers/mock_api.dart';

/// Changer son mot de passe incrémente token_version côté API : toutes les
/// sessions du compte tombent au prochain refresh. L'API réémet une paire pour
/// l'appareil demandeur, que le client doit adopter — sinon changer son mot de
/// passe revient à se déconnecter soi-même.
void main() {
  TestWidgetsFlutterBinding.ensureInitialized();

  late MockApi mock;

  setUp(() {
    SharedPreferences.setMockInitialValues(<String, Object>{});
    mock = MockApi();
    mock.install();
    ApiClient.instance.token = 'old-access';
    ApiClient.instance.refreshToken = 'old-refresh';
  });

  tearDown(() => mock.uninstall());

  test('adopte la paire réémise par l\'API', () async {
    mock.json('PATCH', '/api/v1/me/password', data: <String, dynamic>{
      'ok': true,
      'accessToken': 'new-access',
      'refreshToken': 'new-refresh',
    });

    await ApiClient.instance.changePassword('old-pwd', 'new-pwd');

    expect(ApiClient.instance.token, 'new-access');
    expect(ApiClient.instance.refreshToken, 'new-refresh');
  });

  // Toute enveloppe n'est pas une Map<String, dynamic> : un cast typé direct
  // planterait ici alors qu'il passe sur un littéral Dart.
  test('supporte une enveloppe à clés dynamiques', () async {
    final dynamicKeyed = <dynamic, dynamic>{
      'data': <dynamic, dynamic>{
        'ok': true,
        'accessToken': 'decoded-access',
        'refreshToken': 'decoded-refresh',
      },
    };

    mock.on('PATCH', '/api/v1/me/password', (options) => mock.ok(options, dynamicKeyed));

    await ApiClient.instance.changePassword('old-pwd', 'new-pwd');

    expect(ApiClient.instance.token, 'decoded-access');
    expect(ApiClient.instance.refreshToken, 'decoded-refresh');
  });

  test('garde la session en place si l\'API ne réémet rien', () async {
    mock.json('PATCH', '/api/v1/me/password', data: <String, dynamic>{'ok': true});

    await ApiClient.instance.changePassword('old-pwd', 'new-pwd');

    expect(ApiClient.instance.token, 'old-access');
    expect(ApiClient.instance.refreshToken, 'old-refresh');
  });

  // Réémission impossible côté API : la session est morte de toute façon,
  // autant couper ici plutôt qu'au prochain refresh, sans rapport apparent.
  test('coupe la session quand l\'API signale reauthRequired', () async {
    mock.json('PATCH', '/api/v1/me/password', data: <String, dynamic>{
      'ok': true,
      'reauthRequired': true,
    });

    await ApiClient.instance.changePassword('old-pwd', 'new-pwd');

    expect(ApiClient.instance.token, isNull);
    expect(ApiClient.instance.refreshToken, isNull);
  });
}
