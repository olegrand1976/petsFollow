class DiscoveryProgress {
  const DiscoveryProgress({
    required this.userId,
    required this.startedAt,
    this.completedCards = const [],
    this.streakDays = 0,
  });

  final String userId;
  final DateTime startedAt;
  final List<String> completedCards;
  final int streakDays;

  /// Stage indices (API keys remain `day0`/`day2`/`day4`/`day6` for compat).
  static const journeyDays = [0, 2, 4, 6];

  factory DiscoveryProgress.fromJson(Map<String, dynamic> json) {
    final raw = json['completedCards'];
    return DiscoveryProgress(
      userId: json['userId'] as String? ?? '',
      startedAt: DateTime.tryParse(json['startedAt'] as String? ?? '') ?? DateTime.now(),
      completedCards: raw is List ? raw.map((e) => e.toString()).toList() : const [],
      streakDays: json['streakDays'] as int? ?? 0,
    );
  }

  static String cardKeyForDay(int dayIndex) => 'day$dayIndex';

  /// Next stage unlocks as soon as the previous stage is completed (no calendar wait).
  bool isCardUnlocked(int dayIndex) {
    final i = journeyDays.indexOf(dayIndex);
    if (i <= 0) return true;
    return isCardCompleted(journeyDays[i - 1]);
  }

  bool isCardCompleted(int dayIndex) => completedCards.contains(cardKeyForDay(dayIndex));

  /// True when every journey stage has been marked complete.
  bool get isJourneyComplete => journeyDays.every(isCardCompleted);
}
