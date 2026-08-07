import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:petsfollow_mobile/core/api/api_client.dart';
import 'package:petsfollow_mobile/core/discovery/discovery_controller.dart';
import 'package:petsfollow_mobile/features/home/presentation/home_tab.dart';
import 'package:petsfollow_mobile/l10n/app_localizations_fr.dart';
import 'package:shared_preferences/shared_preferences.dart';

import '../../helpers/fixtures.dart';
import '../../helpers/mock_api.dart';
import '../../helpers/pump_app.dart';

void main() {
  late MockApi mock;
  final l10n = AppLocalizationsFr();

  setUp(() async {
    SharedPreferences.setMockInitialValues({});
    mock = MockApi()..install();
    ApiClient.instance.token = 'test-token';
    ApiClient.instance.userId = 'user-1';
    DiscoveryController.instance.bindUser('user-1');
    await DiscoveryController.instance.clearLocal();
    DiscoveryController.instance.bindUser('user-1');
  });

  tearDown(() async {
    await DiscoveryController.instance.clearLocal();
    mock.uninstall();
    ApiClient.instance.token = null;
    ApiClient.instance.userId = null;
  });

  void stubHomeBase() {
    mock.json('GET', '/api/v1/me', data: {
      ...Fixtures.me(),
      'fullName': 'Demo Client',
    });
    mock.json('GET', '/api/v1/me/feature-modules', data: {
      'moduleCarePlus': false,
      'moduleHorse': false,
      'moduleKennel': false,
      'moduleFamily': false,
    });
    mock.json('GET', '/api/v1/me/discovery', data: {
      'userId': 'user-1',
      'startedAt': DateTime.now().toUtc().toIso8601String(),
      'completedCards': ['day0', 'day2', 'day4', 'day6'],
      'streakDays': 4,
    });
    mock.json('GET', '/api/v1/me/vets', data: [
      {
        'practiceId': 'p-1',
        'vetEmail': 'vet.demo@petsfollow.test',
        'vetFullName': 'Dr Demo',
        'practiceName': 'VetPlus',
      },
    ]);
    mock.json('GET', '/api/v1/pets', data: [Fixtures.pet()]);
    mock.json('GET', '/api/v1/pets/pet-1/timeline', data: const []);
  }

  testWidgets('home shows pet tips section when API returns items', (tester) async {
    stubHomeBase();
    mock.json('GET', '/api/v1/me/pet-tips', data: {
      'items': [
        {
          'id': 'heat_safety',
          'species': ['dog', 'cat'],
          'priority': 'high',
          'title': 'Chaleur : prudence',
          'body': 'Ne laissez jamais votre animal dans une voiture.',
        },
      ],
    });

    await pumpApp(tester, home: const HomeTab());
    await tester.pumpAndSettle();

    expect(find.byKey(const Key('home_pet_tips')), findsOneWidget);
    expect(find.byKey(const Key('home_pet_tips_title')), findsOneWidget);
    expect(find.text(l10n.homePetTipsTitle), findsOneWidget);
    expect(find.byKey(const Key('home_pet_tips_item_heat_safety')), findsOneWidget);
    expect(find.text('Chaleur : prudence'), findsOneWidget);
  });

  testWidgets('home hides pet tips when API fails', (tester) async {
    stubHomeBase();
    mock.on('GET', '/api/v1/me/pet-tips', (options) => mock.err(options, status: 500));

    await pumpApp(tester, home: const HomeTab());
    await tester.pumpAndSettle();

    expect(find.byKey(const Key('home_pet_tips')), findsNothing);
  });
}
