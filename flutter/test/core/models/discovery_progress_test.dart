import 'package:flutter_test/flutter_test.dart';
import 'package:petsfollow_mobile/core/models/discovery_progress.dart';

void main() {
  final start = DateTime(2026, 1, 1);

  DiscoveryProgress progress({List<String> completed = const []}) => DiscoveryProgress(
        userId: 'u1',
        startedAt: start,
        completedCards: completed,
      );

  test('stage 1 is always unlocked', () {
    expect(progress().isCardUnlocked(0), isTrue);
  });

  test('next stage unlocks as soon as previous is completed', () {
    expect(progress().isCardUnlocked(2), isFalse);
    expect(progress(completed: ['day0']).isCardUnlocked(2), isTrue);

    expect(progress(completed: ['day0']).isCardUnlocked(4), isFalse);
    expect(progress(completed: ['day0', 'day2']).isCardUnlocked(4), isTrue);

    expect(progress(completed: ['day0', 'day2', 'day4']).isCardUnlocked(6), isTrue);
  });

  test('full journey can complete on the same calendar day', () {
    final p = progress(completed: ['day0', 'day2', 'day4']);
    expect(p.isCardUnlocked(6), isTrue);
    expect(p.isCardCompleted(0), isTrue);
    expect(p.isCardCompleted(6), isFalse);
  });

  test('isJourneyComplete only when all four stages are done', () {
    expect(progress().isJourneyComplete, isFalse);
    expect(progress(completed: ['day0', 'day2', 'day4']).isJourneyComplete, isFalse);
    expect(
      progress(completed: ['day0', 'day2', 'day4', 'day6']).isJourneyComplete,
      isTrue,
    );
  });
}
