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

  void stubHomeBase({
    required List<String> completedCards,
    List<Map<String, dynamic>>? timeline,
  }) {
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
      'completedCards': completedCards,
      'streakDays': completedCards.length,
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
    mock.json(
      'GET',
      '/api/v1/pets/pet-1/timeline',
      data: timeline ??
          [
            {
              'id': 'hr-1',
              'type': 'heartrate',
              'title': '28 rpm',
              'body': '',
              'createdAt': '2026-08-04T10:00:00Z',
            },
          ],
    );
  }

  testWidgets('home hides discovery and shows recent activity when journey done',
      (tester) async {
    stubHomeBase(completedCards: ['day0', 'day2', 'day4', 'day6']);

    await pumpApp(tester, home: const HomeTab());
    await tester.pumpAndSettle();

    expect(find.text(l10n.discoveryTitle), findsNothing);
    expect(find.byKey(const Key('home_recent_activity_title')), findsOneWidget);
    expect(find.text(l10n.homeRecentActivity), findsOneWidget);
    expect(find.text('Rex'), findsWidgets);
    expect(find.textContaining('28 rpm'), findsOneWidget);
  });

  testWidgets('home shows discovery when journey incomplete', (tester) async {
    stubHomeBase(completedCards: ['day0', 'day2']);

    await pumpApp(tester, home: const HomeTab());
    await tester.pumpAndSettle();

    final discovery = find.text(l10n.discoveryTitle);
    await tester.scrollUntilVisible(
      discovery,
      200,
      scrollable: find.byType(Scrollable).first,
    );
    await tester.pumpAndSettle();
    expect(discovery, findsOneWidget);
    expect(find.byKey(const Key('home_recent_activity_title')), findsNothing);
  });
}
