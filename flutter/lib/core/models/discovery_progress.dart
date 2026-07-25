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

  bool isCardUnlocked(int dayIndex, [DateTime? now]) => daysSinceStart(now) >= dayIndex;

  bool isCardCompleted(int dayIndex) => completedCards.contains(cardKeyForDay(dayIndex));
}
