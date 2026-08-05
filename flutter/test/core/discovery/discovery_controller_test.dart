import 'package:flutter_test/flutter_test.dart';
import 'package:petsfollow_mobile/core/api/api_client.dart';
import 'package:petsfollow_mobile/core/config/app_env.dart';
import 'package:petsfollow_mobile/core/discovery/discovery_controller.dart';
import 'package:shared_preferences/shared_preferences.dart';

import '../../helpers/mock_api.dart';

void main() {
  late MockApi mock;
  final completedPosts = <String>[];

  setUp(() async {
    SharedPreferences.setMockInitialValues({});
    completedPosts.clear();
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

  test('load merges local completions that remote is missing and posts them', () async {
    final sp = await SharedPreferences.getInstance();
    await sp.setStringList('pf_discovery_completed_user-1', ['day0', 'day2', 'day4', 'day6']);
    await sp.setString(
      'pf_discovery_started_at_user-1',
      DateTime.now().toUtc().toIso8601String(),
    );

    mock.json('GET', '/api/v1/me', data: {'id': 'user-1', 'email': 'a@b.c'});
    mock.json('GET', '/api/v1/me/discovery', data: {
      'userId': 'user-1',
      'startedAt': DateTime.now().toUtc().toIso8601String(),
      'completedCards': ['day0', 'day2'],
      'streakDays': 2,
    });
    mock.on('POST', '/api/v1/me/discovery/complete', (options) {
      final key = (options.data as Map)['cardKey']?.toString() ?? '';
      completedPosts.add(key);
      final all = {'day0', 'day2', ...completedPosts};
      return mock.ok(options, {
        'userId': 'user-1',
        'startedAt': DateTime.now().toUtc().toIso8601String(),
        'completedCards': all.toList(),
        'streakDays': all.length,
      });
    });

    final progress = await DiscoveryController.instance.load();
    expect(progress.isJourneyComplete, isTrue);
    expect(completedPosts, containsAll(['day4', 'day6']));
  });

  test('staging closes legacy seed journey older than 36h', () async {
    expect(AppEnv.isStaging, isTrue); // tests default to staging canal
    final started = DateTime.now().toUtc().subtract(const Duration(days: 2));
    mock.json('GET', '/api/v1/me', data: {'id': 'user-1', 'email': 'a@b.c'});
    mock.json('GET', '/api/v1/me/discovery', data: {
      'userId': 'user-1',
      'startedAt': started.toIso8601String(),
      'completedCards': ['day0', 'day2'],
      'streakDays': 2,
    });
    mock.on('POST', '/api/v1/me/discovery/complete', (options) {
      final key = (options.data as Map)['cardKey']?.toString() ?? '';
      completedPosts.add(key);
      final all = {'day0', 'day2', ...completedPosts};
      return mock.ok(options, {
        'userId': 'user-1',
        'startedAt': started.toIso8601String(),
        'completedCards': all.toList(),
        'streakDays': all.length,
      });
    });

    final progress = await DiscoveryController.instance.load();
    expect(progress.isJourneyComplete, isTrue);
    expect(completedPosts, containsAll(['day4', 'day6']));
  });

  test('staging does not auto-close when journey already past day2', () async {
    final started = DateTime.now().toUtc().subtract(const Duration(days: 2));
    mock.json('GET', '/api/v1/me', data: {'id': 'user-1', 'email': 'a@b.c'});
    mock.json('GET', '/api/v1/me/discovery', data: {
      'userId': 'user-1',
      'startedAt': started.toIso8601String(),
      'completedCards': ['day0', 'day2', 'day4'],
      'streakDays': 3,
    });
    mock.on('POST', '/api/v1/me/discovery/complete', (options) {
      completedPosts.add((options.data as Map)['cardKey']?.toString() ?? '');
      return mock.ok(options, {
        'userId': 'user-1',
        'startedAt': started.toIso8601String(),
        'completedCards': ['day0', 'day2', 'day4', 'day6'],
        'streakDays': 4,
      });
    });

    final progress = await DiscoveryController.instance.load();
    expect(progress.isJourneyComplete, isFalse);
    expect(completedPosts, isEmpty);
  });
}
