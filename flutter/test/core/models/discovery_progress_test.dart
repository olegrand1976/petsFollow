import 'package:flutter_test/flutter_test.dart';
import 'package:petsfollow_mobile/core/models/discovery_progress.dart';

void main() {
  final start = DateTime(2026, 1, 1);

  DiscoveryProgress progress({List<String> completed = const []}) => DiscoveryProgress(
        userId: 'u1',
        startedAt: start,
        completedCards: completed,
      );

  test('day0 is always unlocked', () {
    expect(progress().isCardUnlocked(0), isTrue);
  });

  test('day2 locked until day0 completed', () {
    expect(progress().isCardUnlocked(2), isFalse);
    expect(progress(completed: ['day0']).isCardUnlocked(2), isTrue);
  });

  test('day4 locked until day2 completed', () {
    expect(progress(completed: ['day0']).isCardUnlocked(4), isFalse);
    expect(progress(completed: ['day0', 'day2']).isCardUnlocked(4), isTrue);
  });

  test('full journey unlocks same calendar day', () {
    final p = progress(completed: ['day0', 'day2', 'day4']);
    expect(p.isCardUnlocked(6, start), isTrue);
    expect(p.daysSinceStart(start), 0);
  });
}
