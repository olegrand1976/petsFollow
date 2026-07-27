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

  /// Journey step indices (narrative labels may still say "day N").
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

  int daysSinceStart([DateTime? now]) {
    final today = now ?? DateTime.now();
    final start = DateTime(startedAt.year, startedAt.month, startedAt.day);
    final current = DateTime(today.year, today.month, today.day);
    return current.difference(start).inDays;
  }

  /// Sequential unlock: card N opens once the previous journey card is completed.
  /// Calendar days no longer gate progression.
  bool isCardUnlocked(int dayIndex, [DateTime? now]) {
    final i = journeyDays.indexOf(dayIndex);
    if (i <= 0) return true;
    return isCardCompleted(journeyDays[i - 1]);
  }

  bool isCardCompleted(int dayIndex) => completedCards.contains(cardKeyForDay(dayIndex));
}
